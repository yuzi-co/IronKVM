package redfish

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

// mountSettings adds the web UI's /api/redfish routes to the harness engine,
// without the token check the server puts in front of them.
func (h *harness) mountSettings() {
	h.engine.GET("/api/redfish/settings", h.service.GetSettings)
	h.engine.POST("/api/redfish/settings", h.service.SetSettings)
	h.engine.GET("/api/redfish/sessions", h.service.GetSessions)
	h.engine.DELETE("/api/redfish/sessions/:id", h.service.EndSession)
}

// apiResponse is the web UI's response envelope.
type apiResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func (h *harness) api(method, path, body string) apiResponse {
	h.t.Helper()

	w := h.do(method, path, body)
	if w.Code != http.StatusOK {
		h.t.Fatalf("%s %s: status %d %s", method, path, w.Code, w.Body.String())
	}
	var rsp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil {
		h.t.Fatalf("%s %s: %s\n%s", method, path, err, w.Body.String())
	}
	return rsp
}

// Every path under /redfish, known or not, public or not, with any method,
// answers 404 while the service is off.
func TestWhenOffEveryRedfishPathIsNotFound(t *testing.T) {
	h := newHarness(t)
	token := h.admin()
	h.enabled = false

	for _, r := range []struct{ method, path, body string }{
		{http.MethodGet, "/redfish", ""},
		{http.MethodGet, "/redfish/v1", ""},
		{http.MethodGet, "/redfish/v1/", ""},
		{http.MethodHead, "/redfish/v1/", ""},
		{http.MethodOptions, "/redfish/v1/", ""},
		{http.MethodGet, "/redfish/v1/$metadata", ""},
		{http.MethodGet, "/redfish/v1/Systems/1", ""},
		{http.MethodPut, "/redfish/v1/Systems/1", "{}"},
		{http.MethodGet, "/redfish/v1/NoSuchThing", ""},
		{http.MethodPost, sessionsPath, `{"UserName":"admin","Password":"admin"}`},
		{http.MethodPost, resetPath, `{"ResetType":"ForceRestart"}`},
	} {
		w := h.do(r.method, r.path, r.body, token...)
		if w.Code != http.StatusNotFound {
			t.Errorf("%s %s: %d %s", r.method, r.path, w.Code, w.Body.String())
		}
		if w.Header().Get("OData-Version") != "" || w.Header().Get("Allow") != "" {
			t.Errorf("%s %s answered as a Redfish service would: %v", r.method, r.path, w.Header())
		}
	}

	if presses := h.host.presses; len(presses) != 0 {
		t.Fatalf("pressed %v while off", presses)
	}
	if h.limiter.failed != nil || h.limiter.succeeded != nil || h.counted.passwordChecks() != 0 {
		t.Fatal("a login was checked while off")
	}
}

// The switch reads the setting on every request, so it needs no restart.
func TestTheSwitchTakesEffectWithoutARestart(t *testing.T) {
	h := newHarness(t)

	if w := h.do(http.MethodGet, rootPath, ""); w.Code != http.StatusOK {
		t.Fatalf("on: %d", w.Code)
	}
	h.enabled = false
	if w := h.do(http.MethodGet, rootPath, ""); w.Code != http.StatusNotFound {
		t.Fatalf("off: %d", w.Code)
	}
	h.enabled = true
	if w := h.do(http.MethodGet, rootPath, ""); w.Code != http.StatusOK {
		t.Fatalf("on again: %d", w.Code)
	}
}

// A nil Enabled is the service as it was before the setting: always on.
func TestWithoutAnEnabledDependencyTheServiceIsOn(t *testing.T) {
	s := New(Deps{})
	if !s.enabled() {
		t.Fatal("a service without Deps.Enabled is off")
	}
}

