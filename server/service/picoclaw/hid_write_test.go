package picoclaw

import (
	"errors"
	"strings"
	"testing"
)

// errHIDTimeout is what a write to a gadget endpoint the host is not polling
// returns once its deadline passes.
var errHIDTimeout = errors.New("write /dev/hidg2: i/o timeout")

func TestMoveFailsAndKeepsThePointerWhenTheWriteTimesOut(t *testing.T) {
	service, _, hid := newMCPActionTestService(t)
	actionResultOf(t, callKVMActions(t, service, `{"actions":[{"action":"move","x":0.25,"y":0.5}]}`))

	hid.mu.Lock()
	hid.mouseErr = errHIDTimeout
	hid.mu.Unlock()
	text := errorTextOf(t, callKVMActions(t, service, `{"actions":[{"action":"move","dx":10}]}`))
	if !strings.Contains(text, `action 1 of 1 ("move") failed`) || !strings.Contains(text, "i/o timeout") ||
		!strings.Contains(text, "did not take effect") {
		t.Fatalf("error = %q, want the failed action and the write error", text)
	}
	if x, y, known := service.pointer.get(); !known || x != 0.25 || y != 0.5 {
		t.Fatalf("pointer = %v,%v (known %v), want it to stay at 0.25,0.5", x, y, known)
	}

	hid.mu.Lock()
	hid.mouseErr = nil
	hid.mu.Unlock()
	result := actionResultOf(t, callKVMActions(t, service, `{"actions":[{"action":"move","dx":10}]}`))
	if result.Pointer == nil || result.Pointer.PX != 490 {
		t.Fatalf("pointer after the retried move = %+v, want px 490", result.Pointer)
	}
}

func TestBatchStopsAtTheActionWhoseWriteFails(t *testing.T) {
	service, _, hid := newMCPActionTestService(t)
	hid.keyErr = errors.New("write /dev/hidg0: i/o timeout")

	text := errorTextOf(t, callKVMActions(t, service,
		`{"actions":[{"action":"move","x":0.5,"y":0.5},{"action":"hotkey","keys":["shift"]},{"action":"click"}]}`))
	if !strings.Contains(text, `action 2 of 3 ("hotkey") failed`) || !strings.Contains(text, "keyboard report") ||
		!strings.Contains(text, "hidg0: i/o timeout") {
		t.Fatalf("error = %q", text)
	}
	if len(hid.mouse) != 1 {
		t.Fatalf("mouse reports = %d, want only the move before the failed hotkey", len(hid.mouse))
	}
}

func TestSessionEndWritesNothingWhenNothingIsHeld(t *testing.T) {
	service, _, hid := newMCPActionTestService(t)
	actionResultOf(t, callKVMActions(t, service, `{"actions":[{"action":"click","x":0.5,"y":0.5},{"action":"hotkey","keys":["shift"]}]}`))
	keys, mouse := len(hid.keys), len(hid.mouse)
	// From here on every write would time out, as on an endpoint the host
	// does not poll.
	hid.keyErr, hid.mouseErr = errHIDTimeout, errHIDTimeout

	if _, err := service.lock.acquire("test-session"); err != nil {
		t.Fatal(err)
	}
	if err := service.releaseHeldInput(); err != nil {
		t.Fatalf("release after a turn that left nothing held = %v, want no writes", err)
	}
	service.releaseGatewaySession("test-session")
	if len(hid.keys) != keys || len(hid.mouse) != mouse {
		t.Fatal("session end wrote HID reports although nothing was held")
	}
}

func TestSessionEndReleasesWhatIsHeld(t *testing.T) {
	service, _, hid := newMCPActionTestService(t)
	actionResultOf(t, callKVMActions(t, service, `{"actions":[{"action":"move","x":0.25,"y":0.75}]}`))
	// A key and a button left down, as by an action cut off halfway.
	if err := service.sendKeyboardReport([]byte{0x02, 0, 0, 0, 0, 0, 0, 0}); err != nil {
		t.Fatal(err)
	}
	if err := service.sendMouseMoveWithButton(0.25, 0.75, 0x01, 0); err != nil {
		t.Fatal(err)
	}

	if err := service.releaseHeldInput(); err != nil {
		t.Fatal(err)
	}
	if last := hid.keys[len(hid.keys)-1]; strings.Trim(string(last), "\x00") != "" {
		t.Fatalf("last keyboard report = %v, want all keys up", last)
	}
	last := hid.lastMouse()
	if last[0] != 0 {
		t.Fatalf("last mouse report = %v, want all buttons up", last)
	}
	if x, y := absoluteXY(last); x != toAbsoluteHidCoord(0.25) || y != toAbsoluteHidCoord(0.75) {
		t.Fatalf("button release at %d,%d, want it where the pointer is", x, y)
	}
	if keys, buttons := service.held.get(); keys || buttons {
		t.Fatal("input still recorded as held after the release")
	}
}

func TestFailedReleaseIsReportedAndRetried(t *testing.T) {
	service, _, hid := newMCPActionTestService(t)
	if err := service.sendKeyboardReport([]byte{0x02, 0, 0, 0, 0, 0, 0, 0}); err != nil {
		t.Fatal(err)
	}
	hid.keyErr = errHIDTimeout
	if err := service.releaseHeldInput(); err == nil || !strings.Contains(err.Error(), "release keyboard") {
		t.Fatalf("release = %v, want the keyboard write error", err)
	}
	hid.keyErr = nil
	if err := service.releaseHeldInput(); err != nil {
		t.Fatal(err)
	}
	if keys, _ := service.held.get(); keys {
		t.Fatal("keys still recorded as held after a successful retry")
	}
}

func TestEmptyTypePointsToHotkeyForKeys(t *testing.T) {
	service, _, hid := newMCPActionTestService(t)

	text := errorTextOf(t, callKVMActions(t, service, `{"actions":[{"action":"type","text":""}]}`))
	if !strings.Contains(text, `type needs "text"`) || !strings.Contains(text, `{"action":"hotkey","keys":["shift"]}`) {
		t.Fatalf("error = %q, want it to point a key press to hotkey", text)
	}
	if len(hid.keys) != 0 {
		t.Fatal("an empty type action wrote keyboard reports")
	}
}

func TestHotkeyPressesASingleModifier(t *testing.T) {
	service, _, hid := newMCPActionTestService(t)

	actionResultOf(t, callKVMActions(t, service, `{"actions":[{"action":"hotkey","keys":["shift"]}]}`))
	if len(hid.keys) != 2 || hid.keys[0][0] == 0 || strings.Trim(string(hid.keys[0][1:]), "\x00") != "" ||
		strings.Trim(string(hid.keys[1]), "\x00") != "" {
		t.Fatalf("keyboard reports = %v, want Shift down then all keys up", hid.keys)
	}
}
