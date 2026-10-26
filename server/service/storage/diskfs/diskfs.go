// Package diskfs gives the file browser access to the filesystem inside an
// image. The board's kernel reads it: the image is attached to a loop device
// and mounted, so FAT, exFAT, ISO 9660 and UDF are read by the kernel's own
// drivers rather than a parser written here.
//
// An image must not be written here while the host can see it, nor read here
// while the host can write it: two machines caching one filesystem is how an
// image gets destroyed. Callers enforce that (see service/storage); this
// package only tracks its own mounts so concurrent requests share one.
//
// Ported from sipeed/NanoKVM (server/service/storage/diskfs, commit 5572e1de,
// PR #957) and changed for IronKVM: the loop device is set up with the loop
// driver's ioctls, as service/ventoy does, rather than with losetup; an image
// with a partition table is browsed at its first FAT or exFAT partition; discs
// are browsed read-only; and every path inside a mount goes through os.Root.
// Copyright (c) Sipeed. Licensed under the GNU General Public License v3.0.
package diskfs

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strings"
	"time"
)

// Kind is the filesystem an image holds.
type Kind string

const (
	KindExFAT   Kind = "exfat"
	KindFAT32   Kind = "fat32"
	KindFAT16   Kind = "fat16"
	KindFAT12   Kind = "fat12"
	KindISO9660 Kind = "iso9660"
	KindUDF     Kind = "udf"
	KindNTFS    Kind = "ntfs"
	KindUnknown Kind = "unknown"
)

// Browsable reports whether the file browser can list files inside the image.
// NTFS is recognised only to say why it cannot be: neither of the board's
// kernels has a driver for it.
func (k Kind) Browsable() bool {
	switch k {
	case KindExFAT, KindFAT32, KindFAT16, KindFAT12, KindISO9660, KindUDF:
		return true
	default:
		return false
	}
}

// Writable reports whether files can be written into the filesystem. A disc
// filesystem is read-only by design.
func (k Kind) Writable() bool {
	switch k {
	case KindExFAT, KindFAT32, KindFAT16, KindFAT12:
		return true
	default:
		return false
	}
}

// IsFAT reports whether the filesystem is a FAT variant, which holds no file
// of 4 GiB or more.
func (k Kind) IsFAT() bool {
	return k == KindFAT12 || k == KindFAT16 || k == KindFAT32
}

// MaxFATFileSize is the largest file a FAT volume can hold.
const MaxFATFileSize = 1<<32 - 1

// Volume is where an image's filesystem is: the whole image, or one partition
// of it.
type Volume struct {
	Kind Kind
	// Offset and Size locate a partition in bytes. Both are zero when the
	// filesystem starts at the beginning of the image and fills it.
	Offset int64
	Size   int64
	// UDFBridge marks an ISO 9660 disc that carries a UDF volume as well, as
	// Windows installers do. Its UDF side holds the files the ISO 9660 side
	// only lists as a README.
	UDFBridge bool
}

// Entry is one directory entry inside an image.
type Entry struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	IsDir   bool      `json:"isDir"`
	ModTime time.Time `json:"modifiedAt"`
}

// Image is an image's filesystem, opened for reading or for writing. Paths are
// relative to the filesystem's root, slash separated; "/" and "" name the root.
type Image interface {
	Kind() Kind
	Writable() bool
	// List returns at most limit entries of a directory, directories first,
	// and whether there were more.
	List(rel string, limit int) ([]Entry, bool, error)
	// Open returns a regular file for reading.
	Open(rel string) (*os.File, Entry, error)
	// Write stores r as a new file and reports how many bytes landed. An
	// existing name is refused with os.ErrExist.
	Write(rel string, r io.Reader) (int64, error)
	// Remove deletes a file, or a directory with everything in it.
	Remove(rel string) error
	StatFS() (total, available uint64, err error)
	Close() error
}

