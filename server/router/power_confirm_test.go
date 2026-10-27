package router

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/apikey"
	"NanoKVM-Server/service/picoclaw"
)

// A press the agent asked for is confirmed by a person in the web UI: a login
// session of a role that may press power. An API key, even an
// administrator's, cannot confirm, deny or list it, and neither can a viewer.
func TestPowerConfirmationsNeedALoginSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	if _, _, err := store.Authenticate("admin", "admin"); err != nil {
		t.Fatal(err)
	}
	if err := store.Create("operator", "operator-password", authn.RoleUser); err != nil {
		t.Fatal(err)
	}
	if err := store.Create("watcher", "watcher-password", authn.RoleViewer); err != nil {
		t.Fatal(err)
	}
	originalStore := authn.DefaultStore
	authn.DefaultStore = store
	conf := config.GetInstance()
	originalAuthentication, originalSecret, originalDuration :=
		conf.Authentication, conf.JWT.SecretKey, conf.JWT.RefreshTokenDuration
	conf.Authentication, conf.JWT.SecretKey, conf.JWT.RefreshTokenDuration = "enable", "test-secret", 3600
	originalKeys := apikey.File
	apikey.File = filepath.Join(t.TempDir(), "api_keys.json")
	t.Cleanup(func() {
		apikey.File = originalKeys
		authn.DefaultStore = originalStore
		conf.Authentication, conf.JWT.SecretKey, conf.JWT.RefreshTokenDuration =
			originalAuthentication, originalSecret, originalDuration
	})

	token := func(name string) string {
		user, err := store.Get(name)
		if err != nil {
			t.Fatal(err)
		}
		jwt, err := middleware.GenerateJWT(user.Username, user.TokenVersion)
		if err != nil {
			t.Fatal(err)
		}
		return jwt
	}
	key, _, err := apikey.Create("test", "admin")
	if err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	picoclawRouter(r, &picoclaw.Service{})

	call := func(method, path, jwt, apiKey string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, nil)
		if jwt != "" {
			req.AddCookie(&http.Cookie{Name: middleware.CookieName, Value: jwt})
		}
		if apiKey != "" {
			req.Header.Set("X-API-Key", apiKey)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	routes := [][2]string{
		{http.MethodGet, "/api/agent/power/confirmations"},
		{http.MethodPost, "/api/agent/power/confirmations/some-id/confirm"},
		{http.MethodPost, "/api/agent/power/confirmations/some-id/deny"},
	}
	for _, route := range routes {
		if w := call(route[0], route[1], "", key); w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: an admin API key got %d, want refused", route[0], route[1], w.Code)
		}
		if w := call(route[0], route[1], token("watcher"), ""); w.Code != http.StatusForbidden {
			t.Errorf("%s %s: a viewer got %d, want refused", route[0], route[1], w.Code)
		}
		if w := call(route[0], route[1], "", ""); w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: nobody got %d", route[0], route[1], w.Code)
		}
		// A user and an administrator reach the handler, which knows no
		// press by that id.
		for _, name := range []string{"operator", "admin"} {
			w := call(route[0], route[1], token(name), "")
			if w.Code != http.StatusOK {
				t.Errorf("%s %s: %s got %d", route[0], route[1], name, w.Code)
				continue
			}
			if route[0] == http.MethodPost && !strings.Contains(w.Body.String(), "no such press") {
				t.Errorf("%s %s: %s got %s", route[0], route[1], name, w.Body.String())
			}
		}
	}
}
