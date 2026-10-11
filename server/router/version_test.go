package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"NanoKVM-Server/common/apicontract"

	"github.com/gin-gonic/gin"
)

func TestVersionAnswersTheContractWithoutAuthentication(t *testing.T) {
	installedVersion = func() string { return "2.4.3" }
	gin.SetMode(gin.TestMode)
	r := gin.New()
	versionRouter(r)

	// No cookie, no key.
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/version", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}

	var body struct {
		Code int                        `json:"code"`
		Data map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != 0 {
		t.Fatalf("code %d", body.Code)
	}

	// Exactly these fields: anything added here is told to anyone who asks,
	// so it has to be added on purpose, with this test.
	keys := make([]string, 0, len(body.Data))
	for k := range body.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) != 3 || keys[0] != "api" || keys[1] != "features" || keys[2] != "version" {
		t.Fatalf("unexpected fields %v", keys)
	}

	var rsp VersionRsp
	_ = json.Unmarshal(w.Body.Bytes(), &struct{ Data *VersionRsp }{&rsp})
	if rsp.Version != "2.4.3" || rsp.API != apicontract.APIVersion {
		t.Fatalf("answer %+v", rsp)
	}
	if !sort.StringsAreSorted(rsp.Features) || len(rsp.Features) == 0 {
		t.Fatalf("features %v", rsp.Features)
	}
}

func TestSwitchingStaticHandlerFollowsTheStore(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	for dir, body := range map[string]string{a: "ui a", b: "ui b"} {
		if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	current := a
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(switchingStaticHandler(func() string { return current }))

	if got := get(t, r, "/").Body.String(); got != "ui a" {
		t.Fatalf("got %q", got)
	}
	current = b
	if got := get(t, r, "/").Body.String(); got != "ui b" {
		t.Fatalf("after the switch got %q", got)
	}
}

func TestProbeWebDir(t *testing.T) {
	dir := t.TempDir()
	if err := probeWebDir(dir); err == nil {
		t.Fatal("a directory without index.html passed the probe")
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html></html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := probeWebDir(dir); err != nil {
		t.Fatalf("probe: %v", err)
	}
}
