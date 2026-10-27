package picoclaw

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"NanoKVM-Server/service/agent"
	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/service/hid"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	protocol "github.com/modelcontextprotocol/go-sdk/mcp"
	log "github.com/sirupsen/logrus"
)

const (
	mcpServerName    = "nanokvm"
	mcpServerVersion = "1.0.0"
	// The JSON-RPC error code for a call that PicoClaw's control mode refuses.
	// Clients read the reason from the error data. It was -32003 until
	// go-sdk's client read that as its own "client is closing" code and
	// logged every refusal as a closed connection; go-sdk uses -32001 to
	// -32005 and -32042, so this stays clear of them.
	mcpControlConflictCode = -32010
	maxMCPRequestBytes     = 1 << 20
)

// mcpToolDefinitions are the tools the on-board agent is offered. The input
// schemas are plain maps: the handlers decode the arguments themselves, as
// they always have, and the SDK does not validate them.
var mcpToolDefinitions = []*protocol.Tool{
	{
		Name:        "kvm_screenshot",
		Description: "Capture the current HDMI frame from the downstream remote host as a JPEG image. Without arguments the image is scaled down (960 pixels wide unless the NanoKVM owner changed it), which is enough for most steps. To read small text or aim at a small target, pass both width and height, for example the full screen size reported with the previous screenshot.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"width":   map[string]any{"type": "integer", "description": "Target width in pixels (optional). Alone it can only make the default image smaller; with height the image has exactly this size"},
				"height":  map[string]any{"type": "integer", "description": "Target height in pixels (optional). Alone it can only make the default image smaller; with width the image has exactly this size"},
				"quality": map[string]any{"type": "integer", "description": "JPEG quality 1-100 (optional, default set by the NanoKVM owner, normally 60)"},
			},
		},
	},
	{
		Name: "kvm_actions",
		Description: "Send mouse and keyboard actions to the downstream remote host, in order. " +
			"Mouse positions are fractions of the screen from 0 to 1, not pixels: x=0 is the left edge, x=1 the right edge, y=0 the top, y=1 the bottom. " +
			"For a point seen in a screenshot, divide its pixel position by the image width and height (pixel 480,270 in a 960x540 image is x=0.5, y=0.5). " +
			"To move the pointer by a number of screen pixels from where it is, use move with dx/dy. " +
			`Examples: {"action":"click","x":0.5,"y":0.5} clicks the centre; {"action":"move","x":0.1,"y":0.9} goes near the bottom-left corner; ` +
			`{"action":"move","dx":10} moves 10 pixels right; {"action":"move","dy":-20} moves 20 pixels up; {"action":"click"} clicks where the pointer is; ` +
			`{"action":"type","text":"hello"} types text; {"action":"hotkey","keys":["shift"]} presses one key, here a modifier on its own; {"action":"hotkey","keys":["ctrl","c"]} presses keys together; {"action":"scroll","direction":"down","amount":3}; ` +
			`{"action":"drag","from":{"x":0.2,"y":0.2},"to":{"x":0.6,"y":0.2}}; {"action":"wait","duration_ms":1000}. ` +
			"type is only for text: to press any key, such as Enter, Escape, Tab, an arrow or Shift, use hotkey with that key. " +
			"type presses the keys that give the text on the keyboard layout the NanoKVM owner set for the host; if the host uses another, name it in layout. " +
			"The result reports where the pointer ended up. Set screenshot_after to get a screenshot of the result in the same call.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"actions": map[string]any{
					"type":        "array",
					"description": "The actions to run, in order. Each object needs an \"action\" field: click, move, type, hotkey, scroll, drag or wait.",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"action": map[string]any{
								"type": "string",
								"enum": []string{"click", "move", "type", "hotkey", "scroll", "drag", "wait"},
							},
							"x": map[string]any{
								"type": "number", "minimum": 0, "maximum": 1,
								"description": "click, move, scroll: horizontal position as a fraction of the screen width, 0 = left edge, 1 = right edge. Not pixels.",
							},
							"y": map[string]any{
								"type": "number", "minimum": 0, "maximum": 1,
								"description": "click, move, scroll: vertical position as a fraction of the screen height, 0 = top edge, 1 = bottom edge. Not pixels.",
							},
							"dx": map[string]any{
								"type":        "number",
								"description": "move only, instead of x/y: screen pixels to move right from the current pointer position (negative = left).",
							},
							"dy": map[string]any{
								"type":        "number",
								"description": "move only, instead of x/y: screen pixels to move down from the current pointer position (negative = up).",
							},
							"button": map[string]any{
								"type":        "string",
								"enum":        []string{"left", "right", "middle", "back", "forward"},
								"description": "click, drag: mouse button, default left.",
							},
							"text": map[string]any{"type": "string", "description": "type: the characters to type. Not for single keys: press Enter, Shift and other keys with hotkey."},
							"layout": map[string]any{
								"type":        "string",
								"description": "type: the keyboard layout active on the host, only when it differs from the one the NanoKVM owner set (US unless set): " + strings.Join(hid.LayoutIDs(), ", ") + ".",
							},
							"keys": map[string]any{
								"type":        "array",
								"items":       map[string]any{"type": "string"},
								"description": `hotkey: the keys to press together and then release, modifiers first. One key works too, including a modifier alone: ["shift"], ["enter"], ["ctrl","alt","delete"].`,
							},
							"direction": map[string]any{
								"type":        "string",
								"enum":        []string{"up", "down"},
								"description": "scroll: direction, default up.",
							},
							"amount": map[string]any{"type": "integer", "minimum": 1, "description": "scroll: number of wheel steps, default 1."},
							"duration_ms": map[string]any{
								"type":        "integer",
								"minimum":     0,
								"maximum":     maxWaitDurationMS,
								"description": "wait: milliseconds to wait, up to 30000.",
							},
							"from": map[string]any{
								"type":        "object",
								"description": "drag: start point, with x and y as fractions from 0 to 1.",
								"properties":  map[string]any{"x": map[string]any{"type": "number"}, "y": map[string]any{"type": "number"}},
							},
							"to": map[string]any{
								"type":        "object",
								"description": "drag: end point, with x and y as fractions from 0 to 1.",
								"properties":  map[string]any{"x": map[string]any{"type": "number"}, "y": map[string]any{"type": "number"}},
							},
						},
						"required": []string{"action"},
					},
				},
				"screenshot_after": map[string]any{
					"type":        "boolean",
					"description": "If true, take a screenshot after the actions and return it in this result, so no separate kvm_screenshot call is needed to check the outcome. Default false.",
				},
				"settle_ms": map[string]any{
					"type":        "integer",
					"minimum":     0,
					"maximum":     maxScreenshotSettle.Milliseconds(),
					"description": "With screenshot_after: milliseconds to wait after the last action before the screenshot, so windows and pages can draw. Default 500, at most 5000.",
				},
			},
			"required": []string{"actions"},
		},
	},
}

