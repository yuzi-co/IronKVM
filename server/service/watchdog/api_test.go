package watchdog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/config"
)

func newEngine(w *Watchdog) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/settings", w.GetSettings)
	r.POST("/settings", w.SetSettings)
	r.GET("/state", w.GetState)
	r.GET("/log", w.GetLog)
	r.GET("/log/:id/screenshot", w.GetScreenshot)
	return r
}

type envelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func call(t *testing.T, r *gin.Engine, method, path, body string) (*httptest.ResponseRecorder, envelope) {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var env envelope
	if strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
		if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
			t.Fatalf("%s %s: %s", method, path, rec.Body.String())
		}
	}
	return rec, env
}

func TestTheSettingsAreSavedAndValidated(t *testing.T) {
	h := newFakeHost()
	h.settings.Enabled = false
	r := newEngine(New(h.deps(t)))

	good := `{"enabled":true,"timeoutMinutes":7,"action":"power","cooldownMinutes":30,"maxPerHour":2,"pingHost":"10.0.0.5"}`
	if _, env := call(t, r, http.MethodPost, "/settings", good); env.Code != 0 {
		t.Fatalf("saving good settings: %+v", env)
	}

	_, env := call(t, r, http.MethodGet, "/settings", "")
	var got Settings
	if err := json.Unmarshal(env.Data, &got); err != nil {
		t.Fatal(err)
	}
	want := Settings{Enabled: true, TimeoutMinutes: 7, Action: "power", CooldownMinutes: 30, MaxPerHour: 2, PingHost: "10.0.0.5"}
	if got != want {
		t.Fatalf("got %+v", got)
	}

	for _, bad := range []string{
		`{"timeoutMinutes":7,"action":"reset","cooldownMinutes":30,"maxPerHour":2}`,
		`{"enabled":true,"timeoutMinutes":0,"action":"reset","cooldownMinutes":30,"maxPerHour":2}`,
		`{"enabled":true,"timeoutMinutes":7,"action":"nmi","cooldownMinutes":30,"maxPerHour":2}`,
		`{"enabled":true,"timeoutMinutes":7,"action":"reset","cooldownMinutes":30,"maxPerHour":99}`,
		`{"enabled":true,"timeoutMinutes":7,"action":"reset","cooldownMinutes":30,"maxPerHour":2,"pingHost":"host.lan"}`,
	} {
		if _, env := call(t, r, http.MethodPost, "/settings", bad); env.Code == 0 {
			t.Errorf("accepted %s", bad)
		}
	}
	if h.settings != want.toConfig() {
		t.Fatalf("a refused request changed the settings: %+v", h.settings)
	}
}

func TestTheLogAndItsScreenshotAreServed(t *testing.T) {
	h := newFakeHost()
	w := New(h.deps(t))
	r := newEngine(w)

	for range 31 {
		w.tick(context.Background())
		h.now = h.now.Add(SampleInterval)
	}

	_, env := call(t, r, http.MethodGet, "/log", "")
	var entries []Entry
	if err := json.Unmarshal(env.Data, &entries); err != nil || len(entries) != 1 {
		t.Fatalf("log %s, %v", env.Data, err)
	}

	rec, _ := call(t, r, http.MethodGet, "/log/"+entries[0].ID+"/screenshot", "")
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/jpeg" || rec.Body.String() != "desktop" {
		t.Fatalf("screenshot: %d %q %q", rec.Code, rec.Header().Get("Content-Type"), rec.Body.String())
	}

	if _, env := call(t, r, http.MethodGet, "/log/123/screenshot", ""); env.Code == 0 {
		t.Fatal("served a screenshot for an entry that does not exist")
	}

	_, env = call(t, r, http.MethodGet, "/state", "")
	var st State
	if err := json.Unmarshal(env.Data, &st); err != nil {
		t.Fatal(err)
	}
	if st.LastAction == nil || st.ActionsLastHour != 1 || !st.Signal {
		t.Fatalf("state %s", env.Data)
	}
}

func (s Settings) toConfig() config.Watchdog {
	return config.Watchdog{
		Enabled:         s.Enabled,
		TimeoutMinutes:  s.TimeoutMinutes,
		Action:          s.Action,
		CooldownMinutes: s.CooldownMinutes,
		MaxPerHour:      s.MaxPerHour,
		PingHost:        s.PingHost,
	}
}
