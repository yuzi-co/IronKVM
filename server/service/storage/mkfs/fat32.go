package mkfs

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

// A FAT32 volume, laid out as Microsoft's FAT specification (fatgen103)
// describes it: 32 reserved sectors with the boot sector, the FSInfo sector
// and their backups at 6 and 7, two FATs, and the root directory in cluster
// 2. Only the sectors that are not zero are generated; an empty volume needs
// nothing else.

const (
	fat32Reserved   = 32
	fat32NumFATs    = 2
	fat32RootClus   = 2
	fat32FSInfo     = 1
	fat32BackupBoot = 6
	// fat32MinClusters is the fewest clusters a FAT32 volume may have: a
	// host counts the clusters, not the type string, and reads a volume
	// with fewer as FAT16.
	fat32MinClusters = 65525
	fat32MaxClusters = 0x0FFFFFF5 - 2
	fat32EndOfChain  = 0x0FFFFFFF
	fat32Media       = 0xF8
	attrVolumeID     = 0x08
)

// region is part of a volume that is not zero, at Offset bytes into it.
type region struct {
	Offset int64
	Data   []byte
}

// fat32ClusterSectors is the cluster size fatgen103's table gives a volume
// of this many sectors, the one Windows' format uses up to 32 GB. A volume
// too small for FAT32 gets 0.
func fat32ClusterSectors(sectors uint64) uint64 {
	switch {
	case sectors <= 66600:
		return 0
	case sectors <= 532480: // 260 MB
		return 1
	case sectors <= 16777216: // 8 GB
		return 8
	case sectors <= 33554432: // 16 GB
		return 16
	case sectors <= 67108864: // 32 GB
		return 32
	default:
		return 64
	}
}

// newFAT32 lays out an empty FAT32 volume of sectors sectors that starts
// hidden sectors into its disk. label is at most 11 characters a FAT short
// name may hold (see LabelFor).
func newFAT32(sectors, hidden uint64, label string, serial uint32, now time.Time) ([]region, error) {
	if len(label) > 11 {
		return nil, fmt.Errorf("volume label %q is longer than 11 characters", label)
	}
	if sectors > 0xFFFFFFFF {
		return nil, errors.New("a FAT32 volume holds at most 2 TiB")
	}
	spc := fat32ClusterSectors(sectors)
	if spc == 0 {
		return nil, errors.New("the volume is too small for FAT32")
	}

	// fatgen103's FAT size: never too small, at most a sector or so large.
	perSector := (256*spc + fat32NumFATs) / 2
	fatSize := ceilDiv(sectors-fat32Reserved, perSector)

	// Pad the reserved area so the data region, and so every cluster,
	// starts on a cluster boundary, and on 4 KiB at least, as the flash
	// under the image writes in pages.
	align := max(spc, 8)
	dataStart := alignUp(fat32Reserved+fat32NumFATs*fatSize, align)
	reserved := dataStart - fat32NumFATs*fatSize
	if reserved > 0xFFFF || dataStart >= sectors {
		return nil, errors.New("the volume is too small for FAT32")
	}
	count := (sectors - dataStart) / spc
	if count < fat32MinClusters {
		return nil, fmt.Errorf("%d clusters is below FAT32's minimum of %d", count, fat32MinClusters)
	}
	if count > fat32MaxClusters {
		return nil, errors.New("too many clusters for FAT32")
	}
	if fatSize*sectorSize/4 < count+2 {
		return nil, fmt.Errorf("layout error: a FAT of %d sectors cannot map %d clusters", fatSize, count)
	}

	padded := label
	for len(padded) < 11 {
		padded += " "
	}

	boot := make([]byte, sectorSize)
	copy(boot[0:3], []byte{0xEB, 0x58, 0x90})
	copy(boot[3:11], "MSWIN4.1")
	binary.LittleEndian.PutUint16(boot[11:], sectorSize)
	boot[13] = byte(spc)
	binary.LittleEndian.PutUint16(boot[14:], uint16(reserved))
	boot[16] = fat32NumFATs
	boot[21] = fat32Media
	binary.LittleEndian.PutUint16(boot[24:], 63)  // sectors per track
	binary.LittleEndian.PutUint16(boot[26:], 255) // heads
	binary.LittleEndian.PutUint32(boot[28:], uint32(hidden))
	binary.LittleEndian.PutUint32(boot[32:], uint32(sectors))
	binary.LittleEndian.PutUint32(boot[36:], uint32(fatSize))
	binary.LittleEndian.PutUint32(boot[44:], fat32RootClus)
	binary.LittleEndian.PutUint16(boot[48:], fat32FSInfo)
	binary.LittleEndian.PutUint16(boot[50:], fat32BackupBoot)
	boot[64] = 0x80 // drive number
	boot[66] = 0x29 // the next three fields are present
	binary.LittleEndian.PutUint32(boot[67:], serial)
	copy(boot[71:82], padded)
	copy(boot[82:90], "FAT32   ")
	// The boot code halts. Nothing boots from this volume.
	copy(boot[90:], []byte{0xF4, 0xEB, 0xFD})
	boot[510], boot[511] = 0x55, 0xAA

	info := make([]byte, sectorSize)
	binary.LittleEndian.PutUint32(info[0:], 0x41615252)
	binary.LittleEndian.PutUint32(info[484:], 0x61417272)
	binary.LittleEndian.PutUint32(info[488:], uint32(count-1)) // the root takes one
	binary.LittleEndian.PutUint32(info[492:], fat32RootClus+1)
	binary.LittleEndian.PutUint32(info[508:], 0xAA550000)

	fat := make([]byte, 12)
	binary.LittleEndian.PutUint32(fat[0:], 0x0FFFFF00|fat32Media)
	binary.LittleEndian.PutUint32(fat[4:], fat32EndOfChain) // clean, no errors
	binary.LittleEndian.PutUint32(fat[8:], fat32EndOfChain) // the root directory

	regions := []region{
		{0, boot},
		{fat32FSInfo * sectorSize, info},
		{fat32BackupBoot * sectorSize, boot},
		{(fat32BackupBoot + fat32FSInfo) * sectorSize, info},
	}
	for i := uint64(0); i < fat32NumFATs; i++ {
		regions = append(regions, region{int64(reserved+i*fatSize) * sectorSize, fat})
	}

	// The root directory holds the volume label, which is the name hosts
	// show; the one in the boot sector is only a copy.
	if label != "" {
		entry := make([]byte, 32)
		copy(entry[0:11], padded)
		entry[11] = attrVolumeID
		date, clock := fatTimestamp(now)
		binary.LittleEndian.PutUint16(entry[22:], clock)
		binary.LittleEndian.PutUint16(entry[24:], date)
		regions = append(regions, region{int64(dataStart) * sectorSize, entry})
	}
	return regions, nil
}

// fatTimestamp is a FAT directory entry's date and time, in local time as
// FAT keeps it, at two second resolution.
func fatTimestamp(t time.Time) (date, clock uint16) {
	if t.Year() < 1980 {
		t = time.Date(1980, 1, 1, 0, 0, 0, 0, t.Location())
	}
	date = uint16(t.Year()-1980)<<9 | uint16(t.Month())<<5 | uint16(t.Day())
	clock = uint16(t.Hour())<<11 | uint16(t.Minute())<<5 | uint16(t.Second()/2)
	return date, clock
}
