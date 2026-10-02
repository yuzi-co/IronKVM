package picoclaw

import (
	"testing"

	"NanoKVM-Server/common"
	"NanoKVM-Server/config"
)

func usePicoclawScreenshotSettings(t *testing.T, settings config.Picoclaw) {
	t.Helper()
	previous := picoclawScreenshotSettings
	picoclawScreenshotSettings = func() config.Picoclaw { return settings.WithDefaults() }
	t.Cleanup(func() { picoclawScreenshotSettings = previous })
}

func useScreen1080p(t *testing.T) {
	t.Helper()
	before := common.GetScreen().Snapshot()
	common.SetScreen("resolution", 1080)
	common.SetScreen("quality", 80)
	t.Cleanup(func() {
		common.SetScreen("resolution", int(before.Height))
		common.SetScreen("quality", int(before.Quality))
	})
}

func TestModelScreenshotDefaultsTo960At60(t *testing.T) {
	useScreen1080p(t)
	usePicoclawScreenshotSettings(t, config.Picoclaw{})

	width, height, quality := resolveScreenshotRequest(ScreenshotQuery{Format: "base64"}, common.GetScreen().Snapshot())
	if width != 960 || height != 540 || quality != 60 {
		t.Fatalf("got %dx%d q%d, want 960x540 q60", width, height, quality)
	}
}

func TestModelScreenshotFollowsServerSettings(t *testing.T) {
	useScreen1080p(t)
	usePicoclawScreenshotSettings(t, config.Picoclaw{ScreenshotWidth: 640, ScreenshotQuality: 45})

	width, height, quality := resolveScreenshotRequest(ScreenshotQuery{Format: "base64"}, common.GetScreen().Snapshot())
	if width != 640 || height != 360 || quality != 45 {
		t.Fatalf("got %dx%d q%d, want 640x360 q45", width, height, quality)
	}
}

func TestModelScreenshotRequestOverridesServerSettings(t *testing.T) {
	useScreen1080p(t)
	usePicoclawScreenshotSettings(t, config.Picoclaw{ScreenshotWidth: 640, ScreenshotQuality: 45})

	// A model that needs to read small text or hit a small target asks for
	// the full size. A width alone only shrinks the default; width and height
	// together are taken as they are.
	width, height, quality := resolveScreenshotRequest(ScreenshotQuery{Format: "base64", Width: 1920, Height: 1080, Quality: 80}, common.GetScreen().Snapshot())
	if width != 1920 || height != 1080 || quality != 80 {
		t.Fatalf("got %dx%d q%d, want 1920x1080 q80", width, height, quality)
	}
}

func TestRawScreenshotIgnoresModelSettings(t *testing.T) {
	useScreen1080p(t)
	usePicoclawScreenshotSettings(t, config.Picoclaw{ScreenshotWidth: 640, ScreenshotQuality: 45})

	width, height, quality := resolveScreenshotRequest(ScreenshotQuery{}, common.GetScreen().Snapshot())
	if width != 1920 || height != 1080 || quality != 80 {
		t.Fatalf("got %dx%d q%d, want the stream's 1920x1080 q80", width, height, quality)
	}
}

func TestScreenshotCaptionGivesImageAndScreenSize(t *testing.T) {
	got := screenshotCaption(ScreenshotMeta{SourceWidth: 1920, SourceHeight: 1080, CaptureWidth: 960, CaptureHeight: 540})
	if got != "screenshot captured: 960x540 image of a 1920x1080 screen" {
		t.Fatalf("caption = %q", got)
	}
}
