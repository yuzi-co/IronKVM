package stream

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func withSourceFiles(t *testing.T) (width, height, refresh string) {
	t.Helper()
	dir := t.TempDir()
	width = filepath.Join(dir, "width")
	height = filepath.Join(dir, "height")
	refresh = filepath.Join(dir, "refresh")

	oldWidth, oldHeight, oldRefresh := sourceWidthFile, sourceHeightFile, sourceRefreshFile
	sourceWidthFile, sourceHeightFile, sourceRefreshFile = width, height, refresh
	t.Cleanup(func() {
		sourceWidthFile, sourceHeightFile, sourceRefreshFile = oldWidth, oldHeight, oldRefresh
	})

	return width, height, refresh
}

func write(t *testing.T, path, value string, at time.Time) {
	t.Helper()
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, at, at); err != nil {
		t.Fatal(err)
	}
}

func TestStreamInfoReadsTheSourceMode(t *testing.T) {
	width, height, refresh := withSourceFiles(t)
	now := time.Now()
	write(t, width, "1920\n", now)
	write(t, height, "1080\n", now)
	write(t, refresh, "60000\n", now)

	info := readStreamInfo()
	if info.Width != 1920 || info.Height != 1080 || info.Refresh != 60000 {
		t.Fatalf("got %+v, want 1920x1080 at 60000 mHz", info)
	}
}

func TestStreamInfoSaysWhetherFramesCarryCaptureTimes(t *testing.T) {
	t.Cleanup(func() { captureStamps.Store(false) })

	captureStamps.Store(false)
	if readStreamInfo().CaptureStamps {
		t.Fatal("capture stamps reported without capture times")
	}

	captureStamps.Store(true)
	if !readStreamInfo().CaptureStamps {
		t.Fatal("capture stamps not reported")
	}
}

func TestARefreshRateOlderThanTheSizeIsNotReported(t *testing.T) {
	// Sipeed's library writes the size and never the rate, so a rate file
	// from an earlier libkvm-v4l2 run may describe another mode.
	width, height, refresh := withSourceFiles(t)
	now := time.Now()
	write(t, refresh, "60000", now.Add(-time.Hour))
	write(t, width, "1280", now)
	write(t, height, "720", now)

	if info := readStreamInfo(); info.Refresh != 0 || info.Width != 1280 {
		t.Fatalf("got %+v, want 1280 wide and no refresh rate", info)
	}
}

func TestStreamInfoWithoutFilesIsZero(t *testing.T) {
	withSourceFiles(t)
	info := readStreamInfo()
	if info.Width != 0 || info.Height != 0 || info.Refresh != 0 {
		t.Fatalf("got %+v, want zeros", info)
	}
}
