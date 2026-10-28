package netboot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	installer "NanoKVM-Server/service/addons/install"
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
	"NanoKVM-Server/service/upstream"
	"NanoKVM-Server/utils"
	"NanoKVM-Server/utils/atomicfile"
)

// AddonName is the add-on that carries dnsmasq and the boot files, in
// /data/ironkvm/addons/netboot.
const AddonName = "netboot"

// Initd is the add-on's boot script. S04addons copies it to /etc/init.d while
// proxy DHCP on the LAN is on; S03usbdev runs it from /kvmapp for the link.
const Initd = "S85netboot"

// Variables rather than constants so the tests can point them at a scratch
// root and a stub package manager.
var (
	ApkPath = "apk"
	// downloads fetches the pinned files, bounded and checked against their
	// pins (service/addons/install shares it).
	downloads = upstream.Default
)

const (
	dnsmasqPackage = "dnsmasq"
	apkTimeout     = 10 * time.Minute
	fetchTimeout   = 10 * time.Minute
)

// addonDir is the add-on's directory on /data; dnsmasq is in it, and the
// TFTP root under it.
func addonDir() string { return addon.Dir(AddonName) }

func dnsmasqPath() string { return filepath.Join(addonDir(), "dnsmasq") }

func tftpRoot() string { return filepath.Join(addonDir(), "tftp") }

// StagedTFTPRoot is the directory dnsmasq serves by TFTP. /data is exFAT,
// mounted with fmask and dmask 0077, so everything on it is root's alone.
// dnsmasq runs as nobody and checks its TFTP directory at startup, so pointed
// at the add-on's own copy it refuses to start at all. S85netboot copies the
// boot files here, on the slot's ext4, readable by everyone, before it starts
// an instance. The server's HTTP menu runs as root and keeps reading tftpRoot.
const StagedTFTPRoot = "/var/lib/netboot/tftp"

func versionPath() string { return filepath.Join(addonDir(), "version") }

// addonSpec is what the add-on needs from the root filesystem: only its boot
// script, while the LAN side is on. dnsmasq runs from /data and gets no link
// in /usr/sbin, so the image's S80dnsmasq, which starts a dnsmasq it finds
// there, stays idle.
func addonSpec() addon.Spec {
	return addon.Spec{Name: AddonName, Initd: Initd}
}

// dnsmasqInstalled says whether the binary is in place.
func dnsmasqInstalled() bool {
	fi, err := os.Stat(dnsmasqPath())
	return err == nil && fi.Mode().IsRegular()
}

// bootFilesInstalled says whether every boot file is in the TFTP root.
func bootFilesInstalled() bool {
	for _, name := range bootFiles() {
		fi, err := os.Stat(filepath.Join(tftpRoot(), name))
		if err != nil || !fi.Mode().IsRegular() {
			return false
		}
	}
	return true
}

// installed says whether network boot can be turned on.
func installed() bool {
	return dnsmasqInstalled() && bootFilesInstalled()
}

// install puts dnsmasq and the boot files on /data. A part already in place
// is kept, so an install that failed half way costs only the rest next time.
func install() error {
	if !addon.OnData() {
		return errors.New("network boot keeps dnsmasq and its boot files on /data, and needs an IronKVM image with /data mounted")
	}

	ws := filepath.Join(addonDir(), ".fetch")
	_ = os.RemoveAll(ws)
	if err := os.MkdirAll(ws, 0o755); err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(ws) }()

	if !dnsmasqInstalled() {
		if err := installDnsmasq(ws); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(tftpRoot(), 0o755); err != nil {
		return err
	}
	if err := installBootFiles(ws); err != nil {
		return err
	}

	return addon.Record(addonSpec())
}

