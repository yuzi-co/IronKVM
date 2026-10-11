// Package hostpower is the one place that presses the host's ATX buttons on
// behalf of a service that acts alone: the host watchdog and the NUT client.
// The press lengths live here, and so does the power LED check, which keeps a
// short press from switching on a host that is off. A person at the UI, Redfish
// and IPMI press through vm.PressButton directly and decide for themselves.
package hostpower

import (
	"errors"
	"time"
)

// The press lengths. A short press is the UI's default press: ACPI treats it
// as a request to shut down, or to start. A long press holds power long enough
// to force a hung host off, as Redfish's ForceOff does.
const (
	ShortPress = 800 * time.Millisecond
	LongPress  = 5 * time.Second
)

// The buttons PressButton accepts. They are the names service/vm uses.
const (
	ButtonPower = "power"
	ButtonReset = "reset"
)

// State is what the power LED says about the host.
type State int

const (
	// Unknown means the LED is not wired, or could not be read.
	Unknown State = iota
	Off
	On
)

func (s State) String() string {
	switch s {
	case On:
		return "on"
	case Off:
		return "off"
	}
	return "unknown"
}

// ErrUnknownState is what the checked presses return when the LED cannot say
// whether the host is on. A short press on a host that is off would switch it
// on, so nothing presses blind.
var ErrUnknownState = errors.New("the host's power state is unknown: the power LED is not connected or cannot be read")

// Deps is what the helper reaches outside this package. The router wires the
// real ones; tests pass fakes.
type Deps struct {
	// PressButton holds "power" or "reset" for d.
	PressButton func(kind string, d time.Duration) error
	// PowerLEDConnected reports whether the owner has said the LED header is
	// wired. When it is not, the LED is never read.
	PowerLEDConnected func() bool
	// PowerLED reports whether the host's power LED is lit.
	PowerLED func() (bool, error)
	// Sleep defaults to time.Sleep.
	Sleep func(time.Duration)
}

// Power presses the host's buttons and reads its power LED.
type Power struct {
	deps Deps
}

func New(deps Deps) *Power {
	if deps.Sleep == nil {
		deps.Sleep = time.Sleep
	}
	return &Power{deps: deps}
}

// LED reads the power LED. connected is false when the owner has not wired it,
// and then on and err are not meaningful: the LED is not read.
func (p *Power) LED() (connected, on bool, err error) {
	if !p.deps.PowerLEDConnected() {
		return false, false, nil
	}
	on, err = p.deps.PowerLED()
	return true, on, err
}

// State is the host's power state by the LED.
func (p *Power) State() State {
	connected, on, err := p.LED()
	switch {
	case !connected || err != nil:
		return Unknown
	case on:
		return On
	}
	return Off
}

// Reset presses the reset button. A reset does nothing to a host that is off,
// so it needs no check.
func (p *Power) Reset() error {
	return p.deps.PressButton(ButtonReset, ShortPress)
}

// Press holds the power button for d with no check, for the caller that has
// already judged the host, such as the watchdog on a host that looks hung.
func (p *Power) Press(d time.Duration) error {
	return p.deps.PressButton(ButtonPower, d)
}

// Cycle forces the host off with a long press, waits pause, and presses
// briefly to switch it on.
func (p *Power) Cycle(pause time.Duration) error {
	if err := p.Press(LongPress); err != nil {
		return err
	}
	p.deps.Sleep(pause)
	return p.Press(ShortPress)
}

// Shutdown asks a host that is on to shut down, with a short press. It reports
// whether it pressed: a host that is off is left alone.
func (p *Power) Shutdown() (bool, error) { return p.pressIf(On, ShortPress) }

// ForceOff cuts a host that is on with a long press. It reports whether it
// pressed.
func (p *Power) ForceOff() (bool, error) { return p.pressIf(On, LongPress) }

// PowerOn switches on a host that is off, with a short press. It reports
// whether it pressed: a host that is on is left alone, so a host that came
// back by itself is not sent a shutdown request.
func (p *Power) PowerOn() (bool, error) { return p.pressIf(Off, ShortPress) }

func (p *Power) pressIf(want State, d time.Duration) (bool, error) {
	switch p.State() {
	case Unknown:
		return false, ErrUnknownState
	case want:
		return true, p.Press(d)
	}
	return false, nil
}
