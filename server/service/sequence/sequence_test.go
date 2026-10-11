package sequence

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"NanoKVM-Server/proto"
)

// fakeOutput records what a run sends, in order.
type fakeOutput struct {
	mu      sync.Mutex
	events  []string
	reports [][]byte
	closed  atomic.Bool
	fail    error
}

func (o *fakeOutput) report(ctx context.Context, kind stepKind, report []byte) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.fail != nil {
		return o.fail
	}
	o.events = append(o.events, "report")
	o.reports = append(o.reports, append([]byte(nil), report...))
	return nil
}

func (o *fakeOutput) press(button string, hold time.Duration) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.events = append(o.events, "atx:"+button)
	return nil
}

func (o *fakeOutput) close() {
	o.closed.Store(true)
}

func (o *fakeOutput) snapshot() ([]string, [][]byte) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return append([]string(nil), o.events...), append([][]byte(nil), o.reports...)
}

func newTestManager(out *fakeOutput) *Manager {
	m := newManager(func() output { return out })
	m.videoMode = func() videoMode { return videoMode{1920, 1080} }
	m.videoPoll = 5 * time.Millisecond
	return m
}

// waitDone waits for the run to end and returns its status.
func waitDone(t *testing.T, m *Manager) proto.SequenceStatusRsp {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if status := m.Status(); !status.Running {
			return status
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatal("the run did not end")
	return proto.SequenceStatusRsp{}
}

// waitTaps waits until the run has pressed the key at least n times.
func waitTaps(t *testing.T, m *Manager, n int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if m.Status().Taps >= n {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("fewer than %d taps", n)
}

func tapProgram(t *testing.T, req proto.AutoTapReq) program {
	t.Helper()
	p, err := autoTapProgram(req)
	if err != nil {
		t.Fatalf("autoTapProgram: %v", err)
	}
	return p
}

// checkTaps holds that every press is followed by a release, and that the run
// sent nothing else.
func checkTaps(t *testing.T, reports [][]byte, down []byte) {
	t.Helper()
	up := make([]byte, 8)
	if len(reports) == 0 {
		t.Fatal("no reports")
	}
	for i, r := range reports {
		want := down
		if i%2 == 1 {
			want = up
		}
		if !bytes.Equal(r, want) {
			t.Fatalf("report %d = %v, want %v", i, r, want)
		}
	}
}

func TestAutoTapPressesAndReleasesUntilTheTimeout(t *testing.T) {
	out := &fakeOutput{}
	m := newTestManager(out)

	p := tapProgram(t, proto.AutoTapReq{Keys: []int{0x3b}, IntervalMs: 50})
	p.timeout = 300 * time.Millisecond
	if _, err := m.start(p); err != nil {
		t.Fatalf("start: %v", err)
	}

	status := waitDone(t, m)
	if status.Reason != reasonTimeout {
		t.Fatalf("reason = %q, want %q", status.Reason, reasonTimeout)
	}
	if status.Taps < 2 {
		t.Fatalf("taps = %d, want at least 2 in 300 ms at 50 ms", status.Taps)
	}
	if !out.closed.Load() {
		t.Fatal("the output was not closed, so nothing released what the run held")
	}
	_, reports := out.snapshot()
	checkTaps(t, reports, []byte{0, 0, 0x3b, 0, 0, 0, 0, 0})
}

func TestStopEndsTheRunAndReleases(t *testing.T) {
	out := &fakeOutput{}
	m := newTestManager(out)

	if _, err := m.start(tapProgram(t, proto.AutoTapReq{Keys: []int{0x4c}, IntervalMs: 50})); err != nil {
		t.Fatalf("start: %v", err)
	}
	waitTaps(t, m, 1)

	status, err := m.Stop()
	if err != nil {
		t.Fatalf("stop: %v", err)
	}
	if status.Running || status.Reason != reasonStopped {
		t.Fatalf("status = %+v, want stopped", status)
	}
	if !out.closed.Load() {
		t.Fatal("the output was not closed")
	}
	if _, err := m.Stop(); !errors.Is(err, errNotRunning) {
		t.Fatalf("second stop = %v, want errNotRunning", err)
	}
}

func TestUserInputStopsOnlyARunThatAskedForIt(t *testing.T) {
	out := &fakeOutput{}
	m := newTestManager(out)

	if _, err := m.start(tapProgram(t, proto.AutoTapReq{Keys: []int{0x3b}, IntervalMs: 50})); err != nil {
		t.Fatalf("start: %v", err)
	}
	waitTaps(t, m, 1)
	m.noteUserInput()
	time.Sleep(20 * time.Millisecond)
	if !m.Status().Running {
		t.Fatal("user input stopped a run that did not ask to stop on it")
	}
	_, _ = m.Stop()

	out = &fakeOutput{}
	m = newTestManager(out)
	if _, err := m.start(tapProgram(t, proto.AutoTapReq{Keys: []int{0x3b}, IntervalMs: 50, StopOnInput: true})); err != nil {
		t.Fatalf("start: %v", err)
	}
	waitTaps(t, m, 1)
	m.noteUserInput()
	if status := waitDone(t, m); status.Reason != reasonInput {
		t.Fatalf("reason = %q, want %q", status.Reason, reasonInput)
	}
	if m.listening.Load() {
		t.Fatal("the manager still listens for user input after the run")
	}
}

func TestAVideoModeChangeStopsTheRun(t *testing.T) {
	out := &fakeOutput{}
	m := newTestManager(out)

	var mode atomic.Value
	mode.Store(videoMode{1920, 1080})
	m.videoMode = func() videoMode { return mode.Load().(videoMode) }

	if _, err := m.start(tapProgram(t, proto.AutoTapReq{Keys: []int{0x3b}, IntervalMs: 50, StopOnVideo: true})); err != nil {
		t.Fatalf("start: %v", err)
	}
	waitTaps(t, m, 1)

	// No mode published is not a change.
	mode.Store(videoMode{0, 0})
	time.Sleep(30 * time.Millisecond)
	if !m.Status().Running {
		t.Fatal("a zero mode stopped the run")
	}

	mode.Store(videoMode{640, 480})
	if status := waitDone(t, m); status.Reason != reasonVideo {
		t.Fatalf("reason = %q, want %q", status.Reason, reasonVideo)
	}
}

func TestMacroPlaysEveryStepInOrder(t *testing.T) {
	out := &fakeOutput{}
	m := newTestManager(out)

	steps, err := compileSteps([]proto.SequenceStep{
		{Type: StepATX, Button: ButtonPower, Ms: 100},
		{Type: StepKey, Modifiers: 0x01, Keys: []int{0x06}},
		{Type: StepDelay, Ms: 5},
		{Type: StepKey},
		{Type: StepMouse, Absolute: true, Buttons: 1, X: 0x1234, Y: 0x0102},
		{Type: StepMouse, X: -1, Y: 2, Wheel: -1},
	})
	if err != nil {
		t.Fatalf("compileSteps: %v", err)
	}
	if _, err := m.start(program{kind: KindMacro, name: "copy", steps: steps}); err != nil {
		t.Fatalf("start: %v", err)
	}

	status := waitDone(t, m)
	if status.Reason != reasonFinished || status.Step != 6 || status.Total != 6 || status.Name != "copy" {
		t.Fatalf("status = %+v", status)
	}
	events, reports := out.snapshot()
	wantEvents := []string{"atx:power", "report", "report", "report", "report"}
	if len(events) != len(wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	for i := range events {
		if events[i] != wantEvents[i] {
			t.Fatalf("events = %v, want %v", events, wantEvents)
		}
	}
	wantReports := [][]byte{
		{0x01, 0, 0x06, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{1, 0x34, 0x12, 0x02, 0x01, 0},
		{0, 0xff, 2, 0xff, 0},
	}
	for i := range wantReports {
		if !bytes.Equal(reports[i], wantReports[i]) {
			t.Fatalf("report %d = %v, want %v", i, reports[i], wantReports[i])
		}
	}
	if !out.closed.Load() {
		t.Fatal("the output was not closed")
	}
}

func TestOneRunAtATime(t *testing.T) {
	m := newTestManager(&fakeOutput{})

	if _, err := m.start(tapProgram(t, proto.AutoTapReq{Keys: []int{0x3b}, IntervalMs: 50})); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer m.Stop()

	status, err := m.start(tapProgram(t, proto.AutoTapReq{Keys: []int{0x4c}, IntervalMs: 50}))
	if !errors.Is(err, errRunning) {
		t.Fatalf("second start = %v, want errRunning", err)
	}
	if !status.Running || status.Kind != KindAutoTap {
		t.Fatalf("the refusal should report the run in progress, got %+v", status)
	}
}

func TestAFailedWriteEndsTheRun(t *testing.T) {
	out := &fakeOutput{fail: errHID}
	m := newTestManager(out)

	if _, err := m.start(tapProgram(t, proto.AutoTapReq{Keys: []int{0x3b}, IntervalMs: 50})); err != nil {
		t.Fatalf("start: %v", err)
	}
	if status := waitDone(t, m); status.Reason != reasonHID {
		t.Fatalf("reason = %q, want %q", status.Reason, reasonHID)
	}
	if !out.closed.Load() {
		t.Fatal("the output was not closed after the failure")
	}
}

func TestShutdownStopsTheRunAndRefusesNewOnes(t *testing.T) {
	out := &fakeOutput{}
	m := newTestManager(out)

	if _, err := m.start(tapProgram(t, proto.AutoTapReq{Keys: []int{0x3b}, IntervalMs: 50})); err != nil {
		t.Fatalf("start: %v", err)
	}
	waitTaps(t, m, 1)
	m.shutdown(time.Second)

	if status := m.Status(); status.Running || status.Reason != reasonShutdown {
		t.Fatalf("status = %+v, want ended by shutdown", status)
	}
	if !out.closed.Load() {
		t.Fatal("the output was not closed at shutdown")
	}
	if _, err := m.start(tapProgram(t, proto.AutoTapReq{Keys: []int{0x3b}})); !errors.Is(err, errShutdown) {
		t.Fatalf("start after shutdown = %v, want errShutdown", err)
	}
}

func TestReasonForWrappedErrors(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{nil, reasonFinished},
		{errStopped, reasonStopped},
		{errTimeout, reasonTimeout},
		{errInput, reasonInput},
		{errVideo, reasonVideo},
		{errShutdown, reasonShutdown},
		{fmt.Errorf("%w: control lease", errBusy), reasonBusy},
		{fmt.Errorf("%w: gpio", errATX), reasonATX},
		{errHID, reasonHID},
		{errors.New("something else"), reasonHID},
	}
	for _, c := range cases {
		if got := reasonFor(c.err); got != c.want {
			t.Errorf("reasonFor(%v) = %q, want %q", c.err, got, c.want)
		}
	}
}
