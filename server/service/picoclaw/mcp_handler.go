package picoclaw

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// JSON-RPC 2.0 types for MCP HTTP transport

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCP tool definitions

var mcpToolDefinitions = []map[string]interface{}{
	{
		"name":        "kvm_screenshot",
		"description": "Capture the current HDMI frame from the downstream remote host as a JPEG image. Without arguments the image is scaled down (960 pixels wide unless the NanoKVM owner changed it), which is enough for most steps. To read small text or aim at a small target, pass both width and height, for example the full screen size reported with the previous screenshot.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"width":   map[string]interface{}{"type": "integer", "description": "Target width in pixels (optional). Alone it can only make the default image smaller; with height the image has exactly this size"},
				"height":  map[string]interface{}{"type": "integer", "description": "Target height in pixels (optional). Alone it can only make the default image smaller; with width the image has exactly this size"},
				"quality": map[string]interface{}{"type": "integer", "description": "JPEG quality 1-100 (optional, default set by the NanoKVM owner, normally 60)"},
			},
		},
	},
	{
		"name":        "kvm_actions",
		"description": "Send one or more HID actions (click, type, hotkey, scroll, drag, move, wait) to the downstream remote host. Use normalized [0,1] coordinates for mouse actions. Set screenshot_after to get a screenshot of the result in the same call.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"actions": map[string]interface{}{
					"type":        "array",
					"description": "Array of action objects. Each requires an 'action' field (click, move, type, hotkey, scroll, drag, wait).",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"action":    map[string]interface{}{"type": "string"},
							"x":         map[string]interface{}{"type": "number"},
							"y":         map[string]interface{}{"type": "number"},
							"button":    map[string]interface{}{"type": "string"},
							"text":      map[string]interface{}{"type": "string"},
							"keys":      map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
							"direction": map[string]interface{}{"type": "string"},
							"amount":    map[string]interface{}{"type": "integer"},
							"duration_ms": map[string]interface{}{
								"type":        "integer",
								"minimum":     0,
								"maximum":     maxWaitDurationMS,
								"description": "Wait duration in milliseconds, up to 30000",
							},
							"from": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"x": map[string]interface{}{"type": "number"}, "y": map[string]interface{}{"type": "number"}}},
							"to":   map[string]interface{}{"type": "object", "properties": map[string]interface{}{"x": map[string]interface{}{"type": "number"}, "y": map[string]interface{}{"type": "number"}}},
						},
						"required": []string{"action"},
					},
				},
				"screenshot_after": map[string]interface{}{
					"type":        "boolean",
					"description": "If true, take a screenshot after the actions and return it in this result, so no separate kvm_screenshot call is needed to check the outcome. Default false.",
				},
				"settle_ms": map[string]interface{}{
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

// MCPHandler handles MCP HTTP transport (JSON-RPC 2.0 over POST).
func (s *Service) MCPHandler(c *gin.Context) {
	var req jsonRPCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonRPCError{Code: -32700, Message: "parse error"},
		})
		return
	}

	if req.JSONRPC != "2.0" {
		c.JSON(http.StatusOK, jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonRPCError{Code: -32600, Message: "invalid request: jsonrpc must be 2.0"},
		})
		return
	}

	var resp jsonRPCResponse
	switch req.Method {
	case "initialize":
		resp = s.mcpInitialize(req)
	case "tools/list":
		resp = s.mcpToolsList(req)
	case "tools/call":
		resp = s.mcpToolsCall(req, c)
	case "ping":
		resp = jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{}}
	default:
		resp = jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonRPCError{Code: -32601, Message: fmt.Sprintf("method not found: %s", req.Method)},
		}
	}

	c.JSON(http.StatusOK, resp)
}

func mcpModeConflictResponse(req jsonRPCRequest, err *PicoclawError, status controlStatusForMCP) jsonRPCResponse {
	message := "PicoClaw control mode is not active"
	reason := CodeControlRequired
	if err != nil && err.Message != "" {
		message = err.Message
		reason = err.Code
	}
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Error: &jsonRPCError{
			Code:    -32003,
			Message: message,
			Data: map[string]interface{}{
				"reason":        reason,
				"control_mode":  status.Mode,
				"transitioning": status.Transitioning,
			},
		},
	}
}

type controlStatusForMCP struct {
	Mode          string
	Transitioning bool
}

func (s *Service) mcpInitialize(req jsonRPCRequest) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "nanokvm",
				"version": "1.0.0",
			},
		},
	}
}

func (s *Service) mcpToolsList(req jsonRPCRequest) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": mcpToolDefinitions,
		},
	}
}

