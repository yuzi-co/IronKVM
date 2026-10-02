package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"NanoKVM-Server/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// Host is the KVM side of a chat session: who may chat, the input lock and
// what has to happen when a session ends. The server's AI service implements
// it; it does not depend on which agent runs.
type Host interface {
	// Admit decides whether a chat session may open. An *Error refuses it
	// with that error's status and code.
	Admit(sessionID string) error
	// Renew extends the session's hold on the AI input lock. False means
	// another session took it over.
	Renew(sessionID string) bool
	// Opened is called once the browser socket is up.
	Opened(sessionID string)
	// Closed is called once for every admitted session when it ends.
	// opened says whether Opened was called for it.
	Closed(sessionID string, closeCode int, reason string, opened bool)
	// RelayConfig gives the browser socket's limits. It is read after the
	// agent session opens, so it can follow the agent's configuration.
	RelayConfig() RelayConfig
}

// RelayConfig holds the browser socket's limits.
type RelayConfig struct {
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PingInterval    time.Duration
	MaxMessageBytes int64
}

type sessionState int

const (
	stateConnecting sessionState = iota
	stateActive
	stateClosed
)

type chatSession struct {
	id string

	mu     sync.Mutex
	state  sessionState
	client *websocket.Conn
	agent  Session
	opened bool
	cfg    RelayConfig

	writeMu   sync.Mutex
	closeOnce sync.Once
}

// Bridge connects browser chat sockets to the agent. It speaks the normalized
// events and commands of this package to the browser and nothing else.
type Bridge struct {
	chat     Chat
	host     Host
	leases   *TaskLeases
	upgrader websocket.Upgrader

	mu       sync.Mutex
	sessions map[string]*chatSession
}

// NewBridge returns a bridge to chat that takes task capture leases from
// leases.
func NewBridge(chat Chat, host Host, leases *TaskLeases) *Bridge {
	return &Bridge{
		chat:   chat,
		host:   host,
		leases: leases,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin:     middleware.SameOrigin,
		},
		sessions: make(map[string]*chatSession),
	}
}

type relayResult struct {
	code   int
	reason string
}

// Connect is the chat socket route. The query parameter session_id names the
// session; without it the server picks one.
func (b *Bridge) Connect(c *gin.Context) {
	sessionID := strings.TrimSpace(c.Query("session_id"))
	if sessionID == "" {
		sessionID = uuid.NewString()
	}

	session, refusal := b.register(sessionID)
	if refusal != nil {
		writeError(c, refusal)
		return
	}
	if err := b.host.Admit(sessionID); err != nil {
		b.unregister(session)
		writeError(c, asError(err, CodeRuntimeUnavailable))
		return
	}

	agentSession, err := b.chat.OpenSession(context.Background(), sessionID)
	if err != nil {
		b.closeSession(session, CloseRuntimeUnavailable, asError(err, CodeRuntimeUnavailable).Message)
		return
	}
	if !session.setAgent(agentSession) {
		// The session was closed while the agent connected.
		_ = agentSession.Close()
		return
	}

	client, err := b.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("failed to upgrade the chat websocket: %s", err)
		b.closeSession(session, websocket.CloseNormalClosure, "upgrade failed")
		return
	}
	stopWatcher := middleware.WatchWebSocket(c.Request.Context(), client)
	defer stopWatcher()

	cfg := b.host.RelayConfig()
	if !session.activate(client, cfg) {
		writeClose(client, websocket.CloseNormalClosure, "session closed")
		_ = client.Close()
		return
	}
	b.host.Opened(sessionID)
	configureClientConn(client, cfg)

	var wg sync.WaitGroup
	results := make(chan relayResult, 2)
	wg.Add(3)
	go b.pingLoop(session, client, cfg, &wg)
	go b.pumpCommands(session, client, agentSession, results, &wg)
	go b.pumpEvents(session, agentSession, results, &wg)

	result := <-results
	b.closeSession(session, result.code, result.reason)
	wg.Wait()
}

func (b *Bridge) pingLoop(session *chatSession, client *websocket.Conn, cfg RelayConfig, wg *sync.WaitGroup) {
	defer wg.Done()
	interval := cfg.PingInterval
	if interval <= 0 {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if !b.host.Renew(session.id) {
			writeClose(client, CloseTakenOver, "session lock lost")
			_ = client.Close()
			return
		}
		if err := client.WriteControl(websocket.PingMessage, []byte("downstream"), writeDeadline(cfg)); err != nil {
			return
		}
	}
}

