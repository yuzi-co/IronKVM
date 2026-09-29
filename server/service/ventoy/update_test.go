package ventoy

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newRelease builds the Linux archive of a Ventoy release, with its files
// stored as they are, for cat standing in for xzcat.
func newRelease(t *testing.T, version string, efiSize int) (archive []byte, files map[string][]byte) {
	t.Helper()
	files = map[string][]byte{
		bootName: bytes.Repeat([]byte{0xB0}, 512),
		coreName: bytes.Repeat([]byte{0xC0}, 4096),
		efiName:  make([]byte, efiSize),
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, m := range membersOf(version) {
		data := files[m.Name]
		if err := tw.WriteHeader(&tar.Header{Name: m.Member, Mode: 0o644, Size: int64(len(data)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes(), files
}

func writeTo(data []byte) func(string) error {
	return func(dst string) error { return os.WriteFile(dst, data, 0o644) }
}

func setupUpdate(t *testing.T) (*fakeBoard, *Service) {
	t.Helper()
	b, s, _ := newFakeBoard(t)
	oldXz := XzcatPath
	t.Cleanup(func() { XzcatPath = oldXz })
	XzcatPath = "cat"
	writeRelease(t)
	return b, s
}

func readAll(t *testing.T) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, name := range []string{bootName, coreName, efiName} {
		data, err := os.ReadFile(filepath.Join(Dir, name))
		if err != nil {
			t.Fatal(err)
		}
		out[name] = sum(data)
	}
	return out
}

func TestUpdateReplacesTheFilesAndRecordsTheRelease(t *testing.T) {
	_, s := setupUpdate(t)
	archive, files := newRelease(t, "9.9.9", efiSectors*sectorSize)

	if err := s.update("9.9.9", sum(archive), writeTo(archive), func(int) {}); err != nil {
		t.Fatal(err)
	}
	got := readAll(t)
	for name, data := range files {
		if got[name] != sum(data) {
			t.Fatalf("%s was not replaced", name)
		}
	}
	rec, ok := readRelease()
	if !ok || rec.Version != "9.9.9" || rec.Archive != sum(archive) || rec.Files[efiName] != sum(files[efiName]) {
		t.Fatalf("record %+v", rec)
	}
	if installedVersion() != "9.9.9" || s.status().Version != "9.9.9" {
		t.Fatal("the status does not show the new release")
	}
	for _, left := range []string{".update", bootName + ".old"} {
		if _, err := os.Stat(inDir(left)); err == nil {
			t.Fatalf("%s left behind", left)
		}
	}

	// An uninstall forgets the update, so a reinstall is the pinned release.
	if err := s.uninstall(); err != nil {
		t.Fatal(err)
	}
	if installedVersion() != Version {
		t.Fatalf("after uninstall the version is %s", installedVersion())
	}
}

func TestUpdateIsRefusedWhileTheDiskIsInADrive(t *testing.T) {
	b, s := setupUpdate(t)
	b.inDrive = true
	before := readAll(t)

	fetched := false
	err := s.update("9.9.9", "", func(string) error { fetched = true; return nil }, func(int) {})
	if !errors.Is(err, errInDrive) {
		t.Fatalf("update in a drive: %v", err)
	}
	if fetched {
		t.Fatal("downloaded while the disk was in a drive")
	}
	if u := s.Updater(nil); u.Start() == nil || !strings.Contains(u.Status().InUse, "eject") {
		t.Fatalf("the updater does not refuse: %+v", u.Status())
	}
	if after := readAll(t); after[efiName] != before[efiName] {
		t.Fatal("the files changed")
	}
}

func TestAFailedUpdateLeavesTheInstalledRelease(t *testing.T) {
	_, s := setupUpdate(t)
	before := readAll(t)

	good, _ := newRelease(t, "9.9.9", efiSectors*sectorSize)
	odd, _ := newRelease(t, "9.9.9", 1024)
	cases := map[string]struct {
		fetch func(string) error
		want  string
	}{
		"download fails":       {func(string) error { return errors.New("checksum mismatch") }, "checksum"},
		"wrong version inside": {writeTo(good), ""},
		"another layout":       {writeTo(odd), "VTOYEFI"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			version := "9.9.9"
			if name == "wrong version inside" {
				version = "9.9.10"
			}
			err := s.update(version, "", c.fetch, func(int) {})
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("update: %v", err)
			}
			if after := readAll(t); after[bootName] != before[bootName] || after[efiName] != before[efiName] {
				t.Fatal("a failed update changed the files")
			}
			if _, ok := readRelease(); ok {
				t.Fatal("a failed update left a record")
			}
		})
	}
}
