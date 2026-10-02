package agent

import (
	"errors"
	"net/http"
)

// WebSocket close codes the chat socket ends with. The browser reads them.
const (
	CloseLockHeld            = 4001
	CloseRuntimeUnavailable  = 4002
	CloseAuthFailed          = 4003
	CloseTakenOver           = 4004
	CloseUpstreamClosed      = 4005
	CloseControlModeSwitched = 4006
	CloseRuntimeStopped      = 4007
)

// Error codes the routes in this package answer with. They are the codes the
// PicoClaw routes always used, so the browser reads them unchanged.
const (
	CodeLockHeld           = "AI_LOCK_HELD"
	CodeRuntimeUnavailable = "RUNTIME_UNAVAILABLE"
	CodeInvalidRequest     = "INVALID_ACTION"
)

// Error is an error with the code and HTTP status a route answers with.
type Error struct {
	Status  int
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Message }

// NewError returns an *Error with HTTP status 200, the status the server's
// API envelope uses for errors.
func NewError(code, message string) *Error {
	return &Error{Status: http.StatusOK, Code: code, Message: message}
}

var (
	ErrSessionNotFound  = &Error{Status: http.StatusNotFound, Code: CodeRuntimeUnavailable, Message: "session not found"}
	ErrInvalidSessionID = &Error{Status: http.StatusOK, Code: CodeInvalidRequest, Message: "invalid session id"}
	// ErrSessionNotActive: no open chat session has the id.
	ErrSessionNotActive = errors.New("chat session is not active")
)

// CloseError ends a chat session with a WebSocket close code.
type CloseError struct {
	Code   int
	Reason string
}

func (e *CloseError) Error() string { return e.Reason }

// asError returns err as an *Error, wrapping any other error in one with
// the given code.
func asError(err error, code string) *Error {
	var agentErr *Error
	if errors.As(err, &agentErr) {
		return agentErr
	}
	return NewError(code, err.Error())
}
