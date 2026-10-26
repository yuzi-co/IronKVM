package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"NanoKVM-Server/service/controlmode"

	log "github.com/sirupsen/logrus"
)

// ErrControlModeSwitch is the cause an operation is canceled with when the AI
// control mode switches.
var ErrControlModeSwitch = errors.New("agent operation canceled by control mode switch")

// InputLock is the AI input lock as the Lifecycle needs it: one chat session
// holds it at a time.
type InputLock interface {
	Owner() string
	Release(sessionID string) bool
}

// Sessions is the chat side the Lifecycle reaches: the Bridge.
type Sessions interface {
	Broadcast(event Event)
	CloseAll(closeCode int, reason string) int
	SessionCount() int
}

// LifecycleTiming bounds a start and a stop.
type LifecycleTiming struct {
	// StartSettle is how long to wait after Launch before checking.
	StartSettle time.Duration
	// StartTimeout bounds the wait for the agent to become ready.
	StartTimeout time.Duration
	// ReadyPoll is how often the wait checks.
	ReadyPoll time.Duration
	// StopTimeout bounds the wait for the process to exit.
	StopTimeout time.Duration
	// ProbeInterval is how often the background check runs.
	ProbeInterval time.Duration
	// StatusRefresh is how old a status may be before the status route
	// checks again.
	StatusRefresh time.Duration
}

// LifecycleConfig is what NewLifecycle needs.
type LifecycleConfig struct {
	// ID names the agent in lower case in messages, as in "picoclaw
	// runtime is disabled".
	ID string
	// Title names the agent in messages, as in "PicoClaw does not own
	// device control".
	Title   string
	Runtime Runtime
	Status  *StatusStore
	Intent  *IntentStore
	Control *controlmode.Manager
	Lock    InputLock
	// BeginOperation starts a lifecycle operation that CancelOperations
	// cancels, with ErrControlModeSwitch as the cause.
	BeginOperation func(parent context.Context) (context.Context, func())
	// CancelOperations cancels the agent's running operations, its device
	// actions among them, and returns how many there were.
	CancelOperations func() int
	// ReleaseHID releases any input the agent left held on the host.
	ReleaseHID func() error
	Timing     LifecycleTiming
}

// Lifecycle runs an agent: install, start, stop, restore at boot and the
// status the web UI reads, under the AI control mode. It is agent-neutral;
// the agent's Runtime supplies the specifics.
type Lifecycle struct {
	id, title string
	runtime   Runtime
	status    *StatusStore
	intent    *IntentStore
	control   *controlmode.Manager
	lock      InputLock
	begin     func(parent context.Context) (context.Context, func())
	cancelOps func() int
	releaseHI func() error
	timing    LifecycleTiming

	chatMu sync.RWMutex
	chat   Sessions

	// mu serializes start, stop, restore and configuration changes.
	mu            sync.Mutex
	reconcileOnce sync.Once
	probeOnce     sync.Once
}

// NewLifecycle returns a lifecycle for cfg.Runtime.
func NewLifecycle(cfg LifecycleConfig) *Lifecycle {
	l := &Lifecycle{
		id:        cfg.ID,
		title:     cfg.Title,
		runtime:   cfg.Runtime,
		status:    cfg.Status,
		intent:    cfg.Intent,
		control:   cfg.Control,
		lock:      cfg.Lock,
		begin:     cfg.BeginOperation,
		cancelOps: cfg.CancelOperations,
		releaseHI: cfg.ReleaseHID,
		timing:    cfg.Timing,
		chat:      noSessions{},
	}
	if l.control == nil {
		l.control = controlmode.GetManager()
	}
	if l.begin == nil {
		l.begin = func(parent context.Context) (context.Context, func()) {
			ctx, cancel := context.WithCancelCause(parent)
			return ctx, func() { cancel(context.Canceled) }
		}
	}
	if l.cancelOps == nil {
		l.cancelOps = func() int { return 0 }
	}
	if l.releaseHI == nil {
		l.releaseHI = func() error { return nil }
	}
	return l
}

// SetChat connects the lifecycle to the chat bridge, whose sessions it closes
// when the agent stops and tells about control mode changes.
func (l *Lifecycle) SetChat(chat Sessions) {
	if chat == nil {
		chat = noSessions{}
	}
	l.chatMu.Lock()
	l.chat = chat
	l.chatMu.Unlock()
}

