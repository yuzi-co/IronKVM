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
)

// The settings export and import and the diagnostics bundle answer an
// administrator only: an account with the user role is refused like a viewer.
func TestSupportRoutesNeedAnAdministrator(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	if _, _, err := store.Authenticate("admin", "admin"); err != nil {
		t.Fatal(err)
	}
	if err := store.Create("operator", "operator-password", authn.RoleUser); err != nil {
		t.Fatal(err)
	}
	originalStore := authn.DefaultStore
	authn.DefaultStore = store
	conf := config.GetInstance()
	originalAuthentication, originalSecret, originalDuration :=
		conf.Authentication, conf.JWT.SecretKey, conf.JWT.RefreshTokenDuration
	conf.Authentication, conf.JWT.SecretKey, conf.JWT.RefreshTokenDuration = "enable", "test-secret", 3600
	t.Cleanup(func() {
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

	r := gin.New()
	supportRouter(r)

	call := func(method, path, jwt, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if jwt != "" {
			req.AddCookie(&http.Cookie{Name: middleware.CookieName, Value: jwt})
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	routes := [][2]string{
		{http.MethodGet, "/api/settings/export"},
		{http.MethodPost, "/api/settings/preview"},
		{http.MethodPost, "/api/settings/import"},
		{http.MethodGet, "/api/diagnostics/bundle"},
	}
	operator := token("operator")
	for _, route := range routes {
		if w := call(route[0], route[1], operator, `{}`); w.Code != http.StatusForbidden {
			t.Errorf("%s %s: a user got %d", route[0], route[1], w.Code)
		}
		if w := call(route[0], route[1], "", `{}`); w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: nobody got %d", route[0], route[1], w.Code)
		}
	}

	// An administrator reaches the handler. A preview of something that is
	// not a settings file is refused by the handler, not by the gate.
	w := call(http.MethodPost, "/api/settings/preview", token("admin"), `{"document":{"format":"other"}}`)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "not an IronKVM settings file") {
		t.Fatalf("admin preview: %d %s", w.Code, w.Body.String())
	}
}
