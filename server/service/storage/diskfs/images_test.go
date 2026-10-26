package diskfs

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The images the tests browse are built here, byte by byte, so the tests need
// no mkfs or genisoimage: a FAT16 volume and an ISO 9660 disc, each small
// enough to build in memory and real enough for the kernel to mount.

const testSector = 512

// fatFile is a file or a directory of a test volume. Names are 8.3 and upper
// case, which is all a volume without long names can hold.
type fatFile struct {
	name string // "HELLO.TXT", or "DOCS/NOTE.TXT" inside a directory
	data string
	dir  bool
}

// buildFAT16 returns a superfloppy FAT16 volume of 8 MiB holding files. Each
// directory must come before the files inside it.
func buildFAT16(t *testing.T, files ...fatFile) []byte {
	t.Helper()

	const (
		totalSectors = 16384 // 8 MiB
		reserved     = 1
		numFATs      = 2
		rootEntries  = 512
		rootSectors  = rootEntries * 32 / testSector
		fatSectors   = 64 // room for 16384 clusters
		dataStart    = reserved + numFATs*fatSectors + rootSectors
		clusterSize  = testSector
	)

	img := make([]byte, totalSectors*testSector)
	boot := img[:testSector]
	copy(boot, []byte{0xEB, 0x3C, 0x90})
	copy(boot[3:], "MSWIN4.1")
	binary.LittleEndian.PutUint16(boot[0x0B:], testSector)
	boot[0x0D] = 1 // sectors per cluster
	binary.LittleEndian.PutUint16(boot[0x0E:], reserved)
	boot[0x10] = numFATs
	binary.LittleEndian.PutUint16(boot[0x11:], rootEntries)
	binary.LittleEndian.PutUint16(boot[0x13:], totalSectors)
	boot[0x15] = 0xF8
	binary.LittleEndian.PutUint16(boot[0x16:], fatSectors)
	binary.LittleEndian.PutUint16(boot[0x18:], 32)
	binary.LittleEndian.PutUint16(boot[0x1A:], 2)
	boot[0x24] = 0x80
	boot[0x26] = 0x29
	binary.LittleEndian.PutUint32(boot[0x27:], 0x12345678)
	copy(boot[0x2B:], "TESTVOL    ")
	copy(boot[0x36:], "FAT16   ")
	boot[510], boot[511] = 0x55, 0xAA

	fat := make([]uint16, fatSectors*testSector/2)
	fat[0] = 0xFFF8
	fat[1] = 0xFFFF
	next := uint16(2)

	clusterOffset := func(c uint16) int {
		return (dataStart + int(c) - 2) * clusterSize
	}

	// dirs maps a directory name to the cluster holding its entries, and
	// slots to the next free entry there. The root is "".
	dirs := map[string]uint16{"": 0}
	slots := map[string]int{"": 0}
	entryAt := func(dir string) []byte {
		i := slots[dir]
		slots[dir]++
		if dir == "" {
			off := (reserved + numFATs*fatSectors) * testSector
			return img[off+i*32 : off+i*32+32]
		}
		off := clusterOffset(dirs[dir])
		return img[off+i*32 : off+i*32+32]
	}
	writeEntry := func(e []byte, name string, attr byte, cluster uint16, size int) {
		copy(e, shortName(t, name))
		e[11] = attr
		binary.LittleEndian.PutUint16(e[0x1A:], cluster)
		binary.LittleEndian.PutUint32(e[0x1C:], uint32(size))
		// 2026-01-02 03:04:06
		binary.LittleEndian.PutUint16(e[0x16:], 3<<11|4<<5|3)
		binary.LittleEndian.PutUint16(e[0x18:], (2026-1980)<<9|1<<5|2)
	}

	for _, f := range files {
		parent, name := "", f.name
		if i := strings.LastIndexByte(f.name, '/'); i >= 0 {
			parent, name = f.name[:i], f.name[i+1:]
		}
		if _, ok := dirs[parent]; !ok {
			t.Fatalf("buildFAT16: %s comes before its directory", f.name)
		}

		if f.dir {
			c := next
			next++
			fat[c] = 0xFFFF
			dirs[f.name] = c
			slots[f.name] = 0
			writeEntry(entryAt(parent), name, 0x10, c, 0)
			writeEntry(entryAt(f.name), ".", 0x10, c, 0)
			writeEntry(entryAt(f.name), "..", 0x10, dirs[parent], 0)
			continue
		}

		first := uint16(0)
		if len(f.data) > 0 {
			n := (len(f.data) + clusterSize - 1) / clusterSize
			first = next
			for i := 0; i < n; i++ {
				c := next
				next++
				fat[c] = c + 1
				if i == n-1 {
					fat[c] = 0xFFFF
				}
			}
			copy(img[clusterOffset(first):], f.data)
		}
		writeEntry(entryAt(parent), name, 0x20, first, len(f.data))
	}

	for i := 0; i < numFATs; i++ {
		off := (reserved + i*fatSectors) * testSector
		for j, v := range fat {
			binary.LittleEndian.PutUint16(img[off+2*j:], v)
		}
	}
	return img
}