func (l *Lifecycle) sessions() Sessions {
	l.chatMu.RLock()
	defer l.chatMu.RUnlock()
	return l.chat
}

// StatusStore is the store the lifecycle records into.
func (l *Lifecycle) StatusStore() *StatusStore { return l.status }

func (l *Lifecycle) lockLifecycle() func() {
	l.mu.Lock()
	return l.mu.Unlock
}

func (l *Lifecycle) owner() string {
	if l.lock == nil {
		return ""
	}
	return l.lock.Owner()
}

// OperationError returns the error a lifecycle operation answers with once
// ctx is done, and nil before.
func (l *Lifecycle) OperationError(ctx context.Context) *Error {
	if ctx == nil || ctx.Err() == nil {
		return nil
	}

	cause := context.Cause(ctx)
	switch {
	case errors.Is(cause, ErrControlModeSwitch):
		return NewError(CodeControlModeConflict, l.title+" runtime lifecycle canceled because the control mode is switching")
	case errors.Is(cause, context.DeadlineExceeded):
		return NewError(CodeRuntimeUnavailable, l.title+" runtime lifecycle timed out")
	default:
		return NewError(CodeRuntimeUnavailable, l.title+" runtime lifecycle was canceled")
	}
}

// Sleep waits for delay, or less with OperationError's error when ctx ends.
func (l *Lifecycle) Sleep(ctx context.Context, delay time.Duration) *Error {
	if ctx == nil {
		ctx = context.Background()
	}
	if delay <= 0 {
		return l.OperationError(ctx)
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return l.OperationError(ctx)
	case <-timer.C:
		return nil
	}
}

// ensureReady checks the agent and records what it found. A passive check
// (allowLifecycleOverwrite false) leaves a pending start or stop alone and
// does not check a runtime the intent says should not run.
func (l *Lifecycle) ensureReady(ctx context.Context, allowLifecycleOverwrite bool) *Error {
	if lifecycleErr := l.OperationError(ctx); lifecycleErr != nil {
		return lifecycleErr
	}
	setStatus := func(status RuntimeStatus) {
		if allowLifecycleOverwrite {
			l.status.Set(status)
			return
		}
		l.status.SetFromProbe(status)
	}
	currentStatus := l.status.Get()
	if currentStatus.Installing {
		return NewError(CodeRuntimeUnavailable, l.id+" installation is in progress")
	}
	if !allowLifecycleOverwrite && IsLifecyclePending(currentStatus) {
		return NewError(CodeRuntimeUnavailable, l.id+" runtime lifecycle operation is pending")
	}
	if !allowLifecycleOverwrite {
		intent, intentErr := l.intent.Load()
		if intentErr != nil || !intent.DesiredRunning {
			l.applyStoppedStatus()
			if intentErr != nil {
				return NewError(CodeRuntimeUnavailable, intentErr.Error())
			}
			return NewError(CodeRuntimeUnavailable, l.id+" runtime is disabled")
		}
	}

	status, err := l.runtime.Check(ctx)
	if status != nil {
		status.CheckedAt = time.Now()
		status.CurrentSession = l.owner()
		setStatus(*status)
	}
	if err != nil {
		return asError(err, CodeRuntimeUnavailable)
	}
	return nil
}

// Ready checks the agent as the background check does and returns nil when
// it can take a chat session.
func (l *Lifecycle) Ready(ctx context.Context) error {
	if err := l.ensureReady(ctx, false); err != nil {
		return err
	}
	return nil
}

func (l *Lifecycle) startProbeLoop() {
	l.probeOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(l.timing.ProbeInterval)
			defer ticker.Stop()

			for range ticker.C {
				_ = l.ensureReady(context.Background(), false)
			}
		}()
	})
}

func (l *Lifecycle) shouldRefresh(status RuntimeStatus) bool {
	if status.Installing {
		return false
	}
	if status.CheckedAt.IsZero() {
		return true
	}
	return time.Since(status.CheckedAt) >= l.timing.StatusRefresh
}

