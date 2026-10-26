package picoclaw

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"NanoKVM-Server/service/agent"
	"NanoKVM-Server/service/controlmode"
)

type VisionReader interface {
	ReadMjpeg(width uint16, height uint16, quality uint16) (data []byte, result int)
}

// HIDWriter is the part of the HID gadget PicoClaw writes to: the keyboard and
// the absolute pointer. It never writes the relative mouse endpoint: the host
// polls that one only while something on the host has the device open, so a
// report there can wait and time out. A "dx"/"dy" move goes to the remembered
// position plus the delta on the absolute pointer instead.
type HIDWriter interface {
	WriteKeyboardReport(data []byte) error
	WriteAbsoluteMouseReport(data []byte) error
}

type Service struct {
	vision             VisionReader
	hid                HIDWriter
	config             *ConfigStore
	lock               *SessionLock
	runtime            *RuntimeStore
	runtimeIntent      *RuntimeIntentStore
	control            *controlmode.Manager
	releaseHID         func() error
	operations         *controlOperationTracker
	acquireHDMIForRead func(context.Context) (func(), func() bool, error)
	chat               ChatHub
	pointer            pointerTracker
	held               heldInput
	// lifecycle runs PicoClaw through the agent-neutral lifecycle. It is
	// built on first use.
	lifecycle *agent.Lifecycle
	// mcp serves /api/picoclaw/mcp. It is built on first use.
	mcpOnce sync.Once
	mcp     http.Handler
}

type ConfigStore struct {
	mu     sync.RWMutex
	config Config
}

// RuntimeStore and RuntimeIntentStore are the agent lifecycle's stores.
type (
	RuntimeStore       = agent.StatusStore
	RuntimeIntentStore = agent.IntentStore
)

type Config struct {
	GatewayURL       string `json:"gateway_url"`
	ConnectTimeoutMs int    `json:"connect_timeout_ms"`
	ReadTimeoutMs    int    `json:"read_timeout_ms"`
	WriteTimeoutMs   int    `json:"write_timeout_ms"`
	PingIntervalMs   int    `json:"ping_interval_ms"`
	MaxMessageBytes  int    `json:"max_message_bytes"`
	AllowTokenQuery  bool   `json:"allow_token_query"`
	Token            string `json:"token,omitempty"`
}

// The runtime status and the answers of the runtime routes are the agent
// lifecycle's.
type (
	RuntimeStatus        = agent.RuntimeStatus
	RuntimeIntentStatus  = agent.RuntimeIntentStatus
	ControlStatus        = agent.ControlStatus
	RuntimeCapabilities  = agent.RuntimeCapabilities
	RuntimeStartResult   = agent.RuntimeStartResult
	RuntimeInstallResult = agent.RuntimeInstallResult
)

type ScreenshotMeta struct {
	ImageBase64   string `json:"image_base64,omitempty"`
	SourceWidth   uint16 `json:"source_width"`
	SourceHeight  uint16 `json:"source_height"`
	CaptureWidth  uint16 `json:"capture_width"`
	CaptureHeight uint16 `json:"capture_height"`
	Format        string `json:"format"`
}

type ScreenshotQuery struct {
	Format  string `form:"format"`
	Width   uint16 `form:"width"`
	Height  uint16 `form:"height"`
	Quality uint16 `form:"quality"`
}

type Point struct {
	X *float64 `json:"x"`
	Y *float64 `json:"y"`
}

type HotkeyKeys []string

func (k *HotkeyKeys) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*k = nil
		return nil
	}

	var list []string
	if err := json.Unmarshal(data, &list); err == nil {
		*k = normalizeHotkeyKeys(list)
		return nil
	}

	var csv string
	if err := json.Unmarshal(data, &csv); err == nil {
		if strings.TrimSpace(csv) == "" {
			*k = nil
			return nil
		}
		*k = normalizeHotkeyKeys(strings.Split(csv, ","))
		return nil
	}

	return &json.UnmarshalTypeError{Value: "keys", Type: hotkeyKeysType}
}

var hotkeyKeysType = reflect.TypeOf("")

func normalizeHotkeyKeys(input []string) HotkeyKeys {
	keys := make([]string, 0, len(input))
	for _, item := range input {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		keys = append(keys, trimmed)
	}
	return HotkeyKeys(keys)
}

type Action struct {
	Action     string     `json:"action"`
	X          *float64   `json:"x"`
	Y          *float64   `json:"y"`
	DX         *float64   `json:"dx"`
	DY         *float64   `json:"dy"`
	From       *Point     `json:"from"`
	To         *Point     `json:"to"`
	Button     string     `json:"button"`
	Text       string     `json:"text"`
	Layout     string     `json:"layout"`
	Keys       HotkeyKeys `json:"keys"`
	Direction  string     `json:"direction"`
	Amount     int        `json:"amount"`
	DurationMs int        `json:"duration_ms"`
}

type ActionBatch struct {
	Actions []Action `json:"actions"`
}

type ActionResult struct {
	Action          string `json:"action"`
	DurationMs      int64  `json:"duration_ms"`
	HIDWrites       int    `json:"hid_writes"`
	ExecutedActions int    `json:"executed_actions,omitempty"`
	// Pointer is where the actions left the mouse pointer, when the server
	// knows it.
	Pointer *PointerPosition `json:"pointer,omitempty"`
}

// PointerPosition is a mouse pointer position both as fractions of the screen
// and in screen pixels.
type PointerPosition struct {
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	PX int     `json:"px"`
	PY int     `json:"py"`
}

type cachedFrame struct {
	data       []byte
	width      uint16
	height     uint16
	capturedAt time.Time
}

// Close codes of the chat socket. They are the agent package's.
const (
	CloseCodePicoclawLockHeld    = agent.CloseLockHeld
	CloseCodeRuntimeUnavailable  = agent.CloseRuntimeUnavailable
	CloseCodeAuthFailed          = agent.CloseAuthFailed
	CloseCodePicoclawTakenOver   = agent.CloseTakenOver
	CloseCodeUpstreamClosed      = agent.CloseUpstreamClosed
	CloseCodeControlModeSwitched = agent.CloseControlModeSwitched
	CloseCodeRuntimeStopped      = agent.CloseRuntimeStopped
)

// ChatHub is the chat socket side the service reaches: the agent.Bridge.
type ChatHub interface {
	Publish(sessionID string, event agent.Event) error
	Broadcast(event agent.Event)
	Prompt(sessionID string, prompt agent.Prompt) error
	IsActive(sessionID string) bool
	CloseSession(sessionID string, closeCode int, reason string) bool
	CloseAll(closeCode int, reason string) int
	SessionCount() int
}

type LoadImageRequest struct {
	Path     string `json:"path"`
	Prompt   string `json:"prompt"`
	Filename string `json:"filename"`
}
