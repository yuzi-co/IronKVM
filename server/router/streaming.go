package router

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/allowlist"
	"NanoKVM-Server/service/rtsp"
	"NanoKVM-Server/service/srt"
	"NanoKVM-Server/service/stream/audio"
	"NanoKVM-Server/service/stream/egress"
)

// viewTokenAuthenticator checks an RTSP or SRT player's token with the rules
// WHEP's Bearer token follows.
func viewTokenAuthenticator(token string) (egress.Viewer, bool) {
	viewer, status := middleware.AuthenticateViewToken(token)
	if status != 0 {
		return egress.Viewer{}, false
	}
	p := viewer.Principal
	return egress.Viewer{
		Username: p.Username,
		Role:     string(p.Role),
		Via:      p.Via,
		Key:      p.Key,
		Valid:    viewer.Valid,
	}, true
}

func webServerPorts() []int {
	conf := config.GetInstance()
	return []int{conf.Port.Http, conf.Port.Https}
}

// rtspRouter starts the RTSP server (ironkvm-dist#142) when it is on, and
// serves its settings to an administrator.
func rtspRouter(r *gin.Engine) {
	service := rtsp.New(rtsp.Deps{
		Settings:      rtsp.CurrentSettings,
		SetSettings:   rtsp.ApplySettings,
		ReservedPorts: webServerPorts,
		// The allowed networks setting covers RTSP as it does the web UI.
		ClientAllowed:  allowlist.AllowedAddr,
		Authenticate:   viewTokenAuthenticator,
		Tap:            egress.Shared(),
		AudioAvailable: audio.Available,
	})
	if err := service.Apply(); err != nil {
		log.Errorf("rtsp: %s", err)
	}

	admin := r.Group("/api/stream").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	admin.GET("/rtsp", service.GetSettings)
	admin.POST("/rtsp", service.SetSettings)
}

// srtRouter starts the SRT listener (ironkvm-dist#143) when it is on, and
// serves its settings to an administrator.
func srtRouter(r *gin.Engine) {
	service := srt.New(srt.Deps{
		Settings:    srt.CurrentSettings,
		SetSettings: srt.ApplySettings,
		Passphrase:  srt.FilePassphrase{Path: srt.PassphraseFile},
		// The allowed networks setting covers SRT as it does the web UI.
		ClientAllowed:  allowlist.AllowedAddr,
		Authenticate:   viewTokenAuthenticator,
		Tap:            egress.Shared(),
		AudioAvailable: audio.Available,
	})
	if err := service.Apply(); err != nil {
		log.Errorf("srt: %s", err)
	}

	admin := r.Group("/api/stream").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	admin.GET("/srt", service.GetSettings)
	admin.POST("/srt", service.SetSettings)
}
