package webui

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// The vectors were made by the minisign tool, so these hold the parser and the
// verifier to the real format in both its algorithms.
func TestVerifiesSignaturesMadeByMinisign(t *testing.T) {
	pk, err := ParsePublicKey(readTestdata(t, "vector.pub"))
	if err != nil {
		t.Fatal(err)
	}
	msg := readTestdata(t, "vector.txt")
	for _, name := range []string{"vector.txt.minisig", "vector-legacy.minisig"} {
		sig, err := ParseSignature(readTestdata(t, name))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if err := pk.Verify(bytes.NewReader(msg), sig); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if err := pk.Verify(bytes.NewReader(append(msg, 'x')), sig); !errors.Is(err, ErrAltered) {
			t.Fatalf("%s: a changed message should fail as altered, got %v", name, err)
		}
	}
}

func TestEditedTrustedCommentFails(t *testing.T) {
	pk, _ := ParsePublicKey(readTestdata(t, "vector.pub"))
	edited := strings.Replace(string(readTestdata(t, "vector.txt.minisig")), "test vector", "test vectoR", 1)
	sig, err := ParseSignature([]byte(edited))
	if err != nil {
		t.Fatal(err)
	}
	if err := pk.Verify(bytes.NewReader(readTestdata(t, "vector.txt")), sig); !errors.Is(err, ErrAltered) {
		t.Fatalf("an edited trusted comment should fail, got %v", err)
	}
}

func TestOtherKeyIsUnknown(t *testing.T) {
	sig, _ := ParseSignature(readTestdata(t, "vector.txt.minisig"))
	other := newTestKey(t)
	pk, err := ParsePublicKey(other.pubFile())
	if err != nil {
		t.Fatal(err)
	}
	if err := pk.Verify(bytes.NewReader(readTestdata(t, "vector.txt")), sig); !errors.Is(err, ErrUnknownKey) {
		t.Fatalf("a signature by another key should be unknown, got %v", err)
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	if _, err := ParsePublicKey([]byte("untrusted comment: x\nnot base64\n")); err == nil {
		t.Fatal("garbage parsed as a key")
	}
	if _, err := ParseSignature([]byte("untrusted comment: x\nAAAA\n")); err == nil {
		t.Fatal("garbage parsed as a signature")
	}
	// A CRLF key, as a Windows editor leaves it, is still the same key.
	crlf := strings.ReplaceAll(string(readTestdata(t, "vector.pub")), "\n", "\r\n")
	if _, err := ParsePublicKey([]byte(crlf)); err != nil {
		t.Fatalf("a CRLF key should parse: %v", err)
	}
}

func policyEnv(t *testing.T) (Policy, *testKey, string) {
	t.Helper()
	dir := t.TempDir()
	key := newTestKey(t)
	key.writePub(t, filepath.Join(dir, "image-keys"), "release.pub")
	p := Policy{
		KeyDirs:  []string{filepath.Join(dir, "image-keys"), filepath.Join(dir, "owner-keys")},
		CheckOff: filepath.Join(dir, "no-signature-check"),
	}
	file := filepath.Join(dir, "pkg.tar.gz")
	if err := os.WriteFile(file, []byte("package bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p, key, file
}

func TestPolicyAcceptsAKeyFromEitherDirectory(t *testing.T) {
	p, key, file := policyEnv(t)
	key.signFile(t, file)
	v, err := p.VerifyFile(file, file+".minisig")
	if err != nil || !v.Checked || v.Signer != "release.pub" {
		t.Fatalf("image key: verdict %+v, err %v", v, err)
	}

	owner := newTestKey(t)
	owner.writePub(t, p.KeyDirs[1], "mine.pub")
	owner.signFile(t, file)
	v, err = p.VerifyFile(file, file+".minisig")
	if err != nil || v.Signer != "mine.pub" {
		t.Fatalf("owner key: verdict %+v, err %v", v, err)
	}
}

func TestPolicyRefusals(t *testing.T) {
	p, key, file := policyEnv(t)

	if _, err := p.VerifyFile(file, file+".minisig"); !errors.Is(err, ErrNoSignature) {
		t.Fatalf("missing signature: %v", err)
	}

	stranger := newTestKey(t)
	stranger.signFile(t, file)
	if _, err := p.VerifyFile(file, file+".minisig"); !errors.Is(err, ErrUnknownKey) {
		t.Fatalf("unknown signer: %v", err)
	}

	key.signFile(t, file)
	if err := os.WriteFile(file, []byte("package bytez"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := p.VerifyFile(file, file+".minisig"); !errors.Is(err, ErrAltered) {
		t.Fatalf("altered package: %v", err)
	}
}

// The owner's two ways out, as generic/slot has them: the switch file, and an
// empty keyring. Each passes with a warning and Checked false.
func TestPolicyOwnerCanTurnTheCheckOff(t *testing.T) {
	p, _, file := policyEnv(t)
	if err := os.WriteFile(p.CheckOff, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	v, err := p.VerifyFile(file, file+".minisig")
	if err != nil || v.Checked || !strings.Contains(v.Warning, "no-signature-check") {
		t.Fatalf("switch: verdict %+v, err %v", v, err)
	}
	if p.State() != "off" {
		t.Fatalf("state %q, want off", p.State())
	}

	empty := Policy{KeyDirs: []string{t.TempDir()}}
	v, err = empty.VerifyFile(file, file+".minisig")
	if err != nil || v.Checked || !strings.Contains(v.Warning, "keyring is empty") {
		t.Fatalf("empty keyring: verdict %+v, err %v", v, err)
	}
	if empty.State() != "no-keys" {
		t.Fatalf("state %q, want no-keys", empty.State())
	}
}

func TestPolicySkipsAMalformedKey(t *testing.T) {
	p, key, file := policyEnv(t)
	if err := os.WriteFile(filepath.Join(p.KeyDirs[0], "broken.pub"), []byte("nonsense"), 0o644); err != nil {
		t.Fatal(err)
	}
	key.signFile(t, file)
	if v, err := p.VerifyFile(file, file+".minisig"); err != nil || !v.Checked {
		t.Fatalf("a broken key stopped a good one: %+v %v", v, err)
	}
}

func TestDefaultPolicyMatchesTheSlotInstaller(t *testing.T) {
	t.Setenv("IRONKVM_KEYS_DIRS", "")
	os.Unsetenv("IRONKVM_KEYS_DIRS")
	os.Unsetenv("IRONKVM_SIGNATURE_CHECK_OFF")
	p := DefaultPolicy()
	if strings.Join(p.KeyDirs, " ") != "/usr/share/ironkvm/keys /data/ironkvm/keys" ||
		p.CheckOff != "/data/ironkvm/no-signature-check" {
		t.Fatalf("default policy %+v", p)
	}
	t.Setenv("IRONKVM_KEYS_DIRS", "/tmp/a /tmp/b")
	if got := DefaultPolicy().KeyDirs; len(got) != 2 || got[1] != "/tmp/b" {
		t.Fatalf("override: %v", got)
	}
}
