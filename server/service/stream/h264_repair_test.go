package stream

import "testing"

func newTestSubscription() *H264Subscription {
	return &H264Subscription{slot: NewFrameSlot[H264Frame]()}
}

func dataFrame(keyFrame bool) H264Frame {
	result := 1
	if keyFrame {
		result = 3
	}

	return H264Frame{Data: []byte{0x00, 0x00, 0x00, 0x01}, Result: result, KeyFrame: keyFrame}
}

func statusFrame() H264Frame {
	return H264Frame{Result: -1}
}

// take empties the slot, standing in for the delivery goroutine, and reports
// whether anything was there.
func take(subscription *H264Subscription) bool {
	select {
	case <-subscription.slot.Channel():
		return true
	default:
		return false
	}
}

// The defect this repairs. A refused frame is a gap, and every frame after a
// gap is coded against a picture the decoder never received, so delivering them
// gives the viewer a smeared image rather than a late one.
func TestADroppedFrameHoldsTheStreamUntilTheNextKeyframe(t *testing.T) {
	subscription := newTestSubscription()

	// The client is still holding this one.
	subscription.deliver(dataFrame(false))

	// So this one is refused, and that is the gap.
	subscription.deliver(dataFrame(false))

	if !take(subscription) {
		t.Fatal("the first frame never reached the slot")
	}
	if take(subscription) {
		t.Fatal("two frames reached a slot that holds one")
	}

	// The client has caught up, but the stream has not been repaired yet.
	subscription.deliver(dataFrame(false))
	if take(subscription) {
		t.Fatal("a frame after the gap was delivered before a keyframe")
	}

	subscription.deliver(dataFrame(true))
	if !take(subscription) {
		t.Fatal("the keyframe that ends the repair was not delivered")
	}

	// And the stream runs again.
	subscription.deliver(dataFrame(false))
	if !take(subscription) {
		t.Fatal("the stream did not resume after the keyframe")
	}
}

// A status frame carries no picture, and each path reports it under its own
// capture mode. A repair that silenced it would turn a second of held video
// into a stream that reports nothing at all.
func TestAStatusFrameSurvivesARepair(t *testing.T) {
	subscription := newTestSubscription()

	subscription.deliver(dataFrame(false))
	subscription.deliver(dataFrame(false))

	if !take(subscription) {
		t.Fatal("the first frame never reached the slot")
	}

	subscription.deliver(statusFrame())
	if !take(subscription) {
		t.Fatal("a status frame was withheld by the repair")
	}

	// The repair is still running: the status frame does not end it.
	subscription.deliver(dataFrame(false))
	if take(subscription) {
		t.Fatal("a status frame ended the repair")
	}
}

// A status frame that the slot refuses is not a lost picture, so it must not
// start a repair and cost the viewer the rest of the GOP.
func TestARefusedStatusFrameDoesNotStartARepair(t *testing.T) {
	subscription := newTestSubscription()

	subscription.deliver(dataFrame(false))
	subscription.deliver(statusFrame())

	if !take(subscription) {
		t.Fatal("the first frame never reached the slot")
	}

	subscription.deliver(dataFrame(false))
	if !take(subscription) {
		t.Fatal("a refused status frame started a repair")
	}
}

// A keyframe ends the repair only if it is actually delivered. If the client is
// still behind when it arrives, the stream stays held rather than resuming on a
// keyframe the decoder never saw.
func TestAKeyframeTheClientCannotTakeDoesNotEndTheRepair(t *testing.T) {
	subscription := newTestSubscription()

	subscription.deliver(dataFrame(false))
	subscription.deliver(dataFrame(false))

	// The slot is still full, so the keyframe is refused too.
	subscription.deliver(dataFrame(true))

	if !take(subscription) {
		t.Fatal("the first frame never reached the slot")
	}
	if take(subscription) {
		t.Fatal("the refused keyframe reached the slot")
	}

	subscription.deliver(dataFrame(false))
	if take(subscription) {
		t.Fatal("the stream resumed on a keyframe that was never delivered")
	}

	subscription.deliver(dataFrame(true))
	if !take(subscription) {
		t.Fatal("the next keyframe did not end the repair")
	}
}

// ironkvm-dist#73: the delivery goroutine late by a tick or two. With one frame
// of room the second frame was refused and the rest of the GOP lost, a 470 ms
// freeze at 60 fps for a delay of 17. The subscription's slot rides that out.
func TestAPathLateByAFewFramesLosesNone(t *testing.T) {
	subscription := &H264Subscription{slot: NewFrameSlotDepth[H264Frame](h264SubscriptionDepth)}

	subscription.deliver(dataFrame(true))
	for i := 1; i < h264SubscriptionDepth; i++ {
		subscription.deliver(dataFrame(false))
	}
	if subscription.repairing {
		t.Fatal("a path behind by fewer frames than the slot holds started a repair")
	}

	for i := 0; i < h264SubscriptionDepth; i++ {
		if !take(subscription) {
			t.Fatalf("frame %d of %d was lost", i+1, h264SubscriptionDepth)
		}
	}

	// Caught up: the next frame goes straight through.
	subscription.deliver(dataFrame(false))
	if !take(subscription) {
		t.Fatal("the frame after the late spell was not delivered")
	}
}

// A path that stays behind still loses frames to the next keyframe.
func TestAPathFurtherBehindThanTheSlotStillRepairs(t *testing.T) {
	subscription := &H264Subscription{slot: NewFrameSlotDepth[H264Frame](h264SubscriptionDepth)}

	for i := 0; i <= h264SubscriptionDepth; i++ {
		subscription.deliver(dataFrame(i == 0))
	}
	if !subscription.repairing {
		t.Fatal("a refused frame did not start a repair")
	}
}

func TestTheSourceGivesEachPathTheDeeperSlot(t *testing.T) {
	withCapture(t, func(uint16, uint16, uint16) ([]byte, int) { return nil, -1 })

	subscription := newH264Source().subscribe(func() bool { return false })
	defer subscription.Close()

	if got := cap(subscription.Frames()); got != h264SubscriptionDepth {
		t.Fatalf("the subscription's slot holds %d frames, want %d", got, h264SubscriptionDepth)
	}
}
