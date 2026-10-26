// Parts of this file are ported from sipeed/NanoKVM
// (server/service/storage/diskfs/diskfs_test.go, commit 5572e1de, PR #957).
// Copyright (c) Sipeed. Licensed under the GNU General Public License v3.0.

package diskfs

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectFindsTheFilesystem(t *testing.T) {
	fat16 := buildFAT16(t, fatFile{name: "A.TXT", data: "a"})
	tests := []struct {
		name string
		data []byte
		want Volume
	}{
		{"fat12", bootSector("MSDOS5.0", 0x36, "FAT12   "), Volume{Kind: KindFAT12}},
		{"fat16", fat16, Volume{Kind: KindFAT16}},
		// A FAT32 volume also carries "FAT16" at 0x36 on some formatters;
		// the FAT32 field decides.
		{"fat32", func() []byte {
			b := bootSector("MSWIN4.1", 0x52, "FAT32   ")
			copy(b[0x36:], "FAT16   ")
			return b
		}(), Volume{Kind: KindFAT32}},
		{"exfat", bootSector("EXFAT   ", 0, ""), Volume{Kind: KindExFAT}},
		{"ntfs", bootSector("NTFS    ", 0, ""), Volume{Kind: KindNTFS}},
		{"iso9660", buildISO(t, false, map[string]string{"A.TXT": "a"}), Volume{Kind: KindISO9660}},
		{"iso9660 with a udf bridge", buildISO(t, true, map[string]string{"A.TXT": "a"}), Volume{Kind: KindISO9660, UDFBridge: true}},
		{"fat16 in an mbr partition", withMBR(fat16, 0x06), Volume{Kind: KindFAT16, Offset: 2048 * 512, Size: int64(len(fat16))}},
		{"fat16 in a gpt partition", withGPT(fat16), Volume{Kind: KindFAT16, Offset: 2048 * 512, Size: int64(len(fat16))}},
		{"an empty mbr partition", withMBR(make([]byte, 64<<10), 0x83), Volume{Kind: KindUnknown}},
		{"zeros", make([]byte, 64<<10), Volume{Kind: KindUnknown}},
		{"too small", []byte("tiny"), Volume{Kind: KindUnknown}},
	}

	for _, tt := range tests {
		got, err := Detect(writeTemp(t, "image.img", tt.data))
		if err != nil {
			t.Errorf("%s: %v", tt.name, err)
			continue
		}
		if got != tt.want {
			t.Errorf("%s: Detect = %+v, want %+v", tt.name, got, tt.want)
		}
	}
}

// Ported: the volume recognition sequence of a disc decides between ISO 9660
// and UDF, and ends at the first slot that holds no descriptor.
func TestDetectOptical(t *testing.T) {
	write := func(descriptors ...string) string {
		image := make([]byte, 0x8000+(len(descriptors)+2)*opticalSectorSize)
		for i, id := range descriptors {
			copy(image[0x8000+i*opticalSectorSize+1:], id)
		}
		return writeTemp(t, "disc.iso", image)
	}

	tests := []struct {
		name        string
		descriptors []string
		want        Volume
	}{
		{"iso9660", []string{"CD001", "CD001"}, Volume{Kind: KindISO9660}},
		{"iso9660 with udf bridge", []string{"CD001", "CD001", "BEA01", "NSR02", "TEA01"}, Volume{Kind: KindISO9660, UDFBridge: true}},
		{"pure udf 1.02", []string{"BEA01", "NSR02", "TEA01"}, Volume{Kind: KindUDF}},
		{"pure udf 2.x", []string{"BEA01", "NSR03", "TEA01"}, Volume{Kind: KindUDF}},
		{"extended area without nsr", []string{"BEA01", "TEA01"}, Volume{Kind: KindUnknown}},
		{"nothing", nil, Volume{Kind: KindUnknown}},
		{"nsr after a gap", []string{"BEA01", "", "NSR03"}, Volume{Kind: KindUnknown}},
	}

	for _, tt := range tests {
		got, err := Detect(write(tt.descriptors...))
		if err != nil {
			t.Errorf("%s: %v", tt.name, err)
			continue
		}
		if got != tt.want {
			t.Errorf("%s: Detect = %+v, want %+v", tt.name, got, tt.want)
		}
	}
}

