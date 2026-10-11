package webrtc

import (
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// The clock answer must be on the clock the capture times are on, or the
// offset the browser measures with it corrects nothing: the same instant
// reads the same in both.
func TestClockReplyMatchesAbsCaptureTime(t *testing.T) {
	at := time.Unix(1760000000, 123_456_000)

	reply, err := strconv.ParseInt(clockReply(at, ""), 10, 64)
	if err != nil {
		t.Fatalf("reply %q is not an integer: %v", clockReply(at, ""), err)
	}

	ntp := binary.BigEndian.Uint64(absCaptureTime(at))
	seconds := int64(ntp>>32) - ntpEpochOffset
	micros := seconds*1_000_000 + int64((ntp&0xffffffff)*1_000_000>>32)

	if diff := reply - micros; diff < -1 || diff > 1 {
		t.Fatalf("clock reply %d µs, capture time %d µs: %d µs apart", reply, micros, diff)
	}
}

// A clock request on the signalling socket is answered on it at once, with
// the board's clock between the request leaving and the answer arriving.
func TestSignallingAnswersAClockRequest(t *testing.T) {
	upgrade := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	accepted := make(chan *websocket.Conn, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade websocket: %v", err)
			return
		}
		accepted <- conn
	}))
	defer server.Close()

	browser, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer browser.Close()

	serverConn := <-accepted
	defer serverConn.Close()

	handler := NewSignalingHandler(NewClient(serverConn, nil))

	sent := time.Now()
	if err := handler.HandleMessage(&Message{Event: clockEvent, Data: "17"}); err != nil {
		t.Fatalf("handle: %v", err)
	}

	var answer Message
	if err := browser.ReadJSON(&answer); err != nil {
		t.Fatalf("read: %v", err)
	}
	received := time.Now()

	if answer.Event != clockEvent {
		t.Fatalf("answer event %q, want %q", answer.Event, clockEvent)
	}
	micros, id, _ := strings.Cut(answer.Data, " ")
	if id != "17" {
		t.Fatalf("answer data %q does not carry the id back", answer.Data)
	}
	board, err := strconv.ParseInt(micros, 10, 64)
	if err != nil {
		t.Fatalf("answer data %q: %v", answer.Data, err)
	}
	if board < sent.UnixMicro() || board > received.UnixMicro() {
		t.Fatalf("board clock %d outside the round trip [%d, %d]", board, sent.UnixMicro(), received.UnixMicro())
	}
}

func TestClockReplyCarriesOnlyAShortID(t *testing.T) {
	at := time.UnixMicro(1760000000123456)
	if got := clockReply(at, "3"); got != "1760000000123456 3" {
		t.Fatalf("got %q", got)
	}
	if got := clockReply(at, strings.Repeat("x", maxClockID+1)); got != "1760000000123456" {
		t.Fatalf("a long id came back: %q", got)
	}
}
