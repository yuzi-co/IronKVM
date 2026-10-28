package netbird

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	installer "NanoKVM-Server/service/addons/install"
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/upstream"
	"NanoKVM-Server/utils"

	log "github.com/sirupsen/logrus"
)

// Variables rather than constants so the tests can point them at a scratch
// root and a stub package manager. On a distribution image NetbirdPath is a
// link into the add-on on /data.
var (
	NetbirdPath = "/usr/bin/netbird"
	ApkPath     = "apk"
	// FallbackWorkspace is where the package is fetched off a distribution
	// image. On one it is the add-on's directory on /data: /tmp is a 79 MB
	// tmpfs and the binary alone is 43 MB.
	FallbackWorkspace = "/root/.netbird-fetch"
	// downloads fetches the package over HTTPS, bounded and checked against
	// the pin.
	downloads = upstream.Default
)

// addonSpec is what NetBird needs back from the root filesystem after a new
// image: its binary in /usr/bin and its boot script while it is enabled. Its
// identity is already on /data, where S98netbird keeps it. The link names
// the binary in the add-on's image while it is one.
func addonSpec() addon.Spec {
	return spec(installer.LinkFile(addon.NetBird.Name, "netbird"))
}

// spec is addonSpec with the link naming file. A plain install records the
// plain name before the image it replaces, if any, goes.
func spec(file string) addon.Spec {
	return addon.Spec{
		Name:  addon.NetBird.Name,
		Links: []addon.Link{{Path: "/usr/bin/netbird", File: file}},
		Initd: addon.NetBird.Initd,
	}
}

func isInstalled() bool {
	_, err := os.Stat(NetbirdPath)
	return err == nil
}

// install downloads the pinned package, checks it against its pin before apk
// reads it, has apk check its signature, and takes usr/bin/netbird out of it.
// Nothing is installed into the root filesystem: the binary goes to the
// add-on's image on /data, linked from /usr/bin, or to /usr/bin itself off a
// distribution image. A failed or mismatched download leaves nothing behind.
func install() error {
	if _, err := exec.LookPath(ApkPath); err != nil {
		return errors.New("apk not found: NetBird installs from Alpine's packages and needs an IronKVM image")
	}
	res, err := installer.Install(context.Background(), release, installer.Options{
		Client:    downloads,
		Apk:       ApkPath,
		Workspace: FallbackWorkspace,
		Plain:     placeBinary,
	})
	if err != nil {
		return err
	}
	// placeBinary records the link to a plain file itself.
	if res.Image {
		if err := addon.Record(addonSpec()); err != nil {
			return err
		}
	}
	log.Debugf("install netbird from %s", packageFile)
	return nil
}

// placeBinary moves the checked binary into place where the add-on cannot be
// an image (service/addons/install, "Fallback"): the add-on's directory on a
// distribution image, recorded so S04addons puts the link back after a new
// image, and /usr/bin anywhere else, as it always was.
func placeBinary(stage string) error {
	src := filepath.Join(stage, "netbird")
	if err := os.Chmod(src, 0o755); err != nil {
		return err
	}
	dst := NetbirdPath
	onData := addon.OnData()
	if onData {
		dst = filepath.Join(addon.Dir(addon.NetBird.Name), "netbird")
	}
	if err := utils.MoveFile(src, dst); err != nil {
		return err
	}
	if onData {
		return addon.Record(spec("netbird"))
	}
	return nil
}
