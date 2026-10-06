package webrtc

import (
	"testing"
	"time"
)

func TestFrameClockKeepsTheRegularStep(t *testing.T) {
	var c frameClock
	frame := time.Second / 60

	// The first frame, then frames a frame time apart, give or take a few ms.
	for i, us := range []int64{1000, 17667, 35000, 52000, 69000} {
		if got := c.skip(us, frame, clockRate); got != 0 {
			t.Fatalf("frame %d at %d us: skipped %d samples, want 0", i, us, got)
		}
	}
}

func TestFrameClockSkipsMissingFrameTimes(t *testing.T) {
	var c frameClock
	frame := time.Second / 60

	c.skip(0, frame, clockRate)
	// Five frame times later: four were never sent.
	got := c.skip(5*16667, frame, clockRate)
	want := uint32(4 * clockRate / 60)
	if got < want-2 || got > want+2 {
		t.Fatalf("skipped %d samples, want about %d", got, want)
	}

	// Back to the regular step.
	if got := c.skip(6*16667, frame, clockRate); got != 0 {
		t.Fatalf("skipped %d samples after the gap, want 0", got)
	}
}

func TestFrameClockIgnoresRestartsAndLongStops(t *testing.T) {
	var c frameClock
	frame := time.Second / 30

	c.skip(5_000_000, frame, clockRate)
	// The capture loop started again: its clock went back.
	if got := c.skip(1000, frame, clockRate); got != 0 {
		t.Fatalf("skipped %d samples across a restart, want 0", got)
	}
	// A stop longer than maxClockSkip.
	if got := c.skip(1000+int64(maxClockSkip/time.Microsecond)+1, frame, clockRate); got != 0 {
		t.Fatalf("skipped %d samples across a long stop, want 0", got)
	}
	// No frame time known.
	if got := c.skip(100_000_000, 0, clockRate); got != 0 {
		t.Fatalf("skipped %d samples with no frame time, want 0", got)
	}
}
