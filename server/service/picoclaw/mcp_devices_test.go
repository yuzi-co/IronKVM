package picoclaw

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/service/controlmode"
	"NanoKVM-Server/service/hostdevices"
	"NanoKVM-Server/service/hostpower"
)

// fakeHost is the board's buttons, power LED and drives as the agent's tools
// see them.
type fakeHost struct {
	mu       sync.Mutex
	pressed  []string
	inserted []string
	ejected  []string
	mounted  []string
	mountErr error
	// hold, when set, keeps a press down until it is closed.
	hold    chan struct{}
	holding chan struct{}
}

func (h *fakeHost) devices() hostdevices.Devices {
	return hostdevices.Devices{
		PressButton: func(kind string, d time.Duration) error {
			if h.hold != nil {
				h.holding <- struct{}{}
				<-h.hold
			}
			h.mu.Lock()
			defer h.mu.Unlock()
			h.pressed = append(h.pressed, kind+" "+d.String())
			return nil
		},
		PowerLEDConnected: func() bool { return true },
		PowerLED:          func() (bool, error) { return true, nil },
		ListMedia: func() ([]hostdevices.MediaDrive, []string, error) {
			return []hostdevices.MediaDrive{
				{ID: "disk"},
				{ID: "cdrom", File: "/data/win11.iso", ReadOnly: true},
			}, []string{"/data/win11.iso", "/data/rescue.img"}, nil
		},
		InsertMedia: func(drive, file string, readOnly bool) error {
			h.mu.Lock()
			defer h.mu.Unlock()
			h.inserted = append(h.inserted, drive+" "+file+" ro="+map[bool]string{true: "1", false: "0"}[readOnly])
			return nil
		},
		EjectMedia: func(drive string) error {
			h.mu.Lock()
			defer h.mu.Unlock()
			h.ejected = append(h.ejected, drive)
			return nil
		},
		MountURL: func(drive, url string) error {
			h.mu.Lock()
			defer h.mu.Unlock()
			h.mounted = append(h.mounted, drive+" "+url)
			return h.mountErr
		},
	}
}

// deviceTestSession is the chat session that holds the AI input lock in the
// device tests, as the bridge's session does while the user chats.
const deviceTestSession = "chat-1"

func newDeviceTestService(t *testing.T) (*Service, *fakeHost) {
	t.Helper()
	service, _, _ := newMCPActionTestService(t)
	host := &fakeHost{}
	service.SetDevices(host.devices())
	if err := service.lock.Ensure(deviceTestSession); err != nil {
		t.Fatal(err)
	}
	return service, host
}

func listedTools(t *testing.T, service *Service) []string {
	t.Helper()
	result := toolResult(t, mcpRPC(t, service, `{"jsonrpc":"2.0","id":7,"method":"tools/list"}`, nil))
	var names []string
	for _, tool := range result["tools"].([]any) {
		names = append(names, tool.(map[string]any)["name"].(string))
	}
	return names
}

func toolText(t *testing.T, result map[string]any) string {
	t.Helper()
	content := resultContent(t, result)
	if len(content) == 0 {
		t.Fatalf("result = %v, want text", result)
	}
	text, _ := content[0]["text"].(string)
	return text
}

func TestDeviceToolsAreListedOnlyWithTheirTiers(t *testing.T) {
	service, _ := newDeviceTestService(t)

	withAgentTiers(t, audit.TierObserve)
	if got, want := listedTools(t, service), []string{"kvm_screenshot", "kvm_power_status", "kvm_media_list"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("observe only: tools = %v, want %v", got, want)
	}

	withAgentTiers(t, audit.TierObserve, audit.TierInput, audit.TierPower)
	if got, want := listedTools(t, service), []string{"kvm_screenshot", "kvm_actions", "kvm_power_status", "kvm_power", "kvm_media_list"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("with power: tools = %v, want %v", got, want)
	}

	withAgentTiers(t, audit.TierObserve, audit.TierInput, audit.TierPower, audit.TierMedia)
	if got, want := listedTools(t, service), []string{"kvm_screenshot", "kvm_actions", "kvm_power_status", "kvm_power", "kvm_media_list", "kvm_media"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("all tiers: tools = %v, want %v", got, want)
	}
}

// A board whose router gave the agent no devices offers no device tools, even
// with every tier on.
func TestDeviceToolsAreLeftOutWithoutDevices(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)
	withAgentTiers(t, audit.TierObserve, audit.TierInput, audit.TierPower, audit.TierMedia)

	if got, want := listedTools(t, service), []string{"kvm_screenshot", "kvm_actions"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("tools = %v, want %v", got, want)
	}
}

