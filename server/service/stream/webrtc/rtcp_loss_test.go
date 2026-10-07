package webrtc

import (
	"testing"

	"github.com/pion/rtcp"
)

// fakeSent is a sentLog for rtcpLoss.
type fakeSent struct {
	last, key     uint16
	sent, keySent bool
	// gaveUp is a frame this viewer's queue gave up: never sent to it.
	gaveUp [2]uint16
}

func (f fakeSent) sentSequence() (uint16, bool) { return f.last, f.sent }
func (f fakeSent) keySequence() (uint16, bool)  { return f.key, f.keySent }
func (f fakeSent) wasSent(seq uint16) bool {
	return int16(seq-f.gaveUp[0]) < 0 || int16(seq-f.gaveUp[1]) > 0
}

func TestRTCPLossReadsWhatTheViewerReports(t *testing.T) {
	nack := func(seq uint16) rtcp.Packet {
		return &rtcp.TransportLayerNack{Nacks: []rtcp.NackPair{{PacketID: seq}}}
	}
	// 1000 sent last, the last keyframe started at 600, 995 to 997 given up.
	view := fakeSent{last: 1000, sent: true, key: 600, keySent: true, gaveUp: [2]uint16{995, 997}}
	noKey := view
	noKey.keySent = false
	wrapped := fakeSent{last: 10, sent: true, gaveUp: [2]uint16{20, 20}}
	none := fakeSent{gaveUp: [2]uint16{1, 0}}
	far := view
	far.last, far.key = 2000, 1500

	cases := []struct {
		name    string
		packets []rtcp.Packet
		view    fakeSent
		loss    bool
		picture bool
	}{
		{"receiver report", []rtcp.Packet{&rtcp.ReceiverReport{}}, view, false, false},
		{"PLI", []rtcp.Packet{&rtcp.ReceiverReport{}, &rtcp.PictureLossIndication{}}, view, true, true},
		{"FIR", []rtcp.Packet{&rtcp.FullIntraRequest{}}, view, true, true},
		{"NACK the responder still holds", []rtcp.Packet{nack(990)}, view, true, false},
		{"NACK past the responder's history, after the last keyframe", []rtcp.Packet{nack(700)}, view, true, true},
		{"NACK past the responder's history, before the last keyframe", []rtcp.Packet{nack(500)}, view, true, false},
		{"NACK past the responder's history, no keyframe yet", []rtcp.Packet{nack(500)}, noKey, true, true},
		{"NACK past what the track remembers", []rtcp.Packet{nack(2000 - sentRingSize)}, far, true, false},
		{"NACK for a frame this viewer was never sent", []rtcp.Packet{nack(996)}, view, false, false},
		{"NACK across the wrap", []rtcp.Packet{nack(65530)}, wrapped, true, false},
		{"NACK before anything was sent", []rtcp.Packet{nack(5)}, none, false, false},
	}

	for _, c := range cases {
		loss, picture := rtcpLoss(c.packets, c.view)
		if loss != c.loss || picture != c.picture {
			t.Errorf("%s: loss %t, picture %t; want %t, %t", c.name, loss, picture, c.loss, c.picture)
		}
	}
}

func TestTheTrackRemembersWhatItSent(t *testing.T) {
	track, _ := newTestTrack(5)
	if _, ok := track.sentSequence(); ok {
		t.Fatal("a sequence before anything was sent")
	}
	if _, ok := track.keySequence(); ok {
		t.Fatal("a keyframe before anything was sent")
	}

	track.markKey(1)
	if err := track.writePackets(sharedFrame()); err != nil {
		t.Fatal(err)
	}

	if seq, ok := track.sentSequence(); !ok || seq != 2 {
		t.Fatalf("sent %d (%t), want 2", seq, ok)
	}
	if seq, ok := track.keySequence(); !ok || seq != 1 {
		t.Fatalf("keyframe at %d (%t), want 1", seq, ok)
	}
	if !track.wasSent(1) || !track.wasSent(2) || track.wasSent(3) || track.wasSent(1+sentRingSize) {
		t.Fatal("the track misremembers what it sent")
	}
}
