// Package nut is a NUT client for the KVM. It reads a UPS from a NUT server on
// the network and drives the host's ATX buttons from what it finds: a shutdown
// press on low battery, on a forced shutdown or after a time on battery, a long
// press if the host will not go down, and a power-on press when mains is back.
// It is off by default. See
// docs/superpowers/specs/2026-10-11-nut-client-design.md.
package nut

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
	"NanoKVM-Server/service/hostpower"
)

// PollInterval is how often the client reads the UPS. A forced shutdown is
// the primary's call to cut the output, so the faster the KVM sees it the
// better; one LIST VAR is a few hundred bytes.
const PollInterval = 5 * time.Second

const (
	dialTimeout = 3 * time.Second
	// unreachableAfter is how many polls in a row must fail before the server
	// counts as unreachable. One lost packet is not an outage.
	unreachableAfter = 3
	// shutdownGrace is how long a host that was asked to shut down may stay on
	// before the client stops expecting it to go off. After that, a host the
	// owner switches off by hand is not switched on again.
	shutdownGrace = 15 * time.Minute
	maxEvents     = 20
)

// The reasons for a shutdown.
const (
	ReasonFSD        = "fsd"
	ReasonLowBattery = "lowBattery"
	ReasonOnBattery  = "onBattery"
)

// Status says how the client stands, for the page.
const (
	StatusOff          = "off"
	StatusUnconfigured = "unconfigured"
	StatusConnecting   = "connecting"
	StatusOK           = "ok"
	StatusUnreachable  = "unreachable"
)

// The UPS states the page shows, from ups.status.
const (
	UPSUnknown    = "unknown"
	UPSOnline     = "online"
	UPSOnBattery  = "onBattery"
	UPSLowBattery = "lowBattery"
	UPSShutdown   = "shutdown"
)

// Phase says what the client is doing about the host.
const (
	PhaseIdle         = "idle"
	PhaseShuttingDown = "shuttingDown"
	PhaseWaitingMains = "waitingForMains"
	PhaseNeedsLED     = "needsLed"
)

// Event kinds.
const (
	EventUnreachable = "unreachable"
	EventReachable   = "reachable"
	EventOnBattery   = "onBattery"
	EventMainsBack   = "mainsBack"
	EventShutdown    = "shutdown"
	EventForceOff    = "forceOff"
	EventHostOff     = "hostOff"
	EventPowerOn     = "powerOn"
	EventPressFailed = "pressFailed"
)

// Event is something the client did or saw, for the page and the log.
type Event struct {
	Time   time.Time `json:"time"`
	Kind   string    `json:"kind"`
	Detail string    `json:"detail,omitempty"`
}

// Deps is everything the client reaches outside this package. The router wires
// the real ones; tests pass fakes. The NUT server itself is reached over TCP,
// and the tests run a fake one.
type Deps struct {
	Now func() time.Time

	// Settings returns the current settings with the defaults filled in.
	Settings func() config.NUT
	// SetSettings saves the settings and applies them.
	SetSettings func(config.NUT) error
	// Password returns the NUT password, empty when there is none.
	Password func() (string, error)
	// SetPassword stores the password, or removes it when empty.
	SetPassword func(string) error

	// Power presses the host's buttons, behind the power LED check.
	Power *hostpower.Power
}

// Reading is what the UPS reports.
type Reading struct {
	Flags []string
	// Charge is battery.charge in percent, and Runtime battery.runtime in
	// seconds. Both are -1 when the UPS does not report them.
	Charge  float64
	Runtime int
}

// Has reports whether ups.status carries flag.
func (r Reading) Has(flag string) bool {
	for _, f := range r.Flags {
		if f == flag {
			return true
		}
	}
	return false
}

// State is the page's word for the UPS.
func (r Reading) State() string {
	switch {
	case r.Has("FSD"):
		return UPSShutdown
	case r.Has("LB"):
		return UPSLowBattery
	case r.Has("OB"):
		return UPSOnBattery
	case r.Has("OL"):
		return UPSOnline
	}
	return UPSUnknown
}

func readingFrom(vars map[string]string) Reading {
	r := Reading{Flags: strings.Fields(vars["ups.status"]), Charge: -1, Runtime: -1}
	if v, err := strconv.ParseFloat(vars["battery.charge"], 64); err == nil && v >= 0 {
		r.Charge = v
	}
	if v, err := strconv.ParseFloat(vars["battery.runtime"], 64); err == nil && v >= 0 {
		r.Runtime = int(v)
	}
	return r
}

