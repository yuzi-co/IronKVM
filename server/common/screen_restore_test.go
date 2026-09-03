package common

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// resetScreen forces the next GetScreen to build the singleton again, so a test
// can watch it being seeded. The cleanup drops it rather than putting the old
// pointer back: with the settings files pointed at a temporary directory that
// no longer exists, the next build falls back to the defaults, which is what
// every other test in this package expects.
func resetScreen(t *testing.T) {
	t.Helper()

	reset := func() {
		screen = nil
		screenOnce = sync.Once{}
	}

	reset()
	t.Cleanup(reset)
}

// withScreenFiles points the settings at a temporary directory and writes the
// named ones. A key that is not named is left absent, which is how an
// unconfigured board looks.
func withScreenFiles(t *testing.T, stored map[string]string) {
	t.Helper()

	dir := t.TempDir()

	original := ScreenFileMap
	t.Cleanup(func() { ScreenFileMap = original })

	ScreenFileMap = map[string]string{}
	for key, path := range original {
		ScreenFileMap[key] = filepath.Join(dir, filepath.Base(path))
	}

	for key, value := range stored {
		path, ok := ScreenFileMap[key]
		if !ok {
			t.Fatalf("no file is mapped for %q", key)
		}
		if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
			t.Fatalf("write %s: %s", path, err)
		}
	}
}

// The defect this closes. The files outlive the process and nothing read them,
// so a board set to 60 frames served 30 after every restart until somebody
// opened the UI.
func TestStoredSettingsAreRestored(t *testing.T) {
	withScreenFiles(t, map[string]string{
		"fps":        "60",
		"quality":    "5000",
		"resolution": "720",
	})

	values := loadScreenValues()

	if values.FPS != 60 {
		t.Errorf("fps = %d, want 60", values.FPS)
	}
	if values.BitRate != 5000 {
		t.Errorf("bitrate = %d, want 5000", values.BitRate)
	}
	if values.Width != 1280 || values.Height != 720 {
		t.Errorf("resolution = %dx%d, want 1280x720", values.Width, values.Height)
	}
}

// The whole feature is one line in GetScreen. Restoring the values and then
// building the singleton from the defaults anyway would pass every case above.
func TestTheSingletonIsSeededFromTheStoredSettings(t *testing.T) {
	withScreenFiles(t, map[string]string{"fps": "60", "resolution": "720"})
	resetScreen(t)

	values := GetScreen().Snapshot()

	if values.FPS != 60 {
		t.Errorf("fps = %d, want 60", values.FPS)
	}
	if values.Width != 1280 || values.Height != 720 {
		t.Errorf("resolution = %dx%d, want 1280x720", values.Width, values.Height)
	}
}

// An unconfigured board is the ordinary case, not a fault.
func TestNoStoredSettingsLeavesTheDefaults(t *testing.T) {
	withScreenFiles(t, nil)

	if values := loadScreenValues(); values != defaultScreenValues {
		t.Fatalf("values = %+v, want the defaults %+v", values, defaultScreenValues)
	}
}

// One API key carries either a JPEG quality or an H.264 bitrate depending on
// its size, and both are written to the same file. Reading it back has to make
// the same distinction or a board set to quality 60 comes up at bitrate 60.
func TestTheOverloadedQualityFieldIsReadBackTheWayItWasWritten(t *testing.T) {
	withScreenFiles(t, map[string]string{"quality": "60"})

	values := loadScreenValues()

	if values.Quality != 60 {
		t.Errorf("quality = %d, want 60", values.Quality)
	}
	if values.BitRate != defaultScreenValues.BitRate {
		t.Errorf("bitrate = %d, want the default %d", values.BitRate, defaultScreenValues.BitRate)
	}
}

// A file the operator or a half-finished write left unreadable must cost its
// own setting and nothing else.
func TestAnUnreadableSettingDoesNotTakeTheOthersWithIt(t *testing.T) {
	withScreenFiles(t, map[string]string{
		"fps":        "not a number",
		"resolution": "1080",
	})

	values := loadScreenValues()

	if values.FPS != defaultScreenValues.FPS {
		t.Errorf("fps = %d, want the default %d", values.FPS, defaultScreenValues.FPS)
	}
	if values.Width != 1920 || values.Height != 1080 {
		t.Errorf("resolution = %dx%d, want 1920x1080", values.Width, values.Height)
	}
}

// The stored value goes through the same rule the API applies, so a resolution
// the board does not offer is ignored rather than configured.
func TestAnUnknownResolutionIsIgnored(t *testing.T) {
	withScreenFiles(t, map[string]string{"resolution": "1440"})

	values := loadScreenValues()

	if values.Width != defaultScreenValues.Width || values.Height != defaultScreenValues.Height {
		t.Fatalf("resolution = %dx%d, want the default %dx%d",
			values.Width, values.Height, defaultScreenValues.Width, defaultScreenValues.Height)
	}
}

// validateFPS is the rule the API uses, and a stored value has to meet it too:
// a file holding 240 must not configure the capture loop for a 4ms tick.
func TestAStoredFPSIsClamped(t *testing.T) {
	withScreenFiles(t, map[string]string{"fps": "240"})

	if values := loadScreenValues(); values.FPS != 60 {
		t.Fatalf("fps = %d, want 60", values.FPS)
	}
}

// The trailing newline a shell redirect leaves behind is not a parse failure.
func TestAStoredValueMayCarryWhitespace(t *testing.T) {
	withScreenFiles(t, map[string]string{"fps": " 50\n"})

	if values := loadScreenValues(); values.FPS != 50 {
		t.Fatalf("fps = %d, want 50", values.FPS)
	}
}
