package picoclaw

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	serverConfig "NanoKVM-Server/config"
	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/service/controlmode"

	"github.com/gin-gonic/gin"
	protocol "github.com/modelcontextprotocol/go-sdk/mcp"
)

// headerRoundTripper adds the headers that PicoClaw's MCP config names, the
// way PicoClaw's own headerTransport does.
type headerRoundTripper struct {
	headers map[string]string
}

func (h headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	for key, value := range h.headers {
		req.Header.Set(key, value)
	}
	return http.DefaultTransport.RoundTrip(req)
}

// connectLikePicoClaw connects to the route the way the PicoClaw gateway
// (yuzi-co/picoclaw v0.3.1-ironkvm.1, pkg/mcp/manager.go) does: the go-sdk
// streamable client with the standalone SSE stream off, which is what its
// "type": "http" server entry selects, and the URL on 127.0.0.1.
func connectLikePicoClaw(t *testing.T, service *Service) *protocol.ClientSession {
	t.Helper()
	router := gin.New()
	router.POST("/api/picoclaw/mcp", service.MCPHandler)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	client := protocol.NewClient(&protocol.Implementation{Name: "picoclaw", Version: "test"}, nil)
	transport := &protocol.StreamableClientTransport{
		Endpoint:             server.URL + "/api/picoclaw/mcp",
		DisableStandaloneSSE: true,
		HTTPClient:           &http.Client{Transport: headerRoundTripper{headers: map[string]string{serverConfig.PicoclawInternalTokenHeader: "test-token"}}},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func TestPicoClawClientNegotiatesItsProtocolVersion(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)
	session := connectLikePicoClaw(t, service)

	result := session.InitializeResult()
	if result.ProtocolVersion != "2025-11-25" {
		t.Fatalf("protocolVersion = %q, want the client's 2025-11-25", result.ProtocolVersion)
	}
	if result.ServerInfo.Name != "nanokvm" || result.ServerInfo.Version != "1.0.0" {
		t.Fatalf("serverInfo = %+v", result.ServerInfo)
	}
	if session.ID() != "" {
		t.Fatalf("session id = %q, want none: the route keeps no sessions", session.ID())
	}
}

func TestPicoClawClientListsTheToolsInOrder(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)
	withAgentTiers(t, audit.TierObserve, audit.TierInput)
	session := connectLikePicoClaw(t, service)

	var names []string
	for tool, err := range session.Tools(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, tool.Name)
	}
	if len(names) != 2 || names[0] != "kvm_screenshot" || names[1] != "kvm_actions" {
		t.Fatalf("tools = %v", names)
	}
}

func TestPicoClawClientGetsTheScreenshotAsAnImage(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)
	// The screen is looked at for the user chatting in a turn.
	if err := service.lock.Ensure("chat-1"); err != nil {
		t.Fatal(err)
	}
	session := connectLikePicoClaw(t, service)

	result, err := session.CallTool(context.Background(), &protocol.CallToolParams{Name: "kvm_screenshot"})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError || len(result.Content) != 2 {
		t.Fatalf("result = %+v", result)
	}
	caption, ok := result.Content[0].(*protocol.TextContent)
	if !ok || caption.Text != "screenshot captured: 960x540 image of a 1920x1080 screen" {
		t.Fatalf("caption = %#v", result.Content[0])
	}
	image, ok := result.Content[1].(*protocol.ImageContent)
	if !ok || image.MIMEType != "image/jpeg" || !bytes.Equal(image.Data, []byte{0xff, 0xd8, 0xff, 0xd9}) {
		t.Fatalf("image = %#v", result.Content[1])
	}
}

func TestPicoClawClientRunsActions(t *testing.T) {
	service, _, hid := newMCPActionTestService(t)
	if err := service.lock.Ensure("chat"); err != nil {
		t.Fatal(err)
	}
	session := connectLikePicoClaw(t, service)

	result, err := session.CallTool(context.Background(), &protocol.CallToolParams{
		Name:      "kvm_actions",
		Arguments: map[string]any{"actions": []any{map[string]any{"action": "click", "x": 0.5, "y": 0.5}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("result = %+v", result.Content)
	}
	if len(hid.mouse) == 0 {
		t.Fatal("the click sent no mouse reports")
	}
}

// TestPicoClawClientSeesTheControlConflict: the go-sdk client reports the
// conflict as the server's error, not as a closed connection, and the session
// stays usable. The wire answer is pinned by
// TestMCPControlConflictCarriesTheModeInData.
func TestPicoClawClientSeesTheControlConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	path := filepath.Join(t.TempDir(), "mode")
	if err := os.WriteFile(path, []byte(string(controlmode.ModeOff)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	service := &Service{control: controlmode.NewManager(path, controlmode.ModePicoclaw)}
	session := connectLikePicoClaw(t, service)

	_, err := session.CallTool(context.Background(), &protocol.CallToolParams{
		Name:      "kvm_actions",
		Arguments: map[string]any{"actions": []any{map[string]any{"action": "wait", "duration_ms": 1}}},
	})
	if err == nil || !strings.Contains(err.Error(), "PicoClaw device control is not enabled") {
		t.Fatalf("err = %v, want the control conflict", err)
	}
	if strings.Contains(err.Error(), "connection closed") {
		t.Fatalf("err = %v, the client read the conflict as a closed connection", err)
	}
	if _, err := session.CallTool(context.Background(), &protocol.CallToolParams{Name: "kvm_screenshot"}); err != nil {
		t.Fatalf("the next call failed: %v", err)
	}
}
