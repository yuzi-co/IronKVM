package webrtc

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/pion/webrtc/v4"
)

// newSignalingTest builds the server's side of a session: a signalling
// handler over a real socket and peer connection, with the video track a
// viewer is answered against. It returns the handler and the offer a browser
// would send, as the message data.
func newSignalingTest(t *testing.T) (*SignalingHandler, *webrtc.PeerConnection, string) {
	t.Helper()

	serverConn, _ := newSignallingPair(t)

	peer, _, err := createPeerConnection(nil)
	if err != nil {
		t.Fatalf("create peer connection: %s", err)
	}
	t.Cleanup(func() { _ = peer.Close() })

	track, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion-video")
	if err != nil {
		t.Fatalf("create track: %s", err)
	}
	if _, err := peer.AddTrack(track); err != nil {
		t.Fatalf("add track: %s", err)
	}

	client := NewClient(serverConn, peer)
	client.track = &Track{video: track}
	handler := NewSignalingHandler(client)

	// The browser offers, as h264-webrtc.tsx does.
	browser, _, err := createPeerConnection(nil)
	if err != nil {
		t.Fatalf("create browser peer connection: %s", err)
	}
	t.Cleanup(func() { _ = browser.Close() })
	if _, err := browser.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo,
		webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly}); err != nil {
		t.Fatalf("add browser transceiver: %s", err)
	}
	offer, err := browser.CreateOffer(nil)
	if err != nil {
		t.Fatalf("create offer: %s", err)
	}
	data, err := json.Marshal(offer)
	if err != nil {
		t.Fatalf("marshal offer: %s", err)
	}

	return handler, peer, string(data)
}

func candidateMessage(t *testing.T, ip string) *Message {
	t.Helper()

	mid := "0"
	data, err := json.Marshal(webrtc.ICECandidateInit{
		Candidate: fmt.Sprintf("candidate:1 1 udp 2130706431 %s 50000 typ host", ip),
		SDPMid:    &mid,
	})
	if err != nil {
		t.Fatalf("marshal candidate: %s", err)
	}
	return &Message{Event: "video-candidate", Data: string(data)}
}

// waitForRemoteCandidate reports whether the peer connection's ICE agent comes
// to know a remote candidate at ip. The agent takes candidates on its own
// goroutine, so this polls.
func waitForRemoteCandidate(peer *webrtc.PeerConnection, ip string) bool {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		for _, stat := range peer.GetStats() {
			candidate, ok := stat.(webrtc.ICECandidateStats)
			if ok && candidate.Type == webrtc.StatsTypeRemoteCandidate && candidate.IP == ip {
				return true
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// The browser can surface its first candidates before it sends the offer, and
// both travel on the one socket. A candidate that arrives before the offer is
// held, not dropped, and goes to the ICE agent once the offer is applied.
func TestACandidateBeforeTheOfferIsAppliedWithIt(t *testing.T) {
	handler, peer, offer := newSignalingTest(t)

	if err := handler.HandleMessage(candidateMessage(t, "192.0.2.1")); err != nil {
		t.Fatalf("candidate before the offer: %s", err)
	}
	if err := handler.HandleMessage(&Message{Event: "video-offer", Data: offer}); err != nil {
		t.Fatalf("offer: %s", err)
	}

	if !waitForRemoteCandidate(peer, "192.0.2.1") {
		t.Fatal("the candidate sent before the offer never reached the ICE agent")
	}
}

func TestACandidateAfterTheOfferIsAppliedAtOnce(t *testing.T) {
	handler, peer, offer := newSignalingTest(t)

	if err := handler.HandleMessage(&Message{Event: "video-offer", Data: offer}); err != nil {
		t.Fatalf("offer: %s", err)
	}
	if err := handler.HandleMessage(candidateMessage(t, "192.0.2.2")); err != nil {
		t.Fatalf("candidate after the offer: %s", err)
	}

	if !waitForRemoteCandidate(peer, "192.0.2.2") {
		t.Fatal("the candidate sent after the offer never reached the ICE agent")
	}
}

// A viewer that sends candidates and never an offer cannot make the server
// hold an unbounded list.
func TestHeldCandidatesAreBounded(t *testing.T) {
	handler, _, _ := newSignalingTest(t)

	for i := 0; i < maxPendingCandidates+10; i++ {
		_ = handler.HandleMessage(candidateMessage(t, fmt.Sprintf("192.0.2.%d", i%250+1)))
	}

	if got := len(handler.pendingCandidates); got != maxPendingCandidates {
		t.Fatalf("holding %d candidates, want the bound %d", got, maxPendingCandidates)
	}
}
