package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"NanoKVM-Server/service/netboot"
)

func scratchBootMenu(t *testing.T) {
	t.Helper()
	oldDir, oldRecord, oldClient := imageDir, bootMenuRecordPath, imageClient
	t.Cleanup(func() { imageDir, bootMenuRecordPath, imageClient = oldDir, oldRecord, oldClient })
	imageDir = t.TempDir()
	bootMenuRecordPath = filepath.Join(t.TempDir(), "ironkvm", "bootmenu.json")
}

func sumOf(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

func TestTheBootMenuIsThePinnedReleaseUntilAnUpdate(t *testing.T) {
	scratchBootMenu(t)

	if got := currentBootMenu(); got != pinnedBootMenu() {
		t.Fatalf("without a record the button downloads %+v", got)
	}
	if _, ok := installedBootMenu(); ok {
		t.Fatal("a missing ISO counts as installed")
	}

	// A record that does not name the ISO over https is ignored.
	for _, bad := range []bootMenu{
		{Version: "9.9.9", URL: "http://example.com/netboot.xyz.iso", SHA256: sumOf("x")},
		{Version: "9.9.9", URL: "https://example.com/other.iso", SHA256: sumOf("x")},
		{Version: "9.9.9", URL: "https://example.com/netboot.xyz.iso", SHA256: "nope"},
	} {
		if err := writeBootMenuRecord(bad); err != nil {
			t.Fatal(err)
		}
		if got := currentBootMenu(); got != pinnedBootMenu() {
			t.Fatalf("record %+v was trusted", bad)
		}
	}

	rec := bootMenu{Version: "9.9.9", URL: "https://example.com/dl/netboot.xyz.iso", SHA256: sumOf("new iso")}
	if err := writeBootMenuRecord(rec); err != nil {
		t.Fatal(err)
	}
	if got := currentBootMenu(); got != rec {
		t.Fatalf("with a record the button downloads %+v", got)
	}

	// The present check follows the record: the new release's file is
	// present, another file under the name is not.
	if err := os.WriteFile(bootMenuPath(), []byte("new iso"), 0o644); err != nil {
		t.Fatal(err)
	}
	if v, ok := installedBootMenu(); !ok || v != "9.9.9" {
		t.Fatalf("installed %q %v", v, ok)
	}
	if err := os.WriteFile(bootMenuPath(), []byte("something else"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := installedBootMenu(); ok {
		t.Fatal("an unknown file counts as installed")
	}
}

func serveISO(t *testing.T, content string) string {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(content))
	}))
	t.Cleanup(srv.Close)
	imageClient = srv.Client()
	return srv.URL + "/dl/" + netboot.BootMenuISOName
}

func TestUpdateBootMenuReplacesTheISOAndRecordsIt(t *testing.T) {
	scratchBootMenu(t)
	if err := os.WriteFile(bootMenuPath(), []byte("old iso"), 0o644); err != nil {
		t.Fatal(err)
	}
	url := serveISO(t, "new iso")
	s := NewService()

	next := bootMenu{Version: "9.9.9", URL: url, SHA256: sumOf("new iso")}
	if err := s.updateBootMenu(context.Background(), next, func(int) {}); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(bootMenuPath()); string(data) != "new iso" {
		t.Fatalf("the ISO holds %q", data)
	}
	if rec, ok := readBootMenuRecord(); !ok || rec != next {
		t.Fatalf("record %+v", rec)
	}
	if s.downloadStatus != downloadStatusSuccess {
		t.Fatalf("the downloader shows %s", s.downloadStatus)
	}
}

func TestAnUpdateWithTheWrongChecksumChangesNothing(t *testing.T) {
	scratchBootMenu(t)
	if err := os.WriteFile(bootMenuPath(), []byte("old iso"), 0o644); err != nil {
		t.Fatal(err)
	}
	url := serveISO(t, "tampered iso")
	s := NewService()

	err := s.updateBootMenu(context.Background(), bootMenu{Version: "9.9.9", URL: url, SHA256: sumOf("new iso")}, func(int) {})
	if err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("update: %v", err)
	}
	if data, _ := os.ReadFile(bootMenuPath()); string(data) != "old iso" {
		t.Fatalf("the ISO holds %q", data)
	}
	if _, ok := readBootMenuRecord(); ok {
		t.Fatal("a failed update left a record")
	}
	entries, _ := os.ReadDir(imageDir)
	if len(entries) != 1 {
		t.Fatalf("the image directory holds %d entries", len(entries))
	}
}