func TestKindCapabilities(t *testing.T) {
	for kind, want := range map[Kind][2]bool{
		KindFAT12:   {true, true},
		KindFAT16:   {true, true},
		KindFAT32:   {true, true},
		KindExFAT:   {true, true},
		KindISO9660: {true, false},
		KindUDF:     {true, false},
		KindNTFS:    {false, false},
		KindUnknown: {false, false},
	} {
		if kind.Browsable() != want[0] || kind.Writable() != want[1] {
			t.Errorf("%s: Browsable %v Writable %v, want %v", kind, kind.Browsable(), kind.Writable(), want)
		}
	}
}

func TestIsTransferTemp(t *testing.T) {
	tests := map[string]bool{
		".kvm-upload-0a1b2c3d":  true,
		".kvm-upload-00000000":  true,
		".kvm-upload-0A1B2C3D":  false,
		".kvm-upload-0a1b2c3":   false,
		".kvm-upload-notes.txt": false,
		"ubuntu.iso":            false,
	}
	for name, want := range tests {
		if got := IsTransferTemp(name); got != want {
			t.Errorf("IsTransferTemp(%q) = %v, want %v", name, got, want)
		}
	}
}

// openTestDir is a directory standing in for a mounted medium.
func openTestDir(t *testing.T, writable bool) (Image, string) {
	t.Helper()
	dir := t.TempDir()
	image, err := OpenDir(dir, KindExFAT, writable)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = image.Close() })
	return image, dir
}

func mustWrite(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestListSortsDirectoriesFirstAndHidesTransferTemps(t *testing.T) {
	image, dir := openTestDir(t, false)
	mustWrite(t, filepath.Join(dir, "b.txt"), "bb")
	mustWrite(t, filepath.Join(dir, "A.txt"), "a")
	mustWrite(t, filepath.Join(dir, ".kvm-upload-0a1b2c3d"), "partial")
	mustWrite(t, filepath.Join(dir, "zdir", "inner.txt"), "x")

	entries, truncated, err := image.List("/", 100)
	if err != nil {
		t.Fatal(err)
	}
	if truncated {
		t.Error("a short listing was reported truncated")
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name)
	}
	if strings.Join(names, ",") != "zdir,A.txt,b.txt" {
		t.Fatalf("List = %v", names)
	}
	if !entries[0].IsDir || entries[2].Size != 2 {
		t.Errorf("entries = %+v", entries)
	}

	inner, _, err := image.List("zdir", 100)
	if err != nil || len(inner) != 1 || inner[0].Name != "inner.txt" {
		t.Fatalf("List(zdir) = %+v, %v", inner, err)
	}
}

func TestListStopsAtTheLimit(t *testing.T) {
	image, dir := openTestDir(t, false)
	for _, name := range []string{"a", "b", "c", "d"} {
		mustWrite(t, filepath.Join(dir, name), "")
	}
	entries, truncated, err := image.List("/", 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 || !truncated {
		t.Fatalf("List = %d entries, truncated %v; want 3, true", len(entries), truncated)
	}
}

// Nothing a medium holds may lead outside it: not a "..", not a link, not an
// absolute path. The medium is untrusted and the server runs as root.
func TestNothingResolvesOutsideTheMedium(t *testing.T) {
	image, dir := openTestDir(t, true)
	secret := filepath.Join(t.TempDir(), "secret")
	mustWrite(t, secret, "secret")
	if err := os.Symlink(secret, filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Dir(secret), filepath.Join(dir, "linkdir")); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(dir, "sub", "f"), "ok")

	for _, rel := range []string{"link", "linkdir/secret", "../secret", "sub/../../secret", secret} {
		if f, _, err := image.Open(rel); err == nil {
			_ = f.Close()
			t.Errorf("Open(%q) succeeded", rel)
		}
		if _, _, err := image.List(rel, 10); err == nil {
			t.Errorf("List(%q) succeeded", rel)
		}
	}
	for _, rel := range []string{"../escape.txt", "linkdir/escape.txt"} {
		if _, err := image.Write(rel, strings.NewReader("x")); err == nil {
			t.Errorf("Write(%q) succeeded", rel)
		}
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(secret), "escape.txt")); err == nil {
		t.Error("a write escaped the medium")
	}
	if err := image.Remove("linkdir/secret"); err == nil {
		t.Error("Remove through a link succeeded")
	}
	if _, err := os.Stat(secret); err != nil {
		t.Errorf("the file outside is gone: %v", err)
	}

	// A listing does not show a link at all.
	entries, _, _ := image.List("/", 10)
	for _, e := range entries {
		if e.Name == "link" || e.Name == "linkdir" {
			t.Errorf("List shows the link %q", e.Name)
		}
	}
}

