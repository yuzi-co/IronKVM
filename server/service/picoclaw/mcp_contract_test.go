package picoclaw

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"NanoKVM-Server/service/agent"
	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/service/controlmode"

	"github.com/gin-gonic/gin"
)

// These tests pin what MCP clients of /api/picoclaw/mcp see on the wire: the
// PicoClaw gateway and any other agent that uses the route. They send the
// headers the go-sdk client in PicoClaw sends.

var updateGolden = flag.Bool("update", false, "rewrite the golden files in testdata")

const toolsListGolden = "testdata/mcp_tools_list.golden.json"

type rpcEnvelope struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	} `json:"error"`
}

func mcpRPC(t *testing.T, service *Service, body string, headers map[string]string) rpcEnvelope {
	t.Helper()
	router := gin.New()
	router.POST("/", service.MCPHandler)
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
	var envelope rpcEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("body %s: %v", response.Body.String(), err)
	}
	if envelope.JSONRPC != "2.0" || string(envelope.ID) != "7" {
		t.Fatalf("envelope = %s, want jsonrpc 2.0 and id 7", response.Body.String())
	}
	return envelope
}

func callTool(t *testing.T, service *Service, name string, arguments string, headers map[string]string) rpcEnvelope {
	t.Helper()
	body := `{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"` + name + `","arguments":` + arguments + `}}`
	return mcpRPC(t, service, body, headers)
}

func toolResult(t *testing.T, envelope rpcEnvelope) map[string]any {
	t.Helper()
	if envelope.Error != nil {
		t.Fatalf("JSON-RPC error %d: %s", envelope.Error.Code, envelope.Error.Message)
	}
	var result map[string]any
	if err := json.Unmarshal(envelope.Result, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func resultContent(t *testing.T, result map[string]any) []map[string]any {
	t.Helper()
	items, ok := result["content"].([]any)
	if !ok {
		t.Fatalf("result = %v, want a content array", result)
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, item.(map[string]any))
	}
	return out
}

func isErrorResult(result map[string]any) bool {
	isError, _ := result["isError"].(bool)
	return isError
}

func TestMCPInitializeAnswersWithTheServerIdentity(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)

	envelope := mcpRPC(t, service, `{"jsonrpc":"2.0","id":7,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`, nil)
	result := toolResult(t, envelope)
	if result["protocolVersion"] != "2024-11-05" {
		t.Fatalf("protocolVersion = %v", result["protocolVersion"])
	}
	if !reflect.DeepEqual(result["serverInfo"], map[string]any{"name": "nanokvm", "version": "1.0.0"}) {
		t.Fatalf("serverInfo = %v", result["serverInfo"])
	}
	if !reflect.DeepEqual(result["capabilities"], map[string]any{"tools": map[string]any{}}) {
		t.Fatalf("capabilities = %v, want only tools", result["capabilities"])
	}
}

func TestMCPPingAnswersAnEmptyResult(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)

	result := toolResult(t, mcpRPC(t, service, `{"jsonrpc":"2.0","id":7,"method":"ping"}`, nil))
	if len(result) != 0 {
		t.Fatalf("ping result = %v, want {}", result)
	}
}

// TestMCPUnknownMethodIsRefused: go-sdk answers a method it does not know
// with HTTP 400, not a JSON-RPC error. No client of the route sends one:
// PicoClaw calls initialize, tools/list and tools/call.
func TestMCPUnknownMethodIsRefused(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)

	response := performMCPRequest(service, `{"jsonrpc":"2.0","id":7,"method":"nanokvm/unknown"}`)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "nanokvm/unknown") {
		t.Fatalf("status = %d, body = %s, want 400 naming the method", response.Code, response.Body.String())
	}
}

// TestMCPToolsListMatchesGolden pins the tool names, descriptions and input
// schemas. Run with -update to rewrite the golden file after a deliberate
// change. Every tier is on and the board has every device, so the file holds
// every tool.
func TestMCPToolsListMatchesGolden(t *testing.T) {
	service, _ := newDeviceTestService(t)
	withAgentTiers(t, audit.TierObserve, audit.TierInput, audit.TierPower, audit.TierMedia)

	result := toolResult(t, mcpRPC(t, service, `{"jsonrpc":"2.0","id":7,"method":"tools/list"}`, nil))
	tools, err := json.MarshalIndent(result["tools"], "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(toolsListGolden), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(toolsListGolden, append(tools, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	golden, err := os.ReadFile(toolsListGolden)
	if err != nil {
		t.Fatal(err)
	}
	var want, got any
	if err := json.Unmarshal(golden, &want); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(tools, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tools/list differs from %s:\n%s", toolsListGolden, tools)
	}
}

func TestMCPUnknownToolIsInvalidParams(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)
	store := withAgentTiers(t, audit.TierObserve, audit.TierInput)

	envelope := callTool(t, service, "kvm_format_disk", `{}`, nil)
	if envelope.Error == nil || envelope.Error.Code != -32602 || !strings.Contains(envelope.Error.Message, "kvm_format_disk") {
		t.Fatalf("envelope = %+v, want -32602 naming the tool", envelope)
	}
	if events := agentEvents(t, store); len(events) != 0 {
		t.Fatalf("events = %+v, want none for an unknown tool", events)
	}
}

