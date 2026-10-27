package dashcam

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/audit"
)

// GetState answers the state for the page and the save button.
func (s *Service) GetState(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, s.State())
}

// PostSave saves the last minute: the button on the page, or a call of the
// route with an API key, which names the clip "api" instead of "manual".
func (s *Service) PostSave(c *gin.Context) {
	var rsp proto.Response

	kind := TriggerManual
	if id, ok := audit.IdentityOf(c); ok && id.Via == "apikey" {
		kind = TriggerAPI
	}

	p, err := s.Trigger(kind, true)
	if err != nil {
		audit.SetDetail(c, err.Error())
		code := -1
		switch {
		case errors.Is(err, ErrDisabled):
			code = -2
		case errors.Is(err, ErrEmpty):
			code = -3
		case errors.Is(err, ErrTooSoon):
			code = -4
		}
		rsp.ErrRsp(c, code, err.Error())
		return
	}

	detail := "until " + p.Deadline.Format("15:04:05")
	if p.Joined {
		detail = "joined the clip of " + strings.Join(p.Triggers, ", ") + ", " + detail
	}
	audit.SetTarget(c, clipBase(p.Started, p.Triggers[:1]))
	audit.SetDetail(c, detail)
	rsp.OkRspWithData(c, p)
}

// GetClips lists the clips, newest first.
func (s *Service) GetClips(c *gin.Context) {
	var rsp proto.Response

	clips, err := listClips(s.deps.Dir)
	if err != nil {
		log.Errorf("dashcam: list the clips: %s", err)
		rsp.ErrRsp(c, -1, "failed to list the clips")
		return
	}
	rsp.OkRspWithData(c, clips)
}

// clipPath answers the path of a clip that exists, by its name alone.
func (s *Service) clipPath(name string) (string, bool) {
	if !ValidName(name) {
		return "", false
	}
	path := filepath.Join(s.deps.Dir, name)
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return "", false
	}
	return path, true
}

// GetClip sends a clip. Range requests work, so a browser can seek in it.
func (s *Service) GetClip(c *gin.Context) {
	var rsp proto.Response

	name := c.Param("name")
	path, ok := s.clipPath(name)
	if !ok {
		rsp.ErrRsp(c, -1, "no such clip")
		return
	}
	c.FileAttachment(path, name)
}

// DeleteClip deletes a clip. The audit log takes its name from the route.
func (s *Service) DeleteClip(c *gin.Context) {
	var rsp proto.Response

	path, ok := s.clipPath(c.Param("name"))
	if !ok {
		rsp.ErrRsp(c, -1, "no such clip")
		return
	}

	s.writeMu.Lock()
	err := os.Remove(path)
	s.writeMu.Unlock()
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Errorf("dashcam: delete %s: %s", filepath.Base(path), err)
		rsp.ErrRsp(c, -2, "failed to delete the clip")
		return
	}
	rsp.OkRsp(c)
}

// GetSettings answers the settings.
func (s *Service) GetSettings(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, s.Settings())
}

type setSettingsRequest struct {
	Enabled     *bool    `json:"enabled" validate:"required"`
	Seconds     int      `json:"seconds"`
	PostSeconds int      `json:"postSeconds"`
	MaxKbps     int      `json:"maxKbps"`
	StorageMB   int      `json:"storageMB"`
	Triggers    Triggers `json:"triggers"`
}

// SetSettings saves the settings and applies them at once.
func (s *Service) SetSettings(c *gin.Context) {
	var rsp proto.Response
	var req setSettingsRequest

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	settings := Settings{
		Enabled:     *req.Enabled,
		Seconds:     req.Seconds,
		PostSeconds: req.PostSeconds,
		MaxKbps:     req.MaxKbps,
		StorageMB:   req.StorageMB,
		Triggers:    req.Triggers,
	}
	if err := settings.Validate(); err != nil {
		rsp.ErrRsp(c, -1, err.Error())
		return
	}
	if err := s.Configure(settings); err != nil {
		log.Errorf("dashcam: save the settings: %s", err)
		rsp.ErrRsp(c, -2, "failed to save the settings")
		return
	}
	rsp.OkRspWithData(c, s.Settings())
}
