package auth

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/authn/totp"
	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/utils"
)

// Two-factor login for the web UI.
//
// What it covers, and what it does not:
//
//   - The web login. A password that is right for an account with TOTP on
//     gets a challenge instead of a session, and only LoginTOTP turns the
//     challenge into a session cookie.
//   - API keys bypass it, as they bypass the password. A key is minted only
//     from a session (router/auth.go puts the key routes behind
//     CheckSession), so on an account with TOTP on every key was issued by a
//     login that passed the code step.
//   - VNC, IPMI and Redfish stay password only. Their protocols have no
//     place for a second factor, and PiKVM's TOTP covers its web UI alone
//     for the same reason. The settings page says so.
//
// The secret is sealed with authn.Keyring under TOTPKeyFile and kept in the
// account file, which is written 0600 by rename, the same as the password
// hashes. Recovery codes are kept as SHA-256 hashes. Neither the secret nor a
// code is ever logged.

// TOTPKeyFile holds the key TOTP secrets are sealed under. It sits in
// /etc/kvm beside the account file, which /data/identity backs on a board
// with slots, so the pair moves together.
const TOTPKeyFile = "/etc/kvm/totp.key"

// The second step's reply codes, beyond the ones Login already uses (-2 a
// wrong credential, -3 the store unavailable, -4 and -5 lockouts).
const (
	codeTOTPRequired    = -6
	codeTOTPChallengeGo = -7
)

// totpIssuer is the name an authenticator app files the entry under.
const totpIssuer = "IronKVM"

var totpKeyring = authn.NewKeyring(TOTPKeyFile)

// codeFailureDelay slows each wrong code, as loginFailed slows each wrong
// password. A variable so the tests do not sleep.
var codeFailureDelay = time.Second

// startTOTPChallenge answers a correct password for an account that has
// TOTP on.
func (s *Service) startTOTPChallenge(c *gin.Context, user *authn.User, rsp *proto.Response) {
	// The password was right and the sign-in is half done: a login in the
	// log only when the second factor passes.
	audit.SetAction(c, "login-password")
	audit.SetOutcome(c, audit.OutcomeOK)
	audit.SetDetail(c, "second factor required")
	token, err := guard.issue(user.Username, user.TokenVersion)
	if err != nil {
		rsp.ErrRsp(c, -3, "authentication unavailable")
		return
	}
	rsp.Data = &proto.LoginChallengeRsp{TOTPRequired: true, Challenge: token}
	rsp.ErrRsp(c, codeTOTPRequired, "two-factor code required")
}

// LoginTOTP is the second login step.
func (s *Service) LoginTOTP(c *gin.Context) {
	var req proto.LoginTOTPReq
	var rsp proto.Response

	if config.GetInstance().Authentication == "disable" {
		rsp.OkRsp(c)
		return
	}

	clientIP := GetClientIP(c)
	if locked, code, msg := CheckLoginAttempt(clientIP); locked {
		time.Sleep(3 * time.Second)
		rsp.ErrRsp(c, code, msg)
		return
	}
	if err := proto.ParseFormRequest(c, &req); err != nil {
		time.Sleep(codeFailureDelay)
		rsp.ErrRsp(c, -1, "invalid parameters")
		return
	}

	pending, ok := guard.lookup(req.Challenge)
	if !ok {
		rsp.ErrRsp(c, codeTOTPChallengeGo, "the sign-in expired, enter the password again")
		return
	}
	if guard.locked(pending.username) {
		time.Sleep(codeFailureDelay)
		rsp.ErrRsp(c, -5, "too many wrong codes, try again later")
		return
	}

	audit.SetActor(c, pending.username)
	user, err := authn.DefaultStore.Get(pending.username)
	if err != nil || !user.Enabled || user.TokenVersion != pending.tokenVersion || !user.TOTPEnabled() {
		guard.drop(req.Challenge)
		rsp.ErrRsp(c, codeTOTPChallengeGo, "the sign-in expired, enter the password again")
		return
	}

	ok, err = checkSecondFactor(user, req.Code)
	if err != nil {
		log.Errorf("check the second factor of %s: %s", user.Username, err)
		rsp.ErrRsp(c, -3, "authentication unavailable")
		return
	}
	if !ok {
		time.Sleep(codeFailureDelay)
		userLocked := guard.fail(req.Challenge, user.Username)
		addressLocked, code, msg := RecordLoginFailure(clientIP)
		log.Warnf("wrong two-factor code for %s from %s", user.Username, clientIP)
		switch {
		case addressLocked:
			rsp.ErrRsp(c, code, msg)
		case userLocked:
			rsp.ErrRsp(c, -5, "too many wrong codes, try again later")
		default:
			rsp.ErrRsp(c, -2, "invalid code")
		}
		return
	}

	guard.drop(req.Challenge)
	guard.succeed(user.Username)
	ClearLoginAttempt(clientIP)
	token, err := middleware.GenerateJWT(user.Username, user.TokenVersion)
	if err != nil {
		rsp.ErrRsp(c, -3, "generate token failed")
		return
	}
	setSessionCookie(c, token)
	rsp.OkRsp(c)
	log.Infof("user logged in with a second factor: %s", user.Username)
}