var (
	// ErrUnsupported reports an image the file browser cannot read: no
	// filesystem it knows, or one this kernel has no driver for.
	ErrUnsupported = errors.New("the image holds no filesystem the file browser can read")
	// ErrBusy reports an image that cannot be opened the way the caller
	// asked: a write meets a reader, or the loop devices are all taken.
	ErrBusy = errors.New("the image is busy")
	// ErrReadOnly reports a change to an image opened for reading, or to a
	// filesystem that cannot be written.
	ErrReadOnly = errors.New("the image is read-only here")
)

// UploadPrefix marks a partially written upload inside an image. It is deleted
// on failure, renamed into place on success, and never listed.
const UploadPrefix = ".kvm-upload-"

// IsTransferTemp reports whether name is exactly what an upload names its
// partial file: UploadPrefix plus eight lower case hex digits. A user's own
// file that merely starts the same way is not one.
func IsTransferTemp(name string) bool {
	rest, ok := strings.CutPrefix(name, UploadPrefix)
	return ok && len(rest) == 8 && strings.Trim(rest, "0123456789abcdef") == ""
}

const (
	sectorSize = 512
	// opticalSectorSize is the sector size of a disc, and vrsDescriptors the
	// number of volume structure descriptors looked at. A UDF disc keeps its
	// volume recognition sequence right after any ISO 9660 descriptors, and a
	// handful of sectors covers both.
	opticalSectorSize = 2048
	vrsDescriptors    = 32
	// maxPartitions bounds the GPT entries read; 128 is what every
	// partitioning tool writes.
	maxPartitions = 128
)

// Detect reports where an image's filesystem is and what it is, from the boot
// sector, the disc descriptors or the partition table, so the file browser
// knows what to mount before it mounts anything.
func Detect(imagePath string) (Volume, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return Volume{}, err
	}
	defer func() {
		_ = file.Close()
	}()

	info, err := file.Stat()
	if err != nil {
		return Volume{}, err
	}

	boot := make([]byte, sectorSize)
	if _, err := file.ReadAt(boot, 0); err != nil {
		return Volume{Kind: KindUnknown}, nil // too small to hold a filesystem
	}

	if kind := bootSectorKind(boot); kind != KindUnknown {
		return Volume{Kind: kind}, nil
	}

	if v := detectOptical(file); v.Kind != KindUnknown {
		return v, nil
	}

	if v, ok := detectPartition(file, boot, info.Size()); ok {
		return v, nil
	}

	return Volume{Kind: KindUnknown}, nil
}

// bootSectorKind reads a volume boot record. A partition table also ends in
// 0x55AA, so only the FAT type fields and the OEM names decide.
func bootSectorKind(boot []byte) Kind {
	switch string(boot[3:11]) {
	case "EXFAT   ":
		return KindExFAT
	case "NTFS    ":
		return KindNTFS
	}

	if boot[510] != 0x55 || boot[511] != 0xaa {
		return KindUnknown
	}
	// A FAT32 volume may also carry the legacy "FAT16" string at 0x36, so the
	// FAT32 field has to be looked at first.
	if fat := boot[0x52 : 0x52+8]; isFATLabel(fat) {
		return fatKind(string(fat))
	}
	if fat := boot[0x36 : 0x36+8]; isFATLabel(fat) {
		return fatKind(string(fat))
	}
	return KindUnknown
}

// detectOptical walks the volume recognition area 32 KiB into the image. An
// ISO 9660 primary descriptor ("CD001") in the first slot is the common case;
// a pure UDF disc, which many recent installers ship as, has no ISO 9660
// descriptor at all and is only recognised by its NSR02/NSR03 descriptor.
func detectOptical(file io.ReaderAt) Volume {
	id := make([]byte, 5)
	iso, udf := false, false

scan:
	for i := 0; i < vrsDescriptors; i++ {
		offset := int64(0x8000 + i*opticalSectorSize + 1)
		if _, err := file.ReadAt(id, offset); err != nil {
			break
		}

		switch string(id) {
		case "CD001":
			if i == 0 {
				iso = true
			}
		case "NSR02", "NSR03":
			udf = true
		case "BEA01", "TEA01", "BOOT2", "CDW02":
		default:
			// The sequence ends at the first slot that is no descriptor.
			break scan
		}
	}

	switch {
	case iso:
		return Volume{Kind: KindISO9660, UDFBridge: udf}
	case udf:
		return Volume{Kind: KindUDF}
	default:
		return Volume{Kind: KindUnknown}
	}
}

