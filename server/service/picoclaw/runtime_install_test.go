package picoclaw

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"NanoKVM-Server/service/extensions/addon"
)

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// bundle puts a PicoClaw in a scratch /kvmapp/picoclaw/bin, as the image does,
// with the sha256 file the build writes. checksum "" writes the right one.
func bundle(t *testing.T, content, checksum string) {
	t.Helper()
	dir := t.TempDir()
	savedBin, savedSum := picoclawBundledBinary, picoclawBundledChecksum
	t.Cleanup(func() { picoclawBundledBinary, picoclawBundledChecksum = savedBin, savedSum })
	picoclawBundledBinary = filepath.Join(dir, "picoclaw")
	picoclawBundledChecksum = filepath.Join(dir, "picoclaw.sha256")
	if content == "" {
		return
	}
	if checksum == "" {
		checksum = sha256Hex(content)
	}
	if err := os.WriteFile(picoclawBundledBinary, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	line := checksum + "  picoclaw-linux-riscv64-ironkvm\n"
	if err := os.WriteFile(picoclawBundledChecksum, []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestInstallCopiesTheImagesPicoclawOntoData(t *testing.T) {
	fsroot, _ := scratchImage(t, true)
	bundle(t, "fork build", "")

	if err := installBundledPicoclaw(); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(addon.Dir("picoclaw"), "picoclaw")
	if b, err := os.ReadFile(dest); err != nil || string(b) != "fork build" {
		t.Fatalf("the image's binary is not at %s: %q %v", dest, b, err)
	}
	if b, _ := os.ReadFile(dest + ".sha256"); strings.TrimSpace(string(b)) != sha256Hex("fork build") {
		t.Fatalf("the installed sha256 is not recorded: %q", b)
	}
	if target, err := os.Readlink(fsroot + "/usr/bin/picoclaw"); err != nil || target != dest {
		t.Fatalf("/usr/bin/picoclaw links to %q (%v), want %s", target, err, dest)
	}
	if installedPicoclawIsStale() {
		t.Fatal("a fresh install must not count as stale")
	}
}

func TestAnImageWithoutPicoclawCannotInstallIt(t *testing.T) {
	scratchImage(t, true)
	bundle(t, "", "")

	if _, err := readBundledPicoclawChecksum(); !errors.Is(err, errPicoclawNotBundled) {
		t.Fatalf("got %v, want %v", err, errPicoclawNotBundled)
	}
	if err := installBundledPicoclaw(); !errors.Is(err, errPicoclawNotBundled) {
		t.Fatalf("got %v, want %v", err, errPicoclawNotBundled)
	}
	if installedPicoclawIsStale() {
		t.Fatal("with nothing in the image there is nothing to replace the install with")
	}
}

func TestACorruptBinaryIsNotInstalled(t *testing.T) {
	scratchImage(t, true)
	bundle(t, "fork build", sha256Hex("something else"))

	if err := installBundledPicoclaw(); err == nil || !strings.Contains(err.Error(), "sha256 mismatch") {
		t.Fatalf("got %v, want a sha256 mismatch", err)
	}
	dest := filepath.Join(addon.Dir("picoclaw"), "picoclaw")
	for _, p := range []string{dest, dest + ".tmp", dest + ".sha256"} {
		if _, err := os.Stat(p); err == nil {
			t.Fatalf("%s exists after a failed install", p)
		}
	}
}

func TestAChecksumFileWithoutADigestIsRefused(t *testing.T) {
	scratchImage(t, true)
	bundle(t, "fork build", "not-a-digest")
	if _, err := readBundledPicoclawChecksum(); err == nil || errors.Is(err, errPicoclawNotBundled) {
		t.Fatalf("got %v, want a malformed checksum error", err)
	}
}

// Sipeed's v0.2.8 was installed with no recorded sha256, and an older image's
// fork build with another one. Both are replaced before the next start, and
// the settings in home stay.
func TestAnInstallThatIsNotTheImagesIsReplaced(t *testing.T) {
	scratchImage(t, true)
	dest := filepath.Join(addon.Dir("picoclaw"), "picoclaw")
	_ = os.MkdirAll(filepath.Join(addon.Dir("picoclaw"), "home"), 0o755)
	_ = os.WriteFile(filepath.Join(addon.Dir("picoclaw"), "home", "config.json"), []byte("{}"), 0o600)
	_ = os.WriteFile(dest, []byte("sipeed v0.2.8"), 0o755)

	bundle(t, "fork build", "")
	if !installedPicoclawIsStale() {
		t.Fatal("an install with no recorded sha256 must count as stale")
	}
	refreshInstalledPicoclaw()
	if b, _ := os.ReadFile(dest); string(b) != "fork build" {
		t.Fatalf("the old binary was not replaced: %q", b)
	}

	bundle(t, "newer fork build", "")
	if !installedPicoclawIsStale() {
		t.Fatal("an install of another build must count as stale")
	}
	refreshInstalledPicoclaw()
	if b, _ := os.ReadFile(dest); string(b) != "newer fork build" {
		t.Fatalf("the older build was not replaced: %q", b)
	}
	if b, _ := os.ReadFile(filepath.Join(addon.Dir("picoclaw"), "home", "config.json")); string(b) != "{}" {
		t.Fatalf("the settings changed: %q", b)
	}
}
