//go:build linux

package netbird

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/upstream"
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

	saved := struct{ root, dataDir, marker, mounts, fsroot, nb, ws, keys string }{
		addon.Root, addon.DataDir, addon.DistroMarker, addon.Mounts, addon.FsRoot, NetbirdPath, FallbackWorkspace, KeyDir}
	t.Cleanup(func() {
		addon.Root, addon.DataDir, addon.DistroMarker, addon.Mounts, addon.FsRoot = saved.root, saved.dataDir, saved.marker, saved.mounts, saved.fsroot
		NetbirdPath, FallbackWorkspace, KeyDir = saved.nb, saved.ws, saved.keys
	})
	addon.Root = filepath.Join(data, "ironkvm", "addons")
	addon.DataDir = data
	addon.DistroMarker = marker
	addon.Mounts = mounts
	addon.FsRoot = fsroot
	NetbirdPath = fsroot + "/usr/bin/netbird"
	FallbackWorkspace = filepath.Join(base, "fetch")
	KeyDir = filepath.Join(base, "run-keys")
	if err := os.MkdirAll(KeyDir, 0o755); err != nil {
		t.Fatal(err)
	}
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

// fakeApk puts a stub in place of apk-tools 3. It answers verify and extract,
// unpacks the package with tar, and appends every call to the file it
// returns. BAD_SIGNATURE in the environment makes verify and extract refuse
// the package, as apk does a tampered one.
func fakeApk(t *testing.T) (calls string) {
	t.Helper()
	dir := t.TempDir()
	calls = filepath.Join(dir, "calls")
	path := filepath.Join(dir, "apk")
	stub(t, path, fmt.Sprintf(`echo "$*" >> %q
case "$1" in
verify) [ -z "$BAD_SIGNATURE" ] || { echo "$2: UNTRUSTED signature" >&2; exit 1; } ;;
extract)
	[ -z "$BAD_SIGNATURE" ] || { echo "ERROR: $5: UNTRUSTED signature" >&2; exit 1; }
	tar -xzf "$5" -C "$4" ;;
*) echo "unexpected: $*" >&2; exit 2 ;;
esac`, calls))

	saved := ApkPath
	t.Cleanup(func() { ApkPath = saved })
	ApkPath = path
	return calls
}

// fakePackage is a package as the stub apk reads it: a tar.gz holding
// usr/bin/netbird with the given content.
func fakePackage(t *testing.T, binary string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, d := range []string{"usr/", "usr/bin/"} {
		if err := tw.WriteHeader(&tar.Header{Name: d, Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.WriteHeader(&tar.Header{Name: "usr/bin/netbird", Mode: 0o755, Size: int64(len(binary))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte(binary)); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// servePackage serves body over TLS as the pinned package, pinned to sum
// with the bound limit.
func servePackage(t *testing.T, body []byte, sum string, limit int64) {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	savedPin, savedDownloads := releasePackage, downloads
	t.Cleanup(func() { releasePackage, downloads = savedPin, savedDownloads })
	releasePackage = pinnedFile{URL: srv.URL + "/" + packageFile, SHA256: sum, Max: limit}
	downloads = upstream.NewClient(srv.URL, srv.Client())
}

func sha256Hex(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}
