package download

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"

	"NanoKVM-Server/service/netboot"
)

func TestImageFilenameFromURL(t *testing.T) {
	name, err := imageFilenameFromURL("https://example.com/isos/debian-13.iso")
	if err != nil {
		t.Fatalf("expected a normal iso url to be accepted: %s", err)
	}

	if name != "debian-13.iso" {
		t.Fatalf("unexpected filename %q", name)
	}
}

func TestImageFilenameFromURLAppliesTheSameRulesAsUploads(t *testing.T) {
	// The upload path validates the name; the remote path used only
	// filepath.Base, so it accepted names the rest of the service rejects.
	for _, raw := range []string{
		"https://example.com/",
		"https://example.com/..",
		"https://example.com/passwd",
		"https://example.com/a%20b.iso",
		"ftp://example.com/x.iso",
		"https:///x.iso",
		"not a url",
	} {
		if name, err := imageFilenameFromURL(raw); err == nil {
			t.Fatalf("expected %q to be rejected, got %q", raw, name)
		}
	}
}

func TestFitsOnDiskLeavesHeadroom(t *testing.T) {
	// Filling the card completely takes the whole device down, not just the
	// download, so a chunk is kept back.
	available := int64(reservedFreeBytes) + 1000

	if !fitsOnDisk(1000, available) {
		t.Fatal("expected a download that fits to be allowed")
	}

	if fitsOnDisk(1001, available) {
		t.Fatal("expected a download that eats the headroom to be refused")
	}
}

func TestFitsOnDiskAllowsAnUnknownSize(t *testing.T) {
	// A server that sends no Content-Length reports -1; the copy is bounded
	// separately, so this must not fail up front.
	if !fitsOnDisk(-1, reservedFreeBytes*4) {
		t.Fatal("expected an unknown size to be allowed through")
	}
}

func TestAvailableBytesReportsFreeSpace(t *testing.T) {
	available, err := availableBytes(t.TempDir())
	if err != nil {
		t.Fatalf("expected the free space to be readable: %s", err)
	}

	if available <= 0 {
		t.Fatalf("expected a positive amount of free space, got %d", available)
	}
}

// The Boot menu button stores netboot.xyz's ISO under its release name, and
// the pinned sum is one the download accepts.
func TestTheBootMenuISOIsAnImageTheDownloaderAccepts(t *testing.T) {
	name, err := imageFilenameFromURL(netboot.BootMenuISOURL)
	if err != nil || name != "netboot.xyz.iso" {
		t.Fatalf("the boot menu is stored as %q (%v)", name, err)
	}

	sum, err := parseSHA256(netboot.BootMenuISOSHA256)
	if err != nil || len(sum) != 32 {
		t.Fatalf("the pinned sum does not parse: %v", err)
	}
}

func TestFileHasSHA256(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "netboot.xyz.iso")
	if err := os.WriteFile(path, []byte("menu"), 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte("menu"))
	other := sha256.Sum256([]byte("other"))

	if !fileHasSHA256(path, sum[:]) {
		t.Error("a file with the pinned checksum should count as present")
	}
	if fileHasSHA256(path, other[:]) {
		t.Error("a file with another checksum must be downloaded again")
	}
	if fileHasSHA256(filepath.Join(dir, "missing.iso"), sum[:]) {
		t.Error("a missing file is not present")
	}
	if fileHasSHA256(dir, sum[:]) {
		t.Error("a directory is not an image")
	}
	if fileHasSHA256(path, nil) {
		t.Error("without a checksum nothing can be trusted as present")
	}
}
