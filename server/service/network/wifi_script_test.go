package network

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"NanoKVM-Server/proto"
)

const wifiInitScriptSource = "../../../kvmapp/system/init.d/S30wifi"

// genWpaConf runs gen_wpa_conf from the boot script. Sourced without an
// argument, the script only defines its functions.
func genWpaConf(t *testing.T, ssid, pass string) string {
	t.Helper()

	if _, err := os.Stat(wifiInitScriptSource); err != nil {
		t.Skipf("cannot read %s: %s", wifiInitScriptSource, err)
	}
	cmd := exec.Command("sh", "-c", `. "$0" && gen_wpa_conf "$1" "$2"`, wifiInitScriptSource, ssid, pass)
	cmd.Env = append(os.Environ(), "DEVINFO=/nonexistent")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gen_wpa_conf: %s: %s", err, out)
	}
	return string(out)
}

// An open network has no password, and wpa_passphrase refuses an empty one,
// which used to leave a configuration with no network in it.
func TestAnOpenNetworkGetsANetworkBlock(t *testing.T) {
	got := genWpaConf(t, "cafe", "")
	want := "ctrl_interface=/var/run/wpa_supplicant\nnetwork={\n\tssid=\"cafe\"\n\tkey_mgmt=NONE\n}\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAProtectedNetworkGoesThroughWpaPassphrase(t *testing.T) {
	if _, err := exec.LookPath("wpa_passphrase"); err != nil {
		t.Skip("wpa_passphrase is not installed")
	}
	got := genWpaConf(t, "home", "correct horse")
	if !strings.Contains(got, "psk=") {
		t.Fatalf("no psk in %q", got)
	}
}

func TestTheConnectRequestAcceptsAnOpenNetwork(t *testing.T) {
	for _, req := range []proto.ConnectWifiReq{
		{Ssid: "cafe"},
		{Ssid: "home", Password: "12345678"},
		{Ssid: "home", Password: strings.Repeat("a", 63)},
	} {
		if err := proto.ValidateRequest(&req); err != nil {
			t.Errorf("refused %+v: %s", req, err)
		}
	}

	for _, req := range []proto.ConnectWifiReq{
		{Password: "12345678"},
		{Ssid: "home", Password: "short"},
		{Ssid: "home", Password: strings.Repeat("a", 64)},
		{Ssid: strings.Repeat("s", 33), Password: "12345678"},
	} {
		if err := proto.ValidateRequest(&req); err == nil {
			t.Errorf("accepted %+v", req)
		}
	}
}
