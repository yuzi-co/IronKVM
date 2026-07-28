package middleware

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/service/apikey"
)

// withAccounts gives the package an account store holding an administrator and
// an ordinary operator, and points the key file at a temporary path.
func withAccounts(t *testing.T) *authn.Store {
	t.Helper()
	gin.SetMode(gin.TestMode)

	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	if _, ok, err := store.Authenticate("admin", "admin"); err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	if err := store.Create("alice", "valid-password", authn.RoleUser); err != nil {
		t.Fatal(err)
	}

	restore := useTestAuthStore(t, store)
	t.Cleanup(restore)

	original := apikey.File
	apikey.File = filepath.Join(t.TempDir(), "api_keys.json")
	t.Cleanup(func() { apikey.File = original })

	return store
}

// issueKey returns a usable key belonging to username.
func issueKey(t *testing.T, username string) string {
	t.Helper()

	secret, _, err := apikey.Create("automation", username)
	if err != nil {
		t.Fatalf("failed to issue a key: %s", err)
	}

	return secret
}

func protectedEngine() *gin.Engine {
	r := gin.New()
	r.POST("/api/vm/system/reboot", CheckToken(), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return r
}

func keyedRequest(t *testing.T, header string, value string, origin string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/vm/system/reboot", nil)
	req.Host = "nanokvm.local"
	if header != "" {
		req.Header.Set(header, value)
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}

	w := httptest.NewRecorder()
	protectedEngine().ServeHTTP(w, req)

	return w
}

func TestAPIKeyAuthenticatesWithoutASession(t *testing.T) {
	// The point of the feature: a script holds a key, not a login session.
	withAccounts(t)
	secret := issueKey(t, "admin")

	w := keyedRequest(t, "X-API-Key", secret, "")

	if w.Code != http.StatusOK {
		t.Fatalf("a valid api key should be accepted, got %d", w.Code)
	}
}

func TestAPIKeyIsAcceptedAsABearerToken(t *testing.T) {
	withAccounts(t)
	secret := issueKey(t, "admin")

	w := keyedRequest(t, "Authorization", "Bearer "+secret, "")

	if w.Code != http.StatusOK {
		t.Fatalf("a bearer api key should be accepted, got %d", w.Code)
	}
}

func TestWrongAPIKeyIsRejected(t *testing.T) {
	withAccounts(t)
	issueKey(t, "admin")

	for _, value := range []string{"nkvm_not-a-real-key", "", "Bearer nkvm_nope"} {
		w := keyedRequest(t, "X-API-Key", value, "")

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("api key %q should be rejected, got %d", value, w.Code)
		}
	}
}

func TestAPIKeyCarriesOnlyItsOwnersAuthority(t *testing.T) {
	// The device has more than one kind of account now. A key issued by an
	// operator who is not an administrator must not reach the routes that ask
	// for one, or minting a key would be a way around the role.
	withAccounts(t)

	r := gin.New()
	r.POST("/api/auth/users", CheckToken(), RequireRole(authn.RoleAdmin), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	request := func(secret string) int {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/users", nil)
		req.Host = "nanokvm.local"
		req.Header.Set("X-API-Key", secret)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	if code := request(issueKey(t, "alice")); code != http.StatusForbidden {
		t.Fatalf("an operator's key reached an admin route, got %d", code)
	}

	if code := request(issueKey(t, "admin")); code != http.StatusOK {
		t.Fatalf("an administrator's key should reach an admin route, got %d", code)
	}
}

func TestAPIKeyDiesWithItsAccount(t *testing.T) {
	// A key is not an authority of its own. Disabling the account it belongs
	// to has to take the key with it, or shutting an operator out would leave
	// whatever they issued still working.
	store := withAccounts(t)
	secret := issueKey(t, "alice")

	if w := keyedRequest(t, "X-API-Key", secret, ""); w.Code != http.StatusOK {
		t.Fatalf("setup: the key should work first, got %d", w.Code)
	}

	disabled := false
	if _, err := store.Update("admin", "alice", authn.UserPatch{Enabled: &disabled}); err != nil {
		t.Fatal(err)
	}

	if w := keyedRequest(t, "X-API-Key", secret, ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("a disabled account's key was accepted, got %d", w.Code)
	}

	if err := store.Delete("admin", "alice"); err != nil {
		t.Fatal(err)
	}

	if w := keyedRequest(t, "X-API-Key", secret, ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("a deleted account's key was accepted, got %d", w.Code)
	}
}

func TestAPIKeyCannotReachSessionOnlyRoutes(t *testing.T) {
	// Managing keys is session work. A stolen key already has the run of the
	// API, but it must not be able to quietly mint more keys that survive the
	// password change made to shut it out.
	withAccounts(t)
	secret := issueKey(t, "admin")

	r := gin.New()
	r.POST("/api/auth/api-keys", CheckSession(), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/api-keys", nil)
	req.Host = "nanokvm.local"
	req.Header.Set("X-API-Key", secret)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("an api key should not manage keys, got %d", w.Code)
	}
}

func TestSessionStillReachesSessionOnlyRoutes(t *testing.T) {
	store := withAccounts(t)
	admin, err := store.Get("admin")
	if err != nil {
		t.Fatal(err)
	}

	token, err := GenerateJWT(admin.Username, admin.TokenVersion)
	if err != nil {
		t.Fatalf("failed to generate token: %s", err)
	}

	r := gin.New()
	r.POST("/api/auth/api-keys", CheckSession(), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/api-keys", nil)
	req.Host = "nanokvm.local"
	req.AddCookie(&http.Cookie{Name: CookieName, Value: token})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("a logged in session should manage keys, got %d", w.Code)
	}
}

func TestAPIKeyDoesNotBypassTheOriginCheck(t *testing.T) {
	// A key sent from a hostile page would need that page to already know the
	// key, but the origin rule is what the rest of the API relies on and this
	// path must not become the hole in it.
	withAccounts(t)
	secret := issueKey(t, "admin")

	w := keyedRequest(t, "X-API-Key", secret, "https://evil.example")

	if w.Code != http.StatusForbidden {
		t.Fatalf("a cross-origin request should be refused, got %d", w.Code)
	}
}
