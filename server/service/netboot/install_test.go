//go:build linux

package netboot

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"NanoKVM-Server/service/extensions/addon"
)

func TestInstallPutsDnsmasqAndTheBootFilesOnData(t *testing.T) {
	scratchImage(t, true)
	calls := fakeApk(t, "2.92_p2-r0")
	requests := fakeReleases(t)

	if err := install(); err != nil {
		t.Fatal(err)
	}

	if b, err := os.ReadFile(dnsmasqPath()); err != nil || string(b) != "dnsmasq 2.92_p2-r0\n" {
		t.Fatalf("dnsmasq is not on /data: %q %v", b, err)
	}
	if fi, err := os.Stat(dnsmasqPath()); err != nil || fi.Mode().Perm()&0o111 == 0 {
		t.Fatal("dnsmasq must be executable")
	}
	if b, _ := os.ReadFile(versionPath()); string(b) != "2.92_p2-r0\n" {
		t.Fatalf("version %q", b)
	}
	for _, name := range bootFiles() {
		if _, err := os.Stat(filepath.Join(tftpRoot(), name)); err != nil {
			t.Errorf("%s is missing from the TFTP root", name)
		}
	}
	if b, _ := os.ReadFile(filepath.Join(tftpRoot(), ipxeEFIarm64)); string(b) != "ipxe "+ipxeEFIarm64 {
		t.Fatalf("arm64 iPXE holds %q", b)
	}
	if b, _ := os.ReadFile(filepath.Join(tftpRoot(), bootScriptName)); string(b) != bootScript() {
		t.Fatalf("boot.ipxe holds %q", b)
	}
	if !installed() {
		t.Fatal("installed must answer true")
	}

	// Only the three chosen members come out of the archive.
	entries, _ := os.ReadDir(tftpRoot())
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	slices.Sort(names)
	want := bootFiles()
	slices.Sort(want)
	if !slices.Equal(names, want) {
		t.Fatalf("the TFTP root holds %q, want %q", names, want)
	}

	// No link in /usr/sbin: S80dnsmasq would start a dnsmasq it finds there.
	if b, err := os.ReadFile(filepath.Join(addon.Dir(AddonName), "initd")); err != nil || string(b) != Initd+"\n" {
		t.Fatalf("initd is %q (%v)", b, err)
	}
	if _, err := os.Stat(filepath.Join(addon.Dir(AddonName), "links")); err == nil {
		t.Fatal("the add-on records a link")
	}
	if _, err := os.Stat(filepath.Join(addon.Dir(AddonName), ".fetch")); err == nil {
		t.Fatal("the fetch workspace is left on /data")
	}

	got, _ := os.ReadFile(calls)
	if !regexp.MustCompile(`(?s)^update\nfetch -o \S+ dnsmasq\nverify \S+dnsmasq-2.92_p2-r0.apk\nextract --no-chown`).Match(got) {
		t.Fatalf("apk was run as:\n%s", got)
	}

	// A second install finds everything and downloads nothing.
	before := requests.Load()
	if err := install(); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != before {
		t.Fatalf("a second install downloaded %d files", requests.Load()-before)
	}
}

func TestInstallNeedsData(t *testing.T) {
	scratchImage(t, false)
	fakeApk(t, "2.92_p2-r0")
	fakeReleases(t)

	err := install()
	if err == nil || !strings.Contains(err.Error(), "/data") {
		t.Fatalf("an install off a distribution image: %v", err)
	}
}

func TestAPackageWithABadSignatureIsNotInstalled(t *testing.T) {
	scratchImage(t, true)
	fakeApk(t, "2.92_p2-r0")
	fakeReleases(t)
	t.Setenv("BAD_SIGNATURE", "1")

	if err := install(); err == nil || !strings.Contains(err.Error(), "UNTRUSTED") {
		t.Fatalf("install with a bad signature: %v", err)
	}
	if dnsmasqInstalled() {
		t.Fatal("dnsmasq was installed from a package that failed verify")
	}
}

