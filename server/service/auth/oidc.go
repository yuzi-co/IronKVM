package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/audit"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// Login through an OpenID Connect provider: the authorization code flow
// with PKCE, state and nonce. The board is the confidential client (or a
// public one, without a secret, where the IdP allows it).
//
//  1. GET /api/auth/oidc/login makes a state, a nonce and a PKCE verifier,
//     keeps them in memory under the state, puts the state in a cookie that
//     binds the flow to this browser, and sends the browser to the IdP.
//  2. GET /api/auth/oidc/callback takes the code and state back. The state
//     must match the cookie and a flow in memory, which is used up. The code
//     is exchanged with the verifier, and the ID token is verified: the
//     signature against the IdP's key set, iss, aud, exp and the nonce.
//  3. The token's claims name a local account (see mapOIDCUser). The
//     account must be linked to the token's issuer and sub (see
//     resolveOIDCAccount). A login records a login session like the password
//     login does, and is audited.
//  4. An account with no link yet is refused, and the sign-in waits for an
//     administrator to link it or for the user to confirm it with the
//     account's password (see oidc_link.go).
//
// The IdP does the second factor: an OIDC login does not ask for the local
// TOTP code.

const (
	oidcFlowCookie   = "nano-kvm-oidc"
	oidcCallbackPath = "/api/auth/oidc/callback"
	oidcFlowTTL      = 10 * time.Minute
	maxOIDCFlows     = 128
	oidcHTTPTimeout  = 15 * time.Second
	// oidcProviderTTL is how long the discovery document is trusted before
	// it is fetched again. The key set refreshes itself when it meets a key
	// it does not know.
	oidcProviderTTL = time.Hour
)

// The reasons a login is refused, as the login page receives them in
// ?oidc=. Each one has a message there.
const (
	oidcErrDisabled = "disabled"
	oidcErrLocked   = "locked"
	oidcErrProvider = "provider"
	oidcErrDenied   = "denied"
	oidcErrState    = "state"
	oidcErrToken    = "token"
	oidcErrAccount  = "account"
	oidcErrRole     = "role"
	oidcErrServer   = "server"
	// oidcErrBinding is an account linked to another account at the IdP:
	// the username matched, the sub did not.
	oidcErrBinding = "binding"
	// oidcErrUnlinked is an account with no link yet. An administrator links
	// it, or the user confirms the link with the account's password.
	oidcErrUnlinked = "unlinked"
)

// oidcFailureDelay slows a refused callback down, as a wrong password is.
var oidcFailureDelay = 2 * time.Second

type oidcFlow struct {
	verifier    string
	nonce       string
	redirectURL string
	issuer      string
	clientID    string
	expires     time.Time
}

type oidcProvider struct {
	issuer  string
	fetched time.Time
	p       *oidc.Provider
}

var (
	oidcMu         sync.Mutex
	oidcFlows      = map[string]*oidcFlow{}
	cachedProvider *oidcProvider
)

// resetOIDCState forgets the flows in progress and the cached provider. A
// change of settings calls it, and tests.
func resetOIDCState() {
	oidcMu.Lock()
	defer oidcMu.Unlock()
	oidcFlows = map[string]*oidcFlow{}
	cachedProvider = nil
	resetOIDCLinks()
}

func oidcHTTPClient() *http.Client {
	return &http.Client{Timeout: oidcHTTPTimeout}
}

func oidcContext(ctx context.Context) context.Context {
	client := oidcHTTPClient()
	ctx = oidc.ClientContext(ctx, client)
	return context.WithValue(ctx, oauth2.HTTPClient, client)
}

// provider returns the IdP's discovery, fetched at most once an hour.
func provider(ctx context.Context, issuer string) (*oidc.Provider, error) {
	oidcMu.Lock()
	cached := cachedProvider
	oidcMu.Unlock()
	if cached != nil && cached.issuer == issuer && time.Since(cached.fetched) < oidcProviderTTL {
		return cached.p, nil
	}
	// The provider keeps the context's HTTP client for its key set, not the
	// deadline, so the timeout here bounds only the discovery.
	ctx, cancel := context.WithTimeout(oidcContext(ctx), oidcHTTPTimeout)
	defer cancel()
	p, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}
	oidcMu.Lock()
	cachedProvider = &oidcProvider{issuer: issuer, fetched: time.Now(), p: p}
	oidcMu.Unlock()
	return p, nil
}

