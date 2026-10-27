package atomicfile

import (
	"os"
	"path/filepath"
	"testing"
)

func dirEntries(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestWriteFileReplacesTheFileWithTheGivenMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, mode := range []os.FileMode{0o644, 0o600, 0o640} {
		if err := WriteFile(path, []byte("new"), mode); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "new" {
			t.Fatalf("content %q, want new", data)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		// The mode is set on the file, so the umask does not narrow it.
		if got := info.Mode().Perm(); got != mode {
			t.Fatalf("mode %o, want %o", got, mode)
		}
	}

	if names := dirEntries(t, dir); len(names) != 1 {
		t.Fatalf("directory holds %v, want only the file", names)
	}
}

func TestWriteFileCreatesAMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "uuid")
	if err := WriteFile(path, []byte("id\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != "id\n" {
		t.Fatalf("read %q, %v", data, err)
	}
}

func TestWriteFileNeedsTheDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "file")
	if err := WriteFile(path, []byte("x"), 0o600); err == nil {
		t.Fatal("wrote into a directory that does not exist")
	}
}

// A rename that fails leaves the old content and no temporary file behind.
func TestWriteFileCleansUpWhenTheRenameFails(t *testing.T) {
	dir := t.TempDir()
	// A non-empty directory where the file should go cannot be renamed over.
	target := filepath.Join(dir, "target")
	if err := os.MkdirAll(filepath.Join(target, "child"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := WriteFile(target, []byte("x"), 0o600); err == nil {
		t.Fatal("renamed over a non-empty directory")
	}
	if names := dirEntries(t, dir); len(names) != 1 || names[0] != "target" {
		t.Fatalf("directory holds %v, want only target", names)
	}
	if names := dirEntries(t, target); len(names) != 1 || names[0] != "child" {
		t.Fatalf("target holds %v, want it untouched", names)
	}
}
