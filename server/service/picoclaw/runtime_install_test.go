package picoclaw

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/service/controlmode"
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/upstream"
)

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

const testReleaseDir = "picoclaw-vtest-linux-riscv64"

// fakeArchive is a release laid out as dist/release-picoclaw.sh lays it out:
// one directory holding the binary and the license files.
func fakeArchive(t *testing.T, binary string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: testReleaseDir + "/", Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
		t.Fatal(err)
	}
	files := []struct {
		name, body string
		mode       int64
	}{
		{"GO-MODULES", "example.org/a\tv1.0.0\tMIT\n", 0o644},
		{"LICENSE", "MIT License\n", 0o644},
		{"picoclaw", binary, 0o755},
	}
	for _, f := range files {
		if err := tw.WriteHeader(&tar.Header{Name: testReleaseDir + "/" + f.name, Mode: f.mode, Size: int64(len(f.body))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(f.body)); err != nil {
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

// serveRelease serves body over TLS as the release, pinned to sum, with the
// binary in it pinned to binarySum. It returns the number of requests made.
func serveRelease(t *testing.T, body []byte, sum, binarySum string) *atomic.Int32 {
	t.Helper()
	var hits atomic.Int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	pin(t, srv.URL+"/"+testReleaseDir+".tar.gz", sum, int64(len(body)), binarySum)
	downloads = upstream.NewClient(srv.URL, srv.Client())
	return &hits
}

func pin(t *testing.T, url, sum string, size int64, binarySum string) {
	t.Helper()
	savedRelease, savedDownloads, savedFree, savedOverride := release, downloads, freeBytes, picoclawReleaseURLOverride
	t.Cleanup(func() {
		release, downloads, freeBytes, picoclawReleaseURLOverride = savedRelease, savedDownloads, savedFree, savedOverride
	})
	release = pinnedRelease{
		URL: url, SHA256: sum, Size: size, Max: 1 << 20,
		Dir: testReleaseDir, BinarySHA256: binarySum, BinarySize: 1 << 10,
	}
	freeBytes = func(string) (uint64, error) { return 1 << 40, nil }
	picoclawReleaseURLOverride = filepath.Join(t.TempDir(), "no-override")
}

// installed puts an earlier install of PicoClaw on /data: content, with
// recorded as its sha256 record when it is not empty.
func installed(t *testing.T, content, recorded string) string {
	t.Helper()
	dest := filepath.Join(addon.Dir("picoclaw"), "picoclaw")
	if err := os.MkdirAll(filepath.Join(addon.Dir("picoclaw"), "home"), 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(addon.Dir("picoclaw"), "home", "config.json"), []byte("{}"), 0o600)
	if err := os.WriteFile(dest, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	if recorded != "" {
		_ = os.WriteFile(dest+".sha256", []byte(recorded+"\n"), 0o644)
	}
	return dest
}

func assertGone(t *testing.T, paths ...string) {
	t.Helper()
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			t.Fatalf("%s exists after the install", p)
		}
	}
}

func TestInstallDownloadsThePinnedReleaseOntoData(t *testing.T) {
	fsroot, _ := scratchImage(t, true)
	archive := fakeArchive(t, "fork build")
	serveRelease(t, archive, sha256Hex(string(archive)), sha256Hex("fork build"))

	var stages []string
	err := installReleasePicoclaw(func(stage string, _ int, _ string) {
		if len(stages) == 0 || stages[len(stages)-1] != stage {
			stages = append(stages, stage)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	dir := addon.Dir("picoclaw")
	dest := filepath.Join(dir, "picoclaw")
	if b, err := os.ReadFile(dest); err != nil || string(b) != "fork build" {
		t.Fatalf("the release's binary is not at %s: %q %v", dest, b, err)
	}
	if info, _ := os.Stat(dest); info == nil || info.Mode().Perm() != 0o755 {
		t.Fatalf("the binary's mode is not 0755: %v", info)
	}
	if b, _ := os.ReadFile(dest + ".sha256"); strings.TrimSpace(string(b)) != sha256Hex("fork build") {
		t.Fatalf("the installed sha256 is not recorded: %q", b)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "licenses", "LICENSE")); string(b) != "MIT License\n" {
		t.Fatalf("the license is not kept beside the binary: %q", b)
	}
	if _, err := os.Stat(filepath.Join(dir, "licenses", "picoclaw")); err == nil {
		t.Fatal("the binary was copied into the licenses too")
	}
	if target, err := os.Readlink(fsroot + "/usr/bin/picoclaw"); err != nil || target != dest {
		t.Fatalf("/usr/bin/picoclaw links to %q (%v), want %s", target, err, dest)
	}
	assertGone(t, filepath.Join(dir, ".fetch"), dest+".new")
	if installedPicoclawIsStale() {
		t.Fatal("a fresh install of the pin must not count as stale")
	}
	if got := strings.Join(stages, " "); got != "downloading extracting verifying installing" {
		t.Fatalf("stages %q", got)
	}
}

// A download whose sum is not the pin is refused before anything is
// unpacked, and an install already there is left as it was.
func TestAMismatchedDownloadIsRefusedAndTheOldBinaryKept(t *testing.T) {
	scratchImage(t, true)
	dest := installed(t, "older build", sha256Hex("older build"))
	archive := fakeArchive(t, "fork build")
	serveRelease(t, archive, sha256Hex("another archive"), sha256Hex("fork build"))

	err := installReleasePicoclaw(nil)
	if !errors.Is(err, upstream.ErrChecksum) {
		t.Fatalf("got %v, want a checksum mismatch", err)
	}
	if b, _ := os.ReadFile(dest); string(b) != "older build" {
		t.Fatalf("the installed binary changed: %q", b)
	}
	if b, _ := os.ReadFile(dest + ".sha256"); strings.TrimSpace(string(b)) != sha256Hex("older build") {
		t.Fatalf("the installed record changed: %q", b)
	}
	assertGone(t, filepath.Join(addon.Dir("picoclaw"), ".fetch"), filepath.Join(addon.Dir("picoclaw"), "licenses"))
}

// The archive is the pinned one but the binary in it is not: refused, the old
// binary kept.
func TestABinaryThatIsNotThePinnedOneIsRefused(t *testing.T) {
	scratchImage(t, true)
	dest := installed(t, "older build", sha256Hex("older build"))
	archive := fakeArchive(t, "tampered build")
	serveRelease(t, archive, sha256Hex(string(archive)), sha256Hex("fork build"))

	err := installReleasePicoclaw(nil)
	if err == nil || !strings.Contains(err.Error(), "sha256 mismatch") {
		t.Fatalf("got %v, want a sha256 mismatch", err)
	}
	if b, _ := os.ReadFile(dest); string(b) != "older build" {
		t.Fatalf("the installed binary changed: %q", b)
	}
	assertGone(t, filepath.Join(addon.Dir("picoclaw"), ".fetch"), dest+".new")
}

// A failed first install leaves no add-on directory behind.
func TestADownloadThatFailsLeavesNothing(t *testing.T) {
	scratchImage(t, true)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "gone", http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	pin(t, srv.URL+"/x.tar.gz", sha256Hex("x"), 1, sha256Hex("fork build"))
	downloads = upstream.NewClient(srv.URL, srv.Client())

	err := installReleasePicoclaw(nil)
	if err == nil || !strings.Contains(err.Error(), "status 404") {
		t.Fatalf("got %v, want the server's 404", err)
	}
	assertGone(t, addon.Dir("picoclaw"))
}

// An unreachable server is a failed install, not a hang.
func TestAnUnreachableServerFailsTheInstall(t *testing.T) {
	scratchImage(t, true)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url, client := srv.URL, srv.Client()
	srv.Close()
	pin(t, url+"/x.tar.gz", sha256Hex("x"), 1, sha256Hex("fork build"))
	downloads = upstream.NewClient(url, client)

	if err := installReleasePicoclaw(nil); err == nil {
		t.Fatal("an install from a closed server succeeded")
	}
	assertGone(t, addon.Dir("picoclaw"))
}

// Too little space on /data is refused before the download starts, and the
// old binary stays.
func TestTooLittleSpaceIsRefusedBeforeTheDownload(t *testing.T) {
	scratchImage(t, true)
	dest := installed(t, "older build", sha256Hex("older build"))
	archive := fakeArchive(t, "fork build")
	hits := serveRelease(t, archive, sha256Hex(string(archive)), sha256Hex("fork build"))
	var asked string
	freeBytes = func(path string) (uint64, error) { asked = path; return 1 << 20, nil }

	err := installReleasePicoclaw(nil)
	if !errors.Is(err, errNoSpace) {
		t.Fatalf("got %v, want not enough space", err)
	}
	if hits.Load() != 0 {
		t.Fatalf("%d requests were made after the space check failed", hits.Load())
	}
	if asked != addon.Dir("picoclaw") {
		t.Fatalf("the free space of %s was read, want the add-on's directory", asked)
	}
	if b, _ := os.ReadFile(dest); string(b) != "older build" {
		t.Fatalf("the installed binary changed: %q", b)
	}
}

// Off a distribution image the binary goes to /usr/bin/picoclaw, as before,
// and the fetch directory is the fallback.
func TestOffADistributionImageTheBinaryGoesToUsrBin(t *testing.T) {
	scratchImage(t, false)
	base := t.TempDir()
	archive := fakeArchive(t, "fork build")
	serveRelease(t, archive, sha256Hex(string(archive)), sha256Hex("fork build"))
	if dest := picoclawInstallDestination(); dest != picoclawBinaryPath {
		t.Fatalf("off a distribution image the binary goes to %s, want %s", dest, picoclawBinaryPath)
	}
	saved := []string{picoclawFetchFallback, picoclawLicensesFallback}
	t.Cleanup(func() { picoclawFetchFallback, picoclawLicensesFallback = saved[0], saved[1] })
	picoclawFetchFallback = filepath.Join(base, "fetch")
	picoclawLicensesFallback = filepath.Join(base, "licenses")
	if got := picoclawFetchDir(); got != picoclawFetchFallback {
		t.Fatalf("fetch dir %s", got)
	}
	if got := picoclawLicensesDir(); got != picoclawLicensesFallback {
		t.Fatalf("licenses dir %s", got)
	}
}

// The test aid: a file naming another URL for the same, pinned, file. Plain
// http is allowed there and nowhere else.
func TestTheOverrideNamesAnotherCopyOfThePinnedFile(t *testing.T) {
	scratchImage(t, true)
	archive := fakeArchive(t, "fork build")
	tlsHits := serveRelease(t, archive, sha256Hex(string(archive)), sha256Hex("fork build"))
	var plainHits atomic.Int32
	plain := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		plainHits.Add(1)
		_, _ = w.Write(archive)
	}))
	t.Cleanup(plain.Close)

	picoclawReleaseURLOverride = filepath.Join(t.TempDir(), "picoclaw-release-url")
	if err := os.WriteFile(picoclawReleaseURLOverride, []byte(plain.URL+"/asset.tar.gz\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := installReleasePicoclaw(nil); err != nil {
		t.Fatal(err)
	}
	if plainHits.Load() != 1 || tlsHits.Load() != 0 {
		t.Fatalf("override %d requests, release %d, want 1 and 0", plainHits.Load(), tlsHits.Load())
	}

	// The same file over plain http, without the override, is refused.
	release.URL = plain.URL + "/asset.tar.gz"
	_ = os.Remove(picoclawReleaseURLOverride)
	if err := installReleasePicoclaw(nil); err == nil || !strings.Contains(err.Error(), "not an https URL") {
		t.Fatalf("got %v, want plain http refused", err)
	}

	// The override cannot change the sum: another file there is refused.
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(fakeArchive(t, "other build"))
	}))
	t.Cleanup(other.Close)
	_ = os.WriteFile(picoclawReleaseURLOverride, []byte(other.URL+"/asset.tar.gz\n"), 0o644)
	if err := installReleasePicoclaw(nil); !errors.Is(err, upstream.ErrChecksum) {
		t.Fatalf("got %v, want a checksum mismatch", err)
	}
	if b, _ := os.ReadFile(filepath.Join(addon.Dir("picoclaw"), "picoclaw")); string(b) != "fork build" {
		t.Fatalf("the installed binary changed: %q", b)
	}

	// Something that is not a URL is ignored, and the release is used.
	_ = os.WriteFile(picoclawReleaseURLOverride, []byte("file:///etc/passwd\n"), 0o644)
	if got := readReleaseURLOverride(); got != "" {
		t.Fatalf("override %q, want it ignored", got)
	}
}

