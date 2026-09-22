package picoclaw

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"NanoKVM-Server/service/extensions/addon"
)

// scratchImage points the add-on package at a temporary root. onDistro decides
// whether it looks like a distribution image with /data mounted. mount is a
// stub that records the bind and adds the line a real one would.
func scratchImage(t *testing.T, onDistro bool) (fsroot, calls string) {
	t.Helper()
	base := t.TempDir()
	fsroot = filepath.Join(base, "root")
	data := filepath.Join(base, "data")
	for _, d := range []string{fsroot + "/usr/bin", fsroot + "/root", data} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	marker := filepath.Join(base, "deviceinfo")
	if onDistro {
		_ = os.WriteFile(marker, []byte("DEVICE=one-board\n"), 0o644)
	}
	mounts := filepath.Join(base, "mounts")
	_ = os.WriteFile(mounts, []byte("/dev/x "+data+" exfat rw 0 0\n"), 0o644)
	calls = filepath.Join(base, "calls")
	stub := filepath.Join(base, "mount")
	_ = os.WriteFile(stub, []byte("#!/bin/sh\necho \"mount $*\" >> \""+calls+"\"; echo \"x $3 none rw 0 0\" >> \""+mounts+"\"\n"), 0o755)

	saved := []string{addon.Root, addon.DataDir, addon.DistroMarker, addon.Mounts, addon.FsRoot, addon.MountCmd}
	t.Cleanup(func() {
		addon.Root, addon.DataDir, addon.DistroMarker, addon.Mounts, addon.FsRoot, addon.MountCmd =
			saved[0], saved[1], saved[2], saved[3], saved[4], saved[5]
	})
	addon.Root = filepath.Join(data, "ironkvm", "addons")
	addon.DataDir = data
	addon.DistroMarker = marker
	addon.Mounts = mounts
	addon.FsRoot = fsroot
	addon.MountCmd = stub
	return fsroot, calls
}

func TestPicoclawInstallsOntoDataOnADistributionImage(t *testing.T) {
	fsroot, calls := scratchImage(t, true)
	// Settings the owner made before this change, on the slot.
	_ = os.MkdirAll(fsroot+"/root/.picoclaw", 0o755)
	_ = os.WriteFile(fsroot+"/root/.picoclaw/config.json", []byte(`{"model":"m"}`), 0o600)

	dest := picoclawInstallDestination()
	if want := filepath.Join(addon.Dir("picoclaw"), "picoclaw"); dest != want {
		t.Fatalf("the binary goes to %s, want %s on /data", dest, want)
	}
	_ = os.MkdirAll(filepath.Dir(dest), 0o755)
	_ = os.WriteFile(dest, []byte("bin"), 0o755)
	if err := recordPicoclawAddon(); err != nil {
		t.Fatal(err)
	}

	if target, err := os.Readlink(fsroot + "/usr/bin/picoclaw"); err != nil || target != dest {
		t.Fatalf("/usr/bin/picoclaw links to %q (%v), want %s", target, err, dest)
	}
	b, err := os.ReadFile(filepath.Join(addon.Dir("picoclaw"), "home", "config.json"))
	if err != nil || string(b) != `{"model":"m"}` {
		t.Fatalf("the owner's settings were not carried onto /data: %q %v", b, err)
	}
	c, _ := os.ReadFile(calls)
	if !strings.Contains(string(c), "--bind "+filepath.Join(addon.Dir("picoclaw"), "home")+" "+fsroot+"/root/.picoclaw") {
		t.Fatalf("/root/.picoclaw was not bound from /data: %q", c)
	}
	if _, err := os.Stat(filepath.Join(addon.Dir("picoclaw"), "initd")); err == nil {
		t.Fatal("PicoClaw has no boot script, so it must have no initd")
	}
}

func TestPicoclawElsewhereInstallsAsItAlwaysDid(t *testing.T) {
	scratchImage(t, false)
	if dest := picoclawInstallDestination(); dest != picoclawBinaryPath {
		t.Fatalf("off a distribution image the binary goes to %s, want %s", dest, picoclawBinaryPath)
	}
	if err := recordPicoclawAddon(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(addon.Dir("picoclaw")); err == nil {
		t.Fatal("nothing may be written to /data off a distribution image")
	}
}

// The add-on's directory does not exist before the first install, and the
// install is what makes it.
func TestInstallBinaryMakesItsDirectory(t *testing.T) {
	scratchImage(t, true)
	base := t.TempDir()
	source := filepath.Join(base, "picoclaw")
	if err := os.WriteFile(source, []byte("bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	dest := picoclawInstallDestination()
	if _, err := os.Stat(filepath.Dir(dest)); err == nil {
		t.Fatal("the add-on directory must not exist yet, or this proves nothing")
	}
	if err := installPicoclawBinary(source, dest); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(dest); err != nil || string(b) != "bin" {
		t.Fatalf("the binary is not at %s: %q %v", dest, b, err)
	}
}
