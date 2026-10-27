package router

import (
	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/slot"

	"github.com/gin-gonic/gin"
)

// slotRouter is the A/B slot install and the boot decisions that follow it
// (ironkvm-dist#165). Writing a slot and choosing what the board boots are an
// administrator's, like the application update. They also need a signed-in
// session: an API key, which a script or another machine holds, can read the
// slots and follow an install but not write a slot or change what boots.
func slotRouter(r *gin.Engine) {
	service := slot.NewService()
	read := r.Group("/api").Use(middleware.CheckToken(), middleware.RequireRole(authn.RoleAdmin))
	write := r.Group("/api").Use(middleware.CheckSession(), middleware.RequireRole(authn.RoleAdmin))

	read.GET("/slot", service.GetStatus)           // slot status: running, trusted, each slot's record
	read.GET("/slot/images", service.GetImages)    // slot images already on /data
	read.GET("/slot/install", service.GetProgress) // progress of the running or last install
	write.POST("/slot/install", service.Install)   // install an image from a URL or a file on /data
	write.POST("/slot/try", service.Try)           // boot a slot once, on the next boot
	write.POST("/slot/confirm", service.Confirm)   // make the running slot trusted
	write.POST("/slot/revert", service.Revert)     // trust the previous slot again, from the next boot
}
