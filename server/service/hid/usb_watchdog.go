package hid

// usb_watchdog.go supervises the USB gadget for as long as the server runs.
//
// Until now the gadget was repaired at exactly two moments: S03usbdev's boot
// watch, which spends a fixed budget in the first seconds of uptime and then
// exits for good, and whichever operator happened to press "reset USB". A link
// that died at any other moment stayed dead, because nothing was watching. The
// board keeps answering on eth0 the whole time, so the web UI looks healthy and
// only the keyboard is gone. That is the failure this file exists to end.
//
// Three decisions here are specific to this board and are not the obvious ones.
//
// FIRST, health is "configured AND high-speed", not "configured". A gadget that
// enumerated at full speed reports configured and does not work. See
// usb_link.go, which grades the sample, and S03usbdev's usb_link_ok, which
// makes the same test during boot.
//
// SECOND, the escalation never reaches restart_phy. That command unbinds the
// dwc2 controller from its driver, and a bind that then fails leaves the gadget
// wedged with no way back: rewriting UDC does not recover it, and neither does
// another restart_phy. Only a full configfs teardown does, and that needs a
// shell on the device. This board has no remote power cycle, so an unattended
// escalation that can strand it is not an option however rare the failure is.
// restart_phy stays where it is, behind an operator pressing a button and
// watching what happens. The ladder here stops at stop_start, which flips the
// controller's role and rebuilds the configuration, and which is the recovery
// that was actually observed returning a full-speed link to high-speed.
//
// THIRD, it gives up. A board plugged into a host that is genuinely full-speed
// only, or switched off for the weekend, is working as well as it can. After
// the ladder is spent this stops touching the gadget and says so once, rather
// than rebinding for the rest of its uptime.
//
// What it deliberately does NOT cover: a gadget that is configured at
// high-speed and still has dead endpoints because the composition asked for
// more IN endpoints than the controller has FIFOs. That reads healthy here, by
// design. It is a composition fault, no amount of rebinding fixes it, and
// S03usbdev's endpoint budget refuses it before the bind.

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	// How often the link is sampled. Two small sysfs reads, so the cost is
	// noise next to anything else on this board.
	usbPollInterval = 2 * time.Second

	// How long a fault must persist before acting, once the link has been seen
	// working since this server started. Long enough to sit through a host
	// reboot's own disconnect, short enough that a wedged gadget comes back
	// without anybody noticing it went.
	usbFaultDebounceSeen = 15 * time.Second

	// The same, for a link that has never been seen working. That is a bind
	// which silently failed, and it is also the exact reading of a KVM sitting
	// in a computer that is switched off, so it waits far longer before
	// touching anything.
	usbFaultDebounceUnseen = 3 * time.Minute

	// Floor between attempts, doubling while the fault persists, reset the
	// moment the link comes back.
	usbRecoveryBackoffMin = 30 * time.Second
	usbRecoveryBackoffMax = 15 * time.Minute

	// How many attempts before this gives up and leaves the gadget alone. Two
	// rebinds and one rebuild. S03usbdev's boot watch stops after two rebinds
	// for the same reason.
	usbRecoveryAttempts = 3

	// A deliberate gadget operation unbinds and rebinds the UDC. Ignore the
	// link for this long afterwards so the supervisor cannot race the mount
	// path, a HID mode switch, or its own recovery.
	usbSettleAfterMutation = 10 * time.Second

	// How long an endpoint must have been refusing reports before that counts
	// as proof somebody is driving a dead link. A burst around any legitimate
	// gadget operation says nothing; five seconds of it is a person.
	usbHidFaultMinAge = 5 * time.Second
)

// usbMutation records that something deliberately disturbed the gadget.
//
// Guarded by its own mutex rather than any of the HID locks: the supervisor has
// to be able to read it without waiting behind a media operation that takes
// seconds, which is exactly when it most needs the answer.
var (
	usbMutationMu     sync.Mutex
	usbSettleUntil    time.Time
	usbWatchdogNow    = time.Now
	usbWatchdogOnce   sync.Once
	usbRecoverRebind  = func() error { return usbDevCommand("restart") }
	usbRecoverRebuild = func() error { return usbDevCommand("stop_start") }
)

// NoteUSBGadgetMutated tells the supervisor to hold off. Call it around
// anything that unbinds the UDC on purpose: mounting an image, switching HID
// mode, resetting the controller. Without it the transient "not attached" in
// the middle of a legitimate operation looks exactly like the failure this
// watches for.
func NoteUSBGadgetMutated() {
	usbMutationMu.Lock()
	usbSettleUntil = usbWatchdogNow().Add(usbSettleAfterMutation)
	usbMutationMu.Unlock()
}

