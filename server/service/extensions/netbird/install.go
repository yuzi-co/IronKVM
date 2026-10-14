package netbird

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
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

const (
	apkTimeout      = 10 * time.Minute
	downloadTimeout = 10 * time.Minute
)

// addonSpec is what NetBird needs back from the root filesystem after a new
// image: its binary in /usr/bin and its boot script while it is enabled. Its
// identity is already on /data, where S98netbird keeps it.
func addonSpec() addon.Spec {
	return addon.Spec{
		Name:  addon.NetBird.Name,
		Links: []addon.Link{{Path: "/usr/bin/netbird", File: "netbird"}},
		Initd: addon.NetBird.Initd,
	}
}

func isInstalled() bool {
	_, err := os.Stat(NetbirdPath)
	return err == nil
}

func workspace() string {
	if addon.OnData() {
		return filepath.Join(addon.Dir(addon.NetBird.Name), ".fetch")
	}
	return FallbackWorkspace
}

// apk runs the package manager with apk's own deadline. A failure carries its
// output.
func apk(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apkTimeout)
	defer cancel()
	return apkContext(ctx, args...)
}

// apkContext runs the package manager until ctx ends. WaitDelay closes the
// pipes if apk leaves a child holding them, so the deadline holds.
func apkContext(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, ApkPath, args...)
	cmd.WaitDelay = 2 * time.Second
	return vpn.Run(cmd)
}

// install downloads the pinned package, checks it against its pin before apk
// reads it, has apk check its signature, and takes usr/bin/netbird out of it.
// Nothing is installed into the root filesystem: the binary goes to the
// add-on's directory on /data, linked from /usr/bin, or to /usr/bin itself
// off a distribution image. A failed or mismatched download leaves nothing
// behind.
func install() error {
	if _, err := exec.LookPath(ApkPath); err != nil {
		return errors.New("apk not found: NetBird installs from Alpine's packages and needs an IronKVM image")
	}

	ws := workspace()
	_ = os.RemoveAll(ws)
	if err := os.MkdirAll(ws, 0o755); err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(ws)
		if addon.OnData() {
			// Only removes the add-on's directory when a failed first
			// install left it empty.
			_ = os.Remove(addon.Dir(addon.NetBird.Name))
		}
	}()

	pkg := filepath.Join(ws, packageFile)
	if err := download(pkg); err != nil {
		return err
	}
	if _, err := apk("verify", pkg); err != nil {
		return err
	}
	root := filepath.Join(ws, "root")
	if err := extractPackage(pkg, root); err != nil {
		return err
	}
	if err := placeBinary(filepath.Join(root, "usr", "bin", "netbird")); err != nil {
		return err
	}
	log.Debugf("install netbird from %s", packageFile)
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

// extractPackage unpacks the package into dir. apk extract checks the
// signature again and reads both package formats. --no-chown because /data is
// exFAT, which has no owners.
func extractPackage(pkg, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	_, err := apk("extract", "--no-chown", "--destination", dir, pkg)
	return err
}

// placeBinary moves the extracted binary into place and, on a distribution
// image, records the add-on so S04addons puts the link back after a new image.
func placeBinary(src string) error {
	fi, err := os.Stat(src)
	if err != nil || !fi.Mode().IsRegular() {
		return errors.New("the netbird package has no usr/bin/netbird")
	}
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
		return addon.Record(addonSpec())
	}
	return nil
}
