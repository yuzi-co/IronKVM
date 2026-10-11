package router

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"NanoKVM-Server/common/apicontract"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/agent"
	"NanoKVM-Server/service/allowlist"
	"NanoKVM-Server/service/controlmode"
	"NanoKVM-Server/service/picoclaw"
	"NanoKVM-Server/service/vm"
	"NanoKVM-Server/service/webui"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Init(r *gin.Engine) {
	// Engine middleware applies only to the routes registered after it, so
	// these go first: every route, the static UI and the login included, is
	// behind both.
	r.Use(allowlist.Middleware())
	r.Use(middleware.LimitBody(middleware.DefaultBodyLimit, bodyLimits))

	web(r)
	server(r)
	log.Debugf("router init done")
}

func web(r *gin.Engine) {
	execPath, err := os.Executable()
	if err != nil {
		panic("invalid executable path")
	}

	execDir := filepath.Dir(execPath)
	webPath := fmt.Sprintf("%s/web", execDir)

	// The bundled UI in web/, unless a separately installed one in ui/ is
	// current and passes its check (service/webui).
	uiStore = webui.NewStore(filepath.Join(execDir, "ui"), webPath, apicontract.APIVersion, webui.DefaultPolicy())
	uiStore.Probe = probeWebDir
	r.Use(switchingStaticHandler(uiStore.Dir))
}

// bodyLimits are the routes allowed a larger body than
// middleware.DefaultBodyLimit. Each one is an upload whose handler bounds the
// body itself, so the cap here steps aside rather than setting a second number
// that would have to be kept in step with the first.
var bodyLimits = map[string]int64{
	// Capped at the largest package plus 1MB by the handler, which also
	// checks the free space against the declared length first.
	"POST /api/application/update/offline": middleware.NoBodyLimit,
	// Capped at the largest UI package and its signature plus 1MB.
	"POST /api/ui/upload": middleware.NoBodyLimit,
	// An ISO, often several GB. It streams to /data, not to memory, and the
	// write fails when the card is full.
	"POST /api/download/file": middleware.NoBodyLimit,
}

// apiPrefix covers the routes of the web UI's own API.
const apiPrefix = "/api/"

// redfishPrefix is the Redfish service's root. It is registered on the same
// engine, so it skips the static handler as the API does.
const redfishPrefix = "/redfish"

// skipsStatic reports whether a path belongs to a route handler rather than
// to the built web UI.
func skipsStatic(urlPath string) bool {
	return strings.HasPrefix(urlPath, apiPrefix) ||
		urlPath == redfishPrefix ||
		strings.HasPrefix(urlPath, redfishPrefix+"/")
}

// staticHandler serves the built web UI. The static middleware runs before
// routing and stats the filesystem for every request, so an API call or a
// websocket upgrade would pay a syscall against the SD card, and a file that
// happened to sit at the same path would shadow the real handler.
func staticHandler(webPath string) gin.HandlerFunc {
	serve := static.Serve("/", static.LocalFile(webPath, true))

	return func(c *gin.Context) {
		if skipsStatic(c.Request.URL.Path) {
			return
		}

		// Shared caches must not hand a compressed body to a client that
		// cannot read one, whichever branch answers below.
		c.Header("Vary", "Accept-Encoding")
		c.Header("Cache-Control", cacheControlFor(c.Request.URL.Path))

		if servePrecompressed(c, webPath) {
			return
		}

		serve(c)
	}
}

func server(r *gin.Engine) {
	control := controlmode.GetManager()
	picoclawService := picoclaw.NewService(control)
	// PicoClaw is the agent behind the chat routes. Another agent needs only
	// another agent.Agent here.
	chatAgent := picoclawService.Agent()
	chatBridge := agent.NewBridge(chatAgent, picoclawService, agent.NewTaskLeases(vm.AcquireHdmiCaptureLease))
	picoclawService.SetChat(chatBridge)

	versionRouter(r)
	authRouter(r)
	applicationRouter(r)
	uiRouter(r)
	vmRouter(r)
	streamRouter(r)
	roomMicRouter(r)
	storageRouter(r)
	networkRouter(r)
	allowlistRouter(r)
	hidRouter(r)
	sequenceRouter(r)
	controlRouter(r, control, picoclawService)
	mcpRouter(r, control, picoclawService)
	picoclawRouter(r, picoclawService)
	agentRouter(r, chatBridge, chatAgent)
	wsRouter(r)
	downloadRouter(r)
	extensionsRouter(r)
	debugRouter(r)
	metricsRouter(r)
	redfishRouter(r)
	ipmiRouter(r)
	watchdogRouter(r)
	notifyRouter(r)
	hostlogRouter(r)
	nutRouter(r)
	netbootRouter(r)
	vncRouter(r)
	ventoyRouter(r)
}

func LoopbackHTTPAllowedPaths() []string {
	paths := PicoclawLoopbackHTTPAllowedPaths()
	paths = append(paths, HIDLoopbackHTTPAllowedPaths()...)
	return paths
}