// start launches the agent and waits until it is ready. An agent that does
// not become ready is stopped again.
func (l *Lifecycle) start(ctx context.Context) (string, string, *Error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if lifecycleErr := l.OperationError(ctx); lifecycleErr != nil {
		return "", "", lifecycleErr
	}

	command, output, launchErr := l.runtime.Launch(ctx)
	if launchErr != nil {
		return command, output, asError(launchErr, CodeRuntimeStartFailed)
	}

	if lifecycleErr := l.Sleep(ctx, l.timing.StartSettle); lifecycleErr != nil {
		return command, output, lifecycleErr
	}
	if runtimeErr := l.waitReady(ctx, l.timing.StartTimeout); runtimeErr != nil {
		startErr := NewError(CodeRuntimeStartFailed, runtimeErr.Message)
		failureStatus := "unavailable"
		if runtimeErr.Code == CodeControlModeConflict {
			startErr = runtimeErr
			failureStatus = "starting"
		} else {
			if cleanupErr := l.stopAndVerify(true); cleanupErr != nil {
				failureStatus = "error"
				startErr.Message = fmt.Sprintf(
					"%s; failed to stop partially started runtime: %v",
					startErr.Message,
					cleanupErr,
				)
			}
		}
		l.status.Update(func(status *RuntimeStatus) {
			status.Ready = false
			status.Status = failureStatus
			status.LastError = startErr.Message
			status.CurrentSession = ""
			status.CheckedAt = time.Now()
		})
		return command, output, startErr
	}

	return command, output, nil
}

func (l *Lifecycle) waitReady(ctx context.Context, timeout time.Duration) *Error {
	if ctx == nil {
		ctx = context.Background()
	}
	deadline := time.Now().Add(timeout)
	var lastErr *Error

	for {
		if lifecycleErr := l.OperationError(ctx); lifecycleErr != nil {
			return lifecycleErr
		}

		runtimeErr := l.ensureReady(ctx, true)
		if runtimeErr == nil {
			return nil
		}
		lastErr = runtimeErr

		if time.Now().After(deadline) {
			break
		}

		if lifecycleErr := l.Sleep(ctx, l.timing.ReadyPoll); lifecycleErr != nil {
			return lifecycleErr
		}
	}

	if lastErr != nil {
		return lastErr
	}
	return NewError(CodeRuntimeUnavailable, "gateway is unavailable")
}

func (l *Lifecycle) markCanceled(err *Error) {
	if err == nil || err.Code != CodeControlModeConflict {
		return
	}
	l.status.Update(func(status *RuntimeStatus) {
		status.Ready = false
		status.Restoring = false
		if IsLifecyclePending(*status) {
			status.Status = "stopped"
		}
		status.LastError = err.Message
		status.CurrentSession = ""
		status.CheckedAt = time.Now()
	})
}

// stopAndVerify stops the agent's process and waits until it is gone. With
// forceStop it halts even a process it does not see running.
func (l *Lifecycle) stopAndVerify(forceStop bool) error {
	err := stopProcessAndVerify(
		l.title,
		forceStop,
		l.runtime.Running,
		func() error {
			_, _, haltErr := l.runtime.Halt()
			return haltErr
		},
		l.timing.StopTimeout,
		100*time.Millisecond,
	)
	if err != nil {
		return err
	}

	l.status.Update(func(status *RuntimeStatus) {
		status.Ready = false
		status.Status = "stopped"
		status.CurrentSession = ""
		status.CheckedAt = time.Now()
	})
	return nil
}

func stopProcessAndVerify(
	title string,
	forceStop bool,
	isRunning func() (bool, error),
	stop func() error,
	timeout time.Duration,
	pollInterval time.Duration,
) error {
	if !forceStop {
		running, err := isRunning()
		if err != nil {
			return fmt.Errorf("check %s runtime: %w", title, err)
		}
		if !running {
			return nil
		}
	}

	var stopErr error
	if err := stop(); err != nil {
		stopErr = fmt.Errorf("stop %s runtime: %w", title, err)
	}

	deadline := time.Now().Add(timeout)
	for {
		running, err := isRunning()
		if err != nil {
			return errors.Join(stopErr, fmt.Errorf("verify %s stopped: %w", title, err))
		}
		if !running {
			if stopErr != nil {
				log.Warnf("%s stop command returned an error after the runtime stopped: %v", title, stopErr)
			}
			return nil
		}
		if !time.Now().Before(deadline) {
			return errors.Join(stopErr, fmt.Errorf("%s runtime is still running", title))
		}
		time.Sleep(pollInterval)
	}
}

