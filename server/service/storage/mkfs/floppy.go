package mkfs

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

// A floppy image: a bare FAT12 volume with no partition table, laid out as
// DOS's FORMAT lays out a 3.5" disk of that capacity. BIOSes, DOS and the
// BIOS flash tools that still ship as floppy images read the geometry from
// the boot sector, so it has to be the real one: a 1.44 MB disk has 18
// sectors per track on 2 heads, not the 63/255 of a hard disk.
//
// The host sees the image on the mass storage gadget as a small removable
// disk without a partition table (a "superfloppy"), which BIOSes, Windows and
// Linux all read.

// floppyFormat is the geometry and FAT12 layout of one floppy capacity.
type floppyFormat struct {
	sectors     uint16 // in the whole disk
	clusterSecs uint8
	rootEntries uint16
	media       uint8
	fatSectors  uint16
	trackSecs   uint16
	heads       uint16
}

// floppyFormats are the 3.5" capacities, by size in bytes. The layouts are
// the ones DOS and Windows format; mkfs.fat writes the same ones except that
// it gives a 2.88 MB disk 224 root entries rather than 240.
var floppyFormats = map[int64]floppyFormat{
	737280:  {sectors: 1440, clusterSecs: 2, rootEntries: 112, media: 0xF9, fatSectors: 3, trackSecs: 9, heads: 2},
	1474560: {sectors: 2880, clusterSecs: 1, rootEntries: 224, media: 0xF0, fatSectors: 9, trackSecs: 18, heads: 2},
	2949120: {sectors: 5760, clusterSecs: 2, rootEntries: 240, media: 0xF0, fatSectors: 9, trackSecs: 36, heads: 2},
}

const (
	floppyReserved = 1
	floppyNumFATs  = 2
)

// IsFloppySize reports whether an image of size bytes is a floppy disk: 720
// KB, 1.44 MB or 2.88 MB.
func IsFloppySize(size int64) bool {
	_, ok := floppyFormats[size]
	return ok
}

// floppyBootCode prints a message and reboots on a key press, as a DOS data
// disk does: a BIOS that boots from the drive stops there instead of hanging
// on a blank screen. It sits at offset 62, after the extended BPB, and
// expects the BIOS's 0000:7C00 load address.
var floppyBootCode = []byte{
	0x31, 0xC0, // xor ax, ax
	0x8E, 0xD8, // mov ds, ax
	0xBE, 0x5E, 0x7C, // mov si, message (0x7C00 + 94)
	0xFC,       // cld
	0xAC,       // print: lodsb
	0x84, 0xC0, // test al, al
	0x74, 0x0B, // jz wait
	0x56,       // push si
	0xB4, 0x0E, // mov ah, 0x0E (teletype output)
	0xBB, 0x07, 0x00, // mov bx, 7
	0xCD, 0x10, // int 0x10
	0x5E,       // pop si
	0xEB, 0xF0, // jmp print
	0x31, 0xC0, // wait: xor ax, ax
	0xCD, 0x16, // int 0x16 (read a key)
	0xCD, 0x19, // int 0x19 (reboot)
	0xEB, 0xFE, // jmp $
}

const floppyBootMessage = "\r\nThis disk is not bootable. Remove it and press a key to retry.\r\n\x00"

// newFloppy lays out an empty FAT12 floppy volume of size bytes, the whole
// disk. label is at most 11 characters a FAT short name may hold, or empty
// for none.
func newFloppy(size int64, label string, serial uint32, now time.Time) ([]byte, error) {
	f, ok := floppyFormats[size]
	if !ok {
		return nil, fmt.Errorf("%d bytes is no floppy size: 737280, 1474560 or 2949120", size)
	}
	if len(label) > 11 {
		return nil, fmt.Errorf("volume label %q is longer than 11 characters", label)
	}

	rootSectors := uint64(f.rootEntries) * 32 / sectorSize
	rootStart := uint64(floppyReserved) + floppyNumFATs*uint64(f.fatSectors)
	dataStart := rootStart + rootSectors
	count := (uint64(f.sectors) - dataStart) / uint64(f.clusterSecs)
	// FAT12 counts up to 4084 clusters, at 12 bits each after the two
	// reserved entries.
	if count > 4084 || (count+2)*3/2 > uint64(f.fatSectors)*sectorSize {
		return nil, fmt.Errorf("layout error: %d clusters in a FAT of %d sectors", count, f.fatSectors)
	}

	padded := label
	if padded == "" {
		padded = "NO NAME"
	}
	for len(padded) < 11 {
		padded += " "
	}

	disk := make([]byte, size)
	boot := disk[:sectorSize]
	copy(boot[0:3], []byte{0xEB, 0x3C, 0x90})
	copy(boot[3:11], "MSWIN4.1")
	binary.LittleEndian.PutUint16(boot[11:], sectorSize)
	boot[13] = f.clusterSecs
	binary.LittleEndian.PutUint16(boot[14:], floppyReserved)
	boot[16] = floppyNumFATs
	binary.LittleEndian.PutUint16(boot[17:], f.rootEntries)
	binary.LittleEndian.PutUint16(boot[19:], f.sectors)
	boot[21] = f.media
	binary.LittleEndian.PutUint16(boot[22:], f.fatSectors)
	binary.LittleEndian.PutUint16(boot[24:], f.trackSecs)
	binary.LittleEndian.PutUint16(boot[26:], f.heads)
	boot[36] = 0x00 // drive number: the first floppy drive
	boot[38] = 0x29 // the next three fields are present
	binary.LittleEndian.PutUint32(boot[39:], serial)
	copy(boot[43:54], padded)
	copy(boot[54:62], "FAT12   ")
	n := copy(boot[62:], floppyBootCode)
	copy(boot[62+n:510], floppyBootMessage)
	boot[510], boot[511] = 0x55, 0xAA

	// Each FAT starts with the media descriptor and an end of chain for the
	// reserved cluster 1; every other cluster is free.
	for i := uint64(0); i < floppyNumFATs; i++ {
		fat := disk[(floppyReserved+i*uint64(f.fatSectors))*sectorSize:]
		fat[0], fat[1], fat[2] = f.media, 0xFF, 0xFF
	}

	if label != "" {
		entry := disk[rootStart*sectorSize:][:32]
		copy(entry[0:11], padded)
		entry[11] = attrVolumeID
		date, clock := fatTimestamp(now)
		binary.LittleEndian.PutUint16(entry[22:], clock)
		binary.LittleEndian.PutUint16(entry[24:], date)
	}
	return disk, nil
}

// CreateFloppy makes path a new floppy image of size bytes (737280, 1474560
// or 2949120), formatted FAT12 and named label, which may be empty. A path
// that exists already is refused with an error that wraps fs.ErrExist, and
// is left as it was.
func CreateFloppy(path string, size int64, label string) error {
	serial, err := randomUint32()
	if err != nil {
		return err
	}
	disk, err := newFloppy(size, label, serial, time.Now())
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	_, err = file.Write(disk)
	if err == nil {
		err = file.Sync()
	}
	if cerr := file.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		_ = os.Remove(path)
	}
	return err
}
