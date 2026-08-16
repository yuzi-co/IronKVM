package application

import (
	"encoding/base64"
	"fmt"
	"testing"
)

const baseURL = "https://cdn.example.com/nanokvm"

// validPackageName satisfies the pattern validateLatest enforces. The pattern is
// stricter than "a plain file name" - it pins the whole shape - so a fixture
// that only avoided path separators would now be rejected for the wrong reason
// and would prove nothing about the case under test.
const validPackageName = "nanokvm_2.0.0.tar.gz"

// validSha512 is a well-formed digest rather than a real one: validateLatest
// checks that it decodes to 64 bytes and nothing here checks the contents.
var validSha512 = base64.StdEncoding.EncodeToString(make([]byte, 64))

func validLatest() Latest {
	return Latest{
		Version: "1.2.3", Name: "nanokvm_1.2.3.tar.gz",
		Sha512: validSha512, LegacySize: 1,
	}
}

func latestJSON(name string) []byte {
	return latestJSONWithSize(name, 100)
}

func latestJSONWithSize(name string, size uint64) []byte {
	return []byte(fmt.Sprintf(
		`{"version":"2.0.0","name":%q,"sha512":%q,"size":%d}`, name, validSha512, size))
}

func TestValidateLatestV1DoesNotInterpretLegacySizeAsBytes(t *testing.T) {
	latest := validLatest()
	latest.LegacySize = 15048 // historic stable manifests used a non-byte value
	if err := validateLatest(&latest); err != nil {
		t.Fatal(err)
	}
	if err := validateDownloadedSize(&latest, 15406125); err != nil {
		t.Fatalf("v1 must not require equality with legacy size: %v", err)
	}
}

func TestValidateLatestV2RequiresExactByteFields(t *testing.T) {
	latest := validLatest()
	latest.ManifestVersion = 2
	latest.SizeBytes = 100
	latest.UnpackedSizeBytes = 200
	if err := validateLatest(&latest); err != nil {
		t.Fatal(err)
	}
	if err := validateDownloadedSize(&latest, 99); err == nil {
		t.Fatal("v2 size mismatch was accepted")
	}
	if err := validateExpandedSize(&latest, 199); err == nil {
		t.Fatal("v2 expanded size mismatch was accepted")
	}
	latest.ManifestVersion = 3
	if err := validateLatest(&latest); err == nil {
		t.Fatal("unknown manifest version was accepted")
	}
}

func TestParseLatestBuildsTheDownloadURL(t *testing.T) {
	latest, err := parseLatest(latestJSON(validPackageName), baseURL)
	if err != nil {
		t.Fatalf("expected a normal package to be accepted: %s", err)
	}

	if latest.Url != baseURL+"/"+validPackageName {
		t.Fatalf("unexpected url %q", latest.Url)
	}

	if latest.Version != "2.0.0" {
		t.Fatalf("unexpected version %q", latest.Version)
	}
}

func TestParseLatestRejectsANameThatIsNotAPlainFile(t *testing.T) {
	// The name is chosen by whoever serves latest.json, and it lands on disk
	// before the checksum has had a chance to reject the package.
	for _, name := range []string{
		"../../etc/init.d/S99evil",
		"/etc/init.d/S99evil",
		"sub/dir/" + validPackageName,
		"..",
		"",
		".hidden",
		validPackageName + "; reboot",
	} {
		if _, err := parseLatest(latestJSON(name), baseURL); err == nil {
			t.Fatalf("expected %q to be rejected", name)
		}
	}
}

func TestParseLatestRejectsMalformedJSON(t *testing.T) {
	if _, err := parseLatest([]byte("not json"), baseURL); err == nil {
		t.Fatal("expected malformed json to be rejected")
	}
}

func TestParseLatestRejectsAnOversizedPackage(t *testing.T) {
	body := latestJSONWithSize(validPackageName, maxPackageSize+1)

	if _, err := parseLatest(body, baseURL); err == nil {
		t.Fatal("expected an oversized package to be rejected before downloading")
	}
}

