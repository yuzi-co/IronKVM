package webrtc

import "testing"

// Every viewer holds one signaling socket for the whole session, and its write
// buffer is allocated for as long as the socket lives. Signaling is a handful of
// small JSON messages, so a video-sized buffer here is memory the board does
// not have to spare.
func TestSignalingWriteBufferStaysSmall(t *testing.T) {
	if upgrader.WriteBufferSize <= 0 || upgrader.WriteBufferSize > 16*1024 {
		t.Fatalf("signaling write buffer is %d bytes, want 1..16KB", upgrader.WriteBufferSize)
	}
}
