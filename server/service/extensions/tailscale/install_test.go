package tailscale

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/upstream"
)

// fakeRelease is a package laid out as pkgs.tailscale.com's: one directory
// holding the two binaries, each holding its own name.
func fakeRelease(t *testing.T, pad int) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	dir := "tailscale_" + Version + "_riscv64/"
	if err := tw.WriteHeader(&tar.Header{Name: dir, Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"tailscaled", "tailscale"} {
		body := name + strings.Repeat("\n", pad)
		if err := tw.WriteHeader(&tar.Header{Name: dir + name, Mode: 0o755, Size: int64(len(body))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func sha256Hex(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
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
	releasePackage = pinnedFile{URL: srv.URL + "/tailscale_" + Version + "_riscv64.tgz", SHA256: sum, Max: limit}
	downloads = upstream.NewClient(srv.URL, srv.Client())
}

func TestInstallTakesThePinnedPackageOntoData(t *testing.T) {
	fsroot, _ := scratchImage(t, true)
	pkg := fakeRelease(t, 0)
	servePackage(t, pkg, sha256Hex(pkg), 1<<20)

	if err := install(); err != nil {
		t.Fatal(err)
	}
	dir := addon.Dir("tailscale")
	for path, file := range map[string]string{
		fsroot + "/usr/bin/tailscale":   "tailscale",
		fsroot + "/usr/sbin/tailscaled": "tailscaled",
	} {
		if b, err := os.ReadFile(filepath.Join(dir, file)); err != nil || string(b) != file {
			t.Fatalf("%s is not on /data: %q %v", file, b, err)
		}
		if target, err := os.Readlink(path); err != nil || target != filepath.Join(dir, file) {
			t.Fatalf("%s is not a link into the add-on: %q %v", path, target, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".fetch")); err == nil {
		t.Fatal("the package must not be left on /data")
	}
}

func TestInstallElsewhereRemovesTheWorkspace(t *testing.T) {
	fsroot, _ := scratchImage(t, false)
	pkg := fakeRelease(t, 0)
	servePackage(t, pkg, sha256Hex(pkg), 1<<20)

	if err := install(); err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Lstat(fsroot + "/usr/sbin/tailscaled"); err != nil || !fi.Mode().IsRegular() {
		t.Fatalf("tailscaled must be the binary itself off a distribution image: %v", err)
	}
	if _, err := os.Stat(FallbackWorkspace); err == nil {
		t.Fatal("the workspace must be removed")
	}
}

// A package whose sum is not the pin is refused before it is unpacked, and
// leaves nothing behind.
func TestInstallRefusesAMismatchedPackage(t *testing.T) {
	fsroot, _ := scratchImage(t, true)
	pkg := fakeRelease(t, 0)
	servePackage(t, pkg, sha256Hex([]byte("another package")), 1<<20)

	if err := install(); !errors.Is(err, upstream.ErrChecksum) {
		t.Fatalf("got %v", err)
	}
	assertNothingLeft(t, fsroot)
}

// A package larger than its bound is refused, whatever its sum.
func TestInstallRefusesAnOversizePackage(t *testing.T) {
	fsroot, _ := scratchImage(t, true)
	pkg := fakeRelease(t, 4096)
	servePackage(t, pkg, sha256Hex(pkg), int64(len(pkg)-1))

	if err := install(); err == nil || !strings.Contains(err.Error(), "larger than") {
		t.Fatalf("got %v", err)
	}
	assertNothingLeft(t, fsroot)
}

func assertNothingLeft(t *testing.T, fsroot string) {
	t.Helper()
	for _, path := range []string{fsroot + "/usr/bin/tailscale", fsroot + "/usr/sbin/tailscaled"} {
		if _, err := os.Lstat(path); err == nil {
			t.Fatalf("%s must not be installed", path)
		}
	}
	if _, err := os.Stat(addon.Dir("tailscale")); err == nil {
		t.Fatal("a failed first install must leave no add-on directory behind")
	}
}

// The pin is what the license manifest and the update offer name.
func TestThePinIsOneRelease(t *testing.T) {
	want := "https://pkgs.tailscale.com/stable/tailscale_" + Version + "_riscv64.tgz"
	if releasePackage.URL != want {
		t.Fatalf("the pinned URL is %s, want %s", releasePackage.URL, want)
	}
	if len(releasePackage.SHA256) != 64 {
		t.Fatalf("not a SHA-256: %q", releasePackage.SHA256)
	}
}
