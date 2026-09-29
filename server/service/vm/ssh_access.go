package vm

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/utils"
	"NanoKVM-Server/utils/unixcrypt"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Where SSH access lives on the board, and how it survives a slot switch.
//
// /root/.ssh is a directory bind from /data/identity-system/root-ssh, made by
// S02identity before sshd starts. A file renamed into it stays on /data, so the
// atomic write below persists a key with no further step. On a board where the
// bind failed, S02identity falls back to a copy, and saveSSHIdentity (its
// `save`) writes that copy back.
//
// /etc/kvm is a directory bind from /data/identity, like ssh_stop beside it, so
// the keys-only flag is on /data too. The drop-in sshd reads is on the slot's
// own root filesystem, which the next image replaces, so S50sshd writes it
// again from the flag at every start. The flag is the setting; the drop-in is
// how this slot's sshd hears about it.
//
// Package variables so tests can point them at a scratch tree.
var (
	authorizedKeysPath = "/root/.ssh/authorized_keys"
	sshKeysOnlyFlag    = "/etc/kvm/ssh_keys_only"
	sshKeysOnlyDropIn  = "/etc/ssh/sshd_config.d/ironkvm-keys-only.conf"
	sshHostKeyGlob     = "/etc/ssh/ssh_host_*_key.pub"
	shadowPath         = "/etc/shadow"
	sshdPidFile        = "/var/run/sshd.pid"

	sshdEffectiveConfig = readSSHDConfig
	reloadSSHD          = hangUpSSHD
	saveSSHIdentity     = utils.SaveIdentity

	// sshMu serialises every change to the keys file and the keys-only
	// setting, since each checks one against the other.
	sshMu sync.Mutex
)

const sshdBinary = "/usr/sbin/sshd"

// keysOnlyConfig must match what S50sshd writes at boot. sshd takes the first
// value it reads for a keyword, and the Include of sshd_config.d is the first
// line of the image's sshd_config, so this wins over the main file.
const keysOnlyConfig = "# Written by IronKVM: Settings > SSH > Keys only.\n" +
	"PasswordAuthentication no\n" +
	"KbdInteractiveAuthentication no\n"

// maxKeyInput bounds what the add endpoint reads. An RSA 16384 key is under
// 3 KiB; anything near this is not a key.
const maxKeyInput = 16 * 1024

// defaultRootPasswords are the factory passwords a board may still carry. The
// Sipeed images set root's password to "root"; an empty field is checked
// separately.
var defaultRootPasswords = []string{"root"}

var (
	errKeyInvalid     = errors.New("not a valid public key")
	errKeyOptions     = errors.New("keys with options are not accepted")
	errKeyMultiple    = errors.New("add one key at a time")
	errKeyDuplicate   = errors.New("this key is already authorized")
	errKeyNotFound    = errors.New("no key with this fingerprint")
	errLastKey        = errors.New("the last key cannot be removed while keys-only login is on")
	errNoKeys         = errors.New("add an authorized key before turning off password login")
	errNotHonoured    = errors.New("sshd on this image does not read the keys-only setting")
	errReloadFailed   = errors.New("saved, but sshd could not be reloaded; it applies at the next start")
	errKeysFileFailed = errors.New("failed to write the authorized keys")
)

// authorizedLine is one line of authorized_keys. key is nil for a line that
// is not a key: a comment, a blank line, or something sshd would skip. Those
// are written back unchanged.
type authorizedLine struct {
	text string
	key  *proto.SSHKey
}

func parseAuthorizedLine(text string) authorizedLine {
	line := authorizedLine{text: text}
	trimmed := strings.TrimSpace(text)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return line
	}
	pub, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(trimmed))
	if err != nil {
		return line
	}
	line.key = &proto.SSHKey{
		Type:        pub.Type(),
		Comment:     comment,
		Fingerprint: ssh.FingerprintSHA256(pub),
	}
	return line
}