// MCPHandler serves the MCP endpoint that the on-board agent calls over
// loopback. The go-sdk streamable HTTP handler speaks the protocol; this
// package registers the tools and decides who may call them.
func (s *Service) MCPHandler(c *gin.Context) {
	// The tool handlers see only the headers of the request, so the peer
	// address travels in one. Set, not added: a client cannot choose it.
	c.Request.Header.Set(audit.ClientHeader, audit.ClientIP(c))
	relaxMCPRequestHeaders(c.Request)
	ctx := context.WithValue(c.Request.Context(), httpRequestContextKey{}, c.Request.Context())
	s.mcpHTTPHandler().ServeHTTP(c.Writer, c.Request.WithContext(ctx))
}

type httpRequestContextKey struct{}

// withHTTPRequestCancel ends a tool call when its HTTP request ends. go-sdk
// keeps the values of the request context but detaches the call from its
// cancellation, and an agent that gives up on a call must not leave a wait
// holding device control.
func withHTTPRequestCancel(ctx context.Context) (context.Context, func()) {
	requestCtx, ok := ctx.Value(httpRequestContextKey{}).(context.Context)
	if !ok {
		return ctx, func() {}
	}
	ctx, cancel := context.WithCancelCause(ctx)
	stop := context.AfterFunc(requestCtx, func() { cancel(context.Cause(requestCtx)) })
	return ctx, func() {
		stop()
		cancel(context.Canceled)
	}
}

