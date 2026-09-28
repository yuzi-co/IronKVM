//go:build linux

package ventoy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// The device-mapper device that is the Ventoy disk: one linear table over
// read-only loop devices on the head, on each image and on VTOYEFI.

const (
	// dmMiscMinor is device-mapper's control node, 10:236.
	dmMiscMinor = 236
	miscMajor   = 10
)

// procMisc lists the misc devices, device-mapper's control among them when
// the kernel has it.
var procMisc = "/proc/misc"

// kernelHasDM reports whether the kernel has device-mapper.
func kernelHasDM() bool {
	data, err := os.ReadFile(procMisc)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if f := strings.Fields(line); len(f) == 2 && f[1] == "device-mapper" {
			return true
		}
	}
	return false
}

// ensureControl makes /dev/mapper/control when devtmpfs has not.
func ensureControl() error {
	if _, err := os.Stat(dmControl); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dmControl), 0o755); err != nil {
		return err
	}
	err := unix.Mknod(dmControl, unix.S_IFCHR|0o600, int(unix.Mkdev(miscMajor, dmMiscMinor)))
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("make %s: %w", dmControl, err)
	}
	return nil
}

// sources are the files a disk's extents read.
type sources struct {
	Head  string
	EFI   string
	Files []string
}

// assemble creates the device-mapper device name for the disk, and its node
// at node. Nothing it made is left behind when it fails.
func assemble(name, uuid, node string, d *Disk, src sources) (err error) {
	if err := ensureControl(); err != nil {
		return err
	}

	var loops []*loopDevice
	defer func() {
		for _, l := range loops {
			if err != nil {
				l.Detach()
			} else {
				// Device-mapper holds each one now.
				l.Close()
			}
		}
	}()
	attach := func(file string) (*loopDevice, error) {
		l, err := attachLoop(file)
		if err != nil {
			return nil, err
		}
		loops = append(loops, l)
		return l, nil
	}

	head, err := attach(src.Head)
	if err != nil {
		return err
	}
	efi, err := attach(src.EFI)
	if err != nil {
		return err
	}
	files := make(map[int]*loopDevice)

	targets := make([]dmTarget, 0, len(d.Extents))
	for _, e := range d.Extents {
		var dev *loopDevice
		offset := e.Offset
		switch e.Kind {
		case ExtentHead:
			dev = head
		case ExtentZero:
			dev, offset = head, d.ZeroOffset
		case ExtentEFI:
			dev = efi
		case ExtentFile:
			if dev = files[e.Index]; dev == nil {
				if e.Index >= len(src.Files) {
					return fmt.Errorf("layout error: no file %d", e.Index)
				}
				if dev, err = attach(src.Files[e.Index]); err != nil {
					return err
				}
				files[e.Index] = dev
			}
		}
		targets = append(targets, dmTarget{
			Start:  e.Start,
			Length: e.Length,
			Type:   "linear",
			Params: fmt.Sprintf("%s %d", dev.devno(), offset),
		})
	}

	dev, err := dmCreate(name, uuid, targets)
	if err != nil {
		return err
	}
	if err := makeNode(node, dev); err != nil {
		_ = dmRemove(name)
		return err
	}
	return nil
}

// makeNode makes the block node for the device, as dmsetup does without
// udev. The mass storage function reports the path it opened, so a node here
// reads back as itself, where a link would read back as /dev/dm-N.
func makeNode(node string, dev uint64) error {
	_ = os.Remove(node)
	if err := os.MkdirAll(filepath.Dir(node), 0o755); err != nil {
		return err
	}
	if err := unix.Mknod(node, unix.S_IFBLK|0o600, int(dev)); err != nil {
		return fmt.Errorf("make %s: %w", node, err)
	}
	return nil
}

// deviceExists reports whether the device-mapper device name exists.
func deviceExists(name string) bool {
	if !kernelHasDM() {
		return false
	}
	_, _, err := dmStatus(name)
	return err == nil
}

// removeTimeout is how long a remove waits for the drive to close the
// device. The kernel closes it as the eject returns, but a host's read in
// flight can hold it a moment longer.
const removeTimeout = 5 * time.Second

// disassemble removes the device name and its node. Its loop devices free
// themselves.
func disassemble(name, node string) error {
	if !kernelHasDM() {
		return nil
	}
	deadline := time.Now().Add(removeTimeout)
	for {
		err := dmRemove(name)
		if err == nil || errors.Is(err, unix.ENXIO) {
			break
		}
		if !errors.Is(err, unix.EBUSY) || time.Now().After(deadline) {
			return fmt.Errorf("remove %s: %w", name, err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err := os.Remove(node); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// deviceSize is the size of the device name in bytes, or 0.
func deviceSize(name string) int64 {
	dev, _, err := dmStatus(name)
	if err != nil {
		return 0
	}
	data, err := os.ReadFile(fmt.Sprintf("/sys/dev/block/%d:%d/size", unix.Major(dev), unix.Minor(dev)))
	if err != nil {
		return 0
	}
	var sectors int64
	if _, err := fmt.Sscan(strings.TrimSpace(string(data)), &sectors); err != nil {
		return 0
	}
	return sectors * sectorSize
}
