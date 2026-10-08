package stream

import (
	"sync"
	"testing"
	"time"

	"NanoKVM-Server/common"
)

func TestEncoderGopIsTheCleanIntervalWhenKeyframesCanBeAskedFor(t *testing.T) {
	cases := []struct {
		gop       uint8
		fps       int
		onRequest bool
		want      uint8
	}{
		{30, 60, true, 120},  // 2 s
		{30, 30, true, 60},   // 2 s
		{100, 30, true, 100}, // the setting is longer than 2 s
		{30, 120, true, 240},
		{30, 144, true, 255}, // what set_h264_gop can carry
		{30, 60, false, 30},  // the setting, under 1 s
		{60, 30, false, 30},  // at most 1 s
		{30, 0, true, 30},    // no frame rate yet: 2 frames, the setting wins
	}

	for _, c := range cases {
		if got := encoderGop(c.gop, c.fps, c.onRequest); got != c.want {
			t.Errorf("encoderGop(%d, %d, %t) = %d, want %d", c.gop, c.fps, c.onRequest, got, c.want)
		}
	}
}

// at is a fixed clock for the schedule tests.
func at(ms int) time.Time {
	return time.Unix(1000, 0).Add(time.Duration(ms) * time.Millisecond)
}

// read runs one tick of the schedule: decide, then read a frame that is a
// keyframe if one was forced or key says so.
func read(s *keyframeSchedule, ms int, key bool) string {
	reason := s.beforeRead(at(ms), 30, 60)
	s.afterRead(at(ms), key || reason != "", 0, 2000, 60)

	return reason
}

func TestAViewerThatAsksGetsAKeyframeAtTheNextRead(t *testing.T) {
	p := newKeyframePolicy()
	s := newKeyframeSchedule(p, true)
	read(s, 0, true)

	if r := read(s, 500, false); r != "" {
		t.Fatalf("forced %q with nobody asking", r)
	}

	p.request(KeyframeReasonWaiting)
	if r := read(s, 517, false); r != KeyframeReasonWaiting {
		t.Fatalf("forced %q, want %q", r, KeyframeReasonWaiting)
	}
	if r := read(s, 534, false); r != "" {
		t.Fatalf("forced %q again after the keyframe", r)
	}
}

func TestAKeyframeFromTheGopAnswersAPendingRequest(t *testing.T) {
	p := newKeyframePolicy()
	s := newKeyframeSchedule(p, true)
	read(s, 0, true)

	p.request(KeyframeReasonWaiting)
	s.afterRead(at(400), true, 100000, 2000, 60)

	if r := read(s, 417, false); r != "" {
		t.Fatalf("forced %q though the GOP's keyframe came after the request", r)
	}
}

// A viewer that keeps asking must not turn the stream into keyframes.
func TestForcedKeyframesAreKeptApart(t *testing.T) {
	p := newKeyframePolicy()
	s := newKeyframeSchedule(p, true)
	read(s, 0, true)

	p.request(KeyframeReasonPicture)
	if r := read(s, 1000, false); r == "" {
		t.Fatal("the first request was not answered")
	}

	p.request(KeyframeReasonPicture)
	for ms := 1017; ms < 1000+int(keyframeMinGap/time.Millisecond); ms += 17 {
		if r := read(s, ms, false); r != "" {
			t.Fatalf("forced %q at %d ms, %s after the last", r, ms, at(ms).Sub(at(1000)))
		}
	}

	if r := read(s, 1000+int(keyframeMinGap/time.Millisecond), false); r != KeyframeReasonPicture {
		t.Fatalf("the waiting request was not answered after the gap: %q", r)
	}
}

func TestRepeatedLossShortensTheIntervalUntilItStops(t *testing.T) {
	p := newKeyframePolicy()
	s := newKeyframeSchedule(p, true)
	read(s, 0, true)

	viewer := new(int)
	p.reportLoss(viewer, at(100))
	p.reportLoss(viewer, at(1100))
	if r := read(s, 1200, false); r != "" {
		t.Fatalf("forced %q after two losses", r)
	}

	p.reportLoss(viewer, at(2100))
	// GOP 30 at 60 fps: 0.5 s. The last keyframe was at 0.
	if r := read(s, 2117, false); r != keyframeReasonLossy {
		t.Fatalf("forced %q, want the lossy interval", r)
	}
	if r := read(s, 2500, false); r != "" {
		t.Fatalf("forced %q 0.4 s after the last", r)
	}
	if r := read(s, 2617, false); r != keyframeReasonLossy {
		t.Fatalf("forced %q, want the lossy interval at 0.5 s", r)
	}

	// No loss for lossHold: back to the encoder's own interval, so nothing is
	// forced until the backstop, 2.5 s after the last keyframe.
	ms := 2617
	for ms += 500; ms < 2100+int(lossHold/time.Millisecond); ms += 500 {
		read(s, ms, false)
	}
	last := ms - 500
	for ; ms < last+2400; ms += 100 {
		if r := read(s, ms, false); r != "" {
			t.Fatalf("forced %q at %d ms, after the loss ended at 2.1 s", r, ms)
		}
	}
}

