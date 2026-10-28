package tailscale

import (
	"context"
	"os"
	"path/filepath"

	installer "NanoKVM-Server/service/addons/install"
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

func isInstalled() bool {
	_, err1 := os.Stat(TailscalePath)
	_, err2 := os.Stat(TailscaledPath)

	return err1 == nil && err2 == nil
}

// install downloads the pinned release, checks it and both binaries against
// their pins before anything is placed, and makes them the installed ones:
// on a distribution image as the add-on's image, mounted at once, with links
// from /usr/bin and /usr/sbin that S04addons puts back after a new image. A
// failed or mismatched download leaves nothing behind.
func install() error {
	res, err := installer.Install(context.Background(), release, installer.Options{
		Client:    downloads,
		Workspace: FallbackWorkspace,
		Plain:     placeBinaries,
	})
	if err != nil {
		log.Errorf("failed to install tailscale %s: %s", Version, err)
		return err
	}
	// placeBinaries records the links to plain files itself.
	if res.Image {
		if err := addon.Record(addonSpec()); err != nil {
			log.Errorf("failed to record the tailscale add-on: %s", err)
			return err
		}
	}
	log.Debugf("install tailscale %s successfully", Version)
	return nil
}

// placeBinaries moves the two checked binaries into place where the add-on
// cannot be an image (service/addons/install, "Fallback"). On a distribution
// image that is the add-on's directory on /data, with links from /usr/bin and
// /usr/sbin, so the next image still has them. Anywhere else it is /usr/bin
// and /usr/sbin, as it always was.
func placeBinaries(stage string) error {
	tailscaleDst, tailscaledDst := TailscalePath, TailscaledPath
	if addon.OnData() {
		tailscaleDst = filepath.Join(addon.Dir(addonSpec().Name), "tailscale")
		tailscaledDst = filepath.Join(addon.Dir(addonSpec().Name), "tailscaled")
	}

	if err := utils.MoveFile(filepath.Join(stage, "tailscale"), tailscaleDst); err != nil {
		log.Errorf("failed to move tailscale: %s", err)
		return err
	}
	if err := utils.MoveFile(filepath.Join(stage, "tailscaled"), tailscaledDst); err != nil {
		log.Errorf("failed to move tailscaled: %s", err)
		return err
	}
	if addon.OnData() {
		if err := addon.Record(spec(func(f string) string { return f })); err != nil {
			log.Errorf("failed to record the tailscale add-on: %s", err)
			return err
		}
	}
	return nil
}
