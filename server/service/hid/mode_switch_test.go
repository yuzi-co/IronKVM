package hid

import (
	"errors"
	"strings"
	"testing"
)

// switchGadget is the whole of what replaced the reboot, so these cover the
// cases that decide whether an operator is told the truth about which mode the
// board is in.

func TestSwitchGadgetRebuildsTheGadgetOnce(t *testing.T) {
	var actions []string

	err := switchGadget(ModeHidOnly,
		func(action string) error {
			actions = append(actions, action)
			return nil
		},
		func() (string, error) { return ModeHidOnly, nil },
	)
	if err != nil {
		t.Fatalf("switchGadget returned %v, want nil", err)
	}

	// stop_start is what unbinds the gadget before the script reconfigures it.
	// A plain start would leave it bound, and configfs would force an unbind
	// underneath the HID unlink rather than the script asking for one.
	if len(actions) != 1 || actions[0] != "stop_start" {
		t.Fatalf("ran %v, want exactly one stop_start", actions)
	}
}

func TestSwitchGadgetReportsARebuildThatFailed(t *testing.T) {
	boom := errors.New("script exited 1")
	read := false

	err := switchGadget(ModeNormal,
		func(string) error { return boom },
		func() (string, error) {
			read = true
			return ModeNormal, nil
		},
	)
	if !errors.Is(err, boom) {
		t.Fatalf("switchGadget returned %v, want it to wrap %v", err, boom)
	}

	// Reading the mode back after a failed rebuild would report whatever the
	// previous mode left behind, which is how a failure turns into a success.
	if read {
		t.Fatal("switchGadget read the mode back after the rebuild failed")
	}
}

func TestSwitchGadgetFailsWhenTheModeDidNotChange(t *testing.T) {
	// The failure this catches: the script ran and exited 0, but the gadget
	// still reports the mode it started in. That happens whenever the HID
	// functions were not unlinked first, because f_hid then refuses every
	// attribute write and the descriptors never move.
	err := switchGadget(ModeHidOnly,
		func(string) error { return nil },
		func() (string, error) { return ModeNormal, nil },
	)
	if err == nil {
		t.Fatal("switchGadget accepted a gadget that stayed in the old mode")
	}
	if !strings.Contains(err.Error(), ModeNormal) || !strings.Contains(err.Error(), ModeHidOnly) {
		t.Fatalf("switchGadget said %q, want it to name both modes", err)
	}
}

func TestSwitchGadgetReportsAModeItCannotRead(t *testing.T) {
	boom := errors.New("no such file")

	err := switchGadget(ModeNormal,
		func(string) error { return nil },
		func() (string, error) { return "", boom },
	)
	if !errors.Is(err, boom) {
		t.Fatalf("switchGadget returned %v, want it to wrap %v", err, boom)
	}
}

func TestSwitchGadgetAcceptsAModeThatMatches(t *testing.T) {
	for _, mode := range []string{ModeNormal, ModeHidOnly} {
		err := switchGadget(mode,
			func(string) error { return nil },
			func() (string, error) { return mode, nil },
		)
		if err != nil {
			t.Fatalf("switchGadget(%s) returned %v, want nil", mode, err)
		}
	}
}
