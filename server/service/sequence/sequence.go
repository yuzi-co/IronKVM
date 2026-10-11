// Package sequence plays timed input sequences on the host: the key auto-tap
// (repeat F2 or Del until the firmware menu appears) and recorded macros. Both
// are a list of steps, key and pointer reports, pauses and ATX presses, and
// one engine plays both.
//
// The server plays them, not the browser. A browser tab that is closed,
// frozen in the background or reloaded would otherwise stop a macro halfway,
// and an auto-tap exists for the reboot that follows a power press, which is
// when an operator switches tabs. Timing is also steadier here: a background
// tab's timers are throttled to once a second.
//
// One sequence runs at a time. A run ends when its steps end, when it is
// stopped, at its timeout, and optionally at the first key or click from a
// browser or when the video mode changes. Whatever it holds then is released,
// see hidOutput.
package sequence

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
)

const (
	// maxRunTime bounds every run, whatever its own timeout says, so a run
	// nobody stops cannot type for ever.
	maxRunTime = 30 * time.Minute

	// stopWait bounds how long a stop request waits for the run to end. A
	// run checks between steps, and the longest step that does not check is
	// an ATX press, so the wait can run out; the status then still says
	// running, and the run ends after the press.
	stopWait = 3 * time.Second

	// shutdownWait bounds the wait at server shutdown.
	shutdownWait = time.Second
)

const (
	KindAutoTap = "autotap"
	KindMacro   = "macro"
)

// The reasons a run ends, as SequenceStatusRsp.Reason reports them.
const (
	reasonFinished = "finished"
	reasonStopped  = "stopped"
	reasonTimeout  = "timeout"
	reasonInput    = "input"
	reasonVideo    = "video"
	reasonShutdown = "shutdown"
	reasonBusy     = "busy"
	reasonHID      = "hid_error"
	reasonATX      = "atx_error"
)

var (
	errStopped  = errors.New("stopped")
	errTimeout  = errors.New("timed out")
	errInput    = errors.New("stopped by user input")
	errVideo    = errors.New("the video mode changed")
	errShutdown = errors.New("the server is shutting down")
	errBusy     = errors.New("the keyboard is busy")
	errHID      = errors.New("HID write failed")
	errATX      = errors.New("ATX press failed")

	errRunning    = errors.New("a sequence is already running")
	errNotRunning = errors.New("no sequence is running")
	errStopWait   = errors.New("the sequence did not stop in time")
)

type stepKind uint8

const (
	stepKeyboard stepKind = iota
	stepRelative
	stepAbsolute
	stepDelay
	stepATX
)

// step is one step of a run, already turned into what the endpoint takes.
type step struct {
	kind   stepKind
	report []byte
	// wait is the pause of a delay, or how long an ATX button is held.
	wait   time.Duration
	button string
}

// program is what one run plays: steps once, then loop until a stop
// condition. A macro has no loop; an auto-tap is all loop, after an optional
// ATX press.
type program struct {
	kind        string
	name        string
	steps       []step
	loop        []step
	timeout     time.Duration
	stopOnInput bool
	stopOnVideo bool
}

// Manager runs one sequence at a time and keeps its progress for the status
// and stop requests.
type Manager struct {
	mu      sync.Mutex
	cancel  context.CancelCauseFunc
	done    chan struct{}
	started time.Time
	status  proto.SequenceStatusRsp
	closed  bool

	// listening is set while the running sequence stops on user input, so
	// the websocket's per-event call costs one atomic load otherwise.
	listening atomic.Bool

	newOutput func() output
	videoMode func() videoMode
	videoPoll time.Duration
	now       func() time.Time
}

func newManager(newOutput func() output) *Manager {
	return &Manager{
		newOutput: newOutput,
		videoMode: readVideoMode,
		videoPoll: defaultVideoPoll,
		now:       time.Now,
	}
}

var (
	defaultManager     *Manager
	defaultManagerOnce sync.Once
	pressButton        atomic.Value // func(string, time.Duration) error
)

// SetATX gives the engine the call that presses a front-panel button. The
// router passes vm.PressButton, the call behind POST /api/vm/gpio, so a
// macro's ATX step and the power menu share one lock and one set of limits.
func SetATX(press func(string, time.Duration) error) {
	pressButton.Store(press)
}

// GetManager returns the server's sequence manager.
func GetManager() *Manager {
	defaultManagerOnce.Do(func() {
		defaultManager = newManager(func() output {
			press, _ := pressButton.Load().(func(string, time.Duration) error)
			return newHIDOutput(press)
		})
	})
	return defaultManager
}

// NoteUserInput is called for every key press, click or scroll a browser
// sends. It stops a run that asked to stop on user input.
func NoteUserInput() {
	if defaultManager != nil {
		defaultManager.noteUserInput()
	}
}

// Shutdown stops the running sequence and waits briefly for it to release
// what it holds, so a restart does not leave a key down on the host.
func Shutdown() {
	if defaultManager != nil {
		defaultManager.shutdown(shutdownWait)
	}
}

