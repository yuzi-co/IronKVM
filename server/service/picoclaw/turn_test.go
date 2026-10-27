package picoclaw

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"NanoKVM-Server/service/audit"

	"github.com/gin-gonic/gin"
)

// The agent acts on the host only while the user waits for it. After the
// user cancels a turn, or the agent has answered, the chat stays open, but a
// tool call the agent still has running is refused with a message the model
// can act on, and is recorded as a failure for its session, as a call with no
// session is.
func TestDeviceToolsAreRefusedOnceTheTurnHasEnded(t *testing.T) {
	service, host := newDeviceTestService(t)
	hid := service.hid.(*fakeHID)
	vision := service.vision.(*fakeVision)
	store := withAgentTiers(t, audit.TierObserve, audit.TierInput, audit.TierPower, audit.TierMedia)
	service.chat.(*turnChat).endTurn(deviceTestSession)

	for _, call := range []struct{ name, args string }{
		{"kvm_actions", `{"actions":[{"action":"move","x":0.5,"y":0.5}]}`},
		{"kvm_power", `{"button":"power"}`},
		{"kvm_media", `{"action":"eject","drive":"cdrom"}`},
		{"kvm_screenshot", `{}`},
	} {
		result := toolResult(t, callTool(t, service, call.name, call.args, nil))
		if !isErrorResult(result) || toolText(t, result) != noTurnMessage {
			t.Fatalf("%s: result = %v, want refused for the ended turn", call.name, result)
		}
	}
	if len(hid.mouse) != 0 || len(host.pressed) != 0 || len(host.ejected) != 0 || vision.calls != 0 {
		t.Fatalf("mouse %d, pressed %v, ejected %v, screenshots %d after the turn ended",
			len(hid.mouse), host.pressed, host.ejected, vision.calls)
	}
	events := agentEvents(t, store)
	if len(events) != 4 {
		t.Fatalf("events = %+v", events)
	}
	for _, e := range events {
		if e.Outcome != audit.OutcomeFail || e.Session != deviceTestSession || !strings.Contains(e.Detail, noTurnMessage) {
			t.Fatalf("event = %+v, want a failure for %s", e, deviceTestSession)
		}
	}
}

// A call that names a session gets the same answer when that session is not
// in a turn, though no other session holds the lock.
func TestANamedSessionNeedsATurnToo(t *testing.T) {
	service, _, hid := newMCPActionTestService(t)
	withAgentTiers(t, audit.TierObserve, audit.TierInput)
	service.chat.(*turnChat).endTurn("chat-9")

	result := toolResult(t, callTool(t, service, "kvm_actions", `{"actions":[{"action":"move","x":0.5,"y":0.5}]}`,
		map[string]string{agentSessionHeader: "chat-9"}))
	if !isErrorResult(result) || toolText(t, result) != noTurnMessage || len(hid.mouse) != 0 {
		t.Fatalf("result = %v, mouse %d, want refused", result, len(hid.mouse))
	}
}

// A cancel stops a batch that is still running: the step after it is
// refused and what the earlier steps ran is reported.
func TestABatchStopsWhenTheTurnEnds(t *testing.T) {
	service, _, hid := newMCPActionTestService(t)
	withAgentTiers(t, audit.TierObserve, audit.TierInput)
	if err := service.lock.Ensure("chat-1"); err != nil {
		t.Fatal(err)
	}
	hid.afterMouse = func() { service.chat.(*turnChat).endTurn("chat-1") }

	result := toolResult(t, callTool(t, service, "kvm_actions",
		`{"actions":[{"action":"move","x":0.1,"y":0.1},{"action":"move","x":0.9,"y":0.9}]}`, nil))
	text := toolText(t, result)
	if !isErrorResult(result) || !strings.Contains(text, "action 2 of 2") || !strings.Contains(text, noTurnMessage) {
		t.Fatalf("result = %v, want the second step refused", result)
	}
	if len(hid.mouse) != 1 {
		t.Fatalf("mouse writes = %d, want only the first step's", len(hid.mouse))
	}
}

// The loopback routes answer the same refusal, under its own code.
func TestLoopbackRoutesAreRefusedOnceTheTurnHasEnded(t *testing.T) {
	service, vision, hid := newMCPActionTestService(t)
	service.chat.(*turnChat).endTurn("chat-1")

	r := gin.New()
	r.POST("/actions", service.Actions)
	r.GET("/screenshot", service.Screenshot)
	for _, request := range []*http.Request{
		httptest.NewRequest(http.MethodPost, "/actions", strings.NewReader(`{"action":"move","x":0.5,"y":0.5}`)),
		httptest.NewRequest(http.MethodGet, "/screenshot", nil),
	} {
		request.Header.Set(agentSessionHeader, "chat-1")
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		r.ServeHTTP(response, request)

		var body struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Code != CodeTurnNotActive || body.Message != noTurnMessage {
			t.Fatalf("%s: %s, want refused for the ended turn", request.URL.Path, response.Body.String())
		}
	}
	if len(hid.mouse) != 0 || vision.calls != 0 {
		t.Fatalf("mouse %d, screenshots %d after the turn ended", len(hid.mouse), vision.calls)
	}
	if owner := service.lock.Owner(); owner != "" {
		t.Fatalf("a refused call left the lock with %q", owner)
	}
}
