package router

import (
	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/agent"
	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/service/picoclaw"
)

// The agent's routes live under agentBasePath, whichever agent runs. They
// lived under picoclawBasePath while PicoClaw was the only one, and stay
// there too: a web UI installed apart from the server, and a PicoClaw whose
// configuration still names the old MCP address, keep working. A UI uses the
// new paths only on a server that declares the agent-routes feature.
const (
	agentBasePath    = "/api/agent"
	picoclawBasePath = "/api/picoclaw"
)

// agentBasePaths are the bases every agent-neutral route is served under.
var agentBasePaths = []string{agentBasePath, picoclawBasePath}

const (
	picoclawModelConfigPath   = "/model/config"
	picoclawAgentProfilePath  = "/agent/profile"
	agentSessionsPath         = "/sessions"
	agentSessionByIDPath      = "/sessions/:id"
	agentRuntimeStatusPath    = "/runtime/status"
	agentRuntimeSessionPath   = "/runtime/session"
	agentRuntimeInstallPath   = "/runtime/install"
	agentRuntimeUninstallPath = "/runtime/uninstall"
	agentRuntimeStartPath     = "/runtime/start"
	agentRuntimeStopPath      = "/runtime/stop"
	agentScreenshotPath       = "/screenshot"
	agentActionsPath          = "/actions"
	agentMCPPath              = "/mcp"
	agentLoadImagePath        = "/load-image"
)

// chatSocketPaths are the chat socket's path under each base: "gateway" was
// PicoClaw's word for it.
var chatSocketPaths = map[string]string{
	agentBasePath:    "/chat/ws",
	picoclawBasePath: "/gateway/ws",
}

// agentLoopbackPaths are the routes the agent calls over loopback HTTP.
var agentLoopbackPaths = []string{
	agentMCPPath,
	agentRuntimeSessionPath,
	agentScreenshotPath,
	agentActionsPath,
	agentLoadImagePath,
}

func PicoclawLoopbackHTTPAllowedPaths() []string {
	paths := make([]string, 0, len(agentBasePaths)*len(agentLoopbackPaths))
	for _, base := range agentBasePaths {
		for _, path := range agentLoopbackPaths {
			paths = append(paths, base+path)
		}
	}
	return paths
}

// picoclawRouter registers PicoClaw's own settings routes, and the KVM tools
// and the session release, which the PicoClaw service serves for any agent.
func picoclawRouter(r *gin.Engine, service *picoclaw.Service) {
	settingsAPI := agentFrontendAPI(r, picoclawBasePath)
	settingsAPI.POST(picoclawModelConfigPath, service.UpdateModelConfig)
	settingsAPI.POST(picoclawAgentProfilePath, service.UpdateAgentProfile)

	for _, base := range agentBasePaths {
		frontendAPI := agentFrontendAPI(r, base)
		localAPI := r.Group(base).Use(middleware.CheckLoopbackInternalToken())

		localAPI.GET(agentScreenshotPath, audit.AgentGate(service.AuditPrincipal, "screenshot", audit.TierObserve), service.Screenshot)
		localAPI.POST(agentActionsPath, audit.AgentGate(service.AuditPrincipal, "actions", audit.TierInput), service.Actions)
		localAPI.POST(agentMCPPath, service.MCPHandler)
		localAPI.POST(agentLoadImagePath, audit.AgentGate(service.AuditPrincipal, "load-image", audit.TierObserve), service.LoadImage)
		localAPI.GET(agentRuntimeSessionPath, service.GetRuntimeSession)

		frontendAPI.DELETE(agentRuntimeSessionPath, service.ReleaseRuntimeSession)
	}
}

// runtimeRouter registers the runtime routes. They use only the agent
// lifecycle.
func runtimeRouter(r *gin.Engine, lifecycle *agent.Lifecycle) {
	for _, base := range agentBasePaths {
		frontendAPI := agentFrontendAPI(r, base)
		frontendAPI.GET(agentRuntimeStatusPath, lifecycle.Status)
		frontendAPI.POST(agentRuntimeInstallPath, lifecycle.Install)
		frontendAPI.POST(agentRuntimeUninstallPath, lifecycle.Uninstall)
		frontendAPI.POST(agentRuntimeStartPath, lifecycle.Start)
		frontendAPI.POST(agentRuntimeStopPath, lifecycle.Stop)
	}
}

// agentRouter registers the chat socket and the chat history routes. They use
// only the agent interface.
func agentRouter(r *gin.Engine, bridge *agent.Bridge, history agent.History) {
	for _, base := range agentBasePaths {
		frontendAPI := agentFrontendAPI(r, base)
		frontendAPI.GET(chatSocketPaths[base], bridge.Connect)
		frontendAPI.GET(agentSessionsPath, agent.ListSessionsHandler(history))
		frontendAPI.GET(agentSessionByIDPath, agent.ReadSessionHandler(history))
		frontendAPI.DELETE(agentSessionByIDPath, agent.DeleteSessionHandler(history))
	}
}

func agentFrontendAPI(r *gin.Engine, base string) gin.IRoutes {
	return r.Group(base).Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
}
