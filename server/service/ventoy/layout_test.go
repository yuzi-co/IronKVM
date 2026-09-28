package ventoy

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// fakeBoot stands in for the release's boot files: recognisable bytes of
// the real sizes.
func fakeBoot() BootFiles {
	boot := bytes.Repeat([]byte{0xB0}, 512)
	core := bytes.Repeat([]byte{0xC0}, coreSectors*sectorSize)
	return BootFiles{BootImg: boot, CoreImg: core}
}

var testIdentity = Identity{
	UUID:      [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
	Signature: [4]byte{0xAA, 0xBB, 0xCC, 0xDD},
	Serial:    0xCAFE,
}

// tailFill answers a tail read with the byte 0x70+i, so a test can tell
// whose tail the head holds.
func tailFill(i int, buf []byte) error {
	for j := range buf {
		buf[j] = byte(0x70 + i)
	}
	return nil
}

func TestCHS(t *testing.T) {
	for lba, want := range map[uint64][3]byte{
		0:              {0, 1, 0},
		2048:           {32, 33, 0},
		16450559:       {254, 63 | 0xC0, 0xFF}, // the last sector below the saturation, 1023/254/63
		1 << 30:        {254, 0xFF, 0xFF},
		63 * 255 * 300: {0, 1 | byte(300>>2)&0xC0, byte(300 & 0xFF)},
	} {
		if got := chs(lba); got != want {
			t.Errorf("chs(%d) = %v, want %v", lba, got, want)
		}
	}
}

func TestDiskLayoutMatchesVentoy2Disk(t *testing.T) {
	files := []File{
		{Name: "a.iso", Size: 700 << 20},
		{Name: "b.img", Size: 1000},
	}
	d, err := BuildDisk(files, tailFill, fakeBoot(), testIdentity)
	if err != nil {
		t.Fatal(err)
	}
	mbr := d.Head[:sectorSize]

	if !bytes.Equal(mbr[:diskUUIDOffset], bytes.Repeat([]byte{0xB0}, diskUUIDOffset)) {
		t.Fatal("boot.img's code is not in the MBR")
	}
	if !bytes.Equal(mbr[diskUUIDOffset:diskUUIDOffset+16], testIdentity.UUID[:]) {
		t.Fatal("disk UUID")
	}
	if !bytes.Equal(mbr[diskSignatureOffset:diskSignatureOffset+4], testIdentity.Signature[:]) {
		t.Fatal("disk signature")
	}
	if mbr[444] != 0xB0 || mbr[445] != 0xB0 {
		t.Fatal("bytes 444 and 445 come from boot.img, as VentoyWorker.sh copies 446 bytes")
	}
	if mbr[510] != 0x55 || mbr[511] != 0xAA {
		t.Fatal("MBR signature")
	}

	p1, p2 := mbr[446:462], mbr[462:478]
	if p1[0] != 0x80 || p1[4] != 0x07 || binary.LittleEndian.Uint32(p1[8:]) != 2048 {
		t.Fatalf("partition 1: %x", p1)
	}
	if p2[0] != 0 || p2[4] != 0xEF || binary.LittleEndian.Uint32(p2[12:]) != efiSectors {
		t.Fatalf("partition 2: %x", p2)
	}
	part1Len := uint64(binary.LittleEndian.Uint32(p1[12:]))
	part2Start := uint64(binary.LittleEndian.Uint32(p2[8:]))
	if part2Start != 2048+part1Len || part2Start%8 != 0 || part2Start != d.Part2Start {
		t.Fatalf("partition 2 at %d after partition 1 of %d sectors", part2Start, part1Len)
	}
	if d.Sectors != part2Start+efiSectors {
		t.Fatal("the disk ends where partition 2 does")
	}
	for i := range 16 {
		if mbr[478+i] != 0 {
			t.Fatal("partitions 3 and 4 are empty")
		}
	}

	if !bytes.Equal(d.Head[sectorSize:part1Start*sectorSize], fakeBoot().CoreImg) {
		t.Fatal("core.img is not in sectors 1 to 2047")
	}
	if string(d.Head[part1Start*sectorSize+3:part1Start*sectorSize+11]) != "EXFAT   " {
		t.Fatal("partition 1 does not start with the exFAT boot sector")
	}
	if binary.LittleEndian.Uint64(d.Head[part1Start*sectorSize+72:]) != part1Len {
		t.Fatal("the volume is not the whole partition")
	}
}

func TestExtentsCoverTheDisk(t *testing.T) {
	files := []File{
		{Name: "big.iso", Size: 5<<30 + 2048},
		{Name: "empty.iso", Size: 0},
		{Name: "odd.img", Size: 3*512 + 7},
		{Name: "aligned.iso", Size: 64 << 10},
	}
	d, err := BuildDisk(files, tailFill, fakeBoot(), testIdentity)
	if err != nil {
		t.Fatal(err)
	}

	var pos uint64
	fileSectors := map[int]uint64{}
	headSectors := uint64(len(d.Head)) / sectorSize
	for _, e := range d.Extents {
		if e.Start != pos || e.Length == 0 {
			t.Fatalf("extent %+v does not follow %d", e, pos)
		}
		pos += e.Length
		switch e.Kind {
		case ExtentHead:
			if e.Offset+e.Length > headSectors {
				t.Fatalf("extent %+v runs past the head", e)
			}
		case ExtentZero:
			if e.Length > d.ZeroSectors {
				t.Fatalf("zero run %+v is longer than the head's", e)
			}
		case ExtentFile:
			if e.Offset != fileSectors[e.Index] {
				t.Fatalf("file %d read out of order", e.Index)
			}
			fileSectors[e.Index] += e.Length
		case ExtentEFI:
			if e.Start != d.Part2Start || e.Length != efiSectors {
				t.Fatalf("VTOYEFI extent %+v", e)
			}
		}
	}
	if pos != d.Sectors {
		t.Fatalf("extents end at %d, the disk at %d", pos, d.Sectors)
	}
	for i, f := range files {
		if fileSectors[i] != uint64(f.Size)/sectorSize {
			t.Errorf("%s: %d sectors mapped, want %d", f.Name, fileSectors[i], f.Size/sectorSize)
		}
	}
	if !bytes.Equal(d.Head[d.ZeroOffset*sectorSize:], make([]byte, d.ZeroSectors*sectorSize)) {
		t.Fatal("the zero run is not zero")
	}

	// odd.img's tail sector: its 7 bytes, then zeros.
	for _, e := range d.Extents {
		if e.Kind == ExtentHead && e.Length == 1 && e.Start > part1Start {
			tail := d.Head[e.Offset*sectorSize : (e.Offset+1)*sectorSize]
			if !bytes.Equal(tail[:7], bytes.Repeat([]byte{0x72}, 7)) || !bytes.Equal(tail[7:], make([]byte, 505)) {
				t.Fatalf("tail sector %x", tail[:16])
			}
			return
		}
	}
	t.Fatal("no tail sector for odd.img")
}
