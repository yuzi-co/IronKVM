//go:build linux

package audio

import (
	"bytes"
	"fmt"
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

// countedChild returns a newCmd that runs script under sh, and a channel that
// closes once the child has been started n times. The logging tests wait on the
// channel, so they judge a known number of attempts rather than however many a
// loaded host fits into a fixed interval.
func countedChild(n int, script string) (func() *exec.Cmd, <-chan struct{}) {
	var mutex sync.Mutex
	var starts int
	reached := make(chan struct{})

	return func() *exec.Cmd {
		mutex.Lock()
		starts++
		if starts == n {
			close(reached)
		}
		mutex.Unlock()

		return exec.Command("sh", "-c", script)
	}, reached
}

// steadyStateAttempts is how many failed starts the logging tests wait for.
// It is several times quietAfterFailures, so a source that kept logging past
// that point would write well over the limit those tests allow.
const steadyStateAttempts = 30

func TestRunDeliversFullChunks(t *testing.T) {
	source := NewSource()
	// Two chunks of zeros, then exit. head -c reads from /dev/zero.
	source.newCmd = func() *exec.Cmd {
		return exec.Command("sh", "-c", "head -c 7680 /dev/zero")
	}

	got := &collector{}
	delivered := make(chan struct{})
	var deliveredOnce sync.Once
	handle := func(chunk []byte) {
		got.handle(chunk)
		if got.count() >= 2 {
			deliveredOnce.Do(func() { close(delivered) })
		}
	}

	done := make(chan struct{})
	go func() {
		source.Run(handle)
		close(done)
	}()

	// Stop once two chunks are in rather than after a fixed interval, which a
	// loaded host can spend before the child has written anything. Run only
	// returns on Stop, so done closing first means it gave up on its own.
	select {
	case <-delivered:
	case <-done:
		t.Fatalf("Run returned before Stop, after %d chunks", got.count())
	}
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

	// Far more attempts than the old eight-attempt budget.
	const wantAtLeast = 20

	var starts int
	var mutex sync.Mutex
	reached := make(chan struct{})

	source.newCmd = func() *exec.Cmd {
		mutex.Lock()
		starts++
		if starts == wantAtLeast {
			close(reached)
		}
		mutex.Unlock()

		return exec.Command("sh", "-c", "exit 1")
	}

	done := make(chan struct{})
	go func() {
		source.Run(func([]byte) {})
		close(done)
	}()

	// Wait for the attempts themselves rather than for a stretch of time that
	// is meant to hold them. A loaded host fits fewer into any fixed interval,
	// and a Run that retires the child closes done before it gets there.
	select {
	case <-reached:
	case <-done:
		mutex.Lock()
		got := starts
		mutex.Unlock()

		t.Fatalf("Run retired a failing child after %d starts instead of retrying it", got)
	}

	select {
	case <-done:
		t.Fatal("Run retired a failing child instead of retrying it")
	default:
	}

	source.Stop()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after Stop")
	}
}

// Retrying forever must not mean writing forever. The log that fills is
// /tmp/nanokvm-server.log, which S98vidiag and the supervisor both read, and
// the device it fills is the boot SD card.
func TestRunStopsLoggingOnceFailureIsTheSteadyState(t *testing.T) {
	var captured bytes.Buffer
	original := log.StandardLogger().Out
	log.SetOutput(&captured)
	t.Cleanup(func() { log.SetOutput(original) })

	source := NewSource()
	source.minBackoff = time.Millisecond
	source.maxBackoff = time.Millisecond
	newCmd, attempted := countedChild(steadyStateAttempts, "exit 1")
	source.newCmd = newCmd

	done := make(chan struct{})
	go func() {
		source.Run(func([]byte) {})
		close(done)
	}()

	select {
	case <-attempted:
	case <-done:
		t.Fatal("Run returned before Stop")
	}
	source.Stop()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after Stop")
	}

	// A handful of lines describes the fault; the rest only wear the card.
	const wantAtMost = 10

	if lines := strings.Count(captured.String(), "audio capture"); lines > wantAtMost {
		t.Errorf("wrote %d audio capture log lines, want no more than %d", lines, wantAtMost)
	}
}

