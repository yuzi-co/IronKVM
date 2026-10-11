package nut

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"NanoKVM-Server/config"
	"NanoKVM-Server/service/hostpower"
)

// harness is a monitor against a fake NUT server, with a clock the test moves
// and a host the test switches.
type harness struct {
	t   *testing.T
	srv *fakeServer
	now time.Time

	settings config.NUT
	password string

	ledConnected bool
	hostOn       bool
	ledErr       error
	pressErr     error
	presses      []string

	mon *Monitor
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	srv := newFakeServer(t)
	srv.user, srv.password = "kvm", "secret"
	host, port := srv.hostPort()

	h := &harness{
		t:            t,
		srv:          srv,
		now:          time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC),
		password:     "secret",
		ledConnected: true,
		hostOn:       true,
		settings: config.NUT{
			Enabled:             true,
			Host:                host,
			Port:                port,
			UPS:                 "ups",
			Username:            "kvm",
			OnBatteryMinutes:    5,
			ForceOffMinutes:     3,
			PowerOnRestore:      true,
			PowerOnDelaySeconds: 60,
		},
	}
	h.mon = New(Deps{
		Now:      func() time.Time { return h.now },
		Settings: func() config.NUT { return h.settings.WithDefaults() },
		Password: func() (string, error) { return h.password, nil },
		Power: hostpower.New(hostpower.Deps{
			PressButton: func(kind string, d time.Duration) error {
				if h.pressErr != nil {
					return h.pressErr
				}
				h.presses = append(h.presses, fmt.Sprintf("%s %s", kind, d))
				return nil
			},
			PowerLEDConnected: func() bool { return h.ledConnected },
			PowerLED:          func() (bool, error) { return h.hostOn, h.ledErr },
			Sleep:             func(time.Duration) {},
		}),
	})
	return h
}

// tick moves the clock by d, then polls once.
func (h *harness) tick(d time.Duration) {
	h.t.Helper()
	h.now = h.now.Add(d)
	h.mon.tick(context.Background())
}

func (h *harness) pressed() string { return strings.Join(h.presses, ", ") }

func (h *harness) wantPresses(want string) {
	h.t.Helper()
	if got := h.pressed(); got != want {
		h.t.Fatalf("pressed [%s], want [%s]", got, want)
	}
}

func (h *harness) state() State { return h.mon.State() }

const step = PollInterval

func TestItDoesNothingWhileOff(t *testing.T) {
	h := newHarness(t)
	h.settings.Enabled = false
	h.srv.set("OB LB", 5, 30)

	for i := 0; i < 5; i++ {
		h.tick(step)
	}
	if h.srv.acceptCount() != 0 {
		t.Fatal("the client connected while off")
	}
	h.wantPresses("")
	if st := h.state(); st.Status != StatusOff || st.Enabled {
		t.Fatalf("%+v", st)
	}
}

func TestItNeedsAServerAndAUPS(t *testing.T) {
	h := newHarness(t)
	h.settings.Host = ""
	h.tick(step)
	if h.srv.acceptCount() != 0 || h.state().Status != StatusUnconfigured {
		t.Fatalf("%+v", h.state())
	}
}

func TestItReportsTheUPS(t *testing.T) {
	h := newHarness(t)
	h.srv.set("OB DISCHRG", 87, 900)
	h.tick(step)

	st := h.state()
	if st.Status != StatusOK || !st.Reachable || st.UPS != UPSOnBattery ||
		st.Charge == nil || *st.Charge != 87 || st.RuntimeSeconds == nil || *st.RuntimeSeconds != 900 {
		t.Fatalf("%+v", st)
	}
	if st.Host != "on" || !st.LoggedIn {
		t.Fatalf("%+v", st)
	}
	h.wantPresses("")
}

func TestOnBatteryDoesNothingUntilTheTimeIsUp(t *testing.T) {
	h := newHarness(t)
	h.srv.set("OB", 90, 1200)

	h.tick(step)
	h.tick(4 * time.Minute)
	h.wantPresses("")
	if st := h.state(); st.OnBatterySeconds == nil || st.Phase != PhaseIdle {
		t.Fatalf("%+v", st)
	}

	h.tick(time.Minute + step)
	h.wantPresses("power 800ms")
	if st := h.state(); st.Phase != PhaseShuttingDown {
		t.Fatalf("%+v", st)
	}

	// One press per outage.
	h.tick(step)
	h.tick(step)
	h.wantPresses("power 800ms")
}

func TestWithoutATimeOnBatteryNeverShutsDown(t *testing.T) {
	h := newHarness(t)
	h.settings.OnBatteryMinutes = 0
	h.srv.set("OB", 90, 1200)

	h.tick(step)
	h.tick(10 * time.Hour)
	h.wantPresses("")
}

