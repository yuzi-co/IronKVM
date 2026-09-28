//go:build linux

package netboot

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
	"strings"
	"sync/atomic"
	"testing"

	"NanoKVM-Server/config"
	"NanoKVM-Server/service/extensions/addon"
)

// scratchImage points the add-on package and this package's paths at a
// temporary root. onDistro decides whether it looks like a distribution image
// with /data mounted.
func scratchImage(t *testing.T, onDistro bool) (base string) {
	t.Helper()
	base = t.TempDir()
	data := filepath.Join(base, "data")
	for _, d := range []string{data, filepath.Join(base, "kvmapp")} {
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

	saved := struct{ root, dataDir, marker, mounts, pkg, conf, run, usb string }{
		addon.Root, addon.DataDir, addon.DistroMarker, addon.Mounts, addon.PkgInitdDir, ConfDir, RunDir, USBScript}
	t.Cleanup(func() {
		addon.Root, addon.DataDir, addon.DistroMarker, addon.Mounts, addon.PkgInitdDir = saved.root, saved.dataDir, saved.marker, saved.mounts, saved.pkg
		ConfDir, RunDir, USBScript = saved.conf, saved.run, saved.usb
	})
	addon.Root = filepath.Join(data, "ironkvm", "addons")
	addon.DataDir = data
	addon.DistroMarker = marker
	addon.Mounts = mounts
	addon.PkgInitdDir = filepath.Join(base, "kvmapp")
	ConfDir = filepath.Join(base, "etc-kvm", "netboot")
	RunDir = filepath.Join(base, "run")
	USBScript = filepath.Join(base, "init.d", "S03usbdev")
	return base
}

// fakeInstall puts a complete add-on in place without a download.
func fakeInstall(t *testing.T) {
	t.Helper()
	if err := os.MkdirAll(tftpRoot(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dnsmasqPath(), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range bootFiles() {
		if err := os.WriteFile(filepath.Join(tftpRoot(), name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := addon.Record(addonSpec()); err != nil {
		t.Fatal(err)
	}
}

// recordScripts replaces runScript with one that records "<script base> <args>"
// and answers what fail says for that line.
func recordScripts(t *testing.T, fail map[string]error) *[]string {
	t.Helper()
	var calls []string
	saved := runScript
	t.Cleanup(func() { runScript = saved })
	runScript = func(script string, args ...string) error {
		line := strings.TrimSpace(filepath.Base(script) + " " + strings.Join(args, " "))
		calls = append(calls, line)
		return fail[line]
	}
	return &calls
}

// memSettings keeps the settings in memory, as server.yaml would.
type memSettings struct {
	current config.NetBoot
	saves   int
	fail    error
}

func (m *memSettings) get() config.NetBoot { return m.current }

func (m *memSettings) save(s config.NetBoot) error {
	if m.fail != nil {
		return m.fail
	}
	m.current = s
	m.saves++
	return nil
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

// fakeApk puts a stub in place of apk-tools 3. It answers update, fetch,
// verify and extract, builds the package it fetches with tar, and appends
// every call to the file it returns. BAD_SIGNATURE in the environment makes
// verify refuse the package, as apk does a tampered one.
func fakeApk(t *testing.T, version string) (calls string) {
	t.Helper()
	dir := t.TempDir()
	calls = filepath.Join(dir, "calls")
	src := filepath.Join(dir, "src")
	path := filepath.Join(dir, "apk")
	stub(t, path, fmt.Sprintf(`echo "$*" >> %q
case "$1" in
update) echo "OK: 25418 distinct packages available" ;;
fetch)
	mkdir -p %[3]q/usr/sbin
	printf 'dnsmasq %[2]s\n' > %[3]q/usr/sbin/dnsmasq
	tar -czf "$3/dnsmasq-%[2]s.apk" -C %[3]q usr
	echo "Downloading dnsmasq-%[2]s" ;;
verify) [ -z "$BAD_SIGNATURE" ] || { echo "$2: UNTRUSTED signature" >&2; exit 1; } ;;
extract) tar -xzf "$5" -C "$4" ;;
*) echo "unexpected: $*" >&2; exit 2 ;;
esac`, calls, version, src))

	saved := ApkPath
	t.Cleanup(func() { ApkPath = saved })
	ApkPath = path
	return calls
}

func sum(content []byte) string {
	h := sha256.Sum256(content)
	return hex.EncodeToString(h[:])
}

type tarEntry struct {
	name     string
	content  string
	linkname string
}

func tarGz(t *testing.T, entries []tarEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		hdr := &tar.Header{Name: e.name, Mode: 0o644, Size: int64(len(e.content)), Typeflag: tar.TypeReg}
		if e.linkname != "" {
			hdr = &tar.Header{Name: e.name, Linkname: e.linkname, Typeflag: tar.TypeSymlink}
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if e.linkname == "" {
			if _, err := tw.Write([]byte(e.content)); err != nil {
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

// fakeReleases serves an iPXE archive and the netboot.xyz binaries from a
// test server and points the pins at them. It counts the requests.
func fakeReleases(t *testing.T) *atomic.Int32 {
	t.Helper()

	var entries []tarEntry
	var members []archiveMember
	for _, m := range ipxeMembers {
		content := "ipxe " + m.Name
		entries = append(entries, tarEntry{name: m.Member, content: content})
		members = append(members, archiveMember{Member: m.Member, Name: m.Name, SHA256: sum([]byte(content))})
	}
	entries = append(entries,
		tarEntry{name: "ipxeboot/undionly.kpxe", linkname: "x86_64/undionly.kpxe"},
		tarEntry{name: "ipxeboot/riscv64/ipxe.efi", content: "not taken"},
	)
	archive := tarGz(t, entries)

	files := map[string][]byte{"/ipxeboot.tar.gz": archive}
	var xyz []pinnedFile
	for _, f := range netbootXYZFiles {
		content := []byte("netboot.xyz " + f.Name)
		files["/"+f.Name] = content
		xyz = append(xyz, pinnedFile{Name: f.Name, SHA256: sum(content)})
	}

	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		content, ok := files[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(content)
	}))
	t.Cleanup(srv.Close)

	for i := range xyz {
		xyz[i].URL = srv.URL + "/" + xyz[i].Name
	}

	savedArchive, savedMembers, savedXYZ, savedClient := ipxeArchive, ipxeMembers, netbootXYZFiles, httpClient
	t.Cleanup(func() {
		ipxeArchive, ipxeMembers, netbootXYZFiles, httpClient = savedArchive, savedMembers, savedXYZ, savedClient
	})
	ipxeArchive = pinnedFile{Name: "ipxeboot.tar.gz", URL: srv.URL + "/ipxeboot.tar.gz", SHA256: sum(archive)}
	ipxeMembers = members
	netbootXYZFiles = xyz
	httpClient = srv.Client()

	return &requests
}