// The quiet promise covers one line and has to cover both. On a device on
// 2026-08-19 the retry message fell silent after five attempts exactly as
// designed, and arecord's own complaint kept arriving every fifteen seconds for
// hours. That complaint was an idle host, which is no longer a failure at all
// (see TestRunTreatsAnIdleHostAsIdleNotAsAFailure), but a real fault that
// repeats forever needs the same promise.
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
	newCmd, attempted := countedChild(steadyStateAttempts,
		"echo 'arecord: main:850: audio open error: Device or resource busy' >&2; exit 1")
	source.newCmd = newCmd

	done := make(chan struct{})
	go func() {
		source.Run(func([]byte) {})
		close(done)
	}()

	select {
	case <-attempted:
	case <-done:
		t.Fatal("Run returned before Stop")
	}
	source.Stop()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after Stop")
	}

	// Silence is not the goal. An operator who has to diagnose no audio needs
	// the reason the child gave, so it has to appear.
	if !strings.Contains(captured.String(), "Device or resource busy") {
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

// idleComplaint is what arecord said on a board on 2026-09-30 while the host
// played nothing to the gadget.
const idleComplaint = "echo 'arecord: pcm_read:2285: read error: I/O error' >&2; exit 1"

// busyComplaint is a real fault: another reader holds the card.
const busyComplaint = "echo 'arecord: main:850: audio open error: Device or resource busy' >&2; exit 1"

func TestIsHostIdle(t *testing.T) {
	const idle = "arecord: pcm_read:2285: read error: I/O error"

	cases := []struct {
		name      string
		delivered bool
		uptime    time.Duration
		reason    string
		want      bool
	}{
		{"the idle host seen on the board", false, 390 * time.Millisecond, idle, true},
		{"another arecord build's line number", false, time.Second, "arecord: pcm_read:2240: read error: I/O error", true},
		{"audio came first, so the host stopped mid-run", true, 390 * time.Millisecond, idle, false},
		{"ran as long as a healthy child", false, minRunDuration + time.Second, idle, false},
		{"the card is held by someone else", false, 50 * time.Millisecond, "arecord: main:850: audio open error: Device or resource busy", false},
		{"the gadget has no card", false, 50 * time.Millisecond, "arecord: main:850: audio open error: No such file or directory", false},
		{"a silent crash", false, 50 * time.Millisecond, "", false},
	}

	for _, c := range cases {
		if got := isHostIdle(c.delivered, c.uptime, c.reason); got != c.want {
			t.Errorf("%s: isHostIdle = %v, want %v", c.name, got, c.want)
		}
	}
}

// stateLog records what a source reports through onState.
type stateLog struct {
	mutex  sync.Mutex
	states []State
}

func (l *stateLog) record(state State) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.states = append(l.states, state)
}

func (l *stateLog) all() []State {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return append([]State(nil), l.states...)
}

// captureLog sends the standard logger to a buffer for the rest of the test.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()

	var captured bytes.Buffer
	original := log.StandardLogger().Out
	log.SetOutput(&captured)
	t.Cleanup(func() { log.SetOutput(original) })

	return &captured
}

// fastSource is a source that retries at once and reports to states.
func fastSource(states *stateLog) *Source {
	source := NewSource()
	source.minBackoff = time.Millisecond
	source.maxBackoff = time.Millisecond
	source.onState = states.record

	return source
}

// scriptedChild returns a newCmd that runs script(n) for the nth start, and a
// channel that closes once the child has been started until times.
func scriptedChild(until int, script func(n int) string) (func() *exec.Cmd, <-chan struct{}) {
	var mutex sync.Mutex
	var starts int
	reached := make(chan struct{})

	return func() *exec.Cmd {
		mutex.Lock()
		starts++
		n := starts
		if n == until {
			close(reached)
		}
		mutex.Unlock()

		return exec.Command("sh", "-c", script(n))
	}, reached
}

// runUntil runs source until reached closes, then stops it and waits.
func runUntil(t *testing.T, source *Source, handle func([]byte), reached <-chan struct{}) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		source.Run(handle)
		close(done)
	}()

	select {
	case <-reached:
	case <-done:
		t.Fatal("Run returned before Stop")
	case <-time.After(10 * time.Second):
		t.Fatal("the expected attempts never happened")
	}
	source.Stop()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after Stop")
	}
}

// An idle host is this board's ordinary state and not a fault. It gets one
// line at info, saying what is going on, and no warning at all.
func TestRunTreatsAnIdleHostAsIdleNotAsAFailure(t *testing.T) {
	captured := captureLog(t)
	states := &stateLog{}
	source := fastSource(states)
	newCmd, attempted := countedChild(steadyStateAttempts, idleComplaint)
	source.newCmd = newCmd

	runUntil(t, source, func([]byte) {}, attempted)

	out := captured.String()
	if n := strings.Count(out, "the host is not playing audio to the KVM"); n != 1 {
		t.Errorf("said the host is idle %d times, want once:\n%s", n, out)
	}
	if strings.Contains(out, "level=warning") {
		t.Errorf("an idle host produced a warning:\n%s", out)
	}
	if got := states.all(); len(got) != 1 || got[0] != StateIdle {
		t.Errorf("reported %v, want only %v", got, StateIdle)
	}
}

