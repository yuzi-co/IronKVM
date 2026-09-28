package ventoy

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// The Ventoy disk as Ventoy2Disk.sh 1.1.17 lays it out in MBR style: boot.img
// in the MBR, core.img from sector 1, partition 1 from sector 2048, and the
// 32 MB VTOYEFI partition right after it, starting on a multiple of 8
// sectors. See format_ventoy_disk_mbr in tool/ventoy_lib.sh and the install
// branch of tool/VentoyWorker.sh.

const (
	// part1Start is where Ventoy2Disk.sh starts partition 1.
	part1Start = 2048
	// efiSectors is VENTOY_SECTOR_NUM, partition 2's size.
	efiSectors = 65536
	// part2Align is the alignment format_ventoy_disk_mbr gives partition 2.
	part2Align = 8
	// coreSectors is what VentoyWorker.sh copies of core.img: sectors 1 to
	// 2047.
	coreSectors = part1Start - 1
	// bootCodeBytes is what it copies of boot.img: everything before the
	// partition table.
	bootCodeBytes = 446
	// maxDiskSectors is the most an MBR can address.
	maxDiskSectors = 1 << 32

	partTypeExFAT = 0x07
	partTypeEFI   = 0xEF
	partActive    = 0x80
)

// Where VentoyWorker.sh writes the disk's two random identifiers.
const (
	diskUUIDOffset      = 384
	diskSignatureOffset = 440
)

// ExtentKind says where a run of disk sectors comes from.
type ExtentKind int

const (
	// ExtentHead is the generated head file.
	ExtentHead ExtentKind = iota
	// ExtentFile is one of the images, at Extent.Index.
	ExtentFile
	// ExtentZero is zeros. The head holds one run of them, Disk.ZeroOffset,
	// long enough for the longest.
	ExtentZero
	// ExtentEFI is Ventoy's VTOYEFI partition image.
	ExtentEFI
)

func (k ExtentKind) String() string {
	switch k {
	case ExtentHead:
		return "head"
	case ExtentFile:
		return "file"
	case ExtentZero:
		return "zero"
	case ExtentEFI:
		return "efi"
	}
	return fmt.Sprintf("ExtentKind(%d)", int(k))
}

// Extent maps Length sectors of the disk, from Start, to a source from
// sector Offset.
type Extent struct {
	Start  uint64
	Length uint64
	Kind   ExtentKind
	Index  int
	Offset uint64
}

// Disk is a laid-out Ventoy disk.
type Disk struct {
	// Head is the whole head file.
	Head []byte
	// Extents cover the disk from sector 0 to Sectors, in order.
	Extents []Extent
	Sectors uint64
	// Part2Start is VTOYEFI's first sector.
	Part2Start uint64
	// ZeroOffset and ZeroSectors are the run of zeros in the head that
	// every ExtentZero reads.
	ZeroOffset  uint64
	ZeroSectors uint64
}

// Identity is what makes one Ventoy disk different from another. It is drawn
// once and kept, so the host sees the same disk after a rebuild.
type Identity struct {
	UUID      [16]byte
	Signature [4]byte
	Serial    uint32
}

// BootFiles are the two boot images from the Ventoy release, core.img
// already decompressed.
type BootFiles struct {
	BootImg []byte
	CoreImg []byte
}

// TailReader reads the last, partial sector of file i: the size%512 bytes
// that follow its last whole sector.
type TailReader func(i int, buf []byte) error

// chs is the CHS address of a sector as fdisk writes it, with 255 heads and
// 63 sectors per track, saturated at 1023/254/63.
func chs(lba uint64) [3]byte {
	const heads, sectors = 255, 63
	c := lba / (heads * sectors)
	h := (lba / sectors) % heads
	s := lba%sectors + 1
	if c > 1023 {
		c, h, s = 1023, 254, 63
	}
	return [3]byte{byte(h), byte(s) | byte(c>>2)&0xC0, byte(c)}
}

func partitionEntry(entry []byte, status, kind byte, start, length uint64) {
	entry[0] = status
	first := chs(start)
	copy(entry[1:4], first[:])
	entry[4] = kind
	last := chs(start + length - 1)
	copy(entry[5:8], last[:])
	binary.LittleEndian.PutUint32(entry[8:], uint32(start))
	binary.LittleEndian.PutUint32(entry[12:], uint32(length))
}

