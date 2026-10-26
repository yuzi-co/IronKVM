package mcpservice

import (
	"reflect"
	"testing"

	"NanoKVM-Server/service/hid"
)

func mustLayout(t *testing.T, id string) *hid.Layout {
	t.Helper()
	layout, err := hid.GetLayout(id)
	if err != nil {
		t.Fatal(err)
	}
	return layout
}

func TestBuildKeyComboReport(t *testing.T) {
	report, err := buildKeyComboReport([]string{"ControlLeft", "AltLeft", "Delete"})
	if err != nil {
		t.Fatal(err)
	}
	if report[0] != 0x05 || report[2] != 0x4c {
		t.Fatalf("unexpected combo report: %v", report)
	}
	if _, err := buildKeyComboReport([]string{"Unknown"}); err == nil {
		t.Fatal("expected unknown key error")
	}
	if _, err := buildKeyComboReport([]string{"KeyA", "KeyB", "KeyC", "KeyD", "KeyE", "KeyF", "KeyG"}); err == nil {
		t.Fatal("expected six-key limit error")
	}
}

func TestBuildTypeReportsReturnsSkippedRunes(t *testing.T) {
	reports, skipped := buildTypeReports("A界", mustLayout(t, "us"))
	if len(reports) != 2 || string(skipped) != "界" {
		t.Fatalf("reports=%d skipped=%q", len(reports), string(skipped))
	}
	if reports[0][0] != 2 || reports[0][2] != 4 || string(reports[1]) != string(keyUpReport) {
		t.Fatalf("unexpected reports: %v", reports)
	}
}

func TestBuildTypeReportsUsesTheLayoutAndItsDeadKeys(t *testing.T) {
	reports, skipped := buildTypeReports("zäâ", mustLayout(t, "de"))
	if len(skipped) != 0 {
		t.Fatalf("skipped %q", string(skipped))
	}
	// German swaps Z and Y, has ä on its own key, and types â as the dead
	// key ^ and then a. Every key press is followed by its release.
	want := [][]byte{
		{0, 0, 0x1c, 0, 0, 0, 0, 0}, keyUpReport,
		{0, 0, 0x34, 0, 0, 0, 0, 0}, keyUpReport,
		{0, 0, 0x35, 0, 0, 0, 0, 0}, keyUpReport,
		{0, 0, 0x04, 0, 0, 0, 0, 0}, keyUpReport,
	}
	if !reflect.DeepEqual(reports, want) {
		t.Fatalf("reports = %v, want %v", reports, want)
	}
}

func TestSupportedKeyNamesMatchKeyboardDescriptor(t *testing.T) {
	for name, code := range keyCodeMap {
		if code > 0x65 {
			t.Fatalf("key %s uses unsupported descriptor code %#x", name, code)
		}
	}
}