// stopAndCloseSessions closes every chat session, releases the AI input lock
// and stops the agent.
func (l *Lifecycle) stopAndCloseSessions(closeCode int, closeReason string) error {
	startedAt := time.Now()
	sessionCount := l.sessions().CloseAll(closeCode, closeReason)
	closeElapsed := time.Since(startedAt)
	if l.lock != nil {
		l.lock.Release("")
	}

	stopStartedAt := time.Now()
	if err := l.stopAndVerify(false); err != nil {
		log.WithFields(log.Fields{
			"session_count":     sessionCount,
			"close_code":        closeCode,
			"close_sessions_ms": closeElapsed.Milliseconds(),
			"stop_runtime_ms":   time.Since(stopStartedAt).Milliseconds(),
			"total_ms":          time.Since(startedAt).Milliseconds(),
		}).WithError(err).Warnf("%s stop runtime and close sessions failed", l.id)
		return err
	}

	log.WithFields(log.Fields{
		"session_count":     sessionCount,
		"close_code":        closeCode,
		"close_sessions_ms": closeElapsed.Milliseconds(),
		"stop_runtime_ms":   time.Since(stopStartedAt).Milliseconds(),
		"total_ms":          time.Since(startedAt).Milliseconds(),
	}).Infof("%s runtime stopped and gateway sessions closed", l.id)
	return nil
}

// ReleaseControlSessions closes every chat session and releases the AI input
// lock. It returns how many sessions it closed.
func (l *Lifecycle) ReleaseControlSessions(closeCode int, closeReason string) int {
	startedAt := time.Now()
	sessionCount := l.sessions().CloseAll(closeCode, closeReason)
	if l.lock != nil {
		l.lock.Release("")
	}
	log.WithFields(log.Fields{
		"session_count": sessionCount,
		"close_code":    closeCode,
		"elapsed_ms":    time.Since(startedAt).Milliseconds(),
	}).Infof("%s gateway sessions closed for control release", l.id)
	return sessionCount
}

// CancelOperations cancels the agent's running operations.
func (l *Lifecycle) CancelOperations() int {
	return l.cancelOps()
}

// PreemptForMCP ends what the agent holds before external MCP takes device
// control: its operations, its chat sessions and the AI input lock. It
// leaves the runtime and its intent alone.
func (l *Lifecycle) PreemptForMCP() error {
	startedAt := time.Now()
	activeOperations := l.CancelOperations()
	closedSessions := l.ReleaseControlSessions(
		CloseControlModeSwitched,
		"external MCP enabled",
	)
	log.WithFields(log.Fields{
		"active_operations": activeOperations,
		"closed_sessions":   closedSessions,
		"elapsed_ms":        time.Since(startedAt).Milliseconds(),
	}).Infof("%s control leases preempted for MCP", l.id)
	return nil
}

