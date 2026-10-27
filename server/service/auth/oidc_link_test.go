package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/service/audit"
)

// linkCookieOf returns the pending link cookie a refused callback set.
func linkCookieOf(rec *httptest.ResponseRecorder) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == oidcLinkCookie && c.Value != "" && c.MaxAge > 0 {
			return c
		}
	}
	return nil
}

type linkReply struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Username     string          `json:"username"`
		TOTPRequired bool            `json:"totpRequired"`
		Issuer       string          `json:"issuer"`
		SignIns      []pendingSignIn `json:"signIns"`
	} `json:"data"`
}

func (f *oidcFixture) call(method, path string, body any, cookie *http.Cookie) (linkReply, *httptest.ResponseRecorder) {
	f.t.Helper()
	rec := requestJSONRecorder(f.router, method, path, body, cookie)
	var reply linkReply
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &reply); err != nil {
			f.t.Fatalf("%s %s: %s", method, path, rec.Body.String())
		}
	}
	return reply, rec
}

// confirm sends the account's password, and code, for the pending link.
func (f *oidcFixture) confirm(link *http.Cookie, password, code string) (linkReply, *httptest.ResponseRecorder) {
	f.t.Helper()
	return f.call(http.MethodPost, "/api/auth/oidc/link", map[string]string{"password": encryptForRequest(password), "code": code}, link)
}

func (f *oidcFixture) pending(admin *http.Cookie) linkReply {
	f.t.Helper()
	reply, rec := f.call(http.MethodGet, "/api/auth/oidc/pending", nil, admin)
	if rec.Code != http.StatusOK || reply.Code != 0 {
		f.t.Fatalf("pending = %d %s", rec.Code, rec.Body.String())
	}
	return reply
}

func resetPendingSignIns(t *testing.T) {
	oidcLinkMu.Lock()
	pendingSignIns = nil
	oidcLinkMu.Unlock()
	t.Cleanup(func() {
		oidcLinkMu.Lock()
		pendingSignIns = nil
		oidcLinkMu.Unlock()
	})
}

// A single sign-on to an account that has no link signs nobody in and links
// nothing. The administrator sees it on the pending list.
func TestOIDCUnlinkedFirstSignInIsRefusedAndPending(t *testing.T) {
	f := newOIDCFixture(t)
	resetPendingSignIns(t)
	f.idp.claims["preferred_username"] = "alice"
	f.idp.claims["email"] = "alice@corp.example"
	f.idp.claims["name"] = "Alice Example"

	rec := f.login()
	f.refused(rec, "unlinked")
	if e := f.lastEvent(); e.Outcome != audit.OutcomeDenied || e.Actor != "alice" || !strings.Contains(e.Detail, "subject-1") {
		t.Fatalf("refusal audited as %+v", e)
	}
	if issuer, sub := f.link("alice"); issuer != "" || sub != "" {
		t.Fatalf("a refused sign-in linked alice to %q %q", issuer, sub)
	}

	admin := loginCookie(t, f.router, "admin", "admin")
	reply := f.pending(admin)
	if reply.Data.Issuer != f.idp.server.URL || len(reply.Data.SignIns) != 1 {
		t.Fatalf("pending = %+v", reply.Data)
	}
	p := reply.Data.SignIns[0]
	if p.Username != "alice" || p.Issuer != f.idp.server.URL || p.Subject != "subject-1" ||
		p.Email != "alice@corp.example" || p.Name != "Alice Example" || time.Since(time.Unix(p.Time, 0)) > time.Minute {
		t.Fatalf("pending sign-in %+v", p)
	}

	// The administrator picks it: the account is linked, the entry goes, and
	// the next sign-in gets in.
	body := map[string]string{"issuer": p.Issuer, "subject": p.Subject}
	if reply, rec := f.call(http.MethodPut, "/api/auth/users/alice/oidc", body, admin); reply.Code != 0 {
		t.Fatalf("pre-link = %s", rec.Body.String())
	}
	if n := len(f.pending(admin).Data.SignIns); n != 0 {
		t.Fatalf("%d pending sign-ins after the link", n)
	}
	if name, _ := f.account(sessionCookieOf(f.login())); name != "alice" {
		t.Fatalf("pre-linked login signed in as %q", name)
	}
}

