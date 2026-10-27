// Package egresstest is a tap fed by a test instead of the capture loop and
// the audio hub, for the RTSP and SRT loopback tests.
package egresstest

import (
	"time"

	"NanoKVM-Server/common"
	"NanoKVM-Server/service/stream"
	"NanoKVM-Server/service/stream/audio"
	"NanoKVM-Server/service/stream/egress"
)

// Feed is what a test sends the tap. Frames sent with no reader wait until
// one subscribes.
type Feed struct {
	Video chan stream.H264Frame
	Audio chan audio.Frame
	Codec uint8
}

type video struct{ frames chan stream.H264Frame }

func (v video) Frames() <-chan stream.H264Frame { return v.frames }
func (video) Close()                            {}

type sound struct{ frames chan audio.Frame }

func (s sound) Frames() <-chan audio.Frame { return s.frames }
func (sound) Close()                       {}

// New returns a tap on a fresh feed carrying H.264.
func New() (*egress.Tap, *Feed) {
	feed := &Feed{
		Video: make(chan stream.H264Frame),
		Audio: make(chan audio.Frame),
		Codec: common.CodecH264,
	}
	tap := egress.NewTap(egress.Source{
		SubscribeVideo:  func(func() bool) egress.VideoSubscription { return video{feed.Video} },
		SubscribeAudio:  func() egress.AudioSubscription { return sound{feed.Audio} },
		Codec:           func() uint8 { return feed.Codec },
		RequestKeyframe: func() {},
		ReportReaders:   func(string, int, uint64) {},
	})
	return tap, feed
}

// SPS is a 1920x1080 Baseline sequence parameter set.
var SPS = []byte{
	0x67, 0x42, 0xc0, 0x28, 0xd9, 0x00, 0x78, 0x02,
	0x27, 0xe5, 0x84, 0x00, 0x00, 0x03, 0x00, 0x04,
	0x00, 0x00, 0x03, 0x00, 0xf0, 0x3c, 0x60, 0xc9,
	0x20,
}

// PPS is a picture parameter set to go with it.
var PPS = []byte{0x68, 0xce, 0x3c, 0x80}

// KeyFrame is an H.264 IDR access unit with its SPS and PPS, as the encoder
// writes one.
var KeyFrame = append(append(append(append(append([]byte{0, 0, 0, 1}, SPS...),
	0, 0, 0, 1), PPS...),
	0, 0, 0, 1), 0x65, 0x88, 0x84, 0x00, 0x33, 0xff)

// DeltaFrame is an H.264 P slice.
var DeltaFrame = []byte{0, 0, 0, 1, 0x41, 0x9a, 0x02, 0x03, 0x04}

// Pump sends a keyframe and then delta frames, a keyframe every gop frames,
// at fps until stop closes. It returns at once; the frames go out on a
// goroutine.
func (f *Feed) Pump(stop <-chan struct{}, fps, gop int) {
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(fps))
		defer ticker.Stop()
		for i := 0; ; i++ {
			frame := stream.H264Frame{Data: DeltaFrame, Result: 1}
			if i%gop == 0 {
				frame = stream.H264Frame{Data: KeyFrame, Result: 3, KeyFrame: true}
			}
			select {
			case <-stop:
				return
			case f.Video <- frame:
			}
			select {
			case <-stop:
				return
			case <-ticker.C:
			}
		}
	}()
}

// PumpAudio sends an Opus packet every 20 ms until stop closes.
func (f *Feed) PumpAudio(stop <-chan struct{}, packet []byte) {
	go func() {
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()
		for seq := uint64(0); ; seq++ {
			select {
			case <-stop:
				return
			case f.Audio <- audio.Frame{Seq: seq, Data: packet}:
			}
			select {
			case <-stop:
				return
			case <-ticker.C:
			}
		}
	}()
}
