package stream

import (
	"sync"
	"testing"
	"time"

	"NanoKVM-Server/common"
)

// A library that says when its frames were captured gets them on the frames,
// and the frames' timestamps follow the capture, not the reads.
func TestFramesCarryTheCaptureTimes(t *testing.T) {
	var mutex sync.Mutex
	captured := time.Now()
	withCapture(t, func(uint16, uint16, uint16) ([]byte, int) {
		mutex.Lock()
		captured = captured.Add(20 * time.Millisecond)
		mutex.Unlock()
		return []byte{0x00, 0x00, 0x00, 0x01}, 4
	})

	original := videoTimes
	t.Cleanup(func() { videoTimes = original })
	videoTimes = func() (common.VideoTimes, bool) {
		mutex.Lock()
		defer mutex.Unlock()
		return common.VideoTimes{
			CaptureEnd:  captured,
			FrameTime:   16 * time.Millisecond,
			EncodeStart: captured.Add(1 * time.Millisecond),
			EncodeEnd:   captured.Add(13 * time.Millisecond),
			Returned:    captured.Add(14 * time.Millisecond),
		}, true
	}

	source := newH264Source()
	subscription := source.subscribe(nil)
	defer subscription.Close()

	var frames []H264Frame
	deadline := time.After(3 * time.Second)
	for len(frames) < 3 {
		select {
		case frame := <-subscription.Frames():
			frames = append(frames, frame)
		case <-deadline:
			t.Fatalf("%d frames in 3 s", len(frames))
		}
	}

	for i, frame := range frames {
		if !frame.HasTimes {
			t.Fatalf("frame %d has no capture times", i)
		}
		if got := frame.Times.CaptureStart(); !got.Equal(frame.Times.CaptureEnd.Add(-16 * time.Millisecond)) {
			t.Fatalf("frame %d capture start %s", i, got)
		}
	}

	// 20 ms apart in capture, so 20 ms apart in the stream's clock,
	// whenever the reads ran.
	for i := 1; i < len(frames); i++ {
		gap := frames[i].Timestamp - frames[i-1].Timestamp
		if gap < 19_000 || gap > 21_000 {
			t.Fatalf("timestamps %d us apart, want the capture's 20 ms", gap)
		}
	}
}
