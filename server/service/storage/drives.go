package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/utils"
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
	// with every CD it sees. A kernel without forced_eject leaves only the
	// host able to release it.
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
		describeFile(&info)
		drives = append(drives, info)
	}
	return drives, nil
}

// describeFile fills in what the UI warns about for the file a drive serves:
// its size, and whether its path is gone. The kernel keeps a deleted file open
// and the host goes on reading it, so a drive can serve an image the library
// no longer lists.
func describeFile(info *proto.DriveInfo) {
	if info.File == "" {
		return
	}
	st, err := os.Stat(info.File)
	if errors.Is(err, fs.ErrNotExist) {
		info.Missing = true
		return
	}
	if err == nil && st.Mode().IsRegular() {
		info.Size = st.Size()
	}
}

// loadedDrive returns the id of the drive holding file, or "" if none does.
// resolveImage returns the file an insert of file would serve, with every
// link followed, or errInvalidImage when that is not an image file under the
// image directory. The check on the name alone would let a link such as
// /data/x.img -> /etc/shadow through. A path that does not exist has nothing
// to follow and keeps its name; the kernel refuses it when the drive opens it.
func resolveImage(file string) (string, error) {
	file = filepath.Clean(file)
	if !isMountableImage(file) {
		return "", errInvalidImage
	}

	resolved, err := filepath.EvalSymlinks(file)
	if errors.Is(err, fs.ErrNotExist) {
		return file, nil
	}
	if err != nil {
		return "", errInvalidImage
	}

	root := realPath(imageRoot)
	if !utils.IsPathInside(root, resolved) || !hasImageSuffix(resolved) {
		return "", errInvalidImage
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.Mode().IsRegular() {
		return "", errInvalidImage
	}
	return resolved, nil
}

// realPath is path with every link followed, or path itself, cleaned, when
// it cannot be resolved.
func realPath(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return filepath.Clean(path)
}

func loadedDrive(file string) (string, error) {
	drives, err := listDrives()
	if err != nil {
		return "", err
	}
	// Two names for one file are one image: compare where they lead.
	clean := realPath(file)
	for _, d := range drives {
		if d.File != "" && realPath(d.File) == clean {
			return d.ID, nil
		}
	}
	return "", nil
}

// eject empties the drive. When the host has locked the medium, it writes the
// LUN's forced_eject, which the kernel has from ironkvm-dist's patch 0006 on:
// that clears the host's lock and closes the file, and the host sees the
// medium go as it would after a normal eject.
func eject(d driveDef) error {
	err := writeAttr(lunPath(d, "file"), []byte("\n"))
	if !errors.Is(err, syscall.EBUSY) {
		return err
	}
	forced := lunPath(d, "forced_eject")
	if _, err := os.Stat(forced); err != nil {
		return errMediumLocked
	}
	return writeAttr(forced, []byte("1"))
}

// ejectDrive removes the drive's medium. An empty drive is left alone.
func ejectDrive(id string) error {
	released, err := ejectLocked(id)
	release(released)
	return err
}

// ejectLocked is ejectDrive under the lock. It returns the file the drive
// let go of, if any.
func ejectLocked(id string) (string, error) {
	driveMu.Lock()
	defer driveMu.Unlock()

	d, err := findDrive(id)
	if err != nil {
		return "", err
	}
	info, err := readDrive(d)
	if err != nil {
		return "", err
	}
	if info.File == "" {
		return "", nil
	}
	if err := eject(d); err != nil {
		return "", err
	}
	return info.File, nil
}

// insertDrive loads file into the drive, replacing what it holds. ro applies
// to the disk only; the CD drive is always read-only.
func insertDrive(id string, file string, ro bool) error {
	released, err := insertLocked(id, file, ro, resolveImage)
	release(released)
	return err
}

// insertLocked is an insert under the lock. resolve checks the file and
// returns what the drive is to open. It returns the file the drive let go of
// when that was a different one.
func insertLocked(id string, file string, ro bool, resolve func(string) (string, error)) (string, error) {
	driveMu.Lock()
	defer driveMu.Unlock()

	file, err := resolve(file)
	if err != nil {
		return "", err
	}
	d, err := findDrive(id)
	if err != nil {
		return "", err
	}

	// A writable disk and a CD on one backing file would corrupt it.
	holder, err := loadedDrive(file)
	if err != nil {
		return "", err
	}
	if holder != "" && holder != id {
		return "", errInOtherDrive
	}

	info, err := readDrive(d)
	if err != nil {
		return "", err
	}
	released := ""
	if info.File != "" {
		if err := eject(d); err != nil {
			return "", err
		}
		if realPath(info.File) != realPath(file) {
			released = info.File
		}
	}

	// The kernel takes a change to ro only while no medium is present.
	if d.id == DriveDisk {
		flag := "0"
		if ro {
			flag = "1"
		}
		if err := writeAttr(lunPath(d, "ro"), []byte(flag)); err != nil {
			return released, err
		}
	}

	return released, writeAttr(lunPath(d, "file"), []byte(filepath.Clean(file)))
}

// errImageLoaded is wrapped with the drive that holds the image.
var errImageLoaded = errors.New("the image is loaded")

// removeImage deletes an image unless a drive is serving it. Removing it would
// pull the medium out from under the host. The lock keeps an insert from
// loading it between the check and the remove.
func removeImage(file string) error {
	driveMu.Lock()
	defer driveMu.Unlock()

	if err := imageInUse(file); err != nil {
		return err
	}
	return os.Remove(file)
}

// imageInUse returns errImageLoaded, with where, when a drive serves file
// itself or through a device such as the Ventoy disk. The caller holds
// driveMu.
func imageInUse(file string) error {
	holder, err := loadedDrive(file)
	if err != nil {
		return err
	}
	if holder != "" {
		return fmt.Errorf("%w in the %s drive", errImageLoaded, holder)
	}
	dev, holder, err := deviceServing(file)
	if err != nil {
		return err
	}
	if holder != "" {
		return fmt.Errorf("%w on the %s disk in the %s drive", errImageLoaded, dev.Name, holder)
	}
	return nil
}

// replaceImage renames src over the image file unless a drive serves file.
// The lock keeps an insert from loading it between the check and the rename.
func replaceImage(src, file string) error {
	driveMu.Lock()
	defer driveMu.Unlock()

	if err := imageInUse(file); err != nil {
		return err
	}
	return os.Rename(src, file)
}
