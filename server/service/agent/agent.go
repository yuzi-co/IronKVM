// Package agent is the boundary between the IronKVM server and the on-device
// AI agent. The server's chat socket, the chat history routes, the task
// capture lease and the agent's lifecycle (install, start, stop, restore at
// boot, the runtime status and the AI control mode around them; see
// Lifecycle) use only the interfaces here. So do the AI input lock
// (SessionLock) and the KVM side of a chat session (ChatHost), which are the
// same whichever agent runs. An adapter such as the PicoClaw one
// in service/picoclaw translates them to one agent's own protocol, process and
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

// Runtime is the agent's own part of its lifecycle: its install, its process
// and its configuration. The Lifecycle drives it and keeps the rest: the AI
// control mode, the persisted intent to run, the status the web UI reads, the
// locks around a start or stop, and the restore at boot. A Runtime may record
// detail only it knows, such as a config error, into the StatusStore the
// Lifecycle was built with.
//
// Errors that answer a route are *Error values or CodedError ones; any other
// error answers with the code the route uses for failures.
type Runtime interface {
	// InstallPath is where the agent's binary is installed.
	InstallPath() string
	// Version is the release Install installs.
	Version() string
	// Installed reports whether the agent is installed.
	Installed() (bool, error)
	// Stale reports whether the installed agent is not the release
	// Install installs, so an update is available.
	Stale() bool
	// Install downloads and installs the release. It reports its progress
	// as it goes. The Lifecycle runs it in the background.
	Install(progress func(stage string, percent int, lastError string)) error
	// Uninstall removes the agent's binary and settings. The Lifecycle has
	// stopped the agent first.
	Uninstall()

	// Model reports the agent's model settings.
	Model() (Model, error)
	// ApplyConfig writes the agent's configuration with the settings the
	// server enforces (hardening, the KVM MCP server, the agent profile).
	// Launch applies it too.
	ApplyConfig(ctx context.Context) error

	// Running reports whether the agent's process runs.
	Running() (bool, error)
	// Launch starts the agent's process and returns the command it ran and
	// its output. It does not wait for the agent to be ready: the
	// Lifecycle polls Check for that, and stops the process when it does
	// not become ready.
	Launch(ctx context.Context) (command, output string, err error)
	// Halt asks the agent's process to stop, once. The Lifecycle waits
	// until Running reports it gone.
	Halt() (command, output string, err error)
	// Check checks that the agent is installed, configured, running and
	// answering. It returns the status to record, or nil to keep the one
	// recorded, and an error when the agent cannot take a chat session.
	// The Lifecycle sets the status's CheckedAt and CurrentSession.
	Check(ctx context.Context) (*RuntimeStatus, error)
	// Describe adds the agent's own fields to the status the status route
	// answers with.
	Describe(status RuntimeStatus) RuntimeStatus
}

// Model is the agent's model settings.
type Model struct {
	// Configured says the model has what it needs, such as its API key.
	Configured bool
	// Name is the configured model's name.
	Name string
	// Target is the model the agent is set to use, configured or not.
	Target string
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

// TurnReporter is a Session that may say up front that its agent ends every
// turn with turn_done. The capture lease of a turn then lasts until turn_done
// from the first turn on. Without it, the bridge learns this from the first
// turn_done of the session, and until then a reply ends the lease, which in
// a turn that replies before its last tool call is too soon.
type TurnReporter interface {
	ReportsTurnDone() bool
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
