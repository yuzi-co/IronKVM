package network

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The cases mirror valid_target in S01syslogd (probe/build-rootfs.sh).
func TestValidateSyslogTarget(t *testing.T) {
	valid := []string{
		"10.0.0.40",
		"10.0.0.40:5515",
		"logs",
		"logs.example.com",
		"logs.example.com:514",
		"log-host:1",
		"host:65535",
		"[::1]",
		"[::1]:514",
		"[fd00::1]:5515",
		"[FD00::a:1]",
		"[::ffff:10.0.0.1]:514",
		"host.",
	}
	for _, target := range valid {
		if err := validateSyslogTarget(target); err != nil {
			t.Errorf("%q refused: %s", target, err)
		}
	}

	invalid := []string{
		"",
		"-R",
		"-host",
		".host",
		"host:",
		"host:0",
		"host:65536",
		"host:99999",
		"host:123456",
		"host:5x",
		"host:-1",
		":514",
		"::1",
		"fd00::1:514",
		"[]",
		"[]:514",
		"[::1]:",
		"[::1]:0",
		"[::1]:70000",
		"[::g]",
		"[::1]x",
		"[::1]:514:1",
		"host name",
		"host;reboot",
		"host/24",
		"hóst",
		"[a]b]:5",
		strings.Repeat("a", 300),
	}
	for _, target := range invalid {
		if err := validateSyslogTarget(target); err == nil {
			t.Errorf("%q accepted", target)
		}
	}
}

func TestParseSyslogFile(t *testing.T) {
	cases := map[string]string{
		"":                             "",
		"\n\n":                         "",
		"# collector\n":                "",
		"10.0.0.40:5515\n":             "10.0.0.40:5515",
		"10.0.0.40:5515\r\n":           "10.0.0.40:5515",
		"# c\r\n\r\n  logs:514  \r\n":  "logs:514",
		"first\nsecond\n":              "first",
		"  # indented comment\nhost\n": "host",
	}
	for in, want := range cases {
		if got := parseSyslogFile([]byte(in)); got != want {
			t.Errorf("parseSyslogFile(%q) = %q, want %q", in, got, want)
		}
	}
}

func testSyslogEnv(t *testing.T) syslogEnv {
	t.Helper()
	dir := t.TempDir()
	for _, d := range []string{"identity", "etc", "proc", "init"} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return syslogEnv{
		keptFile:   filepath.Join(dir, "identity", "syslog-remote"),
		configFile: filepath.Join(dir, "etc", "kvm", "syslog-remote"),
		initScript: filepath.Join(dir, "init", "S01syslogd"),
		procDir:    filepath.Join(dir, "proc"),
		restart:    func(context.Context, string) error { return nil },
		now:        time.Now,
	}
}

func TestSyslogWriteBothAndClear(t *testing.T) {
	env := testSyslogEnv(t)

	if err := env.write("10.0.0.40:5515"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{env.keptFile, env.configFile} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "10.0.0.40:5515\n" {
			t.Errorf("%s holds %q", name, data)
		}
		info, _ := os.Stat(name)
		if info.Mode().Perm() != 0o600 {
			t.Errorf("%s mode %v", name, info.Mode().Perm())
		}
	}
	if got := env.target(); got != "10.0.0.40:5515" {
		t.Errorf("target() = %q", got)
	}

	// No temporary files are left beside them.
	for _, name := range []string{env.keptFile, env.configFile} {
		entries, _ := os.ReadDir(filepath.Dir(name))
		if len(entries) != 1 {
			t.Errorf("%s has %d entries", filepath.Dir(name), len(entries))
		}
	}

	if err := env.write(""); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{env.keptFile, env.configFile} {
		if _, err := os.Stat(name); !os.IsNotExist(err) {
			t.Errorf("%s still exists: %v", name, err)
		}
	}
	if got := env.target(); got != "" {
		t.Errorf("target() after clearing = %q", got)
	}

	// Clearing what is already clear is not an error.
	if err := env.write(""); err != nil {
		t.Fatal(err)
	}
}

