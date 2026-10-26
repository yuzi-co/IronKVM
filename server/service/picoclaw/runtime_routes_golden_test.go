package picoclaw

import (
	"bytes"
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/service/controlmode"
)

var updateRuntimeRoutesGolden = flag.Bool("update-runtime-routes", false, "rewrite testdata/runtime_routes.golden.json")

// runtimeRoutesHarness drives the runtime, model, profile and session routes
// of one service in a scratch directory, where PicoClaw is not installed.
type runtimeRoutesHarness struct {
	t       *testing.T
	tmp     string
	service *Service
	control *controlmode.Manager
	steps   []runtimeRouteStep
}

type runtimeRouteStep struct {
	Name   string `json:"name"`
	Status int    `json:"http_status"`
	Body   any    `json:"body"`
}

func newRuntimeRoutesHarness(t *testing.T) *runtimeRoutesHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PICOCLAW_HOME", home)
	savedLicenses, savedFetch := picoclawLicensesFallback, picoclawFetchFallback
	t.Cleanup(func() {
		picoclawLicensesFallback, picoclawFetchFallback = savedLicenses, savedFetch
	})
	picoclawLicensesFallback = filepath.Join(tmp, "licenses")
	picoclawFetchFallback = filepath.Join(tmp, "fetch")

	control := controlmode.NewManager(filepath.Join(tmp, "mode"), controlmode.ModeOff)
	return &runtimeRoutesHarness{
		t:       t,
		tmp:     tmp,
		control: control,
		service: newRuntimeRoutesService(control, filepath.Join(tmp, "intent.json")),
	}
}

func (h *runtimeRoutesHarness) do(name, method string, handler gin.HandlerFunc, body string, header map[string]string) {
	h.t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	var reader *strings.Reader
	if body != "" {
		reader = strings.NewReader(body)
	} else {
		reader = strings.NewReader("")
	}
	c.Request = httptest.NewRequest(method, "/api/picoclaw/test", reader)
	if body != "" {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	for key, value := range header {
		c.Request.Header.Set(key, value)
	}
	handler(c)
	h.record(name, recorder.Code, recorder.Body.Bytes())
}

func (h *runtimeRoutesHarness) record(name string, status int, raw []byte) {
	h.t.Helper()
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		h.t.Fatalf("%s: answer is not JSON: %v: %s", name, err, raw)
	}
	h.steps = append(h.steps, runtimeRouteStep{
		Name:   name,
		Status: status,
		Body:   normalizeRuntimeRouteJSON(decoded, h.tmp),
	})
}

// normalizeRuntimeRouteJSON drops the times and the scratch directory, which
// change from run to run.
func normalizeRuntimeRouteJSON(value any, tmp string) any {
	switch v := value.(type) {
	case map[string]any:
		for key := range v {
			switch key {
			case "checked_at", "changed_at", "updated_at", "last_started_at", "last_stopped_at":
				v[key] = "<time>"
				continue
			}
			v[key] = normalizeRuntimeRouteJSON(v[key], tmp)
		}
		return v
	case []any:
		for i := range v {
			v[i] = normalizeRuntimeRouteJSON(v[i], tmp)
		}
		return v
	case string:
		return strings.ReplaceAll(v, tmp, "<tmp>")
	default:
		return v
	}
}

// TestRuntimeRoutesAnswerAsRecorded pins the JSON of the runtime routes, so a
// change to how the lifecycle is organized shows up as a diff of the golden
// file. Run with -update-runtime-routes to rewrite it.
func TestRuntimeRoutesAnswerAsRecorded(t *testing.T) {
	h := newRuntimeRoutesHarness(t)
	routes := runtimeRoutesUnderTest(h.service)

	h.do("status_initial", http.MethodGet, routes.status, "", nil)
	h.do("start_without_control", http.MethodPost, routes.start, "", nil)

	if err := h.control.Switch(controlmode.ModePicoclaw, nil); err != nil {
		t.Fatal(err)
	}
	h.do("status_with_control", http.MethodGet, routes.status, "", nil)
	h.do("start_not_installed", http.MethodPost, routes.start, "", nil)
	h.do("status_after_failed_start", http.MethodGet, routes.status, "", nil)

	h.do("model_config_bad_payload", http.MethodPost, routes.modelConfig, "{", nil)
	h.do("model_config_bad_model", http.MethodPost, routes.modelConfig,
		`{"model":"nomodel","api_base":"https://api.example.invalid/v1","api_key":"k"}`, nil)
	h.do("model_config_ok", http.MethodPost, routes.modelConfig,
		`{"model":"openai/gpt-test","api_base":"https://api.example.invalid/v1","api_key":"k"}`, nil)
	h.do("status_after_model_config", http.MethodGet, routes.status, "", nil)

	h.do("agent_profile_bad", http.MethodPost, routes.agentProfile, `{"profile":"nope"}`, nil)
	h.do("agent_profile_kvm", http.MethodPost, routes.agentProfile, `{"profile":"kvm"}`, nil)

	h.do("release_session_no_header", http.MethodDelete, routes.releaseSession, "", nil)
	h.do("release_session", http.MethodDelete, routes.releaseSession, "", map[string]string{sessionIDHeader: "s1"})

	h.do("stop", http.MethodPost, routes.stop, "", nil)
	h.do("status_after_stop", http.MethodGet, routes.status, "", nil)
	h.do("start_after_stop", http.MethodPost, routes.start, "", nil)

	if err := h.control.Switch(controlmode.ModePicoclaw, nil); err != nil {
		t.Fatal(err)
	}
	h.do("uninstall", http.MethodPost, routes.uninstall, "", nil)

	raw, err := json.Marshal(h.service.RuntimeStatus())
	if err != nil {
		t.Fatal(err)
	}
	h.record("runtime_status_after_uninstall", http.StatusOK, raw)

	if err := h.service.StopRuntimeForMCP(); err != nil {
		t.Fatal(err)
	}
	h.service.PreserveRuntimeForChatOnly("test")
	h.do("status_after_mcp_stop", http.MethodGet, routes.status, "", nil)

	got, err := json.MarshalIndent(h.steps, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')
	path := filepath.Join("testdata", "runtime_routes.golden.json")
	if *updateRuntimeRoutesGolden {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run with -update-runtime-routes to write it)", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("runtime route answers changed; diff testdata/runtime_routes.golden.json against:\n%s", got)
	}
}