// start begins playing p and returns once the run is going.
func (m *Manager) start(p program) (proto.SequenceStatusRsp, error) {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return m.Status(), errShutdown
	}
	if m.done != nil {
		m.mu.Unlock()
		return m.Status(), errRunning
	}

	timeout := p.timeout
	if timeout <= 0 || timeout > maxRunTime {
		timeout = maxRunTime
	}
	ctx, cancel := context.WithCancelCause(context.Background())
	ctx, cancelTimeout := context.WithTimeoutCause(ctx, timeout, errTimeout)
	done := make(chan struct{})
	m.cancel = cancel
	m.done = done
	m.started = m.now()
	m.status = proto.SequenceStatusRsp{
		Running:   true,
		Kind:      p.kind,
		Name:      p.name,
		Total:     len(p.steps),
		TimeoutMs: timeout.Milliseconds(),
	}
	m.listening.Store(p.stopOnInput)
	out := m.newOutput()
	m.mu.Unlock()

	if p.stopOnVideo {
		go m.watchVideo(ctx, cancel)
	}
	go func() {
		defer cancelTimeout()
		m.run(ctx, done, p, out)
	}()
	return m.Status(), nil
}

func (m *Manager) run(ctx context.Context, done chan struct{}, p program, out output) {
	err := m.play(ctx, done, p, out)
	out.close()
	m.finish(done, err)
}

func (m *Manager) play(ctx context.Context, done chan struct{}, p program, out output) error {
	for i, s := range p.steps {
		if err := m.do(ctx, s, out); err != nil {
			return err
		}
		m.update(done, func(st *proto.SequenceStatusRsp) { st.Step = i + 1 })
	}
	if len(p.loop) == 0 {
		return nil
	}
	for {
		for _, s := range p.loop {
			if err := m.do(ctx, s, out); err != nil {
				return err
			}
		}
		m.update(done, func(st *proto.SequenceStatusRsp) { st.Taps++ })
	}
}

func (m *Manager) do(ctx context.Context, s step, out output) error {
	if err := context.Cause(ctx); err != nil {
		return err
	}
	switch s.kind {
	case stepDelay:
		return sleep(ctx, s.wait)
	case stepATX:
		return out.press(s.button, s.wait)
	default:
		return out.report(ctx, s.kind, s.report)
	}
}

func (m *Manager) update(done chan struct{}, change func(*proto.SequenceStatusRsp)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.done == done {
		change(&m.status)
	}
}

func (m *Manager) finish(done chan struct{}, err error) {
	m.mu.Lock()
	if m.done != done {
		m.mu.Unlock()
		return
	}

	m.status.Running = false
	m.status.ElapsedMs = m.now().Sub(m.started).Milliseconds()
	m.status.Reason = reasonFor(err)
	switch m.status.Reason {
	case reasonBusy, reasonHID, reasonATX:
		log.Errorf("%s sequence %q failed: %v", m.status.Kind, m.status.Name, err)
	default:
		log.Debugf("%s sequence %q ended: %s", m.status.Kind, m.status.Name, m.status.Reason)
	}
	m.listening.Store(false)
	m.cancel(nil)
	m.cancel = nil
	m.done = nil
	m.mu.Unlock()

	close(done)
}

func reasonFor(err error) string {
	switch {
	case err == nil:
		return reasonFinished
	case errors.Is(err, errStopped):
		return reasonStopped
	case errors.Is(err, errTimeout):
		return reasonTimeout
	case errors.Is(err, errInput):
		return reasonInput
	case errors.Is(err, errVideo):
		return reasonVideo
	case errors.Is(err, errShutdown):
		return reasonShutdown
	case errors.Is(err, errBusy):
		return reasonBusy
	case errors.Is(err, errATX):
		return reasonATX
	default:
		return reasonHID
	}
}

// Stop ends the running sequence and waits for it to release what it holds.
func (m *Manager) Stop() (proto.SequenceStatusRsp, error) {
	return m.stop(errStopped, stopWait)
}

func (m *Manager) stop(cause error, wait time.Duration) (proto.SequenceStatusRsp, error) {
	m.mu.Lock()
	cancel := m.cancel
	done := m.done
	m.mu.Unlock()

	if cancel == nil || done == nil {
		return m.Status(), errNotRunning
	}

	cancel(cause)
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-done:
		return m.Status(), nil
	case <-timer.C:
		return m.Status(), errStopWait
	}
}

func (m *Manager) noteUserInput() {
	if !m.listening.Load() {
		return
	}
	m.mu.Lock()
	cancel := m.cancel
	m.mu.Unlock()
	if cancel != nil {
		// Only cancel: the caller is the websocket's read loop, and the run
		// releases its keys through the HID workers on its own goroutine.
		cancel(errInput)
	}
}

func (m *Manager) shutdown(wait time.Duration) {
	m.mu.Lock()
	m.closed = true
	m.mu.Unlock()
	_, _ = m.stop(errShutdown, wait)
}

// Status reports the running sequence, or the last one.
func (m *Manager) Status() proto.SequenceStatusRsp {
	m.mu.Lock()
	defer m.mu.Unlock()
	status := m.status
	if status.Running {
		status.ElapsedMs = m.now().Sub(m.started).Milliseconds()
	}
	return status
}

func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-timer.C:
		return nil
	}
}
