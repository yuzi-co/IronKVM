package picoclaw

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"sync"
	"testing"

	"NanoKVM-Server/common"
)

// hdmiVision behaves like libkvm on a board whose stream is set to the
// automatic resolution and that no browser is watching: a 0x0 request gets
// the full HDMI frame, any other size gets a frame scaled to that size.
type hdmiVision struct {
	mu       sync.Mutex
	width    uint16
	height   uint16
	requests [][2]uint16
}

func (v *hdmiVision) ReadMjpeg(width uint16, height uint16, quality uint16) ([]byte, int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.requests = append(v.requests, [2]uint16{width, height})
	if width == 0 || height == 0 {
		width, height = v.width, v.height
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewGray(image.Rect(0, 0, int(width), int(height))), nil); err != nil {
		return nil, -1
	}
	return buf.Bytes(), 0
}

func useAutoResolution(t *testing.T) {
	t.Helper()
	useScreen1080p(t)
	common.SetScreen("resolution", 0)
}

func TestModelScreenshotIsDownscaledWithoutAViewer(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)
	useAutoResolution(t)
	vision := &hdmiVision{width: 1920, height: 1080}
	service.vision = vision

	data, meta, err := service.captureScreenshot(context.Background(), ScreenshotQuery{Format: "base64"})
	if err != nil {
		t.Fatalf("capture: %v", err.Message)
	}
	cfg, decodeErr := jpeg.DecodeConfig(bytes.NewReader(data))
	if decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if cfg.Width != 960 || cfg.Height != 540 {
		t.Fatalf("image is %dx%d, want 960x540 (requests %v)", cfg.Width, cfg.Height, vision.requests)
	}
	if got := screenshotCaption(meta); got != "screenshot captured: 960x540 image of a 1920x1080 screen" {
		t.Fatalf("caption = %q", got)
	}
}

func TestModelScreenshotAtFullSizeWithoutAViewerReadsOnce(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)
	useAutoResolution(t)
	vision := &hdmiVision{width: 1920, height: 1080}
	service.vision = vision

	_, meta, err := service.captureScreenshot(context.Background(), ScreenshotQuery{Format: "base64", Width: 1920, Height: 1080})
	if err != nil {
		t.Fatalf("capture: %v", err.Message)
	}
	if len(vision.requests) != 1 {
		t.Fatalf("requests = %v, want one full-size read", vision.requests)
	}
	if got := screenshotCaption(meta); got != "screenshot captured: 1920x1080 image of a 1920x1080 screen" {
		t.Fatalf("caption = %q", got)
	}
}

func TestModelScreenshotCaptionUsesTheFrameSize(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)
	// The stream is set to 1080p but the request and the frame disagree:
	// the caption follows the frame.
	vision := &hdmiVision{width: 1920, height: 1080}
	service.vision = vision

	_, meta, err := service.captureScreenshot(context.Background(), ScreenshotQuery{Format: "base64", Width: 640})
	if err != nil {
		t.Fatalf("capture: %v", err.Message)
	}
	if meta.CaptureWidth != 640 || meta.CaptureHeight != 360 || meta.SourceWidth != 1920 || meta.SourceHeight != 1080 {
		t.Fatalf("meta = %+v", meta)
	}
}
