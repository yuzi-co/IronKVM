package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	return NewStore(filepath.Join(t.TempDir(), "audit"))
}

func TestRecordThenQueryNewestFirst(t *testing.T) {
	s := newTestStore(t)
	base := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		s.Record(Event{Time: base.Add(time.Duration(i) * time.Second), Actor: "alice", Category: CategoryPower, Action: fmt.Sprintf("a%d", i)})
	}

	page, err := s.Query(Query{})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 5 || len(page.Events) != 5 {
		t.Fatalf("got total %d, %d events; want 5", page.Total, len(page.Events))
	}
	if page.Events[0].Action != "a4" || page.Events[4].Action != "a0" {
		t.Fatalf("not newest first: %v", page.Events)
	}
}

func TestQueryPagesAndFilters(t *testing.T) {
	s := newTestStore(t)
	now := time.Now().Truncate(time.Millisecond)
	for i := 0; i < 10; i++ {
		category, outcome := CategoryAuth, OutcomeOK
		if i%2 == 1 {
			category, outcome = CategoryMedia, OutcomeFail
		}
		s.Record(Event{Time: now.Add(time.Duration(i) * time.Second), Actor: "bob", IP: "10.0.0.7", Category: category, Action: "x", Target: fmt.Sprintf("img%d.iso", i), Outcome: outcome})
	}

	page, _ := s.Query(Query{Limit: 3, Offset: 2})
	if page.Total != 10 || len(page.Events) != 3 || page.Events[0].Target != "img7.iso" {
		t.Fatalf("paging wrong: %+v", page)
	}

	page, _ = s.Query(Query{Category: CategoryMedia})
	if page.Total != 5 {
		t.Fatalf("category filter matched %d, want 5", page.Total)
	}
	page, _ = s.Query(Query{Outcome: OutcomeOK, Text: "IMG4"})
	if page.Total != 1 || page.Events[0].Target != "img4.iso" {
		t.Fatalf("text filter wrong: %+v", page)
	}
	page, _ = s.Query(Query{Actor: "BO", IP: "10.0.0"})
	if page.Total != 10 {
		t.Fatalf("actor/ip filter matched %d, want 10", page.Total)
	}
	page, _ = s.Query(Query{Since: now.Add(8 * time.Second)})
	if page.Total != 2 {
		t.Fatalf("since filter matched %d, want 2", page.Total)
	}
	page, _ = s.Query(Query{Offset: 99})
	if len(page.Events) != 0 || page.Total != 10 {
		t.Fatalf("offset past the end: %+v", page)
	}
}

func TestFieldsCannotForgeALine(t *testing.T) {
	s := newTestStore(t)
	s.Record(Event{Actor: "mallory\n{\"actor\":\"admin\"}", Category: CategoryAuth, Action: "login", Outcome: OutcomeFail, Detail: strings.Repeat("é", 1000)})
	s.Flush()

	data, err := os.ReadFile(filepath.Join(s.dir, activeName))
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(string(data), "\n"); n != 1 {
		t.Fatalf("one event wrote %d lines", n)
	}
	page, _ := s.Query(Query{})
	if len(page.Events) != 1 {
		t.Fatalf("got %d events", len(page.Events))
	}
	detail := page.Events[0].Detail
	if len(detail) > maxDetail {
		t.Fatalf("detail not bounded: %d", len(detail))
	}
	if !strings.HasSuffix(detail, "é") {
		t.Fatalf("detail cut inside a rune: %q", detail)
	}
}

func TestRepeatedFailuresAreFolded(t *testing.T) {
	s := newTestStore(t)
	now := time.Now()
	for i := 0; i < 50; i++ {
		s.Record(Event{Time: now.Add(time.Duration(i) * time.Second / 10), Actor: "root", IP: "10.0.0.9", Category: CategoryAuth, Action: "login", Outcome: OutcomeFail, Detail: "invalid username or password"})
	}
	s.Record(Event{Time: now.Add(6 * time.Second), Actor: "root", IP: "10.0.0.9", Category: CategoryAuth, Action: "login", Outcome: OutcomeOK})

	page, _ := s.Query(Query{})
	if page.Total != 2 {
		t.Fatalf("got %d events, want the failure folded into one plus the success", page.Total)
	}
	if page.Events[1].Repeat != 50 {
		t.Fatalf("repeat = %d, want 50", page.Events[1].Repeat)
	}
}

