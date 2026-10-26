package picoclaw

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/upstream"
	"NanoKVM-Server/utils"
)

// Variables rather than constants so the tests can point them at a scratch
// directory and a local server.
var (
	// picoclawFetchFallback is where the release is fetched off a
	// distribution image. On one it is the add-on's directory on /data.
	picoclawFetchFallback = "/root/.picoclaw-fetch"
	// picoclawLicensesFallback is where the release's license files go off a
	// distribution image. On one they go to the add-on's directory.
	picoclawLicensesFallback = "/usr/share/licenses/picoclaw"
	// picoclawReleaseURLOverride is a test aid, not a setting: a file whose
	// one line replaces the release's URL, so a board can install from a
	// local server before the release is published. The asset is still held
	// to the pinned SHA-256, so it can only name another copy of the same
	// file, and plain http is allowed for it for that reason. Nothing creates
	// it, and a test that does removes it afterwards.
	picoclawReleaseURLOverride = "/data/ironkvm/picoclaw-release-url"
	// downloads fetches the release over HTTPS, bounded and checked against
	// the pin.
	downloads = upstream.Default
	// freeBytes reports the space available to an unprivileged writer in
	// the filesystem that holds path.
	freeBytes = statfsFree
)

const picoclawDownloadTimeout = 10 * time.Minute

// spaceMargin is kept free beyond the archive and the unpacked binary: the
// license files, the add-on's small files, and room for the settings to be
// written afterwards.
const spaceMargin = 2 << 20

var errNoSpace = errors.New("not enough free space")

// installReleasePicoclaw downloads the pinned release, checks the archive and
// then the binary in it against their pins, and moves the binary into place
// with one rename, its license files beside it. Until that rename nothing an
// install before it left is touched, so a failed download, a mismatch or a
// full disk leaves the PicoClaw that was there, if any, as it was. The fetch
// directory goes in every case, and so does the add-on's directory when a
// failed first install left it empty.
func installReleasePicoclaw(progress func(stage string, percent int, lastError string)) error {
	if progress == nil {
		progress = func(string, int, string) {}
	}
	destination := picoclawInstallDestination()
	ws := picoclawFetchDir()
	if err := checkFreeSpace(ws); err != nil {
		return err
	}

	_ = os.RemoveAll(ws)
	if err := os.MkdirAll(ws, 0o755); err != nil {
		return fmt.Errorf("failed to make %s: %w", ws, err)
	}
	defer func() {
		_ = os.RemoveAll(ws)
		if addon.OnData() {
			_ = os.Remove(addon.Dir(picoclawAddon().Name))
		}
	}()

	progress("downloading", 5, "")
	archive := filepath.Join(ws, filepath.Base(release.Dir)+".tar.gz")
	if err := downloadRelease(archive, func(done, total int64) {
		if total <= 0 {
			total = release.Size
		}
		if total > 0 {
			progress("downloading", 5+int(done*65/total), "")
		}
	}); err != nil {
		if errors.Is(err, upstream.ErrChecksum) {
			return fmt.Errorf("the download is not the pinned PicoClaw %s, so nothing was installed: %w", Version, err)
		}
		return fmt.Errorf("failed to download PicoClaw %s: %w", Version, err)
	}

	progress("extracting", 75, "")
	unpacked := filepath.Join(ws, "unpacked")
	if _, err := utils.UnTarGz(archive, unpacked); err != nil {
		return fmt.Errorf("failed to unpack PicoClaw %s: %w", Version, err)
	}
	_ = os.Remove(archive)
	dir := filepath.Join(unpacked, release.Dir)
	binary := filepath.Join(dir, "picoclaw")

	progress("verifying", 85, "")
	if err := checkFile(binary, release.BinarySHA256); err != nil {
		return err
	}

	progress("installing", 90, "")
	if err := placePicoclaw(binary, dir, destination); err != nil {
		return err
	}
	if err := recordPicoclawAddon(); err != nil {
		return fmt.Errorf("failed to record the add-on: %w", err)
	}
	return nil
}

// downloadRelease writes the pinned archive to dst. It is kept only when its
// SHA-256 is the pinned one; anything else, or more than its bound, is removed
// before it takes its name.
func downloadRelease(dst string, progress func(done, total int64)) error {
	ctx, cancel := context.WithTimeout(context.Background(), picoclawDownloadTimeout)
	defer cancel()

	client, rawURL := downloads, release.URL
	if override := readReleaseURLOverride(); override != "" {
		log.Warnf("picoclaw install: downloading from %s, named in %s, instead of the release; the sha256 is still the pinned one", override, picoclawReleaseURLOverride)
		client = upstream.NewClient(downloads.APIBase, downloads.HTTP)
		client.AllowHTTP = true
		rawURL = override
	}
	return client.Download(ctx, rawURL, release.SHA256, release.Max, dst, progress)
}

func readReleaseURLOverride() string {
	raw, err := os.ReadFile(picoclawReleaseURLOverride)
	if err != nil {
		return ""
	}
	value := strings.TrimSpace(string(raw))
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		log.Warnf("picoclaw install: ignoring %s, which holds no http or https URL", picoclawReleaseURLOverride)
		return ""
	}
	return value
}

