// Package framequeue holds the encoded video frames waiting for one viewer.
//
// It has no cgo dependency, so its tests run on the build host rather than on
// the board.
package framequeue

import (
	"sync"
	"time"
)

// Defaults for a WebRTC viewer.
//
// A 1080p keyframe is about a hundred RTP packets, and each one costs the
// writer close to a millisecond on the board (SRTP in software, the interceptor
// chain, the socket), so a keyframe can take 90 to 150 ms to go out. Frames
// arrive every 17 or 33 ms meanwhile. The queue has to hold everything that
// arrives while a keyframe is being written, or the frames after it are lost
// and the viewer waits for the next keyframe.
//
// DefaultMaxDelay is how far behind a viewer may fall before frames are given
// up: well over the slowest keyframe, short enough that a viewer on a link that
// cannot keep up is resynchronised rather than shown a picture seconds old.
// DefaultMaxFrames is only a backstop on the count, a second of frames at 60 fps.
const (
	DefaultMaxFrames = 60
	DefaultMaxDelay  = 500 * time.Millisecond
)

type entry[T any] struct {
	value    T
	key      bool
	enqueued time.Time
}

// Queue is a bounded FIFO of frames for a single viewer, fed by the capture
// loop and drained by that viewer's writer goroutine.
//
// Put never blocks, so one slow viewer cannot hold up capture or the others.
// Frames are coded against the one before, so the queue never leaves a hole in
// the middle of a GOP: when it has to drop, it drops every frame up to the next
// keyframe.
//
//   - A delta frame is queued unless the viewer is too far behind: the queue
//     holds MaxFrames frames, or the oldest one has waited longer than MaxDelay.
//     Then it and every delta frame after it are dropped until a keyframe.
//   - A new queue waits for a keyframe: a viewer that joins a running stream
//     cannot decode the delta frames before one.
//   - A keyframe is always queued. If the viewer is too far behind when it
//     arrives, the frames still waiting are discarded first: the keyframe does
//     not need them, and sending them would only add to the delay.
//
// So a keyframe that is slow to write delays the frames after it instead of
// losing them, and only sustained overload costs frames, with the delay held
// to about MaxDelay plus the time to write what is queued.
type Queue[T any] struct {
	mutex  sync.Mutex
	ready  *sync.Cond
	frames []entry[T]
	head   int
	count  int

	maxDelay      time.Duration
	waitingForKey bool
	closed        bool
	dropped       uint64
	resyncs       uint64

	now func() time.Time
}

// New returns a queue holding at most maxFrames frames, giving up on frames
// once the oldest pending one has waited maxDelay. A maxDelay of zero or less
// bounds the queue by count alone.
func New[T any](maxFrames int, maxDelay time.Duration) *Queue[T] {
	if maxFrames < 1 {
		maxFrames = 1
	}

	q := &Queue[T]{
		frames:        make([]entry[T], maxFrames),
		maxDelay:      maxDelay,
		waitingForKey: true,
		now:           time.Now,
	}
	q.ready = sync.NewCond(&q.mutex)

	return q
}

// behindLocked reports whether the viewer has fallen too far behind to be
// given another frame in sequence.
func (q *Queue[T]) behindLocked(now time.Time) bool {
	if q.count == len(q.frames) {
		return true
	}

	if q.count == 0 || q.maxDelay <= 0 {
		return false
	}

	return now.Sub(q.frames[q.head].enqueued) > q.maxDelay
}

// Put offers a frame and never blocks. It reports whether the frame was queued.
func (q *Queue[T]) Put(value T, key bool) bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.closed {
		return false
	}

	now := q.now()

	if key {
		if q.behindLocked(now) {
			q.dropped += uint64(q.count)
			q.resyncs++
			q.clearLocked()
		}

		q.waitingForKey = false
	} else {
		if q.waitingForKey {
			q.dropped++
			return false
		}

		if q.behindLocked(now) {
			q.dropped++
			q.resyncs++
			q.waitingForKey = true

			return false
		}
	}

	tail := (q.head + q.count) % len(q.frames)
	q.frames[tail] = entry[T]{value: value, key: key, enqueued: now}
	q.count++
	q.ready.Signal()

	return true
}

// Take blocks until a frame is queued and returns it, or returns false once
// the queue is closed.
func (q *Queue[T]) Take() (T, bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for q.count == 0 && !q.closed {
		q.ready.Wait()
	}

	var zero T
	if q.closed {
		return zero, false
	}

	e := q.frames[q.head]
	q.frames[q.head] = entry[T]{}
	q.head = (q.head + 1) % len(q.frames)
	q.count--

	return e.value, true
}

// Close releases the consumer. Frames still queued are discarded: the viewer
// is going away, and its writer should not spend time on them first.
func (q *Queue[T]) Close() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.closed {
		return
	}

	q.closed = true
	q.clearLocked()
	q.ready.Broadcast()
}

func (q *Queue[T]) clearLocked() {
	for i := range q.frames {
		q.frames[i] = entry[T]{}
	}

	q.head = 0
	q.count = 0
}

// Len is the number of frames waiting.
func (q *Queue[T]) Len() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return q.count
}

// Dropped is the number of frames this viewer was too slow to receive.
func (q *Queue[T]) Dropped() uint64 {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return q.dropped
}

// Resyncs is the number of times the viewer fell too far behind and the stream
// was cut back to a keyframe.
func (q *Queue[T]) Resyncs() uint64 {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return q.resyncs
}