// An install that is not the pinned build is offered the update and left
// running as it is: Sipeed's v0.2.8 recorded no sha256, an older fork build
// another one. Nothing replaces it on its own.
func TestAnInstallThatIsNotThePinIsOfferedTheUpdate(t *testing.T) {
	scratchImage(t, true)
	pin(t, "https://example.invalid/x.tar.gz", sha256Hex("x"), 1, sha256Hex("fork build"))

	installed(t, "sipeed v0.2.8", "")
	if !installedPicoclawIsStale() {
		t.Fatal("an install with no recorded sha256 must count as stale")
	}
	installed(t, "older fork build", sha256Hex("older fork build"))
	if !installedPicoclawIsStale() {
		t.Fatal("an install of another build must count as stale")
	}
	status := withRelease(RuntimeStatus{Installed: true})
	if !status.UpdateAvailable || status.Version != Version {
		t.Fatalf("status %+v, want the update offered", status)
	}
	if withRelease(RuntimeStatus{Installed: true, Installing: true}).UpdateAvailable {
		t.Fatal("an update is offered while one runs")
	}
	if withRelease(RuntimeStatus{Installed: false}).UpdateAvailable {
		t.Fatal("an update is offered with nothing installed")
	}

	installed(t, "fork build", sha256Hex("fork build"))
	if installedPicoclawIsStale() || withRelease(RuntimeStatus{Installed: true}).UpdateAvailable {
		t.Fatal("the pinned build counts as stale")
	}
	if b, _ := os.ReadFile(filepath.Join(addon.Dir("picoclaw"), "home", "config.json")); string(b) != "{}" {
		t.Fatalf("the settings changed: %q", b)
	}
}

