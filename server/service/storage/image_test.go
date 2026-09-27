package storage

import (
	"errors"
	"testing"
)

func TestNormalizeMountedImageReportsLegacyEmmcAsNoImage(t *testing.T) {
	// Devices that have not rebooted since the update still have the eMMC
	// path in the gadget, and the UI must not show it as a mounted image.
	if got := normalizeMountedImage("/dev/mmcblk0p3\n"); got != "" {
		t.Fatalf("normalizeMountedImage = %q, want empty", got)
	}
}

func TestNormalizeMountedImageKeepsARealImage(t *testing.T) {
	if got := normalizeMountedImage("/data/ubuntu.iso\n"); got != "/data/ubuntu.iso" {
		t.Fatalf("normalizeMountedImage = %q, want the image path", got)
	}
}

func TestIsMountableImageAcceptsImagesInDataDirectory(t *testing.T) {
	for _, path := range []string{"/data/ubuntu.iso", "/data/win.IMG", "/data/isos/debian.iso"} {
		if !isMountableImage(path) {
			t.Fatalf("%q should be mountable", path)
		}
	}
}

func TestIsMountableImageRejectsBlockDevices(t *testing.T) {
	// Mounting a raw device exposes the whole filesystem of the KVM to the
	// machine it is plugged into.
	for _, path := range []string{"/dev/mmcblk0", "/dev/mmcblk0p3", "/etc/shadow"} {
		if isMountableImage(path) {
			t.Fatalf("%q must not be mountable", path)
		}
	}
}

func TestIsMountableImageRejectsTraversal(t *testing.T) {
	if isMountableImage("/data/../etc/shadow.iso") {
		t.Fatal("a traversal out of the image directory must be rejected")
	}
}

func TestIsMountableImageRejectsOtherExtensions(t *testing.T) {
	for _, path := range []string{"/data/notes.txt", "/data/script.sh", "/data/iso"} {
		if isMountableImage(path) {
			t.Fatalf("%q must not be mountable", path)
		}
	}
}

func TestIsMountableImageRejectsEmptyPath(t *testing.T) {
	if isMountableImage("") {
		t.Fatal("an empty path is not a mountable image")
	}
}

func TestLegacyMountWithCdromGoesToTheCdDrive(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")

	if err := legacyMount("/data/a.iso", true); err != nil {
		t.Fatalf("mount: %s", err)
	}
	if got := readAttr(t, dir, "lun.1", "file"); got != "/data/a.iso" {
		t.Fatalf("lun.1/file is %q", got)
	}
	if got := readAttr(t, dir, "lun.0", "file"); got != "" {
		t.Fatalf("lun.0/file is %q, want it untouched", got)
	}
}

func TestLegacyMountWithoutCdromGoesToAWritableDisk(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")

	if err := legacyMount("/data/a.img", false); err != nil {
		t.Fatalf("mount: %s", err)
	}
	if got := readAttr(t, dir, "lun.0", "file"); got != "/data/a.img" {
		t.Fatalf("lun.0/file is %q", got)
	}
	if got := readAttr(t, dir, "lun.0", "ro"); got != "0" {
		t.Fatalf("lun.0/ro is %q, want 0", got)
	}
}

func TestLegacyMountOfNothingEjectsBoth(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	if err := insertDrive(DriveDisk, "/data/a.img", false); err != nil {
		t.Fatalf("setup: %s", err)
	}
	if err := insertDrive(DriveCdrom, "/data/b.iso", true); err != nil {
		t.Fatalf("setup: %s", err)
	}

	if err := legacyMount("", false); err != nil {
		t.Fatalf("unmount: %s", err)
	}
	if readAttr(t, dir, "lun.0", "file") != "" || readAttr(t, dir, "lun.1", "file") != "" {
		t.Fatalf("drives not empty")
	}
}

func TestLegacyMountOfNothingOnAnOldGadgetEjectsTheDisk(t *testing.T) {
	dir := fakeGadget(t, "lun.0")
	if err := insertDrive(DriveDisk, "/data/a.img", false); err != nil {
		t.Fatalf("setup: %s", err)
	}

	if err := legacyMount("", false); err != nil {
		t.Fatalf("unmount: %s", err)
	}
	if got := readAttr(t, dir, "lun.0", "file"); got != "" {
		t.Fatalf("lun.0/file is %q", got)
	}
}

func TestLegacyCdromMountOnAnOldGadgetFails(t *testing.T) {
	dir := fakeGadget(t, "lun.0")

	err := legacyMount("/data/a.iso", true)
	if !errors.Is(err, errNoCdDrive) {
		t.Fatalf("got %v, want errNoCdDrive", err)
	}
	if got := readAttr(t, dir, "lun.0", "file"); got != "" {
		t.Fatalf("lun.0/file is %q, the disk must not take the CD's image", got)
	}
}

func TestLegacyMountedPrefersTheCd(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	if err := insertDrive(DriveDisk, "/data/a.img", false); err != nil {
		t.Fatalf("setup: %s", err)
	}

	got, err := legacyMounted()
	if err != nil || got != "/data/a.img" {
		t.Fatalf("got %q, %v, want the disk's image", got, err)
	}

	if err := insertDrive(DriveCdrom, "/data/b.iso", true); err != nil {
		t.Fatalf("setup: %s", err)
	}
	got, err = legacyMounted()
	if err != nil || got != "/data/b.iso" {
		t.Fatalf("got %q, %v, want the CD's image", got, err)
	}
}

func TestLegacyCdromReportsACdMedium(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")

	if got, err := legacyCdrom(); err != nil || got != 0 {
		t.Fatalf("empty: got %d, %v, want 0", got, err)
	}
	if err := insertDrive(DriveCdrom, "/data/b.iso", true); err != nil {
		t.Fatalf("setup: %s", err)
	}
	if got, err := legacyCdrom(); err != nil || got != 1 {
		t.Fatalf("loaded: got %d, %v, want 1", got, err)
	}
}

func TestLegacyCdromOnAnOldGadgetIsZero(t *testing.T) {
	fakeGadget(t, "lun.0")

	if got, err := legacyCdrom(); err != nil || got != 0 {
		t.Fatalf("got %d, %v, want 0", got, err)
	}
}