func randomString(n int) (string, error) {
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// putFlow keeps a flow, dropping expired ones and, when full, the oldest.
func putFlow(state string, flow *oidcFlow) {
	oidcMu.Lock()
	defer oidcMu.Unlock()
	now := time.Now()
	var oldest string
	for k, f := range oidcFlows {
		if now.After(f.expires) {
			delete(oidcFlows, k)
			continue
		}
		if oldest == "" || f.expires.Before(oidcFlows[oldest].expires) {
			oldest = k
		}
	}
	if len(oidcFlows) >= maxOIDCFlows && oldest != "" {
		delete(oidcFlows, oldest)
	}
	oidcFlows[state] = flow
}

// takeFlow returns the flow of a state and forgets it: a state is good once.
func takeFlow(state string) (*oidcFlow, bool) {
	oidcMu.Lock()
	defer oidcMu.Unlock()
	flow, ok := oidcFlows[state]
	if !ok {
		return nil, false
	}
	delete(oidcFlows, state)
	if time.Now().After(flow.expires) {
		return nil, false
	}
	return flow, true
}

// redirectURL is the callback address the IdP sends the browser back to: the
// one in the settings, or else the board's as the browser reached it. It has
// to be registered at the IdP exactly.
func redirectURL(settings *oidcSettings, c *gin.Context) string {
	if settings.RedirectURL != "" {
		return settings.RedirectURL
	}
	scheme := "http"
	if middleware.RequestIsSecure(c) {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host + oidcCallbackPath
}

func toLoginPage(c *gin.Context, reason string) {
	c.Redirect(http.StatusFound, "/#/auth/login?oidc="+reason)
}

func setFlowCookie(c *gin.Context, value string, maxAge int) {
	// Lax, not Strict: the callback is a navigation from the IdP's site, and
	// a Strict cookie would not come with it.
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oidcFlowCookie, value, maxAge, "/api/auth/oidc", "", middleware.RequestIsSecure(c), true)
}

// OIDCInfo tells the login page whether to show the provider's button. It is
// public, and carries nothing about the provider but its name.
func (s *Service) OIDCInfo(c *gin.Context) {
	var rsp proto.Response
	settings, err := loadOIDCSettings()
	enabled := err == nil && settings.usable() && config.GetInstance().Authentication != "disable"
	info := oidcInfo{Enabled: enabled}
	if enabled {
		info.ProviderName = settings.ProviderName
		info.HidePasswordLogin = settings.HidePasswordLogin
	}
	rsp.OkRspWithData(c, info)
}

// OIDCLogin starts a login at the IdP.
func (s *Service) OIDCLogin(c *gin.Context) {
	settings, err := loadOIDCSettings()
	if err != nil || !settings.usable() || config.GetInstance().Authentication == "disable" {
		toLoginPage(c, oidcErrDisabled)
		return
	}
	p, err := provider(c.Request.Context(), settings.Issuer)
	if err != nil {
		log.Warnf("oidc discovery at %s: %s", settings.Issuer, err)
		s.oidcRefused(c, "", oidcErrProvider, "discovery failed", false)
		return
	}
	state, err1 := randomString(32)
	nonce, err2 := randomString(32)
	if err1 != nil || err2 != nil {
		toLoginPage(c, oidcErrServer)
		return
	}
	verifier := oauth2.GenerateVerifier()
	flow := &oidcFlow{
		verifier:    verifier,
		nonce:       nonce,
		redirectURL: redirectURL(&settings, c),
		issuer:      settings.Issuer,
		clientID:    settings.ClientID,
		expires:     time.Now().Add(oidcFlowTTL),
	}
	putFlow(state, flow)
	setFlowCookie(c, state, int(oidcFlowTTL/time.Second))

	conf := oauthConfig(&settings, p, flow.redirectURL)
	c.Redirect(http.StatusFound, conf.AuthCodeURL(state, oidc.Nonce(nonce), oauth2.S256ChallengeOption(verifier)))
}

func oauthConfig(settings *oidcSettings, p *oidc.Provider, redirect string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     settings.ClientID,
		ClientSecret: settings.ClientSecret,
		Endpoint:     p.Endpoint(),
		RedirectURL:  redirect,
		Scopes:       settings.Scopes,
	}
}