// checkSecondFactor accepts a current TOTP code or an unused recovery code.
func checkSecondFactor(user *authn.User, code string) (bool, error) {
	if totp.IsNumericCode(code) {
		secret, err := totpKeyring.Open(user.Username, user.TOTPSecret)
		if err != nil {
			return false, err
		}
		return guard.verify(user.Username, secret, code), nil
	}

	used, err := authn.DefaultStore.ConsumeRecoveryCode(user.Username, totp.HashRecoveryCode(code))
	if err != nil || !used {
		return false, err
	}
	log.Warnf("recovery code used by %s", user.Username)
	return true, nil
}

func (s *Service) GetTOTP(c *gin.Context) {
	var rsp proto.Response
	principal, ok := middleware.CurrentPrincipal(c)
	if !ok {
		rsp.ErrRsp(c, -1, "invalid session")
		return
	}
	user, err := authn.DefaultStore.Get(principal.Username)
	if err != nil {
		rsp.ErrRsp(c, -1, "failed to load the account")
		return
	}
	rsp.OkRspWithData(c, &proto.GetTOTPRsp{
		Enabled:       user.TOTPEnabled(),
		RecoveryCodes: len(user.RecoveryCodes),
	})
}

// SetupTOTP makes a secret for the session's account and shows it. Nothing
// changes until EnableTOTP receives a code from it.
func (s *Service) SetupTOTP(c *gin.Context) {
	var req proto.TOTPPasswordReq
	var rsp proto.Response
	user, ok := s.confirmPassword(c, &req, &rsp)
	if !ok {
		return
	}
	if user.TOTPEnabled() {
		rsp.ErrRsp(c, -4, "two-factor login is already on")
		return
	}

	secret, err := totp.NewSecret()
	if err != nil {
		rsp.ErrRsp(c, -5, "failed to make a secret")
		return
	}
	guard.startEnrolment(user.Username, secret)

	account := user.Username
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		account += "@" + hostname
	}
	rsp.OkRspWithData(c, &proto.SetupTOTPRsp{
		Secret: totp.EncodeSecret(secret),
		URI:    totp.URI(totpIssuer, account, secret),
	})
}

// EnableTOTP puts the pending secret in force once a code from it is right,
// and returns the recovery codes, the one time they are shown.
func (s *Service) EnableTOTP(c *gin.Context) {
	var req proto.EnableTOTPReq
	var rsp proto.Response
	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid parameters")
		return
	}
	principal, ok := middleware.CurrentPrincipal(c)
	if !ok {
		rsp.ErrRsp(c, -1, "invalid session")
		return
	}

	secret, expired := guard.confirmEnrolment(principal.Username, req.Code)
	if expired {
		rsp.ErrRsp(c, -3, "the setup expired, start again")
		return
	}
	if secret == nil {
		time.Sleep(codeFailureDelay)
		rsp.ErrRsp(c, -2, "invalid code")
		return
	}

	sealed, err := totpKeyring.Seal(principal.Username, string(secret))
	if err != nil {
		log.Errorf("seal the TOTP secret of %s: %s", principal.Username, err)
		rsp.ErrRsp(c, -4, "failed to save")
		return
	}
	codes, hashes, err := newRecoveryCodes()
	if err != nil {
		rsp.ErrRsp(c, -4, "failed to save")
		return
	}
	if _, err = authn.DefaultStore.SetTOTP(principal.Username, sealed, hashes); err != nil {
		log.Errorf("save the TOTP secret of %s: %s", principal.Username, err)
		rsp.ErrRsp(c, -4, "failed to save")
		return
	}

	rsp.OkRspWithData(c, &proto.RecoveryCodesRsp{RecoveryCodes: codes})
	log.Infof("two-factor login turned on for %s", principal.Username)
}

