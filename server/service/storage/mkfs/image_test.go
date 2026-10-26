package mkfs

import (
	"encoding/binary"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"NanoKVM-Server/service/storage/diskfs"
)

func createTemp(t *testing.T, size int64, fsType FS) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "exchange.img")
	if err := CreateImage(path, size, fsType, "EXCHANGE"); err != nil {
		t.Fatalf("CreateImage(%d, %s): %v", size, fsType, err)
	}
	return path
}

func readSector(t *testing.T, path string, lba int64) []byte {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	buf := make([]byte, sectorSize)
	if _, err := f.ReadAt(buf, lba*sectorSize); err != nil {
		t.Fatal(err)
	}
	return buf
}

// The image has the size asked for, and only its metadata takes space on the
// card: the rest is a hole until the host or the file browser writes it.
func TestCreatedImageIsSizedAndSparse(t *testing.T) {
	for _, c := range []struct {
		fs   FS
		size int64
	}{
		{ExFAT, 1 << 30},
		{ExFAT, 64 << 20},
		{FAT32, 1 << 30},
		{FAT32, 64 << 20},
		{ExFAT, 40 << 30},
		{FAT32, 40 << 30},
	} {
		path := createTemp(t, c.size, c.fs)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() != c.size {
			t.Errorf("%s %d: file of %d bytes", c.fs, c.size, info.Size())
		}
		allocated := info.Sys().(*syscall.Stat_t).Blocks * 512
		if allocated > 16<<20 {
			t.Errorf("%s %d: %d bytes allocated, want a sparse file", c.fs, c.size, allocated)
		}
		_ = os.Remove(path)
	}
}

// One primary partition from 1 MiB to the end, of the type that names its
// filesystem, is what a USB stick formatted by Windows or macOS carries.
func TestCreatedImageHasOneMBRPartition(t *testing.T) {
	for fsType, partType := range map[FS]byte{ExFAT: 0x07, FAT32: 0x0C} {
		const size = 256 << 20
		mbr := readSector(t, createTemp(t, size, fsType), 0)
		if mbr[510] != 0x55 || mbr[511] != 0xAA {
			t.Fatalf("%s: no MBR signature", fsType)
		}
		entry := mbr[446:462]
		if entry[0] != 0x00 || entry[4] != partType {
			t.Errorf("%s: status %#x type %#x, want 0x00 %#x", fsType, entry[0], entry[4], partType)
		}
		start := binary.LittleEndian.Uint32(entry[8:])
		length := binary.LittleEndian.Uint32(entry[12:])
		if start != partitionStart || uint64(start)+uint64(length) != size/sectorSize {
			t.Errorf("%s: partition %d+%d sectors, want %d to the end", fsType, start, length, partitionStart)
		}
		for i := 1; i < 4; i++ {
			if mbr[446+16*i+4] != 0 {
				t.Errorf("%s: partition %d is not empty", fsType, i+1)
			}
		}
		if binary.LittleEndian.Uint32(mbr[440:]) == 0 {
			t.Errorf("%s: no disk signature; Windows wants one", fsType)
		}
	}
}

// The file browser of ironkvm-dist#158 finds the filesystem asked for, at the
// partition.
func TestCreatedImageIsDetected(t *testing.T) {
	for fsType, kind := range map[FS]diskfs.Kind{ExFAT: diskfs.KindExFAT, FAT32: diskfs.KindFAT32} {
		const size = 512 << 20
		vol, err := diskfs.Detect(createTemp(t, size, fsType))
		if err != nil {
			t.Fatal(err)
		}
		if vol.Kind != kind || vol.Offset != partitionStart*sectorSize || vol.Size != size-partitionStart*sectorSize {
			t.Errorf("%s: detected %+v", fsType, vol)
		}
	}
}

