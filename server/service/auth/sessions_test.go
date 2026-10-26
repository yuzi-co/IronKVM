package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/proto"

	"github.com/gin-gonic/gin"
)

// TestMain keeps every test in this package away from the board's session
// file.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "auth-sessions-")
	if err != nil {
		panic(err)
	}
	authn.DefaultSessions = authn.NewSessionStore(filepath.Join(dir, "login_sessions.json"))
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

type sessionsFixture struct {
	t      *testing.T
	store  *authn.Store
	router *gin.Engine
}

// newSessionsFixture wires the session routes as router/auth.go does, with
// admin, alice and bob accounts.
func newSessionsFixture(t *testing.T) *sessionsFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	t.Cleanup(useTestStore(store))
	original := authn.DefaultSessions
	authn.DefaultSessions = authn.NewSessionStore(filepath.Join(t.TempDir(), "login_sessions.json"))
	t.Cleanup(func() { authn.DefaultSessions = original })

	if _, ok, err := store.Authenticate("admin", "admin"); err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	for _, name := range []string{"alice", "bob"} {
		if err := store.Create(name, name+"-password", authn.RoleUser); err != nil {
			t.Fatal(err)
		}
	}

	service := NewService()
	router := gin.New()
	router.POST("/login", service.Login)
	api := router.Group("/").Use(middleware.CheckToken())
	api.GET("/account", service.GetAccount)
	api.POST("/logout", service.Logout)
	api.POST("/password", service.ChangePassword)
	session := router.Group("/").Use(middleware.CheckSession())
	session.GET("/sessions", service.ListSessions)
	session.DELETE("/sessions/:id", service.RevokeSession)
	session.POST("/sessions/revoke-others", service.RevokeOtherSessions)
	session.POST("/totp/disable", service.DisableTOTP)
	admin := router.Group("/").Use(middleware.CheckToken(), middleware.RequireRole(authn.RoleAdmin))
	admin.DELETE("/users/:username", service.DeleteUser)
	admin.POST("/users/:username/password", service.ChangeUserPassword)
	adminSession := router.Group("/").Use(middleware.CheckSession(), middleware.RequireRole(authn.RoleAdmin))
	adminSession.GET("/sessions/all", service.ListAllSessions)
	adminSession.DELETE("/users/:username/totp", service.ResetUserTOTP)

	return &sessionsFixture{t: t, store: store, router: router}
}

func (f *sessionsFixture) login(username string) *http.Cookie {
	f.t.Helper()
	password := username + "-password"
	if username == "admin" {
		password = "admin"
	}
	return loginCookie(f.t, f.router, username, password)
}

func (f *sessionsFixture) status(cookie *http.Cookie) int {
	return requestJSON(f.router, http.MethodGet, "/account", nil, cookie)
}

// call makes a request and decodes the reply's code and data.
func (f *sessionsFixture) call(method, path string, cookie *http.Cookie, body any) (int, int, json.RawMessage) {
	f.t.Helper()
	recorder := requestJSONRecorder(f.router, method, path, body, cookie)
	if recorder.Code != http.StatusOK {
		return recorder.Code, 0, nil
	}
	var reply struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &reply); err != nil {
		f.t.Fatalf("%s %s: %s", method, path, recorder.Body.String())
	}
	return recorder.Code, reply.Code, reply.Data
}

func (f *sessionsFixture) list(path string, cookie *http.Cookie) []proto.LoginSessionInfo {
	f.t.Helper()
	status, code, data := f.call(http.MethodGet, path, cookie, nil)
	if status != http.StatusOK || code != 0 {
		f.t.Fatalf("GET %s = %d/%d", path, status, code)
	}
	var rsp proto.ListLoginSessionsRsp
	if err := json.Unmarshal(data, &rsp); err != nil {
		f.t.Fatal(err)
	}
	return rsp.Sessions
}

func currentOf(t *testing.T, sessions []proto.LoginSessionInfo) proto.LoginSessionInfo {
	t.Helper()
	for _, s := range sessions {
		if s.Current {
			return s
		}
	}
	t.Fatalf("no current session in %+v", sessions)
	return proto.LoginSessionInfo{}
}

