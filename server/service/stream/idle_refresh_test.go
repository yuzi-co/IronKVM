package stream

import (
	"testing"
	"time"
)

// At 2000 kbit/s and 60 fps a delta picture's budget is 4167 bytes: large from
// 3125, small under 1042, and motion is a run of 9 large pictures (150 ms).
const (
	idleTestLarge = 5000
	idleTestSmall = 300
	idleTestKey   = 90000
	frameMs       = 17
)

// readSized runs one tick at 60 fps, 2000 kbit/s: decide, then read a picture
// of size bytes, a keyframe if one was forced or key says so.
func readSized(s *keyframeSchedule, ms int, key bool, size int) string {
	reason := s.beforeRead(at(ms), 30, 60)
	key = key || reason != ""
	if key {
		size = idleTestKey
	}
	s.afterRead(at(ms), key, size, 2000, 60)

	return reason
}

// run reads from ms to end, one picture a frame, and lists when each idle
// refresh was forced. size and key say what the screen and the encoder's own
// GOP give at each time.
func run(s *keyframeSchedule, from, to int, size func(ms int) int, key func(ms int) bool) []int {
	var refreshes []int
	for ms := from; ms < to; ms += frameMs {
		if readSized(s, ms, key(ms), size(ms)) == keyframeReasonIdle {
			refreshes = append(refreshes, ms)
		}
	}

	return refreshes
}

func noKey(int) bool { return false }

func newIdleTestSchedule(t *testing.T) *keyframeSchedule {
	t.Helper()
	// The tests run at 60 fps: on everywhere.
	t.Setenv(idleRefreshEnv, "on")
	s := newKeyframeSchedule(newKeyframePolicy(), true)
	if s.idle == nil {
		t.Fatal("idle refresh is off with " + idleRefreshEnv + "=on")
	}
	readSized(s, 0, true, 0)

	return s
}

func TestTheScreenGoingStillAfterScrollingGetsOneKeyframe(t *testing.T) {
	s := newIdleTestSchedule(t)
	scroll := func(ms int) int {
		if ms >= 600 && ms < 1400 {
			return idleTestLarge
		}

		return idleTestSmall
	}

	got := run(s, frameMs, 2400, scroll, noKey)
	if len(got) != 1 {
		t.Fatalf("refreshes at %v ms, want one", got)
	}
	// The last large picture is the one read at 1394 ms; 200 ms of small
	// ones after it, then the next read.
	if got[0] < 1394+200 || got[0] > 1394+200+2*frameMs {
		t.Fatalf("refresh at %d ms, want about 200 ms after the motion stopped at 1394", got[0])
	}
}

func TestContinuousMotionGetsNoRefresh(t *testing.T) {
	s := newIdleTestSchedule(t)
	gop := func(ms int) bool { return ms%1666 < frameMs }

	if got := run(s, frameMs, 8000, func(int) int { return idleTestLarge }, gop); len(got) != 0 {
		t.Fatalf("refreshes at %v ms during continuous motion", got)
	}
}

// A screen that changes ten times a second: one large picture every 100 ms,
// small ones between, which leave it the bits of several. Not motion.
func TestAScreenChangingTenTimesASecondIsNotMotion(t *testing.T) {
	s := newIdleTestSchedule(t)
	scroll := func(ms int) int {
		if ms >= 300 && ms < 2600 && ms%100 < frameMs {
			return idleTestLarge
		}

		return idleTestSmall
	}
	gop := func(ms int) bool { return ms%1666 < frameMs }

	if got := run(s, frameMs, 3300, scroll, gop); len(got) != 0 {
		t.Fatalf("refreshes at %v ms", got)
	}
}

// The WAVE420L refines every other picture for a second or two after motion
// or a keyframe: runs of one large picture. Not motion.
func TestAnEncoderRefiningEveryOtherPictureIsNotMotion(t *testing.T) {
	s := newIdleTestSchedule(t)
	refine := func(ms int) int {
		if ms >= 700 && ms < 2500 && (ms/frameMs)%2 == 0 {
			return idleTestLarge
		}

		return idleTestSmall
	}

	if got := run(s, frameMs, 3300, refine, noKey); len(got) != 0 {
		t.Fatalf("refreshes at %v ms", got)
	}
}

