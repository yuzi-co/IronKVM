package middleware

import (
	"net/http"

	"NanoKVM-Server/authn"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// viewerRoutes is everything a view-only principal may reach: the video, what
// the page needs to draw it, and its own sign-in state. It is a list of what is
// allowed rather than of what is not, so a route added tomorrow is closed to a
// viewer until someone decides otherwise here.
//
// The check runs where a request is authenticated (allowByRole), so it covers
// every route behind CheckToken, CheckSession and
// CheckTokenOrLoopbackInternalToken, websockets included. A route that takes
// no credential from this package is not behind it; the router test lists
// those.
//
// Keys are "METHOD route-pattern", as gin registers them.
var viewerRoutes = map[string]struct{}{
	"GET /api/auth/account":  {},
	"GET /api/auth/password": {},
	"POST /api/auth/logout":  {},
	// A viewer account may change its own password. A share session has no
	// account, and the handler refuses it.
	"POST /api/auth/password": {},

	"GET /api/stream/mjpeg":       {},
	"GET /api/stream/h264":        {},
	"GET /api/stream/h264/direct": {},
	"GET /api/stream/info":        {},
	"GET /api/stream/screenshot":  {},

	"GET /api/vm/info":             {},
	"GET /api/vm/hardware":         {},
	"GET /api/vm/health":           {},
	"GET /api/vm/screen":           {},
	"GET /api/vm/input-resolution": {},
	"GET /api/vm/input-region":     {},
	"GET /api/vm/hostname":         {},
	"GET /api/vm/web-title":        {},
	"GET /api/vm/hdmi":             {},
	"GET /api/vm/failsafe":         {},

	"GET /api/room-mic": {},
}

// ViewerRoutes returns the allowed routes, for the router test.
func ViewerRoutes() []string {
	routes := make([]string, 0, len(viewerRoutes))
	for route := range viewerRoutes {
		routes = append(routes, route)
	}
	return routes
}

// ViewerMayReach reports whether a viewer may call a route.
func ViewerMayReach(method, route string) bool {
	_, ok := viewerRoutes[method+" "+route]
	return ok
}

// allowByRole ends the request with 403 when the principal is a viewer and the
// route is not one a viewer may reach. It reports whether the request goes on.
func allowByRole(c *gin.Context, principal Principal) bool {
	if principal.Role != authn.RoleViewer {
		return true
	}
	if ViewerMayReach(c.Request.Method, c.FullPath()) {
		return true
	}
	log.Debugf("view-only %s refused %s %s", principal.Username, c.Request.Method, c.FullPath())
	c.JSON(http.StatusForbidden, "forbidden")
	c.Abort()
	return false
}
