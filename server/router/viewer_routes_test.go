package router

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"

	"github.com/gin-gonic/gin"
)

const probeHeader = "X-Route-Probe"

// ungatedRoutes are the routes that take no credential from the middleware
// package, and why. A route missing from here and from the gate is a route
// anybody can call, so adding one means adding a line, with its reason.
var ungatedRoutes = map[string]string{
	"POST /api/auth/login":          "the login",
	"POST /api/auth/login/totp":     "the login, second step",
	"POST /api/auth/share":          "opens a view-only session with a share link's token",
	"POST /api/network/wifi":        "only answers in access point mode",
	"POST /api/network/wifi/verify": "only answers in access point mode",
	"GET /api/version":              "the UI asks before it logs in",
}

// ungatedPaths are routes open under every method, with their reason.
var ungatedPaths = map[string]string{
	"/api/mcp": "answers to the MCP key an administrator issued (service/mcp), not to a login",
}

// isLoopbackRoute is a route that answers to the loopback token, which only
// processes on the board hold.
func isLoopbackRoute(chain []string) bool {
	return chainHas(chain, "CheckLoopbackInternalToken")
}

func isGated(chain []string) bool {
	return chainHas(chain, "CheckToken") ||
		chainHas(chain, "CheckSession") ||
		chainHas(chain, "CheckTokenOrLoopbackInternalToken")
}

func chainHas(chain []string, name string) bool {
	for _, handler := range chain {
		if strings.Contains(handler, "."+name+".func") {
			return true
		}
	}
	return false
}

// concretePath fills a route's parameters in so it can be requested.
func concretePath(route string) string {
	parts := strings.Split(route, "/")
	for index, part := range parts {
		switch {
		case strings.HasPrefix(part, ":"):
			parts[index] = "x"
		case strings.HasPrefix(part, "*"):
			parts[index] = "x"
		}
	}
	return strings.Join(parts, "/")
}

// TestViewOnlyIsRefusedEverywhereExceptTheVideo walks every route the server
// registers and requires of each one that a view-only session is refused,
// unless the route is on the list of what a viewer may reach. It does not take
// the router's word for it: the request is made, and it must come back 403.
//
// A request is only dispatched to a route whose handler chain starts with the
// authentication middleware, which is read from the chain itself first. A route
// without it would run its real handler in this test, so such a route fails the
// test before any request is made.
func TestViewOnlyIsRefusedEverywhereExceptTheVideo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	if _, _, err := store.Authenticate("admin", "admin"); err != nil {
		t.Fatal(err)
	}
	if err := store.Create("guest", "guest-password", authn.RoleViewer); err != nil {
		t.Fatal(err)
	}
	shares := authn.NewShareStore(filepath.Join(t.TempDir(), "share_links.json"))
	originalStore, originalShares := authn.DefaultStore, authn.DefaultShares
	authn.DefaultStore, authn.DefaultShares = store, shares
	conf := config.GetInstance()
	originalAuthentication, originalSecret, originalDuration :=
		conf.Authentication, conf.JWT.SecretKey, conf.JWT.RefreshTokenDuration
	conf.Authentication, conf.JWT.SecretKey, conf.JWT.RefreshTokenDuration = "enable", "test-secret", 3600
	t.Cleanup(func() {
		authn.DefaultStore, authn.DefaultShares = originalStore, originalShares
		conf.Authentication, conf.JWT.SecretKey, conf.JWT.RefreshTokenDuration =
			originalAuthentication, originalSecret, originalDuration
	})

	guest, err := store.Get("guest")
	if err != nil {
		t.Fatal(err)
	}
	viewerToken, err := middleware.GenerateJWT(guest.Username, guest.TokenVersion)
	if err != nil {
		t.Fatal(err)
	}
	link, _, err := shares.Create("", "admin", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	shareToken, err := middleware.GenerateShareJWT(link.ID, link.ExpiresAt)
	if err != nil {
		t.Fatal(err)
	}

	// The probe is the first middleware on the engine. It records the handler
	// chain of the route a request matched, then stops the request.
	chains := make(map[string][]string)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if c.GetHeader(probeHeader) == "" {
			return
		}
		chains[c.Request.Method+" "+c.FullPath()] = c.HandlerNames()
		c.AbortWithStatus(http.StatusTeapot)
	})
	server(r)

	routes := r.Routes()
	if len(routes) < 100 {
		t.Fatalf("only %d routes registered; the walk is not seeing the router", len(routes))
	}

	var dispatched, allowed int
	var problems []string
	for _, route := range routes {
		key := route.Method + " " + route.Path
		request := httptest.NewRequest(route.Method, concretePath(route.Path), nil)
		request.Header.Set(probeHeader, "1")
		recorder := httptest.NewRecorder()
		r.ServeHTTP(recorder, request)
		chain, ok := chains[key]
		if !ok {
			problems = append(problems, key+": the probe did not reach the route")
			continue
		}

		switch {
		case isLoopbackRoute(chain):
			continue
		case !isGated(chain):
			_, known := ungatedRoutes[key]
			_, knownPath := ungatedPaths[route.Path]
			if !known && !knownPath && !strings.HasPrefix(route.Path, "/redfish") {
				problems = append(problems, key+": no authentication in front of it, and not on the list of public routes")
			}
			continue
		}

		if middleware.ViewerMayReach(route.Method, route.Path) {
			allowed++
			// What a viewer may reach is read-only, apart from signing out and
			// changing its own password.
			mutating := route.Method != http.MethodGet
			if mutating && key != "POST /api/auth/logout" && key != "POST /api/auth/password" {
				problems = append(problems, key+": a viewer may call a route that changes state")
			}
			continue
		}

		for name, token := range map[string]string{"viewer account": viewerToken, "share link": shareToken} {
			request = httptest.NewRequest(route.Method, concretePath(route.Path), nil)
			request.AddCookie(&http.Cookie{Name: middleware.CookieName, Value: token})
			recorder = httptest.NewRecorder()
			r.ServeHTTP(recorder, request)
			dispatched++
			if recorder.Code != http.StatusForbidden {
				problems = append(problems, key+": a "+name+" got "+http.StatusText(recorder.Code))
			}
		}
	}

	for _, listed := range middleware.ViewerRoutes() {
		method, path, _ := strings.Cut(listed, " ")
		found := false
		for _, route := range routes {
			if route.Method == method && route.Path == path {
				found = true
			}
		}
		if !found {
			problems = append(problems, listed+": on the viewer list but not a route")
		}
	}

	sort.Strings(problems)
	for _, problem := range problems {
		t.Error(problem)
	}
	if allowed == 0 || dispatched == 0 {
		t.Fatalf("the walk checked nothing: %d allowed routes, %d refused requests", allowed, dispatched)
	}
	t.Logf("%d routes, %d viewer routes, %d refused requests", len(routes), allowed, dispatched)
}