func TestUserListsAndRevokesOnlyTheirOwnSessions(t *testing.T) {
	f := newSessionsFixture(t)
	alice1 := f.login("alice")
	alice2 := f.login("alice")
	bob := f.login("bob")

	mine := f.list("/sessions", alice1)
	if len(mine) != 2 {
		t.Fatalf("alice sees %d sessions: %+v", len(mine), mine)
	}
	for _, s := range mine {
		if s.Username != "alice" {
			t.Fatalf("alice sees %s's session", s.Username)
		}
		if s.IP != "192.0.2.1" || s.CreatedAt == 0 || s.LastUsedAt == 0 || s.ExpiresAt == 0 {
			t.Fatalf("session details: %+v", s)
		}
	}
	current := currentOf(t, mine)
	alice2ID := currentOf(t, f.list("/sessions", alice2)).ID
	if current.ID == alice2ID {
		t.Fatal("both sessions are marked current")
	}
	bobID := currentOf(t, f.list("/sessions", bob)).ID

	// Another user's session reads as one that does not exist, and stays.
	if _, code, _ := f.call(http.MethodDelete, "/sessions/"+bobID, alice1, nil); code == 0 {
		t.Fatal("alice revoked bob's session")
	}
	if f.status(bob) != http.StatusOK {
		t.Fatal("bob's session ended")
	}
	if status, _, _ := f.call(http.MethodGet, "/sessions/all", alice1, nil); status != http.StatusForbidden {
		t.Fatalf("alice listing everyone = %d", status)
	}

	if _, code, _ := f.call(http.MethodDelete, "/sessions/"+alice2ID, alice1, nil); code != 0 {
		t.Fatalf("revoke own session code = %d", code)
	}
	if f.status(alice2) != http.StatusUnauthorized {
		t.Fatal("the revoked session still works")
	}
	if f.status(alice1) != http.StatusOK {
		t.Fatal("the revoking session ended")
	}
}

func TestAdminListsAndRevokesEveryonesSessions(t *testing.T) {
	f := newSessionsFixture(t)
	admin := f.login("admin")
	bob := f.login("bob")
	bobID := currentOf(t, f.list("/sessions", bob)).ID

	all := f.list("/sessions/all", admin)
	users := map[string]bool{}
	for _, s := range all {
		users[s.Username] = true
	}
	if !users["admin"] || !users["bob"] || len(all) != 2 {
		t.Fatalf("admin sees %+v", all)
	}

	if _, code, _ := f.call(http.MethodDelete, "/sessions/"+bobID, admin, nil); code != 0 {
		t.Fatalf("admin revoke code = %d", code)
	}
	if f.status(bob) != http.StatusUnauthorized {
		t.Fatal("bob's revoked session still works")
	}
}

func TestRevokeOtherSessionsKeepsTheCaller(t *testing.T) {
	f := newSessionsFixture(t)
	keep := f.login("alice")
	other := f.login("alice")
	bob := f.login("bob")

	if _, code, _ := f.call(http.MethodPost, "/sessions/revoke-others", keep, nil); code != 0 {
		t.Fatalf("revoke others code = %d", code)
	}
	if f.status(keep) != http.StatusOK || f.status(other) != http.StatusUnauthorized || f.status(bob) != http.StatusOK {
		t.Fatalf("keep=%d other=%d bob=%d", f.status(keep), f.status(other), f.status(bob))
	}
	if got := f.list("/sessions", keep); len(got) != 1 {
		t.Fatalf("sessions left: %+v", got)
	}
}

// Logging out ends this session. The account's other sessions, on other
// devices, are left alone now that each can be revoked by itself.
func TestLogoutEndsOnlyThisSession(t *testing.T) {
	f := newSessionsFixture(t)
	here := f.login("alice")
	there := f.login("alice")
	if code := requestJSON(f.router, http.MethodPost, "/logout", nil, here); code != http.StatusOK {
		t.Fatalf("logout status = %d", code)
	}
	if f.status(here) != http.StatusUnauthorized {
		t.Fatal("the logged out session still works")
	}
	if f.status(there) != http.StatusOK {
		t.Fatal("logout ended another device's session")
	}
}

