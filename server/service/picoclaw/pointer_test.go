package picoclaw

import (
	"encoding/json"
	"strings"
	"testing"
)

func actionResultOf(t *testing.T, result mcpToolResult) ActionResult {
	t.Helper()
	if result.IsError || len(result.Content) == 0 {
		t.Fatalf("result = %+v, want success", result)
	}
	var actionResult ActionResult
	if err := json.Unmarshal([]byte(result.Content[0].Text), &actionResult); err != nil {
		t.Fatalf("action result %q: %v", result.Content[0].Text, err)
	}
	return actionResult
}

func errorTextOf(t *testing.T, result mcpToolResult) string {
	t.Helper()
	if !result.IsError || len(result.Content) == 0 {
		t.Fatalf("result = %+v, want an error", result)
	}
	return result.Content[0].Text
}

func absoluteXY(report []byte) (uint16, uint16) {
	return uint16(report[1]) | uint16(report[2])<<8, uint16(report[3]) | uint16(report[4])<<8
}

func TestRelativeMoveIsMeasuredInScreenPixelsFromThePointer(t *testing.T) {
	service, _, hid := newMCPActionTestService(t)

	first := actionResultOf(t, callKVMActions(t, service, `{"actions":[{"action":"move","x":0.25,"y":0.5}]}`))
	if first.Pointer == nil || first.Pointer.PX != 480 || first.Pointer.PY != 540 {
		t.Fatalf("pointer after absolute move = %+v, want px 480, py 540", first.Pointer)
	}

	second := actionResultOf(t, callKVMActions(t, service, `{"actions":[{"action":"move","dx":10}]}`))
	if second.Pointer == nil || second.Pointer.PX != 490 || second.Pointer.PY != 540 {
		t.Fatalf("pointer after dx 10 = %+v, want px 490, py 540", second.Pointer)
	}
	wantX := toAbsoluteHidCoord(0.25 + 10.0/1919)
	if x, y := absoluteXY(hid.lastMouse()); x != wantX || y != toAbsoluteHidCoord(0.5) {
		t.Fatalf("HID report at %d,%d, want %d,%d", x, y, wantX, toAbsoluteHidCoord(0.5))
	}

	third := actionResultOf(t, callKVMActions(t, service, `{"actions":[{"action":"move","dx":-5000,"dy":-20}]}`))
	if third.Pointer == nil || third.Pointer.PX != 0 || third.Pointer.PY != 520 {
		t.Fatalf("pointer after dx -5000, dy -20 = %+v, want the left edge at py 520", third.Pointer)
	}
}

func TestRelativeMoveNeedsAKnownPointer(t *testing.T) {
	service, _, hid := newMCPActionTestService(t)

	text := errorTextOf(t, callKVMActions(t, service, `{"actions":[{"action":"move","dx":10}]}`))
	if !strings.Contains(text, "not known yet") || !strings.Contains(text, `First move or click with "x" and "y"`) {
		t.Fatalf("error = %q", text)
	}
	if len(hid.mouse) > 1 {
		t.Fatalf("sent %d mouse reports for a refused move", len(hid.mouse))
	}
}

func TestPixelCoordinatesAreRefusedWithAnExplanation(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)

	text := errorTextOf(t, callKVMActions(t, service, `{"actions":[{"action":"wait","duration_ms":0},{"action":"move","x":480,"y":270}]}`))
	for _, want := range []string{
		`action 2 of 2 ("move") failed`,
		`"x" is 480, but it must be a fraction from 0 to 1, not pixels`,
		"divide its pixel position by the image width and height",
		`use "dx"/"dy" instead`,
		"Action 1 ran",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("error = %q, want it to contain %q", text, want)
		}
	}
}

func TestMoveWithMisnamedFieldsNamesThem(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)

	text := errorTextOf(t, callKVMActions(t, service, `{"actions":[{"action":"move","coordinate":[10,0],"direction":"right"}]}`))
	for _, want := range []string{
		"move needs coordinates",
		`{"action":"move","dx":10}`,
		`Unknown fields ignored: "coordinate"`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("error = %q, want it to contain %q", text, want)
		}
	}
}

func TestMoveRefusesAbsoluteAndRelativeTogether(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)

	text := errorTextOf(t, callKVMActions(t, service, `{"actions":[{"action":"move","x":0.5,"y":0.5,"dx":10}]}`))
	if !strings.Contains(text, "use only one pair") {
		t.Fatalf("error = %q", text)
	}
}

func TestClickWithoutCoordinatesClicksAtThePointer(t *testing.T) {
	service, _, hid := newMCPActionTestService(t)

	result := actionResultOf(t, callKVMActions(t, service, `{"actions":[{"action":"move","x":0.25,"y":0.5},{"action":"move","dx":10},{"action":"click"}]}`))
	if result.ExecutedActions != 3 {
		t.Fatalf("result = %+v", result)
	}
	wantX := toAbsoluteHidCoord(0.25 + 10.0/1919)
	for _, report := range hid.mouse[len(hid.mouse)-3:] {
		if x, _ := absoluteXY(report); x != wantX {
			t.Fatalf("click report at x %d, want %d", x, wantX)
		}
	}
	if hid.mouse[len(hid.mouse)-2][0] != 1 {
		t.Fatalf("click did not press the left button: %v", hid.mouse[len(hid.mouse)-2])
	}
}

func TestClickWithoutCoordinatesOrPointerExplains(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)

	text := errorTextOf(t, callKVMActions(t, service, `{"actions":[{"action":"click"}]}`))
	if !strings.Contains(text, `click needs "x" and "y"`) || !strings.Contains(text, `{"x":0.5,"y":0.5} the centre`) {
		t.Fatalf("error = %q", text)
	}
}

func TestPointerIsForgottenWhenTheLockIsTakenAfresh(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)

	actionResultOf(t, callKVMActions(t, service, `{"actions":[{"action":"move","x":0.5,"y":0.5}]}`))
	service.lock.Release("test-session")

	text := errorTextOf(t, callKVMActions(t, service, `{"actions":[{"action":"move","dx":10}]}`))
	if !strings.Contains(text, "not known yet") {
		t.Fatalf("error = %q, want the pointer forgotten", text)
	}
}

func TestUnknownActionListsTheValidOnes(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)

	text := errorTextOf(t, callKVMActions(t, service, `{"actions":[{"action":"mouse_move","x":0.5,"y":0.5}]}`))
	if !strings.Contains(text, `unknown action "mouse_move"`) || !strings.Contains(text, "click, move, type, hotkey, scroll, drag, wait") {
		t.Fatalf("error = %q", text)
	}
}

func TestKVMActionsSchemaDescribesRelativeMoves(t *testing.T) {
	schema, err := json.Marshal(mcpToolDefinitions)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"dx"`, `"dy"`, "not pixels", `{\"action\":\"move\",\"dx\":10}`} {
		if !strings.Contains(string(schema), want) {
			t.Fatalf("tool definitions lack %s", want)
		}
	}
}
