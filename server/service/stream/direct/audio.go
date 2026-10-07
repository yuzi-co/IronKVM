package direct

import (
	"sync"

	"NanoKVM-Server/service/stream/audio"

	"github.com/gorilla/websocket"
)

// audioMessage marks an audio message on the direct stream's websocket.
//
// Byte 0 of a video message is the keyframe flag, which is 0 or 1, so a value
// video never sends is all the browser needs to tell the two apart. Audio is
// sent only to a browser that asked for it with ?audio=1, so an older viewer
// that reads byte 0 as a keyframe flag never sees one.
const audioMessage byte = 0x10

// audioHeaderSize is the marker and a little-endian uint64 sequence number,
// then the Opus packet. The sequence is the hub's: a gap is frames lost on the
// way, and the browser keeps its clock by playing silence for them.
const audioHeaderSize = 9

// audioStateMessage marks a notice of what capture is doing: the marker, then
// one byte of audio.State. It lets the browser say that the host plays nothing
// rather than leave the operator to guess at silence.
//
// It is two bytes long on purpose. Every viewer that asks for audio drops a
// message shorter than audioHeaderSize, so one built before this existed
// ignores it instead of reading it as audio or video.
const audioStateMessage byte = 0x11

// audioQueueFrames bounds what waits for the writer: 160 ms of audio.
const audioQueueFrames = 8

// audioHub is audio.Shared in production. A variable so a test can supply a
// hub with no arecord behind it.
var audioHub = audio.Shared

// audioQueue holds the frames waiting for the client's writer. It keeps the
// newest: the video queue drops to the next keyframe, and audio has none, so a
// full queue drops from the front.
type audioQueue struct {
	mutex  sync.Mutex
	frames []audio.Frame
	closed bool
}

func (q *audioQueue) push(frame audio.Frame) bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.closed {
		return false
	}
	if len(q.frames) >= audioQueueFrames {
		copy(q.frames, q.frames[1:])
		q.frames = q.frames[:len(q.frames)-1]
	}
	q.frames = append(q.frames, frame)

	return true
}

func (q *audioQueue) popAll() []audio.Frame {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.frames) == 0 {
		return nil
	}
	frames := q.frames
	q.frames = nil

	return frames
}

func (q *audioQueue) close() {
	q.mutex.Lock()
	q.closed = true
	q.frames = nil
	q.mutex.Unlock()
}

// offerAudio queues a frame and wakes the writer, which shares its wake-up
// with the video queue.
func (c *client) offerAudio(frame audio.Frame) {
	if c.audio.push(frame) {
		c.queue.wakeWriter()
	}
}

// forwardAudio hands every frame of a subscription to the client until the
// subscription closes.
func forwardAudio(sub *audio.Subscription, c *client) {
	for frame := range sub.Frames() {
		c.offerAudio(frame)
	}
}

// writeAudio sends one frame as a single binary message, or a state notice as
// an audioStateMessage.
func writeAudio(conn *websocket.Conn, frame audio.Frame) error {
	if frame.State != audio.StateUnknown {
		return conn.WriteMessage(websocket.BinaryMessage, []byte{audioStateMessage, byte(frame.State)})
	}

	return writeOpusFrame(conn, audioMessage, frame)
}