func TestOIDCPreLinkNeedsASubjectAndAnAccount(t *testing.T) {
	f := newOIDCFixture(t)
	admin := loginCookie(t, f.router, "admin", "admin")
	for _, tc := range []struct {
		path string
		body map[string]string
	}{
		{"/api/auth/users/alice/oidc", map[string]string{"subject": "  "}},
		{"/api/auth/users/alice/oidc", map[string]string{"subject": strings.Repeat("s", 256)}},
		{"/api/auth/users/nobody/oidc", map[string]string{"subject": "sub"}},
		// The device owner signs in with its password only.
		{"/api/auth/users/admin/oidc", map[string]string{"subject": "sub"}},
	} {
		if reply, rec := f.call(http.MethodPut, tc.path, tc.body, admin); reply.Code == 0 {
			t.Fatalf("%s %v: %s", tc.path, tc.body, rec.Body.String())
		}
	}
	if _, sub := f.link("alice"); sub != "" {
		t.Fatalf("alice linked to %q", sub)
	}

	// Relinking another's account ends that account's sessions.
	f.prelink("alice", "old-sub")
	alice := loginCookie(t, f.router, "alice", "alice-password")
	if reply, rec := f.call(http.MethodPut, "/api/auth/users/alice/oidc", map[string]string{"subject": "new-sub"}, admin); reply.Code != 0 {
		t.Fatalf("relink = %s", rec.Body.String())
	}
	if name, _ := f.account(alice); name != "" {
		t.Fatalf("alice's session survived the relink as %q", name)
	}
}

// The refused sign-in waits for the account's password, in this browser
// only, and the password links the account and signs the user in.
func TestOIDCPasswordConfirmationLinks(t *testing.T) {
	f := newOIDCFixture(t)
	resetPendingSignIns(t)
	f.save(func(s *oidcSettings) { s.UserGroup = "kvm-users"; s.ViewerGroup = "kvm-viewers" })
	f.idp.claims["preferred_username"] = "alice"
	f.idp.claims["groups"] = "kvm-viewers"

	rec := f.login()
	f.refused(rec, "unlinked")
	link := linkCookieOf(rec)
	if link == nil || !link.HttpOnly || link.Path != "/api/auth/oidc" {
		t.Fatalf("link cookie %+v", link)
	}

	if reply, _ := f.call(http.MethodGet, "/api/auth/oidc/link", nil, nil); reply.Code != codeOIDCLinkGone {
		t.Fatalf("without the cookie: %+v", reply)
	}
	reply, _ := f.call(http.MethodGet, "/api/auth/oidc/link", nil, link)
	if reply.Code != 0 || reply.Data.Username != "alice" || reply.Data.TOTPRequired {
		t.Fatalf("link info %+v", reply)
	}

	reply, rec = f.confirm(link, "alice-password", "")
	if reply.Code != 0 {
		t.Fatalf("confirm = %s", rec.Body.String())
	}
	if issuer, sub := f.link("alice"); issuer != f.idp.server.URL || sub != "subject-1" {
		t.Fatalf("alice linked to %q %q", issuer, sub)
	}
	// Signed in, with the role the groups map to.
	if name, role := f.account(sessionCookieOf(rec)); name != "alice" || role != authn.RoleViewer {
		t.Fatalf("signed in as %q/%q", name, role)
	}
	if links := f.eventsOf("oidc-link"); len(links) != 1 || links[0].Target != "alice" || !strings.Contains(links[0].Detail, "password") {
		t.Fatalf("link audited as %+v", links)
	}
	admin := loginCookie(t, f.router, "admin", "admin")
	if n := len(f.pending(admin).Data.SignIns); n != 0 {
		t.Fatalf("%d pending sign-ins after the link", n)
	}
	// The pending link is used up.
	if reply, _ := f.confirm(link, "alice-password", ""); reply.Code != codeOIDCLinkGone {
		t.Fatalf("second confirm %+v", reply)
	}
	// And the next sign-in gets straight in.
	if sessionCookieOf(f.login()) == nil {
		t.Fatal("login after the confirmed link refused")
	}
}

