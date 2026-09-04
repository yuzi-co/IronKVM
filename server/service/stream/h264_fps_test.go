package stream

import (
	"sync"
	"testing"
	"time"

	"NanoKVM-Server/common"
)

// withEncoderFPS records what the loop tells the encoder, for one test.
func withEncoderFPS(t *testing.T) func() []int {
	t.Helper()

	var mutex sync.Mutex
	var told []int

	original := setEncoderFPS
	t.Cleanup(func() { setEncoderFPS = original })
	setEncoderFPS = func(fps int) {
		mutex.Lock()
		defer mutex.Unlock()
		told = append(told, fps)
	}

	return func() []int {
		mutex.Lock()
		defer mutex.Unlock()

		return append([]int(nil), told...)
	}
}

// withScreenFPS puts one FPS setting in place and puts the previous one back.
// The screen singleton is process wide, so the value has to be restored or the
// next test in this package reads it.
func withScreenFPS(t *testing.T, fps int) {
	t.Helper()

	screen := common.GetScreen()
	before := screen.Snapshot().FPS
	t.Cleanup(func() { common.SetScreen("fps", before) })
	common.SetScreen("fps", fps)
}

// The encoder works out what one frame may cost by dividing the bitrate by the
// frame rate it was given. It has a compiled-in default and no way to learn the
// configured rate, so a loop that only spoke up on a change would leave every
// stream that never changes its setting running against the wrong number - and
// that is every stream on a board nobody reconfigures.
func TestTheEncoderIsToldTheFrameRateAtTheStart(t *testing.T) {
	withScreenFPS(t, 30)
	told := withEncoderFPS(t)
	withCapture(t, func(uint16, uint16, uint16) ([]byte, int) {
		return []byte{0x00, 0x00, 0x00, 0x01}, 3
	})

	source := newH264Source()
	subscription := source.subscribe(nil)
	defer subscription.Close()

	waitFor(t, "the encoder to be told the frame rate", func() bool {
		return len(told()) > 0
	})

	if got := told()[0]; got != 30 {
		t.Fatalf("the encoder was told %d, want the configured 30", got)
	}
}

func TestAChangedFrameRateReachesTheEncoder(t *testing.T) {
	withScreenFPS(t, 30)
	told := withEncoderFPS(t)
	withCapture(t, func(uint16, uint16, uint16) ([]byte, int) {
		return []byte{0x00, 0x00, 0x00, 0x01}, 3
	})

	source := newH264Source()
	subscription := source.subscribe(nil)
	defer subscription.Close()

	waitFor(t, "the first frame rate", func() bool { return len(told()) > 0 })

	common.SetScreen("fps", 60)

	waitFor(t, "the new frame rate", func() bool {
		values := told()

		return len(values) > 1 && values[len(values)-1] == 60
	})
}

// The loop reads the settings on every tick. Repeating the same number to the
// encoder is not harmless: set_h264_fps clears enc_h264_init, the next frame
// rebuilds the channel, and a rebuilt channel starts with a keyframe. Doing
// that thirty times a second would be a stream of nothing but keyframes.
func TestAnUnchangedFrameRateIsNotRepeated(t *testing.T) {
	withScreenFPS(t, 60)
	told := withEncoderFPS(t)

	frames := make(chan struct{}, 64)
	withCapture(t, func(uint16, uint16, uint16) ([]byte, int) {
		select {
		case frames <- struct{}{}:
		default:
		}

		return []byte{0x00, 0x00, 0x00, 0x01}, 3
	})

	source := newH264Source()
	subscription := source.subscribe(nil)
	defer subscription.Close()

	// Wait for several captures, so the loop has gone round many more times
	// than the one that speaks to the encoder.
	for i := 0; i < 5; i++ {
		select {
		case <-frames:
		case <-time.After(2 * time.Second):
			t.Fatal("the capture loop did not run")
		}
	}

	if values := told(); len(values) != 1 {
		t.Fatalf("the encoder was told %v, want one call at the start", values)
	}
}
