package stream

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

// readPacer decides when the capture loop reads the next frame.
//
// The loop used to read on a ticker at the frame rate. A read takes the newest
// captured frame, so the frame waited for the tick: anything from nothing to
// a frame time, a phase that drifted with the source and changed from one
// stream to the next (about 8 ms between two runs of the same setting at 30
// fps, ironkvm-dist trial 38), plus up to KVMV_PICKUP_WAIT_MS where the library
// waits for a frame about to complete.
//
// When the library says when each frame was captured (libkvm-v4l2's
// kvmv_video_times), the next read is timed from that instead. The frame to
// take is the first source frame at least a frame period, less half a source
// frame, after the last one taken; the read starts readLead before that frame
// completes. A read blocks until a frame it has not taken completes, and the
// library waits for a frame about to complete rather than take one over half a
// source frame old, so the read takes that frame as soon as it lands. The rate
// is then never above the one asked for: 30 fps from 60 Hz is every second
// frame, from 50 Hz every second frame too (25 fps), where the ticker mixed
// gaps of one and two source frames. A read that took longer than a frame
// time (H.264 through the scaler at 60 fps) is followed by the next at once,
// which takes the frame that completed meanwhile, as the ticker's late tick
// did.
//
// Without capture times (Sipeed's library on slot A), and with
// KVM_READ_PACING=tick, the ticker stays.
type readPacer struct {
	byCapture bool
	period    time.Duration
	ticker    *time.Ticker
	timer     *time.Timer

	last      time.Time     // capture end of the last frame read, wall clock
	frameTime time.Duration // the source's frame time then, 0 when unknown
}

// readLead is how long before the frame is due the paced read starts. It covers
// the timer's lateness and the read's way into the library, so the read is
// already waiting when the capture completes. Too short and the frame waits for
// the read; too long and the read waits for the frame, which costs nothing but
// the wait.
const readLead = 3 * time.Millisecond

const readPacingEnv = "KVM_READ_PACING"

func newReadPacer(period time.Duration) *readPacer {
	return &readPacer{
		byCapture: os.Getenv(readPacingEnv) != "tick",
		period:    period,
		ticker:    time.NewTicker(period),
	}
}

func (p *readPacer) stop() {
	p.ticker.Stop()
	if p.timer != nil {
		p.timer.Stop()
	}
}

// setPeriod follows a frame rate change.
func (p *readPacer) setPeriod(period time.Duration) {
	p.period = period
	p.ticker.Reset(period)
}

// paced says whether reads follow the capture times now.
func (p *readPacer) paced() bool {
	return p.byCapture && !p.last.IsZero()
}

// read records what a read gave: the capture end of its frame and the source's
// frame time, or the zero time when it gave none or the library cannot say.
func (p *readPacer) read(captureEnd time.Time, frameTime time.Duration) {
	if captureEnd.IsZero() || !captureEnd.After(p.last) {
		// No new frame (no signal, an error, or the library's rate guard
		// left the picture out): the next read is due a frame period
		// later, as on the ticker.
		if !p.last.IsZero() && p.byCapture {
			p.last = p.last.Add(p.period)
		}

		return
	}

	if p.byCapture && p.last.IsZero() {
		log.Infof("capture: reads follow the frames' capture times (%s=tick for the ticker)", readPacingEnv)
	}

	p.last = captureEnd
	p.frameTime = frameTime
}

// next is when the frame the next read should take completes: the first
// source frame that ends at least a period less half a source frame after the
// last one.
func (p *readPacer) next() time.Time {
	ft := p.frameTime
	if ft <= 0 || ft > p.period {
		return p.last.Add(p.period)
	}

	frames := (p.period - ft/2 + ft - 1) / ft

	return p.last.Add(frames * ft)
}

// delay is how long a paced read waits from now: until readLead before the
// next frame is due, or not at all when that is past (the frame the read is
// for has completed, or is about to).
func (p *readPacer) delay(now time.Time) time.Duration {
	at := p.next().Add(-readLead)
	if !at.After(now) {
		return 0
	}

	return at.Sub(now)
}

// wait blocks until the next read is due.
func (p *readPacer) wait() {
	if !p.paced() {
		<-p.ticker.C

		return
	}

	d := p.delay(time.Now())
	if d <= 0 {
		return
	}

	if p.timer == nil {
		p.timer = time.NewTimer(d)
	} else {
		p.timer.Reset(d)
	}
	<-p.timer.C
}
