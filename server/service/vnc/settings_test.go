package vnc

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/config"
)

type apiResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func call(t *testing.T, h *harness, method, path, body string) apiResponse {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/settings", h.srv.GetSettings)
	r.POST("/settings", h.srv.SetSettings)
	r.GET("/state", h.srv.GetState)
	r.POST("/disconnect", h.srv.EndSession)

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("%s %s answered %d", method, path, w.Code)
	}

	var rsp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil {
		t.Fatal(err)
	}
	return rsp
}

func TestSettingsAPI(t *testing.T) {
	h := newHarness(t, config.VNC{Port: 0, MaxFPS: 15})
	h.srv.deps.ReservedPorts = func() []int { return []int{80, 443} }

	rsp := call(t, h, http.MethodGet, "/settings", "")
	var settings Settings
	_ = json.Unmarshal(rsp.Data, &settings)
	if rsp.Code != 0 || !settings.Enabled || settings.PasswordSet || settings.MaxFPS != 15 {
		t.Fatalf("GET /settings = %+v %s", rsp, rsp.Data)
	}

	for _, c := range []struct {
		name, body string
	}{
		{"the web port", `{"enabled":true,"port":443,"maxFps":15,"vncAuth":false}`},
		{"a frame rate over 60", `{"enabled":true,"port":5900,"maxFps":61,"vncAuth":false}`},
		{"plain authentication without a password", `{"enabled":true,"port":5900,"maxFps":15,"vncAuth":true}`},
		{"a short password", `{"enabled":true,"port":5900,"maxFps":15,"vncAuth":true,"password":"abc"}`},
		{"no enabled field", `{"port":5900,"maxFps":15,"vncAuth":false}`},
	} {
		if rsp := call(t, h, http.MethodPost, "/settings", c.body); rsp.Code == 0 {
			t.Errorf("%s was accepted", c.name)
		}
	}
	if h.password.password != "" {
		t.Fatal("a refused request saved a password")
	}

	// Off, with plain authentication and a password.
	rsp = call(t, h, http.MethodPost, "/settings", `{"enabled":false,"port":5901,"maxFps":20,"vncAuth":true,"password":"abc12345"}`)
	if rsp.Code != 0 {
		t.Fatalf("POST /settings = %+v", rsp)
	}
	if h.getSettings() != (config.VNC{Enabled: false, Port: 5901, MaxFPS: 20, VNCAuth: true}) || h.password.password != "abc12345" {
		t.Fatalf("saved %+v and password %q", h.getSettings(), h.password.password)
	}
	if h.srv.State().Listening {
		t.Fatal("the server still listens after it was turned off")
	}

	// The password is never sent back.
	rsp = call(t, h, http.MethodGet, "/settings", "")
	if strings.Contains(string(rsp.Data), "abc12345") {
		t.Fatalf("GET /settings returned the password: %s", rsp.Data)
	}
	_ = json.Unmarshal(rsp.Data, &settings)
	if !settings.PasswordSet || !settings.VNCAuth {
		t.Fatalf("GET /settings = %s", rsp.Data)
	}

	// With a password set, plain authentication stays on without a new one.
	if rsp := call(t, h, http.MethodPost, "/settings", `{"enabled":false,"port":5901,"maxFps":20,"vncAuth":true}`); rsp.Code != 0 {
		t.Fatalf("keeping the password: %+v", rsp)
	}

	if rsp := call(t, h, http.MethodPost, "/disconnect", ""); rsp.Code == 0 {
		t.Fatal("disconnect with no session succeeded")
	}
	rsp = call(t, h, http.MethodGet, "/state", "")
	var state State
	_ = json.Unmarshal(rsp.Data, &state)
	if rsp.Code != 0 || state.Listening || state.Session != nil {
		t.Fatalf("GET /state = %s", rsp.Data)
	}
}
