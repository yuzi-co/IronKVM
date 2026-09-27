package hid

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// recordWrites captures every report that reaches the wire.
func recordWrites(t *testing.T) *[][]byte {
	t.Helper()

	restore := writeReport
	t.Cleanup(func() { writeReport = restore })

	var got [][]byte
	writeReport = func(_ string, _ *os.File, data []byte, _ time.Duration) error {
		got = append(got, append([]byte(nil), data...))
		return nil
	}
	return &got
}

// gadgetReportLength points the configfs read at a scratch file holding value,
// or at a missing file when value is empty.
func gadgetReportLength(t *testing.T, value string) {
	t.Helper()

	restore := absoluteReportLengthPath
	t.Cleanup(func() { absoluteReportLengthPath = restore })

	path := filepath.Join(t.TempDir(), "report_length")
	if value != "" {
		if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	absoluteReportLengthPath = path
}

func TestReadAbsoluteReportID(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  byte
	}{
		{"7\n", AbsolutePointerReportID},
		{"7", AbsolutePointerReportID},
		{"6\n", 0},
		{"", 0},
		{"garbage\n", 0},
	} {
		gadgetReportLength(t, tc.value)
		if got := readAbsoluteReportID(); got != tc.want {
			t.Errorf("report_length %q: got ID %d, want %d", tc.value, got, tc.want)
		}
	}
}

func TestOpeningThePointerReadsTheReportID(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  byte
	}{
		{"7\n", AbsolutePointerReportID},
		{"6\n", 0},
		{"", 0},
	} {
		gadgetReportLength(t, tc.value)
		h := &Hid{absReportID: 0xee}

		node := filepath.Join(t.TempDir(), "hidg2")
		if err := os.WriteFile(node, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := h.openDeviceNoLock(h.absoluteMouseDevice(node)); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { h.closeDeviceNoLock(h.absoluteMouseDevice(node)) })

		if h.absReportID != tc.want {
			t.Errorf("report_length %q: absReportID %d, want %d", tc.value, h.absReportID, tc.want)
		}
	}
}

func TestOpeningTheKeyDeviceAlsoReadsTheReportID(t *testing.T) {
	gadgetReportLength(t, "7\n")
	h := &Hid{}

	node := filepath.Join(t.TempDir(), "hidg2")
	if err := os.WriteFile(node, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	device := h.extendedKeyDevice()
	device.path = node
	if err := h.openDeviceNoLock(device); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { h.closeDeviceNoLock(device) })

	if h.absReportID != AbsolutePointerReportID {
		t.Fatalf("absReportID %d after opening through the key device, want %d", h.absReportID, AbsolutePointerReportID)
	}
}

func TestPointerReportsGetTheIDWhenTheGadgetHasIDs(t *testing.T) {
	got := recordWrites(t)
	h := &Hid{absReportID: AbsolutePointerReportID}
	openDevice(t, h.absoluteMouseDevice(HID2))

	report := []byte{0x01, 0x34, 0x12, 0x78, 0x56, 0xff}
	if err := h.WriteAbsoluteMouseReport(report); err != nil {
		t.Fatal(err)
	}

	want := []byte{AbsolutePointerReportID, 0x01, 0x34, 0x12, 0x78, 0x56, 0xff}
	if len(*got) != 1 || !bytes.Equal((*got)[0], want) {
		t.Fatalf("wrote %x, want one report %x", *got, want)
	}
}

func TestPointerReportsStayPlainWithoutIDs(t *testing.T) {
	got := recordWrites(t)
	h := &Hid{}
	openDevice(t, h.absoluteMouseDevice(HID2))

	report := []byte{0x01, 0x34, 0x12, 0x78, 0x56, 0xff}
	if err := h.WriteAbsoluteMouseReport(report); err != nil {
		t.Fatal(err)
	}

	if len(*got) != 1 || !bytes.Equal((*got)[0], report) {
		t.Fatalf("wrote %x, want one report %x", *got, report)
	}
}

// The pointer path already allocates once per report for reasons of its own,
// so this holds the prefix to adding nothing on top of that.
func TestThePrefixDoesNotAllocate(t *testing.T) {
	restore := writeReport
	t.Cleanup(func() { writeReport = restore })
	writeReport = func(string, *os.File, []byte, time.Duration) error { return nil }

	report := []byte{0, 0, 0, 0, 0, 0}
	allocs := func(id byte) float64 {
		h := &Hid{}
		openDevice(t, h.absoluteMouseDevice(HID2))
		h.absReportID = id
		return testing.AllocsPerRun(200, func() {
			_ = h.WriteAbsoluteMouseReport(report)
		})
	}

	plain := allocs(0)
	prefixed := allocs(AbsolutePointerReportID)
	if prefixed > plain {
		t.Fatalf("a prefixed pointer report allocates %.0f times, a plain one %.0f", prefixed, plain)
	}
}

// S03usbdev stop_start can change report_length under an open handle: it keeps
// the function directories, so /dev/hidg2 is never deleted and the handle
// stays valid. The ID read at open time is then stale. On 2026-09-26 the
// server kept sending plain reports to a gadget that declared IDs, and the
// host dropped every pointer report.
func TestAChangedReportLengthDropsThePointerHandle(t *testing.T) {
	gadgetReportLength(t, "6\n")
	h := &Hid{}

	node := filepath.Join(t.TempDir(), "hidg2")
	if err := os.WriteFile(node, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	device := h.absoluteMouseDevice(node)
	if err := h.openDeviceNoLock(device); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { h.closeDeviceNoLock(device) })

	if h.RefreshAbsoluteReportID() {
		t.Fatal("an unchanged report_length dropped the handle")
	}
	if h.g2 == nil {
		t.Fatal("an unchanged report_length closed the handle")
	}

	if err := os.WriteFile(absoluteReportLengthPath, []byte("7\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !h.RefreshAbsoluteReportID() {
		t.Fatal("a changed report_length was not noticed")
	}
	if h.g2 != nil {
		t.Fatal("the handle survived a changed report_length, so the stale ID stays in use")
	}

	if err := h.openDeviceNoLock(device); err != nil {
		t.Fatal(err)
	}
	if h.absReportID != AbsolutePointerReportID {
		t.Fatalf("absReportID %d after the reopen, want %d", h.absReportID, AbsolutePointerReportID)
	}
}

func TestRefreshLeavesAClosedPointerAlone(t *testing.T) {
	gadgetReportLength(t, "7\n")
	h := &Hid{}
	if h.RefreshAbsoluteReportID() {
		t.Fatal("refresh acted with no handle open")
	}
}
