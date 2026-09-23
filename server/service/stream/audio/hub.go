package audio

import "sync"

// subscriptionBuffer is how many frames a listener may fall behind before it
// starts losing them. Eight frames is 160 ms, twice the slack the capture
// stream itself keeps.
const subscriptionBuffer = 8

// Frame is one 20 ms Opus packet and its place in the capture that made it.
//
// Seq counts from 0 for each capture. A listener that sees a gap knows frames
// were lost and can keep its clock by filling the gap with silence. A listener
// that sees Seq go back knows a new capture started.
type Frame struct {
	Seq  uint64
	Data []byte
}

// Capture is what the hub runs. *Stream is the production one.
type Capture interface {
	Start()
	Stop()
	Frames() <-chan []byte
}

// Hub shares one capture between every listener.
//
// The capture device opens exclusively: a second arecord on hw:UAC1Gadget,0
// fails with "Device or resource busy". While WebRTC was the only path with
// audio it owned its stream outright. With H.264 direct carrying audio too, a
// WebRTC viewer and a direct viewer at the same time would each start one, and
// whichever came second would spend its restart budget failing. So both
// subscribe here: the first listener starts the capture and the last one stops
// it.
type Hub struct {
	mutex      sync.Mutex
	subs       map[*Subscription]struct{}
	current    Capture
	newCapture func() Capture
	available  func() bool
}

// Shared is the hub every video path subscribes to.
var Shared = NewHubWith(func() Capture { return NewStream() }, Available)

// NewHubWith builds a hub around a capture factory and an availability check,
// so a test can supply both.
func NewHubWith(newCapture func() Capture, available func() bool) *Hub {
	return &Hub{
		subs:       make(map[*Subscription]struct{}),
		newCapture: newCapture,
		available:  available,
	}
}

// Subscription is one listener's view of the shared capture.
type Subscription struct {
	hub       *Hub
	frames    chan Frame
	closeOnce sync.Once
}

// Frames delivers the capture's frames in order. It closes when this
// subscription closes, when the capture ends by itself, or on StopAll.
func (s *Subscription) Frames() <-chan Frame {
	return s.frames
}

// Subscribe adds a listener, starting capture if it is the first. It returns
// nil when this board cannot capture audio right now. The answer is not cached,
// because the settings page rebuilds the USB gadget while the server runs.
func (h *Hub) Subscribe() *Subscription {
	if !h.available() {
		return nil
	}

	sub := &Subscription{hub: h, frames: make(chan Frame, subscriptionBuffer)}

	h.mutex.Lock()
	h.subs[sub] = struct{}{}
	var start Capture
	if h.current == nil {
		start = h.newCapture()
		h.current = start
	}
	h.mutex.Unlock()

	if start != nil {
		start.Start()
		go h.fanOut(start)
	}

	return sub
}

// fanOut hands each frame to every listener. It never blocks on one: a
// listener that is behind loses a frame, and the sequence number tells it so.
func (h *Hub) fanOut(capture Capture) {
	var seq uint64

	for data := range capture.Frames() {
		frame := Frame{Seq: seq, Data: data}
		seq++

		h.mutex.Lock()
		if h.current == capture {
			for sub := range h.subs {
				select {
				case sub.frames <- frame:
				default:
				}
			}
		}
		h.mutex.Unlock()
	}

	// The capture ended. If it is still the current one, nobody stopped it:
	// the encoder could not be built or the child could not run. Every
	// listener is told by the close, and the next Subscribe starts afresh.
	h.mutex.Lock()
	if h.current == capture {
		h.current = nil
		for sub := range h.subs {
			delete(h.subs, sub)
			sub.closeFrames()
		}
	}
	h.mutex.Unlock()
}

// Close removes the listener, and stops capture if it was the last one. Stop
// runs outside the lock, because it waits on the capture goroutine.
func (s *Subscription) Close() {
	h := s.hub

	h.mutex.Lock()
	_, present := h.subs[s]
	delete(h.subs, s)
	var stop Capture
	if present && len(h.subs) == 0 && h.current != nil {
		stop = h.current
		h.current = nil
	}
	h.mutex.Unlock()

	s.closeFrames()
	if stop != nil {
		stop.Stop()
	}
}

// StopAll ends capture whatever the listener count. main.go calls it through
// webrtc.StopAudioCapture when the server stops, so no arecord outlives it.
func (h *Hub) StopAll() {
	h.mutex.Lock()
	stop := h.current
	h.current = nil
	for sub := range h.subs {
		delete(h.subs, sub)
		sub.closeFrames()
	}
	h.mutex.Unlock()

	if stop != nil {
		stop.Stop()
	}
}

// closeFrames runs under the hub's lock or after the subscription left the
// map, so fanOut can never send on a closed channel.
func (s *Subscription) closeFrames() {
	s.closeOnce.Do(func() {
		close(s.frames)
	})
}
