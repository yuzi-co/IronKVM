package webrtc

import (
	"encoding/json"
	"errors"

	"NanoKVM-Server/service/roommic"
	"NanoKVM-Server/service/stream/audio"

	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
	log "github.com/sirupsen/logrus"
)

// roomMicEvent is the signalling event for the room microphone, both ways.
// The browser sends "on" or "off"; the server answers, and tells every viewer
// with the room track whenever the microphone opens or closes, with a
// roomMicState.
const roomMicEvent = "room-mic"

// roomStreamID is the media stream the room track belongs to. The browser
// tells the two audio tracks apart by it.
const roomStreamID = "room-mic"

// roomSSRC is a placeholder like the others; pion rewrites it per binding.
const roomSSRC = 0x1234ABCF

// roomMics is roommic.Shared in production. A variable so a test can supply a
// manager with no arecord behind it.
var roomMics = roommic.Shared

// roomMicState is what a viewer is told.
type roomMicState struct {
	// Live is true while the microphone is open for anyone.
	Live bool `json:"live"`
	// Listening is true while it is open for this viewer.
	Listening bool `json:"listening"`
	// ListenerCount is how many accounts have it on.
	ListenerCount int `json:"listenerCount"`
	// Listeners names who has it on, for an administrator only.
	Listeners []string `json:"listeners,omitempty"`
	// Allowed is the administrator's setting.
	Allowed bool `json:"allowed"`
	// Error says why switching it on failed: "not-allowed" or "unavailable".
	Error string `json:"error,omitempty"`
}

// newRoomTrack creates the room microphone's track. Opus in SDP is always
// declared with two channels; the encoder sends one and the decoder plays it
// in both.
func newRoomTrack() (*webrtc.TrackLocalStaticRTP, error) {
	return webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{
			MimeType:  webrtc.MimeTypeOpus,
			ClockRate: audioClockRate,
			Channels:  audio.Channels,
		},
		"room-mic",
		roomStreamID,
	)
}

// hasRoomTrack reports whether this session negotiated the room microphone.
func (c *Client) hasRoomTrack() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.track != nil && c.track.room != nil
}

// setRoomMic switches the room microphone on or off for this viewer.
func (c *Client) setRoomMic(on bool) {
	if !c.hasRoomTrack() {
		return
	}

	c.roomMutex.Lock()
	switch {
	case on && c.room == nil:
		listener, err := roomMics.Open(c.user, "webrtc")
		switch {
		case errors.Is(err, roommic.ErrNotAllowed):
			c.roomError = "not-allowed"
		case err != nil:
			c.roomError = "unavailable"
		default:
			c.room = listener
			go c.forwardRoom(listener)
		}
	case !on && c.room != nil:
		listener := c.room
		c.room = nil
		c.roomMutex.Unlock()
		// Close notifies every watcher, this one included, which takes
		// roomMutex, so it runs outside it.
		listener.Close()
		return
	}
	c.roomMutex.Unlock()

	// A refusal changes nothing the manager would announce, so this viewer is
	// answered here. An open has been announced already, and once more does
	// no harm.
	c.sendRoomState()
}

// closeRoomMic lets go of the microphone when the session ends.
func (c *Client) closeRoomMic() {
	c.roomMutex.Lock()
	listener := c.room
	c.room = nil
	c.roomMutex.Unlock()

	if listener != nil {
		listener.Close()
	}
}

// sendRoomState tells this viewer the microphone's state. The lock orders the
// messages: each reads the state while holding it, so the last one sent is
// never older than one sent before it.
func (c *Client) sendRoomState() {
	if c.ws == nil {
		return
	}

	c.roomMutex.Lock()
	defer c.roomMutex.Unlock()

	data, err := json.Marshal(c.roomStateLocked())
	if err != nil {
		return
	}

	_ = c.WriteMessage(roomMicEvent, string(data))
}

// roomStateLocked builds the state this viewer is told and clears the pending
// error. The names of the listeners reach an administrator only. The caller
// holds roomMutex.
func (c *Client) roomStateLocked() roomMicState {
	status := roomMics.Status().ForViewer(c.role)
	state := roomMicState{
		Live:          status.Live,
		Listening:     c.room != nil,
		ListenerCount: status.ListenerCount,
		Listeners:     status.Listeners,
		Allowed:       status.Allowed,
		Error:         c.roomError,
	}
	c.roomError = ""

	return state
}

// forwardRoom packetizes this viewer's microphone frames and writes them to
// its room track. Each listener has its own subscription, so the packetizer is
// this viewer's own and the hub's buffer stands between the capture and a
// slow connection.
func (c *Client) forwardRoom(listener *roommic.Listener) {
	packetizer := rtp.NewPacketizer(
		rtpMTU,
		audioPayloadType,
		roomSSRC,
		&codecs.OpusPayloader{},
		rtp.NewRandomSequencer(),
		audioClockRate,
	)

	var next uint64
	started := false

	for frame := range listener.Frames() {
		if frame.State != audio.StateUnknown {
			continue
		}

		// A gap moves the RTP clock across it, as for the host's audio.
		if started && frame.Seq > next {
			packetizer.SkipSamples(uint32(frame.Seq-next) * audio.SamplesPerFrame)
		}
		next = frame.Seq + 1
		started = true

		c.mutex.Lock()
		track := c.track
		c.mutex.Unlock()
		if track == nil || track.room == nil {
			continue
		}

		if err := writeOpusPackets(track.room, packetizer.Packetize(frame.Data, audio.SamplesPerFrame)); err != nil {
			log.Debugf("room microphone write to %s failed: %s", c.ws.RemoteAddr(), err)
			break
		}
	}

	// The frames ended: this viewer switched off, the session ended, an
	// administrator disallowed the microphone, or the capture failed. In the
	// last cases the listener is still held here.
	c.roomMutex.Lock()
	own := c.room == listener
	if own {
		c.room = nil
	}
	c.roomMutex.Unlock()

	if own {
		listener.Close()
	}
}

// watchRoomMic sends this viewer every change of the microphone's state, and
// its state now. It returns the function that stops it.
func (c *Client) watchRoomMic() func() {
	if !c.hasRoomTrack() {
		return func() {}
	}

	cancel := roomMics.Watch(func() {
		// Off the notifying goroutine: a slow socket must not hold up the
		// viewer that switched the microphone on or off.
		go c.sendRoomState()
	})
	go c.sendRoomState()

	return cancel
}
