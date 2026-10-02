//go:build linux

package gpiocdev

import (
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

// The structures of include/uapi/linux/gpio.h, uAPI v2. The layouts are fixed
// by the kernel and padded so that they are the same on every architecture;
// uapi_test.go holds the sizes against the header.

type chipInfo struct {
	Name  [32]byte
	Label [32]byte
	Lines uint32
}

type lineAttribute struct {
	ID      uint32
	Padding uint32
	Value   uint64 // flags, values or debounce_period_us
}

type lineInfo struct {
	Name     [32]byte
	Consumer [32]byte
	Offset   uint32
	NumAttrs uint32
	Flags    uint64
	Attrs    [10]lineAttribute
	Padding  [4]uint32
}

type lineConfigAttribute struct {
	Attr lineAttribute
	Mask uint64
}

type lineConfig struct {
	Flags    uint64
	NumAttrs uint32
	Padding  [5]uint32
	Attrs    [10]lineConfigAttribute
}

type lineRequest struct {
	Offsets         [64]uint32
	Consumer        [32]byte
	Config          lineConfig
	NumLines        uint32
	EventBufferSize uint32
	Padding         [5]uint32
	Fd              int32
}

type lineValues struct {
	Bits uint64
	Mask uint64
}

const (
	lineFlagInput  = 1 << 2
	lineFlagOutput = 1 << 3

	lineAttrIDOutputValues = 2
)

func ioc(dir, nr, size uintptr) uintptr {
	return dir<<30 | size<<16 | 0xB4<<8 | nr
}

const (
	iocRead  = 2
	iocWrite = 1
)

var (
	ioctlChipInfo  = ioc(iocRead, 0x01, unsafe.Sizeof(chipInfo{}))
	ioctlLineInfo  = ioc(iocRead|iocWrite, 0x05, unsafe.Sizeof(lineInfo{}))
	ioctlLine      = ioc(iocRead|iocWrite, 0x07, unsafe.Sizeof(lineRequest{}))
	ioctlGetValues = ioc(iocRead|iocWrite, 0x0E, unsafe.Sizeof(lineValues{}))
	ioctlSetValues = ioc(iocRead|iocWrite, 0x0F, unsafe.Sizeof(lineValues{}))
)

func ioctl(fd uintptr, req uintptr, arg unsafe.Pointer) error {
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, fd, req, uintptr(arg))
	if errno != 0 {
		return errno
	}
	return nil
}

func findOnChip(chip, name string) (uint32, bool, error) {
	f, err := os.OpenFile(chip, os.O_RDWR|unix.O_CLOEXEC, 0)
	if err != nil {
		return 0, false, err
	}
	defer f.Close()

	var ci chipInfo
	if err := ioctl(f.Fd(), ioctlChipInfo, unsafe.Pointer(&ci)); err != nil {
		return 0, false, err
	}
	for off := uint32(0); off < ci.Lines; off++ {
		li := lineInfo{Offset: off}
		if err := ioctl(f.Fd(), ioctlLineInfo, unsafe.Pointer(&li)); err != nil {
			return 0, false, err
		}
		if cString(li.Name[:]) == name {
			return off, true, nil
		}
	}
	return 0, false, nil
}

func requestLine(chip string, offset uint32, dir Direction, initial int) (*os.File, error) {
	f, err := os.OpenFile(chip, os.O_RDWR|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var req lineRequest
	req.Offsets[0] = offset
	req.NumLines = 1
	copy(req.Consumer[:len(req.Consumer)-1], Consumer)
	if dir == Output {
		req.Config.Flags = lineFlagOutput
		req.Config.NumAttrs = 1
		req.Config.Attrs[0].Attr.ID = lineAttrIDOutputValues
		if initial != 0 {
			req.Config.Attrs[0].Attr.Value = 1
		}
		req.Config.Attrs[0].Mask = 1
	} else {
		req.Config.Flags = lineFlagInput
	}
	if err := ioctl(f.Fd(), ioctlLine, unsafe.Pointer(&req)); err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(req.Fd), chip+":"+Consumer), nil
}

func getValue(f *os.File) (int, error) {
	v := lineValues{Mask: 1}
	if err := ioctl(f.Fd(), ioctlGetValues, unsafe.Pointer(&v)); err != nil {
		return 0, err
	}
	return int(v.Bits & 1), nil
}

func setValue(f *os.File, val int) error {
	v := lineValues{Mask: 1}
	if val != 0 {
		v.Bits = 1
	}
	return ioctl(f.Fd(), ioctlSetValues, unsafe.Pointer(&v))
}
