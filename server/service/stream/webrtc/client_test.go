package webrtc

import (
	"testing"

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

func newTestClient() *Client {
	c := NewClient(nil, nil)
	c.track, _ = newTestTrack(5)

	return c
}

func TestEnqueueAcceptsFrameWhenClientIsKeepingUp(t *testing.T) {
	c := newTestClient()

	c.enqueue(frame("frame"), deltaFrame)

	if c.queue.Len() != 1 {
		t.Fatal("the frame should be queued for the writer")
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

func TestStopReleasesTheWriter(t *testing.T) {
	c := newTestClient()
	c.startWriters()
	c.stop()

	c.enqueue(frame("late"), keyFrame)
	if c.queue.Len() != 0 {
		t.Fatal("a stopped client takes no more frames")
	}
}
