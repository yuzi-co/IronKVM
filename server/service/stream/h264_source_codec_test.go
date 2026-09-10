package stream

import (
	"sync"
	"testing"

	"NanoKVM-Server/common"
)

// The codec reaches the encoder from the screen settings, because there is one
// hardware encoder and it cannot serve two codecs at once. These hold the
// capture loop to passing what the operator chose.

type codecCall struct {
	codec   uint8
	bitRate uint16
	gop     uint8
	fps     uint8
}

// withCodecCapture records what the loop asked the encoder for.
func withCodecCapture(t *testing.T, calls *[]codecCall, mutex *sync.Mutex) {
	t.Helper()

	original := readVideo
	t.Cleanup(func() { readVideo = original })
	readVideo = func(width uint16, height uint16, codec uint8, bitRate uint16, gop uint8, fps uint8) ([]byte, int) {
		mutex.Lock()
		*calls = append(*calls, codecCall{codec: codec, bitRate: bitRate, gop: gop, fps: fps})
		mutex.Unlock()

		return []byte{0, 0, 0, 1, 0x67}, 3
	}
}

func TestTheCaptureLoopAsksForTheConfiguredCodec(t *testing.T) {
	for _, codec := range []uint8{common.CodecH264, common.CodecH265} {
		var mutex sync.Mutex
		var calls []codecCall
		withCodecCapture(t, &calls, &mutex)

		common.SetScreen("codec", int(codec))

		subscription := SubscribeH264(func() bool { return true })
		waitFor(t, "a capture read", func() bool {
			mutex.Lock()
			defer mutex.Unlock()
			return len(calls) > 0
		})
		subscription.Close()

		mutex.Lock()
		got := calls[0].codec
		mutex.Unlock()

		if got != codec {
			t.Fatalf("capture loop asked for codec %d, want %d", got, codec)
		}
	}
}

func TestTheCaptureLoopLeavesGopAndFpsToTheModule(t *testing.T) {
	// 0 means "whatever set_h264_gop and set_capture_fps last set", which is
	// what kvmv_read_img has always done. Passing the stored values here would
	// silently override SetGop, which has its own API and stores nothing.
	var mutex sync.Mutex
	var calls []codecCall
	withCodecCapture(t, &calls, &mutex)

	subscription := SubscribeH264(func() bool { return true })
	waitFor(t, "a capture read", func() bool {
		mutex.Lock()
		defer mutex.Unlock()
		return len(calls) > 0
	})
	subscription.Close()

	mutex.Lock()
	first := calls[0]
	mutex.Unlock()

	if first.gop != 0 || first.fps != 0 {
		t.Fatalf("capture loop passed gop=%d fps=%d, want 0 and 0 so the module keeps deciding",
			first.gop, first.fps)
	}
}
