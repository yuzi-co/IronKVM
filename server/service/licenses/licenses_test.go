package licenses

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// writeTree lays out a manifest directory the way make-manifest.sh does, and
// a secret beside it that nothing may serve.
func writeTree(t *testing.T) (dir, secret string) {
	t.Helper()
	base := t.TempDir()
	dir = filepath.Join(base, "ironkvm")
	files := map[string]string{
		"MANIFEST":           "component\tironkvm\t1\tGPL-3.0-only\n",
		"COMPONENTS":         "ironkvm: the fork\n",
		"NOTICE":             "notices\n",
		"texts/MIT":          "MIT License\n",
		"texts/GPL-2.0-only": "GPL\n",
	}
	for name, body := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	secret = filepath.Join(base, "secret")
	if err := os.WriteFile(secret, []byte("do not serve\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir, secret
}

func read(t *testing.T, m *Manifest, name string) (string, error) {
	t.Helper()
	f, size, err := m.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	body, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if int64(len(body)) != size {
		t.Fatalf("%s: size %d, read %d", name, size, len(body))
	}
	return string(body), nil
}

func TestListNamesEveryRegularFile(t *testing.T) {
	dir, _ := writeTree(t)
	files, err := New(dir).List()
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, file := range files {
		paths = append(paths, file.Path)
		if file.Size <= 0 {
			t.Errorf("%s: size %d", file.Path, file.Size)
		}
	}
	want := []string{"COMPONENTS", "MANIFEST", "NOTICE", "texts/GPL-2.0-only", "texts/MIT"}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("listed %v, want %v", paths, want)
	}
}

func TestListLeavesOutSymlinks(t *testing.T) {
	dir, secret := writeTree(t)
	if err := os.Symlink(secret, filepath.Join(dir, "texts", "leak")); err != nil {
		t.Skip("no symlinks here:", err)
	}
	files, err := New(dir).List()
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if strings.Contains(file.Path, "leak") {
			t.Fatalf("listed the symlink %s", file.Path)
		}
	}
}

func TestOpenServesFilesInTheTree(t *testing.T) {
	dir, _ := writeTree(t)
	m := New(dir)
	for name, want := range map[string]string{"NOTICE": "notices\n", "texts/MIT": "MIT License\n"} {
		got, err := read(t, m, name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got != want {
			t.Fatalf("%s: %q", name, got)
		}
	}
}

func TestOpenRefusesPathsOutOfTheTree(t *testing.T) {
	dir, secret := writeTree(t)
	m := New(dir)
	for _, name := range []string{
		"",
		".",
		"..",
		"../secret",
		"texts/../../secret",
		"texts/../NOTICE",
		"./NOTICE",
		"/etc/passwd",
		secret,
		`..\secret`,
		`texts\MIT`,
		"NOTICE\x00",
		"texts/",
		"texts//MIT",
	} {
		if _, err := read(t, m, name); !errors.Is(err, ErrBadPath) {
			t.Errorf("%q: %v, want ErrBadPath", name, err)
		}
	}
}

func TestOpenRefusesASymlinkOutOfTheTree(t *testing.T) {
	dir, secret := writeTree(t)
	if err := os.Symlink(secret, filepath.Join(dir, "texts", "leak")); err != nil {
		t.Skip("no symlinks here:", err)
	}
	if err := os.Symlink(filepath.Dir(secret), filepath.Join(dir, "up")); err != nil {
		t.Fatal(err)
	}
	m := New(dir)
	for _, name := range []string{"texts/leak", "up/secret"} {
		body, err := read(t, m, name)
		if err == nil {
			t.Errorf("%s: served %q", name, body)
		}
	}
}

func TestOpenServesOnlyRegularFiles(t *testing.T) {
	dir, _ := writeTree(t)
	if _, err := read(t, New(dir), "texts"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("a directory: %v, want ErrNotFound", err)
	}
}

func TestOpenReportsAMissingFile(t *testing.T) {
	dir, _ := writeTree(t)
	if _, err := read(t, New(dir), "texts/NOPE"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("%v, want ErrNotFound", err)
	}
}

func TestOpenRefusesAFileOverTheCap(t *testing.T) {
	dir, _ := writeTree(t)
	big := filepath.Join(dir, "BIG")
	f, err := os.Create(big)
	if err != nil {
		t.Fatal(err)
	}
	// Sparse, so the test writes nothing to disk.
	if err := f.Truncate(MaxFileSize + 1); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if _, err := read(t, New(dir), "BIG"); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("%v, want ErrTooLarge", err)
	}
}

func TestAMissingDirectoryIsNotInstalled(t *testing.T) {
	m := New(filepath.Join(t.TempDir(), "absent"))
	if _, err := m.List(); !errors.Is(err, ErrNotInstalled) {
		t.Fatalf("List: %v, want ErrNotInstalled", err)
	}
	if _, err := read(t, m, "NOTICE"); !errors.Is(err, ErrNotInstalled) {
		t.Fatalf("Open: %v, want ErrNotInstalled", err)
	}
}
