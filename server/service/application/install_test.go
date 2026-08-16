package application

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// stubInstallHook records whether the hook ran and lets a test make it fail.
func stubInstallHook(t *testing.T, err error) *int {
	t.Helper()

	original := runInstallHook
	t.Cleanup(func() { runInstallHook = original })

	var calls int
	runInstallHook = func() error {
		calls++
		return err
	}

	return &calls
}

// useTempDirs points AppDir and BackupDir at a scratch tree, so no test touches
// the real /kvmapp. It returns a prepared source directory holding one file.
func useTempDirs(t *testing.T) string {
	t.Helper()

	originalApp, originalBackup := AppDir, BackupDir
	t.Cleanup(func() { AppDir, BackupDir = originalApp, originalBackup })

	root := t.TempDir()
	AppDir = filepath.Join(root, "kvmapp")
	BackupDir = filepath.Join(root, "old")
	if err := os.MkdirAll(AppDir, 0o755); err != nil {
		t.Fatalf("failed to create AppDir: %s", err)
	}

	source := filepath.Join(root, "source")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatalf("failed to create source: %s", err)
	}
	if err := os.WriteFile(filepath.Join(source, "marker"), []byte("new"), 0o644); err != nil {
		t.Fatalf("failed to write marker: %s", err)
	}

	return source
}

// The boot scripts are the half of the fork that a package alone cannot install:
// nothing copies kvmapp/system/init.d into /etc/init.d. Without this call a
// release ships them in the image and omits them from the tarball, and the two
// diverge with nothing to show it.
func TestInstallPreparedPackageRunsTheInstallHook(t *testing.T) {
	source := useTempDirs(t)
	calls := stubInstallHook(t, nil)

	if err := installPreparedPackage(source); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if *calls != 1 {
		t.Fatalf("expected the install hook to run once, ran %d times", *calls)
	}
}

// The application is already in place when the hook runs. Reporting the failure
// tells the operator the boot scripts are stale. Hiding it would leave a board
// whose application and boot scripts disagree, with nothing said about it.
func TestInstallPreparedPackageReportsAHookFailure(t *testing.T) {
	source := useTempDirs(t)
	stubInstallHook(t, errors.New("syntax check failed"))

	err := installPreparedPackage(source)
	if err == nil {
		t.Fatal("a failed install hook must be reported")
	}
	if !strings.Contains(err.Error(), "boot scripts") {
		t.Fatalf("the error should name what did not happen, got %q", err)
	}

	// The application must still be installed. Undoing it would replace a
	// partial update with a broken one.
	if _, statErr := os.Stat(filepath.Join(AppDir, "marker")); statErr != nil {
		t.Fatalf("the application should still be installed: %s", statErr)
	}
}

// An upstream image carries no hook. There is nothing to install and nothing to
// lose, so an update on such a board must not fail over a file that was never
// meant to be there.
func TestExecInstallHookSucceedsWhenTheScriptIsAbsent(t *testing.T) {
	original := AppDir
	t.Cleanup(func() { AppDir = original })
	AppDir = t.TempDir()

	if err := execInstallHook(); err != nil {
		t.Fatalf("a missing hook must not be an error: %s", err)
	}
}
