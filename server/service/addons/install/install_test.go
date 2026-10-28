package install

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/upstream"
)

func sum(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

type entry struct {
	name, body string
	typ        byte
	link       string
}

func tarGz(t *testing.T, entries []entry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		typ := e.typ
		if typ == 0 {
			typ = tar.TypeReg
		}
		hdr := &tar.Header{Name: e.name, Typeflag: typ, Mode: 0o644, Size: int64(len(e.body)), Linkname: e.link}
		if typ != tar.TypeReg {
			hdr.Size = 0
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if typ == tar.TypeReg {
			if _, err := tw.Write([]byte(e.body)); err != nil {
				t.Fatal(err)
			}
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

func stub(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// board is a scratch distribution image: /data mounted, an add-on root on
// it, and stubs for mount, umount and mksquashfs that log their arguments.
// The mount stub adds the line a real mount would, so the next detach sees
// it, and the umount stub takes it out again.
type board struct {
	base, calls, mounts string
}

func scratch(t *testing.T, onDistro, squashfs, mksquashfs bool) *board {
	t.Helper()
	b := &board{base: t.TempDir()}
	data := filepath.Join(b.base, "data")
	if err := os.MkdirAll(data, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(b.base, "deviceinfo")
	if onDistro {
		_ = os.WriteFile(marker, []byte("DEVICE=one-board\n"), 0o644)
	}
	b.mounts = filepath.Join(b.base, "mounts")
	_ = os.WriteFile(b.mounts, []byte("/dev/x "+data+" exfat rw 0 0\n"), 0o644)
	b.calls = filepath.Join(b.base, "calls")
	fs := "nodev\tproc\n\text4\n"
	if squashfs {
		fs += "\tsquashfs\n"
	}
	filesystems := filepath.Join(b.base, "filesystems")
	_ = os.WriteFile(filesystems, []byte(fs), 0o644)

	mountStub := filepath.Join(b.base, "mount")
	stub(t, mountStub, fmt.Sprintf(`echo "mount $*" >> %q; for last; do :; done; echo "x $last squashfs ro 0 0" >> %q`, b.calls, b.mounts))
	umountStub := filepath.Join(b.base, "umount")
	stub(t, umountStub, fmt.Sprintf(`echo "umount $*" >> %q; for last; do :; done; grep -v " $last " %q > %q.new; mv %q.new %q`, b.calls, b.mounts, b.mounts, b.mounts, b.mounts))
	// The fake mksquashfs writes, as its "image", the staged files' names
	// and contents and the pseudo file, so a test can see what went in.
	mksq := filepath.Join(b.base, "mksquashfs")
	if mksquashfs {
		stub(t, mksq, fmt.Sprintf(`echo "mksquashfs $*" >> %q
src=$1; out=$2; pf=
while [ $# -gt 0 ]; do [ "$1" = -pf ] && pf=$2; shift; done
{ (cd "$src" && find . -type f | LC_ALL=C sort | while read -r f; do echo "== $f"; cat "$f"; done); echo "== pseudo"; cat "$pf"; } > "$out"`, b.calls))
	}

	saved := []string{addon.Root, addon.DataDir, addon.DistroMarker, addon.Mounts, addon.FsRoot, addon.MountCmd, addon.UmountCmd, Mksquashfs, Filesystems, ApkPath}
	savedFree := FreeBytes
	t.Cleanup(func() {
		addon.Root, addon.DataDir, addon.DistroMarker, addon.Mounts, addon.FsRoot, addon.MountCmd, addon.UmountCmd, Mksquashfs, Filesystems, ApkPath =
			saved[0], saved[1], saved[2], saved[3], saved[4], saved[5], saved[6], saved[7], saved[8], saved[9]
		FreeBytes = savedFree
	})
	addon.Root = filepath.Join(data, "ironkvm", "addons")
	addon.DataDir = data
	addon.DistroMarker = marker
	addon.Mounts = b.mounts
	addon.FsRoot = filepath.Join(b.base, "root")
	addon.MountCmd = mountStub
	addon.UmountCmd = umountStub
	Mksquashfs = mksq
	Filesystems = filesystems
	FreeBytes = func(string) (uint64, error) { return 1 << 40, nil }
	return b
}

func (b *board) log(t *testing.T) string {
	t.Helper()
	data, _ := os.ReadFile(b.calls)
	return string(data)
}

// serve serves body over TLS and returns a manifest for it and a client
// that trusts the server.
func serve(t *testing.T, version string, body []byte, members []Member) (Manifest, *upstream.Client) {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return Manifest{
		Name: "tool", Title: "Tool", Version: version, Licence: "MIT",
		Source: "https://example.org/tool",
		Artifact: Artifact{
			URL: srv.URL + "/tool-" + version + ".tar.gz", SHA256: sum(body), Max: 1 << 20,
			Format: FormatTarGz, Members: members,
		},
	}, upstream.NewClient(srv.URL, srv.Client())
}

// release is an archive with a binary, a licence, and files nobody listed.
func release(t *testing.T, binary string) ([]byte, []Member) {
	t.Helper()
	body := tarGz(t, []entry{
		{name: "tool-1/", typ: tar.TypeDir},
		{name: "tool-1/tool", body: binary},
		{name: "tool-1/LICENSE", body: "MIT License\n"},
		{name: "tool-1/sub/", typ: tar.TypeDir},
		{name: "tool-1/sub/deeper", body: "not a licence"},
		{name: "other/evil", body: "not listed"},
		{name: "tool-1/link", typ: tar.TypeSymlink, link: "/etc/shadow"},
	})
	return body, []Member{
		{Path: "tool-1/tool", Name: "tool", SHA256: sum([]byte(binary)), Mode: 0o755},
		{Path: "tool-1/", Licences: true},
	}
}

func plainInto(t *testing.T, dst string, called *bool) func(string) error {
	return func(stage string) error {
		*called = true
		if err := os.MkdirAll(dst, 0o755); err != nil {
			return err
		}
		return os.Rename(filepath.Join(stage, "tool"), filepath.Join(dst, "tool"))
	}
}

func TestAnInstallBecomesAMountedImage(t *testing.T) {
	b := scratch(t, true, true, true)
	body, members := release(t, "tool v1\n")
	m, client := serve(t, "1.0", body, members)
	plainCalled := false

	res, err := Install(context.Background(), m, Options{Client: client, Plain: plainInto(t, t.TempDir(), &plainCalled)})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Image || plainCalled {
		t.Fatalf("image %v, plain called %v", res.Image, plainCalled)
	}
	dir := addon.Dir("tool")
	image := filepath.Join(dir, "tool-1.0.sqfs")
	data, err := os.ReadFile(image)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	for _, want := range []string{"== ./tool\ntool v1\n", "== ./licenses/LICENSE\nMIT License\n", "tool m 755 0 0", "licenses m 755 0 0", "licenses/LICENSE m 644 0 0"} {
		if !strings.Contains(got, want) {
			t.Fatalf("the image lacks %q:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"evil", "deeper", "link"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("the image holds %q, which the manifest does not list:\n%s", unwanted, got)
		}
	}
	if side, _ := os.ReadFile(image + ".sha256"); string(side) != sum(data)+"  tool-1.0.sqfs\n" {
		t.Fatalf("sidecar %q", side)
	}
	if cur, ok := Current("tool"); !ok || cur != "tool-1.0.sqfs" {
		t.Fatalf("current %q %v", cur, ok)
	}
	calls := b.log(t)
	if !strings.Contains(calls, "mksquashfs "+filepath.Join(dir, ".fetch", "stage")+" ") || !strings.Contains(calls, "-comp lz4") {
		t.Fatalf("mksquashfs was not run on the staged files with lz4:\n%s", calls)
	}
	if !strings.Contains(calls, "mount -t squashfs -o loop,ro "+image+" "+filepath.Join(dir, "sqfs")) {
		t.Fatalf("the image was not mounted:\n%s", calls)
	}
	if lic, _ := os.ReadFile(filepath.Join(dir, "licenses", "LICENSE")); string(lic) != "MIT License\n" {
		t.Fatalf("licence %q", lic)
	}
	rec, ok := ReadInstalled("tool")
	if !ok || rec.Image != "tool-1.0.sqfs" || rec.Sums["tool"] != sum([]byte("tool v1\n")) || rec.Licence != "MIT" || rec.Source != m.Source {
		t.Fatalf("installed.json %+v", rec)
	}
	if LinkFile("tool", "tool") != "sqfs/tool" || FilesDir("tool") != filepath.Join(dir, "sqfs") {
		t.Fatalf("link file %s, files dir %s", LinkFile("tool", "tool"), FilesDir("tool"))
	}
	if _, err := os.Stat(filepath.Join(dir, ".fetch")); err == nil {
		t.Fatal("the workspace is left")
	}
}

func TestAnUpdateKeepsThePreviousImageAndRollsBack(t *testing.T) {
	b := scratch(t, true, true, true)
	install := func(version, binary string) {
		t.Helper()
		body, members := release(t, binary)
		m, client := serve(t, version, body, members)
		if _, err := Install(context.Background(), m, Options{Client: client}); err != nil {
			t.Fatal(err)
		}
	}
	install("1.0", "v1")
	install("2.0", "v2")
	dir := addon.Dir("tool")
	if cur, _ := Current("tool"); cur != "tool-2.0.sqfs" {
		t.Fatalf("current %s", cur)
	}
	for _, f := range []string{"tool-1.0.sqfs", "tool-1.0.sqfs.sha256", "tool-2.0.sqfs", "installed.previous.json"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Fatalf("%s is gone after the update", f)
		}
	}
	if !strings.Contains(b.log(t), "umount -l "+filepath.Join(dir, "sqfs")) {
		t.Fatalf("the old image was not detached before the new one was mounted:\n%s", b.log(t))
	}

	if err := Rollback("tool"); err != nil {
		t.Fatal(err)
	}
	if cur, _ := Current("tool"); cur != "tool-1.0.sqfs" {
		t.Fatalf("current after rollback %s", cur)
	}
	if rec, _ := ReadInstalled("tool"); rec.Manifest.Version != "1.0" {
		t.Fatalf("installed.json after rollback names %s", rec.Manifest.Version)
	}

	// A third version keeps the one in use and the one it replaces only.
	install("3.0", "v3")
	if _, err := os.Stat(filepath.Join(dir, "tool-2.0.sqfs")); err == nil {
		t.Fatal("an image two updates old is kept")
	}
	if _, err := os.Stat(filepath.Join(dir, "tool-1.0.sqfs")); err != nil {
		t.Fatal("the image the update replaced is gone")
	}
}

// A reinstall of the version in use must not write over the file the loop
// device reads.
func TestAReinstallOfTheSameVersionGetsANewFile(t *testing.T) {
	scratch(t, true, true, true)
	body, members := release(t, "v1")
	m, client := serve(t, "1.0", body, members)
	for range 2 {
		if _, err := Install(context.Background(), m, Options{Client: client}); err != nil {
			t.Fatal(err)
		}
	}
	if cur, _ := Current("tool"); cur != "tool-1.0.2.sqfs" {
		t.Fatalf("current %s", cur)
	}
	if _, err := os.Stat(filepath.Join(addon.Dir("tool"), "tool-1.0.sqfs")); err != nil {
		t.Fatal("the image in use was removed")
	}
}

func TestAMemberWhoseSumIsNotThePinIsRefused(t *testing.T) {
	b := scratch(t, true, true, true)
	body, members := release(t, "tampered")
	members[0].SHA256 = sum([]byte("the real one"))
	m, client := serve(t, "1.0", body, members)

	_, err := Install(context.Background(), m, Options{Client: client})
	if err == nil || !strings.Contains(err.Error(), "sha256 mismatch") {
		t.Fatalf("got %v", err)
	}
	if _, err := os.Stat(addon.Dir("tool")); err == nil {
		t.Fatal("a failed first install leaves the add-on's directory")
	}
	if strings.Contains(b.log(t), "mksquashfs") || strings.Contains(b.log(t), "mount") {
		t.Fatalf("a refused member was packed or mounted:\n%s", b.log(t))
	}
}

func TestADownloadThatIsNotThePinIsRefused(t *testing.T) {
	scratch(t, true, true, true)
	body, members := release(t, "v1")
	m, client := serve(t, "1.0", body, members)
	m.Artifact.SHA256 = sum([]byte("another archive"))

	if _, err := Install(context.Background(), m, Options{Client: client}); !errors.Is(err, upstream.ErrChecksum) {
		t.Fatalf("got %v", err)
	}
}

func TestPathsThatLeaveTheArtifactOrTheStageAreRefused(t *testing.T) {
	for name, members := range map[string][]Member{
		"saved name with a directory": {{Path: "a/tool", Name: "../tool"}},
		"saved name with a slash":     {{Path: "a/tool", Name: "bin/tool"}},
		"member path with ..":         {{Path: "a/../../etc/passwd", Name: "passwd"}},
		"absolute member path":        {{Path: "/etc/passwd", Name: "passwd"}},
		"licences not a directory":    {{Path: "a/LICENSE", Licences: true}},
		"name with a space":           {{Path: "a/tool", Name: "to ol"}},
		"two members, one name":       {{Path: "a/x", Name: "x"}, {Path: "a/y", Name: "x"}},
		"saved as the licences":       {{Path: "a/x", Name: LicencesDir}},
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			archive := filepath.Join(dir, "a.tar.gz")
			_ = os.WriteFile(archive, tarGz(t, []entry{{name: "a/tool", body: "x"}}), 0o644)
			stage := filepath.Join(dir, "stage")
			if _, err := ExtractTarGz(archive, members, stage); err == nil {
				t.Fatal("accepted")
			}
			if entries, _ := os.ReadDir(dir); len(entries) != 1 {
				t.Fatalf("wrote %v", entries)
			}
		})
	}

	// An archive entry that climbs out is never a member, so it is never
	// written, and the member it pretends to be is missing.
	dir := t.TempDir()
	archive := filepath.Join(dir, "a.tar.gz")
	_ = os.WriteFile(archive, tarGz(t, []entry{{name: "a/../tool", body: "x"}, {name: "../a/tool", body: "x"}}), 0o644)
	_, err := ExtractTarGz(archive, []Member{{Path: "a/tool", Name: "tool"}}, filepath.Join(dir, "stage"))
	if err == nil || !strings.Contains(err.Error(), "has no a/tool") {
		t.Fatalf("got %v", err)
	}
	if entries, _ := os.ReadDir(filepath.Join(dir, "stage")); len(entries) != 0 {
		t.Fatalf("wrote %v", entries)
	}
	if _, err := os.Stat(filepath.Join(dir, "tool")); err == nil {
		t.Fatal("wrote beside the stage")
	}
}

func TestAnUnlistedMemberIsNotWritten(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "a.tar.gz")
	_ = os.WriteFile(archive, tarGz(t, []entry{{name: "a/tool", body: "x"}, {name: "a/extra", body: "y"}, {name: "a/LICENSE", body: "l"}}), 0o644)
	stage := filepath.Join(dir, "stage")
	sums, err := ExtractTarGz(archive, []Member{{Path: "a/tool", Name: "tool", SHA256: sum([]byte("x"))}}, stage)
	if err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(stage)
	if len(entries) != 1 || entries[0].Name() != "tool" || len(sums) != 1 {
		t.Fatalf("stage %v, sums %v", entries, sums)
	}
}

func TestWithoutSquashfsTheFilesArePlacedAsBefore(t *testing.T) {
	for name, kernel := range map[string]bool{"no squashfs in the kernel": false, "no mksquashfs on the slot": true} {
		t.Run(name, func(t *testing.T) {
			b := scratch(t, true, kernel, !kernel)
			body, members := release(t, "v1")
			m, client := serve(t, "1.0", body, members)
			dst := addon.Dir("tool")
			plainCalled := false
			res, err := Install(context.Background(), m, Options{Client: client, Plain: plainInto(t, dst, &plainCalled)})
			if err != nil {
				t.Fatal(err)
			}
			if res.Image || !plainCalled {
				t.Fatalf("image %v, plain called %v", res.Image, plainCalled)
			}
			if got, _ := os.ReadFile(filepath.Join(dst, "tool")); string(got) != "v1" {
				t.Fatalf("plain file %q", got)
			}
			if _, ok := Current("tool"); ok || LinkFile("tool", "tool") != "tool" {
				t.Fatal("a plain install names an image")
			}
			if rec, ok := ReadInstalled("tool"); !ok || rec.Image != "" {
				t.Fatalf("installed.json %+v %v", rec, ok)
			}
			if strings.Contains(b.log(t), "mount") {
				t.Fatalf("mounted:\n%s", b.log(t))
			}
		})
	}
}

// Back on a slot without squashfs, a plain install takes the add-on out of
// image form, so S04addons stops trying to mount what it cannot.
func TestAPlainInstallAfterAnImageDropsTheImage(t *testing.T) {
	b := scratch(t, true, true, true)
	body, members := release(t, "v1")
	m, client := serve(t, "1.0", body, members)
	if _, err := Install(context.Background(), m, Options{Client: client}); err != nil {
		t.Fatal(err)
	}
	Filesystems = filepath.Join(b.base, "no-such-file")
	plainCalled := false
	if _, err := Install(context.Background(), m, Options{Client: client, Plain: plainInto(t, addon.Dir("tool"), &plainCalled)}); err != nil {
		t.Fatal(err)
	}
	dir := addon.Dir("tool")
	for _, f := range []string{"current", "tool-1.0.sqfs", "tool-1.0.sqfs.sha256"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
			t.Fatalf("%s is left", f)
		}
	}
	if addon.IsMounted(filepath.Join(dir, "sqfs")) {
		t.Fatal("the image is still mounted")
	}
}

// The whole image is checked against its sidecar before a mount.
func TestACorruptedImageIsNotMounted(t *testing.T) {
	b := scratch(t, true, true, true)
	body, members := release(t, "v1")
	m, client := serve(t, "1.0", body, members)
	if _, err := Install(context.Background(), m, Options{Client: client}); err != nil {
		t.Fatal(err)
	}
	image := filepath.Join(addon.Dir("tool"), "tool-1.0.sqfs")
	f, _ := os.OpenFile(image, os.O_WRONLY|os.O_APPEND, 0)
	_, _ = f.WriteString("x")
	_ = f.Close()
	_ = os.WriteFile(b.calls, nil, 0o644)

	err := Mount("tool")
	if err == nil || !strings.Contains(err.Error(), "tool-1.0.sqfs") {
		t.Fatalf("got %v", err)
	}
	if strings.Contains(b.log(t), "mount -t") {
		t.Fatalf("a corrupted image was mounted:\n%s", b.log(t))
	}

	_ = os.Remove(image + ".sha256")
	if err := Mount("tool"); err == nil {
		t.Fatal("an image with no sidecar was mounted")
	}
}

func TestTooLittleSpaceIsRefusedBeforeTheDownload(t *testing.T) {
	scratch(t, true, true, true)
	body, members := release(t, "v1")
	m, client := serve(t, "1.0", body, members)
	m.Artifact.Size, m.Artifact.Unpacked = 10<<20, 20<<20
	FreeBytes = func(string) (uint64, error) { return 40 << 20, nil }

	// Plain files need 10 + 20 + 2 MiB; an image 20 MiB more.
	_, err := Install(context.Background(), m, Options{Client: client})
	if !errors.Is(err, ErrNoSpace) || !strings.Contains(err.Error(), "needs 52 MiB, 40 MiB are free") {
		t.Fatalf("got %v", err)
	}
	Filesystems = "/no-such-file"
	plainCalled := false
	if _, err := Install(context.Background(), m, Options{Client: client, Plain: plainInto(t, t.TempDir(), &plainCalled)}); err != nil {
		t.Fatalf("plain files fit: %v", err)
	}
}

func TestOffADistributionImageOnlyPlainFiles(t *testing.T) {
	scratch(t, false, true, true)
	body, members := release(t, "v1")
	m, client := serve(t, "1.0", body, members)
	ws := filepath.Join(t.TempDir(), "fetch")
	dst := t.TempDir()
	plainCalled := false
	res, err := Install(context.Background(), m, Options{Client: client, Workspace: ws, Plain: plainInto(t, dst, &plainCalled)})
	if err != nil || res.Image || !plainCalled {
		t.Fatalf("err %v, image %v, plain %v", err, res.Image, plainCalled)
	}
	if _, err := os.Stat(addon.Root); err == nil {
		t.Fatal("wrote to /data off a distribution image")
	}
	if _, err := os.Stat(ws); err == nil {
		t.Fatal("the workspace is left")
	}
}

func TestAnApkIsVerifiedAndOnlyItsMembersTaken(t *testing.T) {
	b := scratch(t, true, true, true)
	apk := filepath.Join(b.base, "apk")
	stub(t, apk, fmt.Sprintf(`echo "apk $*" >> %q
case "$1" in
verify) [ -z "$BAD_SIGNATURE" ] || { echo "$2: UNTRUSTED signature" >&2; exit 1; } ;;
extract) tar -xzf "$5" -C "$4" ;;
esac`, b.calls))
	body := tarGz(t, []entry{{name: "usr/bin/tool", body: "v1"}, {name: "etc/other", body: "x"}, {name: "usr/bin/link", typ: tar.TypeSymlink, link: "/etc/shadow"}})
	m, client := serve(t, "1.0", body, []Member{{Path: "usr/bin/tool", Name: "tool", Mode: 0o755}})
	m.Artifact.Format = FormatApk

	if _, err := Install(context.Background(), m, Options{Client: client, Apk: apk}); err != nil {
		t.Fatal(err)
	}
	calls := b.log(t)
	if !strings.Contains(calls, "apk verify ") || !strings.Contains(calls, "apk extract --no-chown --destination ") {
		t.Fatalf("apk calls:\n%s", calls)
	}
	data, _ := os.ReadFile(filepath.Join(addon.Dir("tool"), "tool-1.0.sqfs"))
	if !strings.Contains(string(data), "== ./tool\nv1") || strings.Contains(string(data), "other") {
		t.Fatalf("image:\n%s", data)
	}

	t.Setenv("BAD_SIGNATURE", "1")
	m.Version = "2.0"
	if _, err := Install(context.Background(), m, Options{Client: client, Apk: apk}); err == nil || !strings.Contains(err.Error(), "exit status 1") {
		t.Fatalf("an unsigned package was accepted: %v", err)
	}

	// A member that is a link in the package is not a file in it.
	t.Setenv("BAD_SIGNATURE", "")
	m.Version = "3.0"
	m.Artifact.Members = []Member{{Path: "usr/bin/link", Name: "tool"}}
	if _, err := Install(context.Background(), m, Options{Client: client, Apk: apk}); err == nil || !strings.Contains(err.Error(), "has no usr/bin/link") {
		t.Fatalf("got %v", err)
	}
}

// The real mksquashfs, where the test host has one: the command line packs
// an LZ4 image with the manifest's modes, owned by root. CI images without
// squashfs-tools skip it; tools in the dist repo run it in a container.
func TestTheRealMksquashfsCommandLine(t *testing.T) {
	if _, err := exec.LookPath("mksquashfs"); err != nil {
		t.Skip("no mksquashfs here")
	}
	if _, err := exec.LookPath("unsquashfs"); err != nil {
		t.Skip("no unsquashfs here")
	}
	saved := Mksquashfs
	t.Cleanup(func() { Mksquashfs = saved })
	Mksquashfs = "mksquashfs"

	ws := t.TempDir()
	stage := filepath.Join(ws, "stage")
	_ = os.MkdirAll(filepath.Join(stage, LicencesDir), 0o700)
	_ = os.WriteFile(filepath.Join(stage, "tool"), []byte("#!/bin/sh\n"), 0o600)
	_ = os.WriteFile(filepath.Join(stage, LicencesDir, "LICENSE"), []byte("MIT\n"), 0o600)
	out := filepath.Join(ws, "tool.sqfs")
	if err := pack(stage, out, ws, []Member{{Path: "a/tool", Name: "tool", Mode: 0o755}, {Path: "a/", Licences: true}}); err != nil {
		t.Fatal(err)
	}
	list, err := exec.Command("unsquashfs", "-lls", out).CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %s", err, list)
	}
	for _, want := range []string{"-rwxr-xr-x root/root", "squashfs-root/tool", "drwxr-xr-x root/root", "-rw-r--r-- root/root"} {
		if !strings.Contains(string(list), want) {
			t.Fatalf("listing lacks %q:\n%s", want, list)
		}
	}
	stat, _ := exec.Command("unsquashfs", "-s", out).CombinedOutput()
	if !strings.Contains(string(stat), "Compression lz4") {
		t.Fatalf("not lz4:\n%s", stat)
	}
	t.Logf("unsquashfs -lls:\n%s", list)
}

// Fetch stands in for the download of an artifact with no pin: its sum is the
// one installed.json records, and its refusal installs nothing.
func TestFetchReplacesTheDownloadAndItsSumIsRecorded(t *testing.T) {
	scratch(t, true, true, true)
	body, members := release(t, "tool v2\n")
	m, client := serve(t, "2.0", body, members)
	m.Artifact.SHA256 = ""

	refuse := func(ctx context.Context, dst string, progress func(done, total int64)) (string, error) {
		return "", errors.New("not the index's package")
	}
	if _, err := Install(context.Background(), m, Options{Client: client, Fetch: refuse}); err == nil || !strings.Contains(err.Error(), "not the index's package") {
		t.Fatalf("got %v", err)
	}
	if _, err := os.Stat(addon.Dir("tool")); err == nil {
		t.Fatal("a refused fetch must leave nothing behind")
	}

	fetch := func(ctx context.Context, dst string, progress func(done, total int64)) (string, error) {
		if err := client.Fetch(ctx, m.Artifact.URL, m.Artifact.Max, dst, progress); err != nil {
			return "", err
		}
		return sum(body), nil
	}
	if _, err := Install(context.Background(), m, Options{Client: client, Fetch: fetch}); err != nil {
		t.Fatal(err)
	}
	rec, ok := ReadInstalled("tool")
	if !ok || rec.Manifest.Artifact.SHA256 != sum(body) || rec.Manifest.Version != "2.0" {
		t.Fatalf("installed.json %+v", rec.Manifest)
	}
}
