package direct

import (
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"NanoKVM-Server/service/stream/audio"

	"github.com/gorilla/websocket"
)

func TestAudioQueueKeepsTheNewestFramesInOrder(t *testing.T) {
	// A browser that stopped reading is further behind than 160 ms, and the
	// oldest audio is the least worth sending. The video queue drops to the
	// next keyframe; audio has none, so it drops from the front.
	q := &audioQueue{}
	for i := 0; i < audioQueueFrames+3; i++ {
		q.push(audio.Frame{Seq: uint64(i)})
	}

	frames := q.popAll()
	if len(frames) != audioQueueFrames {
		t.Fatalf("held %d frames, want %d", len(frames), audioQueueFrames)
	}
	for i, f := range frames {
		if want := uint64(i + 3); f.Seq != want {
			t.Fatalf("frame %d has sequence %d, want %d", i, f.Seq, want)
		}
	}
	if len(q.popAll()) != 0 {
		t.Fatal("popAll left frames behind")
	}
}

func TestAudioQueueRefusesFramesOnceClosed(t *testing.T) {
	q := &audioQueue{}
	q.close()
	if q.push(audio.Frame{Seq: 1}) {
		t.Fatal("a closed queue accepted a frame")
	}
}

// pair returns a server-side connection and the client that dialled it.
func pair(t *testing.T) (*websocket.Conn, *websocket.Conn, func()) {
	t.Helper()
	serverConn := make(chan *websocket.Conn, 1)
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		serverConn <- conn
	}))

	client, _, err := websocket.DefaultDialer.Dial("ws"+server.URL[len("http"):], nil)
	if err != nil {
		server.Close()
		t.Fatalf("failed to dial: %s", err)
	}
	conn := <-serverConn

	return conn, client, func() {
		_ = client.Close()
		_ = conn.Close()
		server.Close()
	}
}

func TestAnAudioFrameReachesTheBrowserWithItsSequence(t *testing.T) {
	// Byte 0 of a video message is the keyframe flag, 0 or 1. Audio takes a
	// value video never sends, so the browser tells the two apart from one
	// byte, and a browser that never asked for audio never sees one.
	conn, browser, done := pair(t)
	defer done()

	c := newClient(conn)
	c.start()
	defer func() { c.close(); c.wait() }()

	c.offerAudio(audio.Frame{Seq: 513, Data: []byte("opus")})

	_ = browser.SetReadDeadline(time.Now().Add(5 * time.Second))
	messageType, payload, err := browser.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read: %s", err)
	}
	if messageType != websocket.BinaryMessage {
		t.Fatalf("message type is %d, want binary", messageType)
	}
	if len(payload) != audioHeaderSize+4 || payload[0] != audioMessage {
		t.Fatalf("received %v, want the audio marker, a sequence and the packet", payload)
	}
	if seq := binary.LittleEndian.Uint64(payload[1:9]); seq != 513 {
		t.Fatalf("sequence %d, want 513", seq)
	}
	if string(payload[9:]) != "opus" {
		t.Fatalf("packet %q, want %q", payload[9:], "opus")
	}
}

func TestForwardAudioStopsWhenTheSubscriptionCloses(t *testing.T) {
	frames := make(chan []byte, 4)
	hub := audio.NewHubWith(func() audio.Capture { return &chanCapture{frames} }, func() bool { return true })
	sub := hub.Subscribe()

	c := &client{audio: &audioQueue{}, queue: newFrameQueue(defaultQueueFrames, defaultQueueBytes)}
	finished := make(chan struct{})
	go func() { forwardAudio(sub, c); close(finished) }()

	frames <- []byte("a")
	var got []audio.Frame
	deadline := time.Now().Add(time.Second)
	for len(got) == 0 && time.Now().Before(deadline) {
		got = c.audio.popAll()
		time.Sleep(5 * time.Millisecond)
	}
	if len(got) != 1 || string(got[0].Data) != "a" {
		t.Fatalf("forwarded %v, want the one frame", got)
	}

	sub.Close()
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("forwardAudio kept running after its subscription closed")
	}
}

type chanCapture struct{ frames chan []byte }

func (c *chanCapture) Start()                {}
func (c *chanCapture) Stop()                 {}
func (c *chanCapture) Frames() <-chan []byte { return c.frames }
