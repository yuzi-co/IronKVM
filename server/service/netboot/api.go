package netboot

import (
	"errors"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/vpn"
)

// The web UI's Network boot page. These handlers answer under /api/netboot
// with the web UI's response envelope, behind its session check and admin
// role.

type setSettingsRequest struct {
	USB *bool `json:"usb" form:"usb" validate:"required"`
	LAN *bool `json:"lan" form:"lan" validate:"required"`
}

func (s *Service) GetStatus(c *gin.Context) {
	var rsp proto.Response

	rsp.OkRspWithData(c, s.status())
}

func (s *Service) SetSettings(c *gin.Context) {
	var rsp proto.Response
	var req setSettingsRequest

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	next := config.NetBoot{USB: *req.USB, LAN: *req.LAN}
	if err := s.apply(next); err != nil {
		log.Errorf("netboot: set usb=%t lan=%t: %s", next.USB, next.LAN, err)
		rsp.ErrRsp(c, errorCode(err), err.Error())
		return
	}

	log.Infof("netboot: usb=%t lan=%t", next.USB, next.LAN)
	rsp.OkRspWithData(c, s.status())
}

// Install fetches dnsmasq and the boot files, which takes minutes on the
// board's link. The page waits for it.
func (s *Service) Install(c *gin.Context) {
	var rsp proto.Response

	if err := s.install(); err != nil {
		log.Errorf("netboot: install: %s", err)
		rsp.ErrRsp(c, errorCode(err), vpn.Message("install failed", err))
		return
	}

	log.Infof("netboot: installed")
	rsp.OkRspWithData(c, s.status())
}

func (s *Service) Uninstall(c *gin.Context) {
	var rsp proto.Response

	if err := s.uninstall(); err != nil {
		log.Errorf("netboot: uninstall: %s", err)
		rsp.ErrRsp(c, errorCode(err), vpn.Message("uninstall failed", err))
		return
	}

	log.Infof("netboot: uninstalled")
	rsp.OkRspWithData(c, s.status())
}

// errorCode is -2 for a change refused because of the board's state, which
// the page shows as it is, and -3 for one that failed on the way.
func errorCode(err error) int {
	if errors.Is(err, errBusy) || errors.Is(err, errNotInstalled) {
		return -2
	}
	return -3
}

// CurrentSettings reads the settings from the running configuration.
func CurrentSettings() config.NetBoot {
	return config.GetInstance().NetBoot
}

// SaveSettings writes the settings to server.yaml and to the running
// configuration.
func SaveSettings(s config.NetBoot) error {
	conf, err := config.Read()
	if err != nil {
		return err
	}

	conf.NetBoot = s
	if err := config.Write(conf); err != nil {
		return err
	}

	config.GetInstance().NetBoot = s
	return nil
}
