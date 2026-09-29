package download

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/service/netboot"
	"NanoKVM-Server/service/storage"
	"NanoKVM-Server/service/upstream"
)

// The netboot.xyz ISO behind the image downloader's boot menu button. The
// server pins a release of it; an update from netboot.xyz's GitHub releases,
// checked against the release's netboot.xyz-sha256-checksums.txt, replaces
// it with a newer one and records which, so the button then answers from the
// recorded release.

// bootMenuRecordPath is where the record of an updated ISO is kept, next to
// the other add-ons' state on /data. A variable for the tests.
var bootMenuRecordPath = "/data/ironkvm/bootmenu.json"

// bootMenu is a release of the ISO: its version, URL and SHA-256.
type bootMenu struct {
	Version string `json:"version"`
	URL     string `json:"url"`
	SHA256  string `json:"sha256"`
}

func pinnedBootMenu() bootMenu {
	return bootMenu{
		Version: netboot.BootMenuISOVersion,
		URL:     netboot.BootMenuISOURL,
		SHA256:  netboot.BootMenuISOSHA256,
	}
}

func bootMenuPath() string { return filepath.Join(imageDir, netboot.BootMenuISOName) }

// readBootMenuRecord returns the recorded release, if it is whole and names
// the ISO over HTTPS.
func readBootMenuRecord() (bootMenu, bool) {
	var rec bootMenu
	data, err := os.ReadFile(bootMenuRecordPath)
	if err != nil || json.Unmarshal(data, &rec) != nil {
		return bootMenu{}, false
	}
	u, err := url.Parse(rec.URL)
	if err != nil || u.Scheme != "https" || path.Base(u.Path) != netboot.BootMenuISOName || rec.Version == "" {
		return bootMenu{}, false
	}
	if sum, err := parseSHA256(rec.SHA256); err != nil || sum == nil {
		return bootMenu{}, false
	}
	return rec, true
}

// currentBootMenu is the release the button downloads: the recorded one, or
// the pinned one.
func currentBootMenu() bootMenu {
	if rec, ok := readBootMenuRecord(); ok {
		return rec
	}
	return pinnedBootMenu()
}

// installedBootMenu returns the version of the ISO in the image directory, if
// its checksum is the recorded release's or the pinned one's. Any other file
// under the name is not one this server can vouch for.
func installedBootMenu() (string, bool) {
	file := bootMenuPath()
	candidates := []bootMenu{pinnedBootMenu()}
	if rec, ok := readBootMenuRecord(); ok {
		candidates = append([]bootMenu{rec}, candidates...)
	}
	for _, c := range candidates {
		sum, err := parseSHA256(c.SHA256)
		if err == nil && fileHasSHA256(file, sum) {
			return c.Version, true
		}
	}
	return "", false
}

// BootMenuUpdater returns the updater of the netboot.xyz ISO.
func (s *Service) BootMenuUpdater(client *upstream.Client) *upstream.Updater {
	return upstream.NewUpdater(upstream.Component{
		Name:         "netboot.xyz ISO",
		Repo:         netboot.NetbootXYZRepo,
		Pinned:       netboot.BootMenuISOVersion,
		ChecksumFile: func(string) string { return netboot.NetbootXYZChecksums },
		Assets:       func(string) []string { return []string{netboot.BootMenuISOName} },
		Installed:    installedBootMenu,
		InUse:        func() error { return storage.ImageInUse(bootMenuPath()) },
		Install: func(ctx context.Context, rel upstream.Release, sums upstream.Checksums, progress func(int)) error {
			next := bootMenu{
				Version: rel.Version,
				URL:     rel.Assets[netboot.BootMenuISOName],
				SHA256:  sums[netboot.BootMenuISOName],
			}
			return s.updateBootMenu(ctx, next, progress)
		},
	}, client)
}

// updateBootMenu downloads the release through the image downloader, so it
// takes the transfer lock and shows in the downloader's progress, and
// replaces the ISO only once it is verified and while no drive serves it.
func (s *Service) updateBootMenu(ctx context.Context, next bootMenu, progress func(int)) error {
	sum, err := parseSHA256(next.SHA256)
	if err != nil || sum == nil {
		return errors.New("the release publishes no valid SHA-256 for the ISO")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	done, err := s.beginDownload(next.URL, cancel)
	if err != nil {
		return err
	}

	err = s.downloadRemoteImage(ctx, next.URL, sum, netboot.BootMenuISOName, func(percentage string) {
		s.setDownloadProgress(done, percentage)
		if p, err := strconv.ParseFloat(strings.TrimSuffix(percentage, "%"), 64); err == nil {
			progress(int(p))
		}
	}, func(tmp, dest string) error {
		if err := storage.ReplaceImage(tmp, dest); err != nil {
			return err
		}
		return writeBootMenuRecord(next)
	})

	status := downloadStatusSuccess
	switch {
	case errors.Is(err, context.Canceled):
		status = downloadStatusIdle
	case errors.Is(err, errSHA256Mismatch):
		status = downloadStatusChecksumFailed
	case err != nil:
		status = downloadStatusFailed
	}
	s.finishDownload(done, status)
	if errors.Is(err, errSHA256Mismatch) {
		return errors.New("the downloaded ISO does not match the release's checksum file")
	}
	return err
}

func writeBootMenuRecord(rec bootMenu) error {
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(bootMenuRecordPath), 0o755); err != nil {
		return err
	}
	tmp := bootMenuRecordPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, bootMenuRecordPath); err != nil {
		log.Errorf("record the netboot.xyz ISO's release: %s", err)
		return err
	}
	return nil
}
