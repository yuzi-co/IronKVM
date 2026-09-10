package stream

import (
	"sync"
	"testing"
	"time"

	"NanoKVM-Server/common"
)

// The GOP reaches the encoder from the capture loop now, the way the frame rate
// does. It used to be sent from the HTTP handler, which had two costs. The
// handler is the only place that knew the value, so a GOP restored from the
// card was never applied at all; and reaching libkvm from a request built the
// whole capture pipeline, so changing this one menu item on an idle board
// started capture.

// withEncoderGop records what the loop tells the encoder, for one test.
func withEncoderGop(t *testing.T) func() []uint8 {
	t.Helper()

	var mutex sync.Mutex
	var told []uint8

	original := setEncoderGop
	t.Cleanup(func() { setEncoderGop = original })
	setEncoderGop = func(gop uint8) {
		mutex.Lock()
		defer mutex.Unlock()
		told = append(told, gop)
	}

	return func() []uint8 {
		mutex.Lock()
		defer mutex.Unlock()

		return append([]uint8(nil), told...)
	}
}

// withScreenGop puts one GOP setting in place and puts the previous one back.
func withScreenGop(t *testing.T, gop int) {
	t.Helper()

	screen := common.GetScreen()
	before := screen.Snapshot().GOP
	t.Cleanup(func() { common.SetScreen("gop", int(before)) })
	common.SetScreen("gop", gop)
}

// A GOP read back from the card has to reach the encoder without anyone opening
// the menu. Nothing else calls set_h264_gop, so a loop that stayed quiet would
// leave every restored setting unapplied.
func TestTheEncoderIsToldTheGopAtTheStart(t *testing.T) {
	withScreenGop(t, 12)
	told := withEncoderGop(t)
	withCapture(t, func(uint16, uint16, uint16) ([]byte, int) {
		return []byte{0x00, 0x00, 0x00, 0x01}, 3
	})

	source := newH264Source()
	subscription := source.subscribe(nil)
	defer subscription.Close()

	waitFor(t, "the encoder to be told the gop", func() bool {
		return len(told()) > 0
	})

	if got := told()[0]; got != 12 {
		t.Fatalf("the encoder was told %d, want the configured 12", got)
	}
}

func TestAChangedGopReachesTheEncoder(t *testing.T) {
	withScreenGop(t, 30)
	told := withEncoderGop(t)
	withCapture(t, func(uint16, uint16, uint16) ([]byte, int) {
		return []byte{0x00, 0x00, 0x00, 0x01}, 3
	})

	source := newH264Source()
	subscription := source.subscribe(nil)
	defer subscription.Close()

	waitFor(t, "the first gop", func() bool { return len(told()) > 0 })

	common.SetScreen("gop", 60)

	waitFor(t, "the new gop", func() bool {
		values := told()

		return len(values) > 1 && values[len(values)-1] == 60
	})
}

// set_h264_gop clears enc_video_init unconditionally, so the next frame rebuilds
// the encoder channel and a rebuilt channel starts with a keyframe. Repeating
// the same number every tick would be a stream of nothing but keyframes.
func TestAnUnchangedGopIsNotRepeated(t *testing.T) {
	withScreenGop(t, 20)
	told := withEncoderGop(t)

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
