package redfish

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

const resetURL = "/redfish/v1/Systems/1/Actions/ComputerSystem.Reset"

// The spec's ResetType table: what each type presses with the LED on, off,
// and unreadable. "" means nothing is pressed; "refused" means the action
// fails because the LED cannot be read.
var resetTable = []struct {
	resetType          string
	ledOn, ledOff, led string
}{
	{"On", "", "power 800ms", "refused"},
	{"ForceOff", "power 5s", "", "refused"},
	{"GracefulShutdown", "power 800ms", "", "refused"},
	{"ForceRestart", "reset 800ms", "reset 800ms", "reset 800ms"},
	{"PushPowerButton", "power 800ms", "power 800ms", "power 800ms"},
}

func TestEachResetTypePressesWhatTheTableSays(t *testing.T) {
	for _, row := range resetTable {
		for _, state := range []struct {
			name string
			led  *bool
			want string
		}{
			{"on", boolPtr(true), row.ledOn},
			{"off", boolPtr(false), row.ledOff},
			{"unknown", nil, row.led},
		} {
			h := newHarness(t)
			h.host.setLED(state.led)

			w := h.do(http.MethodPost, resetURL, `{"ResetType":"`+row.resetType+`"}`, h.admin()...)

			if state.want == "refused" {
				expectError(t, w, http.StatusConflict, "ActionNotSupported")
				if len(h.host.presses) != 0 {
					t.Fatalf("%s, LED %s: pressed %v", row.resetType, state.name, h.host.presses)
				}
				continue
			}

			if w.Code != http.StatusNoContent {
				t.Fatalf("%s, LED %s: status %d: %s", row.resetType, state.name, w.Code, w.Body.String())
			}
			if w.Header().Get("OData-Version") != "4.0" {
				t.Fatalf("%s, LED %s: no OData-Version header on the 204", row.resetType, state.name)
			}
			got := strings.Join(h.host.presses, ",")
			if got != state.want {
				t.Fatalf("%s, LED %s: pressed %q, want %q", row.resetType, state.name, got, state.want)
			}
		}
	}
}

func TestResetTypesTheBoardCannotDoAreRefused(t *testing.T) {
	h := newHarness(t)

	for _, resetType := range []string{"PowerCycle", "Nmi", "GracefulRestart", "ForceOn", "on", ""} {
		w := h.do(http.MethodPost, resetURL, `{"ResetType":"`+resetType+`"}`, h.admin()...)
		expectError(t, w, http.StatusBadRequest, "ActionParameterValueNotInList")
	}
	if len(h.host.presses) != 0 {
		t.Fatalf("pressed %v", h.host.presses)
	}
}

func TestResetChecksItsParameters(t *testing.T) {
	h := newHarness(t)

	expectError(t, h.do(http.MethodPost, resetURL, `{}`, h.admin()...), http.StatusBadRequest, "ActionParameterMissing")
	expectError(t, h.do(http.MethodPost, resetURL, `{"ResetType":1}`, h.admin()...), http.StatusBadRequest, "ActionParameterValueTypeError")
	expectError(t, h.do(http.MethodPost, resetURL, `{"ResetType":"On","Delay":5}`, h.admin()...), http.StatusBadRequest, "ActionParameterNotSupported")
	expectError(t, h.do(http.MethodPost, resetURL, `not json`, h.admin()...), http.StatusBadRequest, "MalformedJSON")
	if len(h.host.presses) != 0 {
		t.Fatalf("pressed %v", h.host.presses)
	}
}

func TestResetNeedsAnAdmin(t *testing.T) {
	h := newHarness(t)

	w := h.do(http.MethodPost, resetURL, `{"ResetType":"ForceRestart"}`, h.user()...)
	expectError(t, w, http.StatusForbidden, "InsufficientPrivilege")
	if len(h.host.presses) != 0 {
		t.Fatalf("pressed %v", h.host.presses)
	}
}

func TestAFailedPressIsAServerError(t *testing.T) {
	h := newHarness(t)
	h.host.pressErr = errors.New("gpio write failed")

	w := h.do(http.MethodPost, resetURL, `{"ResetType":"ForceRestart"}`, h.admin()...)
	expectError(t, w, http.StatusInternalServerError, "GeneralError")
}

func TestSystemReportsThePowerStateFromTheLED(t *testing.T) {
	for _, tc := range []struct {
		led  *bool
		want any
	}{
		{boolPtr(true), "On"},
		{boolPtr(false), "Off"},
		{nil, nil},
	} {
		h := newHarness(t)
		h.host.setLED(tc.led)

		w := h.do(http.MethodGet, "/redfish/v1/Systems/1", "", h.user()...)
		if w.Code != http.StatusOK {
			t.Fatalf("status %d", w.Code)
		}
		body := decode(t, w)
		if body["PowerState"] != tc.want {
			t.Fatalf("PowerState is %v, want %v", body["PowerState"], tc.want)
		}
		if _, ok := body["PowerState"]; !ok {
			t.Fatal("PowerState is missing rather than null")
		}
	}
}

