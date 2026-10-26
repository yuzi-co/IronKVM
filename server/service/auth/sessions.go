package auth

import (
	"errors"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/audit"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ListSessions lists the login sessions of the signed-in account, with the
// one this request came through marked.
func (s *Service) ListSessions(c *gin.Context) {
	var rsp proto.Response
	principal, ok := sessionPrincipal(c, &rsp)
	if !ok {
		return
	}
	rsp.OkRspWithData(c, &proto.ListLoginSessionsRsp{Sessions: listSessions(principal, principal.Username)})
}

// ListAllSessions lists every account's login sessions, for an administrator.
func (s *Service) ListAllSessions(c *gin.Context) {
	var rsp proto.Response
	principal, ok := sessionPrincipal(c, &rsp)
	if !ok {
		return
	}
	rsp.OkRspWithData(c, &proto.ListLoginSessionsRsp{Sessions: listSessions(principal, "")})
}

// RevokeSession ends one login session: the caller's own, or anyone's for an
// administrator. Another account's session reads to anyone else as one that
// does not exist, so the answer does not say whether an id is in use.
func (s *Service) RevokeSession(c *gin.Context) {
	var rsp proto.Response
	principal, ok := sessionPrincipal(c, &rsp)
	if !ok {
		return
	}
	id := c.Param("id")
	session, found := authn.DefaultSessions.Get(id)
	if !found || (session.Username != principal.Username && principal.Role != authn.RoleAdmin) {
		rsp.ErrRsp(c, -2, authn.ErrSessionNotFound.Error())
		return
	}

	audit.SetTarget(c, session.Username)
	audit.SetDetail(c, sessionAuditDetail(session))
	err := middleware.RevokeLoginSession(id)
	if errors.Is(err, authn.ErrSessionNotFound) {
		rsp.ErrRsp(c, -2, err.Error())
		return
	}
	if err != nil {
		// Ended for now; the file still holds it and a restart would bring
		// it back, so the caller is told it did not stick.
		rsp.ErrRsp(c, -3, "the session ended but could not be removed from the device")
		return
	}
	if id == principal.SessionID {
		clearSessionCookie(c)
	}
	rsp.OkRsp(c)
	log.Infof("login session of %s revoked by %s", session.Username, principal.Username)
}

// RevokeOtherSessions ends every session of the caller's account but the one
// the request came through.
func (s *Service) RevokeOtherSessions(c *gin.Context) {
	var rsp proto.Response
	principal, ok := sessionPrincipal(c, &rsp)
	if !ok {
		return
	}
	if principal.SessionID == "" {
		rsp.ErrRsp(c, -1, "invalid session")
		return
	}
	count := 0
	for _, session := range authn.DefaultSessions.List(principal.Username) {
		if session.ID != principal.SessionID {
			count++
		}
	}
	middleware.RevokeOtherSessions(principal.Username, principal.SessionID)
	audit.SetTarget(c, principal.Username)
	rsp.OkRspWithData(c, &proto.RevokeOtherSessionsRsp{Revoked: count})
	log.Infof("%s revoked %d other login sessions", principal.Username, count)
}

// sessionPrincipal is the account a session request speaks for. A share
// session has no account and so no sessions to list.
func sessionPrincipal(c *gin.Context, rsp *proto.Response) (middleware.Principal, bool) {
	principal, ok := middleware.CurrentPrincipal(c)
	if !ok || principal.ShareID != "" {
		rsp.ErrRsp(c, -1, "invalid session")
		return middleware.Principal{}, false
	}
	return principal, true
}

// listSessions returns the live sessions of username, or of every account for
// an empty one. Sessions whose account no longer takes their token, after a
// password change, a role change or a deletion, are dropped first.
func listSessions(principal middleware.Principal, username string) []proto.LoginSessionInfo {
	if config.GetInstance().Authentication == "disable" {
		return []proto.LoginSessionInfo{}
	}
	pruneStaleSessions()
	sessions := authn.DefaultSessions.List(username)
	result := make([]proto.LoginSessionInfo, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, proto.LoginSessionInfo{
			ID:         session.ID,
			Username:   session.Username,
			CreatedAt:  session.CreatedAt.Unix(),
			LastUsedAt: session.LastUsedAt.Unix(),
			ExpiresAt:  session.ExpiresAt.Unix(),
			IP:         session.IP,
			Client:     session.Client,
			Current:    session.ID == principal.SessionID,
		})
	}
	return result
}

// pruneStaleSessions drops the sessions whose account is gone, disabled or on
// a newer token version. An account file that cannot be read drops nothing.
func pruneStaleSessions() {
	accounts := map[string]*authn.User{}
	for _, session := range authn.DefaultSessions.List("") {
		if _, seen := accounts[session.Username]; seen {
			continue
		}
		user, err := authn.DefaultStore.Get(session.Username)
		if err != nil && !errors.Is(err, authn.ErrUserNotFound) {
			return
		}
		accounts[session.Username] = user
	}
	err := authn.DefaultSessions.Prune(func(session authn.LoginSession) bool {
		user, known := accounts[session.Username]
		if !known {
			// Created after the accounts were read.
			return true
		}
		return user != nil && user.Enabled && user.TokenVersion == session.TokenVersion
	})
	if err != nil {
		log.Warnf("prune login sessions: %s", err)
	}
}

func sessionAuditDetail(session authn.LoginSession) string {
	detail := session.ID
	if len(detail) > 12 {
		detail = detail[:12]
	}
	if session.Client != "" {
		detail += " " + session.Client
	}
	if session.IP != "" {
		detail += " " + session.IP
	}
	return detail
}
