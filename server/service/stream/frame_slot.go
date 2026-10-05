package stream

import (
	"sync"
	"sync/atomic"
)

// FrameSlot holds at most one pending frame for a single client, or a few
// when made with NewFrameSlotDepth.
//
// The capture loop must never block on a client: one viewer on a slow link
// would otherwise stall capture for every other viewer. Producers hand a frame
// to the slot and move on, and each client's writer goroutine takes frames at
// whatever rate it can manage. When a client falls behind, frames are dropped
// rather than buffered, because a queue of stale frames only adds latency.
//
// Producers choose the policy that suits the codec:
//   - Replace keeps the newest frame, for formats where every frame stands
//     alone (MJPEG).
//   - TryPut refuses while the slot is full, so the producer learns the
//     client is behind and can decide what to send next (H.264, where a gap
//     has to be repaired with a keyframe).
type FrameSlot[T any] struct {
	mutex   sync.Mutex
	frames  chan T
	closed  bool
	dropped atomic.Uint64
}

func NewFrameSlot[T any]() *FrameSlot[T] {
	return NewFrameSlotDepth[T](1)
}

// NewFrameSlotDepth makes a slot that holds up to depth frames, oldest first.
// TryPut refuses only when all of them are pending, and Replace discards the
// oldest. It is for a consumer that is close by and normally keeps up, but
// whose goroutine can be late by a frame or two: there a refusal would cost a
// whole GOP for a delay the next tick makes up.
func NewFrameSlotDepth[T any](depth int) *FrameSlot[T] {
	if depth < 1 {
		depth = 1
	}

	return &FrameSlot[T]{
		frames: make(chan T, depth),
	}
}

// TryPut stores a frame only if the slot has room. It reports whether the
// frame was accepted, and never blocks.
func (s *FrameSlot[T]) TryPut(frame T) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.closed {
		return false
	}

	select {
	case s.frames <- frame:
		return true
	default:
		s.dropped.Add(1)
		return false
	}
}

// Replace stores a frame, discarding the oldest pending one when the slot is
// full. It reports whether the slot accepted it, and never blocks.
func (s *FrameSlot[T]) Replace(frame T) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.closed {
		return false
	}

	if len(s.frames) == cap(s.frames) {
		select {
		case <-s.frames:
			s.dropped.Add(1)
		default:
		}
	}

	// The slot has room and no other producer can run while the mutex is
	// held, so this cannot block.
	s.frames <- frame

	return true
}

// Pending reports whether a frame is still waiting to be taken.
func (s *FrameSlot[T]) Pending() bool {
	return len(s.frames) > 0
}

// Channel exposes the pending frame directly, for a consumer that has to wait
// on something else at the same time. It closes with the slot.
func (s *FrameSlot[T]) Channel() <-chan T {
	return s.frames
}

// Take blocks until a frame is available, returning false once the slot is
// closed and drained.
func (s *FrameSlot[T]) Take() (T, bool) {
	frame, ok := <-s.frames

	return frame, ok
}

// Dropped is the number of frames this client was too slow to receive.
func (s *FrameSlot[T]) Dropped() uint64 {
	return s.dropped.Load()
}

// Close releases the consumer. A frame already pending is still delivered.
func (s *FrameSlot[T]) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.closed {
		return
	}

	s.closed = true
	close(s.frames)
}