// The agent's tiers default to observe and input: power and media are off
// until the owner turns them on.
func TestAgentPowerAndMediaAreOffByDefault(t *testing.T) {
	service, _ := newDeviceTestService(t)
	withAgentTiers(t, audit.DefaultTiers().Agent...)

	for _, name := range listedTools(t, service) {
		if name == "kvm_power" || name == "kvm_media" {
			t.Fatalf("%s listed with the default tiers", name)
		}
	}
}

func TestPowerIsRefusedWithThePowerTierOff(t *testing.T) {
	service, host := newDeviceTestService(t)
	store := withAgentTiers(t, audit.TierObserve, audit.TierInput, audit.TierMedia)

	result := toolResult(t, callTool(t, service, "kvm_power", `{"button":"power","press":"long"}`, nil))
	if !isErrorResult(result) || toolText(t, result) != audit.DeniedMessage(audit.SubjectAgent, audit.TierPower) {
		t.Fatalf("result = %v, want the denied message", result)
	}
	if len(host.pressed) != 0 {
		t.Fatalf("pressed %v with the tier off", host.pressed)
	}
	events := agentEvents(t, store)
	if len(events) != 1 || events[0].Outcome != audit.OutcomeDenied || events[0].Tier != "power" || events[0].Action != "kvm_power" {
		t.Fatalf("events = %+v", events)
	}
}

func TestMediaIsRefusedWithTheMediaTierOff(t *testing.T) {
	service, host := newDeviceTestService(t)
	store := withAgentTiers(t, audit.TierObserve, audit.TierInput, audit.TierPower)

	result := toolResult(t, callTool(t, service, "kvm_media", `{"action":"insert","drive":"cdrom","file":"/data/win11.iso"}`, nil))
	if !isErrorResult(result) || toolText(t, result) != audit.DeniedMessage(audit.SubjectAgent, audit.TierMedia) {
		t.Fatalf("result = %v, want the denied message", result)
	}
	if len(host.inserted) != 0 {
		t.Fatalf("inserted %v with the tier off", host.inserted)
	}
	events := agentEvents(t, store)
	if len(events) != 1 || events[0].Outcome != audit.OutcomeDenied || events[0].Tier != "media" {
		t.Fatalf("events = %+v", events)
	}
}

func TestPowerStatusReportsTheLED(t *testing.T) {
	service, _ := newDeviceTestService(t)
	store := withAgentTiers(t, audit.TierObserve)

	result := toolResult(t, callTool(t, service, "kvm_power_status", `{}`, nil))
	if isErrorResult(result) {
		t.Fatalf("result = %v", result)
	}
	var status map[string]any
	if err := json.Unmarshal([]byte(toolText(t, result)), &status); err != nil {
		t.Fatal(err)
	}
	if status["state"] != "on" || status["ledConnected"] != true {
		t.Fatalf("status = %v", status)
	}
	if events := agentEvents(t, store); len(events) != 1 || events[0].Tier != "observe" || events[0].Outcome != audit.OutcomeOK {
		t.Fatalf("events = %+v", events)
	}
}

func TestPowerPressesGoThroughTheButton(t *testing.T) {
	service, host := newDeviceTestService(t)
	store := withAgentTiers(t, audit.TierObserve, audit.TierPower)

	for _, args := range []string{`{"button":"power"}`, `{"button":"power","press":"long"}`, `{"button":"reset","press":"short"}`} {
		if result := toolResult(t, callTool(t, service, "kvm_power", args, nil)); isErrorResult(result) {
			t.Fatalf("%s: result = %v", args, result)
		}
	}
	want := []string{
		"power " + hostpower.ShortPress.String(),
		"power " + hostpower.LongPress.String(),
		"reset " + hostpower.ShortPress.String(),
	}
	if !reflect.DeepEqual(host.pressed, want) {
		t.Fatalf("pressed = %v, want %v", host.pressed, want)
	}

	events := agentEvents(t, store)
	if len(events) != 3 {
		t.Fatalf("events = %+v", events)
	}
	var long *audit.Event
	for i := range events {
		if strings.HasPrefix(events[i].Detail, "long") {
			long = &events[i]
		}
	}
	if long == nil || long.Target != "power" || long.Tier != "power" || long.Outcome != audit.OutcomeOK || long.Detail != "long hold=5000ms" {
		t.Fatalf("events = %+v, want the long press recorded", events)
	}
}

