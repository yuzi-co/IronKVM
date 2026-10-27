package router

import (
	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/common"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/service/dashcam"
	"NanoKVM-Server/service/events"
	"NanoKVM-Server/service/stream"
	"NanoKVM-Server/service/stream/egress"
)

// dashcamFile and dashcamDir are variables so tests keep off /etc/kvm and
// /data.
var (
	dashcamFile = dashcam.File
	dashcamDir  = dashcam.Dir
)

// dashcamRouter starts the dashcam (ironkvm-dist#119) when it is on, and
// serves its clips and settings.
//
// An administrator or a user may save a clip and fetch one, by session or by
// API key, as they may press the power button; a viewer may not, and a
// view-only key acts as one. Deleting a clip and the settings are an
// administrator's.
func dashcamRouter(r *gin.Engine) {
	service := dashcam.New(dashcam.Deps{
		Load: func() (dashcam.Settings, error) { return dashcam.Load(dashcamFile) },
		Save: func(s dashcam.Settings) error { return dashcam.Save(dashcamFile, s) },
		Subscribe: func() dashcam.Reader {
			// Video only: the clip has no sound, so the reader does not
			// start the audio capture.
			return egress.Shared().Subscribe("dashcam", false)
		},
		RequestKeyframe: func() { stream.RequestKeyframe(stream.KeyframeReasonDashcam) },
		VideoAvailable:  func() bool { return !common.VideoDisabled() },
		Bus:             events.Default,
		Record:          audit.Record,
		Dir:             dashcamDir,
	})
	service.Start()

	api := r.Group("/api/dashcam").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin, authn.RoleUser),
	)
	api.GET("", service.GetState)
	api.POST("/save", service.PostSave)
	api.GET("/clips", service.GetClips)
	api.GET("/clips/:name", service.GetClip)

	admin := r.Group("/api/dashcam").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	admin.DELETE("/clips/:name", service.DeleteClip)
	admin.GET("/settings", service.GetSettings)
	admin.POST("/settings", service.SetSettings)
}
