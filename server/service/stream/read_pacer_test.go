package stream

import (
	"testing"
	"time"
)

func TestReadPacerFollowsTheCaptureTimes(t *testing.T) {
	t.Setenv(readPacingEnv, "")

	p := newReadPacer(time.Second / 30)
	defer p.stop()

	if p.paced() {
		t.Fatal("paced before any capture time")
	}

	captured := time.Now()
	p.read(captured)
	if !p.paced() {
		t.Fatal("not paced after a capture time")
	}

	// The next read starts readLead before the next frame is due.
	want := time.Second/30 - readLead
	if got := p.delay(captured); got != want {
		t.Fatalf("delay %s, want %s", got, want)
	}

	// A read that ran past that goes at once.
	if got := p.delay(captured.Add(40 * time.Millisecond)); got != 0 {
		t.Fatalf("delay %s after a long read, want 0", got)
	}

	// A read with no frame moves the schedule on by a frame time.
	p.read(time.Time{})
	if got := p.delay(captured); got != 2*time.Second/30-readLead {
		t.Fatalf("delay %s after a read with no frame", got)
	}

	// A frame rate change applies from the next frame.
	p.read(captured.Add(100 * time.Millisecond))
	p.setPeriod(time.Second / 60)
	if got := p.delay(captured.Add(100 * time.Millisecond)); got != time.Second/60-readLead {
		t.Fatalf("delay %s at 60 fps", got)
	}
}

func TestReadPacerKeepsTheTickerWhenAsked(t *testing.T) {
	t.Setenv(readPacingEnv, "tick")

	p := newReadPacer(time.Second / 30)
	defer p.stop()

	p.read(time.Now())
	if p.paced() {
		t.Fatal("paced with KVM_READ_PACING=tick")
	}
}
