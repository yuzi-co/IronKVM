package vm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecordMdnsStateWritesTheMarkerThatOutlivesAnImage(t *testing.T) {
	// Deleting /etc/init.d/S50avahi-daemon turned mDNS off until the next
	// image put the script back. The marker is on /data through the
	// /etc/kvm bind, and S50avahi-daemon will not start while it exists.
	marker := filepath.Join(t.TempDir(), "mdns_disabled")

	if err := recordMdnsState(marker, false); err != nil {
		t.Fatalf("recording off failed: %s", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("off must leave the marker behind: %s", err)
	}

	if err := recordMdnsState(marker, true); err != nil {
		t.Fatalf("recording on failed: %s", err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("on must remove the marker, stat said %v", err)
	}
}

func TestRecordMdnsStateOnWithNoMarkerIsNotAnError(t *testing.T) {
	// Every board that never turned mDNS off has no marker.
	marker := filepath.Join(t.TempDir(), "mdns_disabled")
	if err := recordMdnsState(marker, true); err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
}

func TestRecordMdnsStateReportsAMarkerItCannotWrite(t *testing.T) {
	// Off with no marker is off until the next image. The switch must say
	// so rather than report a setting it could not keep.
	marker := filepath.Join(t.TempDir(), "no-such-dir", "mdns_disabled")
	if err := recordMdnsState(marker, false); err == nil {
		t.Fatal("expected an error for a marker that cannot be written")
	}
}