func TestGetSettings(t *testing.T) {
	h := newHarness(t)
	h.mountSettings()

	var got struct {
		Enabled     bool     `json:"enabled"`
		HTTPS       bool     `json:"https"`
		ServiceRoot string   `json:"serviceRoot"`
		ResetTypes  []string `json:"resetTypes"`
	}
	read := func() {
		t.Helper()
		rsp := h.api(http.MethodGet, "/api/redfish/settings", "")
		if rsp.Code != 0 {
			t.Fatalf("answered %+v", rsp)
		}
		got.ResetTypes = nil
		if err := json.Unmarshal(rsp.Data, &got); err != nil {
			t.Fatal(err)
		}
	}

	read()
	if !got.Enabled || got.HTTPS || got.ServiceRoot != "/redfish/v1/" {
		t.Fatalf("got %+v", got)
	}
	if strings.Join(got.ResetTypes, ",") != strings.Join(resetTypes, ",") {
		t.Fatalf("reset types %v, want %v", got.ResetTypes, resetTypes)
	}

	// The list follows the power LED setting, as the system resource does.
	h.ledWired = false
	h.enabled = false
	h.https = true
	read()
	if got.Enabled || !got.HTTPS {
		t.Fatalf("got %+v", got)
	}
	if strings.Join(got.ResetTypes, ",") != strings.Join(blindResetTypes, ",") {
		t.Fatalf("reset types %v, want %v", got.ResetTypes, blindResetTypes)
	}
}

func TestSetSettingsSavesTheSwitch(t *testing.T) {
	h := newHarness(t)
	h.mountSettings()

	if rsp := h.api(http.MethodPost, "/api/redfish/settings", `{"enabled":false}`); rsp.Code != 0 {
		t.Fatalf("answered %+v", rsp)
	}
	if h.enabled || len(h.saved) != 1 || h.saved[0] {
		t.Fatalf("enabled=%t saved=%v", h.enabled, h.saved)
	}
	if w := h.do(http.MethodGet, rootPath, ""); w.Code != http.StatusNotFound {
		t.Fatalf("off: %d", w.Code)
	}

	if rsp := h.api(http.MethodPost, "/api/redfish/settings", `{"enabled":true}`); rsp.Code != 0 {
		t.Fatalf("answered %+v", rsp)
	}
	if w := h.do(http.MethodGet, rootPath, ""); w.Code != http.StatusOK {
		t.Fatalf("on: %d", w.Code)
	}
}

// A body without the switch must not turn the service off.
func TestSetSettingsNeedsTheSwitch(t *testing.T) {
	h := newHarness(t)
	h.mountSettings()

	for _, body := range []string{`{}`, `{"enabled":"no"}`, ``} {
		if rsp := h.api(http.MethodPost, "/api/redfish/settings", body); rsp.Code == 0 {
			t.Errorf("%q answered success", body)
		}
	}
	if !h.enabled || len(h.saved) != 0 {
		t.Fatalf("enabled=%t saved=%v", h.enabled, h.saved)
	}
}

// Turning the service off ends every session, so none comes back when it is
// turned on again.
func TestTurningOffEndsEverySession(t *testing.T) {
	h := newHarness(t)
	h.mountSettings()
	admin, user := h.admin(), h.user()

	h.api(http.MethodPost, "/api/redfish/settings", `{"enabled":false}`)
	h.api(http.MethodPost, "/api/redfish/settings", `{"enabled":true}`)

	for _, token := range [][]string{admin, user} {
		if w := h.do(http.MethodGet, systemPath, "", token...); w.Code != http.StatusUnauthorized {
			t.Fatalf("a token from before the switch answered %d", w.Code)
		}
	}
	if n := len(h.service.sessions.list()); n != 0 {
		t.Fatalf("%d sessions left", n)
	}
}

// Turning it on, or a save that fails, leaves the sessions alone.
func TestSessionsSurviveTurningOnAndAFailedSave(t *testing.T) {
	h := newHarness(t)
	h.mountSettings()
	token := h.admin()

	h.api(http.MethodPost, "/api/redfish/settings", `{"enabled":true}`)

	h.saveErr = errors.New("read-only file system")
	if rsp := h.api(http.MethodPost, "/api/redfish/settings", `{"enabled":false}`); rsp.Code == 0 {
		t.Fatal("a failed save answered success")
	}
	if !h.enabled {
		t.Fatal("a failed save turned the service off")
	}

	if w := h.do(http.MethodGet, systemPath, "", token...); w.Code != http.StatusOK {
		t.Fatalf("the session answered %d", w.Code)
	}
}

