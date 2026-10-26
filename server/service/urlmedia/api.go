package urlmedia

import (
	"context"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/audit"
)

type mountReq struct {
	Drive string `json:"drive" form:"drive" validate:"required"`
	URL   string `json:"url" form:"url" validate:"required"`
}

type ejectReq struct {
	Drive string `json:"drive" form:"drive" validate:"required"`
}

// mountTimeout bounds the check of the server and the attach.
const mountTimeout = 30 * time.Second

// GetStatus answers whether the board can mount from a URL, and what it
// serves.
func (s *Service) GetStatus(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, s.Status())
}

// MountURL serves an image from a URL in a drive. The audit log gets the
// drive and the URL's host: the URL itself may carry credentials.
func (s *Service) MountURL(c *gin.Context) {
	var req mountReq
	var rsp proto.Response
	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	audit.SetTarget(c, req.Drive+": "+HostOf(req.URL))

	ctx, cancel := context.WithTimeout(c.Request.Context(), mountTimeout)
	defer cancel()
	st, err := s.Mount(ctx, req.Drive, req.URL)
	if err != nil {
		log.Warnf("urlmedia: mount %s from %s: %s", req.Drive, Redact(req.URL), err)
		code := -2
		if errors.Is(err, ErrUnavailable) {
			code = -3
		}
		rsp.ErrRsp(c, code, err.Error())
		return
	}
	rsp.OkRspWithData(c, st)
}

// EjectURL takes a drive's image from a URL out.
func (s *Service) EjectURL(c *gin.Context) {
	var req ejectReq
	var rsp proto.Response
	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	audit.SetTarget(c, req.Drive)
	if err := s.Eject(req.Drive); err != nil {
		rsp.ErrRsp(c, -2, err.Error())
		return
	}
	rsp.OkRsp(c)
}
