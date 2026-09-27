package redfish

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/auth"
)

const principalKey = "redfish.principal"

// principal is who a request acts as.
type principal struct {
	username string
	admin    bool
	// sessionID is set when the request came with a session token.
	sessionID string
}

func currentPrincipal(c *gin.Context) principal {
	value, _ := c.Get(principalKey)
	p, _ := value.(principal)
	return p
}

// authenticate runs before every /redfish route. It accepts a session token
// or an API key in X-Auth-Token, or HTTP Basic credentials, and lets the
// public routes through without any.
func (s *Service) authenticate(c *gin.Context) {
	// Every answer carries it, the 204s from actions included.
	c.Header("OData-Version", "4.0")

	if s.deps.AuthDisabled() {
		c.Set(principalKey, principal{username: "admin", admin: true})
		c.Next()
		return
	}

	// Basic credentials a browser has cached would otherwise ride along on a
	// cross-site request, as a cookie does.
	if !middleware.SameOrigin(c.Request) {
		writeError(c, http.StatusForbidden, "InsufficientPrivilege", "cross-origin requests are refused")
		return
	}

	if publicRoutes[c.Request.Method+" "+c.FullPath()] {
		c.Next()
		return
	}

	if token := c.GetHeader("X-Auth-Token"); token != "" {
		p, ok := s.tokenPrincipal(token)
		if !ok {
			unauthorized(c)
			return
		}
		c.Set(principalKey, p)
		c.Next()
		return
	}

	if username, password, ok := c.Request.BasicAuth(); ok {
		user, ok := s.checkPassword(c, username, password)
		if !ok {
			return
		}
		c.Set(principalKey, principal{username: user.Username, admin: user.Role == authn.RoleAdmin})
		c.Next()
		return
	}

	unauthorized(c)
}

// tokenPrincipal resolves X-Auth-Token: a session token first, then an API
// key. A session whose account has since been disabled, deleted or had its
// password changed is ended here.
func (s *Service) tokenPrincipal(token string) (principal, bool) {
	if sess, ok := s.sessions.byToken(token); ok {
		user, err := s.deps.Accounts.ValidateToken(sess.username, sess.tokenVersion)
		if err != nil {
			s.sessions.remove(sess.id)
			return principal{}, false
		}
		return principal{username: user.Username, admin: user.Role == authn.RoleAdmin, sessionID: sess.id}, true
	}

	username, ok := s.deps.APIKeyUser(token)
	if !ok {
		return principal{}, false
	}
	user, err := s.deps.Accounts.Get(username)
	if err != nil || !user.Enabled {
		return principal{}, false
	}
	return principal{username: user.Username, admin: user.Role == authn.RoleAdmin}, true
}

// checkPassword checks one set of credentials under the login's brute-force
// limit. On failure it has answered already.
func (s *Service) checkPassword(c *gin.Context, username, password string) (*authn.User, bool) {
	ip := c.RemoteIP()
	if s.deps.Limiter.Locked(ip, username) {
		unauthorized(c)
		return nil, false
	}

	// Credentials that passed within the last minute, for the account as it
	// is now, skip the bcrypt.
	if user, err := s.deps.Accounts.Get(username); err == nil && user.Enabled && s.credentials.has(user, password) {
		return user, true
	}

	user, ok, err := s.deps.Accounts.Authenticate(username, password)
	if err != nil {
		log.Errorf("redfish: load accounts: %s", err)
		writeError(c, http.StatusInternalServerError, "GeneralError", "authentication unavailable")
		return nil, false
	}
	if !ok {
		s.deps.Limiter.Failed(ip, username)
		time.Sleep(s.deps.FailureDelay)
		unauthorized(c)
		return nil, false
	}

	// Guesses sent at once all pass the check above before any of them has
	// failed. A right one that lands after the others locked the account is
	// refused, as it would have been had it come a moment later.
	if s.deps.Limiter.Locked(ip, username) {
		unauthorized(c)
		return nil, false
	}

	s.deps.Limiter.Succeeded(ip, username)
	s.credentials.remember(user, password)
	return user, true
}

func unauthorized(c *gin.Context) {
	c.Header("WWW-Authenticate", `Basic realm="IronKVM"`)
	writeError(c, http.StatusUnauthorized, "NoValidSession", "")
}

// adminOnly wraps a handler that changes something: pressing a button,
// changing media, or ending another user's session.
func adminOnly(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !currentPrincipal(c).admin {
			writeError(c, http.StatusForbidden, "InsufficientPrivilege", "")
			return
		}
		handler(c)
	}
}

// LoginLimiter applies the web login's brute-force limit from service/auth,
// counted per address and account: a success for one account must not wipe
// the failures against another, or a single valid account would buy
// unlimited guesses at the admin's password. A lockout the web login put on
// the whole address applies here too.
type LoginLimiter struct{}

// limiterKey is the record Redfish failures count under in the web login's
// table. It cannot clash with an address, which never holds a "/".
func limiterKey(ip, username string) string {
	return ip + "/" + username
}

func (LoginLimiter) Locked(ip, username string) bool {
	if locked, _, _ := auth.CheckLoginAttempt(ip); locked {
		return true
	}
	locked, _, _ := auth.CheckLoginAttempt(limiterKey(ip, username))
	return locked
}

func (LoginLimiter) Failed(ip, username string) {
	auth.RecordLoginFailure(limiterKey(ip, username))
}

func (LoginLimiter) Succeeded(ip, username string) {
	auth.ClearLoginAttempt(limiterKey(ip, username))
}