func TestLowBatteryShutsDownAtOnce(t *testing.T) {
	h := newHarness(t)
	h.srv.set("OB LB", 8, 60)

	h.tick(step)
	h.wantPresses("power 800ms")
	if ev := h.state().Events; len(ev) == 0 || ev[0].Kind != EventShutdown || ev[0].Detail != ReasonLowBattery {
		t.Fatalf("%+v", ev)
	}
}

func TestAForcedShutdownShutsDownAtOnce(t *testing.T) {
	h := newHarness(t)
	h.srv.set("OB FSD", 50, 600)

	h.tick(step)
	h.wantPresses("power 800ms")
}

// The primary waits for secondaries that are logged in. The client stays
// logged in while the host shuts down, and leaves once its LED is off, and
// does not log in again for the rest of that shutdown.
func TestItLogsOutOnceTheHostIsOff(t *testing.T) {
	h := newHarness(t)
	h.tick(step)
	waitFor(t, "the login", func() bool { return h.srv.loggedIn() == 1 })

	h.srv.set("OB FSD", 50, 600)
	h.tick(step)
	h.wantPresses("power 800ms")
	h.tick(step)
	if h.srv.loggedIn() != 1 {
		t.Fatal("logged out while the host is still on")
	}

	h.hostOn = false
	h.tick(step)
	waitFor(t, "the logout", func() bool { return h.srv.loggedIn() == 0 })
	if !h.srv.saw("LOGOUT") {
		t.Fatal("no LOGOUT")
	}

	connections := h.srv.acceptCount()
	h.tick(step)
	h.tick(step)
	if h.srv.loggedIn() != 0 {
		t.Fatal("logged in again during the shutdown")
	}
	if h.srv.acceptCount() == connections {
		t.Fatal("the client stopped polling after it logged out")
	}
	if st := h.state(); st.Status != StatusOK || st.LoggedIn || st.Phase != PhaseWaitingMains {
		t.Fatalf("%+v", st)
	}
}

func TestAHostThatIsAlreadyOffIsNotWaitedFor(t *testing.T) {
	h := newHarness(t)
	h.hostOn = false
	h.tick(step)
	waitFor(t, "the login", func() bool { return h.srv.loggedIn() == 1 })

	h.srv.set("OB LB", 10, 60)
	h.tick(step)
	h.wantPresses("")
	waitFor(t, "the logout", func() bool { return h.srv.loggedIn() == 0 })
}

func TestARefusedLoginStillReads(t *testing.T) {
	h := newHarness(t)
	h.password = "wrong"
	h.srv.set("OB LB", 10, 60)

	h.tick(step)
	h.wantPresses("power 800ms")
	if st := h.state(); st.LoggedIn || st.LoginError != "ACCESS-DENIED" || st.Status != StatusOK {
		t.Fatalf("%+v", st)
	}
}

func TestWithoutCredentialsItReadsOnly(t *testing.T) {
	h := newHarness(t)
	h.settings.Username = ""
	h.tick(step)
	if h.srv.saw("LOGIN") || h.srv.saw("USERNAME") {
		t.Fatal("it logged in without credentials")
	}
	if st := h.state(); st.Status != StatusOK || st.LoggedIn {
		t.Fatalf("%+v", st)
	}
}

func TestAHostThatStaysOnGetsALongPress(t *testing.T) {
	h := newHarness(t)
	h.srv.set("OB FSD", 50, 600)
	h.tick(step)
	h.wantPresses("power 800ms")

	h.tick(2 * time.Minute)
	h.wantPresses("power 800ms")
	h.tick(time.Minute)
	h.wantPresses("power 800ms, power 5s")

	// Once only.
	h.tick(time.Minute)
	h.wantPresses("power 800ms, power 5s")
}

func TestNoLongPressWithoutATimeout(t *testing.T) {
	h := newHarness(t)
	h.settings.ForceOffMinutes = 0
	h.srv.set("OB FSD", 50, 600)
	h.tick(step)
	h.tick(time.Hour)
	h.wantPresses("power 800ms")
}

func TestNoLongPressOnceTheHostIsOff(t *testing.T) {
	h := newHarness(t)
	h.srv.set("OB FSD", 50, 600)
	h.tick(step)
	h.hostOn = false
	h.tick(step)
	h.tick(time.Hour)
	h.wantPresses("power 800ms")
}

func TestNoLongPressWhenMainsIsBackBeforeTheTimeout(t *testing.T) {
	h := newHarness(t)
	h.srv.set("OB LB", 8, 60)
	h.tick(step)
	h.srv.set("OL", 8, 60)
	h.tick(10 * time.Minute)
	h.wantPresses("power 800ms")
}

