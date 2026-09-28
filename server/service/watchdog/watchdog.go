// Package watchdog resets or power-cycles a hung host. It watches the
// picture, the HDMI signal, the power LED and, optionally, a ping of the host,
// and presses a front-panel button once the host has shown no sign of life for
// the configured time. It is off by default. See
// docs/superpowers/specs/2026-09-28-host-watchdog-design.md.
package watchdog

import (
	"context"
	"crypto/sha256"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
)

// SampleInterval is how often the watchdog looks at the host. Each sample
// encodes one JPEG, so this is the watchdog's whole capture cost.
const SampleInterval = 10 * time.Second

// The actions, which are also the values of config.Watchdog.Action.
const (
	ActionReset = "reset"
	ActionPower = "power"
)

// Why the watchdog acted.
const (
	// ReasonFrozen means the picture did not change.
	ReasonFrozen = "frozen"
	// ReasonNoSignal means the HDMI signal was gone at the time of the
	// action.
	ReasonNoSignal = "noSignal"
)

// Status says what the watchdog is doing, for the page.
const (
	StatusOff        = "off"
	StatusWatching   = "watching"
	StatusHostOff    = "hostOff"
	StatusCaptureOff = "captureOff"
	StatusCooldown   = "cooldown"
	StatusCapped     = "capped"
	StatusActing     = "acting"
)

// The presses. A reset is the UI's default press. A power cycle holds power
// long enough to force a hung host off, as Redfish's ForceOff does, waits, and
// presses again to turn it on: a short press alone does nothing to a host
// that no longer answers ACPI.
const (
	resetPress    = 800 * time.Millisecond
	forceOffPress = 5 * time.Second
	powerOffPause = 5 * time.Second
	powerOnPress  = 800 * time.Millisecond
)

// Capture qualities. The sample only has to be compared, so it takes the
// lowest the capture path allows; the screenshot is for a person to read.
const (
	sampleQuality     = 51
	screenshotQuality = 60
)

// ErrFrameUnchanged is what Deps.Capture returns when the capture library says
// the picture has not changed since its last frame. The library says so when
// a viewer has MJPEG frame detection on.
var ErrFrameUnchanged = errors.New("frame not changed")

// Deps is everything the watchdog reaches outside this package. The router
// wires the real ones; tests pass fakes.
type Deps struct {
	Now   func() time.Time
	Sleep func(time.Duration)

	// Settings returns the current settings with the defaults filled in.
	Settings func() config.Watchdog
	// SetSettings saves the settings and applies them.
	SetSettings func(config.Watchdog) error

	// CaptureDisabled reports that the owner turned HDMI capture off.
	CaptureDisabled func() bool
	// Signal reports whether an HDMI signal is present.
	Signal func() bool
	// Capture takes one JPEG at the stream's size and the given quality.
	Capture func(ctx context.Context, quality int) ([]byte, error)

	// PowerLEDConnected reports whether the owner has said the LED header
	// is wired. When it is not, the LED is never read.
	PowerLEDConnected func() bool
	// PowerLED reports whether the host's power LED is lit.
	PowerLED func() (bool, error)

	// Ping reports whether host answered one echo request.
	Ping func(ctx context.Context, host string) bool

	// PressButton holds "power" or "reset" for d.
	PressButton func(kind string, d time.Duration) error

	Log *Log
}

// Watchdog watches the host and acts when it hangs.
type Watchdog struct {
	deps      Deps
	startOnce sync.Once

	mu sync.Mutex
	// enabled is whether the last sample found the watchdog on, so turning
	// it on restarts the timer.
	enabled bool
	status  string
	// lastAlive is the last sign of life: a picture change, a ping reply, a
	// sample that could not decide, or a reason to start the timer again.
	lastAlive time.Time
	// lastChange is the last picture change seen, and lastPing the last ping
	// reply. Both are zero until one is seen.
	lastChange   time.Time
	lastPing     time.Time
	lastSample   time.Time
	frame        [sha256.Size]byte
	haveFrame    bool
	signal       bool
	ledConnected bool
	ledOn        bool
	pingOK       bool
	lastAction   time.Time
	// actions are the times of the actions in the last hour.
	actions []time.Time
}

