package redfish

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/service/auth"
)

// withProbe adds a GET behind authenticate that reports who the request acts
// as, and a POST behind adminOnly.
func withProbe(h *harness) *harness {
	g := h.engine.Group("", h.service.authenticate)
	g.GET("/redfish/v1/Probe", func(c *gin.Context) {
		p := currentPrincipal(c)
		c.String(http.StatusOK, "%s admin=%t", p.username, p.admin)
	})
	g.POST("/redfish/v1/Probe", adminOnly(func(c *gin.Context) {
		c.String(http.StatusOK, "done")
	}))
	return h
}

func TestNoCredentialsAnswer401WithABasicChallenge(t *testing.T) {
	h := withProbe(newHarness(t))

	w := h.do(http.MethodGet, "/redfish/v1/Probe", "")
	expectError(t, w, http.StatusUnauthorized, "NoValidSession")
	if got := w.Header().Get("WWW-Authenticate"); got != `Basic realm="IronKVM"` {
		t.Fatalf("WWW-Authenticate is %q", got)
	}
}

func TestBasicCredentialsActAsTheAccount(t *testing.T) {
	h := withProbe(newHarness(t))

	for _, tc := range []struct{ username, password, want string }{
		{"admin", "admin", "admin admin=true"},
		{"alice", "valid-password", "alice admin=false"},
	} {
		w := h.do(http.MethodGet, "/redfish/v1/Probe", "", "Authorization", basic(tc.username, tc.password))
		if w.Code != http.StatusOK || w.Body.String() != tc.want {
			t.Fatalf("%s: got %d %q, want %q", tc.username, w.Code, w.Body.String(), tc.want)
		}
	}
	if len(h.limiter.succeeded) != 2 {
		t.Fatalf("successes recorded: %v", h.limiter.succeeded)
	}
}

func TestBadBasicCredentialsCountAsALoginFailure(t *testing.T) {
	h := withProbe(newHarness(t))

	w := h.do(http.MethodGet, "/redfish/v1/Probe", "", "Authorization", basic("admin", "wrong"))
	expectError(t, w, http.StatusUnauthorized, "NoValidSession")
	if len(h.limiter.failed) != 1 || h.limiter.failed[0] != clientIP {
		t.Fatalf("failures recorded: %v", h.limiter.failed)
	}
}

func TestALockedOutAddressIsRefusedEvenWithTheRightPassword(t *testing.T) {
	h := withProbe(newHarness(t))
	h.limiter.locked[clientIP] = true

	w := h.do(http.MethodGet, "/redfish/v1/Probe", "", "Authorization", basic("admin", "admin"))
	expectError(t, w, http.StatusUnauthorized, "NoValidSession")
}

func TestASessionTokenActsAsItsAccount(t *testing.T) {
	h := withProbe(newHarness(t))

	w := h.do(http.MethodGet, "/redfish/v1/Probe", "", h.user()...)
	if w.Code != http.StatusOK || w.Body.String() != "alice admin=false" {
		t.Fatalf("got %d %q", w.Code, w.Body.String())
	}
}

func TestAnUnknownTokenIsRefused(t *testing.T) {
	h := withProbe(newHarness(t))

	w := h.do(http.MethodGet, "/redfish/v1/Probe", "", "X-Auth-Token", "0123")
	expectError(t, w, http.StatusUnauthorized, "NoValidSession")
}

func TestAnIdleSessionExpires(t *testing.T) {
	h := withProbe(newHarness(t))
	token := h.admin()

	h.clock = h.clock.Add(31 * time.Minute)

	w := h.do(http.MethodGet, "/redfish/v1/Probe", "", token...)
	expectError(t, w, http.StatusUnauthorized, "NoValidSession")
}

// A password change or a revoke bumps the account's token version, which
// ends its Redfish sessions as it ends its web ones.
func TestARevokedAccountLosesItsSession(t *testing.T) {
	h := withProbe(newHarness(t))
	token := h.user()

	if _, err := h.accounts.Revoke("alice"); err != nil {
		t.Fatal(err)
	}

	w := h.do(http.MethodGet, "/redfish/v1/Probe", "", token...)
	expectError(t, w, http.StatusUnauthorized, "NoValidSession")
	if len(h.service.sessions.list()) != 0 {
		t.Fatal("the revoked session is still in the store")
	}
}

