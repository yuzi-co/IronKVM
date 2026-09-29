package storage

import (
	"os"

	"NanoKVM-Server/proto"
)

// ImageDirectory is where every image a drive may serve lives.
const ImageDirectory = imageDirectory

// The drive layer's errors, so a caller outside the package, such as
// Redfish, can map each one to its own answer. They are the same values the
// drive layer returns, so errors.Is holds.
var (
	ErrNoDrive      = errNoDrive
	ErrInvalidImage = errInvalidImage
	ErrInOtherDrive = errInOtherDrive
	ErrMediumLocked = errMediumLocked
	ErrImageLoaded  = errImageLoaded
)

// ImageInUse returns ErrImageLoaded, saying where, while a drive serves the
// image, directly or on a disk the server assembled.
func ImageInUse(file string) error {
	if hidOnly() {
		return nil
	}
	driveMu.Lock()
	defer driveMu.Unlock()
	return imageInUse(file)
}

// ReplaceImage renames src over the image file, and refuses with
// ErrImageLoaded while a drive serves file: the host would see its medium
// change under it.
func ReplaceImage(src, file string) error {
	if hidOnly() {
		return os.Rename(src, file)
	}
	return replaceImage(src, file)
}

// ListDrives returns the drives the gadget has, in LUN order. A gadget in
// HID-only mode has none.
func ListDrives() ([]proto.DriveInfo, error) {
	if hidOnly() {
		return []proto.DriveInfo{}, nil
	}
	return listDrives()
}

// InsertDrive loads file into the drive id, with the same checks, lock and
// eject the UI's insert goes through. ro applies to the disk only.
func InsertDrive(id string, file string, ro bool) error {
	if hidOnly() {
		return errNoDrive
	}
	return insertDrive(id, file, ro)
}

// EjectDrive empties the drive id. An empty drive is left alone.
func EjectDrive(id string) error {
	if hidOnly() {
		return errNoDrive
	}
	return ejectDrive(id)
}