func TestOIDCPasswordConfirmationWrongPasswordIsRefused(t *testing.T) {
	f := newOIDCFixture(t)
	conf := config.GetInstance()
	duration, failures := conf.Security.LoginLockoutDuration, conf.Security.LoginMaxFailures
	conf.Security.LoginLockoutDuration, conf.Security.LoginMaxFailures = 60, 3
	t.Cleanup(func() {
		conf.Security.LoginLockoutDuration, conf.Security.LoginMaxFailures = duration, failures
		ClearLoginAttempt("192.0.2.1")
	})
	ClearLoginAttempt("192.0.2.1")
	f.idp.claims["preferred_username"] = "alice"

	rec := f.login()
	link := linkCookieOf(rec)
	if link == nil {
		t.Fatal("no link cookie")
	}
	reply, _ := f.confirm(link, "not-alices-password", "")
	if reply.Code != -2 {
		t.Fatalf("wrong password answered %+v", reply)
	}
	if _, sub := f.link("alice"); sub != "" {
		t.Fatalf("a wrong password linked alice to %q", sub)
	}
	if n := f.sessions("alice"); n != 0 {
		t.Fatalf("%d sessions after a wrong password", n)
	}
	// The failures count toward the login's rate limit: the third locks the
	// address, and then even the right password is turned away.
	f.confirm(link, "wrong-again", "")
	f.confirm(link, "wrong-a-third-time", "")
	if locked, _, _ := CheckLoginAttempt("192.0.2.1"); !locked {
		t.Fatal("three wrong passwords did not lock the address")
	}
	if reply, _ = f.confirm(link, "alice-password", ""); reply.Code == 0 {
		t.Fatal("a locked address confirmed the link")
	}
	if _, sub := f.link("alice"); sub != "" {
		t.Fatalf("alice linked to %q from a locked address", sub)
	}
}

func TestOIDCPasswordConfirmationIsDroppedAfterTooManyTries(t *testing.T) {
	f := newOIDCFixture(t)
	conf := config.GetInstance()
	failures := conf.Security.LoginMaxFailures
	conf.Security.LoginMaxFailures = 100
	t.Cleanup(func() {
		conf.Security.LoginMaxFailures = failures
		ClearLoginAttempt("192.0.2.1")
	})
	ClearLoginAttempt("192.0.2.1")
	f.idp.claims["preferred_username"] = "alice"
	link := linkCookieOf(f.login())
	for i := 0; i < oidcLinkAttempts; i++ {
		f.confirm(link, fmt.Sprintf("wrong-%d", i), "")
	}
	if reply, _ := f.confirm(link, "alice-password", ""); reply.Code != codeOIDCLinkGone {
		t.Fatalf("after %d wrong passwords: %+v", oidcLinkAttempts, reply)
	}
}

func TestOIDCPasswordConfirmationAsksForTheSecondFactor(t *testing.T) {
	f := newOIDCFixture(t)
	if _, err := f.store.SetTOTP("alice", "sealed-secret", []string{"unused"}); err != nil {
		t.Fatal(err)
	}
	f.idp.claims["preferred_username"] = "alice"
	link := linkCookieOf(f.login())
	if reply, _ := f.call(http.MethodGet, "/api/auth/oidc/link", nil, link); !reply.Data.TOTPRequired {
		t.Fatalf("link info %+v", reply)
	}
	if reply, _ := f.confirm(link, "alice-password", ""); reply.Code != -2 {
		t.Fatalf("the password alone answered %+v", reply)
	}
	if _, sub := f.link("alice"); sub != "" {
		t.Fatalf("alice linked to %q without the second factor", sub)
	}
}

