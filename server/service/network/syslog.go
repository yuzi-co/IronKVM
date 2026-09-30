package network

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"NanoKVM-Server/proto"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Remote syslog forwarding is an owner setting that the image's init script
// applies, not the server: /etc/init.d/S01syslogd reads one line, HOST or
// HOST:PORT, and runs `syslogd -L -R HOST:PORT` when it is valid. See
// probe/build-rootfs.sh in ironkvm-dist. The script reads the kept copy under
// /data/identity first, because it runs before S02identity binds that
// directory over /etc/kvm; /etc/kvm/syslog-remote is the fallback for a board
// with no kept identity. The server writes both, so either name holds the
// value whichever one the script finds.
type syslogEnv struct {
	keptFile   string
	configFile string
	initScript string
	procDir    string
	socket     string
	restart    func(ctx context.Context, script string) error
	now        func() time.Time
}

var syslogPaths = syslogEnv{
	keptFile:   "/data/identity/syslog-remote",
	configFile: "/etc/kvm/syslog-remote",
	initScript: "/etc/init.d/S01syslogd",
	procDir:    "/proc",
	socket:     "/dev/log",
	restart:    runInitScript,
	now:        time.Now,
}

// The script's stop waits up to five seconds for the old daemons to exit.
const syslogRestartTimeout = 20 * time.Second

// syslogTargetMax keeps a pasted paragraph out of the file. A DNS name is at
// most 253 characters, and a port and brackets add a few more.
const syslogTargetMax = 262

type syslogState struct {
	// Target is the collector the setting names, empty when it names none.
	Target string `json:"target"`
	// Active is the collector the running syslogd forwards to, from its
	// command line, empty when it forwards nowhere or is not running.
	Active string `json:"active"`
	// Supported says whether this image's init script reads the setting.
	// Images built before it existed start syslogd without looking.
	Supported bool `json:"supported"`
}

type setSyslogReq struct {
	Target string `json:"target"`
}

type syslogTestRsp struct {
	Message string `json:"message"`
}

func (s *Service) GetSyslog(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, syslogPaths.state())
}

func (s *Service) SetSyslog(c *gin.Context) {
	var req setSyslogReq
	var rsp proto.Response

	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	target := strings.TrimSpace(req.Target)
	if target != "" {
		if err := validateSyslogTarget(target); err != nil {
			rsp.ErrRsp(c, -1, err.Error())
			return
		}
	}

	env := syslogPaths
	if err := env.write(target); err != nil {
		log.Errorf("failed to save the remote syslog target: %s", err)
		rsp.ErrRsp(c, -3, "failed to save the setting")
		return
	}
	_ = exec.Command("sync").Run()

	// The setting is on /data and outlives the image, so an image that does
	// not read it yet keeps it for the one that will. Its script is not
	// restarted: that would change nothing.
	if !env.supported() {
		rsp.OkRspWithData(c, env.state())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), syslogRestartTimeout)
	defer cancel()
	if err := env.restart(ctx, env.initScript); err != nil {
		log.Errorf("failed to restart syslogd: %s", err)
		rsp.ErrRsp(c, -4, "saved, but the logger did not restart: "+err.Error())
		return
	}

	log.Infof("remote syslog target set to %q", target)
	rsp.OkRspWithData(c, env.state())
}

func (s *Service) TestSyslog(c *gin.Context) {
	var rsp proto.Response

	env := syslogPaths
	msg := "IronKVM test message " + env.now().Format(time.RFC3339)
	if err := sendLocalSyslog(env.socket, "ironkvm", msg, env.now()); err != nil {
		log.Errorf("failed to send a syslog test message: %s", err)
		rsp.ErrRsp(c, -1, "the system logger is not accepting messages")
		return
	}

	rsp.OkRspWithData(c, &syslogTestRsp{Message: msg})
}

func (e syslogEnv) state() *syslogState {
	return &syslogState{
		Target:    e.target(),
		Active:    activeSyslogTarget(e.procDir),
		Supported: e.supported(),
	}
}

// target reads the setting the way the script does: the kept copy when it is
// readable, the first line that is not blank and not a comment, carriage
// returns stripped.
func (e syslogEnv) target() string {
	name := e.configFile
	if f, err := os.Open(e.keptFile); err == nil {
		_ = f.Close()
		name = e.keptFile
	}

	data, err := os.ReadFile(name)
	if err != nil {
		return ""
	}
	return parseSyslogFile(data)
}

func parseSyslogFile(data []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(strings.ReplaceAll(scanner.Text(), "\r", ""))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line
	}
	return ""
}