// The pin itself: an https URL on the fork's releases, named for the version,
// with sums and bounds that make sense together.
func TestThePinIsTheForksRelease(t *testing.T) {
	if !strings.HasPrefix(release.URL, "https://github.com/yuzi-co/picoclaw/releases/download/") {
		t.Fatalf("release URL %s is not the fork's release", release.URL)
	}
	if !strings.HasSuffix(release.URL, "/"+release.Dir+".tar.gz") || !strings.Contains(release.Dir, Version) {
		t.Fatalf("release URL %s does not name %s and %s", release.URL, release.Dir, Version)
	}
	// The release tag is the version, <upstream>-ironkvm.N (ironkvm-dist's
	// docs/versioning.md): the build takes its version from git describe, so
	// the binary reports the tag it was released under.
	tag := strings.Split(strings.TrimPrefix(release.URL, "https://github.com/yuzi-co/picoclaw/releases/download/"), "/")[0]
	if tag != Version || !regexp.MustCompile(`^v\d+\.\d+\.\d+-ironkvm\.[1-9]\d*$`).MatchString(tag) {
		t.Fatalf("tag %s is not the version %s, <upstream>-ironkvm.N", tag, Version)
	}
	for _, sum := range []string{release.SHA256, release.BinarySHA256} {
		if b, err := hex.DecodeString(sum); err != nil || len(b) != sha256.Size {
			t.Fatalf("%q is not a SHA-256", sum)
		}
	}
	if release.Size <= 0 || release.Size > release.Max || release.BinarySize <= release.Size {
		t.Fatalf("sizes %d, %d, bound %d", release.Size, release.BinarySize, release.Max)
	}
}

