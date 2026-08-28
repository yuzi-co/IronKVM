package hid

import (
	"errors"
	"os"
	"sync"
	"syscall"
	"time"

	"NanoKVM-Server/proto"
)

// A HID endpoint can fail in a way that produces no error the operator can act
// on. The gadget driver only completes a write once the target fetches the
// report, so an interface the target is not polling swallows every write until
// the deadline expires. Nothing is broken on this side: the device node is
// present, the gadget is bound, and the USB link is configured.
//
// It happens per endpoint. On one observed target the absolute mouse stopped
// being fetched for forty minutes while the keyboard and the relative mouse kept
// working, so the pointer did not move and nothing anywhere said why. A USB
// re-enumeration cleared it; relative mouse mode also worked throughout. Which
// of the two an operator should reach for depends on the target, and the server
// cannot tell - but it can say which endpoint stopped, which is what neither the
// log nor the web UI did before.
//
// Do not read a stall as a diagnosis of the target. It says one thing only: the
// reports are not being collected.
const (
	hidStateUnknown   = "unknown"   // nothing has been written yet
	hidStateAccepting = "accepting" // the target is fetching reports
	hidStateStalled   = "stalled"   // writes time out: the target is not fetching
	hidStateDetached  = "detached"  // the gadget is not enumerated: reports go nowhere
	hidStateError     = "error"     // the write failed some other way
)

// endpointHealth remembers the last outcome of a write to one HID endpoint.
// The zero value is usable and means "nothing written yet".
type endpointHealth struct {
	mu       sync.Mutex
	state    string
	detail   string
	since    time.Time // when the current state began
	observed time.Time // when the current state was last confirmed
}

// hidTransition describes what one write did to an endpoint's state. Callers
// log on a change and stay quiet otherwise, so From matters as much as To: the
// first successful write is a change, and it is not a recovery.
type hidTransition struct {
	Changed bool
	From    string
	To      string
}

// record takes the outcome of one write and reports the transition it caused.
func (h *endpointHealth) record(err error, now time.Time) hidTransition {
	state, detail := classifyWriteResult(err)

	h.mu.Lock()
	defer h.mu.Unlock()

	from := h.state
	if from == "" {
		from = hidStateUnknown
	}

	h.observed = now
	if h.state == state && h.detail == detail {
		return hidTransition{Changed: false, From: from, To: state}
	}

	h.state = state
	h.detail = detail
	h.since = now
	return hidTransition{Changed: true, From: from, To: state}
}

// classifyWriteResult separates the three ways a report can fail to arrive,
// because they are three different faults and only one of them is repairable
// from this side.
//
// A deadline is back-pressure from a live host. The link is up and the target
// is simply not fetching from this one endpoint, so it stays "stalled": the
// operator's answer is a different mouse mode or a reset, and nothing automatic
// should touch the gadget for it.
//
// ESHUTDOWN is the opposite. f_hid returns it when the function is disabled,
// which means the gadget is not enumerated and every report to every endpoint
// goes nowhere. ENODEV is the same fact arriving as the node disappearing under
// a write, which is what a rebind racing a keystroke looks like. Those two are
// the ones the supervisor in usb_watchdog.go reads, and they are why this is
// not one undifferentiated "error" any more.
func classifyWriteResult(err error) (state string, detail string) {
	switch {
	case err == nil:
		return hidStateAccepting, ""
	case errors.Is(err, os.ErrDeadlineExceeded):
		return hidStateStalled, ""
	case errors.Is(err, syscall.ESHUTDOWN), errors.Is(err, syscall.ENODEV):
		return hidStateDetached, err.Error()
	default:
		return hidStateError, err.Error()
	}
}

// linkFaultSince reports how long this endpoint has been failing with an error
// that means the gadget is not enumerated, and the zero time when it is not.
func (h *endpointHealth) linkFaultSince() time.Time {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.state != hidStateDetached {
		return time.Time{}
	}
	return h.since
}

func (h *endpointHealth) snapshot(now time.Time) proto.HidDeviceStatus {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.state == "" {
		return proto.HidDeviceStatus{State: hidStateUnknown}
	}

	return proto.HidDeviceStatus{
		State:         h.state,
		Detail:        h.detail,
		StateForMs:    millisBetween(h.since, now),
		ObservedMsAgo: millisBetween(h.observed, now),
	}
}

func millisBetween(from, to time.Time) int64 {
	elapsed := to.Sub(from)
	if elapsed < 0 {
		return 0
	}
	return elapsed.Milliseconds()
}
