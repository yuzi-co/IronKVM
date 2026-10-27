package picoclaw

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"NanoKVM-Server/middleware"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/agent"
	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/service/hostdevices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	protocol "github.com/modelcontextprotocol/go-sdk/mcp"
)

// The agent's power and reset presses wait for a person (ironkvm-dist#167).
// kvm_power does not press: it records the press as a confirmation tied to
// the chat session and its turn, and answers CONFIRMATION_REQUIRED. The web
// UI shows it to the signed-in users who may press power, and one of them
// confirms or denies it. Only a confirmation from a login session presses,
// and only once. The agent learns the outcome from kvm_power_confirmation,
// which waits for it. Presses from the web UI, IPMI, Home Assistant, external
// MCP and the REST API do not come through here and are unchanged.
//
// A chat channel other than the web UI, such as Telegram, cannot confirm a
// press: the model is told to ask the user to confirm it in the web UI.

// The states of a confirmation. pressing is the short span between a user's
// confirmation and the end of the press; it is not shown as pending.
const (
	confirmationPending  = "pending"
	confirmationPressing = "pressing"
	confirmationPressed  = "pressed"
	confirmationDenied   = "denied"
	confirmationExpired  = "expired"
	confirmationFailed   = "failed"
)

// CodeConfirmationRequired is kvm_power's answer: nothing was pressed yet.
const CodeConfirmationRequired = "CONFIRMATION_REQUIRED"

// The error codes of the confirm and deny routes, in the web UI's
// { code, msg } shape.
const (
	confirmationNotFound   = -1
	confirmationNotPending = -2
	confirmationPressError = -3
)

const powerConfirmationToolName = "kvm_power_confirmation"

var (
	// powerConfirmationTTL is how long a press waits for a person.
	powerConfirmationTTL = 60 * time.Second
	// powerConfirmationWait bounds one kvm_power_confirmation call, well
	// inside the turn's time limit; the agent calls again to wait longer.
	powerConfirmationWait = 25 * time.Second
	// powerConfirmationKeep is how long an ended confirmation stays readable,
	// so that the agent can still learn how it ended.
	powerConfirmationKeep = 5 * time.Minute
	// turnCheckInterval is how often a wait checks that the turn goes on.
	turnCheckInterval = time.Second
)

// The texts the model reads. They say what to do next.
const (
	confirmationRequiredMessage = "Not pressed yet: the server needs a person to confirm this press. Tell the user which button and press you asked for and ask them to confirm it in the IronKVM web UI, which shows it to every signed-in user who may press power; a reply in this chat, or in a channel such as Telegram, does not confirm it. Then call " + powerConfirmationToolName + " with this id to wait for their answer."
	confirmationWaitingMessage  = "Still waiting for the user to confirm the press in the IronKVM web UI. Call " + powerConfirmationToolName + " again to keep waiting."
	confirmationPressedMessage  = "The user confirmed and the button was pressed. Check " + powerStatusToolName + " or a screenshot for the result."
	confirmationDeniedMessage   = "The user denied the press; the button was not pressed. Do not ask for it again unless the user does."
	confirmationTimedOutMessage = "Nobody confirmed the press in time; the button was not pressed."
	confirmationTurnEndMessage  = "The user's turn ended before the press was confirmed; the button was not pressed."
)

// powerConfirmation is a confirmation and what the store keeps beside it.
type powerConfirmation struct {
	agent.PowerConfirmation
	press hostdevices.Press
	// turn is the chat session's turn the press was asked in.
	turn  uint64
	ended time.Time
	done  chan struct{}
	timer *time.Timer
}

// confirmationStore holds the confirmations, in memory: a restart forgets
// them, and a press nobody confirmed is not pressed.
type confirmationStore struct {
	mu    sync.Mutex
	items map[string]*powerConfirmation
}

// add records c unless its session already has a press waiting, which it
// returns instead.
func (s *confirmationStore) add(c *powerConfirmation, now time.Time) (agent.PowerConfirmation, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked(now)
	for _, item := range s.items {
		if item.SessionID == c.SessionID && (item.State == confirmationPending || item.State == confirmationPressing) {
			return item.PowerConfirmation, false
		}
	}
	if s.items == nil {
		s.items = make(map[string]*powerConfirmation)
	}
	s.items[c.ID] = c
	return c.PowerConfirmation, true
}