func (b *Bridge) pumpCommands(session *chatSession, client *websocket.Conn, agentSession Session, results chan<- relayResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		messageType, data, err := client.ReadMessage()
		if err != nil {
			results <- relayResult{code: clientCloseCode(err), reason: closeReason(err)}
			return
		}
		if !b.host.Renew(session.id) {
			results <- relayResult{code: CloseTakenOver, reason: "session lock lost"}
			return
		}
		if messageType != websocket.TextMessage {
			continue
		}
		var command Command
		if err := json.Unmarshal(data, &command); err != nil {
			log.Debugf("ignoring an unreadable chat command: %v", err)
			continue
		}

		switch command.Type {
		case CommandPrompt:
			maxRuntimeMs := 0
			if command.MaxRuntimeMs != nil {
				maxRuntimeMs = *command.MaxRuntimeMs
			}
			b.leases.Prompt(session.id, command.RequestID, maxRuntimeMs)
			err = agentSession.Prompt(context.Background(), Prompt{
				RequestID:    command.RequestID,
				Text:         command.Text,
				MaxSteps:     command.MaxSteps,
				MaxRuntimeMs: command.MaxRuntimeMs,
			})
		case CommandCancel:
			b.leases.Cancel(session.id, command.RequestID)
			err = agentSession.Cancel(context.Background(), command.RequestID)
		default:
			continue
		}
		if err != nil {
			results <- relayResult{code: clientCloseCode(err), reason: closeReason(err)}
			return
		}
	}
}

func (b *Bridge) pumpEvents(session *chatSession, agentSession Session, results chan<- relayResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		event, err := agentSession.Next()
		if err != nil {
			results <- relayResult{code: agentCloseCode(err), reason: closeReason(err)}
			return
		}
		if !b.host.Renew(session.id) {
			results <- relayResult{code: CloseTakenOver, reason: "session lock lost"}
			return
		}
		b.leases.Observe(session.id, event)
		if err := session.write(event); err != nil {
			results <- relayResult{code: agentCloseCode(err), reason: closeReason(err)}
			return
		}
	}
}

// Publish sends an event the server made, such as an observation, to one
// open session. It returns ErrSessionNotActive when no open session has the
// id.
func (b *Bridge) Publish(sessionID string, event Event) error {
	session := b.activeSession(sessionID)
	if session == nil {
		return ErrSessionNotActive
	}
	return session.write(event)
}

// Broadcast sends an event the server made to every open session.
func (b *Bridge) Broadcast(event Event) {
	for _, session := range b.snapshot() {
		session.mu.Lock()
		active := session.state == stateActive && session.client != nil
		session.mu.Unlock()
		if active {
			_ = session.write(event)
		}
	}
}

// Prompt sends a prompt the server made to one open session's agent. It
// takes no capture lease. It returns ErrSessionNotActive when no open
// session has the id.
func (b *Bridge) Prompt(sessionID string, prompt Prompt) error {
	session := b.activeSession(sessionID)
	if session == nil {
		return ErrSessionNotActive
	}
	session.mu.Lock()
	agentSession := session.agent
	session.mu.Unlock()
	return agentSession.Prompt(context.Background(), prompt)
}

// IsActive reports whether an open session has the id.
func (b *Bridge) IsActive(sessionID string) bool {
	return b.activeSession(sessionID) != nil
}

// CloseSession ends one session and releases its task capture leases, also
// when no session has the id.
func (b *Bridge) CloseSession(sessionID string, closeCode int, reason string) bool {
	b.mu.Lock()
	session := b.sessions[sessionID]
	b.mu.Unlock()
	if session != nil {
		b.closeSession(session, closeCode, reason)
	}
	b.leases.ReleaseSession(sessionID)
	return session != nil
}

// CloseAll ends every session and returns how many there were.
func (b *Bridge) CloseAll(closeCode int, reason string) int {
	sessions := b.snapshot()
	for _, session := range sessions {
		b.closeSession(session, closeCode, reason)
	}
	return len(sessions)
}

// SessionCount counts the sessions that are open or opening.
func (b *Bridge) SessionCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.sessions)
}