func TestMCPScreenshotResultShape(t *testing.T) {
	service, vision, _ := newMCPActionTestService(t)
	// The screen is looked at for the user chatting in a turn.
	if err := service.lock.Ensure("chat-1"); err != nil {
		t.Fatal(err)
	}
	store := withAgentTiers(t, audit.TierObserve, audit.TierInput)

	result := toolResult(t, callTool(t, service, "kvm_screenshot", `{}`, nil))
	if isErrorResult(result) {
		t.Fatalf("result = %v", result)
	}
	content := resultContent(t, result)
	if len(content) != 2 {
		t.Fatalf("content = %v, want caption and image", content)
	}
	if !reflect.DeepEqual(content[0], map[string]any{"type": "text", "text": "screenshot captured: 960x540 image of a 1920x1080 screen"}) {
		t.Fatalf("caption = %v", content[0])
	}
	wantImage := map[string]any{
		"type":     "image",
		"mimeType": "image/jpeg",
		"data":     base64.StdEncoding.EncodeToString([]byte{0xff, 0xd8, 0xff, 0xd9}),
	}
	if !reflect.DeepEqual(content[1], wantImage) {
		t.Fatalf("image = %v", content[1])
	}
	if vision.calls != 1 || vision.width != 960 || vision.height != 540 {
		t.Fatalf("vision calls = %d at %dx%d", vision.calls, vision.width, vision.height)
	}

	events := agentEvents(t, store)
	if len(events) != 1 || events[0].Outcome != audit.OutcomeOK || events[0].Tier != "observe" ||
		events[0].Actor != "picoclaw" || events[0].Via != "agent" || events[0].Action != "kvm_screenshot" {
		t.Fatalf("events = %+v", events)
	}
}

func TestMCPScreenshotPassesTheSizeArguments(t *testing.T) {
	service, vision, _ := newMCPActionTestService(t)
	// The screen is looked at for the user chatting in a turn.
	if err := service.lock.Ensure("chat-1"); err != nil {
		t.Fatal(err)
	}

	result := toolResult(t, callTool(t, service, "kvm_screenshot", `{"width":1280,"height":720,"quality":80}`, nil))
	if isErrorResult(result) {
		t.Fatalf("result = %v", result)
	}
	if vision.width != 1280 || vision.height != 720 {
		t.Fatalf("captured at %dx%d, want 1280x720", vision.width, vision.height)
	}
	if text := resultContent(t, result)[0]["text"]; text != "screenshot captured: 1280x720 image of a 1920x1080 screen" {
		t.Fatalf("caption = %v", text)
	}
}

func TestMCPScreenshotFailureIsAToolError(t *testing.T) {
	service, vision, _ := newMCPActionTestService(t)
	store := withAgentTiers(t, audit.TierObserve, audit.TierInput)
	vision.result = -1

	result := toolResult(t, callTool(t, service, "kvm_screenshot", `{}`, nil))
	if !isErrorResult(result) {
		t.Fatalf("result = %v, want isError", result)
	}
	content := resultContent(t, result)
	if len(content) != 1 || content[0]["type"] != "text" || content[0]["text"] == "" {
		t.Fatalf("content = %v, want one text item", content)
	}
	events := agentEvents(t, store)
	if len(events) != 1 || events[0].Outcome != audit.OutcomeFail || !strings.Contains(events[0].Detail, content[0]["text"].(string)) {
		t.Fatalf("events = %+v", events)
	}
}

func TestMCPScreenshotDeniedWithTheObserveTierOff(t *testing.T) {
	service, vision, _ := newMCPActionTestService(t)
	store := withAgentTiers(t, audit.TierInput)

	result := toolResult(t, callTool(t, service, "kvm_screenshot", `{}`, nil))
	content := resultContent(t, result)
	if !isErrorResult(result) || len(content) != 1 || content[0]["text"] != audit.DeniedMessage(audit.SubjectAgent, audit.TierObserve) {
		t.Fatalf("result = %v, want the denied message", result)
	}
	if vision.calls != 0 {
		t.Fatal("captured a frame with the tier off")
	}
	if events := agentEvents(t, store); len(events) != 1 || events[0].Outcome != audit.OutcomeDenied {
		t.Fatalf("events = %+v", events)
	}
}

func TestMCPActionsResultShape(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)

	result := toolResult(t, callTool(t, service, "kvm_actions", `{"actions":[{"action":"move","x":0.5,"y":0.5}]}`, map[string]string{sessionIDHeader: "s1"}))
	if isErrorResult(result) {
		t.Fatalf("result = %v", result)
	}
	content := resultContent(t, result)
	if len(content) != 1 || content[0]["type"] != "text" {
		t.Fatalf("content = %v", content)
	}
	var action map[string]any
	if err := json.Unmarshal([]byte(content[0]["text"].(string)), &action); err != nil {
		t.Fatalf("action text %q: %v", content[0]["text"], err)
	}
	for _, key := range []string{"action", "duration_ms", "hid_writes", "pointer"} {
		if _, ok := action[key]; !ok {
			t.Fatalf("action result %v lacks %q", action, key)
		}
	}
	if action["action"] != "move" {
		t.Fatalf("action = %v", action["action"])
	}
}

