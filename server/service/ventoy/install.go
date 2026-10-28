package ventoy

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	installer "NanoKVM-Server/service/addons/install"
	"NanoKVM-Server/service/upstream"
)

// Variables rather than constants so the tests can point them at a scratch
// directory and a local server. xzcat is installer.XzcatPath.
var (
	// Dir holds the release's files, the head and the state. It is on
	// /data, and its files have no image suffix, so the image list skips
	// them.
	Dir = "/data/ironkvm/ventoy"
	// downloads fetches the archive, bounded and checked against the pin.
	downloads = upstream.Default
)

const fetchTimeout = 10 * time.Minute

func inDir(name string) string { return filepath.Join(Dir, name) }

// installed says whether every file the disk needs from the release is in
// place.
func installed() bool {
	for _, m := range releaseMembers {
		fi, err := os.Stat(inDir(m.Name))
		if err != nil || !fi.Mode().IsRegular() {
			return false
		}
	}
	return true
}

// install downloads the pinned release and keeps the three files the disk
// needs, each checked against its own sum.
func install() error {
	ws := inDir(".fetch")
	_ = os.RemoveAll(ws)
	if err := os.MkdirAll(ws, 0o755); err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(ws) }()

	archive := filepath.Join(ws, "ventoy.tar.gz")
	if err := fetchPinned(releaseArchive, archive); err != nil {
		return err
	}
	return extractMembers(archive, releaseMembers, ws, Dir)
}

// uninstall removes the release's files, the head and the record of an
// update, so a reinstall is the pinned release. The state stays, so the
// owner's selection survives a reinstall.
func uninstall() error {
	names := []string{headName, releaseName}
	for _, m := range releaseMembers {
		names = append(names, m.Name)
	}
	for _, name := range names {
		if err := os.Remove(inDir(name)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

// fetchPinned downloads f to dst, and keeps it only when its SHA-256 is the
// pinned one.
func fetchPinned(f pinnedFile, dst string) error {
	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()
	return downloads.Download(ctx, f.URL, f.SHA256, f.Max, dst, nil)
}

// extractMembers takes the named members out of the archive into ws,
// decompresses the xz ones, checks each result and moves it into dir.
// Nothing else in the archive is written, so a member's path never becomes
// a path on the board. A member the archive lacks is an error.
func extractMembers(archive string, members []releaseMember, ws, dir string) error {
	if _, err := extractChecked(archive, members, ws); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, m := range members {
		if err := os.Rename(filepath.Join(ws, m.Name), filepath.Join(dir, m.Name)); err != nil {
			return err
		}
	}
	return nil
}

// extractChecked takes the named members out of the archive into ws under
// their saved names, decompressed and checked by the shared installer, and
// returns their sums. A member with no pinned sum, from a release newer than
// the pin, is checked by the archive's sum alone.
func extractChecked(archive string, members []releaseMember, ws string) (map[string]string, error) {
	ms := make([]installer.Member, len(members))
	for i, m := range members {
		ms[i] = installer.Member{Path: m.Member, Name: m.Name, SHA256: m.SHA256, XZ: m.XZ, Max: memberMax}
	}
	return installer.ExtractTarGz(archive, ms, ws)
}
