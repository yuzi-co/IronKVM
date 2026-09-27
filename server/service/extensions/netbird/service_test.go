//go:build linux

package netbird

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"

	"github.com/gin-gonic/gin"
)

// lifecycle is scratchImage off a distribution image, with a stub netbird
// that answers from the board's status fixture, a stub boot script in the
// package copy, and scratch pid files, /proc and /etc/init.d. Every stub
// appends to the file it returns. STUB_HANG makes `netbird status` hang.
func lifecycle(t *testing.T) (calls string) {
	t.Helper()
	fsroot := scratchImage(t, false)
	base := filepath.Dir(fsroot)
	calls = filepath.Join(base, "calls")
	fixture, err := filepath.Abs("testdata/status-connected.json")
	if err != nil {
		t.Fatal(err)
	}

	savedTs, savedNb := addon.Tailscale, addon.NetBird
	savedInitd, savedPkg, savedProc := addon.InitdDir, addon.PkgInitdDir, addon.ProcDir
	savedStatus, savedCache, savedFetch, savedInstall := statusTimeout, updates, fetchLatest, installPackage
	t.Cleanup(func() {
		addon.Tailscale, addon.NetBird = savedTs, savedNb
		addon.InitdDir, addon.PkgInitdDir, addon.ProcDir = savedInitd, savedPkg, savedProc
		statusTimeout, updates, fetchLatest, installPackage = savedStatus, savedCache, savedFetch, savedInstall
	})
	addon.InitdDir = filepath.Join(base, "etc-init.d")
	addon.PkgInitdDir = filepath.Join(base, "kvmapp-init.d")
	addon.ProcDir = filepath.Join(base, "proc")
	for _, d := range []string{addon.InitdDir, addon.PkgInitdDir, addon.ProcDir, filepath.Join(base, "run")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	addon.Tailscale.PidFile = filepath.Join(base, "run", "tailscaled.pid")
	addon.NetBird.PidFile = filepath.Join(base, "run", "netbird.pid")
	updates = &vpn.VersionCache{TTL: vpn.UpdateTTL}

	r := strings.NewReplacer("CALLS", calls, "FIXTURE", fixture, "PIDFILE", addon.NetBird.PidFile, "PROCDIR", addon.ProcDir)
	stub(t, filepath.Join(addon.PkgInitdDir, "S98netbird"), r.Replace(`echo "script $1" >> "CALLS"
case "$1" in
start)
	if [ -n "$STUB_RUN" ]; then
		sleep 0.2
		echo 4242 > "PIDFILE"
		mkdir -p "PROCDIR/4242"
		printf "/usr/bin/netbird\000service\000run\000" > "PROCDIR/4242/cmdline"
	fi
	echo "Starting netbird: ${STUB_RESULT:-OK}" ;;
stop) echo "Stopping netbird: ${STUB_STOP:-OK}" ;;
esac`))
	stub(t, NetbirdPath, r.Replace(`echo "netbird $*" | sed "s| [^ ]*netbird-setup-key[^ ]*| KEYFILE|" >> "CALLS"
case "$1" in
version) echo "0.78.2" ;;
status) [ -z "$STUB_HANG" ] || exec sleep 30; cat "FIXTURE" ;;
up)
	if [ "$2" = "--no-browser" ]; then
		printf 'Use this URL to log in:\n\nhttps://login.netbird.io/activate?user_code=ABCD-EFGH \n\n'
		exit 0
	fi
	if [ "$2" = "--setup-key-file" ]; then echo "key $(cat "$3") $(stat -c %a "$3")" >> "CALLS"; fi
	echo "Connected" ;;
deregister) echo "Deregistered successfully" ;;
esac`))
	return calls
}

func fakeRunning(t *testing.T, d addon.Daemon, pid int, argv0 string) {
	t.Helper()
	if err := os.WriteFile(d.PidFile, []byte(strconv.Itoa(pid)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(addon.ProcDir, strconv.Itoa(pid))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cmdline"), []byte(argv0+"\x00service\x00run\x00"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func call(t *testing.T, h gin.HandlerFunc, body string) proto.Response {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h(c)
	var rsp proto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil {
		t.Fatalf("response %q: %v", w.Body.String(), err)
	}
	return rsp
}

func data(t *testing.T, rsp proto.Response) map[string]any {
	t.Helper()
	m, ok := rsp.Data.(map[string]any)
	if rsp.Code != 0 || !ok {
		t.Fatalf("got %d %q %v", rsp.Code, rsp.Msg, rsp.Data)
	}
	return m
}

func TestEntryPointsRefuseWhileTailscaleRuns(t *testing.T) {
	calls := lifecycle(t)
	fakeRunning(t, addon.Tailscale, 100, "/usr/sbin/tailscaled")
	s := NewService()
	for name, c := range map[string]struct {
		h    gin.HandlerFunc
		body string
	}{
		"install": {s.Install, ""},
		"start":   {s.Start, ""},
		"up":      {s.Up, ""},
		"login":   {s.Login, `{"setupKey":"A1B2C3D4-E5F6-4789-ABCD-0123456789EF"}`},
		"boot":    {s.Boot, `{"enabled":true}`},
	} {
		rsp := call(t, c.h, c.body)
		if rsp.Code == 0 || !strings.Contains(rsp.Msg, "Tailscale is running or starts at boot") {
			t.Fatalf("%s: got %d %q", name, rsp.Code, rsp.Msg)
		}
	}
	if got := callsOf(t, calls); got != "" {
		t.Fatalf("nothing may run while refused, ran:\n%s", got)
	}
}

func TestLoginTrimsAndUppercasesTheSetupKey(t *testing.T) {
	calls := lifecycle(t)
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	if rsp := call(t, NewService().Login, `{"setupKey":"  a1b2c3d4-e5f6-4789-abcd-0123456789ef \n"}`); rsp.Code != 0 {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	if got := callsOf(t, calls); got != "netbird up --setup-key-file KEYFILE\nkey A1B2C3D4-E5F6-4789-ABCD-0123456789EF 600\n" {
		t.Fatalf("calls:\n%s", got)
	}
}

func TestLoginStartsAStoppedDaemonFirst(t *testing.T) {
	calls := lifecycle(t)
	if rsp := call(t, NewService().Login, `{"setupKey":"A1B2C3D4-E5F6-4789-ABCD-0123456789EF"}`); rsp.Code != 0 {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	if got := callsOf(t, calls); got != "script start\nnetbird up --setup-key-file KEYFILE\nkey A1B2C3D4-E5F6-4789-ABCD-0123456789EF 600\n" {
		t.Fatalf("calls:\n%s", got)
	}
}

func TestLoginWithoutAKeyReturnsTheSSOURL(t *testing.T) {
	lifecycle(t)
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	d := data(t, call(t, NewService().Login, ""))
	if d["url"] != "https://login.netbird.io/activate?user_code=ABCD-EFGH" {
		t.Fatalf("got %v", d)
	}
}

func TestGetStatusRunning(t *testing.T) {
	lifecycle(t)
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	d := data(t, call(t, NewService().GetStatus, ""))
	if d["state"] != "running" || d["ip"] != "100.73.212.105" || d["account"] != "api.netbird.io" || d["blockedBy"] != "" {
		t.Fatalf("got %v", d)
	}
}

func TestStatusNotInstalledStillNamesTheBlocker(t *testing.T) {
	lifecycle(t)
	if err := os.Remove(NetbirdPath); err != nil {
		t.Fatal(err)
	}
	fakeRunning(t, addon.Tailscale, 100, "/usr/sbin/tailscaled")
	d := data(t, call(t, NewService().GetStatus, ""))
	if d["state"] != "notInstall" || d["blockedBy"] != "tailscale" {
		t.Fatalf("got %v", d)
	}
}

// With the management server unreachable the CLI can hang. The page must not.
func TestStatusDoesNotHangOnAStuckCLI(t *testing.T) {
	lifecycle(t)
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	t.Setenv("STUB_HANG", "1")
	statusTimeout = 300 * time.Millisecond

	start := time.Now()
	d := data(t, call(t, NewService().GetStatus, ""))
	if d["state"] != "notRunning" {
		t.Fatalf("got %v", d)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("status took %s", elapsed)
	}
}

func TestLogoutDeregisters(t *testing.T) {
	calls := lifecycle(t)
	if rsp := call(t, NewService().Logout, ""); rsp.Code != 0 {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	if got := callsOf(t, calls); got != "netbird deregister\n" {
		t.Fatalf("calls:\n%s", got)
	}
}

func TestGetUpdateReportsCurrentAndLatest(t *testing.T) {
	lifecycle(t)
	fetchLatest = func() (string, error) { return "0.78.2", nil }
	d := data(t, call(t, NewService().GetUpdate, ""))
	if d["current"] != "0.78.2" || d["latest"] != "0.78.2" {
		t.Fatalf("got %v", d)
	}
}

func TestUpdateKeepsAStoppedDaemonStopped(t *testing.T) {
	calls := lifecycle(t)
	installPackage = func() error {
		f, err := os.OpenFile(calls, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		_, err = f.WriteString("install\n")
		return err
	}
	s := NewService()

	if rsp := call(t, s.Update, ""); rsp.Code != 0 {
		t.Fatalf("update: %q", rsp.Msg)
	}
	if got := callsOf(t, calls); got != "install\n" {
		t.Fatalf("a stopped daemon must stay stopped:\n%s", got)
	}

	_ = os.Remove(calls)
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	if rsp := call(t, s.Update, ""); rsp.Code != 0 {
		t.Fatalf("update: %q", rsp.Msg)
	}
	if got := callsOf(t, calls); got != "script stop\ninstall\nscript start\n" {
		t.Fatalf("a running daemon must run again:\n%s", got)
	}
}

const setupKey = "A1B2C3D4-E5F6-4789-ABCD-0123456789EF"

// A key that is not a NetBird key is refused before anything runs, and the
// message does not repeat what was pasted.
func TestLoginRefusesAKeyInAnotherFormat(t *testing.T) {
	calls := lifecycle(t)
	rsp := call(t, NewService().Login, `{"setupKey":"tskey-auth-SECRETVALUE"}`)
	if rsp.Code != -1 || strings.Contains(rsp.Msg, "SECRETVALUE") || !strings.Contains(rsp.Msg, "not in NetBird's format") {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	if got := callsOf(t, calls); got != "" {
		t.Fatalf("nothing may run for a malformed key, ran:\n%s", got)
	}
}

func TestLoginLeavesNoKeyFileBehind(t *testing.T) {
	lifecycle(t)
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	if rsp := call(t, NewService().Login, `{"setupKey":"`+setupKey+`"}`); rsp.Code != 0 {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	if entries, _ := os.ReadDir(KeyDir); len(entries) != 0 {
		t.Fatalf("the key file must be removed, found %v", entries)
	}
}

// Two starts at once, NetBird's through the page and Tailscale's through the
// same lock: exactly one goes ahead.
func TestStartRacesTailscaleForTheLock(t *testing.T) {
	lifecycle(t)
	t.Setenv("STUB_RUN", "1")
	done := make(chan proto.Response)
	go func() { done <- call(t, NewService().Start, "") }()
	time.Sleep(20 * time.Millisecond)
	tsErr := addon.Exclusive("tailscale", func() error {
		fakeRunning(t, addon.Tailscale, 100, "/usr/sbin/tailscaled")
		return nil
	})
	rsp := <-done
	if (rsp.Code == 0) == (tsErr == nil) {
		t.Fatalf("exactly one start may go ahead: netbird %d %q, tailscale %v", rsp.Code, rsp.Msg, tsErr)
	}
}

// Restart of a stopped daemon is a start, and is refused like one.
func TestRestartOfAStoppedDaemonIsRefusedWhileTailscaleRuns(t *testing.T) {
	calls := lifecycle(t)
	fakeRunning(t, addon.Tailscale, 100, "/usr/sbin/tailscaled")
	rsp := call(t, NewService().Restart, "")
	if rsp.Code == 0 || !strings.Contains(rsp.Msg, "Tailscale is running or starts at boot") {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	if got := callsOf(t, calls); got != "" {
		t.Fatalf("nothing may run while refused, ran:\n%s", got)
	}
}

func TestStartRefreshesTheBootCopy(t *testing.T) {
	lifecycle(t)
	stale := filepath.Join(addon.InitdDir, "S98netbird")
	if err := os.WriteFile(stale, []byte("#!/bin/sh\n# old\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if rsp := call(t, NewService().Start, ""); rsp.Code != 0 {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	got, _ := os.ReadFile(stale)
	want, _ := os.ReadFile(addon.NetBird.Script())
	if string(got) != string(want) {
		t.Fatalf("the boot copy was not refreshed: %q", got)
	}
}

// installBlocks makes installPackage wait until the returned channel closes.
func installBlocks(t *testing.T) (release chan struct{}, started chan struct{}) {
	t.Helper()
	release, started = make(chan struct{}), make(chan struct{})
	binary, err := os.ReadFile(NetbirdPath)
	if err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(NetbirdPath)
	installPackage = func() error {
		close(started)
		<-release
		return os.WriteFile(NetbirdPath, binary, 0o755)
	}
	return release, started
}

func TestSecondInstallUpdateOrUninstallIsBusy(t *testing.T) {
	lifecycle(t)
	release, started := installBlocks(t)
	s := NewService()
	done := make(chan proto.Response)
	go func() { done <- call(t, s.Install, "") }()
	<-started
	for name, h := range map[string]gin.HandlerFunc{"install": s.Install, "update": s.Update, "uninstall": s.Uninstall} {
		if rsp := call(t, h, ""); rsp.Code != -1 || !strings.Contains(rsp.Msg, "NetBird is busy") {
			t.Fatalf("%s: got %d %q", name, rsp.Code, rsp.Msg)
		}
	}
	close(release)
	if rsp := <-done; rsp.Code != 0 {
		t.Fatalf("the first install: %d %q", rsp.Code, rsp.Msg)
	}
}

// The download runs without the lock; the start after it checks again.
func TestInstallChecksAgainBeforeItStarts(t *testing.T) {
	calls := lifecycle(t)
	_ = os.Remove(NetbirdPath)
	installPackage = func() error {
		fakeRunning(t, addon.Tailscale, 100, "/usr/sbin/tailscaled")
		stub(t, NetbirdPath, "exit 0")
		return nil
	}
	rsp := call(t, NewService().Install, "")
	if rsp.Code != -1 || !strings.Contains(rsp.Msg, "Tailscale is running or starts at boot") {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	if strings.Contains(callsOf(t, calls), "script start") {
		t.Fatal("netbird must not start while tailscale runs")
	}
}

func TestUpdateAndUninstallStopOnAFailedStop(t *testing.T) {
	calls := lifecycle(t)
	installPackage = func() error { t.Error("nothing may be installed after a failed stop"); return nil }
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	t.Setenv("STUB_STOP", "FAIL")
	s := NewService()
	for name, h := range map[string]gin.HandlerFunc{"update": s.Update, "uninstall": s.Uninstall} {
		if rsp := call(t, h, ""); rsp.Code != -1 || !strings.Contains(rsp.Msg, "Stopping netbird: FAIL") {
			t.Fatalf("%s: got %d %q", name, rsp.Code, rsp.Msg)
		}
	}
	if !isInstalled() {
		t.Fatal("uninstall must keep the binary when the daemon did not stop")
	}
	if strings.Contains(callsOf(t, calls), "script start") {
		t.Fatal("nothing may start after a failed stop")
	}
}
