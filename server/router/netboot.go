package router

import (
	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/netboot"
	"NanoKVM-Server/service/storage"
	"NanoKVM-Server/service/vm"
)

func netbootRouter(r *gin.Engine) {
	service := netboot.New(netboot.Deps{
		Settings:     netboot.CurrentSettings,
		SaveSettings: netboot.SaveSettings,
		Link:         netbootLink,
		ImageDir:     storage.ImageDirectory,
	})
	service.Start()

	// The web UI's Network boot page. The menu and the images are served
	// by the service's own listener on the USB link, never here.
	admin := r.Group("/api/netboot").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	admin.GET("/status", service.GetStatus)
	admin.POST("/settings", service.SetSettings)
	admin.POST("/install", service.Install)
	admin.POST("/uninstall", service.Uninstall)
}

func netbootLink() netboot.Link {
	link := vm.CurrentUSBLink()

	return netboot.Link{
		Mode:   link.Mode,
		Prefix: link.Prefix,
		Board:  link.Board,
		Host:   link.Host,
		Active: link.Active,
	}
}