// OIDCCallback finishes a login the IdP sent the browser back from.
func (s *Service) OIDCCallback(c *gin.Context) {
	clientIP := GetClientIP(c)
	// The flow cookie is spent whatever happens next.
	setFlowCookie(c, "", -1)

	settings, err := loadOIDCSettings()
	if err != nil || !settings.usable() || config.GetInstance().Authentication == "disable" {
		toLoginPage(c, oidcErrDisabled)
		return
	}
	if locked, _, _ := CheckLoginAttempt(clientIP); locked {
		s.oidcRefused(c, "", oidcErrLocked, "too many failed attempts", false)
		return
	}

	state := c.Query("state")
	cookie, _ := c.Cookie(oidcFlowCookie)
	if state == "" || cookie == "" || subtle.ConstantTimeCompare([]byte(state), []byte(cookie)) != 1 {
		s.oidcRefused(c, "", oidcErrState, "state does not match this browser", true)
		return
	}
	flow, ok := takeFlow(state)
	if !ok || flow.issuer != settings.Issuer || flow.clientID != settings.ClientID {
		s.oidcRefused(c, "", oidcErrState, "unknown or expired state", true)
		return
	}
	if idpErr := c.Query("error"); idpErr != "" {
		// The user cancelled at the IdP, or the IdP refused them. Not a
		// guess at anything, so not counted against the address.
		s.oidcRefused(c, "", oidcErrDenied, "provider answered "+clip(idpErr, 64), false)
		return
	}

	claims, err := s.exchange(c.Request.Context(), &settings, flow, c.Query("code"))
	if err != nil {
		log.Infof("oidc login refused: %s", err)
		s.oidcRefused(c, "", oidcErrToken, err.Error(), true)
		return
	}

	identity, err := identityOf(claims)
	if err != nil {
		s.oidcRefused(c, "", oidcErrToken, err.Error(), true)
		return
	}
	username, groupRole, reason, detail := mapOIDCUser(&settings, claims)
	if reason != "" {
		s.oidcRefused(c, username, reason, detail, true)
		return
	}
	user, link, reason, detail := resolveOIDCAccount(&settings, username, groupRole, identity)
	if link != "" {
		auditOIDCLink(c, username, link)
	}
	if reason == oidcErrUnlinked {
		s.oidcUnlinked(c, user, groupRole, identity, claims)
		return
	}
	if reason != "" {
		s.oidcRefused(c, username, reason, detail, true)
		return
	}

	token, err := middleware.IssueSession(c, user.Username, user.TokenVersion)
	if err != nil {
		log.Errorf("oidc login for %s: issue session: %s", user.Username, err)
		s.oidcRefused(c, user.Username, oidcErrServer, "could not record the session", false)
		return
	}
	ClearLoginAttempt(clientIP)
	setSessionCookie(c, token)
	audit.Request(c, audit.Event{
		Actor:    user.Username,
		Via:      "oidc",
		Role:     string(user.Role),
		Category: audit.CategoryAuth,
		Action:   "login-oidc",
		Outcome:  audit.OutcomeOK,
		Detail:   clip(settings.Issuer, 128),
	})
	log.Infof("user logged in through oidc: %s", user.Username)
	c.Redirect(http.StatusFound, "/")
}

// exchange trades the code for tokens and verifies the ID token, returning
// its claims.
func (s *Service) exchange(ctx context.Context, settings *oidcSettings, flow *oidcFlow, code string) (map[string]any, error) {
	if code == "" {
		return nil, errors.New("no code")
	}
	p, err := provider(ctx, settings.Issuer)
	if err != nil {
		return nil, fmt.Errorf("discovery: %w", err)
	}
	ctx, cancel := context.WithTimeout(oidcContext(ctx), oidcHTTPTimeout)
	defer cancel()

	token, err := oauthConfig(settings, p, flow.redirectURL).Exchange(ctx, code, oauth2.VerifierOption(flow.verifier))
	if err != nil {
		return nil, fmt.Errorf("code exchange: %w", err)
	}
	raw, ok := token.Extra("id_token").(string)
	if !ok || raw == "" {
		return nil, errors.New("no id_token in the token response")
	}
	// Verify checks the signature against the IdP's keys, the issuer, the
	// audience (our client ID) and the expiry.
	idToken, err := p.Verifier(&oidc.Config{ClientID: settings.ClientID}).Verify(ctx, raw)
	if err != nil {
		return nil, fmt.Errorf("id token: %w", err)
	}
	if idToken.Nonce == "" || subtle.ConstantTimeCompare([]byte(idToken.Nonce), []byte(flow.nonce)) != 1 {
		return nil, errors.New("id token: nonce does not match")
	}
	claims := map[string]any{}
	if err = idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("id token claims: %w", err)
	}
	return claims, nil
}

