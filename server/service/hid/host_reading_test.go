package hid

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

// fakePending stands in for the kernel's pending_ms.
type fakePending struct {
	mu      sync.Mutex
	pending time.Duration
	err     error
	reads   int
	closed  bool
}

func (f *fakePending) set(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pending = d
}

func (f *fakePending) Pending() (time.Duration, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reads++
	return f.pending, f.err
}

func (f *fakePending) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

// fakeReadTimer records what the watch asked for and fires only by hand.
type fakeReadTimer struct {
	after   time.Duration
	fire    func()
	resets  int
	stopped bool
}

func (t *fakeReadTimer) Reset(d time.Duration) bool {
	t.after = d
	t.resets++
	t.stopped = false
	return true
}

func (t *fakeReadTimer) Stop() bool {
	t.stopped = true
	return true
}

// withPending makes every endpoint's kernel report the given source; nil is a
// kernel without pending_ms. It returns how often the kernel was probed.
func withPending(t *testing.T, source *fakePending) *int {
	t.Helper()

	restore := openPendingReader
	t.Cleanup(func() { openPendingReader = restore })

	probes := 0
	openPendingReader = func(string) pendingReader {
		probes++
		if source == nil {
			return nil
		}
		return source
	}
	return &probes
}

// withFakeTimer replaces the recheck timer and returns the one the watch made,
// or nil when it made none.
func withFakeTimer(t *testing.T) **fakeReadTimer {
	t.Helper()

	restore := newReadTimer
	t.Cleanup(func() { newReadTimer = restore })

	made := new(*fakeReadTimer)
	newReadTimer = func(d time.Duration, f func()) readTimer {
		timer := &fakeReadTimer{after: d, fire: f}
		*made = timer
		return timer
	}
	return made
}

// acceptWrites makes every write succeed, as every write does under
// keep_latest, and records what was written.
func acceptWrites(t *testing.T) *[][]byte {
	t.Helper()

	restore := writeReport
	t.Cleanup(func() { writeReport = restore })

	var written [][]byte
	writeReport = func(_ string, _ *os.File, data []byte, _ time.Duration) error {
		written = append(written, append([]byte(nil), data...))
		return nil
	}
	return &written
}

// Slot A's vendor kernel has no pending_ms. The write timeout is then the only
// signal, exactly as before: a timed-out write is a stall and closes the
// handle, a successful one is accepting, and nothing else runs per report.
func TestWithoutPendingMsTheWriteTimeoutIsTheSignal(t *testing.T) {
	probes := withPending(t, nil)
	timer := withFakeTimer(t)

	h := &Hid{}
	device := h.keyboardDevice(filepath.Join(t.TempDir(), "hidg0"))

	acceptWrites(t)
	openDevice(t, device)
	for i := 0; i < 5; i++ {
		if err := h.writeHID(device, make([]byte, KeyboardReportLen)); err != nil {
			t.Fatalf("write %d: %s", i, err)
		}
	}
	if got := h.kbHealth.snapshot(time.Now()).State; got != hidStateAccepting {
		t.Fatalf("state = %q, want %q", got, hidStateAccepting)
	}
	if *probes != 1 {
		t.Fatalf("the kernel was probed %d times for one handle, want 1", *probes)
	}
	if *timer != nil {
		t.Fatal("a recheck was armed without pending_ms")
	}

	stallWrites(t)
	err := h.writeHID(device, make([]byte, KeyboardReportLen))
	if !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("err = %v, want a deadline", err)
	}
	if errors.Is(err, errHostNotReading) {
		t.Fatal("the stall came from pending_ms, which this kernel does not have")
	}
	if got := h.kbHealth.snapshot(time.Now()).State; got != hidStateStalled {
		t.Fatalf("state = %q, want %q", got, hidStateStalled)
	}
	if device.get() != nil {
		t.Fatal("the timed-out write left the handle open, unlike before")
	}
}

