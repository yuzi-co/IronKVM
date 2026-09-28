package netboot

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

const (
	netbootScript = "../../../kvmapp/system/init.d/S85netboot"
	usbdevScript  = "../../../kvmapp/system/init.d/S03usbdev"
)

// scriptDefault reads the default in ${VAR:-value} from a script.
func scriptDefault(t *testing.T, script, name string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.FromSlash(script))
	if err != nil {
		t.Fatalf("cannot read %s: %s", script, err)
	}
	m := regexp.MustCompile(`\$\{` + name + `:-([^}]*)\}`).FindSubmatch(body)
	if m == nil {
		t.Fatalf("%s has no default for %s", script, name)
	}
	return string(m[1])
}

// The server writes where S85netboot reads, and reads what it writes.
func TestTheScriptsAndTheServerAgreeOnThePaths(t *testing.T) {
	for _, tc := range []struct{ script, name, want string }{
		{netbootScript, "DNSMASQ", "/data/ironkvm/addons/" + AddonName + "/dnsmasq"},
		{netbootScript, "CONFDIR", "/etc/kvm/netboot"},
		{netbootScript, "RUN", "/tmp/netboot"},
		{usbdevScript, "USB_NETBOOT", "/kvmapp/system/init.d/" + Initd},
	} {
		if got := scriptDefault(t, tc.script, tc.name); got != tc.want {
			t.Errorf("%s: %s is %s, want %s", filepath.Base(tc.script), tc.name, got, tc.want)
		}
	}
	if ConfDir != "/etc/kvm/netboot" || RunDir != "/tmp/netboot" || USBScript != "/etc/init.d/S03usbdev" {
		t.Errorf("the server's paths: %s %s %s", ConfDir, RunDir, USBScript)
	}
}
