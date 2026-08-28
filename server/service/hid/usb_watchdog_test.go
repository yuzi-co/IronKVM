package hid

import (
	"testing"
	"time"
)

var epoch = time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)

// freezeClock pins the supervisor's clock so observe() records a known
// faultSince, and restores it afterwards.
func freezeClock(t *testing.T, at time.Time) {
	t.Helper()
	previous := usbWatchdogNow
	usbWatchdogNow = func() time.Time { return at }
	t.Cleanup(func() {
		usbWatchdogNow = previous
		// The settle window is package state. A test that opens one against a
		// frozen clock would otherwise leave it open for whatever runs next.
		usbMutationMu.Lock()
		usbSettleUntil = time.Time{}
		usbMutationMu.Unlock()
	})
}

// detached drives the supervisor to a state where the link has been dead since
// `since`, with `seen` recording whether it ever worked.
func detachedSince(seen bool, since time.Time) *usbWatchdog {
	return &usbWatchdog{sawHealthy: seen, faultSince: since}
}

func TestHealthyLinkNeverActs(t *testing.T) {
	w := &usbWatchdog{sawHealthy: true}

	if got := w.decide(epoch.Add(time.Hour), time.Time{}); got != usbActionNone {
		t.Fatalf("a healthy link asked for %v", got)
	}
}

func TestDebounceHoldsBeforeActing(t *testing.T) {
	w := detachedSince(true, epoch)

	if got := w.decide(epoch.Add(usbFaultDebounceSeen-time.Second), time.Time{}); got != usbActionNone {
		t.Fatalf("acted %s into a %s debounce", usbFaultDebounceSeen-time.Second, usbFaultDebounceSeen)
	}
	if got := w.decide(epoch.Add(usbFaultDebounceSeen), time.Time{}); got != usbActionRebind {
		t.Fatalf("after the debounce got %v, want a rebind", got)
	}
}

// A link that has never worked is indistinguishable from a KVM sitting in a
// switched-off computer, so it waits far longer.
func TestUnseenLinkWaitsLonger(t *testing.T) {
	w := detachedSince(false, epoch)

	if got := w.decide(epoch.Add(usbFaultDebounceSeen+time.Second), time.Time{}); got != usbActionNone {
		t.Fatalf("a never-healthy link acted after the short debounce: %v", got)
	}
	if got := w.decide(epoch.Add(usbFaultDebounceUnseen), time.Time{}); got != usbActionRebind {
		t.Fatalf("after the long debounce got %v, want a rebind", got)
	}
}

// Somebody typing into a dead link is not an idle powered-off host, and that
// evidence collapses the long wait to the short one.
func TestSustainedHidFaultsShortenTheWait(t *testing.T) {
	w := detachedSince(false, epoch)
	now := epoch.Add(usbFaultDebounceSeen)

	if got := w.decide(now, time.Time{}); got != usbActionNone {
		t.Fatalf("acted without HID evidence: %v", got)
	}
	if got := w.decide(now, now.Add(-usbHidFaultMinAge)); got != usbActionRebind {
		t.Fatalf("with %s of HID faults got %v, want a rebind", usbHidFaultMinAge, got)
	}
}

// A burst of failures around a legitimate gadget operation is not a person
// driving a dead link.
func TestBriefHidFaultsDoNotShortenTheWait(t *testing.T) {
	w := detachedSince(false, epoch)
	now := epoch.Add(usbFaultDebounceSeen)

	if got := w.decide(now, now.Add(-time.Second)); got != usbActionNone {
		t.Fatalf("one second of HID faults triggered %v", got)
	}
}

func TestBackoffHoldsBetweenAttempts(t *testing.T) {
	w := detachedSince(true, epoch)
	now := epoch.Add(usbFaultDebounceSeen)

	if got := w.decide(now, time.Time{}); got != usbActionRebind {
		t.Fatalf("first decision was %v", got)
	}
	// poll owns this bookkeeping; decide only names the action.
	w.attempt++
	w.nextAttemptAfter = now.Add(usbBackoff(w.attempt))

	if got := w.decide(now.Add(usbRecoveryBackoffMin-time.Second), time.Time{}); got != usbActionNone {
		t.Fatalf("acted inside the backoff: %v", got)
	}
	if got := w.decide(now.Add(usbRecoveryBackoffMin), time.Time{}); got != usbActionRebind {
		t.Fatalf("after the backoff got %v, want the second rebind", got)
	}
}

// The whole ladder: rebind, rebind, rebuild, then stop. It must never reach
// anything beyond the rebuild, because restart_phy can strand a board that has
// no remote power cycle.
func TestLadderEscalatesThenGivesUp(t *testing.T) {
	w := detachedSince(true, epoch)
	now := epoch.Add(usbFaultDebounceSeen)

	want := []usbAction{usbActionRebind, usbActionRebind, usbActionRebuild, usbActionGiveUp}
	for i, expected := range want {
		got := w.decide(now, time.Time{})
		if got != expected {
			t.Fatalf("rung %d was %v, want %v", i+1, got, expected)
		}
		if got == usbActionGiveUp {
			break
		}
		w.attempt++
		now = now.Add(usbBackoff(w.attempt))
		w.nextAttemptAfter = time.Time{}
	}

	// poll owns this bookkeeping; decide only names the action.
	w.gaveUp = true

	// It stays given up, rather than saying so every two seconds forever.
	if got := w.decide(now.Add(24*time.Hour), time.Time{}); got != usbActionNone {
		t.Fatalf("acted a day after giving up: %v", got)
	}
}

