package router

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

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
