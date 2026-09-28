package mjpeg

import (
	"NanoKVM-Server/service/stream"
)

// Subscription is a viewer inside this process, such as the VNC server. It
// takes frames from the same capture loop as the HTTP viewers, so a browser and
// a subscriber that watch at once share one hardware read per frame, and
// duplicate suppression and the HDMI viewer count apply to it as well.
//
// It has no writer goroutine. The subscriber takes frames from Frames when it
// is ready for one. A frame it has not taken stays in the slot, and while every
// viewer holds a frame the capture loop reads nothing, so the capture rate
// follows the subscriber the same way it follows a browser on a slow link.
type Subscription struct {
	streamer *Streamer
	client   *client
}

// Subscribe adds a viewer inside this process to the MJPEG stream. The caller
// must Close it.
func Subscribe() *Subscription {
	return streamer.subscribe()
}

func (s *Streamer) subscribe() *Subscription {
	// The writer of an HTTP client closes done when it lets go of the
	// response. A subscription has no writer and no response, so there is
	// nothing to wait for.
	done := make(chan struct{})
	close(done)

	sub := &Subscription{
		streamer: s,
		client: &client{
			slot:   stream.NewFrameSlot[[]byte](),
			done:   done,
			failed: make(chan struct{}),
		},
	}
	s.addClient(sub, sub.client)

	return sub
}

// Frames hands out the newest frame the subscriber has not taken yet. The
// frame is shared with the capture loop and must not be modified. The channel
// closes with the subscription.
func (s *Subscription) Frames() <-chan []byte {
	return s.client.slot.Channel()
}

// Refresh makes the capture loop send its next frame even if it is identical to
// the last one, for a subscriber that has to show the picture again and holds
// no copy of it.
func (s *Subscription) Refresh() {
	s.streamer.forceNext.Store(true)
}

// Close removes the subscriber from the stream. It is safe to call more than
// once.
func (s *Subscription) Close() {
	s.streamer.removeClient(s)
}
