package jiggler

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"NanoKVM-Server/service/controlmode"
	"NanoKVM-Server/service/hid"
	"NanoKVM-Server/service/inputcontrol"
)

// The keys the key jiggler can press.
//
// F15 is the default. No common OS or application binds it, which is why
// keep-awake tools such as Caffeine press it. Shift and Control are offered for
// hosts that ignore F15; both are sent as the left key alone, which does
// nothing by itself.
const (
	KeyF15     = "f15"
	KeyShift   = "shift"
	KeyControl = "ctrl"
)

// DefaultKey is the key used until one is chosen, and in place of a key this
// build does not know.
const DefaultKey = KeyF15

// KeyConfigFile holds the key jiggler's state: "on" or "off" on the first line
// and the key on the second. Unlike the mouse jiggler's file it stays when the
// key jiggler is off, so the chosen key survives a reboot. A missing file is
// off with F15. A variable so tests can point it somewhere writable.
var KeyConfigFile = "/etc/kvm/key-jiggler"

const (
	keyStateOn  = "on"
	keyStateOff = "off"
)

// parseKeyConfig reads KeyConfigFile's contents. Anything but "on" reads as
// off, so a damaged file never starts pressing keys, and an unknown or missing
// key reads as DefaultKey.
func parseKeyConfig(content string) (enabled bool, key string) {
	lines := strings.Split(strings.ReplaceAll(content, "\r", ""), "\n")
	enabled = strings.TrimSpace(lines[0]) == keyStateOn
	key = DefaultKey
	if len(lines) > 1 && ValidKey(strings.TrimSpace(lines[1])) {
		key = strings.TrimSpace(lines[1])
	}
	return enabled, key
}

func formatKeyConfig(enabled bool, key string) string {
	state := keyStateOff
	if enabled {
		state = keyStateOn
	}
	return state + "\n" + key + "\n"
}

// readKeyConfig reads the key jiggler's state at boot. A missing or unreadable
// file is off with the default key.
func readKeyConfig() (enabled bool, key string) {
	content, err := os.ReadFile(KeyConfigFile)
	if err != nil {
		return false, DefaultKey
	}
	return parseKeyConfig(string(content))
}

// SetKeyJiggler turns the key jiggler on or off and sets the key it presses.
// The file is written first, so a key jiggler that reports itself on is one the
// next boot also finds on, with the same key.
func (j *Jiggler) SetKeyJiggler(enabled bool, key string) error {
	if !ValidKey(key) {
		return fmt.Errorf("unknown jiggler key %q", key)
	}

	if err := os.WriteFile(KeyConfigFile, []byte(formatKeyConfig(enabled, key)), 0644); err != nil {
		return err
	}

	j.mutex.Lock()
	j.keyEnabled = enabled
	j.key = key
	j.mutex.Unlock()

	// Outside the lock, because Run takes it. Turning the key jiggler off
	// needs nothing more: the loop stops by itself once neither is on.
	if enabled {
		j.Run()
	}

	return nil
}

// KeyJiggler reports whether the key jiggler is on and the key it presses. The
// key is reported while it is off too.
func (j *Jiggler) KeyJiggler() (enabled bool, key string) {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	return j.keyEnabled, j.key
}

// keyHold is how long the key stays down, long enough for a host polling the
// keyboard to see it and short enough that it never auto-repeats.
const keyHold = 50 * time.Millisecond

// ValidKey reports whether key is one the jiggler can press.
func ValidKey(key string) bool {
	_, ok := keyReport(key)
	return ok
}

// keyReport builds the boot keyboard report that holds key down: modifiers,
// reserved, then six key usages. Shift and Control are modifier bits with no
// usage; F15 is usage 0x6A with no modifier.
func keyReport(key string) ([]byte, bool) {
	report := make([]byte, hid.KeyboardReportLen)
	switch key {
	case KeyF15:
		report[2] = 0x6a
	case KeyShift:
		report[0] = 0x02 // Left Shift
	case KeyControl:
		report[0] = 0x01 // Left Control
	default:
		return nil, false
	}
	return report, true
}

// press sends one press and release of key through the keyboard endpoint,
// gated the same way as the mouse move: it gives way to a control mode change
// and to anyone typing, and skips the press rather than wait for them.
func press(key string) {
	report, ok := keyReport(key)
	if !ok {
		return
	}

	_, releaseMode, err := controlmode.GetManager().AcquireStable()
	if err != nil {
		return
	}
	defer releaseMode()

	ctx, release, err := inputcontrol.GetCoordinator().BeginBackground(context.Background())
	if err != nil {
		return
	}
	defer release()

	_ = pressKey(ctx, hid.GetHid().WriteKeyboardReport, report, keyHold)
}

// pressKey writes report, holds it, and writes an empty report. The release is
// written whatever happened before it, including a press that failed or a hold
// cut short by manual input, and tried a second time if it fails: a key left
// down on the host repeats or turns every click into a Shift-click.
func pressKey(ctx context.Context, write func([]byte) error, report []byte, hold time.Duration) error {
	err := write(report)
	if err == nil {
		_ = waitMove(ctx, hold)
	}

	releaseReport := make([]byte, hid.KeyboardReportLen)
	releaseErr := write(releaseReport)
	if releaseErr != nil {
		releaseErr = write(releaseReport)
	}
	if err == nil {
		err = releaseErr
	}
	return err
}