func usbGadgetSettling() bool {
	usbMutationMu.Lock()
	defer usbMutationMu.Unlock()
	return usbWatchdogNow().Before(usbSettleUntil)
}

// usbWatchdog is the supervisor's state between samples. It is a struct so that
// decide, which holds all the judgement, is a pure function of observed state
// and needs no filesystem, no sleeps and no gadget to test.
type usbWatchdog struct {
	// sawHealthy records that the link worked at least once. It is what
	// separates "this died" from "this was never alive", and it chooses which
	// debounce applies.
	sawHealthy bool

	// faultSince is when the current unbroken run of faulty samples began, and
	// the zero time when the link is not currently faulty.
	faultSince time.Time

	// nextAttemptAfter throttles escalation; attempt drives both the backoff
	// and how far up the ladder to go.
	nextAttemptAfter time.Time
	attempt          int

	// gaveUp stops the ladder after the last rung, and stops the log line that
	// says so from repeating every two seconds.
	gaveUp bool

	lastState string
}

type usbAction int

const (
	usbActionNone usbAction = iota

	// usbActionRebind writes an empty UDC and binds it again. Cheap, targeted,
	// and it resolves the common case: a stale session left behind when a host
	// went away.
	usbActionRebind

	// usbActionRebuild is S03usbdev's stop_start. It returns the controller to
	// host role and composes the gadget again, which is a real renegotiation
	// rather than a re-attachment. This is what recovered a full-speed link.
	usbActionRebuild

	// usbActionGiveUp ends the ladder. There is no rung above the rebuild, so
	// what is left is to stop touching the gadget and say so once.
	usbActionGiveUp
)

// StartUSBWatchdog starts the supervisor. Safe to call more than once; only the
// first call starts anything.
//
// Best-effort in the same spirit as the rest of the gadget code: it must never
// be able to take the server down, so it recovers from a panic and only logs.
func StartUSBWatchdog() {
	usbWatchdogOnce.Do(func() {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("usb watchdog panicked, no longer supervising the gadget: %v", r)
				}
			}()

			w := &usbWatchdog{}
			for {
				time.Sleep(usbPollInterval)
				w.poll()
			}
		}()
	})
}

func (w *usbWatchdog) poll() {
	link, err := readUSBLink()
	if err != nil {
		// No controller to read at all. That is not a link fault, it is a
		// kernel with no gadget controller, and nothing here can repair it.
		// Logged on a change only, because saying it every two seconds helps
		// nobody.
		w.observe(usbLink{State: "unavailable"})
		return
	}

	w.observe(link)

	// A deliberate operation is in flight. Whatever the link reads right now is
	// about that operation, not about a fault.
	if usbGadgetSettling() {
		return
	}

	now := usbWatchdogNow()
	action := w.decide(now, GetHid().linkFaultSince())
	if action == usbActionNone {
		return
	}

	if action == usbActionGiveUp {
		// A host that is genuinely full-speed only, or switched off, is a board
		// working as well as it can. It must not spend the rest of its uptime
		// rebinding, and it must not say so every two seconds either.
		log.Warnf("usb watchdog: %s, and %d attempt(s) did not recover it. Leaving the gadget alone. "+
			"Reset the USB controller from Settings, or check the cable and the host",
			link.describe(), w.attempt)
		w.gaveUp = true
		return
	}

	w.attempt++
	w.nextAttemptAfter = now.Add(usbBackoff(w.attempt))

	// Hold the supervisor off its own repair. Both rungs unbind the UDC, so
	// without this the next sample reads the recovery itself as a fresh fault.
	NoteUSBGadgetMutated()

	held := now.Sub(w.faultSince).Truncate(time.Second)

	switch action {
	case usbActionRebind:
		log.Warnf("usb watchdog: %s, and has for %s. Rebinding the gadget (attempt %d of %d)",
			link.describe(), held, w.attempt, usbRecoveryAttempts)
		w.recover(usbRecoverRebind, "rebind")
	case usbActionRebuild:
		log.Warnf("usb watchdog: %s, and %d rebind(s) did not fix it. Rebuilding the gadget",
			link.describe(), w.attempt-1)
		w.recover(usbRecoverRebuild, "rebuild")
	}
}