// relaxMCPRequestHeaders keeps the endpoint open to a client that omits the
// headers the streamable HTTP transport asks for, as the hand-written handler
// that came before it was. Every answer is JSON, so the Accept header changes
// nothing that the client gets. A Content-Type that names something other
// than JSON is still refused.
func relaxMCPRequestHeaders(r *http.Request) {
	r.Header.Set("Accept", "application/json, text/event-stream")
	if r.Header.Get("Content-Type") == "" {
		r.Header.Set("Content-Type", "application/json")
	}
}

func (s *Service) mcpHTTPHandler() http.Handler {
	s.mcpOnce.Do(func() {
		server := s.newMCPServer()
		handler := protocol.NewStreamableHTTPHandler(
			func(*http.Request) *protocol.Server { return server },
			&protocol.StreamableHTTPOptions{Stateless: true, JSONResponse: true},
		)
		s.mcp = http.MaxBytesHandler(handler, maxMCPRequestBytes)
	})
	return s.mcp
}

func (s *Service) newMCPServer() *protocol.Server {
	server := protocol.NewServer(
		&protocol.Implementation{Name: mcpServerName, Version: mcpServerVersion},
		&protocol.ServerOptions{
			Capabilities: &protocol.ServerCapabilities{Tools: &protocol.ToolCapabilities{}},
			// No Mcp-Session-Id: every request stands alone, so a client has
			// no session to resume, lose or delete.
			GetSessionID: func() string { return "" },
		},
	)
	handlers := map[string]protocol.ToolHandler{
		"kvm_screenshot": s.mcpScreenshot,
		"kvm_actions":    s.mcpLeasedActions,
	}
	for _, tool := range mcpToolDefinitions {
		server.AddTool(tool, s.agentTool(tool.Name, handlers[tool.Name]))
	}
	for _, device := range s.deviceTools() {
		server.AddTool(device.tool, s.agentTool(device.tool.Name, device.handler))
	}
	// The agent is offered only the tools whose tier is on.
	server.AddReceivingMiddleware(onlyAllowedAgentTools)
	return server
}

// agentTool wraps a tool handler with its tier check and its audit record. A
// call whose tier is off never reaches the handler: it gets a tool error and
// is recorded as denied. The record names the agent and the chat session the
// call acts for.
func (s *Service) agentTool(name string, next protocol.ToolHandler) protocol.ToolHandler {
	tier := agentToolTiers[name]
	return func(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
		arguments := toolArguments(req)
		target, detail := describeAgentCall(name, arguments)
		event := audit.Event{
			Actor:    s.agentName(),
			Via:      "agent",
			Session:  s.actingSession(requestSession(req)),
			IP:       requestHeader(req, audit.ClientHeader),
			Category: audit.CategoryAgent,
			Action:   name,
			Tier:     string(tier),
			Target:   target,
			Detail:   detail,
		}
		if !audit.Allowed(audit.SubjectAgent, tier) {
			event.Outcome = audit.OutcomeDenied
			audit.Record(event)
			return mcpToolError(audit.DeniedMessage(audit.SubjectAgent, tier)), nil
		}

		callCtx, release := withHTTPRequestCancel(ctx)
		result, err := next(callCtx, req)
		release()
		switch {
		case err != nil:
			event.Outcome = audit.OutcomeFail
			event.Detail = joinAuditDetail(event.Detail, mcpErrorMessage(err))
		case result != nil && result.IsError:
			event.Outcome = audit.OutcomeFail
			event.Detail = joinAuditDetail(event.Detail, mcpResultText(result))
		default:
			event.Outcome = audit.OutcomeOK
		}
		audit.Record(event)
		return result, err
	}
}

// onlyAllowedAgentTools drops from a tools/list answer the tools whose tier
// is off. It also lists the tools in the order of agentToolOrder, which the
// SDK sorts by name, so that the agent's prompt keeps its order.
func onlyAllowedAgentTools(next protocol.MethodHandler) protocol.MethodHandler {
	return func(ctx context.Context, method string, req protocol.Request) (protocol.Result, error) {
		result, err := next(ctx, method, req)
		if err != nil || method != "tools/list" {
			return result, err
		}
		if list, ok := result.(*protocol.ListToolsResult); ok && list != nil {
			listed := make(map[string]*protocol.Tool, len(list.Tools))
			for _, tool := range list.Tools {
				listed[tool.Name] = tool
			}
			kept := make([]*protocol.Tool, 0, len(list.Tools))
			for _, name := range agentToolOrder {
				tool, ok := listed[name]
				if tier, known := agentToolTiers[name]; ok && known && audit.Allowed(audit.SubjectAgent, tier) {
					kept = append(kept, tool)
				}
			}
			list.Tools = kept
		}
		return result, err
	}
}