// readAuthorizedKeys reads the file a line at a time. A missing file is an
// empty one.
func readAuthorizedKeys() ([]authorizedLine, error) {
	data, err := os.ReadFile(authorizedKeysPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var lines []authorizedLine
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, parseAuthorizedLine(scanner.Text()))
	}
	return lines, scanner.Err()
}

func keysOf(lines []authorizedLine) []proto.SSHKey {
	keys := []proto.SSHKey{}
	for _, line := range lines {
		if line.key != nil {
			keys = append(keys, *line.key)
		}
	}
	return keys
}

func writeAuthorizedKeys(lines []authorizedLine) error {
	var buf bytes.Buffer
	for _, line := range lines {
		buf.WriteString(line.text)
		buf.WriteByte('\n')
	}

	dir := filepath.Dir(authorizedKeysPath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	// sshd's StrictModes refuses a key file under a directory others can
	// write, and 0700 is what ssh itself makes.
	if err := os.Chmod(dir, 0o700); err != nil {
		return err
	}
	return writeFileAtomic(authorizedKeysPath, buf.Bytes(), 0o600)
}

// writeFileAtomic replaces path by renaming a synced temporary file over it,
// so a power cut leaves the old file or the new one and never half of either.
// The temporary file is made in the same directory, which on /root/.ssh keeps
// the rename inside the bind from /data.
func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer func() { _ = os.Remove(name) }()

	if _, err = tmp.Write(data); err == nil {
		err = tmp.Chmod(mode)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = os.Rename(name, path); err != nil {
		return err
	}

	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		_ = d.Close()
	}
	return nil
}

// normaliseKey parses what the owner pasted and returns the line to store:
// the key re-encoded, with its comment.
func normaliseKey(input string) (string, *proto.SSHKey, error) {
	input = strings.TrimSpace(input)
	if input == "" || len(input) > maxKeyInput {
		return "", nil, errKeyInvalid
	}
	if strings.ContainsAny(input, "\r\n") {
		return "", nil, errKeyMultiple
	}

	pub, comment, options, _, err := ssh.ParseAuthorizedKey([]byte(input))
	if err != nil {
		return "", nil, errKeyInvalid
	}
	if len(options) > 0 {
		return "", nil, errKeyOptions
	}

	line := strings.TrimSuffix(string(ssh.MarshalAuthorizedKey(pub)), "\n")
	if comment != "" {
		line += " " + comment
	}
	return line, &proto.SSHKey{
		Type:        pub.Type(),
		Comment:     comment,
		Fingerprint: ssh.FingerprintSHA256(pub),
	}, nil
}

func addAuthorizedKey(input string) (*proto.SSHKey, error) {
	text, key, err := normaliseKey(input)
	if err != nil {
		return nil, err
	}

	sshMu.Lock()
	defer sshMu.Unlock()

	lines, err := readAuthorizedKeys()
	if err != nil {
		return nil, err
	}
	for _, existing := range keysOf(lines) {
		if existing.Fingerprint == key.Fingerprint {
			return nil, errKeyDuplicate
		}
	}

	lines = append(lines, authorizedLine{text: text, key: key})
	if err := writeAuthorizedKeys(lines); err != nil {
		return nil, err
	}
	persistSSHIdentity()
	return key, nil
}

func deleteAuthorizedKey(fingerprint string) error {
	sshMu.Lock()
	defer sshMu.Unlock()

	lines, err := readAuthorizedKeys()
	if err != nil {
		return err
	}

	kept := lines[:0:0]
	removed := 0
	for _, line := range lines {
		if line.key != nil && line.key.Fingerprint == fingerprint {
			removed++
			continue
		}
		kept = append(kept, line)
	}
	if removed == 0 {
		return errKeyNotFound
	}
	if keysOnlyEnabled() && len(keysOf(kept)) == 0 {
		return errLastKey
	}

	if err := writeAuthorizedKeys(kept); err != nil {
		return err
	}
	persistSSHIdentity()
	return nil
}

