package stream

import (
	"sync"
	"time"

	"NanoKVM-Server/common"
)

// H264Frame is one encoded frame as the capture pipeline returned it.
//
// Data is shared by every subscriber and nobody may write to it. Result is
// libkvm's own return, passed through so each delivery path can report it
// under its own capture mode. A frame with a negative Result carries no data
// and exists only so that report still happens.
type H264Frame struct {
	Data      []byte
	Result    int
	KeyFrame  bool
	Timestamp int64
	Duration  time.Duration
}

// H264Source runs one capture loop for both H.264 delivery paths.
//
// There is one encoder behind ReadH264 and it hands each caller whatever frame
// is ready when it asks. Two loops calling it therefore do not each get the
// stream: they divide it, and a decoder given every second frame of a GOP has
// nothing to do with the half it gets. Direct mode and WebRTC mode each ran
// their own ticker and their own ReadH264, so a board with one viewer on each
// served two broken streams instead of one good one, and paid for the capture
// twice to do it.
//
// The loop starts with the first subscriber and stops after the last one
// leaves.
type H264Source struct {
	mutex       sync.Mutex
	subscribers map[*H264Subscription]struct{}
	running     bool
}

// H264Subscription is one delivery path's view of the capture loop.
//
// demand is asked every tick. Direct mode acknowledges frames and stops asking
// for them while a viewer is behind, so a subscriber that wants nothing must be
// able to say so: with every subscriber quiet the source reads nothing at all,
// which is what kept an idle board off the encoder before this existed.
//
// It runs on the capture goroutine, and both paths share that goroutine now.
// So demand must not block, must not write to a socket, and must not take a
// lock that anything holds across a write. One path that broke this rule would
// stall the other path's viewers as well as its own, which is the whole failure
// the frame slots below exist to prevent, arriving by a different door.
//
// Both implementations hold to it today. Each reads its client list from an
// atomic snapshot, and direct mode's hasCaptureDemand takes a per-client mutex
// that is only ever held for bookkeeping: the websocket writer pops a frame
// under that mutex and releases it before it writes.
type H264Subscription struct {
	source *H264Source
	slot   *FrameSlot[H264Frame]
	demand func() bool
	once   sync.Once
}

func newH264Source() *H264Source {
	return &H264Source{subscribers: make(map[*H264Subscription]struct{})}
}

var defaultH264Source = newH264Source()

// readH264 is a variable so a test can drive the loop without capture
// hardware. Off-device the stub answers -1 to everything, which exercises the
// status path and never the delivery one.
var readH264 = func(width uint16, height uint16, bitRate uint16) ([]byte, int) {
	return common.GetKvmVision().ReadH264(width, height, bitRate)
}

// SubscribeH264 joins the shared capture loop, starting it if it is not
// running. demand reports whether this path has anything to send a frame to;
// pass a function that returns true to take every frame.
func SubscribeH264(demand func() bool) *H264Subscription {
	return defaultH264Source.subscribe(demand)
}

func (s *H264Source) subscribe(demand func() bool) *H264Subscription {
	if demand == nil {
		demand = func() bool { return true }
	}

	subscription := &H264Subscription{
		source: s,
		slot:   NewFrameSlot[H264Frame](),
		demand: demand,
	}

	s.mutex.Lock()
	s.subscribers[subscription] = struct{}{}
	start := !s.running
	s.running = true
	s.mutex.Unlock()

	if start {
		go s.run()
	}

	return subscription
}

// Frames is the pending frame, for a caller that also has to notice its own
// viewers leaving. It closes with the subscription.
func (p *H264Subscription) Frames() <-chan H264Frame {
	return p.slot.Channel()
}

// Dropped counts the frames this path was not ready to take.
func (p *H264Subscription) Dropped() uint64 {
	return p.slot.Dropped()
}

// Close leaves the capture loop. It is safe to call more than once.
func (p *H264Subscription) Close() {
	p.once.Do(func() {
		p.source.remove(p)
		p.slot.Close()
	})
}

func (s *H264Source) remove(subscription *H264Subscription) {
	s.mutex.Lock()
	delete(s.subscribers, subscription)
	s.mutex.Unlock()
}

// snapshot lists the current subscribers and reports whether any remain.
func (s *H264Source) snapshot() []*H264Subscription {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	subscribers := make([]*H264Subscription, 0, len(s.subscribers))
	for subscription := range s.subscribers {
		subscribers = append(subscribers, subscription)
	}

	return subscribers
}

// stopIfIdle marks the loop stopped when the last subscriber has gone. It runs
// under the same mutex subscribe uses, so a subscriber arriving at this moment
// either finds running still true and is served by this loop, or finds it false
// and starts the next one.
func (s *H264Source) stopIfIdle() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.subscribers) > 0 {
		return false
	}

	s.running = false

	return true
}

func (s *H264Source) run() {
	screen := common.GetScreen()
	common.CheckScreen()
	values := screen.Snapshot()
	fps := values.FPS
	duration := time.Second / time.Duration(fps)

	startTime := time.Now()

	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for range ticker.C {
		subscribers := s.snapshot()
		if len(subscribers) == 0 {
			if s.stopIfIdle() {
				return
			}

			continue
		}

		values = screen.Snapshot()
		if values.FPS != fps && values.FPS != 0 {
			fps = values.FPS
			duration = time.Second / time.Duration(fps)
			ticker.Reset(duration)
		}

		// Ask before reading. A read costs the encoder whether or not anyone
		// takes the frame.
		wanted := subscribers[:0:0]
		for _, subscription := range subscribers {
			if subscription.demand() {
				wanted = append(wanted, subscription)
			}
		}

		if len(wanted) == 0 {
			continue
		}

		data, result := readH264(values.Width, values.Height, values.BitRate)

		frame := H264Frame{
			Data:      data,
			Result:    result,
			KeyFrame:  result == 3,
			Timestamp: time.Since(startTime).Microseconds(),
			Duration:  duration,
		}

		// A negative result still goes out, because each path reports capture
		// status under its own mode and a status nobody reports is a stream
		// that fails silently.
		for _, subscription := range wanted {
			subscription.slot.TryPut(frame)
		}

		if result < 0 || len(data) == 0 {
			continue
		}

		// Counted once per captured frame. Both paths used to count, so a
		// board with a viewer on each reported twice the frame rate it had.
		GetFrameRateCounter().Update()
	}
}