func TestADownloadThatDoesNotMatchItsPinIsRefused(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("tampered"))
	}))
	defer srv.Close()
	saved := httpClient
	defer func() { httpClient = saved }()
	httpClient = srv.Client()

	dst := filepath.Join(t.TempDir(), "file")
	err := fetchPinned(srv.URL, sum([]byte("original")), 1<<20, dst)
	if err == nil || !strings.Contains(err.Error(), "sha256") {
		t.Fatalf("a wrong sum: %v", err)
	}
	if _, err := os.Stat(dst); err == nil {
		t.Fatal("a file that failed its pin was kept")
	}
	if _, err := os.Stat(dst + ".part"); err == nil {
		t.Fatal("the partial file was left")
	}

	if err := fetchPinned(srv.URL, sum([]byte("tampered")), 4, dst); err == nil {
		t.Fatal("a file larger than its bound was accepted")
	}
	if err := fetchPinned(srv.URL, sum([]byte("tampered")), 1<<20, dst); err != nil {
		t.Fatal(err)
	}
}

func TestExtractTakesOnlyTheNamedMembersEachCheckedByItsSum(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "a.tar.gz")
	content := tarGz(t, []tarEntry{
		{name: "ipxeboot/x86_64/ipxe.efi", content: "efi"},
		{name: "ipxeboot/ipxe.efi", linkname: "x86_64/ipxe.efi"},
		{name: "../escape", content: "out"},
		{name: "ipxeboot/other", content: "other"},
	})
	if err := os.WriteFile(archive, content, 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}

	good := []archiveMember{{Member: "ipxeboot/x86_64/ipxe.efi", Name: "ipxe.efi", SHA256: sum([]byte("efi"))}}
	if err := extractMembers(archive, good, out); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(out)
	if len(entries) != 1 || entries[0].Name() != "ipxe.efi" {
		t.Fatalf("extracted %v", entries)
	}
	if _, err := os.Stat(filepath.Join(dir, "escape")); err == nil {
		t.Fatal("a member outside the archive's root was written")
	}

	bad := []archiveMember{{Member: "ipxeboot/x86_64/ipxe.efi", Name: "bad.efi", SHA256: sum([]byte("else"))}}
	if err := extractMembers(archive, bad, out); err == nil || !strings.Contains(err.Error(), "sha256") {
		t.Fatalf("a member with the wrong sum: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "bad.efi")); err == nil {
		t.Fatal("a member with the wrong sum was kept")
	}

	missing := []archiveMember{{Member: "ipxeboot/arm64/ipxe.efi", Name: "arm.efi", SHA256: sum([]byte("x"))}}
	if err := extractMembers(archive, missing, out); err == nil || !strings.Contains(err.Error(), "ipxeboot/arm64/ipxe.efi") {
		t.Fatalf("a missing member: %v", err)
	}
}

// The pins are what was checked by hand against the releases.
func TestThePinsAreWellFormed(t *testing.T) {
	hexSum := regexp.MustCompile(`^[0-9a-f]{64}$`)

	all := append([]pinnedFile{ipxeArchive, {Name: "iso", URL: BootMenuISOURL, SHA256: BootMenuISOSHA256}}, netbootXYZFiles...)
	for _, f := range all {
		if !hexSum.MatchString(f.SHA256) {
			t.Errorf("%s: %q is not a SHA-256", f.Name, f.SHA256)
		}
		if !strings.HasPrefix(f.URL, "https://github.com/") {
			t.Errorf("%s: %q is not a release asset", f.Name, f.URL)
		}
	}
	for _, m := range ipxeMembers {
		if !hexSum.MatchString(m.SHA256) {
			t.Errorf("%s: %q is not a SHA-256", m.Member, m.SHA256)
		}
		if !strings.HasPrefix(m.Member, "ipxeboot/") {
			t.Errorf("%s is not in the archive's root", m.Member)
		}
	}
	for _, f := range netbootXYZFiles {
		if !strings.Contains(f.URL, "/"+netbootXYZVersion+"/") || filepath.Base(f.URL) != f.Name {
			t.Errorf("%s is fetched from %s", f.Name, f.URL)
		}
	}
	if !strings.Contains(BootMenuISOURL, "/"+netbootXYZVersion+"/netboot.xyz.iso") {
		t.Errorf("the ISO is %s", BootMenuISOURL)
	}
	if !strings.Contains(ipxeArchive.URL, "/v"+ipxeVersion+"/") {
		t.Errorf("the iPXE archive is %s", ipxeArchive.URL)
	}

	files := bootFiles()
	seen := map[string]bool{}
	for _, name := range files {
		if seen[name] {
			t.Errorf("%s is installed twice", name)
		}
		seen[name] = true
	}
}
