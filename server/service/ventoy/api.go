package ventoy

import (
	"errors"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
)

// The image manager's Ventoy section. These handlers answer under
// /api/ventoy with the web UI's response envelope, behind its session check
// and admin role.

type setImagesRequest struct {
	Images []string `json:"images" form:"images"`
}

func (s *Service) GetStatus(c *gin.Context) {
	var rsp proto.Response

	rsp.OkRspWithData(c, s.status())
}

func (s *Service) SetImages(c *gin.Context) {
	var rsp proto.Response
	var req setImagesRequest

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	if err := s.setImages(req.Images); err != nil {
		log.Errorf("ventoy: set images: %s", err)
		rsp.ErrRsp(c, errorCode(err), err.Error())
		return
	}

	log.Infof("ventoy: %d images selected", len(req.Images))
	rsp.OkRspWithData(c, s.status())
}

// Install fetches the release, which takes a minute on the board's link.
// The page waits for it.
func (s *Service) Install(c *gin.Context) {
	s.run(c, "install", s.install)
}

func (s *Service) Uninstall(c *gin.Context) {
	s.run(c, "uninstall", s.uninstall)
}

func (s *Service) Insert(c *gin.Context) {
	s.run(c, "insert", s.insert)
}

func (s *Service) Eject(c *gin.Context) {
	s.run(c, "eject", s.eject)
}

func (s *Service) run(c *gin.Context, what string, action func() error) {
	var rsp proto.Response

	if err := action(); err != nil {
		log.Errorf("ventoy: %s: %s", what, err)
		rsp.ErrRsp(c, errorCode(err), err.Error())
		return
	}

	log.Infof("ventoy: %s done", what)
	rsp.OkRspWithData(c, s.status())
}

// errorCode is -2 for a change refused because of the board's state, which
// the page shows as it is, and -3 for one that failed on the way.
func errorCode(err error) int {
	for _, refused := range []error{errNoKernel, errNotReady, errNoImages, errInDrive, errNoData, errNotImage} {
		if errors.Is(err, refused) {
			return -2
		}
	}
	return -3
}
