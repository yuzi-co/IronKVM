package stream

import (
	"math"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Adaptive keyframes (ironkvm-dist#72, trial 64).
//
// A keyframe costs the most of any frame: 85 to 120 kB at 1080p, about a
// hundred RTP packets, 80 to 130 ms of the board's send path on slot B, and
// every frame behind it waits that long (trials 56, 61, 63). A fixed GOP pays
// that every half second at 60 fps whether or not anyone needs it. A viewer
// only needs one when it starts, after it fell behind, or after it lost
// packets that retransmission could not repair.
//
// So the encoder is asked for a keyframe when a viewer needs one, and the
// periodic ones are spaced by time and by the network:
//
//   - Clean: one every keyframeRareInterval (2 s), or the GOP setting's
//     interval if that is longer. The encoder's own GOP is set to this many
//     frames, so it makes these itself and the server forces none.
//   - On demand, at once: a viewer that waits for a keyframe (it joined, or its
//     queue gave up frames) and a WebRTC viewer that reports a lost picture
//     (PLI, FIR, or a NACK the retransmission history no longer holds). Forced
//     keyframes are kept keyframeMinGap apart, so a viewer that keeps
//     reporting cannot turn the stream into keyframes.
//   - Lossy: once one viewer has reported loss lossEventsFrequent times within
//     lossWindow, the server forces a keyframe whenever the GOP setting's
//     interval (GOP / fps, 0.5 s at GOP 30 and 60 fps, as every stream had
//     before) passes without one. It goes back to clean lossHold after that
//     viewer's last loss, or when it leaves.
//
// The encoder serves every viewer, so the most demanding one decides. H.264
// Direct and MJPEG run over TCP and lose nothing: they only ask on joining or
// after their own queue gave up frames, and never make the stream lossy.
//
// Where the library cannot be asked for a keyframe, the GOP setting stays the
// encoder's GOP, capped at keyframeFallbackInterval, and nothing is forced.
const (
	keyframeRareInterval     = 2 * time.Second
	keyframeFallbackInterval = time.Second
	keyframeMinGap           = 250 * time.Millisecond

	lossWindow         = 5 * time.Second
	lossEventsFrequent = 3
	lossDebounce       = 200 * time.Millisecond
	lossHold           = 10 * time.Second

	// maxEncoderGop is what set_h264_gop can carry. Both libraries clamp
	// further: Sipeed's and libkvm-v4l2 to 100, the Coda980 to 99. So at 60
	// fps the clean interval is 1.65 to 1.67 s rather than 2 s (trial 64).
	maxEncoderGop = 255
)

// Reasons a keyframe is asked for, for the log.
const (
	KeyframeReasonWaiting = "a viewer waits for one"
	KeyframeReasonPicture = "a viewer lost a picture"
	keyframeReasonLossy   = "interval while a viewer loses packets"
	keyframeReasonOverdue = "no keyframe from the encoder's GOP"
)

// keyframeDiagEnv logs every forced keyframe and how many reads it took.
const keyframeDiagEnv = "KVM_KEYFRAME_DIAG"

type lossRecord struct {
	events []time.Time
	until  time.Time
}

// keyframePolicy holds what viewers asked for between two reads of the
// capture loop. Viewers call it from their own goroutines; the capture loop
// reads it once a tick.
type keyframePolicy struct {
	mutex   sync.Mutex
	pending string
	viewers map[any]*lossRecord
	lossy   bool
}

func newKeyframePolicy() *keyframePolicy {
	return &keyframePolicy{viewers: make(map[any]*lossRecord)}
}

var defaultKeyframePolicy = newKeyframePolicy()

// RequestKeyframe asks for a keyframe as soon as the policy allows: at the
// next read, unless one was forced less than keyframeMinGap ago.
func RequestKeyframe(reason string) {
	defaultKeyframePolicy.request(reason)
}

// ReportLoss records that viewer lost packets: a NACK, PLI or FIR. Repeated
// reports make the stream lossy (see above). viewer is any comparable key
// that stays the same for one viewer.
func ReportLoss(viewer any) {
	defaultKeyframePolicy.reportLoss(viewer, time.Now())
}

// ForgetViewer drops a viewer's loss record when it leaves, so a lossy viewer
// that is gone no longer decides the interval for the others.
func ForgetViewer(viewer any) {
	defaultKeyframePolicy.forget(viewer)
}

func (p *keyframePolicy) request(reason string) {
	p.mutex.Lock()
	if p.pending == "" {
		p.pending = reason
	}
	p.mutex.Unlock()
}

func (p *keyframePolicy) reportLoss(viewer any, now time.Time) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	record := p.viewers[viewer]
	if record == nil {
		record = &lossRecord{}
		p.viewers[viewer] = record
	}

	// One loss usually arrives as several reports: a NACK per RTCP interval
	// until the packet comes, then a PLI. They count once.
	if n := len(record.events); n > 0 && now.Sub(record.events[n-1]) < lossDebounce {
		return
	}

	kept := record.events[:0]
	for _, at := range record.events {
		if now.Sub(at) < lossWindow {
			kept = append(kept, at)
		}
	}
	record.events = append(kept, now)

	if len(record.events) >= lossEventsFrequent || now.Before(record.until) {
		record.until = now.Add(lossHold)
	}
}

func (p *keyframePolicy) forget(viewer any) {
	p.mutex.Lock()
	delete(p.viewers, viewer)
	p.mutex.Unlock()
}

// state answers whether any viewer is lossy now, and takes the pending
// request if forced keyframes are allowed now (allowed is false while the
// last forced one is less than keyframeMinGap old: the request then waits).
// changed reports a switch between clean and lossy, for the log.
func (p *keyframePolicy) state(now time.Time, allowed bool) (lossy bool, changed bool, pending string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for viewer, record := range p.viewers {
		if now.Before(record.until) {
			lossy = true
		} else if n := len(record.events); n == 0 || now.Sub(record.events[n-1]) >= lossWindow {
			delete(p.viewers, viewer)
		}
	}

	changed = lossy != p.lossy
	p.lossy = lossy

	if allowed {
		pending = p.pending
		p.pending = ""
	}

	return lossy, changed, pending
}

// keySeen clears a pending request: the keyframe just read reaches every
// viewer that asked before it, since a viewer asks after a frame it could
// not use and the keyframe comes after that frame.
func (p *keyframePolicy) keySeen() {
	p.mutex.Lock()
	p.pending = ""
	p.mutex.Unlock()
}

// frames converts an interval into frames at fps, at least 1.
func frames(interval time.Duration, fps int) int {
	if fps < 1 {
		fps = 1
	}

	n := int(math.Round(interval.Seconds() * float64(fps)))
	if n < 1 {
		n = 1
	}

	return n
}

// gopInterval is the GOP setting as time at fps.
func gopInterval(gop uint8, fps int) time.Duration {
	if fps < 1 {
		fps = 1
	}

	return time.Duration(gop) * time.Second / time.Duration(fps)
}

// encoderGop is the GOP the encoder is given. With keyframes on request it is
// the clean interval: keyframeRareInterval, or the GOP setting if longer.
// Without, the GOP setting, at most keyframeFallbackInterval.
func encoderGop(gop uint8, fps int, onRequest bool) uint8 {
	n := int(gop)
	if onRequest {
		n = max(n, frames(keyframeRareInterval, fps))
	} else {
		n = min(n, frames(keyframeFallbackInterval, fps))
	}

	return uint8(min(max(n, 1), maxEncoderGop))
}

// keyframeSchedule is the capture loop's side of the policy. Only the capture
// goroutine uses it.
type keyframeSchedule struct {
	policy    *keyframePolicy
	onRequest bool
	diag      bool

	lastKey    time.Time
	lastForced time.Time

	// For the diagnostic log: the read count when a keyframe was forced.
	reads       int
	forcedAt    int
	forcedFor   string
	forcedTime  time.Time
	awaitingKey bool
}

func newKeyframeSchedule(policy *keyframePolicy, onRequest bool) *keyframeSchedule {
	return &keyframeSchedule{
		policy:    policy,
		onRequest: onRequest,
		diag:      os.Getenv(keyframeDiagEnv) != "",
	}
}

// beforeRead decides whether to force a keyframe for the next read, and
// answers the reason, or "" for none. gop and fps are the settings.
func (s *keyframeSchedule) beforeRead(now time.Time, gop uint8, fps int) string {
	if !s.onRequest {
		return ""
	}

	// A stream opens on a keyframe: both libraries make one when they are
	// told the GOP at the start.
	if s.lastKey.IsZero() {
		s.lastKey = now
	}

	allowed := s.lastForced.IsZero() || now.Sub(s.lastForced) >= keyframeMinGap
	lossy, changed, pending := s.policy.state(now, allowed)

	if changed {
		if lossy {
			log.Infof("keyframes: a viewer is losing packets; at least every %s", gopInterval(gop, fps))
		} else {
			log.Infof("keyframes: no loss for %s; every %s again", lossHold, max(keyframeRareInterval, gopInterval(gop, fps)))
		}
	}

	if !allowed {
		return ""
	}

	since := now.Sub(s.lastKey)

	reason := ""
	switch {
	case pending != "":
		if since < keyframeMinGap {
			// One went out a moment ago, before this viewer asked. Ask again
			// once the gap has passed.
			s.policy.request(pending)
			return ""
		}
		reason = pending
	case lossy && since >= gopInterval(gop, fps):
		reason = keyframeReasonLossy
	case since >= max(keyframeRareInterval, gopInterval(gop, fps))*5/4:
		// The encoder's GOP should have made one by now. It may have been
		// clamped below what was asked, or not have taken the GOP at all.
		reason = keyframeReasonOverdue
	}

	if reason != "" {
		s.lastForced = now
		if s.diag {
			s.forcedAt = s.reads
			s.forcedFor = reason
			s.forcedTime = now
			s.awaitingKey = true
		}
	}

	return reason
}

// afterRead records what the read gave.
func (s *keyframeSchedule) afterRead(now time.Time, key bool) {
	s.reads++
	if !key {
		return
	}

	s.lastKey = now
	s.policy.keySeen()

	if s.diag && s.awaitingKey {
		s.awaitingKey = false
		log.Infof("keyframe forced (%s): came with read %d after the request, %d ms",
			s.forcedFor, s.reads-s.forcedAt, now.Sub(s.forcedTime).Milliseconds())
	}
}