// mbr is sector 0: boot.img's code, the identifiers and the partition table.
func mbr(boot []byte, id Identity, part2Start uint64) []byte {
	sector := make([]byte, sectorSize)
	copy(sector[:bootCodeBytes], boot)
	copy(sector[diskUUIDOffset:], id.UUID[:])
	copy(sector[diskSignatureOffset:], id.Signature[:])
	partitionEntry(sector[446:462], partActive, partTypeExFAT, part1Start, part2Start-part1Start)
	partitionEntry(sector[462:478], 0, partTypeEFI, part2Start, efiSectors)
	sector[510], sector[511] = 0x55, 0xAA
	return sector
}

// BuildDisk lays out the Ventoy disk for files, in the order given. tail
// supplies the partial last sector of a file whose size is not a multiple
// of 512, which the head then holds: a loop device ends at the file's last
// whole sector.
func BuildDisk(files []File, tail TailReader, boot BootFiles, id Identity) (*Disk, error) {
	if len(boot.BootImg) < bootCodeBytes {
		return nil, errors.New("boot.img is shorter than the MBR's code")
	}
	if len(boot.CoreImg) > coreSectors*sectorSize {
		return nil, errors.New("core.img does not fit before partition 1")
	}

	vol, err := newVolume(files, volumeOptions{
		PartitionOffset: part1Start,
		EndAlign:        part2Align,
		Serial:          id.Serial,
		Label:           "Ventoy",
	})
	if err != nil {
		return nil, err
	}

	part2Start := part1Start + vol.Sectors
	sectors := part2Start + efiSectors
	if sectors > maxDiskSectors {
		return nil, errors.New("the images are too large for an MBR disk")
	}

	metaSectors := uint64(len(vol.Meta)) / sectorSize
	head := make([]byte, 0, (part1Start+metaSectors+uint64(len(files)))*sectorSize)
	head = append(head, mbr(boot.BootImg, id, part2Start)...)
	head = append(head, boot.CoreImg...)
	head = append(head, make([]byte, part1Start*sectorSize-len(head))...)
	head = append(head, vol.Meta...)

	d := &Disk{Sectors: sectors, Part2Start: part2Start}
	add := func(e Extent) {
		if e.Length == 0 {
			return
		}
		// A zero run next to another one is one run.
		if n := len(d.Extents); n > 0 && e.Kind == ExtentZero && d.Extents[n-1].Kind == ExtentZero {
			d.Extents[n-1].Length += e.Length
			return
		}
		d.Extents = append(d.Extents, e)
	}

	add(Extent{Start: 0, Length: part1Start + metaSectors, Kind: ExtentHead})
	pos := part1Start + metaSectors
	for i, f := range files {
		if f.Size == 0 {
			continue
		}
		start := part1Start + vol.Start[i]
		if start != pos {
			return nil, fmt.Errorf("layout error: file %d starts at %d, not %d", i, start, pos)
		}
		whole := uint64(f.Size) / sectorSize
		add(Extent{Start: pos, Length: whole, Kind: ExtentFile, Index: i})
		pos += whole
		if rest := uint64(f.Size) % sectorSize; rest > 0 {
			sector := make([]byte, sectorSize)
			if err := tail(i, sector[:rest]); err != nil {
				return nil, fmt.Errorf("read the end of %s: %w", f.Name, err)
			}
			add(Extent{Start: pos, Length: 1, Kind: ExtentHead, Offset: uint64(len(head)) / sectorSize})
			head = append(head, sector...)
			pos++
		}
		end := start + vol.Allocated[i]
		add(Extent{Start: pos, Length: end - pos, Kind: ExtentZero})
		pos = end
	}
	add(Extent{Start: pos, Length: part2Start - pos, Kind: ExtentZero})
	add(Extent{Start: part2Start, Length: efiSectors, Kind: ExtentEFI})

	for _, e := range d.Extents {
		if e.Kind == ExtentZero && e.Length > d.ZeroSectors {
			d.ZeroSectors = e.Length
		}
	}
	d.ZeroOffset = uint64(len(head)) / sectorSize
	head = append(head, make([]byte, d.ZeroSectors*sectorSize)...)
	d.Head = head

	return d, nil
}
