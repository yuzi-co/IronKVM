package jiggler

import (
	"context"
	"time"

	"NanoKVM-Server/service/controlmode"
	"NanoKVM-Server/service/hid"
	"NanoKVM-Server/service/inputcontrol"
)

// The keys the jiggler can press instead of moving the mouse.
//
// F15 is the default for the key method. No common OS or application binds it,
// which is why keep-awake tools such as Caffeine press it. Shift and Control
// are offered for hosts that ignore F15; both are sent as the left key alone,
// which does nothing by itself.
const (
	KeyF15     = "f15"
	KeyShift   = "shift"
	KeyControl = "ctrl"
)

// DefaultKey is the key used when the key method is chosen without one.
const DefaultKey = KeyF15

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
