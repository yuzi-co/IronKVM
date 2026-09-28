package hid

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/proto"
)

func mustLayout(t *testing.T, id string) *Layout {
	t.Helper()
	layout, err := GetLayout(id)
	if err != nil {
		t.Fatal(err)
	}
	return layout
}

func planCodes(plan pastePlan) [][]Char {
	codes := make([][]Char, 0, len(plan.steps))
	for _, step := range plan.steps {
		codes = append(codes, step.strokes)
	}
	return codes
}

func planIndexes(plan pastePlan) []int {
	indexes := make([]int, 0, len(plan.steps))
	for _, step := range plan.steps {
		indexes = append(indexes, step.index)
	}
	return indexes
}

func TestPlanPasteLineEndings(t *testing.T) {
	enter := []Char{{0, usageEnter}}
	a := []Char{{0, 4}}

	tests := []struct {
		name    string
		text    string
		want    [][]Char
		indexes []int
	}{
		{"LF", "a\na", [][]Char{a, enter, a}, []int{0, 1, 2}},
		{"CRLF", "a\r\na", [][]Char{a, enter, a}, []int{0, 1, 3}},
		{"CR", "a\ra", [][]Char{a, enter, a}, []int{0, 1, 2}},
		{"blank lines", "\r\n\r\n", [][]Char{enter, enter}, []int{0, 2}},
		{"CR then CRLF", "\r\r\n", [][]Char{enter, enter}, []int{0, 1}},
		{"LF then CR", "\n\r", [][]Char{enter, enter}, []int{0, 1}},
		{"trailing CR", "a\r", [][]Char{a, enter}, []int{0, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := planPaste(tt.text, mustLayout(t, "us"))
			if got := planCodes(plan); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("strokes = %+v, want %+v", got, tt.want)
			}
			if got := planIndexes(plan); !reflect.DeepEqual(got, tt.indexes) {
				t.Errorf("indexes = %v, want %v", got, tt.indexes)
			}
			if plan.untypeableCount != 0 {
				t.Errorf("untypeable = %+v", plan.untypeable)
			}
		})
	}
}

func TestPlanPasteTabs(t *testing.T) {
	for _, id := range LayoutIDs() {
		plan := planPaste("\tx\t", mustLayout(t, id))
		codes := planCodes(plan)
		if id == "ru" {
			// x is untypeable on Russian; the tabs still are.
			if len(codes) != 2 || codes[0][0] != (Char{0, usageTab}) || codes[1][0] != (Char{0, usageTab}) {
				t.Errorf("%s: strokes = %+v", id, codes)
			}
			continue
		}
		if len(codes) != 3 || codes[0][0] != (Char{0, usageTab}) || codes[2][0] != (Char{0, usageTab}) {
			t.Errorf("%s: strokes = %+v", id, codes)
		}
	}
}

func TestPlanPasteDeadKeys(t *testing.T) {
	plan := planPaste("Crème brûlée", mustLayout(t, "fr"))
	if plan.untypeableCount != 0 {
		t.Fatalf("untypeable = %+v", plan.untypeable)
	}
	// è and é have their own keys on French; û takes dead ^ and u.
	if len(plan.steps) != 12 {
		t.Fatalf("steps = %d, want 12", len(plan.steps))
	}
	u := plan.steps[8]
	if u.index != 8 || !reflect.DeepEqual(u.strokes, []Char{{0, 47}, {0, 24}}) {
		t.Errorf("û = %+v", u)
	}
	if plan.keystrokes != 13 {
		t.Errorf("keystrokes = %d, want 13", plan.keystrokes)
	}
}

func TestPlanPasteComposesCombiningMarks(t *testing.T) {
	// e and U+0301 COMBINING ACUTE ACCENT, as a macOS file name holds é.
	plan := planPaste("cé!", mustLayout(t, "de"))
	if plan.untypeableCount != 0 {
		t.Fatalf("untypeable = %+v", plan.untypeable)
	}
	want := [][]Char{{{0, 6}}, {{0, 46}, {0, 8}}, {{modShift, 30}}}
	if got := planCodes(plan); !reflect.DeepEqual(got, want) {
		t.Errorf("strokes = %+v, want %+v", got, want)
	}
	if got := planIndexes(plan); !reflect.DeepEqual(got, []int{0, 1, 3}) {
		t.Errorf("indexes = %v", got)
	}

	// A mark that composes to nothing the layout has is reported with its
	// letter, and US cannot type é at all.
	plan = planPaste("éx", mustLayout(t, "us"))
	if plan.untypeableCount != 1 || plan.untypeable[0] != (proto.PasteUntypeable{Index: 0, Line: 1, Column: 1, Char: "é"}) {
		t.Errorf("untypeable = %+v", plan.untypeable)
	}
	if got := planIndexes(plan); !reflect.DeepEqual(got, []int{2}) {
		t.Errorf("indexes = %v", got)
	}
}

