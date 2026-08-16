package application

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"NanoKVM-Server/utils"
	log "github.com/sirupsen/logrus"
)

// runInstallHook is a variable so tests can avoid running a shell script.
var runInstallHook = execInstallHook

// installHookTimeout bounds the hook. It copies a handful of small files, so a
// run that takes a minute is a run that has hung.
const installHookTimeout = 60 * time.Second

// execInstallHook runs the install script the package carries.
//
// The script installs the boot scripts, which a package update otherwise cannot
// touch: nothing else copies kvmapp/system/init.d into /etc/init.d, so without
// it a release would ship the fork's boot behaviour in the SD image and omit it
// from the tarball.
//
// An upstream image carries no such script. There is nothing to install and
// nothing to lose, so its absence is not an error.
func execInstallHook() error {
	path := filepath.Join(AppDir, "system", "install.sh")

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if info.Mode()&0o111 == 0 {
		return fmt.Errorf("%s is not executable", path)
	}

	ctx, cancel := context.WithTimeout(context.Background(), installHookTimeout)
	defer cancel()

	output, err := exec.CommandContext(ctx, path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}

	log.Infof("install hook: %s", strings.TrimSpace(string(output)))
	return nil
}

var (
	mutex      sync.Mutex
	isUpdating bool
)

func acquireUpdateLock() bool {
	mutex.Lock()
	defer mutex.Unlock()

	if isUpdating {
		return false
	}
	isUpdating = true
	return true
}

func releaseUpdateLock() {
	mutex.Lock()
	defer mutex.Unlock()
	isUpdating = false
}

func installPreparedPackage(sourceDir string) error {
	if err := backupCurrentApp(); err != nil {
		return err
	}

	if err := applyUpdate(sourceDir); err != nil {
		return err
	}

	if err := utils.ChmodRecursively(AppDir, 0o755); err != nil {
		return fmt.Errorf("failed to chmod: %w", err)
	}

	// The application is in place by now, and BackupDir holds the previous one.
	// A failure here is reported rather than undone: an application that updated
	// while its boot scripts did not is a partial state the operator has to know
	// about, and rolling the application back would not repair the scripts.
	if err := runInstallHook(); err != nil {
		log.Errorf("install hook failed: %s", err)
		return fmt.Errorf("application installed but the boot scripts were not updated: %w", err)
	}

	return nil
}

func backupCurrentApp() error {
	if err := os.RemoveAll(BackupDir); err != nil {
		return fmt.Errorf("failed to remove backup: %w", err)
	}

	if err := utils.MoveFilesRecursively(AppDir, BackupDir); err != nil {
		return fmt.Errorf("failed to backup app: %w", err)
	}

	return nil
}

func applyUpdate(sourceDir string) error {
	if err := utils.MoveFilesRecursively(sourceDir, AppDir); err != nil {
		// Try to restore backup on failure
		if restoreErr := utils.MoveFilesRecursively(BackupDir, AppDir); restoreErr != nil {
			log.Errorf("Failed to restore backup after update failure: %v", restoreErr)
		}
		return fmt.Errorf("failed to move update in place: %w", err)
	}
	return nil
}