func TestItPowersOnWhenMainsIsBack(t *testing.T) {
	h := newHarness(t)
	h.srv.set("OB FSD", 50, 600)
	h.tick(step)
	h.hostOn = false
	h.tick(step)
	h.wantPresses("power 800ms")

	// Mains returns. The delay passes first.
	h.srv.set("OL CHRG", 55, 600)
	h.tick(step)
	h.wantPresses("power 800ms")
	if st := h.state(); st.PowerOnInSeconds == nil {
		t.Fatalf("%+v", st)
	}
	h.tick(30 * time.Second)
	h.wantPresses("power 800ms")
	h.tick(31 * time.Second)
	h.wantPresses("power 800ms, power 800ms")
	if ev := h.state().Events; ev[0].Kind != EventPowerOn {
		t.Fatalf("%+v", ev)
	}

	// Once.
	h.tick(step)
	h.tick(time.Hour)
	h.wantPresses("power 800ms, power 800ms")
}

func TestItWaitsForTheMinimumCharge(t *testing.T) {
	h := newHarness(t)
	h.settings.PowerOnMinCharge = 30
	h.srv.set("OB LB", 8, 60)
	h.tick(step)
	h.hostOn = false
	h.tick(step)

	h.srv.set("OL CHRG", 12, 60)
	h.tick(step)
	h.tick(5 * time.Minute)
	h.wantPresses("power 800ms")

	h.srv.set("OL CHRG", 31, 600)
	h.tick(step)
	h.tick(time.Minute + step)
	h.wantPresses("power 800ms, power 800ms")
}

func TestTheDelayStartsOverWhenMainsDropsAgain(t *testing.T) {
	h := newHarness(t)
	h.srv.set("OB LB", 8, 60)
	h.tick(step)
	h.hostOn = false
	h.tick(step)

	h.srv.set("OL", 8, 60)
	h.tick(step)
	h.tick(50 * time.Second)
	h.srv.set("OB", 8, 60)
	h.tick(step)
	h.srv.set("OL", 8, 60)
	h.tick(step)
	h.tick(50 * time.Second)
	h.wantPresses("power 800ms")
	h.tick(11 * time.Second)
	h.wantPresses("power 800ms, power 800ms")
}

// Mains returns while the UPS still has its forced shutdown set, or its low
// battery flag: that is not "mains is back".
func TestItDoesNotPowerOnWhileTheUPSStillSaysShutdown(t *testing.T) {
	h := newHarness(t)
	h.srv.set("OB FSD", 50, 600)
	h.tick(step)
	h.hostOn = false
	h.tick(step)

	h.srv.set("OL FSD", 50, 600)
	h.tick(step)
	h.tick(10 * time.Minute)
	h.wantPresses("power 800ms")
}

// The power race: mains returns before the UPS cuts its output, so the host
// is down but the board never lost power. The host went off without a press
// from here (the UPS ran out, or the host's own NUT shut it down), and the
// client switches it on.
func TestThePowerRaceIsStillPoweredOn(t *testing.T) {
	h := newHarness(t)
	h.srv.set("OB", 70, 1000)
	h.tick(step)
	h.wantPresses("")

	h.hostOn = false
	h.tick(step)
	h.wantPresses("")

	h.srv.set("OL CHRG", 70, 1000)
	h.tick(step)
	h.tick(61 * time.Second)
	h.wantPresses("power 800ms")
}

// A host that is off by choice while mains is fine stays off.
func TestAHostSwitchedOffByHandStaysOff(t *testing.T) {
	h := newHarness(t)
	h.tick(step)
	h.hostOn = false
	h.tick(step)
	h.tick(time.Hour)
	h.wantPresses("")
}

// An outage that never takes the host down owes it nothing, so a host switched
// off by hand afterwards stays off.
func TestAnOutageTheHostSurvivedOwesNothing(t *testing.T) {
	h := newHarness(t)
	h.srv.set("OB", 90, 1200)
	h.tick(step)
	h.srv.set("OL", 90, 1200)
	h.tick(step)
	h.hostOn = false
	h.tick(step)
	h.tick(time.Hour)
	h.wantPresses("")
}

// The BIOS restored power by itself: the LED is on again, and the client does
// not press, which would shut the host down.
func TestAHostThatCameBackByItselfIsLeftAlone(t *testing.T) {
	h := newHarness(t)
	h.srv.set("OB LB", 8, 60)
	h.tick(step)
	h.hostOn = false
	h.tick(step)

	h.srv.set("OL", 8, 60)
	h.tick(step)
	h.hostOn = true
	h.tick(61 * time.Second)
	h.tick(step)
	h.wantPresses("power 800ms")
}

func TestNoPowerOnWhenTheOptionIsOff(t *testing.T) {
	h := newHarness(t)
	h.settings.PowerOnRestore = false
	h.srv.set("OB LB", 8, 60)
	h.tick(step)
	h.hostOn = false
	h.tick(step)
	h.srv.set("OL", 8, 60)
	h.tick(step)
	h.tick(time.Hour)
	h.wantPresses("power 800ms")
}

