package ventoy

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"NanoKVM-Server/service/storage"
)

// fakeBoard replaces the kernel and the drives with a model: a device that
// exists or not, and a disk drive that holds it or not.
type fakeBoard struct {
	kernel    bool
	exists    bool
	inDrive   bool
	assembled *Disk
	sources   sources
	removed   int
	failAt    string
	released  func()
}

func newFakeBoard(t *testing.T) (*fakeBoard, *Service, string) {
	t.Helper()
	b := &fakeBoard{kernel: true}
	images := t.TempDir()

	oldDir := Dir
	Dir = t.TempDir()
	saved := []any{kernelSupport, assembleDisk, removeDisk, diskExists, diskSize, insertDevice, ejectDrive, deviceDrive, registerDevice}
	t.Cleanup(func() {
		Dir = oldDir
		kernelSupport = saved[0].(func() bool)
		assembleDisk = saved[1].(func(string, string, string, *Disk, sources) error)
		removeDisk = saved[2].(func(string, string) error)
		diskExists = saved[3].(func(string) bool)
		diskSize = saved[4].(func(string) int64)
		insertDevice = saved[5].(func(string) error)
		ejectDrive = saved[6].(func(string) error)
		deviceDrive = saved[7].(func(string) (string, error))
		registerDevice = saved[8].(func(storage.Device))
	})

	kernelSupport = func() bool { return b.kernel }
	assembleDisk = func(name, uuid, node string, d *Disk, src sources) error {
		if b.failAt == "assemble" {
			return errors.New("assemble failed")
		}
		b.exists, b.assembled, b.sources = true, d, src
		return nil
	}
	removeDisk = func(name, node string) error {
		if b.inDrive {
			return errors.New("busy")
		}
		b.exists = false
		b.removed++
		return nil
	}
	diskExists = func(string) bool { return b.exists }
	diskSize = func(string) int64 {
		if b.assembled == nil {
			return 0
		}
		return int64(b.assembled.Sectors) * sectorSize
	}
	insertDevice = func(path string) error {
		if b.failAt == "insert" {
			return errors.New("insert failed")
		}
		if path != DevicePath {
			t.Fatalf("inserted %s", path)
		}
		b.inDrive = true
		return nil
	}
	ejectDrive = func(id string) error {
		if id != storage.DriveDisk {
			t.Fatalf("ejected %s", id)
		}
		b.inDrive = false
		return nil
	}
	deviceDrive = func(string) (string, error) {
		if b.inDrive {
			return storage.DriveDisk, nil
		}
		return "", nil
	}
	registerDevice = func(d storage.Device) { b.released = d.Released }

	s := New(Deps{ImageDir: images})
	return b, s, images
}

