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
)

// Linking an account to an identity provider account, and the list of
// refused sign-ins to pick one from, answer an administrator's session only
// (ironkvm-dist#169): not a user, not nobody, and not an administrator's API
// key.
func TestOIDCPreLinkRoutesNeedAnAdministratorSession(t *testing.T) {
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

	r := gin.New()
	authRouter(r)

	const body = `{"issuer":"https://idp.example","subject":"sub-1"}`
	call := func(method, path, jwt, key string) int {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if jwt != "" {
			req.AddCookie(&http.Cookie{Name: middleware.CookieName, Value: jwt})
		}
		if key != "" {
			req.Header.Set("X-API-Key", key)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	key, _, err := apikey.Create("test", "admin")
	if err != nil {
		t.Fatal(err)
	}
	routes := [][2]string{
		{http.MethodGet, "/api/auth/oidc/pending"},
		{http.MethodPut, "/api/auth/users/operator/oidc"},
	}
	operator := token("operator")
	for _, route := range routes {
		if code := call(route[0], route[1], operator, ""); code != http.StatusForbidden {
			t.Errorf("%s %s: a user got %d", route[0], route[1], code)
		}
		if code := call(route[0], route[1], "", ""); code != http.StatusUnauthorized {
			t.Errorf("%s %s: nobody got %d", route[0], route[1], code)
		}
		if code := call(route[0], route[1], "", key); code < 400 {
			t.Errorf("%s %s: an admin key got %d, want it refused", route[0], route[1], code)
		}
	}
	if user, _ := store.Get("operator"); user.OIDCLinked() {
		t.Fatal("a refused request linked the account")
	}
}