// persistSSHIdentity writes the keys back to /data on a board whose
// /root/.ssh bind failed. With the bind live it has nothing to do. A failure
// does not fail the request: the key is in place for this boot either way.
func persistSSHIdentity() {
	if err := saveSSHIdentity(); err != nil {
		log.Warnf("authorized keys changed but the identity write-back failed: %s", err)
	}
}

func keysOnlyEnabled() bool {
	_, err := os.Stat(sshKeysOnlyFlag)
	return err == nil
}

func setKeysOnly(enabled bool) error {
	sshMu.Lock()
	defer sshMu.Unlock()

	if enabled {
		lines, err := readAuthorizedKeys()
		if err != nil {
			return err
		}
		if len(keysOf(lines)) == 0 {
			return errNoKeys
		}

		if err := os.MkdirAll(filepath.Dir(sshKeysOnlyDropIn), 0o755); err != nil {
			return err
		}
		if err := writeFileAtomic(sshKeysOnlyDropIn, []byte(keysOnlyConfig), 0o644); err != nil {
			return err
		}
		// Ask sshd what it would do with the drop-in in place. This also
		// proves the configuration parses, which matters: SIGHUP makes sshd
		// re-execute itself, and one that cannot read its configuration
		// exits and takes the listener with it.
		config, err := sshdEffectiveConfig()
		if err != nil || config["passwordauthentication"] != "no" {
			_ = os.Remove(sshKeysOnlyDropIn)
			if err != nil {
				log.Errorf("sshd -T failed with the keys-only drop-in: %s", err)
			}
			return errNotHonoured
		}
		if err := os.WriteFile(sshKeysOnlyFlag, nil, 0o644); err != nil {
			_ = os.Remove(sshKeysOnlyDropIn)
			return err
		}
	} else {
		// The flag first: it is what the next boot reads.
		if err := os.Remove(sshKeysOnlyFlag); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err := os.Remove(sshKeysOnlyDropIn); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	if err := reloadSSHD(); err != nil {
		log.Errorf("failed to reload sshd: %s", err)
		return errReloadFailed
	}
	return nil
}

// readSSHDConfig asks sshd for its effective configuration. Keys come back
// lower case; for a keyword given more than once, such as port, the first.
func readSSHDConfig() (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, sshdBinary, "-T").Output()
	if err != nil {
		return nil, err
	}
	return parseSSHDConfig(out), nil
}

func parseSSHDConfig(out []byte) map[string]string {
	config := map[string]string{}
	for _, line := range strings.Split(string(out), "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), " ")
		if !ok {
			continue
		}
		key = strings.ToLower(key)
		if _, seen := config[key]; !seen {
			config[key] = strings.TrimSpace(value)
		}
	}
	return config
}

// sshdPid returns the listener's pid, or 0 when it is not running.
func sshdPid() int {
	data, err := os.ReadFile(sshdPidFile)
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 1 {
		return 0
	}
	if err := syscall.Kill(pid, 0); err != nil {
		return 0
	}
	return pid
}

// hangUpSSHD makes the listening sshd re-read its configuration. SIGHUP goes
// to the listener only: every open session is its own process and keeps
// running. An sshd that is not running reads the drop-in when it starts.
func hangUpSSHD() error {
	pid := sshdPid()
	if pid == 0 {
		return nil
	}
	return syscall.Kill(pid, syscall.SIGHUP)
}

func readHostKeys() []proto.SSHKey {
	keys := []proto.SSHKey{}
	paths, _ := filepath.Glob(sshHostKeyGlob)
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		pub, _, _, _, err := ssh.ParseAuthorizedKey(data)
		if err != nil {
			continue
		}
		keys = append(keys, proto.SSHKey{Type: pub.Type(), Fingerprint: ssh.FingerprintSHA256(pub)})
	}
	return keys
}