// DisableTOTP turns the session account's second factor off.
func (s *Service) DisableTOTP(c *gin.Context) {
	var req proto.TOTPPasswordReq
	var rsp proto.Response
	user, ok := s.confirmPassword(c, &req, &rsp)
	if !ok {
		return
	}
	if _, err := authn.DefaultStore.ClearTOTP(user.Username); err != nil {
		rsp.ErrRsp(c, -4, "failed to save")
		return
	}
	guard.forget(user.Username)
	rsp.OkRsp(c)
	log.Infof("two-factor login turned off for %s", user.Username)
}

// RegenerateRecoveryCodes replaces the session account's recovery codes.
func (s *Service) RegenerateRecoveryCodes(c *gin.Context) {
	var req proto.TOTPPasswordReq
	var rsp proto.Response
	user, ok := s.confirmPassword(c, &req, &rsp)
	if !ok {
		return
	}
	if !user.TOTPEnabled() {
		rsp.ErrRsp(c, -4, "two-factor login is off")
		return
	}
	codes, hashes, err := newRecoveryCodes()
	if err != nil {
		rsp.ErrRsp(c, -5, "failed to save")
		return
	}
	if _, err = authn.DefaultStore.SetRecoveryCodes(user.Username, hashes); err != nil {
		rsp.ErrRsp(c, -5, "failed to save")
		return
	}
	rsp.OkRspWithData(c, &proto.RecoveryCodesRsp{RecoveryCodes: codes})
	log.Infof("recovery codes replaced for %s", user.Username)
}

// ResetUserTOTP is an administrator turning another account's second factor
// off, for a user who lost both the authenticator and the recovery codes.
func (s *Service) ResetUserTOTP(c *gin.Context) {
	var rsp proto.Response
	username := c.Param("username")
	principal, _ := middleware.CurrentPrincipal(c)
	// One's own second factor goes through DisableTOTP, which asks for the
	// password, so an administrator's unattended session cannot drop it.
	if principal.Username == username {
		rsp.ErrRsp(c, -2, "turn your own two-factor login off from the account page")
		return
	}
	if _, err := authn.DefaultStore.ClearTOTP(username); err != nil {
		rsp.ErrRsp(c, -1, err.Error())
		return
	}
	guard.forget(username)
	rsp.OkRsp(c)
	log.Infof("two-factor login reset for %s by %s", username, principal.Username)
}

// confirmPassword parses a TOTPPasswordReq and checks its password against
// the session's account. A change to the second factor asks for it, so a
// session left open on someone else's screen cannot turn 2FA off or read new
// recovery codes. Wrong passwords count against the address like a login.
func (s *Service) confirmPassword(c *gin.Context, req *proto.TOTPPasswordReq, rsp *proto.Response) (*authn.User, bool) {
	if err := proto.ParseFormRequest(c, req); err != nil {
		rsp.ErrRsp(c, -1, "invalid parameters")
		return nil, false
	}
	principal, ok := middleware.CurrentPrincipal(c)
	if !ok {
		rsp.ErrRsp(c, -1, "invalid session")
		return nil, false
	}
	clientIP := GetClientIP(c)
	if locked, code, msg := CheckLoginAttempt(clientIP); locked {
		rsp.ErrRsp(c, code, msg)
		return nil, false
	}
	password, err := utils.DecodeDecrypt(req.Password)
	if err == nil && password != "" {
		var user *authn.User
		user, ok, err = authn.DefaultStore.Authenticate(principal.Username, password)
		if err != nil {
			rsp.ErrRsp(c, -3, "authentication unavailable")
			return nil, false
		}
		if ok {
			return user, true
		}
	}
	time.Sleep(codeFailureDelay)
	if locked, code, msg := RecordLoginFailure(clientIP); locked {
		rsp.ErrRsp(c, code, msg)
		return nil, false
	}
	rsp.ErrRsp(c, -2, "password is incorrect")
	return nil, false
}

func newRecoveryCodes() (codes, hashes []string, err error) {
	codes, err = totp.NewRecoveryCodes()
	if err != nil {
		return nil, nil, err
	}
	hashes = make([]string, len(codes))
	for i, code := range codes {
		hashes[i] = totp.HashRecoveryCode(code)
	}
	return codes, hashes, nil
}
