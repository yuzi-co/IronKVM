package totp

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"net/url"
	"strings"
	"testing"
	"time"
)

// RFC 4226 appendix D: the six-digit HOTP values of its 20-byte secret.
func TestHOTPRFC4226Vectors(t *testing.T) {
	secret := []byte("12345678901234567890")
	want := []string{"755224", "287082", "359152", "969429", "338314",
		"254676", "287922", "162583", "399871", "520489"}
	for counter, code := range want {
		if got := HOTP(secret, uint64(counter), 6, sha1.New); got != code {
			t.Errorf("counter %d: got %s, want %s", counter, got, code)
		}
	}
}

// RFC 6238 appendix B: eight-digit TOTP values at the listed times, for each
// of the three hashes, with the 30 second step.
func TestTOTPRFC6238Vectors(t *testing.T) {
	seeds := map[string]struct {
		secret  []byte
		newHash func() hash.Hash
	}{
		"SHA1":   {[]byte("12345678901234567890"), sha1.New},
		"SHA256": {[]byte("12345678901234567890123456789012"), sha256.New},
		"SHA512": {[]byte("1234567890123456789012345678901234567890123456789012345678901234"), sha512.New},
	}
	vectors := []struct {
		unix int64
		mode string
		code string
	}{
		{59, "SHA1", "94287082"},
		{59, "SHA256", "46119246"},
		{59, "SHA512", "90693936"},
		{1111111109, "SHA1", "07081804"},
		{1111111109, "SHA256", "68084774"},
		{1111111109, "SHA512", "25091201"},
		{1111111111, "SHA1", "14050471"},
		{1111111111, "SHA256", "67062674"},
		{1111111111, "SHA512", "99943326"},
		{1234567890, "SHA1", "89005924"},
		{1234567890, "SHA256", "91819424"},
		{1234567890, "SHA512", "93441116"},
		{2000000000, "SHA1", "69279037"},
		{2000000000, "SHA256", "90698825"},
		{2000000000, "SHA512", "38618901"},
		{20000000000, "SHA1", "65353130"},
		{20000000000, "SHA256", "77737706"},
		{20000000000, "SHA512", "47863826"},
	}
	for _, v := range vectors {
		seed := seeds[v.mode]
		got := HOTP(seed.secret, Counter(time.Unix(v.unix, 0)), 8, seed.newHash)
		if got != v.code {
			t.Errorf("%s at %d: got %s, want %s", v.mode, v.unix, got, v.code)
		}
	}
}

// The six-digit code is the last six digits of the eight-digit one, which is
// what links Code to the RFC's table.
func TestCodeIsSixDigitTruncation(t *testing.T) {
	secret := []byte("12345678901234567890")
	if got := Code(secret, time.Unix(1111111109, 0)); got != "081804" {
		t.Fatalf("got %s, want 081804", got)
	}
}

func TestVerifyAcceptsOneStepEitherSide(t *testing.T) {
	secret := []byte("12345678901234567890")
	now := time.Unix(1111111109, 0)
	for _, offset := range []time.Duration{-Period, 0, Period} {
		code := Code(secret, now.Add(offset))
		counter, ok := Verify(secret, code, now, 0)
		if !ok {
			t.Fatalf("offset %s: refused", offset)
		}
		if counter != Counter(now.Add(offset)) {
			t.Fatalf("offset %s: matched step %d, want %d", offset, counter, Counter(now.Add(offset)))
		}
	}
}

func TestVerifyRefusesTwoStepsAway(t *testing.T) {
	secret := []byte("12345678901234567890")
	now := time.Unix(1111111109, 0)
	for _, offset := range []time.Duration{-2 * Period, 2 * Period} {
		if _, ok := Verify(secret, Code(secret, now.Add(offset)), now, 0); ok {
			t.Fatalf("offset %s: accepted", offset)
		}
	}
}

