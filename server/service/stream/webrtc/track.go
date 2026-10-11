package webrtc

import (
	"encoding/binary"
	"time"

	"github.com/pion/rtp"
	log "github.com/sirupsen/logrus"
)

// defaultPlayoutDelayExtensionID is used until signalling tells us what the
// peer negotiated.
const defaultPlayoutDelayExtensionID uint8 = 5

// audioClockRate is the RTP clock for Opus. Opus always uses 48000 in RTP,
// whatever rate the encoder was created with.
const audioClockRate = 48000

// rtpExtensionProfile is the one-byte header extension profile (RFC 5285).
const rtpExtensionProfile = 0xBEDE

// playoutDelayExtensionURI has to match on both sides of the negotiation.
const playoutDelayExtensionURI = "http://www.webrtc.org/experiments/rtp-hdrext/playout-delay"

// absCaptureTimeExtensionURI is the absolute capture time header extension
// (http://www.webrtc.org/experiments/rtp-hdrext/abs-capture-time), which PiKVM
// measures its latency with. Each frame's first packet carries the time the
// frame's first line reached the capture (common.VideoTimes.CaptureStart), on
// the board's wall clock, the clock its RTCP sender reports use. Chrome
// offers it only when the page asks (RTCRtpTransceiver
// setHeaderExtensionsToNegotiate; it is "stopped" by default), and then
// reports it per frame as captureTimestamp in
// RTCRtpReceiver.getSynchronizationSources().
const absCaptureTimeExtensionURI = "http://www.webrtc.org/experiments/rtp-hdrext/abs-capture-time"

// ntpEpochOffset is the seconds from 1900 (NTP) to 1970 (Unix).
const ntpEpochOffset = 2208988800

// absCaptureTime is the extension's data for a frame captured at t: the
// capture time as an NTP timestamp (UQ32.32), then the estimated capture clock
// offset, 0 because the board both captures and sends (so the browser's
// senderCaptureTimeOffset is 0 rather than undefined). 16 bytes, the most a
// one-byte header extension carries. nil for the zero time.
func absCaptureTime(t time.Time) []byte {
	if t.IsZero() {
		return nil
	}

	ns := t.UnixNano()
	seconds := uint64(ns/int64(time.Second)) + ntpEpochOffset
	fraction := (uint64(ns%int64(time.Second)) << 32) / uint64(time.Second)

	data := make([]byte, 16)
	binary.BigEndian.PutUint64(data, seconds<<32|fraction)

	return data
}

// playoutDelayExtensionData asks the receiver not to buffer ahead of playback.
// It is the same for every client and never changes, so it is marshalled once
// instead of once per track.
var playoutDelayExtensionData = marshalPlayoutDelay()

func marshalPlayoutDelay() []byte {
	data, err := (&rtp.PlayoutDelayExtension{MinDelay: 0, MaxDelay: 0}).Marshal()
	if err != nil {
		log.Errorf("failed to marshal the playout delay extension: %v", err)
		return nil
	}

	return data
}

// rtpWriter is the part of a peer connection's track that a Track writes to.
// Naming it keeps the packet handling testable without a peer connection.
type rtpWriter interface {
	WriteRTP(*rtp.Packet) error
}

// setPlayoutDelayExtensionID records what signalling negotiated. It runs on the
// websocket goroutine while the capture goroutine may already be writing frames
// for this client, so the value is held atomically.
func (t *Track) setPlayoutDelayExtensionID(id uint8) {
	t.extensionID.Store(uint32(id))
}

func (t *Track) playoutDelayExtensionID() uint8 {
	if id := t.extensionID.Load(); id != 0 {
		return uint8(id)
	}

	return defaultPlayoutDelayExtensionID
}

// setAbsCaptureTimeExtensionID records what signalling negotiated for the
// capture time, 0 for nothing: then no packet carries it.
func (t *Track) setAbsCaptureTimeExtensionID(id uint8) {
	t.absCaptureID.Store(uint32(id))
}

