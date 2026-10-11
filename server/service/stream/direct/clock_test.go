package direct

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestAClockRequestIsAnsweredWithTheStreamClock(t *testing.T) {
	// The status bar sets frame timestamps against this answer, so it has to
	// be on the clock the frames are stamped with, and in the format the
	// worker reads: the marker, then little-endian microseconds.
	conn, browser, done := pair(t)
	defer done()

	c := newClient(conn)
	c.start()
	defer func() { c.close(); c.wait() }()

	before := clockMicros()
	c.handleControl(websocket.BinaryMessage, []byte{clockRequestMessage})

	_ = browser.SetReadDeadline(time.Now().Add(5 * time.Second))
	messageType, payload, err := browser.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read: %s", err)
	}
	after := clockMicros()

	if messageType != websocket.BinaryMessage || len(payload) != 9 || payload[0] != clockReplyMessage {
		t.Fatalf("received %v, want the clock marker and eight bytes", payload)
	}
	clock := int64(binary.LittleEndian.Uint64(payload[1:]))
	if clock < before || clock > after {
		t.Fatalf("clock %d, want between %d and %d", clock, before, after)
	}
}

func TestAClockRequestOfTheWrongLengthIsIgnored(t *testing.T) {
	c := newClient(nil)
	c.handleControl(websocket.BinaryMessage, []byte{clockRequestMessage, 0})
	if c.clockAsked.Load() {
		t.Fatal("a two-byte message is not a clock request")
	}
	c.handleControl(websocket.TextMessage, []byte{clockRequestMessage})
	if c.clockAsked.Load() {
		t.Fatal("a text message is not a clock request")
	}
}

func TestTheClockReplyIsNoOtherMessage(t *testing.T) {
	// The worker tells messages apart by their first byte and their length.
	for _, marker := range []byte{0, 1, audioMessage, audioStateMessage, roomAudioMessage, roomStateMessage} {
		if marker == clockReplyMessage {
			t.Fatalf("the clock reply shares marker %#x", marker)
		}
	}
	for _, request := range []byte{frameAckMessage, streamResyncMessage, roomMicControl} {
		if request == clockRequestMessage {
			t.Fatalf("the clock request shares marker %#x", request)
		}
	}
}
