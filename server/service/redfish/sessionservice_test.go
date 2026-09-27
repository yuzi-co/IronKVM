package redfish

import (
	"net/http"
	"strings"
	"testing"
)

const sessionsURL = "/redfish/v1/SessionService/Sessions"

// login posts to the session collection and returns the response.
func (h *harness) login(username, password string) (int, string, string) {
	h.t.Helper()

	w := h.do(http.MethodPost, sessionsURL, `{"UserName":"`+username+`","Password":"`+password+`"}`)
	return w.Code, w.Header().Get("X-Auth-Token"), w.Header().Get("Location")
}

func TestLoginReturnsATokenAndTheSessionsLocation(t *testing.T) {
	h := newHarness(t)

	w := h.do(http.MethodPost, sessionsURL, `{"UserName":"admin","Password":"admin"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	token := w.Header().Get("X-Auth-Token")
	location := w.Header().Get("Location")
	if token == "" || !strings.HasPrefix(location, sessionsURL+"/") {
		t.Fatalf("token %q, location %q", token, location)
	}
	body := decode(t, w)
	if body["UserName"] != "admin" || body["@odata.id"] != location {
		t.Fatalf("body is %v", body)
	}
	if strings.Contains(w.Body.String(), token) {
		t.Fatal("the token is in the body")
	}

	// The token works, and the session is where Location says.
	if w := h.do(http.MethodGet, location, "", "X-Auth-Token", token); w.Code != http.StatusOK {
		t.Fatalf("GET %s with the token: %d %s", location, w.Code, w.Body.String())
	}
	if len(h.limiter.succeeded) != 1 {
		t.Fatalf("successes recorded: %v", h.limiter.succeeded)
	}
}

func TestLoginWithABadPasswordCountsAsAFailure(t *testing.T) {
	h := newHarness(t)

	w := h.do(http.MethodPost, sessionsURL, `{"UserName":"admin","Password":"wrong"}`)
	expectError(t, w, http.StatusUnauthorized, "NoValidSession")
	if w.Header().Get("X-Auth-Token") != "" {
		t.Fatal("a failed login returned a token")
	}
	if len(h.limiter.failed) != 1 {
		t.Fatalf("failures recorded: %v", h.limiter.failed)
	}
}

func TestLoginFromALockedOutAddressIsRefused(t *testing.T) {
	h := newHarness(t)
	h.limiter.locked[clientIP] = true

	if code, token, _ := h.login("admin", "admin"); code != http.StatusUnauthorized || token != "" {
		t.Fatalf("got %d with token %q", code, token)
	}
}

func TestLoginNeedsBothProperties(t *testing.T) {
	h := newHarness(t)

	w := h.do(http.MethodPost, sessionsURL, `{"UserName":"admin"}`)
	expectError(t, w, http.StatusBadRequest, "PropertyMissing")

	w = h.do(http.MethodPost, sessionsURL, `{"UserName":"admin",`)
	expectError(t, w, http.StatusBadRequest, "MalformedJSON")

	w = h.do(http.MethodPost, sessionsURL, `{"UserName":"admin","Password":"admin"}`, "Content-Type", "text/plain")
	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("text/plain body: %d", w.Code)
	}
}

func TestWithAuthenticationDisabledAnyLoginSucceeds(t *testing.T) {
	h := newHarness(t)
	h.authDisabled = true

	if code, token, _ := h.login("anyone", "anything"); code != http.StatusCreated || token == "" {
		t.Fatalf("got %d with token %q", code, token)
	}
}

func TestLogoutEndsTheSession(t *testing.T) {
	h := newHarness(t)
	_, token, location := h.login("alice", "valid-password")

	w := h.do(http.MethodDelete, location, "", "X-Auth-Token", token)
	if w.Code != http.StatusNoContent {
		t.Fatalf("DELETE: %d %s", w.Code, w.Body.String())
	}

	w = h.do(http.MethodGet, "/redfish/v1/SessionService", "", "X-Auth-Token", token)
	expectError(t, w, http.StatusUnauthorized, "NoValidSession")
}

func TestOnlyAnAdminEndsSomeoneElsesSession(t *testing.T) {
	h := newHarness(t)
	_, _, adminSession := h.login("admin", "admin")
	_, _, aliceSession := h.login("alice", "valid-password")

	w := h.do(http.MethodDelete, adminSession, "", h.user()...)
	expectError(t, w, http.StatusForbidden, "InsufficientPrivilege")

	if w := h.do(http.MethodDelete, aliceSession, "", h.admin()...); w.Code != http.StatusNoContent {
		t.Fatalf("admin deleting alice's session: %d %s", w.Code, w.Body.String())
	}
}

func TestSessionCollectionListsTheLiveSessions(t *testing.T) {
	h := newHarness(t)
	h.login("admin", "admin")
	token := h.user()

	w := h.do(http.MethodGet, sessionsURL, "", token...)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if count := decode(t, w)["Members@odata.count"]; count != float64(2) {
		t.Fatalf("Members@odata.count is %v, want 2", count)
	}
}

func TestAnUnknownSessionIsNotFound(t *testing.T) {
	h := newHarness(t)

	w := h.do(http.MethodGet, sessionsURL+"/nope", "", h.admin()...)
	expectError(t, w, http.StatusNotFound, "ResourceMissingAtURI")

	w = h.do(http.MethodDelete, sessionsURL+"/nope", "", h.admin()...)
	expectError(t, w, http.StatusNotFound, "ResourceMissingAtURI")
}

func TestSessionServiceReportsTheIdleTimeout(t *testing.T) {
	h := newHarness(t)

	w := h.do(http.MethodGet, "/redfish/v1/SessionService", "", h.user()...)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if timeout := decode(t, w)["SessionTimeout"]; timeout != float64(1800) {
		t.Fatalf("SessionTimeout is %v", timeout)
	}
}
