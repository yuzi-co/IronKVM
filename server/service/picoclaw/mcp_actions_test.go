package picoclaw

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"NanoKVM-Server/config"
	"NanoKVM-Server/service/controlmode"

	"github.com/gin-gonic/gin"
)

type fakeVision struct {
	mu     sync.Mutex
	calls  int
	width  uint16
	height uint16
	result int
}

func (v *fakeVision) ReadMjpeg(width uint16, height uint16, quality uint16) ([]byte, int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.calls++
	v.width, v.height = width, height
	if v.result != 0 {
		return nil, v.result
	}
	return []byte{0xff, 0xd8, 0xff, 0xd9}, 0
}

type fakeHID struct {
	mu    sync.Mutex
	mouse [][]byte
	keys  [][]byte
}

func (h *fakeHID) WriteHid0(data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.keys = append(h.keys, append([]byte(nil), data...))
}

func (h *fakeHID) WriteHid1(data []byte) {}

func (h *fakeHID) WriteHid2(data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.mouse = append(h.mouse, append([]byte(nil), data...))
}

func (h *fakeHID) lastMouse() []byte {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.mouse) == 0 {
		return nil
	}
	return h.mouse[len(h.mouse)-1]
}

func newMCPActionTestService(t *testing.T) (*Service, *fakeVision, *fakeHID) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	useScreen1080p(t)
	usePicoclawScreenshotSettings(t, config.Picoclaw{})
	previousDelay := screenshotRetryDelay
	screenshotRetryDelay = time.Millisecond
	t.Cleanup(func() { screenshotRetryDelay = previousDelay })

	vision := &fakeVision{}
	hid := &fakeHID{}
	service := &Service{
		vision:     vision,
		hid:        hid,
		config:     &ConfigStore{},
		lock:       &SessionLock{},
		runtime:    &RuntimeStore{},
		control:    controlmode.NewManager(filepath.Join(t.TempDir(), "mode"), controlmode.ModePicoclaw),
		operations: newControlOperationTracker(),
		acquireHDMIForRead: func(context.Context) (func(), func() bool, error) {
			return func() {}, nil, nil
		},
		releaseHID: func() error { return nil },
	}
	if err := service.control.Switch(controlmode.ModePicoclaw, func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	return service, vision, hid
}

type mcpToolResult struct {
	IsError bool `json:"isError"`
	Content []struct {
		Type     string `json:"type"`
		Text     string `json:"text"`
		Data     string `json:"data"`
		MimeType string `json:"mimeType"`
	} `json:"content"`
}

func callKVMActions(t *testing.T, service *Service, arguments string) mcpToolResult {
	t.Helper()
	router := gin.New()
	router.POST("/", service.MCPHandler)
	body := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"kvm_actions","arguments":` + arguments + `}}`
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(sessionIDHeader, "test-session")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}

	var envelope struct {
		Result mcpToolResult `json:"result"`
		Error  any           `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Error != nil {
		t.Fatalf("JSON-RPC error: %s", response.Body.String())
	}
	return envelope.Result
}

func TestKVMActionsWithoutScreenshotAfterReturnsTextOnly(t *testing.T) {
	service, vision, _ := newMCPActionTestService(t)

	result := callKVMActions(t, service, `{"actions":[{"action":"move","x":0.5,"y":0.5}]}`)
	if result.IsError || len(result.Content) != 1 || result.Content[0].Type != "text" {
		t.Fatalf("result = %+v, want one text item", result)
	}
	if vision.calls != 0 {
		t.Fatalf("captured %d frames, want none", vision.calls)
	}
}

func TestKVMActionsScreenshotAfterReturnsImage(t *testing.T) {
	service, vision, hid := newMCPActionTestService(t)

	started := time.Now()
	result := callKVMActions(t, service, `{"actions":[{"action":"click","x":0.25,"y":0.75}],"screenshot_after":true,"settle_ms":50}`)
	if elapsed := time.Since(started); elapsed < 50*time.Millisecond {
		t.Fatalf("returned after %s, before the 50 ms settle", elapsed)
	}
	if result.IsError || len(result.Content) != 3 {
		t.Fatalf("result = %+v, want action text, caption and image", result)
	}
	if !strings.Contains(result.Content[0].Text, `"hid_writes"`) {
		t.Fatalf("first item = %q, want the action result", result.Content[0].Text)
	}
	caption := result.Content[1].Text
	if caption != "screenshot captured: 960x540 image of a 1920x1080 screen, taken 50 ms after the actions" {
		t.Fatalf("caption = %q", caption)
	}
	image := result.Content[2]
	if image.Type != "image" || image.MimeType != "image/jpeg" || image.Data == "" {
		t.Fatalf("image item = %+v", image)
	}
	if vision.calls != 1 || vision.width != 960 || vision.height != 540 {
		t.Fatalf("vision calls = %d at %dx%d, want one at 960x540", vision.calls, vision.width, vision.height)
	}
	if len(hid.mouse) == 0 {
		t.Fatal("the click sent no mouse reports")
	}
}

func TestKVMActionsScreenshotFailureKeepsActionSuccess(t *testing.T) {
	service, vision, _ := newMCPActionTestService(t)
	vision.result = -1

	result := callKVMActions(t, service, `{"actions":[{"action":"move","x":0.1,"y":0.1}],"screenshot_after":true,"settle_ms":0}`)
	if result.IsError {
		t.Fatalf("result = %+v, want success: the actions ran", result)
	}
	if len(result.Content) != 2 || !strings.HasPrefix(result.Content[1].Text, "actions done; screenshot failed:") {
		t.Fatalf("result = %+v, want the failure as text", result)
	}
}

func TestKVMActionsFailureTakesNoScreenshot(t *testing.T) {
	service, vision, _ := newMCPActionTestService(t)

	result := callKVMActions(t, service, `{"actions":[{"action":"fly"}],"screenshot_after":true}`)
	if !result.IsError {
		t.Fatalf("result = %+v, want an error", result)
	}
	if vision.calls != 0 {
		t.Fatalf("captured %d frames after a failed action", vision.calls)
	}
}

func TestSettleDurationDefaultsAndBounds(t *testing.T) {
	value := func(v int) *int { return &v }
	for _, test := range []struct {
		in   *int
		want time.Duration
	}{
		{nil, defaultScreenshotSettle},
		{value(0), 0},
		{value(-10), 0},
		{value(1200), 1200 * time.Millisecond},
		{value(60_000), maxScreenshotSettle},
	} {
		if got := settleDuration(test.in); got != test.want {
			t.Fatalf("settleDuration(%v) = %s, want %s", test.in, got, test.want)
		}
	}
}
