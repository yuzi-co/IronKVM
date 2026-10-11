package picoclaw

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"NanoKVM-Server/service/audit"

	"github.com/gin-gonic/gin"
)

// TestMain keeps the audit events and the tier settings of every test in this
// package off the board's own files.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "picoclaw-audit")
	if err != nil {
		panic(err)
	}
	audit.UseStore(audit.NewStore(filepath.Join(dir, "log")))
	audit.SettingsFile = filepath.Join(dir, "audit.json")
	audit.TiersFile = filepath.Join(dir, "agent-tiers.json")
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// withAgentTiers sets the agent's tiers for a test and puts the defaults back.
func withAgentTiers(t *testing.T, tiers ...audit.Tier) *audit.Store {
	t.Helper()
	store := audit.NewStore(filepath.Join(t.TempDir(), "log"))
	restore := audit.UseStore(store)
	if _, err := audit.SaveTiers(audit.Tiers{MCP: audit.DefaultTiers().MCP, Agent: tiers}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		restore()
		_ = os.Remove(audit.TiersFile)
		audit.ResetForTest()
	})
	return store
}

func agentEvents(t *testing.T, store *audit.Store) []audit.Event {
	t.Helper()
	page, err := store.Query(audit.Query{})
	if err != nil {
		t.Fatal(err)
	}
	return page.Events
}

func TestAgentActionsAreRefusedWithTheInputTierOff(t *testing.T) {
	service, _, hid := newMCPActionTestService(t)
	store := withAgentTiers(t, audit.TierObserve)

	body := performMCPRequest(service, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"kvm_actions","arguments":{"actions":[{"action":"click","x":0.5,"y":0.5}]}}}`).Body.String()
	if !strings.Contains(body, `"isError":true`) || !strings.Contains(body, "input tier is off") {
		t.Fatalf("not refused with the tier: %s", body)
	}
	if len(hid.mouse) != 0 {
		t.Fatal("the click reached the HID with the tier off")
	}

	events := agentEvents(t, store)
	if len(events) != 1 || events[0].Outcome != audit.OutcomeDenied || events[0].Tier != "input" ||
		events[0].Actor != "picoclaw" || events[0].Action != "kvm_actions" {
		t.Fatalf("events = %+v", events)
	}
}

func TestAgentToolsListFollowsTheTiers(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)
	withAgentTiers(t, audit.TierObserve)

	body := performMCPRequest(service, jsonRPCBody("tools/list")).Body.String()
	if !strings.Contains(body, "kvm_screenshot") || strings.Contains(body, "kvm_actions") {
		t.Fatalf("tools/list = %s", body)
	}
}

func TestAgentCallsAreAuditedWithoutWhatWasTyped(t *testing.T) {
	service, _, _ := newMCPActionTestService(t)
	store := withAgentTiers(t, audit.TierObserve, audit.TierInput)

	result := callKVMActions(t, service, `{"actions":[{"action":"move","x":0.1,"y":0.1},{"action":"type","text":"s3cret-pass"}]}`)
	if result.IsError {
		t.Fatalf("result = %+v", result)
	}

	events := agentEvents(t, store)
	if len(events) != 1 {
		t.Fatalf("events = %+v", events)
	}
	e := events[0]
	if e.Outcome != audit.OutcomeOK || e.Tier != "input" || e.Detail != "actions=2 move,type" || strings.Contains(e.Detail, "s3cret") {
		t.Fatalf("event wrong: %+v", e)
	}
}

func TestAgentGateRefusesAndRecordsALoopbackRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := withAgentTiers(t, audit.TierObserve)

	ran := false
	r := gin.New()
	r.POST("/actions", audit.AgentGate("actions", audit.TierInput), func(c *gin.Context) { ran = true })
	r.GET("/screenshot", audit.AgentGate("screenshot", audit.TierObserve), func(c *gin.Context) { c.Status(http.StatusOK) })

	denied := httptest.NewRecorder()
	r.ServeHTTP(denied, httptest.NewRequest("POST", "/actions", nil))
	allowed := httptest.NewRecorder()
	r.ServeHTTP(allowed, httptest.NewRequest("GET", "/screenshot", nil))

	if denied.Code != http.StatusForbidden || ran {
		t.Fatalf("denied route: status %d, ran %t", denied.Code, ran)
	}
	if allowed.Code != http.StatusOK {
		t.Fatalf("allowed route: status %d", allowed.Code)
	}
	events := agentEvents(t, store)
	if len(events) != 2 || events[0].Outcome != audit.OutcomeOK || events[0].Tier != "observe" ||
		events[1].Outcome != audit.OutcomeDenied || events[1].Tier != "input" {
		t.Fatalf("events = %+v", events)
	}
}
