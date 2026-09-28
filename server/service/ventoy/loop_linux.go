//go:build linux

package ventoy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// Read-only loop devices, set up with the loop driver's ioctls, as
// `losetup -r` does.

const (
	loFlagsReadOnly  = 1
	loFlagsAutoclear = 4
	loopMajor        = 7
)

var loopControl = "/dev/loop-control"

// loopDevice is a loop device the caller holds open. With autoclear set, the
// kernel frees it on its last close: once device-mapper has opened it,
// closing this one leaves device-mapper the only holder.
type loopDevice struct {
	fd    int
	Path  string
	Major uint32
	Minor uint32
}

// attachLoop attaches file to a free loop device, read-only.
func attachLoop(file string) (*loopDevice, error) {
	backing, err := unix.Open(file, unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", file, err)
	}
	defer func() { _ = unix.Close(backing) }()

	ctl, err := unix.Open(loopControl, unix.O_RDWR|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", loopControl, err)
	}
	defer func() { _ = unix.Close(ctl) }()

	// Another process can take the free device between the two calls, and
	// the kernel then answers EBUSY. Ask again.
	for attempt := 0; attempt < 8; attempt++ {
		n, err := unix.IoctlRetInt(ctl, unix.LOOP_CTL_GET_FREE)
		if err != nil {
			return nil, fmt.Errorf("find a free loop device: %w", err)
		}
		path := fmt.Sprintf("/dev/loop%d", n)
		fd, err := openLoop(path, n)
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

		info := unix.LoopInfo64{Flags: loFlagsReadOnly | loFlagsAutoclear}
		copy(info.File_name[:len(info.File_name)-1], filepath.Base(file))
		if err := unix.IoctlLoopSetStatus64(fd, &info); err != nil {
			_ = unix.IoctlSetInt(fd, unix.LOOP_CLR_FD, 0)
			_ = unix.Close(fd)
			return nil, fmt.Errorf("configure %s: %w", path, err)
		}
		return &loopDevice{fd: fd, Path: path, Major: loopMajor, Minor: uint32(n)}, nil
	}
	return nil, errors.New("no free loop device")
}

// openLoop opens the loop device n, making its node if devtmpfs has not.
func openLoop(path string, n int) (int, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if errors.Is(err, unix.ENOENT) {
		if err := unix.Mknod(path, unix.S_IFBLK|0o600, int(unix.Mkdev(loopMajor, uint32(n)))); err != nil && !errors.Is(err, os.ErrExist) {
			return -1, fmt.Errorf("make %s: %w", path, err)
		}
		fd, err = unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC, 0)
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

// Detach frees a loop device that nothing else took, after a failed build.
func (l *loopDevice) Detach() {
	if l.fd >= 0 {
		_ = unix.IoctlSetInt(l.fd, unix.LOOP_CLR_FD, 0)
	}
	l.Close()
}

func (l *loopDevice) devno() string { return fmt.Sprintf("%d:%d", l.Major, l.Minor) }
