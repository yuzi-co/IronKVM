package tailscale

import (
	"os"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Service struct{}

// Variables rather than constants so the tests can point them at a scratch
// root. On a distribution image both are links into the add-on on /data.
var (
	TailscalePath  = "/usr/bin/tailscale"
	TailscaledPath = "/usr/sbin/tailscaled"
)

// addonSpec is what Tailscale needs back from the root filesystem after a new
// image: its two binaries in their usual places and its boot script while it is
// enabled. Its login is already on /data, where S98tailscaled keeps it.
func addonSpec() addon.Spec {
	return addon.Spec{
		Name: addon.Tailscale.Name,
		Links: []addon.Link{
			{Path: "/usr/bin/tailscale", File: "tailscale"},
			{Path: "/usr/sbin/tailscaled", File: "tailscaled"},
		},
		Initd: addon.Tailscale.Initd,
	}
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Install(c *gin.Context) {
	var rsp proto.Response

	if vpn.Refuse(c, addon.Tailscale) {
		return
	}

	if !isInstalled() {
		if err := install(); err != nil {
			rsp.ErrRsp(c, -1, vpn.Message("install failed", err))
			return
		}

		if err := NewCli().Start(); err != nil {
			log.Errorf("failed to start tailscale after install: %s", err)
		}
	}

	rsp.OkRsp(c)
	log.Debugf("install tailscale successfully")
}

func (s *Service) Uninstall(c *gin.Context) {
	var rsp proto.Response

	if err := NewCli().Stop(); err != nil {
		log.Debugf("failed to stop tailscale before uninstall: %s", err)
	}
	if err := addon.SetBoot(addon.Tailscale, false); err != nil {
		log.Errorf("failed to turn off tailscale at boot: %s", err)
	}

	if addon.OnData() {
		if err := addon.Remove(addonSpec()); err != nil {
			log.Errorf("failed to remove the tailscale add-on: %s", err)
		}
	}
	_ = os.Remove(TailscalePath)
	_ = os.Remove(TailscaledPath)

	rsp.OkRsp(c)
	log.Debugf("uninstall tailscale successfully")
}

// Start starts the daemon. Start at boot is the boot route's alone now, and
// the Go memory limit is S98tailscaled's, derived from the addons group.
func (s *Service) Start(c *gin.Context) {
	var rsp proto.Response

	if vpn.Refuse(c, addon.Tailscale) {
		return
	}

	if err := NewCli().Start(); err != nil {
		log.Errorf("failed to run tailscale start: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("start failed", err))
		return
	}

	rsp.OkRsp(c)
	log.Debugf("tailscale start successfully")
}

func (s *Service) Restart(c *gin.Context) {
	var rsp proto.Response

	if err := NewCli().Restart(); err != nil {
		log.Errorf("failed to run tailscale restart: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("restart failed", err))
		return
	}

	rsp.OkRsp(c)
	log.Debugf("tailscale restart successfully")
}

func (s *Service) Stop(c *gin.Context) {
	var rsp proto.Response

	if err := NewCli().Stop(); err != nil {
		log.Errorf("failed to run tailscale stop: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("stop failed", err))
		return
	}

	rsp.OkRsp(c)
	log.Debugf("tailscale stop successfully")
}

func (s *Service) Up(c *gin.Context) {
	var rsp proto.Response

	if vpn.Refuse(c, addon.Tailscale) {
		return
	}

	if err := NewCli().Up(); err != nil {
		log.Errorf("failed to run tailscale up: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("tailscale up failed", err))
		return
	}

	rsp.OkRsp(c)
	log.Debugf("run tailscale up successfully")
}

func (s *Service) Down(c *gin.Context) {
	var rsp proto.Response

	if err := NewCli().Down(); err != nil {
		log.Errorf("failed to run tailscale down: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("tailscale down failed", err))
		return
	}

	rsp.OkRsp(c)
	log.Debugf("run tailscale down successfully")
}

func (s *Service) Login(c *gin.Context) {
	var rsp proto.Response

	if vpn.Refuse(c, addon.Tailscale) {
		return
	}

	cli := NewCli()
	status, err := cli.Status()
	if err != nil {
		if err := cli.Start(); err != nil {
			rsp.ErrRsp(c, -1, vpn.Message("start failed", err))
			return
		}
		status, err = cli.Status()
	}

	if err != nil {
		log.Errorf("failed to get tailscale status: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("unknown status", err))
		return
	}

	if status.BackendState == "Running" {
		rsp.OkRspWithData(c, &proto.VpnLoginRsp{})
		return
	}

	url, err := cli.Login()
	if err != nil {
		log.Errorf("failed to run tailscale login: %s", err)
		rsp.ErrRsp(c, -2, vpn.Message("login failed", err))
		return
	}

	rsp.OkRspWithData(c, &proto.VpnLoginRsp{Url: url})
	log.Debugf("tailscale login url: %s", url)
}

func (s *Service) Logout(c *gin.Context) {
	var rsp proto.Response

	if err := NewCli().Logout(); err != nil {
		log.Errorf("failed to run tailscale logout: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("logout failed", err))
		return
	}

	rsp.OkRsp(c)
	log.Debugf("tailscale logout successfully")
}

func (s *Service) GetStatus(c *gin.Context) {
	var rsp proto.Response

	st := proto.VpnStatus{State: proto.VpnNotInstall}
	if isInstalled() {
		st = proto.VpnStatus{State: proto.VpnNotRunning}
		if ts, err := NewCli().Status(); err != nil {
			log.Debugf("failed to get tailscale status: %s", err)
		} else if st, err = toVpnStatus(ts); err != nil {
			log.Errorf("%s", err)
			rsp.ErrRsp(c, -1, err.Error())
			return
		}
	}

	vpn.Fill(&st, addon.Tailscale)
	rsp.OkRspWithData(c, &st)
}

// Boot turns start at boot on or off.
func (s *Service) Boot(c *gin.Context) {
	vpn.Boot(c, addon.Tailscale, isInstalled())
}