func TestItNeverPressesWithoutThePowerLED(t *testing.T) {
	h := newHarness(t)
	h.ledConnected = false
	h.srv.set("OB LB", 8, 60)
	h.tick(step)
	h.tick(time.Hour)
	h.wantPresses("")
	if st := h.state(); st.Phase != PhaseNeedsLED || st.Host != "unknown" {
		t.Fatalf("%+v", st)
	}

	h.ledConnected, h.ledErr = true, errors.New("gpio")
	h.tick(step)
	h.wantPresses("")
}

func TestAFailedPressIsRetriedAndLoggedOnce(t *testing.T) {
	h := newHarness(t)
	h.pressErr = errors.New("gpio busy")
	h.srv.set("OB LB", 8, 60)

	h.tick(step)
	h.tick(step)
	h.tick(step)
	failed := 0
	for _, e := range h.state().Events {
		if e.Kind == EventPressFailed {
			failed++
		}
	}
	if failed != 1 {
		t.Fatalf("%d failures logged", failed)
	}

	h.pressErr = nil
	h.tick(step)
	h.wantPresses("power 800ms")
}

func TestAnUnreachableServerIsReportedAndNeverActedOn(t *testing.T) {
	h := newHarness(t)
	h.tick(step)
	if st := h.state(); st.Status != StatusOK {
		t.Fatalf("%+v", st)
	}

	h.srv.close()
	h.tick(step)
	h.tick(step)
	if st := h.state(); st.Status == StatusUnreachable {
		t.Fatal("unreachable after two failures")
	}
	h.tick(step)
	st := h.state()
	if st.Status != StatusUnreachable || st.Reachable || st.UPS != UPSUnknown || st.Charge != nil {
		t.Fatalf("%+v", st)
	}
	if st.Events[0].Kind != EventUnreachable {
		t.Fatalf("%+v", st.Events)
	}

	h.tick(time.Hour)
	h.wantPresses("")
	// The event is logged once, not on every poll.
	n := 0
	for _, e := range h.state().Events {
		if e.Kind == EventUnreachable {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("%d unreachable events", n)
	}
}

// The NUT server is the host: it goes down with the host, and the board sees
// mains return only when the server does.
func TestItPowersOnAfterTheServerComesBack(t *testing.T) {
	h := newHarness(t)
	h.srv.set("OB", 60, 900)
	h.tick(step)

	// The host dies, and its NUT server with it.
	h.srv.close()
	h.hostOn = false
	for i := 0; i < 4; i++ {
		h.tick(step)
	}
	h.wantPresses("")
	if h.state().Status != StatusUnreachable {
		t.Fatal("not unreachable")
	}

	// It comes back (the BIOS or a timer started it), serving mains again.
	// The LED is still off here: the host is not up yet.
	srv := newFakeServer(t)
	srv.user, srv.password = "kvm", "secret"
	srv.set("OL", 60, 900)
	h.settings.Host, h.settings.Port = srv.hostPort()
	h.mon.Reconnect()
	h.tick(step)
	h.tick(61 * time.Second)
	h.wantPresses("power 800ms")
	if ev := h.state().Events; ev[0].Kind != EventPowerOn {
		t.Fatalf("%+v", ev)
	}
}

func TestAShutdownInProgressSurvivesAnUnreachableServer(t *testing.T) {
	h := newHarness(t)
	h.srv.set("OB FSD", 50, 600)
	h.tick(step)
	h.wantPresses("power 800ms")

	h.srv.close()
	for i := 0; i < 4; i++ {
		h.tick(step)
	}
	// Still on after the timeout: the long press does not need the server.
	h.tick(4 * time.Minute)
	h.wantPresses("power 800ms, power 5s")
}

func TestChangedSettingsReconnect(t *testing.T) {
	h := newHarness(t)
	h.tick(step)
	n := h.srv.acceptCount()
	h.tick(step)
	if h.srv.acceptCount() != n {
		t.Fatal("it reconnected on every poll")
	}
	h.mon.Reconnect()
	h.tick(step)
	if h.srv.acceptCount() != n+1 {
		t.Fatal("it did not reconnect")
	}
}

func TestTurningItOffMidOutageForgetsTheOutage(t *testing.T) {
	h := newHarness(t)
	h.srv.set("OB LB", 8, 60)
	h.tick(step)
	h.hostOn = false
	h.settings.Enabled = false
	h.tick(step)
	h.settings.Enabled = true
	h.srv.set("OL", 8, 60)
	h.tick(step)
	h.tick(time.Hour)
	h.wantPresses("power 800ms")
}
