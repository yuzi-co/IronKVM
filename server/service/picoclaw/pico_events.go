package picoclaw

import (
	"encoding/json"
	"strings"

	"NanoKVM-Server/service/agent"

	"github.com/google/uuid"
)

// picoEvent turns one pico protocol message from PicoClaw into a normalized
// chat event. ok is false for a message the browser does not need (pong,
// typing.stop, anything unknown).
//
// PicoClaw 0.3 puts the message id, the error code and the error text in the
// payload, and sends its reasoning and its tool calls as messages of their
// own, marked by payload.kind. Older builds kept some of them at the top
// level, so both places are read.
func picoEvent(data []byte) (event agent.Event, ok bool) {
	var message map[string]any
	if err := json.Unmarshal(data, &message); err != nil {
		return agent.Event{
			Type:    agent.EventError,
			Code:    "INVALID_MESSAGE",
			Message: "Failed to parse gateway message",
		}, true
	}
	payload, _ := message["payload"].(map[string]any)
	if payload == nil {
		payload = map[string]any{}
	}

	switch messageType, _ := message["type"].(string); messageType {
	case "typing.start":
		return agent.Event{Type: agent.EventTurnStarted}, true
	case "typing.stop", "pong":
		return agent.Event{}, false
	case "error":
		return agent.Event{
			Type:      agent.EventError,
			Code:      firstString(message["code"], payload["code"], "ERROR"),
			Message:   firstString(message["message"], payload["message"], "Gateway error"),
			RequestID: nonEmptyString(payload["request_id"]),
		}, true
	case "turn.done":
		return picoTurnDone(payload), true
	case "control.mode_changed":
		// Only the server sends this; an agent that sent it would change
		// nothing.
		return agent.Event{}, false
	case "message.create", "message.update":
		return agent.Event{
			Type:      agent.EventAgentMessage,
			MessageID: firstString(payload["message_id"], message["id"], uuid.NewString()),
			Kind:      picoMessageKind(payload),
			Text:      picoText(message, payload),
		}, true
	case "message.delete":
		id := firstString(payload["message_id"], message["id"], "")
		if id == "" {
			return agent.Event{}, false
		}
		return agent.Event{Type: agent.EventAgentMessageRemoved, MessageID: id}, true
	}

	if image := picoImage(message, payload); image != "" {
		return agent.Event{
			Type:        agent.EventObservation,
			MessageID:   firstString(message["id"], uuid.NewString()),
			Text:        picoText(message, payload),
			ImageBase64: image,
		}, true
	}
	if action := picoAction(message, payload); action != "" {
		event := agent.Event{
			Type:       agent.EventToolCall,
			ToolCallID: firstString(message["id"], uuid.NewString()),
			Title:      action,
		}
		if x, isNumber := payload["x"].(float64); isNumber {
			event.X = &x
		}
		if y, isNumber := payload["y"].(float64); isNumber {
			event.Y = &y
		}
		return event, true
	}
	return agent.Event{}, false
}

// picoMessageKind: reasoning and tool calls are hidden; a placeholder is a
// "thinking" text a reply replaces; anything else is the reply.
func picoMessageKind(payload map[string]any) agent.MessageKind {
	kind, _ := payload["kind"].(string)
	thought, _ := payload["thought"].(bool)
	if kind == "thought" || kind == "tool_calls" || thought {
		return agent.KindHidden
	}
	if placeholder, _ := payload["placeholder"].(bool); placeholder {
		return agent.KindPlaceholder
	}
	return agent.KindReply
}

func picoTurnDone(payload map[string]any) agent.Event {
	ids := []string{}
	if list, isList := payload["request_ids"].([]any); isList {
		for _, item := range list {
			if id := nonEmptyString(item); id != "" {
				ids = append(ids, id)
			}
		}
	}
	if first := nonEmptyString(payload["request_id"]); first != "" && !containsString(ids, first) {
		ids = append([]string{first}, ids...)
	}

	stopReason := agent.StopEndTurn
	switch payload["status"] {
	case "error":
		stopReason = agent.StopError
	case "canceled":
		stopReason = agent.StopCancelled
	}

	event := agent.Event{Type: agent.EventTurnDone, RequestIDs: ids, StopReason: stopReason}
	if usage, isObject := payload["usage"].(map[string]any); isObject {
		count := func(key string) int {
			value, _ := usage[key].(float64)
			return int(value)
		}
		event.Usage = &agent.Usage{
			InputTokens:  count("input_tokens"),
			OutputTokens: count("output_tokens"),
			TotalTokens:  count("total_tokens"),
		}
		if calls, isNumber := usage["llm_calls"].(float64); isNumber {
			llmCalls := int(calls)
			event.Usage.LLMCalls = &llmCalls
		}
	}
	return event
}

// picoText reads the text of a message: a string, or a list of strings and
// {text} parts joined by newlines. "null" and "undefined" count as no text.
func picoText(message, payload map[string]any) string {
	content := firstPresent(payload["content"], message["content"], payload["text"], message["text"])
	switch value := content.(type) {
	case string:
		return literalText(value)
	case []any:
		parts := make([]string, 0, len(value))
		for _, item := range value {
			text := ""
			switch part := item.(type) {
			case string:
				text = literalText(part)
			case map[string]any:
				if raw, has := part["text"]; has {
					text = literalText(stringOf(raw))
				}
			}
			if text != "" {
				parts = append(parts, text)
			}
		}
		return strings.TrimSpace(strings.Join(parts, "\n"))
	}
	return ""
}

func picoImage(message, payload map[string]any) string {
	data, _ := firstPresent(payload["data"], message["data"]).(map[string]any)
	if data == nil {
		data = map[string]any{}
	}
	image, _ := firstTruthy(data["image_base64"], payload["image_base64"], message["image_base64"]).(string)
	return image
}

func picoAction(message, payload map[string]any) string {
	action, _ := firstTruthy(payload["action"], message["action"], payload["tool_name"], message["tool_name"]).(string)
	return literalText(action)
}

func literalText(value string) string {
	trimmed := strings.TrimSpace(value)
	switch strings.ToLower(trimmed) {
	case "null", "undefined":
		return ""
	}
	return trimmed
}

// firstPresent is the JavaScript a ?? b: the first value that is not null.
func firstPresent(values ...any) any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

// firstTruthy is the JavaScript a || b for strings and objects.
func firstTruthy(values ...any) any {
	for _, value := range values {
		switch typed := value.(type) {
		case nil:
			continue
		case string:
			if typed == "" {
				continue
			}
		case bool:
			if !typed {
				continue
			}
		case float64:
			if typed == 0 {
				continue
			}
		}
		return value
	}
	return nil
}

// firstString returns the first non-empty string among values; the last
// value is the fallback and is returned as it is.
func firstString(values ...any) string {
	for _, value := range values {
		if text := nonEmptyString(value); text != "" {
			return text
		}
	}
	return ""
}

func nonEmptyString(value any) string {
	text, _ := value.(string)
	return text
}

func stringOf(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case nil:
		return ""
	default:
		raw, _ := json.Marshal(typed)
		return string(raw)
	}
}

func containsString(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