func (s *confirmationStore) pruneLocked(now time.Time) {
	for id, item := range s.items {
		if !item.ended.IsZero() && now.Sub(item.ended) > powerConfirmationKeep {
			delete(s.items, id)
		}
	}
}

// get returns a confirmation and the channel that closes when it ends.
func (s *confirmationStore) get(id string) (agent.PowerConfirmation, uint64, <-chan struct{}, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return agent.PowerConfirmation{}, 0, nil, false
	}
	return item.PowerConfirmation, item.turn, item.done, true
}

// pending lists the presses that wait for a person, oldest first.
func (s *confirmationStore) pending() []agent.PowerConfirmation {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := []agent.PowerConfirmation{}
	for _, item := range s.items {
		if item.State == confirmationPending {
			list = append(list, item.PowerConfirmation)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.Before(list[j].CreatedAt) })
	return list
}

// claim takes a pending confirmation for the one person who confirms it: it
// moves to pressing, so a second confirmation finds it no longer pending.
func (s *confirmationStore) claim(id, by string) (*powerConfirmation, agent.PowerConfirmation, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return nil, agent.PowerConfirmation{}, false
	}
	if item.State != confirmationPending {
		return nil, item.PowerConfirmation, true
	}
	item.State = confirmationPressing
	item.DecidedBy = by
	if item.timer != nil {
		item.timer.Stop()
	}
	claimed := *item
	return &claimed, item.PowerConfirmation, true
}

// end moves a confirmation that is in one of the from states to state, and
// wakes whoever waits for it. It reports whether it did.
func (s *confirmationStore) end(id, state, message, by string, now time.Time, from ...string) (agent.PowerConfirmation, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return agent.PowerConfirmation{}, false
	}
	allowed := false
	for _, f := range from {
		allowed = allowed || item.State == f
	}
	if !allowed {
		return item.PowerConfirmation, false
	}
	item.State = state
	item.Message = message
	if by != "" {
		item.DecidedBy = by
	}
	item.ended = now
	if item.timer != nil {
		item.timer.Stop()
	}
	close(item.done)
	return item.PowerConfirmation, true
}

// requestPowerPress is kvm_power. withAgentSession has checked the chat
// session, its turn, device control and the AI input lock; the press waits
// for a person all the same.
func (s *Service) requestPowerPress(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var args powerArguments
	if !decodeArguments(req, &args) {
		return mcpToolError("invalid arguments: give button (power or reset) and press (short or long)"), nil
	}
	press := args.press()
	button, _, err := hostdevices.CheckPress(press)
	if err != nil {
		return mcpToolError(err.Error()), nil
	}
	pressName := strings.ToLower(strings.TrimSpace(press.Press))
	if pressName == "" {
		pressName = hostdevices.PressShort
	}

	sessionID := s.actingSession(requestSession(req))
	now := time.Now()
	c := &powerConfirmation{
		PowerConfirmation: agent.PowerConfirmation{
			ID:        uuid.NewString(),
			Agent:     s.agentName(),
			SessionID: sessionID,
			Button:    button,
			Press:     pressName,
			State:     confirmationPending,
			CreatedAt: now,
			ExpiresAt: now.Add(powerConfirmationTTL),
		},
		press: press,
		turn:  s.chat.Turn(sessionID),
		done:  make(chan struct{}),
	}
	if existing, added := s.confirmations.add(c, now); !added {
		// One press at a time per chat: a second short press would undo the
		// first.
		return mcpToolError(fmt.Sprintf("a %s %s press (id %s) is already waiting for the user or being made: call %s with that id instead of asking again",
			existing.Press, existing.Button, existing.ID, powerConfirmationToolName)), nil
	}
	id := c.ID
	c.timer = time.AfterFunc(powerConfirmationTTL, func() {
		s.endConfirmation(id, confirmationExpired, confirmationTimedOutMessage, "", confirmationPending)
	})

	noteToolDetail(ctx, "awaiting confirmation "+id)
	s.publishConfirmation(c.PowerConfirmation)
	return mcpJSONResult(map[string]any{
		"code":             CodeConfirmationRequired,
		"id":               id,
		"pressed":          false,
		"button":           button,
		"press":            pressName,
		"expiresInSeconds": int(powerConfirmationTTL / time.Second),
		"message":          confirmationRequiredMessage,
	}), nil
}

