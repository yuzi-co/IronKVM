//go:build linux

package netbird

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
)

func TestInstallOnADistributionImagePutsTheBinaryOnData(t *testing.T) {
	fsroot := scratchImage(t, true)
	calls := fakeApk(t, "0.79.0")
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
	for _, want := range []string{"update\n", "fetch -o ", "verify "} {
		i := strings.Index(got, want)
		if i <= last {
			t.Fatalf("want %q after the previous step in:\n%s", want, got)
		}
		last = i
	}
}

func TestInstallElsewhereUsesUsrBin(t *testing.T) {
	fsroot := scratchImage(t, false)
	fakeApk(t, "0.79.0")
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
	fakeApk(t, "0.79.0")
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

func TestVersionFromSearch(t *testing.T) {
	for out, want := range map[string]string{
		"netbird-0.78.2-r0\n":                           "0.78.2",
		"netbird-openrc-0.78.2-r0\nnetbird-0.78.2-r1\n": "0.78.2",
		"netbird-0.79.0_rc1-r0":                         "0.79.0_rc1",
	} {
		if got, err := versionFromSearch([]byte(out)); err != nil || got != want {
			t.Fatalf("%q: got %q %v, want %q", out, got, err, want)
		}
	}
	if _, err := versionFromSearch([]byte("")); err == nil {
		t.Fatal("no package must be an error")
	}
}

func TestLatestVersionAsksApk(t *testing.T) {
	scratchImage(t, true)
	calls := fakeApk(t, "0.79.0")
	v, err := latestVersion()
	if err != nil || v != "0.79.0" {
		t.Fatalf("got %q %v", v, err)
	}
	if got := callsOf(t, calls); got != "update\nsearch -e netbird\n" {
		t.Fatalf("calls:\n%s", got)
	}
}

// The page waits on this check. A mirror that does not answer must not hold
// it for apk's ten minutes.
func TestLatestVersionGivesUpAfterTheCheckTimeout(t *testing.T) {
	scratchImage(t, true)
	fakeApk(t, "0.79.0")
	slow := filepath.Join(t.TempDir(), "apk")
	stub(t, slow, "sleep 30")
	savedApk, savedTimeout := ApkPath, checkTimeout
	t.Cleanup(func() { ApkPath, checkTimeout = savedApk, savedTimeout })
	ApkPath, checkTimeout = slow, 300*time.Millisecond

	start := time.Now()
	if _, err := latestVersion(); err == nil {
		t.Fatal("a check that timed out must fail")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("the check took %s", elapsed)
	}
}
