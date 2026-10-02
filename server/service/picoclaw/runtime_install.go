package picoclaw

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/service/extensions/addon"
)

// errPicoclawNotBundled is what an install answers on an image that does not
// carry PicoClaw. There is no download to fall back to.
var errPicoclawNotBundled = errors.New("PicoClaw is not included in this image")

func (s *Service) installRuntime() (string, *PicoclawError) {
	if s == nil {
		return "", newPicoclawError(CodeRuntimeUnavailable, "picoclaw service is unavailable")
	}
	s.ensureDependencies()
	log.Debugf("picoclaw install: start, binary=%s, source=%s", picoclawBinaryPath, picoclawBundledBinary)

	currentStatus := s.runtime.Get()
	if currentStatus.Installing {
		log.Debugf("picoclaw install: install already in progress")
		return "picoclaw installation is already in progress", nil
	}

	if installed, err := isPicoclawInstalled(); err == nil && installed && !installedPicoclawIsStale() {
		settings, _ := loadPicoclawGatewaySettings()
		log.Debugf("picoclaw install: binary already exists at %s", picoclawBinaryPath)
		s.runtime.Set(RuntimeStatus{
			Ready:           false,
			Installed:       true,
			Installing:      false,
			InstallProgress: 100,
			InstallStage:    "installed",
			InstallPath:     picoclawBinaryPath,
			ModelConfigured: settings.ModelConfigured,
			ModelName:       settings.ModelName,
			Status:          "installed",
			CheckedAt:       time.Now(),
		})
		return "picoclaw is already installed", nil
	}

	if _, err := readBundledPicoclawChecksum(); err != nil {
		log.Errorf("picoclaw install: %v", err)
		return "", newPicoclawError(CodeRuntimeUnavailable, err.Error())
	}

	s.runtime.Set(RuntimeStatus{
		Ready:           false,
		Installed:       false,
		Installing:      true,
		InstallProgress: 0,
		InstallStage:    "preparing",
		InstallPath:     picoclawBinaryPath,
		Status:          "installing",
		CheckedAt:       time.Now(),
	})

	go s.runInstallRuntime()
	return "picoclaw installation started", nil
}

func (s *Service) runInstallRuntime() {
	s.ensureDependencies()

	s.setInstallProgress("installing", 20, "")
	if err := installBundledPicoclaw(); err != nil {
		log.Errorf("picoclaw install: %v", err)
		s.finishInstallFailure("install_failed", err.Error())
		return
	}
	log.Debugf("picoclaw install: install completed successfully")

	s.runtime.Set(RuntimeStatus{
		Ready:           false,
		Installed:       true,
		Installing:      false,
		InstallProgress: 100,
		InstallStage:    "installed",
		InstallPath:     picoclawBinaryPath,
		Status:          "installed",
		CheckedAt:       time.Now(),
	})
}

// installBundledPicoclaw copies the image's PicoClaw to where it is installed,
// checks the copy against the image's sha256, records that sha256 beside it,
// and makes the add-on's link and bind.
func installBundledPicoclaw() error {
	digest, err := readBundledPicoclawChecksum()
	if err != nil {
		return err
	}
	destination := picoclawInstallDestination()
	if err := installPicoclawBinary(picoclawBundledBinary, destination, digest); err != nil {
		return err
	}
	if err := os.WriteFile(installedChecksumPath(destination), []byte(digest+"\n"), 0o644); err != nil {
		return fmt.Errorf("failed to record the installed sha256: %w", err)
	}
	if err := recordPicoclawAddon(); err != nil {
		return fmt.Errorf("failed to record the add-on: %w", err)
	}
	return nil
}

// refreshInstalledPicoclaw replaces an installed PicoClaw that is not the one
// the image carries: Sipeed's v0.2.8 from an older server, or the fork's build
// from an older image. It runs before every start, so a new image brings its
// PicoClaw with it. The settings are left alone.
func refreshInstalledPicoclaw() {
	if !installedPicoclawIsStale() {
		return
	}
	log.Infof("picoclaw: installing the PicoClaw this image carries over the one installed")
	if err := installBundledPicoclaw(); err != nil {
		log.Errorf("picoclaw: failed to install the PicoClaw this image carries: %v", err)
	}
}

// installedPicoclawIsStale reports whether the image carries a PicoClaw and
// the installed one is not it. It compares the sha256 recorded at install with
// the image's, and reads no binary.
func installedPicoclawIsStale() bool {
	want, err := readBundledPicoclawChecksum()
	if err != nil {
		return false
	}
	have, err := os.ReadFile(installedChecksumPath(picoclawInstallDestination()))
	return err != nil || strings.TrimSpace(string(have)) != want
}

func installedChecksumPath(binary string) string {
	return binary + ".sha256"
}

