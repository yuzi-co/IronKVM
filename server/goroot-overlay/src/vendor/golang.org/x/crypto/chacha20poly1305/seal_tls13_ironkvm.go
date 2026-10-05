// NanoKVM addition to the toolchain's vendored x/crypto, applied through
// server/goroot-overlay (ironkvm-dist#68). Not part of x/crypto.

package chacha20poly1305

import (
	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/internal/alias"
	"golang.org/x/crypto/internal/poly1305"
)

// SealTLS13 is Seal of plaintext followed by the one byte contentType, the
// inner plaintext of a TLS 1.3 record, with the two not next to each other in
// memory. The result is byte for byte what Seal gives for the joined input.
//
// crypto/tls (with the overlay's gather_ironkvm.go) uses it to seal a record
// from the caller's buffer into the one the socket is written from. Seal
// needs the inner plaintext in one piece, so crypto/tls copies every payload
// next to its header first and seals it there.
func (c *chacha20poly1305) SealTLS13(dst, nonce, plaintext []byte, contentType byte, additionalData []byte) []byte {
	if len(nonce) != NonceSize {
		panic("chacha20poly1305: bad nonce length passed to Seal")
	}
	if uint64(len(plaintext)) > (1<<38)-65 {
		panic("chacha20poly1305: plaintext too large")
	}

	n := len(plaintext) + 1
	ret, out := sliceForAppend(dst, n+poly1305.TagSize)
	ciphertext, tag := out[:n], out[n:]
	if alias.InexactOverlap(out, plaintext) {
		panic("chacha20poly1305: invalid buffer overlap")
	}

	var polyKey [32]byte
	s, _ := chacha20.NewUnauthenticatedCipher(c.key[:], nonce)
	s.XORKeyStream(polyKey[:], polyKey[:])
	s.SetCounter(1) // set the counter to 1, skipping 32 bytes
	s.XORKeyStream(ciphertext[:n-1], plaintext)
	ciphertext[n-1] = contentType
	s.XORKeyStream(ciphertext[n-1:], ciphertext[n-1:])

	p := poly1305.New(&polyKey)
	writeWithPadding(p, additionalData)
	writeWithPadding(p, ciphertext)
	writeUint64(p, len(additionalData))
	writeUint64(p, n)
	p.Sum(tag[:0])

	return ret
}
