package tailscale

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/upstream"
	"NanoKVM-Server/utils"

	log "github.com/sirupsen/logrus"
)

// Variables rather than constants so the tests can point them at a scratch
// directory and a local server.
var (
	// FallbackWorkspace is where the package is fetched off a distribution
	// image. On one it is the add-on's directory on /data, where the binaries
	// end up.
	FallbackWorkspace = "/root/.tailscale"
	// downloads fetches the package over HTTPS, bounded and checked against
	// the pin.
	downloads = upstream.Default
)

const downloadTimeout = 10 * time.Minute

func isInstalled() bool {
	_, err1 := os.Stat(TailscalePath)
	_, err2 := os.Stat(TailscaledPath)

	return err1 == nil && err2 == nil
}

func workspace() string {
	if addon.OnData() {
		return filepath.Join(addon.Dir(addon.Tailscale.Name), ".fetch")
	}
	return FallbackWorkspace
}

// install downloads the pinned release, checks it against its pin before
// anything is unpacked, and moves the two binaries into place. A failed or
// mismatched download leaves nothing behind: the workspace goes, and so does
// the add-on's directory when a failed first install left it empty.
func install() error {
	ws := workspace()
	_ = os.RemoveAll(ws)
	if err := os.MkdirAll(ws, 0o755); err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(ws)
		if addon.OnData() {
			_ = os.Remove(addon.Dir(addon.Tailscale.Name))
		}
	}()

	tarFile := filepath.Join(ws, "tailscale_riscv64.tgz")
	if err := download(tarFile); err != nil {
		log.Errorf("failed to download tailscale %s: %s", Version, err)
		return err
	}

	dir, err := utils.UnTarGz(tarFile, ws)
	if err != nil {
		log.Errorf("failed to decompress tailscale: %s", err)
		return err
	}

	if err := placeBinaries(dir); err != nil {
		return err
	}

	log.Debugf("install tailscale %s successfully", Version)
	return nil
}

// placeBinaries moves the two unpacked binaries into place. On a distribution
// image that is the add-on's directory on /data, with links from /usr/bin and
// /usr/sbin, so the next image still has them. Anywhere else it is /usr/bin and
// /usr/sbin, as it always was.
func placeBinaries(dir string) error {
	tailscaleDst, tailscaledDst := TailscalePath, TailscaledPath
	onData := addon.OnData()
	if onData {
		tailscaleDst = filepath.Join(addon.Dir(addonSpec().Name), "tailscale")
		tailscaledDst = filepath.Join(addon.Dir(addonSpec().Name), "tailscaled")
	}

	if err := utils.MoveFile(fmt.Sprintf("%s/tailscale", dir), tailscaleDst); err != nil {
		log.Errorf("failed to move tailscale: %s", err)
		return err
	}
	if err := utils.MoveFile(fmt.Sprintf("%s/tailscaled", dir), tailscaledDst); err != nil {
		log.Errorf("failed to move tailscaled: %s", err)
		return err
	}

	if onData {
		if err := addon.Record(addonSpec()); err != nil {
			log.Errorf("failed to record the tailscale add-on: %s", err)
			return err
		}
	}
	return nil
}

// download writes the pinned package to dst. It is kept only when its
// SHA-256 is the pinned one; anything else, or more than its bound, is
// removed before it takes its name.
func download(dst string) error {
	ctx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
	defer cancel()
	return downloads.Download(ctx, releasePackage.URL, releasePackage.SHA256, releasePackage.Max, dst, nil)
}
