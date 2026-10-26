package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestWHEPRoutes checks that the WHEP endpoint is registered, takes no
// session middleware (it answers to a Bearer token of its own), and that its
// setting sits behind the administrator role.
func TestWHEPRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	chains := make(map[string][]string)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		chains[c.Request.Method+" "+c.FullPath()] = c.HandlerNames()
		c.AbortWithStatus(http.StatusTeapot)
	})
	whepRouter(r)

	for _, key := range []string{
		"OPTIONS /api/whep",
		"POST /api/whep",
		"OPTIONS /api/whep/:id",
		"PATCH /api/whep/:id",
		"DELETE /api/whep/:id",
		"GET /api/stream/whep",
		"POST /api/stream/whep",
	} {
		method, path, _ := strings.Cut(key, " ")
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(method, concretePath(path), nil))
		chain, ok := chains[key]
		if !ok {
			t.Errorf("%s: not a route", key)
			continue
		}
		admin := strings.HasPrefix(path, "/api/stream/")
		if admin && (!chainHas(chain, "CheckToken") || !chainHas(chain, "RequireRole")) {
			t.Errorf("%s: not behind the administrator role: %v", key, chain)
		}
		if !admin && isGated(chain) {
			t.Errorf("%s: behind the session middleware, which would take the cookie: %v", key, chain)
		}
	}
}
