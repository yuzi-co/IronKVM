package router

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"NanoKVM-Server/config"
	"NanoKVM-Server/service/ipmi"
	"NanoKVM-Server/service/vm"
)

func TestIPMIButtonNamesAreTheOnesVMPresses(t *testing.T) {
	if ipmi.ButtonPower != vm.ButtonPower || ipmi.ButtonReset != vm.ButtonReset {
		t.Fatalf("ipmi presses %q and %q, vm knows %q and %q",
			ipmi.ButtonPower, ipmi.ButtonReset, vm.ButtonPower, vm.ButtonReset)
	}
}

// The GUID goes out the way SMBIOS writes one: the first three fields of
// the UUID byte-swapped, the rest as they are.
func TestTheIPMIGUIDIsTheRedfishUUIDInSMBIOSOrder(t *testing.T) {
	got := ipmiGUID("00112233-4455-6677-8899-aabbccddeeff")
	want := [16]byte{0x33, 0x22, 0x11, 0x00, 0x55, 0x44, 0x77, 0x66, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	if got != want {
		t.Fatalf("GUID %x, want %x", got, want)
	}
}

// The page's routes need a web UI login, and registering them with the
// service off opens no socket.
func TestTheIPMISettingsNeedAWebLogin(t *testing.T) {
	conf := config.GetInstance()
	original := conf.IPMI
	conf.IPMI.Enabled = false
	t.Cleanup(func() { conf.IPMI = original })

	originalUUID, originalKey := redfishUUIDFile, ipmiKeyFile
	redfishUUIDFile = filepath.Join(t.TempDir(), "redfish-uuid")
	ipmiKeyFile = filepath.Join(t.TempDir(), "ipmi.key")
	t.Cleanup(func() { redfishUUIDFile, ipmiKeyFile = originalUUID, originalKey })

	r := newTestEngine(t, t.TempDir())
	ipmiRouter(r)

	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/api/ipmi/settings"},
		{http.MethodPost, "/api/ipmi/settings"},
		{http.MethodPost, "/api/ipmi/users/admin/password"},
		{http.MethodDelete, "/api/ipmi/users/admin/password"},
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(route.method, route.path, nil))
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without a login answered %d %s", route.method, route.path, w.Code, w.Body.String())
		}
	}
}
