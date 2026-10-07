package webrtc

import "testing"

// countProcs replaces moreProcs for one test and reports the holders.
func countProcs(t *testing.T) *int {
	t.Helper()

	held := 0
	old := moreProcs
	moreProcs = func() func() {
		held++
		released := false

		return func() {
			if released {
				t.Fatal("a second P released twice")
			}
			released = true
			held--
		}
	}
	t.Cleanup(func() { moreProcs = old })

	return &held
}

func TestSoloProcsHoldsASecondPForOneViewerOnly(t *testing.T) {
	held := countProcs(t)

	var solo soloProcs
	steps := []struct {
		viewers int
		held    int
	}{
		{0, 0},
		{1, 1},
		{1, 1}, // no second hold while the count stays
		{2, 0},
		{3, 0},
		{1, 1},
		{0, 0},
	}

	for i, step := range steps {
		solo.update(step.viewers)
		if *held != step.held {
			t.Fatalf("step %d, %d viewers: %d holds, want %d", i, step.viewers, *held, step.held)
		}
	}
}

func TestSoloProcsLetsGoWhenTheStreamEnds(t *testing.T) {
	held := countProcs(t)

	var solo soloProcs
	solo.update(1)
	solo.update(0)
	solo.update(0)

	if *held != 0 {
		t.Fatalf("%d holds after the stream ended", *held)
	}
}
