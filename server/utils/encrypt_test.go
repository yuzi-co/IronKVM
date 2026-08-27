package utils

import (
	"encoding/base64"
	"net/url"
	"testing"
)

// The web UI sends encodeURIComponent(CryptoJS.AES.encrypt(...)), so every
// ciphertext below was produced by the library this package used before, and
// the round trip proves the wire format did not move.
var wireFixtures = map[string]string{
	"root":    "U2FsdGVkX1/ZV15WSO06EgR2jLY9Kq7w4qrrfwB0rTQ=",
	"hunter2": "U2FsdGVkX19LCQKm5jVy20O1/wVFzuh5eBF/9UOAWMI=",
	"":        "U2FsdGVkX1/EM0z0ef6531opUP4m9m7PIweP+b4vT0w=",
	"a-longer-password-that-spans-two-blocks": "U2FsdGVkX1+8+ckMEkEls+rbRlrp17IcUkENPnlwtwSGcs/o9WOWVJSbXer+W3Ks+rwJMbsfxkOkTqXxKDgkqw==",
}

func TestDecryptReadsTheWireFormat(t *testing.T) {
	for plaintext, ciphertext := range wireFixtures {
		got, err := Decrypt(ciphertext)
		if err != nil {
			t.Fatalf("Decrypt(%q) returned %v, want no error", ciphertext, err)
		}
		if got != plaintext {
			t.Fatalf("Decrypt(%q) = %q, want %q", ciphertext, got, plaintext)
		}
	}
}

func TestDecryptTakesAnEmptyStringAsNoPassword(t *testing.T) {
	got, err := Decrypt("")
	if err != nil || got != "" {
		t.Fatalf("Decrypt(\"\") = %q, %v, want \"\", nil", got, err)
	}
}

// Every case here made aes256.Decrypt panic. The login route runs this on a
// request body before any credential is checked, so an anonymous caller
// reached the panic, and gin turned it into a 500 rather than an answer.
func TestDecryptRefusesMalformedCiphertextInsteadOfPanicking(t *testing.T) {
	salted := func(body []byte) string {
		return base64.StdEncoding.EncodeToString(append([]byte("Salted__12345678"), body...))
	}

	cases := map[string]string{
		"header alone, no cipher block":          salted(nil),
		"cipher block cut short":                 salted([]byte("abc")),
		"whole block of garbage":                 salted([]byte("0123456789abcdef")),
		"padding byte larger than the plaintext": "U2FsdGVkX18AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"not base64 at all":                      "not base64 at all",
		"too short to hold a salt":               base64.StdEncoding.EncodeToString([]byte("Salted__")),
		"right length, wrong magic":              base64.StdEncoding.EncodeToString([]byte("Peppered1234567890abcdef01234567")),
	}

	for name, ciphertext := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := Decrypt(ciphertext)
			if err == nil {
				t.Fatalf("Decrypt(%q) = %q with no error, want a refusal", ciphertext, got)
			}
			if got != "" {
				t.Fatalf("Decrypt(%q) returned %q beside its error, want an empty string", ciphertext, got)
			}
		})
	}
}

// The web UI percent-encodes, and an API client that posts a form does not,
// because gin has already decoded it. Both have to arrive at the same password.
func TestDecodeDecryptTakesBothRepresentations(t *testing.T) {
	ciphertext := wireFixtures["hunter2"]

	for name, data := range map[string]string{
		"raw base64":      ciphertext,
		"percent-encoded": url.QueryEscape(ciphertext),
	} {
		t.Run(name, func(t *testing.T) {
			got, err := DecodeDecrypt(data)
			if err != nil {
				t.Fatalf("DecodeDecrypt(%q) returned %v, want no error", data, err)
			}
			if got != "hunter2" {
				t.Fatalf("DecodeDecrypt(%q) = %q, want %q", data, got, "hunter2")
			}
		})
	}
}

// A base64 '+' becomes a space if the value is unescaped a second time, so a
// raw ciphertext that carries one must not take the percent-decoding path.
func TestDecodeDecryptDoesNotUnescapeRawBase64(t *testing.T) {
	ciphertext := wireFixtures["a-longer-password-that-spans-two-blocks"]
	if !containsPlus(ciphertext) {
		t.Fatalf("fixture %q carries no '+', so it cannot prove anything", ciphertext)
	}

	got, err := DecodeDecrypt(ciphertext)
	if err != nil {
		t.Fatalf("DecodeDecrypt returned %v, want no error", err)
	}
	if got != "a-longer-password-that-spans-two-blocks" {
		t.Fatalf("DecodeDecrypt = %q, want the whole password", got)
	}
}

func TestDecodeDecryptRefusesMalformedInput(t *testing.T) {
	if _, err := DecodeDecrypt("U2FsdGVkX18AAAAAAAAAAA=="); err == nil {
		t.Fatal("DecodeDecrypt returned no error on a truncated ciphertext, want a refusal")
	}
}

func containsPlus(s string) bool {
	for _, r := range s {
		if r == '+' {
			return true
		}
	}
	return false
}
