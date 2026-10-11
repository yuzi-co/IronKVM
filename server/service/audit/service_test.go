package audit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type envelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func serviceEngine(t *testing.T) (*gin.Engine, *Store) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	oldStd := std
	std = newTestStore(t)
	oldSettings := SettingsFile
	SettingsFile = filepath.Join(t.TempDir(), "audit.json")
	ResetForTest()
	t.Cleanup(func() {
		std, SettingsFile = oldStd, oldSettings
		ResetForTest()
	})

	service := NewService()
	r := gin.New()
	r.GET("/api/audit", service.List)
	r.GET("/api/audit/export", service.Export)
	r.GET("/api/audit/settings", service.GetSettings)
	r.POST("/api/audit/settings", service.SetSettings)
	return r, std
}

func get(r *gin.Engine, target string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", target, nil))
	return w
}

func TestListPagesAndFiltersOverHTTP(t *testing.T) {
	r, s := serviceEngine(t)
	now := time.Now()
	for i := 0; i < 6; i++ {
		category := CategoryAuth
		if i >= 4 {
			category = CategoryPower
		}
		s.Record(Event{Time: now.Add(time.Duration(i) * time.Second), Actor: "alice", Category: category, Action: "x"})
	}

	var rsp envelope
	w := get(r, "/api/audit?category=power&limit=1")
	if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil || rsp.Code != 0 {
		t.Fatalf("bad answer: %s", w.Body.String())
	}
	var page Page
	if err := json.Unmarshal(rsp.Data, &page); err != nil {
		t.Fatal(err)
	}
	if page.Total != 2 || len(page.Events) != 1 {
		t.Fatalf("page = %+v", page)
	}

	for _, bad := range []string{"limit=abc", "offset=-1", "since=yesterday"} {
		var rsp envelope
		_ = json.Unmarshal(get(r, "/api/audit?"+bad).Body.Bytes(), &rsp)
		if rsp.Code == 0 {
			t.Errorf("%s was accepted", bad)
		}
	}
}

func TestEmptyLogListsAsAnEmptyArray(t *testing.T) {
	r, _ := serviceEngine(t)
	body := get(r, "/api/audit").Body.String()
	if !strings.Contains(body, `"events":[]`) {
		t.Fatalf("empty page should carry an empty array, got %s", body)
	}
}

func TestExportIsJSONLinesOldestFirst(t *testing.T) {
	r, s := serviceEngine(t)
	now := time.Now()
	s.Record(Event{Time: now, Actor: "a", Category: CategoryAuth, Action: "first"})
	s.Record(Event{Time: now.Add(time.Second), Actor: "a", Category: CategoryAuth, Action: "second"})

	w := get(r, "/api/audit/export")
	if w.Code != http.StatusOK || !strings.Contains(w.Header().Get("Content-Disposition"), "attachment") {
		t.Fatalf("not a download: %d %v", w.Code, w.Header())
	}
	lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
	if len(lines) != 2 || !strings.Contains(lines[0], "first") || !strings.Contains(lines[1], "second") {
		t.Fatalf("export = %q", w.Body.String())
	}
}

func TestSettingsPersistAndDriveTheForward(t *testing.T) {
	r, s := serviceEngine(t)
	var forwarded []string
	s.SetSink(func(e Event) {
		if CurrentSettings().Syslog {
			forwarded = append(forwarded, Syslog(e))
		}
	})

	s.Record(Event{Category: CategoryAuth, Action: "before"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/audit/settings", strings.NewReader(`{"syslog":true}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if !CurrentSettings().Syslog {
		t.Fatalf("setting not applied: %s", w.Body.String())
	}

	s.Record(Event{Category: CategoryAuth, Action: "after"})
	if len(forwarded) != 1 || !strings.Contains(forwarded[0], "action=after") {
		t.Fatalf("forwarded = %q", forwarded)
	}

	// A fresh read of the file sees it, as a restart would.
	ResetForTest()
	if !CurrentSettings().Syslog {
		t.Fatal("setting lost after a reload")
	}
}

func TestSettingsRejectAMissingField(t *testing.T) {
	r, _ := serviceEngine(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/audit/settings", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var rsp envelope
	_ = json.Unmarshal(w.Body.Bytes(), &rsp)
	if rsp.Code == 0 {
		t.Fatal("an empty body changed the settings")
	}
}