// A code that was accepted cannot be used again while it is still inside the
// window, and neither can an older one.
func TestVerifyRefusesReplay(t *testing.T) {
	secret := []byte("12345678901234567890")
	now := time.Unix(1111111109, 0)
	code := Code(secret, now)
	counter, ok := Verify(secret, code, now, 0)
	if !ok {
		t.Fatal("first use refused")
	}
	if _, ok := Verify(secret, code, now, counter); ok {
		t.Fatal("same code accepted twice")
	}
	if _, ok := Verify(secret, code, now.Add(Period), counter); ok {
		t.Fatal("same code accepted again one step later")
	}
	if _, ok := Verify(secret, Code(secret, now.Add(-Period)), now, counter); ok {
		t.Fatal("an older code accepted after a newer one")
	}
	// The next step's code still works.
	if _, ok := Verify(secret, Code(secret, now.Add(Period)), now.Add(Period), counter); !ok {
		t.Fatal("next code refused")
	}
}

func TestVerifyRefusesMalformed(t *testing.T) {
	secret := []byte("12345678901234567890")
	now := time.Unix(1111111109, 0)
	for _, code := range []string{"", "12345", "1234567", "abcdef", "08180"} {
		if _, ok := Verify(secret, code, now, 0); ok {
			t.Fatalf("%q accepted", code)
		}
	}
	if _, ok := Verify(secret, "081 804", now, 0); !ok {
		t.Fatal("a code typed with a space refused")
	}
}

func TestSecretRoundTrip(t *testing.T) {
	secret, err := NewSecret()
	if err != nil {
		t.Fatal(err)
	}
	if len(secret) != SecretSize {
		t.Fatalf("secret is %d bytes", len(secret))
	}
	encoded := EncodeSecret(secret)
	if strings.Contains(encoded, "=") {
		t.Fatalf("padding in %s", encoded)
	}
	decoded, err := DecodeSecret(strings.ToLower(encoded[:8]) + " " + encoded[8:])
	if err != nil || string(decoded) != string(secret) {
		t.Fatalf("round trip failed: %v", err)
	}
	if _, err := DecodeSecret("not base32!"); err == nil {
		t.Fatal("garbage decoded")
	}
}

func TestURI(t *testing.T) {
	uri := URI("IronKVM", "alice", []byte("12345678901234567890"))
	parsed, err := url.Parse(uri)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Scheme != "otpauth" || parsed.Host != "totp" || parsed.Path != "/IronKVM:alice" {
		t.Fatalf("unexpected uri %s", uri)
	}
	query := parsed.Query()
	if query.Get("secret") != "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ" || query.Get("issuer") != "IronKVM" ||
		query.Get("digits") != "6" || query.Get("period") != "30" || query.Get("algorithm") != "SHA1" {
		t.Fatalf("unexpected query %v", query)
	}
}

func TestRecoveryCodes(t *testing.T) {
	codes, err := NewRecoveryCodes()
	if err != nil {
		t.Fatal(err)
	}
	if len(codes) != RecoveryCodeCount {
		t.Fatalf("%d codes", len(codes))
	}
	seen := map[string]bool{}
	for _, code := range codes {
		if len(code) != 19 || strings.Count(code, "-") != 3 {
			t.Fatalf("unexpected shape %q", code)
		}
		if seen[code] {
			t.Fatalf("duplicate %q", code)
		}
		seen[code] = true
		if IsNumericCode(code) {
			t.Fatalf("%q looks like a TOTP code", code)
		}
	}
	code := codes[0]
	typed := strings.ToLower(strings.ReplaceAll(code, "-", " "))
	if HashRecoveryCode(code) != HashRecoveryCode(typed) {
		t.Fatal("a code typed in lower case with spaces hashes differently")
	}
	if HashRecoveryCode(codes[0]) == HashRecoveryCode(codes[1]) {
		t.Fatal("two codes hash the same")
	}
	if HashRecoveryCode("ABCD-0000") != HashRecoveryCode("abcd-oooo") {
		t.Fatal("O and 0 should be the same character")
	}
}
