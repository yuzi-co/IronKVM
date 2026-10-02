package picoclaw

import (
	"reflect"
	"testing"

	"NanoKVM-Server/service/agent"
)

func mustPicoEvent(t *testing.T, raw string) agent.Event {
	t.Helper()
	event, ok := picoEvent([]byte(raw))
	if !ok {
		t.Fatalf("%s gave no event", raw)
	}
	return event
}

func TestPicoMessageIDComesFromThePayload(t *testing.T) {
	if id := mustPicoEvent(t, `{"type":"message.create","payload":{"content":"hi","message_id":"m1"}}`).MessageID; id != "m1" {
		t.Fatalf("id %q", id)
	}
	if id := mustPicoEvent(t, `{"type":"message.create","id":"m2","payload":{}}`).MessageID; id != "m2" {
		t.Fatalf("id %q", id)
	}
	if id := mustPicoEvent(t, `{"type":"message.create","id":"","payload":{}}`).MessageID; id == "" {
		t.Fatal("a message without an id got no id")
	}
}

func TestPicoReasoningAndToolCallsAreHidden(t *testing.T) {
	for _, raw := range []string{
		`{"type":"message.create","payload":{"kind":"thought","thought":true,"content":"x"}}`,
		`{"type":"message.create","payload":{"thought":true,"content":"x"}}`,
		`{"type":"message.update","payload":{"kind":"tool_calls","content":""}}`,
	} {
		if kind := mustPicoEvent(t, raw).Kind; kind != agent.KindHidden {
			t.Fatalf("%s: kind %q", raw, kind)
		}
	}
}

func TestPicoPlaceholderAndReply(t *testing.T) {
	event := mustPicoEvent(t, `{"type":"message.create","payload":{"placeholder":true,"content":"Thinking"}}`)
	if event.Kind != agent.KindPlaceholder || event.Text != "Thinking" {
		t.Fatalf("%+v", event)
	}
	event = mustPicoEvent(t, `{"type":"message.update","payload":{"content":" Done ","context_usage":{}}}`)
	if event.Type != agent.EventAgentMessage || event.Kind != agent.KindReply || event.Text != "Done" {
		t.Fatalf("%+v", event)
	}
	if text := mustPicoEvent(t, `{"type":"message.create","content":"null"}`).Text; text != "" {
		t.Fatalf("text %q", text)
	}
	if text := mustPicoEvent(t, `{"type":"message.create","payload":{"content":["a",{"text":"b"},{"x":1}]}}`).Text; text != "a\nb" {
		t.Fatalf("text %q", text)
	}
}

func TestPicoErrorFields(t *testing.T) {
	event := mustPicoEvent(t, `{"type":"error","payload":{"code":"command_disabled","message":"control commands are disabled","request_id":"r1"}}`)
	if event.Type != agent.EventError || event.Code != "command_disabled" || event.Message != "control commands are disabled" || event.RequestID != "r1" {
		t.Fatalf("%+v", event)
	}
	event = mustPicoEvent(t, `{"type":"error","code":"AI_LOCK_HELD","message":"held"}`)
	if event.Code != "AI_LOCK_HELD" || event.Message != "held" || event.RequestID != "" {
		t.Fatalf("%+v", event)
	}
	event = mustPicoEvent(t, `{"type":"error"}`)
	if event.Code != "ERROR" || event.Message != "Gateway error" {
		t.Fatalf("%+v", event)
	}
}

