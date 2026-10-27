package auth

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/utils"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Linking an account to an identity provider account before its first
// single sign-on (ironkvm-dist#169).
//
// A single sign-on to an account with no link is refused: the username alone
// does not say who is behind the IdP account, and some providers let a user
// rename themselves. The account is linked in one of two ways:
//
//   - An administrator links it, entering the issuer and sub or picking a
//     refused sign-in from the pending list, which keeps the last few.
//   - The user confirms it. The refused callback leaves a pending link behind
//     a cookie for a few minutes, and the login page asks for the account's
//     password, and its TOTP code when it has one. The right answer links the
//     account and signs the user in. An account whose password nobody knows,
//     made by a single sign-on, waits for an administrator.

const (
	oidcLinkCookie = "nano-kvm-oidc-link"
	oidcLinkTTL    = 5 * time.Minute
	maxOIDCLinks   = 64
	// oidcLinkAttempts is how many wrong passwords a pending link takes
	// before it is dropped. The address's login failures count too.
	oidcLinkAttempts = 5

	maxPendingSignIns = 20
	pendingSignInTTL  = 24 * time.Hour
	// maxClaimLength clips the claims a pending sign-in shows.
	maxClaimLength = 128
)

// The answers of the confirmation step beyond -1 (bad request), -2 (wrong
// password or code) and -3 (server). A locked address answers with the code
// of the login's rate limit.
const (
	codeOIDCLinkLocked = -5
	codeOIDCLinkGone   = -7
	codeOIDCLinkTaken  = -8
)

// pendingLink is a refused single sign-on that the account's password can
// still turn into a link.
type pendingLink struct {
	username     string
	issuer       string
	subject      string
	groupRole    authn.Role
	tokenVersion uint64
	expires      time.Time
	attempts     int
}

// pendingSignIn is a refused single sign-on to an account with no link, as
// an administrator sees it to link the account.
type pendingSignIn struct {
	Username string `json:"username"`
	Issuer   string `json:"issuer"`
	Subject  string `json:"subject"`
	Email    string `json:"email,omitempty"`
	Name     string `json:"name,omitempty"`
	// Time is when the sign-in was refused, in Unix seconds.
	Time int64 `json:"time"`
}

var (
	oidcLinkMu sync.Mutex
	oidcLinks  = map[string]*pendingLink{}
	// pendingSignIns are kept in memory, oldest first. A restart forgets
	// them; the user signs in again to be listed.
	pendingSignIns []pendingSignIn
	// oidcNow is the clock of pending links and sign-ins. Tests move it.
	oidcNow = time.Now
)

// resetOIDCLinks forgets the pending links: they were made under settings
// that changed. The pending sign-ins stay, each with its issuer.
func resetOIDCLinks() {
	oidcLinkMu.Lock()
	defer oidcLinkMu.Unlock()
	oidcLinks = map[string]*pendingLink{}
}

// putLink keeps a pending link, dropping expired ones and, when full, the
// oldest.
func putLink(token string, link *pendingLink) {
	oidcLinkMu.Lock()
	defer oidcLinkMu.Unlock()
	now := oidcNow()
	var oldest string
	for k, l := range oidcLinks {
		if now.After(l.expires) {
			delete(oidcLinks, k)
			continue
		}
		if oldest == "" || l.expires.Before(oidcLinks[oldest].expires) {
			oldest = k
		}
	}
	if len(oidcLinks) >= maxOIDCLinks && oldest != "" {
		delete(oidcLinks, oldest)
	}
	oidcLinks[token] = link
}

// lookupLink returns a copy of the pending link of a token.
func lookupLink(token string) (pendingLink, bool) {
	oidcLinkMu.Lock()
	defer oidcLinkMu.Unlock()
	link, ok := oidcLinks[token]
	if !ok || token == "" {
		return pendingLink{}, false
	}
	if oidcNow().After(link.expires) {
		delete(oidcLinks, token)
		return pendingLink{}, false
	}
	return *link, true
}

// failLink counts a wrong password against a pending link, and drops the
// link when it has taken enough.
func failLink(token string) {
	oidcLinkMu.Lock()
	defer oidcLinkMu.Unlock()
	if link, ok := oidcLinks[token]; ok {
		link.attempts++
		if link.attempts >= oidcLinkAttempts {
			delete(oidcLinks, token)
		}
	}
}