// rootPasswordState reads root's entry in /etc/shadow and says whether it is a
// factory password. It returns only the verdict; the hash stays here.
func rootPasswordState() string {
	data, err := os.ReadFile(shadowPath)
	if err != nil {
		return "unknown"
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.SplitN(line, ":", 3)
		if len(fields) < 2 || fields[0] != "root" {
			continue
		}
		return classifyRootHash(fields[1])
	}
	return "unknown"
}

func classifyRootHash(hash string) string {
	if hash == "" {
		return "empty"
	}
	// A locked account, or one with no password login at all.
	if strings.HasPrefix(hash, "!") || strings.HasPrefix(hash, "*") {
		return "set"
	}
	for _, password := range defaultRootPasswords {
		ok, err := unixcrypt.Verify(password, hash)
		if err != nil {
			return "unknown"
		}
		if ok {
			return "default"
		}
	}
	return "set"
}

func sshPort(config map[string]string) int {
	if port, err := strconv.Atoi(config["port"]); err == nil && port > 0 {
		return port
	}
	return 22
}

func (s *Service) GetSSHKeys(c *gin.Context) {
	var rsp proto.Response

	lines, err := readAuthorizedKeys()
	if err != nil {
		log.Errorf("failed to read authorized keys: %s", err)
		rsp.ErrRsp(c, -1, "failed to read the authorized keys")
		return
	}
	rsp.OkRspWithData(c, &proto.GetSSHKeysRsp{Keys: keysOf(lines)})
}

func (s *Service) AddSSHKey(c *gin.Context) {
	var rsp proto.Response
	var req proto.AddSSHKeyReq
	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	key, err := addAuthorizedKey(req.Key)
	switch {
	case errors.Is(err, errKeyInvalid), errors.Is(err, errKeyMultiple):
		rsp.ErrRsp(c, -2, err.Error())
	case errors.Is(err, errKeyOptions):
		rsp.ErrRsp(c, -3, err.Error())
	case errors.Is(err, errKeyDuplicate):
		rsp.ErrRsp(c, -4, err.Error())
	case err != nil:
		log.Errorf("failed to add an authorized key: %s", err)
		rsp.ErrRsp(c, -5, errKeysFileFailed.Error())
	default:
		log.Infof("authorized key added: %s %s", key.Type, key.Fingerprint)
		rsp.OkRspWithData(c, key)
	}
}

func (s *Service) DeleteSSHKey(c *gin.Context) {
	var rsp proto.Response
	var req proto.DeleteSSHKeyReq
	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	err := deleteAuthorizedKey(req.Fingerprint)
	switch {
	case errors.Is(err, errKeyNotFound):
		rsp.ErrRsp(c, -2, err.Error())
	case errors.Is(err, errLastKey):
		rsp.ErrRsp(c, -3, err.Error())
	case err != nil:
		log.Errorf("failed to remove an authorized key: %s", err)
		rsp.ErrRsp(c, -5, errKeysFileFailed.Error())
	default:
		log.Infof("authorized key removed: %s", req.Fingerprint)
		rsp.OkRsp(c)
	}
}

func (s *Service) SetSSHKeysOnly(c *gin.Context) {
	var rsp proto.Response
	var req proto.SetSSHKeysOnlyReq
	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	err := setKeysOnly(req.Enabled)
	switch {
	case errors.Is(err, errNoKeys):
		rsp.ErrRsp(c, -2, err.Error())
	case errors.Is(err, errNotHonoured):
		rsp.ErrRsp(c, -3, err.Error())
	case errors.Is(err, errReloadFailed):
		rsp.ErrRsp(c, -4, err.Error())
	case err != nil:
		log.Errorf("failed to change keys-only login: %s", err)
		rsp.ErrRsp(c, -5, fmt.Sprintf("failed to change keys-only login: %s", err))
	default:
		log.Infof("ssh keys-only login set to %t", req.Enabled)
		rsp.OkRsp(c)
	}
}
