package mjpeg

import (
	"sync/atomic"
	"testing"
	"time"
)

// A subscriber gets frames from the same loop as the HTTP viewers, and the loop
// reads nothing more while the subscriber still holds the frame it was given.
func TestASubscriberReceivesFramesAndPacesTheLoop(t *testing.T) {
	withScreenFPS(t, 60)
	withCaptureFPS(t)
	withRefreshInterval(t, time.Minute)

	reads := withReadMjpeg(t, func(n int) ([]byte, int) {
		return []byte{byte(n), byte(n >> 8)}, 0
	})

	s := NewStreamer()
	sub := s.subscribe()
	defer sub.Close()

	var frame []byte
	select {
	case frame = <-sub.Frames():
	case <-time.After(2 * time.Second):
		t.Fatal("the subscriber received no frame")
	}
	if len(frame) != 2 {
		t.Fatalf("the subscriber received %v, want a two-byte frame", frame)
	}

	// The next frame is read and left in the slot. After that the loop has
	// to wait for the subscriber.
	waitFor(t, "a pending frame", func() bool { return sub.client.slot.Pending() })
	held := reads()
	time.Sleep(150 * time.Millisecond)
	if got := reads(); got != held {
		t.Fatalf("the loop read %d frames while the subscriber held one", got-held)
	}

	select {
	case <-sub.Frames():
	case <-time.After(time.Second):
		t.Fatal("the pending frame was not delivered")
	}
	waitFor(t, "a read after the frame was taken", func() bool { return reads() > held })
}

// Closing the last subscription stops the loop and closes the channel, so a
// subscriber that waits on it wakes up.
func TestClosingASubscriptionStopsTheLoop(t *testing.T) {
	withScreenFPS(t, 60)
	withCaptureFPS(t)
	withReadMjpeg(t, func(n int) ([]byte, int) { return []byte{byte(n)}, 0 })

	s := NewStreamer()
	sub := s.subscribe()
	sub.Close()
	sub.Close()

	waitFor(t, "the loop to stop", func() bool { return atomic.LoadInt32(&s.running) == 0 })

	deadline := time.After(time.Second)
	for {
		select {
		case _, ok := <-sub.Frames():
			if !ok {
				return
			}
		case <-deadline:
			t.Fatal("the frame channel stayed open after Close")
		}
	}
}

// Refresh sends the next frame even when it matches the last one.
func TestRefreshResendsAnUnchangedFrame(t *testing.T) {
	withScreenFPS(t, 60)
	withCaptureFPS(t)
	withRefreshInterval(t, time.Minute)
	withReadMjpeg(t, func(int) ([]byte, int) { return []byte("still"), 0 })

	s := NewStreamer()
	sub := s.subscribe()
	defer sub.Close()

	select {
	case <-sub.Frames():
	case <-time.After(2 * time.Second):
		t.Fatal("the subscriber received no frame")
	}

	select {
	case <-sub.Frames():
		t.Fatal("an unchanged frame was sent without a refresh")
	case <-time.After(150 * time.Millisecond):
	}

	sub.Refresh()
	select {
	case <-sub.Frames():
	case <-time.After(2 * time.Second):
		t.Fatal("Refresh did not resend the frame")
	}
}
