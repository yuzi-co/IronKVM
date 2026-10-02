package agent_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"NanoKVM-Server/service/agent"
	"NanoKVM-Server/service/agent/agenttest"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type closedCall struct {
	sessionID string
	code      int
	reason    string
	opened    bool
}

type fakeHost struct {
	mu        sync.Mutex
	admitErr  error
	lockLost  bool
	openedIDs []string
	closed    chan closedCall
}

func newFakeHost() *fakeHost {
	return &fakeHost{closed: make(chan closedCall, 8)}
}

func (h *fakeHost) Admit(string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.admitErr
}

func (h *fakeHost) Renew(string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return !h.lockLost
}

func (h *fakeHost) Opened(sessionID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.openedIDs = append(h.openedIDs, sessionID)
}

func (h *fakeHost) Closed(sessionID string, code int, reason string, opened bool) {
	h.closed <- closedCall{sessionID, code, reason, opened}
}

func (h *fakeHost) RelayConfig() agent.RelayConfig {
	return agent.RelayConfig{ReadTimeout: 10 * time.Second, WriteTimeout: time.Second, PingInterval: time.Minute, MaxMessageBytes: 1 << 20}
}

type leaseCounter struct {
	mu       sync.Mutex
	held     int
	acquired int
}

func (l *leaseCounter) acquire() func() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.held++
	l.acquired++
	return func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		l.held--
	}
}

func (l *leaseCounter) counts() (int, int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.held, l.acquired
}

type harness struct {
	agent  *agenttest.Agent
	host   *fakeHost
	leases *leaseCounter
	bridge *agent.Bridge
	server *httptest.Server
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := &harness{agent: agenttest.New(), host: newFakeHost(), leases: &leaseCounter{}}
	h.bridge = agent.NewBridge(h.agent, h.host, agent.NewTaskLeases(h.leases.acquire))
	router := gin.New()
	router.GET("/ws", h.bridge.Connect)
	h.server = httptest.NewServer(router)
	t.Cleanup(h.server.Close)
	return h
}

func (h *harness) dial(t *testing.T, sessionID string) (*websocket.Conn, *agenttest.Session) {
	t.Helper()
	url := "ws" + strings.TrimPrefix(h.server.URL, "http") + "/ws?session_id=" + sessionID
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	select {
	case session := <-h.agent.Opened:
		return conn, session
	case <-time.After(5 * time.Second):
		t.Fatal("the agent session did not open")
		return nil, nil
	}
}

func readEvent(t *testing.T, conn *websocket.Conn) map[string]any {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read event: %v", err)
	}
	var event map[string]any
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("decode event %s: %v", data, err)
	}
	return event
}