func TestPlanPasteReportsUntypeable(t *testing.T) {
	text := "ab€\r\n\tжx\rДа 😀"
	plan := planPaste(text, mustLayout(t, "us"))

	want := []proto.PasteUntypeable{
		{Index: 2, Line: 1, Column: 3, Char: "€"},
		{Index: 6, Line: 2, Column: 2, Char: "ж"},
		{Index: 9, Line: 3, Column: 1, Char: "Д"},
		{Index: 10, Line: 3, Column: 2, Char: "а"},
		{Index: 12, Line: 3, Column: 4, Char: "😀"},
	}
	if !reflect.DeepEqual(plan.untypeable, want) {
		t.Errorf("untypeable = %+v\nwant %+v", plan.untypeable, want)
	}
	if plan.untypeableCount != len(want) {
		t.Errorf("untypeableCount = %d", plan.untypeableCount)
	}
	// a b Enter Tab x Enter space
	if got := planIndexes(plan); !reflect.DeepEqual(got, []int{0, 1, 3, 5, 7, 8, 11}) {
		t.Errorf("indexes = %v", got)
	}
}

func TestPlanPasteCapsTheReport(t *testing.T) {
	text := strings.Repeat("ж", maxUntypeableReported+50) + "a"
	plan := planPaste(text, mustLayout(t, "us"))
	if len(plan.untypeable) != maxUntypeableReported {
		t.Errorf("reported %d, want %d", len(plan.untypeable), maxUntypeableReported)
	}
	if plan.untypeableCount != maxUntypeableReported+50 {
		t.Errorf("untypeableCount = %d", plan.untypeableCount)
	}
	if len(plan.steps) != 1 || plan.steps[0].index != maxUntypeableReported+50 {
		t.Errorf("steps = %+v", plan.steps)
	}
}

func TestPlanPasteRussian(t *testing.T) {
	plan := planPaste("Привет, мир!", mustLayout(t, "ru"))
	if plan.untypeableCount != 0 {
		t.Fatalf("untypeable = %+v", plan.untypeable)
	}
	want := [][]Char{
		{{modShift, 10}}, {{0, 11}}, {{0, 5}}, {{0, 7}}, {{0, 23}}, {{0, 17}}, // Привет
		{{modShift, 56}}, {{0, usageSpace}}, // ", "
		{{0, 25}}, {{0, 5}}, {{0, 11}}, // мир
		{{modShift, 30}}, // !
	}
	if got := planCodes(plan); !reflect.DeepEqual(got, want) {
		t.Errorf("strokes = %+v\nwant %+v", got, want)
	}
}

// fakeTyper records the key presses of a paste. Each step can be held until
// the test releases it.
type fakeTyper struct {
	mu      sync.Mutex
	typed   [][]Char
	closed  bool
	failAt  int // 1-based step that fails, 0 for none
	failErr error
	gate    chan struct{}
	entered chan int
}

func (f *fakeTyper) typeStep(ctx context.Context, strokes []Char, _ time.Duration) error {
	f.mu.Lock()
	n := len(f.typed) + 1
	f.mu.Unlock()

	if f.entered != nil {
		f.entered <- n
	}
	if f.gate != nil {
		<-f.gate
	}
	if n == f.failAt {
		return f.failErr
	}

	f.mu.Lock()
	f.typed = append(f.typed, strokes)
	f.mu.Unlock()
	return nil
}

func (f *fakeTyper) close() {
	f.mu.Lock()
	f.closed = true
	f.mu.Unlock()
}

