package direct

import (
	"encoding/binary"
	"errors"
	"sync"

	"NanoKVM-Server/service/roommic"
	"NanoKVM-Server/service/stream/audio"

	"github.com/gorilla/websocket"
)

// The room microphone on the direct stream's websocket. It is offered only to
// a browser that asked with ?room=1, on a kernel with the onboard card.
const (
	// roomMicControl is the browser's switch: the marker, then 1 for on or 0
	// for off.
	roomMicControl byte = 4

	// roomAudioMessage is one microphone frame, laid out like audioMessage:
	// the marker, a little-endian uint64 sequence, the Opus packet.
	roomAudioMessage byte = 0x12

	// roomStateMessage is the microphone's state for this viewer: the marker,
	// then live, listening and allowed as 0 or 1, then a reason the last
	// switch-on failed (roomError*).
	roomStateMessage byte = 0x13
	roomStateSize         = 5
)

// Why switching the microphone on failed, the last byte of a state message.
const (
	roomErrorNone        byte = 0
	roomErrorNotAllowed  byte = 1
	roomErrorUnavailable byte = 2
)

// roomMics is roommic.Shared in production. A variable so a test can supply a
// manager with no arecord behind it.
var roomMics = roommic.Shared

// roomSession is one direct viewer's part in the room microphone.
type roomSession struct {
	client *client
	user   string

	// frames waits for the writer like the host's audio does.
	frames audioQueue

	mutex    sync.Mutex
	listener *roommic.Listener
	// pending is the newest state not yet written; a newer one replaces it.
	pending []byte
	// failure is the reason the last switch-on failed, sent once.
	failure byte
}

func newRoomSession(c *client, user string) *roomSession {
	return &roomSession{client: c, user: user}
}

// set switches the microphone on or off for this viewer.
func (r *roomSession) set(on bool) {
	r.mutex.Lock()
	switch {
	case on && r.listener == nil:
		listener, err := roomMics.Open(r.user, "direct")
		switch {
		case errors.Is(err, roommic.ErrNotAllowed):
			r.failure = roomErrorNotAllowed
		case err != nil:
			r.failure = roomErrorUnavailable
		default:
			r.listener = listener
			go r.forward(listener)
		}
	case !on && r.listener != nil:
		listener := r.listener
		r.listener = nil
		r.mutex.Unlock()
		listener.Close()
		return
	}
	r.mutex.Unlock()

	r.queueState()
}

// forward hands the listener's frames to the writer until they end.
func (r *roomSession) forward(listener *roommic.Listener) {
	for frame := range listener.Frames() {
		if frame.State != audio.StateUnknown {
			continue
		}
		if r.frames.push(frame) {
			r.client.queue.wakeWriter()
		}
	}

	// The frames ended without this viewer switching off: an administrator
	// disallowed the microphone or the capture failed.
	r.mutex.Lock()
	own := r.listener == listener
	if own {
		r.listener = nil
	}
	r.mutex.Unlock()

	if own {
		listener.Close()
	}
}

// queueState puts this viewer's current state in front of the writer.
func (r *roomSession) queueState() {
	// The status is read under the lock, so the last state queued is never
	// older than one queued before it.
	r.mutex.Lock()
	status := roomMics.Status()
	state := []byte{roomStateMessage, flag(status.Live), flag(r.listener != nil), flag(status.Allowed), r.failure}
	r.failure = roomErrorNone
	r.pending = state
	r.mutex.Unlock()

	r.client.queue.wakeWriter()
}

// takeState returns the state waiting to be written, or nil.
func (r *roomSession) takeState() []byte {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	state := r.pending
	r.pending = nil
	return state
}

// close lets go of the microphone when the connection ends.
func (r *roomSession) close() {
	r.frames.close()

	r.mutex.Lock()
	listener := r.listener
	r.listener = nil
	r.mutex.Unlock()

	if listener != nil {
		listener.Close()
	}
}

// write sends what is waiting: the state first, then the frames.
func (r *roomSession) write(conn *websocket.Conn, deadline func() error) (bool, error) {
	wrote := false

	if state := r.takeState(); state != nil {
		if err := deadline(); err != nil {
			return wrote, err
		}
		if err := conn.WriteMessage(websocket.BinaryMessage, state); err != nil {
			return wrote, err
		}
		wrote = true
	}

	for _, frame := range r.frames.popAll() {
		if err := deadline(); err != nil {
			return wrote, err
		}
		if err := writeOpusFrame(conn, roomAudioMessage, frame); err != nil {
			return wrote, err
		}
		wrote = true
	}

	return wrote, nil
}

// handleControl takes the browser's switch. It reports whether the message
// was one.
func (r *roomSession) handleControl(messageType int, data []byte) bool {
	if messageType != websocket.BinaryMessage || len(data) != 2 || data[0] != roomMicControl {
		return false
	}

	r.set(data[1] == 1)
	return true
}

func flag(b bool) byte {
	if b {
		return 1
	}
	return 0
}

// writeOpusFrame sends one frame with its marker and sequence number.
func writeOpusFrame(conn *websocket.Conn, marker byte, frame audio.Frame) error {
	writer, err := conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return err
	}

	var header [audioHeaderSize]byte
	header[0] = marker
	binary.LittleEndian.PutUint64(header[1:], frame.Seq)
	if _, err := writer.Write(header[:]); err != nil {
		_ = writer.Close()
		return err
	}
	if _, err := writer.Write(frame.Data); err != nil {
		_ = writer.Close()
		return err
	}

	return writer.Close()
}

// watch queues a state for every change of the microphone, and one now. It
// returns the function that stops it.
func (r *roomSession) watch() func() {
	// Off the notifying goroutine: set holds the session's lock while it opens
	// the microphone, and the manager notifies from there.
	cancel := roomMics.Watch(func() { go r.queueState() })
	r.queueState()

	return cancel
}
