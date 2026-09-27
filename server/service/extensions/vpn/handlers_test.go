//go:build linux

package vpn

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"

	"github.com/gin-gonic/gin"
)

// scratchNetBird points NetBird's pid file, /proc and both init.d directories
// at a temporary tree, off a distribution image, and gives it a boot script
// with body.
func scratchNetBird(t *testing.T, body string) {
	t.Helper()
	base := t.TempDir()
	saved := struct {
		initd, pkg, proc, marker string
		ts, nb                   addon.Daemon
	}{addon.InitdDir, addon.PkgInitdDir, addon.ProcDir, addon.DistroMarker, addon.Tailscale, addon.NetBird}
	t.Cleanup(func() {
		addon.InitdDir, addon.PkgInitdDir, addon.ProcDir, addon.DistroMarker = saved.initd, saved.pkg, saved.proc, saved.marker
		addon.Tailscale, addon.NetBird = saved.ts, saved.nb
	})
	addon.InitdDir = filepath.Join(base, "etc-init.d")
	addon.PkgInitdDir = filepath.Join(base, "kvmapp-init.d")
	addon.ProcDir = filepath.Join(base, "proc")
	addon.DistroMarker = filepath.Join(base, "absent")
	addon.Tailscale.PidFile = filepath.Join(base, "tailscaled.pid")
	addon.NetBird.PidFile = filepath.Join(base, "netbird.pid")
	for _, d := range []string{addon.InitdDir, addon.PkgInitdDir, addon.ProcDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(addon.NetBird.Script(), []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func fakeRunning(t *testing.T, d addon.Daemon, pid int) {
	t.Helper()
	if err := os.WriteFile(d.PidFile, []byte(strconv.Itoa(pid)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(addon.ProcDir, strconv.Itoa(pid))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cmdline"), []byte("/usr/bin/"+d.Process+"\x00"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func respond(t *testing.T, h func(c *gin.Context)) proto.Response {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	h(c)
	var rsp proto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil {
		t.Fatalf("response %q: %v", w.Body.String(), err)
	}
	return rsp
}

func TestStopDaemonWithNothingRunningIsNotAnError(t *testing.T) {
	scratchNetBird(t, `echo "Stopping netbird: FAIL"`)
	if err := StopDaemon(addon.NetBird); err != nil {
		t.Fatal(err)
	}
}

func TestStopDaemonThatLeftItRunningFails(t *testing.T) {
	scratchNetBird(t, `echo "Stopping netbird: FAIL"`)
	fakeRunning(t, addon.NetBird, 4242)
	err := StopDaemon(addon.NetBird)
	if err == nil || !strings.Contains(Message("stop failed", err), "Stopping netbird: FAIL") {
		t.Fatalf("got %v", err)
	}
}

func TestBusyAnswersASecondCaller(t *testing.T) {
	var b Busy
	var end func()
	respond(t, func(c *gin.Context) {
		end = b.Begin(c, addon.NetBird)
		c.JSON(http.StatusOK, proto.Response{})
	})
	if end == nil {
		t.Fatal("the first caller must get in")
	}
	rsp := respond(t, func(c *gin.Context) {
		if b.Begin(c, addon.NetBird) != nil {
			t.Error("a second caller must not get in")
		}
	})
	if rsp.Code != -1 || !strings.Contains(rsp.Msg, "NetBird is busy") {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	end()
	respond(t, func(c *gin.Context) {
		if e := b.Begin(c, addon.NetBird); e == nil {
			t.Error("after the first one ends the next gets in")
		} else {
			e()
			c.JSON(http.StatusOK, proto.Response{})
		}
	})
}

func TestGuardRefusesWithoutRunningTheAction(t *testing.T) {
	scratchNetBird(t, "")
	fakeRunning(t, addon.Tailscale, 100)
	ran := false
	rsp := respond(t, func(c *gin.Context) {
		if Guard(c, addon.NetBird, "start failed", func() error { ran = true; return nil }) {
			t.Error("a refused action is not a success")
		}
	})
	if ran || !strings.Contains(rsp.Msg, "Tailscale is running or starts at boot") {
		t.Fatalf("ran %v, got %q", ran, rsp.Msg)
	}
}

func TestGuardReportsWhatFailed(t *testing.T) {
	scratchNetBird(t, "")
	rsp := respond(t, func(c *gin.Context) {
		Guard(c, addon.NetBird, "start failed", func() error {
			return &CmdError{Err: errors.New("exit status 1"), Tail: "Starting netbird: FAIL"}
		})
	})
	if rsp.Code != -1 || rsp.Msg != "start failed:\nStarting netbird: FAIL" {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
}
