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
	"NanoKVM-Server/service/slot"
)

// Writing a slot and choosing what the board boots answer an administrator
// only: an account with the user role is refused, and nobody signed in is
// refused before the tool runs.
func TestSlotRoutesNeedAnAdministrator(t *testing.T) {
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
	// No tool, so a request that got past the gate would be answered by
	// the handler with "no slot tool", never by running anything.
	originalBin := slot.Bin
	slot.Bin = filepath.Join(t.TempDir(), "slot")
	t.Cleanup(func() {
		authn.DefaultStore = originalStore
		conf.Authentication, conf.JWT.SecretKey, conf.JWT.RefreshTokenDuration =
			originalAuthentication, originalSecret, originalDuration
		slot.Bin = originalBin
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
	slotRouter(r)

	call := func(method, path, jwt string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(`{"slot":"b"}`))
		req.Header.Set("Content-Type", "application/json")
		if jwt != "" {
			req.AddCookie(&http.Cookie{Name: middleware.CookieName, Value: jwt})
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	routes := [][2]string{
		{http.MethodGet, "/api/slot"},
		{http.MethodGet, "/api/slot/images"},
		{http.MethodGet, "/api/slot/install"},
		{http.MethodPost, "/api/slot/install"},
		{http.MethodPost, "/api/slot/try"},
		{http.MethodPost, "/api/slot/confirm"},
		{http.MethodPost, "/api/slot/revert"},
	}
	operator := token("operator")
	for _, route := range routes {
		if w := call(route[0], route[1], operator); w.Code != http.StatusForbidden {
			t.Errorf("%s %s: a user got %d", route[0], route[1], w.Code)
		}
		if w := call(route[0], route[1], ""); w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: nobody got %d", route[0], route[1], w.Code)
		}
	}

	w := call(http.MethodGet, "/api/slot", token("admin"))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "no slot tool") {
		t.Fatalf("admin status: %d %s", w.Code, w.Body.String())
	}
}
