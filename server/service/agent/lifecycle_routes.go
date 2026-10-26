package agent

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// writeErrorWithData answers with err and, when data is set, data beside it.
// It is the envelope the runtime routes always used.
func writeErrorWithData(c *gin.Context, err *Error, data any) {
	if err == nil {
		return
	}

	payload := gin.H{
		"code":    err.Code,
		"message": err.Message,
	}
	if data != nil {
		payload["data"] = data
	}

	statusCode := err.Status
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	c.JSON(statusCode, payload)
}

// Status answers the runtime status route. It checks the agent again when the
// recorded status is old.
func (l *Lifecycle) Status(c *gin.Context) {
	l.startProbeLoop()
	modeStatus, modeErr := l.control.Status()
	if modeErr != nil {
		writeErrorWithData(c, NewError(CodeRuntimeUnavailable, modeErr.Error()), nil)
		return
	}
	status := l.status.Get()
	if l.shouldRefresh(status) {
		_ = l.ensureReady(context.Background(), false)
		status = l.status.Get()
	}

	if installed, err := l.runtime.Installed(); err == nil {
		status.Installed = installed
		l.status.Update(func(current *RuntimeStatus) {
			current.Installed = installed
		})
	}

	status = l.applyIntentStatus(ApplyControlModeStatus(status, modeStatus))
	writeSuccess(c, l.runtime.Describe(status))
}

// Start answers the runtime start route: it starts the agent, while the agent
// owns device control, and records that it should run.
func (l *Lifecycle) Start(c *gin.Context) {
	defer l.recoverRoute(c, "start")
	log.Infof("%s runtime start requested", l.id)

	// Lock order: hold a stable control lease before entering the runtime
	// lifecycle section, so MCP/off transitions wait instead of racing the
	// start and the desired-running commit.
	releaseControl, controlErr := l.AcquireControl()
	if controlErr != nil {
		status := l.RuntimeStatus()
		log.WithFields(log.Fields{
			"code":          controlErr.Code,
			"runtimeStatus": status.Status,
			"ready":         status.Ready,
			"controlMode":   status.ControlMode,
		}).Warnf("%s runtime start rejected by control mode", l.id)
		writeErrorWithData(c, controlErr, gin.H{"status": status})
		return
	}
	defer releaseControl()

	operationCtx, releaseOperation := l.begin(c.Request.Context())
	defer releaseOperation()

	unlock := l.lockLifecycle()
	defer unlock()

	if lifecycleErr := l.OperationError(operationCtx); lifecycleErr != nil {
		writeErrorWithData(c, lifecycleErr, gin.H{"status": l.RuntimeStatus()})
		return
	}

	if currentStatus := l.status.Get(); currentStatus.Installing {
		runtimeErr := NewError(CodeRuntimeUnavailable, l.id+" installation is in progress")
		writeErrorWithData(c, runtimeErr, gin.H{"status": l.RuntimeStatus()})
		return
	}

	if readyErr := l.ensureReady(operationCtx, true); readyErr == nil {
		l.setIntentDesired(true, "web")
		status := l.RuntimeStatus()
		log.WithFields(log.Fields{
			"runtimeStatus": status.Status,
			"ready":         status.Ready,
			"controlMode":   status.ControlMode,
		}).Infof("%s runtime start skipped because runtime is already ready", l.id)
		writeSuccess(c, RuntimeStartResult{
			Started: true,
			Status:  status,
		})
		return
	}

	l.status.Update(func(status *RuntimeStatus) {
		status.Ready = false
		status.Restoring = false
		status.Status = "starting"
		status.LastError = ""
		status.CheckedAt = time.Now()
	})

	command, output, startErr := l.start(operationCtx)
	if startErr != nil {
		l.markCanceled(startErr)
		l.setIntentError(startErr.Message)
		status := l.RuntimeStatus()
		log.WithFields(log.Fields{
			"code":          startErr.Code,
			"runtimeStatus": status.Status,
			"ready":         status.Ready,
			"controlMode":   status.ControlMode,
		}).Warnf("%s runtime start returning structured error", l.id)
		writeErrorWithData(c, startErr, gin.H{
			"command": command,
			"output":  output,
			"status":  status,
		})
		return
	}

	if controlErr := l.requireControl(); controlErr != nil {
		l.setIntentError(controlErr.Message)
		status := l.RuntimeStatus()
		log.WithFields(log.Fields{
			"code":          controlErr.Code,
			"runtimeStatus": status.Status,
			"ready":         status.Ready,
			"controlMode":   status.ControlMode,
		}).Warnf("%s runtime start lost control before intent commit", l.id)
		writeErrorWithData(c, controlErr, gin.H{
			"command": command,
			"output":  output,
			"status":  status,
		})
		return
	}

	l.setIntentDesired(true, "web")
	status := l.RuntimeStatus()
	log.WithFields(log.Fields{
		"runtimeStatus": status.Status,
		"ready":         status.Ready,
		"controlMode":   status.ControlMode,
	}).Infof("%s runtime start returning success", l.id)
	writeSuccess(c, RuntimeStartResult{
		Started: true,
		Command: command,
		Output:  output,
		Status:  status,
	})
}

