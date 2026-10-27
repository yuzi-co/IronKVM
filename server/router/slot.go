package router

import (
	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/slot"

	"github.com/gin-gonic/gin"
)

// slotRouter is the A/B slot install and the boot decisions that follow it
// (ironkvm-dist#165). Writing a slot and choosing what the board boots are an
// administrator's, like the application update.
func slotRouter(r *gin.Engine) {
	service := slot.NewService()
	api := r.Group("/api").Use(middleware.CheckToken(), middleware.RequireRole(authn.RoleAdmin))

	api.GET("/slot", service.GetStatus)           // slot status: running, trusted, each slot's record
	api.GET("/slot/images", service.GetImages)    // slot images already on /data
	api.GET("/slot/install", service.GetProgress) // progress of the running or last install
	api.POST("/slot/install", service.Install)    // install an image from a URL or a file on /data
	api.POST("/slot/try", service.Try)            // boot a slot once, on the next boot
	api.POST("/slot/confirm", service.Confirm)    // make the running slot trusted
	api.POST("/slot/revert", service.Revert)      // trust the previous slot again, from the next boot
}