func toolArguments(req *protocol.CallToolRequest) json.RawMessage {
	if req == nil || req.Params == nil {
		return nil
	}
	return req.Params.Arguments
}

func requestHeader(req *protocol.CallToolRequest, name string) string {
	if req == nil || req.Extra == nil {
		return ""
	}
	return req.Extra.Header.Get(name)
}

func mcpErrorMessage(err error) string {
	var rpcErr *jsonrpc.Error
	if errors.As(err, &rpcErr) {
		return rpcErr.Message
	}
	return err.Error()
}

// mcpControlConflict is the JSON-RPC error for a call that PicoClaw's control
// mode refuses. The data names the reason and the current mode, so that the
// agent can tell the user why it cannot act.
func mcpControlConflict(err *PicoclawError, status controlStatusForMCP) error {
	message := "PicoClaw control mode is not active"
	reason := CodeControlRequired
	if err != nil && err.Message != "" {
		message = err.Message
		reason = err.Code
	}
	data, _ := json.Marshal(map[string]interface{}{
		"reason":        reason,
		"control_mode":  status.Mode,
		"transitioning": status.Transitioning,
	})
	return &jsonrpc.Error{Code: mcpControlConflictCode, Message: message, Data: data}
}

type controlStatusForMCP struct {
	Mode          string
	Transitioning bool
}

// mcpLeasedActions runs kvm_actions under a PicoClaw control mode lease and a
// control operation, which a switch of the control mode cancels.
func (s *Service) mcpLeasedActions(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	releaseMode, modeErr := s.acquireControlMode()
	if modeErr != nil {
		return nil, s.controlConflictForMCP(modeErr)
	}
	defer releaseMode()

	operationCtx, releaseOperation := s.beginControlOperation(ctx)
	defer releaseOperation()
	return s.mcpActions(operationCtx, toolArguments(req), requestSession(req)), nil
}

// controlConflictForMCP is the control conflict error for a refused lease,
// with the control mode as it now is.
func (s *Service) controlConflictForMCP(modeErr *PicoclawError) error {
	controlStatus := controlStatusForMCP{}
	if status, err := s.control.Status(); err == nil {
		controlStatus = controlStatusForMCP{
			Mode:          string(status.Mode),
			Transitioning: status.Transitioning,
		}
	}
	return mcpControlConflict(modeErr, controlStatus)
}

// mcpScreenshot captures the screen for the session the call acts for, as
// kvm_actions acts for it, and only while that session has a turn in
// progress: the screen may show what the user would not show an agent that
// nobody asked to look.
func (s *Service) mcpScreenshot(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	if turnErr := s.requireTurn(s.actingSession(requestSession(req))); turnErr != nil {
		return mcpToolError(turnErr.Message), nil
	}

	var params struct {
		Width   uint16 `json:"width"`
		Height  uint16 `json:"height"`
		Quality uint16 `json:"quality"`
	}
	if args := toolArguments(req); args != nil {
		_ = json.Unmarshal(args, &params)
	}

	query := ScreenshotQuery{
		Format:  "base64",
		Width:   params.Width,
		Height:  params.Height,
		Quality: params.Quality,
	}

	data, meta, err := s.captureScreenshot(ctx, query)
	if err != nil {
		return mcpToolError(err.Message), nil
	}

	s.publishMCPObservation(requestSession(req), "screenshot captured", data)

	return &protocol.CallToolResult{
		Meta: protocol.Meta{
			"source_width":   meta.SourceWidth,
			"source_height":  meta.SourceHeight,
			"capture_width":  meta.CaptureWidth,
			"capture_height": meta.CaptureHeight,
		},
		Content: []protocol.Content{
			&protocol.TextContent{Text: screenshotCaption(meta)},
			&protocol.ImageContent{Data: data, MIMEType: "image/jpeg"},
		},
	}, nil
}