// Stop answers the runtime stop route: it releases device control, records
// that the agent should not run, closes its chat sessions and stops it.
func (l *Lifecycle) Stop(c *gin.Context) {
	defer l.recoverRoute(c, "stop")
	log.Infof("%s runtime stop requested", l.id)

	if err := l.releaseControlToOff("runtime_stop"); err != nil {
		status := l.RuntimeStatus()
		stopErr := NewError(CodeRuntimeStartFailed, err.Error())
		log.WithFields(log.Fields{
			"code":          stopErr.Code,
			"runtimeStatus": status.Status,
			"ready":         status.Ready,
			"controlMode":   status.ControlMode,
		}).Warnf("%s runtime stop returning structured error", l.id)
		writeErrorWithData(c, stopErr, gin.H{"status": status})
		return
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
	l.setIntentDesired(false, "web")

	err := l.stopAndCloseSessions(CloseRuntimeStopped, l.title+" runtime stopped")
	if err != nil {
		l.setIntentError(err.Error())
		status := l.RuntimeStatus()
		stopErr := NewError(CodeRuntimeStartFailed, err.Error())
		log.WithFields(log.Fields{
			"code":          stopErr.Code,
			"runtimeStatus": status.Status,
			"ready":         status.Ready,
			"controlMode":   status.ControlMode,
		}).Warnf("%s runtime stop returning structured error", l.id)
		writeErrorWithData(c, stopErr, gin.H{"status": status})
		return
	}

	status := l.RuntimeStatus()
	log.WithFields(log.Fields{
		"runtimeStatus": status.Status,
		"ready":         status.Ready,
		"controlMode":   status.ControlMode,
	}).Infof("%s runtime stop returning success", l.id)
	writeSuccess(c, RuntimeStartResult{
		Started: false,
		Status:  status,
	})
}

func (l *Lifecycle) recoverRoute(c *gin.Context, operation string) {
	if recovered := recover(); recovered != nil {
		message := fmt.Sprintf("%s runtime %s panicked: %v", l.id, operation, recovered)
		log.Errorf("%s\n%s", message, debug.Stack())
		if c.Writer.Written() {
			return
		}
		writeErrorWithData(
			c,
			NewError(CodeRuntimeStartFailed, message),
			gin.H{"status": l.safeStatusForRecovery()},
		)
	}
}

func (l *Lifecycle) safeStatusForRecovery() (status RuntimeStatus) {
	defer func() {
		if recover() != nil {
			status = RuntimeStatus{
				Ready:       false,
				Installed:   false,
				InstallPath: l.status.InstallPath(),
				Status:      "error",
			}
		}
	}()
	return l.RuntimeStatus()
}

// Install answers the runtime install route. The install runs in the
// background; the status route follows it.
func (l *Lifecycle) Install(c *gin.Context) {
	output, err := l.install()
	if err != nil {
		writeErrorWithData(c, err, nil)
		return
	}

	currentStatus := l.status.Get()
	writeSuccess(c, RuntimeInstallResult{
		Installed: currentStatus.Installed && !currentStatus.Installing,
		Binary:    l.runtime.InstallPath(),
		Output:    output,
		Status:    l.RuntimeStatus(),
	})
}

func (l *Lifecycle) install() (string, *Error) {
	log.Debugf("%s install: start, binary=%s, release=%s", l.id, l.runtime.InstallPath(), l.runtime.Version())

	currentStatus := l.status.Get()
	if currentStatus.Installing {
		log.Debugf("%s install: install already in progress", l.id)
		return l.id + " installation is already in progress", nil
	}

	installed, err := l.runtime.Installed()
	if err == nil && installed && !l.runtime.Stale() {
		model, _ := l.runtime.Model()
		log.Debugf("%s install: %s is already installed at %s", l.id, l.runtime.Version(), l.runtime.InstallPath())
		l.status.Set(RuntimeStatus{
			Ready:           false,
			Installed:       true,
			Installing:      false,
			InstallProgress: 100,
			InstallStage:    "installed",
			ModelConfigured: model.Configured,
			ModelName:       model.Name,
			Status:          "installed",
			CheckedAt:       time.Now(),
		})
		return l.id + " is already installed", nil
	}
	// An update replaces the binary of a running agent, and the status
	// would then describe a process that is not the installed one.
	if err == nil && installed {
		if running, runErr := l.runtime.Running(); runErr == nil && running {
			return "", NewError(CodeRuntimeStartFailed, "stop "+l.title+" before updating it")
		}
	}

	l.status.Set(RuntimeStatus{
		Ready:           false,
		Installed:       false,
		Installing:      true,
		InstallProgress: 0,
		InstallStage:    "preparing",
		Status:          "installing",
		CheckedAt:       time.Now(),
	})

	go l.runInstall()
	return l.id + " installation started", nil
}

func (l *Lifecycle) runInstall() {
	if err := l.runtime.Install(l.setInstallProgress); err != nil {
		log.Errorf("%s install: %v", l.id, err)
		l.finishInstallFailure("install_failed", err.Error())
		return
	}
	log.Infof("%s install: %s installed", l.id, l.runtime.Version())

	model, _ := l.runtime.Model()
	l.status.Set(RuntimeStatus{
		Ready:           false,
		Installed:       true,
		Installing:      false,
		InstallProgress: 100,
		InstallStage:    "installed",
		ModelConfigured: model.Configured,
		ModelName:       model.Name,
		Status:          "installed",
		CheckedAt:       time.Now(),
	})
}

func (l *Lifecycle) setInstallProgress(stage string, progress int, lastError string) {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	l.status.Update(func(status *RuntimeStatus) {
		status.Ready = false
		status.Installed = false
		status.Installing = true
		status.InstallProgress = progress
		status.InstallStage = stage
		status.Status = "installing"
		status.LastError = lastError
		status.CheckedAt = time.Now()
	})
}

// finishInstallFailure records a failed install. A failed update leaves the
// agent that was installed, so the status says it is still there.
func (l *Lifecycle) finishInstallFailure(status string, message string) {
	installed, _ := l.runtime.Installed()
	l.status.Set(RuntimeStatus{
		Ready:           false,
		Installed:       installed,
		Installing:      false,
		InstallProgress: 0,
		InstallStage:    status,
		Status:          status,
		LastError:       message,
		CheckedAt:       time.Now(),
	})
}

// Uninstall answers the runtime uninstall route: it releases device control,
// stops the agent, records that it should not run and removes it.
func (l *Lifecycle) Uninstall(c *gin.Context) {
	currentStatus := l.status.Get()
	uninstallOutput := l.id + " uninstalled successfully"
	if currentStatus.Installing {
		writeErrorWithData(c, NewError(CodeRuntimeStartFailed, "cannot uninstall while installation is in progress"), nil)
		return
	}

	if err := l.releaseControlToOff("runtime_uninstall"); err != nil {
		status := l.RuntimeStatus()
		writeErrorWithData(c, NewError(CodeRuntimeStartFailed, "control release failed before uninstall: "+err.Error()), gin.H{"status": status})
		return
	}

	unlock := l.lockLifecycle()
	defer unlock()

	currentStatus = l.status.Get()
	if currentStatus.Installing {
		writeErrorWithData(c, NewError(CodeRuntimeStartFailed, "cannot uninstall while installation is in progress"), nil)
		return
	}
	running, err := l.runtime.Running()
	if err != nil {
		writeErrorWithData(c, NewError(CodeRuntimeUnavailable, "failed to check "+l.id+" runtime status"), nil)
		return
	}

	if running || currentStatus.Ready || currentStatus.Status == "ready" || IsLifecyclePending(currentStatus) {
		l.status.Update(func(status *RuntimeStatus) {
			status.Ready = false
			status.Restoring = false
			status.Status = "stopping"
			status.LastError = ""
			status.CheckedAt = time.Now()
		})
		if stopErr := l.stopAndCloseSessions(CloseRuntimeStopped, l.title+" runtime stopped"); stopErr != nil {
			uninstallOutput = l.id + " uninstalled successfully (stop failed before uninstall: " + stopErr.Error() + ")"
		}
	}
	l.setIntentDesired(false, "web")

	l.runtime.Uninstall()

	l.status.Set(RuntimeStatus{
		Ready:           false,
		Installed:       false,
		Installing:      false,
		InstallProgress: 0,
		InstallStage:    "",
		Status:          "not_installed",
		CheckedAt:       time.Now(),
		CurrentSession:  "",
	})

	writeSuccess(c, RuntimeInstallResult{
		Installed: false,
		Binary:    l.runtime.InstallPath(),
		Output:    uninstallOutput,
		Status:    l.RuntimeStatus(),
	})
}

// Reconfiguration is a change to the agent's settings.
type Reconfiguration struct {
	// Apply writes the settings. Its error answers the route.
	Apply func() error
	// Restart restarts a running agent after Apply, so it reads the new
	// settings. It needs the agent to own device control.
	Restart bool
	// RestartFailed leads the message a failed restart answers with.
	RestartFailed string
	// Refresh rereads the settings when the agent is not restarted.
	Refresh func()
	// Source records who asked, in the runtime intent.
	Source string
}

// Reconfigure applies change under the lifecycle lock and then restarts the
// agent or checks it again. On failure it answers the route and returns
// false; on success the caller answers.
func (l *Lifecycle) Reconfigure(c *gin.Context, change Reconfiguration) bool {
	currentStatus := l.status.Get()
	shouldRestart := change.Restart && (currentStatus.Ready || currentStatus.Status == "ready")
	if shouldRestart {
		// Restart follows Start's lock order: a stable control lease
		// first, then the lifecycle lock around the config write and the
		// stop and start.
		releaseControl, controlErr := l.AcquireControl()
		if controlErr != nil {
			writeErrorWithData(c, controlErr, gin.H{"status": l.RuntimeStatus()})
			return false
		}
		defer releaseControl()
	}

	unlock := l.lockLifecycle()
	defer unlock()
	currentStatus = l.status.Get()
	if change.Restart && !shouldRestart && (currentStatus.Ready || currentStatus.Status == "ready") {
		releaseControl, controlErr := l.AcquireControl()
		if controlErr != nil {
			writeErrorWithData(c, controlErr, gin.H{"status": l.RuntimeStatus()})
			return false
		}
		defer releaseControl()
		shouldRestart = true
	}

	if err := change.Apply(); err != nil {
		writeErrorWithData(c, asError(err, CodeRuntimeUnavailable), nil)
		return false
	}

	if shouldRestart {
		if _, _, haltErr := l.runtime.Halt(); haltErr != nil {
			writeErrorWithData(c, NewError(CodeRuntimeUnavailable, change.RestartFailed+asError(haltErr, CodeRuntimeStartFailed).Message), nil)
			return false
		}
		if _, _, startErr := l.start(context.Background()); startErr != nil {
			writeErrorWithData(c, NewError(CodeRuntimeUnavailable, change.RestartFailed+startErr.Message), nil)
			return false
		}
		l.setIntentDesired(true, change.Source)
	} else {
		if change.Refresh != nil {
			change.Refresh()
		}
		_ = l.ensureReady(context.Background(), true)
	}
	return true
}
