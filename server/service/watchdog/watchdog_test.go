package watchdog

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"NanoKVM-Server/config"
)

// fakeHost stands in for the clock, the capture path, the LED, the ping and
// the buttons.
type fakeHost struct {
	mu sync.Mutex

	now      time.Time
	settings config.Watchdog

	captureOff bool
	signal     bool
	// frame is what a capture returns; captureErr fails it instead.
	frame      string
	captureErr error
	captures   []int

	ledConnected bool
	ledOn        bool
	ledErr       error

	pingOK bool
	pings  []string

	presses []string
	sleeps  []time.Duration
}

func newFakeHost() *fakeHost {
	return &fakeHost{
		now: time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC),
		settings: config.Watchdog{
			Enabled: true,
			Action:  ActionReset,
		}.WithDefaults(),
		signal: true,
		frame:  "desktop",
	}
}

func (h *fakeHost) deps(t *testing.T) Deps {
	t.Helper()

	return Deps{
		Now: func() time.Time {
			h.mu.Lock()
			defer h.mu.Unlock()
			return h.now
		},
		Sleep: func(d time.Duration) {
			h.mu.Lock()
			defer h.mu.Unlock()
			h.sleeps = append(h.sleeps, d)
		},
		Settings: func() config.Watchdog {
			h.mu.Lock()
			defer h.mu.Unlock()
			return h.settings
		},
		SetSettings: func(s config.Watchdog) error {
			h.mu.Lock()
			defer h.mu.Unlock()
			h.settings = s.WithDefaults()
			return nil
		},
		CaptureDisabled: func() bool { return h.captureOff },
		Signal:          func() bool { return h.signal },
		Capture: func(_ context.Context, quality int) ([]byte, error) {
			h.mu.Lock()
			defer h.mu.Unlock()
			h.captures = append(h.captures, quality)
			if h.captureErr != nil {
				return nil, h.captureErr
			}
			return []byte(h.frame), nil
		},
		PowerLEDConnected: func() bool { return h.ledConnected },
		PowerLED:          func() (bool, error) { return h.ledOn, h.ledErr },
		Ping: func(_ context.Context, host string) bool {
			h.pings = append(h.pings, host)
			return h.pingOK
		},
		PressButton: func(kind string, d time.Duration) error {
			h.mu.Lock()
			defer h.mu.Unlock()
			h.presses = append(h.presses, fmt.Sprintf("%s %s", kind, d))
			return nil
		},
		Log: OpenLog(t.TempDir()),
	}
}

// run samples every SampleInterval for d, calling change before each sample
// when it is not nil.
func run(w *Watchdog, h *fakeHost, d time.Duration, change func(i int)) {
	for i := 0; time.Duration(i)*SampleInterval <= d; i++ {
		if change != nil {
			change(i)
		}
		w.tick(context.Background())
		h.mu.Lock()
		h.now = h.now.Add(SampleInterval)
		h.mu.Unlock()
	}
}

func (h *fakeHost) pressed() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]string(nil), h.presses...)
}

func TestNothingHappensWhileTheWatchdogIsOff(t *testing.T) {
	h := newFakeHost()
	h.settings.Enabled = false
	w := New(h.deps(t))

	run(w, h, time.Hour, nil)

	if got := h.pressed(); len(got) != 0 {
		t.Fatalf("pressed %v", got)
	}
	// Off costs nothing: no capture and no ping.
	if len(h.captures) != 0 {
		t.Fatalf("captured %d frames while off", len(h.captures))
	}
	if st := w.State(); st.Status != StatusOff {
		t.Fatalf("status %q", st.Status)
	}
}

// Off, the watchdog samples nothing, so what its page shows about the signal
// and the LED must come from the board now, not from the last sample taken
// before it was turned off, or from none at all.
func TestTheStateWhileOffIsReadLive(t *testing.T) {
	h := newFakeHost()
	h.settings.Enabled = false
	h.ledConnected = true
	h.ledOn = true
	w := New(h.deps(t))

	st := w.State()
	if !st.Signal || !st.LEDConnected || !st.LEDOn {
		t.Fatalf("signal %v, LED connected %v, LED on %v, want all true", st.Signal, st.LEDConnected, st.LEDOn)
	}
	if len(h.captures) != 0 {
		t.Fatalf("captured %d frames for the state", len(h.captures))
	}

	h.signal = false
	h.ledErr = errors.New("unreadable")
	if st := w.State(); st.Signal || st.LEDOn {
		t.Fatalf("signal %v, LED on %v, want both false", st.Signal, st.LEDOn)
	}
}

