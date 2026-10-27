package dashcam

import (
	"testing"
	"time"
)

// gopStream makes n samples at fps with a keyframe every gop samples, each
// carrying size bytes of payload.
func gopStream(n, fps, gop, size int) []Sample {
	out := make([]Sample, n)
	frame := time.Second / time.Duration(fps)
	for i := range out {
		payload := make([]byte, size)
		out[i] = Sample{
			Key:   i%gop == 0,
			NALUs: [][]byte{payload},
			PTS:   time.Second + time.Duration(i)*frame,
		}
	}
	return out
}

func checkStartsOnKey(t *testing.T, r *Ring) {
	t.Helper()
	snap := r.Snapshot()
	if len(snap) > 0 && !snap[0].Key {
		t.Fatalf("the ring starts on a delta frame at PTS %s", snap[0].PTS)
	}
}

func TestRingRefusesDeltaFramesUntilAKeyframe(t *testing.T) {
	r := NewRing(10*time.Second, 1<<20)
	stream := gopStream(25, 10, 10, 100)

	for i := 1; i < 10; i++ {
		if r.Push(stream[i]) {
			t.Fatalf("sample %d, a delta frame, was kept by an empty ring", i)
		}
	}
	if !r.Waiting() || r.Len() != 0 {
		t.Fatalf("the ring holds %d samples before any keyframe", r.Len())
	}
	for i := 10; i < 25; i++ {
		if !r.Push(stream[i]) {
			t.Fatalf("sample %d was refused after the keyframe", i)
		}
	}
	if r.Len() != 15 {
		t.Fatalf("the ring holds %d samples, want 15", r.Len())
	}
	checkStartsOnKey(t, r)
}

// The time bound keeps the window and at most one GOP more, from a keyframe.
func TestRingKeepsTheWindowFromAKeyframe(t *testing.T) {
	const window = 3 * time.Second
	r := NewRing(window, 1<<30)
	// 1 s GOPs at 10 fps for a minute.
	for _, s := range gopStream(600, 10, 10, 100) {
		r.Push(s)
		checkStartsOnKey(t, r)
	}

	if span := r.Span(); span < window || span >= window+time.Second {
		t.Fatalf("the ring spans %s, want %s up to one GOP more", span, window)
	}
	// 3.9 s from the first keyframe to the last sample: four GOPs.
	if r.Len() != 40 {
		t.Fatalf("the ring holds %d samples, want 40", r.Len())
	}
}

// The byte bound drops whole GOPs from the front.
func TestRingKeepsToItsBudget(t *testing.T) {
	const budget = 25_000
	r := NewRing(time.Hour, budget)
	for _, s := range gopStream(500, 10, 10, 900) {
		r.Push(s)
		if r.Bytes() > budget {
			t.Fatalf("the ring holds %d bytes, over its budget of %d", r.Bytes(), budget)
		}
		checkStartsOnKey(t, r)
	}
	// A GOP is 10 x (900 + 64 + 24) = 9880 bytes. The ring drops one only
	// when it must, so it never holds less than the budget minus a GOP.
	if r.Bytes() <= budget-9880 {
		t.Fatalf("the ring holds %d bytes, more was dropped than the budget needs", r.Bytes())
	}
}

// One GOP larger than the budget leaves nothing to start from but the next
// keyframe.
func TestRingDropsAGOPLargerThanItsBudget(t *testing.T) {
	r := NewRing(time.Hour, 5_000)
	stream := gopStream(30, 10, 20, 900)
	for _, s := range stream[:6] {
		r.Push(s)
	}
	if !r.Waiting() {
		t.Fatalf("the ring holds %d samples of a GOP over its budget", r.Len())
	}
	// The rest of the GOP is refused, and the next keyframe starts again.
	for _, s := range stream[6:20] {
		if r.Push(s) {
			t.Fatal("a delta frame was kept after the ring dropped its GOP")
		}
	}
	if !r.Push(stream[20]) || r.Len() != 1 {
		t.Fatalf("the next keyframe was not kept: %d samples", r.Len())
	}
}

// The ring wraps for as long as the stream runs, and holds no more than its
// bounds whatever the length.
func TestRingWrapsWithoutGrowing(t *testing.T) {
	r := NewRing(2*time.Second, 1<<20)
	maxLen := 0
	for _, s := range gopStream(20_000, 30, 30, 50) {
		r.Push(s)
		maxLen = max(maxLen, r.Len())
	}
	// 2 s plus a GOP of 1 s at 30 fps, and one frame: 60 frame times of
	// 33.333333 ms fall a few nanoseconds short of the window.
	if maxLen > 91 {
		t.Fatalf("the ring grew to %d samples", maxLen)
	}
	checkStartsOnKey(t, r)
	if cap(r.samples) > 4*90 {
		t.Fatalf("the ring's array grew to %d entries", cap(r.samples))
	}
}

// A clock that goes back is a new tap; the ring starts again on it.
func TestRingStartsAgainWhenTheClockGoesBack(t *testing.T) {
	r := NewRing(time.Hour, 1<<20)
	for _, s := range gopStream(20, 10, 10, 10) {
		r.Push(s)
	}
	restart := Sample{Key: false, NALUs: [][]byte{{1}}, PTS: 0}
	if r.Push(restart) || !r.Waiting() {
		t.Fatalf("a delta frame on a new clock was kept, or the old samples stayed: %d", r.Len())
	}
}

func TestRingSnapshotIsACopy(t *testing.T) {
	r := NewRing(time.Hour, 1<<20)
	for _, s := range gopStream(20, 10, 10, 10) {
		r.Push(s)
	}
	snap := r.Snapshot()
	r.Reset()
	if len(snap) != 20 || snap[0].NALUs == nil {
		t.Fatalf("the snapshot changed with the ring: %d samples", len(snap))
	}
}

func TestRingSetBoundsTrims(t *testing.T) {
	r := NewRing(time.Hour, 1<<20)
	for _, s := range gopStream(100, 10, 10, 10) {
		r.Push(s)
	}
	r.SetBounds(2*time.Second, 1<<20)
	if span := r.Span(); span < 2*time.Second || span >= 3*time.Second {
		t.Fatalf("the ring spans %s after a shorter window", span)
	}
	checkStartsOnKey(t, r)
}
