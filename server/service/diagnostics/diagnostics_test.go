package diagnostics

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const sentinel = "SENTINEL"

type fakeBoard struct {
	t    *testing.T
	root string
}

func newFakeBoard(t *testing.T) *fakeBoard {
	return &fakeBoard{t: t, root: t.TempDir()}
}

func (b *fakeBoard) write(path, content string) {
	b.t.Helper()
	full := filepath.Join(b.root, path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		b.t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		b.t.Fatal(err)
	}
}

// secrets are the values seeded on the fake board, each holding the
// sentinel.
var secrets = map[string]string{
	"/etc/kvm/.jwt_secret":              sentinel + "-jwtfile-9f8e7d",
	"/etc/kvm/totp.key":                 sentinel + "-totp-JBSWY3DP",
	"/etc/kvm/.picoclaw_internal_token": sentinel + "-picoclaw-tok",
	"/etc/kvm/first-boot-password":      sentinel + "-firstboot",
	"/boot/ironkvm-first-password.txt":  sentinel + "-firstcard",
	"/etc/kvm/wifi.ssid":                sentinel + "-ssid",
	"/etc/kvm/wifi.pass":                sentinel + "-wifipass",
	"/etc/kvm/vnc.passwd":               sentinel + "-vncpass",
	"/etc/kvm/nut.passwd":               sentinel + "-nutpass",
	"/etc/kvm/ipmi.key":                 sentinel + "-ipmikey",
}

// jsonSecrets are values inside the JSON secret files and server.yaml, as
// they are stored there. The proxy's password also appears alone, as a log
// line that names only the password would hold it.
var jsonSecrets = []string{
	"$2a$10$" + sentinel + "-pwdhash", sentinel + "-ipmipw", sentinel + "-totpsecret", sentinel + "-recovery1",
	sentinel + "-apikeyhash", sentinel + "-sharehash", "nag_mcp_" + sentinel + "-mcpkey",
	"https://hooks.example/" + sentinel + "-webhook", sentinel + "-ntfytoken", sentinel + "-topic",
	sentinel + "-botTok", sentinel + "-chat",
	"https://updates.example/" + sentinel + "-updateurl",
	sentinel + "-yamljwt", sentinel + "-turnuser", sentinel + "-turncred",
	"http://user:" + sentinel + "-proxy@proxy.lan:3128", sentinel + "-proxy",
}

func seedBoard(b *fakeBoard) {
	for path, value := range secrets {
		b.write(path, value+"\n")
	}
	b.write("/etc/kvm/pwd", `{"version":2,"users":[{"username":"admin","password":"$2a$10$`+sentinel+`-pwdhash","ipmiPassword":"`+sentinel+`-ipmipw","totpSecret":"`+sentinel+`-totpsecret","recoveryCodes":["`+sentinel+`-recovery1"]}]}`)
	b.write("/etc/kvm/api_keys.json", `[{"id":"k1","name":"laptop","hash":"`+sentinel+`-apikeyhash","createdAt":1700000000}]`)
	b.write("/etc/kvm/share_links.json", `{"version":1,"links":[{"id":"s1","tokenHash":"`+sentinel+`-sharehash"}]}`)
	b.write("/etc/kvm/mcp.json", `{"apiKey":"nag_mcp_`+sentinel+`-mcpkey"}`)
	b.write("/etc/kvm/notify.json", `{"webhook":{"url":"https://hooks.example/`+sentinel+`-webhook"},"ntfy":{"token":"`+sentinel+`-ntfytoken","topic":"`+sentinel+`-topic"},"telegram":{"botToken":"`+sentinel+`-botTok","chatId":"`+sentinel+`-chat"}}`)
	b.write("/etc/kvm/application-update.json", `{"enabled":true,"url":"https://updates.example/`+sentinel+`-updateurl"}`)
	b.write("/etc/kvm/server.key", "-----BEGIN PRIVATE KEY-----\n"+sentinel+"PEMBODY\n-----END PRIVATE KEY-----\n")
	b.write("/etc/kvm/server.yaml", `proto: https
port:
  http: 80
  https: 443
logger:
  level: info
  file: stdout
authentication: enable
jwt:
  secretKey: `+sentinel+`-yamljwt
  refreshTokenDuration: 2678400
turn:
  turnAddr: turn:turn.example:3478
  turnUser: `+sentinel+`-turnuser
  turnCred: `+sentinel+`-turncred
proxy: http://user:`+sentinel+`-proxy@proxy.lan:3128
security:
  allowedNetworks:
    - 10.0.0.0/8
`)

	// The server log and the kernel log echo every secret, as a careless
	// log line would, and some secrets the board does not hold.
	var log strings.Builder
	for _, v := range secrets {
		log.WriteString("level=info msg=\"loaded " + v + "\"\n")
	}
	for _, v := range jsonSecrets {
		log.WriteString("level=debug msg=\"value " + v + "\"\n")
	}
	log.WriteString("Authorization: Bearer " + sentinel + "abcdefgh12345\n")
	log.WriteString("token=" + sentinel + "-unknowntoken other=1\n")
	log.WriteString(`{"password": "` + sentinel + `-jsonpass"}` + "\n")
	log.WriteString("fetch https://bob:" + sentinel + "-urlpass@example.com/x\n")
	log.WriteString("session eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ" + sentinel + "In0.c2lnbmF0dXJlX" + sentinel + "\n")
	log.WriteString("tailscale up --authkey=tskey-auth-" + sentinel + "-abc\n")
	log.WriteString("key nkvm_" + sentinel + "abcdef0123\n")
	log.WriteString("bot 123456789:AAH" + sentinel + "abcdefghijklmnopqrstuvwxyz0\n")
	log.WriteString("-----BEGIN RSA PRIVATE KEY-----\n" + sentinel + "rsa\n-----END RSA PRIVATE KEY-----\n")
	b.write(ServerLogFile, log.String())

	b.write("/proc/42/stat", "42 (NanoKVM-Server) S 1 42 42 0 -1")
	b.write("/proc/42/status", "Name:\tNanoKVM-Server\nVmRSS:\t   23456 kB\n")
	b.write("/proc/42/cmdline", "tailscale\x00up\x00--authkey\x00"+sentinel+"-cmdline")
	b.write("/proc/meminfo", "MemTotal: 160000 kB\n")
	b.write("/proc/mounts", "/dev/mmcblk0p2 / ext4 rw 0 0\n")
	b.write("/proc/version", "Linux version 5.10.4\n")
	b.write("/boot/slot", "b\n")
	b.write("/kvmapp/version", "2.4.3\n")
	b.write("/sys/kernel/config/usb_gadget/g0/UDC", "4340000.usb\n")
	b.write("/sys/kernel/config/usb_gadget/g0/functions/hid.GS0/report_length", "8")
	b.write("/sys/class/udc/4340000.usb/state", "configured\n")
}

