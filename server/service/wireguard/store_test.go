package wireguard

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func useTempStore(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "kvm", "wireguard.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	old := configFile
	configFile = path
	t.Cleanup(func() { configFile = old })
	return path
}

func TestStoreRoundTripIsPrivate(t *testing.T) {
	path := useTempStore(t)

	c := validConfig()
	c.StartAtBoot = true
	c.Peers[0].PresharedKey = testPSK
	if err := Save(c); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("the file is %o, want 600", perm)
	}

	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, c) {
		t.Errorf("got %+v, want %+v", got, c)
	}

	// No temporary file is left beside it.
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Errorf("%d files in the directory", len(entries))
	}
}

func TestStoreTightensAnExistingFile(t *testing.T) {
	path := useTempStore(t)
	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Save(validConfig()); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(path)
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("the file is %o, want 600", perm)
	}
}

func TestLoadMissingIsEmpty(t *testing.T) {
	useTempStore(t)
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.Configured() {
		t.Error("a board without the file is configured")
	}
}

func TestRemove(t *testing.T) {
	path := useTempStore(t)
	if err := Save(validConfig()); err != nil {
		t.Fatal(err)
	}
	if err := Remove(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("the file is still there")
	}
	if err := Remove(); err != nil {
		t.Errorf("removing twice: %v", err)
	}
}
