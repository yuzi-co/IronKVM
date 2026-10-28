//go:build linux

package netbird

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
	"NanoKVM-Server/service/upstream"
)

// pinnedPackage serves a good package, pinned to its own sum.
func pinnedPackage(t *testing.T) {
	t.Helper()
	pkg := fakePackage(t, "netbird 0.79.0\n")
	servePackage(t, pkg, sha256Hex(pkg), 1<<20)
}

func TestInstallOnADistributionImagePutsTheBinaryOnData(t *testing.T) {
	fsroot := scratchImage(t, true)
	calls := fakeApk(t)
	pinnedPackage(t)
	if err := install(); err != nil {
		t.Fatal(err)
	}

	dir := addon.Dir("netbird")
	bin := filepath.Join(dir, "netbird")
	if b, err := os.ReadFile(bin); err != nil || string(b) != "netbird 0.79.0\n" {
		t.Fatalf("the binary is not on /data: %q %v", b, err)
	}
	if fi, err := os.Stat(bin); err != nil || fi.Mode().Perm()&0o111 == 0 {
		t.Fatal("the binary must be executable")
	}
	if target, err := os.Readlink(fsroot + "/usr/bin/netbird"); err != nil || target != bin {
		t.Fatalf("/usr/bin/netbird is not a link into the add-on: %q %v", target, err)
	}
	if b, err := os.ReadFile(filepath.Join(dir, "initd")); err != nil || string(b) != "S98netbird\n" {
		t.Fatalf("initd is %q (%v)", b, err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".fetch")); err == nil {
		t.Fatal("the fetched package must not be left on /data")
	}
	if !isInstalled() {
		t.Fatal("isInstalled must follow the link and answer true")
	}

	got := callsOf(t, calls)
	last := -1
	for _, want := range []string{"verify ", "extract "} {
		i := strings.Index(got, want)
		if i <= last {
			t.Fatalf("want %q after the previous step in:\n%s", want, got)
		}
		last = i
	}
	if !strings.Contains(got, packageFile) {
		t.Fatalf("apk must read the pinned package:\n%s", got)
	}
}

func TestInstallElsewhereUsesUsrBin(t *testing.T) {
	fsroot := scratchImage(t, false)
	fakeApk(t)
	pinnedPackage(t)
	if err := install(); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Lstat(fsroot + "/usr/bin/netbird")
	if err != nil || !fi.Mode().IsRegular() {
		t.Fatalf("/usr/bin/netbird must be the binary itself off a distribution image: %v", err)
	}
	if _, err := os.Stat(addon.Dir("netbird")); err == nil {
		t.Fatal("nothing may be written to /data off a distribution image")
	}
	if _, err := os.Stat(FallbackWorkspace); err == nil {
		t.Fatal("the workspace must be removed")
	}
}

func TestInstallRefusesAnUnsignedPackage(t *testing.T) {
	fsroot := scratchImage(t, true)
	fakeApk(t)
	pinnedPackage(t)
	t.Setenv("BAD_SIGNATURE", "1")

	err := install()
	if err == nil || !strings.Contains(vpn.Message("install failed", err), "UNTRUSTED signature") {
		t.Fatalf("got %v", err)
	}
	if _, err := os.Lstat(fsroot + "/usr/bin/netbird"); err == nil {
		t.Fatal("an unsigned package must not be installed")
	}
	if _, err := os.Stat(addon.Dir("netbird")); err == nil {
		t.Fatal("a failed first install must leave no add-on directory behind")
	}
}

// The vendor firmware has no apk. NetBird needs an IronKVM image.
func TestInstallWithoutApkSaysWhy(t *testing.T) {
	scratchImage(t, true)
	saved := ApkPath
	t.Cleanup(func() { ApkPath = saved })
	ApkPath = filepath.Join(t.TempDir(), "no-apk")

	err := install()
	if err == nil || !strings.Contains(err.Error(), "needs an IronKVM image") {
		t.Fatalf("got %v", err)
	}
	if _, err := os.Stat(addon.Dir("netbird")); err == nil {
		t.Fatal("nothing may be left behind")
	}
}

// A package whose sum is not the pin is refused before apk reads it, and
// leaves nothing behind.
func TestInstallRefusesAMismatchedPackage(t *testing.T) {
	fsroot := scratchImage(t, true)
	calls := fakeApk(t)
	pkg := fakePackage(t, "netbird 0.79.0\n")
	servePackage(t, pkg, sha256Hex([]byte("another package")), 1<<20)

	err := install()
	if !errors.Is(err, upstream.ErrChecksum) {
		t.Fatalf("got %v", err)
	}
	assertNothingLeft(t, fsroot, calls)
}

// A package larger than its bound is refused, whatever its sum.
func TestInstallRefusesAnOversizePackage(t *testing.T) {
	fsroot := scratchImage(t, true)
	calls := fakeApk(t)
	pkg := fakePackage(t, strings.Repeat("x", 4096))
	servePackage(t, pkg, sha256Hex(pkg), int64(len(pkg)-1))

	err := install()
	if err == nil || !strings.Contains(err.Error(), "larger than") {
		t.Fatalf("got %v", err)
	}
	assertNothingLeft(t, fsroot, calls)
}

func assertNothingLeft(t *testing.T, fsroot, calls string) {
	t.Helper()
	if got := callsOf(t, calls); got != "" {
		t.Fatalf("apk must not see a package that failed its pin:\n%s", got)
	}
	if _, err := os.Lstat(fsroot + "/usr/bin/netbird"); err == nil {
		t.Fatal("nothing may be installed")
	}
	if _, err := os.Stat(addon.Dir("netbird")); err == nil {
		t.Fatal("a failed first install must leave no add-on directory behind")
	}
}

// The pin is what the license manifest and the update offer name.
func TestThePinIsOneRelease(t *testing.T) {
	if !strings.Contains(release.Artifact.URL, "/netbird-"+Version+"-r") || !strings.HasSuffix(release.Artifact.URL, "/"+packageFile) {
		t.Fatalf("%s is not netbird %s as %s", release.Artifact.URL, Version, packageFile)
	}
	if !strings.HasPrefix(release.Artifact.URL, "https://") || len(release.Artifact.SHA256) != 64 {
		t.Fatalf("not a pin: %+v", release.Artifact)
	}
}
