package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// withSecretPath points the persistence at a temporary file for one test.
func withSecretPath(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), ".jwt_secret")

	original := jwtSecretPath
	t.Cleanup(func() { jwtSecretPath = original })
	jwtSecretPath = path

	return path
}

func requirePOSIXPermissions(t *testing.T) {
	t.Helper()

	if runtime.GOOS == "windows" {
		t.Skip("file modes are not POSIX permissions here")
	}
}

// The defect this closes. A key generated into memory on every start
// invalidated every session on every restart, and a restart on this board is
// not a quick event.
func TestTheSecretKeySurvivesARestart(t *testing.T) {
	withSecretPath(t)

	first, err := loadOrCreateSecretKey()
	if err != nil {
		t.Fatalf("first start: %s", err)
	}
	if first == "" {
		t.Fatal("first start produced no key")
	}

	second, err := loadOrCreateSecretKey()
	if err != nil {
		t.Fatalf("second start: %s", err)
	}

	if second != first {
		t.Fatal("the second start signed sessions with a different key")
	}
}

func TestTheStoredKeyIsReadableOnlyByItsOwner(t *testing.T) {
	requirePOSIXPermissions(t)

	path := withSecretPath(t)

	if _, err := loadOrCreateSecretKey(); err != nil {
		t.Fatalf("first start: %s", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %s", err)
	}

	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("mode is %o, want 600", perm)
	}
}

// A restore, or a release that wrote the file less carefully, can leave the key
// readable by more than root. The key is already exposed by then, so repairing
// the mode is worth more than refusing to use it.
func TestALooseModeIsTightened(t *testing.T) {
	requirePOSIXPermissions(t)

	path := withSecretPath(t)

	key, err := loadOrCreateSecretKey()
	if err != nil {
		t.Fatalf("first start: %s", err)
	}

	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatalf("chmod: %s", err)
	}

	again, err := loadOrCreateSecretKey()
	if err != nil {
		t.Fatalf("second start: %s", err)
	}
	if again != key {
		t.Fatal("the key changed when only its mode was wrong")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %s", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("mode is %o after a start, want 600", perm)
	}
}

// A truncated or empty file must never become the signing key for every session
// on the device. There is no repair for it and no reason to want one, so it is
// replaced.
func TestAnUnusableStoredKeyIsReplaced(t *testing.T) {
	path := withSecretPath(t)

	for _, content := range []string{"", "\n", "short", "   \n  "} {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("write %q: %s", content, err)
		}

		key, err := loadOrCreateSecretKey()
		if err != nil {
			t.Fatalf("start with %q stored: %s", content, err)
		}

		if len(key) < minSecretKeyLength {
			t.Fatalf("start with %q stored produced key %q", content, key)
		}

		stored, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read back: %s", err)
		}
		if len(stored) < minSecretKeyLength {
			t.Fatalf("start with %q stored left %q on disk", content, stored)
		}
	}
}

// A board that cannot write the file still has to serve. Losing persistence
// costs the behaviour this change replaces; refusing to start costs the KVM.
func TestAnUnwritablePathStillYieldsAKey(t *testing.T) {
	original := jwtSecretPath
	t.Cleanup(func() { jwtSecretPath = original })

	// A path whose parent is a file, so both MkdirAll and the write fail.
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("write blocker: %s", err)
	}
	jwtSecretPath = filepath.Join(blocker, ".jwt_secret")

	key, err := loadOrCreateSecretKey()
	if err != nil {
		t.Fatalf("an unwritable path must not stop the server: %s", err)
	}
	if len(key) < minSecretKeyLength {
		t.Fatalf("key is %q", key)
	}
}

// The entropy failure is still fatal. A key nobody could generate is not one to
// carry on without, and the caller turns this into log.Fatalf.
func TestNoEntropyIsStillAnError(t *testing.T) {
	withSecretPath(t)

	originalReader := secretKeyReader
	t.Cleanup(func() { secretKeyReader = originalReader })
	secretKeyReader = failingReader{}

	if key, err := loadOrCreateSecretKey(); err == nil {
		t.Fatalf("expected an error, got key %q", key)
	}
}

// The temporary file the atomic write uses must not be left behind, because the
// directory it lands in is the one an operator reads to see what a board holds.
func TestNoTemporaryFileIsLeftBehind(t *testing.T) {
	path := withSecretPath(t)

	if _, err := loadOrCreateSecretKey(); err != nil {
		t.Fatalf("first start: %s", err)
	}

	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatalf("read dir: %s", err)
	}

	if len(entries) != 1 || entries[0].Name() != filepath.Base(path) {
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}

		t.Fatalf("directory holds %v, want only %s", names, filepath.Base(path))
	}
}
