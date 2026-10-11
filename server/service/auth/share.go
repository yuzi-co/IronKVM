package auth

import (
	"errors"
	"time"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/proto"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func shareLinkInfo(link *authn.ShareLink) proto.ShareLinkInfo {
	return proto.ShareLinkInfo{
		ID:        link.ID,
		Label:     link.Label,
		CreatedBy: link.CreatedBy,
		CreatedAt: link.CreatedAt.Unix(),
		ExpiresAt: link.ExpiresAt.Unix(),
	}
}

// ListShareLinks lists the links that are still good.
func (s *Service) ListShareLinks(c *gin.Context) {
	var rsp proto.Response
	links, err := authn.DefaultShares.List()
	if err != nil {
		rsp.ErrRsp(c, -1, "failed to load share links")
		return
	}
	result := make([]proto.ShareLinkInfo, 0, len(links))
	for index := range links {
		result = append(result, shareLinkInfo(&links[index]))
	}
	rsp.OkRspWithData(c, &proto.ListShareLinksRsp{Links: result})
}

// CreateShareLink makes a view-only link. The token is in the answer and
// nowhere else.
func (s *Service) CreateShareLink(c *gin.Context) {
	var req proto.CreateShareLinkReq
	var rsp proto.Response
	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid parameters")
		return
	}
	principal, _ := middleware.CurrentPrincipal(c)
	ttl := time.Duration(req.TTLSeconds) * time.Second
	link, token, err := authn.DefaultShares.Create(req.Label, principal.Username, ttl)
	switch {
	case errors.Is(err, authn.ErrShareTTL), errors.Is(err, authn.ErrShareLimit):
		rsp.ErrRsp(c, -2, err.Error())
		return
	case err != nil:
		log.Errorf("create share link: %s", err)
		rsp.ErrRsp(c, -3, "failed to create the share link")
		return
	}
	rsp.OkRspWithData(c, &proto.CreateShareLinkRsp{ShareLinkInfo: shareLinkInfo(link), Token: token})
	log.Infof("share link %s created by %s, expires %s", link.ID, principal.Username, link.ExpiresAt.Format(time.RFC3339))
}

// RevokeShareLink ends a link and every session it opened.
func (s *Service) RevokeShareLink(c *gin.Context) {
	var rsp proto.Response
	id := c.Param("id")
	if err := authn.DefaultShares.Revoke(id); err != nil {
		if errors.Is(err, authn.ErrShareNotFound) {
			rsp.ErrRsp(c, -1, err.Error())
			return
		}
		log.Errorf("revoke share link %s: %s", id, err)
		rsp.ErrRsp(c, -2, "failed to revoke the share link")
		return
	}
	middleware.RevokeUserSessions(authn.SharePrincipalPrefix + id)
	rsp.OkRsp(c)
	log.Infof("share link %s revoked", id)
}

// RedeemShareLink opens a view-only session for whoever holds a link's token.
// It takes no credential, so it is held to the login's rate limit.
func (s *Service) RedeemShareLink(c *gin.Context) {
	var req proto.RedeemShareLinkReq
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
		time.Sleep(3 * time.Second)
		rsp.ErrRsp(c, -1, "invalid parameters")
		return
	}
	link, ok := authn.DefaultShares.Redeem(req.Token)
	if !ok {
		s.loginFailed(c, clientIP, &rsp)
		return
	}
	token, err := middleware.GenerateShareJWT(link.ID, link.ExpiresAt)
	if err != nil {
		rsp.ErrRsp(c, -3, "generate token failed")
		return
	}
	setSessionCookie(c, token)
	rsp.OkRsp(c)
	log.Infof("share link %s opened", link.ID)
}
