//go:build linux

package diskfs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// Loop devices, set up with the loop driver's ioctls, as service/ventoy sets
// up its own: no losetup is needed, and a kernel built with no loop devices
// made in advance (BLK_DEV_LOOP_MIN_COUNT=0, slot B) hands out a fresh one.

const (
	loFlagsReadOnly  = 1
	loFlagsAutoclear = 4
	loopMajor        = 7
	// loopAttempts bounds the retries when another process takes the free
	// device between the two calls.
	loopAttempts = 8
)

var loopControl = "/dev/loop-control"

// loopDevice is a loop device the caller holds open. With autoclear set, the
// kernel frees it on its last close: once a mount holds it, closing this one
// leaves the mount the only holder, and the unmount frees it.
type loopDevice struct {
	fd   int
	Path string
}

// attachLoop attaches the byte range offset..offset+size of file to a free
// loop device, the whole file when size is zero.
func attachLoop(file string, offset, size int64, writable bool) (*loopDevice, error) {
	mode := unix.O_RDONLY
	if writable {
		mode = unix.O_RDWR
	}

	backing, err := unix.Open(file, mode|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", file, err)
	}
	defer func() { _ = unix.Close(backing) }()

	ctl, err := unix.Open(loopControl, unix.O_RDWR|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("%w: open %s: %v", ErrUnsupported, loopControl, err)
	}
	defer func() { _ = unix.Close(ctl) }()

	for attempt := 0; attempt < loopAttempts; attempt++ {
		n, err := unix.IoctlRetInt(ctl, unix.LOOP_CTL_GET_FREE)
		if err != nil {
			return nil, fmt.Errorf("%w: no free loop device: %v", ErrBusy, err)
		}
		path := fmt.Sprintf("/dev/loop%d", n)
		fd, err := openLoop(path, n, mode)
		if err != nil {
			return nil, err
		}
		if err := unix.IoctlSetInt(fd, unix.LOOP_SET_FD, backing); err != nil {
			_ = unix.Close(fd)
			if errors.Is(err, unix.EBUSY) {
				continue
			}
			return nil, fmt.Errorf("attach %s to %s: %w", file, path, err)
		}

		flags := uint32(loFlagsAutoclear)
		if !writable {
			flags |= loFlagsReadOnly
		}
		info := unix.LoopInfo64{Offset: uint64(offset), Sizelimit: uint64(size), Flags: flags}
		copy(info.File_name[:len(info.File_name)-1], filepath.Base(file))
		if err := unix.IoctlLoopSetStatus64(fd, &info); err != nil {
			_ = unix.IoctlSetInt(fd, unix.LOOP_CLR_FD, 0)
			_ = unix.Close(fd)
			return nil, fmt.Errorf("configure %s: %w", path, err)
		}
		return &loopDevice{fd: fd, Path: path}, nil
	}
	return nil, fmt.Errorf("%w: no free loop device", ErrBusy)
}

// openLoop opens the loop device n, making its node if devtmpfs has not.
func openLoop(path string, n int, mode int) (int, error) {
	fd, err := unix.Open(path, mode|unix.O_CLOEXEC, 0)
	if errors.Is(err, unix.ENOENT) {
		if err := unix.Mknod(path, unix.S_IFBLK|0o600, int(unix.Mkdev(loopMajor, uint32(n)))); err != nil && !errors.Is(err, os.ErrExist) {
			return -1, fmt.Errorf("make %s: %w", path, err)
		}
		fd, err = unix.Open(path, mode|unix.O_CLOEXEC, 0)
	}
	if err != nil {
		return -1, fmt.Errorf("open %s: %w", path, err)
	}
	return fd, nil
}

// Close lets go of the loop device. The kernel frees it when nothing else
// holds it.
func (l *loopDevice) Close() {
	if l.fd >= 0 {
		_ = unix.Close(l.fd)
		l.fd = -1
	}
}

// Detach frees a loop device that no mount took.
func (l *loopDevice) Detach() {
	if l.fd >= 0 {
		_ = unix.IoctlSetInt(l.fd, unix.LOOP_CLR_FD, 0)
	}
	l.Close()
}
