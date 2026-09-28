package mjpeg

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// recordFrameDetect swaps the hardware call for a recorder and returns the
// values it was handed, newest last.
func recordFrameDetect(t *testing.T) func() []uint8 {
	t.Helper()

	var (
		mutex  sync.Mutex
		values []uint8
	)

	original := setFrameDetect
	originalEnabled := detectEnabled
	// The pause tests pause detection that is on.
	detectEnabled = true
	setFrameDetect = func(frames uint8) {
		mutex.Lock()
		defer mutex.Unlock()
		values = append(values, frames)
	}

	t.Cleanup(func() {
		setFrameDetect = original
		detectEnabled = originalEnabled
		resetFrameDetectPause()
	})

	resetFrameDetectPause()

	return func() []uint8 {
		mutex.Lock()
		defer mutex.Unlock()

		return append([]uint8(nil), values...)
	}
}

func TestPauseDurationDefaultsWhenUnset(t *testing.T) {
	if got := pauseDuration(0); got != defaultPauseDuration {
		t.Fatalf("pauseDuration(0) = %s, want %s", got, defaultPauseDuration)
	}
}

func TestPauseDurationDefaultsWhenNegative(t *testing.T) {
	if got := pauseDuration(-5); got != defaultPauseDuration {
		t.Fatalf("pauseDuration(-5) = %s, want %s", got, defaultPauseDuration)
	}
}

// Detection is what notices the screen has changed. A caller must not be able
// to switch it off for a week.
func TestPauseDurationIsCapped(t *testing.T) {
	if got := pauseDuration(1_000_000_000); got != maxPauseDuration {
		t.Fatalf("pauseDuration(1e9) = %s, want %s", got, maxPauseDuration)
	}
}

func TestPauseFrameDetectStopsThenResumes(t *testing.T) {
	values := recordFrameDetect(t)

	pauseFrameDetect(80 * time.Millisecond)

	if got := values(); len(got) != 1 || got[0] != 0 {
		t.Fatalf("after pause, calls = %v, want [0]", got)
	}

	time.Sleep(200 * time.Millisecond)

	got := values()
	if len(got) != 2 || got[1] != FrameDetectInterval {
		t.Fatalf("after the pause elapsed, calls = %v, want [0 %d]", got, FrameDetectInterval)
	}
}

// A second, shorter request must not cut short a pause someone else is still
// relying on.
func TestShorterPauseDoesNotCutLongerOneShort(t *testing.T) {
	values := recordFrameDetect(t)

	pauseFrameDetect(400 * time.Millisecond)
	pauseFrameDetect(50 * time.Millisecond)

	time.Sleep(200 * time.Millisecond)

	if got := values(); len(got) != 1 {
		t.Fatalf("detection resumed early: calls = %v, want just [0]", got)
	}

	time.Sleep(400 * time.Millisecond)

	got := values()
	if len(got) != 2 || got[1] != FrameDetectInterval {
		t.Fatalf("detection never resumed: calls = %v", got)
	}
}

// A pause returns to the operator's setting. Detection that is off stays off,
// and is not switched on by a pause that ends.
func TestPauseLeavesDisabledDetectionOff(t *testing.T) {
	values := recordFrameDetect(t)
	setFrameDetectEnabled(false)

	pauseFrameDetect(50 * time.Millisecond)
	time.Sleep(150 * time.Millisecond)

	if got := values(); len(got) != 1 || got[0] != 0 {
		t.Fatalf("calls = %v, want just the [0] from switching it off", got)
	}
	if frameDetectEnabled() {
		t.Fatal("setting reads as on after a pause")
	}
}

// Switching detection off during a pause keeps it off when the pause ends.
func TestSettingDuringPauseWins(t *testing.T) {
	values := recordFrameDetect(t)

	pauseFrameDetect(80 * time.Millisecond)
	setFrameDetectEnabled(false)
	time.Sleep(200 * time.Millisecond)

	got := values()
	if len(got) != 2 || got[1] != 0 {
		t.Fatalf("calls = %v, want [0 0]", got)
	}
}

func TestGetFrameDetectReportsSetting(t *testing.T) {
	recordFrameDetect(t)
	gin.SetMode(gin.TestMode)

	for _, enabled := range []bool{true, false} {
		setFrameDetectEnabled(enabled)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/stream/mjpeg/detect", nil)
		GetFrameDetect(c)

		var body struct {
			Code int `json:"code"`
			Data struct {
				Enabled bool `json:"enabled"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode %q: %v", w.Body.String(), err)
		}
		if body.Code != 0 || body.Data.Enabled != enabled {
			t.Fatalf("enabled=%t: answer %s", enabled, w.Body.String())
		}
	}
}