// StopForMCP stops the agent for external MCP and records that it should
// not run.
func (l *Lifecycle) StopForMCP() error {
	startedAt := time.Now()

	currentStatus := l.status.Get()
	statusRequiresStop :=
		currentStatus.Ready ||
			currentStatus.Status == "ready" ||
			IsLifecyclePending(currentStatus)
	running, runningErr := l.runtime.Running()
	if runningErr != nil {
		if !statusRequiresStop {
			log.WithFields(log.Fields{
				"status":     currentStatus.Status,
				"ready":      currentStatus.Ready,
				"elapsed_ms": time.Since(startedAt).Milliseconds(),
			}).WithError(runningErr).Warnf("skipping %s runtime stop for MCP because runtime is not active", l.title)
			l.setIntentDesired(false, "mcp_preempt")
			return nil
		}
		return fmt.Errorf("check %s runtime for MCP: %w", l.title, runningErr)
	}
	if !running && !statusRequiresStop {
		l.setIntentDesired(false, "mcp_preempt")
		log.WithFields(log.Fields{
			"status":     currentStatus.Status,
			"ready":      currentStatus.Ready,
			"elapsed_ms": time.Since(startedAt).Milliseconds(),
		}).Infof("%s runtime stop skipped for MCP because runtime is not active", l.title)
		return nil
	}

	unlock := l.lockLifecycle()
	defer unlock()

	l.status.Update(func(status *RuntimeStatus) {
		status.Ready = false
		status.Restoring = false
		status.Status = "stopping"
		status.LastError = ""
		status.CheckedAt = time.Now()
	})
	if err := l.stopAndVerify(false); err != nil {
		log.WithFields(log.Fields{
			"elapsed_ms": time.Since(startedAt).Milliseconds(),
		}).WithError(err).Warnf("%s stop runtime for MCP failed", l.id)
		return fmt.Errorf("stop %s runtime for MCP: %w", l.title, err)
	}
	l.setIntentDesired(false, "mcp_preempt")
	log.WithFields(log.Fields{
		"elapsed_ms": time.Since(startedAt).Milliseconds(),
	}).Infof("%s runtime stopped for MCP", l.id)
	return nil
}

// PreserveForChatOnly keeps a ready agent, or one with chat sessions, set to
// run after the AI control mode was released.
func (l *Lifecycle) PreserveForChatOnly(source string) {
	status := l.status.Get()
	if !status.Ready && l.sessions().SessionCount() == 0 {
		return
	}
	l.setIntentDesired(true, source)
}

// RuntimeStatus is the status with the control mode and the intent applied.
func (l *Lifecycle) RuntimeStatus() RuntimeStatus {
	status := l.status.Get()
	modeStatus, err := l.control.Status()
	if err != nil {
		status.ControlMode = string(controlmode.ModeOff)
		return status
	}
	return l.applyIntentStatus(ApplyControlModeStatus(status, modeStatus))
}

func (l *Lifecycle) applyIntentStatus(status RuntimeStatus) RuntimeStatus {
	intent, err := l.intent.Load()
	status.RuntimeIntent = intent
	if err != nil && status.RuntimeIntent.LastError == "" {
		status.RuntimeIntent.LastError = err.Error()
	}
	status.Restoring = status.Restoring || status.Status == "restoring"
	return l.withRelease(status)
}

// withRelease adds the release an install installs and whether the installed
// agent is older than it.
func (l *Lifecycle) withRelease(status RuntimeStatus) RuntimeStatus {
	status.Version = l.runtime.Version()
	status.UpdateAvailable = status.Installed && !status.Installing && l.runtime.Stale()
	return status
}

// ApplyControlModeStatus adds the AI control mode to status and what it
// allows the agent to do.
func ApplyControlModeStatus(status RuntimeStatus, modeStatus controlmode.Status) RuntimeStatus {
	status.ControlMode = string(modeStatus.Mode)
	status.Transitioning = modeStatus.Transitioning
	canControl := modeStatus.Mode == controlmode.ModePicoclaw && !modeStatus.Transitioning
	runtimeUsable := status.Ready && status.Installed && status.ModelConfigured && !status.Installing
	chat := runtimeUsable && modeStatus.Mode != controlmode.ModeMCP && !modeStatus.Transitioning
	status.Control = ControlStatus{
		Mode:          string(modeStatus.Mode),
		Transitioning: modeStatus.Transitioning,
		CanControl:    canControl,
		LastError:     modeStatus.LastError,
		ChangedAt:     modeStatus.ChangedAt,
	}
	status.Capabilities = RuntimeCapabilities{
		Chat:          chat,
		ReadOnlyTools: chat,
		DeviceWrite:   canControl,
	}
	return status
}

// AcquireControl holds the AI control mode stable while the agent owns
// device control. The release function ends the hold.
func (l *Lifecycle) AcquireControl() (func(), *Error) {
	release, err := l.control.AcquireWrite(controlmode.ModePicoclaw)
	if err != nil {
		return nil, l.ControlError(err)
	}
	return release, nil
}

// requireControl returns nil while the agent owns device control.
func (l *Lifecycle) requireControl() *Error {
	if err := l.control.RequireWrite(controlmode.ModePicoclaw); err != nil {
		return l.ControlError(err)
	}
	return nil
}

