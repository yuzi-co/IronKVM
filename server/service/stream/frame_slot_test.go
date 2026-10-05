package stream

import (
	"sync"
	"testing"
)

func TestFrameSlotTakeReturnsWhatWasPut(t *testing.T) {
	slot := NewFrameSlot[int]()

	if !slot.TryPut(7) {
		t.Fatal("an empty slot should accept a frame")
	}

	value, ok := slot.Take()
	if !ok || value != 7 {
		t.Fatalf("expected 7, got %d (ok=%v)", value, ok)
	}
}

func TestTryPutRefusesWhenAFrameIsPending(t *testing.T) {
	// This is what tells an H.264 producer that the client is behind.
	slot := NewFrameSlot[int]()
	slot.TryPut(1)

	if slot.TryPut(2) {
		t.Fatal("a slot holding a frame must refuse another")
	}

	value, _ := slot.Take()
	if value != 1 {
		t.Fatalf("the pending frame should be untouched, got %d", value)
	}
}

func TestReplaceKeepsTheNewestFrame(t *testing.T) {
	// MJPEG frames are independent, so a slow client should see the newest
	// frame rather than a stale one.
	slot := NewFrameSlot[int]()
	slot.TryPut(1)

	slot.Replace(2)

	value, _ := slot.Take()
	if value != 2 {
		t.Fatalf("expected the newest frame, got %d", value)
	}
}

func TestDroppedCountsEvictedAndRefusedFrames(t *testing.T) {
	slot := NewFrameSlot[int]()

	slot.TryPut(1)
	slot.TryPut(2) // refused
	slot.Replace(3)

	if slot.Dropped() != 2 {
		t.Fatalf("expected 2 dropped frames, got %d", slot.Dropped())
	}
}

func TestTakeUnblocksOnClose(t *testing.T) {
	slot := NewFrameSlot[int]()

	done := make(chan struct{})
	go func() {
		defer close(done)

		if _, ok := slot.Take(); ok {
			t.Error("Take on a closed slot should report no frame")
		}
	}()

	slot.Close()
	<-done
}

func TestPendingFrameIsStillDeliveredAfterClose(t *testing.T) {
	slot := NewFrameSlot[int]()
	slot.TryPut(9)
	slot.Close()

	value, ok := slot.Take()
	if !ok || value != 9 {
		t.Fatalf("a frame already queued should still be delivered, got %d (ok=%v)", value, ok)
	}
}

func TestPutOnClosedSlotIsRefusedNotPanicking(t *testing.T) {
	slot := NewFrameSlot[int]()
	slot.Close()

	if slot.TryPut(1) {
		t.Fatal("a closed slot must not accept frames")
	}
	if slot.Replace(2) {
		t.Fatal("a closed slot must not accept frames")
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	slot := NewFrameSlot[int]()

	slot.Close()
	slot.Close()
}

func TestFrameSlotSurvivesConcurrentUse(t *testing.T) {
	// Run with -race: one producer never blocks, one consumer drains.
	slot := NewFrameSlot[int]()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 2000; i++ {
			slot.Replace(i)
		}
		slot.Close()
	}()

	go func() {
		defer wg.Done()
		for {
			if _, ok := slot.Take(); !ok {
				return
			}
		}
	}()

	wg.Wait()
}

func TestADeepSlotTakesFramesUntilItIsFull(t *testing.T) {
	slot := NewFrameSlotDepth[int](3)

	for i := 1; i <= 3; i++ {
		if !slot.TryPut(i) {
			t.Fatalf("a slot of three refused frame %d", i)
		}
	}
	if slot.TryPut(4) {
		t.Fatal("a full slot must refuse another frame")
	}
	if slot.Dropped() != 1 {
		t.Fatalf("dropped = %d, want 1", slot.Dropped())
	}

	for want := 1; want <= 3; want++ {
		if got, _ := slot.Take(); got != want {
			t.Fatalf("took %d, want %d: frames must leave in order", got, want)
		}
	}
}

func TestReplaceOnADeepSlotDiscardsOnlyWhenFull(t *testing.T) {
	slot := NewFrameSlotDepth[int](2)

	slot.Replace(1)
	slot.Replace(2)
	if slot.Dropped() != 0 {
		t.Fatalf("a slot with room discarded %d frames", slot.Dropped())
	}

	slot.Replace(3)
	if slot.Dropped() != 1 {
		t.Fatalf("dropped = %d, want 1", slot.Dropped())
	}
	for _, want := range []int{2, 3} {
		if got, _ := slot.Take(); got != want {
			t.Fatalf("took %d, want %d: the oldest frame goes first", got, want)
		}
	}
}

func TestASlotOfDepthZeroHoldsOne(t *testing.T) {
	slot := NewFrameSlotDepth[int](0)

	if !slot.TryPut(1) || slot.TryPut(2) {
		t.Fatal("depth 0 should behave as a slot of one")
	}
}
