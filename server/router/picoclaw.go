package router

import (
	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/agent"
	"NanoKVM-Server/service/audit"
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

// picoclawRouter registers PicoClaw's own routes: its model and profile
// settings, its session release, and the KVM tools it calls back over
// loopback.
func picoclawRouter(r *gin.Engine, service *picoclaw.Service) {
	frontendAPI := picoclawFrontendAPI(r)
	localAPI := r.Group(picoclawBasePath).Use(middleware.CheckLoopbackInternalToken())

	localAPI.GET(picoclawScreenshotPath, audit.AgentGate(service.AuditPrincipal, "screenshot", audit.TierObserve), service.Screenshot)
	localAPI.POST(picoclawActionsPath, audit.AgentGate(service.AuditPrincipal, "actions", audit.TierInput), service.Actions)
	localAPI.POST(picoclawMCPPath, service.MCPHandler)
	localAPI.POST(picoclawLoadImagePath, audit.AgentGate(service.AuditPrincipal, "load-image", audit.TierObserve), service.LoadImage)
	localAPI.GET(picoclawRuntimeSessionPath, service.GetRuntimeSession)

	frontendAPI.POST(picoclawModelConfigPath, service.UpdateModelConfig)
	frontendAPI.POST(picoclawAgentProfilePath, service.UpdateAgentProfile)
	frontendAPI.DELETE(picoclawRuntimeSessionPath, service.ReleaseRuntimeSession)
}

// runtimeRouter registers the runtime routes. They use only the agent
// lifecycle, and keep the /api/picoclaw paths the web UI has always used.
func runtimeRouter(r *gin.Engine, lifecycle *agent.Lifecycle) {
	frontendAPI := picoclawFrontendAPI(r)
	frontendAPI.GET(picoclawRuntimeStatusPath, lifecycle.Status)
	frontendAPI.POST(picoclawRuntimeInstallPath, lifecycle.Install)
	frontendAPI.POST(picoclawRuntimeUninstallPath, lifecycle.Uninstall)
	frontendAPI.POST(picoclawRuntimeStartPath, lifecycle.Start)
	frontendAPI.POST(picoclawRuntimeStopPath, lifecycle.Stop)
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
