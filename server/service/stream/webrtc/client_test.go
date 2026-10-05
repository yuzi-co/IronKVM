package webrtc

import (
	"testing"
	"time"

	"github.com/pion/rtp"
)

const (
	keyFrame   = true
	deltaFrame = false
)

func frame(label string) []*rtp.Packet {
	return []*rtp.Packet{{Payload: []byte(label)}}
}

func takeLabel(t *testing.T, c *Client) string {
	t.Helper()

	packets, ok := c.queue.Take()
	if !ok {
		t.Fatal("queue was closed")
	}

	return string(packets[0].Payload)
}

func labels(w *recordingWriter) []string {
	var out []string
	for _, p := range w.packets() {
		out = append(out, string(p.Payload))
	}

	return out
}

func newTestClient() *Client {
	c := NewClient(nil, nil)
	c.track, _ = newTestTrack(5)

	return c
}

func TestEnqueueAcceptsFrameWhenClientIsKeepingUp(t *testing.T) {
	c := newTestClient()

	c.enqueue(frame("key"), keyFrame)
	c.enqueue(frame("frame"), deltaFrame)

	if c.queue.Len() != 2 {
		t.Fatal("the frames should be queued for the writer")
	}
}

// A viewer that joins a running stream starts at the next keyframe: the
// delta frames before it reference pictures it never received.
func TestEnqueueStartsANewClientAtAKeyframe(t *testing.T) {
	c := newTestClient()

	c.enqueue(frame("delta"), deltaFrame)

	if c.queue.Len() != 0 {
		t.Fatal("a delta frame before the first keyframe should not be queued")
	}
}

// Trial 38 (#72): a 1080p keyframe takes the writer 90 to 150 ms, and the
// frames that arrived meanwhile were dropped along with the rest of the GOP.
// They must wait for the writer instead.
func TestEnqueueKeepsFramesWhileTheWriterIsBusy(t *testing.T) {
	c := newTestClient()
	c.enqueue(frame("key"), keyFrame)
	c.enqueue(frame("p1"), deltaFrame)
	c.enqueue(frame("p2"), deltaFrame)
	c.enqueue(frame("p3"), deltaFrame)

	if c.queue.Dropped() != 0 {
		t.Fatalf("no frame should be dropped, dropped=%d", c.queue.Dropped())
	}

	for _, want := range []string{"key", "p1", "p2", "p3"} {
		if got := takeLabel(t, c); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	}
}

// The capture loop must never wait on a viewer: one client on a slow link
// would otherwise hold up the frame for everyone else. A client that stays
// behind loses frames up to the next keyframe.
func TestEnqueueNeverBlocksOnAClientThatStaysBehind(t *testing.T) {
	c := newTestClient()
	c.enqueue(frame("key"), keyFrame)

	for i := 0; i < 1000; i++ {
		c.enqueue(frame("delta"), deltaFrame)
	}

	if c.queue.Dropped() == 0 {
		t.Fatal("a client that never drains should drop frames")
	}

	c.enqueue(frame("key2"), keyFrame)
	if c.queue.Len() == 0 {
		t.Fatal("a keyframe should resume the stream")
	}
}

// Frames that arrive before the peer connection is up wait for it, so the
// keyframe capture starts with is the viewer's first picture.
func TestWriterHoldsFramesUntilTheConnectionIsReady(t *testing.T) {
	c := NewClient(nil, nil)
	track, writer := newTestTrack(5)
	c.track = track
	c.startWriters()
	defer c.stop()

	c.enqueue(frame("key"), keyFrame)
	c.enqueue(frame("p1"), deltaFrame)

	time.Sleep(20 * time.Millisecond)
	if n := len(labels(writer)); n != 0 {
		t.Fatalf("nothing should be written before the connection is ready, got %d packets", n)
	}

	c.markReady()

	deadline := time.Now().Add(time.Second)
	for len(labels(writer)) < 2 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}

	got := labels(writer)
	if len(got) != 2 || got[0] != "key" || got[1] != "p1" {
		t.Fatalf("expected the held keyframe and the frame after it, got %v", got)
	}
}

func TestStopReleasesTheWriter(t *testing.T) {
	c := newTestClient()
	c.startWriters()
	c.stop()

	c.enqueue(frame("late"), keyFrame)
	if c.queue.Len() != 0 {
		t.Fatal("a stopped client takes no more frames")
	}
}
