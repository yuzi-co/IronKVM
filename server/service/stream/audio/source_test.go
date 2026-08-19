//go:build linux

package audio

import (
	"bytes"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

// collector gathers what Run hands out, from whichever goroutine calls it.
type collector struct {
	mutex  sync.Mutex
	chunks [][]byte
}

func (c *collector) handle(chunk []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.chunks = append(c.chunks, append([]byte(nil), chunk...))
}

func (c *collector) count() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return len(c.chunks)
}

func TestRunDeliversFullChunks(t *testing.T) {
	source := NewSource()
	// Two chunks of zeros, then exit. head -c reads from /dev/zero.
	source.newCmd = func() *exec.Cmd {
		return exec.Command("sh", "-c", "head -c 7680 /dev/zero")
	}

	got := &collector{}

	done := make(chan struct{})
	go func() {
		source.Run(got.handle)
		close(done)
	}()

	time.Sleep(500 * time.Millisecond)
	source.Stop()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after Stop")
	}

	if got.count() < 2 {
		t.Fatalf("got %d chunks, want at least 2", got.count())
	}

	if n := len(got.chunks[0]); n != ChunkBytes {
		t.Errorf("first chunk is %d bytes, want %d", n, ChunkBytes)
	}
}

func TestRunReturnsAfterStop(t *testing.T) {
	source := NewSource()
	// A child that never writes and never exits, which is what arecord does
	// while the host plays nothing.
	source.newCmd = func() *exec.Cmd {
		return exec.Command("sh", "-c", "sleep 60")
	}

	done := make(chan struct{})
	go func() {
		source.Run(func([]byte) {})
		close(done)
	}()

	time.Sleep(200 * time.Millisecond)
	source.Stop()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after Stop")
	}
}

// TestRunKeepsRetryingAFailingChild states the rule that matters on this
// device: a host that is not streaming audio is the ordinary idle state, not a
// fault. Capture has to be there when the host finally plays something, and
// StartAudioStream fires on an ICE state change that a settled connection
// never produces again - so a Source that retires itself is audio that never
// comes back without a page reload.
func TestRunKeepsRetryingAFailingChild(t *testing.T) {
	source := NewSource()
	source.minBackoff = time.Millisecond
	source.maxBackoff = 2 * time.Millisecond

	var starts int
	var mutex sync.Mutex

	source.newCmd = func() *exec.Cmd {
		mutex.Lock()
		starts++
		mutex.Unlock()

		return exec.Command("sh", "-c", "exit 1")
	}

	done := make(chan struct{})
	go func() {
		source.Run(func([]byte) {})
		close(done)
	}()

	// Long enough for far more attempts than the old eight-attempt budget.
	time.Sleep(500 * time.Millisecond)

	select {
	case <-done:
		t.Fatal("Run retired a failing child instead of retrying it")
	default:
	}

	const wantAtLeast = 20

	mutex.Lock()
	got := starts
	mutex.Unlock()

	if got < wantAtLeast {
		t.Errorf("started the child %d times, want at least %d", got, wantAtLeast)
	}

	source.Stop()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after Stop")
	}
}

// Retrying forever must not mean writing forever. The log that fills is
// /tmp/nanokvm-server.log, which S99vidiag and the supervisor both read, and
// the device it fills is the boot SD card.
func TestRunStopsLoggingOnceFailureIsTheSteadyState(t *testing.T) {
	var captured bytes.Buffer
	original := log.StandardLogger().Out
	log.SetOutput(&captured)
	t.Cleanup(func() { log.SetOutput(original) })

	source := NewSource()
	source.minBackoff = time.Millisecond
	source.maxBackoff = time.Millisecond
	source.newCmd = func() *exec.Cmd {
		return exec.Command("sh", "-c", "exit 1")
	}

	done := make(chan struct{})
	go func() {
		source.Run(func([]byte) {})
		close(done)
	}()

	time.Sleep(500 * time.Millisecond)
	source.Stop()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after Stop")
	}

	// Dozens of attempts fit in half a second. A handful of lines describes
	// the fault; the rest only wear the card.
	const wantAtMost = 10

	if lines := strings.Count(captured.String(), "audio capture"); lines > wantAtMost {
		t.Errorf("wrote %d audio capture log lines, want no more than %d", lines, wantAtMost)
	}
}