// The case this exists for: under keep_latest the write succeeds while the
// host has left the report before it unread, and pending_ms says so.
func TestAReportLeftUnreadFailsTheWriteAsAStall(t *testing.T) {
	source := &fakePending{pending: 150 * time.Millisecond}
	withPending(t, source)
	withFakeTimer(t)
	written := acceptWrites(t)
	hook := test.NewGlobal()
	defer hook.Reset()

	h := &Hid{}
	path := filepath.Join(t.TempDir(), "hidg0")
	device := h.keyboardDevice(path)
	openDevice(t, device)

	for i := 0; i < 20; i++ {
		err := h.writeHID(device, make([]byte, KeyboardReportLen))
		if !errors.Is(err, os.ErrDeadlineExceeded) {
			t.Fatalf("write %d: err = %v, want a stall", i, err)
		}
	}

	if len(*written) != 20 {
		t.Fatalf("%d reports reached the kernel, want 20: keep_latest keeps the newest", len(*written))
	}
	got := h.kbHealth.snapshot(time.Now())
	if got.State != hidStateStalled {
		t.Fatalf("state = %q, want %q", got.State, hidStateStalled)
	}
	if device.get() == nil {
		t.Fatal("the handle was closed, but nothing is wrong with it")
	}
	if n := countEntries(hook, log.ErrorLevel, "not fetching reports"); n != 1 {
		t.Fatalf("the stall warning was logged %d times, want once", n)
	}
	if h.kbHealth.stalledWrites.Load() != 20 {
		t.Fatalf("stalled writes = %d, want 20", h.kbHealth.stalledWrites.Load())
	}
}

func TestAHostThatReadsIsAcceptingAndArmsTheRecheck(t *testing.T) {
	source := &fakePending{pending: 3 * time.Millisecond}
	withPending(t, source)
	timer := withFakeTimer(t)
	acceptWrites(t)

	h := &Hid{}
	device := h.keyboardDevice(filepath.Join(t.TempDir(), "hidg0"))
	openDevice(t, device)

	if err := h.writeHID(device, make([]byte, KeyboardReportLen)); err != nil {
		t.Fatalf("write: %s", err)
	}
	if got := h.kbHealth.snapshot(time.Now()).State; got != hidStateAccepting {
		t.Fatalf("state = %q, want %q", got, hidStateAccepting)
	}
	if *timer == nil {
		t.Fatal("no recheck armed")
	}
	if want := hostReadStallAfter - 3*time.Millisecond; (*timer).after != want {
		t.Fatalf("recheck in %s, want %s", (*timer).after, want)
	}

	// The next write resets the same timer rather than making another.
	first := *timer
	if err := h.writeHID(device, make([]byte, KeyboardReportLen)); err != nil {
		t.Fatalf("write: %s", err)
	}
	if *timer != first || first.resets != 1 {
		t.Fatalf("timer made again or not reset (resets %d)", first.resets)
	}
}

// A host that stops reading after the last report is reported without
// anything more being typed, as it would be once the next write timed out.
func TestTheRecheckReportsAStallWithoutAnotherWrite(t *testing.T) {
	source := &fakePending{pending: 2 * time.Millisecond}
	withPending(t, source)
	timer := withFakeTimer(t)
	written := acceptWrites(t)
	hook := test.NewGlobal()
	defer hook.Reset()

	h := &Hid{}
	device := h.keyboardDevice(filepath.Join(t.TempDir(), "hidg0"))
	openDevice(t, device)

	press := []byte{0x02, 0, 0, 0, 0, 0, 0, 0}
	if err := h.writeHID(device, press); err != nil {
		t.Fatalf("write: %s", err)
	}

	source.set(120 * time.Millisecond)
	(*timer).fire()

	got := h.kbHealth.snapshot(time.Now())
	if got.State != hidStateStalled || !got.WasAccepting {
		t.Fatalf("state %q, wasAccepting %t; want stalled after accepting", got.State, got.WasAccepting)
	}
	if n := countEntries(hook, log.ErrorLevel, "not fetching reports"); n != 1 {
		t.Fatalf("the stall warning was logged %d times, want once", n)
	}
	// The recheck reports; it does not release. A key held with nothing more
	// typed stays held, as on the vendor kernel.
	if len(*written) != 1 {
		t.Fatalf("the recheck wrote %d reports, want none", len(*written)-1)
	}
}