func readClose(t *testing.T, conn *websocket.Conn) *websocket.CloseError {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			var closeErr *websocket.CloseError
			if !errors.As(err, &closeErr) {
				t.Fatalf("socket ended without a close frame: %v", err)
			}
			return closeErr
		}
	}
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestBridgeRelaysAPromptTurnThroughTheInterface(t *testing.T) {
	h := newHarness(t)
	conn, session := h.dial(t, "s1")

	if err := conn.WriteJSON(map[string]any{
		"type": "session/prompt", "requestId": "r1", "text": "hello", "maxSteps": 4, "maxRuntimeMs": 60000,
	}); err != nil {
		t.Fatal(err)
	}
	prompt := <-session.Prompts
	if prompt.RequestID != "r1" || prompt.Text != "hello" || prompt.MaxSteps == nil || *prompt.MaxSteps != 4 ||
		prompt.MaxRuntimeMs == nil || *prompt.MaxRuntimeMs != 60000 {
		t.Fatalf("the agent got %+v", prompt)
	}
	if held, _ := h.leases.counts(); held != 1 {
		t.Fatalf("a prompt holds %d capture leases, want 1", held)
	}

	session.Emit(agent.Event{Type: agent.EventTurnStarted})
	if event := readEvent(t, conn); event["type"] != "turn_started" {
		t.Fatalf("got %v", event)
	}
	session.Emit(agent.Event{Type: agent.EventAgentMessage, MessageID: "m1", Kind: agent.KindHidden, Text: "thinking"})
	session.Emit(agent.Event{Type: agent.EventAgentMessage, MessageID: "m2", Kind: agent.KindPlaceholder, Text: "..."})
	for range 2 {
		readEvent(t, conn)
	}
	if held, _ := h.leases.counts(); held != 1 {
		t.Fatal("a hidden or placeholder message ended the capture lease")
	}

	llmCalls := 2
	session.Emit(agent.Event{
		Type: agent.EventTurnDone, RequestIDs: []string{"r1"}, StopReason: agent.StopEndTurn,
		Usage: &agent.Usage{InputTokens: 10, OutputTokens: 2, TotalTokens: 12, LLMCalls: &llmCalls},
	})
	event := readEvent(t, conn)
	if event["type"] != "turn_done" || event["stopReason"] != "end_turn" {
		t.Fatalf("got %v", event)
	}
	usage, _ := event["usage"].(map[string]any)
	if usage["totalTokens"] != float64(12) || usage["llmCalls"] != float64(2) {
		t.Fatalf("usage %v", usage)
	}
	waitFor(t, "turn_done to end the lease", func() bool { held, _ := h.leases.counts(); return held == 0 })

	if err := conn.WriteJSON(map[string]any{"type": "session/cancel"}); err != nil {
		t.Fatal(err)
	}
	if requestID := <-session.Cancels; requestID != "" {
		t.Fatalf("cancel named %q", requestID)
	}
}

func TestBridgeClosesWithUpstreamClosedWhenTheAgentGoesAway(t *testing.T) {
	h := newHarness(t)
	conn, session := h.dial(t, "s1")

	session.End(errors.New("connection reset"))
	if closeErr := readClose(t, conn); closeErr.Code != agent.CloseUpstreamClosed {
		t.Fatalf("close code %d, want %d", closeErr.Code, agent.CloseUpstreamClosed)
	}
	call := <-h.host.closed
	if call.sessionID != "s1" || call.code != agent.CloseUpstreamClosed || !call.opened {
		t.Fatalf("host.Closed got %+v", call)
	}
	select {
	case <-session.Closed():
	case <-time.After(5 * time.Second):
		t.Fatal("the agent session was not closed")
	}
}

func TestBridgePassesTheAgentsCloseCode(t *testing.T) {
	h := newHarness(t)
	conn, session := h.dial(t, "s1")
	session.End(&agent.CloseError{Code: agent.CloseAuthFailed, Reason: "bad token"})
	closeErr := readClose(t, conn)
	if closeErr.Code != agent.CloseAuthFailed || closeErr.Text != "bad token" {
		t.Fatalf("close %d %q", closeErr.Code, closeErr.Text)
	}
}

func TestBridgePublishesServerEvents(t *testing.T) {
	h := newHarness(t)
	conn, _ := h.dial(t, "s1")

	if err := h.bridge.Publish("s1", agent.Event{Type: agent.EventObservation, MessageID: "o1", ImageBase64: "abc", MimeType: "image/jpeg"}); err != nil {
		t.Fatal(err)
	}
	if event := readEvent(t, conn); event["type"] != "observation" || event["imageBase64"] != "abc" {
		t.Fatalf("got %v", event)
	}
	h.bridge.Broadcast(agent.Event{Type: agent.EventControlModeChanged, Control: &agent.ControlMode{Mode: "mcp"}})
	event := readEvent(t, conn)
	control, _ := event["control"].(map[string]any)
	if event["type"] != "control_mode_changed" || control["mode"] != "mcp" || control["canControl"] != false {
		t.Fatalf("got %v", event)
	}
	if err := h.bridge.Publish("other", agent.Event{Type: agent.EventObservation}); !errors.Is(err, agent.ErrSessionNotActive) {
		t.Fatalf("publish to an unknown session = %v", err)
	}
}

