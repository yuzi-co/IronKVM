// NanoKVM (ironkvm-dist#72, run sheet trial 66): AES-GCM for riscv64 in the
// goroot overlay. See goroot-overlay/overlay.sh and
// goroot-overlay/aesgcm_riscv64_gen.py, which writes gcm_riscv64.s.
//
// The generic GCM (gcm_generic.go) encrypts H again and rebuilds a 16-entry
// table of its multiples on every Seal and Open, then multiplies four bits at
// a time. Here the table is 256 multiples of H (Shoup's 8-bit method), built
// once per key in initGCM, and the multiplication and the counter mode are
// assembly. Variable time, like the generic code it replaces.

//go:build riscv64 && !purego

package gcm

import (
	"crypto/internal/fips140/aes"
	"crypto/internal/fips140/subtle"
	"crypto/internal/fips140deps/byteorder"
)

// ghashTable8 holds b*H for every byte b, in the bit order of the generic
// code's field elements: the byte's top bit is the coefficient of x^0. lo and
// hi are the gcmFieldElement low and high words; the assembly finds hi 2048
// bytes after lo.
type ghashTable8 struct {
	lo, hi [256]uint64
}

// ghashReduce8[b] reduces the eight coefficients x^128 to x^135 that leave
// the high word when an element is multiplied by x^8 (b is the high word's
// low byte, so its bit k is x^(127-k) before the shift): the remainder,
// already in place in the top 15 bits of the low word.
var ghashReduce8 = func() (t [256]uint64) {
	for b := range t {
		for k := 0; k < 8; k++ {
			if b&(1<<k) != 0 {
				t[b] ^= 0xe100000000000000 >> (7 - k)
			}
		}
	}
	return
}()

//go:noescape
func ghashBlocksAsm(t *ghashTable8, y *gcmFieldElement, data *byte, n int)

func checkGenericIsExpected() {}

type gcmPlatformData struct {
	table ghashTable8
}

func initGCM(g *GCM) {
	var h [gcmBlockSize]byte
	aes.EncryptBlockInternal(&g.cipher, h[:], h[:])
	t := &g.table
	x := gcmFieldElement{byteorder.BEUint64(h[:8]), byteorder.BEUint64(h[8:])}
	// The top bit of a byte is x^0, so 0x80 is H and each lower bit is the
	// one above it times x.
	for b := 0x80; b > 0; b >>= 1 {
		t.lo[b], t.hi[b] = x.low, x.high
		x = ghashDouble(&x)
	}
	for b := 2; b < 256; b <<= 1 {
		for j := 1; j < b; j++ {
			t.lo[b|j] = t.lo[b] ^ t.lo[j]
			t.hi[b|j] = t.hi[b] ^ t.hi[j]
		}
	}
}

// ghashUpdate8 extends y with data, zero-padding a partial last block.
func ghashUpdate8(t *ghashTable8, y *gcmFieldElement, data []byte) {
	if n := len(data) / gcmBlockSize; n > 0 {
		ghashBlocksAsm(t, y, &data[0], n)
		data = data[n*gcmBlockSize:]
	}
	if len(data) > 0 {
		var partial [gcmBlockSize]byte
		copy(partial[:], data)
		ghashBlocksAsm(t, y, &partial[0], 1)
	}
}

// gcmAuth8 is gcmAuthGeneric with the per-key table.
func gcmAuth8(out []byte, g *GCM, tagMask *[gcmBlockSize]byte, ciphertext, additionalData []byte) {
	var y gcmFieldElement
	ghashUpdate8(&g.table, &y, additionalData)
	ghashUpdate8(&g.table, &y, ciphertext)
	var lenBlock [gcmBlockSize]byte
	byteorder.BEPutUint64(lenBlock[:8], uint64(len(additionalData))*8)
	byteorder.BEPutUint64(lenBlock[8:], uint64(len(ciphertext))*8)
	ghashBlocksAsm(&g.table, &y, &lenBlock[0], 1)
	var s [gcmBlockSize]byte
	byteorder.BEPutUint64(s[:8], y.low)
	byteorder.BEPutUint64(s[8:], y.high)
	subtle.XORBytes(out, s[:], tagMask[:])
}

// counter8 is deriveCounterGeneric followed by the tag mask, as both seal
// and open start.
func counter8(g *GCM, counter, tagMask *[gcmBlockSize]byte, nonce []byte) {
	if len(nonce) == gcmStandardNonceSize {
		copy(counter[:], nonce)
		counter[gcmBlockSize-1] = 1
	} else {
		var h [gcmBlockSize]byte
		aes.EncryptBlockInternal(&g.cipher, h[:], h[:])
		deriveCounterGeneric(&h, counter, nonce)
	}
	aes.EncryptBlockInternal(&g.cipher, tagMask[:], counter[:])
	gcmInc32(counter)
}

func seal(out []byte, g *GCM, nonce, plaintext, data []byte) {
	var counter, tagMask [gcmBlockSize]byte
	counter8(g, &counter, &tagMask, nonce)
	aes.GCMCounterCrypt(&g.cipher, out, plaintext, &counter)
	var tag [gcmTagSize]byte
	gcmAuth8(tag[:], g, &tagMask, out[:len(plaintext)], data)
	copy(out[len(plaintext):], tag[:])
}

func open(out []byte, g *GCM, nonce, ciphertext, data []byte) error {
	var counter, tagMask [gcmBlockSize]byte
	counter8(g, &counter, &tagMask, nonce)
	tag := ciphertext[len(ciphertext)-g.tagSize:]
	ciphertext = ciphertext[:len(ciphertext)-g.tagSize]
	var expectedTag [gcmTagSize]byte
	gcmAuth8(expectedTag[:], g, &tagMask, ciphertext, data)
	if subtle.ConstantTimeCompare(expectedTag[:g.tagSize], tag) != 1 {
		return errOpen
	}
	aes.GCMCounterCrypt(&g.cipher, out, ciphertext, &counter)
	return nil
}
