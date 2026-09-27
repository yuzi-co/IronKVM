package netbird

import (
	"errors"
	"io"
	"os"
	"strings"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Service struct{}

// The update check's cache, the lookup and the installer. Variables so the
// tests need neither the network nor apk.
var (
	updates        = &vpn.VersionCache{TTL: vpn.UpdateTTL}
	fetchLatest    = latestVersion
	installPackage = install
)

// busy keeps install, update and uninstall apart: they share the fetch
// workspace and the binary.
var busy vpn.Busy

func NewService() *Service {
	return &Service{}
}

// Install fetches the package without the VPN lock, which a download of
// minutes must not hold, and starts the daemon under it, after checking again.
func (s *Service) Install(c *gin.Context) {
	var rsp proto.Response

	end := busy.Begin(c, addon.NetBird)
	if end == nil {
		return
	}
	defer end()

	if vpn.Refuse(c, addon.NetBird) {
		return
	}

	if !isInstalled() {
		if err := installPackage(); err != nil {
			log.Errorf("failed to install netbird: %s", err)
			rsp.ErrRsp(c, -1, vpn.Message("install failed", err))
			return
		}

		if err := vpn.StartChecked(addon.NetBird, NewCli().Start); err != nil {
			log.Errorf("failed to start netbird after install: %s", err)
			rsp.ErrRsp(c, -1, vpn.Message("NetBird is installed but did not start", err))
			return
		}
	}

	rsp.OkRsp(c)
	log.Debugf("install netbird successfully")
}

// Uninstall removes the binary and the add-on. The identity on /data stays, as
// Tailscale's does, so a reinstall is the same peer. A daemon that does not
// stop keeps its binary.
func (s *Service) Uninstall(c *gin.Context) {
	var rsp proto.Response

	end := busy.Begin(c, addon.NetBird)
	if end == nil {
		return
	}
	defer end()

	if err := NewCli().Stop(); err != nil {
		log.Errorf("failed to stop netbird before uninstall: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("stop failed", err))
		return
	}
	if err := addon.SetBoot(addon.NetBird, false); err != nil {
		log.Errorf("failed to turn off netbird at boot: %s", err)
	}
	if addon.OnData() {
		if err := addon.Remove(addonSpec()); err != nil {
			log.Errorf("failed to remove the netbird add-on: %s", err)
		}
	}
	_ = os.Remove(NetbirdPath)

	rsp.OkRsp(c)
	log.Debugf("uninstall netbird successfully")
}

func (s *Service) Start(c *gin.Context) {
	var rsp proto.Response

	if !vpn.Guard(c, addon.NetBird, "start failed", NewCli().Start) {
		return
	}
	rsp.OkRsp(c)
}

// Restart of a stopped daemon is a start, and goes through the same check.
func (s *Service) Restart(c *gin.Context) {
	var rsp proto.Response

	defer addon.LockVPN()()
	if !addon.Running(addon.NetBird) && vpn.Refuse(c, addon.NetBird) {
		return
	}
	if err := NewCli().Restart(); err != nil {
		log.Errorf("failed to restart netbird: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("restart failed", err))
		return
	}
	rsp.OkRsp(c)
}

func (s *Service) Stop(c *gin.Context) {
	var rsp proto.Response

	if err := NewCli().Stop(); err != nil {
		log.Errorf("failed to stop netbird: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("stop failed", err))
		return
	}
	rsp.OkRsp(c)
}

func (s *Service) Up(c *gin.Context) {
	var rsp proto.Response

	if !vpn.Guard(c, addon.NetBird, "netbird up failed", NewCli().Up) {
		return
	}
	rsp.OkRsp(c)
}

func (s *Service) Down(c *gin.Context) {
	var rsp proto.Response

	if err := NewCli().Down(); err != nil {
		log.Errorf("failed to run netbird down: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("netbird down failed", err))
		return
	}
	rsp.OkRsp(c)
}

// Login joins with {setupKey}, or with an empty body starts an SSO login and
// returns its URL. The key is never logged.
func (s *Service) Login(c *gin.Context) {
	var rsp proto.Response
	var req proto.NetbirdLoginReq

	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	key := ""
	if strings.TrimSpace(req.SetupKey) != "" {
		var err error
		if key, err = NormalizeSetupKey(req.SetupKey); err != nil {
			rsp.ErrRsp(c, -1, err.Error())
			return
		}
	}

	// The check, the start and the join hold the lock together.
	defer addon.LockVPN()()
	if vpn.Refuse(c, addon.NetBird) {
		return
	}

	cli := NewCli()
	if !addon.Running(addon.NetBird) {
		if err := cli.Start(); err != nil {
			rsp.ErrRsp(c, -1, vpn.Message("start failed", err))
			return
		}
	}

	if key != "" {
		if err := cli.JoinWithSetupKey(key); err != nil {
			log.Errorf("failed to join netbird with a setup key: %s", err)
			rsp.ErrRsp(c, -2, vpn.Message("join failed", err))
			return
		}
		rsp.OkRspWithData(c, &proto.VpnLoginRsp{})
		return
	}

	url, err := cli.LoginSSO()
	if err != nil {
		log.Errorf("failed to start the netbird sso login: %s", err)
		rsp.ErrRsp(c, -2, vpn.Message("login failed", err))
		return
	}
	rsp.OkRspWithData(c, &proto.VpnLoginRsp{Url: url})
}

// Logout is netbird deregister: it removes the peer from the account. The page
// warns before it calls this.
func (s *Service) Logout(c *gin.Context) {
	var rsp proto.Response

	if err := NewCli().Deregister(); err != nil {
		log.Errorf("failed to deregister netbird: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("deregister failed", err))
		return
	}
	rsp.OkRsp(c)
}

// GetStatus asks the CLI only while the daemon runs; without it the CLI would
// only fail to connect.
func (s *Service) GetStatus(c *gin.Context) {
	var rsp proto.Response

	st := proto.VpnStatus{State: proto.VpnNotInstall}
	if isInstalled() {
		st = proto.VpnStatus{State: proto.VpnNotRunning}
		if addon.Running(addon.NetBird) {
			if nb, err := NewCli().Status(); err != nil {
				log.Debugf("failed to get netbird status: %s", err)
			} else if st, err = toVpnStatus(nb); err != nil {
				log.Errorf("%s", err)
				rsp.ErrRsp(c, -1, err.Error())
				return
			}
		}
	}

	vpn.Fill(&st, addon.NetBird)
	rsp.OkRspWithData(c, &st)
}

// Boot turns start at boot on or off.
func (s *Service) Boot(c *gin.Context) {
	vpn.Boot(c, addon.NetBird, isInstalled())
}

// GetUpdate reports the installed version and the one Alpine offers, which is
// looked up at most once an hour.
func (s *Service) GetUpdate(c *gin.Context) {
	var rsp proto.Response

	if !isInstalled() {
		rsp.ErrRsp(c, -1, "netbird is not installed")
		return
	}
	current, err := NewCli().Version()
	if err != nil {
		rsp.ErrRsp(c, -1, vpn.Message("version failed", err))
		return
	}
	latest, err := updates.Latest(fetchLatest)
	if err != nil {
		rsp.ErrRsp(c, -1, vpn.Message("update check failed", err))
		return
	}
	rsp.OkRspWithData(c, &proto.VpnUpdateRsp{Current: current, Latest: latest})
}

// Update installs Alpine's current package over the installed binary. The
// identity stays on /data, and the daemon runs again only if it ran before,
// and only if Tailscale did not start meanwhile.
func (s *Service) Update(c *gin.Context) {
	var rsp proto.Response

	end := busy.Begin(c, addon.NetBird)
	if end == nil {
		return
	}
	defer end()

	if !isInstalled() {
		rsp.ErrRsp(c, -1, "netbird is not installed")
		return
	}

	cli := NewCli()
	wasRunning := addon.Running(addon.NetBird)
	if wasRunning {
		if err := cli.Stop(); err != nil {
			rsp.ErrRsp(c, -1, vpn.Message("stop failed", err))
			return
		}
	}

	err := installPackage()
	updates.Reset()

	if wasRunning {
		if startErr := vpn.StartChecked(addon.NetBird, cli.Start); startErr != nil && err == nil {
			err = startErr
		}
	}
	if err != nil {
		log.Errorf("failed to update netbird: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("update failed", err))
		return
	}
	rsp.OkRsp(c)
}