// oidcRefused sends the browser back to the login page with the reason,
// audits the refusal, and counts it against the address when it could be
// someone trying their luck.
func (s *Service) oidcRefused(c *gin.Context, actor, reason, detail string, count bool) {
	if count {
		time.Sleep(oidcFailureDelay)
		RecordLoginFailure(GetClientIP(c))
	}
	outcome := audit.OutcomeFail
	switch reason {
	case oidcErrAccount, oidcErrRole, oidcErrLocked, oidcErrDenied, oidcErrBinding, oidcErrUnlinked:
		outcome = audit.OutcomeDenied
	}
	audit.Request(c, audit.Event{
		Actor:    actor,
		Via:      "oidc",
		Category: audit.CategoryAuth,
		Action:   "login-oidc",
		Outcome:  outcome,
		Detail:   clip(reason+": "+detail, 200),
	})
	toLoginPage(c, reason)
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// oidcIdentity is the IdP account a login came from: the ID token's issuer
// and sub. The pair is what an account is linked to.
type oidcIdentity struct {
	issuer  string
	subject string
}

// identityOf reads the issuer and sub of a verified ID token. Both are
// required claims, and the verification has already matched the issuer to
// the one in the settings.
func identityOf(claims map[string]any) (oidcIdentity, error) {
	issuer, _ := claims["iss"].(string)
	subject, _ := claims["sub"].(string)
	if issuer == "" || subject == "" {
		return oidcIdentity{}, errors.New("id token without iss or sub")
	}
	return oidcIdentity{issuer: issuer, subject: subject}, nil
}

// auditOIDCLink records that a login linked an account to an IdP account.
// detail says to which issuer, and the one it replaced, if any.
func auditOIDCLink(c *gin.Context, username, detail string) {
	audit.Request(c, audit.Event{
		Actor:    username,
		Via:      "oidc",
		Category: audit.CategoryAuth,
		Action:   "oidc-link",
		Target:   username,
		Outcome:  audit.OutcomeOK,
		Detail:   clip(detail, 200),
	})
	log.Infof("account %s %s", username, detail)
}

// mapOIDCUser reads the local username, and the role the groups give if
// they are mapped, from the ID token's claims. reason is set when the claims
// name no usable account.
func mapOIDCUser(settings *oidcSettings, claims map[string]any) (username string, groupRole authn.Role, reason, detail string) {
	value, _ := claims[settings.UsernameClaim].(string)
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "", oidcErrAccount, "no " + settings.UsernameClaim + " claim"
	}

	username = value
	if settings.UsernameClaim == claimEmail {
		// The address must be the IdP's verified one, and in the domain
		// whose local parts are this board's usernames. An IdP that does not
		// say whether it verified the address is taken at its word.
		if verified, present := claims["email_verified"]; present {
			if b, ok := verified.(bool); !ok || !b {
				return clip(value, 64), "", oidcErrAccount, "email not verified"
			}
		}
		local, domain, found := strings.Cut(value, "@")
		if !found || !strings.EqualFold(domain, settings.EmailDomain) {
			return clip(value, 64), "", oidcErrAccount, "email outside " + settings.EmailDomain
		}
		username = local
	}
	if err := authn.ValidateUsername(username); err != nil {
		return clip(value, 64), "", oidcErrAccount, "not a valid local username"
	}

	if settings.groupsMapped() {
		groupRole = roleFromGroups(settings, claims[settings.GroupsClaim])
		if groupRole == "" {
			return username, "", oidcErrRole, "in none of the mapped groups"
		}
	}
	return username, groupRole, "", ""
}

// roleFromGroups returns the highest role whose group the user is in.
func roleFromGroups(settings *oidcSettings, claim any) authn.Role {
	var groups []string
	switch v := claim.(type) {
	case string:
		groups = []string{v}
	case []any:
		for _, g := range v {
			if s, ok := g.(string); ok {
				groups = append(groups, s)
			}
		}
	}
	in := func(name string) bool { return name != "" && contains(groups, name) }
	switch {
	case in(settings.AdminGroup):
		return authn.RoleAdmin
	case in(settings.UserGroup):
		return authn.RoleUser
	case in(settings.ViewerGroup):
		return authn.RoleViewer
	}
	return ""
}