// readBundledPicoclawChecksum returns the sha256 the image records for its
// PicoClaw, in the sha256sum format the build writes, or errPicoclawNotBundled
// when the image carries no PicoClaw.
func readBundledPicoclawChecksum() (string, error) {
	if info, err := os.Stat(picoclawBundledBinary); err != nil || info.IsDir() {
		return "", errPicoclawNotBundled
	}
	raw, err := os.ReadFile(picoclawBundledChecksum)
	if err != nil {
		return "", errPicoclawNotBundled
	}
	fields := strings.Fields(string(raw))
	if len(fields) == 0 || !isValidSHA256Digest(fields[0]) {
		return "", fmt.Errorf("%s holds no sha256", picoclawBundledChecksum)
	}
	return strings.ToLower(fields[0]), nil
}

func isValidSHA256Digest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func (s *Service) setInstallProgress(stage string, progress int, lastError string) {
	s.ensureDependencies()
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	s.runtime.UpdateInstallStatus(func(status *RuntimeStatus) {
		status.Ready = false
		status.Installed = false
		status.Installing = true
		status.InstallProgress = progress
		status.InstallStage = stage
		status.Status = "installing"
		status.LastError = lastError
		status.CheckedAt = time.Now()
	})
}

func (s *Service) finishInstallFailure(status string, message string) {
	s.ensureDependencies()
	s.runtime.Set(RuntimeStatus{
		Ready:           false,
		Installed:       false,
		Installing:      false,
		InstallProgress: 0,
		InstallStage:    status,
		InstallPath:     picoclawBinaryPath,
		Status:          status,
		LastError:       message,
		CheckedAt:       time.Now(),
	})
}

// installPicoclawBinary copies source to destination through a temporary file
// and renames it into place only when the bytes written hash to wantSHA256.
func installPicoclawBinary(source string, destination string, wantSHA256 string) error {
	// The destination directory need not exist: on a distribution image it is
	// the add-on's own directory on /data, and the first install of PicoClaw on
	// a board is what creates it. Measured on the reference board on
	// 2026-09-22, before this line: "failed to create destination binary: open
	// /data/ironkvm/addons/picoclaw/picoclaw.tmp: no such file or directory",
	// twice, and the web UI offered to install again.
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	inFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to open picoclaw binary: %w", err)
	}
	defer inFile.Close()

	tempDestination := destination + ".tmp"
	outFile, err := os.OpenFile(tempDestination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create destination binary: %w", err)
	}

	hasher := sha256.New()
	if _, err := io.Copy(io.MultiWriter(outFile, hasher), inFile); err != nil {
		_ = outFile.Close()
		_ = os.Remove(tempDestination)
		return fmt.Errorf("failed to write destination binary: %w", err)
	}
	if err := outFile.Close(); err != nil {
		_ = os.Remove(tempDestination)
		return fmt.Errorf("failed to finalize destination binary: %w", err)
	}
	if got := hex.EncodeToString(hasher.Sum(nil)); got != strings.ToLower(wantSHA256) {
		_ = os.Remove(tempDestination)
		return fmt.Errorf("sha256 mismatch: %s is %s, want %s", source, got, wantSHA256)
	}
	if err := os.Chmod(tempDestination, 0o755); err != nil {
		_ = os.Remove(tempDestination)
		return fmt.Errorf("failed to set destination mode: %w", err)
	}
	if err := os.Rename(tempDestination, destination); err != nil {
		_ = os.Remove(tempDestination)
		return fmt.Errorf("failed to install picoclaw binary: %w", err)
	}
	return nil
}

// picoclawAddon is what PicoClaw needs back from the root filesystem after a
// new image: its binary at /usr/bin/picoclaw and its settings, the model and
// its API key among them, at /root/.picoclaw. It has no boot script: the server
// starts it from /kvmapp when the intent in /etc/kvm says so, and /etc/kvm is
// already on /data.
func picoclawAddon() addon.Spec {
	return addon.Spec{
		Name:  "picoclaw",
		Links: []addon.Link{{Path: picoclawBinaryPath, File: "picoclaw"}},
		Binds: []addon.Bind{{Dir: "/root/.picoclaw", Sub: "home"}},
	}
}

// picoclawInstallDestination is where the binary is written. On a distribution
// image that is the add-on's directory on /data, which the next image keeps.
func picoclawInstallDestination() string {
	if addon.OnData() {
		return filepath.Join(addon.Dir(picoclawAddon().Name), "picoclaw")
	}
	return picoclawBinaryPath
}

// recordPicoclawAddon makes the link and the bind on a distribution image, and
// does nothing anywhere else. Settings already in /root/.picoclaw are carried
// into the add-on before the bind covers them.
func recordPicoclawAddon() error {
	if !addon.OnData() {
		return nil
	}
	return addon.Record(picoclawAddon())
}