// endConfirmation ends a confirmation and tells the chat session.
func (s *Service) endConfirmation(id, state, message, by string, from ...string) (agent.PowerConfirmation, bool) {
	c, ended := s.confirmations.end(id, state, message, by, time.Now(), from...)
	if ended {
		s.publishConfirmation(c)
	}
	return c, ended
}

// publishConfirmation shows the chat session a confirmation's change, so an
// open chat panel need not wait for its next poll.
func (s *Service) publishConfirmation(c agent.PowerConfirmation) {
	s.ensureDependencies()
	_ = s.chat.Publish(c.SessionID, agent.Event{
		Type:         agent.EventPowerConfirmation,
		SessionID:    c.SessionID,
		Confirmation: &c,
	})
}

type confirmationArguments struct {
	ID string `json:"id"`
}

// mcpPowerConfirmation is kvm_power_confirmation: it waits up to
// powerConfirmationWait for a press to be confirmed, denied or to expire.
// It holds neither device control nor the input lock while it waits.
func (s *Service) mcpPowerConfirmation(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var args confirmationArguments
	if !decodeArguments(req, &args) || strings.TrimSpace(args.ID) == "" {
		return mcpToolError("invalid arguments: give the id kvm_power returned"), nil
	}
	sessionID := s.actingSession(requestSession(req))
	c, err := s.awaitConfirmation(ctx, sessionID, strings.TrimSpace(args.ID), powerConfirmationWait)
	if err != nil {
		return mcpToolError(err.Error()), nil
	}
	answer := map[string]any{
		"id":      c.ID,
		"state":   c.State,
		"pressed": c.State == confirmationPressed,
		"button":  c.Button,
		"press":   c.Press,
		"message": agentConfirmationMessage(c),
	}
	if c.State == confirmationPressing {
		answer["state"] = confirmationPending
	}
	if c.State == confirmationPending {
		answer["expiresInSeconds"] = max(0, int(time.Until(c.ExpiresAt).Seconds()))
	}
	if c.State == confirmationFailed {
		result := mcpJSONResult(answer)
		result.IsError = true
		return result, nil
	}
	return mcpJSONResult(answer), nil
}

// awaitConfirmation waits for a confirmation of the session to end, at most
// wait. A turn that ends while it waits ends the confirmation.
func (s *Service) awaitConfirmation(ctx context.Context, sessionID, id string, wait time.Duration) (agent.PowerConfirmation, error) {
	c, turn, done, ok := s.confirmations.get(id)
	if !ok || sessionID == "" || c.SessionID != sessionID {
		return agent.PowerConfirmation{}, fmt.Errorf("no press of this chat session has the id %s", id)
	}
	deadline := time.NewTimer(wait)
	defer deadline.Stop()
	check := time.NewTicker(turnCheckInterval)
	defer check.Stop()
	for {
		select {
		case <-done:
		case <-ctx.Done():
		case <-deadline.C:
		case <-check.C:
			if s.chat.Turn(sessionID) == turn {
				continue
			}
			s.endConfirmation(id, confirmationExpired, confirmationTurnEndMessage, "", confirmationPending)
		}
		c, _, _, _ = s.confirmations.get(id)
		return c, nil
	}
}

func agentConfirmationMessage(c agent.PowerConfirmation) string {
	switch c.State {
	case confirmationPressed:
		return confirmationPressedMessage
	case confirmationDenied:
		return confirmationDeniedMessage
	case confirmationExpired, confirmationFailed:
		return c.Message
	}
	return confirmationWaitingMessage
}

