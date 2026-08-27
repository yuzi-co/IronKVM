package stream

import (
	"sync/atomic"
	"testing"
	"time"
)

// withCapture points the source at a fake encoder for the length of one test.
func withCapture(t *testing.T, read func(uint16, uint16, uint16) ([]byte, int)) {
	t.Helper()

	original := readH264
	t.Cleanup(func() { readH264 = original })
	readH264 = read
}

func waitFor(t *testing.T, what string, condition func() bool) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for %s", what)
}

// The point of the shared source. One encoder sits behind ReadH264 and hands
// each caller whatever frame is ready, so two loops calling it split the stream
// rather than each receiving it. Both subscribers have to see the same frames.
func TestEverySubscriberSeesTheSameFrames(t *testing.T) {
	var reads atomic.Int64
	withCapture(t, func(uint16, uint16, uint16) ([]byte, int) {
		reads.Add(1)
		return []byte{0x00, 0x00, 0x00, 0x01}, 3
	})

	source := newH264Source()
	first := source.subscribe(nil)
	defer first.Close()
	second := source.subscribe(nil)
	defer second.Close()

	for _, subscription := range []*H264Subscription{first, second} {
		select {
		case frame, ok := <-subscription.Frames():
			if !ok {
				t.Fatal("the subscription closed before a frame arrived")
			}
			if len(frame.Data) != 4 || !frame.KeyFrame || frame.Result != 3 {
				t.Fatalf("frame = %+v, want the captured keyframe", frame)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("a subscriber received no frame")
		}
	}

	// Both were served, and the encoder was not read twice to do it.
	if got := reads.Load(); got > 2 {
		t.Errorf("the encoder was read %d times for one frame each to two subscribers", got)
	}
}

// A subscriber that is not taking frames must not stop the other one. This is
// the whole reason the hand-off is a slot rather than a channel the producer
// waits on.
func TestASubscriberThatTakesNothingDoesNotStopTheOthers(t *testing.T) {
	withCapture(t, func(uint16, uint16, uint16) ([]byte, int) {
		return []byte{0x01}, 1
	})

	source := newH264Source()
	stalled := source.subscribe(nil)
	defer stalled.Close()
	live := source.subscribe(nil)
	defer live.Close()

	taken := 0
	deadline := time.After(2 * time.Second)
	for taken < 5 {
		select {
		case <-live.Frames():
			taken++
		case <-deadline:
			t.Fatalf("the live subscriber took %d frames while the other took none", taken)
		}
	}

	if dropped := stalled.Dropped(); dropped == 0 {
		t.Error("the stalled subscriber reports no dropped frames, so nothing was refused for it")
	}
}

// Reading the encoder costs the board whether or not anyone wants the frame,
// and direct mode stops asking while its viewer is behind. With every
// subscriber quiet the loop must not read at all.
func TestNoSubscriberWantsAFrameSoNothingIsRead(t *testing.T) {
	var reads atomic.Int64
	withCapture(t, func(uint16, uint16, uint16) ([]byte, int) {
		reads.Add(1)
		return []byte{0x01}, 1
	})

	source := newH264Source()
	var wanted atomic.Bool
	subscription := source.subscribe(func() bool { return wanted.Load() })
	defer subscription.Close()

	time.Sleep(200 * time.Millisecond)
	if got := reads.Load(); got != 0 {
		t.Fatalf("the encoder was read %d times with nobody asking", got)
	}

	wanted.Store(true)
	waitFor(t, "a read once a subscriber asks", func() bool { return reads.Load() > 0 })
}

// The loop belongs to nobody once the last path leaves, and a loop left running
// reads the encoder for an empty room.
func TestTheLoopStopsAfterTheLastSubscriberLeaves(t *testing.T) {
	var reads atomic.Int64
	withCapture(t, func(uint16, uint16, uint16) ([]byte, int) {
		reads.Add(1)
		return []byte{0x01}, 1
	})

	source := newH264Source()
	subscription := source.subscribe(nil)
	waitFor(t, "the loop to start", func() bool { return reads.Load() > 0 })

	subscription.Close()
	waitFor(t, "the loop to stop", func() bool {
		source.mutex.Lock()
		defer source.mutex.Unlock()

		return !source.running
	})

	settled := reads.Load()
	time.Sleep(200 * time.Millisecond)
	if got := reads.Load(); got != settled {
		t.Errorf("the encoder was read %d more times after the last subscriber left", got-settled)
	}
}

// A later subscriber has to start the loop again rather than wait on one that
// has stopped.
func TestASubscriberArrivingAfterTheStopStartsTheLoopAgain(t *testing.T) {
	var reads atomic.Int64
	withCapture(t, func(uint16, uint16, uint16) ([]byte, int) {
		reads.Add(1)
		return []byte{0x01}, 1
	})

	source := newH264Source()
	first := source.subscribe(nil)
	waitFor(t, "the first loop to start", func() bool { return reads.Load() > 0 })
	first.Close()
	waitFor(t, "the first loop to stop", func() bool {
		source.mutex.Lock()
		defer source.mutex.Unlock()

		return !source.running
	})

	settled := reads.Load()
	second := source.subscribe(nil)
	defer second.Close()
	waitFor(t, "the loop to run again", func() bool { return reads.Load() > settled })
}

// Closing twice is what a deferred Close beside an explicit one does, and it
// must not panic on the slot's channel.
func TestClosingASubscriptionTwiceIsSafe(t *testing.T) {
	withCapture(t, func(uint16, uint16, uint16) ([]byte, int) { return nil, -1 })

	source := newH264Source()
	subscription := source.subscribe(nil)
	subscription.Close()
	subscription.Close()
}

// A capture failure still reaches every path, because each one reports the
// status under its own mode and a status nobody reports is a stream that fails
// with nothing said.
func TestACaptureFailureIsDeliveredRatherThanSwallowed(t *testing.T) {
	withCapture(t, func(uint16, uint16, uint16) ([]byte, int) { return nil, -1 })

	source := newH264Source()
	subscription := source.subscribe(nil)
	defer subscription.Close()

	select {
	case frame := <-subscription.Frames():
		if frame.Result != -1 || len(frame.Data) != 0 {
			t.Fatalf("frame = %+v, want the failure passed through", frame)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the capture failure never reached the subscriber")
	}
}
