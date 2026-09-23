package audio

import (
	"sync"
	"testing"
	"time"
)

// fakeCapture stands in for a Stream: frames are pushed by the test, and End
// closes the channel the way a capture that gave up does.
type fakeCapture struct {
	frames  chan []byte
	mutex   sync.Mutex
	started int
	stopped int
	endOnce sync.Once
}

func newFakeCapture() *fakeCapture { return &fakeCapture{frames: make(chan []byte, 16)} }

func (f *fakeCapture) Start() { f.mutex.Lock(); f.started++; f.mutex.Unlock() }
func (f *fakeCapture) Stop()  { f.mutex.Lock(); f.stopped++; f.mutex.Unlock(); f.End() }
func (f *fakeCapture) End()   { f.endOnce.Do(func() { close(f.frames) }) }
func (f *fakeCapture) Frames() <-chan []byte {
	return f.frames
}
func (f *fakeCapture) counts() (int, int) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.started, f.stopped
}

type fakeFactory struct {
	mutex    sync.Mutex
	captures []*fakeCapture
}

func (f *fakeFactory) make() Capture {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	c := newFakeCapture()
	f.captures = append(f.captures, c)
	return c
}

func (f *fakeFactory) all() []*fakeCapture {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return append([]*fakeCapture(nil), f.captures...)
}

func newTestHub(available bool) (*Hub, *fakeFactory) {
	factory := &fakeFactory{}
	return NewHubWith(factory.make, func() bool { return available }), factory
}

func receive(t *testing.T, s *Subscription) Frame {
	t.Helper()
	select {
	case f, ok := <-s.Frames():
		if !ok {
			t.Fatal("the subscription closed while a frame was expected")
		}
		return f
	case <-time.After(time.Second):
		t.Fatal("no frame arrived")
	}
	return Frame{}
}

func closedWithin(s *Subscription) bool {
	deadline := time.After(time.Second)
	for {
		select {
		case _, ok := <-s.Frames():
			if !ok {
				return true
			}
		case <-deadline:
			return false
		}
	}
}

func TestHubStartsOneCaptureForEveryListener(t *testing.T) {
	// The capture device opens exclusively. A second arecord for a second
	// viewer fails with "Device or resource busy", so WebRTC and direct must
	// share one capture rather than each start their own.
	hub, factory := newTestHub(true)
	a := hub.Subscribe()
	b := hub.Subscribe()
	defer a.Close()
	defer b.Close()

	captures := factory.all()
	if len(captures) != 1 {
		t.Fatalf("started %d captures for two listeners, want 1", len(captures))
	}
	if started, _ := captures[0].counts(); started != 1 {
		t.Fatalf("the capture was started %d times, want 1", started)
	}
}

func TestHubHandsEveryListenerEveryFrameInOrder(t *testing.T) {
	hub, factory := newTestHub(true)
	a := hub.Subscribe()
	b := hub.Subscribe()
	defer a.Close()
	defer b.Close()

	c := factory.all()[0]
	c.frames <- []byte{1}
	c.frames <- []byte{2}

	for _, s := range []*Subscription{a, b} {
		first, second := receive(t, s), receive(t, s)
		if first.Seq != 0 || second.Seq != 1 || first.Data[0] != 1 || second.Data[0] != 2 {
			t.Fatalf("got %+v then %+v, want sequence 0 and 1 in order", first, second)
		}
	}
}

func TestHubStopsCaptureWhenTheLastListenerLeaves(t *testing.T) {
	// While the host plays nothing arecord blocks in a read, so nothing on the
	// read path notices that nobody listens. The last Close has to kill it.
	hub, factory := newTestHub(true)
	a := hub.Subscribe()
	b := hub.Subscribe()
	c := factory.all()[0]

	a.Close()
	if _, stopped := c.counts(); stopped != 0 {
		t.Fatal("capture stopped while a listener remained")
	}

	b.Close()
	if _, stopped := c.counts(); stopped != 1 {
		t.Fatal("capture kept running after the last listener left")
	}

	b.Close()
	if _, stopped := c.counts(); stopped != 1 {
		t.Fatal("a second Close of the same subscription stopped capture again")
	}
}

func TestHubClosesListenersWhenCaptureEndsByItself(t *testing.T) {
	// An encoder that cannot be built ends the stream at once. A listener has
	// to learn that, or it waits for frames that never come, and the next
	// listener has to get a fresh capture rather than the dead one.
	hub, factory := newTestHub(true)
	a := hub.Subscribe()
	factory.all()[0].End()

	if !closedWithin(a) {
		t.Fatal("the listener was not told that capture ended")
	}

	b := hub.Subscribe()
	defer b.Close()
	if n := len(factory.all()); n != 2 {
		t.Fatalf("started %d captures after the first ended, want 2", n)
	}

	factory.all()[1].frames <- []byte{9}
	if f := receive(t, b); f.Seq != 0 {
		t.Fatalf("a fresh capture numbered its first frame %d, want 0", f.Seq)
	}
	a.Close()
}

func TestHubGivesNoSubscriptionWithoutAudio(t *testing.T) {
	hub, factory := newTestHub(false)
	if s := hub.Subscribe(); s != nil {
		t.Fatal("subscribed on a board with no capture card")
	}
	if n := len(factory.all()); n != 0 {
		t.Fatalf("started %d captures on a board with no capture card", n)
	}
}

func TestHubSlowListenerLosesFramesWithoutStallingOthers(t *testing.T) {
	hub, factory := newTestHub(true)
	slow := hub.Subscribe()
	fast := hub.Subscribe()
	defer slow.Close()
	defer fast.Close()

	c := factory.all()[0]
	for i := 0; i < subscriptionBuffer+6; i++ {
		c.frames <- []byte{byte(i)}
		if f := receive(t, fast); f.Seq != uint64(i) {
			t.Fatalf("the fast listener got sequence %d, want %d", f.Seq, i)
		}
	}
}

func TestHubStopAllKillsCaptureAndClosesListeners(t *testing.T) {
	hub, factory := newTestHub(true)
	a := hub.Subscribe()
	hub.StopAll()

	if _, stopped := factory.all()[0].counts(); stopped != 1 {
		t.Fatal("StopAll left capture running")
	}
	if !closedWithin(a) {
		t.Fatal("StopAll left a listener open")
	}
	a.Close()
}
