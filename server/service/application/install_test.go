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

// useTempUpdateMarker points the stand-off marker at a scratch path.
func useTempUpdateMarker(t *testing.T) string {
	t.Helper()

	original := updateMarkerPath
	t.Cleanup(func() { updateMarkerPath = original })

	updateMarkerPath = filepath.Join(t.TempDir(), "nanokvm-updating")
	return updateMarkerPath
}

// S98supervise polls every five seconds and restarts a server it finds gone. An
// update stops the server and then moves /kvmapp, so without a stand-off the
// supervisor starts a server in the middle of that. On 2026-08-17 it did: the
// server it started died three seconds later with its own files being moved,
// and the board rebooted.
func TestInstallPreparedPackageMarksTheUpdate(t *testing.T) {
	source := useTempDirs(t)
	marker := useTempUpdateMarker(t)
	stubInstallHook(t, nil)

	if err := installPreparedPackage(source); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("the marker should outlive a successful install: %s", err)
	}
}

// It must outlive the install, because the restart that follows is inside the
// window it protects: the process that would remove it is the one being
// replaced. The next server to start clears it instead.
func TestTheUpdateMarkerIsNotRemovedOnSuccess(t *testing.T) {
	source := useTempDirs(t)
	marker := useTempUpdateMarker(t)
	stubInstallHook(t, nil)

	if err := installPreparedPackage(source); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if _, err := os.Stat(marker); os.IsNotExist(err) {
		t.Fatal("the marker was removed, so the supervisor can act during the restart")
	}
}

// An install that fails restores the old application and does not restart, so
// there is nothing left to protect. Leaving the marker would suspend the
// supervisor for the whole stand-off over a board that is already running.
func TestAFailedInstallClearsTheUpdateMarker(t *testing.T) {
	source := useTempDirs(t)
	marker := useTempUpdateMarker(t)
	stubInstallHook(t, errors.New("boom"))

	if err := installPreparedPackage(source); err == nil {
		t.Fatal("expected the hook failure to be reported")
	}

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("a failed install must not leave the supervisor standing off")
	}
}

// The next server to start is what ends the stand-off, so this runs at startup.
func TestClearUpdateMarkerRemovesIt(t *testing.T) {
	marker := useTempUpdateMarker(t)
	if err := os.WriteFile(marker, nil, 0o644); err != nil {
		t.Fatalf("failed to write the marker: %s", err)
	}

	ClearUpdateMarker()

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("the marker survived startup, so the supervisor stays stood off")
	}
}

// Every boot calls it and almost none of them are after an update.
func TestClearUpdateMarkerIsQuietWhenThereIsNoMarker(t *testing.T) {
	useTempUpdateMarker(t)
	ClearUpdateMarker()
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
