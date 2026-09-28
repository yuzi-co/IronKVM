package stream

import "testing"

// useStartedCounter sets what CurrentFPS sees for one test. Other tests in
// this package start the real counter, so the state is set rather than
// assumed.
func useStartedCounter(t *testing.T, c *FrameRateCounter) {
	t.Helper()

	previous := startedCounter.Swap(c)
	t.Cleanup(func() { startedCounter.Store(previous) })
}

// Before a stream starts the counter, the rate is 0 and reported as not
// started. Reading it does not start it: its ticker writes now_fps to the card
// every three seconds, and a scrape is not a stream. The metrics package
// checks the same thing against the real counter.
func TestCurrentFPSBeforeTheCounterStarts(t *testing.T) {
	useStartedCounter(t, nil)

	if fps, started := CurrentFPS(); fps != 0 || started {
		t.Fatalf("CurrentFPS() = %d, %t, want 0, false", fps, started)
	}
	if startedCounter.Load() != nil {
		t.Fatal("CurrentFPS started the counter")
	}
}

func TestCurrentFPSReadsAStartedCounter(t *testing.T) {
	c := &FrameRateCounter{}
	c.fps.Store(24)
	useStartedCounter(t, c)

	if fps, started := CurrentFPS(); fps != 24 || !started {
		t.Fatalf("CurrentFPS() = %d, %t, want 24, true", fps, started)
	}
}
