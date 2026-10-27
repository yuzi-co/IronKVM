package picoclaw

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"NanoKVM-Server/service/agent"
	"NanoKVM-Server/service/audit"
)

// requestPress calls kvm_power and returns the id of the confirmation it
// asked for, failing unless it pressed nothing and asked for one.
func requestPress(t *testing.T, service *Service, args string) string {
	t.Helper()
	result := toolResult(t, callTool(t, service, "kvm_power", args, nil))
	if isErrorResult(result) {
		t.Fatalf("kvm_power %s = %v", args, result)
	}
	var answer struct {
		Code    string `json:"code"`
		ID      string `json:"id"`
		Pressed bool   `json:"pressed"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(toolText(t, result)), &answer); err != nil {
		t.Fatal(err)
	}
	if answer.Code != CodeConfirmationRequired || answer.ID == "" || answer.Pressed {
		t.Fatalf("kvm_power %s = %+v, want a confirmation asked for", args, answer)
	}
	return answer.ID
}

// awaitAnswer calls kvm_power_confirmation for id and returns its answer.
func awaitAnswer(t *testing.T, service *Service, id string, headers map[string]string) (map[string]any, bool) {
	t.Helper()
	result := toolResult(t, callTool(t, service, "kvm_power_confirmation", `{"id":"`+id+`"}`, headers))
	var answer map[string]any
	if err := json.Unmarshal([]byte(toolText(t, result)), &answer); err != nil {
		return map[string]any{"text": toolText(t, result)}, isErrorResult(result)
	}
	return answer, isErrorResult(result)
}

func shortWaits(t *testing.T, ttl, wait time.Duration) {
	t.Helper()
	previousTTL, previousWait, previousCheck := powerConfirmationTTL, powerConfirmationWait, turnCheckInterval
	powerConfirmationTTL, powerConfirmationWait, turnCheckInterval = ttl, wait, 5*time.Millisecond
	t.Cleanup(func() {
		powerConfirmationTTL, powerConfirmationWait, turnCheckInterval = previousTTL, previousWait, previousCheck
	})
}

// An agent's press does not press: it waits for a person, and the chat
// panel and the web UI's list show it.
func TestAnAgentPressWaitsForConfirmation(t *testing.T) {
	service, host := newDeviceTestService(t)
	withAgentTiers(t, audit.TierPower)

	id := requestPress(t, service, `{"button":"power","press":"long"}`)
	if len(host.pressed) != 0 {
		t.Fatalf("pressed %v before anyone confirmed", host.pressed)
	}
	pending := service.confirmations.pending()
	if len(pending) != 1 || pending[0].ID != id || pending[0].Button != "power" || pending[0].Press != "long" ||
		pending[0].SessionID != deviceTestSession || pending[0].Agent != "picoclaw" {
		t.Fatalf("pending = %+v", pending)
	}
	if left := time.Until(pending[0].ExpiresAt); left <= 0 || left > powerConfirmationTTL {
		t.Fatalf("expires in %s", left)
	}
	published := service.chat.(*turnChat).published
	if len(published) != 1 || published[0].Type != agent.EventPowerConfirmation ||
		published[0].Confirmation == nil || published[0].Confirmation.State != confirmationPending {
		t.Fatalf("published = %+v, want the pending press", published)
	}

	// A second press while one waits is refused, not queued.
	result := toolResult(t, callTool(t, service, "kvm_power", `{"button":"reset"}`, nil))
	if !isErrorResult(result) || !strings.Contains(toolText(t, result), id) {
		t.Fatalf("second press = %v, want refused naming %s", result, id)
	}
	if len(host.pressed) != 0 || len(service.confirmations.pending()) != 1 {
		t.Fatalf("pressed %v, pending %+v", host.pressed, service.confirmations.pending())
	}
}

// A confirmation presses exactly once, and the agent waiting for it learns
// it was pressed. A second confirmation is refused.
func TestAConfirmationPressesOnce(t *testing.T) {
	service, host := newDeviceTestService(t)
	withAgentTiers(t, audit.TierPower)

	id := requestPress(t, service, `{"button":"reset"}`)
	answers := make(chan map[string]any, 1)
	go func() {
		answer, _ := awaitAnswer(t, service, id, nil)
		answers <- answer
	}()

	confirmed, code, message := service.confirmPowerPress(id, "alice")
	if code != 0 || confirmed.State != confirmationPressed || confirmed.DecidedBy != "alice" {
		t.Fatalf("confirm = %+v %d %s", confirmed, code, message)
	}
	select {
	case answer := <-answers:
		if answer["state"] != confirmationPressed || answer["pressed"] != true {
			t.Fatalf("agent's answer = %v, want pressed", answer)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the agent's wait did not end with the confirmation")
	}

	again, code, _ := service.confirmPowerPress(id, "bob")
	if code != confirmationNotPending || again.State != confirmationPressed {
		t.Fatalf("second confirm = %+v %d, want refused", again, code)
	}
	if _, code, _ := service.denyPowerPress(id, "bob"); code != confirmationNotPending {
		t.Fatalf("deny after confirm = %d, want refused", code)
	}
	if len(host.pressed) != 1 {
		t.Fatalf("pressed = %v, want one press", host.pressed)
	}
	if pending := service.confirmations.pending(); len(pending) != 0 {
		t.Fatalf("pending = %+v after the press", pending)
	}
	if _, code, _ := service.confirmPowerPress("no-such-id", "alice"); code != confirmationNotFound {
		t.Fatalf("unknown id = %d", code)
	}
}

// A denial presses nothing and tells the agent.
func TestADenialTellsTheAgent(t *testing.T) {
	service, host := newDeviceTestService(t)
	withAgentTiers(t, audit.TierPower)

	id := requestPress(t, service, `{"button":"power"}`)
	denied, code, _ := service.denyPowerPress(id, "alice")
	if code != 0 || denied.State != confirmationDenied || denied.DecidedBy != "alice" {
		t.Fatalf("deny = %+v %d", denied, code)
	}
	answer, isError := awaitAnswer(t, service, id, nil)
	if isError || answer["state"] != confirmationDenied || answer["pressed"] != false || answer["message"] != confirmationDeniedMessage {
		t.Fatalf("agent's answer = %v", answer)
	}
	if _, code, _ := service.confirmPowerPress(id, "bob"); code != confirmationNotPending {
		t.Fatalf("confirm after deny = %d, want refused", code)
	}
	if len(host.pressed) != 0 {
		t.Fatalf("pressed %v after a denial", host.pressed)
	}
}

// A press nobody confirms expires: the agent learns it, and a late
// confirmation presses nothing.
func TestAnUnconfirmedPressExpires(t *testing.T) {
	service, host := newDeviceTestService(t)
	withAgentTiers(t, audit.TierPower)
	shortWaits(t, 30*time.Millisecond, 5*time.Second)

	id := requestPress(t, service, `{"button":"power"}`)
	answer, _ := awaitAnswer(t, service, id, nil)
	if answer["state"] != confirmationExpired || answer["message"] != confirmationTimedOutMessage {
		t.Fatalf("agent's answer = %v, want expired", answer)
	}
	if _, code, _ := service.confirmPowerPress(id, "alice"); code != confirmationNotPending {
		t.Fatalf("late confirm = %d, want refused", code)
	}
	if len(host.pressed) != 0 {
		t.Fatalf("pressed %v after expiry", host.pressed)
	}
	// The agent may ask again once the first has ended.
	requestPress(t, service, `{"button":"power"}`)
}

// A wait ends after its own limit with the press still pending, so the
// agent calls again within its turn.
func TestAWaitEndsPendingAfterItsLimit(t *testing.T) {
	service, host := newDeviceTestService(t)
	withAgentTiers(t, audit.TierPower)
	shortWaits(t, time.Minute, 20*time.Millisecond)

	id := requestPress(t, service, `{"button":"power"}`)
	answer, isError := awaitAnswer(t, service, id, nil)
	if isError || answer["state"] != confirmationPending || answer["message"] != confirmationWaitingMessage {
		t.Fatalf("agent's answer = %v, want still pending", answer)
	}
	if _, ok := answer["expiresInSeconds"]; !ok {
		t.Fatalf("agent's answer = %v, want the time left", answer)
	}
	if len(host.pressed) != 0 {
		t.Fatalf("pressed %v", host.pressed)
	}
}

// The press belongs to the turn it was asked in. Once that turn ends, a
// waiting agent is told, and a confirmation presses nothing, also in the
// session's next turn.
func TestAPressEndsWithItsTurn(t *testing.T) {
	service, host := newDeviceTestService(t)
	withAgentTiers(t, audit.TierPower)
	shortWaits(t, time.Minute, 5*time.Second)
	chat := service.chat.(*turnChat)

	id := requestPress(t, service, `{"button":"power"}`)
	chat.endTurn(deviceTestSession)
	answer, _ := awaitAnswer(t, service, id, nil)
	if answer["state"] != confirmationExpired || answer["message"] != confirmationTurnEndMessage {
		t.Fatalf("agent's answer = %v, want expired with the turn", answer)
	}

	chat.nextTurn(deviceTestSession)
	id = requestPress(t, service, `{"button":"power"}`)
	chat.endTurn(deviceTestSession)
	chat.nextTurn(deviceTestSession)
	ended, code, _ := service.confirmPowerPress(id, "alice")
	if code != confirmationPressError || ended.State != confirmationExpired {
		t.Fatalf("confirm in a later turn = %+v %d, want refused", ended, code)
	}
	if len(host.pressed) != 0 {
		t.Fatalf("pressed %v for a turn that had ended", host.pressed)
	}
}

// The owner can turn the power tier off while a press waits: the
// confirmation then presses nothing.
func TestAConfirmationNeedsThePowerTierStill(t *testing.T) {
	service, host := newDeviceTestService(t)
	withAgentTiers(t, audit.TierPower)

	id := requestPress(t, service, `{"button":"power"}`)
	withAgentTiers(t, audit.TierObserve)
	ended, code, _ := service.confirmPowerPress(id, "alice")
	if code != confirmationPressError || ended.State != confirmationFailed {
		t.Fatalf("confirm = %+v %d, want refused", ended, code)
	}
	if len(host.pressed) != 0 {
		t.Fatalf("pressed %v with the tier off", host.pressed)
	}
}

// Only the session that asked can wait for its press.
func TestAnotherSessionCannotWaitForAPress(t *testing.T) {
	service, _ := newDeviceTestService(t)
	withAgentTiers(t, audit.TierPower)

	id := requestPress(t, service, `{"button":"power"}`)
	answer, isError := awaitAnswer(t, service, id, map[string]string{agentSessionHeader: "chat-2"})
	if !isError || !strings.Contains(answer["text"].(string), "no press of this chat session") {
		t.Fatalf("answer = %v, want refused", answer)
	}
}
