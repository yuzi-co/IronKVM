package mjpeg

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// withReadMjpeg puts a frame source in place of the capture library and reports
// how many times the loop asked for a frame.
func withReadMjpeg(t *testing.T, source func(n int) ([]byte, int)) func() int {
	t.Helper()

	calls := make(chan struct{}, 4096)

	original := readMjpeg
	t.Cleanup(func() { readMjpeg = original })

	readMjpeg = func(_ uint16, _ uint16, _ uint16) ([]byte, int) {
		calls <- struct{}{}

		return source(len(calls))
	}

	return func() int { return len(calls) }
}

// withRefreshInterval keeps the periodic resend out of a test's way, or brings
// it close enough to observe.
func withRefreshInterval(t *testing.T, d time.Duration) {
	t.Helper()

	original := refreshInterval
	t.Cleanup(func() { refreshInterval = original })
	refreshInterval = d
}

func clientOf(s *Streamer, c *gin.Context) *client {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.clients[c]
}

// drain takes frames as fast as they arrive, standing in for a writer goroutine
// on a link that keeps up. It reports how many it received.
func drain(c *client, stop <-chan struct{}) func() int {
	received := make(chan struct{}, 4096)

	go func() {
		for {
			select {
			case <-stop:
				return
			case _, ok := <-c.slot.Channel():
				if !ok {
					return
				}
				received <- struct{}{}
			}
		}
	}()

	return func() int { return len(received) }
}

func TestAnIdenticalFrameIsNotSentTwice(t *testing.T) {
	withScreenFPS(t, 30)
	withCaptureFPS(t)
	withRefreshInterval(t, time.Minute)

	frame := []byte("the same picture, over and over")
	withReadMjpeg(t, func(int) ([]byte, int) { return frame, 0 })

	s := NewStreamer()
	ctx := addBareClient(s)
	s.forceNext.Store(true)

	stop := make(chan struct{})
	defer close(stop)
	received := drain(clientOf(s, ctx), stop)

	done := make(chan struct{})
	go func() { defer close(done); s.run() }()

	waitFor(t, "the first frame", func() bool { return received() > 0 })
	time.Sleep(400 * time.Millisecond)

	removeBareClient(s, ctx)
	<-done

	// 400ms at 30fps is about a dozen ticks. Only the first frame is new.
	if got := received(); got != 1 {
		t.Fatalf("the viewer received %d copies of one unchanged frame, want 1", got)
	}

	if s.Suppressed() == 0 {
		t.Fatal("nothing was recorded as suppressed")
	}
}

func TestAChangedFrameIsSent(t *testing.T) {
	withScreenFPS(t, 30)
	withCaptureFPS(t)
	withRefreshInterval(t, time.Minute)

	withReadMjpeg(t, func(n int) ([]byte, int) {
		return []byte{byte(n), byte(n >> 8)}, 0
	})

	s := NewStreamer()
	ctx := addBareClient(s)

	stop := make(chan struct{})
	defer close(stop)
	received := drain(clientOf(s, ctx), stop)

	done := make(chan struct{})
	go func() { defer close(done); s.run() }()

	waitFor(t, "several changed frames", func() bool { return received() >= 5 })

	removeBareClient(s, ctx)
	<-done

	if s.Suppressed() != 0 {
		t.Fatalf("%d frames were suppressed though every one differed", s.Suppressed())
	}
}

func TestNothingIsReadWhileEveryViewerStillHoldsAFrame(t *testing.T) {
	withScreenFPS(t, 30)
	withCaptureFPS(t)
	withRefreshInterval(t, time.Minute)

	reads := withReadMjpeg(t, func(n int) ([]byte, int) {
		return []byte{byte(n), byte(n >> 8)}, 0
	})

	s := NewStreamer()
	// Nothing drains this one, so its slot stays full after the first frame.
	ctx := addBareClient(s)

	done := make(chan struct{})
	go func() { defer close(done); s.run() }()

	waitFor(t, "the first read", func() bool { return reads() > 0 })

	settled := reads()
	time.Sleep(400 * time.Millisecond)
	after := reads()

	removeBareClient(s, ctx)
	<-done

	// The ticker fired about a dozen times in that window. Without the check
	// every one of them would have read, encoded and thrown away a frame.
	if after > settled+1 {
		t.Fatalf("read %d more frames while the viewer held one, want none", after-settled)
	}
}

func TestShouldSend(t *testing.T) {
	first := []byte("aaaa")
	same := []byte("aaaa")
	other := []byte("bbbb")

	t.Run("the first frame of a stream", func(t *testing.T) {
		withRefreshInterval(t, time.Minute)
		s := NewStreamer()

		if !s.shouldSend(first) {
			t.Fatal("the first frame was suppressed, so a viewer would see nothing")
		}
	})

	t.Run("an identical frame", func(t *testing.T) {
		withRefreshInterval(t, time.Minute)
		s := NewStreamer()
		s.shouldSend(first)

		if s.shouldSend(same) {
			t.Fatal("an identical frame was sent again")
		}
	})

	t.Run("a changed frame", func(t *testing.T) {
		withRefreshInterval(t, time.Minute)
		s := NewStreamer()
		s.shouldSend(first)

		if !s.shouldSend(other) {
			t.Fatal("a changed frame was suppressed")
		}
	})

	t.Run("a new viewer forces one through", func(t *testing.T) {
		withRefreshInterval(t, time.Minute)
		s := NewStreamer()
		s.shouldSend(first)

		s.forceNext.Store(true)
		if !s.shouldSend(same) {
			t.Fatal("a viewer that arrived on a still screen would have waited for the host to change something")
		}

		if s.forceNext.Load() {
			t.Fatal("the force was not spent, so it would send the next duplicate too")
		}
	})

	t.Run("the refresh resends one", func(t *testing.T) {
		withRefreshInterval(t, time.Millisecond)
		s := NewStreamer()
		s.shouldSend(first)
		time.Sleep(5 * time.Millisecond)

		if !s.shouldSend(same) {
			t.Fatal("the periodic refresh did not resend, so a lost frame would never be repaired")
		}
	})

	t.Run("a duplicate is compared against the screen, not against the last send", func(t *testing.T) {
		withRefreshInterval(t, time.Minute)
		s := NewStreamer()

		s.shouldSend(first)
		s.shouldSend(same)  // suppressed
		s.shouldSend(other) // changed, sent

		if s.shouldSend(other) {
			t.Fatal("the frame after a send was compared against the wrong thing")
		}
	})
}

func TestAllPending(t *testing.T) {
	s := NewStreamer()
	a := clientOf(s, addBareClient(s))
	b := clientOf(s, addBareClient(s))

	if allPending([]*client{a, b}) {
		t.Fatal("two empty slots reported as pending, so no frame would ever be read")
	}

	a.enqueue([]byte("x"))
	if allPending([]*client{a, b}) {
		t.Fatal("one empty slot reported as pending, so the second viewer would be starved")
	}

	b.enqueue([]byte("x"))
	if !allPending([]*client{a, b}) {
		t.Fatal("both slots full and not reported as pending")
	}
}
