package webrtc

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"
)

// The capture time is an NTP timestamp, seconds since 1900 in the high half
// and the fraction of a second in the low half, then a zero clock offset.
func TestAbsCaptureTimeIsNTP(t *testing.T) {
	at := time.Unix(1760000000, 250_000_000)

	data := absCaptureTime(at)
	if len(data) != 16 {
		t.Fatalf("data is %d bytes, want 16", len(data))
	}

	ntp := binary.BigEndian.Uint64(data)
	if seconds := ntp >> 32; seconds != 1760000000+2208988800 {
		t.Fatalf("seconds %d", seconds)
	}
	if fraction := uint32(ntp); fraction != 1<<30 {
		t.Fatalf("fraction %#x, want a quarter (%#x)", fraction, uint32(1<<30))
	}
	if offset := binary.BigEndian.Uint64(data[8:]); offset != 0 {
		t.Fatalf("clock offset %d, want 0", offset)
	}

	if absCaptureTime(time.Time{}) != nil {
		t.Fatal("the zero time should carry no extension")
	}
}

// Only a frame's first packet carries the capture time, only when the viewer
// negotiated it, and the playout delay stays on every packet.
func TestWritePacketsPutsTheCaptureTimeOnTheFirstPacket(t *testing.T) {
	capture := absCaptureTime(time.Unix(1760000000, 0))

	track, writer := newTestTrack(3)
	if err := track.writePackets(sharedFrame(), capture); err != nil {
		t.Fatalf("writePackets failed: %s", err)
	}
	for _, packet := range writer.packets() {
		if packet.Header.GetExtension(9) != nil {
			t.Fatal("a viewer that did not negotiate it got the capture time")
		}
	}

	track, writer = newTestTrack(3)
	track.setAbsCaptureTimeExtensionID(9)
	if err := track.writePackets(sharedFrame(), capture); err != nil {
		t.Fatalf("writePackets failed: %s", err)
	}

	got := writer.packets()
	if len(got) != 2 {
		t.Fatalf("wrote %d packets", len(got))
	}
	if !bytes.Equal(got[0].Header.GetExtension(9), capture) {
		t.Fatal("the first packet lacks the capture time")
	}
	if got[1].Header.GetExtension(9) != nil {
		t.Fatal("the second packet carries the capture time too")
	}
	for i, packet := range got {
		if packet.Header.GetExtension(3) == nil {
			t.Fatalf("packet %d lost the playout delay", i)
		}
	}

	// It must fit the one-byte header form alongside the playout delay.
	if _, err := got[0].Marshal(); err != nil {
		t.Fatalf("the first packet does not marshal: %s", err)
	}

	// And a frame without a capture time carries none.
	track, writer = newTestTrack(3)
	track.setAbsCaptureTimeExtensionID(9)
	if err := track.writePackets(sharedFrame(), nil); err != nil {
		t.Fatalf("writePackets failed: %s", err)
	}
	if writer.packets()[0].Header.GetExtension(9) != nil {
		t.Fatal("a frame with no capture time carries one")
	}
}