// A feed that hosts its manifest and its packages on different hosts has to name
// the package outright. GitHub Pages is the right place for a few hundred bytes
// of JSON and GitHub Releases is the right place for a 26 MB tarball, but a
// release asset lives under a per-tag path that no fixed base URL can reach.
func TestParseLatestUsesTheManifestURLWhenPresent(t *testing.T) {
	body := []byte(fmt.Sprintf(
		`{"manifest_version":2,"version":"1.0.0","name":"nanokvm_1.0.0.tar.gz",`+
			`"url":%q,"sha512":%q,"size":100,"size_bytes":100,"unpacked_size_bytes":200}`,
		"https://github.com/yuzi-co/IronKVM/releases/download/v1.0.0/nanokvm_1.0.0.tar.gz",
		validSha512))

	latest, err := parseLatest(body, baseURL)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	want := "https://github.com/yuzi-co/IronKVM/releases/download/v1.0.0/nanokvm_1.0.0.tar.gz"
	if latest.Url != want {
		t.Fatalf("download URL = %q, want %q", latest.Url, want)
	}
}

// Without the field nothing changes for a feed that serves both from one
// directory, which is every feed that exists today, including Sipeed's.
func TestParseLatestJoinsTheBaseURLWhenTheManifestHasNoURL(t *testing.T) {
	body := []byte(fmt.Sprintf(
		`{"manifest_version":2,"version":"1.0.0","name":"nanokvm_1.0.0.tar.gz",`+
			`"sha512":%q,"size":100,"size_bytes":100,"unpacked_size_bytes":200}`, validSha512))

	latest, err := parseLatest(body, baseURL)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if latest.Url != baseURL+"/nanokvm_1.0.0.tar.gz" {
		t.Fatalf("download URL = %q, want the joined base URL", latest.Url)
	}
}

// The package is written to the SD card the device boots from, and the manifest
// is the only thing that says where it comes from. Plain HTTP would let anyone
// on the path replace it, and the sha512 that would catch that comes from the
// same document.
func TestParseLatestRefusesANonHTTPSManifestURL(t *testing.T) {
	for _, raw := range []string{
		"http://example.com/nanokvm_1.0.0.tar.gz",
		"ftp://example.com/nanokvm_1.0.0.tar.gz",
		"/releases/nanokvm_1.0.0.tar.gz",
		"nanokvm_1.0.0.tar.gz",
		"file:///tmp/nanokvm_1.0.0.tar.gz",
		"https:///nanokvm_1.0.0.tar.gz",
	} {
		body := []byte(fmt.Sprintf(
			`{"manifest_version":2,"version":"1.0.0","name":"nanokvm_1.0.0.tar.gz",`+
				`"url":%q,"sha512":%q,"size":100,"size_bytes":100,"unpacked_size_bytes":200}`,
			raw, validSha512))

		if _, err := parseLatest(body, baseURL); err == nil {
			t.Fatalf("expected %q to be refused", raw)
		}
	}
}

// IronKVM ships packages under its own name. An official Sipeed package must
// still install, because being able to return to the official firmware is the
// reason the rename stayed shallow.
func TestValidateLatestAcceptsBothPackagePrefixes(t *testing.T) {
	for _, name := range []string{"ironkvm_1.0.0.tar.gz", "nanokvm_2.4.3.tar.gz"} {
		latest := Latest{Version: "1.0.0", Name: name, Sha512: validSha512, LegacySize: 100}
		if err := validateLatest(&latest); err != nil {
			t.Fatalf("%s should be accepted: %s", name, err)
		}
	}
}

// The name decides both the URL and the file written to the SD card, and it is
// written before the checksum has had a chance to reject the package. It has to
// stay a plain file name of one known shape.
func TestValidateLatestStillRefusesAnUnsafePackageName(t *testing.T) {
	for _, name := range []string{
		"../ironkvm_1.0.0.tar.gz",
		"sub/ironkvm_1.0.0.tar.gz",
		"ironkvm_1.0.0.tar.gz.sh",
		"evilkvm_1.0.0.tar.gz",
		"ironkvm_1.0.tar.gz",
		"ironkvm_1.0.0.zip",
		"IRONKVM_1.0.0.tar.gz",
	} {
		latest := Latest{Version: "1.0.0", Name: name, Sha512: validSha512, LegacySize: 100}
		if err := validateLatest(&latest); err == nil {
			t.Fatalf("expected %q to be refused", name)
		}
	}
}