// Remove takes the add-on away, settings and licenses included, and a second
// install puts it back.
func TestUninstallRemovesTheAddonAndAReinstallBringsItBack(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fsroot, _ := scratchImage(t, true)
	savedUmount := addon.UmountCmd
	t.Cleanup(func() { addon.UmountCmd = savedUmount })
	addon.UmountCmd = "true"
	archive := fakeArchive(t, "fork build")
	serveRelease(t, archive, sha256Hex(string(archive)), sha256Hex("fork build"))
	home := filepath.Join(t.TempDir(), "home")
	_ = os.MkdirAll(home, 0o755)
	t.Setenv("PICOCLAW_HOME", home)
	if err := installReleasePicoclaw(nil); err != nil {
		t.Fatal(err)
	}

	tmp := t.TempDir()
	service := &Service{
		control:       controlmode.NewManager(filepath.Join(tmp, "mode"), controlmode.ModeOff),
		runtime:       &RuntimeStore{},
		runtimeIntent: NewRuntimeIntentStore(filepath.Join(tmp, "intent.json")),
		releaseHID:    func() error { return nil },
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/picoclaw/runtime/uninstall", nil)
	service.UninstallRuntime(c)
	if recorder.Code != http.StatusOK {
		t.Fatalf("uninstall answered %d: %s", recorder.Code, recorder.Body.String())
	}
	var body struct {
		Data RuntimeInstallResult `json:"data"`
	}
	_ = json.Unmarshal(recorder.Body.Bytes(), &body)
	if body.Data.Installed || body.Data.Status.Status != "not_installed" {
		t.Fatalf("after uninstall: %s", recorder.Body.String())
	}
	assertGone(t, addon.Dir("picoclaw"), fsroot+"/usr/bin/picoclaw", home)

	if err := installReleasePicoclaw(nil); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(filepath.Join(addon.Dir("picoclaw"), "picoclaw")); string(b) != "fork build" {
		t.Fatalf("the reinstall did not put the binary back: %q", b)
	}
}
