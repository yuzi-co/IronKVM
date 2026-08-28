package hid

import (
	"errors"
	"os"
	"syscall"
	"testing"
	"time"
)

func at(seconds int) time.Time {
	return time.Unix(1700000000, 0).Add(time.Duration(seconds) * time.Second)
}

func TestHealthStartsUnknown(t *testing.T) {
	var h endpointHealth

	got := h.snapshot(at(0))
	if got.State != hidStateUnknown {
		t.Fatalf("state = %q, want %q", got.State, hidStateUnknown)
	}
	if got.Detail != "" {
		t.Fatalf("detail = %q, want empty", got.Detail)
	}
}

func TestHealthFirstSuccessIsAChange(t *testing.T) {
	var h endpointHealth

	if changed := h.record(nil, at(0)).Changed; !changed {
		t.Fatal("first successful write should report a state change")
	}
	if got := h.snapshot(at(0)).State; got != hidStateAccepting {
		t.Fatalf("state = %q, want %q", got, hidStateAccepting)
	}
}

func TestHealthRepeatedSuccessIsNotAChange(t *testing.T) {
	var h endpointHealth

	h.record(nil, at(0))
	if changed := h.record(nil, at(1)).Changed; changed {
		t.Fatal("a second successful write should not report a change")
	}
}

// The distinction that matters: a write that times out means the target is not
// fetching from that endpoint, which is a different fault from the device node
// being gone. Only the first one warrants "switch to relative mouse".
func TestHealthTimeoutStalls(t *testing.T) {
	var h endpointHealth

	h.record(nil, at(0))
	if changed := h.record(os.ErrDeadlineExceeded, at(1)).Changed; !changed {
		t.Fatal("the first timeout should report a state change")
	}

	got := h.snapshot(at(1))
	if got.State != hidStateStalled {
		t.Fatalf("state = %q, want %q", got.State, hidStateStalled)
	}
}

func TestHealthOtherErrorsAreNotStalls(t *testing.T) {
	var h endpointHealth

	h.record(nil, at(0))
	h.record(errors.New("no such device"), at(1))

	got := h.snapshot(at(1))
	if got.State != hidStateError {
		t.Fatalf("state = %q, want %q", got.State, hidStateError)
	}
	if got.Detail != "no such device" {
		t.Fatalf("detail = %q, want the error text", got.Detail)
	}
}

// This is the whole point of the type: 20 identical failures per second became
// 20 log lines per second, and the operator learned nothing after the first.
func TestHealthRepeatedTimeoutsAreNotChanges(t *testing.T) {
	var h endpointHealth

	h.record(os.ErrDeadlineExceeded, at(0))
	for i := 1; i < 100; i++ {
		if changed := h.record(os.ErrDeadlineExceeded, at(i)).Changed; changed {
			t.Fatalf("timeout %d reported a change; only the first should", i)
		}
	}
}

// A stall that has lasted ten minutes has to read as ten minutes, so the clock
// starts when the state began and not when it was last confirmed.
func TestHealthStalledForMeasuresFromTheFirstFailure(t *testing.T) {
	var h endpointHealth

	h.record(os.ErrDeadlineExceeded, at(10))
	h.record(os.ErrDeadlineExceeded, at(20))
	h.record(os.ErrDeadlineExceeded, at(30))

	got := h.snapshot(at(70))
	if want := int64(60000); got.StateForMs != want {
		t.Fatalf("stateForMs = %d, want %d", got.StateForMs, want)
	}
}

// Once the mouse mode is switched away, nothing writes to the dead endpoint any
// more, so its state stops being refreshed. A consumer has to be able to tell a
// live observation from a stale one rather than reporting a fault that may have
// cured itself.
func TestHealthReportsHowOldTheObservationIs(t *testing.T) {
	var h endpointHealth

	h.record(os.ErrDeadlineExceeded, at(10))

	got := h.snapshot(at(25))
	if want := int64(15000); got.ObservedMsAgo != want {
		t.Fatalf("observedMsAgo = %d, want %d", got.ObservedMsAgo, want)
	}
}

func TestHealthRecoveryIsAChange(t *testing.T) {
	var h endpointHealth

	h.record(os.ErrDeadlineExceeded, at(0))
	h.record(os.ErrDeadlineExceeded, at(1))

	if changed := h.record(nil, at(2)).Changed; !changed {
		t.Fatal("recovering should report a state change")
	}
	if got := h.snapshot(at(2)).State; got != hidStateAccepting {
		t.Fatalf("state = %q, want %q", got, hidStateAccepting)
	}
}

