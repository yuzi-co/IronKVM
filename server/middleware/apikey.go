package middleware

import (
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/service/apikey"
)

const (
	apiKeyHeader = "X-API-Key"
	bearerPrefix = "Bearer "
)

// authenticateByAPIKey authenticates a caller that holds a key instead of a
// login session. Scripts have no browser to keep a cookie in, and the
// alternative people reached for was switching authentication off entirely.
//
// The key acts as the user it was issued to, and with that user's role. It
// carries no authority of its own: an account that is disabled or deleted
// takes its keys with it, and a key issued by an operator who is not an
// administrator does not reach the routes that ask for one.
//
// The origin rule is applied to these requests exactly as it is to session
// ones. A key travels in a header a cross-site form cannot set, so it is not
// reachable by the attack the rule exists for, but nothing is gained by
// carving out an exception and a second path through the check is a second
// place for it to go wrong.
func authenticateByAPIKey(r *http.Request) (Principal, bool) {
	secret := presentedAPIKey(r)
	if secret == "" {
		return Principal{}, false
	}

	key, ok := apikey.Verify(secret)
	if !ok {
		return Principal{}, false
	}

	if key.Username == "" {
		log.Warnf("api key %s belongs to no account; issue a new one", key.ID)
		return Principal{}, false
	}

	user, err := authn.DefaultStore.Get(key.Username)
	if err != nil {
		log.Debugf("api key %s names an account that is gone: %s", key.ID, err)
		return Principal{}, false
	}

	if !user.Enabled {
		log.Debugf("api key %s belongs to a disabled account", key.ID)
		return Principal{}, false
	}

	return Principal{Username: user.Username, Role: user.Role}, true
}

// presentedAPIKey pulls a key out of either header clients expect to use.
func presentedAPIKey(r *http.Request) string {
	if key := strings.TrimSpace(r.Header.Get(apiKeyHeader)); key != "" {
		return key
	}

	authorization := r.Header.Get("Authorization")
	if !strings.HasPrefix(authorization, bearerPrefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(authorization, bearerPrefix))
}
