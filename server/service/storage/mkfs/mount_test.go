package mkfs

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"NanoKVM-Server/service/storage/diskfs"
)

// These tests mount a created image through a loop device, as the board's
// file browser does. They need root and the loop driver, so they skip in an
// unprivileged container; run the test binary as root to include them:
//
//	go test -c -tags novision -o mkfs.test ./service/storage/mkfs
//	sudo ./mkfs.test -test.run Mount -test.v
func needLoopMount(t *testing.T) {
	t.Helper()
	if os.Geteuid() != 0 {
		t.Skip("needs root to loop mount")
	}
	if _, err := os.Stat("/dev/loop-control"); err != nil {
		t.Skip("no loop device support: ", err)
	}
	old := diskfs.MountRoot
	diskfs.MountRoot = t.TempDir()
	t.Cleanup(func() {
		diskfs.CleanupStale()
		diskfs.MountRoot = old
	})
}

// fsck checks the partition with the host's own checker, when it has one:
// the kernel mounting a volume is not proof that a careful reader accepts it.
func fsck(t *testing.T, path string, fsType FS) {
	t.Helper()
	tool, args := "fsck.exfat", []string{"-n"}
	if fsType == FAT32 {
		tool, args = "fsck.vfat", []string{"-n", "-v"}
	}
	if _, err := exec.LookPath(tool); err != nil {
		t.Logf("no %s here, the volume is not checked by it", tool)
		return
	}
	// The checkers read a partition, not a disk: hand them the volume
	// through a read-only loop device at the partition's offset.
	out, err := exec.Command("losetup", "--show", "-f", "-r", "-o",
		"1048576", path).Output()
	if err != nil {
		t.Logf("losetup: %v; the volume is not checked", err)
		return
	}
	dev := strings.TrimSpace(string(out))
	defer func() { _ = exec.Command("losetup", "-d", dev).Run() }()

	res, err := exec.Command(tool, append(args, dev)...).CombinedOutput()
	if err != nil {
		t.Errorf("%s %s: %v\n%s", tool, dev, err, res)
	} else {
		t.Logf("%s: clean\n%s", tool, res)
	}
}

// A created image mounts writable, takes a file, and gives it back when it is
// mounted again; its free space is about its size.
func TestMountCreatedImage(t *testing.T) {
	needLoopMount(t)

	// 256 MiB takes the smallest clusters, 20 GiB the 16 KiB (FAT32) and
	// 32 KiB (exFAT) ones Windows picks for a disk of that size.
	for _, size := range []int64{256 << 20, 20 << 30} {
		for _, fsType := range []FS{ExFAT, FAT32} {
			t.Run(fmt.Sprintf("%s-%dMiB", fsType, size>>20), func(t *testing.T) {
				testMountCreated(t, size, fsType)
			})
		}
	}
}

func testMountCreated(t *testing.T, size int64, fsType FS) {
	t.Helper()
	path := createTemp(t, size, fsType)
	fsck(t, path, fsType)

	image, err := diskfs.Mount(path, true)
	if err != nil {
		t.Fatal(err)
	}
	if string(image.Kind()) != string(fsType) || !image.Writable() {
		t.Errorf("Kind %s, Writable %v", image.Kind(), image.Writable())
	}
	total, free, err := image.StatFS()
	if err != nil {
		t.Fatal(err)
	}
	if want := uint64(size) * 9 / 10; total < want || free < want {
		t.Errorf("total %d, free %d of a %d byte image", total, free, size)
	}
	entries, _, err := image.List("/", 100)
	if err != nil || len(entries) != 0 {
		t.Errorf("a new image lists %v, %v", entries, err)
	}
	body := strings.Repeat("exchanged with the host\n", 4096)
	if _, err := image.Write("report.txt", strings.NewReader(body)); err != nil {
		t.Fatal(err)
	}
	if err := image.Close(); err != nil {
		t.Fatal(err)
	}
	fsck(t, path, fsType)

	again, err := diskfs.Mount(path, false)
	if err != nil {
		t.Fatal(err)
	}
	defer again.Close()
	f, _, err := again.Open("report.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	data, _ := io.ReadAll(f)
	if string(data) != body {
		t.Errorf("read back %d bytes, want %d", len(data), len(body))
	}
}
