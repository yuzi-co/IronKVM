package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"NanoKVM-Server/proto"
)

// The two virtual drives. Each is one LUN of the gadget's mass storage
// function, and neither ever changes type: S03usbdev creates lun.1 as a
// CD-ROM, and lun.0 stays a disk. So a media change is a write to the LUN's
// file, which the host sees as an insert or an eject, and the gadget is never
// re-enumerated.
const (
	DriveDisk  = "disk"
	DriveCdrom = "cdrom"
)

type driveDef struct {
	id  string
	lun string
}

var driveDefs = []driveDef{
	{id: DriveDisk, lun: "lun.0"},
	{id: DriveCdrom, lun: "lun.1"},
}

// massStorageDir is the gadget's mass storage function. Tests point it at a
// temporary directory laid out the same way.
var massStorageDir = "/sys/kernel/config/usb_gadget/g0/functions/mass_storage.disk0"

// massStorageLink is the function's entry in the gadget config. Turning the
// virtual disk off removes it and keeps the function directory, so the LUNs
// outlive it; only this link says whether the host can see them.
var massStorageLink = "/sys/kernel/config/usb_gadget/g0/configs/c.1/mass_storage.disk0"

// writeAttr writes one configfs attribute. Tests replace it, because a plain
// file cannot return the errors the kernel does.
var writeAttr = func(path string, data []byte) error {
	return os.WriteFile(path, data, 0o666)
}

// driveMu serialises every change to the drives, and the delete of an image,
// so the check that an image is in no other drive holds until its write.
var driveMu sync.Mutex

var (
	errNoDrive      = errors.New("no such drive")
	errInvalidImage = errors.New("not an image under /data")
	errInOtherDrive = errors.New("the image is loaded in the other drive")
	// The kernel refuses to eject a medium the host has locked, as Linux does
	// while a CD is mounted. 5.10 has no forced_eject, so only the host can
	// release it.
	errMediumLocked = errors.New("the host holds the medium, eject it on the host first")
)

func lunPath(d driveDef, attr string) string {
	return filepath.Join(massStorageDir, d.lun, attr)
}

func driveExists(d driveDef) bool {
	if _, err := os.Lstat(massStorageLink); err != nil {
		return false
	}
	info, err := os.Stat(filepath.Join(massStorageDir, d.lun))
	return err == nil && info.IsDir()
}

// findDrive returns the drive with this id, if the gadget has its LUN. A
// gadget built before lun.1 existed has only the disk.
func findDrive(id string) (driveDef, error) {
	for _, d := range driveDefs {
		if d.id == id && driveExists(d) {
			return d, nil
		}
	}
	return driveDef{}, errNoDrive
}

func readDrive(d driveDef) (proto.DriveInfo, error) {
	file, err := os.ReadFile(lunPath(d, "file"))
	if err != nil {
		return proto.DriveInfo{}, err
	}
	ro, err := os.ReadFile(lunPath(d, "ro"))
	if err != nil {
		return proto.DriveInfo{}, err
	}

	return proto.DriveInfo{
		ID:   d.id,
		Type: d.id,
		File: normalizeMountedImage(string(file)),
		Ro:   strings.TrimSpace(string(ro)) == "1",
	}, nil
}

// listDrives returns the drives the gadget has, in LUN order.
func listDrives() ([]proto.DriveInfo, error) {
	drives := []proto.DriveInfo{}
	for _, d := range driveDefs {
		if !driveExists(d) {
			continue
		}
		info, err := readDrive(d)
		if err != nil {
			return nil, err
		}
		drives = append(drives, info)
	}
	return drives, nil
}

// loadedDrive returns the id of the drive holding file, or "" if none does.
func loadedDrive(file string) (string, error) {
	drives, err := listDrives()
	if err != nil {
		return "", err
	}
	clean := filepath.Clean(file)
	for _, d := range drives {
		if d.File != "" && filepath.Clean(d.File) == clean {
			return d.ID, nil
		}
	}
	return "", nil
}

func eject(d driveDef) error {
	err := writeAttr(lunPath(d, "file"), []byte("\n"))
	if errors.Is(err, syscall.EBUSY) {
		return errMediumLocked
	}
	return err
}

// ejectDrive removes the drive's medium. An empty drive is left alone.
func ejectDrive(id string) error {
	driveMu.Lock()
	defer driveMu.Unlock()

	d, err := findDrive(id)
	if err != nil {
		return err
	}
	info, err := readDrive(d)
	if err != nil {
		return err
	}
	if info.File == "" {
		return nil
	}
	return eject(d)
}

// insertDrive loads file into the drive, replacing what it holds. ro applies
// to the disk only; the CD drive is always read-only.
func insertDrive(id string, file string, ro bool) error {
	driveMu.Lock()
	defer driveMu.Unlock()

	if !isMountableImage(file) {
		return errInvalidImage
	}
	d, err := findDrive(id)
	if err != nil {
		return err
	}

	// A writable disk and a CD on one backing file would corrupt it.
	holder, err := loadedDrive(file)
	if err != nil {
		return err
	}
	if holder != "" && holder != id {
		return errInOtherDrive
	}

	info, err := readDrive(d)
	if err != nil {
		return err
	}
	if info.File != "" {
		if err := eject(d); err != nil {
			return err
		}
	}

	// The kernel takes a change to ro only while no medium is present.
	if d.id == DriveDisk {
		flag := "0"
		if ro {
			flag = "1"
		}
		if err := writeAttr(lunPath(d, "ro"), []byte(flag)); err != nil {
			return err
		}
	}

	return writeAttr(lunPath(d, "file"), []byte(filepath.Clean(file)))
}

// errImageLoaded is wrapped with the drive that holds the image.
var errImageLoaded = errors.New("the image is loaded")

// removeImage deletes an image unless a drive is serving it. Removing it would
// pull the medium out from under the host. The lock keeps an insert from
// loading it between the check and the remove.
func removeImage(file string) error {
	driveMu.Lock()
	defer driveMu.Unlock()

	holder, err := loadedDrive(file)
	if err != nil {
		return err
	}
	if holder != "" {
		return fmt.Errorf("%w in the %s drive", errImageLoaded, holder)
	}
	return os.Remove(file)
}
