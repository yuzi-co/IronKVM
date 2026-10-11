package router

import (
	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/sequence"
)

// sequenceRouter serves the key auto-tap and the macros, which the server
// plays as timed input sequences (service/sequence).
func sequenceRouter(r *gin.Engine) {
	sequence.SetATX(auditedPress("macro"))
	service := sequence.NewService()
	api := r.Group("/api").Use(middleware.CheckToken())

	api.POST("/hid/autotap", service.StartAutoTap) // start pressing a key over and over
	api.POST("/hid/macro/play", service.PlayMacro) // play a saved macro, or the steps sent
	api.GET("/hid/sequence", service.GetStatus)    // the running sequence, or the last one
	api.POST("/hid/sequence/stop", service.Stop)   // stop the running sequence
	api.GET("/hid/macros", service.GetMacros)      // the saved macros

	// Saved macros are shared by every user, as the shortcuts are, so only
	// an admin changes them.
	admin := r.Group("/api").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	admin.POST("/hid/macro", service.SaveMacro)
	admin.DELETE("/hid/macro", service.DeleteMacro)
}