// Monitor reads the UPS and acts on the host.
type Monitor struct {
	deps      Deps
	startOnce sync.Once
	// reconnect asks the next tick to drop the connection: the settings or
	// the password changed.
	reconnect atomic.Bool

	mu       sync.Mutex
	status   string
	client   *Client
	loggedIn bool
	// loginErr is the ERR code of a refused login or authentication.
	loginErr string
	// noLogin keeps the client logged out after it logged out on purpose, so
	// a shutdown that is under way is not announced to the primary again.
	noLogin bool

	failures  int
	reachable bool
	lastPoll  time.Time
	reading   *Reading
	host      hostpower.State

	onBatterySince time.Time

	// shutdown is the reason of the shutdown under way, empty when none.
	shutdown        string
	shutdownAt      time.Time
	shutdownHostOff bool
	forced          bool

	// restore is set when an outage took, or may have taken, the host down,
	// and the client owes it a power-on. sawOff says the host has been seen off
	// since; a host that never went off is not switched on.
	restore bool
	// restoreFrom is when the debt began, and restoreByPress says a shutdown
	// press from here began it, not just an outage seen.
	restoreFrom    time.Time
	restoreByPress bool
	sawOff         bool
	powerOnSince   time.Time

	ledWarned bool
	events    []Event
}

// New returns a monitor. Start runs it.
func New(deps Deps) *Monitor {
	if deps.Now == nil {
		deps.Now = time.Now
	}
	return &Monitor{deps: deps, status: StatusOff, host: hostpower.Unknown}
}

// Start polls every PollInterval for the rest of uptime. Calling it again
// does nothing.
func (m *Monitor) Start() {
	m.startOnce.Do(func() {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("nut: panicked, no longer watching the UPS: %v", r)
				}
			}()

			ticker := time.NewTicker(PollInterval)
			defer ticker.Stop()
			for range ticker.C {
				m.tick(context.Background())
			}
		}()
	})
}

// Reconnect drops the connection at the next tick, so new settings or a new
// password take effect.
func (m *Monitor) Reconnect() { m.reconnect.Store(true) }

// tick reads the UPS once and acts on it.
func (m *Monitor) tick(ctx context.Context) {
	now := m.deps.Now()
	s := m.deps.Settings()

	if m.reconnect.Swap(false) {
		m.dropClient()
	}

	if !s.Enabled {
		m.dropClient()
		m.mu.Lock()
		m.resetLocked(StatusOff)
		m.mu.Unlock()
		return
	}
	if s.Host == "" || s.UPS == "" {
		m.dropClient()
		m.mu.Lock()
		m.resetLocked(StatusUnconfigured)
		m.mu.Unlock()
		return
	}

	reading, err := m.poll(ctx, s)

	m.mu.Lock()
	m.lastPoll = now
	if err != nil {
		m.failures++
		if m.failures == unreachableAfter {
			m.reachable = false
			m.status = StatusUnreachable
			log.Warnf("nut: %s unreachable: %s", s.Host, err)
			m.addEventLocked(now, EventUnreachable, err.Error())
		} else if m.failures < unreachableAfter && m.reading == nil {
			m.status = StatusConnecting
		}
		m.mu.Unlock()
		m.step(now, s, nil)
		return
	}
	if !m.reachable && m.failures >= unreachableAfter {
		m.addEventLocked(now, EventReachable, "")
		log.Infof("nut: %s reachable again", s.Host)
	}
	m.failures = 0
	m.reachable = true
	m.status = StatusOK
	m.reading = &reading
	m.mu.Unlock()

	m.step(now, s, &reading)
}

// resetLocked forgets what was learned about the UPS and the outage, for a
// client that is off or not set up. The events stay.
func (m *Monitor) resetLocked(status string) {
	m.status = status
	m.failures, m.reachable = 0, false
	m.lastPoll = time.Time{}
	m.reading = nil
	m.host = hostpower.Unknown
	m.onBatterySince = time.Time{}
	m.noLogin = false
	m.shutdown, m.shutdownHostOff, m.forced = "", false, false
	m.restore, m.restoreByPress, m.sawOff, m.powerOnSince = false, false, false, time.Time{}
	m.ledWarned = false
}

func (m *Monitor) dropClient() {
	m.mu.Lock()
	c := m.client
	m.client = nil
	m.loggedIn = false
	m.loginErr = ""
	m.mu.Unlock()
	if c != nil {
		_ = c.Close()
	}
}

// poll reads the UPS, connecting first when it has to. A failed read drops the
// connection, so the next poll starts again.
func (m *Monitor) poll(ctx context.Context, s config.NUT) (Reading, error) {
	m.mu.Lock()
	c := m.client
	m.mu.Unlock()

	if c == nil {
		var err error
		if c, err = m.connect(ctx, s); err != nil {
			return Reading{}, err
		}
	}

	vars, err := c.Vars(s.UPS)
	if err != nil {
		m.dropClient()
		return Reading{}, err
	}
	return readingFrom(vars), nil
}

