package ventoy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"NanoKVM-Server/service/upstream"
)

// Ventoy can be updated past the pinned release, from the project's GitHub
// releases. Each release publishes sha256.txt with the sum of its Linux
// archive; the archive is checked against it, and the three files the disk
// needs are then taken out of it as the pinned release's are. The layout the
// disk is built with is Ventoy2Disk.sh 1.1.17's, so a release whose files do
// not have the sizes that layout expects is refused before anything is
// replaced.

const (
	ventoyRepo      = "ventoy/Ventoy"
	ventoyChecksums = "sha256.txt"
	releaseName     = "release.json"
)

func archiveName(version string) string { return "ventoy-" + version + "-linux.tar.gz" }

// membersOf names the files a release's archive holds the disk's files in,
// with no pinned sums.
func membersOf(version string) []releaseMember {
	prefix := "./ventoy-" + version
	return []releaseMember{
		{Member: prefix + "/boot/boot.img", Name: bootName},
		{Member: prefix + "/boot/core.img.xz", Name: coreName, XZ: true},
		{Member: prefix + "/ventoy/ventoy.disk.img.xz", Name: efiName, XZ: true},
	}
}

// releaseRecord is what an update leaves in Dir: the release the files came
// from, the archive's published sum and the files' sums. Without it the
// files are the pinned release's. An uninstall removes it, so the next
// install is the pinned release again.
type releaseRecord struct {
	Version string            `json:"version"`
	Archive string            `json:"archive"`
	Files   map[string]string `json:"files"`
}

func readRelease() (releaseRecord, bool) {
	var rec releaseRecord
	data, err := os.ReadFile(inDir(releaseName))
	if err != nil {
		return rec, false
	}
	if json.Unmarshal(data, &rec) != nil || rec.Version == "" {
		return releaseRecord{}, false
	}
	return rec, true
}

// installedVersion is the Ventoy release in Dir.
func installedVersion() string {
	if rec, ok := readRelease(); ok {
		return rec.Version
	}
	return Version
}

// checkLayout refuses files the disk's layout cannot hold: boot code shorter
// than the MBR's, a core image past partition 1's start, or a VTOYEFI
// partition of another size than the one the device table maps.
func checkLayout(dir string) error {
	size := func(name string) int64 {
		fi, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			return -1
		}
		return fi.Size()
	}
	if n := size(bootName); n < bootCodeBytes {
		return fmt.Errorf("boot.img is %d bytes, shorter than the MBR's code", n)
	}
	if n := size(coreName); n <= 0 || n > coreSectors*sectorSize {
		return fmt.Errorf("core.img is %d bytes, and does not fit before partition 1", n)
	}
	if n := size(efiName); n != efiSectors*sectorSize {
		return fmt.Errorf("the VTOYEFI image is %d bytes, not the %d this disk layout maps", n, efiSectors*sectorSize)
	}
	return nil
}

// Updater returns the updater of the Ventoy release.
func (s *Service) Updater(client *upstream.Client) *upstream.Updater {
	var u *upstream.Updater
	u = upstream.NewUpdater(upstream.Component{
		Name:         "Ventoy",
		Repo:         ventoyRepo,
		Pinned:       Version,
		ChecksumFile: func(string) string { return ventoyChecksums },
		Assets:       func(v string) []string { return []string{archiveName(v)} },
		Installed: func() (string, bool) {
			if !installed() {
				return "", false
			}
			return installedVersion(), true
		},
		InUse: func() error {
			if in, err := inDrive(); err != nil {
				return err
			} else if in {
				return errInDrive
			}
			return nil
		},
		Install: func(ctx context.Context, rel upstream.Release, sums upstream.Checksums, progress func(int)) error {
			return s.update(rel.Version, sums[archiveName(rel.Version)], func(dst string) error {
				return u.Download(ctx, rel, sums, archiveName(rel.Version), releaseArchive.Max, dst, progress, 0, 80)
			}, progress)
		},
	}, client)
	return u
}

// update replaces the release's files with those of version, whose archive
// fetch downloads and checks. The disk must be in no drive, as the device
// reads these files; an idle device is removed first, and the next insert
// builds it from the new files.
func (s *Service) update(version, archiveSum string, fetch func(dst string) error, progress func(int)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.deps.OnData() {
		return errNoData
	}
	if !installed() {
		return errNotReady
	}
	if in, err := inDrive(); err != nil {
		return err
	} else if in {
		return errInDrive
	}

	ws := inDir(".update")
	_ = os.RemoveAll(ws)
	if err := os.MkdirAll(ws, 0o755); err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(ws) }()

	archive := filepath.Join(ws, archiveName(version))
	if err := fetch(archive); err != nil {
		return err
	}
	staged := filepath.Join(ws, "files")
	if err := os.MkdirAll(staged, 0o755); err != nil {
		return err
	}
	members := membersOf(version)
	sums, err := extractChecked(archive, members, staged)
	if err != nil {
		return err
	}
	_ = os.Remove(archive)
	progress(90)
	if err := checkLayout(staged); err != nil {
		return err
	}

	rec := releaseRecord{Version: version, Archive: archiveSum, Files: sums}
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	names := make([]string, len(members))
	for i, m := range members {
		names[i] = m.Name
	}

	// The drive was checked under mu, which an insert also takes, so no
	// device reads the files while they are swapped.
	if err := s.removeIdle(); err != nil {
		return err
	}
	return upstream.Swap(staged, Dir, names, func() error {
		return writeAtomic(inDir(releaseName), data)
	})
}
