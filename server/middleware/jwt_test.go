package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

func TestCheckTokenUsesLiveRoleAndVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	admin, ok, err := store.Authenticate("admin", "admin")
	if err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	if err = store.Create("alice", "valid-password", authn.RoleUser); err != nil {
		t.Fatal(err)
	}
	alice, err := store.Get("alice")
	if err != nil {
		t.Fatal(err)
	}

	restore := useTestAuthStore(t, store)
	defer restore()
	adminToken, err := GenerateJWT(admin.Username, admin.TokenVersion)
	if err != nil {
		t.Fatal(err)
	}
	userToken, err := GenerateJWT(alice.Username, alice.TokenVersion)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.GET("/admin", CheckToken(), RequireRole(authn.RoleAdmin), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	if status := requestWithToken(router, "/admin", ""); status != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d", status)
	}
	if status := requestWithToken(router, "/admin", adminToken); status != http.StatusNoContent {
		t.Fatalf("admin status = %d", status)
	}
	if status := requestWithToken(router, "/admin", userToken); status != http.StatusForbidden {
		t.Fatalf("user status = %d", status)
	}
	if _, err = store.Revoke("admin"); err != nil {
		t.Fatal(err)
	}
	if status := requestWithToken(router, "/admin", adminToken); status != http.StatusUnauthorized {
		t.Fatalf("revoked status = %d", status)
	}
}

func TestParseJWTRejectsOtherHMACMethods(t *testing.T) {
	conf := config.GetInstance()
	originalSecret := conf.JWT.SecretKey
	conf.JWT.SecretKey = "test-secret"
	defer func() { conf.JWT.SecretKey = originalSecret }()

	now := time.Now()
	claims := Token{
		Username:     "admin",
		TokenVersion: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "admin",
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS384, claims).SignedString([]byte(conf.JWT.SecretKey))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ParseJWT(token); err == nil {
		t.Fatal("HS384 token was accepted")
	}
}

func TestParseJWTRejectsExpiredToken(t *testing.T) {
	conf := config.GetInstance()
	originalSecret := conf.JWT.SecretKey
	conf.JWT.SecretKey = "test-secret"
	defer func() { conf.JWT.SecretKey = originalSecret }()

	claims := Token{
		Username:     "admin",
		TokenVersion: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "admin",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(conf.JWT.SecretKey))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ParseJWT(token); err == nil {
		t.Fatal("expired token was accepted")
	}
}

func TestRevokeUserSessionsCancelsActiveRequests(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	unregister := activeSessions.register("alice", cancel)
	defer unregister()

	RevokeUserSessions("alice")
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("active session was not cancelled")
	}
}

func TestAccountFileResetCancelsActiveSession(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pwd")
	store := authn.NewStore(path)
	user, ok, err := store.Authenticate("admin", "admin")
	if err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	restore := useTestAuthStore(t, store)
	defer restore()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go watchSessionState(ctx, cancel, user.Username, user.TokenVersion, 10*time.Millisecond)
	if err = os.Remove(path); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("active session survived account-file reset")
	}
}

func TestWatchWebSocketClosesRevokedConnection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		connection, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return
		}
		defer connection.Close()
		stop := WatchWebSocket(ctx, connection)
		defer stop()
		_, _, _ = connection.ReadMessage()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[len("http"):]
	connection, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	cancel()
	_ = connection.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, err = connection.ReadMessage()
	if !websocket.IsCloseError(err, SessionRevokedCloseCode) {
		t.Fatalf("close error = %v, want code %d", err, SessionRevokedCloseCode)
	}
}

// The origin rule this fork applies is its own: it guards every authenticated
// request, not only the websocket upgrades, because the session travels in a
// cookie and a page the operator has open can POST to any endpoint without an
// upgrade. It compares the hostname alone, because the device serves the same
// UI over http and https, and it stands down when authentication is switched
// off. See middleware/origin.go.
func TestCheckTokenAcceptsValidTokenFromSameOrigin(t *testing.T) {
	if status := requestWithOriginHeader(t, "https://nanokvm.local"); status != http.StatusNoContent {
		t.Fatalf("same-origin request with a valid token = %d, want %d", status, http.StatusNoContent)
	}
}

func TestCheckTokenRejectsValidTokenFromForeignOrigin(t *testing.T) {
	// The cookie is valid; only the Origin differs. This is the CSRF and
	// cross-site websocket hijacking case.
	if status := requestWithOriginHeader(t, "https://evil.example.com"); status != http.StatusForbidden {
		t.Fatalf("cross-site request = %d, want %d", status, http.StatusForbidden)
	}
}

func TestCheckTokenAcceptsValidTokenWithoutOrigin(t *testing.T) {
	// A non-browser client sends no Origin, and cannot be driven by a web page.
	if status := requestWithOriginHeader(t, ""); status != http.StatusNoContent {
		t.Fatalf("request without an origin = %d, want %d", status, http.StatusNoContent)
	}
}

func requestWithOriginHeader(t *testing.T, origin string) int {
	t.Helper()
	gin.SetMode(gin.TestMode)

	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	admin, ok, err := store.Authenticate("admin", "admin")
	if err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	restore := useTestAuthStore(t, store)
	defer restore()

	token, err := GenerateJWT(admin.Username, admin.TokenVersion)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.POST("/api/vm/system/reboot", CheckToken(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodPost, "/api/vm/system/reboot", nil)
	request.Host = "nanokvm.local"
	request.AddCookie(&http.Cookie{Name: CookieName, Value: token})
	if origin != "" {
		request.Header.Set("Origin", origin)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder.Code
}

func requestWithToken(handler http.Handler, path, token string) int {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.AddCookie(&http.Cookie{Name: CookieName, Value: token})
	handler.ServeHTTP(recorder, request)
	return recorder.Code
}

func useTestAuthStore(t *testing.T, store *authn.Store) func() {
	t.Helper()
	originalStore := authn.DefaultStore
	conf := config.GetInstance()
	originalAuthentication := conf.Authentication
	originalSecret := conf.JWT.SecretKey
	originalDuration := conf.JWT.RefreshTokenDuration
	authn.DefaultStore = store
	conf.Authentication = "enable"
	conf.JWT.SecretKey = "test-secret"
	conf.JWT.RefreshTokenDuration = 3600
	return func() {
		authn.DefaultStore = originalStore
		conf.Authentication = originalAuthentication
		conf.JWT.SecretKey = originalSecret
		conf.JWT.RefreshTokenDuration = originalDuration
	}
}