// placePicoclaw makes the verified binary the installed one, and puts the
// release's other files, its licenses, beside it. The binary goes in with a
// rename in the same directory, the sha256 record after it, so a crash in
// between leaves a binary whose record is stale, which offers the update
// again rather than hiding it.
func placePicoclaw(binary, dir, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	staged := destination + ".new"
	_ = os.Remove(staged)
	if err := os.Rename(binary, staged); err != nil {
		// Off a distribution image the fetch directory and /usr/bin can be
		// different filesystems.
		if err := installPicoclawBinary(binary, staged, release.BinarySHA256); err != nil {
			return err
		}
	}
	if err := os.Chmod(staged, 0o755); err != nil {
		_ = os.Remove(staged)
		return fmt.Errorf("failed to set the binary's mode: %w", err)
	}
	if err := os.Rename(staged, destination); err != nil {
		_ = os.Remove(staged)
		return fmt.Errorf("failed to install picoclaw binary: %w", err)
	}
	if err := writeFileAtomic(installedChecksumPath(destination), []byte(release.BinarySHA256+"\n"), 0o644); err != nil {
		return fmt.Errorf("failed to record the installed sha256: %w", err)
	}
	if err := placeLicenses(dir); err != nil {
		log.Warnf("picoclaw install: the license files were not kept: %v", err)
	}
	return nil
}

// placeLicenses replaces the license directory with the release's files other
// than the binary.
func placeLicenses(dir string) error {
	target := picoclawLicensesDir()
	staged := target + ".new"
	_ = os.RemoveAll(staged)
	if err := os.MkdirAll(staged, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Name() == "picoclaw" || !e.Type().IsRegular() {
			continue
		}
		if err := os.Rename(filepath.Join(dir, e.Name()), filepath.Join(staged, e.Name())); err != nil {
			_ = os.RemoveAll(staged)
			return err
		}
	}
	_ = os.RemoveAll(target)
	return os.Rename(staged, target)
}

// checkFile reads path and compares its SHA-256 with want.
func checkFile(path, want string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("the release holds no %s: %w", filepath.Base(path), err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	if got := hex.EncodeToString(h.Sum(nil)); got != strings.ToLower(want) {
		return fmt.Errorf("sha256 mismatch: the release's picoclaw is %s, want %s", got, want)
	}
	return nil
}

// checkFreeSpace refuses the install before the download when the filesystem
// it lands on cannot hold the archive and the unpacked binary at once. The
// fetch directory may not exist yet, so the nearest directory that does is
// asked.
func checkFreeSpace(ws string) error {
	need := uint64(release.Size + release.BinarySize + spaceMargin)
	dir := ws
	for {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil
		}
		dir = parent
	}
	free, err := freeBytes(dir)
	if err != nil {
		log.Warnf("picoclaw install: cannot read the free space of %s: %v", dir, err)
		return nil
	}
	if free < need {
		return fmt.Errorf("%w on %s: PicoClaw %s needs %d MiB, %d MiB are free", errNoSpace, dir, Version, mib(need), mib(free))
	}
	return nil
}

func mib(n uint64) uint64 { return (n + 1<<20 - 1) >> 20 }

func statfsFree(path string) (uint64, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, err
	}
	return st.Bavail * uint64(st.Bsize), nil
}

// installedPicoclawIsStale reports whether the installed PicoClaw is not the
// pinned build: Sipeed's v0.2.8 from an older server, which recorded no
// sha256, or the fork's build from an older image or release. It compares the
// sha256 recorded at install with the pin, and reads no binary. The web UI
// offers the update; nothing replaces the binary on its own.
func installedPicoclawIsStale() bool {
	have, err := os.ReadFile(installedChecksumPath(picoclawInstallDestination()))
	return err != nil || strings.TrimSpace(string(have)) != release.BinarySHA256
}

func installedChecksumPath(binary string) string {
	return binary + ".sha256"
}

// installPicoclawBinary copies source to destination through a temporary file
// and renames it into place only when the bytes written hash to wantSHA256.
func installPicoclawBinary(source string, destination string, wantSHA256 string) error {
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
// starts it through /kvmapp/system/init.d/S96picoclaw when the intent in
// /etc/kvm says so, and /etc/kvm is already on /data.
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

// picoclawFetchDir is where the release is downloaded and unpacked: inside the
// add-on's directory on a distribution image, so the binary moves into place
// with a rename on the same filesystem.
func picoclawFetchDir() string {
	if addon.OnData() {
		return filepath.Join(addon.Dir(picoclawAddon().Name), ".fetch")
	}
	return picoclawFetchFallback
}

// picoclawLicensesDir is where the release's license files are kept.
func picoclawLicensesDir() string {
	if addon.OnData() {
		return filepath.Join(addon.Dir(picoclawAddon().Name), "licenses")
	}
	return picoclawLicensesFallback
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

// removePicoclaw removes PicoClaw's settings, binary, add-on and license
// files. PicoClaw is stopped.
func removePicoclaw() {
	if configPath, err := resolvePicoclawConfigPath(); err == nil {
		_ = os.RemoveAll(filepath.Dir(configPath))
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			_ = os.RemoveAll(filepath.Join(home, ".picoclaw"))
		} else {
			_ = os.RemoveAll("/root/.picoclaw")
		}
	}
	if addon.OnData() {
		if err := addon.Remove(picoclawAddon()); err != nil {
			log.Errorf("picoclaw uninstall: failed to remove the add-on: %v", err)
		}
	}
	_ = os.Remove(picoclawBinaryPath)
	_ = os.Remove(installedChecksumPath(picoclawBinaryPath))
	if !addon.OnData() {
		_ = os.RemoveAll(picoclawLicensesFallback)
		_ = os.RemoveAll(picoclawFetchFallback)
	}
}
