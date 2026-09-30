package router

import (
	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/extensions/netbird"
	"NanoKVM-Server/service/extensions/tailscale"

	"github.com/gin-gonic/gin"
)

func extensionsRouter(r *gin.Engine) {
	api := r.Group("/api/extensions").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)

	ts := tailscale.NewService()

	api.POST("/tailscale/install", ts.Install)       // install tailscale
	api.POST("/tailscale/uninstall", ts.Uninstall)   // uninstall tailscale
	api.GET("/tailscale/status", ts.GetStatus)       // get tailscale status
	api.POST("/tailscale/up", ts.Up)                 // run tailscale up
	api.POST("/tailscale/down", ts.Down)             // run tailscale down
	api.POST("/tailscale/login", ts.Login)           // tailscale login
	api.POST("/tailscale/logout", ts.Logout)         // tailscale logout
	api.POST("/tailscale/start", ts.Start)           // tailscale start
	api.POST("/tailscale/stop", ts.Stop)             // tailscale stop
	api.POST("/tailscale/restart", ts.Restart)       // tailscale restart
	api.POST("/tailscale/boot", ts.Boot)             // tailscale start at boot on or off
	api.GET("/tailscale/update", ts.GetUpdate)       // tailscale current and latest version
	api.POST("/tailscale/update", ts.Update)         // install the latest tailscale
	api.POST("/tailscale/connect", ts.Connect)       // start if needed, then up
	api.POST("/tailscale/disconnect", ts.Disconnect) // down, then stop

	nb := netbird.NewService()

	api.POST("/netbird/install", nb.Install)       // install netbird from Alpine's package
	api.POST("/netbird/uninstall", nb.Uninstall)   // uninstall netbird
	api.GET("/netbird/status", nb.GetStatus)       // get netbird status
	api.POST("/netbird/up", nb.Up)                 // run netbird up
	api.POST("/netbird/down", nb.Down)             // run netbird down
	api.POST("/netbird/login", nb.Login)           // join with a setup key, or start an SSO login
	api.POST("/netbird/logout", nb.Logout)         // netbird deregister
	api.POST("/netbird/start", nb.Start)           // netbird start
	api.POST("/netbird/stop", nb.Stop)             // netbird stop
	api.POST("/netbird/restart", nb.Restart)       // netbird restart
	api.POST("/netbird/boot", nb.Boot)             // netbird start at boot on or off
	api.GET("/netbird/update", nb.GetUpdate)       // netbird current and latest version
	api.POST("/netbird/update", nb.Update)         // install the latest netbird
	api.POST("/netbird/connect", nb.Connect)       // start if needed, then up
	api.POST("/netbird/disconnect", nb.Disconnect) // down, then stop
}
