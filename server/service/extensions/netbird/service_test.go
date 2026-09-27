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

	r := strings.NewReplacer("CALLS", calls, "FIXTURE", fixture)
	stub(t, filepath.Join(addon.PkgInitdDir, "S98netbird"), r.Replace(`echo "script $1" >> "CALLS"
case "$1" in
start) echo "Starting netbird: ${STUB_RESULT:-OK}" ;;
stop) echo "Stopping netbird: OK" ;;
esac`))
	stub(t, NetbirdPath, r.Replace(`echo "netbird $*" >> "CALLS"
case "$1" in
version) echo "0.78.2" ;;
status) [ -z "$STUB_HANG" ] || sleep 30; cat "FIXTURE" ;;
up)
	if [ "$2" = "--no-browser" ]; then
		printf 'Use this URL to log in:\n\nhttps://login.netbird.io/activate?user_code=ABCD-EFGH \n\n'
		exit 0
	fi
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
		"login":   {s.Login, `{"setupKey":"KEY-123"}`},
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

func TestLoginTrimsTheSetupKey(t *testing.T) {
	calls := lifecycle(t)
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	if rsp := call(t, NewService().Login, `{"setupKey":"  KEY-123 \n"}`); rsp.Code != 0 {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	if got := callsOf(t, calls); got != "netbird up --setup-key KEY-123\n" {
		t.Fatalf("calls:\n%s", got)
	}
}

func TestLoginStartsAStoppedDaemonFirst(t *testing.T) {
	calls := lifecycle(t)
	if rsp := call(t, NewService().Login, `{"setupKey":"KEY-123"}`); rsp.Code != 0 {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	if got := callsOf(t, calls); got != "script start\nnetbird up --setup-key KEY-123\n" {
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