func fakeRun(b *fakeBoard) func(context.Context, string, ...string) ([]byte, error) {
	return func(ctx context.Context, name string, args ...string) ([]byte, error) {
		switch name {
		case "dmesg":
			out := "[    0.000000] Linux version 5.10.4\n[    1.0] nanokvm-slot: booting b\n"
			for _, v := range secrets {
				out += "[    2.0] wlan0: psk=" + v + "\n"
			}
			return []byte(out), nil
		case "ip":
			return []byte("2: eth0: <UP> mtu 1500\n    inet 10.0.30.222/24\n"), nil
		}
		return nil, errors.New("not found")
	}
}

func newTestCollector(b *fakeBoard) *Collector {
	c := NewCollector()
	c.Root = b.root
	c.Run = fakeRun(b)
	c.Now = func() time.Time { return time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC) }
	c.Versions = func() map[string]string { return map[string]string{"application": "2.4.3", "build": "dev.1"} }
	c.Settings = func() (any, error) { return map[string]any{"logging": map[string]string{"level": "info"}}, nil }
	c.Video = func() any { return map[string]any{"videoDisabled": false} }
	return c
}

func untar(t *testing.T, data []byte) map[string]string {
	t.Helper()
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(gz)
	files := map[string]string{}
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		files[strings.TrimPrefix(h.Name, "ironkvm-diagnostics/")] = string(body)
	}
	return files
}

func bundle(t *testing.T, c *Collector) map[string]string {
	t.Helper()
	var buf bytes.Buffer
	if err := c.Write(context.Background(), &buf); err != nil {
		t.Fatal(err)
	}
	return untar(t, buf.Bytes())
}

func TestBundleCarriesNoSecret(t *testing.T) {
	b := newFakeBoard(t)
	seedBoard(b)

	files := bundle(t, newTestCollector(b))

	for name, content := range files {
		if i := strings.Index(content, sentinel); i >= 0 {
			t.Errorf("%s holds a secret near %q", name, content[max(0, i-60):min(len(content), i+40)])
		}
	}
}

