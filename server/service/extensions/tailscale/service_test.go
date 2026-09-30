//go:build linux

package tailscale

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/utils"

	"github.com/gin-gonic/gin"
)

func stub(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// lifecycle is scratchImage off a distribution image, with stub binaries, a
// stub boot script in the package copy, and scratch pid files, /proc and
// /etc/init.d. Every stub appends to the file it returns. STUB_RESULT=FAIL
// makes the script's start print FAIL; STUB_FAIL makes `tailscale up` fail.
func lifecycle(t *testing.T) (calls string) {
	t.Helper()
	fsroot, _ := scratchImage(t, false)
	base := filepath.Dir(fsroot)
	calls = filepath.Join(base, "calls")

	savedTs, savedNb := addon.Tailscale, addon.NetBird
	savedInitd, savedPkg, savedProc := addon.InitdDir, addon.PkgInitdDir, addon.ProcDir
	savedLimit := utils.GoMemLimitFile
	t.Cleanup(func() {
		addon.Tailscale, addon.NetBird = savedTs, savedNb
		addon.InitdDir, addon.PkgInitdDir, addon.ProcDir = savedInitd, savedPkg, savedProc
		utils.GoMemLimitFile = savedLimit
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
	utils.GoMemLimitFile = filepath.Join(base, "GOMEMLIMIT")

	stub(t, filepath.Join(addon.PkgInitdDir, "S98tailscaled"), `echo "script $1" >> "`+calls+`"
case "$1" in
start) echo "GOMEMLIMIT set to 56MiB"; echo "Starting tailscaled[1.2.3]: ${STUB_RESULT:-OK}" ;;
stop) echo "Stopping tailscaled: ${STUB_STOP:-OK}" ;;
esac`)
	stub(t, TailscalePath, `echo "tailscale $*" >> "`+calls+`"
case "$1" in
version) echo "1.88.3"; echo "  tailscale commit: abc" ;;
up) if [ -n "$STUB_FAIL" ]; then echo "backend error: tailscaled is not running" >&2; exit 1; fi ;;
status) echo "{\"BackendState\": \"${STUB_STATE:-Stopped}\"}" ;;
esac`)
	stub(t, TailscaledPath, `exit 0`)
	return calls
}

// fakeRunning writes d's pid file and a /proc entry whose command line starts
// with argv0.
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

func callsOf(t *testing.T, calls string) string {
	t.Helper()
	b, err := os.ReadFile(calls)
	if err != nil {
		return ""
	}
	return string(b)
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

func TestEntryPointsRefuseWhileNetBirdRuns(t *testing.T) {
	calls := lifecycle(t)
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	s := NewService()
	for name, h := range map[string]gin.HandlerFunc{
		"install": s.Install, "start": s.Start, "up": s.Up, "login": s.Login,
	} {
		rsp := call(t, h, "")
		if rsp.Code == 0 || !strings.Contains(rsp.Msg, "NetBird is running or starts at boot") {
			t.Fatalf("%s: got %d %q", name, rsp.Code, rsp.Msg)
		}
	}
	if rsp := call(t, s.Boot, `{"enabled":true}`); rsp.Code == 0 || !strings.Contains(rsp.Msg, "NetBird") {
		t.Fatalf("boot: got %d %q", rsp.Code, rsp.Msg)
	}
	if got := callsOf(t, calls); got != "" {
		t.Fatalf("nothing may run while refused, ran:\n%s", got)
	}
	if addon.BootEnabled(addon.Tailscale) {
		t.Fatal("start at boot must stay off")
	}
}

func TestEntryPointsRefuseWhileNetBirdStartsAtBoot(t *testing.T) {
	calls := lifecycle(t)
	if err := os.WriteFile(filepath.Join(addon.InitdDir, "S98netbird"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if rsp := call(t, NewService().Start, ""); rsp.Code == 0 || !strings.Contains(rsp.Msg, "NetBird") {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	if got := callsOf(t, calls); got != "" {
		t.Fatalf("nothing may run while refused, ran:\n%s", got)
	}
}

func TestBootOffIsAllowedWhileBlocked(t *testing.T) {
	lifecycle(t)
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	if rsp := call(t, NewService().Boot, `{"enabled":false}`); rsp.Code != 0 {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
}

func TestStartAndStopLeaveStartAtBootAlone(t *testing.T) {
	calls := lifecycle(t)
	s := NewService()
	if rsp := call(t, s.Start, ""); rsp.Code != 0 {
		t.Fatalf("start: %q", rsp.Msg)
	}
	if addon.BootEnabled(addon.Tailscale) {
		t.Fatal("start must not turn on start at boot")
	}
	if rsp := call(t, s.Boot, `{"enabled":true}`); rsp.Code != 0 {
		t.Fatalf("boot: %q", rsp.Msg)
	}
	if rsp := call(t, s.Stop, ""); rsp.Code != 0 {
		t.Fatalf("stop: %q", rsp.Msg)
	}
	if !addon.BootEnabled(addon.Tailscale) {
		t.Fatal("stop must not turn off start at boot")
	}
	if got := callsOf(t, calls); got != "script start\nscript stop\n" {
		t.Fatalf("calls:\n%s", got)
	}
}

func TestStartFailureCarriesTheScriptOutput(t *testing.T) {
	lifecycle(t)
	t.Setenv("STUB_RESULT", "FAIL")
	rsp := call(t, NewService().Start, "")
	if rsp.Code != -1 || !strings.Contains(rsp.Msg, "Starting tailscaled[1.2.3]: FAIL") {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
}

func TestUpFailureCarriesTheCLIOutput(t *testing.T) {
	lifecycle(t)
	t.Setenv("STUB_FAIL", "1")
	rsp := call(t, NewService().Up, "")
	if rsp.Code != -1 || !strings.Contains(rsp.Msg, "backend error: tailscaled is not running") {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
}

func TestStartNoLongerWritesGoMemLimit(t *testing.T) {
	lifecycle(t)
	if rsp := call(t, NewService().Start, ""); rsp.Code != 0 {
		t.Fatalf("start: %q", rsp.Msg)
	}
	if utils.IsGoMemLimitExist() {
		t.Fatal("S98tailscaled derives the limit; the server must not write it")
	}
}

// Restart of a stopped daemon is a start, and is refused like one.
func TestRestartOfAStoppedDaemonIsRefusedWhileNetBirdRuns(t *testing.T) {
	calls := lifecycle(t)
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	rsp := call(t, NewService().Restart, "")
	if rsp.Code == 0 || !strings.Contains(rsp.Msg, "NetBird is running or starts at boot") {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	if got := callsOf(t, calls); got != "" {
		t.Fatalf("nothing may run while refused, ran:\n%s", got)
	}
}

func TestStartRefreshesTheBootCopy(t *testing.T) {
	lifecycle(t)
	stale := filepath.Join(addon.InitdDir, "S98tailscaled")
	if err := os.WriteFile(stale, []byte("#!/bin/sh\n# old\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if rsp := call(t, NewService().Start, ""); rsp.Code != 0 {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	got, _ := os.ReadFile(stale)
	want, _ := os.ReadFile(addon.Tailscale.Script())
	if string(got) != string(want) {
		t.Fatalf("the boot copy was not refreshed: %q", got)
	}
}

func TestSecondInstallUpdateOrUninstallIsBusy(t *testing.T) {
	lifecycle(t)
	savedInstall := installPackage
	t.Cleanup(func() { installPackage = savedInstall })
	binary, err := os.ReadFile(TailscalePath)
	if err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(TailscalePath)
	release, started := make(chan struct{}), make(chan struct{})
	installPackage = func() error {
		close(started)
		<-release
		return os.WriteFile(TailscalePath, binary, 0o755)
	}
	s := NewService()
	done := make(chan proto.Response)
	go func() { done <- call(t, s.Install, "") }()
	<-started
	for name, h := range map[string]gin.HandlerFunc{"install": s.Install, "update": s.Update, "uninstall": s.Uninstall} {
		if rsp := call(t, h, ""); rsp.Code != -1 || !strings.Contains(rsp.Msg, "Tailscale is busy") {
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
	savedInstall := installPackage
	t.Cleanup(func() { installPackage = savedInstall })
	binary, err := os.ReadFile(TailscalePath)
	if err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(TailscalePath)
	installPackage = func() error {
		fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
		return os.WriteFile(TailscalePath, binary, 0o755)
	}
	rsp := call(t, NewService().Install, "")
	if rsp.Code != -1 || !strings.Contains(rsp.Msg, "NetBird is running or starts at boot") {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	if strings.Contains(callsOf(t, calls), "script start") {
		t.Fatal("tailscale must not start while netbird runs")
	}
}

func TestUpdateAndUninstallStopOnAFailedStop(t *testing.T) {
	calls := lifecycle(t)
	savedInstall := installPackage
	t.Cleanup(func() { installPackage = savedInstall })
	installPackage = func() error { t.Error("nothing may be installed after a failed stop"); return nil }
	fakeRunning(t, addon.Tailscale, 4343, "/usr/sbin/tailscaled")
	t.Setenv("STUB_STOP", "FAIL")
	s := NewService()
	for name, h := range map[string]gin.HandlerFunc{"update": s.Update, "uninstall": s.Uninstall} {
		if rsp := call(t, h, ""); rsp.Code != -1 || !strings.Contains(rsp.Msg, "Stopping tailscaled: FAIL") {
			t.Fatalf("%s: got %d %q", name, rsp.Code, rsp.Msg)
		}
	}
	if !isInstalled() {
		t.Fatal("uninstall must keep the binaries when the daemon did not stop")
	}
	if strings.Contains(callsOf(t, calls), "script start") {
		t.Fatal("nothing may start after a failed stop")
	}
}

func TestConnectStartsThenUps(t *testing.T) {
	calls := lifecycle(t)
	if rsp := call(t, NewService().Connect, ""); rsp.Code != 0 {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	want := "script start\ntailscale status --json\ntailscale up --accept-dns=false\n"
	if got := callsOf(t, calls); got != want {
		t.Fatalf("calls:\n%s", got)
	}
}

func TestConnectIsRefusedWhileNetBirdRuns(t *testing.T) {
	calls := lifecycle(t)
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	if rsp := call(t, NewService().Connect, ""); rsp.Code == 0 || !strings.Contains(rsp.Msg, "NetBird") {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	if got := callsOf(t, calls); got != "" {
		t.Fatalf("nothing may run while refused, ran:\n%s", got)
	}
}

func TestDisconnectDownsThenStops(t *testing.T) {
	calls := lifecycle(t)
	fakeRunning(t, addon.Tailscale, 4343, "/usr/sbin/tailscaled")
	if rsp := call(t, NewService().Disconnect, ""); rsp.Code != 0 {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	if got := callsOf(t, calls); got != "tailscale down\nscript stop\n" {
		t.Fatalf("calls:\n%s", got)
	}
}
