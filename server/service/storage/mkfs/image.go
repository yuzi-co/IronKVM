package mkfs

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"
	"unicode/utf16"
)

// FS is a filesystem an exchange image can be formatted with.
type FS string

const (
	// ExFAT reads and writes on Windows, macOS and Linux, and holds files of
	// any size.
	ExFAT FS = "exfat"
	// FAT32 is for hosts too old for exFAT and for UEFI firmware, which
	// only has to read FAT. It holds no file of 4 GiB or more.
	FAT32 FS = "fat32"
)

const (
	// MinImageSize is the smallest exchange image: room for FAT32's
	// minimum cluster count with 512 byte clusters, after the 1 MiB before
	// the partition.
	MinImageSize = 64 << 20
	// MaxImageSize is the most an MBR partition table addresses.
	MaxImageSize = 2 << 40

	// partitionStart is where the partition starts, in sectors: 1 MiB, as
	// Windows, macOS and parted place the first partition.
	partitionStart = 2048

	partTypeExFAT = 0x07 // also NTFS; hosts read the boot sector to tell
	partTypeFAT32 = 0x0C // FAT32 with LBA addressing
)

// CreateImage makes path a new disk image of size bytes for exchanging files
// with the host: an MBR with one partition from 1 MiB to the end, formatted
// fs and named label. The file is sparse: only the partition table and the
// filesystem's metadata take space until something writes the rest. A path
// that exists already is refused with an error that wraps fs.ErrExist, and
// is left as it was.
//
// The image carries a partition table rather than being a bare volume (a
// "superfloppy") because every host reads that from a USB disk: Windows
// mounts a superfloppy only from removable media and cannot repartition or
// format it in Disk Management, and some UEFI firmware does not look for a
// filesystem on a disk without a partition table at all.
func CreateImage(path string, size int64, fs FS, label string) error {
	switch {
	case size < MinImageSize:
		return fmt.Errorf("an image of %d bytes is below the minimum of %d MiB", size, MinImageSize>>20)
	case size > MaxImageSize:
		return fmt.Errorf("an image of %d bytes is past the 2 TiB an MBR can hold", size)
	case size%sectorSize != 0:
		return fmt.Errorf("an image of %d bytes is not a whole number of sectors", size)
	}

	sectors := uint64(size)/sectorSize - partitionStart
	serial, err := randomUint32()
	if err != nil {
		return err
	}

	var (
		regions  []region
		partType byte
	)
	switch fs {
	case ExFAT:
		vol, err := NewExFAT(nil, ExFATOptions{
			PartitionOffset: partitionStart,
			Sectors:         sectors,
			Serial:          serial,
			Label:           label,
		})
		if err != nil {
			return err
		}
		regions = []region{{0, vol.Meta}}
		partType = partTypeExFAT
	case FAT32:
		if regions, err = newFAT32(sectors, partitionStart, label, serial, time.Now()); err != nil {
			return err
		}
		partType = partTypeFAT32
	default:
		return fmt.Errorf("unknown filesystem %q", fs)
	}

	diskID, err := randomUint32()
	if err != nil {
		return err
	}
	mbr := partitionTable(diskID|1, partType, partitionStart, sectors)

	// O_EXCL makes the refusal of an existing name atomic.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if err := writeImage(file, size, mbr, regions); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return err
	}
	return nil
}

func writeImage(file *os.File, size int64, mbr []byte, regions []region) error {
	if err := file.Truncate(size); err != nil {
		return err
	}
	if _, err := file.WriteAt(mbr, 0); err != nil {
		return err
	}
	for _, r := range regions {
		if err := writeSparse(file, partitionStart*sectorSize+r.Offset, r.Data); err != nil {
			return err
		}
	}
	return file.Sync()
}

// writeSparse writes data at offset, skipping each 4 KiB block that is all
// zero. The file is new, so those blocks read as zero already, and writing
// them would take the space the image is meant to leave free: an exFAT FAT
// for a large volume is megabytes of zeros.
func writeSparse(file *os.File, offset int64, data []byte) error {
	const block = 4096
	zero := make([]byte, block)
	for start := 0; start < len(data); start += block {
		end := min(start+block, len(data))
		if bytes.Equal(data[start:end], zero[:end-start]) {
			continue
		}
		if _, err := file.WriteAt(data[start:end], offset+int64(start)); err != nil {
			return err
		}
	}
	return nil
}

// partitionTable is an MBR holding one partition. It has no boot code and
// no active partition, so a BIOS moves on to the next disk.
func partitionTable(diskID uint32, partType byte, start, sectors uint64) []byte {
	mbr := make([]byte, sectorSize)
	binary.LittleEndian.PutUint32(mbr[440:], diskID)
	entry := mbr[446:462]
	copy(entry[1:4], chs(start))
	entry[4] = partType
	copy(entry[5:8], chs(start+sectors-1))
	binary.LittleEndian.PutUint32(entry[8:], uint32(start))
	binary.LittleEndian.PutUint32(entry[12:], uint32(sectors))
	mbr[510], mbr[511] = 0x55, 0xAA
	return mbr
}

// chs is the cylinder, head and sector of lba under the 255 heads and 63
// sectors per track every BIOS since the 1990s assumes, or the maximum
// 1023/254/63 past what that can address.
func chs(lba uint64) []byte {
	const heads, spt = 255, 63
	c := lba / (heads * spt)
	if c > 1023 {
		return []byte{0xFE, 0xFF, 0xFF}
	}
	h := (lba / spt) % heads
	s := lba%spt + 1
	return []byte{byte(h), byte(s) | byte(c>>2)&0xC0, byte(c)}
}

func randomUint32() (uint32, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}

// fallbackLabel names a volume whose image name gives nothing to use.
const fallbackLabel = "EXCHANGE"

// LabelFor makes a volume label from an image's name, which is what the host
// shows the disk as. FAT keeps a label in upper case, in the characters of
// a short name, so everything else becomes "_"; exFAT keeps any character a
// file name can hold. Both stop at 11 characters.
func LabelFor(name string, fs FS) string {
	var label string
	if fs == FAT32 {
		var b strings.Builder
		for _, r := range strings.ToUpper(strings.TrimSpace(name)) {
			if b.Len() == 11 {
				break
			}
			if r < 0x80 && (unicode.IsUpper(r) || unicode.IsDigit(r) || strings.ContainsRune(" !#$%&'()-@^_`{}~", r)) {
				b.WriteRune(r)
			} else {
				b.WriteByte('_')
			}
		}
		label = b.String()
	} else {
		var units []uint16
		for _, r := range strings.TrimSpace(name) {
			if r < 0x20 || strings.ContainsRune(`"*/:<>?\|`, r) {
				continue
			}
			u := utf16.Encode([]rune{r})
			if len(units)+len(u) > 11 {
				break
			}
			units = append(units, u...)
		}
		label = string(utf16.Decode(units))
	}

	label = strings.TrimRight(label, " ")
	if !strings.ContainsFunc(label, func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }) {
		return fallbackLabel
	}
	return label
}
