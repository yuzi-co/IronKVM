package router

import (
	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/hostpower"
	"NanoKVM-Server/service/nut"
	"NanoKVM-Server/service/vm"
)

func nutRouter(r *gin.Engine) {
	password := nut.FilePassword{Path: nut.PasswordFile}

	service := nut.New(nut.Deps{
		Settings:    nut.CurrentSettings,
		SetSettings: nut.ApplySettings,
		Password:    password.Get,
		SetPassword: password.Set,

		Power: hostpower.New(hostpower.Deps{
			PressButton:       vm.PressButton,
			PowerLEDConnected: vm.PowerLEDConnected,
			PowerLED:          vm.PowerLED,
		}),
	})
	service.Start()

	// The state is for the status bar, which every viewer has.
	r.Group("/api/nut").Use(middleware.CheckToken()).GET("/state", service.GetState)

	// The web UI's NUT page.
	admin := r.Group("/api/nut").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	admin.GET("/settings", service.GetSettings)
	admin.POST("/settings", service.SetSettings)
}
