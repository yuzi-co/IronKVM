package router

import (
	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/urlmedia"
)

// urlMediaRouter serves images to the drives straight from a URL, through
// an nbd device (ironkvm-dist#106). Only a kernel with NBD can: the status
// says so, with the reason, on one without.
func urlMediaRouter(r *gin.Engine) {
	service := urlmedia.Default()
	service.Start()

	api := r.Group("/api").Use(middleware.CheckToken(), middleware.RequireRole(authn.RoleAdmin))
	api.GET("/storage/url", service.GetStatus)       // what the drives serve from URLs
	api.POST("/storage/url/mount", service.MountURL) // serve an image from a URL in a drive
	api.POST("/storage/url/eject", service.EjectURL) // take it out
}
