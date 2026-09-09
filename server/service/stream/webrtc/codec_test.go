package webrtc

import (
	"testing"

	"NanoKVM-Server/common"

	"github.com/pion/webrtc/v4"
)

// The board has one hardware encoder, so the codec is global and a WebRTC
// session cannot choose its own. What a session can do is record which codec it
// negotiated, so the sender never packetizes HEVC into a track that told the
// peer it carries H.264.

func TestBothCodecsHaveAMimeType(t *testing.T) {
	cases := map[uint8]string{
		common.CodecH264: webrtc.MimeTypeH264,
		common.CodecH265: webrtc.MimeTypeH265,
	}

	for codec, want := range cases {
		if got := mimeTypeForCodec(codec); got != want {
			t.Fatalf("mimeTypeForCodec(%d) = %q, want %q", codec, got, want)
		}
	}
}

func TestAnUnknownCodecFallsBackToH264(t *testing.T) {
	// The setting is validated before it is stored, so this should be
	// unreachable. It still must not answer with an empty mime type, which
	// pion would reject at track creation and which would read as a bug
	// somewhere else entirely.
	if got := mimeTypeForCodec(200); got != webrtc.MimeTypeH264 {
		t.Fatalf("mimeTypeForCodec(200) = %q, want the H.264 fallback", got)
	}
}

func TestEachCodecGetsItsOwnPacketizer(t *testing.T) {
	m := NewWebRTCManager()

	h264 := m.packetizerFor(common.CodecH264)
	h265 := m.packetizerFor(common.CodecH265)

	if h264 == nil || h265 == nil {
		t.Fatal("a codec had no packetizer")
	}
	if h264 == h265 {
		t.Fatal("both codecs share one packetizer, so HEVC would be cut up as H.264")
	}
}

func TestAPacketizerIsReusedRatherThanRebuiltPerFrame(t *testing.T) {
	// The packetizer carries the RTP sequence number. A fresh one per frame
	// would restart the sequence and the peer would treat the stream as
	// permanently reordered.
	m := NewWebRTCManager()

	first := m.packetizerFor(common.CodecH265)
	second := m.packetizerFor(common.CodecH265)

	if first != second {
		t.Fatal("packetizerFor built a second packetizer for the same codec")
	}
}
