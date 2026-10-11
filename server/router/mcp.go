package router

import (
	"net/http"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/controlmode"
	"NanoKVM-Server/service/hid"
	mcpservice "NanoKVM-Server/service/mcp"
	"NanoKVM-Server/service/mcp/capture/kvm"
	"NanoKVM-Server/service/picoclaw"
	"NanoKVM-Server/service/storage"
	"NanoKVM-Server/service/vm"

	"github.com/gin-gonic/gin"
)

func mcpRouter(r *gin.Engine, control *controlmode.Manager, picoclawService *picoclaw.Service) {
	service := mcpservice.NewServiceWithPreempt(
		control,
		picoclawService.PreemptControlLeasesForMCP,
		picoclawService.StopRuntimeForMCP,
		hid.ReleaseAllHIDStateBestEffort,
		func(status controlmode.Status) {
			picoclawService.PublishControlModeChangedFrom(status, "mcp_config")
		},
	)
	management := r.Group("/api/mcp").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	management.GET("/config", service.GetConfig)
	management.POST("/config", service.SetConfig)
	management.POST("/key/regenerate", service.RegenerateAPIKey)

	handler := mcpservice.NewMCPHandler(control, kvmcapture.New(), mcpDevices())
	handler = http.NewCrossOriginProtection().Handler(handler)
	r.Any("/api/mcp", mcpservice.APIKeyMiddleware(control), gin.WrapH(handler))
}

// mcpDevices gives MCP the power button and the virtual media. The tools
// exist only behind the power and media tiers, which are off by default, and
// the audit log records each call, so the press goes through vm.PressButton
// directly rather than the audited wrapper the other callers use.
func mcpDevices() mcpservice.Devices {
	return mcpservice.Devices{
		PressButton: vm.PressButton,
		ListMedia: func() ([]mcpservice.MediaDrive, []string, error) {
			drives, err := storage.Drives()
			if err != nil {
				return nil, nil, err
			}
			images, err := storage.Images()
			if err != nil {
				return nil, nil, err
			}
			out := make([]mcpservice.MediaDrive, len(drives))
			for i, d := range drives {
				out[i] = mcpservice.MediaDrive{ID: d.ID, File: d.File, ReadOnly: d.ReadOnly}
			}
			return out, images, nil
		},
		InsertMedia: storage.Insert,
		EjectMedia:  storage.Eject,
	}
}