func TestFAT32BootSector(t *testing.T) {
	const size = 1 << 30
	path := createTemp(t, size, FAT32)
	boot := readSector(t, path, partitionStart)

	if binary.LittleEndian.Uint16(boot[11:]) != sectorSize {
		t.Errorf("bytes per sector %d", binary.LittleEndian.Uint16(boot[11:]))
	}
	spc := uint64(boot[13])
	reserved := uint64(binary.LittleEndian.Uint16(boot[14:]))
	fats := uint64(boot[16])
	total := uint64(binary.LittleEndian.Uint32(boot[32:]))
	fatSize := uint64(binary.LittleEndian.Uint32(boot[36:]))
	if binary.LittleEndian.Uint16(boot[17:]) != 0 || binary.LittleEndian.Uint16(boot[19:]) != 0 || binary.LittleEndian.Uint16(boot[22:]) != 0 {
		t.Error("FAT12/16 fields are set on a FAT32 volume")
	}
	if total != size/sectorSize-partitionStart {
		t.Errorf("TotSec32 %d", total)
	}
	if binary.LittleEndian.Uint32(boot[28:]) != partitionStart {
		t.Errorf("HiddSec %d", binary.LittleEndian.Uint32(boot[28:]))
	}
	clusters := (total - reserved - fats*fatSize) / spc
	if clusters < 65525 {
		t.Errorf("%d clusters: below FAT32's minimum, a host reads it as FAT16", clusters)
	}
	if fatSize*sectorSize/4 < clusters+2 {
		t.Errorf("a FAT of %d sectors cannot map %d clusters", fatSize, clusters)
	}
	if (reserved+fats*fatSize)%spc != 0 {
		t.Errorf("the data region at %d is not aligned to the %d-sector cluster", reserved+fats*fatSize, spc)
	}
	if string(boot[71:82]) != "EXCHANGE   " || string(boot[82:90]) != "FAT32   " {
		t.Errorf("label %q type %q", boot[71:82], boot[82:90])
	}

	info := readSector(t, path, partitionStart+1)
	if binary.LittleEndian.Uint32(info[0:]) != 0x41615252 || binary.LittleEndian.Uint32(info[484:]) != 0x61417272 {
		t.Error("no FSInfo sector")
	}
	if free := uint64(binary.LittleEndian.Uint32(info[488:])); free != clusters-1 {
		t.Errorf("FSInfo counts %d free clusters, want %d", free, clusters-1)
	}
	if backup := readSector(t, path, partitionStart+6); string(backup) != string(boot) {
		t.Error("the backup boot sector differs")
	}

	fat := readSector(t, path, int64(partitionStart+reserved))
	for i, want := range []uint32{0x0FFFFFF8, 0x0FFFFFFF, 0x0FFFFFFF} {
		if got := binary.LittleEndian.Uint32(fat[4*i:]); got != want {
			t.Errorf("FAT entry %d = %#x, want %#x", i, got, want)
		}
	}
	second := readSector(t, path, int64(partitionStart+reserved+fatSize))
	if string(second) != string(fat) {
		t.Error("the second FAT differs from the first")
	}
}

func TestExFATFillsThePartition(t *testing.T) {
	const size = 2 << 30
	boot := readSector(t, createTemp(t, size, ExFAT), partitionStart)
	if string(boot[3:11]) != "EXFAT   " {
		t.Fatalf("OEM name %q", boot[3:11])
	}
	if got := binary.LittleEndian.Uint64(boot[64:]); got != partitionStart {
		t.Errorf("PartitionOffset %d", got)
	}
	length := binary.LittleEndian.Uint64(boot[72:])
	if length != size/sectorSize-partitionStart {
		t.Errorf("VolumeLength %d, want %d", length, size/sectorSize-partitionStart)
	}
	heap := uint64(binary.LittleEndian.Uint32(boot[88:]))
	count := uint64(binary.LittleEndian.Uint32(boot[92:]))
	spc := uint64(1) << boot[109]
	if count != (length-heap)/spc {
		t.Errorf("%d clusters, want every whole cluster after the heap: %d", count, (length-heap)/spc)
	}
	if spc != 64 {
		t.Errorf("%d sectors per cluster; 32 KB is Windows' choice for 2 GB", spc)
	}
}

func TestCreateImageRefuses(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "taken.img")
	if err := os.WriteFile(existing, []byte("keep me"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CreateImage(existing, 64<<20, ExFAT, "X"); !errors.Is(err, fs.ErrExist) {
		t.Errorf("an existing file: %v, want ErrExist", err)
	}
	if data, _ := os.ReadFile(existing); string(data) != "keep me" {
		t.Errorf("the existing file was changed: %q", data)
	}

	for _, c := range []struct {
		name string
		size int64
		fs   FS
	}{
		{"below the minimum", MinImageSize - 1<<20, ExFAT},
		{"not whole sectors", 64<<20 + 100, ExFAT},
		{"FAT32 past 2 TiB", 3 << 40, FAT32},
		{"an unknown filesystem", 64 << 20, FS("ntfs")},
	} {
		path := filepath.Join(dir, "new.img")
		if err := CreateImage(path, c.size, c.fs, "X"); err == nil {
			t.Errorf("%s: created", c.name)
		}
		if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("%s: left a file behind", c.name)
		}
	}
}

func TestLabelFor(t *testing.T) {
	for _, c := range []struct {
		name string
		fs   FS
		want string
	}{
		{"exchange", FAT32, "EXCHANGE"},
		{"my files.v2", FAT32, "MY FILES_V2"},
		{"a-very-long-name", FAT32, "A-VERY-LONG"},
		{"Déjà vu", FAT32, "D_J_ VU"},
		{"Déjà vu", ExFAT, "Déjà vu"},
		{"Exchange disk 2026", ExFAT, "Exchange di"},
		{"...", FAT32, "EXCHANGE"},
		{"", ExFAT, "EXCHANGE"},
	} {
		if got := LabelFor(c.name, c.fs); got != c.want {
			t.Errorf("LabelFor(%q, %s) = %q, want %q", c.name, c.fs, got, c.want)
		}
	}
}
