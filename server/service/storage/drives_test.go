package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// fakeGadget lays out the named LUNs the way configfs shows them with no
// medium: file holds a newline, ro holds the flag. It points the drive layer
// at the directory for the length of the test.
func fakeGadget(t *testing.T, luns ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, lun := range luns {
		path := filepath.Join(dir, lun)
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatalf("setup: %s", err)
		}
		ro := "0\n"
		if lun == "lun.1" {
			ro = "1\n"
		}
		if err := os.WriteFile(filepath.Join(path, "file"), []byte("\n"), 0o666); err != nil {
			t.Fatalf("setup: %s", err)
		}
		if err := os.WriteFile(filepath.Join(path, "ro"), []byte(ro), 0o666); err != nil {
			t.Fatalf("setup: %s", err)
		}
	}

	oldDir := massStorageDir
	massStorageDir = dir
	t.Cleanup(func() { massStorageDir = oldDir })
	return dir
}

// recordWrites replaces writeAttr with one that logs "lun/attr=value" and
// then writes, unless fail returns an error for that write.
func recordWrites(t *testing.T, fail func(path string, data []byte) error) *[]string {
	t.Helper()
	var writes []string
	oldWrite := writeAttr
	writeAttr = func(path string, data []byte) error {
		rel := filepath.Base(filepath.Dir(path)) + "/" + filepath.Base(path)
		writes = append(writes, rel+"="+strings.TrimSpace(string(data)))
		if fail != nil {
			if err := fail(path, data); err != nil {
				return err
			}
		}
		return os.WriteFile(path, data, 0o666)
	}
	t.Cleanup(func() { writeAttr = oldWrite })
	return &writes
}

func readAttr(t *testing.T, dir string, lun string, attr string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, lun, attr))
	if err != nil {
		t.Fatalf("read %s/%s: %s", lun, attr, err)
	}
	return strings.TrimSpace(string(data))
}

// lockedEject makes every eject of a loaded medium fail the way the kernel
// does when the host has locked it.
func lockedEject(path string, data []byte) error {
	if filepath.Base(path) == "file" && strings.TrimSpace(string(data)) == "" {
		return &os.PathError{Op: "write", Path: path, Err: syscall.EBUSY}
	}
	return nil
}

func TestListDrivesShowsBothLunsEmpty(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")

	drives, err := listDrives()
	if err != nil {
		t.Fatalf("listDrives: %s", err)
	}
	if len(drives) != 2 {
		t.Fatalf("got %d drives, want 2: %+v", len(drives), drives)
	}
	if drives[0].ID != DriveDisk || drives[0].Type != DriveDisk || drives[0].File != "" || drives[0].Ro {
		t.Fatalf("disk is %+v", drives[0])
	}
	if drives[1].ID != DriveCdrom || drives[1].Type != DriveCdrom || drives[1].File != "" || !drives[1].Ro {
		t.Fatalf("cdrom is %+v", drives[1])
	}
}

func TestListDrivesOnAnOldGadgetShowsOnlyTheDisk(t *testing.T) {
	fakeGadget(t, "lun.0")

	drives, err := listDrives()
	if err != nil {
		t.Fatalf("listDrives: %s", err)
	}
	if len(drives) != 1 || drives[0].ID != DriveDisk {
		t.Fatalf("got %+v, want only the disk", drives)
	}
}

func TestListDrivesWithTheDiskFunctionOffIsEmpty(t *testing.T) {
	fakeGadget(t)

	drives, err := listDrives()
	if err != nil {
		t.Fatalf("listDrives: %s", err)
	}
	if len(drives) != 0 {
		t.Fatalf("got %+v, want none", drives)
	}
}

