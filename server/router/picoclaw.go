package router

import (
	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/agent"
	"NanoKVM-Server/service/picoclaw"
)

const (
	picoclawBasePath             = "/api/picoclaw"
	picoclawModelConfigPath      = "/model/config"
	picoclawAgentProfilePath     = "/agent/profile"
	picoclawSessionsPath         = "/sessions"
	picoclawSessionByIDPath      = "/sessions/:id"
	picoclawRuntimeStatusPath    = "/runtime/status"
	picoclawRuntimeSessionPath   = "/runtime/session"
	picoclawRuntimeInstallPath   = "/runtime/install"
	picoclawRuntimeUninstallPath = "/runtime/uninstall"
	picoclawRuntimeStartPath     = "/runtime/start"
	picoclawRuntimeStopPath      = "/runtime/stop"
	picoclawGatewayWSPath        = "/gateway/ws"
	picoclawScreenshotPath       = "/screenshot"
	picoclawActionsPath          = "/actions"
	picoclawMCPPath              = "/mcp"
	picoclawLoadImagePath        = "/load-image"
)

var picoclawLoopbackHTTPAllowedPaths = []string{
	picoclawBasePath + picoclawMCPPath,
	picoclawBasePath + picoclawRuntimeSessionPath,
	picoclawBasePath + picoclawScreenshotPath,
	picoclawBasePath + picoclawActionsPath,
	picoclawBasePath + picoclawLoadImagePath,
}

func PicoclawLoopbackHTTPAllowedPaths() []string {
	return append([]string(nil), picoclawLoopbackHTTPAllowedPaths...)
}

// picoclawRouter registers PicoClaw's own routes: its runtime, model and
// profile settings, and the KVM tools it calls back over loopback.
func picoclawRouter(r *gin.Engine, service *picoclaw.Service) {
	frontendAPI := picoclawFrontendAPI(r)
	localAPI := r.Group(picoclawBasePath).Use(middleware.CheckLoopbackInternalToken())

	localAPI.GET(picoclawScreenshotPath, service.Screenshot)
	localAPI.POST(picoclawActionsPath, service.Actions)
	localAPI.POST(picoclawMCPPath, service.MCPHandler)
	localAPI.POST(picoclawLoadImagePath, service.LoadImage)
	localAPI.GET(picoclawRuntimeSessionPath, service.GetRuntimeSession)

	frontendAPI.POST(picoclawModelConfigPath, service.UpdateModelConfig)
	frontendAPI.POST(picoclawAgentProfilePath, service.UpdateAgentProfile)
	frontendAPI.GET(picoclawRuntimeStatusPath, service.GetRuntimeStatus)
	frontendAPI.DELETE(picoclawRuntimeSessionPath, service.ReleaseRuntimeSession)
	frontendAPI.POST(picoclawRuntimeInstallPath, service.InstallRuntime)
	frontendAPI.POST(picoclawRuntimeUninstallPath, service.UninstallRuntime)
	frontendAPI.POST(picoclawRuntimeStartPath, service.StartRuntime)
	frontendAPI.POST(picoclawRuntimeStopPath, service.StopRuntime)
}

// agentRouter registers the chat socket and the chat history routes. They use
// only the agent interface. They keep the /api/picoclaw paths the web UI has
// always used.
func agentRouter(r *gin.Engine, bridge *agent.Bridge, history agent.History) {
	frontendAPI := picoclawFrontendAPI(r)
	frontendAPI.GET(picoclawGatewayWSPath, bridge.Connect)
	frontendAPI.GET(picoclawSessionsPath, agent.ListSessionsHandler(history))
	frontendAPI.GET(picoclawSessionByIDPath, agent.ReadSessionHandler(history))
	frontendAPI.DELETE(picoclawSessionByIDPath, agent.DeleteSessionHandler(history))
}

func picoclawFrontendAPI(r *gin.Engine) gin.IRoutes {
	return r.Group(picoclawBasePath).Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
}
