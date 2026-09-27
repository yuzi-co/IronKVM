package tailscale

import (
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
	"NanoKVM-Server/utils"
	"os"

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

const GoMemLimit int64 = 75

// addonSpec is what Tailscale needs back from the root filesystem after a new
// image: its two binaries in their usual places and its boot script while it is
// enabled. Its login is already on /data, where S98tailscaled keeps it.
func addonSpec() addon.Spec {
	return addon.Spec{
		Name: "tailscale",
		Links: []addon.Link{
			{Path: "/usr/bin/tailscale", File: "tailscale"},
			{Path: "/usr/sbin/tailscaled", File: "tailscaled"},
		},
		Initd: "S98tailscaled",
	}
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Install(c *gin.Context) {
	var rsp proto.Response

	if !isInstalled() {
		if err := install(); err != nil {
			rsp.ErrRsp(c, -1, "install failed")
			return
		}

		_ = NewCli().Start()
	}

	rsp.OkRsp(c)
	log.Debugf("install tailscale successfully")
}

func (s *Service) Uninstall(c *gin.Context) {
	var rsp proto.Response

	_ = NewCli().Stop()
	_ = utils.DelGoMemLimit()

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

func (s *Service) Start(c *gin.Context) {
	var rsp proto.Response

	err := NewCli().Start()
	if err != nil {
		rsp.ErrRsp(c, -1, "start failed")
		log.Errorf("failed to run tailscale start: %s", err)
		return
	}

	if !utils.IsGoMemLimitExist() {
		_ = utils.SetGoMemLimit(GoMemLimit)
	}

	rsp.OkRsp(c)
	log.Debugf("tailscale start successfully")
}

func (s *Service) Restart(c *gin.Context) {
	var rsp proto.Response

	err := NewCli().Restart()
	if err != nil {
		rsp.ErrRsp(c, -1, "restart failed")
		log.Errorf("failed to run tailscale restart: %s", err)
		return
	}

	rsp.OkRsp(c)
	log.Debugf("tailscale restart successfully")
}

func (s *Service) Stop(c *gin.Context) {
	var rsp proto.Response

	err := NewCli().Stop()
	if err != nil {
		rsp.ErrRsp(c, -1, "stop failed")
		log.Errorf("failed to run tailscale stop: %s", err)
		return
	}

	_ = utils.DelGoMemLimit()

	rsp.OkRsp(c)
	log.Debugf("tailscale stop successfully")
}

func (s *Service) Up(c *gin.Context) {
	var rsp proto.Response

	err := NewCli().Up()
	if err != nil {
		rsp.ErrRsp(c, -1, "tailscale up failed")
		log.Errorf("failed to run tailscale up: %s", err)
		return
	}

	rsp.OkRsp(c)
	log.Debugf("run tailscale up successfully")
}

func (s *Service) Down(c *gin.Context) {
	var rsp proto.Response

	err := NewCli().Down()
	if err != nil {
		rsp.ErrRsp(c, -1, "tailscale down failed")
		log.Errorf("failed to run tailscale down: %s", err)
		return
	}

	rsp.OkRsp(c)
	log.Debugf("run tailscale down successfully")
}

func (s *Service) Login(c *gin.Context) {
	var rsp proto.Response

	// check tailscale status
	cli := NewCli()
	status, err := cli.Status()
	if err != nil {
		_ = cli.Start()
		status, err = cli.Status()
	}

	if err != nil {
		log.Errorf("failed to get tailscale status: %s", err)
		rsp.ErrRsp(c, -1, "unknown status")
		return
	}

	if status.BackendState == "Running" {
		rsp.OkRspWithData(c, &proto.LoginTailscaleRsp{})
		return
	}

	// get login url
	url, err := cli.Login()
	if err != nil {
		log.Errorf("failed to run tailscale login: %s", err)
		rsp.ErrRsp(c, -2, "login failed")
		return
	}

	if !utils.IsGoMemLimitExist() {
		_ = utils.SetGoMemLimit(GoMemLimit)
	}

	rsp.OkRspWithData(c, &proto.LoginTailscaleRsp{
		Url: url,
	})

	log.Debugf("tailscale login url: %s", url)
}

func (s *Service) Logout(c *gin.Context) {
	var rsp proto.Response

	err := NewCli().Logout()
	if err != nil {
		rsp.ErrRsp(c, -1, "logout failed")
		log.Errorf("failed to run tailscale logout: %s", err)
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