func TestTheRecheckLeavesAHostThatReadAlone(t *testing.T) {
	source := &fakePending{pending: 2 * time.Millisecond}
	withPending(t, source)
	timer := withFakeTimer(t)
	acceptWrites(t)

	h := &Hid{}
	device := h.keyboardDevice(filepath.Join(t.TempDir(), "hidg0"))
	openDevice(t, device)
	if err := h.writeHID(device, make([]byte, KeyboardReportLen)); err != nil {
		t.Fatalf("write: %s", err)
	}

	source.set(0)
	resets := (*timer).resets
	(*timer).fire()

	if got := h.kbHealth.snapshot(time.Now()).State; got != hidStateAccepting {
		t.Fatalf("state = %q, want %q", got, hidStateAccepting)
	}
	if (*timer).resets != resets {
		t.Fatal("re-armed although nothing was waiting")
	}
}

// The kernel counts in jiffies, so a report may read a tick short of the limit
// when the timer fires. It is checked again rather than given up on.
func TestTheRecheckWaitsForAReportStillUnderTheLimit(t *testing.T) {
	source := &fakePending{pending: 2 * time.Millisecond}
	withPending(t, source)
	timer := withFakeTimer(t)
	acceptWrites(t)

	h := &Hid{}
	device := h.keyboardDevice(filepath.Join(t.TempDir(), "hidg0"))
	openDevice(t, device)
	if err := h.writeHID(device, make([]byte, KeyboardReportLen)); err != nil {
		t.Fatalf("write: %s", err)
	}

	source.set(96 * time.Millisecond)
	(*timer).fire()
	if got := h.kbHealth.snapshot(time.Now()).State; got != hidStateAccepting {
		t.Fatalf("state = %q, want %q", got, hidStateAccepting)
	}
	if (*timer).after != hostReadRecheckMin {
		t.Fatalf("re-armed for %s, want %s", (*timer).after, hostReadRecheckMin)
	}

	source.set(106 * time.Millisecond)
	(*timer).fire()
	if got := h.kbHealth.snapshot(time.Now()).State; got != hidStateStalled {
		t.Fatalf("state = %q, want %q", got, hidStateStalled)
	}
}

func TestAClosedHandleIgnoresALateRecheck(t *testing.T) {
	source := &fakePending{pending: 2 * time.Millisecond}
	withPending(t, source)
	timer := withFakeTimer(t)
	acceptWrites(t)

	h := &Hid{}
	device := h.keyboardDevice(filepath.Join(t.TempDir(), "hidg0"))
	openDevice(t, device)
	if err := h.writeHID(device, make([]byte, KeyboardReportLen)); err != nil {
		t.Fatalf("write: %s", err)
	}

	h.kbMutex.Lock()
	h.closeDeviceNoLock(device)
	h.kbMutex.Unlock()
	if !(*timer).stopped || !source.closed {
		t.Fatalf("closing left the timer (stopped %t) or the source (closed %t)", (*timer).stopped, source.closed)
	}

	source.set(500 * time.Millisecond)
	(*timer).fire()
	if got := h.kbHealth.snapshot(time.Now()).State; got != hidStateAccepting {
		t.Fatalf("state = %q, want %q", got, hidStateAccepting)
	}
}

