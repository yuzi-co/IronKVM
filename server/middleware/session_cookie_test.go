package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// tlsState stands in for a finished handshake. Only its presence matters.
var tlsState = tls.ConnectionState{HandshakeComplete: true}

func contextFor(request *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = request

	return c, recorder
}

// The Secure attribute has to follow the request, not the configuration. The
// server can be configured for HTTPS and still be answering a plain HTTP
// request, and a Secure cookie on that response is stored and never sent back.
func TestRequestIsSecureReadsTheRequest(t *testing.T) {
	plain := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	if c, _ := contextFor(plain); RequestIsSecure(c) {
		t.Error("a plain request reported itself as secure")
	}

	encrypted := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	encrypted.TLS = &tlsState
	if c, _ := contextFor(encrypted); !RequestIsSecure(c) {
		t.Error("a TLS request reported itself as insecure")
	}

	proxied := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	proxied.Header.Set("X-Forwarded-Proto", "HTTPS")
	if c, _ := contextFor(proxied); !RequestIsSecure(c) {
		t.Error("a request forwarded from a TLS proxy reported itself as insecure")
	}

	downgraded := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	downgraded.Header.Set("X-Forwarded-Proto", "http")
	if c, _ := contextFor(downgraded); RequestIsSecure(c) {
		t.Error("a request forwarded over plain HTTP reported itself as secure")
	}
}

// The session cookie set on a plain HTTP response must be readable on the next
// plain HTTP request. Marking it Secure is what made the login page reappear
// after HTTPS was switched off, with nothing logged.
func TestSessionCookieOverPlainHttpIsNotSecure(t *testing.T) {
	c, recorder := contextFor(httptest.NewRequest(http.MethodPost, "/api/auth/login", nil))
	SetSessionCookie(c, "a-token")

	cookie := sessionCookieFrom(t, recorder)
	if cookie.Secure {
		t.Error("the cookie set over plain HTTP is Secure, so the browser will never send it back")
	}

	if !cookie.HttpOnly {
		t.Error("the cookie is readable from script")
	}
}

func TestSessionCookieOverTlsIsSecure(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	request.TLS = &tlsState

	c, recorder := contextFor(request)
	SetSessionCookie(c, "a-token")

	if cookie := sessionCookieFrom(t, recorder); !cookie.Secure {
		t.Error("the cookie set over TLS is not Secure")
	}
}

// A browser holding a Secure cookie ignores a deletion that arrives without
// the attribute, so the caller decides and the helper does not guess.
func TestClearSessionCookieCarriesTheAttributeItIsGiven(t *testing.T) {
	for _, secure := range []bool{true, false} {
		c, recorder := contextFor(httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil))
		ClearSessionCookie(c, secure)

		cookie := sessionCookieFrom(t, recorder)
		if cookie.Secure != secure {
			t.Errorf("cleared with Secure=%v, want %v", cookie.Secure, secure)
		}

		if cookie.MaxAge >= 0 {
			t.Errorf("cleared with MaxAge=%d, want a negative age so the browser drops it", cookie.MaxAge)
		}
	}
}

func sessionCookieFrom(t *testing.T, recorder *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()

	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == CookieName {
			return cookie
		}
	}

	t.Fatal("no session cookie was set")
	return nil
}