func TestPowerArgumentErrorsAreToolErrors(t *testing.T) {
	service, host := newDeviceTestService(t)
	withAgentTiers(t, audit.TierPower)

	for _, args := range []string{`{"button":"eject"}`, `{"button":"power","press":"forever"}`} {
		if result := toolResult(t, callTool(t, service, "kvm_power", args, nil)); !isErrorResult(result) {
			t.Fatalf("%s: result = %v, want an error", args, result)
		}
	}
	if len(host.pressed) != 0 {
		t.Fatalf("pressed %v", host.pressed)
	}
}

func TestConcurrentPowerPressIsRefused(t *testing.T) {
	service, host := newDeviceTestService(t)
	withAgentTiers(t, audit.TierPower)
	host.hold = make(chan struct{})
	host.holding = make(chan struct{}, 1)

	first := make(chan map[string]any, 1)
	go func() {
		first <- toolResult(t, callTool(t, service, "kvm_power", `{"button":"power","press":"long"}`, nil))
	}()
	select {
	case <-host.holding:
	case <-time.After(5 * time.Second):
		t.Fatal("the first press never reached the button")
	}

	second := toolResult(t, callTool(t, service, "kvm_power", `{"button":"reset"}`, nil))
	if !isErrorResult(second) || !strings.Contains(toolText(t, second), "another button press is still in progress") {
		t.Fatalf("second press = %v, want refused", second)
	}
	close(host.hold)
	if result := <-first; isErrorResult(result) {
		t.Fatalf("first press = %v", result)
	}
	if len(host.pressed) != 1 {
		t.Fatalf("pressed = %v, want only the first", host.pressed)
	}
}

