//go:build linux && ventoyimage

package ventoy

// The image tests build real Ventoy disks and check them with the kernel's
// exFAT driver and exfatprogs. They need root, the exfat module, device-mapper
// and fsck.exfat, so they run only with the ventoyimage tag, in a privileged
// container:
//
//	docker run --rm --privileged -v <repo>:/src -w /src/server golang:1.25 sh -c \
//	  'apt-get update && apt-get install -y exfatprogs && \
//	   go test -tags novision,ventoyimage -run Image -v ./service/ventoy/'
//
// VENTOY_RELEASE_DIR names a directory holding boot.bin, core.bin and
// vtoyefi.bin from the pinned release; without it the boot files are
// stand-ins. VENTOY_QEMU_ISOS (colon-separated) and VENTOY_QEMU_OUT make
// TestImageForQEMU write a disk of those ISOs for a boot test.

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"golang.org/x/sys/unix"
)

// testFile is a file the test writes: size bytes, random data in the given
// chunks and holes elsewhere.
type testFile struct {
	name   string
	size   int64
	chunks [][2]int64 // offset, length
}

func imageTestFiles() []testFile {
	const mb = 1 << 20
	files := []testFile{
		{name: "Ubuntu 24.04 Живой диск 日本語 😀 Ωmega.iso", size: 3*mb + 123, chunks: [][2]int64{{0, 3*mb + 123}}},
		{name: "empty.iso", size: 0},
		{name: "one-byte.img", size: 1, chunks: [][2]int64{{0, 1}}},
		{name: "sector.iso", size: 512, chunks: [][2]int64{{0, 512}}},
		// Over 4 GB, with data on both sides of the 4 GB mark and in the
		// last, partial sector.
		{name: "big-over-4g.img", size: 5<<30 + 511, chunks: [][2]int64{
			{0, mb}, {4<<30 - mb/2, mb}, {5<<30 - mb, mb + 511},
		}},
		{name: strings.Repeat("n", 200) + ".iso", size: 70 << 10, chunks: [][2]int64{{0, 70 << 10}}},
	}
	rng := rand.New(rand.NewSource(31))
	for i := range 150 {
		size := rng.Int63n(300 << 10)
		files = append(files, testFile{
			name:   fmt.Sprintf("many-%03d.iso", i),
			size:   size,
			chunks: [][2]int64{{0, size}},
		})
	}
	return files
}