// An account made by a single sign-on has no password of its own. Once
// unlinked, only an administrator links it again.
func TestOIDCAccountWithoutAPasswordCanOnlyBePreLinked(t *testing.T) {
	f := newOIDCFixture(t)
	resetPendingSignIns(t)
	f.save(func(s *oidcSettings) { s.AutoCreate = true })
	f.idp.claims["preferred_username"] = "dave"
	if sessionCookieOf(f.login()) == nil {
		t.Fatal("auto-create login refused")
	}
	if _, err := f.store.ClearOIDC("dave"); err != nil {
		t.Fatal(err)
	}
	rec := f.loginFresh()
	f.refused(rec, "unlinked")
	if c := linkCookieOf(rec); c != nil {
		t.Fatalf("an account without a password was offered the password step: %+v", c)
	}
	admin := loginCookie(t, f.router, "admin", "admin")
	if p := f.pending(admin).Data.SignIns; len(p) != 1 || p[0].Username != "dave" {
		t.Fatalf("pending %+v", p)
	}
}

func TestOIDCPendingSignInsAreBounded(t *testing.T) {
	resetPendingSignIns(t)
	start := time.Now()
	now := start
	oidcNow = func() time.Time { return now }
	t.Cleanup(func() { oidcNow = time.Now })

	for i := 0; i < maxPendingSignIns+5; i++ {
		now = now.Add(time.Second)
		recordPendingSignIn(pendingSignIn{Username: "alice", Issuer: "https://idp", Subject: fmt.Sprintf("sub-%d", i)})
	}
	list := listPendingSignIns()
	if len(list) != maxPendingSignIns {
		t.Fatalf("%d pending sign-ins kept, want %d", len(list), maxPendingSignIns)
	}
	// Newest first, and the oldest are the ones dropped.
	if list[0].Subject != fmt.Sprintf("sub-%d", maxPendingSignIns+4) || list[len(list)-1].Subject != "sub-5" {
		t.Fatalf("kept %s .. %s", list[0].Subject, list[len(list)-1].Subject)
	}

	// The same IdP account again is one entry, moved to the top.
	now = now.Add(time.Second)
	recordPendingSignIn(pendingSignIn{Username: "alice", Issuer: "https://idp", Subject: "sub-10"})
	list = listPendingSignIns()
	if len(list) != maxPendingSignIns || list[0].Subject != "sub-10" {
		t.Fatalf("repeat: %d entries, first %s", len(list), list[0].Subject)
	}

	// A day later they are gone.
	now = now.Add(pendingSignInTTL)
	if list = listPendingSignIns(); len(list) != 0 {
		t.Fatalf("%d pending sign-ins after a day", len(list))
	}
}

// The pending list and the pre-link answer an administrator's session only.
func TestOIDCPreLinkRoutesNeedAnAdministrator(t *testing.T) {
	f := newOIDCFixture(t)
	alice := loginCookie(t, f.router, "alice", "alice-password")
	routes := [][2]string{
		{http.MethodGet, "/api/auth/oidc/pending"},
		{http.MethodPut, "/api/auth/users/alice/oidc"},
	}
	for _, route := range routes {
		if code := requestJSON(f.router, route[0], route[1], map[string]string{"subject": "s"}, alice); code != http.StatusForbidden {
			t.Errorf("%s %s: a user got %d", route[0], route[1], code)
		}
		if code := requestJSON(f.router, route[0], route[1], map[string]string{"subject": "s"}, nil); code != http.StatusUnauthorized {
			t.Errorf("%s %s: nobody got %d", route[0], route[1], code)
		}
	}
	if _, sub := f.link("alice"); sub != "" {
		t.Fatalf("alice linked to %q", sub)
	}
}
