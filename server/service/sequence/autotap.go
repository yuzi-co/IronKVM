package sequence

import (
	"fmt"
	"time"

	"NanoKVM-Server/proto"
)

const (
	defaultTapInterval = 150 * time.Millisecond
	minTapInterval     = 50 * time.Millisecond
	maxTapInterval     = 5 * time.Second

	defaultTapTimeout = 60 * time.Second
	maxTapTimeout     = 10 * time.Minute

	// tapHold is how long each press holds the key. Firmware polls a USB
	// keyboard every few milliseconds, so this is seen, and it stays well
	// below every auto-repeat delay, so a press types one key and no more.
	tapHold = 40 * time.Millisecond
)

// autoTapProgram turns a request into a run: an optional ATX press, then
// press, hold, release and wait, until a stop condition. Each press is
// released inside the loop, not at the end, so the key is up between taps
// whatever ends the run.
func autoTapProgram(req proto.AutoTapReq) (program, error) {
	if len(req.Keys) == 0 && req.Modifiers == 0 {
		return program{}, fmt.Errorf("no key to tap")
	}
	down, err := keyboardReport(req.Modifiers, req.Keys)
	if err != nil {
		return program{}, err
	}
	up, _ := keyboardReport(0, nil)

	interval := defaultTapInterval
	if req.IntervalMs != 0 {
		interval = time.Duration(req.IntervalMs) * time.Millisecond
		if interval < minTapInterval || interval > maxTapInterval {
			return program{}, fmt.Errorf("the interval must be between %d and %d ms",
				minTapInterval.Milliseconds(), maxTapInterval.Milliseconds())
		}
	}

	timeout := defaultTapTimeout
	if req.TimeoutS != 0 {
		timeout = time.Duration(req.TimeoutS) * time.Second
		if timeout <= 0 || timeout > maxTapTimeout {
			return program{}, fmt.Errorf("the timeout must be between 1 and %d s", int(maxTapTimeout.Seconds()))
		}
	}

	hold := min(tapHold, interval/2)
	p := program{
		kind:    KindAutoTap,
		name:    tapName(req.Label, req.Modifiers, req.Keys),
		timeout: timeout,
		loop: []step{
			{kind: stepKeyboard, report: down},
			{kind: stepDelay, wait: hold},
			{kind: stepKeyboard, report: up},
			{kind: stepDelay, wait: interval - hold},
		},
		stopOnInput: req.StopOnInput,
		stopOnVideo: req.StopOnVideo,
	}

	if req.ATX != "" {
		press, err := compileStep(proto.SequenceStep{Type: StepATX, Button: req.ATX, Ms: req.ATXMs})
		if err != nil {
			return program{}, err
		}
		p.steps = []step{press}
	}
	return p, nil
}

// tapName is the name the status shows: the label the client sent, or the
// usages when it sent none, so a client that did not start the run can still
// say what it taps.
func tapName(label string, modifiers int, keys []int) string {
	if label != "" {
		return truncate(label, maxNameLen)
	}
	name := ""
	if modifiers != 0 {
		name = fmt.Sprintf("mod:%#02x", modifiers)
	}
	for _, key := range keys {
		if name != "" {
			name += "+"
		}
		name += fmt.Sprintf("%#02x", key)
	}
	return name
}