// A host that reads again is reported as such by the next write, as on the
// vendor kernel.
func TestAHostThatReadsAgainIsAcceptingAgain(t *testing.T) {
	source := &fakePending{pending: 300 * time.Millisecond}
	withPending(t, source)
	withFakeTimer(t)
	acceptWrites(t)
	hook := test.NewGlobal()
	defer hook.Reset()

	h := &Hid{}
	device := h.keyboardDevice(filepath.Join(t.TempDir(), "hidg0"))
	openDevice(t, device)
	_ = h.writeHID(device, make([]byte, KeyboardReportLen))

	source.set(1 * time.Millisecond)
	if err := h.writeHID(device, make([]byte, KeyboardReportLen)); err != nil {
		t.Fatalf("write after the host read again: %s", err)
	}
	if got := h.kbHealth.snapshot(time.Now()).State; got != hidStateAccepting {
		t.Fatalf("state = %q, want %q", got, hidStateAccepting)
	}
	if n := countEntries(hook, log.InfoLevel, "accepting reports again"); n != 1 {
		t.Fatalf("the recovery was logged %d times, want once", n)
	}
}

// Release on resume. The keyboard worker reacts to the stall as it does to a
// timed-out write: it drops the queued reports and writes a release. Under
// keep_latest the release reaches the kernel as the newest report, which is
// the one a host that reads again ends on.
func TestAStalledKeyboardEndsOnARelease(t *testing.T) {
	source := &fakePending{pending: 150 * time.Millisecond}
	withPending(t, source)
	withFakeTimer(t)
	written := acceptWrites(t)

	h := &Hid{}
	path := filepath.Join(t.TempDir(), "hidg0")
	openDevice(t, h.keyboardDevice(path))

	press := []byte{0x02, 0, 0x04, 0, 0, 0, 0, 0}
	queue := make(chan QueuedReport, 1)
	completed := make(chan bool, 1)
	queue <- QueuedReport{Data: press, Complete: func(ok bool) { completed <- ok }}
	close(queue)

	done := make(chan struct{})
	go func() {
		h.keyboardReports(queue, path)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("keyboard worker did not stop")
	}

	if ok := <-completed; ok {
		t.Fatal("the press completed as delivered while the host was not reading")
	}
	if len(*written) < 2 {
		t.Fatalf("wrote %d reports, want the press and a release", len(*written))
	}
	if !bytes.Equal((*written)[0], press) {
		t.Fatalf("first report %v, want the press", (*written)[0])
	}
	if last := (*written)[len(*written)-1]; !bytes.Equal(last, keyboardReleaseReport()) {
		t.Fatalf("the kernel's newest report is %v, want the release", last)
	}
	if got := h.kbHealth.snapshot(time.Now()).State; got != hidStateStalled {
		t.Fatalf("state = %q, want %q", got, hidStateStalled)
	}
}

// A read that fails is no signal at all, not a stall. The source is dropped
// and the next write asks the kernel again, which is what a gadget rebuilt
// under an open handle needs.
func TestAFailedReadIsNotAStallAndProbesAgain(t *testing.T) {
	source := &fakePending{err: errors.New("no such device")}
	probes := withPending(t, source)
	withFakeTimer(t)
	acceptWrites(t)

	h := &Hid{}
	device := h.keyboardDevice(filepath.Join(t.TempDir(), "hidg0"))
	openDevice(t, device)

	for i := 0; i < 2; i++ {
		if err := h.writeHID(device, make([]byte, KeyboardReportLen)); err != nil {
			t.Fatalf("write %d: %s", i, err)
		}
	}
	if !source.closed {
		t.Fatal("the failed source was not closed")
	}
	if *probes != 2 {
		t.Fatalf("probed %d times, want once per write after a failure", *probes)
	}
}

