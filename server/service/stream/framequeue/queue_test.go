package framequeue

import (
	"sync"
	"testing"
	"time"
)

const (
	key   = true
	delta = false
)

type clock struct{ t time.Time }

func (c *clock) now() time.Time          { return c.t }
func (c *clock) advance(d time.Duration) { c.t = c.t.Add(d) }

func newQueue(n int, d time.Duration) (*Queue[string], *clock) {
	c := &clock{t: time.Unix(1000, 0)}
	q := New[string](n, d)
	q.now = c.now

	return q, c
}

func drain(t *testing.T, q *Queue[string]) []string {
	t.Helper()

	var got []string
	for q.Len() > 0 {
		v, ok := q.Take()
		if !ok {
			t.Fatal("queue closed while frames were pending")
		}
		got = append(got, v)
	}

	return got
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestFramesComeOutInOrder(t *testing.T) {
	q, _ := newQueue(8, time.Second)

	for _, f := range []string{"k", "p1", "p2", "p3"} {
		if !q.Put(f, f == "k") {
			t.Fatalf("%s refused", f)
		}
	}

	if got := drain(t, q); !equal(got, []string{"k", "p1", "p2", "p3"}) {
		t.Fatalf("got %v", got)
	}
}

// The case that broke WebRTC at 1080p: the writer is still sending a keyframe
// when several more frames arrive. They must wait, not be dropped.
func TestSlowKeyframeDelaysTheGOPInsteadOfTruncatingIt(t *testing.T) {
	q, c := newQueue(DefaultMaxFrames, DefaultMaxDelay)

	q.Put("k", key)
	if v, _ := q.Take(); v != "k" {
		t.Fatalf("expected the keyframe, got %q", v)
	}

	// 150 ms writing the keyframe at 60 fps: nine frames arrive meanwhile.
	var want []string
	for i := 0; i < 9; i++ {
		c.advance(16 * time.Millisecond)
		f := string(rune('a' + i))
		if !q.Put(f, delta) {
			t.Fatalf("frame %d refused while the writer is inside a keyframe", i)
		}
		want = append(want, f)
	}

	if got := drain(t, q); !equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if q.Dropped() != 0 || q.Resyncs() != 0 {
		t.Fatalf("dropped=%d resyncs=%d", q.Dropped(), q.Resyncs())
	}
}

func TestFullQueueDropsToTheNextKeyframe(t *testing.T) {
	q, _ := newQueue(3, 0)

	q.Put("k", key)
	q.Put("p1", delta)
	q.Put("p2", delta)

	if q.Put("p3", delta) {
		t.Fatal("a full queue must refuse a delta frame")
	}

	// The writer catches up, but the stream already has a hole in it.
	drain(t, q)
	if q.Put("p4", delta) {
		t.Fatal("a delta frame after a drop must wait for a keyframe")
	}

	if !q.Put("k2", key) {
		t.Fatal("a keyframe resumes the stream")
	}
	if !q.Put("p5", delta) {
		t.Fatal("delta frames after the keyframe are accepted again")
	}

	if got := drain(t, q); !equal(got, []string{"k2", "p5"}) {
		t.Fatalf("got %v", got)
	}
	if q.Dropped() != 2 || q.Resyncs() != 1 {
		t.Fatalf("dropped=%d resyncs=%d", q.Dropped(), q.Resyncs())
	}
}

func TestFramesQueuedBeforeTheDropAreStillSent(t *testing.T) {
	// They are contiguous with what the viewer already has, so they decode.
	q, _ := newQueue(2, 0)
	q.Put("k", key)
	q.Put("p1", delta)
	q.Put("p2", delta)

	if got := drain(t, q); !equal(got, []string{"k", "p1"}) {
		t.Fatalf("got %v", got)
	}
}

func TestOldBacklogIsDroppedAfterMaxDelay(t *testing.T) {
	q, c := newQueue(100, 100*time.Millisecond)
	q.Put("k", key)
	c.advance(50 * time.Millisecond)
	q.Put("p1", delta)
	c.advance(51 * time.Millisecond)

	if q.Put("p2", delta) {
		t.Fatal("a viewer whose oldest frame waited over MaxDelay is behind")
	}
}

func TestKeyframeFlushesAStaleBacklog(t *testing.T) {
	q, c := newQueue(100, 100*time.Millisecond)
	q.Put("k", key)
	q.Put("p1", delta)
	c.advance(200 * time.Millisecond)

	if !q.Put("k2", key) {
		t.Fatal("a keyframe is always queued")
	}

	if got := drain(t, q); !equal(got, []string{"k2"}) {
		t.Fatalf("the stale frames should have been discarded, got %v", got)
	}
	if q.Dropped() != 2 || q.Resyncs() != 1 {
		t.Fatalf("dropped=%d resyncs=%d", q.Dropped(), q.Resyncs())
	}
}

func TestANewQueueStartsAtAKeyframe(t *testing.T) {
	// A viewer joining a running stream cannot decode the delta frames before
	// the next keyframe, so they are not sent.
	q, _ := newQueue(8, time.Second)

	if q.Put("p0", delta) {
		t.Fatal("a delta frame before the first keyframe must be refused")
	}
	q.Put("k", key)
	q.Put("p1", delta)

	if got := drain(t, q); !equal(got, []string{"k", "p1"}) {
		t.Fatalf("got %v", got)
	}
}

func TestKeyframeKeepsAFreshBacklog(t *testing.T) {
	// A viewer a frame or two behind finishes the old GOP first.
	q, c := newQueue(100, 100*time.Millisecond)
	q.Put("k", key)
	q.Put("p1", delta)
	c.advance(20 * time.Millisecond)
	q.Put("k2", key)

	if got := drain(t, q); !equal(got, []string{"k", "p1", "k2"}) {
		t.Fatalf("got %v", got)
	}
}

func TestKeyframeIntoAFullQueueReplacesTheBacklog(t *testing.T) {
	q, _ := newQueue(2, 0)
	q.Put("k", key)
	q.Put("p1", delta)

	if !q.Put("k2", key) {
		t.Fatal("a keyframe is always queued")
	}
	if got := drain(t, q); !equal(got, []string{"k2"}) {
		t.Fatalf("got %v", got)
	}
}

func TestRingWrapsAround(t *testing.T) {
	q, _ := newQueue(3, 0)
	q.Put("k", key)

	for i := 0; i < 10; i++ {
		f := string(rune('a' + i))
		if !q.Put(f, delta) {
			t.Fatalf("frame %d refused", i)
		}
		q.Take()
	}

	if got := drain(t, q); !equal(got, []string{"j"}) {
		t.Fatalf("got %v", got)
	}
}

func TestCloseReleasesABlockedTake(t *testing.T) {
	q, _ := newQueue(4, 0)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if _, ok := q.Take(); ok {
			t.Error("Take after Close should report the queue closed")
		}
	}()

	time.Sleep(10 * time.Millisecond)
	q.Close()
	wg.Wait()
}

func TestCloseDiscardsPendingAndRefusesMore(t *testing.T) {
	q, _ := newQueue(4, 0)
	q.Put("k", key)
	q.Close()
	q.Close()

	if _, ok := q.Take(); ok {
		t.Fatal("a closed queue hands out nothing")
	}
	if q.Put("k2", key) {
		t.Fatal("a closed queue refuses frames")
	}
}

func TestTakeWaitsForPut(t *testing.T) {
	q, _ := newQueue(4, 0)
	got := make(chan string)

	go func() {
		v, _ := q.Take()
		got <- v
	}()

	time.Sleep(10 * time.Millisecond)
	q.Put("k", key)

	select {
	case v := <-got:
		if v != "k" {
			t.Fatalf("got %q", v)
		}
	case <-time.After(time.Second):
		t.Fatal("Take did not wake on Put")
	}
}
