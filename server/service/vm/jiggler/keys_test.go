package jiggler

import (
	"bytes"
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"
)

func TestKeyReportBytes(t *testing.T) {
	cases := map[string][]byte{
		KeyF15:     {0x00, 0x00, 0x6a, 0x00, 0x00, 0x00, 0x00, 0x00},
		KeyShift:   {0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		KeyControl: {0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	for key, want := range cases {
		got, ok := keyReport(key)
		if !ok {
			t.Fatalf("%s: expected a report", key)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("%s: got % x, want % x", key, got, want)
		}
	}

	for _, key := range []string{"", "F15", "alt", "a"} {
		if _, ok := keyReport(key); ok {
			t.Fatalf("%q: expected no report", key)
		}
	}
}

var releaseReport = []byte{0, 0, 0, 0, 0, 0, 0, 0}

// recorder is a keyboard endpoint that keeps what was written and fails the
// writes it is told to.
type recorder struct {
	writes [][]byte
	fail   map[int]bool
}

func (r *recorder) write(report []byte) error {
	n := len(r.writes)
	r.writes = append(r.writes, append([]byte(nil), report...))
	if r.fail[n] {
		return errors.New("write failed")
	}
	return nil
}

func TestPressKeyPressesThenReleases(t *testing.T) {
	report, _ := keyReport(KeyShift)
	r := &recorder{}

	if err := pressKey(context.Background(), r.write, report, time.Millisecond); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(r.writes) != 2 || !bytes.Equal(r.writes[0], report) || !bytes.Equal(r.writes[1], releaseReport) {
		t.Fatalf("expected press then release, got % x", r.writes)
	}
}

// Manual input cancels the background context. The hold is cut short, and the
// release still goes out before the lane is handed over.
func TestPressKeyReleasesWhenCancelled(t *testing.T) {
	report, _ := keyReport(KeyF15)
	r := &recorder{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	_ = pressKey(ctx, r.write, report, time.Hour)
	if time.Since(start) > time.Second {
		t.Fatal("expected a cancelled hold to end at once")
	}
	if len(r.writes) != 2 || !bytes.Equal(r.writes[1], releaseReport) {
		t.Fatalf("expected a release after a cancelled hold, got % x", r.writes)
	}
}

// A press whose write failed may still have reached the host, so the release
// is sent anyway.
func TestPressKeyReleasesAfterAFailedPress(t *testing.T) {
	report, _ := keyReport(KeyControl)
	r := &recorder{fail: map[int]bool{0: true}}

	if err := pressKey(context.Background(), r.write, report, time.Millisecond); err == nil {
		t.Fatal("expected the press error")
	}
	if len(r.writes) != 2 || !bytes.Equal(r.writes[1], releaseReport) {
		t.Fatalf("expected a release after a failed press, got % x", r.writes)
	}
}

func TestPressKeyRetriesAFailedRelease(t *testing.T) {
	report, _ := keyReport(KeyF15)
	r := &recorder{fail: map[int]bool{1: true}}

	if err := pressKey(context.Background(), r.write, report, time.Millisecond); err != nil {
		t.Fatalf("expected the retried release to succeed, got %s", err)
	}
	if len(r.writes) != 3 || !bytes.Equal(r.writes[2], releaseReport) {
		t.Fatalf("expected a second release, got % x", r.writes)
	}
}

func TestParseKeyConfig(t *testing.T) {
	cases := []struct {
		content string
		enabled bool
		key     string
	}{
		{"on\nf15\n", true, KeyF15},
		{"on\r\nshift\r\n", true, KeyShift},
		{"off\nctrl\n", false, KeyControl},
		// A key this build does not know, or none at all, reads as F15.
		{"on\nhyper\n", true, KeyF15},
		{"on", true, KeyF15},
		{"off\n", false, KeyF15},
		// Anything but "on" is off, so a damaged file never presses keys.
		{"", false, KeyF15},
		{"yes\nshift\n", false, KeyShift},
	}
	for _, c := range cases {
		enabled, key := parseKeyConfig(c.content)
		if enabled != c.enabled || key != c.key {
			t.Fatalf("%q: got (%t, %q), want (%t, %q)", c.content, enabled, key, c.enabled, c.key)
		}
	}
}

func TestReadKeyConfigWithoutAFileIsOffWithF15(t *testing.T) {
	newJiggler(t)

	enabled, key := readKeyConfig()
	if enabled || key != KeyF15 {
		t.Fatalf("got (%t, %q), want off with f15", enabled, key)
	}
}

// The key jiggler keeps its file while off, so the chosen key is what the next
// boot reads back.
func TestSetKeyJigglerPersistsTheKeyWhileOff(t *testing.T) {
	j := newJiggler(t)

	if err := j.SetKeyJiggler(true, KeyShift); err != nil {
		t.Fatal(err)
	}
	if enabled, key := readKeyConfig(); !enabled || key != KeyShift {
		t.Fatalf("got (%t, %q) back, want on with shift", enabled, key)
	}

	if err := j.SetKeyJiggler(false, KeyControl); err != nil {
		t.Fatal(err)
	}
	if enabled, key := readKeyConfig(); enabled || key != KeyControl {
		t.Fatalf("got (%t, %q) back, want off with ctrl", enabled, key)
	}
	if enabled, key := j.KeyJiggler(); enabled || key != KeyControl {
		t.Fatalf("got (%t, %q), want off with ctrl", enabled, key)
	}

	if err := j.SetKeyJiggler(true, "hyper"); err == nil {
		t.Fatal("expected an unknown key to be refused")
	}
	if err := j.SetKeyJiggler(true, ""); err == nil {
		t.Fatal("expected an empty key to be refused")
	}
	if enabled, key := j.KeyJiggler(); enabled || key != KeyControl {
		t.Fatalf("a refused key changed the state to (%t, %q)", enabled, key)
	}
}

// The two jigglers are independent: turning one on or off leaves the other and
// its file alone.
func TestKeyAndMouseJigglersAreIndependent(t *testing.T) {
	j := newJiggler(t)

	if err := j.Enable("absolute"); err != nil {
		t.Fatal(err)
	}
	if err := j.SetKeyJiggler(true, KeyF15); err != nil {
		t.Fatal(err)
	}
	if err := j.Disable(); err != nil {
		t.Fatal(err)
	}
	if enabled, _ := j.KeyJiggler(); !enabled {
		t.Fatal("disabling the mouse jiggler turned the key jiggler off")
	}

	if err := j.Enable("relative"); err != nil {
		t.Fatal(err)
	}
	if err := j.SetKeyJiggler(false, KeyF15); err != nil {
		t.Fatal(err)
	}
	if !j.IsEnabled() {
		t.Fatal("disabling the key jiggler turned the mouse jiggler off")
	}
	content, _ := os.ReadFile(ConfigFile)
	if string(content) != "relative" {
		t.Fatalf("mouse jiggler file changed to %q", content)
	}
}

// recordActions makes j's loop report its moves and presses instead of
// sending them. Only for tests that call step themselves, with no loop running.
func recordActions(j *Jiggler) (moves, presses *[]string) {
	moves, presses = &[]string{}, &[]string{}
	j.move = func(mode string) { *moves = append(*moves, mode) }
	j.press = func(key string) { *presses = append(*presses, key) }
	return moves, presses
}

func TestStepRunsWhicheverJigglersAreOn(t *testing.T) {
	cases := []struct {
		mouse, key             bool
		wantMoves, wantPresses int
	}{
		{true, false, 1, 0},
		{false, true, 0, 1},
		{true, true, 1, 1},
	}
	for _, c := range cases {
		j := newJiggler(t)
		j.enabled = c.mouse
		j.keyEnabled = c.key
		j.key = KeyControl
		moves, presses := recordActions(j)
		j.lastUpdated = time.Now().Add(-time.Hour)

		if !j.step() {
			t.Fatalf("%+v: expected the loop to keep running", c)
		}
		if len(*moves) != c.wantMoves || len(*presses) != c.wantPresses {
			t.Fatalf("%+v: got moves %v presses %v", c, *moves, *presses)
		}
		if c.key && (*presses)[0] != KeyControl {
			t.Fatalf("pressed %q", (*presses)[0])
		}
	}
}

// Real input holds the key jiggler off exactly as it holds the mouse off.
func TestStepWaitsForIdleBeforePressing(t *testing.T) {
	j := newJiggler(t)
	j.enabled = false
	j.keyEnabled = true
	j.interval = time.Hour
	_, presses := recordActions(j)
	j.running = true
	j.Update()

	if !j.step() {
		t.Fatal("expected the loop to keep running")
	}
	if len(*presses) != 0 {
		t.Fatalf("pressed %v while the target was busy", *presses)
	}
}

// The loop starts for the key jiggler alone and stops once neither is on.
func TestLoopRunsForTheKeyJigglerAlone(t *testing.T) {
	j := newJiggler(t)
	j.enabled = false

	if j.Run() {
		t.Fatal("expected no loop with both jigglers off")
	}
	if err := j.SetKeyJiggler(true, KeyF15); err != nil {
		t.Fatal(err)
	}
	j.mutex.Lock()
	running := j.running
	j.mutex.Unlock()
	if !running {
		t.Fatal("expected enabling the key jiggler to start the loop")
	}

	if err := j.SetKeyJiggler(false, KeyF15); err != nil {
		t.Fatal(err)
	}
	if _, running := j.tick(); running {
		t.Fatal("expected the loop to stop with both jigglers off")
	}
}

// Shutdown waits for a press in flight, and nothing starts the loop after it.
func TestShutdownWaitsForAPressInFlight(t *testing.T) {
	j := newJiggler(t)
	j.enabled = false
	j.keyEnabled = true
	j.lastUpdated = time.Now().Add(-time.Hour)

	inPress := make(chan struct{})
	var mu sync.Mutex
	released := false
	j.press = func(string) {
		close(inPress)
		time.Sleep(50 * time.Millisecond)
		mu.Lock()
		released = true
		mu.Unlock()
	}

	go j.step()
	<-inPress
	j.Shutdown()

	mu.Lock()
	defer mu.Unlock()
	if !released {
		t.Fatal("expected Shutdown to return only after the press finished")
	}
	if j.Run() {
		t.Fatal("expected no loop to start after Shutdown")
	}
	if err := j.SetKeyJiggler(true, KeyF15); err != nil {
		t.Fatal(err)
	}
	j.mutex.Lock()
	running := j.running
	j.mutex.Unlock()
	if running {
		t.Fatal("expected enabling after Shutdown to start no loop")
	}
}