func writeTestFiles(t *testing.T, dir string, files []testFile) ([]File, []string) {
	t.Helper()
	rng := rand.New(rand.NewSource(28))
	out := make([]File, len(files))
	paths := make([]string, len(files))
	for i, f := range files {
		p := filepath.Join(dir, fmt.Sprintf("src-%03d", i))
		fh, err := os.Create(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range f.chunks {
			buf := make([]byte, c[1])
			rng.Read(buf)
			if _, err := fh.WriteAt(buf, c[0]); err != nil {
				t.Fatal(err)
			}
		}
		if err := fh.Truncate(f.size); err != nil {
			t.Fatal(err)
		}
		if err := fh.Close(); err != nil {
			t.Fatal(err)
		}
		mod := time.Date(2026, 9, 28, 12, 0, i%60, 0, time.UTC)
		if err := os.Chtimes(p, mod, mod); err != nil {
			t.Fatal(err)
		}
		out[i] = File{Name: f.name, Size: f.size, ModTime: mod}
		paths[i] = p
	}
	return out, paths
}

// testBoot returns the boot files and the VTOYEFI image to use.
func testBoot(t *testing.T, dir string) (BootFiles, string) {
	t.Helper()
	if rel := os.Getenv("VENTOY_RELEASE_DIR"); rel != "" {
		boot, err := os.ReadFile(filepath.Join(rel, bootName))
		if err != nil {
			t.Fatal(err)
		}
		core, err := os.ReadFile(filepath.Join(rel, coreName))
		if err != nil {
			t.Fatal(err)
		}
		return BootFiles{BootImg: boot, CoreImg: core}, filepath.Join(rel, efiName)
	}
	efi := filepath.Join(dir, efiName)
	data := bytes.Repeat([]byte("VTOYEFI!"), efiSectors*sectorSize/8)
	if err := os.WriteFile(efi, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return fakeBoot(), efi
}

func tailFrom(paths []string, files []File) TailReader {
	return func(i int, buf []byte) error {
		f, err := os.Open(paths[i])
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		_, err = f.ReadAt(buf, files[i].Size-int64(len(buf)))
		return err
	}
}

// writeDisk assembles the disk by concatenation into out, sparse: what is
// zero is never written.
func writeDisk(t *testing.T, d *Disk, paths []string, efi, out string) {
	t.Helper()
	img, err := os.Create(out)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = img.Close() }()
	if err := img.Truncate(int64(d.Sectors) * sectorSize); err != nil {
		t.Fatal(err)
	}

	copyRange := func(src string, srcOff, dstOff, n int64) {
		f, err := os.Open(src)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = f.Close() }()
		buf := make([]byte, 1<<20)
		zero := make([]byte, 1<<20)
		for done := int64(0); done < n; {
			chunk := min(int64(len(buf)), n-done)
			if _, err := f.ReadAt(buf[:chunk], srcOff+done); err != nil && err != io.EOF {
				t.Fatal(err)
			}
			if !bytes.Equal(buf[:chunk], zero[:chunk]) {
				if _, err := img.WriteAt(buf[:chunk], dstOff+done); err != nil {
					t.Fatal(err)
				}
			}
			done += chunk
		}
	}

	for _, e := range d.Extents {
		start, n := int64(e.Start)*sectorSize, int64(e.Length)*sectorSize
		switch e.Kind {
		case ExtentHead:
			if _, err := img.WriteAt(d.Head[e.Offset*sectorSize:e.Offset*sectorSize+e.Length*sectorSize], start); err != nil {
				t.Fatal(err)
			}
		case ExtentFile:
			copyRange(paths[e.Index], int64(e.Offset)*sectorSize, start, n)
		case ExtentEFI:
			copyRange(efi, int64(e.Offset)*sectorSize, start, n)
		case ExtentZero:
		}
	}
}

func run(t *testing.T, name string, args ...string) string {
	t.Helper()
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, out)
	}
	return string(out)
}

// checkVolume runs fsck.exfat on the partition dev, mounts it and compares
// every file with its source.
func checkVolume(t *testing.T, dev string, files []File, paths []string) {
	t.Helper()
	out := run(t, "fsck.exfat", "-n", "-v", dev)
	t.Logf("fsck.exfat %s:\n%s", dev, out)
	if strings.Contains(strings.ToLower(out), "error") || strings.Contains(out, "corrupt") {
		t.Fatalf("fsck.exfat reports a problem")
	}

	mnt := t.TempDir()
	run(t, "mount", "-t", "exfat", "-o", "ro", dev, mnt)
	defer run(t, "umount", mnt)

	label := strings.TrimSpace(run(t, "blkid", "-o", "value", "-s", "LABEL", dev))
	if label != "Ventoy" {
		t.Errorf("label %q, want Ventoy", label)
	}

	entries, err := os.ReadDir(mnt)
	if err != nil {
		t.Fatal(err)
	}
	var got, want []string
	for _, e := range entries {
		got = append(got, e.Name())
	}
	for _, f := range files {
		want = append(want, f.Name)
	}
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("the volume lists\n%q\nwant\n%q", got, want)
	}

	for i, f := range files {
		if !utf8.ValidString(f.Name) {
			t.Fatalf("bad name %q", f.Name)
		}
		fi, err := os.Stat(filepath.Join(mnt, f.Name))
		if err != nil {
			t.Fatal(err)
		}
		if fi.Size() != f.Size {
			t.Errorf("%s: size %d, want %d", f.Name, fi.Size(), f.Size)
			continue
		}
		if !fi.ModTime().Equal(f.ModTime) {
			t.Errorf("%s: modified %s, want %s", f.Name, fi.ModTime().UTC(), f.ModTime)
		}
		if a, b := hashFile(t, filepath.Join(mnt, f.Name)), hashFile(t, paths[i]); a != b {
			t.Errorf("%s: contents differ", f.Name)
		}
	}
}