func writeRelease(t *testing.T) {
	t.Helper()
	boot := fakeBoot()
	for name, data := range map[string][]byte{bootName: boot.BootImg, coreName: boot.CoreImg, efiName: make([]byte, 1024)} {
		if err := os.WriteFile(filepath.Join(Dir, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func writeImage(t *testing.T, dir, name string, size int) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, make([]byte, size), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestInsertBuildsAndServesTheSelection(t *testing.T) {
	b, s, images := newFakeBoard(t)
	writeRelease(t)
	a := writeImage(t, images, "a.iso", 4096)
	c := writeImage(t, images, "sub/c.img", 1000)

	if err := s.setImages([]string{c, a, a}); err != nil {
		t.Fatal(err)
	}
	if err := s.insert(); err != nil {
		t.Fatal(err)
	}
	if !b.inDrive || b.assembled == nil {
		t.Fatal("the disk was not built and inserted")
	}
	if len(b.sources.Files) != 2 || b.sources.Files[0] != a || b.sources.Files[1] != c {
		t.Fatalf("sources %v", b.sources.Files)
	}
	head, err := os.ReadFile(filepath.Join(Dir, headName))
	if err != nil || len(head) != len(b.assembled.Head) {
		t.Fatalf("head.bin not written: %v", err)
	}
	if got := s.servedImages(); len(got) != 2 {
		t.Fatalf("served %v", got)
	}

	st := s.status()
	if !st.InDrive || st.Size != int64(b.assembled.Sectors)*sectorSize || len(st.Images) != 2 {
		t.Fatalf("status %+v", st)
	}
}

func TestTheSetIsFixedWhileInADrive(t *testing.T) {
	b, s, images := newFakeBoard(t)
	writeRelease(t)
	a := writeImage(t, images, "a.iso", 4096)
	if err := s.setImages([]string{a}); err != nil {
		t.Fatal(err)
	}
	if err := s.insert(); err != nil {
		t.Fatal(err)
	}

	if err := s.setImages(nil); !errors.Is(err, errInDrive) {
		t.Fatalf("set change in a drive: %v", err)
	}
	if err := s.uninstall(); !errors.Is(err, errInDrive) {
		t.Fatalf("uninstall in a drive: %v", err)
	}
	if err := s.eject(); err != nil {
		t.Fatal(err)
	}
	if b.inDrive || b.exists {
		t.Fatal("eject left the device")
	}
	if err := s.setImages(nil); err != nil {
		t.Fatal(err)
	}
}

func TestReleasedRemovesTheDevice(t *testing.T) {
	b, s, images := newFakeBoard(t)
	writeRelease(t)
	s.Start()
	a := writeImage(t, images, "a.iso", 4096)
	if err := s.setImages([]string{a}); err != nil {
		t.Fatal(err)
	}
	if err := s.insert(); err != nil {
		t.Fatal(err)
	}

	// Another image went into the disk drive.
	b.inDrive = false
	b.released()
	if b.exists || len(s.servedImages()) != 0 {
		t.Fatal("the device outlived its drive")
	}
}

func TestInsertRefusals(t *testing.T) {
	b, s, images := newFakeBoard(t)

	b.kernel = false
	if err := s.insert(); !errors.Is(err, errNoKernel) {
		t.Fatalf("no kernel: %v", err)
	}
	b.kernel = true
	if err := s.insert(); !errors.Is(err, errNotReady) {
		t.Fatalf("not installed: %v", err)
	}
	writeRelease(t)
	if err := s.insert(); !errors.Is(err, errNoImages) {
		t.Fatalf("no images: %v", err)
	}

	a := writeImage(t, images, "a.iso", 4096)
	if err := s.setImages([]string{a}); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(a); err != nil {
		t.Fatal(err)
	}
	if err := s.insert(); !errors.Is(err, errNoImages) {
		t.Fatalf("only a missing image: %v", err)
	}
	if st := s.status(); len(st.Missing) != 1 || st.Missing[0] != a {
		t.Fatalf("missing %v", st.Missing)
	}
}

func TestAFailedInsertLeavesNothing(t *testing.T) {
	b, s, images := newFakeBoard(t)
	writeRelease(t)
	a := writeImage(t, images, "a.iso", 4096)
	if err := s.setImages([]string{a}); err != nil {
		t.Fatal(err)
	}
	b.failAt = "insert"
	if err := s.insert(); err == nil {
		t.Fatal("insert succeeded")
	}
	if b.exists || len(s.servedImages()) != 0 {
		t.Fatal("a failed insert left the device")
	}
}

func TestSetImagesTakesOnlyImagesInTheImageDirectory(t *testing.T) {
	_, s, images := newFakeBoard(t)
	for _, bad := range []string{"/etc/shadow", filepath.Join(images, "../x.iso"), filepath.Join(images, "a.txt")} {
		if err := s.setImages([]string{bad}); !errors.Is(err, errNotImage) {
			t.Errorf("%s: %v", bad, err)
		}
	}
}

func TestStartRemovesAnIdleDeviceAndAdoptsAServedOne(t *testing.T) {
	b, s, images := newFakeBoard(t)
	b.exists = true
	s.Start()
	if b.exists {
		t.Fatal("an idle device from an earlier run was kept")
	}

	a := writeImage(t, images, "a.iso", 4096)
	if err := s.setImages([]string{a}); err != nil {
		t.Fatal(err)
	}
	b.exists, b.inDrive = true, true
	s2 := New(Deps{ImageDir: images})
	s2.Start()
	if !b.exists || len(s2.servedImages()) != 1 {
		t.Fatal("a device in a drive was not adopted")
	}
}

func TestIdentityIsKept(t *testing.T) {
	b, s, images := newFakeBoard(t)
	writeRelease(t)
	a := writeImage(t, images, "a.iso", 4096)
	if err := s.setImages([]string{a}); err != nil {
		t.Fatal(err)
	}
	if err := s.insert(); err != nil {
		t.Fatal(err)
	}
	first := append([]byte(nil), b.assembled.Head[:sectorSize]...)
	if err := s.eject(); err != nil {
		t.Fatal(err)
	}
	if err := s.insert(); err != nil {
		t.Fatal(err)
	}
	if string(first) != string(b.assembled.Head[:sectorSize]) {
		t.Fatal("the MBR changed between two builds of one set")
	}
}

func TestUniqueNames(t *testing.T) {
	got := uniqueNames([]string{"/data/a.iso", "/data/x/A.iso", "/data/y/a.iso", "/data/b.img"})
	want := []string{"a.iso", "A (2).iso", "a (3).iso", "b.img"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}