func (b *Bridge) register(sessionID string) (*chatSession, *Error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.sessions[sessionID]; ok {
		return nil, &Error{Status: http.StatusConflict, Code: CodeLockHeld, Message: "session is already connected"}
	}
	session := &chatSession{id: sessionID, state: stateConnecting}
	b.sessions[sessionID] = session
	return session, nil
}

func (b *Bridge) unregister(session *chatSession) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.sessions[session.id] == session {
		delete(b.sessions, session.id)
	}
}

func (b *Bridge) snapshot() []*chatSession {
	b.mu.Lock()
	defer b.mu.Unlock()
	sessions := make([]*chatSession, 0, len(b.sessions))
	for _, session := range b.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (b *Bridge) activeSession(sessionID string) *chatSession {
	b.mu.Lock()
	session := b.sessions[sessionID]
	b.mu.Unlock()
	if session == nil {
		return nil
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.state != stateActive || session.agent == nil {
		return nil
	}
	return session
}

func (b *Bridge) closeSession(session *chatSession, closeCode int, reason string) {
	session.closeOnce.Do(func() {
		session.mu.Lock()
		session.state = stateClosed
		agentSession := session.agent
		client := session.client
		opened := session.opened
		session.mu.Unlock()

		b.leases.Forget(session.id)
		if agentSession != nil {
			_ = agentSession.Close()
		}
		if client != nil {
			session.writeMu.Lock()
			writeClose(client, closeCode, reason)
			session.writeMu.Unlock()
			_ = client.Close()
		}
		b.unregister(session)
		b.host.Closed(session.id, closeCode, reason, opened)
	})
}

func (s *chatSession) setAgent(agentSession Session) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state == stateClosed {
		return false
	}
	s.agent = agentSession
	return true
}

func (s *chatSession) activate(client *websocket.Conn, cfg RelayConfig) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state == stateClosed {
		return false
	}
	s.client = client
	s.cfg = cfg
	s.opened = true
	s.state = stateActive
	return true
}

func (s *chatSession) write(event Event) error {
	s.mu.Lock()
	client := s.client
	cfg := s.cfg
	s.mu.Unlock()
	if client == nil {
		return fmt.Errorf("chat socket is not open")
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_ = client.SetWriteDeadline(writeDeadline(cfg))
	return client.WriteMessage(websocket.TextMessage, data)
}

func configureClientConn(conn *websocket.Conn, cfg RelayConfig) {
	readTimeout := cfg.ReadTimeout
	if readTimeout <= 0 {
		readTimeout = 60 * time.Second
	}
	if cfg.MaxMessageBytes > 0 {
		conn.SetReadLimit(cfg.MaxMessageBytes)
	}
	_ = conn.SetReadDeadline(time.Now().Add(readTimeout))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(readTimeout))
	})
}

func writeDeadline(cfg RelayConfig) time.Time {
	timeout := cfg.WriteTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return time.Now().Add(timeout)
}

func writeClose(conn *websocket.Conn, code int, reason string) {
	if conn == nil {
		return
	}
	message := websocket.FormatCloseMessage(code, reason)
	_ = conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(2*time.Second))
}

// clientCloseCode is the close code for a session the browser side ended:
// a normal end of the browser socket, or a failure to reach the agent with
// a command.
func clientCloseCode(err error) int {
	var closeErr *websocket.CloseError
	if errors.As(err, &closeErr) {
		switch closeErr.Code {
		case websocket.CloseNormalClosure, websocket.CloseGoingAway,
			websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived:
			return websocket.CloseNormalClosure
		}
		return closeErr.Code
	}
	return websocket.CloseNormalClosure
}

// agentCloseCode is the close code for a session the agent side ended. An
// agent that just went away ends it with CloseUpstreamClosed.
func agentCloseCode(err error) int {
	var closeErr *CloseError
	if errors.As(err, &closeErr) && closeErr.Code != 0 && closeErr.Code != websocket.CloseNormalClosure {
		return closeErr.Code
	}
	return CloseUpstreamClosed
}

func closeReason(err error) string {
	if err == nil {
		return "closed"
	}
	var agentClose *CloseError
	if errors.As(err, &agentClose) && agentClose.Reason != "" {
		return agentClose.Reason
	}
	var closeErr *websocket.CloseError
	if errors.As(err, &closeErr) && closeErr.Text != "" {
		return closeErr.Text
	}
	return fmt.Sprintf("relay closed: %v", err)
}