// New returns a watchdog. Start runs it.
func New(deps Deps) *Watchdog {
	if deps.Now == nil {
		deps.Now = time.Now
	}
	if deps.Sleep == nil {
		deps.Sleep = time.Sleep
	}
	return &Watchdog{deps: deps, status: StatusOff}
}

// Start samples the host every SampleInterval for the rest of uptime. Calling
// it again does nothing.
func (w *Watchdog) Start() {
	w.startOnce.Do(func() {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("watchdog panicked, no longer watching the host: %v", r)
				}
			}()

			ticker := time.NewTicker(SampleInterval)
			defer ticker.Stop()
			for range ticker.C {
				w.tick(context.Background())
			}
		}()
	})
}

// rearm starts the timer again from now: the host gets the whole timeout
// before the watchdog may act. A change of settings calls it.
func (w *Watchdog) rearm() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.rearmLocked(w.deps.Now())
}

func (w *Watchdog) rearmLocked(now time.Time) {
	w.lastAlive = now
	w.haveFrame = false
}

// tick takes one sample and acts if the host is hung.
func (w *Watchdog) tick(ctx context.Context) {
	now := w.deps.Now()
	s := w.deps.Settings()

	if !s.Enabled {
		w.mu.Lock()
		w.enabled = false
		w.status = StatusOff
		w.mu.Unlock()
		return
	}

	w.mu.Lock()
	if !w.enabled {
		w.enabled = true
		w.rearmLocked(now)
	}
	w.mu.Unlock()

	// If the LED says the host is off, or cannot say, the host is not
	// "should be up", and nothing the picture does means a hang.
	ledConnected := w.deps.PowerLEDConnected()
	ledOn := false
	if ledConnected {
		on, err := w.deps.PowerLED()
		if err != nil || !on {
			w.pause(now, StatusHostOff, ledConnected, false)
			return
		}
		ledOn = true
	}

	// With capture turned off there is no picture and no signal to judge.
	if w.deps.CaptureDisabled() {
		w.pause(now, StatusCaptureOff, ledConnected, ledOn)
		return
	}

	// The sample comes first: it resumes capture if the idle timer stopped
	// it, and only then is the signal state current.
	sampleCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	jpeg, sampleErr := w.deps.Capture(sampleCtx, sampleQuality)
	cancel()
	signal := w.deps.Signal()

	pingOK := false
	if s.PingHost != "" {
		pingOK = w.deps.Ping(ctx, s.PingHost)
	}

	w.mu.Lock()
	w.ledConnected, w.ledOn = ledConnected, ledOn
	w.observeLocked(now, signal, jpeg, sampleErr, pingOK)
	act, reason, stuck := w.decideLocked(now, s)
	if act {
		w.status = StatusActing
	}
	w.mu.Unlock()

	if act {
		w.act(ctx, now, actionFor(s.Action, ledConnected), reason, stuck)
	}
}

// actionFor is the action the watchdog may take. SetSettings refuses a power
// cycle without the power LED, but the LED setting can be turned off later.
// Without it a host shut down on purpose looks hung, and a power cycle would
// switch it back on, so the watchdog resets instead: a reset does nothing to a
// host that is off.
func actionFor(action string, ledConnected bool) string {
	if action == ActionPower && !ledConnected {
		log.Warnf("watchdog: the power LED is not connected, resetting instead of a power cycle")
		return ActionReset
	}
	return action
}

// pause records that the watchdog cannot judge the host now, and starts the
// timer again so the host gets the whole timeout once it can.
func (w *Watchdog) pause(now time.Time, status string, ledConnected, ledOn bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.status = status
	w.ledConnected, w.ledOn = ledConnected, ledOn
	w.rearmLocked(now)
}