func TestSuccessesAreNeverFolded(t *testing.T) {
	s := newTestStore(t)
	for i := 0; i < 3; i++ {
		s.Record(Event{Actor: "agent", Category: CategoryAgent, Action: "tool", Target: "click"})
	}
	page, _ := s.Query(Query{})
	if page.Total != 3 {
		t.Fatalf("got %d events, want 3", page.Total)
	}
}

func TestRotationKeepsTheLogUnderItsCap(t *testing.T) {
	s := newTestStore(t)
	line := strings.Repeat("x", 200)
	count := (segmentBytes * segments * 3) / 260 // several times the cap
	for i := 0; i < count; i++ {
		s.Record(Event{Actor: "u", Category: CategoryAuth, Action: "a", Detail: line, Target: fmt.Sprint(i)})
	}
	s.Flush()

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		t.Fatal(err)
	}
	var total int64
	for _, entry := range entries {
		info, _ := entry.Info()
		total += info.Size()
	}
	if len(entries) > segments {
		t.Fatalf("%d files kept, want at most %d", len(entries), segments)
	}
	if total > segmentBytes*segments {
		t.Fatalf("log is %d bytes, cap is %d", total, segmentBytes*segments)
	}

	// The newest event survived and is first.
	page, _ := s.Query(Query{Limit: 1})
	if len(page.Events) != 1 || page.Events[0].Target != fmt.Sprint(count-1) {
		t.Fatalf("newest event wrong: %+v", page.Events)
	}
}

func TestEventsWaitForTheTimer(t *testing.T) {
	s := newTestStore(t)
	s.Record(Event{Category: CategoryAuth, Action: "login"})
	if _, err := os.Stat(filepath.Join(s.dir, activeName)); err == nil {
		t.Fatal("a single event hit the disk at once; writes should be batched")
	}
	s.Flush()
	if _, err := os.Stat(filepath.Join(s.dir, activeName)); err != nil {
		t.Fatalf("flush wrote nothing: %v", err)
	}
}

func TestBatchFlushesAtThreshold(t *testing.T) {
	s := newTestStore(t)
	for i := 0; i < flushBatch; i++ {
		s.Record(Event{Category: CategoryAuth, Action: "a", Target: fmt.Sprint(i)})
	}
	if _, err := os.Stat(filepath.Join(s.dir, activeName)); err != nil {
		t.Fatalf("a full batch was not written: %v", err)
	}
}

func TestReadSkipsDamagedLines(t *testing.T) {
	s := newTestStore(t)
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		t.Fatal(err)
	}
	body := "not json\n{\"time\":\"2026-10-01T00:00:00Z\",\"category\":\"auth\",\"action\":\"login\",\"outcome\":\"ok\"}\n{broken\n"
	if err := os.WriteFile(filepath.Join(s.dir, activeName), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	page, err := s.Query(Query{})
	if err != nil || page.Total != 1 {
		t.Fatalf("got %+v, %v; want the one good line", page, err)
	}
}

func TestSinkSeesEveryEvent(t *testing.T) {
	s := newTestStore(t)
	var got []string
	s.SetSink(func(e Event) { got = append(got, Syslog(e)) })
	s.Record(Event{Actor: "alice smith", Via: "session", IP: "10.0.0.1", Category: CategoryPower, Action: "press", Target: "reset", Outcome: OutcomeOK})
	want := `audit: actor="alice smith" via=session ip=10.0.0.1 category=power action=press target=reset outcome=ok`
	if len(got) != 1 || got[0] != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
