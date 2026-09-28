package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// registerTestDevice registers a device serving images, for the length of
// the test, and returns a channel that receives each release.
func registerTestDevice(t *testing.T, path string, images ...string) chan struct{} {
	t.Helper()
	released := make(chan struct{}, 4)
	devicesMu.Lock()
	saved := devices
	devices = nil
	devicesMu.Unlock()
	t.Cleanup(func() {
		devicesMu.Lock()
		devices = saved
		devicesMu.Unlock()
	})

	RegisterDevice(Device{
		Path:     path,
		Name:     "Ventoy",
		Images:   func() []string { return images },
		Released: func() { released <- struct{}{} },
	})
	return released
}

func waitReleased(t *testing.T, released chan struct{}, want bool) {
	t.Helper()
	select {
	case <-released:
		if !want {
			t.Fatal("released, but the drive still holds the device")
		}
	case <-time.After(200 * time.Millisecond):
		if want {
			t.Fatal("the release hook did not run")
		}
	}
}

func TestInsertDeviceLoadsTheDiskReadOnly(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	useHidOnly(t, false)
	writes := recordWrites(t, nil)
	registerTestDevice(t, "/dev/mapper/ventoy")

	if err := InsertDevice("/dev/mapper/ventoy"); err != nil {
		t.Fatalf("insert: %s", err)
	}
	want := []string{"lun.0/ro=1", "lun.0/file=/dev/mapper/ventoy"}
	if len(*writes) != len(want) || (*writes)[0] != want[0] || (*writes)[1] != want[1] {
		t.Fatalf("writes %v, want %v", *writes, want)
	}
	if drive, err := DeviceDrive("/dev/mapper/ventoy"); err != nil || drive != DriveDisk {
		t.Fatalf("DeviceDrive = %q, %v", drive, err)
	}
}

func TestInsertDeviceTakesOnlyARegisteredDevice(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	useHidOnly(t, false)
	registerTestDevice(t, "/dev/mapper/ventoy")

	for _, path := range []string{"/dev/mapper/other", "/dev/mmcblk0", "/data/a.img", "/dev/mapper/../mmcblk0"} {
		if err := InsertDevice(path); !errors.Is(err, errNotADevice) {
			t.Errorf("%s: got %v", path, err)
		}
	}
	// And an image insert never takes the device.
	if err := insertDrive(DriveDisk, "/dev/mapper/ventoy", true); !errors.Is(err, errInvalidImage) {
		t.Fatalf("insertDrive took the device: %v", err)
	}
}

func TestEjectAndReplaceReleaseTheDevice(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	useHidOnly(t, false)
	released := registerTestDevice(t, "/dev/mapper/ventoy")

	if err := InsertDevice("/dev/mapper/ventoy"); err != nil {
		t.Fatal(err)
	}
	// Inserting it again is no release.
	if err := InsertDevice("/dev/mapper/ventoy"); err != nil {
		t.Fatal(err)
	}
	waitReleased(t, released, false)

	if err := insertDrive(DriveDisk, "/data/a.img", false); err != nil {
		t.Fatal(err)
	}
	waitReleased(t, released, true)

	if err := InsertDevice("/dev/mapper/ventoy"); err != nil {
		t.Fatal(err)
	}
	if err := ejectDrive(DriveDisk); err != nil {
		t.Fatal(err)
	}
	waitReleased(t, released, true)
}

func TestRemoveImageRefusesAnImageOnADeviceInADrive(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	useHidOnly(t, false)
	image := filepath.Join(t.TempDir(), "a.iso")
	if err := os.WriteFile(image, []byte("iso"), 0o666); err != nil {
		t.Fatalf("setup: %s", err)
	}
	registerTestDevice(t, "/dev/mapper/ventoy", image)

	// Built but in no drive: the image may go, and the next build drops it.
	other := filepath.Join(filepath.Dir(image), "b.iso")
	if err := os.WriteFile(other, []byte("iso"), 0o666); err != nil {
		t.Fatalf("setup: %s", err)
	}
	registerTestDevice(t, "/dev/mapper/ventoy", image, other)
	if err := removeImage(other); err != nil {
		t.Fatalf("remove while the device is in no drive: %s", err)
	}

	if err := InsertDevice("/dev/mapper/ventoy"); err != nil {
		t.Fatal(err)
	}
	err := removeImage(image)
	if !errors.Is(err, errImageLoaded) {
		t.Fatalf("got %v, want errImageLoaded", err)
	}
	if err.Error() != "the image is loaded on the Ventoy disk in the disk drive" {
		t.Fatalf("message is %q", err.Error())
	}
	if _, err := os.Stat(image); err != nil {
		t.Fatalf("the image was removed: %s", err)
	}
}
