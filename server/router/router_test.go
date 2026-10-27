package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestEngine(t *testing.T, webPath string) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(staticHandler(webPath))

	return r
}

func get(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))

	return w
}

func TestStaticFilesAreStillServed(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatalf("failed to create the directory: %s", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte("web ui"), 0o644); err != nil {
		t.Fatalf("failed to write the file: %s", err)
	}

	r := newTestEngine(t, dir)

	if body := get(t, r, "/assets/app.js").Body.String(); body != "web ui" {
		t.Fatalf("expected the static file to be served, got %q", body)
	}
}

func TestAPIPathsNeverReachTheFilesystem(t *testing.T) {
	// The static middleware runs an os.Stat before routing, so every API call
	// and websocket upgrade pays a syscall against the SD card. Worse, a file
	// that happens to sit at the same path shadows the real handler.
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "api", "vm"), 0o755); err != nil {
		t.Fatalf("failed to create the directory: %s", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "api", "vm", "info"), []byte("decoy"), 0o644); err != nil {
		t.Fatalf("failed to write the file: %s", err)
	}

	r := newTestEngine(t, dir)
	r.GET("/api/vm/info", func(c *gin.Context) {
		c.String(http.StatusOK, "handler")
	})

	if body := get(t, r, "/api/vm/info").Body.String(); body != "handler" {
		t.Fatalf("expected the API handler to answer, got %q", body)
	}
}

func TestUnknownAPIPathsAreNotServedFromDisk(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "api"), 0o755); err != nil {
		t.Fatalf("failed to create the directory: %s", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "api", "secret"), []byte("leak"), 0o644); err != nil {
		t.Fatalf("failed to write the file: %s", err)
	}

	r := newTestEngine(t, dir)

	w := get(t, r, "/api/secret")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d with body %q", w.Code, w.Body.String())
	}
}

// The behaviour of the file server underneath the static handler, without the
// compressed copies static_test.go covers: the entry document at the root, the
// directory listing, and a path that is not on disk passing on to the routes.
// There is no fallback to index.html for an unknown path; the UI routes in
// the browser from "/".
func TestTheFileServerAnswersAsItAlwaysHas(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatalf("failed to create the directory: %s", err)
	}
	files := map[string]string{
		"index.html":    "<html>entry</html>",
		"assets/app.js": "web ui",
		"assets/b.css":  "body{}",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(name)), []byte(body), 0o644); err != nil {
			t.Fatalf("failed to write the file: %s", err)
		}
	}

	r := newTestEngine(t, dir)
	r.GET("/route", func(c *gin.Context) { c.String(http.StatusOK, "route") })
	r.NoRoute(func(c *gin.Context) { c.String(http.StatusNotFound, "no route") })

	w := get(t, r, "/")
	if w.Code != http.StatusOK || w.Body.String() != files["index.html"] {
		t.Fatalf("/ answered %d %q, want index.html", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Fatalf("/ has Content-Type %q, want text/html", got)
	}

	// The file server sends a request for index.html itself to the
	// directory, so the document has one address.
	w = get(t, r, "/index.html")
	if w.Code != http.StatusMovedPermanently || w.Header().Get("Location") != "./" {
		t.Fatalf("/index.html answered %d to %q, want 301 to ./", w.Code, w.Header().Get("Location"))
	}

	w = get(t, r, "/assets/app.js")
	if w.Code != http.StatusOK || w.Body.String() != files["assets/app.js"] {
		t.Fatalf("/assets/app.js answered %d %q", w.Code, w.Body.String())
	}

	// A directory without an index is listed. The listing is written
	// without a status, so it goes out with the 404 gin set for a path no
	// route matched.
	w = get(t, r, "/assets/")
	if w.Code != http.StatusNotFound || !strings.Contains(w.Body.String(), `href="app.js"`) || !strings.Contains(w.Body.String(), `href="b.css"`) {
		t.Fatalf("/assets/ answered %d %q, want a listing", w.Code, w.Body.String())
	}

	// A path that is not on disk is left to the routes, and so is one that
	// is not a route either.
	if w = get(t, r, "/route"); w.Code != http.StatusOK || w.Body.String() != "route" {
		t.Fatalf("/route answered %d %q, want the route", w.Code, w.Body.String())
	}
	if w = get(t, r, "/missing"); w.Code != http.StatusNotFound || w.Body.String() != "no route" {
		t.Fatalf("/missing answered %d %q, want the NoRoute handler", w.Code, w.Body.String())
	}
}
