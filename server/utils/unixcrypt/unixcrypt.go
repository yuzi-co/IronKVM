// Package unixcrypt checks a password against a crypt(3) hash of the kinds a
// board's /etc/shadow holds: MD5-crypt ($1$), SHA-256-crypt ($5$) and
// SHA-512-crypt ($6$).
//
// It exists for one question, whether root still has a factory password, so it
// verifies and does not generate hashes for storage. Nothing in the module
// already implements these schemes, and the three are short enough to carry.
// The SHA schemes follow Ulrich Drepper's specification ("Unix crypt using
// SHA-256 and SHA-512"); MD5-crypt follows Poul-Henning Kamp's original.
package unixcrypt

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"errors"
	"hash"
	"strconv"
	"strings"
)

// ErrUnsupported is returned for a hash of a scheme this package does not
// implement, such as yescrypt ($y$) or bcrypt ($2b$).
var ErrUnsupported = errors.New("unsupported crypt scheme")

// Verify reports whether password produces hash. It returns ErrUnsupported
// when the hash is of a scheme it cannot compute, so the caller can tell "no"
// from "cannot say".
func Verify(password, hashed string) (bool, error) {
	computed, err := Crypt(password, hashed)
	if err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare([]byte(computed), []byte(hashed)) == 1, nil
}

// Crypt computes the hash of password with the scheme, rounds and salt that
// setting names. setting may be a full hash; everything after the salt is
// ignored, as crypt(3) does.
func Crypt(password, setting string) (string, error) {
	switch {
	case strings.HasPrefix(setting, "$1$"):
		return md5Crypt([]byte(password), setting[3:]), nil
	case strings.HasPrefix(setting, "$5$"):
		return shaCrypt(sha256.New, "$5$", sha256Order, []byte(password), setting[3:]), nil
	case strings.HasPrefix(setting, "$6$"):
		return shaCrypt(sha512.New, "$6$", sha512Order, []byte(password), setting[3:]), nil
	}
	return "", ErrUnsupported
}

const alphabet = "./0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// encode24 appends n characters for the three bytes b2, b1, b0, least
// significant six bits first.
func encode24(out []byte, b2, b1, b0 byte, n int) []byte {
	w := uint(b2)<<16 | uint(b1)<<8 | uint(b0)
	for ; n > 0; n-- {
		out = append(out, alphabet[w&0x3f])
		w >>= 6
	}
	return out
}

// The byte order each scheme feeds encode24, as index triples. The last group
// of each is the short one.
var (
	md5Order = [][3]int{{0, 6, 12}, {1, 7, 13}, {2, 8, 14}, {3, 9, 15}, {4, 10, 5}}

	sha256Order = [][3]int{
		{0, 10, 20}, {21, 1, 11}, {12, 22, 2}, {3, 13, 23}, {24, 4, 14},
		{15, 25, 5}, {6, 16, 26}, {27, 7, 17}, {18, 28, 8}, {9, 19, 29},
	}

	sha512Order = [][3]int{
		{0, 21, 42}, {22, 43, 1}, {44, 2, 23}, {3, 24, 45}, {25, 46, 4},
		{47, 5, 26}, {6, 27, 48}, {28, 49, 7}, {50, 8, 29}, {9, 30, 51},
		{31, 52, 10}, {53, 11, 32}, {12, 33, 54}, {34, 55, 13}, {56, 14, 35},
		{15, 36, 57}, {37, 58, 16}, {59, 17, 38}, {18, 39, 60}, {40, 61, 19},
		{62, 20, 41},
	}
)

// saltOf returns the salt from the part of a setting after the magic: up to
// the next '$', at most max bytes.
func saltOf(rest string, max int) string {
	if i := strings.IndexByte(rest, '$'); i >= 0 {
		rest = rest[:i]
	}
	if len(rest) > max {
		rest = rest[:max]
	}
	return rest
}

