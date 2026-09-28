package router

import (
	"crypto/tls"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/common"
	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/redfish"
	"NanoKVM-Server/service/stream/mjpeg"
	"NanoKVM-Server/service/vnc"
	"NanoKVM-Server/utils"
)

func vncRouter(r *gin.Engine) {
	service := vnc.New(vnc.Deps{
		Settings:    vnc.CurrentSettings,
		SetSettings: vnc.ApplySettings,
		ReservedPorts: func() []int {
			conf := config.GetInstance()
			return []int{conf.Port.Http, conf.Port.Https}
		},

		Subscribe: func() vnc.FrameSubscription {
			return mjpeg.Subscribe()
		},
		ScreenSize: func() (int, int) {
			values := common.GetScreen().Snapshot()
			return int(values.Width), int(values.Height)
		},
		NewInput: vnc.NewHIDInput,

		Accounts: authn.DefaultStore,
		Limiter:  redfish.LoginLimiter{},
		// The web login waits as long after a wrong password.
		FailureDelay: 2 * time.Second,
		Password:     vnc.FilePassword{Path: vnc.PasswordFile},
		Certificate:  vncCertificate,

		Name: vncDesktopName,
	})
	if err := service.Apply(); err != nil {
		log.Errorf("vnc: %s", err)
	}

	// The web UI's VNC page.
	admin := r.Group("/api/vnc").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	admin.GET("/settings", service.GetSettings)
	admin.POST("/settings", service.SetSettings)
	admin.GET("/state", service.GetState)
	admin.POST("/disconnect", service.EndSession)
}

// vncCertificate loads the certificate the HTTPS listener uses. A board that
// serves plain HTTP may have none yet, and then the shipped one is made the
// same way the HTTPS switch makes it.
func vncCertificate() (*tls.Certificate, error) {
	conf := config.GetInstance()
	if cert, err := tls.LoadX509KeyPair(conf.Cert.Crt, conf.Cert.Key); err == nil {
		return &cert, nil
	}

	if err := utils.EnsureCert(); err != nil {
		return nil, err
	}
	cert, err := tls.LoadX509KeyPair(utils.CertFile, utils.KeyFile)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

// vncDesktopName is the name a VNC client shows for the session: the board's
// host name.
func vncDesktopName() string {
	name, err := os.Hostname()
	if err != nil || name == "" {
		return "IronKVM"
	}
	return name
}