// Endpoints are separate: a keyboard the host stopped reading says nothing
// about the relative mouse.
func TestEachEndpointHasItsOwnSource(t *testing.T) {
	restore := openPendingReader
	t.Cleanup(func() { openPendingReader = restore })
	sources := map[string]*fakePending{
		"hidg0": {pending: 200 * time.Millisecond},
		"hidg1": {pending: 1 * time.Millisecond},
	}
	openPendingReader = func(path string) pendingReader {
		if source, ok := sources[filepath.Base(path)]; ok {
			return source
		}
		return nil
	}
	withFakeTimer(t)
	acceptWrites(t)

	dir := t.TempDir()
	h := &Hid{}
	keyboard := h.keyboardDevice(filepath.Join(dir, "hidg0"))
	mouse := h.relativeMouseDevice(filepath.Join(dir, "hidg1"))
	openDevice(t, keyboard)
	openDevice(t, mouse)

	if err := h.writeHID(keyboard, make([]byte, KeyboardReportLen)); !errors.Is(err, errHostNotReading) {
		t.Fatalf("keyboard: err = %v, want a stall", err)
	}
	if err := h.writeHID(mouse, make([]byte, RelativeMouseReportLen)); err != nil {
		t.Fatalf("mouse: %s", err)
	}
	if got := h.relHealth.snapshot(time.Now()).State; got != hidStateAccepting {
		t.Fatalf("mouse state = %q, want %q", got, hidStateAccepting)
	}
}

func TestTheSysfsSourceReadsTheKernelsValue(t *testing.T) {
	dir := t.TempDir()
	restore := hidgClassDir
	t.Cleanup(func() { hidgClassDir = restore })
	hidgClassDir = dir

	if reader := openPendingReader("/dev/hidg0"); reader != nil {
		t.Fatal("a kernel without the file gave a source")
	}

	attr := filepath.Join(dir, "hidg0", "pending_ms")
	if err := os.MkdirAll(filepath.Dir(attr), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(attr, []byte("0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	reader := openPendingReader("/dev/hidg0")
	if reader == nil {
		t.Fatal("no source for /dev/hidg0")
	}
	defer reader.Close()

	got, err := reader.Pending()
	if err != nil || got != 0 {
		t.Fatalf("Pending() = %s, %v; want 0", got, err)
	}

	// sysfs prints the value afresh on each read from offset 0.
	if err := os.WriteFile(attr, []byte("4294\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err = reader.Pending()
	if err != nil || got != 4294*time.Millisecond {
		t.Fatalf("Pending() = %s, %v; want 4.294s", got, err)
	}
}

func TestParseMillis(t *testing.T) {
	cases := []struct {
		in   string
		want uint64
		ok   bool
	}{
		{"0\n", 0, true},
		{"17\n", 17, true},
		{"4294967295\n", 4294967295, true},
		{"12", 12, true},
		{"\n", 0, false},
		{"", 0, false},
		{"-1\n", 0, false},
		{"1a\n", 0, false},
		{"99999999999999999999\n", 0, false},
	}
	for _, c := range cases {
		got, err := parseMillis([]byte(c.in))
		if (err == nil) != c.ok || got != c.want {
			t.Errorf("parseMillis(%q) = %d, %v; want %d, ok %t", c.in, got, err, c.want, c.ok)
		}
	}
}

// The check runs on every report, so with the source open it must not add an
// allocation to the write.
func TestTheCheckDoesNotAllocatePerReport(t *testing.T) {
	dir := t.TempDir()
	restore := hidgClassDir
	t.Cleanup(func() { hidgClassDir = restore })
	hidgClassDir = dir
	attr := filepath.Join(dir, "hidg0", "pending_ms")
	if err := os.MkdirAll(filepath.Dir(attr), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(attr, []byte("1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	withFakeTimer(t)

	h := &Hid{}
	device := h.keyboardDevice(HID0)
	if err := h.checkHostReadingNoLock(device); err != nil {
		t.Fatalf("first check: %s", err)
	}
	t.Cleanup(func() { h.kbRead.forgetNoLock() })

	allocs := testing.AllocsPerRun(200, func() {
		if err := h.checkHostReadingNoLock(device); err != nil {
			t.Fatalf("check: %s", err)
		}
	})
	if allocs > 0 {
		t.Fatalf("the check allocates %.0f times per report, want 0", allocs)
	}
}
