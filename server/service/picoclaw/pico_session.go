package picoclaw

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"NanoKVM-Server/service/agent"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// picoGatewayMessage is one message of PicoClaw's pico protocol.
type picoGatewayMessage struct {
	Type      string         `json:"type"`
	ID        string         `json:"id,omitempty"`
	SessionID string         `json:"session_id,omitempty"`
	Timestamp int64          `json:"timestamp,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
}

// picoSession is an agent.Session over the pico channel WebSocket.
type picoSession struct {
	sessionID string
	conn      *websocket.Conn
	cfg       Config

	writeMu   sync.Mutex
	closeOnce sync.Once
	done      chan struct{}
}

var (
	_ agent.Session      = (*picoSession)(nil)
	_ agent.TurnReporter = (*picoSession)(nil)
)

func newPicoSession(sessionID string, conn *websocket.Conn, cfg Config) *picoSession {
	session := &picoSession{
		sessionID: sessionID,
		conn:      conn,
		cfg:       cfg,
		done:      make(chan struct{}),
	}
	readTimeout := time.Duration(cfg.ReadTimeoutMs) * time.Millisecond
	conn.SetReadLimit(int64(cfg.MaxMessageBytes))
	_ = conn.SetReadDeadline(time.Now().Add(readTimeout))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(readTimeout))
	})
	go session.pingLoop()
	return session
}

func (p *picoSession) Prompt(_ context.Context, prompt agent.Prompt) error {
	payload := map[string]any{"content": prompt.Text}
	if prompt.MaxSteps != nil {
		payload["max_steps"] = *prompt.MaxSteps
	}
	if prompt.MaxRuntimeMs != nil {
		payload["max_runtime_ms"] = *prompt.MaxRuntimeMs
	}
	requestID := prompt.RequestID
	if requestID == "" {
		requestID = uuid.NewString()
	}
	return p.write(picoGatewayMessage{
		Type:      "message.send",
		ID:        requestID,
		SessionID: p.sessionID,
		Timestamp: time.Now().UnixMilli(),
		Payload:   payload,
	})
}

// ReportsTurnDone says that PicoClaw ends every turn with turn.done: the
// pinned release, from the IronKVM fork, sends it for a finished, a failed and
// a cancelled turn alike.
func (p *picoSession) ReportsTurnDone() bool { return true }

// picoStopCommand is PicoClaw's command to stop the session's running turn.
const picoStopCommand = "/stop"

// Cancel stops the session's running turn. The pico protocol has no cancel
// message: PicoClaw answers an unknown type such as "message.cancel" with an
// error and lets the turn run on. It does honour a /stop chat command sent to
// a session that has a turn running: it aborts that turn, which then ends with
// a turn.done of status "canceled" that lists the /stop message's id with the
// prompt's. /stop stops the whole turn, so requestID only matters to the
// bridge's leases; the message gets an id of its own.
func (p *picoSession) Cancel(_ context.Context, _ string) error {
	return p.write(picoGatewayMessage{
		Type:      "message.send",
		ID:        uuid.NewString(),
		SessionID: p.sessionID,
		Timestamp: time.Now().UnixMilli(),
		Payload:   map[string]any{"content": picoStopCommand},
	})
}

func (p *picoSession) Next() (agent.Event, error) {
	for {
		messageType, data, err := p.conn.ReadMessage()
		if err != nil {
			return agent.Event{}, &agent.CloseError{
				Code:   closeCodeFromError(err),
				Reason: closeReasonFromError(err),
			}
		}
		if messageType == websocket.BinaryMessage {
			return agent.Event{}, &agent.CloseError{
				Code:   CloseCodeUpstreamClosed,
				Reason: "upstream sent unsupported binary message",
			}
		}
		if event, ok := picoEvent(data); ok {
			return event, nil
		}
	}
}

func (p *picoSession) Close() error {
	p.closeOnce.Do(func() {
		close(p.done)
		writeGatewayClose(p.conn, websocket.CloseNormalClosure, "relay closing")
		_ = p.conn.Close()
	})
	return nil
}

func (p *picoSession) write(message picoGatewayMessage) error {
	raw, err := json.Marshal(message)
	if err != nil {
		return err
	}
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	_ = p.conn.SetWriteDeadline(gatewayWriteDeadline(p.cfg))
	return p.conn.WriteMessage(websocket.TextMessage, raw)
}

func (p *picoSession) pingLoop() {
	interval := time.Duration(p.cfg.PingIntervalMs) * time.Millisecond
	if interval <= 0 {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-p.done:
			return
		case <-ticker.C:
			if err := p.conn.WriteControl(websocket.PingMessage, []byte("upstream"), gatewayWriteDeadline(p.cfg)); err != nil {
				return
			}
		}
	}
}

func gatewayWriteDeadline(cfg Config) time.Time {
	timeout := time.Duration(cfg.WriteTimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return time.Now().Add(timeout)
}

func writeGatewayClose(conn *websocket.Conn, code int, reason string) {
	if conn == nil {
		return
	}
	message := websocket.FormatCloseMessage(code, reason)
	_ = conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(2*time.Second))
}

// closeCodeFromError maps the end of the gateway connection to a close code.
// A gateway that just went away gives CloseCodeUpstreamClosed.
func closeCodeFromError(err error) int {
	if closeErr, ok := err.(*websocket.CloseError); ok {
		switch closeErr.Code {
		case websocket.CloseNormalClosure, websocket.CloseGoingAway,
			websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived:
			return CloseCodeUpstreamClosed
		}
		return closeErr.Code
	}
	return CloseCodeUpstreamClosed
}

func closeReasonFromError(err error) string {
	if err == nil {
		return "closed"
	}
	if closeErr, ok := err.(*websocket.CloseError); ok && closeErr.Text != "" {
		return closeErr.Text
	}
	return fmt.Sprintf("relay closed: %v", err)
}
