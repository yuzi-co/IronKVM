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

func TestParseConfig(t *testing.T) {
	cases := []struct {
		content, mode, key string
	}{
		// Files older builds wrote: the mode alone, with or without a newline.
		{"relative", "relative", ""},
		{"absolute\n", "absolute", ""},
		{"", "", ""},
		{"relative\nf15", "relative", KeyF15},
		{"absolute\r\nshift\r\n", "absolute", KeyShift},
		{"relative\nctrl\n", "relative", KeyControl},
		// A key this build does not know falls back to the mouse.
		{"relative\nhyper", "relative", ""},
	}
	for _, c := range cases {
		mode, key := parseConfig(c.content)
		if mode != c.mode || key != c.key {
			t.Fatalf("%q: got (%q, %q), want (%q, %q)", c.content, mode, key, c.mode, c.key)
		}
	}
}

func TestEnableWritesTheKeyAndReadsItBack(t *testing.T) {
	j := newJiggler(t)

	if err := j.Enable("absolute", KeyShift); err != nil {
		t.Fatalf("failed to enable: %s", err)
	}
	content, err := os.ReadFile(ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	if mode, key := parseConfig(string(content)); mode != "absolute" || key != KeyShift {
		t.Fatalf("got (%q, %q) back from %q", mode, key, content)
	}

	// The mouse method writes exactly what older builds wrote.
	if err := j.Enable("relative", ""); err != nil {
		t.Fatalf("failed to enable: %s", err)
	}
	content, _ = os.ReadFile(ConfigFile)
	if string(content) != "relative" {
		t.Fatalf("expected the old format, got %q", content)
	}

	if err := j.Enable("relative", "hyper"); err == nil {
		t.Fatal("expected an unknown key to be refused")
	}
}

// The key survives a Disable, so the menu shows the method the jiggler resumes
// with, and a second Disable is not an error.
func TestDisableKeepsTheKey(t *testing.T) {
	j := newJiggler(t)

	if err := j.Enable("relative", KeyF15); err != nil {
		t.Fatal(err)
	}
	if err := j.Disable(); err != nil {
		t.Fatal(err)
	}
	if err := j.Disable(); err != nil {
		t.Fatalf("expected a second disable to succeed, got %s", err)
	}
	if j.GetKey() != KeyF15 {
		t.Fatalf("expected the key to be kept, got %q", j.GetKey())
	}
}

// The loop presses the key rather than moving the mouse when one is set.
func TestLoopPressesTheKey(t *testing.T) {
	j := newJiggler(t)
	j.key = KeyControl

	pressed := make(chan string, 1)
	j.press = func(key string) {
		select {
		case pressed <- key:
		default:
		}
	}
	j.move = func(string) { t.Error("expected no mouse move") }
	j.lastUpdated = time.Now().Add(-time.Hour)

	if !j.step() {
		t.Fatal("expected the loop to keep running")
	}
	select {
	case key := <-pressed:
		if key != KeyControl {
			t.Fatalf("pressed %q", key)
		}
	default:
		t.Fatal("expected a press")
	}
}

// Shutdown waits for a press in flight, and nothing starts the loop after it.
func TestShutdownWaitsForAPressInFlight(t *testing.T) {
	j := newJiggler(t)
	j.key = KeyF15
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
}
