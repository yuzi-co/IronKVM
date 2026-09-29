package netboot

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"NanoKVM-Server/service/upstream"
)

// The netboot.xyz boot files can be updated past the pinned release, from the
// project's GitHub releases. Each release publishes
// netboot.xyz-sha256-checksums.txt, and every file is checked against it.
// iPXE is not updated: its releases publish no checksum file, and the pinned
// sums of its archive's members are the only check this server has for it.
// dnsmasq comes from Alpine's signed packages and is not updated here either.

// The netboot.xyz releases, exported for the boot menu ISO's updater in the
// download service.
const (
	NetbootXYZRepo      = "netbootxyz/netboot.xyz"
	NetbootXYZChecksums = "netboot.xyz-sha256-checksums.txt"
	// BootMenuISOName is the ISO's asset name, the same in every release,
	// and the name it is saved under.
	BootMenuISOName = "netboot.xyz.iso"
	// BootMenuISOVersion is the pinned ISO's release.
	BootMenuISOVersion = netbootXYZVersion
)

// bootRecord is what an update leaves in the add-on's directory: the
// release the boot files came from and their sums. Without it the files are
// the pinned release's. It goes with the add-on on an uninstall.
type bootRecord struct {
	Version string            `json:"version"`
	Files   map[string]string `json:"files"`
}

const bootRecordName = "netboot.xyz.json"

func bootRecordPath() string { return filepath.Join(addonDir(), bootRecordName) }

func readBootRecord() (bootRecord, bool) {
	var rec bootRecord
	data, err := os.ReadFile(bootRecordPath())
	if err != nil {
		return rec, false
	}
	if json.Unmarshal(data, &rec) != nil || rec.Version == "" {
		return bootRecord{}, false
	}
	return rec, true
}

// bootFilesVersion is the netboot.xyz release in the TFTP root.
func bootFilesVersion() string {
	if rec, ok := readBootRecord(); ok {
		return rec.Version
	}
	return netbootXYZVersion
}

func netbootXYZNames() []string {
	names := make([]string, len(netbootXYZFiles))
	for i, f := range netbootXYZFiles {
		names[i] = f.Name
	}
	return names
}

// Updater returns the updater of the netboot.xyz boot files.
func (s *Service) Updater(client *upstream.Client) *upstream.Updater {
	var u *upstream.Updater
	u = upstream.NewUpdater(upstream.Component{
		Name:         "netboot.xyz boot files",
		Repo:         NetbootXYZRepo,
		Pinned:       netbootXYZVersion,
		ChecksumFile: func(string) string { return NetbootXYZChecksums },
		Assets:       func(string) []string { return netbootXYZNames() },
		Installed: func() (string, bool) {
			if !bootFilesInstalled() {
				return "", false
			}
			return bootFilesVersion(), true
		},
		InUse: func() error { return nil },
		Install: func(ctx context.Context, rel upstream.Release, sums upstream.Checksums, progress func(int)) error {
			return s.updateBootFiles(ctx, rel, sums, func(ctx context.Context, name, dst string, from, span int) error {
				return u.Download(ctx, rel, sums, name, bootFileMax, dst, progress, from, span)
			})
		},
	}, client)
	return u
}

// updateBootFiles downloads the release's boot files next to the TFTP root,
// and swaps them in once all are verified. A host booting meanwhile gets
// either release, whole: each file is renamed into place. The LAN's dnsmasq
// serves a copy of the TFTP root, so it restarts to copy the new files.
func (s *Service) updateBootFiles(ctx context.Context, rel upstream.Release, sums upstream.Checksums,
	fetch func(ctx context.Context, name, dst string, from, span int) error) error {
	// An update waits for a settings change or the reconcile loop; an
	// install or uninstall that holds busy has finished before it starts.
	s.busy.Lock()
	defer s.busy.Unlock()

	if !bootFilesInstalled() {
		return errNotInstalled
	}

	ws := filepath.Join(addonDir(), ".update")
	_ = os.RemoveAll(ws)
	if err := os.MkdirAll(ws, 0o755); err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(ws) }()

	names := netbootXYZNames()
	for i, name := range names {
		span := 90 / len(names)
		if err := fetch(ctx, name, filepath.Join(ws, name), i*span, span); err != nil {
			return err
		}
	}

	rec := bootRecord{Version: rel.Version, Files: map[string]string{}}
	for _, name := range names {
		rec.Files[name] = sums[name]
	}
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	commit := func() error { _, err := replaceFile(bootRecordPath(), string(data)+"\n"); return err }
	if err := upstream.Swap(ws, tftpRoot(), names, commit); err != nil {
		return err
	}

	settings := s.deps.Settings()
	if settings.LAN && running("lan") {
		s.restartLAN(true)
	}
	return nil
}
