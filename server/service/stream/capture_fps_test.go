package stream

import (
	"sync"
	"testing"

	"NanoKVM-Server/common"
)

// withCaptureFPS records what the loop tells the capture channel, for one test.
func withCaptureFPS(t *testing.T) func() []int {
	t.Helper()

	var mutex sync.Mutex
	var told []int

	original := setCaptureFPS
	t.Cleanup(func() { setCaptureFPS = original })
	setCaptureFPS = func(fps int) {
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

// The capture channel hands out every frame the source produces unless it is
// told a rate. On a 60Hz source that is twice what a default stream reads, and
// each frame nobody reads is still 3.1MB written to memory at 1080p. Nothing on
// the device sets it, so the loop that knows the rate has to say it, and it has
// to say it at the start rather than only on a later change: a board nobody
// reconfigures never sees a change.
func TestTheCaptureChannelIsToldTheFrameRateAtTheStart(t *testing.T) {
	withScreenFPS(t, 30)
	told := withCaptureFPS(t)
	withCapture(t, func(uint16, uint16, uint16) ([]byte, int) {
		return []byte{0x00, 0x00, 0x00, 0x01}, 3
	})

	source := newH264Source()
	subscription := source.subscribe(nil)
	defer subscription.Close()

	waitFor(t, "the capture channel to be told the frame rate", func() bool {
		return len(told()) > 0
	})

	if got := told()[0]; got != 30 {
		t.Fatalf("the capture channel was told %d, want the configured 30", got)
	}
}

func TestAChangedFrameRateReachesTheCaptureChannel(t *testing.T) {
	withScreenFPS(t, 30)
	told := withCaptureFPS(t)
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

// The encoder and the capture channel are told the same number, and a reader
// who finds only one of the two calls should be able to tell that is wrong.
func TestTheEncoderAndTheCaptureChannelAreToldTogether(t *testing.T) {
	withScreenFPS(t, 30)
	toldCapture := withCaptureFPS(t)
	toldEncoder := withEncoderFPS(t)
	withCapture(t, func(uint16, uint16, uint16) ([]byte, int) {
		return []byte{0x00, 0x00, 0x00, 0x01}, 3
	})

	source := newH264Source()
	subscription := source.subscribe(nil)
	defer subscription.Close()

	waitFor(t, "both to be told", func() bool {
		return len(toldCapture()) > 0 && len(toldEncoder()) > 0
	})

	common.SetScreen("fps", 60)

	waitFor(t, "both to hear the change", func() bool {
		c, e := toldCapture(), toldEncoder()

		return len(c) > 1 && c[len(c)-1] == 60 && len(e) > 1 && e[len(e)-1] == 60
	})
}
