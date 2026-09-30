package webrtc

import (
	"strings"
	"testing"
)

// Every viewer holds one signaling socket for the whole session, and its write
// buffer is allocated for as long as the socket lives. Signaling is a handful of
// small JSON messages, so a video-sized buffer here is memory the board does
// not have to spare.
func TestSignalingWriteBufferStaysSmall(t *testing.T) {
	if upgrader.WriteBufferSize <= 0 || upgrader.WriteBufferSize > 16*1024 {
		t.Fatalf("signaling write buffer is %d bytes, want 1..16KB", upgrader.WriteBufferSize)
	}
}

// Every viewer's peer connection comes from one API rather than a fresh one
// per connection.
func TestPeerConnectionsShareOneAPI(t *testing.T) {
	first, err := sharedAPI()
	if err != nil {
		t.Fatalf("build api: %s", err)
	}

	second, err := sharedAPI()
	if err != nil {
		t.Fatalf("build api: %s", err)
	}

	if first != second {
		t.Fatal("sharedAPI built a second API")
	}
}

// Sharing the API must not make connections share state: closing one viewer's
// connection has to leave the next one able to negotiate the full feature set.
func TestClosingOneConnectionLeavesTheSharedAPIUsable(t *testing.T) {
	_ = offerFor(t) // builds and closes a connection

	sdp := offerFor(t)
	if !strings.Contains(sdp, " nack\r\n") {
		t.Fatalf("second offer lost nack:\n%s", sdp)
	}
	if !strings.Contains(sdp, playoutDelayExtensionURI) {
		t.Fatalf("second offer lost the playout delay extension:\n%s", sdp)
	}
}
