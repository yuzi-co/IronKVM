package webrtc

import (
	"testing"

	"NanoKVM-Server/common"
)

// WebRTC declares H.264 in its SDP and packetizes with codecs.H264Payloader.
// An HEVC access unit through that produces a stream no viewer can decode, and
// it produces it silently: the connection succeeds and the picture never
// arrives. There is one hardware encoder, so the codec is global and this path
// cannot opt out of it. It refuses instead.

func TestWebRTCServesH264(t *testing.T) {
	if !codecIsDeliverable(common.CodecH264) {
		t.Fatal("webrtc refused H.264, which is the codec it is built for")
	}
}

func TestWebRTCRefusesH265RatherThanSendingGarbage(t *testing.T) {
	if codecIsDeliverable(common.CodecH265) {
		t.Fatal("webrtc accepted H.265; the H264Payloader would emit a stream " +
			"no viewer can decode, and nothing would report it")
	}
}