// When the host starts playing, audio flows on the next attempt, the log says
// so once, and the state follows.
func TestRunPicksUpAudioOnceTheHostPlays(t *testing.T) {
	captured := captureLog(t)
	states := &stateLog{}
	source := fastSource(states)

	const idleAttempts = 3
	newCmd, _ := scriptedChild(0, func(n int) string {
		if n <= idleAttempts {
			return idleComplaint
		}
		// The host plays: two chunks, then the read stays open.
		return fmt.Sprintf("head -c %d /dev/zero; exec sleep 60", 2*ChunkBytes)
	})
	source.newCmd = newCmd

	got := &collector{}
	reached := make(chan struct{})
	var once sync.Once
	runUntil(t, source, func(chunk []byte) {
		got.handle(chunk)
		if got.count() >= 2 {
			once.Do(func() { close(reached) })
		}
	}, reached)

	out := captured.String()
	if n := strings.Count(out, "the host is not playing audio to the KVM"); n != 1 {
		t.Errorf("said the host is idle %d times, want once:\n%s", n, out)
	}
	if n := strings.Count(out, "the host is playing audio to the KVM again"); n != 1 {
		t.Errorf("said the host plays again %d times, want once:\n%s", n, out)
	}
	if strings.Contains(out, "level=warning") {
		t.Errorf("an idle host that started playing produced a warning:\n%s", out)
	}

	want := []State{StateIdle, StatePlaying}
	if reported := states.all(); fmt.Sprint(reported) != fmt.Sprint(want) {
		t.Errorf("reported %v, want %v", reported, want)
	}
}

// A real fault keeps its warning and tells the viewer, and an idle host after
// it is announced afresh.
func TestRunReportsARealFailureAsFailing(t *testing.T) {
	captured := captureLog(t)
	states := &stateLog{}
	source := fastSource(states)
	newCmd, reached := scriptedChild(4, func(n int) string {
		if n == 2 {
			return busyComplaint
		}
		return idleComplaint
	})
	source.newCmd = newCmd

	runUntil(t, source, func([]byte) {}, reached)

	out := captured.String()
	if !strings.Contains(out, "level=warning") || !strings.Contains(out, "Device or resource busy") {
		t.Errorf("a real fault was not warned about:\n%s", out)
	}
	if n := strings.Count(out, "the host is not playing audio to the KVM"); n != 2 {
		t.Errorf("said the host is idle %d times, want twice, before and after the fault:\n%s", n, out)
	}

	got := states.all()
	if len(got) < 3 || got[0] != StateIdle || got[1] != StateFailing || got[2] != StateIdle {
		t.Errorf("reported %v, want %v, %v, %v", got, StateIdle, StateFailing, StateIdle)
	}
}

// A named source, the room microphone, has no host that may be idle. An I/O
// error from it is a failure, logged under its own prefix with the reason,
// and never in the host's words.
func TestANamedSourceLogsItsFailureUnderItsOwnName(t *testing.T) {
	captured := captureLog(t)
	states := &stateLog{}
	source := NewSourceFor(Format{Device: "hw:test,0", Channels: 1, Bitrate: 32000, Name: "room microphone"})
	source.minBackoff = time.Millisecond
	source.maxBackoff = time.Millisecond
	source.onState = states.record
	newCmd, attempted := countedChild(steadyStateAttempts, idleComplaint)
	source.newCmd = newCmd

	runUntil(t, source, func([]byte) {}, attempted)

	out := captured.String()
	if strings.Contains(out, "host") {
		t.Errorf("the microphone's failure was told in the host's words:\n%s", out)
	}
	want := "room microphone: capture failed: arecord: pcm_read:2285: read error: I/O error"
	if n := strings.Count(out, want); n != quietAfterFailures {
		t.Errorf("said %q %d times, want %d:\n%s", want, n, quietAfterFailures, out)
	}
	if strings.Contains(out, "audio capture") {
		t.Errorf("the microphone used the host audio's prefix:\n%s", out)
	}
	if got := states.all(); len(got) != 1 || got[0] != StateFailing {
		t.Errorf("reported %v, want only %v", got, StateFailing)
	}
}