// installDnsmasq fetches Alpine's dnsmasq package, checks its signature, and
// takes usr/sbin/dnsmasq out of it, as the NetBird add-on takes its binary.
// Only that package: it needs musl, which the image has, and dnsmasq-common
// holds the sample configuration and the package's scripts, which nothing
// here uses.
func installDnsmasq(ws string) error {
	if _, err := exec.LookPath(ApkPath); err != nil {
		return errors.New("apk not found: dnsmasq installs from Alpine's packages and needs an IronKVM image")
	}

	if _, err := apk("update"); err != nil {
		return err
	}
	if _, err := apk("fetch", "-o", ws, dnsmasqPackage); err != nil {
		return err
	}
	pkg, err := fetchedPackage(ws)
	if err != nil {
		return err
	}

	// The shared installer has apk check the signature again while it
	// unpacks the package, and takes only usr/sbin/dnsmasq out of it.
	stage := filepath.Join(ws, "stage")
	if _, err := installer.ExtractApk(pkg, []installer.Member{
		{Path: "usr/sbin/dnsmasq", Name: "dnsmasq", Mode: 0o755},
	}, stage, ApkPath); err != nil {
		return err
	}
	if err := utils.MoveFile(filepath.Join(stage, "dnsmasq"), dnsmasqPath()); err != nil {
		return err
	}

	version := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(pkg), dnsmasqPackage+"-"), ".apk")
	_ = os.WriteFile(versionPath(), []byte(version+"\n"), 0o644)

	log.Debugf("install dnsmasq from %s", filepath.Base(pkg))
	return nil
}

// fetchedPackage is the one dnsmasq-<version>.apk that apk fetch wrote.
// dnsmasq-common and dnsmasq-dnssec are other packages, and their names do
// not have a digit after the dash.
func fetchedPackage(dir string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, dnsmasqPackage+"-[0-9]*.apk"))
	if err != nil {
		return "", err
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("apk fetch left %d dnsmasq packages in %s", len(matches), dir)
	}
	return matches[0], nil
}

func apk(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apkTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, ApkPath, args...)
	cmd.WaitDelay = 2 * time.Second
	return vpn.Run(cmd)
}

// installBootFiles downloads what is missing from the TFTP root and writes
// boot.ipxe. Every file is checked against its pin before it takes its name.
func installBootFiles(ws string) error {
	root := tftpRoot()

	missing := false
	for _, m := range ipxeMembers {
		if !exists(filepath.Join(root, m.Name)) {
			missing = true
		}
	}
	if missing {
		archive := filepath.Join(ws, ipxeArchive.Name)
		if err := fetchPinned(ipxeArchive.URL, ipxeArchive.SHA256, ipxeArchiveMax, archive); err != nil {
			return fmt.Errorf("iPXE %s: %w", ipxeVersion, err)
		}
		if err := extractMembers(archive, ipxeMembers, root); err != nil {
			return fmt.Errorf("iPXE %s: %w", ipxeVersion, err)
		}
	}

	for _, f := range netbootXYZFiles {
		dst := filepath.Join(root, f.Name)
		if exists(dst) {
			continue
		}
		if err := fetchPinned(f.URL, f.SHA256, bootFileMax, dst); err != nil {
			return fmt.Errorf("netboot.xyz %s: %w", netbootXYZVersion, err)
		}
	}

	return writeFile(filepath.Join(root, bootScriptName), bootScript())
}

func exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

// fetchPinned downloads url to dst, and keeps it only when its SHA-256 is
// sum. It reads at most limit bytes: a pinned file has a known size, and a
// server that sends more is not sending it.
func fetchPinned(url, sum string, limit int64, dst string) error {
	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()
	return downloads.Download(ctx, url, sum, limit, dst, nil)
}

// extractMembers takes the named members out of a .tar.gz into dir, each
// checked against its own sum, through the shared installer. Nothing else in
// the archive is written, so a member's path never becomes a path on the
// board.
func extractMembers(archive string, members []archiveMember, dir string) error {
	ms := make([]installer.Member, len(members))
	for i, m := range members {
		ms[i] = installer.Member{Path: m.Member, Name: m.Name, SHA256: m.SHA256, Max: bootFileMax}
	}
	_, err := installer.ExtractTarGz(archive, ms, dir)
	return err
}

// writeFile replaces a file whole.
func writeFile(name, content string) error {
	_, err := replaceFile(name, content)
	return err
}

// replaceFile replaces a file whole, and says whether its content changed. A
// torn write would leave dnsmasq a file it refuses, so the new content goes to
// a temporary file first.
func replaceFile(name, content string) (bool, error) {
	if current, err := os.ReadFile(name); err == nil && string(current) == content {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		return false, err
	}
	return true, atomicfile.WriteFile(name, []byte(content), 0o644)
}

// removeFile removes a file and says whether there was one.
func removeFile(name string) (bool, error) {
	err := os.Remove(name)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}

// uninstall removes the add-on: dnsmasq, the boot files and the records.
func uninstall() error {
	if !addon.OnData() {
		return nil
	}
	return addon.Remove(addonSpec())
}
