package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// useImageRoot points the image directory at a fresh temporary one, and
// returns it with a second directory outside it.
func useImageRoot(t *testing.T) (root string, outside string) {
	t.Helper()

	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	root = filepath.Join(base, "data")
	outside = filepath.Join(base, "etc")
	for _, dir := range []string{root, outside} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	original := imageRoot
	imageRoot = root
	t.Cleanup(func() { imageRoot = original })
	return root, outside
}

func touch(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("image"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func symlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
}

// A link under the image directory that leads out of it would serve any
// file on the board to the host, /etc/shadow included.
func TestInsertRefusesALinkThatLeavesTheImageDirectory(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	useHidOnly(t, false)
	root, outside := useImageRoot(t)

	touch(t, filepath.Join(outside, "shadow.img"))
	symlink(t, filepath.Join(outside, "shadow.img"), filepath.Join(root, "x.img"))
	symlink(t, outside, filepath.Join(root, "etc"))

	for _, file := range []string{filepath.Join(root, "x.img"), filepath.Join(root, "etc", "shadow.img")} {
		if err := InsertDrive(DriveDisk, file, true); !errors.Is(err, ErrInvalidImage) {
			t.Fatalf("InsertDrive(%s) = %v, want ErrInvalidImage", file, err)
		}
	}
	if got := readAttr(t, dir, "lun.0", "file"); got != "" {
		t.Fatalf("lun.0/file is %q, want it untouched", got)
	}
}

func TestInsertRefusesADirectoryNamedLikeAnImage(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	useHidOnly(t, false)
	root, _ := useImageRoot(t)

	if err := os.Mkdir(filepath.Join(root, "x.iso"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := InsertDrive(DriveCdrom, filepath.Join(root, "x.iso"), true); !errors.Is(err, ErrInvalidImage) {
		t.Fatalf("InsertDrive(directory) = %v, want ErrInvalidImage", err)
	}
}

// A link inside the image directory is followed, and the drive serves the
// file it names.
func TestInsertFollowsALinkInsideTheImageDirectory(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	useHidOnly(t, false)
	root, _ := useImageRoot(t)

	real := filepath.Join(root, "win11.iso")
	touch(t, real)
	symlink(t, "win11.iso", filepath.Join(root, "latest.iso"))

	if err := InsertDrive(DriveCdrom, filepath.Join(root, "latest.iso"), true); err != nil {
		t.Fatalf("InsertDrive: %s", err)
	}
	if got := readAttr(t, dir, "lun.1", "file"); got != real {
		t.Fatalf("lun.1/file is %q, want %q", got, real)
	}
}

// Two names for one file are one image: a writable disk and a CD on it would
// corrupt it.
func TestInsertSeesThroughALinkToAnImageInTheOtherDrive(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	useHidOnly(t, false)
	root, _ := useImageRoot(t)

	touch(t, filepath.Join(root, "a.img"))
	symlink(t, "a.img", filepath.Join(root, "alias.img"))

	if err := InsertDrive(DriveCdrom, filepath.Join(root, "a.img"), true); err != nil {
		t.Fatal(err)
	}
	if err := InsertDrive(DriveDisk, filepath.Join(root, "alias.img"), false); !errors.Is(err, ErrInOtherDrive) {
		t.Fatalf("InsertDrive(alias) = %v, want ErrInOtherDrive", err)
	}
}
