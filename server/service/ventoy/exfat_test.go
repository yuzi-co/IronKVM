package ventoy

import (
	"encoding/binary"
	"strings"
	"testing"
	"time"
	"unicode/utf16"
)

func TestUpcaseTableIsTheSpecificationsTable(t *testing.T) {
	data := make([]byte, 2*len(upcaseTable))
	for i, u := range upcaseTable {
		binary.LittleEndian.PutUint16(data[2*i:], u)
	}
	if len(data) != 5836 {
		t.Fatalf("table is %d bytes, want 5836", len(data))
	}
	if got := tableChecksum(data); got != 0xE619D30D {
		t.Fatalf("table checksum %#x, want 0xe619d30d", got)
	}
}

func TestUpcaseMap(t *testing.T) {
	for in, want := range map[rune]rune{'a': 'A', 'z': 'Z', 'A': 'A', '1': '1', 'ä': 'Ä', 'я': 'Я', 'ω': 'Ω'} {
		if got := rune(upcaseMap[in]); got != want {
			t.Errorf("upcase %q = %q, want %q", in, got, want)
		}
	}
	if upcaseKey("Äbc.iso") != upcaseKey("äBC.ISO") {
		t.Error("names that differ in case only must compare equal")
	}
}

func TestNameHashIgnoresCase(t *testing.T) {
	a := nameHash(utf16.Encode([]rune("ubuntu.iso")))
	b := nameHash(utf16.Encode([]rune("UBUNTU.ISO")))
	if a != b {
		t.Fatalf("hash %#x != %#x", a, b)
	}
}

func TestEncodeNameRejectsWhatExFATRejects(t *testing.T) {
	for _, name := range []string{"", "a/b", "a:b", "a*b", "a?b", `a\b`, "a|b", "a<b", "a>b", `a"b`, "a\x01b", ".", strings.Repeat("x", 256)} {
		if _, err := encodeName(name); err == nil {
			t.Errorf("%q accepted", name)
		}
	}
	units, err := encodeName("😀 ünïcode.iso")
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 14 {
		t.Fatalf("got %d units, want 14 (one surrogate pair)", len(units))
	}
}

func TestTimestamp(t *testing.T) {
	ts, ms := timestamp(time.Date(2026, 9, 28, 13, 45, 31, 250_000_000, time.UTC))
	want := uint32(46)<<25 | 9<<21 | 28<<16 | 13<<11 | 45<<5 | 15
	if ts != want || ms != 125 {
		t.Fatalf("got %#x/%d, want %#x/125", ts, ms, want)
	}
	if ts, _ := timestamp(time.Unix(0, 0)); ts != 1<<21|1<<16 {
		t.Fatalf("1970 clamps to 1980-01-01, got %#x", ts)
	}
}

