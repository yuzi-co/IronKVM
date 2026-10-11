package mcpservice

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/service/controlmode"
	"NanoKVM-Server/service/inputcontrol"
)

// isolateAudit keeps the events and the tier settings of a test off the
// board's own files, and returns the store the events go to.
func isolateAudit(t *testing.T) *audit.Store {
	t.Helper()
	store := audit.NewStore(filepath.Join(t.TempDir(), "audit"))
	restore := audit.UseStore(store)
	oldSettings, oldTiers := audit.SettingsFile, audit.TiersFile
	audit.SettingsFile = filepath.Join(t.TempDir(), "audit.json")
	audit.TiersFile = filepath.Join(t.TempDir(), "agent-tiers.json")
	audit.ResetForTest()
	t.Cleanup(func() {
		restore()
		audit.SettingsFile, audit.TiersFile = oldSettings, oldTiers
		audit.ResetForTest()
	})
	return store
}

type mcpSession struct {
	t       *testing.T
	handler http.Handler
	id      string
	next    int
}

func newSession(t *testing.T, devices Devices) *mcpSession {
	t.Helper()
	control := controlmode.NewManager(filepath.Join(t.TempDir(), "mode"), controlmode.ModeMCP)
	handler := newMCPHandler(control, &inputcontrol.Coordinator{}, newRemoteWithHID(&recordingHID{}), fakeSnapshotter{snapshot: Snapshot{
		OK: true, Width: 800, Height: 600, JPEG: []byte{0xff, 0xd8, 0xff},
	}}, devices)
	s := &mcpSession{t: t, handler: handler, next: 10}
	rsp := s.post(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`)
	s.id = rsp.Header().Get("Mcp-Session-Id")
	return s
}

func (s *mcpSession) post(body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/mcp", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	// What APIKeyMiddleware does, and what a hostile client tries first.
	req.Header.Set(audit.ClientHeader, "10.9.9.9")
	if s.id != "" {
		req.Header.Set("Mcp-Session-Id", s.id)
		req.Header.Set("Mcp-Protocol-Version", "2025-03-26")
	}
	s.handler.ServeHTTP(rec, req)
	return rec
}

func (s *mcpSession) tools() string {
	s.next++
	return s.post(fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"tools/list","params":{}}`, s.next)).Body.String()
}

