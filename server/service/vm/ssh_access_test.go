package vm

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"

	"NanoKVM-Server/utils/unixcrypt"
)

// sshFixture is a scratch tree standing in for /root/.ssh, /etc/kvm, /etc/ssh
// and /etc/shadow, and records of what would have been done to sshd.
type sshFixture struct {
	root     string
	reloads  int
	saves    int
	sshdSays map[string]string
	sshdErr  error
	reloadFn func() error
	// listening is what /proc/net/tcp would say listens.
	listening map[int]bool
}

func useSSHFixture(t *testing.T) *sshFixture {
	t.Helper()
	f := &sshFixture{
		root:     t.TempDir(),
		sshdSays: map[string]string{"port": "22", "passwordauthentication": "yes"},
	}

	originals := []string{authorizedKeysPath, sshKeysOnlyFlag, sshKeysOnlyDropIn, sshHostKeyGlob, shadowPath, sshdPidFile,
		sshPortFile, sshPortDropIn}
	originalConfig, originalReload, originalSave := sshdEffectiveConfig, reloadSSHD, saveSSHIdentity
	originalListeners, originalReserved := tcpListeners, reservedPorts
	t.Cleanup(func() {
		authorizedKeysPath, sshKeysOnlyFlag, sshKeysOnlyDropIn = originals[0], originals[1], originals[2]
		sshHostKeyGlob, shadowPath, sshdPidFile = originals[3], originals[4], originals[5]
		sshPortFile, sshPortDropIn = originals[6], originals[7]
		sshdEffectiveConfig, reloadSSHD, saveSSHIdentity = originalConfig, originalReload, originalSave
		tcpListeners, reservedPorts = originalListeners, originalReserved
	})

	authorizedKeysPath = filepath.Join(f.root, "root", ".ssh", "authorized_keys")
	sshKeysOnlyFlag = filepath.Join(f.root, "etc", "kvm", "ssh_keys_only")
	sshKeysOnlyDropIn = filepath.Join(f.root, "etc", "ssh", "sshd_config.d", "ironkvm-keys-only.conf")
	sshHostKeyGlob = filepath.Join(f.root, "etc", "ssh", "ssh_host_*_key.pub")
	shadowPath = filepath.Join(f.root, "etc", "shadow")
	sshdPidFile = filepath.Join(f.root, "run", "sshd.pid")
	sshPortFile = filepath.Join(f.root, "etc", "kvm", "ssh_port")
	sshPortDropIn = filepath.Join(f.root, "etc", "ssh", "sshd_config.d", "ironkvm-port.conf")
	f.listening = map[int]bool{80: true, 443: true}
	tcpListeners = func() (map[int]bool, error) { return f.listening, nil }
	reservedPorts = func() []int { return []int{80, 443, 5900, 8069} }
	if err := os.MkdirAll(filepath.Join(f.root, "etc", "kvm"), 0o755); err != nil {
		t.Fatal(err)
	}

	// sshd honours the drop-in when it exists, as the image's sshd does.
	sshdEffectiveConfig = func() (map[string]string, error) {
		if f.sshdErr != nil {
			return nil, f.sshdErr
		}
		out := map[string]string{}
		for k, v := range f.sshdSays {
			out[k] = v
		}
		if _, err := os.Stat(sshKeysOnlyDropIn); err == nil && f.sshdSays["honoursDropIn"] != "no" {
			out["passwordauthentication"] = "no"
		}
		// The drop-in's Port line comes first, since the Include heads
		// sshd_config; a main file with its own Port line adds a second.
		if data, err := os.ReadFile(sshPortDropIn); err == nil && f.sshdSays["honoursDropIn"] != "no" {
			var port int
			for _, line := range strings.Split(string(data), "\n") {
				if _, err := fmt.Sscanf(line, "Port %d", &port); err == nil {
					break
				}
			}
			if f.sshdSays["mainPort"] != "" {
				out["port"] = fmt.Sprintf("%d,%s", port, f.sshdSays["mainPort"])
			} else {
				out["port"] = strconv.Itoa(port)
			}
		}
		return out, nil
	}
	reloadSSHD = func() error {
		f.reloads++
		if f.reloadFn != nil {
			return f.reloadFn()
		}
		return nil
	}
	saveSSHIdentity = func() error { f.saves++; return nil }
	return f
}

