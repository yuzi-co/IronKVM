package router

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/common"
	"NanoKVM-Server/middleware"
	mcpservice "NanoKVM-Server/service/mcp"
	kvmcapture "NanoKVM-Server/service/mcp/capture/kvm"
	"NanoKVM-Server/service/vm"
	"NanoKVM-Server/service/watchdog"
	"NanoKVM-Server/utils"
)

// watchdogLogDir is a variable so tests can keep the log off /data.
var watchdogLogDir = watchdog.LogDir

// frameUnchanged is the capture library's answer when MJPEG frame detection
// found the picture the same as the last frame.
const frameUnchanged = 5

func watchdogRouter(r *gin.Engine) {
	service := watchdog.New(watchdog.Deps{
		Settings:    watchdog.CurrentSettings,
		SetSettings: watchdog.ApplySettings,

		CaptureDisabled: utils.IsHdmiDisabled,
		Signal: func() bool {
			return !utils.IsHdmiDisabled() && common.GetKvmVision().HasHDMISignal()
		},
		Capture: watchdogCapture(kvmcapture.New()),

		PowerLEDConnected: vm.PowerLEDConnected,
		PowerLED:          vm.PowerLED,

		Ping:        watchdog.Ping,
		PressButton: vm.PressButton,

		Log: watchdog.OpenLog(watchdogLogDir),
	})
	service.Start()

	// The web UI's Watchdog page.
	admin := r.Group("/api/watchdog").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	admin.GET("/settings", service.GetSettings)
	admin.POST("/settings", service.SetSettings)
	admin.GET("/state", service.GetState)
	admin.GET("/log", service.GetLog)
	admin.GET("/log/:id/screenshot", service.GetScreenshot)
}

// watchdogCapture takes the watchdog's samples and screenshots through the
// same path as the MCP screenshot tool: at the stream's current size, so a
// viewer's stream is never resized, and under the capture lease, so capture
// resumes if the idle timer stopped it.
func watchdogCapture(snapshotter mcpservice.Snapshotter) func(context.Context, int) ([]byte, error) {
	return func(ctx context.Context, quality int) ([]byte, error) {
		timeoutMS := 1000
		snap, err := snapshotter.Capture(ctx, mcpservice.SnapshotRequest{
			Quality:   quality,
			TimeoutMS: &timeoutMS,
		})
		if err != nil {
			return nil, err
		}
		if snap.OK {
			return snap.JPEG, nil
		}
		if snap.RetCode == frameUnchanged {
			return nil, watchdog.ErrFrameUnchanged
		}
		return nil, errors.New(snap.Message)
	}
}
