package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/service/network"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
)

// noteSignedInAdmin reports a request that an administrator's session
// authenticated, which is the proof an address trial waits for. See
// service/network/ethernet_reach.go.
//
// It is a var so a test can watch the seam. Without one, deleting the call in
// authenticateBySession leaves every test in the tree green and the feature
// gone, because the network package's own tests call it directly.
var noteSignedInAdmin = network.NoteSignedInAdminRequest

const (
	principalContextKey = "principal"
	tokenContextKey     = "token"
	CookieName          = "nano-kvm-token"
	sessionRecheckDelay = 5 * time.Second
)

type Principal struct {
	Username string
	Role     authn.Role
}

type Token struct {
	Username     string `json:"username"`
	TokenVersion uint64 `json:"tokenVersion"`
	jwt.RegisteredClaims
}

func CheckToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !allowByOrigin(c.Request) {
			abortForbiddenOrigin(c)
			return
		}

		principal, token, ok := authenticate(c)
		if !ok {
			abortUnauthorized(c)
			return
		}

		c.Set(principalContextKey, principal)
		c.Set(tokenContextKey, token)
		if token == nil || token.ExpiresAt == nil {
			c.Next()
			return
		}

		requestContext, cancel := context.WithCancel(c.Request.Context())
		c.Request = c.Request.WithContext(requestContext)
		unregister := activeSessions.register(principal.Username, cancel)
		timer := time.AfterFunc(time.Until(token.ExpiresAt.Time), cancel)
		go watchSessionState(requestContext, cancel, principal.Username, token.TokenVersion, sessionRecheckDelay)
		defer func() {
			timer.Stop()
			unregister()
			cancel()
		}()

		c.Next()
	}
}

// CheckSession guards routes that an api key must not reach: the ones that
// issue and revoke the keys themselves. A stolen key already has the run of
// the API, but this stops it minting further keys that would outlive the
// password change made to shut it out.
func CheckSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !allowByOrigin(c.Request) {
			abortForbiddenOrigin(c)
			return
		}

		principal, token, ok := authenticateBySession(c)
		if !ok {
			abortUnauthorized(c)
			return
		}

		c.Set(principalContextKey, principal)
		c.Set(tokenContextKey, token)
		c.Next()
	}
}

func watchSessionState(ctx context.Context, cancel context.CancelFunc, username string, tokenVersion uint64, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := authn.DefaultStore.ValidateToken(username, tokenVersion); err != nil {
				cancel()
				return
			}
		}
	}
}

func RequireRole(roles ...authn.Role) gin.HandlerFunc {
	allowed := make(map[authn.Role]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(c *gin.Context) {
		principal, ok := CurrentPrincipal(c)
		if !ok {
			abortUnauthorized(c)
			return
		}
		if _, ok = allowed[principal.Role]; !ok {
			c.JSON(http.StatusForbidden, "forbidden")
			c.Abort()
			return
		}
		c.Next()
	}
}

func CurrentPrincipal(c *gin.Context) (Principal, bool) {
	value, exists := c.Get(principalContextKey)
	if !exists {
		return Principal{}, false
	}
	principal, ok := value.(Principal)
	return principal, ok
}

func CheckLoopbackInternalToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		if allowByLoopbackInternalToken(c.Request) {
			c.Next()
			return
		}
		abortUnauthorized(c)
	}
}

func CheckTokenOrLoopbackInternalToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !allowByOrigin(c.Request) {
			abortForbiddenOrigin(c)
			return
		}

		if allowByLoopbackInternalToken(c.Request) {
			c.Next()
			return
		}

		principal, token, ok := authenticate(c)
		if !ok {
			abortUnauthorized(c)
			return
		}
		c.Set(principalContextKey, principal)
		c.Set(tokenContextKey, token)
		c.Next()
	}
}

// authenticate accepts either of the two credentials that speak for a user: a
// login session, or an api key issued to one. CheckSession uses the session
// half on its own.
func authenticate(c *gin.Context) (Principal, *Token, bool) {
	if principal, token, ok := authenticateBySession(c); ok {
		return principal, token, true
	}

	if principal, ok := authenticateByAPIKey(c.Request); ok {
		return principal, nil, true
	}

	return Principal{}, nil, false
}

func authenticateBySession(c *gin.Context) (Principal, *Token, bool) {
	conf := config.GetInstance()
	if conf.Authentication == "disable" {
		// Not reported as a client that reached the board. With the check off
		// there is no signed-in client to speak of, and an address trial that
		// confirmed itself on any request at all would confirm on a port scan.
		return Principal{Username: "admin", Role: authn.RoleAdmin}, nil, true
	}

	cookie, err := c.Cookie(CookieName)
	if err != nil {
		return Principal{}, nil, false
	}
	token, err := ParseJWT(cookie)
	if err != nil {
		return Principal{}, nil, false
	}
	user, err := authn.DefaultStore.ValidateToken(token.Username, token.TokenVersion)
	if err != nil {
		log.Debugf("validate session for %q: %s", token.Username, err)
		return Principal{}, nil, false
	}

	// Reported from the one place that decides a session is genuine, rather
	// than from each route. Only an administrator's: every ethernet route is
	// behind RequireRole(RoleAdmin), and a change must not be ratified through
	// a session that could not have made it. This runs before RequireRole, so
	// the role is read here rather than left to the route.
	if user.Role == authn.RoleAdmin {
		noteSignedInAdmin(c)
	}

	return Principal{Username: user.Username, Role: user.Role}, token, true
}

func abortUnauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, "unauthorized")
	c.Abort()
}

func abortForbiddenOrigin(c *gin.Context) {
	log.Debugf("rejected request from origin %s to host %s", c.Request.Header.Get("Origin"), c.Request.Host)

	c.JSON(http.StatusForbidden, "forbidden origin")
	c.Abort()
}

func GenerateJWT(username string, tokenVersion uint64) (string, error) {
	conf := config.GetInstance()
	now := time.Now()
	expireDuration := time.Duration(conf.JWT.RefreshTokenDuration) * time.Second
	claims := Token{
		Username:     username,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expireDuration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(conf.JWT.SecretKey))
}

func ParseJWT(jwtToken string) (*Token, error) {
	conf := config.GetInstance()
	parsed, err := jwt.ParseWithClaims(
		jwtToken,
		&Token{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(conf.JWT.SecretKey), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		log.Debugf("parse jwt error: %s", err)
		return nil, err
	}
	claims, ok := parsed.Claims.(*Token)
	if !ok || !parsed.Valid || claims.Username == "" || claims.Subject != claims.Username || claims.TokenVersion == 0 {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}
