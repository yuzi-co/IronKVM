package hid

import (
	"os"
	"path/filepath"
	"testing"
)

// fakeUDC lays out a sysfs tree the way the gadget framework does, so
// readUSBLink can be driven without a device.
func fakeUDC(t *testing.T, attrs map[string]string) {
	t.Helper()

	dir := filepath.Join(t.TempDir(), "udc")
	if len(attrs) > 0 {
		udc := filepath.Join(dir, "4340000.usb")
		if err := os.MkdirAll(udc, 0o755); err != nil {
			t.Fatalf("build the fake udc tree: %s", err)
		}
		for name, value := range attrs {
			// The kernel's attributes end in a newline. Writing them without
			// one would let a missing TrimSpace pass.
			if err := os.WriteFile(filepath.Join(udc, name), []byte(value+"\n"), 0o644); err != nil {
				t.Fatalf("write %s: %s", name, err)
			}
		}
	} else if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("build the fake udc tree: %s", err)
	}

	previous := udcClassDir
	udcClassDir = dir
	t.Cleanup(func() { udcClassDir = previous })
}

func TestReadUSBLinkReadsBothAttributes(t *testing.T) {
	fakeUDC(t, map[string]string{"state": "configured", "current_speed": "high-speed"})

	link, err := readUSBLink()
	if err != nil {
		t.Fatalf("readUSBLink: %s", err)
	}
	if link.State != "configured" || link.Speed != "high-speed" {
		t.Fatalf("got %+v, want configured/high-speed", link)
	}
	if link.health() != linkHealthy {
		t.Fatalf("configured at high-speed graded %v, want healthy", link.health())
	}
}

// The failure this whole file exists for. A full-speed link is addressed, has
// its descriptors read and its configuration selected, and cannot carry three
// HID interrupt endpoints alongside the console, the disk and the speaker. A
// health test that stops at "configured" reports this board as working.
func TestConfiguredAtFullSpeedIsNotHealthy(t *testing.T) {
	link := usbLink{State: "configured", Speed: "full-speed"}

	if link.health() != linkDegraded {
		t.Fatalf("configured at full-speed graded %v, want degraded", link.health())
	}
	if link.health() == linkHealthy {
		t.Fatal("a full-speed link must never grade healthy")
	}
}

func TestLinkHealthGrades(t *testing.T) {
	cases := []struct {
		state string
		speed string
		want  linkHealth
	}{
		{"configured", "high-speed", linkHealthy},
		{"configured", "full-speed", linkDegraded},
		{"configured", "UNKNOWN", linkDegraded},
		{"configured", "", linkDegraded},
		{"not attached", "UNKNOWN", linkDetached},

		// Everything in the enumeration sequence is in progress, not a fault.
		{"attached", "UNKNOWN", linkPending},
		{"powered", "UNKNOWN", linkPending},
		{"default", "UNKNOWN", linkPending},
		{"addressed", "UNKNOWN", linkPending},

		// A host that goes to sleep suspends the bus. The gadget is fine, and
		// nothing must rebind it for this.
		{"suspended", "high-speed", linkPending},

		{"unavailable", "", linkPending},
	}

	for _, tc := range cases {
		got := usbLink{State: tc.state, Speed: tc.speed}.health()
		if got != tc.want {
			t.Errorf("%q at %q graded %v, want %v", tc.state, tc.speed, got, tc.want)
		}
	}
}

func TestReadUSBLinkWithoutController(t *testing.T) {
	fakeUDC(t, nil)

	if _, err := readUSBLink(); err == nil {
		t.Fatal("an empty udc class must be an error, not a detached link")
	}
}

func TestReadUSBLinkMissingClassDir(t *testing.T) {
	previous := udcClassDir
	udcClassDir = filepath.Join(t.TempDir(), "absent")
	t.Cleanup(func() { udcClassDir = previous })

	if _, err := readUSBLink(); err == nil {
		t.Fatal("a missing udc class dir must be an error")
	}
}

// A controller with no current_speed still grades, and it grades degraded
// rather than healthy: the supervisor asks for a rebind instead of reporting
// health it cannot confirm.
func TestReadUSBLinkWithoutSpeedAttribute(t *testing.T) {
	fakeUDC(t, map[string]string{"state": "configured"})

	link, err := readUSBLink()
	if err != nil {
		t.Fatalf("readUSBLink: %s", err)
	}
	if link.health() != linkDegraded {
		t.Fatalf("graded %v with no speed attribute, want degraded", link.health())
	}
}

func TestDescribeNamesTheTwoFaultsApart(t *testing.T) {
	detached := usbLink{State: "not attached"}.describe()
	degraded := usbLink{State: "configured", Speed: "full-speed"}.describe()

	if detached == degraded {
		t.Fatal("the two faults must not describe identically")
	}
	if detached != "the host has not enumerated the gadget" {
		t.Errorf("detached describes as %q", detached)
	}
	if degraded != "the host enumerated the gadget at full-speed" {
		t.Errorf("degraded describes as %q", degraded)
	}
}