// A login that passed the switch just before it was turned off must not
// leave a session behind.
func TestALoginThatRacesTheSwitchLeavesNoSession(t *testing.T) {
	h := newHarness(t)
	h.counted.during = func() { h.enabled = false }

	w := h.do(http.MethodPost, sessionsPath, `{"UserName":"admin","Password":"admin"}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("login answered %d %s", w.Code, w.Body.String())
	}
	if w.Header().Get("X-Auth-Token") != "" {
		t.Fatal("a token was handed out")
	}
	if n := len(h.service.sessions.list()); n != 0 {
		t.Fatalf("%d sessions left", n)
	}
}

type sessionInfo struct {
	ID        string `json:"id"`
	User      string `json:"user"`
	CreatedAt string `json:"createdAt"`
	LastUsed  string `json:"lastUsed"`
}

func (h *harness) listSessions() ([]sessionInfo, string) {
	h.t.Helper()

	rsp := h.api(http.MethodGet, "/api/redfish/sessions", "")
	if rsp.Code != 0 {
		h.t.Fatalf("answered %+v", rsp)
	}
	var out []sessionInfo
	if err := json.Unmarshal(rsp.Data, &out); err != nil {
		h.t.Fatal(err)
	}
	return out, string(rsp.Data)
}

func TestGetSessionsListsEverySessionWithoutItsToken(t *testing.T) {
	h := newHarness(t)
	h.mountSettings()

	if got, raw := h.listSessions(); len(got) != 0 || raw != "[]" {
		t.Fatalf("no sessions listed as %s", raw)
	}

	w := h.do(http.MethodPost, sessionsPath, `{"UserName":"alice","Password":"valid-password"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("login: %d %s", w.Code, w.Body.String())
	}
	token := w.Header().Get("X-Auth-Token")
	id := strings.TrimPrefix(w.Header().Get("Location"), sessionsPath+"/")

	created := h.clock
	h.clock = h.clock.Add(time.Minute)
	if w := h.do(http.MethodGet, systemPath, "", "X-Auth-Token", token); w.Code != http.StatusOK {
		t.Fatalf("use: %d", w.Code)
	}
	h.admin()

	got, raw := h.listSessions()
	if len(got) != 2 {
		t.Fatalf("listed %s", raw)
	}
	want := sessionInfo{
		ID:        id,
		User:      "alice",
		CreatedAt: created.Format(time.RFC3339),
		LastUsed:  h.clock.Format(time.RFC3339),
	}
	if got[0] != want {
		t.Fatalf("got %+v, want %+v", got[0], want)
	}
	if got[1].User != "admin" {
		t.Fatalf("second session %+v", got[1])
	}
	if strings.Contains(raw, token) || strings.Contains(strings.ToLower(raw), "token") {
		t.Fatalf("the list carries a token: %s", raw)
	}
}

func TestEndSessionMakesItsTokenFail(t *testing.T) {
	h := newHarness(t)
	h.mountSettings()
	token, other := h.user(), h.admin()

	got, _ := h.listSessions()
	if rsp := h.api(http.MethodDelete, "/api/redfish/sessions/"+got[0].ID, ""); rsp.Code != 0 {
		t.Fatalf("answered %+v", rsp)
	}

	if w := h.do(http.MethodGet, systemPath, "", token...); w.Code != http.StatusUnauthorized {
		t.Fatalf("the ended session answered %d", w.Code)
	}
	if w := h.do(http.MethodGet, systemPath, "", other...); w.Code != http.StatusOK {
		t.Fatalf("the other session answered %d", w.Code)
	}

	if rsp := h.api(http.MethodDelete, "/api/redfish/sessions/"+got[0].ID, ""); rsp.Code == 0 {
		t.Fatal("ending it twice answered success")
	}
}