func volumeFor(t *testing.T, files []File) *volume {
	t.Helper()
	v, err := newVolume(files, volumeOptions{PartitionOffset: 2048, EndAlign: 8, Serial: 0x1234, Label: "Ventoy"})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestBootRegion(t *testing.T) {
	v := volumeFor(t, []File{{Name: "a.iso", Size: 1 << 30}})
	boot := v.Meta[:12*sectorSize]

	if string(boot[3:11]) != "EXFAT   " || boot[510] != 0x55 || boot[511] != 0xAA {
		t.Fatal("not an exFAT boot sector")
	}
	for i := 11; i < 64; i++ {
		if boot[i] != 0 {
			t.Fatalf("MustBeZero byte %d is %#x", i, boot[i])
		}
	}
	le32 := func(off int) uint32 { return binary.LittleEndian.Uint32(boot[off:]) }
	if got := binary.LittleEndian.Uint64(boot[64:]); got != 2048 {
		t.Errorf("PartitionOffset %d", got)
	}
	if got := binary.LittleEndian.Uint64(boot[72:]); got != v.Sectors {
		t.Errorf("VolumeLength %d, want %d", got, v.Sectors)
	}
	if (2048+v.Sectors)%8 != 0 {
		t.Errorf("the volume ends at %d, not on a multiple of 8", 2048+v.Sectors)
	}
	if boot[109] != 6 {
		t.Errorf("1 GB of data wants 32 KB clusters, shift %d", boot[109])
	}
	fatLen := le32(84)
	count := le32(92)
	if uint64(fatLen)*sectorSize < (uint64(count)+2)*4 {
		t.Errorf("FAT of %d sectors cannot hold %d clusters", fatLen, count)
	}
	if uint64(le32(88))+uint64(count)*v.ClusterSectors > v.Sectors {
		t.Error("the heap runs past the volume")
	}
	if boot[112] != 100 {
		t.Errorf("PercentInUse %d, want 100", boot[112])
	}

	sum := bootChecksum(boot)
	for i := 11 * sectorSize; i < 12*sectorSize; i += 4 {
		if binary.LittleEndian.Uint32(boot[i:]) != sum {
			t.Fatal("checksum sector does not repeat the checksum")
		}
	}
	if string(v.Meta[12*sectorSize:24*sectorSize]) != string(boot) {
		t.Fatal("the backup boot region differs")
	}
}

func TestClusterSize(t *testing.T) {
	for _, c := range []struct {
		size  int64
		shift byte
	}{
		{0, 3}, {100 << 20, 3}, {1 << 30, 6}, {40 << 30, 8},
	} {
		v := volumeFor(t, []File{{Name: "a.iso", Size: c.size}})
		if v.Meta[109] != c.shift {
			t.Errorf("%d bytes: shift %d, want %d", c.size, v.Meta[109], c.shift)
		}
	}
}

func TestEmptyVolumeMeetsTheMinimum(t *testing.T) {
	v := volumeFor(t, nil)
	if v.Sectors < minVolumeSectors {
		t.Fatalf("volume of %d sectors, below %d", v.Sectors, minVolumeSectors)
	}
}

func TestFilesAreContiguousAndChained(t *testing.T) {
	files := []File{
		{Name: "one.iso", Size: 100 << 20},
		{Name: "empty.iso", Size: 0},
		{Name: "odd.img", Size: 12345},
	}
	v := volumeFor(t, files)
	fat := v.Meta[fatOffset*sectorSize:]
	entry := func(c uint64) uint32 { return binary.LittleEndian.Uint32(fat[c*4:]) }
	if entry(0) != fatMedia || entry(1) != fatEndOfChain {
		t.Fatal("reserved FAT entries")
	}

	clusterBytes := v.ClusterSectors * sectorSize
	for i, f := range files {
		if f.Size == 0 {
			if v.Start[i] != 0 || v.Allocated[i] != 0 {
				t.Errorf("%s: an empty file has no clusters", f.Name)
			}
			continue
		}
		clusters := ceilDiv(uint64(f.Size), clusterBytes)
		if v.Allocated[i] != clusters*v.ClusterSectors {
			t.Errorf("%s: %d sectors, want %d", f.Name, v.Allocated[i], clusters*v.ClusterSectors)
		}
		first := (v.Start[i]-v.HeapOffset)/v.ClusterSectors + firstCluster
		for c := first; c < first+clusters-1; c++ {
			if entry(c) != uint32(c+1) {
				t.Fatalf("%s: FAT[%d] = %#x", f.Name, c, entry(c))
			}
		}
		if entry(first+clusters-1) != fatEndOfChain {
			t.Fatalf("%s: chain does not end", f.Name)
		}
	}
	if v.Start[2] != v.Start[0]+v.Allocated[0] {
		t.Error("the files do not follow each other")
	}
}

func TestRootDirectory(t *testing.T) {
	name := "Ubuntu 24.04 😀 Ωmega long name for three name entries.iso"
	v := volumeFor(t, []File{{Name: name, Size: 3 << 20, ModTime: time.Date(2026, 1, 2, 3, 4, 6, 0, time.UTC)}})
	root := v.Meta[(v.Start[0]-v.ClusterSectors)*sectorSize:]
	// The root directory is the cluster before the file's.
	if root[0] != entryLabel || root[1] != 6 {
		t.Fatalf("label entry %#x/%d", root[0], root[1])
	}
	if string(utf16.Decode([]uint16{binary.LittleEndian.Uint16(root[2:]), binary.LittleEndian.Uint16(root[4:])})) != "Ve" {
		t.Fatal("label text")
	}
	if root[32] != entryBitmap || root[64] != entryUpcase {
		t.Fatal("bitmap and up-case entries")
	}

	set := root[96:]
	units := utf16.Encode([]rune(name))
	nameEntries := int(ceilDiv(uint64(len(units)), 15))
	if set[0] != entryFile || int(set[1]) != 1+nameEntries {
		t.Fatalf("file entry %#x, secondary count %d", set[0], set[1])
	}
	n := (2 + nameEntries) * entrySize
	if binary.LittleEndian.Uint16(set[2:]) != setChecksum(set[:n]) {
		t.Fatal("set checksum")
	}
	stream := set[32:]
	if stream[0] != entryStream || stream[1] != flagAllocation|flagNoFatChain || int(stream[3]) != len(units) {
		t.Fatalf("stream entry %#x flags %#x length %d", stream[0], stream[1], stream[3])
	}
	if binary.LittleEndian.Uint16(stream[4:]) != nameHash(units) {
		t.Fatal("name hash")
	}
	if binary.LittleEndian.Uint64(stream[8:]) != 3<<20 || binary.LittleEndian.Uint64(stream[24:]) != 3<<20 {
		t.Fatal("lengths")
	}
	var got []uint16
	for i := 0; i < nameEntries; i++ {
		e := set[(2+i)*entrySize:]
		if e[0] != entryName {
			t.Fatalf("name entry %d type %#x", i, e[0])
		}
		for j := 0; j < 15; j++ {
			got = append(got, binary.LittleEndian.Uint16(e[2+2*j:]))
		}
	}
	if string(utf16.Decode(got[:len(units)])) != name {
		t.Fatalf("name %q", string(utf16.Decode(got)))
	}
	if set[n] != 0 {
		t.Fatal("the directory does not end after the set")
	}
}

func TestNewVolumeRefusesDuplicateNames(t *testing.T) {
	_, err := newVolume([]File{{Name: "a.iso"}, {Name: "A.ISO"}}, volumeOptions{})
	if err == nil {
		t.Fatal("names that differ in case only were accepted")
	}
}
