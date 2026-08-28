package hid

// usb_link.go reads what the gadget framework says about the USB link.
//
// The rest of the server tests the gadget by asking whether
// /sys/kernel/config/usb_gadget/g0/UDC is non-empty. That is a binding record.
// It says configfs found a controller to attach to, and it stays populated
// across a link that has since died, so it cannot see the failures this file
// exists to report.
//
// Two attributes carry the truth, and BOTH are needed:
//
//   - state, which reaches "configured" only once the host has accepted the
//     gadget. That is a different fact from the gadget being bound.
//   - current_speed, because reaching "configured" is not enough on this board.
//
// The second one is the part a reader coming from other forks will not expect.
// A NanoKVM that enumerates at full speed is addressed, has its descriptors
// read and its configuration selected, and still does not work: the periodic
// bandwidth in a full-speed frame does not hold three HID interrupt endpoints
// alongside the console, the disk and the speaker. The host schedules what fits
// and silently stops polling the rest. On 2026-08-19 that left this board with
// a working absolute mouse and a dead keyboard, reporting "configured"
// throughout.
//
// S03usbdev's usb_link_ok makes the same two-part test for the boot window.
// This is that test for the rest of uptime. Keep the two in step.

import (
	"os"
	"path/filepath"
	"strings"
)

// udcClassDir is where the gadget framework exposes one directory per device
// controller. A variable so the tests can point it at a fake tree.
var udcClassDir = "/sys/class/udc"

const (
	// udcStateConfigured is the only state in which reports actually go
	// somewhere.
	udcStateConfigured = "configured"

	// udcStateDetached is what a controller with no session reads as. A gadget
	// nobody has plugged in and a gadget whose host went away are the same
	// string, which is the whole difficulty in acting on it.
	udcStateDetached = "not attached"

	// udcSpeedHigh is the only negotiated speed that can carry this gadget.
	udcSpeedHigh = "high-speed"
)

// usbLink is one sample of the controller's own view of the link.
type usbLink struct {
	// State is the raw contents of the state attribute, or "unavailable" when
	// no controller could be read at all.
	State string

	// Speed is the raw contents of current_speed. It is only meaningful while
	// State is configured; before that the controller reports "UNKNOWN".
	Speed string
}

// linkHealth grades one sample. The three outcomes want different responses,
// which is why this is not a boolean.
type linkHealth int

const (
	// linkHealthy: enumerated, and at a speed that can carry the gadget.
	linkHealthy linkHealth = iota

	// linkDetached: no session. Either the host went away, the gadget lost its
	// binding, or nothing was ever plugged in. Those read identically here.
	linkDetached

	// linkDegraded: the host took the gadget at a speed that cannot carry it.
	// Enumerated, addressed, configured, and useless.
	linkDegraded

	// linkPending: mid-enumeration, suspended, or no controller to read. Not
	// health, but not evidence of a fault either. A host that goes to sleep
	// reports "suspended" and is working perfectly.
	linkPending
)

func (l usbLink) health() linkHealth {
	switch l.State {
	case udcStateConfigured:
		if l.Speed == udcSpeedHigh {
			return linkHealthy
		}
		return linkDegraded
	case udcStateDetached:
		return linkDetached
	default:
		return linkPending
	}
}

// describe names the fault in the words an operator needs, because "not
// enumerated" and "enumerated uselessly" send them in different directions.
// S03usbdev's usb_link_fault prints the same two sentences.
func (l usbLink) describe() string {
	switch l.health() {
	case linkHealthy:
		return "the host has the gadget at " + l.Speed
	case linkDetached:
		return "the host has not enumerated the gadget"
	case linkDegraded:
		return "the host enumerated the gadget at " + l.Speed
	default:
		return "the link reads " + l.State
	}
}

// udcReadDir and udcReadFile are indirected so the whole reader can be driven
// off a temporary directory in a test. Nothing else in this package needs to
// stub the filesystem, so they stay unexported and local.
var (
	udcReadDir  = os.ReadDir
	udcReadFile = os.ReadFile
)

// readUSBLink samples the first readable device controller.
//
// The board has exactly one, and S03usbdev's usb_link_speed makes the same
// assumption by taking the first match of a glob. Reading a second one would
// need a rule for which of two disagreeing controllers to believe, and there is
// no such rule to write.
//
// A missing state attribute is an error rather than a fault: a kernel with no
// gadget controller is not a link this can repair, and grading it as detached
// would have the supervisor rebinding forever against nothing.
func readUSBLink() (usbLink, error) {
	entries, err := udcReadDir(udcClassDir)
	if err != nil {
		return usbLink{}, err
	}

	var firstErr error
	for _, entry := range entries {
		state, err := readUDCAttr(entry.Name(), "state")
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		// current_speed is allowed to be missing. The state alone still grades
		// the link as detached or pending, and only the configured case needs
		// the speed. An empty speed there grades degraded, which is the safe
		// way round: it asks for a rebind rather than reporting health it
		// cannot confirm.
		speed, _ := readUDCAttr(entry.Name(), "current_speed")

		return usbLink{State: state, Speed: speed}, nil
	}

	if firstErr == nil {
		firstErr = os.ErrNotExist
	}
	return usbLink{}, firstErr
}

func readUDCAttr(udc string, attr string) (string, error) {
	data, err := udcReadFile(filepath.Join(udcClassDir, udc, attr))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
