package hostpower

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

type fake struct {
	connected bool
	on        bool
	ledErr    error
	presses   []string
	sleeps    []time.Duration
}

func (f *fake) power() *Power {
	return New(Deps{
		PressButton: func(kind string, d time.Duration) error {
			f.presses = append(f.presses, fmt.Sprintf("%s %s", kind, d))
			return nil
		},
		PowerLEDConnected: func() bool { return f.connected },
		PowerLED:          func() (bool, error) { return f.on, f.ledErr },
		Sleep:             func(d time.Duration) { f.sleeps = append(f.sleeps, d) },
	})
}

func TestStateFollowsTheLED(t *testing.T) {
	f := &fake{connected: true, on: true}
	p := f.power()
	if got := p.State(); got != On {
		t.Fatalf("lit LED: %v", got)
	}
	f.on = false
	if got := p.State(); got != Off {
		t.Fatalf("dark LED: %v", got)
	}
	f.ledErr = errors.New("gpio")
	if got := p.State(); got != Unknown {
		t.Fatalf("unreadable LED: %v", got)
	}
	f.ledErr, f.connected = nil, false
	if got := p.State(); got != Unknown {
		t.Fatalf("unwired LED: %v", got)
	}
}

func TestShutdownPressesOnlyAHostThatIsOn(t *testing.T) {
	f := &fake{connected: true, on: true}
	pressed, err := f.power().Shutdown()
	if err != nil || !pressed || len(f.presses) != 1 || f.presses[0] != "power 800ms" {
		t.Fatalf("on: %v %v %v", pressed, err, f.presses)
	}

	f = &fake{connected: true, on: false}
	pressed, err = f.power().Shutdown()
	if err != nil || pressed || len(f.presses) != 0 {
		t.Fatalf("off: %v %v %v", pressed, err, f.presses)
	}
}

func TestPowerOnPressesOnlyAHostThatIsOff(t *testing.T) {
	f := &fake{connected: true, on: true}
	pressed, err := f.power().PowerOn()
	if err != nil || pressed || len(f.presses) != 0 {
		t.Fatalf("on: %v %v %v", pressed, err, f.presses)
	}

	f.on = false
	pressed, err = f.power().PowerOn()
	if err != nil || !pressed || len(f.presses) != 1 {
		t.Fatalf("off: %v %v %v", pressed, err, f.presses)
	}
}

func TestForceOffIsALongPress(t *testing.T) {
	f := &fake{connected: true, on: true}
	if _, err := f.power().ForceOff(); err != nil || f.presses[0] != "power 5s" {
		t.Fatalf("%v %v", err, f.presses)
	}
}

func TestNothingPressesBlind(t *testing.T) {
	f := &fake{connected: false}
	p := f.power()
	for name, fn := range map[string]func() (bool, error){
		"shutdown": p.Shutdown, "forceoff": p.ForceOff, "poweron": p.PowerOn,
	} {
		pressed, err := fn()
		if pressed || !errors.Is(err, ErrUnknownState) {
			t.Fatalf("%s: %v %v", name, pressed, err)
		}
	}
	if len(f.presses) != 0 {
		t.Fatalf("pressed: %v", f.presses)
	}
}

func TestResetIsAShortPressOfTheResetButton(t *testing.T) {
	f := &fake{}
	if err := f.power().Reset(); err != nil || len(f.presses) != 1 || f.presses[0] != "reset 800ms" {
		t.Fatalf("%v %v", err, f.presses)
	}
}

func TestCycleHoldsPausesAndPresses(t *testing.T) {
	f := &fake{}
	if err := f.power().Cycle(5 * time.Second); err != nil {
		t.Fatal(err)
	}
	if len(f.presses) != 2 || f.presses[0] != "power 5s" || f.presses[1] != "power 800ms" {
		t.Fatalf("%v", f.presses)
	}
	if len(f.sleeps) != 1 || f.sleeps[0] != 5*time.Second {
		t.Fatalf("%v", f.sleeps)
	}
}
