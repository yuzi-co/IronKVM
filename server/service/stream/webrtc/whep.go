package webrtc

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"NanoKVM-Server/authn"

	"github.com/pion/webrtc/v4"
	log "github.com/sirupsen/logrus"
)

// A WHEP session (ironkvm-dist#145) is a viewer that negotiated over HTTP
// rather than over the signalling socket: the player POSTs its offer and gets
// the answer back with every one of the board's candidates in it. It is the
// same Client the socket path builds, on the same manager, so it counts as a
// WebRTC viewer for capture, the idle shutdown and the metrics, and gets the
// same frames, keyframes and NACK handling. It differs in what it lacks: no
// socket, so no room microphone, clock or codec-change message, and no data
// channel.

var (
	// ErrWHEPOffer is an offer that cannot be answered: not SDP, or with
	// nothing this board can send.
	ErrWHEPOffer = errors.New("the offer cannot be answered")
	// ErrWHEPDataChannel is an offer with a data channel.
	ErrWHEPDataChannel = errors.New("WHEP carries no data channel")
	// ErrWHEPFragment is a PATCH body that is not an SDP fragment.
	ErrWHEPFragment = errors.New("the body is not an SDP fragment")
	// ErrWHEPRestart is a PATCH that asks for an ICE restart, which these
	// sessions do not do.
	ErrWHEPRestart = errors.New("ICE restarts are not supported")
)

var (
	// whepGatherTimeout bounds the wait for the board's own candidates. Host
	// candidates come at once; a STUN server that does not answer is what
	// takes the time, and the answer goes out with what there is by then.
	whepGatherTimeout = 5 * time.Second

	// whepConnectTimeout ends a session that never connects: a player that
	// took its answer and went away, or one that cannot reach the board.
	whepConnectTimeout = 30 * time.Second
)

// WHEPSession is one WHEP viewer.
type WHEPSession struct {
	client  *Client
	handler *SignalingHandler

	// ufrag is the player's ICE user fragment from its offer. A PATCH that
	// names another one is an ICE restart.
	ufrag string

	mutex     sync.Mutex
	connected bool
	reason    string
	done      chan struct{}
	closeOnce sync.Once
}

// NewWHEPSession answers a WHEP player's offer for the viewer named user. The
// answer carries the board's candidates (non-trickle): a player that trickles
// its own sends them with AddCandidates.
func NewWHEPSession(offer string, user string) (*WHEPSession, string, error) {
	if !strings.HasPrefix(strings.TrimSpace(offer), "v=0") {
		return nil, "", ErrWHEPOffer
	}
	// WHEP carries no data channel, and nothing a view-only token opened may
	// come near the board's HID. An offer with one is refused whole: RFC 9725
	// allows no partially successful answer (section 4.4.3).
	if strings.Contains(offer, "\nm=application ") {
		return nil, "", ErrWHEPDataChannel
	}

	iceServers := createICEServers()
	pc, udp, err := createPeerConnection(iceServers)
	if err != nil {
		return nil, "", err
	}

	client := NewClient(nil, pc)
	client.udp = udp
	client.user = user
	client.role = authn.RoleViewer
	if err := client.addTracks(false); err != nil {
		_ = pc.Close()
		return nil, "", err
	}

	s := &WHEPSession{
		client:  client,
		handler: NewSignalingHandler(client),
		ufrag:   sdpAttribute(offer, "ice-ufrag"),
		done:    make(chan struct{}),
	}

	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Debugf("whep: %s ICE %s", client.peerName(), state)
		s.handler.updateVideoStreamState(state)

		switch state {
		case webrtc.ICEConnectionStateConnected, webrtc.ICEConnectionStateCompleted:
			s.mutex.Lock()
			s.connected = true
			s.mutex.Unlock()
		case webrtc.ICEConnectionStateFailed, webrtc.ICEConnectionStateClosed:
			// Closing the peer connection from its own callback can wait on
			// the callback.
			go s.Close("connection " + state.String())
		}
	})

	answer, err := s.negotiate(pc, offer)
	if err != nil {
		s.Close("negotiation failed")
		return nil, "", err
	}

	time.AfterFunc(whepConnectTimeout, func() {
		s.mutex.Lock()
		connected := s.connected
		s.mutex.Unlock()
		if !connected {
			s.Close("never connected")
		}
	})

	return s, answer, nil
}

