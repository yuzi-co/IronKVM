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
)

// TestMain keeps every test in this package away from the board's session
// file, including those that issue a token without the helper.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "middleware-sessions-")
	if err != nil {
		panic(err)
	}
	authn.DefaultSessions = authn.NewSessionStore(filepath.Join(dir, "login_sessions.json"))
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

func sessionIDOf(t *testing.T, token string) string {
	t.Helper()
	parsed, err := ParseJWT(token)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Session == "" {
		t.Fatal("token carries no session")
	}
	return authn.SessionID(parsed.Session)
}

func newSessionTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/whoami", CheckToken(), func(c *gin.Context) {
		principal, _ := CurrentPrincipal(c)
		c.String(http.StatusOK, principal.SessionID)
	})
	return router
}

// A revoked login session's token is refused on the next request, while the
// same user's other sessions go on.
func TestRevokedLoginSessionIsRefused(t *testing.T) {
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	admin, ok, err := store.Authenticate("admin", "admin")
	if err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	defer useTestAuthStore(t, store)()

	first, err := GenerateJWT(admin.Username, admin.TokenVersion)
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateJWT(admin.Username, admin.TokenVersion)
	if err != nil {
		t.Fatal(err)
	}
	router := newSessionTestRouter()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/whoami", nil)
	request.AddCookie(&http.Cookie{Name: CookieName, Value: first})
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || recorder.Body.String() != sessionIDOf(t, first) {
		t.Fatalf("principal session = %d %q", recorder.Code, recorder.Body.String())
	}

	if err = RevokeLoginSession(sessionIDOf(t, first)); err != nil {
		t.Fatal(err)
	}
	if status := requestWithToken(router, "/whoami", first); status != http.StatusUnauthorized {
		t.Fatalf("revoked session status = %d", status)
	}
	if status := requestWithToken(router, "/whoami", second); status != http.StatusOK {
		t.Fatalf("other session status = %d", status)
	}
}

// A token signed with the right key but naming no session, such as one issued
// before sessions were recorded, cannot be listed or revoked, so it is not
// accepted.
func TestTokenWithoutLoginSessionIsRefused(t *testing.T) {
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	admin, ok, err := store.Authenticate("admin", "admin")
	if err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	defer useTestAuthStore(t, store)()

	now := time.Now()
	claims := Token{
		Username:     admin.Username,
		TokenVersion: admin.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   admin.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.GetInstance().JWT.SecretKey))
	if err != nil {
		t.Fatal(err)
	}
	if status := requestWithToken(newSessionTestRouter(), "/whoami", token); status != http.StatusUnauthorized {
		t.Fatalf("token without a session status = %d", status)
	}

	// Nor one naming a session the store never issued.
	claims.Session = "made-up"
	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.GetInstance().JWT.SecretKey))
	if err != nil {
		t.Fatal(err)
	}
	if status := requestWithToken(newSessionTestRouter(), "/whoami", token); status != http.StatusUnauthorized {
		t.Fatalf("token with an unknown session status = %d", status)
	}
}

// RevokeLoginSession ends that session's running requests, websockets
// included, and no other session's.
func TestRevokeLoginSessionEndsOnlyItsRequests(t *testing.T) {
	revokedCtx, revokedCancel := context.WithCancel(context.Background())
	defer revokedCancel()
	otherCtx, otherCancel := context.WithCancel(context.Background())
	defer otherCancel()
	defer activeSessions.register(session{username: "alice", sessionID: "s-revoked", cancel: revokedCancel})()
	defer activeSessions.register(session{username: "alice", sessionID: "s-other", cancel: otherCancel})()

	_ = RevokeLoginSession("s-revoked")
	select {
	case <-revokedCtx.Done():
	case <-time.After(time.Second):
		t.Fatal("the revoked session's request was not cancelled")
	}
	if otherCtx.Err() != nil {
		t.Fatal("another session's request was cancelled")
	}

	RevokeOtherSessions("alice", "s-other")
	if otherCtx.Err() != nil {
		t.Fatal("the kept session's request was cancelled")
	}
}

// RevokeOtherSessions keeps the caller's session and ends the rest of the
// user's, records and running requests both.
func TestRevokeOtherSessionsKeepsTheCurrentOne(t *testing.T) {
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	admin, ok, err := store.Authenticate("admin", "admin")
	if err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	defer useTestAuthStore(t, store)()

	keep, _ := GenerateJWT(admin.Username, admin.TokenVersion)
	other, _ := GenerateJWT(admin.Username, admin.TokenVersion)
	otherCtx, otherCancel := context.WithCancel(context.Background())
	defer otherCancel()
	defer activeSessions.register(session{username: admin.Username, sessionID: sessionIDOf(t, other), cancel: otherCancel})()

	RevokeOtherSessions(admin.Username, sessionIDOf(t, keep))
	router := newSessionTestRouter()
	if status := requestWithToken(router, "/whoami", keep); status != http.StatusOK {
		t.Fatalf("kept session status = %d", status)
	}
	if status := requestWithToken(router, "/whoami", other); status != http.StatusUnauthorized {
		t.Fatalf("other session status = %d", status)
	}
	if otherCtx.Err() == nil {
		t.Fatal("the other session's request was not cancelled")
	}
}

// RevokeUserSessions, which every account change already calls, drops the
// records too: the token version it follows does not cover a change that
// leaves the version alone, such as an administrator resetting a second
// factor.
func TestRevokeUserSessionsDropsTheRecords(t *testing.T) {
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	admin, ok, err := store.Authenticate("admin", "admin")
	if err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	defer useTestAuthStore(t, store)()

	token, _ := GenerateJWT(admin.Username, admin.TokenVersion)
	RevokeUserSessions(admin.Username)
	if status := requestWithToken(newSessionTestRouter(), "/whoami", token); status != http.StatusUnauthorized {
		t.Fatalf("session after RevokeUserSessions status = %d", status)
	}
}

// The session records the peer address. X-Forwarded-For and X-Real-IP are
// written by the client, and nothing here sits behind a proxy it trusts.
func TestIssueSessionRecordsThePeerNotAForwardedHeader(t *testing.T) {
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	admin, ok, err := store.Authenticate("admin", "admin")
	if err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	defer useTestAuthStore(t, store)()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	c.Request.RemoteAddr = "10.0.30.7:51234"
	c.Request.Header.Set("X-Forwarded-For", "203.0.113.9")
	c.Request.Header.Set("X-Real-IP", "203.0.113.9")
	c.Request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:131.0) Gecko/20100101 Firefox/131.0")

	token, err := IssueSession(c, admin.Username, admin.TokenVersion)
	if err != nil {
		t.Fatal(err)
	}
	recorded, ok := authn.DefaultSessions.Get(sessionIDOf(t, token))
	if !ok {
		t.Fatal("no session recorded")
	}
	if recorded.IP != "10.0.30.7" || recorded.Client != "Firefox on Linux" || recorded.Username != admin.Username {
		t.Fatalf("recorded %+v", recorded)
	}

	// A later request from a spoofing client does not move the address either.
	router := newSessionTestRouter()
	request := httptest.NewRequest(http.MethodGet, "/whoami", nil)
	request.RemoteAddr = "10.0.30.8:4000"
	request.Header.Set("X-Forwarded-For", "203.0.113.10")
	request.AddCookie(&http.Cookie{Name: CookieName, Value: token})
	router.ServeHTTP(httptest.NewRecorder(), request)
	if recorded, _ = authn.DefaultSessions.Get(sessionIDOf(t, token)); recorded.IP != "10.0.30.8" {
		t.Fatalf("last address = %q", recorded.IP)
	}
}
