package direct

import (
	"bytes"
	"encoding/binary"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"NanoKVM-Server/service/roommic"
	"NanoKVM-Server/service/stream/audio"

	"github.com/gorilla/websocket"
)

// roomCapture stands in for arecord; the test feeds it frames.
type roomCapture struct {
	frames chan []byte
	once   sync.Once
}

func (c *roomCapture) Start()                {}
func (c *roomCapture) Frames() <-chan []byte { return c.frames }
func (c *roomCapture) Stop()                 { c.once.Do(func() { close(c.frames) }) }

func withRoomManager(t *testing.T) chan *roomCapture {
	t.Helper()

	originalFile := roommic.SettingsFile
	roommic.SettingsFile = filepath.Join(t.TempDir(), "room-mic")
	originalManager := roomMics

	captures := make(chan *roomCapture, 4)
	roomMics = roommic.NewManager(func() bool { return true }, func() audio.Capture {
		capture := &roomCapture{frames: make(chan []byte, 4)}
		captures <- capture
		return capture
	})

	t.Cleanup(func() {
		roomMics.CloseAll("the test ended")
		roomMics = originalManager
		roommic.SettingsFile = originalFile
	})

	return captures
}

func TestTheRoomSwitchIsRefusedUntilAllowed(t *testing.T) {
	withRoomManager(t)
	room := newRoomSession(newClient(nil), "alice")

	if !room.handleControl(websocket.BinaryMessage, []byte{roomMicControl, 1}) {
		t.Fatal("the switch was not taken as one")
	}

	want := []byte{roomStateMessage, 0, 0, 0, roomErrorNotAllowed}
	if got := room.takeState(); !bytes.Equal(got, want) {
		t.Fatalf("state %v, want %v", got, want)
	}
	if roomMics.Live() {
		t.Fatal("the microphone opened without the setting")
	}
}

func TestTheRoomSwitchOpensAndClosesTheMicrophone(t *testing.T) {
	captures := withRoomManager(t)
	if err := roomMics.SetSettings(roommic.Settings{Allowed: true, Gain: roommic.DefaultGain}, "admin"); err != nil {
		t.Fatal(err)
	}

	room := newRoomSession(newClient(nil), "alice")
	room.set(true)

	if !roomMics.Live() {
		t.Fatal("switching on did not open the microphone")
	}
	if want := []byte{roomStateMessage, 1, 1, 1, roomErrorNone}; !bytes.Equal(room.takeState(), want) {
		t.Fatal("the state after switching on is wrong")
	}

	capture := <-captures
	capture.frames <- []byte{0xAA}

	deadline := time.Now().Add(2 * time.Second)
	var frames []audio.Frame
	for len(frames) == 0 && time.Now().Before(deadline) {
		frames = room.frames.popAll()
		time.Sleep(5 * time.Millisecond)
	}
	if len(frames) != 1 || !bytes.Equal(frames[0].Data, []byte{0xAA}) {
		t.Fatalf("frames %v, want the one captured", frames)
	}

	room.handleControl(websocket.BinaryMessage, []byte{roomMicControl, 0})
	if roomMics.Live() {
		t.Fatal("switching off left the microphone open")
	}

	// A message that is not the switch is left to the video controls.
	if room.handleControl(websocket.BinaryMessage, []byte{frameAckMessage, 0, 0, 0, 0, 0, 0, 0, 0}) {
		t.Fatal("a frame ack was taken for the switch")
	}
}

// A room frame is laid out like a host audio frame, with its own marker.
func TestARoomFrameCarriesItsOwnMarker(t *testing.T) {
	server, client, done := pair(t)
	defer done()

	if err := writeOpusFrame(server, roomAudioMessage, audio.Frame{Seq: 7, Data: []byte{1, 2}}); err != nil {
		t.Fatal(err)
	}

	_, data, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if data[0] != roomAudioMessage || binary.LittleEndian.Uint64(data[1:9]) != 7 || !bytes.Equal(data[9:], []byte{1, 2}) {
		t.Fatalf("message %v", data)
	}
}

// The direct stream's state message has no room for names: whoever reads it,
// it says live, listening, allowed and an error code, and nothing else.
func TestTheDirectRoomStateCarriesNoNames(t *testing.T) {
	withRoomManager(t)
	if err := roomMics.SetSettings(roommic.Settings{Allowed: true, Gain: roommic.DefaultGain}, "admin"); err != nil {
		t.Fatal(err)
	}

	alice := newRoomSession(newClient(nil), "alice")
	alice.set(true)
	defer alice.close()

	viewer := newRoomSession(newClient(nil), "bob")
	viewer.queueState()

	state := viewer.takeState()
	if len(state) != roomStateSize || bytes.Contains(state, []byte("alice")) {
		t.Fatalf("state %v", state)
	}
	if want := []byte{roomStateMessage, 1, 0, 1, roomErrorNone}; !bytes.Equal(state, want) {
		t.Fatalf("state %v, want %v", state, want)
	}
}