// One loss arrives as several reports: NACKs until the packet comes, then a
// PLI. They must not count as three.
func TestReportsOfOneLossCountOnce(t *testing.T) {
	p := newKeyframePolicy()
	viewer := new(int)
	for ms := 0; ms < 150; ms += 50 {
		p.reportLoss(viewer, at(ms))
	}

	if lossy, _, _ := p.state(at(200), true); lossy {
		t.Fatal("one burst of reports made the stream lossy")
	}
}

func TestALossyViewerThatLeavesNoLongerCounts(t *testing.T) {
	p := newKeyframePolicy()
	viewer := new(int)
	for ms := 0; ms < 3000; ms += 1000 {
		p.reportLoss(viewer, at(ms))
	}

	if lossy, _, _ := p.state(at(3000), true); !lossy {
		t.Fatal("three losses in 3 s did not make the stream lossy")
	}

	p.forget(viewer)
	if lossy, _, _ := p.state(at(3001), true); lossy {
		t.Fatal("still lossy after the viewer left")
	}
}

func TestTheLossOfOneViewerDecidesForAll(t *testing.T) {
	p := newKeyframePolicy()
	clean, lossy := new(int), new(int)
	p.reportLoss(clean, at(0))
	for ms := 0; ms < 3000; ms += 1000 {
		p.reportLoss(lossy, at(ms))
	}

	if got, _, _ := p.state(at(3000), true); !got {
		t.Fatal("one lossy viewer did not make the stream lossy")
	}
}

func TestAnOverdueKeyframeIsForced(t *testing.T) {
	p := newKeyframePolicy()
	s := newKeyframeSchedule(p, true)
	read(s, 0, true)

	for ms := 17; ms < 2500; ms += 17 {
		if r := read(s, ms, false); r != "" {
			t.Fatalf("forced %q at %d ms, before the backstop", r, ms)
		}
	}

	if r := read(s, 2500, false); r != keyframeReasonOverdue {
		t.Fatalf("forced %q at 2.5 s with no keyframe since 0, want the backstop", r)
	}
}

func TestNothingIsForcedWithoutKeyframeRequests(t *testing.T) {
	p := newKeyframePolicy()
	s := newKeyframeSchedule(p, false)
	read(s, 0, true)

	p.request(KeyframeReasonWaiting)
	for ms := 0; ms < 10000; ms += 100 {
		if r := read(s, ms, false); r != "" {
			t.Fatalf("forced %q", r)
		}
	}
}

// withKeyframeRequests makes the loop believe the library takes requests,
// and records them.
func withKeyframeRequests(t *testing.T) func() []uint8 {
	t.Helper()

	var mutex sync.Mutex
	var asked []uint8

	originalKind, originalRequest := keyframeRequests, requestKeyframe
	t.Cleanup(func() { keyframeRequests, requestKeyframe = originalKind, originalRequest })
	keyframeRequests = func() common.KeyframeRequestKind { return common.KeyframeRequestV4L2 }
	requestKeyframe = func(gop uint8) bool {
		mutex.Lock()
		defer mutex.Unlock()
		asked = append(asked, gop)

		return true
	}

	return func() []uint8 {
		mutex.Lock()
		defer mutex.Unlock()

		return append([]uint8(nil), asked...)
	}
}

// The loop asks with the GOP the encoder holds: libkvm-v4l2 takes the request
// through set_h264_gop, and any other number would change the GOP as well.
func TestTheLoopAsksTheEncoderWithTheGopItHolds(t *testing.T) {
	withScreenGop(t, 30)
	withScreenFPS(t, 30)
	told := withEncoderGop(t)
	asked := withKeyframeRequests(t)
	withCapture(t, func(uint16, uint16, uint16) ([]byte, int) {
		return []byte{0x00, 0x00, 0x00, 0x01}, 4
	})

	source := newH264Source()
	subscription := source.subscribe(nil)
	defer subscription.Close()

	waitFor(t, "the gop", func() bool { return len(told()) > 0 })
	if got := told()[0]; got != 60 {
		t.Fatalf("the encoder was told GOP %d, want 2 s at 30 fps", got)
	}

	RequestKeyframe(KeyframeReasonWaiting)
	waitFor(t, "the keyframe request", func() bool { return len(asked()) > 0 })

	if got := asked()[0]; got != 60 {
		t.Fatalf("asked with GOP %d, want the encoder's 60", got)
	}
}