func TestBundleHoldsEverySource(t *testing.T) {
	b := newFakeBoard(t)
	seedBoard(b)

	files := bundle(t, newTestCollector(b))

	for _, name := range []string{"README.txt", "versions.txt", "slot.txt", "dmesg.txt", "server-log.txt",
		"config.json", "network.txt", "usb-gadget.txt", "video.json", "system.txt", "errors.txt"} {
		if _, ok := files[name]; !ok {
			t.Errorf("%s missing", name)
		}
	}
	checks := map[string][]string{
		"versions.txt":   {"application: 2.4.3", "build: dev.1", "Linux version 5.10.4", "/boot/ver\n(absent)"},
		"slot.txt":       {"=== /boot/slot\nb", "/dev/mmcblk0p2 / ext4"},
		"dmesg.txt":      {"nanokvm-slot: booting b"},
		"config.json":    {`"level": "info"`, `"httpsPort": 443`, `"proxySet": true`, `"10.0.0.0/8"`, `"name": "server.yaml"`},
		"network.txt":    {"=== ip addr", "inet 10.0.30.222/24"},
		"usb-gadget.txt": {"UDC: 4340000.usb", "functions: hid.GS0", "state: configured"},
		"video.json":     {`"videoDisabled": false`},
		"system.txt":     {"42 S 23456 NanoKVM-Server", "MemTotal"},
	}
	for name, wants := range checks {
		for _, want := range wants {
			if !strings.Contains(files[name], want) {
				t.Errorf("%s does not hold %q:\n%s", name, want, files[name])
			}
		}
	}
	// config.json lists /etc/kvm, and never a file's content.
	if strings.Contains(files["config.json"], "BEGIN PRIVATE KEY") {
		t.Error("config.json holds a file's content")
	}
}

func TestLogsAreCut(t *testing.T) {
	b := newFakeBoard(t)
	b.write(ServerLogFile, strings.Repeat("old line that is not wanted\n", 30000)+"the last line\n")
	c := newTestCollector(b)
	c.Run = func(context.Context, string, ...string) ([]byte, error) {
		return []byte(strings.Repeat("kernel line\n", 60000) + "last kernel line\n"), nil
	}

	files := bundle(t, c)

	for _, name := range []string{"server-log.txt", "dmesg.txt"} {
		if len(files[name]) > LogLimit+128 {
			t.Errorf("%s is %d bytes", name, len(files[name]))
		}
		if !strings.Contains(files[name], "earlier output cut") {
			t.Errorf("%s does not say it was cut", name)
		}
	}
	if !strings.HasSuffix(files["server-log.txt"], "the last line\n") || !strings.HasSuffix(files["dmesg.txt"], "last kernel line\n") {
		t.Error("the end of a log was lost")
	}
}

func TestASourceThatHangsIsAbandoned(t *testing.T) {
	old := SourceTimeout
	SourceTimeout = 100 * time.Millisecond
	t.Cleanup(func() { SourceTimeout = old })

	release := make(chan struct{})
	t.Cleanup(func() { close(release) })

	b := newFakeBoard(t)
	c := newTestCollector(b)
	// A command that ignores its context, as a read stuck in the kernel
	// would.
	c.Run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		<-release
		return nil, nil
	}

	start := time.Now()
	files := bundle(t, c)
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Fatalf("the bundle took %s", elapsed)
	}
	if !strings.Contains(files["errors.txt"], "dmesg.txt: timed out") || !strings.Contains(files["errors.txt"], "network.txt: timed out") {
		t.Fatalf("errors.txt:\n%s", files["errors.txt"])
	}
}

func TestOneBundleAtATime(t *testing.T) {
	c := newTestCollector(newFakeBoard(t))
	c.busy.Store(true)
	if err := c.Write(context.Background(), io.Discard); !errors.Is(err, ErrBusy) {
		t.Fatalf("err = %v", err)
	}
}

func TestRedactorPatterns(t *testing.T) {
	r := NewRedactor([]string{"hunter2hunter2", "short"})
	for _, tc := range []struct{ in, want string }{
		{"pw hunter2hunter2 ok", "pw [REDACTED] ok"},
		{"short stays", "short stays"},
		{"password=abc123 next", "password=[REDACTED] next"},
		{`"turnCred": "x", "apiKey": "abcdef"`, `"turnCred": "x", "apiKey": "[REDACTED]"`},
		{"wpa_passphrase=letmein", "wpa_passphrase=[REDACTED]"},
		{"Authorization: Bearer abcdefghij", "Authorization: Bearer [REDACTED]"},
		{"https://user:pw@host/path", "https://[REDACTED]@host/path"},
		{"--authkey=tskey-auth-abcdef123", "--authkey=[REDACTED]"},
		{"key tskey-auth-abcdef123", "key [REDACTED]"},
		{"nag_mcp_abcdefghijk", "[REDACTED]"},
		{"eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ4In0.c2lnbmF0dXJl", "[REDACTED]"},
		{"-----BEGIN EC PRIVATE KEY-----\nabc\n-----END EC PRIVATE KEY-----", "[REDACTED]"},
		{"[    1.0] usb 1-1: new high-speed USB device", "[    1.0] usb 1-1: new high-speed USB device"},
		{"inet 10.0.30.222/24 brd 10.0.30.255", "inet 10.0.30.222/24 brd 10.0.30.255"},
	} {
		if got := r.Redact(tc.in); got != tc.want {
			t.Errorf("Redact(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestParseStat(t *testing.T) {
	name, state := parseStat("7 (kworker/0:1 (x)) I 2 0 0")
	if name != "kworker/0:1 (x)" || state != "I" {
		t.Fatalf("%q %q", name, state)
	}
}
