package router

import (
	"net/netip"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/events"
	"NanoKVM-Server/service/hostlog"
	"NanoKVM-Server/service/vm"
)

// hostlogDiskDir is a variable so tests can keep the log off /data.
var hostlogDiskDir = hostlog.DiskDir

func hostlogRouter(r *gin.Engine) {
	service := hostlog.New(hostlog.Deps{
		Settings:    hostlog.CurrentSettings,
		SetSettings: hostlog.ApplySettings,

		DefaultSources: hostlogDefaultSources,
		USBBoard: func() (netip.Addr, bool) {
			link := vm.CurrentUSBLink()
			return link.Board, link.Active && link.Board.IsValid()
		},
		Publish: events.Publish,

		DiskDir: hostlogDiskDir,
	})
	service.Start()

	// The web UI's Host log page. The stream is a websocket, which the
	// browser opens with the session cookie.
	admin := r.Group("/api/hostlog").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	admin.GET("/settings", service.GetSettings)
	admin.POST("/settings", service.SetSettings)
	admin.GET("/state", service.GetState)
	admin.GET("/lines", service.GetLines)
	admin.GET("/download", service.Download)
	admin.DELETE("/lines", service.ClearLog)
	admin.GET("/stream", service.Stream)
}

// hostlogDefaultSources is whom the receiver hears when the owner names
// nobody: the host on the USB network link while it is up, else the host the
// watchdog pings. With neither it hears any sender.
func hostlogDefaultSources() []netip.Prefix {
	if link := vm.CurrentUSBLink(); link.Active && link.Host.IsValid() {
		return []netip.Prefix{netip.PrefixFrom(link.Host, link.Host.BitLen())}
	}
	if host := config.GetInstance().Watchdog.PingHost; host != "" {
		if a, err := netip.ParseAddr(host); err == nil {
			return []netip.Prefix{netip.PrefixFrom(a.Unmap(), a.Unmap().BitLen())}
		}
	}
	return nil
}