// observeLocked folds one sample into the state.
func (w *Watchdog) observeLocked(now time.Time, signal bool, jpeg []byte, sampleErr error, pingOK bool) {
	w.lastSample = now
	w.signal = signal
	w.pingOK = pingOK

	if pingOK {
		w.lastPing = now
		w.lastAlive = now
	}

	switch {
	case !signal:
		// No signal is not a sign of life. The return of the signal is,
		// because the next frame has nothing to compare with.
		w.haveFrame = false
	case errors.Is(sampleErr, ErrFrameUnchanged):
		// The library compared the frames itself.
	case sampleErr != nil || len(jpeg) == 0:
		// A failed capture proves nothing either way, and a false reset
		// costs more than a missed one.
		w.lastAlive = now
	default:
		// The input is digital and the encoder deterministic, so a still
		// picture gives the same bytes.
		sum := sha256.Sum256(jpeg)
		if !w.haveFrame || sum != w.frame {
			w.frame = sum
			w.haveFrame = true
			w.lastChange = now
			w.lastAlive = now
		}
	}
}

// decideLocked reports whether to act now, why, and for how long the host has
// shown no sign of life.
func (w *Watchdog) decideLocked(now time.Time, s config.Watchdog) (bool, string, time.Duration) {
	stuck := now.Sub(w.lastAlive)
	if stuck < time.Duration(s.TimeoutMinutes)*time.Minute {
		w.status = StatusWatching
		return false, "", stuck
	}

	if !w.lastAction.IsZero() && now.Sub(w.lastAction) < time.Duration(s.CooldownMinutes)*time.Minute {
		w.status = StatusCooldown
		return false, "", stuck
	}

	w.pruneActionsLocked(now)
	if len(w.actions) >= s.MaxPerHour {
		w.status = StatusCapped
		return false, "", stuck
	}

	reason := ReasonFrozen
	if !w.signal {
		reason = ReasonNoSignal
	}
	return true, reason, stuck
}

// pruneActionsLocked forgets the actions older than an hour.
func (w *Watchdog) pruneActionsLocked(now time.Time) {
	kept := w.actions[:0]
	for _, t := range w.actions {
		if now.Sub(t) < time.Hour {
			kept = append(kept, t)
		}
	}
	w.actions = kept
}

// act saves a screenshot, presses, and logs the action.
func (w *Watchdog) act(ctx context.Context, now time.Time, action, reason string, stuck time.Duration) {
	log.Warnf("watchdog: the host shows no sign of life for %s (%s), pressing %s",
		stuck.Round(time.Second), reason, action)

	var screenshot []byte
	if reason != ReasonNoSignal {
		shotCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		jpeg, err := w.deps.Capture(shotCtx, screenshotQuality)
		cancel()
		if err != nil {
			log.Warnf("watchdog: no screenshot before the press: %s", err)
		} else {
			screenshot = jpeg
		}
	}

	pressErr := w.press(action)
	if pressErr != nil {
		log.Errorf("watchdog: press %s: %s", action, pressErr)
	}
	countAction(action)

	entry := Entry{
		ID:           strconv.FormatInt(now.UnixNano(), 10),
		Time:         now.UTC(),
		Action:       action,
		Reason:       reason,
		StuckSeconds: int(stuck / time.Second),
	}
	if pressErr != nil {
		entry.Error = pressErr.Error()
	}
	if w.deps.Log != nil {
		if err := w.deps.Log.Add(entry, screenshot); err != nil {
			log.Errorf("watchdog: save the action log: %s", err)
		}
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.lastAction = now
	w.actions = append(w.actions, now)
	// The host gets the whole timeout to come back up.
	w.rearmLocked(w.deps.Now())
	w.status = StatusWatching
}

// press carries out the action.
func (w *Watchdog) press(action string) error {
	if action != ActionPower {
		return w.deps.PressButton(ActionReset, resetPress)
	}

	if err := w.deps.PressButton(ActionPower, forceOffPress); err != nil {
		return err
	}
	w.deps.Sleep(powerOffPause)
	return w.deps.PressButton(ActionPower, powerOnPress)
}

// The actions taken since the server started, for /api/metrics.
var resets, powerCycles atomic.Uint64

func countAction(action string) {
	if action == ActionPower {
		powerCycles.Add(1)
		return
	}
	resets.Add(1)
}

// ActionCounts returns how many resets and power cycles the watchdog has
// pressed since the server started.
func ActionCounts() (reset, power uint64) {
	return resets.Load(), powerCycles.Load()
}