func TestInsertIntoTheCdDriveWritesOnlyItsFile(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	writes := recordWrites(t, nil)

	if err := insertDrive(DriveCdrom, "/data/win11.iso", true); err != nil {
		t.Fatalf("insert: %s", err)
	}

	if got := readAttr(t, dir, "lun.1", "file"); got != "/data/win11.iso" {
		t.Fatalf("lun.1/file is %q", got)
	}
	if got := readAttr(t, dir, "lun.0", "file"); got != "" {
		t.Fatalf("lun.0/file is %q, want it untouched", got)
	}
	if strings.Join(*writes, " ") != "lun.1/file=/data/win11.iso" {
		t.Fatalf("writes %v, want only lun.1/file", *writes)
	}
}

func TestInsertIntoTheDiskSetsRoBeforeTheFile(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	if err := os.WriteFile(filepath.Join(dir, "lun.0", "file"), []byte("/data/old.img\n"), 0o666); err != nil {
		t.Fatalf("setup: %s", err)
	}
	writes := recordWrites(t, nil)

	if err := insertDrive(DriveDisk, "/data/drivers.img", true); err != nil {
		t.Fatalf("insert: %s", err)
	}

	// The kernel takes a change to ro only while no medium is present.
	want := "lun.0/file= lun.0/ro=1 lun.0/file=/data/drivers.img"
	if strings.Join(*writes, " ") != want {
		t.Fatalf("writes %v, want %s", *writes, want)
	}
	if got := readAttr(t, dir, "lun.0", "ro"); got != "1" {
		t.Fatalf("lun.0/ro is %q", got)
	}
}

func TestInsertWritableDiskClearsRo(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	if err := os.WriteFile(filepath.Join(dir, "lun.0", "ro"), []byte("1\n"), 0o666); err != nil {
		t.Fatalf("setup: %s", err)
	}

	if err := insertDrive(DriveDisk, "/data/scratch.img", false); err != nil {
		t.Fatalf("insert: %s", err)
	}
	if got := readAttr(t, dir, "lun.0", "ro"); got != "0" {
		t.Fatalf("lun.0/ro is %q, want 0", got)
	}
}

func TestInsertNeverWritesRoOnTheCdDrive(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	writes := recordWrites(t, nil)

	if err := insertDrive(DriveCdrom, "/data/a.iso", false); err != nil {
		t.Fatalf("insert: %s", err)
	}
	for _, w := range *writes {
		if strings.HasPrefix(w, "lun.1/ro=") {
			t.Fatalf("wrote %s, the CD drive stays read-only", w)
		}
	}
}

func TestInsertRefusesAnImageInTheOtherDrive(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	if err := insertDrive(DriveCdrom, "/data/a.iso", true); err != nil {
		t.Fatalf("setup: %s", err)
	}

	err := insertDrive(DriveDisk, "/data/a.iso", false)
	if !errors.Is(err, errInOtherDrive) {
		t.Fatalf("got %v, want errInOtherDrive", err)
	}
	if got := readAttr(t, dir, "lun.0", "file"); got != "" {
		t.Fatalf("lun.0/file is %q, want it untouched", got)
	}
}

func TestInsertRefusesTheSameImageSpelledDifferently(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	if err := insertDrive(DriveCdrom, "/data/a.iso", true); err != nil {
		t.Fatalf("setup: %s", err)
	}

	err := insertDrive(DriveDisk, "/data/./a.iso", false)
	if !errors.Is(err, errInOtherDrive) {
		t.Fatalf("got %v, want errInOtherDrive", err)
	}
}

func TestInsertTheImageTheDriveAlreadyHolds(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	if err := insertDrive(DriveDisk, "/data/a.img", false); err != nil {
		t.Fatalf("setup: %s", err)
	}

	if err := insertDrive(DriveDisk, "/data/a.img", true); err != nil {
		t.Fatalf("re-insert: %s", err)
	}
	if got := readAttr(t, dir, "lun.0", "ro"); got != "1" {
		t.Fatalf("lun.0/ro is %q, want 1", got)
	}
}