func TestSystemOffersTheResetAction(t *testing.T) {
	h := newHarness(t)

	body := decode(t, h.do(http.MethodGet, "/redfish/v1/Systems/1", "", h.user()...))
	reset := body["Actions"].(map[string]any)["#ComputerSystem.Reset"].(map[string]any)
	if reset["target"] != resetURL {
		t.Fatalf("target is %v", reset["target"])
	}
	var allowed []string
	for _, v := range reset["ResetType@Redfish.AllowableValues"].([]any) {
		allowed = append(allowed, v.(string))
	}
	if strings.Join(allowed, ",") != "On,ForceOff,GracefulShutdown,ForceRestart,PushPowerButton" {
		t.Fatalf("AllowableValues are %v", allowed)
	}
}

func TestSystemETagFollowsThePowerState(t *testing.T) {
	h := newHarness(t)

	on := h.do(http.MethodGet, "/redfish/v1/Systems/1", "", h.user()...).Header().Get("ETag")
	h.host.setLED(boolPtr(false))
	off := h.do(http.MethodGet, "/redfish/v1/Systems/1", "", h.user()...).Header().Get("ETag")

	if on == "" || off == "" || on == off {
		t.Fatalf("ETags %q and %q", on, off)
	}
}

func TestChassisReportsThePowerState(t *testing.T) {
	h := newHarness(t)
	h.host.setLED(boolPtr(false))

	w := h.do(http.MethodGet, "/redfish/v1/Chassis/1", "", h.user()...)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	body := decode(t, w)
	if body["PowerState"] != "Off" || body["ChassisType"] != "Other" {
		t.Fatalf("chassis is %v", body)
	}
}

// Two ForceOff calls at once must not both see the LED on: the second would
// press for 5 s on a host that is already off, and turn it back on.
func TestConcurrentResetsReadTheLEDAfterEachOther(t *testing.T) {
	h := newHarness(t)
	h.host.pressDelay = 50 * time.Millisecond
	h.host.offAfterPower = true

	admin := h.admin()
	var wg sync.WaitGroup
	codes := make([]int, 2)
	for i := range codes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			codes[i] = h.do(http.MethodPost, resetURL, `{"ResetType":"ForceOff"}`, admin...).Code
		}()
	}
	wg.Wait()

	for _, code := range codes {
		if code != http.StatusNoContent {
			t.Fatalf("answers %v, want both 204", codes)
		}
	}
	if len(h.host.presses) != 1 {
		t.Fatalf("presses %v, want exactly one", h.host.presses)
	}
}

// With the LED header not wired, the LED line says "off" whatever the host
// does. The state-dependent types would then press power on a running host
// for On, so they are refused, and press nothing, whatever the line reads.
func TestWithTheLEDNotWiredOnlyTheBlindTypesPress(t *testing.T) {
	for _, row := range resetTable {
		for _, led := range []*bool{boolPtr(true), boolPtr(false)} {
			h := newHarness(t)
			h.ledWired = false
			h.host.setLED(led)

			w := h.do(http.MethodPost, resetURL, `{"ResetType":"`+row.resetType+`"}`, h.admin()...)

			switch row.resetType {
			case "ForceRestart", "PushPowerButton":
				if w.Code != http.StatusNoContent {
					t.Fatalf("%s: status %d: %s", row.resetType, w.Code, w.Body.String())
				}
				if got := strings.Join(h.host.presses, ","); got != row.led {
					t.Fatalf("%s: pressed %q, want %q", row.resetType, got, row.led)
				}
			default:
				expectError(t, w, http.StatusBadRequest, "ActionNotSupported")
				if !strings.Contains(w.Body.String(), "power LED is not connected") {
					t.Fatalf("%s: the message does not say why: %s", row.resetType, w.Body.String())
				}
				if len(h.host.presses) != 0 {
					t.Fatalf("%s: pressed %v", row.resetType, h.host.presses)
				}
			}
		}
	}
}

func TestWithTheLEDNotWiredThePowerStateIsUnknown(t *testing.T) {
	h := newHarness(t)
	h.ledWired = false
	h.host.setLED(boolPtr(false))

	for _, path := range []string{"/redfish/v1/Systems/1", "/redfish/v1/Chassis/1"} {
		body := decode(t, h.do(http.MethodGet, path, "", h.user()...))
		if state, ok := body["PowerState"]; !ok || state != nil {
			t.Fatalf("%s: PowerState is %v, want null", path, body["PowerState"])
		}
	}
}

func TestWithTheLEDNotWiredOnlyTheBlindTypesAreOffered(t *testing.T) {
	h := newHarness(t)
	h.ledWired = false

	body := decode(t, h.do(http.MethodGet, "/redfish/v1/Systems/1", "", h.user()...))
	reset := body["Actions"].(map[string]any)["#ComputerSystem.Reset"].(map[string]any)
	var allowed []string
	for _, v := range reset["ResetType@Redfish.AllowableValues"].([]any) {
		allowed = append(allowed, v.(string))
	}
	if strings.Join(allowed, ",") != "ForceRestart,PushPowerButton" {
		t.Fatalf("AllowableValues are %v", allowed)
	}
}

func TestSystemETagFollowsTheLEDSetting(t *testing.T) {
	h := newHarness(t)

	wired := h.do(http.MethodGet, "/redfish/v1/Systems/1", "", h.user()...).Header().Get("ETag")
	h.ledWired = false
	notWired := h.do(http.MethodGet, "/redfish/v1/Systems/1", "", h.user()...).Header().Get("ETag")

	if wired == "" || notWired == "" || wired == notWired {
		t.Fatalf("ETags %q and %q", wired, notWired)
	}
}