func hashFile(t *testing.T, name string) string {
	t.Helper()
	f, err := os.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.CopyBuffer(h, f, make([]byte, 1<<20)); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// attachPart attaches the disk's partition 1 from the image file as a loop
// device, through a second read-only device-mapper device when the image is
// itself a device, or by writing the partition to its own file.
func partitionFile(t *testing.T, disk *Disk, img, out string) string {
	t.Helper()
	src, err := os.Open(img)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = src.Close() }()
	dst, err := os.Create(out)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = dst.Close() }()
	n := int64(disk.Part2Start-part1Start) * sectorSize
	if err := dst.Truncate(n); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 1<<20)
	zero := make([]byte, 1<<20)
	for off := int64(0); off < n; off += int64(len(buf)) {
		chunk := min(int64(len(buf)), n-off)
		if _, err := src.ReadAt(buf[:chunk], part1Start*sectorSize+off); err != nil && err != io.EOF {
			t.Fatal(err)
		}
		if !bytes.Equal(buf[:chunk], zero[:chunk]) {
			if _, err := dst.WriteAt(buf[:chunk], off); err != nil {
				t.Fatal(err)
			}
		}
	}
	return out
}

func sameContents(t *testing.T, a, b string, n int64) {
	t.Helper()
	fa, err := os.Open(a)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fa.Close() }()
	fb, err := os.Open(b)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fb.Close() }()
	ba, bb := make([]byte, 4<<20), make([]byte, 4<<20)
	for off := int64(0); off < n; off += int64(len(ba)) {
		chunk := min(int64(len(ba)), n-off)
		if _, err := io.ReadFull(io.NewSectionReader(fa, off, chunk), ba[:chunk]); err != nil {
			t.Fatal(err)
		}
		if _, err := io.ReadFull(io.NewSectionReader(fb, off, chunk), bb[:chunk]); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(ba[:chunk], bb[:chunk]) {
			t.Fatalf("%s and %s differ in the %d bytes from %d", a, b, chunk, off)
		}
	}
	// Nothing past the end.
	if extra, _ := fa.ReadAt(ba[:1], n); extra != 0 {
		t.Fatalf("%s is longer than %d bytes", a, n)
	}
}

func buildTestDisk(t *testing.T) (dir string, disk *Disk, files []File, paths []string, efi string) {
	t.Helper()
	dir = os.Getenv("VENTOY_TEST_DIR")
	if dir == "" {
		dir = t.TempDir()
	} else {
		dir = filepath.Join(dir, t.Name())
		_ = os.RemoveAll(dir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
	}
	files, paths = writeTestFiles(t, dir, imageTestFiles())
	boot, efi := testBoot(t, dir)
	disk, err := BuildDisk(files, tailFrom(paths, files), boot, testIdentity)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d files, disk of %d sectors, head of %d bytes, %d extents, zero run %d sectors",
		len(files), disk.Sectors, len(disk.Head), len(disk.Extents), disk.ZeroSectors)
	return dir, disk, files, paths, efi
}

// TestImageByConcatenation assembles the disk as a file, the way the extent
// map says, and checks partition 1.
func TestImageByConcatenation(t *testing.T) {
	dir, disk, files, paths, efi := buildTestDisk(t)
	img := filepath.Join(dir, "disk.img")
	writeDisk(t, disk, paths, efi, img)

	part := partitionFile(t, disk, img, filepath.Join(dir, "part1.img"))
	loop, err := attachLoop(part)
	if err != nil {
		t.Fatal(err)
	}
	defer loop.Detach()
	checkVolume(t, loop.Path, files, paths)
}

