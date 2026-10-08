// NanoKVM (ironkvm-dist#72, run sheet trial 66): AES encryption in riscv64
// assembly for the goroot overlay. See goroot-overlay/overlay.sh and
// goroot-overlay/aesgcm_riscv64_gen.py, which writes aes_riscv64.s.
//
// The assembly is the generic code's T-table method (encryptBlockGeneric)
// with T-Head instructions for the table lookups, so it is variable time like
// the generic code it replaces. Decryption and key expansion stay generic:
// SRTP and TLS only ever encrypt with AES (GCM and counter mode).

//go:build riscv64 && !purego

package aes

import "crypto/internal/fips140/subtle"

//go:noescape
func encryptBlockAsm(nr int, xk *uint32, dst, src *byte)

//go:noescape
func ctr32BlocksAsm(nr int, xk *uint32, dst, src *byte, n int, counter *byte)

type block struct {
	blockExpanded
}

func newBlock(c *Block, key []byte) *Block {
	newBlockExpanded(&c.blockExpanded, key)
	return c
}

func encryptBlock(c *Block, dst, src []byte) {
	_, _ = dst[BlockSize-1], src[BlockSize-1]
	encryptBlockAsm(c.rounds, &c.enc[0], &dst[0], &src[0])
}

func decryptBlock(c *Block, dst, src []byte) {
	decryptBlockGeneric(&c.blockExpanded, dst, src)
}

func checkGenericIsExpected() {}

// GCMCounterCrypt encrypts src into out with AES in GCM's counter mode: the
// last four bytes of counter are a big-endian block counter that wraps at
// 2^32 (unlike CTR). counter is the first block's and is left at the next
// one's. out must be at least as long as src; they may be the same slice but
// must not overlap otherwise. It is the gcm package's gcmCounterCryptGeneric.
func GCMCounterCrypt(b *Block, out, src []byte, counter *[BlockSize]byte) {
	n := len(src) / BlockSize
	if n > 0 {
		_ = out[n*BlockSize-1]
		ctr32BlocksAsm(b.rounds, &b.enc[0], &out[0], &src[0], n, &counter[0])
	}
	if rest := len(src) - n*BlockSize; rest > 0 {
		var mask [BlockSize]byte
		encryptBlockAsm(b.rounds, &b.enc[0], &mask[0], &counter[0])
		c := uint32(counter[12])<<24 | uint32(counter[13])<<16 | uint32(counter[14])<<8 | uint32(counter[15])
		c++
		counter[12], counter[13], counter[14], counter[15] = byte(c>>24), byte(c>>16), byte(c>>8), byte(c)
		subtle.XORBytes(out[n*BlockSize:], src[n*BlockSize:], mask[:rest])
	}
}
