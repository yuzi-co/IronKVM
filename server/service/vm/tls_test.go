package vm

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/middleware"
)

// Switching HTTPS on used to mint a certificate every time, unconditionally.
//
// That is not a small waste. A self-signed certificate has to be trusted by
// hand, and replacing it voids the trust the operator installed, so every
// switch cost them a manual reinstall on every machine that talks to the board.
// Switching HTTPS off leaves the certificate on disk, so the one they already
// trust is sitting right there to be reused.
//
// utils.EnsureCert keeps a certificate that still answers for the device and
// replaces one that does not. These tests pin that enableTls goes through it,
// because reverting to GenerateCert would compile, pass every certificate test
// in utils, and quietly bring the manual step back.

func TestEnablingTlsGoesThroughTheCertificateThatMayBeKept(t *testing.T) {
	original := ensureTlsCert
	t.Cleanup(func() { ensureTlsCert = original })

	called := 0
	ensureTlsCert = func() error {
		called++
		return nil
	}

	// config.Write reaches the real configuration file, so this stops at the
	// first step. The certificate decision is what is under test and it happens
	// before anything is written.
	_ = enableTls()

	if called != 1 {
		t.Fatalf("enableTls consulted the certificate %d time(s), want 1", called)
	}
}

func TestEnablingTlsStopsWhenTheCertificateCannotBeMade(t *testing.T) {
	original := ensureTlsCert
	t.Cleanup(func() { ensureTlsCert = original })

	want := errors.New("no certificate")
	ensureTlsCert = func() error { return want }

	// Writing proto: https with no usable certificate leaves a board that
	// starts a TLS listener it cannot serve from, reachable on neither scheme.
	if err := enableTls(); !errors.Is(err, want) {
		t.Fatalf("enableTls returned %v, want %v", err, want)
	}
}

// Switching HTTPS off leaves the browser holding a cookie it can no longer use.
//
// The cookie was set with Secure while HTTPS was on, so the browser will not
// send it over plain HTTP, and will not let a plain HTTP response replace or
// delete it either. The operator then logs in over HTTP, the form posts, the
// server answers with a new cookie the browser refuses to store, and the login
// page comes back. Nothing is logged, and the board looks broken.
//
// The response to this request is the last one that travels over HTTPS, so it
// is the only chance to take the cookie back.
func TestDisablingTlsTakesBackTheSecureSessionCookie(t *testing.T) {
	original := disableTlsConfig
	t.Cleanup(func() { disableTlsConfig = original })
	disableTlsConfig = func() error { return nil }

	cookie := setTlsAndReadSessionCookie(t, false)
	if cookie == nil {
		t.Fatal("disabling TLS left the Secure session cookie in the browser")
	}

	if !cookie.Secure {
		t.Error("the deletion is not marked Secure, so a browser holding a Secure cookie ignores it")
	}

	if cookie.MaxAge >= 0 {
		t.Errorf("the deletion carries MaxAge=%d, want a negative age", cookie.MaxAge)
	}

	if cookie.Value != "" {
		t.Errorf("the deletion carries the value %q, want an empty one", cookie.Value)
	}
}

// Going the other way keeps the session. The cookie set over plain HTTP is not
// Secure, so it still travels once the listener moves to HTTPS, and logging the
// operator out for no reason is a worse answer than leaving them alone.
func TestEnablingTlsLeavesTheSessionAlone(t *testing.T) {
	originalCert := ensureTlsCert
	t.Cleanup(func() { ensureTlsCert = originalCert })
	ensureTlsCert = func() error { return nil }

	originalEnable := enableTlsConfig
	t.Cleanup(func() { enableTlsConfig = originalEnable })
	enableTlsConfig = func() error { return nil }

	if cookie := setTlsAndReadSessionCookie(t, true); cookie != nil {
		t.Errorf("enabling TLS cleared the session cookie: %+v", cookie)
	}
}

func setTlsAndReadSessionCookie(t *testing.T, enabled bool) *http.Cookie {
	t.Helper()

	originalRestart := restartServer
	t.Cleanup(func() { restartServer = originalRestart })
	restartServer = func() {}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	body := fmt.Sprintf(`{"enabled":%t}`, enabled)
	request := httptest.NewRequest(http.MethodPost, "/api/vm/tls", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	c.Request = request

	service := &Service{}
	service.SetTls(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("SetTls answered %d", recorder.Code)
	}

	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == middleware.CookieName {
			return cookie
		}
	}

	return nil
}
