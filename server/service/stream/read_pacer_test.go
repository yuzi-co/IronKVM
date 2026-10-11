package stream

import (
	"testing"
	"time"
)

const (
	hz60  = time.Second / 60
	hz50  = time.Second / 50
	hz110 = time.Second / 110
)

func TestReadPacerFollowsTheCaptureTimes(t *testing.T) {
	t.Setenv(readPacingEnv, "")

	p := newReadPacer(time.Second / 30)
	defer p.stop()

	if p.paced() {
		t.Fatal("paced before any capture time")
	}

	captured := time.Now()
	p.read(captured, hz60)
	if !p.paced() {
		t.Fatal("not paced after a capture time")
	}

	// 30 fps from 60 Hz: the second source frame after the last, readLead
	// before it completes.
	if got, want := p.delay(captured), 2*hz60-readLead; got != want {
		t.Fatalf("delay %s, want %s", got, want)
	}

	// A read that ran past that goes at once.
	if got := p.delay(captured.Add(40 * time.Millisecond)); got != 0 {
		t.Fatalf("delay %s after a long read, want 0", got)
	}

	// A read with no frame moves the schedule on by a frame period.
	p.read(time.Time{}, 0)
	if got, want := p.delay(captured), time.Second/30+2*hz60-readLead; got != want {
		t.Fatalf("delay %s after a read with no frame, want %s", got, want)
	}

	// 60 fps from 60 Hz: the next source frame.
	p.read(captured.Add(100*time.Millisecond), hz60)
	p.setPeriod(time.Second / 60)
	if got, want := p.delay(captured.Add(100*time.Millisecond)), hz60-readLead; got != want {
		t.Fatalf("delay %s at 60 fps, want %s", got, want)
	}
}

// Never faster than asked: the frame taken is the first one at least a
// period, less half a source frame, after the last.
func TestReadPacerNeverReadsFasterThanAsked(t *testing.T) {
	t.Setenv(readPacingEnv, "")

	cases := []struct {
		fps   int
		frame time.Duration
		next  time.Duration
	}{
		{30, hz50, 2 * hz50},   // 25 fps, not 50
		{60, hz110, 2 * hz110}, // 55 fps, not 110
		{30, hz110, 4 * hz110}, // 27.5 fps
		{60, hz60, hz60},
		{30, 0, time.Second / 30},            // source frame time unknown
		{60, time.Second / 50, hz60},         // a source slower than asked
		{60, 16683 * time.Microsecond, hz60}, // 59.94 Hz
	}

	for _, c := range cases {
		p := newReadPacer(time.Second / time.Duration(c.fps))
		captured := time.Now()
		p.read(captured, c.frame)
		if got := p.next().Sub(captured); got != c.next {
			t.Errorf("%d fps, source frame %s: next frame after %s, want %s", c.fps, c.frame, got, c.next)
		}
		p.stop()
	}
}

func TestReadPacerKeepsTheTickerWhenAsked(t *testing.T) {
	t.Setenv(readPacingEnv, "tick")

	p := newReadPacer(time.Second / 30)
	defer p.stop()

	p.read(time.Now(), hz60)
	if p.paced() {
		t.Fatal("paced with KVM_READ_PACING=tick")
	}
}