func (s *Service) mcpToolsCall(req jsonRPCRequest, c *gin.Context) jsonRPCResponse {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonRPCError{Code: -32602, Message: "invalid params"},
		}
	}

	switch params.Name {
	case "kvm_screenshot":
		return s.mcpScreenshot(req, params.Arguments, c)
	case "kvm_actions":
		releaseMode, modeErr := s.acquireControlMode()
		if modeErr != nil {
			controlStatus := controlStatusForMCP{}
			if status, err := s.control.Status(); err == nil {
				controlStatus = controlStatusForMCP{
					Mode:          string(status.Mode),
					Transitioning: status.Transitioning,
				}
			}
			return mcpModeConflictResponse(req, modeErr, controlStatus)
		}
		defer releaseMode()

		operationCtx, releaseOperation := s.beginControlOperation(c.Request.Context())
		defer releaseOperation()
		c.Request = c.Request.WithContext(operationCtx)
		return s.mcpActions(req, params.Arguments, c)
	default:
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonRPCError{Code: -32602, Message: fmt.Sprintf("unknown tool: %s", params.Name)},
		}
	}
}

func (s *Service) mcpScreenshot(req jsonRPCRequest, args json.RawMessage, c *gin.Context) jsonRPCResponse {
	var params struct {
		Width   uint16 `json:"width"`
		Height  uint16 `json:"height"`
		Quality uint16 `json:"quality"`
	}
	if args != nil {
		_ = json.Unmarshal(args, &params)
	}

	query := ScreenshotQuery{
		Format:  "base64",
		Width:   params.Width,
		Height:  params.Height,
		Quality: params.Quality,
	}

	data, meta, err := s.captureScreenshot(c.Request.Context(), query)
	if err != nil {
		return mcpToolError(req, err.Message)
	}

	b64 := base64.StdEncoding.EncodeToString(data)
	s.publishMCPObservation(c, "screenshot captured", b64)

	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": screenshotCaption(meta),
				},
				{
					"type":     "image",
					"data":     b64,
					"mimeType": "image/jpeg",
				},
			},
			"meta": map[string]interface{}{
				"source_width":   meta.SourceWidth,
				"source_height":  meta.SourceHeight,
				"capture_width":  meta.CaptureWidth,
				"capture_height": meta.CaptureHeight,
			},
		},
	}
}

func (s *Service) publishMCPObservation(c *gin.Context, text string, imageBase64 string) {
	if strings.TrimSpace(imageBase64) == "" {
		return
	}

	sessionID, session, err := s.requireActiveGatewaySession(c)
	if err != nil || session == nil || session.Downstream == nil {
		return
	}

	message := newPicoGatewayObservationMessage(sessionID, map[string]any{
		"content":      text,
		"image_base64": imageBase64,
		"mime_type":    "image/jpeg",
	})
	if writeErr := session.writeDownstreamJSON(s.config.Get(), message); writeErr != nil {
		log.Warnf("failed to deliver picoclaw observation message: %v", writeErr)
	}
}

func (s *Service) mcpActions(req jsonRPCRequest, args json.RawMessage, c *gin.Context) jsonRPCResponse {
	var params struct {
		Actions         []Action `json:"actions"`
		ScreenshotAfter bool     `json:"screenshot_after"`
		SettleMs        *int     `json:"settle_ms"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return mcpToolError(req, "invalid actions payload")
	}
	if len(params.Actions) == 0 {
		return mcpToolError(req, "empty actions array")
	}

	sessionID := c.GetHeader(sessionIDHeader)
	if sessionID == "" {
		sessionID = s.lock.Owner()
	}

	ctx := c.Request.Context()
	result, err := s.executeActions(ctx, sessionID, params.Actions)
	if err != nil {
		return mcpToolError(req, err.Message)
	}

	resultJSON, _ := json.Marshal(result)
	content := []map[string]interface{}{
		{
			"type": "text",
			"text": string(resultJSON),
		},
	}
	if params.ScreenshotAfter {
		content = append(content, s.screenshotAfterActions(ctx, c, settleDuration(params.SettleMs))...)
	}

	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": content,
		},
	}
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
func (s *Service) screenshotAfterActions(ctx context.Context, c *gin.Context, settle time.Duration) []map[string]interface{} {
	if waitErr := waitForControlOperation(ctx, settle); waitErr != nil {
		return []map[string]interface{}{{
			"type": "text",
			"text": "actions done; no screenshot taken: " + waitErr.Message,
		}}
	}

	data, meta, captureErr := s.captureScreenshot(ctx, ScreenshotQuery{Format: "base64"})
	if captureErr != nil {
		return []map[string]interface{}{{
			"type": "text",
			"text": "actions done; screenshot failed: " + captureErr.Message,
		}}
	}

	b64 := base64.StdEncoding.EncodeToString(data)
	s.publishMCPObservation(c, "screenshot captured", b64)
	return []map[string]interface{}{
		{
			"type": "text",
			"text": screenshotCaption(meta) + fmt.Sprintf(", taken %d ms after the actions", settle.Milliseconds()),
		},
		{
			"type":     "image",
			"data":     b64,
			"mimeType": "image/jpeg",
		},
	}
}

func mcpToolError(req jsonRPCRequest, message string) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"isError": true,
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": message,
				},
			},
		},
	}
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
