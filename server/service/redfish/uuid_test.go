package redfish

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestLoadUUIDCreatesItOnceAndKeepsIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "redfish-uuid")

	first := LoadUUID(path)
	if _, err := uuid.Parse(first); err != nil {
		t.Fatalf("LoadUUID returned %q, not a UUID", first)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("the UUID was not written: %s", err)
	}
	if strings.TrimSpace(string(data)) != first {
		t.Fatalf("file holds %q, want %q", data, first)
	}

	if second := LoadUUID(path); second != first {
		t.Fatalf("second LoadUUID = %q, want %q", second, first)
	}
}

func TestLoadUUIDReplacesAFileThatIsNotAUUID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "redfish-uuid")
	if err := os.WriteFile(path, []byte("garbage\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	id := LoadUUID(path)
	if _, err := uuid.Parse(id); err != nil {
		t.Fatalf("LoadUUID returned %q, not a UUID", id)
	}
	if LoadUUID(path) != id {
		t.Fatal("the replacement UUID was not kept")
	}
}

func TestLoadUUIDStillAnswersWhenItCannotWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-dir", "redfish-uuid")

	if _, err := uuid.Parse(LoadUUID(path)); err != nil {
		t.Fatal("LoadUUID returned no UUID when the file could not be written")
	}
}
