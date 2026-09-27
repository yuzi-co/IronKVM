//go:build linux

package tailscale

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
)

func TestVersionFromPackageURL(t *testing.T) {
	v, err := versionFromPackageURL("https://pkgs.tailscale.com/stable/tailscale_1.90.1_riscv64.tgz")
	if err != nil || v != "1.90.1" {
		t.Fatalf("got %q %v", v, err)
	}
	if _, err := versionFromPackageURL("https://pkgs.tailscale.com/stable/tailscale_latest_riscv64.tgz"); err == nil {
		t.Fatal("the unresolved alias has no version")
	}
}

// A fake release server: the latest alias redirects to a versioned package.
func TestResolveRedirectFollowsToThePackage(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/stable/tailscale_latest_riscv64.tgz", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/stable/tailscale_1.90.1_riscv64.tgz", http.StatusFound)
	})
	mux.HandleFunc("/stable/tailscale_1.90.1_riscv64.tgz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Errorf("the check must not download the package, got %s", r.Method)
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resolved, err := resolveRedirect(srv.Client(), srv.URL+"/stable/tailscale_latest_riscv64.tgz")
	if err != nil {
		t.Fatal(err)
	}
	if v, err := versionFromPackageURL(resolved); err != nil || v != "1.90.1" {
		t.Fatalf("got %q %v from %s", v, err, resolved)
	}
}

func TestGetUpdateReportsCurrentAndLatestAndCaches(t *testing.T) {
	lifecycle(t)
	savedCache, savedFetch := updates, fetchLatest
	t.Cleanup(func() { updates, fetchLatest = savedCache, savedFetch })
	updates = &vpn.VersionCache{TTL: vpn.UpdateTTL}
	fetches := 0
	fetchLatest = func() (string, error) { fetches++; return "1.90.1", nil }

	for i := 0; i < 2; i++ {
		rsp := call(t, NewService().GetUpdate, "")
		data, _ := rsp.Data.(map[string]any)
		if rsp.Code != 0 || data["current"] != "1.88.3" || data["latest"] != "1.90.1" {
			t.Fatalf("got %d %q %v", rsp.Code, rsp.Msg, rsp.Data)
		}
	}
	if fetches != 1 {
		t.Fatalf("the release server was asked %d times within the hour", fetches)
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
	if got := callsOf(t, calls); got != "install\n" {
		t.Fatalf("a stopped daemon must stay stopped:\n%s", got)
	}

	_ = os.Remove(calls)
	fakeRunning(t, addon.Tailscale, 4343, "/usr/sbin/tailscaled")
	if rsp := call(t, s.Update, ""); rsp.Code != 0 {
		t.Fatalf("update: %q", rsp.Msg)
	}
	if got := callsOf(t, calls); got != "script stop\ninstall\nscript start\n" {
		t.Fatalf("a running daemon must run again:\n%s", got)
	}
}