// The quiet promise covers one line and has to cover both. On a device on
// 2026-08-19 the retry message fell silent after five attempts exactly as
// designed, and arecord's own complaint kept arriving every fifteen seconds for
// hours: "audio capture said: arecord: pcm_read:2240: read error: I/O error".
// A host that plays nothing is the ordinary idle state of this board, so that
// is the steady state, not an incident.
//
// TestRunStopsLoggingOnceFailureIsTheSteadyState does not catch it, because the
// child it runs exits without writing anything, and a silent child has no
// complaint to repeat.
func TestRunStopsRepeatingTheChildsComplaint(t *testing.T) {
	var captured bytes.Buffer
	original := log.StandardLogger().Out
	log.SetOutput(&captured)
	t.Cleanup(func() { log.SetOutput(original) })

	source := NewSource()
	source.minBackoff = time.Millisecond
	source.maxBackoff = time.Millisecond
	source.newCmd = func() *exec.Cmd {
		return exec.Command("sh", "-c", "echo 'pcm_read:2240: read error: I/O error' >&2; exit 1")
	}

	done := make(chan struct{})
	go func() {
		source.Run(func([]byte) {})
		close(done)
	}()

	time.Sleep(500 * time.Millisecond)
	source.Stop()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after Stop")
	}

	// Silence is not the goal. An operator who has to diagnose no audio needs
	// the reason the child gave, so it has to appear.
	if !strings.Contains(captured.String(), "read error") {
		t.Error("the child's reason was never reported, so nothing says why audio is off")
	}

	const wantAtMost = 10

	if lines := strings.Count(captured.String(), "audio capture"); lines > wantAtMost {
		t.Errorf("wrote %d audio capture log lines, want no more than %d", lines, wantAtMost)
	}
}

func TestStopBeforeRun(t *testing.T) {
	source := NewSource()
	// This child would run indefinitely, but Stop is called before Run.
	source.newCmd = func() *exec.Cmd {
		return exec.Command("sh", "-c", "sleep 60")
	}

	source.Stop()

	done := make(chan struct{})
	go func() {
		source.Run(func([]byte) {})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return immediately when Stop was called before Run")
	}
}

func TestStopCalledTwice(t *testing.T) {
	source := NewSource()
	source.newCmd = func() *exec.Cmd {
		return exec.Command("sh", "-c", "sleep 60")
	}

	done := make(chan struct{})
	go func() {
		source.Run(func([]byte) {})
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	source.Stop()
	source.Stop() // Should be safe to call again

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after calling Stop twice")
	}
}

// TestMinRunDurationReset checks that a child which emits a chunk and exits at
// once is not counted as healthy. Nothing gives up any more, so the visible
// consequence is the backoff: a healthy child returns it to minBackoff, and a
// crash loop must instead let it climb.
func TestMinRunDurationReset(t *testing.T) {
	source := NewSource()
	source.minBackoff = 5 * time.Millisecond
	source.maxBackoff = 200 * time.Millisecond

	var starts int
	var mutex sync.Mutex

	source.newCmd = func() *exec.Cmd {
		mutex.Lock()
		starts++
		mutex.Unlock()

		// Emit one chunk quickly, then exit. This is shorter than minRunDuration.
		return exec.Command("sh", "-c", "head -c 3840 /dev/zero")
	}

	done := make(chan struct{})
	go func() {
		source.Run(func([]byte) {})
		close(done)
	}()

	time.Sleep(time.Second)
	source.Stop()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after Stop")
	}

	// Climbing from 5 ms to the 200 ms cap allows about ten attempts in a
	// second. Treating this child as healthy would hold the backoff at 5 ms
	// and produce an order of magnitude more.
	const tooMany = 30

	mutex.Lock()
	defer mutex.Unlock()

	if starts >= tooMany {
		t.Errorf("started the child %d times, want fewer than %d: the backoff did not climb",
			starts, tooMany)
	}
}
