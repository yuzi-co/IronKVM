package stream

import (
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Idle refresh (ironkvm-dist#72, trial 67).
//
// With keyframes rare on a clean network (keyframe_policy.go: one every 1.7 to
// 2 s), a screen that stops moving stays blurred until the next one. While the
// screen moves, every delta picture spends its share of the bitrate on the
// motion and little on detail. Once it stops, the Coda980 refines for a few
// hundred milliseconds at most and then sends delta pictures of a few hundred
// bytes that add nothing (trials 61 and 67), so the picture keeps much of the
// motion's blur until the GOP's keyframe repaints it, up to 2 s later.
//
// So when the screen goes still after motion, one keyframe is forced. Motion
// and stillness are read from what the server already has, the size of each
// delta picture against its share of the bitrate (bitrate / 8 / fps, the
// budget):
//
//   - large: idleLargeShare of the budget or more. The encoder spends all it
//     may on the change and still falls short: what it sends is blurred.
//   - small: under idleSmallShare. Nothing much changed.
//   - motion: a run of large pictures that lasts idleMotionFor (at least
//     idleMotionMinFrames). A small picture ends a run; a medium one neither
//     ends nor counts.
//   - still: after motion, idleStillFor with only small pictures. That arms
//     the refresh.
//
// What does not count as motion (trial 67, slot A and slot B):
//
//   - a cursor blink, a clock's second, one redraw of a status line, a key
//     typed: a picture or two;
//   - a screen that changes ten times a second: one large picture, then small
//     ones that leave it the bits of several;
//   - an encoder refining what it sent: the WAVE420L every other picture for
//     a second or two after motion or a keyframe, the Coda980 and Sipeed's
//     H.265 a few large pictures after a keyframe. Large pictures within
//     idleSettle of a keyframe on a screen that is not moving do not count.
//
// Sizes cannot tell motion the encoder kept up with from motion it did not.
// Text scrolled a line a picture is predicted whole by the Coda980's motion
// search and is 45 to 47 dB the moment it stops, yet its pictures are as
// large as those of text scrolled eight lines a picture, which stops at 17 to
// 27 dB. A refresh after the first costs some of that sharpness, but only
// sooner: on slot B a still screen settles at its keyframe's quality, and the
// GOP's next keyframe would bring it down as far within 1.7 to 2 s (trial 65).
//
// Keyframes:
//
//   - one in the motion, or within idleKeyRepaint after its end, does not
//     end it or disarm. The encoder made it with the motion's bits spent, and
//     it is as blurred as the screen (27 to 35 kB and 17 to 21 dB on slot B
//     at 60 fps, against 80 to 120 kB and 30 to 35 dB two seconds later). The
//     refresh then waits idleKeyGap after it.
//   - one later than that has painted the still screen: it disarms.
//
// An armed refresh goes out at the next read once a keyframe of any kind is
// idleKeyGap old and the last refresh idleMinGap old, and once the policy's
// keyframeMinGap allows a forced one. New motion before that disarms it. One
// that cannot go out within idleMaxLate of the motion's end is dropped: by
// then the encoder has refined the picture or a GOP keyframe has repainted it.
//
// It also waits while the encoder's own GOP is about to make a keyframe
// (within idleGopSoon before the next one is due, or idleGopLate after),
// once the motion is idleKeyRepaint old and that keyframe would paint the
// still screen. A forced keyframe does not restart the GOP of the Coda980 or
// of Sipeed's library (trial 67: their next keyframe came on the GOP's
// schedule, 85 ms after a refresh once). The GOP's interval is learnt from two
// keyframes in a row that nobody asked for, and its next one is counted from
// the last of those. Where a forced keyframe restarts the GOP (the WAVE420L)
// the guess is wrong after a refresh, and the refresh then waits idleGopLate
// longer at most.
//
// Where it acts by default: on libkvm-v4l2 (slot B) at 30 fps and below.
// Measured with text scrolled eight lines a picture at 2000 kbit/s (trial 67):
//
//   - slot B, H.264, 30 fps: sharp (within 1 dB of the still screen's best)
//     1.7 s after the stop instead of 3.9 s through H.264 Direct, 2.8 s
//     instead of 3.3 s through WebRTC; 2.7 to 5.3 dB more from 0.5 to 2 s.
//   - slot B, H.264, 60 fps: no better. libkvm-v4l2's H.264 buffer is 300 ms
//     there, and a keyframe within a second of the motion was 21 to 27 kB and
//     17 to 19 dB in two stills of three: the motion has emptied it.
//   - slot B, H.265: the WAVE420L refines every other picture for a second or
//     two after motion, which is not motion here and is still not still: the
//     refresh comes after idleMaxLate and is dropped. Its own refinement
//     reaches 40 dB within 1 to 2 s.
//   - slot A (Sipeed's library): its keyframes after motion are 27 to 55 kB
//     and 18 to 26 dB, blurrier than the delta pictures it refines with (its
//     H.264 ones on a still screen are half the budget at 60 fps, so it
//     seldom goes still at all), and a refresh made the picture worse or no
//     better on both codecs.
//
// KVM_IDLE_REFRESH: off (or 0) turns it off, on (or 1) makes it act on every
// library and frame rate; unset or auto is the default above. The others tune
// it for trials:
// KVM_IDLE_REFRESH_LARGE and _SMALL (percent of the budget), _STILL_MS,
// _MOTION_MS, and _BOOST, the bitrate in percent of the setting for the read
// that carries the refresh keyframe, so a library whose keyframe follows the
// bitrate makes a larger one (0 or 100: none; slot B only).
const (
	idleRefreshEnv = "KVM_IDLE_REFRESH"

	// idleAutoMaxFPS is the highest frame rate at which it acts by default:
	// libkvm-v4l2's H.264 buffer is 500 ms there (1000 x 30 / GOP, the GOP 2 s).
	idleAutoMaxFPS = 30

	idleLargeShare      = 0.75
	idleSmallShare      = 0.25
	idleStillFor        = 200 * time.Millisecond
	idleMotionFor       = 150 * time.Millisecond
	idleMotionMinFrames = 3
	idleMinGap          = time.Second
	idleKeyGap          = 500 * time.Millisecond
	idleKeyRepaint      = 600 * time.Millisecond
	idleMaxLate         = 1200 * time.Millisecond
	idleSettle          = 600 * time.Millisecond
	idleGopSoon         = 400 * time.Millisecond
	idleGopLate         = 100 * time.Millisecond
)

type idleRefresh struct {
	// everywhere: KVM_IDLE_REFRESH=on, past the default's library and frame
	// rate.
	everywhere bool

	largeShare float64
	smallShare float64
	stillFor   time.Duration
	motionFor  time.Duration
	boost      int

	run         int  // large pictures in the current run
	motion      bool // a run of idleMotionFor since the screen was last still
	lastBusy    time.Time
	motionEnd   time.Time // the last large picture of a run that was motion
	settleUntil time.Time
	armed       bool
	lastRefresh time.Time

	// The encoder's own GOP: its last keyframe, its interval, and whether a
	// forced keyframe came since that one.
	lastGopKey  time.Time
	gopPeriod   time.Duration
	forcedSince bool
}

func newIdleRefresh() *idleRefresh {
	return &idleRefresh{
		largeShare: idleLargeShare,
		smallShare: idleSmallShare,
		stillFor:   idleStillFor,
		motionFor:  idleMotionFor,
	}
}

// newIdleRefreshFromEnv is newIdleRefresh with the environment's settings, or
// nil when KVM_IDLE_REFRESH turns it off.
func newIdleRefreshFromEnv() *idleRefresh {
	r := newIdleRefresh()
	switch v := strings.ToLower(strings.TrimSpace(os.Getenv(idleRefreshEnv))); v {
	case "0", "off", "false", "no":
		log.Infof("keyframes: idle refresh off (%s)", idleRefreshEnv)

		return nil
	case "1", "on", "true", "yes":
		r.everywhere = true
	case "", "auto":
	default:
		log.Warnf("%s=%q ignored: off, on or auto", idleRefreshEnv, v)
	}

	if v, ok := envUint(idleRefreshEnv + "_LARGE"); ok {
		r.largeShare = float64(v) / 100
	}
	if v, ok := envUint(idleRefreshEnv + "_SMALL"); ok {
		r.smallShare = float64(v) / 100
	}
	if v, ok := envUint(idleRefreshEnv + "_STILL_MS"); ok {
		r.stillFor = time.Duration(v) * time.Millisecond
	}
	if v, ok := envUint(idleRefreshEnv + "_MOTION_MS"); ok {
		r.motionFor = time.Duration(v) * time.Millisecond
	}
	if v, ok := envUint(idleRefreshEnv + "_BOOST"); ok && v > 100 {
		r.boost = v
	}

	return r
}

func envUint(name string) (int, bool) {
	s := strings.TrimSpace(os.Getenv(name))
	if s == "" {
		return 0, false
	}

	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		log.Warnf("%s=%q ignored: not a whole number", name, s)

		return 0, false
	}

	return v, true
}

// motionFrames is idleMotionFor in pictures at fps, at least
// idleMotionMinFrames.
func (r *idleRefresh) motionFrames(fps int) int {
	return max(frames(r.motionFor, fps), idleMotionMinFrames)
}

// observe takes one picture the encoder made: whether it is a keyframe and
// whether one was forced, its size in bytes, and the bitrate (kbit/s) and
// frame rate it was made for.
func (r *idleRefresh) observe(now time.Time, key bool, forced bool, size int, bitRate int, fps int) {
	if key {
		r.noteGop(now, forced)
		switch {
		case r.motion || r.run > 0:
			// In the motion, or just after it while the encoder still refines:
			// it goes on.
		case r.armed && now.Sub(r.motionEnd) <= idleKeyRepaint:
			// Made with the motion's bits spent: the refresh still goes out,
			// idleKeyGap after it.
		default:
			r.armed = false
			r.settleUntil = now.Add(idleSettle)
		}

		return
	}

	if fps < 1 || bitRate < 1 {
		return
	}

	budget := float64(bitRate) * 1000 / 8 / float64(fps)
	switch {
	case float64(size) >= r.largeShare*budget:
		r.lastBusy = now
		if !r.motion && now.Before(r.settleUntil) {
			// The encoder refining a keyframe of a screen that is not moving.
			return
		}
		r.run++
		r.armed = false
		if r.run >= r.motionFrames(fps) {
			r.motion = true
			r.motionEnd = now
		}
	case float64(size) >= r.smallShare*budget:
		r.lastBusy = now
	default:
		r.run = 0
		if r.motion && now.Sub(r.lastBusy) >= r.stillFor {
			r.armed = true
			r.motion = false
		}
	}
}

// noteGop follows the encoder's own GOP: a keyframe nobody asked for after
// another one nobody asked for gives its interval.
func (r *idleRefresh) noteGop(now time.Time, forced bool) {
	if forced {
		r.forcedSince = true

		return
	}

	if !r.lastGopKey.IsZero() && !r.forcedSince {
		r.gopPeriod = now.Sub(r.lastGopKey)
	}
	r.lastGopKey = now
	r.forcedSince = false
}

// due answers whether an armed refresh may go out now. lastKey is the last
// keyframe of any kind.
func (r *idleRefresh) due(now time.Time, lastKey time.Time) bool {
	if !r.armed {
		return false
	}
	if now.Sub(r.motionEnd) > idleMaxLate {
		r.armed = false

		return false
	}
	if !lastKey.IsZero() && now.Sub(lastKey) < idleKeyGap {
		return false
	}
	if !r.lastRefresh.IsZero() && now.Sub(r.lastRefresh) < idleMinGap {
		return false
	}

	// A GOP keyframe this long after the motion paints the still screen
	// itself; one sooner is as blurred as the screen.
	if now.Sub(r.motionEnd) >= idleKeyRepaint && r.gopSoon(now) {
		return false
	}

	return true
}

// gopSoon answers whether the encoder's GOP is due to make a keyframe within
// idleGopSoon, or was due less than idleGopLate ago.
func (r *idleRefresh) gopSoon(now time.Time) bool {
	if r.gopPeriod <= 0 || r.lastGopKey.IsZero() {
		return false
	}

	next := r.lastGopKey.Add(r.gopPeriod)
	for next.Before(now.Add(-idleGopLate)) {
		next = next.Add(r.gopPeriod)
	}

	return next.Sub(now) <= idleGopSoon
}

// idleActive answers whether the idle refresh acts at fps; it observes every
// picture either way.
func (s *keyframeSchedule) idleActive(fps int) bool {
	if s.idle == nil {
		return false
	}

	return s.idle.everywhere || (s.v4l2 && fps <= idleAutoMaxFPS)
}

// idlePlan says where the idle refresh acts, for the log.
func (s *keyframeSchedule) idlePlan(fps int) string {
	switch {
	case s.idle == nil:
		return "off"
	case s.idle.everywhere:
		return "on (" + idleRefreshEnv + "=on)"
	case s.idleActive(fps):
		return "on"
	default:
		return "inactive (libkvm-v4l2 at 30 fps and below only)"
	}
}

// idleBoostMaxKbps is the most kvmv_read_video takes (libkvm-v4l2's
// KVMV_BITRATE_MAX_KBPS).
const idleBoostMaxKbps = 10000

// refreshBitRate is the bitrate in kbit/s for the read that carries a refresh
// keyframe: the setting, or with KVM_IDLE_REFRESH_BOOST that share of it.
// libkvm-v4l2 sets a changed bitrate on the encoder before the picture it
// reads, and the next read puts the setting back the same way.
func (s *keyframeSchedule) refreshBitRate(bitRate uint16) uint16 {
	if s.idle == nil || s.idle.boost <= 100 {
		return bitRate
	}

	return uint16(min(int(bitRate)*s.idle.boost/100, idleBoostMaxKbps))
}

// fired records a refresh asked for at now.
func (r *idleRefresh) fired(now time.Time) {
	r.armed = false
	r.lastRefresh = now
}
