package webrtc

import "time"

// maxClockSkip bounds the time one gap may move the RTP clock on. A longer
// gap is a stream that stopped and started again, which needs no catching up.
const maxClockSkip = 10 * time.Second

// frameClock keeps a video packetizer's RTP clock in step with capture time
// when frames go missing (ironkvm-dist#35).
//
// Packetize advances the RTP timestamp by one frame time per frame sent. When
// the encoder delivers fewer frames than the capture loop asks for (libkvm-v4l2
// leaves pictures out to hold the bitrate on a busy screen, and either library
// can fall behind the asked rate), the RTP clock then runs slow against the wall
// clock. The browser reads that as every frame arriving later than the last,
// and its jitter buffer grows the delay to match. Skipping the RTP clock over
// the frame times nobody sent keeps it on the capture clock.
//
// Only whole frame times are skipped: a frame that comes a little late or early
// keeps the regular step, as before.
type frameClock struct {
	lastUs  int64
	started bool
}

// skip returns the samples to skip before packetizing a frame captured at
// timestampUs, when one frame time is frameTime and clockRate samples make a
// second. It assumes that frame is then packetized and the one before it
// advanced the clock by one frame time.
func (c *frameClock) skip(timestampUs int64, frameTime time.Duration, clockRate uint32) uint32 {
	last, started := c.lastUs, c.started
	c.lastUs, c.started = timestampUs, true

	if !started || frameTime <= 0 || timestampUs <= last {
		return 0
	}

	gap := time.Duration(timestampUs-last) * time.Microsecond
	if gap > maxClockSkip {
		return 0
	}

	// Frame times that passed with no frame, rounded to the nearest.
	missed := (gap + frameTime/2) / frameTime
	if missed <= 1 {
		return 0
	}

	return uint32((missed - 1) * frameTime * time.Duration(clockRate) / time.Second)
}
