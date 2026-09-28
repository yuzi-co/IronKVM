//go:build linux

package ventoy

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

// The device-mapper ioctls, as linux/dm-ioctl.h defines them. dmsetup is not
// on the board, and the four calls a linear table needs are short enough to
// issue directly.

const (
	dmIoctlType = 0xfd

	dmDevCreateCmd  = 3
	dmDevRemoveCmd  = 4
	dmDevSuspendCmd = 6
	dmDevStatusCmd  = 7
	dmTableLoadCmd  = 9

	dmNameLen = 128
	dmUUIDLen = 129

	// dmIoctlSize is sizeof(struct dm_ioctl) and dmTargetSpecSize
	// sizeof(struct dm_target_spec), both the same on every architecture.
	dmIoctlSize      = 312
	dmTargetSpecSize = 40

	dmReadonlyFlag = 1 << 0

	// dmBufferSize is the smallest buffer a call gets. The calls here
	// return no data, and dmsetup starts at the same size.
	dmBufferSize = 16 << 10
)

// dmVersion is the interface version this code speaks: major 4, as every
// kernel since 2.6 has.
var dmVersion = [3]uint32{4, 0, 0}

// dmControl is device-mapper's control node.
var dmControl = "/dev/mapper/control"

// dmIoctl is struct dm_ioctl.
type dmIoctl struct {
	Version     [3]uint32
	DataSize    uint32
	DataStart   uint32
	TargetCount uint32
	OpenCount   int32
	Flags       uint32
	EventNr     uint32
	Padding     uint32
	Dev         uint64
	Name        [dmNameLen]byte
	UUID        [dmUUIDLen]byte
	Data        [7]byte
}

// dmTarget is one line of a table: Length sectors from Start, served by
// Type with Params.
type dmTarget struct {
	Start, Length uint64
	Type, Params  string
}

// dmRequest is _IOWR(DM_IOCTL, cmd, struct dm_ioctl).
func dmRequest(cmd uintptr) uintptr {
	const read, write = 2, 1
	return (read|write)<<30 | dmIoctlSize<<16 | dmIoctlType<<8 | cmd
}

// dmCall issues one device-mapper ioctl on the device name, with payload
// after the header, and returns the kernel's header. uuid is only for
// DM_DEV_CREATE.
func dmCall(cmd uintptr, name, uuid string, flags uint32, targets int, payload []byte) (*dmIoctl, error) {
	if len(name) >= dmNameLen || len(uuid) >= dmUUIDLen {
		return nil, fmt.Errorf("device name %q or uuid %q is too long", name, uuid)
	}
	size := max(dmIoctlSize+len(payload), dmBufferSize)
	buf := make([]byte, size)

	hdr := dmIoctl{
		Version:     dmVersion,
		DataSize:    uint32(size),
		DataStart:   dmIoctlSize,
		TargetCount: uint32(targets),
		Flags:       flags,
	}
	copy(hdr.Name[:], name)
	copy(hdr.UUID[:], uuid)
	var w bytes.Buffer
	if err := binary.Write(&w, binary.NativeEndian, &hdr); err != nil {
		return nil, err
	}
	copy(buf, w.Bytes())
	copy(buf[dmIoctlSize:], payload)

	fd, err := unix.Open(dmControl, unix.O_RDWR|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", dmControl, err)
	}
	defer func() { _ = unix.Close(fd) }()

	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), dmRequest(cmd), uintptr(unsafe.Pointer(&buf[0])))
	if errno != 0 {
		return nil, errno
	}

	var out dmIoctl
	if err := binary.Read(bytes.NewReader(buf[:dmIoctlSize]), binary.NativeEndian, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// dmTable encodes targets as DM_TABLE_LOAD takes them: each dm_target_spec
// followed by its parameter string, with next the offset from this spec to
// the following one, kept 8-byte aligned.
func dmTable(targets []dmTarget) []byte {
	var out []byte
	for _, t := range targets {
		entry := make([]byte, dmTargetSpecSize, dmTargetSpecSize+len(t.Params)+8)
		binary.NativeEndian.PutUint64(entry[0:], t.Start)
		binary.NativeEndian.PutUint64(entry[8:], t.Length)
		copy(entry[24:40], t.Type)
		entry = append(entry, t.Params...)
		entry = append(entry, 0)
		for len(entry)%8 != 0 {
			entry = append(entry, 0)
		}
		binary.NativeEndian.PutUint32(entry[20:], uint32(len(entry)))
		out = append(out, entry...)
	}
	return out
}

// dmCreate creates the device name, loads targets as a read-only table and
// resumes it, and returns its device number. A failure after the create
// removes the device again.
func dmCreate(name, uuid string, targets []dmTarget) (uint64, error) {
	hdr, err := dmCall(dmDevCreateCmd, name, uuid, 0, 0, nil)
	if err != nil {
		return 0, fmt.Errorf("create %s: %w", name, err)
	}
	if _, err := dmCall(dmTableLoadCmd, name, "", dmReadonlyFlag, len(targets), dmTable(targets)); err != nil {
		_ = dmRemove(name)
		return 0, fmt.Errorf("load the table of %s: %w", name, err)
	}
	// A suspend call without the suspend flag resumes the device, which
	// makes the loaded table the live one.
	if _, err := dmCall(dmDevSuspendCmd, name, "", 0, 0, nil); err != nil {
		_ = dmRemove(name)
		return 0, fmt.Errorf("resume %s: %w", name, err)
	}
	return hdr.Dev, nil
}

// dmRemove removes the device name. It fails with EBUSY while the device is
// open, as it is while a drive serves it.
func dmRemove(name string) error {
	_, err := dmCall(dmDevRemoveCmd, name, "", 0, 0, nil)
	return err
}

// dmStatus returns the device name's number and how many times it is open,
// or os.ErrNotExist.
func dmStatus(name string) (uint64, int, error) {
	hdr, err := dmCall(dmDevStatusCmd, name, "", 0, 0, nil)
	if errors.Is(err, unix.ENXIO) {
		return 0, 0, os.ErrNotExist
	}
	if err != nil {
		return 0, 0, err
	}
	return hdr.Dev, int(hdr.OpenCount), nil
}