// writePackets sends one frame to this client's peer connection.
//
// The frame is packetized once for everyone, so these packets are shared. pion
// rewrites the SSRC and payload type of whatever header it is handed, and every
// client writes on its own goroutine, so each packet goes out through a copy.
// Only the header is copied - the payload is the encoded frame itself, which
// nobody modifies.
//
// capture, when not nil, is the frame's abs-capture-time data (absCaptureTime).
// Only the first packet carries it: the receiver takes the rest of the frame's
// packets, which share its RTP timestamp, from that one, and the extension
// costs 17 bytes a frame rather than a packet.
func (t *Track) writePackets(packets []*rtp.Packet, capture []byte) error {
	id := t.playoutDelayExtensionID()
	absID := uint8(t.absCaptureID.Load())

	for i, source := range packets {
		packet := rtp.Packet{Header: source.Header, Payload: source.Payload}

		packet.Header.Extension = true
		packet.Header.ExtensionProfile = rtpExtensionProfile
		// The struct copy still points at the source's extension slice, and
		// appending to that would reach back into the shared packet.
		packet.Header.Extensions = nil

		if err := packet.Header.SetExtension(id, playoutDelayExtensionData); err != nil {
			log.Errorf("failed to set extension: %v", err)
			return err
		}

		if i == 0 && absID != 0 && capture != nil {
			if err := packet.Header.SetExtension(absID, capture); err != nil {
				log.Errorf("failed to set the capture time extension: %v", err)
				return err
			}
		}

		if err := t.video.WriteRTP(&packet); err != nil {
			log.Errorf("failed to write RTP: %v", err)
			return err
		}

		v := uint32(packet.SequenceNumber) | sentValid
		t.sentRing[packet.SequenceNumber%sentRingSize].Store(v)
		t.sent.Store(v)
	}

	return nil
}

// sentValid marks t.sent and the entries of t.sentRing as holding a sequence
// number.
const sentValid = 1 << 16

// sentRingSize is how many of the latest sequence numbers a track remembers
// writing. It is more than the NACK responder keeps (nackResponderSize).
const sentRingSize = 1024

// markKey records that the frame about to be written, starting at seq, is a
// keyframe. Only the writer calls it.
func (t *Track) markKey(seq uint16) {
	t.keySent.Store(uint32(seq) | sentValid)
}

// keySequence is the first sequence number of the last keyframe written, and
// false before the first.
func (t *Track) keySequence() (uint16, bool) {
	v := t.keySent.Load()

	return uint16(v), v&sentValid != 0
}

// wasSent reports whether seq is among the last sentRingSize sequence numbers
// and was written to this viewer. The packetizer is shared, so the frames a
// viewer's queue gave up leave holes in its sequence: the viewer NACKs them,
// but nothing was lost on the network, and its queue has already asked for a
// keyframe.
func (t *Track) wasSent(seq uint16) bool {
	return t.sentRing[seq%sentRingSize].Load() == uint32(seq)|sentValid
}

// sentSequence is the newest video sequence number written to this viewer,
// and false before the first. The RTCP reader compares a NACK with it.
func (t *Track) sentSequence() (uint16, bool) {
	v := t.sent.Load()

	return uint16(v), v&sentValid != 0
}

// writeAudioPackets sends one audio frame to this client's peer connection.
//
// The packets are shared between clients, so each one goes out through a copy.
// Audio carries no playout delay extension: that hint is about video rendering
// and the browser's own jitter buffer handles the rest.
func (t *Track) writeAudioPackets(packets []*rtp.Packet) error {
	return writeOpusPackets(t.audio, packets)
}

// writeOpusPackets sends packets that may be shared with other clients, each
// through a copy of its header, with no header extension.
func writeOpusPackets(w rtpWriter, packets []*rtp.Packet) error {
	for _, source := range packets {
		packet := rtp.Packet{Header: source.Header, Payload: source.Payload}

		packet.Header.Extension = false
		packet.Header.Extensions = nil

		if err := w.WriteRTP(&packet); err != nil {
			log.Errorf("failed to write audio RTP: %v", err)
			return err
		}
	}

	return nil
}