func md5Crypt(password []byte, rest string) string {
	salt := []byte(saltOf(rest, 8))
	magic := []byte("$1$")

	alt := md5.New()
	alt.Write(password)
	alt.Write(salt)
	alt.Write(password)
	final := alt.Sum(nil)

	ctx := md5.New()
	ctx.Write(password)
	ctx.Write(magic)
	ctx.Write(salt)
	for n := len(password); n > 0; n -= 16 {
		ctx.Write(final[:min(n, 16)])
	}
	for n := len(password); n > 0; n >>= 1 {
		if n&1 != 0 {
			ctx.Write([]byte{0})
		} else {
			ctx.Write(password[:1])
		}
	}
	final = ctx.Sum(nil)

	for i := 0; i < 1000; i++ {
		round := md5.New()
		if i&1 != 0 {
			round.Write(password)
		} else {
			round.Write(final)
		}
		if i%3 != 0 {
			round.Write(salt)
		}
		if i%7 != 0 {
			round.Write(password)
		}
		if i&1 != 0 {
			round.Write(final)
		} else {
			round.Write(password)
		}
		final = round.Sum(nil)
	}

	out := append([]byte("$1$"), salt...)
	out = append(out, '$')
	for _, g := range md5Order {
		out = encode24(out, final[g[0]], final[g[1]], final[g[2]], 4)
	}
	out = encode24(out, 0, 0, final[11], 2)
	return string(out)
}

const (
	defaultRounds = 5000
	minRounds     = 1000
	maxRounds     = 999999999
)

func shaCrypt(newHash func() hash.Hash, magic string, order [][3]int, password []byte, rest string) string {
	rounds := defaultRounds
	customRounds := false
	if strings.HasPrefix(rest, "rounds=") {
		if end := strings.IndexByte(rest, '$'); end > 0 {
			if n, err := strconv.ParseUint(rest[len("rounds="):end], 10, 64); err == nil {
				rounds = int(max(min(n, maxRounds), minRounds))
				customRounds = true
				rest = rest[end+1:]
			}
		}
	}
	salt := []byte(saltOf(rest, 16))

	b := newHash()
	b.Write(password)
	b.Write(salt)
	b.Write(password)
	digestB := b.Sum(nil)
	size := len(digestB)

	a := newHash()
	a.Write(password)
	a.Write(salt)
	n := len(password)
	for ; n > size; n -= size {
		a.Write(digestB)
	}
	a.Write(digestB[:n])
	for n = len(password); n > 0; n >>= 1 {
		if n&1 != 0 {
			a.Write(digestB)
		} else {
			a.Write(password)
		}
	}
	digestA := a.Sum(nil)

	dp := newHash()
	for range password {
		dp.Write(password)
	}
	p := repeat(dp.Sum(nil), len(password))

	ds := newHash()
	for i := 0; i < 16+int(digestA[0]); i++ {
		ds.Write(salt)
	}
	s := repeat(ds.Sum(nil), len(salt))

	c := digestA
	for i := 0; i < rounds; i++ {
		round := newHash()
		if i&1 != 0 {
			round.Write(p)
		} else {
			round.Write(c)
		}
		if i%3 != 0 {
			round.Write(s)
		}
		if i%7 != 0 {
			round.Write(p)
		}
		if i&1 != 0 {
			round.Write(c)
		} else {
			round.Write(p)
		}
		c = round.Sum(nil)
	}

	out := []byte(magic)
	if customRounds {
		out = append(out, "rounds="+strconv.Itoa(rounds)+"$"...)
	}
	out = append(out, salt...)
	out = append(out, '$')
	for _, g := range order {
		out = encode24(out, c[g[0]], c[g[1]], c[g[2]], 4)
	}
	if size == sha256.Size {
		out = encode24(out, 0, c[31], c[30], 3)
	} else {
		out = encode24(out, 0, 0, c[63], 2)
	}
	return string(out)
}

// repeat returns digest repeated to exactly n bytes.
func repeat(digest []byte, n int) []byte {
	out := make([]byte, 0, n)
	for len(out)+len(digest) <= n {
		out = append(out, digest...)
	}
	return append(out, digest[:n-len(out)]...)
}