func TestPowerNeedsPicoClawControl(t *testing.T) {
	service, host := newDeviceTestService(t)
	withAgentTiers(t, audit.TierPower, audit.TierMedia)
	path := filepath.Join(t.TempDir(), "mode")
	if err := os.WriteFile(path, []byte(string(controlmode.ModeMCP)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	service.control = controlmode.NewManager(path, controlmode.ModePicoclaw)

	for _, call := range []struct{ name, args string }{
		{"kvm_power", `{"button":"power"}`},
		{"kvm_media", `{"action":"eject","drive":"cdrom"}`},
	} {
		envelope := callTool(t, service, call.name, call.args, nil)
		if envelope.Error == nil || envelope.Error.Code != -32010 {
			t.Fatalf("%s: envelope = %+v, want the control conflict", call.name, envelope)
		}
	}
	if len(host.pressed) != 0 || len(host.ejected) != 0 {
		t.Fatalf("pressed %v, ejected %v without control", host.pressed, host.ejected)
	}
}

func TestMediaListShowsDrivesAndImages(t *testing.T) {
	service, _ := newDeviceTestService(t)
	withAgentTiers(t, audit.TierObserve)

	result := toolResult(t, callTool(t, service, "kvm_media_list", `{}`, nil))
	text := toolText(t, result)
	if isErrorResult(result) || !strings.Contains(text, `"/data/rescue.img"`) || !strings.Contains(text, `"cdrom"`) {
		t.Fatalf("result = %v", result)
	}
}

func TestMediaInsertIsReadOnlyByDefaultAndEjectEmpties(t *testing.T) {
	service, host := newDeviceTestService(t)
	store := withAgentTiers(t, audit.TierMedia)

	result := toolResult(t, callTool(t, service, "kvm_media", `{"action":"insert","drive":"disk","file":"/data/rescue.img"}`, nil))
	if isErrorResult(result) || !strings.Contains(toolText(t, result), `"drives"`) {
		t.Fatalf("insert = %v, want the drives after it", result)
	}
	result = toolResult(t, callTool(t, service, "kvm_media", `{"action":"insert","drive":"disk","file":"/data/rescue.img","readOnly":false}`, nil))
	if isErrorResult(result) {
		t.Fatalf("writable insert = %v", result)
	}
	result = toolResult(t, callTool(t, service, "kvm_media", `{"action":"insert","drive":"cdrom","file":"/data/win11.iso","readOnly":false}`, nil))
	if !isErrorResult(result) {
		t.Fatalf("writable cdrom = %v, want refused", result)
	}
	if result := toolResult(t, callTool(t, service, "kvm_media", `{"action":"eject","drive":"cdrom"}`, nil)); isErrorResult(result) {
		t.Fatalf("eject = %v", result)
	}

	if want := []string{"disk /data/rescue.img ro=1", "disk /data/rescue.img ro=0"}; !reflect.DeepEqual(host.inserted, want) {
		t.Fatalf("inserted = %v, want %v", host.inserted, want)
	}
	if !reflect.DeepEqual(host.ejected, []string{"cdrom"}) {
		t.Fatalf("ejected = %v", host.ejected)
	}
	events := agentEvents(t, store)
	if len(events) != 4 {
		t.Fatalf("events = %+v", events)
	}
	found := false
	for _, e := range events {
		if e.Detail == "insert disk ro" && e.Target == "/data/rescue.img" && e.Outcome == audit.OutcomeOK {
			found = true
		}
	}
	if !found {
		t.Fatalf("events = %+v, want the read-only insert recorded", events)
	}
}

func TestMediaURLInsertKeepsCredentialsOutOfTheAuditLog(t *testing.T) {
	service, host := newDeviceTestService(t)
	store := withAgentTiers(t, audit.TierMedia)
	url := "https://alice:secret@images.example/win11.iso?token=hunter2"

	result := toolResult(t, callTool(t, service, "kvm_media", `{"action":"insert","drive":"cdrom","url":"`+url+`"}`, nil))
	if isErrorResult(result) {
		t.Fatalf("result = %v", result)
	}
	if !reflect.DeepEqual(host.mounted, []string{"cdrom " + url}) {
		t.Fatalf("mounted = %v", host.mounted)
	}

	// A failure whose message quotes the URL must not carry it into the log.
	host.mountErr = errors.New(`fetch "` + url + `": connection refused`)
	result = toolResult(t, callTool(t, service, "kvm_media", `{"action":"insert","drive":"cdrom","url":"`+url+`"}`, nil))
	if !isErrorResult(result) {
		t.Fatalf("result = %v, want the mount error", result)
	}

	events := agentEvents(t, store)
	if len(events) != 2 {
		t.Fatalf("events = %+v", events)
	}
	for _, e := range events {
		line := e.Target + " " + e.Detail
		if strings.Contains(line, "secret") || strings.Contains(line, "hunter2") || strings.Contains(line, "alice") {
			t.Fatalf("event %+v carries the URL's credentials", e)
		}
		if e.Target != "images.example" || !strings.Contains(e.Detail, "url") {
			t.Fatalf("event %+v, want the host and a url insert", e)
		}
	}
}

// A press and a media change act for the chat session that holds the AI input
// lock, as kvm_actions does. With no session, as after the user closed the
// chat while the agent's turn ran on, they are refused.
func TestPowerAndMediaNeedAChatSession(t *testing.T) {
	service, host := newDeviceTestService(t)
	store := withAgentTiers(t, audit.TierPower, audit.TierMedia)
	service.lock.Release(deviceTestSession)

	for _, call := range []struct{ name, args string }{
		{"kvm_power", `{"button":"power"}`},
		{"kvm_media", `{"action":"eject","drive":"cdrom"}`},
	} {
		result := toolResult(t, callTool(t, service, call.name, call.args, nil))
		if !isErrorResult(result) || !strings.Contains(toolText(t, result), "no chat session holds device control") {
			t.Fatalf("%s: result = %v, want refused", call.name, result)
		}
	}
	if len(host.pressed) != 0 || len(host.ejected) != 0 {
		t.Fatalf("pressed %v, ejected %v without a session", host.pressed, host.ejected)
	}
	for _, e := range agentEvents(t, store) {
		if e.Outcome != audit.OutcomeFail || e.Session != "" {
			t.Fatalf("event = %+v, want a failure with no session", e)
		}
	}
}

// Another session's lock refuses the call that names its own session.
func TestPowerForAnotherSessionIsRefused(t *testing.T) {
	service, host := newDeviceTestService(t)
	withAgentTiers(t, audit.TierPower)

	result := toolResult(t, callTool(t, service, "kvm_power", `{"button":"power"}`, map[string]string{agentSessionHeader: "chat-2"}))
	if !isErrorResult(result) || len(host.pressed) != 0 {
		t.Fatalf("result = %v, pressed %v, want refused", result, host.pressed)
	}
}

// The audit record names the agent and the session it acted for.
func TestDeviceCallsRecordTheAgentAndItsSession(t *testing.T) {
	service, _ := newDeviceTestService(t)
	store := withAgentTiers(t, audit.TierObserve, audit.TierPower)

	toolResult(t, callTool(t, service, "kvm_power", `{"button":"power"}`, nil))
	toolResult(t, callTool(t, service, "kvm_power_status", `{}`, map[string]string{agentSessionHeader: "chat-1"}))

	events := agentEvents(t, store)
	if len(events) != 2 {
		t.Fatalf("events = %+v", events)
	}
	for _, e := range events {
		if e.Actor != "picoclaw" || e.Session != deviceTestSession || e.Outcome != audit.OutcomeOK {
			t.Fatalf("event = %+v, want picoclaw acting for %s", e, deviceTestSession)
		}
	}
}