func dropLink(token string) {
	oidcLinkMu.Lock()
	defer oidcLinkMu.Unlock()
	delete(oidcLinks, token)
}

// pruneSignInsLocked drops the pending sign-ins that are too old.
func pruneSignInsLocked(now time.Time) {
	kept := pendingSignIns[:0]
	for _, p := range pendingSignIns {
		if now.Sub(time.Unix(p.Time, 0)) < pendingSignInTTL {
			kept = append(kept, p)
		}
	}
	pendingSignIns = kept
}

// recordPendingSignIn adds a refused sign-in to the list. The same IdP
// account for the same username is listed once, at its latest attempt, and
// the list keeps the newest maxPendingSignIns.
func recordPendingSignIn(p pendingSignIn) {
	oidcLinkMu.Lock()
	defer oidcLinkMu.Unlock()
	now := oidcNow()
	p.Time = now.Unix()
	pruneSignInsLocked(now)
	kept := pendingSignIns[:0]
	for _, q := range pendingSignIns {
		if q.Username != p.Username || q.Issuer != p.Issuer || q.Subject != p.Subject {
			kept = append(kept, q)
		}
	}
	pendingSignIns = append(kept, p)
	if extra := len(pendingSignIns) - maxPendingSignIns; extra > 0 {
		pendingSignIns = append([]pendingSignIn(nil), pendingSignIns[extra:]...)
	}
}

// listPendingSignIns returns the pending sign-ins, newest first.
func listPendingSignIns() []pendingSignIn {
	oidcLinkMu.Lock()
	defer oidcLinkMu.Unlock()
	pruneSignInsLocked(oidcNow())
	out := append([]pendingSignIn(nil), pendingSignIns...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Time > out[j].Time })
	return out
}

// forgetPendingSignIns drops the pending sign-ins of an IdP account that is
// now linked.
func forgetPendingSignIns(issuer, subject string) {
	oidcLinkMu.Lock()
	defer oidcLinkMu.Unlock()
	kept := pendingSignIns[:0]
	for _, p := range pendingSignIns {
		if p.Issuer != issuer || p.Subject != subject {
			kept = append(kept, p)
		}
	}
	pendingSignIns = kept
}

func setLinkCookie(c *gin.Context, value string, maxAge int) {
	// Lax, as the flow cookie: it is set on the callback, a navigation from
	// the IdP's site, and read by the login page's own requests.
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oidcLinkCookie, value, maxAge, "/api/auth/oidc", "", middleware.RequestIsSecure(c), true)
}

func claimText(claims map[string]any, name string) string {
	value, _ := claims[name].(string)
	return clip(strings.TrimSpace(value), maxClaimLength)
}

// oidcUnlinked refuses a single sign-on to an account with no link. The
// sign-in goes on the administrator's pending list, and, when the account has
// a password of its own, waits for it behind a cookie.
func (s *Service) oidcUnlinked(c *gin.Context, user *authn.User, groupRole authn.Role, identity oidcIdentity, claims map[string]any) {
	recordPendingSignIn(pendingSignIn{
		Username: user.Username,
		Issuer:   identity.issuer,
		Subject:  identity.subject,
		Email:    claimText(claims, "email"),
		Name:     claimText(claims, "name"),
	})
	detail := "not linked yet, sub " + clip(identity.subject, 64)
	if !user.NoPassword {
		token, err := randomString(32)
		if err == nil {
			putLink(token, &pendingLink{
				username:     user.Username,
				issuer:       identity.issuer,
				subject:      identity.subject,
				groupRole:    groupRole,
				tokenVersion: user.TokenVersion,
				expires:      oidcNow().Add(oidcLinkTTL),
			})
			setLinkCookie(c, token, int(oidcLinkTTL/time.Second))
			detail += ", the password may confirm it"
		}
	}
	// Not counted against the address: the IdP vouched for the sign-in, and
	// only the link is missing.
	s.oidcRefused(c, user.Username, oidcErrUnlinked, detail, false)
}

// oidcLinkInfo is what the login page learns about the sign-in that waits
// for a password.
type oidcLinkInfo struct {
	Username     string `json:"username"`
	TOTPRequired bool   `json:"totpRequired"`
}

