package hid

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecordHidModeWritesTheMarkerThatOutlivesAnImage(t *testing.T) {
	// HID-only mode was only the S03usbhid copy over /etc/init.d/S03usbdev,
	// and the next image put the normal script back. The marker is on /data
	// through the /etc/kvm bind, and S03usbdev hands over while it exists.
	marker := filepath.Join(t.TempDir(), "hid_only")

	if err := recordHidMode(marker, ModeHidOnly); err != nil {
		t.Fatalf("recording hid-only failed: %s", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("hid-only must leave the marker behind: %s", err)
	}

	if err := recordHidMode(marker, ModeNormal); err != nil {
		t.Fatalf("recording normal failed: %s", err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("normal must remove the marker, stat said %v", err)
	}
}

func TestRecordHidModeNormalWithNoMarkerIsNotAnError(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "hid_only")
	if err := recordHidMode(marker, ModeNormal); err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
}

func TestRecordHidModeReportsAMarkerItCannotWrite(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "no-such-dir", "hid_only")
	if err := recordHidMode(marker, ModeHidOnly); err == nil {
		t.Fatal("expected an error for a marker that cannot be written")
	}
}
