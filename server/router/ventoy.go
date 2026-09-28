package router

import (
	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/storage"
	"NanoKVM-Server/service/ventoy"
)

func ventoyRouter(r *gin.Engine) {
	service := ventoy.New(ventoy.Deps{
		ImageDir: storage.ImageDirectory,
		OnData:   addon.OnData,
	})
	service.Start()

	// The image manager's Ventoy section.
	admin := r.Group("/api/ventoy").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	admin.GET("/status", service.GetStatus)
	admin.POST("/images", service.SetImages)
	admin.POST("/install", service.Install)
	admin.POST("/uninstall", service.Uninstall)
	admin.POST("/insert", service.Insert)
	admin.POST("/eject", service.Eject)
}
