package watchdog

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestTheLogKeepsTheLastActionsAndTheirScreenshots(t *testing.T) {
	dir := t.TempDir()
	l := OpenLog(dir)

	start := time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)
	for i := range logLimit + 2 {
		at := start.Add(time.Duration(i) * time.Minute)
		e := Entry{ID: strconv.FormatInt(at.UnixNano(), 10), Time: at, Action: ActionReset, Reason: ReasonFrozen}
		if err := l.Add(e, []byte("jpeg")); err != nil {
			t.Fatal(err)
		}
	}

	entries := l.Entries()
	if len(entries) != logLimit {
		t.Fatalf("kept %d entries", len(entries))
	}
	if !entries[0].Time.After(entries[1].Time) {
		t.Fatal("the newest entry is not first")
	}

	// The two oldest fell off and took their screenshots with them.
	shots, _ := filepath.Glob(filepath.Join(dir, "*.jpg"))
	if len(shots) != logLimit {
		t.Fatalf("%d screenshots on disk", len(shots))
	}
	oldest := strconv.FormatInt(start.UnixNano(), 10)
	if _, err := os.Stat(filepath.Join(dir, oldest+".jpg")); !os.IsNotExist(err) {
		t.Fatalf("the oldest screenshot is still there: %v", err)
	}

	// The log survives a restart.
	if again := OpenLog(dir).Entries(); len(again) != logLimit || again[0].ID != entries[0].ID {
		t.Fatalf("reloaded %d entries", len(again))
	}
}

// Only an entry's own screenshot is served, whatever the request names.
func TestTheScreenshotPathComesFromTheLog(t *testing.T) {
	dir := t.TempDir()
	l := OpenLog(dir)
	if err := l.Add(Entry{ID: "100", Action: ActionReset}, []byte("jpeg")); err != nil {
		t.Fatal(err)
	}
	if err := l.Add(Entry{ID: "200", Action: ActionReset}, nil); err != nil {
		t.Fatal(err)
	}

	if path, ok := l.Screenshot("100"); !ok || path != filepath.Join(dir, "100.jpg") {
		t.Fatalf("100: %q %t", path, ok)
	}
	for _, id := range []string{"200", "300", "../log", "100.jpg", ""} {
		if path, ok := l.Screenshot(id); ok {
			t.Errorf("%q gave %q", id, path)
		}
	}
}

// A log file that does not parse gives an empty log rather than no watchdog.
func TestABrokenLogFileStartsAnEmptyLog(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, logFile), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	if entries := OpenLog(dir).Entries(); len(entries) != 0 {
		t.Fatalf("got %d entries", len(entries))
	}
}