// resolveOIDCAccount finds, or makes, the local account a login is for,
// checks or makes its link to the IdP account, and sets its role from the
// groups when they are mapped. link describes a link it made, for the audit
// log, and is empty when the account was already linked to this identity.
//
// The username alone trusts the IdP not to let a user rename themselves into
// another user's name, which some providers allow. The link closes that: an
// account is linked to an issuer and sub by an administrator, by its own
// password, or by the login that created it, and a login with the same issuer
// and another sub is refused. An account with no link is refused with
// oidcErrUnlinked, and user is set so that the caller can offer the password
// confirmation (ironkvm-dist#169). A link made under
// another issuer is replaced (see authn.Store.BindOIDC), so changing the
// issuer in the settings links every account again at its next login rather
// than locking them all out.
func resolveOIDCAccount(settings *oidcSettings, username string, groupRole authn.Role, identity oidcIdentity) (user *authn.User, link, reason, detail string) {
	user, err := authn.DefaultStore.Get(username)
	if errors.Is(err, authn.ErrUserNotFound) {
		if !settings.AutoCreate {
			return nil, "", oidcErrAccount, "no local account"
		}
		role := settings.DefaultRole
		if groupRole != "" {
			role = groupRole
		}
		// Nobody knows the password; the account signs in through the IdP
		// until an administrator sets one.
		password, err := randomString(32)
		if err != nil {
			return nil, "", oidcErrServer, "random"
		}
		// The account is linked as it is made. When another login made it
		// first, the link check below decides as for any existing account.
		err = authn.DefaultStore.CreateOIDC(username, password, role, identity.issuer, identity.subject)
		switch {
		case err == nil:
			log.Infof("account %s created for an oidc login, role %s", username, role)
			link = "created and linked to " + identity.issuer
		case !errors.Is(err, authn.ErrUserExists):
			return nil, "", oidcErrServer, "create account: " + err.Error()
		}
		if user, err = authn.DefaultStore.Get(username); err != nil {
			return nil, link, oidcErrServer, "read account"
		}
	} else if err != nil {
		return nil, "", oidcErrServer, "read account: " + err.Error()
	}

	// The device owner's account also holds the root password; it signs in
	// with that password only, so an IdP account named like it gets nothing.
	if user.SystemAccount {
		return nil, link, oidcErrAccount, "the device owner account signs in with its password only"
	}
	if !user.Enabled {
		return nil, link, oidcErrAccount, "account disabled"
	}

	// The link is checked before the groups touch the role, so an IdP
	// account that is not this one cannot change it either.
	bound, previous, err := authn.DefaultStore.BindOIDC(user.Username, identity.issuer, identity.subject)
	switch {
	case errors.Is(err, authn.ErrOIDCNotLinked):
		return user, link, oidcErrUnlinked, "not linked to an identity provider account yet"
	case errors.Is(err, authn.ErrOIDCSubjectMismatch):
		return nil, link, oidcErrBinding, "linked to another account at " + identity.issuer + ", not sub " + clip(identity.subject, 64)
	case err != nil:
		return nil, link, oidcErrServer, "link account: " + err.Error()
	case previous == "":
		link = "linked to " + identity.issuer
	case previous != identity.issuer:
		link = "linked to " + identity.issuer + " in place of " + previous
	}
	user, err = applyGroupRole(bound, groupRole)
	if err != nil {
		return nil, link, oidcErrRole, "set role from groups: " + err.Error()
	}
	return user, link, "", ""
}

// applyGroupRole gives a linked account the role its IdP groups map to, when
// they are mapped and it differs.
func applyGroupRole(user *authn.User, groupRole authn.Role) (*authn.User, error) {
	if groupRole == "" || groupRole == user.Role {
		return user, nil
	}
	updated, err := authn.DefaultStore.Update("", user.Username, authn.UserPatch{Role: &groupRole})
	if err != nil {
		return nil, err
	}
	// The role changed: the account's other sessions end, as when an
	// administrator changes it.
	middleware.RevokeUserSessions(user.Username)
	return updated, nil
}