// A clock's second, a cursor blink, one redraw of a status line: a picture or
// two, never a run of them.
func TestSmallChangesOnAStillScreenGetNoRefresh(t *testing.T) {
	s := newIdleTestSchedule(t)
	gop := func(ms int) bool { return ms%2000 < frameMs }
	ticks := func(ms int) int {
		switch {
		case ms%1000 < frameMs:
			return 20000 // the clock, or a status line redrawn
		case ms%1000 < 2*frameMs:
			return idleTestLarge
		case ms%530 < frameMs:
			return 1500 // a cursor blink
		}

		return idleTestSmall
	}

	if got := run(s, frameMs, 10000, ticks, gop); len(got) != 0 {
		t.Fatalf("refreshes at %v ms on a still screen", got)
	}
}

func TestRefreshesAreASecondApart(t *testing.T) {
	s := newIdleTestSchedule(t)
	// Two flicks of 200 ms, the second 650 ms after the first refresh: past
	// idleSettle, and still before idleMinGap when it ends.
	flicks := func(ms int) int {
		if (ms >= 1000 && ms < 1200) || (ms >= 2050 && ms < 2250) {
			return idleTestLarge
		}

		return idleTestSmall
	}

	got := run(s, frameMs, 4000, flicks, noKey)
	if len(got) != 2 {
		t.Fatalf("refreshes at %v ms, want two", got)
	}
	if gap := got[1] - got[0]; gap < int(idleMinGap/time.Millisecond) {
		t.Fatalf("refreshes %d ms apart", gap)
	}
}

// A keyframe from the GOP in the motion or just after it was made with the
// motion's bits spent: the refresh still goes out, idleKeyGap after it.
func TestAKeyframeJustAfterTheMotionDefersTheRefresh(t *testing.T) {
	scroll := func(ms int) int {
		if ms >= 600 && ms < 1400 {
			return idleTestLarge
		}

		return idleTestSmall
	}

	for _, key := range []int{1200, 1500} {
		s := newIdleTestSchedule(t)
		gop := func(ms int) bool { return ms >= key && ms < key+frameMs }
		got := run(s, frameMs, 2600, scroll, gop)
		if len(got) != 1 || got[0] < key+int(idleKeyGap/time.Millisecond) {
			t.Fatalf("keyframe at %d ms: refreshes at %v ms, want one at least %s after it", key, got, idleKeyGap)
		}
	}
}

// A keyframe idleKeyRepaint or more after the motion has painted the still
// screen: an armed refresh that still waits is dropped.
func TestALaterKeyframeReplacesAWaitingRefresh(t *testing.T) {
	r := newIdleRefresh()
	r.armed = true
	r.motionEnd = at(0)

	r.observe(at(500), true, false, idleTestKey, 2000, 60)
	if !r.armed {
		t.Fatal("a keyframe 500 ms after the motion dropped the refresh")
	}

	r.observe(at(700), true, false, idleTestKey, 2000, 60)
	if r.armed {
		t.Fatal("a keyframe 700 ms after the motion left the refresh armed")
	}
}

// After a keyframe of a still screen the encoder spends a few large pictures
// refining it. That is not motion.
func TestRefiningAKeyframeIsNotMotion(t *testing.T) {
	s := newIdleTestSchedule(t)
	gop := func(ms int) bool { return ms%1666 < frameMs }
	refine := func(ms int) int {
		if since := ms % 1666; since >= 250 && since < 400 {
			return idleTestLarge
		}

		return idleTestSmall
	}

	if got := run(s, frameMs, 8000, refine, gop); len(got) != 0 {
		t.Fatalf("refreshes at %v ms after keyframes of a still screen", got)
	}
}

func TestIdleRefreshCanBeTurnedOff(t *testing.T) {
	for _, v := range []string{"0", "off", "OFF"} {
		t.Setenv(idleRefreshEnv, v)
		if s := newKeyframeSchedule(newKeyframePolicy(), true); s.idle != nil {
			t.Fatalf("%s=%s left it on", idleRefreshEnv, v)
		}
	}

	t.Setenv(idleRefreshEnv, "")
	if s := newKeyframeSchedule(newKeyframePolicy(), false); s.idle != nil {
		t.Fatal("on without keyframe requests")
	}
}

