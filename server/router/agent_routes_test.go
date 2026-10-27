package router

import (
	"slices"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// Every agent route the UI and the agent used under /api/picoclaw is served
// under /api/agent too, by the same handler, and the old path keeps working.
// Only PicoClaw's own settings stay under /api/picoclaw alone.
func TestAgentRoutesAreServedUnderBothBases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	server(r)

	handlers := make(map[string]string)
	for _, route := range r.Routes() {
		handlers[route.Method+" "+route.Path] = route.Handler
	}

	picoclawOnly := map[string]bool{
		"POST " + picoclawBasePath + picoclawModelConfigPath:  true,
		"POST " + picoclawBasePath + picoclawAgentProfilePath: true,
	}
	renamed := map[string]string{
		chatSocketPaths[picoclawBasePath]: chatSocketPaths[agentBasePath],
	}

	var legacy int
	for _, route := range r.Routes() {
		key := route.Method + " " + route.Path
		if !strings.HasPrefix(route.Path, picoclawBasePath+"/") || picoclawOnly[key] {
			continue
		}
		legacy++
		path := strings.TrimPrefix(route.Path, picoclawBasePath)
		if to, ok := renamed[path]; ok {
			path = to
		}
		neutral := route.Method + " " + agentBasePath + path
		if handlers[neutral] != route.Handler {
			t.Errorf("%s: %s is served by %q, want %q", key, neutral, handlers[neutral], route.Handler)
		}
	}
	if legacy < 15 {
		t.Fatalf("only %d routes under %s; the walk is not seeing the router", legacy, picoclawBasePath)
	}
	for key := range picoclawOnly {
		if _, ok := handlers[key]; !ok {
			t.Errorf("%s is gone", key)
		}
	}
}

func TestTheAgentsLoopbackPathsAreAllowedUnderBothBases(t *testing.T) {
	paths := LoopbackHTTPAllowedPaths()
	for _, path := range []string{"/api/agent/mcp", "/api/picoclaw/mcp", "/api/agent/actions", "/api/picoclaw/actions"} {
		if !slices.Contains(paths, path) {
			t.Errorf("%s is not allowed over loopback HTTP: %v", path, paths)
		}
	}
}