// detectPartition finds the first partition of an MBR or a GPT that holds a
// filesystem the file browser can read. A USB stick image, or one the host
// partitioned and formatted, looks like this.
func detectPartition(file io.ReaderAt, mbr []byte, imageSize int64) (Volume, bool) {
	if mbr[510] != 0x55 || mbr[511] != 0xaa {
		return Volume{}, false
	}

	var parts []extent

	for i := 0; i < 4; i++ {
		entry := mbr[446+16*i : 446+16*(i+1)]
		partType := entry[4]
		start := uint64(binary.LittleEndian.Uint32(entry[8:]))
		sectors := uint64(binary.LittleEndian.Uint32(entry[12:]))
		if partType == 0 || sectors == 0 {
			continue
		}
		if partType == 0xEE {
			parts = append(parts, gptPartitions(file)...)
			break
		}
		// Extended partitions hold more tables, not a filesystem.
		if partType == 0x05 || partType == 0x0F || partType == 0x85 {
			continue
		}
		parts = append(parts, extent{start, sectors})
	}

	for _, p := range parts {
		offset := int64(p.start * sectorSize)
		size := int64(p.sectors * sectorSize)
		if offset <= 0 || size <= 0 || offset >= imageSize {
			continue
		}
		if offset+size > imageSize {
			size = imageSize - offset
		}

		boot := make([]byte, sectorSize)
		if _, err := file.ReadAt(boot, offset); err != nil {
			continue
		}
		if kind := bootSectorKind(boot); kind.Browsable() {
			return Volume{Kind: kind, Offset: offset, Size: size}, true
		}
	}
	return Volume{}, false
}

// gptPartitions reads the partition entries a GPT header at LBA 1 points to.
func gptPartitions(file io.ReaderAt) []extent {
	hdr := make([]byte, sectorSize)
	if _, err := file.ReadAt(hdr, sectorSize); err != nil || string(hdr[:8]) != "EFI PART" {
		return nil
	}
	entriesLBA := binary.LittleEndian.Uint64(hdr[72:])
	count := binary.LittleEndian.Uint32(hdr[80:])
	entrySize := binary.LittleEndian.Uint32(hdr[84:])
	if entrySize < 128 || entrySize > 4096 || entriesLBA == 0 || entriesLBA > 1<<40 {
		return nil
	}
	if count > maxPartitions {
		count = maxPartitions
	}

	var parts []extent
	entry := make([]byte, entrySize)
	zero := make([]byte, 16)
	for i := uint32(0); i < count; i++ {
		offset := int64(entriesLBA)*sectorSize + int64(i)*int64(entrySize)
		if _, err := file.ReadAt(entry, offset); err != nil {
			break
		}
		if string(entry[:16]) == string(zero) {
			continue
		}
		first := binary.LittleEndian.Uint64(entry[32:])
		last := binary.LittleEndian.Uint64(entry[40:])
		if last < first {
			continue
		}
		parts = append(parts, extent{first, last - first + 1})
	}
	return parts
}

func isFATLabel(label []byte) bool {
	l := string(label[:5])
	return len(label) == 8 && (l == "FAT12" || l == "FAT16" || l == "FAT32")
}

func fatKind(label string) Kind {
	switch label[:5] {
	case "FAT12":
		return KindFAT12
	case "FAT16":
		return KindFAT16
	default:
		return KindFAT32
	}
}

// extent is a partition, in sectors.
type extent struct{ start, sectors uint64 }