func TestMCPActionsArgumentErrorsAreToolErrors(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)

	for _, test := range []struct{ arguments, want string }{
		{`{"actions":"click"}`, "invalid actions payload"},
		{`{"actions":[]}`, "empty actions array"},
	} {
		result := toolResult(t, callTool(t, service, "kvm_actions", test.arguments, nil))
		content := resultContent(t, result)
		if !isErrorResult(result) || len(content) != 1 || content[0]["text"] != test.want {
			t.Fatalf("%s: result = %v, want %q", test.arguments, result, test.want)
		}
	}
}

// TestMCPActionsUseTheSessionHeaderOrTheLockOwner pins how a call names its
// session: the X-PicoClaw-Session-ID header, or the session that holds the AI
// input lock when the header is absent, as it is for PicoClaw's MCP client.
func TestMCPActionsUseTheSessionHeaderOrTheLockOwner(t *testing.T) {
	service, _, hid := newMCPActionTestService(t)
	if err := service.lock.Ensure("owner"); err != nil {
		t.Fatal(err)
	}
	move := `{"actions":[{"action":"move","x":0.5,"y":0.5}]}`

	result := toolResult(t, callTool(t, service, "kvm_actions", move, nil))
	if isErrorResult(result) {
		t.Fatalf("without a header: %v, want the lock owner's call to run", result)
	}

	writes := len(hid.mouse)
	result = toolResult(t, callTool(t, service, "kvm_actions", move, map[string]string{sessionIDHeader: "intruder"}))
	if !isErrorResult(result) {
		t.Fatalf("with another session: %v, want a lock error", result)
	}
	if len(hid.mouse) != writes {
		t.Fatal("another session's call reached the HID")
	}
}

func TestMCPControlConflictCarriesTheModeInData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	path := filepath.Join(t.TempDir(), "mode")
	if err := os.WriteFile(path, []byte(string(controlmode.ModeMCP)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	service := &Service{control: controlmode.NewManager(path, controlmode.ModePicoclaw)}

	envelope := callTool(t, service, "kvm_actions", `{"actions":[{"action":"move","x":0.5,"y":0.5}]}`, nil)
	if envelope.Error == nil || envelope.Error.Code != -32010 || envelope.Error.Message != "external MCP owns device control" {
		t.Fatalf("envelope = %+v", envelope)
	}
	var data map[string]any
	if err := json.Unmarshal(envelope.Error.Data, &data); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"reason": CodeControlOwnedByMCP, "control_mode": "mcp", "transitioning": false}
	if !reflect.DeepEqual(data, want) {
		t.Fatalf("data = %v, want %v", data, want)
	}
}

type recordingChat struct {
	noChat
	mu        sync.Mutex
	active    string
	published []agent.Event
}

func (c *recordingChat) IsActive(sessionID string) bool { return sessionID == c.active }
func (c *recordingChat) InTurn(sessionID string) bool   { return sessionID == c.active }
func (c *recordingChat) Turn(sessionID string) uint64 {
	if sessionID == c.active {
		return 1
	}
	return 0
}

func (c *recordingChat) Publish(sessionID string, event agent.Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if sessionID != c.active {
		return agent.ErrSessionNotActive
	}
	c.published = append(c.published, event)
	return nil
}

func TestMCPScreenshotIsPublishedToTheActiveChat(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)
	chat := &recordingChat{active: "s1"}
	service.SetChat(chat)

	result := toolResult(t, callTool(t, service, "kvm_screenshot", `{}`, map[string]string{sessionIDHeader: "s1"}))
	if isErrorResult(result) {
		t.Fatalf("result = %v", result)
	}
	chat.mu.Lock()
	defer chat.mu.Unlock()
	if len(chat.published) != 1 {
		t.Fatalf("published = %+v, want one observation", chat.published)
	}
	event := chat.published[0]
	if event.Type != agent.EventObservation || event.MimeType != "image/jpeg" || event.ImageBase64 != resultContent(t, result)[1]["data"] {
		t.Fatalf("observation = %+v", event)
	}
}

// TestMCPActionsStopWhenTheClientGoesAway pins that a call ends with its HTTP
// request: a client that gives up does not leave a wait holding device
// control.
func TestMCPActionsStopWhenTheClientGoesAway(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)
	router := gin.New()
	router.POST("/mcp", service.MCPHandler)
	server := httptest.NewServer(router)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	body := `{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"kvm_actions","arguments":{"actions":[{"action":"wait","duration_ms":30000}]}}}`
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/mcp", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	request.Header.Set(sessionIDHeader, "s1")
	if response, err := http.DefaultClient.Do(request); err == nil {
		answer, _ := io.ReadAll(response.Body)
		response.Body.Close()
		t.Fatalf("the 30 s wait answered before the client gave up: %s", answer)
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		service.operations.mu.Lock()
		active := len(service.operations.active)
		service.operations.mu.Unlock()
		if active == 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("%d control operations still active after the client left", active)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
