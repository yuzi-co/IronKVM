package agent

import (
	"errors"
	"net/http"
	"time"

	"NanoKVM-Server/service/controlmode"

	log "github.com/sirupsen/logrus"
)

// ChatHub is the chat side the KVM services reach while a chat runs: the
// Bridge. The agent's device tools publish observations through it, ask it
// whether a session has a turn in progress and close sessions with it.
type ChatHub interface {
	Publish(sessionID string, event Event) error
	Broadcast(event Event)
	Prompt(sessionID string, prompt Prompt) error
	IsActive(sessionID string) bool
	InTurn(sessionID string) bool
	Turn(sessionID string) uint64
	CloseSession(sessionID string, closeCode int, reason string) bool
	CloseAll(closeCode int, reason string) int
	SessionCount() int
}

var _ ChatHub = (*Bridge)(nil)

// NoChat stands in for the Bridge until one is connected, as in tests.
type NoChat struct{}

func (NoChat) Publish(string, Event) error           { return ErrSessionNotActive }
func (NoChat) Broadcast(Event)                       {}
func (NoChat) Prompt(string, Prompt) error           { return ErrSessionNotActive }
func (NoChat) IsActive(string) bool                  { return false }
func (NoChat) InTurn(string) bool                    { return false }
func (NoChat) Turn(string) uint64                    { return 0 }
func (NoChat) CloseSession(string, int, string) bool { return false }
func (NoChat) CloseAll(int, string) int              { return 0 }
func (NoChat) SessionCount() int                     { return 0 }

// ChatHostConfig is what NewChatHost needs. Lifecycle and Lock are required;
// the agent's own parts may be nil.
type ChatHostConfig struct {
	// Lifecycle is the agent's lifecycle: the AI control mode, the
	// runtime status and the agent's title come from it.
	Lifecycle *Lifecycle
	// Lock is the AI input lock.
	Lock *SessionLock
	// RelayConfig gives the agent's limits for the browser socket.
	RelayConfig func() RelayConfig
	// Opened is the agent's own part of a session's start.
	Opened func(sessionID string)
	// Closed is the agent's own cleanup when a session ends, before the
	// lock is released. opened is as in Host.Closed.
	Closed func(sessionID string, opened bool)
	// ReleaseInput lets go of input the agent left held on the host. It
	// runs when the session that ended held the AI input lock while the
	// agent was in control.
	ReleaseInput func() error
}

// ChatHost is the KVM side of the chat bridge (Host): it decides who may chat,
// holds the AI input lock and cleans up when a chat session ends. None of it
// depends on which agent runs; the agent's own part comes in ChatHostConfig.
type ChatHost struct {
	lifecycle    *Lifecycle
	lock         *SessionLock
	relayConfig  func() RelayConfig
	opened       func(sessionID string)
	closed       func(sessionID string, opened bool)
	releaseInput func() error
}

var _ Host = (*ChatHost)(nil)

// NewChatHost returns the chat host for cfg's agent.
func NewChatHost(cfg ChatHostConfig) *ChatHost {
	return &ChatHost{
		lifecycle:    cfg.Lifecycle,
		lock:         cfg.Lock,
		relayConfig:  cfg.RelayConfig,
		opened:       cfg.Opened,
		closed:       cfg.Closed,
		releaseInput: cfg.ReleaseInput,
	}
}

// Admit lets a chat session open while the AI control mode allows chat and
// the session can hold the AI input lock.
func (h *ChatHost) Admit(sessionID string) error {
	modeStatus, err := h.lifecycle.control.Status()
	if err != nil {
		return NewError(CodeRuntimeUnavailable, err.Error())
	}
	if modeStatus.Transitioning || modeStatus.Mode == controlmode.ModeMCP {
		controlErr := h.lifecycle.ControlError(nil)
		return &Error{Status: http.StatusConflict, Code: controlErr.Code, Message: controlErr.Message}
	}
	if lockErr := h.lock.Ensure(sessionID); lockErr != nil {
		return &Error{Status: http.StatusConflict, Code: CodeLockHeld, Message: h.lockMessage(lockErr)}
	}
	return nil
}

// lockMessage is the reason Admit gives for a refusal of the lock.
func (h *ChatHost) lockMessage(err error) string {
	if errors.Is(err, ErrNoSessionID) {
		return NoSessionMessage
	}
	return "another " + h.lifecycle.title + " session is running"
}

func (h *ChatHost) Renew(sessionID string) bool {
	return h.lock.Renew(sessionID)
}

func (h *ChatHost) Opened(sessionID string) {
	if h.opened != nil {
		h.opened(sessionID)
	}
}

// Closed releases what a chat session held: what the agent keeps for it, the
// AI input lock and any input left held on the host. A session that ended
// because the agent was not reachable marks the runtime unavailable.
func (h *ChatHost) Closed(sessionID string, closeCode int, reason string, opened bool) {
	if h.closed != nil {
		h.closed(sessionID, opened)
	}
	h.Release(sessionID)

	store := h.lifecycle.status
	status := store.Get()
	status.CurrentSession = h.lock.Owner()
	if closeCode == CloseUpstreamClosed || closeCode == CloseRuntimeUnavailable {
		status.Ready = false
		if closeCode == CloseUpstreamClosed {
			status.Status = "unavailable"
		}
		status.LastError = reason
		status.CheckedAt = time.Now()
		store.Set(status)
	}
}

// Release releases the AI input lock if sessionID holds it, and then lets go
// of the input the agent left held on the host.
func (h *ChatHost) Release(sessionID string) {
	// Only the agent sends input while it is in control, so only what it
	// left held needs releasing. After a turn that pressed nothing this
	// writes nothing, and in particular never touches the relative mouse
	// endpoint, which times out whenever the host is not polling it.
	var releaseInput func() error
	if h.lifecycle.control.Current() == controlmode.ModePicoclaw {
		releaseInput = h.releaseInput
	}
	if _, err := releaseOwnedSession(h.lock, sessionID, releaseInput); err != nil {
		log.Warnf("%s session %s ended with input still held on the host: %v", h.lifecycle.title, sessionID, err)
	}
}

func releaseOwnedSession(lock *SessionLock, sessionID string, releaseHID func() error) (bool, error) {
	if lock == nil || !lock.ReleaseOwned(sessionID) {
		return false, nil
	}
	if releaseHID == nil {
		return true, nil
	}
	return true, releaseHID()
}

// RelayConfig gives the browser socket the agent's limits.
func (h *ChatHost) RelayConfig() RelayConfig {
	if h.relayConfig == nil {
		return RelayConfig{}
	}
	return h.relayConfig()
}