func waitPaste(t *testing.T, m *pasteManager) proto.PasteStatusRsp {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		status := m.getStatus()
		if status.Status != pasteStatusTyping {
			return status
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("paste did not finish")
	return proto.PasteStatusRsp{}
}

func TestPasteManagerTypesInTheBackground(t *testing.T) {
	m := newPasteManager()
	if status := m.getStatus(); status.Status != pasteStatusIdle {
		t.Fatalf("initial status = %+v", status)
	}

	plan := planPaste("héllo\r\n", mustLayout(t, "de"))
	typer := &fakeTyper{}
	status, err := m.start(plan, time.Millisecond, typer)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != pasteStatusTyping || status.Total != 6 {
		t.Errorf("start status = %+v", status)
	}

	status = waitPaste(t, m)
	if status != (proto.PasteStatusRsp{Status: pasteStatusDone, Typed: 6, Total: 6}) {
		t.Errorf("final status = %+v", status)
	}
	if !typer.closed {
		t.Error("typer not closed")
	}
	if !reflect.DeepEqual(typer.typed, planCodes(plan)) {
		t.Errorf("typed = %+v", typer.typed)
	}
}

func TestPasteManagerCancel(t *testing.T) {
	m := newPasteManager()
	typer := &fakeTyper{gate: make(chan struct{}), entered: make(chan int, 16)}
	plan := planPaste(strings.Repeat("a", 10), mustLayout(t, "us"))
	if _, err := m.start(plan, time.Millisecond, typer); err != nil {
		t.Fatal(err)
	}

	// Let three characters through, then hold the fourth.
	for i := 0; i < 3; i++ {
		<-typer.entered
		typer.gate <- struct{}{}
	}
	<-typer.entered

	// A second paste is refused while one runs.
	other := &fakeTyper{}
	if status, err := m.start(plan, time.Millisecond, other); !errors.Is(err, errPasteInProgress) || status.Status != pasteStatusTyping {
		t.Errorf("second start = %+v, %v", status, err)
	}

	// Cancel while the fourth character is being typed. It finishes; nothing
	// after it starts.
	m.mu.Lock()
	cancel := m.cancel
	m.mu.Unlock()
	cancel(errPasteCanceled)
	typer.gate <- struct{}{}
	status := waitPaste(t, m)

	if status.Status != pasteStatusCanceled || status.Typed != 4 || status.Total != 10 {
		t.Errorf("status = %+v", status)
	}
	if len(typer.typed) != 4 || !typer.closed {
		t.Errorf("typed %d, closed %v", len(typer.typed), typer.closed)
	}
	if _, err := m.stop(time.Second); !errors.Is(err, errNoPaste) {
		t.Errorf("stop with nothing running = %v", err)
	}

	// The next paste starts afresh.
	if _, err := m.start(plan, time.Millisecond, &fakeTyper{}); err != nil {
		t.Fatal(err)
	}
	if status := waitPaste(t, m); status.Status != pasteStatusDone || status.Typed != 10 {
		t.Errorf("next paste = %+v", status)
	}
}

func TestPasteManagerCancelDuringDelay(t *testing.T) {
	m := newPasteManager()
	typer := &fakeTyper{entered: make(chan int, 16)}
	plan := planPaste("abc", mustLayout(t, "us"))
	if _, err := m.start(plan, time.Hour, typer); err != nil {
		t.Fatal(err)
	}
	<-typer.entered

	status, err := m.stop(5 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != pasteStatusCanceled || status.Typed != 1 {
		t.Errorf("status = %+v", status)
	}
}

func TestPasteManagerFailure(t *testing.T) {
	tests := []struct {
		err  error
		code string
	}{
		{errors.New("write /dev/hidg0: broken pipe"), pasteErrorHID},
		{errPasteControlBusy, pasteErrorControlBusy},
	}
	for _, tt := range tests {
		m := newPasteManager()
		typer := &fakeTyper{failAt: 3, failErr: tt.err}
		if _, err := m.start(planPaste("abcdef", mustLayout(t, "us")), time.Millisecond, typer); err != nil {
			t.Fatal(err)
		}
		status := waitPaste(t, m)
		want := proto.PasteStatusRsp{Status: pasteStatusFailed, Typed: 2, Total: 6, Error: tt.code}
		if status != want {
			t.Errorf("%v: status = %+v, want %+v", tt.err, status, want)
		}
		if !typer.closed {
			t.Error("typer not closed")
		}
	}
}

// The keyboard is held for one character at a time, so a control mode switch
// never waits for a whole paste. It waits at most for the longest character.
func TestPasteHoldsTheKeyboardPerCharacter(t *testing.T) {
	const modeSwitchWait = 30 * time.Second
	longest := 0
	for _, layout := range layouts {
		for _, keys := range layout.composed {
			longest = max(longest, len(keys))
		}
		longest = max(longest, 2) // a dead key and Space
	}
	if hold := time.Duration(longest) * maxPasteDelay; hold >= modeSwitchWait {
		t.Errorf("one character holds the keyboard for %s", hold)
	}
}

func TestPasteDelay(t *testing.T) {
	tests := []struct {
		ms   int
		want time.Duration
		ok   bool
	}{
		{0, defaultPasteDelay, true},
		{10, 10 * time.Millisecond, true},
		{500, 500 * time.Millisecond, true},
		{9, 0, false},
		{501, 0, false},
		{-1, 0, false},
	}
	for _, tt := range tests {
		got, err := pasteDelay(tt.ms)
		if (err == nil) != tt.ok || got != tt.want {
			t.Errorf("pasteDelay(%d) = %s, %v", tt.ms, got, err)
		}
	}
}

func postJSON(t *testing.T, handler gin.HandlerFunc, body any) proto.Response {
	t.Helper()
	gin.SetMode(gin.TestMode)
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/hid/paste", strings.NewReader(string(data)))
	c.Request.Header.Set("Content-Type", "application/json")
	handler(c)

	var rsp proto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil {
		t.Fatalf("response %q: %v", w.Body.String(), err)
	}
	return rsp
}

func TestPasteHandlerRefusesBeforeTyping(t *testing.T) {
	s := &Service{paste: newPasteManager()}

	tests := []struct {
		name string
		body map[string]any
		code int
	}{
		{"empty", map[string]any{"content": ""}, -1},
		{"unknown layout", map[string]any{"content": "a", "layout": "xx"}, -1},
		{"delay out of range", map[string]any{"content": "a", "delay": 5000}, -1},
		{"too long", map[string]any{"content": strings.Repeat("a", maxPasteRunes+1)}, -2},
		{"untypeable", map[string]any{"content": "aжb", "layout": "us"}, -4},
		{"legacy langue", map[string]any{"content": "aжb", "langue": "en"}, -4},
		{"nothing typeable", map[string]any{"content": "ж", "skipUntypeable": true}, -4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp := postJSON(t, s.Paste, tt.body)
			if rsp.Code != tt.code {
				t.Errorf("code = %d (%s), want %d", rsp.Code, rsp.Msg, tt.code)
			}
		})
	}
	if status := s.paste.getStatus(); status.Status != pasteStatusIdle {
		t.Errorf("a refused paste changed the status to %+v", status)
	}

	rsp := postJSON(t, s.Paste, map[string]any{"content": "a\nжb", "layout": "us"})
	data, _ := json.Marshal(rsp.Data)
	var check proto.PasteCheckRsp
	if err := json.Unmarshal(data, &check); err != nil {
		t.Fatal(err)
	}
	want := []proto.PasteUntypeable{{Index: 2, Line: 2, Column: 1, Char: "ж"}}
	if !reflect.DeepEqual(check.Untypeable, want) || check.UntypeableCount != 1 {
		t.Errorf("check = %+v", check)
	}
}

func TestCheckPasteHandler(t *testing.T) {
	s := &Service{paste: newPasteManager()}
	rsp := postJSON(t, s.CheckPaste, map[string]any{"content": "àé\r\n", "layout": "de", "delay": 20})
	if rsp.Code != 0 {
		t.Fatalf("code = %d (%s)", rsp.Code, rsp.Msg)
	}
	data, _ := json.Marshal(rsp.Data)
	var check proto.PasteCheckRsp
	if err := json.Unmarshal(data, &check); err != nil {
		t.Fatal(err)
	}
	want := proto.PasteCheckRsp{
		Characters: 3, Keystrokes: 5, DurationMs: 100,
		Untypeable: []proto.PasteUntypeable{},
	}
	if !reflect.DeepEqual(check, want) {
		t.Errorf("check = %+v, want %+v", check, want)
	}
}
