package tailscale

import (
	"os"
	"path/filepath"
	"testing"

	"NanoKVM-Server/service/extensions/addon"
)

// scratchImage points the add-on package and the two binary paths at a
// temporary root. onDistro decides whether it looks like a distribution image
// with /data mounted.
func scratchImage(t *testing.T, onDistro bool) (fsroot, unpacked string) {
	t.Helper()
	base := t.TempDir()
	fsroot = filepath.Join(base, "root")
	data := filepath.Join(base, "data")
	unpacked = filepath.Join(base, "unpacked")
	for _, d := range []string{fsroot + "/usr/bin", fsroot + "/usr/sbin", data, unpacked} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{"tailscale", "tailscaled"} {
		if err := os.WriteFile(filepath.Join(unpacked, f), []byte(f), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	marker := filepath.Join(base, "deviceinfo")
	if onDistro {
		_ = os.WriteFile(marker, []byte("DEVICE=one-board\n"), 0o644)
	}
	mounts := filepath.Join(base, "mounts")
	_ = os.WriteFile(mounts, []byte("/dev/x "+data+" exfat rw 0 0\n"), 0o644)

	saved := struct{ root, dataDir, marker, mounts, fsroot, ts, tsd string }{
		addon.Root, addon.DataDir, addon.DistroMarker, addon.Mounts, addon.FsRoot, TailscalePath, TailscaledPath}
	t.Cleanup(func() {
		addon.Root, addon.DataDir, addon.DistroMarker, addon.Mounts, addon.FsRoot = saved.root, saved.dataDir, saved.marker, saved.mounts, saved.fsroot
		TailscalePath, TailscaledPath = saved.ts, saved.tsd
	})
	addon.Root = filepath.Join(data, "ironkvm", "addons")
	addon.DataDir = data
	addon.DistroMarker = marker
	addon.Mounts = mounts
	addon.FsRoot = fsroot
	TailscalePath = fsroot + "/usr/bin/tailscale"
	TailscaledPath = fsroot + "/usr/sbin/tailscaled"
	return fsroot, unpacked
}

func TestPlaceBinariesOnADistributionImageGoesOntoData(t *testing.T) {
	fsroot, unpacked := scratchImage(t, true)
	if err := placeBinaries(unpacked); err != nil {
		t.Fatal(err)
	}
	dir := addon.Dir("tailscale")
	for path, file := range map[string]string{
		fsroot + "/usr/bin/tailscale":   "tailscale",
		fsroot + "/usr/sbin/tailscaled": "tailscaled",
	} {
		if b, err := os.ReadFile(filepath.Join(dir, file)); err != nil || string(b) != file {
			t.Fatalf("%s is not on /data: %v", file, err)
		}
		if target, err := os.Readlink(path); err != nil || target != filepath.Join(dir, file) {
			t.Fatalf("%s is not a link into the add-on: %q %v", path, target, err)
		}
	}
	b, err := os.ReadFile(filepath.Join(dir, "initd"))
	if err != nil || string(b) != "S98tailscaled\n" {
		t.Fatalf("initd is %q (%v)", b, err)
	}
	if !isInstalled() {
		t.Fatal("isInstalled must follow the links and answer true")
	}
}

func TestPlaceBinariesElsewhereKeepsTheOldPaths(t *testing.T) {
	fsroot, unpacked := scratchImage(t, false)
	if err := placeBinaries(unpacked); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{fsroot + "/usr/bin/tailscale", fsroot + "/usr/sbin/tailscaled"} {
		fi, err := os.Lstat(path)
		if err != nil || !fi.Mode().IsRegular() {
			t.Fatalf("%s must be the binary itself off a distribution image: %v", path, err)
		}
	}
	if _, err := os.Stat(addon.Dir("tailscale")); err == nil {
		t.Fatal("nothing may be written to /data off a distribution image")
	}
}

func TestRecordEnabledOnlyOnData(t *testing.T) {
	scratchImage(t, true)
	if err := recordEnabled(true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(addon.Dir("tailscale"), "enabled")); err != nil {
		t.Fatal("start must record enabled on /data")
	}
	if err := recordEnabled(false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(addon.Dir("tailscale"), "enabled")); err == nil {
		t.Fatal("stop must clear enabled on /data")
	}

	scratchImage(t, false)
	if err := recordEnabled(true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(addon.Dir("tailscale")); err == nil {
		t.Fatal("nothing may be written to /data off a distribution image")
	}
}