// ControlError says why the agent does not own device control.
func (l *Lifecycle) ControlError(err error) *Error {
	status, statusErr := l.control.Status()
	if statusErr != nil {
		return NewError(CodeRuntimeUnavailable, statusErr.Error())
	}

	message := l.title + " does not own device control"
	code := CodeControlRequired
	if status.Transitioning {
		code = CodeControlTransitioning
		message = "device control is switching"
	} else {
		switch status.Mode {
		case controlmode.ModeMCP:
			code = CodeControlOwnedByMCP
			message = "external MCP owns device control"
		case controlmode.ModeOff:
			code = CodeControlRequired
			message = l.title + " device control is not enabled"
		default:
			code = CodeControlModeConflict
		}
	}
	controlErr := NewError(code, message)
	if err != nil && controlErr.Message == "" {
		controlErr.Message = err.Error()
	}
	return controlErr
}

// PublishControlModeChanged tells every open chat session that the AI
// control mode changed.
func (l *Lifecycle) PublishControlModeChanged(status controlmode.Status, source string) {
	l.sessions().Broadcast(ControlModeChangedEvent(status, source))
}

// ControlModeChangedEvent is the chat event for a control mode change.
func ControlModeChangedEvent(status controlmode.Status, source string) Event {
	return Event{
		Type: EventControlModeChanged,
		Control: &ControlMode{
			Mode:          string(status.Mode),
			Transitioning: status.Transitioning,
			CanControl:    status.Mode == controlmode.ModePicoclaw && !status.Transitioning,
			LastError:     status.LastError,
			ChangedAt:     status.ChangedAt,
			Source:        source,
		},
	}
}

// releaseControlToOff switches the AI control mode from the agent to off,
// canceling the agent's operations and releasing held input. It tells the
// chat sessions when it switched.
func (l *Lifecycle) releaseControlToOff(source string) error {
	switched, err := l.control.SwitchIfWithCleanup(
		controlmode.ModePicoclaw,
		controlmode.ModeOff,
		func() error {
			l.CancelOperations()
			return nil
		},
		l.releaseHI,
	)
	if err != nil {
		return err
	}
	if switched {
		if status, statusErr := l.control.Status(); statusErr == nil {
			l.PublishControlModeChanged(status, source)
		}
	}
	return nil
}

func (l *Lifecycle) setIntentDesired(desired bool, updatedBy string) {
	if err := l.intent.SetDesiredRunning(desired, updatedBy); err != nil {
		log.WithFields(log.Fields{
			"desired_running": desired,
			"updated_by":      updatedBy,
		}).WithError(err).Warnf("failed to persist %s runtime intent", l.title)
	}
}

func (l *Lifecycle) setIntentError(message string) {
	if message == "" {
		return
	}
	if err := l.intent.SetLastError(message); err != nil {
		log.WithError(err).Warnf("failed to persist %s runtime intent error", l.title)
	}
}

// applyStoppedStatus records a stopped agent with what is known of its
// install and model.
func (l *Lifecycle) applyStoppedStatus() {
	installed, installedErr := l.runtime.Installed()
	installedKnown := installedErr == nil
	model, modelErr := l.runtime.Model()

	l.status.Update(func(status *RuntimeStatus) {
		status.Ready = false
		status.Restoring = false
		status.CurrentSession = ""
		status.CheckedAt = time.Now()
		if installedKnown {
			status.Installed = installed
		}
		if installedKnown && !installed {
			status.ModelConfigured = false
			status.ModelName = ""
			status.Status = "not_installed"
			status.LastError = ""
			return
		}
		if modelErr == nil {
			status.ModelConfigured = model.Configured
			status.ModelName = model.Name
			if !model.Configured {
				status.ModelName = model.Target
				status.Status = "model_not_configured"
				return
			}
		}
		status.Status = "stopped"
		if status.LastError == l.id+" runtime is stopped" {
			status.LastError = ""
		}
	})
}

// noSessions stands in for the bridge until SetChat is called.
type noSessions struct{}

func (noSessions) Broadcast(Event)          {}
func (noSessions) CloseAll(int, string) int { return 0 }
func (noSessions) SessionCount() int        { return 0 }
