package agent_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"NanoKVM-Server/service/agent"
	"NanoKVM-Server/service/agent/agenttest"

	"github.com/gin-gonic/gin"
)

func historyRouter(history agent.History) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/sessions", agent.ListSessionsHandler(history))
	router.GET("/sessions/:id", agent.ReadSessionHandler(history))
	router.DELETE("/sessions/:id", agent.DeleteSessionHandler(history))
	return router
}

func call(t *testing.T, router *gin.Engine, method, path string) (int, map[string]any) {
	t.Helper()
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(method, path, nil))
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return recorder.Code, body
}

func TestHistoryRoutesPageReadAndDelete(t *testing.T) {
	fake := agenttest.New()
	for i := range 3 {
		fake.Store(agent.SessionDetail{
			ID:       fmt.Sprintf("s%d", i),
			Summary:  fmt.Sprintf("chat %d", i),
			Messages: []agent.SessionMessage{{Role: "user", Content: "hi"}},
			Updated:  fmt.Sprintf("2026-10-0%dT00:00:00Z", i+1),
		})
	}
	router := historyRouter(fake)

	_, body := call(t, router, http.MethodGet, "/sessions?offset=1&limit=1")
	items, _ := body["data"].([]any)
	if body["code"] != float64(0) || len(items) != 1 || items[0].(map[string]any)["id"] != "s1" {
		t.Fatalf("page = %v", body)
	}
	if item := items[0].(map[string]any); item["message_count"] != float64(1) {
		t.Fatalf("item = %v", item)
	}
	_, body = call(t, router, http.MethodGet, "/sessions?offset=9")
	if items, ok := body["data"].([]any); !ok || len(items) != 0 {
		t.Fatalf("a page past the end = %v, want an empty list", body)
	}

	_, body = call(t, router, http.MethodGet, "/sessions/s2")
	detail, _ := body["data"].(map[string]any)
	if detail["id"] != "s2" || len(detail["messages"].([]any)) != 1 {
		t.Fatalf("detail = %v", body)
	}

	_, body = call(t, router, http.MethodDelete, "/sessions/s2")
	if data, _ := body["data"].(map[string]any); data["deleted"] != true {
		t.Fatalf("delete = %v", body)
	}
	status, body := call(t, router, http.MethodGet, "/sessions/s2")
	if status != http.StatusNotFound || body["code"] != agent.CodeRuntimeUnavailable || body["message"] != "session not found" {
		t.Fatalf("a deleted session = %d %v", status, body)
	}
}
