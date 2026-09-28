package controlmode

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestManagerDefaultsToPicoclaw(t *testing.T) {
	manager := NewManager(filepath.Join(t.TempDir(), "mode"), ModePicoclaw)
	if got := manager.Current(); got != ModePicoclaw {
		t.Fatalf("mode = %q, want %q", got, ModePicoclaw)
	}
}

func TestSwitchPreemptsBeforeWaitingForActiveWrite(t *testing.T) {
	manager := NewManager(filepath.Join(t.TempDir(), "mode"), ModeMCP)
	release, err := manager.AcquireWrite(ModeMCP)
	if err != nil {
		t.Fatal(err)
	}

	preempted := make(chan struct{})
	blocked := make(chan struct{})
	preemptedBeforeWait := false
	var blockedOnce sync.Once
	manager.activity.blocked = func() {
		blockedOnce.Do(func() {
			select {
			case <-preempted:
				preemptedBeforeWait = true
			default:
			}
			close(blocked)
		})
	}

	done := make(chan error, 1)
	go func() {
		done <- manager.Switch(ModePicoclaw, func() error {
			close(preempted)
			return nil
		})
	}()

	// No deadline here. A switch that waited first would sit in the gate until
	// its own activity timeout and then return an error, which the second case
	// reports.
	select {
	case <-blocked:
	case err := <-done:
		t.Fatalf("switch returned %v without waiting for the active write", err)
	}
	if !preemptedBeforeWait {
		t.Fatal("preempt callback was not called before waiting")
	}
	if _, err := manager.AcquireWrite(ModeMCP); err == nil {
		t.Fatal("new write acquired control during transition")
	}

	release()
	if err := <-done; err != nil {
		t.Fatalf("switch did not resume after active write was released: %v", err)
	}
	if got := manager.Current(); got != ModePicoclaw {
		t.Fatalf("mode = %q, want picoclaw", got)
	}
}

func TestSwitchRunsCleanupAfterActiveWritesDrain(t *testing.T) {
	manager := NewManager(filepath.Join(t.TempDir(), "mode"), ModePicoclaw)
	release, err := manager.AcquireWrite(ModePicoclaw)
	if err != nil {
		t.Fatal(err)
	}

	blocked := make(chan struct{})
	var blockedOnce sync.Once
	manager.activity.blocked = func() {
		blockedOnce.Do(func() { close(blocked) })
	}

	cleanupCalled := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- manager.SwitchWithCleanup(ModeMCP, nil, func() error {
			close(cleanupCalled)
			return nil
		})
	}()

	// Once the switch is blocked in the gate, cleanup can only run after the
	// release below, so the check that follows does not depend on timing.
	select {
	case <-blocked:
	case err := <-done:
		t.Fatalf("switch returned %v without waiting for the active write", err)
	}
	select {
	case <-cleanupCalled:
		t.Fatal("cleanup ran before active write drained")
	default:
	}

	release()
	if err := <-done; err != nil {
		t.Fatalf("switch did not finish: %v", err)
	}
	select {
	case <-cleanupCalled:
	default:
		t.Fatal("cleanup was not called")
	}
}

func TestSwitchCleanupFailureFailsClosed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mode")
	manager := NewManager(path, ModePicoclaw)
	wantErr := errors.New("hid release failed")

	if err := manager.SwitchWithCleanup(ModeMCP, nil, func() error { return wantErr }); !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
	if got := manager.Current(); got != ModeOff {
		t.Fatalf("mode = %q, want off after cleanup failure", got)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "off\n" {
		t.Fatalf("mode file = %q, want off", data)
	}
}

func TestSwitchIfKeepsDifferentMode(t *testing.T) {
	manager := NewManager(filepath.Join(t.TempDir(), "mode"), ModePicoclaw)
	switched, err := manager.SwitchIf(ModeMCP, ModeOff, nil)
	if err != nil {
		t.Fatal(err)
	}
	if switched || manager.Current() != ModePicoclaw {
		t.Fatalf("switched=%v mode=%q, want unchanged picoclaw", switched, manager.Current())
	}
}

func TestSwitchTimeoutDoesNotLeaveManagerTransitioning(t *testing.T) {
	manager := NewManager(filepath.Join(t.TempDir(), "mode"), ModeMCP)
	manager.activityWaitTimeout = 20 * time.Millisecond
	cleanupCalled := false

	status, release, err := manager.AcquireStable()
	if err != nil {
		t.Fatal(err)
	}
	if status.Mode != ModeMCP {
		t.Fatalf("mode = %q, want MCP", status.Mode)
	}

	err = manager.SwitchWithCleanup(ModePicoclaw, nil, func() error {
		cleanupCalled = true
		return nil
	})
	if !errors.Is(err, ErrActivityWaitTimeout) {
		t.Fatalf("switch error = %v, want %v", err, ErrActivityWaitTimeout)
	}
	if cleanupCalled {
		t.Fatal("cleanup ran even though exclusive activity lease was not acquired")
	}
	status, err = manager.Status()
	if err != nil {
		t.Fatal(err)
	}
	if status.Mode != ModeMCP || status.Transitioning {
		t.Fatalf("status after timeout = %+v, want stable MCP", status)
	}

	release()
	if err := manager.Switch(ModePicoclaw, nil); err != nil {
		t.Fatalf("switch after lease release failed: %v", err)
	}
}

func TestInvalidModeFailsClosed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mode")
	if err := os.WriteFile(path, []byte("invalid\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	manager := NewManager(path, ModePicoclaw)
	status, err := manager.Status()
	if err != nil {
		t.Fatal(err)
	}
	if status.Mode != ModeOff {
		t.Fatalf("mode = %q, want off", status.Mode)
	}
	if !strings.Contains(status.LastError, "invalid") {
		t.Fatalf("last_error = %q, want invalid mode detail", status.LastError)
	}
}

func TestStatusReloadsExternallyModifiedModeFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mode")
	if err := os.WriteFile(path, []byte("off\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	manager := NewManager(path, ModePicoclaw)
	if got := manager.Current(); got != ModeOff {
		t.Fatalf("initial mode = %q, want off", got)
	}

	if err := os.WriteFile(path, []byte("mcp\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	modTime := time.Now().Add(time.Second)
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatal(err)
	}
	if got := manager.Current(); got != ModeMCP {
		t.Fatalf("mode after external write = %q, want MCP", got)
	}
}
