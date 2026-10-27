package application

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

type archiveEntry struct {
	name     string
	typeFlag byte
	data     string
}

func writeTestArchive(t *testing.T, entries []archiveEntry) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "update.tar.gz")
	out, err := os.Create(file)
	if err != nil {
		t.Fatal(err)
	}
	gzipWriter := gzip.NewWriter(out)
	tarWriter := tar.NewWriter(gzipWriter)
	for _, entry := range entries {
		header := &tar.Header{Name: entry.name, Typeflag: entry.typeFlag, Mode: 0o755}
		if entry.typeFlag == tar.TypeReg || entry.typeFlag == tar.TypeRegA {
			header.Size = int64(len(entry.data))
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if entry.data != "" {
			if _, err := tarWriter.Write([]byte(entry.data)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
	return file
}

func TestInspectArchiveRejectsUnsafeEntries(t *testing.T) {
	root := "nanokvm_1.2.3"
	for _, entry := range []archiveEntry{
		{name: "../escape", typeFlag: tar.TypeReg, data: "x"},
		{name: "/escape", typeFlag: tar.TypeReg, data: "x"},
		{name: root + "/link", typeFlag: tar.TypeSymlink},
	} {
		archive := writeTestArchive(t, []archiveEntry{{name: root, typeFlag: tar.TypeDir}, entry})
		if _, err := inspectUpdateArchive(archive, root); err == nil {
			t.Fatalf("unsafe entry %+v was accepted", entry)
		}
	}
}

func TestInspectAndExtractArchive(t *testing.T) {
	root := "nanokvm_1.2.3"
	archive := writeTestArchive(t, []archiveEntry{
		{name: root, typeFlag: tar.TypeDir},
		{name: root + "/version", typeFlag: tar.TypeReg, data: "1.2.3\n"},
		{name: root + "/server/NanoKVM-Server", typeFlag: tar.TypeReg, data: "server"},
		{name: root + "/kvm_system/kvm_system", typeFlag: tar.TypeReg, data: "system"},
		{name: root + "/system/init.d/S95nanokvm", typeFlag: tar.TypeReg, data: "init"},
	})
	info, err := inspectUpdateArchive(archive, root)
	if err != nil {
		t.Fatal(err)
	}
	if info.expandedBytes != 22 { // data totals: 6 + 6 + 6 + 4
		t.Fatalf("unexpected expanded size %d", info.expandedBytes)
	}
	destination := t.TempDir()
	extracted, err := extractUpdateArchive(archive, destination, root)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateExtractedPackage(extracted, "1.2.3"); err != nil {
		t.Fatal(err)
	}
}

// extractUpdateArchive checks every entry itself, not only through
// inspectUpdateArchive, and writes nothing outside the destination: not for a
// "../" or absolute name, not for a link entry, and not through a symbolic
// link already in the destination.
func TestExtractArchiveWritesNothingOutsideTheDestination(t *testing.T) {
	root := "nanokvm_1.2.3"
	cases := map[string][]archiveEntry{
		"dot-dot":  {{name: root + "/../escape", typeFlag: tar.TypeReg, data: "x"}},
		"absolute": {{name: "/escape", typeFlag: tar.TypeReg, data: "x"}},
		"symlink":  {{name: root + "/link", typeFlag: tar.TypeSymlink}},
		"hardlink": {{name: root + "/link", typeFlag: tar.TypeLink}},
	}
	for name, entries := range cases {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			destination := filepath.Join(parent, "dest")
			archive := writeTestArchive(t, append([]archiveEntry{{name: root, typeFlag: tar.TypeDir}}, entries...))
			if _, err := extractUpdateArchive(archive, destination, root); err == nil {
				t.Fatal("the entry was accepted")
			}
			if _, err := os.Lstat(filepath.Join(parent, "escape")); err == nil {
				t.Fatal("a file was written outside the destination")
			}
			if _, err := os.Lstat(filepath.Join(destination, root, "link")); err == nil {
				t.Fatal("a link was created")
			}
		})
	}

	t.Run("symlink-on-disk", func(t *testing.T) {
		outside := t.TempDir()
		destination := t.TempDir()
		if err := os.Symlink(outside, filepath.Join(destination, root)); err != nil {
			t.Fatal(err)
		}
		archive := writeTestArchive(t, []archiveEntry{
			{name: root, typeFlag: tar.TypeDir},
			{name: root + "/version", typeFlag: tar.TypeReg, data: "1.2.3\n"},
		})
		if _, err := extractUpdateArchive(archive, destination, root); err == nil {
			t.Fatal("extracted through a symlink")
		}
		if names, _ := os.ReadDir(outside); len(names) != 0 {
			t.Fatalf("%d entries were written outside the destination", len(names))
		}
	})
}