// oidcLinkReq confirms a pending link. Password is encrypted as the login
// form sends it; Code is a TOTP or recovery code, for an account with TOTP.
type oidcLinkReq struct {
	Password string `json:"password"`
	Code     string `json:"code"`
}

// GetOIDCLink tells the login page whether a refused single sign-on in this
// browser waits for the account's password, and for which account.
func (s *Service) GetOIDCLink(c *gin.Context) {
	var rsp proto.Response
	token, _ := c.Cookie(oidcLinkCookie)
	link, ok := lookupLink(token)
	if !ok {
		rsp.ErrRsp(c, codeOIDCLinkGone, "no sign-in waits for a password")
		return
	}
	user, err := authn.DefaultStore.Get(link.username)
	rsp.OkRspWithData(c, oidcLinkInfo{
		Username:     link.username,
		TOTPRequired: err == nil && user.TOTPEnabled(),
	})
}

// ConfirmOIDCLink links an account to the IdP account of a refused single
// sign-on when the account's password, and its second factor if it has one,
// are right, and signs the user in.
func (s *Service) ConfirmOIDCLink(c *gin.Context) {
	var rsp proto.Response
	var req oidcLinkReq

	clientIP := GetClientIP(c)
	if locked, code, msg := CheckLoginAttempt(clientIP); locked {
		time.Sleep(oidcFailureDelay)
		rsp.ErrRsp(c, code, msg)
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.ErrRsp(c, -1, "invalid parameters")
		return
	}
	settings, err := loadOIDCSettings()
	token, _ := c.Cookie(oidcLinkCookie)
	link, ok := lookupLink(token)
	disabled := err != nil || !settings.usable() || config.GetInstance().Authentication == "disable"
	if disabled || !ok || link.issuer != settings.Issuer {
		setLinkCookie(c, "", -1)
		rsp.ErrRsp(c, codeOIDCLinkGone, "the sign-in expired, sign in with the provider again")
		return
	}
	audit.SetActor(c, link.username)
	user, err := authn.DefaultStore.Get(link.username)
	if err != nil || !user.Enabled || user.SystemAccount || user.NoPassword || user.TokenVersion != link.tokenVersion {
		dropLink(token)
		setLinkCookie(c, "", -1)
		rsp.ErrRsp(c, codeOIDCLinkGone, "the sign-in expired, sign in with the provider again")
		return
	}

	password, err := utils.DecodeDecrypt(req.Password)
	passwordOK := false
	if err == nil && password != "" {
		_, passwordOK, err = authn.DefaultStore.Authenticate(user.Username, password)
		if err != nil {
			log.Errorf("check the password of %s for an oidc link: %s", user.Username, err)
			rsp.ErrRsp(c, -3, "authentication unavailable")
			return
		}
	}
	if !passwordOK {
		s.linkFailed(c, token, clientIP, &rsp, "invalid password")
		return
	}
	if user.TOTPEnabled() {
		if guard.locked(user.Username) {
			time.Sleep(oidcFailureDelay)
			rsp.ErrRsp(c, codeOIDCLinkLocked, "too many wrong codes, try again later")
			return
		}
		codeOK, err := checkSecondFactor(user, strings.TrimSpace(req.Code))
		if err != nil {
			log.Errorf("check the second factor of %s for an oidc link: %s", user.Username, err)
			rsp.ErrRsp(c, -3, "authentication unavailable")
			return
		}
		if !codeOK {
			if guard.fail("", user.Username) {
				dropLink(token)
			}
			s.linkFailed(c, token, clientIP, &rsp, "invalid code")
			return
		}
		guard.succeed(user.Username)
	}

	linked, err := authn.DefaultStore.ConfirmOIDC(user.Username, link.issuer, link.subject)
	dropLink(token)
	setLinkCookie(c, "", -1)
	if errors.Is(err, authn.ErrOIDCSubjectMismatch) {
		rsp.ErrRsp(c, codeOIDCLinkTaken, "the account was linked to another identity provider account meanwhile")
		return
	}
	if err != nil {
		log.Errorf("link %s to %s: %s", user.Username, link.issuer, err)
		rsp.ErrRsp(c, -3, "could not link the account")
		return
	}
	ClearLoginAttempt(clientIP)
	forgetPendingSignIns(link.issuer, link.subject)
	auditOIDCLink(c, user.Username, "linked to "+link.issuer+", confirmed with the account password")

	user, err = applyGroupRole(linked, link.groupRole)
	if err != nil {
		rsp.ErrRsp(c, -3, "could not set the role from the groups")
		return
	}
	session, err := middleware.IssueSession(c, user.Username, user.TokenVersion)
	if err != nil {
		log.Errorf("oidc link for %s: issue session: %s", user.Username, err)
		rsp.ErrRsp(c, -3, "could not record the session")
		return
	}
	setSessionCookie(c, session)
	audit.SetDetail(c, clip(link.issuer, 128))
	rsp.OkRsp(c)
	log.Infof("user linked to %s and logged in through oidc: %s", link.issuer, user.Username)
}

