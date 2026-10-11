package common

import "time"

// VideoTimes is when the picture behind the last access unit ReadVideo
// returned was captured and encoded, on the wall clock (the clock RTCP sender
// reports use), from libkvm-v4l2's kvmv_video_times. Sipeed's library has no
// such call, so on slot A there are none.
type VideoTimes struct {
	// CaptureEnd is the capture buffer's timestamp. The SG2002 capture driver
	// takes it at the frame-done interrupt, so it is when the last line of
	// the frame arrived from the HDMI receiver.
	CaptureEnd time.Time
	// FrameTime is the source's frame time, 0 while the library has not
	// measured it yet.
	FrameTime time.Duration
	// EncodeStart and EncodeEnd bracket the encoder's work on the picture.
	EncodeStart time.Time
	EncodeEnd   time.Time
	// Returned is when the read handed the access unit over.
	Returned time.Time
}

// CaptureStart is about when the first line of the frame arrived: the frame
// time before CaptureEnd. It is what PiKVM counts its latency from ("the first
// byte of the HDMI frame"). The whole frame time is taken, blanking included,
// so this is early by the vertical blanking (about 0.7 ms at 1080p60) and a
// latency counted from it errs long, not short. Without a frame time it is
// CaptureEnd.
func (t VideoTimes) CaptureStart() time.Time {
	return t.CaptureEnd.Add(-t.FrameTime)
}