func TestRecoveryResetsOnceTheLinkIsBack(t *testing.T) {
	freezeClock(t, epoch)
	w := &usbWatchdog{sawHealthy: true, attempt: 2, gaveUp: true, faultSince: epoch.Add(-time.Hour)}

	w.observe(usbLink{State: udcStateConfigured, Speed: udcSpeedHigh})

	if !w.faultSince.IsZero() || w.attempt != 0 || w.gaveUp || !w.nextAttemptAfter.IsZero() {
		t.Fatalf("a healthy sample left %+v", w)
	}
}

// A host that goes to sleep reports "suspended". Nothing must rebind for it,
// and a fault timer that was already running must not be cleared by it either.
func TestPendingStatesNeitherStartNorClearTheTimer(t *testing.T) {
	freezeClock(t, epoch)

	fresh := &usbWatchdog{sawHealthy: true}
	fresh.observe(usbLink{State: "suspended", Speed: udcSpeedHigh})
	if !fresh.faultSince.IsZero() {
		t.Fatal("a suspended link started the fault timer")
	}

	running := detachedSince(true, epoch.Add(-time.Minute))
	running.observe(usbLink{State: "addressed"})
	if running.faultSince != epoch.Add(-time.Minute) {
		t.Fatalf("a pending sample moved the fault timer to %s", running.faultSince)
	}
}

// A degraded link is a fault in its own right. Before this, only "not attached"
// counted, and a board enumerated at full speed was left alone forever.
func TestDegradedLinkStartsTheFaultTimer(t *testing.T) {
	freezeClock(t, epoch)
	w := &usbWatchdog{sawHealthy: true}

	w.observe(usbLink{State: udcStateConfigured, Speed: "full-speed"})

	if w.faultSince.IsZero() {
		t.Fatal("a full-speed link did not start the fault timer")
	}
	if got := w.decide(epoch.Add(usbFaultDebounceSeen), time.Time{}); got != usbActionRebind {
		t.Fatalf("a degraded link asked for %v, want a rebind", got)
	}
}

func TestDetachedLinkStartsTheFaultTimerOnce(t *testing.T) {
	freezeClock(t, epoch)
	w := &usbWatchdog{sawHealthy: true}

	w.observe(usbLink{State: udcStateDetached})
	first := w.faultSince

	usbWatchdogNow = func() time.Time { return epoch.Add(time.Minute) }
	w.observe(usbLink{State: udcStateDetached})

	if w.faultSince != first {
		t.Fatalf("a second detached sample moved the fault timer from %s to %s", first, w.faultSince)
	}
}

func TestBackoffDoublesAndSaturates(t *testing.T) {
	if got := usbBackoff(1); got != usbRecoveryBackoffMin {
		t.Errorf("first backoff is %s, want %s", got, usbRecoveryBackoffMin)
	}
	if got := usbBackoff(2); got != 2*usbRecoveryBackoffMin {
		t.Errorf("second backoff is %s, want %s", got, 2*usbRecoveryBackoffMin)
	}
	// The shift overflows long before this; it must saturate rather than wrap
	// to a negative duration and act every poll.
	for _, attempt := range []int{10, 40, 63, 64, 100} {
		if got := usbBackoff(attempt); got != usbRecoveryBackoffMax {
			t.Errorf("backoff at attempt %d is %s, want %s", attempt, got, usbRecoveryBackoffMax)
		}
	}
}

func TestSettleWindowSuppressesTheSupervisor(t *testing.T) {
	freezeClock(t, epoch)

	NoteUSBGadgetMutated()
	if !usbGadgetSettling() {
		t.Fatal("the settle window did not open")
	}

	usbWatchdogNow = func() time.Time { return epoch.Add(usbSettleAfterMutation) }
	if usbGadgetSettling() {
		t.Fatalf("the settle window outlasted %s", usbSettleAfterMutation)
	}
}

// poll is where the sysfs reader and the decision meet, so it is worth one test
// of its own. It stops short of the recovery: that path reopens /dev/hidg*,
// which no test host has.
func TestPollObservesTheLinkAndRespectsTheSettleWindow(t *testing.T) {
	freezeClock(t, epoch)
	fakeUDC(t, map[string]string{"state": "configured", "current_speed": "high-speed"})

	w := &usbWatchdog{}
	w.poll()
	if !w.sawHealthy || !w.faultSince.IsZero() {
		t.Fatalf("a healthy sample left sawHealthy=%v faultSince=%s", w.sawHealthy, w.faultSince)
	}

	fakeUDC(t, map[string]string{"state": "not attached", "current_speed": "UNKNOWN"})
	w.poll()
	if w.faultSince.IsZero() {
		t.Fatal("a detached sample did not start the fault timer")
	}

	// Long past the debounce, so poll would act. The settle window must stop it
	// anyway: a mount, a mode switch or its own last repair is in flight, and
	// the transient it would read is that operation, not a fault.
	usbWatchdogNow = func() time.Time { return epoch.Add(time.Hour) }
	NoteUSBGadgetMutated()
	w.poll()
	if w.attempt != 0 {
		t.Fatalf("poll acted inside the settle window: attempt=%d", w.attempt)
	}
}
