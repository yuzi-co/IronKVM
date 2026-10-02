package picoclaw

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"NanoKVM-Server/service/agent"

	"github.com/gorilla/websocket"
)

// fakePicoGateway accepts one pico channel connection and hands it to the
// test.
func fakePicoGateway(t *testing.T) (*websocket.Conn, <-chan *websocket.Conn) {
	t.Helper()
	conns := make(chan *websocket.Conn, 1)
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		conns <- conn
	}))
	t.Cleanup(server.Close)

	client, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	return client, conns
}

func testPicoConfig() Config {
	return Config{ReadTimeoutMs: 5000, WriteTimeoutMs: 1000, PingIntervalMs: 60000, MaxMessageBytes: 1 << 20}
}

func TestPicoSessionSpeaksThePicoProtocol(t *testing.T) {
	client, conns := fakePicoGateway(t)
	session := newPicoSession("sess", client, testPicoConfig())
	defer session.Close()
	gateway := <-conns

	maxSteps, maxRuntime := 8, 60000
	if err := session.Prompt(context.Background(), agent.Prompt{RequestID: "r1", Text: "hello", MaxSteps: &maxSteps, MaxRuntimeMs: &maxRuntime}); err != nil {
		t.Fatal(err)
	}
	var sent picoGatewayMessage
	if err := gateway.ReadJSON(&sent); err != nil {
		t.Fatal(err)
	}
	if sent.Type != "message.send" || sent.ID != "r1" || sent.SessionID != "sess" ||
		sent.Payload["content"] != "hello" || sent.Payload["max_steps"] != float64(8) || sent.Payload["max_runtime_ms"] != float64(60000) {
		t.Fatalf("PicoClaw got %+v", sent)
	}

	if err := session.Cancel(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
	if err := gateway.ReadJSON(&sent); err != nil {
		t.Fatal(err)
	}
	if sent.Type != "message.cancel" || sent.SessionID != "sess" {
		t.Fatalf("PicoClaw got %+v", sent)
	}

	for _, message := range []string{
		`{"type":"typing.start"}`,
		`{"type":"typing.stop"}`,
		`{"type":"message.create","payload":{"message_id":"m1","content":"Done"}}`,
		`{"type":"turn.done","payload":{"request_ids":["r1"],"status":"ok"}}`,
	} {
		if err := gateway.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			t.Fatal(err)
		}
	}
	var types []agent.EventType
	for range 3 {
		event, err := session.Next()
		if err != nil {
			t.Fatal(err)
		}
		types = append(types, event.Type)
	}
	want := []agent.EventType{agent.EventTurnStarted, agent.EventAgentMessage, agent.EventTurnDone}
	if !reflect.DeepEqual(types, want) {
		t.Fatalf("events %v, want %v", types, want)
	}

	_ = gateway.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"))
	_, err := session.Next()
	var closeErr *agent.CloseError
	if !errors.As(err, &closeErr) || closeErr.Code != CloseCodeUpstreamClosed || closeErr.Reason != "bye" {
		t.Fatalf("end of the gateway = %v, want close %d bye", err, CloseCodeUpstreamClosed)
	}
}

func TestPicoSessionRefusesBinaryMessages(t *testing.T) {
	client, conns := fakePicoGateway(t)
	session := newPicoSession("sess", client, testPicoConfig())
	defer session.Close()
	gateway := <-conns
	_ = gateway.WriteMessage(websocket.BinaryMessage, []byte{1})
	_, err := session.Next()
	var closeErr *agent.CloseError
	if !errors.As(err, &closeErr) || closeErr.Code != CloseCodeUpstreamClosed {
		t.Fatalf("binary message = %v", err)
	}
}

func TestPicoSessionCloseUnblocksNext(t *testing.T) {
	client, _ := fakePicoGateway(t)
	session := newPicoSession("sess", client, testPicoConfig())
	done := make(chan error, 1)
	go func() {
		_, err := session.Next()
		done <- err
	}()
	_ = session.Close()
	_ = session.Close()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("Next returned no error after Close")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Close did not unblock Next")
	}
}