// recover runs one rung of the ladder and puts the HID descriptors back.
//
// The descriptors have to be closed across it. A handle held open over a rebind
// refers to a function the host can no longer see, so it would survive the
// repair and fail every write afterwards. Reopening retries rather than taking
// one shot, because /dev/hidg* reappear a moment after the bind and a single
// attempt loses the race often enough to matter.
func (w *usbWatchdog) recover(run func() error, what string) {
	h := GetHid()
	h.Lock()
	h.CloseNoLock()
	defer func() {
		if err := h.OpenNoLockWithRetry(hidReopenTimeout, hidReopenRetryDelay); err != nil {
			log.Errorf("usb watchdog: reopen the HID devices after %s: %s", what, err)
		}
		h.Unlock()
	}()

	if err := run(); err != nil {
		log.Errorf("usb watchdog: %s failed: %s", what, err)
	}
}

func usbBackoff(attempt int) time.Duration {
	backoff := usbRecoveryBackoffMin << (attempt - 1)
	if backoff > usbRecoveryBackoffMax || backoff <= 0 {
		backoff = usbRecoveryBackoffMax
	}
	return backoff
}

// observe folds one sample into the supervisor's state and logs transitions.
//
// The transitions are logged on purpose, at info. A link failure in the field
// currently leaves almost nothing behind: the kernel ring on this board wraps
// in about forty minutes because the video pipeline writes to it every second,
// so by the time anybody looks, every dwc2 line has gone. A timestamped
// sequence of link states is what separates "the host cut VBUS and we never
// came back" from "we never enumerated in the first place", and those are
// different bugs with different fixes.
func (w *usbWatchdog) observe(link usbLink) {
	state := link.State
	if link.health() == linkDegraded {
		// The speed is the whole story in this case, and a bare "configured"
		// in the log would hide it.
		state = link.State + " at " + link.Speed
	}

	if state != w.lastState {
		if w.lastState == "" {
			log.Infof("usb watchdog: link state %q", state)
		} else {
			log.Infof("usb watchdog: link state %q -> %q", w.lastState, state)
		}
		w.lastState = state
	}

	switch link.health() {
	case linkHealthy:
		if !w.faultSince.IsZero() || w.attempt > 0 {
			log.Infof("usb watchdog: the link is healthy again")
		}
		w.sawHealthy = true
		w.faultSince = time.Time{}
		w.attempt = 0
		w.nextAttemptAfter = time.Time{}
		w.gaveUp = false
	case linkDetached, linkDegraded:
		if w.faultSince.IsZero() {
			w.faultSince = usbWatchdogNow()
		}
	case linkPending:
		// Mid-enumeration, suspended, or no controller. Not health, but not
		// evidence of a fault either: a host that goes to sleep reports
		// "suspended" and is working perfectly. Any running fault timer is left
		// alone rather than cleared, so a link flapping through these states
		// still escalates eventually.
	}
}

// decide chooses whether to act and how far to escalate. It reads the struct
// and changes nothing on it, so the whole judgement can be tested without a
// filesystem, a clock or a gadget. poll owns the bookkeeping that follows.
//
// hidFaultSince is when the HID endpoints started refusing reports, from
// Hid.linkFaultSince. It is what resolves the ambiguity the link cannot: a
// wedged gadget and a healthy one in a switched-off computer both read "not
// attached", but only one of them has somebody typing into it. Sustained HID
// faults collapse the long wait down to the short one.
func (w *usbWatchdog) decide(now time.Time, hidFaultSince time.Time) usbAction {
	if w.faultSince.IsZero() || w.gaveUp {
		return usbActionNone
	}

	debounce := usbFaultDebounceUnseen
	if w.sawHealthy {
		debounce = usbFaultDebounceSeen
	}
	if !hidFaultSince.IsZero() && now.Sub(hidFaultSince) >= usbHidFaultMinAge {
		debounce = usbFaultDebounceSeen
	}

	if now.Sub(w.faultSince) < debounce {
		return usbActionNone
	}
	if !w.nextAttemptAfter.IsZero() && now.Before(w.nextAttemptAfter) {
		return usbActionNone
	}

	if w.attempt >= usbRecoveryAttempts {
		return usbActionGiveUp
	}

	// Rebind first, twice. It is cheap and it fixes the common case. The
	// rebuild is the last rung, and there is nothing above it: restart_phy can
	// strand a board that has no remote power cycle, so it stays an operator's
	// decision.
	if w.attempt >= usbRecoveryAttempts-1 {
		return usbActionRebuild
	}
	return usbActionRebind
}
