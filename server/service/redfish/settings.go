package redfish

import (
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
)

// The web UI's Redfish page. These handlers answer under /api/redfish with the
// web UI's response envelope, behind its own session check and admin role;
// none of them is reachable through /redfish.

// Settings is what the page shows about the service.
type Settings struct {
	Enabled bool `json:"enabled"`
	// HTTPS is whether the board serves over HTTPS, which most Redfish
	// clients need.
	HTTPS bool `json:"https"`
	// ServiceRoot is the service root's path on the board.
	ServiceRoot string `json:"serviceRoot"`
	// ResetTypes are the ResetType values the system offers now, which
	// depend on the power LED setting.
	ResetTypes []string `json:"resetTypes"`
}

type setSettingsRequest struct {
	Enabled *bool `json:"enabled" form:"enabled" validate:"required"`
}

// SessionInfo is one Redfish session as the page lists it. It never carries
// the token.
type SessionInfo struct {
	ID        string `json:"id"`
	User      string `json:"user"`
	CreatedAt string `json:"createdAt"`
	LastUsed  string `json:"lastUsed"`
}

func (s *Service) GetSettings(c *gin.Context) {
	var rsp proto.Response

	rsp.OkRspWithData(c, &Settings{
		Enabled:     s.enabled(),
		HTTPS:       s.deps.HTTPS(),
		ServiceRoot: rootPath,
		ResetTypes:  s.offeredResetTypes(),
	})
}

// SetSettings turns the service on or off. Turning it off ends every
// session, so a token from before does not work when it is turned on again.
func (s *Service) SetSettings(c *gin.Context) {
	var rsp proto.Response
	var req setSettingsRequest

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	if s.deps.SetEnabled == nil {
		rsp.ErrRsp(c, -2, "the setting cannot be changed")
		return
	}

	s.switchMu.Lock()
	defer s.switchMu.Unlock()

	on := *req.Enabled
	if err := s.deps.SetEnabled(on); err != nil {
		log.Errorf("redfish: save the enabled setting: %s", err)
		rsp.ErrRsp(c, -2, "failed to save the setting")
		return
	}
	// Off first, then the sessions: a login that finishes in between sees
	// the service off and drops its own session.
	if !on {
		s.sessions.removeAll()
	}

	log.Infof("redfish enabled: %t", on)
	rsp.OkRsp(c)
}

// GetSessions lists the open Redfish sessions, oldest first.
func (s *Service) GetSessions(c *gin.Context) {
	var rsp proto.Response

	sessions := s.sessions.list()
	out := make([]SessionInfo, 0, len(sessions))
	for _, sess := range sessions {
		out = append(out, SessionInfo{
			ID:        sess.id,
			User:      sess.username,
			CreatedAt: sess.created.UTC().Format(time.RFC3339),
			LastUsed:  sess.lastUsed.UTC().Format(time.RFC3339),
		})
	}

	rsp.OkRspWithData(c, out)
}

// EndSession ends one Redfish session: its token answers 401 from then on.
func (s *Service) EndSession(c *gin.Context) {
	var rsp proto.Response

	if !s.sessions.remove(c.Param("id")) {
		rsp.ErrRsp(c, -1, "no such session")
		return
	}

	rsp.OkRsp(c)
}
