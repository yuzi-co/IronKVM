//go:build linux

package tailscale

import (
	"os"
	"strings"
	"testing"

	"NanoKVM-Server/service/extensions/addon"
)

// The update on offer is the pin, and finding it asks nothing but the
// installed CLI.
func TestGetUpdateOffersThePin(t *testing.T) {
	calls := lifecycle(t)
	rsp := call(t, NewService().GetUpdate, "")
	data, _ := rsp.Data.(map[string]any)
	if rsp.Code != 0 || data["current"] != "1.88.3" || data["latest"] != Version {
		t.Fatalf("got %d %q %v", rsp.Code, rsp.Msg, rsp.Data)
	}
	if got := callsOf(t, calls); got != "tailscale version\n" {
		t.Fatalf("calls:\n%s", got)
	}
}

// An install from before the pin that is the pin or newer keeps running as
// it is: the update would be a reinstall or a downgrade.
func TestUpdateRefusesAnInstallAtOrPastThePin(t *testing.T) {
	calls := lifecycle(t)
	savedInstall := installPackage
	t.Cleanup(func() { installPackage = savedInstall })
	installPackage = func() error { t.Error("nothing may be installed over the pin or a newer release"); return nil }
	fakeRunning(t, addon.Tailscale, 4343, "/usr/sbin/tailscaled")
	for _, v := range []string{Version, "1.200.0"} {
		stub(t, TailscalePath, `echo "tailscale $*" >> "`+calls+`"
[ "$1" = version ] && echo "`+v+`"`)
		rsp := call(t, NewService().Update, "")
		if rsp.Code != -1 || !strings.Contains(rsp.Msg, "the pinned release is "+Version) {
			t.Fatalf("%s: got %d %q", v, rsp.Code, rsp.Msg)
		}
	}
	if got := callsOf(t, calls); strings.Contains(got, "script") {
		t.Fatalf("the daemon must be left alone:\n%s", got)
	}
}

func appendLine(path, line string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString(line + "\n")
	return err
}

func TestUpdateRestartsOnlyADaemonThatRan(t *testing.T) {
	calls := lifecycle(t)
	savedInstall := installPackage
	t.Cleanup(func() { installPackage = savedInstall })
	installPackage = func() error { return appendLine(calls, "install") }
	s := NewService()

	if rsp := call(t, s.Update, ""); rsp.Code != 0 {
		t.Fatalf("update: %q", rsp.Msg)
	}
	if got := callsOf(t, calls); got != "tailscale version\ninstall\n" {
		t.Fatalf("a stopped daemon must stay stopped:\n%s", got)
	}

	_ = os.Remove(calls)
	fakeRunning(t, addon.Tailscale, 4343, "/usr/sbin/tailscaled")
	if rsp := call(t, s.Update, ""); rsp.Code != 0 {
		t.Fatalf("update: %q", rsp.Msg)
	}
	if got := callsOf(t, calls); got != "tailscale version\nscript stop\ninstall\nscript start\n" {
		t.Fatalf("a running daemon must run again:\n%s", got)
	}
}