func TestInsertRejectsAPathOutsideData(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	writes := recordWrites(t, nil)

	for _, file := range []string{"/etc/shadow", "/data/../etc/x.iso", "/dev/mmcblk0p3", "/data/a.txt"} {
		if err := insertDrive(DriveDisk, file, false); !errors.Is(err, errInvalidImage) {
			t.Fatalf("%s: got %v, want errInvalidImage", file, err)
		}
	}
	if len(*writes) != 0 {
		t.Fatalf("writes %v, want none", *writes)
	}
}

func TestUnknownDriveIdsTouchNothing(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	writes := recordWrites(t, nil)

	for _, id := range []string{"", "lun.0", "../lun.0", "CDROM", "floppy"} {
		if err := insertDrive(id, "/data/a.iso", true); !errors.Is(err, errNoDrive) {
			t.Fatalf("insert %q: got %v, want errNoDrive", id, err)
		}
		if err := ejectDrive(id); !errors.Is(err, errNoDrive) {
			t.Fatalf("eject %q: got %v, want errNoDrive", id, err)
		}
	}
	if len(*writes) != 0 {
		t.Fatalf("writes %v, want none", *writes)
	}
}

func TestInsertIntoAMissingCdDriveFails(t *testing.T) {
	fakeGadget(t, "lun.0")

	if err := insertDrive(DriveCdrom, "/data/a.iso", true); !errors.Is(err, errNoDrive) {
		t.Fatalf("got %v, want errNoDrive", err)
	}
}

func TestEjectEmptiesTheDrive(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	if err := insertDrive(DriveCdrom, "/data/a.iso", true); err != nil {
		t.Fatalf("setup: %s", err)
	}

	if err := ejectDrive(DriveCdrom); err != nil {
		t.Fatalf("eject: %s", err)
	}
	if got := readAttr(t, dir, "lun.1", "file"); got != "" {
		t.Fatalf("lun.1/file is %q, want empty", got)
	}
}

func TestEjectAnEmptyDriveWritesNothing(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	writes := recordWrites(t, lockedEject)

	if err := ejectDrive(DriveDisk); err != nil {
		t.Fatalf("eject: %s", err)
	}
	if len(*writes) != 0 {
		t.Fatalf("writes %v, want none", *writes)
	}
}

func TestEjectALockedMediumReportsIt(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	if err := insertDrive(DriveCdrom, "/data/a.iso", true); err != nil {
		t.Fatalf("setup: %s", err)
	}
	recordWrites(t, lockedEject)

	err := ejectDrive(DriveCdrom)
	if !errors.Is(err, errMediumLocked) {
		t.Fatalf("got %v, want errMediumLocked", err)
	}
	if err.Error() != "the host holds the medium, eject it on the host first" {
		t.Fatalf("message is %q", err.Error())
	}
}

func TestInsertOverALockedMediumKeepsIt(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	if err := insertDrive(DriveCdrom, "/data/a.iso", true); err != nil {
		t.Fatalf("setup: %s", err)
	}
	recordWrites(t, lockedEject)

	if err := insertDrive(DriveCdrom, "/data/b.iso", true); !errors.Is(err, errMediumLocked) {
		t.Fatalf("got %v, want errMediumLocked", err)
	}
	if got := readAttr(t, dir, "lun.1", "file"); got != "/data/a.iso" {
		t.Fatalf("lun.1/file is %q, want the old image kept", got)
	}
}

func TestLoadedDriveFindsTheHolder(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	if err := insertDrive(DriveCdrom, "/data/a.iso", true); err != nil {
		t.Fatalf("setup: %s", err)
	}

	for file, want := range map[string]string{"/data/a.iso": DriveCdrom, "/data/./a.iso": DriveCdrom, "/data/b.iso": ""} {
		got, err := loadedDrive(file)
		if err != nil {
			t.Fatalf("%s: %s", file, err)
		}
		if got != want {
			t.Fatalf("%s: got %q, want %q", file, got, want)
		}
	}
}