// shortName is the 11 byte directory form of an 8.3 name.
func shortName(t *testing.T, name string) []byte {
	t.Helper()
	out := []byte("           ")
	if name == "." || name == ".." {
		copy(out, name)
		return out
	}
	base, ext, _ := strings.Cut(name, ".")
	if len(base) > 8 || len(ext) > 3 || strings.ToUpper(name) != name {
		t.Fatalf("shortName: %q is not an upper case 8.3 name", name)
	}
	copy(out, base)
	copy(out[8:], ext)
	return out
}

const isoSector = 2048

// buildISO returns an ISO 9660 disc with the files in its root directory, and
// with a UDF volume recognition sequence after the ISO descriptors when
// udfBridge is set, which is what makes Windows installers bridge discs.
func buildISO(t *testing.T, udfBridge bool, files map[string]string) []byte {
	t.Helper()

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	// Directory records are sorted by name.
	sortStrings(names)

	const (
		pvdLBA  = 16
		rootLBA = 24
	)
	dataLBA := rootLBA + 1
	extents := map[string]int{}
	lba := dataLBA
	for _, name := range names {
		extents[name] = lba
		lba += (len(files[name]) + isoSector - 1) / isoSector
		if len(files[name]) == 0 {
			lba++
		}
	}
	total := lba

	img := make([]byte, total*isoSector)
	sector := func(n int) []byte { return img[n*isoSector : (n+1)*isoSector] }

	both32 := func(b []byte, v uint32) {
		binary.LittleEndian.PutUint32(b, v)
		binary.BigEndian.PutUint32(b[4:], v)
	}
	both16 := func(b []byte, v uint16) {
		binary.LittleEndian.PutUint16(b, v)
		binary.BigEndian.PutUint16(b[2:], v)
	}
	record := func(id []byte, extent, size int, dir bool) []byte {
		n := 33 + len(id)
		if n%2 == 1 {
			n++
		}
		r := make([]byte, n)
		r[0] = byte(n)
		both32(r[2:], uint32(extent))
		both32(r[10:], uint32(size))
		copy(r[18:], []byte{126, 1, 2, 3, 4, 5, 0}) // 2026-01-02 03:04:05
		if dir {
			r[25] = 2
		}
		both16(r[28:], 1)
		r[32] = byte(len(id))
		copy(r[33:], id)
		return r
	}

	pvd := sector(pvdLBA)
	pvd[0] = 1
	copy(pvd[1:], "CD001")
	pvd[6] = 1
	copy(pvd[8:], strings.Repeat(" ", 32))
	copy(pvd[40:], "TESTDISC"+strings.Repeat(" ", 24))
	both32(pvd[80:], uint32(total))
	both16(pvd[120:], 1)
	both16(pvd[124:], 1)
	both16(pvd[128:], isoSector)
	both32(pvd[132:], 10)
	binary.LittleEndian.PutUint32(pvd[140:], 18)
	binary.BigEndian.PutUint32(pvd[148:], 19)
	copy(pvd[156:], record([]byte{0}, rootLBA, isoSector, true))
	pvd[881] = 1

	// The set terminator, and on a bridge disc the UDF volume recognition
	// sequence right after it, as mkisofs -udf lays them out. The UDF volume
	// itself is absent, so a UDF mount of this disc fails and the ISO 9660
	// side is what can be read.
	term := sector(pvdLBA + 1)
	term[0] = 255
	copy(term[1:], "CD001")
	term[6] = 1
	if udfBridge {
		for i, id := range []string{"BEA01", "NSR02", "TEA01"} {
			s := sector(pvdLBA + 2 + i)
			copy(s[1:], id)
			s[6] = 1
		}
	}

	// Path tables: the root alone. Linux does not read them, other readers do.
	lpt := sector(21)
	lpt[0] = 1
	binary.LittleEndian.PutUint32(lpt[2:], rootLBA)
	binary.LittleEndian.PutUint16(lpt[6:], 1)
	mpt := sector(22)
	mpt[0] = 1
	binary.BigEndian.PutUint32(mpt[2:], rootLBA)
	binary.BigEndian.PutUint16(mpt[6:], 1)

	root := sector(rootLBA)
	off := 0
	for _, r := range [][]byte{
		record([]byte{0}, rootLBA, isoSector, true),
		record([]byte{1}, rootLBA, isoSector, true),
	} {
		off += copy(root[off:], r)
	}
	for _, name := range names {
		off += copy(root[off:], record([]byte(name+";1"), extents[name], len(files[name]), false))
		copy(img[extents[name]*isoSector:], files[name])
	}
	return img
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// withMBR puts volume behind a DOS partition table, as partition 1 of type
// partType starting at sector 2048.
func withMBR(volume []byte, partType byte) []byte {
	const start = 2048
	img := make([]byte, start*testSector+len(volume))
	entry := img[446:]
	entry[4] = partType
	binary.LittleEndian.PutUint32(entry[8:], start)
	binary.LittleEndian.PutUint32(entry[12:], uint32(len(volume)/testSector))
	img[510], img[511] = 0x55, 0xAA
	copy(img[start*testSector:], volume)
	return img
}

// withGPT puts volume behind a GPT, as its second partition starting at
// sector 2048; the first entry is an empty slot.
func withGPT(volume []byte) []byte {
	const start = 2048
	img := make([]byte, start*testSector+len(volume)+34*testSector)
	// The protective MBR.
	img[446+4] = 0xEE
	binary.LittleEndian.PutUint32(img[446+8:], 1)
	binary.LittleEndian.PutUint32(img[446+12:], 0xFFFFFFFF)
	img[510], img[511] = 0x55, 0xAA

	hdr := img[testSector : 2*testSector]
	copy(hdr, "EFI PART")
	binary.LittleEndian.PutUint64(hdr[72:], 2)   // partition entries LBA
	binary.LittleEndian.PutUint32(hdr[80:], 128) // number of entries
	binary.LittleEndian.PutUint32(hdr[84:], 128) // entry size

	entry := img[2*testSector+128 : 2*testSector+256]
	copy(entry, "basic data guid!") // any non-zero type GUID
	binary.LittleEndian.PutUint64(entry[32:], start)
	binary.LittleEndian.PutUint64(entry[40:], start+uint64(len(volume)/testSector)-1)
	copy(img[start*testSector:], volume)
	return img
}

// bootSector returns a 64 KiB volume whose first sector carries only what
// detection reads: the jump, the OEM name and the FAT type fields.
func bootSector(oem string, at int, label string) []byte {
	img := make([]byte, 64<<10)
	copy(img, []byte{0xEB, 0x58, 0x90})
	copy(img[3:], oem)
	if label != "" {
		copy(img[at:], label)
	}
	img[510], img[511] = 0x55, 0xAA
	return img
}

func writeTemp(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
