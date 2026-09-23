package direct

import (
	"encoding/binary"
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

// writeAudio sends one frame as a single binary message.
func writeAudio(conn *websocket.Conn, frame audio.Frame) error {
	writer, err := conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return err
	}

	var header [audioHeaderSize]byte
	header[0] = audioMessage
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
