package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"

	"github.com/gin-gonic/gin"
)

// licensesFixture serves a manifest directory with sessions for an admin and
// a viewer, and a secret beside the directory.
type licensesFixture struct {
	engine *gin.Engine
	admin  string
	viewer string
}

func newLicensesFixture(t *testing.T, dir string) licensesFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)

	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	if _, _, err := store.Authenticate("admin", "admin"); err != nil {
		t.Fatal(err)
	}
	if err := store.Create("guest", "guest-password", authn.RoleViewer); err != nil {
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
	registerLicenses(r, dir)
	return licensesFixture{engine: r, admin: token("admin"), viewer: token("guest")}
}

func (f licensesFixture) get(target, token string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, target, nil)
	if token != "" {
		request.AddCookie(&http.Cookie{Name: middleware.CookieName, Value: token})
	}
	recorder := httptest.NewRecorder()
	f.engine.ServeHTTP(recorder, request)
	return recorder
}

func writeLicenseTree(t *testing.T) (dir, secret string) {
	t.Helper()
	base := t.TempDir()
	dir = filepath.Join(base, "ironkvm")
	if err := os.MkdirAll(filepath.Join(dir, "texts"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"NOTICE":     "notices\n",
		"COMPONENTS": "components\n",
		"MANIFEST":   "manifest\n",
		"texts/MIT":  "MIT License\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(name)), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	secret = filepath.Join(base, "secret")
	if err := os.WriteFile(secret, []byte("do not serve\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir, secret
}

func TestLicensesListsTheManifest(t *testing.T) {
	dir, _ := writeLicenseTree(t)
	f := newLicensesFixture(t, dir)

	w := f.get("/api/licenses", f.admin)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	var body struct {
		Code int `json:"code"`
		Data struct {
			Files []struct {
				Path string `json:"path"`
				Size int64  `json:"size"`
			} `json:"files"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, file := range body.Data.Files {
		paths = append(paths, file.Path)
	}
	if body.Code != 0 || strings.Join(paths, " ") != "COMPONENTS MANIFEST NOTICE texts/MIT" {
		t.Fatalf("code %d, files %v", body.Code, paths)
	}
}

func TestLicensesServesAFileAsPlainText(t *testing.T) {
	dir, _ := writeLicenseTree(t)
	f := newLicensesFixture(t, dir)

	w := f.get("/api/licenses/file?path=texts%2FMIT", f.admin)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	if got := w.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Fatalf("Content-Type %q", got)
	}
	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options %q", got)
	}
	if w.Body.String() != "MIT License\n" {
		t.Fatalf("body %q", w.Body)
	}
}

func TestLicensesRefusesTraversal(t *testing.T) {
	dir, secret := writeLicenseTree(t)
	if err := os.Symlink(secret, filepath.Join(dir, "texts", "leak")); err != nil {
		t.Fatal(err)
	}
	f := newLicensesFixture(t, dir)

	for _, query := range []string{
		"path=../secret",
		"path=%2e%2e%2fsecret",
		"path=%2E%2E/secret",
		"path=texts/..%2f..%2fsecret",
		"path=" + secret,
		"path=%2fetc%2fpasswd",
		`path=..%5csecret`,
		"path=texts/leak",
		"path=",
		"",
	} {
		w := f.get("/api/licenses/file?"+query, f.admin)
		if w.Code == http.StatusOK || strings.Contains(w.Body.String(), "do not serve") {
			t.Errorf("%s: status %d, body %q", query, w.Code, w.Body)
		}
	}
}

func TestLicensesWithoutTheDirectoryAreNotFound(t *testing.T) {
	f := newLicensesFixture(t, filepath.Join(t.TempDir(), "absent"))

	for _, target := range []string{"/api/licenses", "/api/licenses/file?path=NOTICE"} {
		w := f.get(target, f.admin)
		if w.Code != http.StatusNotFound {
			t.Fatalf("%s: status %d", target, w.Code)
		}
		if !strings.Contains(w.Body.String(), "not installed") {
			t.Fatalf("%s: body %q", target, w.Body)
		}
	}
}

func TestLicensesNeedASessionAndAViewerHasOne(t *testing.T) {
	dir, _ := writeLicenseTree(t)
	f := newLicensesFixture(t, dir)

	for _, target := range []string{"/api/licenses", "/api/licenses/file?path=NOTICE"} {
		if w := f.get(target, ""); w.Code != http.StatusUnauthorized {
			t.Errorf("%s without a session: status %d", target, w.Code)
		}
		if w := f.get(target, f.viewer); w.Code != http.StatusOK {
			t.Errorf("%s as a viewer: status %d", target, w.Code)
		}
	}
}
