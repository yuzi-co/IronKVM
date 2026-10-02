package agent

import "time"

// EventType names a normalized chat event. The server sends each event to the
// browser as one JSON text message.
type EventType string

const (
	// EventTurnStarted: the agent started working on a turn. ACP has no
	// such event; its turn starts with the session/prompt request.
	EventTurnStarted EventType = "turn_started"
	// EventAgentMessage: an assistant message was created or replaced.
	// Fields: messageId, kind, text. ACP's agent_message_chunk appends
	// text to a message; this event replaces the whole text, because the
	// agents we run send the full text on every update, so it has a name
	// of its own.
	EventAgentMessage EventType = "agent_message"
	// EventAgentMessageRemoved: an assistant message was withdrawn, for
	// example a placeholder. Field: messageId. ACP has no equivalent.
	EventAgentMessageRemoved EventType = "agent_message_removed"
	// EventToolCall: the agent called a tool the sidebar shows as an
	// action. Fields: toolCallId, title, x, y. Shaped after ACP tool_call.
	EventToolCall EventType = "tool_call"
	// EventObservation: a screenshot the KVM tools took for the agent.
	// Fields: messageId, text, imageBase64, mimeType. The server sends it,
	// not the agent; ACP would carry it in a tool_call_update.
	EventObservation EventType = "observation"
	// EventError: an error. Fields: code, message, and requestId when the
	// error is about one prompt and the session is still fine.
	EventError EventType = "error"
	// EventTurnDone: the agent finished a turn. Fields: requestIds,
	// stopReason, usage. ACP returns this as the session/prompt response.
	EventTurnDone EventType = "turn_done"
	// EventControlModeChanged: the AI control mode changed. Field: control.
	// The server sends it, not the agent.
	EventControlModeChanged EventType = "control_mode_changed"
)

// MessageKind says how the browser shows an agent_message.
type MessageKind string

const (
	// KindReply is the assistant's answer. For an agent that does not send
	// turn_done it also ends the turn.
	KindReply MessageKind = "reply"
	// KindPlaceholder is a "thinking" text that a reply replaces.
	KindPlaceholder MessageKind = "placeholder"
	// KindHidden is reasoning or tool-call bookkeeping. The sidebar does not
	// show it; the observations and actions show what the tools did.
	KindHidden MessageKind = "hidden"
)

// StopReason says why a turn ended. The values are ACP's, plus StopError.
type StopReason string

const (
	StopEndTurn   StopReason = "end_turn"
	StopCancelled StopReason = "cancelled"
	// StopError is not in ACP, which reports a failed turn as an error
	// response to session/prompt.
	StopError StopReason = "error"
)

// Usage counts the tokens of a turn. The names follow ACP's usage object;
// LLMCalls is our own.
type Usage struct {
	InputTokens  int  `json:"inputTokens"`
	OutputTokens int  `json:"outputTokens"`
	TotalTokens  int  `json:"totalTokens"`
	LLMCalls     *int `json:"llmCalls,omitempty"`
}

// ControlMode is the AI control state carried by control_mode_changed.
type ControlMode struct {
	Mode          string    `json:"mode"`
	Transitioning bool      `json:"transitioning"`
	CanControl    bool      `json:"canControl"`
	LastError     string    `json:"lastError,omitempty"`
	ChangedAt     time.Time `json:"changedAt,omitzero"`
	Source        string    `json:"source,omitempty"`
}

// Event is one normalized chat event. Type decides which fields are set.
type Event struct {
	Type      EventType `json:"type"`
	SessionID string    `json:"sessionId,omitempty"`

	MessageID string      `json:"messageId,omitempty"`
	Kind      MessageKind `json:"kind,omitempty"`
	Text      string      `json:"text,omitempty"`

	ToolCallID string   `json:"toolCallId,omitempty"`
	Title      string   `json:"title,omitempty"`
	X          *float64 `json:"x,omitempty"`
	Y          *float64 `json:"y,omitempty"`

	ImageBase64 string `json:"imageBase64,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`

	Code      string `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
	RequestID string `json:"requestId,omitempty"`

	RequestIDs []string   `json:"requestIds,omitempty"`
	StopReason StopReason `json:"stopReason,omitempty"`
	Usage      *Usage     `json:"usage,omitempty"`

	Control *ControlMode `json:"control,omitempty"`
}

// Command types the browser sends on the chat socket. The names are ACP's
// method names; the fields are flattened into one JSON object.
const (
	CommandPrompt = "session/prompt"
	CommandCancel = "session/cancel"
)

// Command is one message from the browser.
//
//	{"type":"session/prompt","requestId":"r1","text":"hi","maxSteps":8,"maxRuntimeMs":60000}
//	{"type":"session/cancel"}
type Command struct {
	Type         string `json:"type"`
	RequestID    string `json:"requestId,omitempty"`
	Text         string `json:"text,omitempty"`
	MaxSteps     *int   `json:"maxSteps,omitempty"`
	MaxRuntimeMs *int   `json:"maxRuntimeMs,omitempty"`
}
