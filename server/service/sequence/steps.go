package sequence

import (
	"fmt"
	"time"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/hid"
)

const (
	StepKey   = "key"
	StepMouse = "mouse"
	StepDelay = "delay"
	StepATX   = "atx"

	ButtonPower = "power"
	ButtonReset = "reset"
)

const (
	// maxSteps bounds a macro. A recorded key press is four steps, a press,
	// a pause, a release and a pause, so this is some five hundred keys.
	maxSteps = 2000

	// maxDelay bounds one pause. A longer wait is several pauses.
	maxDelay = 60 * time.Second

	// maxATXPress matches the cap POST /api/vm/gpio puts on a press.
	maxATXPress = 10 * time.Second

	defaultATXPress = 800 * time.Millisecond

	maxKeys         = 6
	maxAbsolute     = 0x7fff
	maxRelativeMove = 127
)

// compileSteps checks steps and turns them into reports. It refuses the whole
// list at the first step it cannot play, naming it, rather than playing part
// of a macro.
func compileSteps(steps []proto.SequenceStep) ([]step, error) {
	if len(steps) == 0 {
		return nil, fmt.Errorf("the macro has no steps")
	}
	if len(steps) > maxSteps {
		return nil, fmt.Errorf("the macro has %d steps, at most %d are allowed", len(steps), maxSteps)
	}

	out := make([]step, 0, len(steps))
	for i, s := range steps {
		compiled, err := compileStep(s)
		if err != nil {
			return nil, fmt.Errorf("step %d: %w", i+1, err)
		}
		out = append(out, compiled)
	}
	return out, nil
}

func compileStep(s proto.SequenceStep) (step, error) {
	switch s.Type {
	case StepKey:
		report, err := keyboardReport(s.Modifiers, s.Keys)
		if err != nil {
			return step{}, err
		}
		return step{kind: stepKeyboard, report: report}, nil
	case StepMouse:
		return mouseStep(s)
	case StepDelay:
		if s.Ms < 0 || time.Duration(s.Ms)*time.Millisecond > maxDelay {
			return step{}, fmt.Errorf("a pause must be between 0 and %d ms", maxDelay.Milliseconds())
		}
		return step{kind: stepDelay, wait: time.Duration(s.Ms) * time.Millisecond}, nil
	case StepATX:
		if s.Button != ButtonPower && s.Button != ButtonReset {
			return step{}, fmt.Errorf("unknown ATX button %q", s.Button)
		}
		hold := time.Duration(s.Ms) * time.Millisecond
		if s.Ms == 0 {
			hold = defaultATXPress
		}
		if hold <= 0 || hold > maxATXPress {
			return step{}, fmt.Errorf("an ATX press must be between 1 and %d ms", maxATXPress.Milliseconds())
		}
		return step{kind: stepATX, button: s.Button, wait: hold}, nil
	default:
		return step{}, fmt.Errorf("unknown step type %q", s.Type)
	}
}

// keyboardReport builds a boot keyboard report: modifiers, a reserved byte
// and up to six usages.
func keyboardReport(modifiers int, keys []int) ([]byte, error) {
	if modifiers < 0 || modifiers > 0xff {
		return nil, fmt.Errorf("modifiers %d out of range", modifiers)
	}
	if len(keys) > maxKeys {
		return nil, fmt.Errorf("%d keys at once, at most %d", len(keys), maxKeys)
	}
	report := make([]byte, hid.KeyboardReportLen)
	report[0] = byte(modifiers)
	for i, key := range keys {
		// 0x04 is the first key usage; 0x01 to 0x03 are error codes, and
		// the modifiers, 0xe0 to 0xe7, go in the first byte. The usages
		// above them are what the web UI's key map sends for media keys.
		if key < 0x04 || key > 0xff || (key >= 0xe0 && key <= 0xe7) {
			return nil, fmt.Errorf("key usage %#x out of range", key)
		}
		report[2+i] = byte(key)
	}
	return report, nil
}

func mouseStep(s proto.SequenceStep) (step, error) {
	if s.Buttons < 0 || s.Buttons > 0xff {
		return step{}, fmt.Errorf("buttons %d out of range", s.Buttons)
	}
	if s.Wheel < -maxRelativeMove || s.Wheel > maxRelativeMove ||
		s.HWheel < -maxRelativeMove || s.HWheel > maxRelativeMove {
		return step{}, fmt.Errorf("wheel out of range")
	}

	if s.Absolute {
		if s.X < 0 || s.X > maxAbsolute || s.Y < 0 || s.Y > maxAbsolute {
			return step{}, fmt.Errorf("pointer position %d,%d out of range", s.X, s.Y)
		}
		if s.HWheel != 0 {
			return step{}, fmt.Errorf("the absolute pointer has no horizontal wheel")
		}
		report := []byte{
			byte(s.Buttons),
			byte(s.X), byte(s.X >> 8),
			byte(s.Y), byte(s.Y >> 8),
			byte(int8(s.Wheel)),
		}
		return step{kind: stepAbsolute, report: report}, nil
	}

	if s.X < -maxRelativeMove || s.X > maxRelativeMove || s.Y < -maxRelativeMove || s.Y > maxRelativeMove {
		return step{}, fmt.Errorf("pointer move %d,%d out of range", s.X, s.Y)
	}
	report := []byte{
		byte(s.Buttons),
		byte(int8(s.X)), byte(int8(s.Y)),
		byte(int8(s.Wheel)), byte(int8(s.HWheel)),
	}
	return step{kind: stepRelative, report: report}, nil
}
