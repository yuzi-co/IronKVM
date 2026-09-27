package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/config"
)

func metricsEngine(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	metricsRouter(r)

	return r
}

// Under /api/ so the static middleware returns early and no file on the SD
// card can shadow it.
func TestMetricsIsRegisteredUnderTheApiPrefix(t *testing.T) {
	for _, route := range metricsEngine(t).Routes() {
		if route.Method == http.MethodGet && route.Path == "/api/metrics" {
			if !strings.HasPrefix(route.Path, apiPrefix) {
				t.Fatalf("route %q does not start with %q", route.Path, apiPrefix)
			}
			return
		}
	}

	t.Fatal("GET /api/metrics is not registered")
}

// No session and no API key is refused. Prometheus sends its key as
// Authorization: Bearer, which CheckToken accepts; middleware/apikey_test.go
// covers that path.
func TestMetricsNeedsAToken(t *testing.T) {
	conf := config.GetInstance()
	original := conf.Authentication
	conf.Authentication = "enable"
	t.Cleanup(func() { conf.Authentication = original })

	w := httptest.NewRecorder()
	metricsEngine(t).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/metrics", nil))

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}