// newKey returns a fresh ed25519 public key as an authorized_keys line. Keys
// are made per test so no fixture carries one.
func newKey(t *testing.T, comment string) (string, ssh.PublicKey) {
	t.Helper()
	public, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pub, err := ssh.NewPublicKey(public)
	if err != nil {
		t.Fatal(err)
	}
	line := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pub)))
	if comment != "" {
		line += " " + comment
	}
	return line, pub
}

// keygenFingerprint is what `ssh-keygen -lf` prints: SHA256 of the key's wire
// encoding, base64 without padding. Computed here without the ssh package's
// helper, so the test checks the format rather than restating the code.
func keygenFingerprint(pub ssh.PublicKey) string {
	sum := sha256.Sum256(pub.Marshal())
	return "SHA256:" + base64.RawStdEncoding.EncodeToString(sum[:])
}

func (f *sshFixture) writeKeys(t *testing.T, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(authorizedKeysPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(authorizedKeysPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func TestReadAuthorizedKeysListsKeysAndSkipsTheRest(t *testing.T) {
	f := useSSHFixture(t)
	first, firstPub := newKey(t, "laptop")
	second, secondPub := newKey(t, "")
	f.writeKeys(t, "# managed by hand\n\n"+first+"\nnot a key at all\n"+second+"\n")

	lines, err := readAuthorizedKeys()
	if err != nil {
		t.Fatal(err)
	}
	keys := keysOf(lines)
	if len(keys) != 2 {
		t.Fatalf("got %d keys, want 2: %+v", len(keys), keys)
	}
	if keys[0].Type != "ssh-ed25519" || keys[0].Comment != "laptop" || keys[0].Fingerprint != keygenFingerprint(firstPub) {
		t.Errorf("first key = %+v", keys[0])
	}
	if keys[1].Comment != "" || keys[1].Fingerprint != keygenFingerprint(secondPub) {
		t.Errorf("second key = %+v", keys[1])
	}
}

func TestReadAuthorizedKeysWithNoFileIsEmpty(t *testing.T) {
	useSSHFixture(t)
	lines, err := readAuthorizedKeys()
	if err != nil || len(keysOf(lines)) != 0 {
		t.Fatalf("got %v, %v; want no keys and no error", lines, err)
	}
}

func TestAddAuthorizedKeyWritesAPrivateFileAndSavesIdentity(t *testing.T) {
	f := useSSHFixture(t)
	line, pub := newKey(t, "me@desk")

	key, err := addAuthorizedKey("  " + line + "  \n")
	if err != nil {
		t.Fatal(err)
	}
	if key.Fingerprint != keygenFingerprint(pub) {
		t.Errorf("fingerprint %s, want %s", key.Fingerprint, keygenFingerprint(pub))
	}
	if got := readFile(t, authorizedKeysPath); got != line+"\n" {
		t.Errorf("file = %q, want %q", got, line+"\n")
	}

	info, _ := os.Stat(authorizedKeysPath)
	if info.Mode().Perm() != 0o600 {
		t.Errorf("file mode %o, want 600", info.Mode().Perm())
	}
	dir, _ := os.Stat(filepath.Dir(authorizedKeysPath))
	if dir.Mode().Perm() != 0o700 {
		t.Errorf("dir mode %o, want 700", dir.Mode().Perm())
	}
	if f.saves != 1 {
		t.Errorf("identity saved %d times, want 1", f.saves)
	}

	// No temporary file is left beside it.
	entries, _ := os.ReadDir(filepath.Dir(authorizedKeysPath))
	if len(entries) != 1 {
		t.Errorf("directory holds %d entries, want only authorized_keys", len(entries))
	}
}

func TestAddAuthorizedKeyTightensAnOpenDirectory(t *testing.T) {
	useSSHFixture(t)
	if err := os.MkdirAll(filepath.Dir(authorizedKeysPath), 0o777); err != nil {
		t.Fatal(err)
	}
	_ = os.Chmod(filepath.Dir(authorizedKeysPath), 0o777)
	line, _ := newKey(t, "")
	if _, err := addAuthorizedKey(line); err != nil {
		t.Fatal(err)
	}
	dir, _ := os.Stat(filepath.Dir(authorizedKeysPath))
	if dir.Mode().Perm() != 0o700 {
		t.Errorf("dir mode %o, want 700; sshd's StrictModes refuses keys under it otherwise", dir.Mode().Perm())
	}
}

func TestAddAuthorizedKeyKeepsWhatIsAlreadyThere(t *testing.T) {
	f := useSSHFixture(t)
	existing, _ := newKey(t, "old")
	f.writeKeys(t, "# a note\n"+existing+"\n")
	added, _ := newKey(t, "new")

	if _, err := addAuthorizedKey(added); err != nil {
		t.Fatal(err)
	}
	want := "# a note\n" + existing + "\n" + added + "\n"
	if got := readFile(t, authorizedKeysPath); got != want {
		t.Errorf("file = %q, want %q", got, want)
	}
}

func TestAddAuthorizedKeyRefusesWhatItShould(t *testing.T) {
	f := useSSHFixture(t)
	existing, _ := newKey(t, "")
	f.writeKeys(t, existing+"\n")
	other, _ := newKey(t, "")
	third, _ := newKey(t, "")

	for _, test := range []struct {
		name  string
		input string
		want  error
	}{
		{"empty", "   ", errKeyInvalid},
		{"garbage", "ssh-ed25519 notbase64!!", errKeyInvalid},
		{"a private key header", "-----BEGIN OPENSSH PRIVATE KEY-----", errKeyInvalid},
		{"options", `command="/bin/true" ` + other, errKeyOptions},
		{"two keys", other + "\n" + third, errKeyMultiple},
		{"a duplicate with another comment", existing + " renamed", errKeyDuplicate},
	} {
		if _, err := addAuthorizedKey(test.input); !errors.Is(err, test.want) {
			t.Errorf("%s: err = %v, want %v", test.name, err, test.want)
		}
	}
	if got := readFile(t, authorizedKeysPath); got != existing+"\n" {
		t.Errorf("a refused add changed the file: %q", got)
	}
}

func TestDeleteAuthorizedKeyRemovesOnlyThatKey(t *testing.T) {
	f := useSSHFixture(t)
	keep, _ := newKey(t, "keep")
	drop, dropPub := newKey(t, "drop")
	f.writeKeys(t, "# note\n"+keep+"\n"+drop+"\nnot a key\n")

	if err := deleteAuthorizedKey(keygenFingerprint(dropPub)); err != nil {
		t.Fatal(err)
	}
	want := "# note\n" + keep + "\nnot a key\n"
	if got := readFile(t, authorizedKeysPath); got != want {
		t.Errorf("file = %q, want %q", got, want)
	}
	if f.saves != 1 {
		t.Errorf("identity saved %d times, want 1", f.saves)
	}
	if err := deleteAuthorizedKey(keygenFingerprint(dropPub)); !errors.Is(err, errKeyNotFound) {
		t.Errorf("second delete err = %v, want errKeyNotFound", err)
	}
}

func TestDeleteRefusesTheLastKeyWhileKeysOnlyIsOn(t *testing.T) {
	f := useSSHFixture(t)
	only, pub := newKey(t, "")
	f.writeKeys(t, only+"\n")
	if err := setKeysOnly(true); err != nil {
		t.Fatal(err)
	}

	if err := deleteAuthorizedKey(keygenFingerprint(pub)); !errors.Is(err, errLastKey) {
		t.Fatalf("err = %v, want errLastKey", err)
	}
	if got := readFile(t, authorizedKeysPath); got != only+"\n" {
		t.Errorf("the refused delete changed the file: %q", got)
	}

	// With keys-only off the same delete goes through.
	if err := setKeysOnly(false); err != nil {
		t.Fatal(err)
	}
	if err := deleteAuthorizedKey(keygenFingerprint(pub)); err != nil {
		t.Errorf("delete with keys-only off: %v", err)
	}
}

func TestKeysOnlyRefusesABoardWithNoKeys(t *testing.T) {
	f := useSSHFixture(t)
	f.writeKeys(t, "# only a comment\n")

	if err := setKeysOnly(true); !errors.Is(err, errNoKeys) {
		t.Fatalf("err = %v, want errNoKeys", err)
	}
	if exists(sshKeysOnlyFlag) || exists(sshKeysOnlyDropIn) || f.reloads != 0 {
		t.Errorf("a refused enable left flag=%v dropin=%v reloads=%d", exists(sshKeysOnlyFlag), exists(sshKeysOnlyDropIn), f.reloads)
	}
}

func TestKeysOnlyWritesTheFlagAndTheDropInAndReloads(t *testing.T) {
	f := useSSHFixture(t)
	line, _ := newKey(t, "")
	f.writeKeys(t, line+"\n")

	if err := setKeysOnly(true); err != nil {
		t.Fatal(err)
	}
	if !exists(sshKeysOnlyFlag) {
		t.Error("no flag on /etc/kvm, so the next boot would forget the setting")
	}
	dropIn := readFile(t, sshKeysOnlyDropIn)
	for _, want := range []string{"PasswordAuthentication no\n", "KbdInteractiveAuthentication no\n"} {
		if !strings.Contains(dropIn, want) {
			t.Errorf("drop-in lacks %q:\n%s", want, dropIn)
		}
	}
	if f.reloads != 1 {
		t.Errorf("sshd reloaded %d times, want 1", f.reloads)
	}

	if err := setKeysOnly(false); err != nil {
		t.Fatal(err)
	}
	if exists(sshKeysOnlyFlag) || exists(sshKeysOnlyDropIn) {
		t.Error("turning keys-only off left the flag or the drop-in behind")
	}
	if f.reloads != 2 {
		t.Errorf("sshd reloaded %d times, want 2", f.reloads)
	}
}

// An sshd that does not read sshd_config.d would take the drop-in and ignore
// it. The switch must not then claim password logins are off.
func TestKeysOnlyRefusesWhenSSHDIgnoresTheDropIn(t *testing.T) {
	f := useSSHFixture(t)
	line, _ := newKey(t, "")
	f.writeKeys(t, line+"\n")
	f.sshdSays["honoursDropIn"] = "no"

	if err := setKeysOnly(true); !errors.Is(err, errNotHonoured) {
		t.Fatalf("err = %v, want errNotHonoured", err)
	}
	if exists(sshKeysOnlyFlag) || exists(sshKeysOnlyDropIn) || f.reloads != 0 {
		t.Errorf("left flag=%v dropin=%v reloads=%d", exists(sshKeysOnlyFlag), exists(sshKeysOnlyDropIn), f.reloads)
	}
}

// sshd -T failing means the configuration may not parse. A SIGHUP then would
// make sshd exit, so nothing is reloaded and the drop-in is taken back.
func TestKeysOnlyDoesNotReloadAnSSHDThatCannotReadItsConfig(t *testing.T) {
	f := useSSHFixture(t)
	line, _ := newKey(t, "")
	f.writeKeys(t, line+"\n")
	f.sshdErr = errors.New("exit status 255")

	if err := setKeysOnly(true); !errors.Is(err, errNotHonoured) {
		t.Fatalf("err = %v, want errNotHonoured", err)
	}
	if exists(sshKeysOnlyDropIn) || f.reloads != 0 {
		t.Errorf("left dropin=%v reloads=%d", exists(sshKeysOnlyDropIn), f.reloads)
	}
}

func TestKeysOnlyKeepsTheSettingWhenTheReloadFails(t *testing.T) {
	f := useSSHFixture(t)
	line, _ := newKey(t, "")
	f.writeKeys(t, line+"\n")
	f.reloadFn = func() error { return errors.New("no such process") }

	if err := setKeysOnly(true); !errors.Is(err, errReloadFailed) {
		t.Fatalf("err = %v, want errReloadFailed", err)
	}
	if !exists(sshKeysOnlyFlag) || !exists(sshKeysOnlyDropIn) {
		t.Error("the setting should stand, to apply at sshd's next start")
	}
}

func TestHangUpSSHDWithNoListenerDoesNothing(t *testing.T) {
	useSSHFixture(t)
	if err := hangUpSSHD(); err != nil {
		t.Errorf("no pid file: %v", err)
	}
	_ = os.MkdirAll(filepath.Dir(sshdPidFile), 0o755)
	_ = os.WriteFile(sshdPidFile, []byte("not a pid\n"), 0o644)
	if err := hangUpSSHD(); err != nil {
		t.Errorf("garbage pid file: %v", err)
	}
}

func TestParseSSHDConfigTakesTheFirstValue(t *testing.T) {
	config := parseSSHDConfig([]byte("port 2222\nport 22\nPasswordAuthentication no\nlistenaddress 0.0.0.0:2222\n"))
	if config["port"] != "2222,22" || sshPort(config) != 2222 {
		t.Errorf("port = %q", config["port"])
	}
	if config["passwordauthentication"] != "no" {
		t.Errorf("passwordauthentication = %q", config["passwordauthentication"])
	}
	if sshPort(map[string]string{}) != 22 {
		t.Error("a config without a port should read as 22")
	}
}

func TestReadHostKeysFingerprintsEachPublicKey(t *testing.T) {
	useSSHFixture(t)
	dir := filepath.Dir(sshHostKeyGlob)
	_ = os.MkdirAll(dir, 0o755)
	line, pub := newKey(t, "root@board")
	_ = os.WriteFile(filepath.Join(dir, "ssh_host_ed25519_key.pub"), []byte(line+"\n"), 0o644)
	// The private half and an unreadable one are both ignored.
	_ = os.WriteFile(filepath.Join(dir, "ssh_host_ed25519_key"), []byte("private\n"), 0o600)
	_ = os.WriteFile(filepath.Join(dir, "ssh_host_rsa_key.pub"), []byte(""), 0o644)

	keys := readHostKeys()
	if len(keys) != 1 || keys[0].Type != "ssh-ed25519" || keys[0].Fingerprint != keygenFingerprint(pub) || keys[0].Comment != "" {
		t.Errorf("host keys = %+v", keys)
	}
}

func writeShadow(t *testing.T, rootField string) {
	t.Helper()
	content := "bin:!::0:::::\nroot:" + rootField + ":19000:0:99999:7:::\nsshd:!::0:::::\n"
	if err := os.WriteFile(shadowPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestRootPasswordState(t *testing.T) {
	useSSHFixture(t)

	if got := rootPasswordState(); got != "unknown" {
		t.Errorf("no shadow file: %s, want unknown", got)
	}

	// The factory password, hashed at test time rather than stored here.
	factory, _ := unixcrypt.Crypt("root", "$6$fixturesalt")
	factoryMD5, _ := unixcrypt.Crypt("root", "$1$fixture")
	changed, _ := unixcrypt.Crypt("a different one", "$6$fixturesalt")

	for _, test := range []struct {
		field string
		want  string
	}{
		{"", "empty"},
		{factory, "default"},
		{factoryMD5, "default"},
		{changed, "set"},
		{"!", "set"},
		{"*", "set"},
		{"!" + factory, "set"},
		{"$y$j9T$salt$hash", "unknown"},
	} {
		writeShadow(t, test.field)
		if got := rootPasswordState(); got != test.want {
			t.Errorf("root field %q: %s, want %s", test.field, got, test.want)
		}
	}
}

func sshEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	service := NewService()
	r := gin.New()
	r.GET("/api/vm/ssh", service.GetSSHState)
	r.GET("/api/vm/ssh/keys", service.GetSSHKeys)
	r.POST("/api/vm/ssh/keys", service.AddSSHKey)
	r.DELETE("/api/vm/ssh/keys", service.DeleteSSHKey)
	r.POST("/api/vm/ssh/keys-only", service.SetSSHKeysOnly)
	r.POST("/api/vm/ssh/port", service.SetSSHPort)
	return r
}

func TestSSHEndpointsAddListAndDelete(t *testing.T) {
	useSSHFixture(t)
	r := sshEngine()
	line, pub := newKey(t, "desk")

	if rsp := serve(r, http.MethodPost, "/api/vm/ssh/keys", `{"key":"`+line+`"}`); rsp["code"] != float64(0) {
		t.Fatalf("add answered %v", rsp)
	}
	if rsp := serve(r, http.MethodPost, "/api/vm/ssh/keys", `{"key":"`+line+`"}`); rsp["code"] != float64(-4) {
		t.Errorf("duplicate add answered %v, want code -4", rsp)
	}
	if rsp := serve(r, http.MethodPost, "/api/vm/ssh/keys", `{"key":"hello"}`); rsp["code"] != float64(-2) {
		t.Errorf("bad add answered %v, want code -2", rsp)
	}

	rsp := serve(r, http.MethodGet, "/api/vm/ssh/keys", "")
	keys := rsp["data"].(map[string]any)["keys"].([]any)
	if len(keys) != 1 || keys[0].(map[string]any)["fingerprint"] != keygenFingerprint(pub) {
		t.Fatalf("list answered %v", rsp)
	}

	if rsp := serve(r, http.MethodPost, "/api/vm/ssh/keys-only", `{"enabled":true}`); rsp["code"] != float64(0) {
		t.Fatalf("keys-only answered %v", rsp)
	}
	body := `{"fingerprint":"` + keygenFingerprint(pub) + `"}`
	if rsp := serve(r, http.MethodDelete, "/api/vm/ssh/keys", body); rsp["code"] != float64(-3) {
		t.Errorf("deleting the last key under keys-only answered %v, want code -3", rsp)
	}
	serve(r, http.MethodPost, "/api/vm/ssh/keys-only", `{"enabled":false}`)
	if rsp := serve(r, http.MethodDelete, "/api/vm/ssh/keys", body); rsp["code"] != float64(0) {
		t.Errorf("delete answered %v", rsp)
	}
	if rsp := serve(r, http.MethodPost, "/api/vm/ssh/keys-only", `{"enabled":true}`); rsp["code"] != float64(-2) {
		t.Errorf("keys-only with no keys answered %v, want code -2", rsp)
	}
}

func TestSSHStateReportsTheVerdictAndNeverTheHash(t *testing.T) {
	f := useSSHFixture(t)
	factory, _ := unixcrypt.Crypt("root", "$6$fixturesalt")
	writeShadow(t, factory)
	f.sshdSays["port"] = "2222"
	line, _ := newKey(t, "")
	f.writeKeys(t, line+"\n")

	r := sshEngine()
	request := serve(r, http.MethodGet, "/api/vm/ssh", "")
	data := request["data"].(map[string]any)
	if data["rootPassword"] != "default" || data["port"] != float64(2222) || data["keyCount"] != float64(1) ||
		data["passwordAuth"] != "yes" || data["keysOnly"] != false {
		t.Errorf("state = %v", data)
	}

	raw := serveRaw(r, http.MethodGet, "/api/vm/ssh")
	if strings.Contains(raw, "fixturesalt") || strings.Contains(raw, factory[len(factory)-20:]) {
		t.Error("the response carries the password hash")
	}
}

func serveRaw(r *gin.Engine, method, path string) string {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(method, path, nil))
	return w.Body.String()
}
