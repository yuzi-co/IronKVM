package storage

import (
	"errors"
	"path/filepath"
	"strings"
	"sync"
)

// Device is a block device the server assembles from images, such as the
// Ventoy disk. It is not an image under the image directory, so it goes into
// a drive through InsertDevice, and the checks on images do not apply to it.
// Its images count as loaded while it is in a drive.
type Device struct {
	// Path is the device node a drive serves. It is under /dev/mapper.
	Path string
	// Name names it in messages.
	Name string
	// Images lists the images it reads. Storage calls it with its lock
	// held, so it must not call back into this package.
	Images func() []string
	// Released runs after a drive lets go of the device, by an eject or
	// by another insert, in its own goroutine.
	Released func()
}

// deviceDir is where the assembled devices live.
const deviceDir = "/dev/mapper"

var errNotADevice = errors.New("not a device the server assembled")

var (
	devicesMu sync.Mutex
	devices   []Device
)

// RegisterDevice makes a device known, so InsertDevice takes it and the
// delete guard counts its images. A second registration of one path
// replaces the first.
func RegisterDevice(d Device) {
	devicesMu.Lock()
	defer devicesMu.Unlock()

	for i := range devices {
		if devices[i].Path == d.Path {
			devices[i] = d
			return
		}
	}
	devices = append(devices, d)
}

func findDevice(path string) (Device, bool) {
	devicesMu.Lock()
	defer devicesMu.Unlock()

	for _, d := range devices {
		if d.Path == path {
			return d, true
		}
	}
	return Device{}, false
}

func registeredDevices() []Device {
	devicesMu.Lock()
	defer devicesMu.Unlock()

	return append([]Device(nil), devices...)
}

// resolveDevice accepts only a registered device under /dev/mapper.
func resolveDevice(path string) (string, error) {
	path = filepath.Clean(path)
	if !strings.HasPrefix(path, deviceDir+"/") {
		return "", errNotADevice
	}
	if _, ok := findDevice(path); !ok {
		return "", errNotADevice
	}
	return path, nil
}

// InsertDevice loads a registered device into the disk drive, read-only, with
// the same lock and eject as an image's insert. Only the server calls it,
// never with a path a client sent.
func InsertDevice(path string) error {
	if hidOnly() {
		return errNoDrive
	}
	released, err := insertLocked(DriveDisk, path, true, resolveDevice)
	release(released)
	return err
}

// DeviceDrive returns the drive holding the device, or "" if none does.
func DeviceDrive(path string) (string, error) {
	if hidOnly() {
		return "", nil
	}
	driveMu.Lock()
	defer driveMu.Unlock()

	return loadedDrive(path)
}

// deviceServing returns the device that serves file from a drive, and that
// drive. The caller holds driveMu.
func deviceServing(file string) (Device, string, error) {
	clean := realPath(file)
	for _, d := range registeredDevices() {
		holder, err := loadedDrive(d.Path)
		if err != nil {
			return Device{}, "", err
		}
		if holder == "" || d.Images == nil {
			continue
		}
		for _, image := range d.Images() {
			if realPath(image) == clean {
				return d, holder, nil
			}
		}
	}
	return Device{}, "", nil
}

// release tells a registered device that a drive let go of it. It runs
// after the lock is released, and the hook in its own goroutine, so the
// hook may call back into this package.
func release(file string) {
	if file == "" {
		return
	}
	if d, ok := findDevice(filepath.Clean(file)); ok && d.Released != nil {
		go d.Released()
	}
}
