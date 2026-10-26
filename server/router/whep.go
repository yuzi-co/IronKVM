package router

import (
	"sync"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/whep"
)

// whepRouter serves the WHEP endpoint (ironkvm-dist#145) and, to an
// administrator, its switch.
func whepRouter(r *gin.Engine) {
	service := whep.New(whep.Deps{
		Enabled:    whepEnabled,
		SetEnabled: setWHEPEnabled,
	})

	// The endpoint takes its own Bearer token and never the cookie.
	service.Register(r)

	admin := r.Group("/api").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	admin.GET("/stream/whep", service.GetSettings)
	admin.POST("/stream/whep", service.SetSettings)
}

// whepSettingMu guards config's WHEP.Enabled, which the owner can change
// while every WHEP request reads it.
var whepSettingMu sync.RWMutex

func whepEnabled() bool {
	whepSettingMu.RLock()
	defer whepSettingMu.RUnlock()

	return config.GetInstance().WHEP.Enabled
}

// saveWHEPSetting writes the setting to server.yaml. Tests replace it.
var saveWHEPSetting = func(enabled bool) error {
	conf, err := config.Read()
	if err != nil {
		return err
	}

	conf.WHEP.Enabled = enabled

	return config.Write(conf)
}

func setWHEPEnabled(enabled bool) error {
	whepSettingMu.Lock()
	defer whepSettingMu.Unlock()

	if err := saveWHEPSetting(enabled); err != nil {
		return err
	}
	config.GetInstance().WHEP.Enabled = enabled
	return nil
}
