package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestWebRTCSettingsAreAdminOnly checks that the WebRTC network settings,
// the STUN server among them, sit behind the administrator role.
func TestWebRTCSettingsAreAdminOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)

	chains := make(map[string][]string)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		chains[c.Request.Method+" "+c.FullPath()] = c.HandlerNames()
		c.AbortWithStatus(http.StatusTeapot)
	})
	networkRouter(r)

	for _, key := range []string{
		"GET /api/network/webrtc",
		"POST /api/network/webrtc",
		"GET /api/network/webrtc/stun",
		"POST /api/network/webrtc/stun",
	} {
		method, path, _ := strings.Cut(key, " ")
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(method, path, nil))
		chain, ok := chains[key]
		if !ok {
			t.Errorf("%s: not a route", key)
			continue
		}
		if !chainHas(chain, "CheckToken") || !chainHas(chain, "RequireRole") {
			t.Errorf("%s: not behind the administrator role: %v", key, chain)
		}
	}
}