// pressConfirmed presses for a confirmation a person gave, if the agent may
// still press for that session: the power tier is on, the turn the press was
// asked in goes on, the agent holds device control and the session the input
// lock, as when kvm_power pressed at once.
func (s *Service) pressConfirmed(c *powerConfirmation) (state, message string) {
	s.ensureDependencies()
	if !audit.Allowed(audit.SubjectAgent, audit.TierPower) {
		return confirmationFailed, "The owner turned the agent's power tier off; the button was not pressed."
	}
	if turn := s.chat.Turn(c.SessionID); turn == 0 || turn != c.turn {
		return confirmationExpired, confirmationTurnEndMessage
	}
	release, modeErr := s.acquireControlMode()
	if modeErr != nil {
		return confirmationFailed, "The agent does not hold device control: " + modeErr.Message
	}
	defer release()
	acquired, lockErr := s.lock.AcquireTemporary(c.SessionID)
	if lockErr != nil {
		return confirmationFailed, lockErr.Message
	}
	if acquired {
		defer s.lock.Release(c.SessionID)
	}
	if s.presser == nil {
		return confirmationFailed, "This board has no power buttons for the agent."
	}
	if err := s.presser.Press(c.press); err != nil {
		return confirmationFailed, "The press failed: " + err.Error()
	}
	return confirmationPressed, ""
}

// ListPowerConfirmations answers the presses that wait for a person.
func (s *Service) ListPowerConfirmations(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, s.confirmations.pending())
}

// ConfirmPowerConfirmation presses for a pending confirmation, once. The
// route takes only a login session, never an API key.
func (s *Service) ConfirmPowerConfirmation(c *gin.Context) {
	answerConfirmation(c, s.confirmPowerPress)
}

// DenyPowerConfirmation refuses a pending confirmation.
func (s *Service) DenyPowerConfirmation(c *gin.Context) {
	answerConfirmation(c, s.denyPowerPress)
}

func answerConfirmation(c *gin.Context, answer func(id, by string) (agent.PowerConfirmation, int, string)) {
	var rsp proto.Response
	confirmation, code, message := answer(c.Param("id"), confirmingUser(c))
	if confirmation.ID != "" {
		audit.SetTarget(c, confirmation.Button)
		audit.SetDetail(c, auditConfirmationDetail(confirmation))
		rsp.Data = confirmation
	}
	if code != 0 {
		rsp.ErrRsp(c, code, message)
		return
	}
	rsp.OkRspWithData(c, confirmation)
}

// confirmPowerPress is by's confirmation of the press id: it presses once,
// and answers the confirmation as it ended, with a nonzero code when nothing
// was pressed.
func (s *Service) confirmPowerPress(id, by string) (agent.PowerConfirmation, int, string) {
	claimed, current, found := s.confirmations.claim(id, by)
	if !found {
		return agent.PowerConfirmation{}, confirmationNotFound, "no such press waits for confirmation"
	}
	if claimed == nil {
		return current, confirmationNotPending, "the press is no longer waiting for confirmation: " + current.State
	}
	state, message := s.pressConfirmed(claimed)
	ended, _ := s.endConfirmation(claimed.ID, state, message, by, confirmationPressing)
	if state != confirmationPressed {
		return ended, confirmationPressError, message
	}
	return ended, 0, ""
}

// denyPowerPress is by's refusal of the press id.
func (s *Service) denyPowerPress(id, by string) (agent.PowerConfirmation, int, string) {
	if _, _, _, found := s.confirmations.get(id); !found {
		return agent.PowerConfirmation{}, confirmationNotFound, "no such press waits for confirmation"
	}
	ended, ok := s.endConfirmation(id, confirmationDenied, confirmationDeniedMessage, by, confirmationPending)
	if !ok {
		return ended, confirmationNotPending, "the press is no longer waiting for confirmation: " + ended.State
	}
	return ended, 0, ""
}

func confirmingUser(c *gin.Context) string {
	principal, _ := middleware.CurrentPrincipal(c)
	return principal.Username
}

func auditConfirmationDetail(c agent.PowerConfirmation) string {
	return fmt.Sprintf("%s press for %s session %s", c.Press, c.Agent, c.SessionID)
}
