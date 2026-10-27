package auth

import (
	"errors"
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

func (s *Service) Login(c *gin.Context) {
	var req proto.LoginReq
	var rsp proto.Response

	conf := config.GetInstance()
	if conf.Authentication == "disable" {
		// There is no session to hand out. Clients of Sipeed's API still
		// expect a token from a successful login, Home Assistant's nanokvm
		// integration among them, and take "disabled" to mean this.
		rsp.OkRspWithData(c, &proto.LoginRsp{Token: proto.LoginDisabledToken})
		return
	}

	clientIP := GetClientIP(c)
	if locked, code, msg := CheckLoginAttempt(clientIP); locked {
		time.Sleep(3 * time.Second)
		rsp.ErrRsp(c, code, msg)
		return
	}
	if err := proto.ParseFormRequest(c, &req); err != nil {
		time.Sleep(3 * time.Second)
		rsp.ErrRsp(c, -1, "invalid parameters")
		return
	}
	audit.SetActor(c, req.Username)
	password, err := utils.DecodeDecrypt(req.Password)
	if err != nil || password == "" {
		s.loginFailed(c, clientIP, &rsp)
		return
	}
	user, ok, err := authn.DefaultStore.Authenticate(req.Username, password)
	if err != nil {
		log.Errorf("load account during login: %s", err)
		time.Sleep(2 * time.Second)
		rsp.ErrRsp(c, -3, "authentication unavailable")
		return
	}
	if !ok {
		s.loginFailed(c, clientIP, &rsp)
		return
	}
	// The address's failed attempts are kept until the second step passes:
	// a right password alone is not a login.
	if user.TOTPEnabled() {
		s.startTOTPChallenge(c, user, &rsp)
		return
	}

	ClearLoginAttempt(clientIP)
	token, err := middleware.IssueSession(c, user.Username, user.TokenVersion)
	if err != nil {
		time.Sleep(time.Second)
		rsp.ErrRsp(c, -3, "generate token failed")
		return
	}
	setSessionCookie(c, token)
	rsp.OkRsp(c)
	log.Infof("user logged in: %s", user.Username)
}

func (s *Service) loginFailed(c *gin.Context, clientIP string, rsp *proto.Response) {
	time.Sleep(2 * time.Second)
	if locked, code, msg := RecordLoginFailure(clientIP); locked {
		rsp.ErrRsp(c, code, msg)
		return
	}
	rsp.ErrRsp(c, -2, "invalid username or password")
}

func (s *Service) Logout(c *gin.Context) {
	var rsp proto.Response
	principal, ok := middleware.CurrentPrincipal(c)
	conf := config.GetInstance()
	if ok && principal.ShareID != "" {
		// A share session belongs to the link, not to an account: there is no
		// token version to revoke, and the link stays good for whoever else
		// holds it.
		clearSessionCookie(c)
		rsp.OkRsp(c)
		return
	}
	// Logging out ends this login session alone. Before sessions were kept
	// one by one, revoking a token meant raising the account's token version,
	// which signed the account out on every device; the list in Settings >
	// Account does that now, for the sessions the user picks.
	if ok && conf.Authentication != "disable" && conf.JWT.RevokeTokensOnLogout && principal.SessionID != "" {
		err := middleware.RevokeLoginSession(principal.SessionID)
		if err != nil && !errors.Is(err, authn.ErrSessionNotFound) {
			rsp.ErrRsp(c, -1, "failed to revoke session")
			return
		}
		log.Infof("user logged out: %s", principal.Username)
	}
	clearSessionCookie(c)
	rsp.OkRsp(c)
}

func (s *Service) GetAccount(c *gin.Context) {
	var rsp proto.Response
	principal, ok := middleware.CurrentPrincipal(c)
	if !ok {
		rsp.ErrRsp(c, -1, "get account failed")
		return
	}
	// An account that cannot be read reports false for both, the answers
	// that promise less.
	user, err := authn.DefaultStore.Get(principal.Username)
	rsp.OkRspWithData(c, &proto.GetAccountRsp{
		Username:      principal.Username,
		Role:          string(principal.Role),
		SystemAccount: err == nil && isSystemPasswordOwner(user),
		TOTPEnabled:   err == nil && user.TOTPEnabled(),
	})
}

func setSessionCookie(c *gin.Context, token string) {
	middleware.SetSessionCookie(c, token)
}

func clearSessionCookie(c *gin.Context) {
	middleware.ClearSessionCookie(c, middleware.RequestIsSecure(c))
}
