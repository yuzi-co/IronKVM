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
