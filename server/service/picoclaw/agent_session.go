package picoclaw

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"NanoKVM-Server/service/agent"
	"NanoKVM-Server/service/audit"

	"github.com/gin-gonic/gin"
	protocol "github.com/modelcontextprotocol/go-sdk/mcp"
)

// agentSessionHeader names the chat session a call acts for. The session is
// the bridge's (agent.Bridge), which opened the agent's own session under the
// same id, so the name holds for any agent. sessionIDHeader is the name it had
// while PicoClaw was the only agent; it is still read.
const agentSessionHeader = "X-Agent-Session-ID"

// namedSession is the session a request's headers name, empty when they name
// none.
func namedSession(header http.Header) string {
	if id := strings.TrimSpace(header.Get(agentSessionHeader)); id != "" {
		return id
	}
	return strings.TrimSpace(header.Get(sessionIDHeader))
}

// requestSession is the session a tool call's headers name.
func requestSession(req *protocol.CallToolRequest) string {
	if req == nil || req.Extra == nil {
		return ""
	}
	return namedSession(req.Extra.Header)
}

// actingSession is the chat session a device action acts for: the one the
// call names, or else the one that holds the AI input lock. An agent's MCP
// client sends the same headers on every call, so it rarely names one; the
// lock has one owner at a time, the session the user is chatting in. Empty
// when no session holds the lock: nobody is chatting, so nobody asked the
// agent to act.
func (s *Service) actingSession(named string) string {
	if named != "" || s.lock == nil {
		return named
	}
	return s.lock.Owner()
}

// The refusals of a device call that no user asked for. The model reads them
// as tool errors, so they say what to do next.
const (
	noSessionMessage = "no chat session holds device control: the agent acts only for a user chatting with it"
	noTurnMessage    = "the user's turn has ended: do not act on the host now; wait for the user's next prompt"
)

// requireTurn refuses a device call unless sessionID is a chat session whose
// agent is working on a prompt the user sent. An open chat is not enough:
// after the user cancels a turn, or once the agent has answered, tool calls
// the agent still has running must not reach the host. The turn is the
// bridge's own record (agent.Bridge.InTurn).
func (s *Service) requireTurn(sessionID string) *PicoclawError {
	if sessionID == "" {
		return newPicoclawError(CodeSessionIDInvalid, noSessionMessage)
	}
	s.ensureDependencies()
	if !s.chat.InTurn(sessionID) {
		err := newPicoclawError(CodeTurnNotActive, noTurnMessage)
		err.SessionID = sessionID
		return err
	}
	return nil
}

// AuditPrincipal is who a loopback route acts as, for its audit record: the
// agent, and the session the request acts for.
func (s *Service) AuditPrincipal(c *gin.Context) audit.Principal {
	return audit.Principal{
		Actor:   s.agentName(),
		Session: s.actingSession(namedSession(c.Request.Header)),
	}
}

func (s *Service) agentName() string {
	return s.Agent().Name()
}

// withAgentSession runs a device tool (a power press, a media change) for the
// chat session that holds the AI input lock, under a control mode lease, as
// kvm_actions runs, and only while that session has a turn in progress.
// Without it, a turn the agent kept running after the user cancelled it or
// closed the chat could still press the power button: kvm_actions is refused
// then, and these are no less consequential.
func (s *Service) withAgentSession(next protocol.ToolHandler) protocol.ToolHandler {
	return func(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
		release, modeErr := s.acquireControlMode()
		if modeErr != nil {
			return nil, s.controlConflictForMCP(modeErr)
		}
		defer release()

		sessionID := s.actingSession(requestSession(req))
		if turnErr := s.requireTurn(sessionID); turnErr != nil {
			return mcpToolError(turnErr.Message), nil
		}
		acquired, lockErr := s.acquireLock(sessionID)
		if lockErr != nil {
			return mcpToolError(lockErr.Message), nil
		}
		if acquired {
			defer s.lock.Release(sessionID)
		}
		return next(ctx, req)
	}
}

// lockHeldMessage refuses the AI input lock while another chat session holds
// it.
const lockHeldMessage = "another PicoClaw session is running"

// acquireLock takes the AI input lock for sessionID as
// agent.SessionLock.AcquireTemporary does, with its refusal as the routes and
// tools answer it.
func (s *Service) acquireLock(sessionID string) (bool, *PicoclawError) {
	acquired, err := s.lock.AcquireTemporary(sessionID)
	return acquired, lockError(err)
}

// lockError is a refusal of the AI input lock as the routes and tools answer
// it, nil for nil.
func lockError(err error) *PicoclawError {
	if err == nil {
		return nil
	}
	var held *agent.LockHeldError
	if errors.As(err, &held) {
		lockErr := newPicoclawError(CodePicoclawLockHeld, lockHeldMessage)
		lockErr.SessionID = held.Owner
		return lockErr
	}
	return newPicoclawError(CodeSessionIDInvalid, noSessionMessage)
}
