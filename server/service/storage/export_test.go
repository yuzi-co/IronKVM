package storage

import (
	"errors"
	"testing"
)

func useHidOnly(t *testing.T, on bool) {
	t.Helper()
	original := hidOnly
	hidOnly = func() bool { return on }
	t.Cleanup(func() { hidOnly = original })
}

func TestExportedDriveCallsReachTheDriveLayer(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	useHidOnly(t, false)

	if err := InsertDrive(DriveCdrom, "/data/win11.iso", true); err != nil {
		t.Fatalf("InsertDrive: %s", err)
	}
	if got := readAttr(t, dir, "lun.1", "file"); got != "/data/win11.iso" {
		t.Fatalf("lun.1/file is %q", got)
	}

	drives, err := ListDrives()
	if err != nil {
		t.Fatalf("ListDrives: %s", err)
	}
	if len(drives) != 2 || drives[1].File != "/data/win11.iso" {
		t.Fatalf("ListDrives = %+v", drives)
	}

	if err := EjectDrive(DriveCdrom); err != nil {
		t.Fatalf("EjectDrive: %s", err)
	}
	if got := readAttr(t, dir, "lun.1", "file"); got != "" {
		t.Fatalf("lun.1/file is %q after eject", got)
	}
}

func TestExportedErrorsAreTheDriveLayersOwn(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	useHidOnly(t, false)

	if err := InsertDrive(DriveDisk, "/etc/shadow", true); !errors.Is(err, ErrInvalidImage) {
		t.Fatalf("InsertDrive(/etc/shadow) = %v, want ErrInvalidImage", err)
	}
	if err := InsertDrive("lun.0", "/data/a.img", true); !errors.Is(err, ErrNoDrive) {
		t.Fatalf("InsertDrive(lun.0) = %v, want ErrNoDrive", err)
	}
	if err := InsertDrive(DriveCdrom, "/data/a.iso", true); err != nil {
		t.Fatal(err)
	}
	if err := InsertDrive(DriveDisk, "/data/a.iso", true); !errors.Is(err, ErrInOtherDrive) {
		t.Fatalf("InsertDrive into the second drive = %v, want ErrInOtherDrive", err)
	}
}

func TestExportedEjectOfALockedMediumReportsIt(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	useHidOnly(t, false)
	if err := InsertDrive(DriveCdrom, "/data/a.iso", true); err != nil {
		t.Fatal(err)
	}
	recordWrites(t, lockedEject)

	if err := EjectDrive(DriveCdrom); !errors.Is(err, ErrMediumLocked) {
		t.Fatalf("EjectDrive = %v, want ErrMediumLocked", err)
	}
}

// In HID-only mode the gadget has no mass storage function, so there is
// nothing to list and nothing to change, whatever configfs still holds.
func TestExportedDriveCallsInHidOnlyModeSeeNoDrives(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	useHidOnly(t, true)

	drives, err := ListDrives()
	if err != nil || len(drives) != 0 {
		t.Fatalf("ListDrives = %+v, %v, want none", drives, err)
	}
	if err := InsertDrive(DriveCdrom, "/data/a.iso", true); !errors.Is(err, ErrNoDrive) {
		t.Fatalf("InsertDrive = %v, want ErrNoDrive", err)
	}
	if err := EjectDrive(DriveCdrom); !errors.Is(err, ErrNoDrive) {
		t.Fatalf("EjectDrive = %v, want ErrNoDrive", err)
	}
	if got := readAttr(t, dir, "lun.1", "file"); got != "" {
		t.Fatalf("lun.1/file is %q, want it untouched", got)
	}
}
