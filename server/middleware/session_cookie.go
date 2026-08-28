package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/config"
)

// RequestIsSecure reports whether this request reached the server over TLS.
//
// It asks the request, not the configuration. The configuration says what the
// server offers; only the request says how the browser actually arrived, and
// the Secure attribute has to follow the second one. A cookie marked Secure on
// a plain HTTP response is stored by the browser and then never sent back over
// that scheme, so the session silently fails to establish and the login page
// reappears with nothing logged.
func RequestIsSecure(c *gin.Context) bool {
	if c.Request.TLS != nil {
		return true
	}

	// A reverse proxy that terminates TLS says so here. Anyone can send this
	// header, but the worst a caller can do with it is mark their own cookie
	// Secure and lock themselves out; it grants nothing.
	return strings.EqualFold(strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")), "https")
}

// SetSessionCookie writes the session cookie for this request.
func SetSessionCookie(c *gin.Context, token string) {
	conf := config.GetInstance()

	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		CookieName,
		token,
		int(conf.JWT.RefreshTokenDuration),
		"/",
		"",
		RequestIsSecure(c),
		true,
	)
}

// ClearSessionCookie deletes the session cookie.
//
// The Secure attribute has to match the cookie being deleted. A browser holding
// a Secure cookie ignores a deletion that arrives over plain HTTP, so a caller
// that wants a Secure cookie gone has to say so on a response that is itself
// encrypted.
func ClearSessionCookie(c *gin.Context, secure bool) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(CookieName, "", -1, "/", "", secure, true)
}
