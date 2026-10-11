package middleware

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"NanoKVM-Server/authn"

	"github.com/gin-gonic/gin"
)

func useTestShares(t *testing.T) *authn.ShareStore {
	t.Helper()
	original := authn.DefaultShares
	shares := authn.NewShareStore(filepath.Join(t.TempDir(), "share_links.json"))
	authn.DefaultShares = shares
	t.Cleanup(func() { authn.DefaultShares = original })
	return shares
}

func viewerEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := func(c *gin.Context) { c.Status(http.StatusNoContent) }
	api := router.Group("/api").Use(CheckToken())
	api.GET("/stream/h264", handler)
	api.GET("/ws", handler)
	api.POST("/hid/paste", handler)
	api.GET("/vm/info", handler)
	api.POST("/vm/gpio", handler)
	return router
}

func do(handler http.Handler, method, path, token string) int {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, nil)
	request.AddCookie(&http.Cookie{Name: CookieName, Value: token})
	handler.ServeHTTP(recorder, request)
	return recorder.Code
}

func TestViewerReachesOnlyTheListedRoutes(t *testing.T) {
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	if _, _, err := store.Authenticate("admin", "admin"); err != nil {
		t.Fatal(err)
	}
	if err := store.Create("guest", "guest-password", authn.RoleViewer); err != nil {
		t.Fatal(err)
	}
	if err := store.Create("alice", "alice-password", authn.RoleUser); err != nil {
		t.Fatal(err)
	}
	defer useTestAuthStore(t, store)()

	guest, _ := store.Get("guest")
	alice, _ := store.Get("alice")
	guestToken, err := GenerateJWT(guest.Username, guest.TokenVersion)
	if err != nil {
		t.Fatal(err)
	}
	aliceToken, err := GenerateJWT(alice.Username, alice.TokenVersion)
	if err != nil {
		t.Fatal(err)
	}
	router := viewerEngine()

	cases := []struct {
		method, path string
		viewer       int
	}{
		{http.MethodGet, "/api/stream/h264", http.StatusNoContent},
		{http.MethodGet, "/api/vm/info", http.StatusNoContent},
		{http.MethodGet, "/api/ws", http.StatusForbidden},
		{http.MethodPost, "/api/hid/paste", http.StatusForbidden},
		{http.MethodPost, "/api/vm/gpio", http.StatusForbidden},
		// A method the route list does not name is not allowed either.
		{http.MethodPost, "/api/vm/info", http.StatusNotFound},
	}
	for _, c := range cases {
		if got := do(router, c.method, c.path, guestToken); got != c.viewer {
			t.Errorf("viewer %s %s = %d, want %d", c.method, c.path, got, c.viewer)
		}
	}

	// A user is unchanged: the same routes stay open to them.
	for _, path := range []string{"/api/ws", "/api/stream/h264", "/api/vm/info"} {
		if got := do(router, http.MethodGet, path, aliceToken); got != http.StatusNoContent {
			t.Errorf("user GET %s = %d", path, got)
		}
	}
	if got := do(router, http.MethodPost, "/api/hid/paste", aliceToken); got != http.StatusNoContent {
		t.Errorf("user POST /api/hid/paste = %d", got)
	}
}

func TestShareSessionIsAViewerAndEndsWithItsLink(t *testing.T) {
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	defer useTestAuthStore(t, store)()
	shares := useTestShares(t)
	router := viewerEngine()

	link, _, err := shares.Create("", "admin", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	token, err := GenerateShareJWT(link.ID, link.ExpiresAt)
	if err != nil {
		t.Fatal(err)
	}

	if got := do(router, http.MethodGet, "/api/stream/h264", token); got != http.StatusNoContent {
		t.Fatalf("share session on the video = %d", got)
	}
	for _, c := range []struct{ method, path string }{
		{http.MethodGet, "/api/ws"},
		{http.MethodPost, "/api/hid/paste"},
		{http.MethodPost, "/api/vm/gpio"},
	} {
		if got := do(router, c.method, c.path, token); got != http.StatusForbidden {
			t.Errorf("share session %s %s = %d, want 403", c.method, c.path, got)
		}
	}

	if err = shares.Revoke(link.ID); err != nil {
		t.Fatal(err)
	}
	if got := do(router, http.MethodGet, "/api/stream/h264", token); got != http.StatusUnauthorized {
		t.Fatalf("share session after the link was revoked = %d", got)
	}
}

func TestShareSessionCannotOutliveItsLink(t *testing.T) {
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	defer useTestAuthStore(t, store)()
	shares := useTestShares(t)
	now := time.Now()
	shares.SetClock(func() time.Time { return now })

	link, _, err := shares.Create("", "admin", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	token, err := GenerateShareJWT(link.ID, link.ExpiresAt)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseJWT(token)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.ExpiresAt.After(link.ExpiresAt.Add(time.Second)) {
		t.Fatalf("the session token outlives its link: %v > %v", parsed.ExpiresAt, link.ExpiresAt)
	}

	router := viewerEngine()
	now = now.Add(2 * time.Minute)
	if got := do(router, http.MethodGet, "/api/stream/h264", token); got != http.StatusUnauthorized {
		t.Fatalf("share session after the link expired = %d", got)
	}
}

func TestRevokingAShareLinkEndsItsRunningSession(t *testing.T) {
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	defer useTestAuthStore(t, store)()
	shares := useTestShares(t)

	link, _, err := shares.Create("", "admin", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	token, err := GenerateShareJWT(link.ID, link.ExpiresAt)
	if err != nil {
		t.Fatal(err)
	}

	registry := newSessionRegistry(10 * time.Millisecond)
	original := activeSessions
	activeSessions = registry
	defer func() { activeSessions = original }()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	finished := make(chan struct{})
	started := make(chan struct{})
	router.GET("/api/stream/h264", CheckToken(), func(c *gin.Context) {
		close(started)
		<-c.Request.Context().Done()
		close(finished)
	})
	go func() {
		request := httptest.NewRequest(http.MethodGet, "/api/stream/h264", nil)
		request.AddCookie(&http.Cookie{Name: CookieName, Value: token})
		router.ServeHTTP(httptest.NewRecorder(), request)
	}()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("the session did not start")
	}

	// Revoked in the store alone, as a hand edit of the file would be: the
	// recheck has to find it.
	if err = shares.Revoke(link.ID); err != nil {
		t.Fatal(err)
	}
	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("the running session outlived its revoked link")
	}
}
