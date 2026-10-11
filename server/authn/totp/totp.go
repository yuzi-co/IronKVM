// Package totp is the time-based one-time password of RFC 6238, on the HOTP
// of RFC 4226, in the one shape every authenticator app supports: HMAC-SHA1,
// six digits, a 30 second step.
//
// It is written here rather than taken from github.com/pquerna/otp because
// the algorithm is thirty lines, and that module also brings an image and a
// barcode encoder. The QR code is drawn by the web page from the otpauth URI.
package totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"net/url"
	"strings"
	"time"
)

const (
	// Digits is the length of a code.
	Digits = 6
	// Period is the length of one time step.
	Period = 30 * time.Second
	// Skew is how many steps either side of now a code is still accepted
	// for, so a phone clock that is off by up to 30 seconds still works.
	Skew = 1
	// SecretSize is the size of a generated secret: 160 bits, the length of
	// the HMAC-SHA1 output that RFC 4226 recommends.
	SecretSize = 20
)

var encoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// NewSecret returns a random secret.
func NewSecret() ([]byte, error) {
	secret := make([]byte, SecretSize)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	return secret, nil
}

// EncodeSecret is the base32 form an authenticator app takes.
func EncodeSecret(secret []byte) string {
	return encoding.EncodeToString(secret)
}

// DecodeSecret reads the base32 form back. Case, spaces and padding are
// ignored, as apps that show the secret group it differently.
func DecodeSecret(text string) ([]byte, error) {
	text = strings.ToUpper(strings.NewReplacer(" ", "", "-", "", "=", "").Replace(text))
	secret, err := encoding.DecodeString(text)
	if err != nil || len(secret) == 0 {
		return nil, errors.New("invalid secret")
	}
	return secret, nil
}

// Counter is the time step t falls in.
func Counter(t time.Time) uint64 {
	return uint64(t.Unix()) / uint64(Period/time.Second)
}

// HOTP is RFC 4226's code for one counter value, with the hash and the number
// of digits as parameters so the RFC's own test vectors can be checked.
func HOTP(secret []byte, counter uint64, digits int, newHash func() hash.Hash) string {
	var message [8]byte
	binary.BigEndian.PutUint64(message[:], counter)
	mac := hmac.New(newHash, secret)
	mac.Write(message[:])
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	value := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff

	modulus := uint32(1)
	for i := 0; i < digits; i++ {
		modulus *= 10
	}
	return fmt.Sprintf("%0*d", digits, value%modulus)
}

// Code is the code an authenticator shows at t.
func Code(secret []byte, t time.Time) string {
	return HOTP(secret, Counter(t), Digits, sha1.New)
}

// Verify checks a code against the steps within Skew of t, and returns the
// step it matched. A step at or below last is refused even when the code is
// right: that is a code already used, and accepting it again is a replay.
// last is zero when no code has been accepted yet.
func Verify(secret []byte, code string, t time.Time, last uint64) (uint64, bool) {
	code = NormalizeCode(code)
	if len(code) != Digits {
		return 0, false
	}
	now := Counter(t)
	matched := uint64(0)
	found := false
	// Every candidate is computed and compared, whether or not an earlier one
	// matched, so the time taken says nothing about which step it was.
	for delta := -Skew; delta <= Skew; delta++ {
		if delta < 0 && now < uint64(-delta) {
			continue
		}
		counter := uint64(int64(now) + int64(delta))
		candidate := HOTP(secret, counter, Digits, sha1.New)
		if subtle.ConstantTimeCompare([]byte(candidate), []byte(code)) == 1 && counter > last {
			if !found || counter > matched {
				matched = counter
			}
			found = true
		}
	}
	return matched, found
}

// NormalizeCode drops the spaces and dashes people type inside a code.
func NormalizeCode(code string) string {
	return strings.NewReplacer(" ", "", "-", "").Replace(strings.TrimSpace(code))
}

// IsNumericCode reports whether code has the shape of a TOTP code rather than
// a recovery code.
func IsNumericCode(code string) bool {
	code = NormalizeCode(code)
	if len(code) != Digits {
		return false
	}
	for _, r := range code {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// URI is the otpauth URI the QR code carries, in the Key Uri Format that
// Google Authenticator defined and every other app follows.
func URI(issuer, account string, secret []byte) string {
	label := url.PathEscape(issuer) + ":" + url.PathEscape(account)
	query := url.Values{}
	query.Set("secret", EncodeSecret(secret))
	query.Set("issuer", issuer)
	query.Set("algorithm", "SHA1")
	query.Set("digits", fmt.Sprint(Digits))
	query.Set("period", fmt.Sprint(int(Period/time.Second)))
	return "otpauth://totp/" + label + "?" + query.Encode()
}

// RecoveryCodeCount is how many recovery codes an enrolment hands out.
const RecoveryCodeCount = 10

// recoveryAlphabet is Crockford's base32 without the letters people misread.
const recoveryAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// NewRecoveryCodes returns RecoveryCodeCount codes of 16 characters each, 80
// bits, written as four groups of four.
//
// The entropy is what lets the store keep a plain SHA-256 of each instead of
// a bcrypt hash. A bcrypt check costs about a second on the board's C906, and
// a login with a recovery code would pay it once per stored code. Eighty
// random bits cannot be searched from a hash the way a chosen password can.
func NewRecoveryCodes() ([]string, error) {
	codes := make([]string, RecoveryCodeCount)
	raw := make([]byte, 16)
	for i := range codes {
		if _, err := rand.Read(raw); err != nil {
			return nil, err
		}
		var b strings.Builder
		for j, v := range raw {
			if j > 0 && j%4 == 0 {
				b.WriteByte('-')
			}
			b.WriteByte(recoveryAlphabet[v&0x1f])
		}
		codes[i] = b.String()
	}
	return codes, nil
}

// HashRecoveryCode is what the store keeps for a recovery code. Case, spaces
// and dashes do not matter, and neither do the letters people confuse with
// digits.
func HashRecoveryCode(code string) string {
	normalized := strings.ToUpper(NormalizeCode(code))
	normalized = strings.NewReplacer("O", "0", "I", "1", "L", "1").Replace(normalized)
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
