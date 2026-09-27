package storage

import "NanoKVM-Server/proto"

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
)

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
