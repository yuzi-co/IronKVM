package ventoy

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sum(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// serveRelease serves a release archive holding members, pins it and them,
// and stands cat in for xzcat, so an "xz" member is stored as it is.
func serveRelease(t *testing.T, members map[string][]byte, extra map[string][]byte) {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, data := range mergeMaps(members, extra) {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(data)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	archive := buf.Bytes()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archive)
	}))
	t.Cleanup(srv.Close)

	oldArchive, oldMembers, oldDir, oldXz := releaseArchive, releaseMembers, Dir, XzcatPath
	t.Cleanup(func() {
		releaseArchive, releaseMembers, Dir, XzcatPath = oldArchive, oldMembers, oldDir, oldXz
	})
	Dir = t.TempDir()
	XzcatPath = "cat"
	releaseArchive = pinnedFile{URL: srv.URL, SHA256: sum(archive), Max: 1 << 20}
	releaseMembers = nil
	for _, m := range oldMembers {
		data := members[m.Member]
		releaseMembers = append(releaseMembers, releaseMember{Member: m.Member, Name: m.Name, XZ: m.XZ, SHA256: sum(data)})
	}
}

func mergeMaps(a, b map[string][]byte) map[string][]byte {
	out := map[string][]byte{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

func releaseContent() map[string][]byte {
	content := map[string][]byte{}
	for _, m := range releaseMembers {
		content[m.Member] = []byte("content of " + m.Name)
	}
	return content
}

func TestInstallKeepsTheThreeFiles(t *testing.T) {
	content := releaseContent()
	serveRelease(t, content, map[string][]byte{"./ventoy-1.1.17/../../etc/evil": []byte("x")})

	if err := install(); err != nil {
		t.Fatal(err)
	}
	if !installed() {
		t.Fatal("not installed after install")
	}
	entries, err := os.ReadDir(Dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("the directory holds %d entries, want the three files", len(entries))
	}
	for _, m := range releaseMembers {
		data, err := os.ReadFile(filepath.Join(Dir, m.Name))
		if err != nil || !strings.HasPrefix(string(data), "content of ") {
			t.Fatalf("%s: %q %v", m.Name, data, err)
		}
	}

	if err := uninstall(); err != nil {
		t.Fatal(err)
	}
	if installed() {
		t.Fatal("installed after uninstall")
	}
}

func TestInstallRefusesAChangedArchive(t *testing.T) {
	serveRelease(t, releaseContent(), nil)
	releaseArchive.SHA256 = strings.Repeat("0", 64)

	if err := install(); err == nil || !strings.Contains(err.Error(), "sha256") {
		t.Fatalf("got %v", err)
	}
	if installed() {
		t.Fatal("a file of a changed archive was kept")
	}
}

func TestInstallRefusesAChangedMember(t *testing.T) {
	serveRelease(t, releaseContent(), nil)
	releaseMembers[1].SHA256 = strings.Repeat("0", 64)

	if err := install(); err == nil || !strings.Contains(err.Error(), "sha256") {
		t.Fatalf("got %v", err)
	}
	if installed() {
		t.Fatal("a changed member was kept")
	}
	if _, err := os.Stat(filepath.Join(Dir, releaseMembers[0].Name)); err == nil {
		t.Fatal("a member was moved into place before all were checked")
	}
}

func TestInstallRefusesAMissingMember(t *testing.T) {
	content := releaseContent()
	delete(content, releaseMembers[2].Member)
	serveRelease(t, content, nil)

	if err := install(); err == nil || !strings.Contains(err.Error(), "has no") {
		t.Fatalf("got %v", err)
	}
}
