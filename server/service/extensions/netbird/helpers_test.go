//go:build linux

package netbird

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"NanoKVM-Server/service/extensions/addon"
)

// scratchImage points the add-on package and NetbirdPath at a temporary root.
// onDistro decides whether it looks like a distribution image with /data
// mounted. Off one, the fetch goes to a scratch FallbackWorkspace.
func scratchImage(t *testing.T, onDistro bool) (fsroot string) {
	t.Helper()
	base := t.TempDir()
	fsroot = filepath.Join(base, "root")
	data := filepath.Join(base, "data")
	for _, d := range []string{fsroot + "/usr/bin", data} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	marker := filepath.Join(base, "deviceinfo")
	if onDistro {
		if err := os.WriteFile(marker, []byte("DEVICE=one-board\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mounts := filepath.Join(base, "mounts")
	if err := os.WriteFile(mounts, []byte("/dev/x "+data+" exfat rw 0 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	saved := struct{ root, dataDir, marker, mounts, fsroot, nb, ws string }{
		addon.Root, addon.DataDir, addon.DistroMarker, addon.Mounts, addon.FsRoot, NetbirdPath, FallbackWorkspace}
	t.Cleanup(func() {
		addon.Root, addon.DataDir, addon.DistroMarker, addon.Mounts, addon.FsRoot = saved.root, saved.dataDir, saved.marker, saved.mounts, saved.fsroot
		NetbirdPath, FallbackWorkspace = saved.nb, saved.ws
	})
	addon.Root = filepath.Join(data, "ironkvm", "addons")
	addon.DataDir = data
	addon.DistroMarker = marker
	addon.Mounts = mounts
	addon.FsRoot = fsroot
	NetbirdPath = fsroot + "/usr/bin/netbird"
	FallbackWorkspace = filepath.Join(base, "fetch")
	return fsroot
}

func stub(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func callsOf(t *testing.T, calls string) string {
	t.Helper()
	b, err := os.ReadFile(calls)
	if err != nil {
		return ""
	}
	return string(b)
}

// fakeApk puts a stub in place of apk-tools 3. It answers update, search,
// fetch, verify and extract, builds the package it fetches with tar, and
// appends every call to the file it returns. BAD_SIGNATURE in the environment
// makes verify and extract refuse the package, as apk does a tampered one.
func fakeApk(t *testing.T, version string) (calls string) {
	t.Helper()
	dir := t.TempDir()
	calls = filepath.Join(dir, "calls")
	src := filepath.Join(dir, "src")
	path := filepath.Join(dir, "apk")
	stub(t, path, fmt.Sprintf(`echo "$*" >> %q
case "$1" in
update) echo "OK: 25418 distinct packages available" ;;
search) echo "netbird-%[2]s-r0" ;;
fetch)
	mkdir -p %[3]q/usr/bin
	printf 'netbird %[2]s\n' > %[3]q/usr/bin/netbird
	tar -czf "$3/netbird-%[2]s-r0.apk" -C %[3]q usr
	echo "Downloading netbird-%[2]s-r0" ;;
verify) [ -z "$BAD_SIGNATURE" ] || { echo "$2: UNTRUSTED signature" >&2; exit 1; } ;;
extract)
	[ -z "$BAD_SIGNATURE" ] || { echo "ERROR: $5: UNTRUSTED signature" >&2; exit 1; }
	tar -xzf "$5" -C "$4" ;;
*) echo "unexpected: $*" >&2; exit 2 ;;
esac`, calls, version, src))

	saved := ApkPath
	t.Cleanup(func() { ApkPath = saved })
	ApkPath = path
	return calls
}