func (s *mcpSession) call(name, arguments string) string {
	s.next++
	return s.post(fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"tools/call","params":{"name":%q,"arguments":%s}}`, s.next, name, arguments)).Body.String()
}

func events(t *testing.T, store *audit.Store) []audit.Event {
	t.Helper()
	page, err := store.Query(audit.Query{})
	if err != nil {
		t.Fatal(err)
	}
	return page.Events
}

func testDevices(pressed *[]string) Devices {
	return Devices{
		PressButton: func(kind string, d time.Duration) error {
			*pressed = append(*pressed, fmt.Sprintf("%s %s", kind, d))
			return nil
		},
		ListMedia: func() ([]MediaDrive, []string, error) {
			return []MediaDrive{{ID: "cdrom"}}, []string{"/data/a.iso"}, nil
		},
		InsertMedia: func(drive, file string, ro bool) error { return nil },
		EjectMedia:  func(drive string) error { return nil },
	}
}

func TestDefaultTiersKeepWhatMCPCouldAlwaysDo(t *testing.T) {
	isolateAudit(t)
	var pressed []string
	list := newSession(t, testDevices(&pressed)).tools()

	for _, name := range []string{"cube_screenshot", "cube_type_text", "cube_press_keys", "cube_move_mouse", "cube_click_mouse", "cube_scroll_mouse"} {
		if !strings.Contains(list, `"name":"`+name+`"`) {
			t.Errorf("default tiers dropped %s", name)
		}
	}
	for _, name := range []string{"cube_power", "cube_media"} {
		if strings.Contains(list, `"name":"`+name+`"`) {
			t.Errorf("%s is offered although its tier is off by default", name)
		}
	}
}

func TestCallToAnOffTierIsRefused(t *testing.T) {
	isolateAudit(t)
	var pressed []string
	s := newSession(t, testDevices(&pressed))

	out := s.call("cube_power", `{"button":"reset"}`)
	if !strings.Contains(out, "tier is off") {
		t.Fatalf("call to a hidden tool was not refused with the tier: %s", out)
	}
	if len(pressed) != 0 {
		t.Fatalf("the button was pressed with the power tier off: %v", pressed)
	}
}

func TestTurningATierOnOffersAndRunsTheTool(t *testing.T) {
	store := isolateAudit(t)
	var pressed []string
	s := newSession(t, testDevices(&pressed))

	if _, err := audit.SaveTiers(audit.Tiers{
		MCP:   []audit.Tier{audit.TierObserve, audit.TierInput, audit.TierPower},
		Agent: []audit.Tier{audit.TierObserve, audit.TierInput},
	}); err != nil {
		t.Fatal(err)
	}

	if list := s.tools(); !strings.Contains(list, `"name":"cube_power"`) || strings.Contains(list, `"name":"cube_media"`) {
		t.Fatalf("tools/list does not follow the tiers: %s", list)
	}
	if out := s.call("cube_power", `{"button":"reset","holdMs":300}`); strings.Contains(out, `"isError":true`) {
		t.Fatalf("power call failed: %s", out)
	}
	if len(pressed) != 1 || pressed[0] != "reset 300ms" {
		t.Fatalf("pressed = %v", pressed)
	}

	got := events(t, store)
	if len(got) != 1 {
		t.Fatalf("got %d events, want 1: %+v", len(got), got)
	}
	e := got[0]
	if e.Actor != "mcp" || e.Via != "mcp" || e.Category != audit.CategoryAgent || e.Action != "cube_power" ||
		e.Tier != "power" || e.Target != "reset" || e.Outcome != audit.OutcomeOK || e.IP != "10.9.9.9" {
		t.Fatalf("event wrong: %+v", e)
	}
}

func TestDeniedCallIsAudited(t *testing.T) {
	store := isolateAudit(t)
	var pressed []string
	s := newSession(t, testDevices(&pressed))
	s.call("cube_media", `{"action":"list"}`)

	got := events(t, store)
	if len(got) != 1 || got[0].Outcome != audit.OutcomeDenied || got[0].Tier != "media" || got[0].Action != "cube_media" {
		t.Fatalf("events = %+v", got)
	}
}

func TestInputCallsAreAuditedWithoutWhatWasTyped(t *testing.T) {
	store := isolateAudit(t)
	s := newSession(t, Devices{})

	s.call("cube_type_text", `{"text":"hunter2-secret","delayMs":1}`)
	s.call("cube_screenshot", `{}`)

	got := events(t, store)
	if len(got) != 2 {
		t.Fatalf("got %d events, want 2: %+v", len(got), got)
	}
	for _, e := range got {
		if strings.Contains(e.Detail+e.Target+e.Action, "hunter2") {
			t.Fatalf("typed text reached the log: %+v", e)
		}
	}
	// Newest first: the screenshot, then the typing.
	if got[0].Action != "cube_screenshot" || got[0].Tier != "observe" ||
		got[1].Action != "cube_type_text" || got[1].Tier != "input" || got[1].Detail != "chars=14" {
		t.Fatalf("events wrong: %+v", got)
	}
}

func TestEveryToolHasATier(t *testing.T) {
	for _, name := range []string{
		screenshotToolName, typeTextToolName, pressKeysToolName, moveMouseToolName,
		clickMouseToolName, scrollMouseToolName, powerToolName, mediaToolName,
	} {
		if _, ok := toolTiers[name]; !ok {
			t.Errorf("%s has no tier", name)
		}
	}
	if toolTier("cube_unheard_of") != audit.TierAdmin {
		t.Error("a tool without a decision should need the admin tier")
	}
}