func TestTheRefreshReadsBitrateIsTheSettingUnlessBoosted(t *testing.T) {
	t.Setenv(idleRefreshEnv, "")
	t.Setenv(idleRefreshEnv+"_BOOST", "")
	s := newKeyframeSchedule(newKeyframePolicy(), true)
	if got := s.refreshBitRate(2000); got != 2000 {
		t.Fatalf("refresh bitrate %d without a boost", got)
	}

	t.Setenv(idleRefreshEnv+"_BOOST", "300")
	s = newKeyframeSchedule(newKeyframePolicy(), true)
	if got := s.refreshBitRate(2000); got != 6000 {
		t.Fatalf("refresh bitrate %d at 300%%, want 6000", got)
	}
	if got := s.refreshBitRate(5000); got != idleBoostMaxKbps {
		t.Fatalf("refresh bitrate %d, want the library's most", got)
	}
}

func TestTheThresholdsFollowTheBudget(t *testing.T) {
	// At 30 fps the budget is 8333 bytes: 5000 is small-ish, not large.
	r := newIdleRefresh()
	for ms := 0; ms < 1000; ms += 33 {
		r.observe(at(ms), false, false, idleTestLarge, 2000, 30)
	}
	for ms := 1000; ms < 1400; ms += 33 {
		r.observe(at(ms), false, false, idleTestSmall, 2000, 30)
	}
	if r.armed {
		t.Fatal("5000-byte pictures at 30 fps counted as motion")
	}
}

// Once the motion is idleKeyRepaint old, a refresh that still waits gives way
// to a keyframe the GOP is about to make. An expected one that does not come
// (the WAVE420L restarts its GOP at a forced keyframe) holds it back
// idleGopLate at most.
func TestAWaitingRefreshGivesWayToTheGopsKeyframe(t *testing.T) {
	r := newIdleRefresh()
	r.armed = true
	r.motionEnd = at(0)
	// The GOP: keyframes nobody asked for at -2666 and -1000 ms, so the next
	// at 666.
	r.noteGop(at(-2666), false)
	r.noteGop(at(-1000), false)

	if !r.due(at(300), time.Time{}) {
		t.Fatal("a refresh 300 ms after the motion gave way to the GOP: its keyframe would be as blurred")
	}
	if r.due(at(650), time.Time{}) {
		t.Fatal("a refresh 650 ms after the motion went out 16 ms before the GOP's keyframe")
	}
	if r.due(at(750), time.Time{}) {
		t.Fatal("the refresh went out within idleGopLate of the GOP's keyframe")
	}
	if !r.due(at(800), time.Time{}) {
		t.Fatal("the refresh still waits for a GOP keyframe that did not come")
	}
}

// A refresh that could not go out within idleMaxLate of the motion's end is
// dropped: the encoder has refined the picture by then, or a GOP keyframe has
// repainted it.
func TestALateRefreshIsDropped(t *testing.T) {
	r := newIdleRefresh()
	r.armed = true
	r.motionEnd = at(0)

	if r.due(at(int(idleMaxLate/time.Millisecond)+20), time.Time{}) {
		t.Fatal("a refresh went out after idleMaxLate")
	}
	if r.armed {
		t.Fatal("a refresh past idleMaxLate is still armed")
	}
}

// By default it acts on libkvm-v4l2 at 30 fps and below, where it measured
// better, and observes everywhere.
func TestByDefaultItActsOnLibkvmV4L2AtThirtyFPSAndBelow(t *testing.T) {
	t.Setenv(idleRefreshEnv, "")
	s := newKeyframeSchedule(newKeyframePolicy(), true)
	if s.idle == nil {
		t.Fatal("idle refresh is off by default")
	}

	cases := []struct {
		v4l2 bool
		fps  int
		want bool
	}{
		{true, 30, true},
		{true, 24, true},
		{true, 60, false},
		{false, 30, false},
		{false, 60, false},
	}
	for _, c := range cases {
		s.v4l2 = c.v4l2
		if got := s.idleActive(c.fps); got != c.want {
			t.Errorf("libkvm-v4l2 %t at %d fps: active %t, want %t", c.v4l2, c.fps, got, c.want)
		}
	}

	t.Setenv(idleRefreshEnv, "on")
	s = newKeyframeSchedule(newKeyframePolicy(), true)
	if !s.idleActive(60) {
		t.Fatal(idleRefreshEnv + "=on does not act on Sipeed's library at 60 fps")
	}
}
