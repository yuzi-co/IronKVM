package agent

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestStopProcessAndVerifyForceStopRunsStopBeforeChecking(t *testing.T) {
	stopCalled := false
	err := stopProcessAndVerify(
		"PicoClaw",
		true,
		func() (bool, error) {
			if !stopCalled {
				return false, errors.New("runtime checked before compensating stop")
			}
			return false, nil
		},
		func() error {
			stopCalled = true
			return nil
		},
		time.Second,
		time.Millisecond,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !stopCalled {
		t.Fatal("compensating stop was not called")
	}
}

func TestStopProcessAndVerifyWaitsForExit(t *testing.T) {
	checks := 0
	stopCalled := false
	err := stopProcessAndVerify(
		"PicoClaw",
		false,
		func() (bool, error) {
			checks++
			switch checks {
			case 1:
				return true, nil
			case 2:
				if !stopCalled {
					return false, errors.New("runtime verified before stop")
				}
				return true, nil
			default:
				return false, nil
			}
		},
		func() error {
			stopCalled = true
			return nil
		},
		time.Second,
		0,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !stopCalled || checks != 3 {
		t.Fatalf("stopCalled=%v checks=%d, want true and 3", stopCalled, checks)
	}
}

func TestStopProcessAndVerifyIgnoresStopErrorAfterExit(t *testing.T) {
	checks := 0
	wantStopErr := errors.New("stop command failed after cleanup")
	err := stopProcessAndVerify(
		"PicoClaw",
		false,
		func() (bool, error) {
			checks++
			return checks == 1, nil
		},
		func() error { return wantStopErr },
		time.Second,
		0,
	)
	if err != nil {
		t.Fatalf("error = %v, want nil after runtime exits", err)
	}
	if checks != 2 {
		t.Fatalf("checks = %d, want 2", checks)
	}
}

func TestStopProcessAndVerifyReportsStopAndVerificationFailures(t *testing.T) {
	wantStopErr := errors.New("stop command failed")
	err := stopProcessAndVerify(
		"PicoClaw",
		true,
		func() (bool, error) { return true, nil },
		func() error { return wantStopErr },
		0,
		0,
	)
	if !errors.Is(err, wantStopErr) {
		t.Fatalf("error = %v, want wrapped stop error", err)
	}
	if !strings.Contains(err.Error(), "PicoClaw runtime is still running") {
		t.Fatalf("error = %v, want verification failure", err)
	}
}

func TestReadyWaitHonorsLifecycleCancellation(t *testing.T) {
	l := NewLifecycle(LifecycleConfig{ID: "picoclaw", Title: "PicoClaw"})
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(ErrControlModeSwitch)

	err := l.waitReady(ctx, time.Hour)
	if err == nil || err.Code != CodeControlModeConflict {
		t.Fatalf("error = %+v, want %s", err, CodeControlModeConflict)
	}
}

// stoppedRuntime is a Runtime whose process is not running.
type stoppedRuntime struct{ Runtime }

func (stoppedRuntime) Running() (bool, error) { return false, nil }

type lockStub struct{ owner string }

func (l *lockStub) Owner() string { return l.owner }

func (l *lockStub) Release(string) bool {
	l.owner = ""
	return true
}

func TestStopAndCloseSessionsReleasesAStaleLockWithoutHID(t *testing.T) {
	lock := &lockStub{owner: "stale-session"}
	l := NewLifecycle(LifecycleConfig{
		ID:         "picoclaw",
		Title:      "PicoClaw",
		Runtime:    stoppedRuntime{},
		Status:     NewStatusStore("/usr/bin/picoclaw", RuntimeStatus{}),
		Lock:       lock,
		ReleaseHID: func() error { return errors.New("hid unavailable") },
	})

	if err := l.stopAndCloseSessions(CloseRuntimeStopped, "test stop"); err != nil {
		t.Fatalf("stopAndCloseSessions error = %v, want nil", err)
	}
	if lock.owner != "" {
		t.Fatalf("lock owner = %q, want released", lock.owner)
	}
}
