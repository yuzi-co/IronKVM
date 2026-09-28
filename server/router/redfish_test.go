package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/config"
	"NanoKVM-Server/service/redfish"
	"NanoKVM-Server/service/vm"
)

func TestRedfishButtonNamesAreTheOnesVMPresses(t *testing.T) {
	if redfish.ButtonPower != vm.ButtonPower || redfish.ButtonReset != vm.ButtonReset {
		t.Fatalf("redfish presses %q and %q, vm knows %q and %q",
			redfish.ButtonPower, redfish.ButtonReset, vm.ButtonPower, vm.ButtonReset)
	}
}

// newRedfishEngine is the server's engine as far as Redfish is concerned: the
// static handler in front, then the Redfish routes.
func newRedfishEngine(t *testing.T, webPath string) *gin.Engine {
	t.Helper()

	original := redfishUUIDFile
	redfishUUIDFile = filepath.Join(t.TempDir(), "redfish-uuid")
	t.Cleanup(func() { redfishUUIDFile = original })

	r := newTestEngine(t, webPath)
	redfishRouter(r)
	return r
}

func TestRedfishIsServedBesideTheWebUI(t *testing.T) {
	r := newRedfishEngine(t, t.TempDir())

	w := get(t, r, "/redfish")
	if w.Code != http.StatusOK {
		t.Fatalf("GET /redfish: %d %s", w.Code, w.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body["v1"] != "/redfish/v1/" {
		t.Fatalf("GET /redfish: %s", w.Body.String())
	}
}

// The static handler serves a directory's index.html, so a web/redfish/v1/
// directory would answer /redfish/v1/ in the service root's place.
func TestRedfishPathsNeverReachTheFilesystem(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "redfish", "v1"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "redfish", "v1", "index.html"), []byte("decoy"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := newRedfishEngine(t, dir)

	w := get(t, r, "/redfish/v1/")
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body["@odata.id"] != "/redfish/v1/" {
		t.Fatalf("GET /redfish/v1/ answered %d %q", w.Code, w.Body.String())
	}
}

// Only /redfish and what is below it skips the static handler.
func TestAFileThatMerelyStartsWithRedfishIsStillServed(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "redfish-help.html"), []byte("help"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := newRedfishEngine(t, dir)

	if body := get(t, r, "/redfish-help.html").Body.String(); body != "help" {
		t.Fatalf("got %q", body)
	}
}

// useRedfishSetting sets redfish.enabled for one test and records what the
// switch saves instead of writing /etc/kvm/server.yaml.
func useRedfishSetting(t *testing.T, enabled bool) *[]bool {
	t.Helper()

	conf := config.GetInstance()
	original := conf.Redfish
	conf.Redfish.Enabled = &enabled

	var saved []bool
	originalSave := saveRedfishSetting
	saveRedfishSetting = func(on bool) error {
		saved = append(saved, on)
		return nil
	}

	t.Cleanup(func() {
		conf.Redfish = original
		saveRedfishSetting = originalSave
	})
	return &saved
}

// The service reads redfish.enabled from the running configuration, and the
// switch saves it and applies it there, without a restart.
func TestRedfishFollowsTheEnabledSetting(t *testing.T) {
	saved := useRedfishSetting(t, false)
	r := newRedfishEngine(t, t.TempDir())

	if w := get(t, r, "/redfish/v1/"); w.Code != http.StatusNotFound {
		t.Fatalf("off: GET /redfish/v1/ answered %d", w.Code)
	}

	if err := setRedfishEnabled(true); err != nil {
		t.Fatal(err)
	}
	if w := get(t, r, "/redfish/v1/"); w.Code != http.StatusOK {
		t.Fatalf("on: GET /redfish/v1/ answered %d", w.Code)
	}
	if !redfishEnabled() || len(*saved) != 1 || !(*saved)[0] {
		t.Fatalf("enabled=%t saved=%v", redfishEnabled(), *saved)
	}
}

// The page's routes need a web UI login; the Redfish credentials do not
// reach them.
func TestTheRedfishSettingsNeedAWebLogin(t *testing.T) {
	useRedfishSetting(t, true)
	r := newRedfishEngine(t, t.TempDir())

	for _, path := range []string{"/api/redfish/settings", "/api/redfish/sessions"} {
		w := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.SetBasicAuth("admin", "admin")
		r.ServeHTTP(w, request)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("GET %s without a login answered %d %s", path, w.Code, w.Body.String())
		}
	}
}
