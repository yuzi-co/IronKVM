package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/apikey"
)

// The dashcam's routes: an administrator or a user saves and fetches clips, by
// session or API key; a viewer and a view-only key are refused everything;
// deleting a clip and the settings are an administrator's.
func TestDashcamRoutesByRole(t *testing.T) {
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
	originalFile, originalDir := dashcamFile, dashcamDir
	dashcamFile = filepath.Join(t.TempDir(), "dashcam.json")
	dashcamDir = t.TempDir()
	t.Cleanup(func() {
		apikey.File = originalKeys
		authn.DefaultStore = originalStore
		conf.Authentication, conf.JWT.SecretKey, conf.JWT.RefreshTokenDuration =
			originalAuthentication, originalSecret, originalDuration
		dashcamFile, dashcamDir = originalFile, originalDir
	})

	clip := "20261027-120000-manual.mp4"
	if err := os.WriteFile(filepath.Join(dashcamDir, clip), []byte("clip"), 0o644); err != nil {
		t.Fatal(err)
	}

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
	dashcamRouter(r)

	type credential struct{ cookie, key string }
	call := func(method, path string, cred credential) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(`{"enabled":false,"seconds":60,"postSeconds":10,"storageMB":200}`))
		req.Header.Set("Content-Type", "application/json")
		if cred.cookie != "" {
			req.AddCookie(&http.Cookie{Name: middleware.CookieName, Value: cred.cookie})
		}
		if cred.key != "" {
			req.Header.Set("X-API-Key", cred.key)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	adminKey, _, err := apikey.Create("script", "admin")
	if err != nil {
		t.Fatal(err)
	}
	viewKey, _, err := apikey.CreateWithScope("player", "admin", apikey.ScopeView)
	if err != nil {
		t.Fatal(err)
	}

	userRoutes := [][2]string{
		{http.MethodGet, "/api/dashcam"},
		{http.MethodPost, "/api/dashcam/save"},
		{http.MethodGet, "/api/dashcam/clips"},
		{http.MethodGet, "/api/dashcam/clips/" + clip},
	}
	adminRoutes := [][2]string{
		{http.MethodGet, "/api/dashcam/settings"},
		{http.MethodPost, "/api/dashcam/settings"},
		{http.MethodDelete, "/api/dashcam/clips/" + clip},
	}

	cases := []struct {
		who         string
		cred        credential
		user, admin int
	}{
		{"nobody", credential{}, http.StatusUnauthorized, http.StatusUnauthorized},
		{"a viewer", credential{cookie: token("watcher")}, http.StatusForbidden, http.StatusForbidden},
		{"a view-only key", credential{key: viewKey}, http.StatusForbidden, http.StatusForbidden},
		{"a user", credential{cookie: token("operator")}, http.StatusOK, http.StatusForbidden},
		{"an administrator's key", credential{key: adminKey}, http.StatusOK, http.StatusOK},
		{"an administrator", credential{cookie: token("admin")}, http.StatusOK, http.StatusOK},
	}
	for _, tc := range cases {
		for _, route := range userRoutes {
			if w := call(route[0], route[1], tc.cred); w.Code != tc.user {
				t.Errorf("%s: %s %s got %d, want %d", tc.who, route[0], route[1], w.Code, tc.user)
			}
		}
		for _, route := range adminRoutes {
			if route[0] == http.MethodDelete && tc.admin == http.StatusOK {
				continue
			}
			if w := call(route[0], route[1], tc.cred); w.Code != tc.admin {
				t.Errorf("%s: %s %s got %d, want %d", tc.who, route[0], route[1], w.Code, tc.admin)
			}
		}
	}

	// The clip downloads as an attachment under its own name.
	w := call(http.MethodGet, "/api/dashcam/clips/"+clip, credential{cookie: token("operator")})
	if w.Body.String() != "clip" || !strings.Contains(w.Header().Get("Content-Disposition"), clip) {
		t.Fatalf("download: %q, %q", w.Body.String(), w.Header().Get("Content-Disposition"))
	}
	// A name that is not a clip's is refused, even for a file that is there,
	// and a path does not reach the handler at all.
	if err := os.WriteFile(filepath.Join(dashcamDir, "notes.mp4"), []byte("notes"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"notes.mp4", "..%2Fdashcam.json"} {
		w = call(http.MethodGet, "/api/dashcam/clips/"+name, credential{cookie: token("admin")})
		if w.Code == http.StatusOK && !strings.Contains(w.Body.String(), "no such clip") {
			t.Fatalf("%s got %d %s", name, w.Code, w.Body.String())
		}
	}
	// The dashcam is off, so a save is refused as such.
	w = call(http.MethodPost, "/api/dashcam/save", credential{cookie: token("operator")})
	if !strings.Contains(w.Body.String(), "the dashcam is off") {
		t.Fatalf("save with the dashcam off: %s", w.Body.String())
	}

	w = call(http.MethodDelete, "/api/dashcam/clips/"+clip, credential{cookie: token("admin")})
	if !strings.Contains(w.Body.String(), `"code":0`) {
		t.Fatalf("delete: %s", w.Body.String())
	}
	if _, err := os.Stat(filepath.Join(dashcamDir, clip)); !os.IsNotExist(err) {
		t.Fatalf("the clip is still there: %v", err)
	}
}
