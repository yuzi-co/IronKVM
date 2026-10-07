package router

import (
	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/roommic"
)

// roomMicRouter serves the room microphone's state to every viewer and its
// settings to administrators. Switching the microphone on and off for one
// viewer happens on the video connection, which carries the sound.
func roomMicRouter(r *gin.Engine) {
	api := r.Group("/api").Use(middleware.CheckToken())
	admin := r.Group("/api").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)

	api.GET("/room-mic", roommic.GetStatus)
	admin.POST("/room-mic", roommic.SetSettings)
}
