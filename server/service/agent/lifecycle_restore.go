package agent

import (
	"context"
	"errors"
	"time"

	"NanoKVM-Server/service/controlmode"

	log "github.com/sirupsen/logrus"
)

// StartRestore restores the persisted intent once, in the background: it
// starts an agent that should run and owns device control, and stops one
// that should not run.
func (l *Lifecycle) StartRestore() {
	l.reconcileOnce.Do(func() {
		go l.restore("startup")
	})
}

func (l *Lifecycle) restore(source string) {
	startedAt := time.Now()

	unlock := l.lockLifecycle()
	defer unlock()

	intent, intentErr := l.intent.Load()
	if intentErr != nil {
		l.setIntentError(intentErr.Error())
		log.WithError(intentErr).Warnf("%s runtime restore skipped because intent is invalid", l.title)
		return
	}

	modeStatus, modeErr := l.control.Status()
	if modeErr != nil {
		l.setIntentError(modeErr.Error())
		log.WithError(modeErr).Warnf("%s runtime restore skipped because control mode is unavailable", l.title)
		return
	}

	fields := log.Fields{
		"source":          source,
		"desired_running": intent.DesiredRunning,
		"control_mode":    string(modeStatus.Mode),
	}

	if !intent.DesiredRunning {
		l.restoreDisabled(source)
		log.WithFields(fields).Infof("%s runtime restore skipped because desired_running is false", l.title)
		return
	}
	if modeStatus.Mode == controlmode.ModeMCP {
		l.status.Update(func(status *RuntimeStatus) {
			status.Ready = false
			status.Status = "blocked_by_mcp"
			status.CheckedAt = time.Now()
		})
		log.WithFields(fields).Infof("%s runtime restore skipped because MCP owns device control", l.title)
		return
	}
	if modeStatus.Mode != controlmode.ModePicoclaw {
		l.status.Update(func(status *RuntimeStatus) {
			status.Ready = false
			status.Status = "stopped"
			status.CheckedAt = time.Now()
		})
		log.WithFields(fields).Infof("%s runtime restore skipped because %s does not own device control", l.title, l.title)
		return
	}

	releaseControl, controlErr := l.AcquireControl()
	if controlErr != nil {
		l.setIntentError(controlErr.Message)
		log.WithFields(fields).Warnf("%s runtime restore skipped because %s control is not stable", l.title, l.title)
		return
	}
	defer releaseControl()

	intent, intentErr = l.intent.Load()
	if intentErr != nil {
		l.setIntentError(intentErr.Error())
		log.WithError(intentErr).Warnf("%s runtime restore skipped because intent changed to an invalid state", l.title)
		return
	}
	if !intent.DesiredRunning {
		fields["desired_running"] = intent.DesiredRunning
		log.WithFields(fields).Infof("%s runtime restore skipped because desired_running changed while waiting", l.title)
		return
	}

	l.status.Update(func(status *RuntimeStatus) {
		status.Ready = false
		status.Restoring = true
		status.Status = "restoring"
		status.LastError = ""
		status.CheckedAt = time.Now()
	})

	if readyErr := l.ensureReady(context.Background(), true); readyErr == nil {
		fields["elapsed_ms"] = time.Since(startedAt).Milliseconds()
		l.status.Update(func(status *RuntimeStatus) {
			status.Restoring = false
			status.Status = "ready"
			status.CheckedAt = time.Now()
		})
		log.WithFields(fields).Infof("%s runtime restore found an already-ready runtime", l.title)
		return
	}

	currentStatus := l.status.Get()
	switch currentStatus.Status {
	case "not_installed", "model_not_configured", "config_error":
		message := currentStatus.LastError
		if message == "" {
			message = currentStatus.ConfigError
		}
		if message == "" {
			message = l.title + " runtime restore prerequisites are not satisfied"
		}
		l.setIntentError(message)
		l.status.Update(func(status *RuntimeStatus) {
			status.Restoring = false
			status.LastError = message
			status.CheckedAt = time.Now()
		})
		fields["elapsed_ms"] = time.Since(startedAt).Milliseconds()
		fields["runtime_status"] = currentStatus.Status
		log.WithFields(fields).Warnf("%s runtime restore skipped because prerequisites are not satisfied", l.title)
		return
	}

	command, output, startErr := l.start(context.Background())
	fields["elapsed_ms"] = time.Since(startedAt).Milliseconds()
	fields["command"] = command
	if output != "" {
		fields["output"] = output
	}
	if startErr != nil {
		l.setIntentError(startErr.Message)
		l.status.Update(func(status *RuntimeStatus) {
			status.Restoring = false
			if status.Status == "restoring" {
				status.Status = "unavailable"
			}
			status.LastError = startErr.Message
			status.CheckedAt = time.Now()
		})
		log.WithFields(fields).WithError(errors.New(startErr.Message)).Warnf("%s runtime restore failed", l.title)
		return
	}

	intent, intentErr = l.intent.Load()
	if intentErr != nil {
		l.setIntentError(intentErr.Error())
		log.WithError(intentErr).Warnf("%s runtime restore result discarded because intent became invalid", l.title)
		return
	}
	if !intent.DesiredRunning {
		log.WithFields(fields).Infof("%s runtime restore result discarded because desired_running is now false", l.title)
		return
	}
	if controlErr := l.requireControl(); controlErr != nil {
		l.setIntentError(controlErr.Message)
		log.WithFields(fields).Warnf("%s runtime restore result discarded because control changed", l.title)
		return
	}

	l.setIntentDesired(true, "restore")
	l.status.Update(func(status *RuntimeStatus) {
		status.Restoring = false
		status.CheckedAt = time.Now()
	})
	log.WithFields(fields).Infof("%s runtime restored from persisted intent", l.title)
}

// restoreDisabled stops an agent the intent says should not run.
func (l *Lifecycle) restoreDisabled(source string) {
	status := l.status.Get()
	running, err := l.runtime.Running()
	fields := log.Fields{
		"source":         source,
		"runtime_status": status.Status,
		"ready":          status.Ready,
	}
	if err != nil {
		l.setIntentError(err.Error())
		l.status.Update(func(status *RuntimeStatus) {
			status.Ready = false
			status.Restoring = false
			status.Status = "unavailable"
			status.LastError = err.Error()
			status.CheckedAt = time.Now()
		})
		log.WithFields(fields).WithError(err).Warnf("%s disabled runtime intent could not check runtime", l.title)
		return
	}

	if !running && !status.Ready && !IsLifecyclePending(status) {
		l.applyStoppedStatus()
		log.WithFields(fields).Infof("%s disabled runtime intent kept runtime stopped", l.title)
		return
	}

	l.status.Update(func(status *RuntimeStatus) {
		status.Ready = false
		status.Restoring = false
		status.Status = "stopping"
		status.LastError = ""
		status.CheckedAt = time.Now()
	})
	if err := l.stopAndCloseSessions(CloseRuntimeStopped, l.title+" runtime disabled"); err != nil {
		l.setIntentError(err.Error())
		l.status.Update(func(status *RuntimeStatus) {
			status.Ready = false
			status.Restoring = false
			status.Status = "unavailable"
			status.LastError = err.Error()
			status.CheckedAt = time.Now()
		})
		log.WithFields(fields).WithError(err).Warnf("%s disabled runtime intent failed to stop runtime", l.title)
		return
	}

	l.applyStoppedStatus()
	log.WithFields(fields).Infof("%s disabled runtime intent stopped runtime", l.title)
}