func TestAStillPictureIsResetAfterTheTimeout(t *testing.T) {
	h := newFakeHost()
	w := New(h.deps(t))

	// One sample short of the timeout: nothing yet.
	run(w, h, 5*time.Minute-SampleInterval, nil)
	if got := h.pressed(); len(got) != 0 {
		t.Fatalf("pressed %v before the timeout", got)
	}

	run(w, h, 0, nil)
	got := h.pressed()
	if len(got) != 1 || got[0] != "reset 800ms" {
		t.Fatalf("pressed %v", got)
	}

	entries := w.deps.Log.Entries()
	if len(entries) != 1 {
		t.Fatalf("log has %d entries", len(entries))
	}
	e := entries[0]
	if e.Action != ActionReset || e.Reason != ReasonFrozen || e.StuckSeconds != 300 || e.Error != "" {
		t.Fatalf("entry %+v", e)
	}

	// The screenshot was taken before the press, at the screenshot quality,
	// and is kept beside the log.
	if last := h.captures[len(h.captures)-1]; last != screenshotQuality {
		t.Fatalf("the last capture used quality %d", last)
	}
	path, ok := w.deps.Log.Screenshot(e.ID)
	if !ok {
		t.Fatal("no screenshot for the entry")
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != "desktop" {
		t.Fatalf("screenshot %q, %v", data, err)
	}
}

func TestAChangingPictureIsLeftAlone(t *testing.T) {
	h := newFakeHost()
	w := New(h.deps(t))

	run(w, h, 2*time.Hour, func(i int) { h.frame = fmt.Sprintf("frame %d", i) })

	if got := h.pressed(); len(got) != 0 {
		t.Fatalf("pressed %v", got)
	}
	if st := w.State(); st.Status != StatusWatching || st.LastChange == nil {
		t.Fatalf("state %+v", st)
	}
}

// A picture that changes only now and then still resets the timer each time.
func TestOneChangeRestartsTheTimer(t *testing.T) {
	h := newFakeHost()
	w := New(h.deps(t))

	run(w, h, 4*time.Minute, nil)
	h.frame = "a clock that ticked"
	run(w, h, 4*time.Minute, nil)

	if got := h.pressed(); len(got) != 0 {
		t.Fatalf("pressed %v", got)
	}
}

func TestAHostWhoseLEDIsOffIsLeftAlone(t *testing.T) {
	h := newFakeHost()
	h.ledConnected = true
	h.ledOn = false
	w := New(h.deps(t))

	run(w, h, time.Hour, nil)

	if got := h.pressed(); len(got) != 0 {
		t.Fatalf("pressed %v", got)
	}
	if st := w.State(); st.Status != StatusHostOff {
		t.Fatalf("status %q", st.Status)
	}
}

// An LED that cannot be read says nothing, and nothing is pressed on it.
func TestAnLEDThatCannotBeReadIsLeftAlone(t *testing.T) {
	h := newFakeHost()
	h.ledConnected = true
	h.ledErr = errors.New("no such line")
	w := New(h.deps(t))

	run(w, h, time.Hour, nil)

	if got := h.pressed(); len(got) != 0 {
		t.Fatalf("pressed %v", got)
	}
}

// The host gets the whole timeout after its LED comes on.
func TestTheTimerStartsWhenTheLEDComesOn(t *testing.T) {
	h := newFakeHost()
	h.ledConnected = true
	w := New(h.deps(t))

	run(w, h, 30*time.Minute, nil)
	h.ledOn = true
	run(w, h, 4*time.Minute, nil)
	if got := h.pressed(); len(got) != 0 {
		t.Fatalf("pressed %v within the timeout of the LED coming on", got)
	}

	run(w, h, time.Minute, nil)
	if got := h.pressed(); len(got) != 1 {
		t.Fatalf("pressed %v", got)
	}
}

func TestNoSignalIsResetAfterTheTimeout(t *testing.T) {
	h := newFakeHost()
	h.signal = false
	h.captureErr = errors.New("no HDMI signal")
	w := New(h.deps(t))

	run(w, h, 5*time.Minute, nil)

	if got := h.pressed(); len(got) != 1 {
		t.Fatalf("pressed %v", got)
	}
	e := w.deps.Log.Entries()[0]
	if e.Reason != ReasonNoSignal || e.Screenshot != "" {
		t.Fatalf("entry %+v", e)
	}
}

// With frame detection on, the library itself says the picture is still.
func TestTheLibrarySayingUnchangedCountsAsStill(t *testing.T) {
	h := newFakeHost()
	w := New(h.deps(t))

	run(w, h, time.Minute, nil)
	h.captureErr = fmt.Errorf("capture: %w", ErrFrameUnchanged)
	run(w, h, 4*time.Minute, nil)

	if got := h.pressed(); len(got) != 1 {
		t.Fatalf("pressed %v", got)
	}
}

// A capture that fails with a signal present proves nothing.
func TestFailedCapturesAreLeftAlone(t *testing.T) {
	h := newFakeHost()
	h.captureErr = errors.New("capture is temporarily unavailable")
	w := New(h.deps(t))

	run(w, h, time.Hour, nil)

	if got := h.pressed(); len(got) != 0 {
		t.Fatalf("pressed %v", got)
	}
}

func TestAHostThatAnswersPingIsLeftAlone(t *testing.T) {
	h := newFakeHost()
	h.settings.PingHost = "192.168.1.10"
	h.pingOK = true
	w := New(h.deps(t))

	run(w, h, time.Hour, nil)

	if got := h.pressed(); len(got) != 0 {
		t.Fatalf("pressed %v", got)
	}
	if len(h.pings) == 0 || h.pings[0] != "192.168.1.10" {
		t.Fatalf("pinged %v", h.pings)
	}
}

func TestAHostThatStopsAnsweringPingIsReset(t *testing.T) {
	h := newFakeHost()
	h.settings.PingHost = "192.168.1.10"
	h.pingOK = true
	w := New(h.deps(t))

	run(w, h, 10*time.Minute, nil)
	h.pingOK = false
	run(w, h, 5*time.Minute, nil)

	if got := h.pressed(); len(got) != 1 {
		t.Fatalf("pressed %v", got)
	}
}

func TestNothingIsPressedWithinTheCooldown(t *testing.T) {
	h := newFakeHost()
	h.settings.CooldownMinutes = 20
	w := New(h.deps(t))

	// The first press at 5 minutes. The host stays frozen: the next timeout
	// ends at 10 minutes, inside the cooldown, which ends at 25.
	run(w, h, 24*time.Minute, nil)
	if got := h.pressed(); len(got) != 1 {
		t.Fatalf("pressed %v inside the cooldown", got)
	}
	if st := w.State(); st.Status != StatusCooldown {
		t.Fatalf("status %q", st.Status)
	}

	run(w, h, time.Minute, nil)
	if got := h.pressed(); len(got) != 2 {
		t.Fatalf("pressed %v after the cooldown", got)
	}
}

func TestNothingIsPressedOverTheHourlyCap(t *testing.T) {
	h := newFakeHost()
	h.settings.TimeoutMinutes = 1
	h.settings.CooldownMinutes = 1
	h.settings.MaxPerHour = 3
	w := New(h.deps(t))

	run(w, h, 59*time.Minute, nil)
	if got := h.pressed(); len(got) != 3 {
		t.Fatalf("pressed %d times in an hour", len(got))
	}
	if st := w.State(); st.Status != StatusCapped || st.ActionsLastHour != 3 {
		t.Fatalf("state %+v", st)
	}

	// An hour after the first action, one slot is free again.
	run(w, h, 2*time.Minute, nil)
	if got := h.pressed(); len(got) != 4 {
		t.Fatalf("pressed %d times", len(got))
	}
}

// A power cycle forces the host off, waits, and turns it on.
func TestThePowerActionIsAPowerCycle(t *testing.T) {
	h := newFakeHost()
	h.settings.Action = ActionPower
	h.ledConnected, h.ledOn = true, true
	w := New(h.deps(t))

	run(w, h, 5*time.Minute, nil)

	got := strings.Join(h.pressed(), ", ")
	if got != "power 5s, power 800ms" {
		t.Fatalf("pressed %s", got)
	}
	if len(h.sleeps) != 1 || h.sleeps[0] != powerOffPause {
		t.Fatalf("slept %v", h.sleeps)
	}
	reset, power := ActionCounts()
	if power == 0 {
		t.Fatalf("counted reset=%d power=%d", reset, power)
	}
}

// The LED setting can be turned off after a power cycle was chosen. A host
// shut down on purpose then looks hung, so the watchdog resets instead.
func TestWithoutThePowerLEDAPowerCycleBecomesAReset(t *testing.T) {
	h := newFakeHost()
	h.settings.Action = ActionPower
	w := New(h.deps(t))

	run(w, h, 5*time.Minute, nil)

	got := strings.Join(h.pressed(), ", ")
	if got != "reset "+resetPress.String() {
		t.Fatalf("pressed %s", got)
	}
}

// With capture turned off there is nothing to judge.
func TestNothingHappensWithCaptureOff(t *testing.T) {
	h := newFakeHost()
	h.captureOff = true
	w := New(h.deps(t))

	run(w, h, time.Hour, nil)

	if got := h.pressed(); len(got) != 0 {
		t.Fatalf("pressed %v", got)
	}
	if len(h.captures) != 0 {
		t.Fatalf("captured with capture off")
	}
}

// Turning the watchdog on never acts at once on a host that was frozen before.
func TestTurningTheWatchdogOnStartsTheTimer(t *testing.T) {
	h := newFakeHost()
	h.settings.Enabled = false
	w := New(h.deps(t))

	run(w, h, time.Hour, nil)
	h.settings.Enabled = true
	run(w, h, 5*time.Minute-SampleInterval, nil)

	if got := h.pressed(); len(got) != 0 {
		t.Fatalf("pressed %v", got)
	}
}

// A failed press is logged with its error and counts toward the cap.
func TestAFailedPressIsLogged(t *testing.T) {
	h := newFakeHost()
	deps := h.deps(t)
	deps.PressButton = func(string, time.Duration) error { return errors.New("gpio busy") }
	w := New(deps)

	run(w, h, 5*time.Minute, nil)

	entries := w.deps.Log.Entries()
	if len(entries) != 1 || entries[0].Error != "gpio busy" {
		t.Fatalf("entries %+v", entries)
	}
	if st := w.State(); st.ActionsLastHour != 1 {
		t.Fatalf("state %+v", st)
	}
}
