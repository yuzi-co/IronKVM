package vm

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/sys/unix"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/storage"
)

// healthStoragePath is the filesystem the image library lives on. The
// uploads, downloads and the add-ons all write there too, so it is the one
// that fills up.
var healthStoragePath = storage.ImageDirectory

// healthStatFS reads the space on a filesystem. A variable so the tests can
// answer for a path that is not a mount on the test host.
var healthStatFS = func(path string) (total, available uint64, err error) {
	var s unix.Statfs_t
	if err := unix.Statfs(path, &s); err != nil {
		return 0, 0, err
	}
	return s.Blocks * uint64(s.Bsize), s.Bavail * uint64(s.Bsize), nil
}

// healthVpns are the VPN add-ons the health check looks at. A variable so the
// tests can point it at daemons of their own.
var healthVpns = func() []addon.Daemon { return []addon.Daemon{addon.Tailscale, addon.NetBird} }

// GetHealth reports the few board readings the toolbar turns into alerts: the
// SoC temperature, the space left for images, and whether each VPN the owner
// set to start at boot is running. The browser decides what counts as wrong,
// so the thresholds live beside the words that explain them.
//
// Every reading is a file read or a statfs: the toolbar asks every half
// minute, and running `tailscale status` for it would cost more than the
// answer is worth. A VPN counts as up while its daemon runs, which misses a
// daemon that runs but has lost its connection; the VPN page tells those
// apart.
func (s *Service) GetHealth(c *gin.Context) {
	var rsp proto.Response

	data := &proto.GetHealthRsp{Vpn: []proto.HealthVpn{}}

	if celsius, ok := readCpuTemp(); ok {
		data.Temperature = &celsius
	}

	if total, available, err := healthStatFS(healthStoragePath); err == nil && total > 0 {
		data.Storage = &proto.HealthStorage{
			Path:      healthStoragePath,
			Total:     total,
			Available: available,
		}
	}

	for _, d := range healthVpns() {
		if !addon.BootEnabled(d) {
			continue
		}
		data.Vpn = append(data.Vpn, proto.HealthVpn{
			Name:    d.Name,
			Title:   d.Title,
			Running: addon.Running(d),
		})
	}

	rsp.OkRspWithData(c, data)
}
