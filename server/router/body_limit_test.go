package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/webui"
)

// Every exception to the body cap names a route that exists. A typo, or a
// route that moved, would otherwise leave an upload refused at 1MB with
// nothing to say why.
func TestEveryBodyLimitExceptionIsARoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	original := uiStore
	uiStore = webui.NewStore(t.TempDir(), t.TempDir(), 1, webui.Policy{})
	t.Cleanup(func() { uiStore = original })

	applicationRouter(r)
	uiRouter(r)
	downloadRouter(r)
	supportRouter(r)

	routes := map[string]bool{}
	for _, route := range r.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for key := range bodyLimits {
		if !routes[key] {
			t.Errorf("bodyLimits names %q, which is not a registered route", key)
		}
	}
}

// The login needs no session, so it is the route anyone can reach: a body
// over the cap is refused before the handler parses it.
func TestTheLoginRefusesABodyOverTheCap(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.LimitBody(middleware.DefaultBodyLimit, bodyLimits))
	authRouter(r)

	body := `{"username":"admin","password":"` + strings.Repeat("a", middleware.DefaultBodyLimit) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("a login body over the cap answered %d %s", w.Code, w.Body.String())
	}
}
