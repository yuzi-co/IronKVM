package webrtc

import (
	"strings"
	"testing"

	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/nack"
	"github.com/pion/webrtc/v4"
)

// offerFor builds the offer a browser would be answered against. Nothing here
// touches the network: an offer is generated before ICE gathering starts.
func offerFor(t *testing.T) string {
	t.Helper()

	mediaEngine, err := createMediaEngine()
	if err != nil {
		t.Fatalf("create media engine: %s", err)
	}

	connection, err := createPeerConnection(nil, mediaEngine)
	if err != nil {
		t.Fatalf("create peer connection: %s", err)
	}
	defer func() { _ = connection.Close() }()

	track, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264},
		"video",
		"pion-video",
	)
	if err != nil {
		t.Fatalf("create track: %s", err)
	}

	if _, err := connection.AddTrack(track); err != nil {
		t.Fatalf("add track: %s", err)
	}

	offer, err := connection.CreateOffer(nil)
	if err != nil {
		t.Fatalf("create offer: %s", err)
	}

	return offer.SDP
}

// The negotiated feedback line is the precondition for everything below it. A
// responder that is never asked for a retransmission is dead weight, and the
// browser only asks when the answer advertises nack. RegisterDefaultCodecs
// supplies it today; hand-registering codecs later must not lose it.
func TestOfferNegotiatesNack(t *testing.T) {
	sdp := offerFor(t)

	if !strings.Contains(sdp, " nack\r\n") {
		t.Fatalf("offer does not advertise nack:\n%s", sdp)
	}
}

// pion accepts only powers of two between 1 and 32768, and it reports a bad
// size from the constructor rather than at the first retransmission. Catch a
// careless edit of the constant here instead of on the device.
func TestNackResponderSizeIsAccepted(t *testing.T) {
	if _, err := nack.NewResponderInterceptor(nack.ResponderSize(nackResponderSize)); err != nil {
		t.Fatalf("responder size %d rejected: %s", nackResponderSize, err)
	}
}

// The registry is built once per connection and a failure there must not be
// answered with a peer connection that silently has no interceptors.
func TestInterceptorRegistryIsPopulated(t *testing.T) {
	registry, err := createInterceptorRegistry()
	if err != nil {
		t.Fatalf("create interceptor registry: %s", err)
	}

	if registry == nil {
		t.Fatal("registry is nil")
	}

	// An empty registry is not an error and not a nil: Build hands back a NoOp,
	// which is precisely the silence this package used to ship. Assert against
	// that rather than against a count, because the count is an implementation
	// detail and the silence is the defect.
	built, err := registry.Build("test")
	if err != nil {
		t.Fatalf("build registry: %s", err)
	}

	if _, empty := built.(*interceptor.NoOp); empty {
		t.Fatal("registry built a NoOp, so no interceptor was registered")
	}
}