func (s *Service) publishMCPObservation(sessionHeader string, text string, image []byte) {
	if len(image) == 0 {
		return
	}

	sessionID, err := s.activeGatewaySession(sessionHeader)
	if err != nil {
		return
	}

	event := agent.Event{
		Type:        agent.EventObservation,
		MessageID:   uuid.NewString(),
		Text:        text,
		ImageBase64: base64.StdEncoding.EncodeToString(image),
		MimeType:    "image/jpeg",
	}
	if publishErr := s.chat.Publish(sessionID, event); publishErr != nil {
		log.Warnf("failed to deliver picoclaw observation message: %v", publishErr)
	}
}

func (s *Service) mcpActions(ctx context.Context, args json.RawMessage, sessionHeader string) *protocol.CallToolResult {
	var params struct {
		Actions         []Action `json:"actions"`
		ScreenshotAfter bool     `json:"screenshot_after"`
		SettleMs        *int     `json:"settle_ms"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return mcpToolError("invalid actions payload")
	}
	if len(params.Actions) == 0 {
		return mcpToolError("empty actions array")
	}

	result, err := s.executeActions(ctx, s.actingSession(sessionHeader), params.Actions)
	if err != nil {
		return mcpToolError(describeActionError(err, params.Actions, args))
	}

	resultJSON, _ := json.Marshal(result)
	content := []protocol.Content{&protocol.TextContent{Text: string(resultJSON)}}
	if params.ScreenshotAfter {
		content = append(content, s.screenshotAfterActions(ctx, sessionHeader, settleDuration(params.SettleMs))...)
	}
	return &protocol.CallToolResult{Content: content}
}

// The pause between the last action and the screenshot that screenshot_after
// takes, so that a menu, a window or a page has time to draw.
const (
	defaultScreenshotSettle = 500 * time.Millisecond
	maxScreenshotSettle     = 5 * time.Second
)

func settleDuration(settleMs *int) time.Duration {
	if settleMs == nil {
		return defaultScreenshotSettle
	}
	settle := time.Duration(*settleMs) * time.Millisecond
	if settle < 0 {
		return 0
	}
	if settle > maxScreenshotSettle {
		return maxScreenshotSettle
	}
	return settle
}

// screenshotAfterActions waits for the screen to settle and returns the
// screenshot as tool result content. The actions have already run, so a
// failed capture is reported as text and does not turn the result into an
// error.
func (s *Service) screenshotAfterActions(ctx context.Context, sessionHeader string, settle time.Duration) []protocol.Content {
	if waitErr := waitForControlOperation(ctx, settle); waitErr != nil {
		return []protocol.Content{&protocol.TextContent{
			Text: "actions done; no screenshot taken: " + waitErr.Message,
		}}
	}

	data, meta, captureErr := s.captureScreenshot(ctx, ScreenshotQuery{Format: "base64"})
	if captureErr != nil {
		return []protocol.Content{&protocol.TextContent{
			Text: "actions done; screenshot failed: " + captureErr.Message,
		}}
	}

	s.publishMCPObservation(sessionHeader, "screenshot captured", data)
	return []protocol.Content{
		&protocol.TextContent{
			Text: screenshotCaption(meta) + fmt.Sprintf(", taken %d ms after the actions", settle.Milliseconds()),
		},
		&protocol.ImageContent{Data: data, MIMEType: "image/jpeg"},
	}
}

func mcpToolError(message string) *protocol.CallToolResult {
	return &protocol.CallToolResult{
		IsError: true,
		Content: []protocol.Content{&protocol.TextContent{Text: message}},
	}
}

// mcpResultText is the first text of a tool result.
func mcpResultText(result *protocol.CallToolResult) string {
	for _, content := range result.Content {
		if text, ok := content.(*protocol.TextContent); ok {
			return text.Text
		}
	}
	return ""
}

// screenshotCaption tells the model what it is looking at: the size of the
// image it got and the size of the remote screen, so that it can ask for a
// sharper image or reason about screen pixels.
func screenshotCaption(meta ScreenshotMeta) string {
	if meta.SourceWidth == 0 || meta.SourceHeight == 0 {
		return fmt.Sprintf("screenshot captured: %dx%d image", meta.CaptureWidth, meta.CaptureHeight)
	}
	return fmt.Sprintf(
		"screenshot captured: %dx%d image of a %dx%d screen",
		meta.CaptureWidth, meta.CaptureHeight, meta.SourceWidth, meta.SourceHeight,
	)
}