func TestAccountChangesRevokeTheUsersSessions(t *testing.T) {
	t.Run("password reset by an administrator", func(t *testing.T) {
		f := newSessionsFixture(t)
		admin := f.login("admin")
		alice := f.login("alice")
		body := map[string]any{"password": encryptForRequest("another-password")}
		if _, code, _ := f.call(http.MethodPost, "/users/alice/password", admin, body); code != 0 {
			t.Fatalf("reset code = %d", code)
		}
		if f.status(alice) != http.StatusUnauthorized {
			t.Fatal("alice's session survived the reset")
		}
		for _, s := range f.list("/sessions/all", admin) {
			if s.Username == "alice" {
				t.Fatalf("alice's session is still listed: %+v", s)
			}
		}
	})

	t.Run("two-factor reset by an administrator", func(t *testing.T) {
		f := newSessionsFixture(t)
		admin := f.login("admin")
		alice := f.login("alice")
		if _, err := f.store.SetTOTP("alice", "sealed", nil); err != nil {
			t.Fatal(err)
		}
		if _, code, _ := f.call(http.MethodDelete, "/users/alice/totp", admin, nil); code != 0 {
			t.Fatalf("reset code = %d", code)
		}
		if f.status(alice) != http.StatusUnauthorized {
			t.Fatal("alice's session survived the two-factor reset")
		}
		if f.status(admin) != http.StatusOK {
			t.Fatal("the administrator's session ended")
		}
	})

	t.Run("two-factor turned off by the user", func(t *testing.T) {
		f := newSessionsFixture(t)
		here := f.login("alice")
		there := f.login("alice")
		if _, err := f.store.SetTOTP("alice", "sealed", nil); err != nil {
			t.Fatal(err)
		}
		body := map[string]any{"password": encryptForRequest("alice-password")}
		if _, code, _ := f.call(http.MethodPost, "/totp/disable", here, body); code != 0 {
			t.Fatalf("disable code = %d", code)
		}
		if f.status(here) != http.StatusOK {
			t.Fatal("the session that made the change ended")
		}
		if f.status(there) != http.StatusUnauthorized {
			t.Fatal("another session survived the two-factor change")
		}
	})

	t.Run("user deleted", func(t *testing.T) {
		f := newSessionsFixture(t)
		admin := f.login("admin")
		f.login("alice")
		if _, code, _ := f.call(http.MethodDelete, "/users/alice", admin, nil); code != 0 {
			t.Fatalf("delete code = %d", code)
		}
		for _, s := range f.list("/sessions/all", admin) {
			if s.Username == "alice" {
				t.Fatalf("a deleted user's session is listed: %+v", s)
			}
		}
	})

	t.Run("own password change", func(t *testing.T) {
		f := newSessionsFixture(t)
		here := f.login("alice")
		there := f.login("alice")
		body := map[string]any{
			"currentPassword": encryptForRequest("alice-password"),
			"password":        encryptForRequest("another-password"),
		}
		if _, code, _ := f.call(http.MethodPost, "/password", here, body); code != 0 {
			t.Fatalf("change code = %d", code)
		}
		if f.status(there) != http.StatusUnauthorized {
			t.Fatal("another session survived the password change")
		}
	})
}

// The address in the list is the peer's. A login that claims another in
// X-Forwarded-For does not get it recorded.
func TestSessionAddressIgnoresForwardedHeaders(t *testing.T) {
	f := newSessionsFixture(t)
	body := map[string]any{"username": "alice", "password": encryptForRequest("alice-password")}
	request := jsonRequest(http.MethodPost, "/login", body)
	request.Header.Set("X-Forwarded-For", "203.0.113.9")
	request.Header.Set("X-Real-IP", "203.0.113.9")
	recorder := httptest.NewRecorder()
	f.router.ServeHTTP(recorder, request)
	var cookie *http.Cookie
	for _, c := range recorder.Result().Cookies() {
		if c.Name == middleware.CookieName {
			cookie = c
		}
	}
	if cookie == nil {
		t.Fatal("no session cookie")
	}
	if got := currentOf(t, f.list("/sessions", cookie)).IP; got != "192.0.2.1" {
		t.Fatalf("recorded address %q", got)
	}
}