// A first successful write is a change, but it is not a recovery, and calling
// it one at boot would report a fault that never happened.
func TestHealthFirstSuccessComesFromUnknown(t *testing.T) {
	var h endpointHealth

	got := h.record(nil, at(0))
	if got.From != hidStateUnknown {
		t.Fatalf("from = %q, want %q", got.From, hidStateUnknown)
	}

	recovery := func() hidTransition {
		h.record(os.ErrDeadlineExceeded, at(1))
		return h.record(nil, at(2))
	}()
	if recovery.From != hidStateStalled {
		t.Fatalf("from = %q, want %q", recovery.From, hidStateStalled)
	}
}

func TestHealthMovingBetweenFaultsIsAChange(t *testing.T) {
	var h endpointHealth

	h.record(os.ErrDeadlineExceeded, at(0))
	if changed := h.record(errors.New("no such device"), at(1)).Changed; !changed {
		t.Fatal("moving from a stall to a different fault should report a change")
	}
	if changed := h.record(os.ErrDeadlineExceeded, at(2)).Changed; !changed {
		t.Fatal("moving back to a stall should report a change")
	}
}

// A wrapped timeout still has to read as a timeout: writeHID returns the error
// with context attached.
func TestHealthUnwrapsTimeouts(t *testing.T) {
	var h endpointHealth

	h.record(wrapped{os.ErrDeadlineExceeded}, at(0))

	if got := h.snapshot(at(0)).State; got != hidStateStalled {
		t.Fatalf("state = %q, want %q", got, hidStateStalled)
	}
}

type wrapped struct{ err error }

func (w wrapped) Error() string { return "timeout after 50ms: " + w.err.Error() }
func (w wrapped) Unwrap() error { return w.err }

// ESHUTDOWN is f_hid's answer when the function is disabled: the gadget is not
// enumerated and every report to every endpoint is lost. That is a different
// fault from a stall, which is one endpoint the host has stopped polling on an
// otherwise working link, and the supervisor acts on only one of the two.
func TestHealthShutdownIsDetachedNotAnError(t *testing.T) {
	for _, err := range []error{syscall.ESHUTDOWN, syscall.ENODEV} {
		var h endpointHealth

		h.record(nil, at(0))
		h.record(&os.PathError{Op: "write", Path: HID0, Err: err}, at(1))

		got := h.snapshot(at(1))
		if got.State != hidStateDetached {
			t.Fatalf("%v gave state %q, want %q", err, got.State, hidStateDetached)
		}
	}
}

func TestHealthDeadlinesAreNeverDetached(t *testing.T) {
	var h endpointHealth

	h.record(os.ErrDeadlineExceeded, at(0))

	if got := h.snapshot(at(0)); got.State != hidStateStalled {
		t.Fatalf("a deadline gave state %q, want %q", got.State, hidStateStalled)
	}
	if !h.linkFaultSince().IsZero() {
		t.Fatal("back-pressure from a live host must not read as a link fault")
	}
}

// The supervisor compares this against a minimum age, so it has to be when the
// run of faults started, not when the last one landed.
func TestLinkFaultSinceIsTheStartOfTheRun(t *testing.T) {
	var h endpointHealth

	h.record(syscall.ESHUTDOWN, at(1))
	h.record(syscall.ESHUTDOWN, at(2))
	h.record(syscall.ESHUTDOWN, at(3))

	if got := h.linkFaultSince(); !got.Equal(at(1)) {
		t.Fatalf("linkFaultSince = %s, want %s", got, at(1))
	}

	h.record(nil, at(4))
	if !h.linkFaultSince().IsZero() {
		t.Fatal("one successful report must clear the link fault")
	}
}

// Hid.linkFaultSince folds three endpoints into one answer, and the earliest is
// when the link went.
func TestHidLinkFaultSinceTakesTheEarliest(t *testing.T) {
	var h Hid

	h.relHealth.record(syscall.ESHUTDOWN, at(5))
	h.kbHealth.record(syscall.ESHUTDOWN, at(2))
	h.absHealth.record(os.ErrDeadlineExceeded, at(1))

	if got := h.linkFaultSince(); !got.Equal(at(2)) {
		t.Fatalf("linkFaultSince = %s, want %s", got, at(2))
	}
}

func TestHidLinkFaultSinceIsZeroWithoutFaults(t *testing.T) {
	var h Hid

	h.kbHealth.record(nil, at(1))
	h.relHealth.record(os.ErrDeadlineExceeded, at(1))

	if got := h.linkFaultSince(); !got.IsZero() {
		t.Fatalf("linkFaultSince = %s with no link fault, want the zero time", got)
	}
}