// connect dials the server. With a user and a password, and unless a shutdown
// has logged the client out, it logs in as a client of the UPS: that is what a
// primary waits for. A refused login leaves a connection that can still read.
func (m *Monitor) connect(ctx context.Context, s config.NUT) (*Client, error) {
	addr := net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
	c, err := Dial(ctx, addr, dialTimeout)
	if err != nil {
		return nil, err
	}

	password := ""
	if s.Username != "" && m.deps.Password != nil {
		if password, err = m.deps.Password(); err != nil {
			log.Errorf("nut: read the password: %s", err)
		}
	}

	m.mu.Lock()
	noLogin := m.noLogin
	m.mu.Unlock()

	loggedIn, loginErr := false, ""
	if s.Username != "" && password != "" && !noLogin {
		err := c.Authenticate(s.Username, password)
		if err == nil {
			err = c.Login(s.UPS)
		}
		var se *ServerError
		switch {
		case err == nil:
			loggedIn = true
		case errors.As(err, &se):
			loginErr = se.Code
			log.Warnf("nut: login refused: %s", se.Code)
		default:
			_ = c.Close()
			return nil, err
		}
	}

	m.mu.Lock()
	m.client, m.loggedIn, m.loginErr = c, loggedIn, loginErr
	m.mu.Unlock()
	return c, nil
}

// step decides from one reading what to do with the host. reading is nil when
// the server could not be read: what was decided before still stands, but
// nothing new is decided from a UPS that cannot be seen.
func (m *Monitor) step(now time.Time, s config.NUT, reading *Reading) {
	host := m.deps.Power.State()

	m.mu.Lock()
	logout := m.stepLocked(now, s, reading, host)
	m.mu.Unlock()

	// The goodbye is one write with a short deadline, kept out of the lock so
	// that the page never waits for it.
	if logout != nil {
		_ = logout.Logout()
	}
}

func (m *Monitor) stepLocked(now time.Time, s config.NUT, reading *Reading, host hostpower.State) (logout *Client) {
	m.host = host
	if host == hostpower.Unknown {
		if !m.ledWarned {
			log.Warnf("nut: the host's power LED is not connected or cannot be read, so the client does not press")
			m.ledWarned = true
		}
		return nil
	}
	m.ledWarned = false

	if host == hostpower.Off {
		m.sawOff = true
	}

	reason := ""
	if reading != nil {
		reason = m.reasonLocked(now, s, reading)
	}

	// A shutdown ends when its reason is gone: mains is back and the UPS has
	// no forced shutdown. A host that is still shutting down is left to finish.
	if reading != nil && m.shutdown != "" && reason == "" {
		m.shutdown, m.shutdownHostOff, m.forced = "", false, false
		m.noLogin = false
	}

	switch {
	case reading != nil && reason != "" && m.shutdown == "":
		logout = m.startShutdownLocked(now, reason, host)
	case m.shutdown != "":
		logout = m.followShutdownLocked(now, s, host)
	}

	if reading != nil {
		m.restoreLocked(now, s, reading, host)
	}
	return logout
}

// reasonLocked says why the host should be shut down now, empty when it
// should not. It also keeps the clock of the time on battery.
func (m *Monitor) reasonLocked(now time.Time, s config.NUT, r *Reading) string {
	if r.Has("OB") {
		if m.onBatterySince.IsZero() {
			m.onBatterySince = now
			m.addEventLocked(now, EventOnBattery, "")
			log.Warnf("nut: the UPS is on battery")
		}
	} else if !m.onBatterySince.IsZero() {
		m.onBatterySince = time.Time{}
		m.addEventLocked(now, EventMainsBack, "")
		log.Infof("nut: mains is back")
	}

	switch {
	case r.Has("FSD"):
		return ReasonFSD
	case r.Has("LB"):
		return ReasonLowBattery
	case s.OnBatteryMinutes > 0 && !m.onBatterySince.IsZero() &&
		now.Sub(m.onBatterySince) >= time.Duration(s.OnBatteryMinutes)*time.Minute:
		return ReasonOnBattery
	}
	return ""
}

// leaveLocked takes the connection out of use for a goodbye. The client stays
// logged out until the shutdown's reason is gone, so a primary that is waiting
// for its secondaries is not made to wait for the KVM again.
func (m *Monitor) leaveLocked() *Client {
	c := m.client
	m.client, m.loggedIn, m.noLogin = nil, false, true
	return c
}