func TestSyslogKeptCopyWins(t *testing.T) {
	env := testSyslogEnv(t)
	_ = os.MkdirAll(filepath.Dir(env.configFile), 0o755)
	_ = os.WriteFile(env.configFile, []byte("image-copy:514\n"), 0o600)
	if got := env.target(); got != "image-copy:514" {
		t.Errorf("fallback target() = %q", got)
	}
	_ = os.WriteFile(env.keptFile, []byte("kept:514\n"), 0o600)
	if got := env.target(); got != "kept:514" {
		t.Errorf("target() = %q, want the kept copy", got)
	}
}

func TestSyslogWriteWithoutKeptIdentity(t *testing.T) {
	env := testSyslogEnv(t)
	_ = os.RemoveAll(filepath.Dir(env.keptFile))

	if err := env.write("logs"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(env.keptFile); !os.IsNotExist(err) {
		t.Errorf("kept copy written without /data/identity: %v", err)
	}
	if got := env.target(); got != "logs" {
		t.Errorf("target() = %q", got)
	}
}

func writeCmdline(t *testing.T, procDir, pid string, args ...string) {
	t.Helper()
	dir := filepath.Join(procDir, pid)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	data := strings.Join(args, "\x00") + "\x00"
	if err := os.WriteFile(filepath.Join(dir, "cmdline"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestActiveSyslogTarget(t *testing.T) {
	env := testSyslogEnv(t)
	if got := activeSyslogTarget(env.procDir); got != "" {
		t.Errorf("no processes: %q", got)
	}

	writeCmdline(t, env.procDir, "1", "init")
	writeCmdline(t, env.procDir, "90", "klogd")
	writeCmdline(t, env.procDir, "self", "syslogd", "-R", "not-a-pid:1")
	// A kernel thread has an empty cmdline.
	_ = os.MkdirAll(filepath.Join(env.procDir, "2"), 0o755)
	_ = os.WriteFile(filepath.Join(env.procDir, "2", "cmdline"), nil, 0o644)

	writeCmdline(t, env.procDir, "80", "syslogd")
	if got := activeSyslogTarget(env.procDir); got != "" {
		t.Errorf("local syslogd: %q", got)
	}

	writeCmdline(t, env.procDir, "80", "/sbin/syslogd", "-L", "-R", "10.0.0.40:5515")
	if got := activeSyslogTarget(env.procDir); got != "10.0.0.40:5515" {
		t.Errorf("forwarding syslogd: %q", got)
	}

	writeCmdline(t, env.procDir, "80", "syslogd", "-L", "-R[::1]:514")
	if got := activeSyslogTarget(env.procDir); got != "[::1]:514" {
		t.Errorf("attached -R: %q", got)
	}
}

func TestSyslogSupported(t *testing.T) {
	env := testSyslogEnv(t)
	if env.supported() {
		t.Error("supported without a script")
	}

	_ = os.WriteFile(env.initScript, []byte("#!/bin/sh\nsyslogd\nklogd\n"), 0o755)
	if env.supported() {
		t.Error("an old script counted as supported")
	}

	_ = os.WriteFile(env.initScript, []byte("#!/bin/sh\nCONFIG=${SYSLOG_REMOTE:-/etc/kvm/syslog-remote}\n"), 0o755)
	if !env.supported() {
		t.Error("a script that reads the setting counted as unsupported")
	}

	writeCmdline(t, env.procDir, "80", "syslogd", "-L", "-R", "logs:514")
	_ = os.WriteFile(env.keptFile, []byte("logs:514\n"), 0o600)
	state := env.state()
	if state.Target != "logs:514" || state.Active != "logs:514" || !state.Supported {
		t.Errorf("state() = %+v", state)
	}
}

func TestSendLocalSyslog(t *testing.T) {
	socket := filepath.Join(t.TempDir(), "log")
	conn, err := net.ListenUnixgram("unixgram", &net.UnixAddr{Name: socket, Net: "unixgram"})
	if err != nil {
		t.Skipf("no unix datagram sockets here: %s", err)
	}
	defer func() { _ = conn.Close() }()

	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	if err := sendLocalSyslog(socket, "ironkvm", "IronKVM test message x", now); err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 512)
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	got := string(buf[:n])
	if !strings.HasPrefix(got, "<13>Sep 30 12:00:00 ironkvm[") || !strings.HasSuffix(got, "]: IronKVM test message x") {
		t.Errorf("sent %q", got)
	}

	if err := sendLocalSyslog(filepath.Join(t.TempDir(), "none"), "ironkvm", "x", now); err == nil {
		t.Error("sending to a missing socket succeeded")
	}
}