func TestBridgeServerPromptReachesTheAgentWithoutALease(t *testing.T) {
	h := newHarness(t)
	_, session := h.dial(t, "s1")
	if err := h.bridge.Prompt("s1", agent.Prompt{RequestID: "x", Text: "load"}); err != nil {
		t.Fatal(err)
	}
	if prompt := <-session.Prompts; prompt.Text != "load" {
		t.Fatalf("got %+v", prompt)
	}
	if _, acquired := h.leases.counts(); acquired != 0 {
		t.Fatal("a server prompt took a capture lease")
	}
}

func TestBridgeCloseAllEndsEverySession(t *testing.T) {
	h := newHarness(t)
	conn, _ := h.dial(t, "s1")
	waitFor(t, "the session to be active", func() bool { return h.bridge.IsActive("s1") })
	if n := h.bridge.CloseAll(agent.CloseRuntimeStopped, "stopped"); n != 1 {
		t.Fatalf("closed %d sessions, want 1", n)
	}
	if closeErr := readClose(t, conn); closeErr.Code != agent.CloseRuntimeStopped {
		t.Fatalf("close code %d", closeErr.Code)
	}
	if h.bridge.SessionCount() != 0 {
		t.Fatal("a closed session is still counted")
	}
}

func TestBridgeClosesWhenTheLockIsLost(t *testing.T) {
	h := newHarness(t)
	conn, _ := h.dial(t, "s1")
	h.host.mu.Lock()
	h.host.lockLost = true
	h.host.mu.Unlock()
	_ = conn.WriteJSON(map[string]any{"type": "session/prompt", "requestId": "r1", "text": "x"})
	if closeErr := readClose(t, conn); closeErr.Code != agent.CloseTakenOver {
		t.Fatalf("close code %d, want %d", closeErr.Code, agent.CloseTakenOver)
	}
}

func TestBridgeRefusesASecondConnectionForTheSameSession(t *testing.T) {
	h := newHarness(t)
	h.dial(t, "s1")
	waitFor(t, "the session to be active", func() bool { return h.bridge.IsActive("s1") })
	response, err := http.Get(h.server.URL + "/ws?session_id=s1")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var body map[string]any
	_ = json.NewDecoder(response.Body).Decode(&body)
	if response.StatusCode != http.StatusConflict || body["code"] != agent.CodeLockHeld {
		t.Fatalf("status %d body %v", response.StatusCode, body)
	}
}

func TestBridgeAnswersTheHostsRefusal(t *testing.T) {
	h := newHarness(t)
	h.host.admitErr = &agent.Error{Status: http.StatusConflict, Code: "CONTROL_OWNED_BY_MCP", Message: "external MCP owns device control"}
	response, err := http.Get(h.server.URL + "/ws?session_id=s1")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var body map[string]any
	_ = json.NewDecoder(response.Body).Decode(&body)
	if response.StatusCode != http.StatusConflict || body["code"] != "CONTROL_OWNED_BY_MCP" {
		t.Fatalf("status %d body %v", response.StatusCode, body)
	}
	if h.bridge.SessionCount() != 0 {
		t.Fatal("a refused session is still counted")
	}
}

func TestBridgeReportsAnAgentThatCannotOpen(t *testing.T) {
	h := newHarness(t)
	h.agent.OpenErr = agent.NewError(agent.CodeRuntimeUnavailable, "gateway is unavailable")
	url := "ws" + strings.TrimPrefix(h.server.URL, "http") + "/ws?session_id=s1"
	if _, _, err := websocket.DefaultDialer.Dial(url, nil); err == nil {
		t.Fatal("the socket opened although the agent did not")
	}
	call := <-h.host.closed
	if call.code != agent.CloseRuntimeUnavailable || call.reason != "gateway is unavailable" || call.opened {
		t.Fatalf("host.Closed got %+v", call)
	}
}
