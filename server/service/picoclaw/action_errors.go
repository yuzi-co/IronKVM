package picoclaw

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// actionFieldNames are the JSON fields an action object may carry.
var actionFieldNames = func() map[string]struct{} {
	names := make(map[string]struct{})
	actionType := reflect.TypeOf(Action{})
	for i := 0; i < actionType.NumField(); i++ {
		name, _, _ := strings.Cut(actionType.Field(i).Tag.Get("json"), ",")
		if name != "" && name != "-" {
			names[name] = struct{}{}
		}
	}
	return names
}()

// describeActionError turns an action failure into the message the model
// reads: which action failed, and any fields it sent that the server ignores,
// since a misnamed field is the usual reason a small model's action fails.
func describeActionError(err *PicoclawError, actions []Action, args json.RawMessage) string {
	if err == nil {
		return ""
	}
	if err.Index == nil || *err.Index < 0 || *err.Index >= len(actions) {
		return err.Message
	}

	index := *err.Index
	message := fmt.Sprintf("action %d of %d (%q) failed: %s", index+1, len(actions), actions[index].Action, err.Message)
	if unknown := unknownActionFields(args, index); len(unknown) > 0 {
		message += fmt.Sprintf(". Unknown fields ignored: %s", strings.Join(unknown, ", "))
	}
	switch {
	case index == 1:
		message += ". Action 1 ran"
	case index > 1:
		message += fmt.Sprintf(". Actions 1 to %d ran", index)
	}
	return message
}

func unknownActionFields(args json.RawMessage, index int) []string {
	var raw struct {
		Actions []map[string]json.RawMessage `json:"actions"`
	}
	if json.Unmarshal(args, &raw) != nil || index >= len(raw.Actions) {
		return nil
	}
	var unknown []string
	for name := range raw.Actions[index] {
		if _, ok := actionFieldNames[name]; !ok {
			unknown = append(unknown, fmt.Sprintf("%q", name))
		}
	}
	sort.Strings(unknown)
	return unknown
}