// supported reports whether the init script knows the setting. The scripts
// that do name the file; the older ones never mention it.
func (e syslogEnv) supported() bool {
	data, err := os.ReadFile(e.initScript)
	return err == nil && bytes.Contains(data, []byte("syslog-remote"))
}

// write stores the target in both files, or removes both when it is empty.
// The kept copy is written only where /data/identity exists: a board without
// it has nothing binding it over /etc/kvm, and the script falls back to the
// /etc/kvm name.
func (e syslogEnv) write(target string) error {
	for _, name := range []string{e.keptFile, e.configFile} {
		if target == "" {
			if err := os.Remove(name); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			continue
		}

		if name == e.keptFile {
			if _, err := os.Stat(filepath.Dir(name)); err != nil {
				continue
			}
		} else if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
			return err
		}

		if err := writeFileAtomic(name, []byte(target+"\n"), 0o600); err != nil {
			return err
		}
	}
	return nil
}

func writeFileAtomic(name string, data []byte, perm os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(name), "."+filepath.Base(name)+".*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	// exFAT has no permissions to set, and a failure there is not one.
	_ = os.Chmod(tmpName, perm)
	return os.Rename(tmpName, name)
}

func runInitScript(ctx context.Context, script string) error {
	out, err := exec.CommandContext(ctx, script, "restart").CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return err
		}
		return fmt.Errorf("%w: %s", err, msg)
	}
	return nil
}

// activeSyslogTarget finds a running syslogd and returns its -R value.
func activeSyslogTarget(procDir string) string {
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if _, err := strconv.Atoi(entry.Name()); err != nil {
			continue
		}
		data, err := os.ReadFile(filepath.Join(procDir, entry.Name(), "cmdline"))
		if err != nil || len(data) == 0 {
			continue
		}
		args := strings.Split(strings.TrimRight(string(data), "\x00"), "\x00")
		if filepath.Base(args[0]) != "syslogd" {
			continue
		}
		return remoteFromArgs(args[1:])
	}
	return ""
}

func remoteFromArgs(args []string) string {
	for i, arg := range args {
		switch {
		case arg == "-R":
			if i+1 < len(args) {
				return args[i+1]
			}
		case strings.HasPrefix(arg, "-R"):
			return arg[2:]
		}
	}
	return ""
}

// validateSyslogTarget accepts what the script's valid_target accepts: a host
// name, an IPv4 address or a bracketed IPv6 address, with an optional port
// from 1 to 65535. It is stricter in one place: a bracketed address followed
// by a bare colon is refused rather than read as no port.
func validateSyslogTarget(target string) error {
	if target == "" {
		return errors.New("enter a host or host:port")
	}
	if len(target) > syslogTargetMax {
		return errors.New("the target is too long")
	}

	var host, port string
	hasPort := false
	switch {
	case strings.HasPrefix(target, "[") && strings.Contains(target, "]:"):
		i := strings.Index(target, "]:")
		host, port, hasPort = target[:i+1], target[i+2:], true
	case strings.HasPrefix(target, "[") && strings.HasSuffix(target, "]"):
		host = target
	case strings.Count(target, ":") > 1:
		return errors.New("put an IPv6 address in brackets, like [fd00::1]:514")
	case strings.Contains(target, ":"):
		i := strings.LastIndex(target, ":")
		host, port, hasPort = target[:i], target[i+1:], true
	default:
		host = target
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		inner := host[1 : len(host)-1]
		if inner == "" || strings.Trim(inner, "0123456789abcdefABCDEF:.") != "" {
			return errors.New("the address in brackets is not an IPv6 address")
		}
	} else if host == "" || host[0] == '-' || host[0] == '.' ||
		strings.Trim(host, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-") != "" {
		return errors.New("the host may hold only letters, digits, dots and dashes, and may not start with a dot or dash")
	}

	if !hasPort {
		return nil
	}
	if port == "" || len(port) > 5 || strings.Trim(port, "0123456789") != "" {
		return errors.New("the port must be a number from 1 to 65535")
	}
	if n, _ := strconv.Atoi(port); n < 1 || n > 65535 {
		return errors.New("the port must be a number from 1 to 65535")
	}
	return nil
}

// sendLocalSyslog writes one message to the local syslogd socket, in the
// format busybox syslogd reads from /dev/log. user.notice, like logger(1).
func sendLocalSyslog(socket, tag, msg string, now time.Time) error {
	conn, err := net.DialTimeout("unixgram", socket, 2*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	line := fmt.Sprintf("<13>%s %s[%d]: %s", now.Format(time.Stamp), tag, os.Getpid(), msg)
	_, err = conn.Write([]byte(line))
	return err
}
