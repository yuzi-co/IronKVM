package ws

import (
	"sync"
	"time"
)

// A key held down in the browser stays down on the host until a key-up report
// arrives. When the connection stalls, nothing arrives: a tab that is frozen, a
// laptop lid that closes or a network that drops without a FIN leaves the
// socket open on this side, and the read loop only notices at the heartbeat
// timeout, 90 seconds later. Until then the host auto-repeats the key.
//
// So the browser keeps a held input alive. While any key or mouse button is
// down it sends a keepalive message every holdKeepaliveInterval, and the
// holdWatch here releases whatever is held when neither a keepalive nor an
// input report has arrived for holdReleaseTimeout. Six keepalives fit in the
// timeout, so one late packet does not cut a held key short.
//
// The timeout is below the host's auto-repeat delay. A Linux console starts
// repeating after 250 ms, X after 660 ms and Windows after about 500 ms, so a
// key released at 300 ms types at most one extra character on the fastest of
// them, rather than ninety seconds of them.
//
// A client opts in by sending a keepalive. The web client sends one when the
// socket opens. A client that never does is one written before the keepalive
// existed, and releasing its held keys after 300 ms would break every long key
// press it makes, so for such a client the watch never fires.
const (
	holdKeepaliveInterval = 50 * time.Millisecond
	holdReleaseTimeout    = 300 * time.Millisecond
)

// holdKind names one of the inputs the watch tracks. Each has its own deadline:
// a key pressed after a mouse button is held starts its own timer, and the
// release that fires names only what is still held.
type holdKind uint8

const (
	holdKeyboard holdKind = iota
	holdRelativeMouse
	holdAbsoluteMouse
	holdKinds
)

// holdTimer is the part of *time.Timer the watch uses, so the tests can drive
// expiry by hand.
type holdTimer interface {
	Stop() bool
}

type holdWatch struct {
	mutex    sync.Mutex
	timeout  time.Duration
	newTimer func(time.Duration, func()) holdTimer
	release  func(holdKind)

	enabled bool
	closed  bool
	held    [holdKinds]bool
	timers  [holdKinds]holdTimer
	// generation tells a timer that fired from one that was re-armed after it
	// had already started to run, so a stale expiry releases nothing.
	generation [holdKinds]uint64
}

func newHoldWatch(timeout time.Duration, release func(holdKind)) *holdWatch {
	return &holdWatch{
		timeout: timeout,
		release: release,
		newTimer: func(d time.Duration, f func()) holdTimer {
			return time.AfterFunc(d, f)
		},
	}
}

// Keepalive records a keepalive message. The first one turns the watch on, and
// each re-arms the deadline of everything held.
func (w *holdWatch) Keepalive() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.closed {
		return
	}
	w.enabled = true
	for kind := range holdKinds {
		if w.held[kind] {
			w.armLocked(kind)
		}
	}
}

// Input records a report of one kind. A report that holds something re-arms
// that kind's deadline; one that holds nothing disarms it.
func (w *holdWatch) Input(kind holdKind, held bool) {
	if kind >= holdKinds {
		return
	}
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.closed {
		return
	}
	w.held[kind] = held
	if held && w.enabled {
		w.armLocked(kind)
		return
	}
	w.disarmLocked(kind)
}

// Close stops every timer. The client's own close releases what is held, so
// the watch has nothing left to do.
func (w *holdWatch) Close() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.closed = true
	for kind := range holdKinds {
		w.disarmLocked(kind)
		w.held[kind] = false
	}
}

func (w *holdWatch) armLocked(kind holdKind) {
	w.disarmLocked(kind)
	generation := w.generation[kind]
	w.timers[kind] = w.newTimer(w.timeout, func() { w.expire(kind, generation) })
}

func (w *holdWatch) disarmLocked(kind holdKind) {
	if w.timers[kind] != nil {
		w.timers[kind].Stop()
		w.timers[kind] = nil
	}
	w.generation[kind]++
}

func (w *holdWatch) expire(kind holdKind, generation uint64) {
	w.mutex.Lock()
	if w.closed || !w.held[kind] || w.generation[kind] != generation {
		w.mutex.Unlock()
		return
	}
	w.held[kind] = false
	w.timers[kind] = nil
	w.generation[kind]++
	w.mutex.Unlock()

	// Outside the lock: the release queues a report, and the read loop may
	// be waiting on this lock while it holds the queue.
	w.release(kind)
}