func (m *Monitor) startShutdownLocked(now time.Time, reason string, host hostpower.State) *Client {
	if host == hostpower.Off {
		// Nothing to shut down, and nothing to wait for: the primary should
		// not wait for this client either.
		m.shutdown, m.shutdownAt, m.shutdownHostOff, m.forced = reason, now, true, false
		return m.leaveLocked()
	}

	pressed, err := m.deps.Power.Shutdown()
	if err != nil {
		m.pressFailedLocked(now, "shutdown", err)
		return nil
	}
	if !pressed {
		return nil
	}

	log.Warnf("nut: shutting the host down (%s)", reason)
	m.addEventLocked(now, EventShutdown, reason)
	m.shutdown, m.shutdownAt = reason, now
	m.shutdownHostOff, m.forced = false, false
	m.restore, m.restoreFrom, m.restoreByPress, m.sawOff = true, now, true, false
	return nil
}

// followShutdownLocked watches a shutdown under way: it logs out once the host
// is off, and presses power for five seconds if the host stays on too long.
func (m *Monitor) followShutdownLocked(now time.Time, s config.NUT, host hostpower.State) *Client {
	if host == hostpower.Off {
		if m.shutdownHostOff {
			return nil
		}
		m.shutdownHostOff = true
		m.addEventLocked(now, EventHostOff, "")
		log.Infof("nut: the host is off")
		return m.leaveLocked()
	}

	if s.ForceOffMinutes <= 0 || m.forced || m.shutdownHostOff ||
		now.Sub(m.shutdownAt) < time.Duration(s.ForceOffMinutes)*time.Minute {
		return nil
	}
	pressed, err := m.deps.Power.ForceOff()
	if err != nil {
		m.pressFailedLocked(now, "forceOff", err)
		return nil
	}
	m.forced = true
	if pressed {
		log.Warnf("nut: the host is still on %d min after the shutdown press, forcing it off", s.ForceOffMinutes)
		m.addEventLocked(now, EventForceOff, "")
	}
	return nil
}

// restoreLocked switches the host on when mains is back. The press needs, all
// at once: an outage that took the host down (or that began with it running),
// mains back with no low battery and no forced shutdown, the host off, enough
// charge, and the delay passed. The delay is what lets a UPS that cuts its
// output after a forced shutdown finish doing so, and a mains return before
// the cut is the power race: the host is down, the UPS never lost power, and
// "restore on AC power" in the BIOS never fires. The press is checked against
// the LED again at the moment of pressing, so a host that came back by itself
// is not sent a shutdown request.
func (m *Monitor) restoreLocked(now time.Time, s config.NUT, r *Reading, host hostpower.State) {
	if !s.PowerOnRestore {
		m.restore, m.powerOnSince = false, time.Time{}
		return
	}

	online := r.Has("OL") && !r.Has("OB") && !r.Has("LB") && !r.Has("FSD")

	// An outage that finds the host running may take it down without a
	// press from here: the UPS runs out, or the host dies. Remember it.
	if r.Has("OB") && host == hostpower.On && !m.restore {
		m.restore, m.restoreFrom, m.restoreByPress, m.sawOff = true, now, false, false
	}
	if !m.restore {
		return
	}

	// The host is on and mains is back. After an outage that never took it
	// down, nothing is owed. After a shutdown press it may still be shutting
	// down, and is owed a power-on once it is off; but a host that went off
	// and came back by itself, or that ignored the press for a long time, is
	// not.
	if host == hostpower.On && online &&
		(!m.restoreByPress || m.sawOff || now.Sub(m.restoreFrom) > shutdownGrace) {
		m.restore, m.powerOnSince = false, time.Time{}
		return
	}

	chargeOK := s.PowerOnMinCharge <= 0 || (r.Charge >= 0 && r.Charge >= float64(s.PowerOnMinCharge))
	if !online || host != hostpower.Off || !chargeOK {
		m.powerOnSince = time.Time{}
		return
	}

	if m.powerOnSince.IsZero() {
		m.powerOnSince = now
	}
	if now.Sub(m.powerOnSince) < time.Duration(s.PowerOnDelaySeconds)*time.Second {
		return
	}

	pressed, err := m.deps.Power.PowerOn()
	if err != nil {
		m.pressFailedLocked(now, "powerOn", err)
		return
	}
	m.restore, m.powerOnSince = false, time.Time{}
	if pressed {
		log.Infof("nut: mains is back, switching the host on")
		m.addEventLocked(now, EventPowerOn, "")
	}
}

func (m *Monitor) pressFailedLocked(now time.Time, what string, err error) {
	log.Errorf("nut: press (%s): %s", what, err)
	detail := what + ": " + err.Error()
	if n := len(m.events); n > 0 && m.events[n-1].Kind == EventPressFailed && m.events[n-1].Detail == detail {
		return
	}
	m.addEventLocked(now, EventPressFailed, detail)
}

func (m *Monitor) addEventLocked(now time.Time, kind, detail string) {
	m.events = append(m.events, Event{Time: now.UTC(), Kind: kind, Detail: detail})
	if len(m.events) > maxEvents {
		m.events = append([]Event(nil), m.events[len(m.events)-maxEvents:]...)
	}
}
