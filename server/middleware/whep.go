package middleware

import (
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/service/apikey"
)

// WHEPViewer is who opened a WHEP session (ironkvm-dist#145), and how to ask
// whether the credential still holds while the session lasts.
type WHEPViewer struct {
	Principal Principal
	// Valid reports whether the credential is still good: the key not
	// revoked and its account still enabled, the share link not revoked or
	// expired.
	Valid func() bool
}

// AuthenticateWHEP checks the credential of a WHEP request. It returns 0 when
// the request may go on, or the status to refuse it with: 401 without a
// credential this server knows, 403 for a credential that could do more than
// watch.
//
// The only credential is "Authorization: Bearer", as WHEP players send it
// (RFC 9725, section 4.7.1): a view-only API key, the key of a view-only
// account, or the token of a share link. The session cookie is never read.
// The endpoint answers other origins, so a cookie it took would let any page a
// signed-in user visits pull the video; the origin check the rest of the API
// relies on is not applied here for the same reason.
//
// A key with the run of its account is refused even though it could watch the
// video through the rest of the API: a WHEP token ends up in an OBS profile,
// a shell history or a player's URL bar, and one that can also type on the
// host must not.
func AuthenticateWHEP(r *http.Request) (WHEPViewer, int) {
	authorization := r.Header.Get("Authorization")
	token := ""
	if strings.HasPrefix(authorization, bearerPrefix) {
		token = strings.TrimSpace(strings.TrimPrefix(authorization, bearerPrefix))
	}

	return AuthenticateViewToken(token)
}

// AuthenticateViewToken checks a token that may only watch, however it
// arrived: WHEP's Bearer header, the password of an RTSP request
// (ironkvm-dist#142) or the stream ID of an SRT caller (ironkvm-dist#143).
// Each of those ends up in a player's settings or URL, so the rules are
// WHEP's: a view-only API key, the key of a view-only account, or a share
// link token, and never a key that could also type on the host. It returns
// 0, 401 or 403 as AuthenticateWHEP does.
func AuthenticateViewToken(token string) (WHEPViewer, int) {
	if config.GetInstance().Authentication == "disable" {
		// The whole board is open; there is nobody to check.
		return WHEPViewer{
			Principal: Principal{Username: "admin", Role: authn.RoleViewer, Via: "open"},
			Valid:     func() bool { return true },
		}, 0
	}

	if token == "" {
		return WHEPViewer{}, http.StatusUnauthorized
	}

	if strings.HasPrefix(token, apikey.SecretPrefix) {
		return whepByKey(token)
	}
	return whepByShare(token)
}

// whepByKey accepts a key that can only watch.
func whepByKey(secret string) (WHEPViewer, int) {
	principal, ok := whepKeyPrincipal(secret)
	if !ok {
		return WHEPViewer{}, http.StatusUnauthorized
	}
	if principal.Role != authn.RoleViewer {
		log.Debugf("whep: refused key %q of %s, which can do more than watch", principal.Key, principal.Username)
		return WHEPViewer{Principal: principal}, http.StatusForbidden
	}

	return WHEPViewer{
		Principal: principal,
		Valid: func() bool {
			again, ok := whepKeyPrincipal(secret)
			return ok && again.Role == authn.RoleViewer
		},
	}, 0
}

// whepKeyPrincipal is authenticateByAPIKey for a secret taken from the WHEP
// request's Bearer header.
func whepKeyPrincipal(secret string) (Principal, bool) {
	request := &http.Request{Header: http.Header{}}
	request.Header.Set(apiKeyHeader, secret)
	return authenticateByAPIKey(request)
}

// whepByShare accepts the token of a share link, the same token
// POST /api/auth/share redeems. It is always a viewer.
func whepByShare(token string) (WHEPViewer, int) {
	link, ok := authn.DefaultShares.Redeem(token)
	if !ok {
		return WHEPViewer{}, http.StatusUnauthorized
	}
	id := link.ID

	return WHEPViewer{
		Principal: Principal{
			Username: authn.SharePrincipalPrefix + id,
			Role:     authn.RoleViewer,
			ShareID:  id,
			Via:      "share",
		},
		Valid: func() bool {
			_, ok := authn.DefaultShares.Valid(id)
			return ok
		},
	}, 0
}