// linkFailed answers a wrong password or code: slowly, counted against the
// pending link and the address.
func (s *Service) linkFailed(c *gin.Context, token, clientIP string, rsp *proto.Response, msg string) {
	time.Sleep(oidcFailureDelay)
	failLink(token)
	if locked, code, lockMsg := RecordLoginFailure(clientIP); locked {
		rsp.ErrRsp(c, code, lockMsg)
		return
	}
	rsp.ErrRsp(c, -2, msg)
}

// oidcPendingRsp is the pending list, with the issuer a link is made under
// when the administrator enters only a sub.
type oidcPendingRsp struct {
	Issuer  string          `json:"issuer"`
	SignIns []pendingSignIn `json:"signIns"`
}

// GetOIDCPending lists the refused single sign-ons to accounts with no link,
// for an administrator to link one.
func (s *Service) GetOIDCPending(c *gin.Context) {
	var rsp proto.Response
	settings, _ := loadOIDCSettings()
	rsp.OkRspWithData(c, oidcPendingRsp{Issuer: settings.Issuer, SignIns: listPendingSignIns()})
}

// linkUserOIDCReq is an administrator's pre-link. An empty Issuer is the one
// in the settings.
type linkUserOIDCReq struct {
	Issuer  string `json:"issuer"`
	Subject string `json:"subject"`
}

// The sub is at most 255 ASCII characters (OpenID Connect Core 2), and the
// issuer a URL.
const (
	maxOIDCSubject = 255
	maxOIDCIssuer  = 512
)

// LinkUserOIDC is an administrator linking an account to an IdP account
// before its first single sign-on, or to another one. When it replaces
// another link, the account's sessions end as on an unlink; an administrator
// relinking their own account keeps the session they are doing it from.
func (s *Service) LinkUserOIDC(c *gin.Context) {
	var rsp proto.Response
	var req linkUserOIDCReq
	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.ErrRsp(c, -1, "invalid parameters")
		return
	}
	issuer, subject := strings.TrimSpace(req.Issuer), strings.TrimSpace(req.Subject)
	if issuer == "" {
		settings, _ := loadOIDCSettings()
		issuer = settings.Issuer
	}
	if issuer == "" || subject == "" || len(issuer) > maxOIDCIssuer || len(subject) > maxOIDCSubject {
		rsp.ErrRsp(c, -1, "an issuer and a subject are needed")
		return
	}
	username := c.Param("username")
	user, err := authn.DefaultStore.Get(username)
	if err != nil {
		rsp.ErrRsp(c, -2, err.Error())
		return
	}
	if user.SystemAccount {
		rsp.ErrRsp(c, -3, "the device owner account signs in with its password only")
		return
	}
	_, previousIssuer, previousSubject, err := authn.DefaultStore.LinkOIDC(username, issuer, subject)
	if err != nil {
		rsp.ErrRsp(c, -2, err.Error())
		return
	}
	forgetPendingSignIns(issuer, subject)

	principal, _ := middleware.CurrentPrincipal(c)
	if previousSubject != "" && (previousIssuer != issuer || previousSubject != subject) {
		if principal.Username == username {
			middleware.RevokeOtherSessions(username, principal.SessionID)
		} else {
			middleware.RevokeUserSessions(username)
		}
	}
	audit.SetDetail(c, clip("linked to "+issuer+" sub "+subject, 200))
	rsp.OkRsp(c)
	log.Infof("single sign-on link of %s set to %s by %s", username, issuer, principal.Username)
}
