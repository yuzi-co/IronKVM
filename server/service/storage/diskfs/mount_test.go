package diskfs

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

// These tests mount real images through a loop device, as the board does. They
// need root and the loop driver, so they skip in an unprivileged container;
// run the test binary as root to include them:
//
//	go test -c -tags novision -o diskfs.test ./service/storage/diskfs
//	sudo ./diskfs.test -test.run Mount -test.v
func needLoopMount(t *testing.T) {
	t.Helper()
	if os.Geteuid() != 0 {
		t.Skip("needs root to loop mount")
	}
	if _, err := os.Stat(loopControl); err != nil {
		t.Skip("no loop device support: ", err)
	}
	old := MountRoot
	MountRoot = t.TempDir()
	t.Cleanup(func() {
		CleanupStale()
		MountRoot = old
	})
}

func listNames(t *testing.T, image Image, rel string) string {
	t.Helper()
	entries, _, err := image.List(rel, 100)
	if err != nil {
		t.Fatalf("List(%q): %v", rel, err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name)
	}
	return strings.Join(names, ",")
}

func readFile(t *testing.T, image Image, rel string) string {
	t.Helper()
	f, _, err := image.Open(rel)
	if err != nil {
		t.Fatalf("Open(%q): %v", rel, err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestMountBrowsesAFATImage(t *testing.T) {
	needLoopMount(t)
	path := writeTemp(t, "stick.img", buildFAT16(t,
		fatFile{name: "HELLO.TXT", data: "hello from fat"},
		fatFile{name: "DOCS", dir: true},
		fatFile{name: "DOCS/NOTE.TXT", data: strings.Repeat("n", 1500)},
	))

	image, err := Mount(path, false)
	if err != nil {
		t.Fatal(err)
	}
	if image.Kind() != KindFAT16 || image.Writable() {
		t.Errorf("Kind %s, Writable %v", image.Kind(), image.Writable())
	}
	if got := listNames(t, image, "/"); got != "DOCS,HELLO.TXT" {
		t.Errorf("root lists %s", got)
	}
	if got := readFile(t, image, "DOCS/NOTE.TXT"); got != strings.Repeat("n", 1500) {
		t.Errorf("NOTE.TXT holds %d bytes", len(got))
	}
	if _, err := image.Write("NEW.TXT", strings.NewReader("x")); !errors.Is(err, ErrReadOnly) {
		t.Errorf("a read-only mount took a write: %v", err)
	}
	if err := image.Close(); err != nil {
		t.Fatal(err)
	}
	if open, _ := State(path); open {
		t.Error("the image is still open after Close")
	}
}

// A file written into an image is there when the image is mounted again, and
// the write leaves the mount point and loop device behind it free.
func TestMountWritesIntoAFATImage(t *testing.T) {
	needLoopMount(t)
	path := writeTemp(t, "exchange.img", buildFAT16(t, fatFile{name: "DOCS", dir: true}))

	image, err := Mount(path, true)
	if err != nil {
		t.Fatal(err)
	}
	if open, writable := State(path); !open || !writable {
		t.Errorf("State = %v, %v while mounted writable", open, writable)
	}
	// A reader shares the writable mount; a second writer is not needed.
	reader, err := Mount(path, false)
	if err != nil {
		t.Fatalf("a reader could not share the writable mount: %v", err)
	}
	if _, err := image.Write("DOCS/report.txt", strings.NewReader("written on the board")); err != nil {
		t.Fatal(err)
	}
	_ = reader.Close()
	if err := image.Close(); err != nil {
		t.Fatal(err)
	}

	again, err := Mount(path, false)
	if err != nil {
		t.Fatal(err)
	}
	defer again.Close()
	if got := readFile(t, again, "DOCS/report.txt"); got != "written on the board" {
		t.Errorf("read back %q", got)
	}
}

// A read-only mount in use refuses to become writable: the reader would see
// a filesystem change under a mount that promised it would not.
func TestAWriterWaitsForTheReaders(t *testing.T) {
	needLoopMount(t)
	path := writeTemp(t, "busy.img", buildFAT16(t))

	reader, err := Mount(path, false)
	if err != nil {
		t.Fatal(err)
	}
	if w, err := Mount(path, true); !errors.Is(err, ErrBusy) {
		if err == nil {
			_ = w.Close()
		}
		t.Fatalf("a writable mount over a reader: %v, want ErrBusy", err)
	}
	_ = reader.Close()

	w, err := Mount(path, true)
	if err != nil {
		t.Fatal(err)
	}
	_ = w.Close()
}

func TestMountBrowsesAPartitionedImage(t *testing.T) {
	needLoopMount(t)
	path := writeTemp(t, "disk.img", withMBR(buildFAT16(t, fatFile{name: "PART.TXT", data: "inside partition 1"}), 0x0E))

	image, err := Mount(path, false)
	if err != nil {
		t.Fatal(err)
	}
	defer image.Close()
	if got := readFile(t, image, "PART.TXT"); got != "inside partition 1" {
		t.Errorf("read %q", got)
	}
}

func TestMountBrowsesAnISO(t *testing.T) {
	needLoopMount(t)
	for _, bridge := range []bool{false, true} {
		path := writeTemp(t, "disc.iso", buildISO(t, bridge, map[string]string{
			"README.TXT": "read me",
			"BIG.BIN":    strings.Repeat("b", 5000),
		}))

		image, err := Mount(path, false)
		if err != nil {
			t.Fatalf("bridge %v: %v", bridge, err)
		}
		// Linux shows plain ISO 9660 names in lower case, without ";1".
		if got := listNames(t, image, "/"); got != "big.bin,readme.txt" {
			t.Errorf("bridge %v: root lists %s", bridge, got)
		}
		if got := readFile(t, image, "readme.txt"); got != "read me" {
			t.Errorf("bridge %v: readme.txt holds %q", bridge, got)
		}
		_ = image.Close()

		if w, err := Mount(path, true); !errors.Is(err, ErrReadOnly) {
			if err == nil {
				_ = w.Close()
			}
			t.Errorf("bridge %v: a writable ISO mount: %v, want ErrReadOnly", bridge, err)
		}
	}
}

func TestMountRefusesWhatItCannotRead(t *testing.T) {
	needLoopMount(t)
	path := writeTemp(t, "blank.img", make([]byte, 1<<20))
	if image, err := Mount(path, false); !errors.Is(err, ErrUnsupported) {
		if err == nil {
			_ = image.Close()
		}
		t.Fatalf("Mount of a blank image: %v, want ErrUnsupported", err)
	}
	entries, _ := os.ReadDir(MountRoot)
	if len(entries) != 0 {
		t.Errorf("left %d mount points", len(entries))
	}
}
