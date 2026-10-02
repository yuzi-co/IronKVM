// Package agent is the boundary between the IronKVM server and the on-device
// AI agent. The server's chat socket, the chat history routes and the task
// capture lease use only the interfaces here; an adapter such as the PicoClaw
// one in service/picoclaw translates them to one agent's own protocol and
// files.
//
// The chat events the browser receives are normalized (see Event). Where the
// Agent Client Protocol (agentclientprotocol.com) has a matching shape, the
// names follow it: the session/prompt and session/cancel commands, the
// tool_call event with toolCallId and title, the stopReason values of a turn
// and the camelCase usage counters. Where ACP does not fit, the event is our
// own; Event documents each case.
package agent

import "context"

// Agent is one on-device agent behind the server.
type Agent interface {
	// Name is a short identifier for logs, such as "picoclaw".
	Name() string
	Runtime
	Chat
	History
}

// Runtime controls the agent process.
type Runtime interface {
	// Start starts the agent and returns once it is ready or has failed.
	Start(ctx context.Context) error
	// Stop stops the agent and returns once it no longer runs.
	Stop(ctx context.Context) error
	// Ready returns nil when the agent can take a chat session.
	Ready(ctx context.Context) error
	// ApplyConfig writes the agent's configuration with the settings the
	// server enforces (hardening, the KVM MCP server, the agent profile).
	// Start applies it too.
	ApplyConfig(ctx context.Context) error
}

// Chat opens chat sessions with the agent.
type Chat interface {
	// OpenSession connects to the agent for the given session id. The id
	// is the browser's; an agent that keeps history keys it by this id.
	OpenSession(ctx context.Context, sessionID string) (Session, error)
}

// Session is one open chat connection to the agent. Prompt, Cancel and Close
// may be called from several goroutines at once; Next is called from one.
type Session interface {
	// Prompt sends a user message. It starts a turn, or joins the running
	// one for an agent that batches messages into a turn.
	Prompt(ctx context.Context, prompt Prompt) error
	// Cancel asks the agent to stop the turn that handles requestID, or the
	// current turn when requestID is empty.
	Cancel(ctx context.Context, requestID string) error
	// Next blocks until the agent's next event. It returns a *CloseError
	// (or another error) once the session can give no more events. Close
	// unblocks it.
	Next() (Event, error)
	// Close ends the connection. It is safe to call more than once.
	Close() error
}

// Prompt is a user message for the agent.
type Prompt struct {
	// RequestID identifies the prompt in the turn_done event and the
	// capture lease. The browser picks it.
	RequestID string
	Text      string
	// MaxSteps and MaxRuntimeMs are limits the user set, nil when unset.
	MaxSteps     *int
	MaxRuntimeMs *int
}

// History reads and deletes the agent's stored chat sessions.
type History interface {
	// ListSessions returns every stored session, newest first.
	ListSessions(ctx context.Context) ([]SessionSummary, error)
	// ReadSession returns ErrSessionNotFound or ErrInvalidSessionID when
	// it cannot give the session.
	ReadSession(ctx context.Context, id string) (SessionDetail, error)
	// DeleteSession returns ErrSessionNotFound or ErrInvalidSessionID when
	// it cannot delete the session.
	DeleteSession(ctx context.Context, id string) error
}

// SessionSummary is one entry of the chat history list. The JSON names are
// the ones the history routes have always returned.
type SessionSummary struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Preview      string `json:"preview"`
	MessageCount int    `json:"message_count"`
	Created      string `json:"created"`
	Updated      string `json:"updated"`
}

// SessionMessage is one user or assistant message of a stored session.
type SessionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// SessionDetail is a stored session with its messages.
type SessionDetail struct {
	ID       string           `json:"id"`
	Messages []SessionMessage `json:"messages"`
	Summary  string           `json:"summary,omitempty"`
	Created  string           `json:"created"`
	Updated  string           `json:"updated"`
}