func TestPicoTurnDone(t *testing.T) {
	event := mustPicoEvent(t, `{"type":"turn.done","payload":{"request_id":"req-1","request_ids":["req-1","req-2"],"status":"ok","usage":{"input_tokens":4380,"output_tokens":328,"total_tokens":4708,"llm_calls":2}}}`)
	llmCalls := 2
	want := agent.Event{
		Type:       agent.EventTurnDone,
		RequestIDs: []string{"req-1", "req-2"},
		StopReason: agent.StopEndTurn,
		Usage:      &agent.Usage{InputTokens: 4380, OutputTokens: 328, TotalTokens: 4708, LLMCalls: &llmCalls},
	}
	if !reflect.DeepEqual(event, want) {
		t.Fatalf("got %+v", event)
	}

	event = mustPicoEvent(t, `{"type":"turn.done","payload":{"status":"error"}}`)
	if event.StopReason != agent.StopError || len(event.RequestIDs) != 0 || event.Usage != nil {
		t.Fatalf("%+v", event)
	}
	if reason := mustPicoEvent(t, `{"type":"turn.done","payload":{"status":"canceled"}}`).StopReason; reason != agent.StopCancelled {
		t.Fatalf("stop reason %q", reason)
	}
	if reason := mustPicoEvent(t, `{"type":"turn.done","payload":{"status":"odd"}}`).StopReason; reason != agent.StopEndTurn {
		t.Fatalf("stop reason %q", reason)
	}
	if ids := mustPicoEvent(t, `{"type":"turn.done","payload":{"request_id":"r"}}`).RequestIDs; !reflect.DeepEqual(ids, []string{"r"}) {
		t.Fatalf("ids %v", ids)
	}
}

func TestPicoOtherMessages(t *testing.T) {
	if event := mustPicoEvent(t, `{"type":"typing.start"}`); event.Type != agent.EventTurnStarted {
		t.Fatalf("%+v", event)
	}
	for _, raw := range []string{`{"type":"typing.stop"}`, `{"type":"pong"}`, `{"type":"message.delete","payload":{}}`, `{"type":"other"}`} {
		if event, ok := picoEvent([]byte(raw)); ok {
			t.Fatalf("%s gave %+v", raw, event)
		}
	}
	if event := mustPicoEvent(t, `{"type":"message.delete","payload":{"message_id":"m1"}}`); event.Type != agent.EventAgentMessageRemoved || event.MessageID != "m1" {
		t.Fatalf("%+v", event)
	}
	event := mustPicoEvent(t, `{"type":"media.create","id":"o1","payload":{"content":"shot","data":{"image_base64":"abc"}}}`)
	if event.Type != agent.EventObservation || event.MessageID != "o1" || event.ImageBase64 != "abc" || event.Text != "shot" {
		t.Fatalf("%+v", event)
	}
	event = mustPicoEvent(t, `{"type":"tool","id":"t1","payload":{"action":"click","x":0.5,"y":0.25}}`)
	if event.Type != agent.EventToolCall || event.Title != "click" || *event.X != 0.5 || *event.Y != 0.25 {
		t.Fatalf("%+v", event)
	}
	if event := mustPicoEvent(t, `not json`); event.Type != agent.EventError || event.Code != "INVALID_MESSAGE" {
		t.Fatalf("%+v", event)
	}
}

// The capture lease follows the turn when PicoClaw's real message sequence
// goes through the translation: PicoClaw sends no top-level id, and only
// turn.done ends the lease.
func TestPicoTurnHoldsTheCaptureLeaseUntilTurnDone(t *testing.T) {
	released := false
	leases := agent.NewTaskLeases(func() func() { return func() { released = true } })
	leases.Prompt("sess", "req-1", 0)
	for _, raw := range []string{
		`{"type":"typing.start"}`,
		`{"type":"message.create","payload":{"message_id":"m1","kind":"thought","thought":true,"content":"x"}}`,
		`{"type":"message.create","payload":{"message_id":"m2","kind":"tool_calls","content":""}}`,
		`{"type":"message.create","payload":{"message_id":"m3","placeholder":true,"content":"Thinking"}}`,
		`{"type":"typing.stop"}`,
	} {
		if event, ok := picoEvent([]byte(raw)); ok {
			leases.Observe("sess", event)
		}
		if released {
			t.Fatalf("lease released before turn.done, at %s", raw)
		}
	}
	leases.Observe("sess", mustPicoEvent(t, `{"type":"turn.done","payload":{"request_ids":["req-1"],"status":"ok"}}`))
	if !released {
		t.Fatal("turn.done did not release the lease")
	}
}