func TestAnAPIKeyInXAuthTokenActsAsItsAccount(t *testing.T) {
	h := withProbe(newHarness(t))
	h.keys["nkvm_secret"] = "alice"

	w := h.do(http.MethodGet, "/redfish/v1/Probe", "", "X-Auth-Token", "nkvm_secret")
	if w.Code != http.StatusOK || w.Body.String() != "alice admin=false" {
		t.Fatalf("got %d %q", w.Code, w.Body.String())
	}
}

func TestAnAPIKeyOfADisabledAccountIsRefused(t *testing.T) {
	h := withProbe(newHarness(t))
	h.keys["nkvm_secret"] = "alice"
	disabled := false
	if _, err := h.accounts.Update("admin", "alice", authn.UserPatch{Enabled: &disabled}); err != nil {
		t.Fatal(err)
	}

	w := h.do(http.MethodGet, "/redfish/v1/Probe", "", "X-Auth-Token", "nkvm_secret")
	expectError(t, w, http.StatusUnauthorized, "NoValidSession")
}

func TestWithAuthenticationDisabledEveryRequestIsAdmin(t *testing.T) {
	h := withProbe(newHarness(t))
	h.authDisabled = true

	w := h.do(http.MethodGet, "/redfish/v1/Probe", "")
	if w.Code != http.StatusOK || w.Body.String() != "admin admin=true" {
		t.Fatalf("got %d %q", w.Code, w.Body.String())
	}
}

// A browser that has cached Basic credentials for the board sends them on a
// cross-site request too. The Origin header gives such a request away.
func TestACrossOriginRequestIsRefused(t *testing.T) {
	h := withProbe(newHarness(t))

	w := h.do(http.MethodPost, "/redfish/v1/Probe", "",
		"Authorization", basic("admin", "admin"), "Origin", "https://evil.example")
	expectError(t, w, http.StatusForbidden, "InsufficientPrivilege")

	// httptest requests are addressed to example.com.
	w = h.do(http.MethodPost, "/redfish/v1/Probe", "",
		"Authorization", basic("admin", "admin"), "Origin", "http://example.com")
	if w.Code != http.StatusOK {
		t.Fatalf("same-origin request: %d %s", w.Code, w.Body.String())
	}
}

func TestAdminOnlyRefusesAUser(t *testing.T) {
	h := withProbe(newHarness(t))

	w := h.do(http.MethodPost, "/redfish/v1/Probe", "", h.user()...)
	expectError(t, w, http.StatusForbidden, "InsufficientPrivilege")

	w = h.do(http.MethodPost, "/redfish/v1/Probe", "", h.admin()...)
	if w.Code != http.StatusOK {
		t.Fatalf("admin: %d %s", w.Code, w.Body.String())
	}
}

// LoginLimiter is the web login's own limit, so failures over Redfish lock
// the address out of both.
func TestLoginLimiterUsesTheLoginsBruteForceLimit(t *testing.T) {
	conf := config.GetInstance()
	original := conf.Security
	conf.Security.LoginLockoutDuration = 60
	conf.Security.LoginMaxFailures = 2
	t.Cleanup(func() { conf.Security = original })

	ip := fmt.Sprintf("198.51.100.%d", time.Now().UnixNano()%250+1)
	t.Cleanup(func() { auth.ClearLoginAttempt(ip) })

	var limiter LoginLimiter
	limiter.Failed(ip)
	if limiter.Locked(ip) {
		t.Fatal("locked after one failure")
	}
	limiter.Failed(ip)
	if !limiter.Locked(ip) {
		t.Fatal("not locked after the second failure")
	}
	if locked, _, _ := auth.CheckLoginAttempt(ip); !locked {
		t.Fatal("the web login does not see the lockout")
	}

	limiter.Succeeded(ip)
	if limiter.Locked(ip) {
		t.Fatal("still locked after a success cleared the record")
	}
}
