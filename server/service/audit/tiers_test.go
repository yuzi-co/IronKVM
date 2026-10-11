package audit

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func useTierFile(t *testing.T) string {
	t.Helper()
	old := TiersFile
	TiersFile = filepath.Join(t.TempDir(), "agent-tiers.json")
	ResetForTest()
	t.Cleanup(func() {
		TiersFile = old
		ResetForTest()
	})
	return TiersFile
}

func TestDefaultTiersAreWhatMCPAndTheAgentAlwaysHad(t *testing.T) {
	useTierFile(t)
	want := []Tier{TierObserve, TierInput}
	got := CurrentTiers()
	if !reflect.DeepEqual(got.MCP, want) || !reflect.DeepEqual(got.Agent, want) {
		t.Fatalf("defaults = %+v, want observe and input for both", got)
	}
	for _, subject := range []string{SubjectMCP, SubjectAgent} {
		if !Allowed(subject, TierObserve) || !Allowed(subject, TierInput) {
			t.Errorf("%s lost observe or input by default", subject)
		}
		for _, tier := range []Tier{TierPower, TierMedia, TierAdmin} {
			if Allowed(subject, tier) {
				t.Errorf("%s has %s by default", subject, tier)
			}
		}
	}
}

func TestSaveTiersPersistsNormalizesAndRejects(t *testing.T) {
	path := useTierFile(t)

	saved, err := SaveTiers(Tiers{
		MCP:   []Tier{TierPower, TierObserve, TierPower},
		Agent: []Tier{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(saved.MCP, []Tier{TierObserve, TierPower}) || len(saved.Agent) != 0 {
		t.Fatalf("saved = %+v", saved)
	}
	if Allowed(SubjectAgent, TierObserve) {
		t.Fatal("an empty agent list should allow nothing")
	}

	// A restart reads the same.
	ResetForTest()
	if got := CurrentTiers(); !reflect.DeepEqual(got.MCP, saved.MCP) || len(got.Agent) != 0 {
		t.Fatalf("after reload = %+v", got)
	}

	if _, err := SaveTiers(Tiers{MCP: []Tier{"root"}}); err == nil {
		t.Fatal("an unknown tier was accepted")
	}
	if got := CurrentTiers(); !reflect.DeepEqual(got.MCP, saved.MCP) {
		t.Fatalf("a refused save changed the tiers: %+v", got)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func TestAnUnknownTierInTheFileIsIgnored(t *testing.T) {
	path := useTierFile(t)
	// What a newer slot might have written.
	if err := os.WriteFile(path, []byte(`{"mcp":["observe","teleport"],"agent":["input"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got := CurrentTiers()
	if !reflect.DeepEqual(got.MCP, []Tier{TierObserve}) || !reflect.DeepEqual(got.Agent, []Tier{TierInput}) {
		t.Fatalf("tiers = %+v", got)
	}
}

func TestAMissingKeyKeepsItsDefault(t *testing.T) {
	path := useTierFile(t)
	if err := os.WriteFile(path, []byte(`{"mcp":["observe"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got := CurrentTiers()
	if !reflect.DeepEqual(got.Agent, []Tier{TierObserve, TierInput}) {
		t.Fatalf("agent tiers = %v, want the defaults", got.Agent)
	}
}

func TestTierRoutes(t *testing.T) {
	useTierFile(t)
	gin.SetMode(gin.TestMode)
	service := NewService()
	r := gin.New()
	r.GET("/api/agent/tiers", service.GetTiers)
	r.POST("/api/agent/tiers", service.SetTiers)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/api/agent/tiers", nil))
	var rsp envelope
	if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil || rsp.Code != 0 {
		t.Fatalf("get: %s", w.Body.String())
	}
	var data struct {
		MCP       []string `json:"mcp"`
		Agent     []string `json:"agent"`
		Available []string `json:"available"`
	}
	if err := json.Unmarshal(rsp.Data, &data); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(data.Available, []string{"observe", "input", "power", "media", "admin"}) ||
		!reflect.DeepEqual(data.MCP, []string{"observe", "input"}) {
		t.Fatalf("get = %+v", data)
	}

	post := func(body string) envelope {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/agent/tiers", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		var rsp envelope
		_ = json.Unmarshal(w.Body.Bytes(), &rsp)
		return rsp
	}
	if rsp := post(`{"mcp":["observe","power"],"agent":["observe"]}`); rsp.Code != 0 {
		t.Fatalf("set: %+v", rsp)
	}
	if !Allowed(SubjectMCP, TierPower) || Allowed(SubjectAgent, TierInput) {
		t.Fatal("the tiers were not applied")
	}
	for _, bad := range []string{`{}`, `{"mcp":["observe"]}`, `{"mcp":["x"],"agent":[]}`, `nope`} {
		if rsp := post(bad); rsp.Code == 0 {
			t.Errorf("%s was accepted", bad)
		}
	}
}