// negotiate applies the offer and returns the answer with the board's
// candidates in it.
func (s *WHEPSession) negotiate(pc *webrtc.PeerConnection, offer string) (string, error) {
	if err := pc.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: offer}); err != nil {
		log.Debugf("whep: offer refused: %s", err)
		return "", ErrWHEPOffer
	}

	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		log.Debugf("whep: no answer: %s", err)
		return "", ErrWHEPOffer
	}

	gathered := webrtc.GatheringCompletePromise(pc)
	if err := pc.SetLocalDescription(answer); err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), whepGatherTimeout)
	defer cancel()
	select {
	case <-gathered:
	case <-ctx.Done():
		log.Debugf("whep: answering with the candidates gathered in %s", whepGatherTimeout)
	}

	if !s.applyVideoParameters(pc) {
		log.Debugf("whep: the offer takes no video this board can send")
		return "", ErrWHEPOffer
	}

	return pc.LocalDescription().SDP, nil
}

// applyVideoParameters reads what the player negotiated for the video track:
// whether it takes the codec at all, and the header extension IDs. It reports
// false when the video was not accepted.
func (s *WHEPSession) applyVideoParameters(pc *webrtc.PeerConnection) bool {
	track := s.client.track
	if track == nil {
		return false
	}

	for _, transceiver := range pc.GetTransceivers() {
		sender := transceiver.Sender()
		if sender == nil || sender.Track() == nil || sender.Track().Kind() != webrtc.RTPCodecTypeVideo {
			continue
		}
		if transceiver.Mid() == "" {
			continue
		}

		params := sender.GetParameters()
		if len(params.Codecs) == 0 {
			return false
		}

		track.setAbsCaptureTimeExtensionID(0)
		track.disablePlayoutDelay()
		for _, ext := range params.HeaderExtensions {
			switch ext.URI {
			case playoutDelayExtensionURI:
				track.setPlayoutDelayExtensionID(uint8(ext.ID))
			case absCaptureTimeExtensionURI:
				track.setAbsCaptureTimeExtensionID(uint8(ext.ID))
			}
		}
		return true
	}

	return false
}

// AddCandidates applies a PATCH body, an SDP fragment with the player's ICE
// candidates (RFC 8840). A candidate the agent cannot use is dropped, as RFC
// 9725 asks; a fragment that names new ICE credentials is a restart and is
// refused.
func (s *WHEPSession) AddCandidates(fragment string) error {
	var mid string
	var candidates []webrtc.ICECandidateInit

	for _, raw := range strings.Split(fragment, "\n") {
		line := strings.TrimRight(raw, "\r")
		switch {
		case line == "":
		case strings.HasPrefix(line, "a=ice-ufrag:"):
			if ufrag := strings.TrimPrefix(line, "a=ice-ufrag:"); s.ufrag != "" && ufrag != s.ufrag {
				return ErrWHEPRestart
			}
		case strings.HasPrefix(line, "a=mid:"):
			mid = strings.TrimPrefix(line, "a=mid:")
		case strings.HasPrefix(line, "a=candidate:"):
			candidateMid := mid
			candidates = append(candidates, webrtc.ICECandidateInit{
				Candidate: strings.TrimPrefix(line, "a="),
				SDPMid:    &candidateMid,
			})
		case strings.HasPrefix(line, "a="), strings.HasPrefix(line, "m="):
			// ice-pwd, end-of-candidates, group, ice-options and the media
			// line itself carry nothing to apply here.
		default:
			return ErrWHEPFragment
		}
	}

	for _, candidate := range candidates {
		if err := s.client.video.AddICECandidate(candidate); err != nil {
			log.Debugf("whep: dropped candidate %q: %s", candidate.Candidate, err)
		}
	}

	return nil
}

// Close ends the session for reason: the viewer leaves the manager and the
// peer connection closes. It is safe to call more than once; the first
// reason is the one kept.
func (s *WHEPSession) Close(reason string) {
	s.closeOnce.Do(func() {
		s.mutex.Lock()
		s.reason = reason
		s.mutex.Unlock()

		s.handler.Close()
		s.client.Close()
		close(s.done)
		log.Debugf("whep: %s ended: %s", s.client.peerName(), reason)
	})
}

// Done is closed once the session has ended.
func (s *WHEPSession) Done() <-chan struct{} {
	return s.done
}

// Reason is why the session ended, empty while it lasts.
func (s *WHEPSession) Reason() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.reason
}

// ICEServers are the STUN and TURN servers a new session uses, for the Link
// headers of the answer (RFC 9725, section 4.6).
func ICEServers() []webrtc.ICEServer {
	return createICEServers()
}

// sdpAttribute returns the first value of a=name: in an SDP, or "".
func sdpAttribute(sdp, name string) string {
	prefix := "a=" + name + ":"
	for _, raw := range strings.Split(sdp, "\n") {
		line := strings.TrimRight(raw, "\r")
		if strings.HasPrefix(line, prefix) {
			return strings.TrimPrefix(line, prefix)
		}
	}
	return ""
}
