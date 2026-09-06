package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// registeredDebugRoutes collects the paths debugRouter adds, so a route that
// disappears is a test failure rather than a 404 found by hand on a device.
func registeredDebugRoutes(t *testing.T) map[string]bool {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	debugRouter(r)

	routes := make(map[string]bool)
	for _, route := range r.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	return routes
}

func TestEveryProfileIsReachable(t *testing.T) {
	routes := registeredDebugRoutes(t)

	want := []string{
		"GET /api/debug/pprof",
		"GET /api/debug/pprof/",
		"GET /api/debug/pprof/cmdline",
		"GET /api/debug/pprof/profile",
		"GET /api/debug/pprof/symbol",
		"POST /api/debug/pprof/symbol",
		"GET /api/debug/pprof/trace",
		"GET /api/debug/pprof/allocs",
		"GET /api/debug/pprof/block",
		"GET /api/debug/pprof/goroutine",
		"GET /api/debug/pprof/heap",
		"GET /api/debug/pprof/mutex",
		"GET /api/debug/pprof/threadcreate",
	}

	for _, route := range want {
		if !routes[route] {
			t.Errorf("route %q is not registered", route)
		}
	}
}

// The static middleware returns early on the /api/ prefix. A profile mounted
// anywhere else would be answered by the SD card instead.
func TestProfilesSitUnderTheApiPrefix(t *testing.T) {
	for route := range registeredDebugRoutes(t) {
		_, path, found := cutMethod(route)
		if !found {
			t.Fatalf("unexpected route format: %q", route)
		}
		if len(path) < len(apiPrefix) || path[:len(apiPrefix)] != apiPrefix {
			t.Errorf("route %q does not start with %q", path, apiPrefix)
		}
	}
}

func cutMethod(route string) (method string, path string, found bool) {
	for i := 0; i < len(route); i++ {
		if route[i] == ' ' {
			return route[:i], route[i+1:], true
		}
	}

	return "", "", false
}

func TestSecondsIsClamped(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{name: "an over-long request is cut to the cap", query: "seconds=3600", want: "30"},
		{name: "the cap itself is left alone", query: "seconds=30", want: "30"},
		{name: "a short request is left alone", query: "seconds=5", want: "5"},
		{name: "a fractional request is left alone", query: "seconds=0.5", want: "0.5"},
		{name: "an unparseable value is left for the handler to reject", query: "seconds=soon", want: "soon"},
		{name: "no value stays absent", query: "", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var seen string
			handler := clampSeconds(func(_ http.ResponseWriter, r *http.Request) {
				seen = r.URL.Query().Get("seconds")
			})

			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodGet, "/api/debug/pprof/profile?"+test.query, nil)

			handler(c)

			if seen != test.want {
				t.Errorf("seconds = %q, want %q", seen, test.want)
			}
		})
	}
}

// The other query parameters have to survive the rewrite, because the clamp
// re-encodes the whole query rather than editing one field in place.
func TestClampKeepsTheOtherParameters(t *testing.T) {
	var request *http.Request
	handler := clampSeconds(func(_ http.ResponseWriter, r *http.Request) {
		request = r
	})

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/debug/pprof/profile?seconds=600&debug=1", nil)

	handler(c)

	if got := request.URL.Query().Get("seconds"); got != "30" {
		t.Errorf("seconds = %q, want 30", got)
	}
	if got := request.URL.Query().Get("debug"); got != "1" {
		t.Errorf("debug = %q, want 1", got)
	}
}

// A profile is a read of the whole process. Losing the role check would put
// that behind any session at all, so the guard is asserted rather than
// assumed.
func TestProfilesAreNotReachableWithoutAToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	debugRouter(r)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/debug/pprof/heap", nil))

	if w.Code == http.StatusOK {
		t.Errorf("an unauthenticated request reached the heap profile: %d", w.Code)
	}
}
