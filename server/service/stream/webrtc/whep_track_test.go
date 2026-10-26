package webrtc

import "testing"

// A WHEP player that did not negotiate the playout delay extension gets no
// extension at all. The browser path falls back to id 5, which a player that
// mapped 5 to something else would misread (ironkvm-dist#145).
func TestWritePacketsWithoutPlayoutDelay(t *testing.T) {
	track, writer := newTestTrack(0)
	track.disablePlayoutDelay()

	if err := track.writePackets(sharedFrame(), nil); err != nil {
		t.Fatalf("writePackets: %s", err)
	}
	for _, packet := range writer.written {
		if packet.Header.Extension || len(packet.Header.Extensions) != 0 {
			t.Fatalf("packet %d carries extensions %+v", packet.SequenceNumber, packet.Header.Extensions)
		}
	}
}

// The capture time still goes out when the player negotiated it, without the
// playout delay.
func TestWritePacketsWithOnlyTheCaptureTime(t *testing.T) {
	track, writer := newTestTrack(0)
	track.disablePlayoutDelay()
	track.setAbsCaptureTimeExtensionID(4)

	if err := track.writePackets(sharedFrame(), make([]byte, 16)); err != nil {
		t.Fatalf("writePackets: %s", err)
	}
	first := writer.written[0]
	if first.Header.GetExtension(4) == nil || first.Header.GetExtension(defaultPlayoutDelayExtensionID) != nil {
		t.Fatalf("first packet extensions %+v, want only the capture time", first.Header.Extensions)
	}
	if writer.written[1].Header.Extension {
		t.Fatal("the second packet carries an extension")
	}
}

// A WHEP client has no websocket. Nothing that would write to one may panic.
func TestClientWithoutASocket(t *testing.T) {
	c := NewClient(nil, nil)
	if err := c.WriteMessage("heartbeat", ""); err != nil {
		t.Fatalf("WriteMessage without a socket: %s", err)
	}
	if got := c.peerName(); got == "" {
		t.Fatal("no name for a client without a socket")
	}
}