// TestImageThroughDeviceMapper assembles the disk the way the board does,
// with loop devices and the device-mapper ioctls, compares it with the
// concatenated image, and checks partition 1 through a second linear
// device over the first.
func TestImageThroughDeviceMapper(t *testing.T) {
	if !kernelHasDM() {
		t.Fatal("this kernel has no device-mapper")
	}
	dir, disk, files, paths, efi := buildTestDisk(t)
	head := filepath.Join(dir, headName)
	if err := os.WriteFile(head, disk.Head, 0o644); err != nil {
		t.Fatal(err)
	}

	const name, node = "ventoytest", "/dev/mapper/ventoytest"
	_ = disassemble(name+"-p1", node+"-p1")
	_ = disassemble(name, node)
	if err := assemble(name, "IRONKVM-VENTOY-TEST", node, disk, sources{Head: head, EFI: efi, Files: paths}); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := disassemble(name, node); err != nil {
			t.Errorf("disassemble: %s", err)
		}
		waitLoopsFree(t, dir)
	}()

	if got := deviceSize(name); got != int64(disk.Sectors)*sectorSize {
		t.Fatalf("device of %d bytes, want %d", got, disk.Sectors*sectorSize)
	}
	// The device is read-only: the kernel says so, and refuses a write.
	dev, _, err := dmStatus(name)
	if err != nil {
		t.Fatal(err)
	}
	ro, err := os.ReadFile(fmt.Sprintf("/sys/dev/block/%d:%d/ro", unix.Major(dev), unix.Minor(dev)))
	if err != nil || strings.TrimSpace(string(ro)) != "1" {
		t.Fatalf("ro is %q, %v", ro, err)
	}
	if f, err := os.OpenFile(node, os.O_WRONLY, 0); err == nil {
		buf := make([]byte, 4096)
		_, werr := f.Write(buf)
		_ = f.Close()
		if werr == nil {
			t.Fatal("a write to the device succeeded")
		}
	}

	img := filepath.Join(dir, "disk.img")
	writeDisk(t, disk, paths, efi, img)
	sameContents(t, node, img, int64(disk.Sectors)*sectorSize)

	p1, err := dmCreate(name+"-p1", "", []dmTarget{{
		Start:  0,
		Length: disk.Part2Start - part1Start,
		Type:   "linear",
		Params: fmt.Sprintf("%d:%d %d", unix.Major(dev), unix.Minor(dev), part1Start),
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := makeNode(node+"-p1", p1); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := disassemble(name+"-p1", node+"-p1"); err != nil {
			t.Errorf("disassemble p1: %s", err)
		}
	}()
	checkVolume(t, node+"-p1", files, paths)
}

// waitLoopsFree checks that no loop device still reads a file in dir once
// the device is gone: autoclear frees them.
func waitLoopsFree(t *testing.T, dir string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		held := []string{}
		backing, _ := filepath.Glob("/sys/block/loop*/loop/backing_file")
		for _, b := range backing {
			data, _ := os.ReadFile(b)
			if strings.HasPrefix(strings.TrimSpace(string(data)), dir) {
				held = append(held, b)
			}
		}
		if len(held) == 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("loop devices still attached: %v", held)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// TestImageForQEMU writes a disk of real ISOs, for a boot test.
func TestImageForQEMU(t *testing.T) {
	isos, out := os.Getenv("VENTOY_QEMU_ISOS"), os.Getenv("VENTOY_QEMU_OUT")
	if isos == "" || out == "" || os.Getenv("VENTOY_RELEASE_DIR") == "" {
		t.Skip("set VENTOY_QEMU_ISOS, VENTOY_QEMU_OUT and VENTOY_RELEASE_DIR")
	}
	paths := strings.Split(isos, ":")
	names := uniqueNames(paths)
	files := make([]File, len(paths))
	for i, p := range paths {
		fi, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		files[i] = File{Name: names[i], Size: fi.Size(), ModTime: fi.ModTime()}
	}
	boot, efi := testBoot(t, t.TempDir())
	disk, err := BuildDisk(files, tailFrom(paths, files), boot, testIdentity)
	if err != nil {
		t.Fatal(err)
	}
	writeDisk(t, disk, paths, efi, out)
	t.Logf("wrote %s: %d sectors, partition 2 at %d", out, disk.Sectors, disk.Part2Start)
}

// TestImageInstallRealRelease downloads the pinned release and checks every
// pin. It needs the network and xzcat.
func TestImageInstallRealRelease(t *testing.T) {
	if os.Getenv("VENTOY_REAL_INSTALL") == "" {
		t.Skip("set VENTOY_REAL_INSTALL=1")
	}
	old := Dir
	Dir = t.TempDir()
	t.Cleanup(func() { Dir = old })
	if err := install(); err != nil {
		t.Fatal(err)
	}
	if !installed() {
		t.Fatal("not installed")
	}
	entries, _ := os.ReadDir(Dir)
	for _, e := range entries {
		fi, _ := e.Info()
		t.Logf("%s %d", e.Name(), fi.Size())
	}
}