func TestOpenReadsARegularFileOnly(t *testing.T) {
	image, dir := openTestDir(t, false)
	mustWrite(t, filepath.Join(dir, "docs", "a.txt"), "hello")

	if f, _, err := image.Open("docs"); err == nil {
		_ = f.Close()
		t.Error("Open opened a directory")
	}

	f, entry, err := image.Open("docs/a.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	data, _ := io.ReadAll(f)
	if string(data) != "hello" || entry.Size != 5 || entry.Name != "a.txt" {
		t.Errorf("Open read %q, entry %+v", data, entry)
	}
}

func TestWriteStoresTheFileAndLeavesNoTemporary(t *testing.T) {
	image, dir := openTestDir(t, true)
	if err := os.Mkdir(filepath.Join(dir, "in"), 0o755); err != nil {
		t.Fatal(err)
	}

	n, err := image.Write("in/new.bin", strings.NewReader("payload"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 7 {
		t.Errorf("Write = %d bytes, want 7", n)
	}
	data, err := os.ReadFile(filepath.Join(dir, "in", "new.bin"))
	if err != nil || string(data) != "payload" {
		t.Fatalf("stored %q, %v", data, err)
	}
	left, _ := os.ReadDir(filepath.Join(dir, "in"))
	if len(left) != 1 {
		t.Errorf("the directory holds %d entries, want 1", len(left))
	}

	// An existing name is never overwritten.
	if _, err := image.Write("in/new.bin", strings.NewReader("other")); !errors.Is(err, os.ErrExist) {
		t.Errorf("Write over an existing file: %v, want ErrExist", err)
	}
	if data, _ := os.ReadFile(filepath.Join(dir, "in", "new.bin")); string(data) != "payload" {
		t.Errorf("the existing file now holds %q", data)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("connection lost") }

// A transfer that fails leaves nothing behind, and the name stays free.
func TestAFailedWriteLeavesNothing(t *testing.T) {
	image, dir := openTestDir(t, true)

	if _, err := image.Write("lost.bin", io.MultiReader(strings.NewReader("half"), failingReader{})); err == nil {
		t.Fatal("Write succeeded")
	}
	left, _ := os.ReadDir(dir)
	if len(left) != 0 {
		t.Errorf("left behind %d entries", len(left))
	}
}

func TestAReadOnlyImageRefusesChanges(t *testing.T) {
	image, dir := openTestDir(t, false)
	mustWrite(t, filepath.Join(dir, "keep.txt"), "x")

	if _, err := image.Write("new.txt", strings.NewReader("x")); !errors.Is(err, ErrReadOnly) {
		t.Errorf("Write: %v, want ErrReadOnly", err)
	}
	if err := image.Remove("keep.txt"); !errors.Is(err, ErrReadOnly) {
		t.Errorf("Remove: %v, want ErrReadOnly", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "keep.txt")); err != nil {
		t.Errorf("the file is gone: %v", err)
	}
}

func TestRemoveTakesAFileOrADirectoryButNeverTheRoot(t *testing.T) {
	image, dir := openTestDir(t, true)
	mustWrite(t, filepath.Join(dir, "a.txt"), "x")
	mustWrite(t, filepath.Join(dir, "d", "e", "f.txt"), "x")

	for _, rel := range []string{"/", "", "."} {
		if err := image.Remove(rel); err == nil {
			t.Errorf("Remove(%q) succeeded", rel)
		}
	}
	if err := image.Remove("a.txt"); err != nil {
		t.Fatal(err)
	}
	if err := image.Remove("d"); err != nil {
		t.Fatal(err)
	}
	left, _ := os.ReadDir(dir)
	if len(left) != 0 {
		t.Errorf("left %d entries", len(left))
	}
}

func TestStatFSReportsTheFilesystem(t *testing.T) {
	image, _ := openTestDir(t, false)
	total, free, err := image.StatFS()
	if err != nil {
		t.Fatal(err)
	}
	if total == 0 || free > total {
		t.Errorf("StatFS = %d total, %d free", total, free)
	}
}
