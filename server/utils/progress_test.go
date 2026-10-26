package utils

import (
	"bytes"
	"sync"
	"testing"
	"time"
)

func TestFormatPercentage(t *testing.T) {
	tests := []struct {
		loaded, total int64
		want          string
	}{
		{0, 100, ""},
		{10, 0, ""},
		{1, 3, "33.33%"},
		{50, 100, "50.00%"},
		{100, 100, "100.00%"},
	}
	for _, tt := range tests {
		if got := FormatPercentage(tt.loaded, tt.total); got != tt.want {
			t.Errorf("FormatPercentage(%d, %d) = %q, want %q", tt.loaded, tt.total, got, tt.want)
		}
	}
}

func withProgressInterval(t *testing.T, d time.Duration) {
	t.Helper()
	old := progressInterval
	progressInterval = d
	t.Cleanup(func() { progressInterval = old })
}

// The writer passes every byte on and reports the running count, against the
// expected total, on its ticker.
func TestProgressWriterReportsTheBytesWritten(t *testing.T) {
	withProgressInterval(t, 5*time.Millisecond)

	var mu sync.Mutex
	var reports [][2]int64
	var out bytes.Buffer
	pw := NewProgressWriter(&out, 10, func(loaded, total int64) {
		mu.Lock()
		reports = append(reports, [2]int64{loaded, total})
		mu.Unlock()
	})
	defer pw.Stop()

	if _, err := pw.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if _, err := pw.Write([]byte("world")); err != nil {
		t.Fatal(err)
	}
	if out.String() != "helloworld" {
		t.Fatalf("passed on %q", out.String())
	}
	if pw.Written() != 10 {
		t.Fatalf("Written = %d, want 10", pw.Written())
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		mu.Lock()
		last := [2]int64{}
		if len(reports) > 0 {
			last = reports[len(reports)-1]
		}
		mu.Unlock()
		if last == [2]int64{10, 10} {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("last report %v, want [10 10]", last)
		}
		time.Sleep(time.Millisecond)
	}
}

// With no expected total there is no percentage to report, and Stop may be
// called twice, or on nil, so callers can defer it.
func TestProgressWriterWithoutATotalStaysQuiet(t *testing.T) {
	withProgressInterval(t, time.Millisecond)

	called := make(chan struct{}, 1)
	pw := NewProgressWriter(&bytes.Buffer{}, 0, func(int64, int64) {
		select {
		case called <- struct{}{}:
		default:
		}
	})
	_, _ = pw.Write([]byte("data"))
	time.Sleep(20 * time.Millisecond)
	pw.Stop()
	pw.Stop()

	var nilWriter *ProgressWriter
	nilWriter.Stop()

	select {
	case <-called:
		t.Fatal("reported progress without a total")
	default:
	}
}
