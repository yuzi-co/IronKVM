// NanoKVM (ironkvm-dist#72): the riscv64 AES assembly against the generic
// code it replaces, on random keys and inputs. See goroot-overlay/overlay.sh.

//go:build riscv64 && !purego

package aes

import (
	"bytes"
	"crypto/internal/fips140deps/byteorder"
	"math/rand/v2"
	"testing"
)

func randBytes(r *rand.Rand, n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(r.Uint32())
	}
	return b
}

func TestEncryptBlockAsmMatchesGeneric(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 66))
	for _, kl := range []int{16, 24, 32} {
		for i := 0; i < 2000; i++ {
			c, err := New(randBytes(r, kl))
			if err != nil {
				t.Fatal(err)
			}
			src := randBytes(r, BlockSize)
			got := make([]byte, BlockSize)
			want := make([]byte, BlockSize)
			encryptBlock(c, got, src)
			encryptBlockGeneric(&c.blockExpanded, want, src)
			if !bytes.Equal(got, want) {
				t.Fatalf("key %d bytes, block %x: asm %x, generic %x", kl, src, got, want)
			}
			// In place.
			encryptBlock(c, src, src)
			if !bytes.Equal(src, want) {
				t.Fatalf("in place: %x, want %x", src, want)
			}
		}
	}
}

func gcmCounterCryptRef(c *Block, out, src []byte, counter *[BlockSize]byte) {
	var mask [BlockSize]byte
	for len(src) > 0 {
		encryptBlockGeneric(&c.blockExpanded, mask[:], counter[:])
		byteorder.BEPutUint32(counter[12:], byteorder.BEUint32(counter[12:])+1)
		n := min(len(src), BlockSize)
		for i := 0; i < n; i++ {
			out[i] = src[i] ^ mask[i]
		}
		out, src = out[n:], src[n:]
	}
}

func TestGCMCounterCryptMatchesGeneric(t *testing.T) {
	r := rand.New(rand.NewPCG(2, 66))
	for _, kl := range []int{16, 24, 32} {
		for _, n := range []int{0, 1, 15, 16, 17, 31, 32, 33, 100, 1200, 1216, 4000} {
			for i := 0; i < 20; i++ {
				c, _ := New(randBytes(r, kl))
				var ctr0 [BlockSize]byte
				copy(ctr0[:], randBytes(r, BlockSize))
				if i%4 == 0 {
					// Wrap the 32-bit counter inside the message.
					byteorder.BEPutUint32(ctr0[12:], 0xffffffff-uint32(r.IntN(3)))
				}
				src := randBytes(r, n)
				ctrA, ctrB := ctr0, ctr0
				got := make([]byte, n)
				want := make([]byte, n)
				GCMCounterCrypt(c, got, src, &ctrA)
				gcmCounterCryptRef(c, want, src, &ctrB)
				if !bytes.Equal(got, want) || ctrA != ctrB {
					t.Fatalf("key %d, %d bytes, counter %x: asm %x (next %x), want %x (next %x)",
						kl, n, ctr0, got, ctrA, want, ctrB)
				}
				// In place.
				ctrA = ctr0
				GCMCounterCrypt(c, src, src, &ctrA)
				if !bytes.Equal(src, want) {
					t.Fatalf("in place, %d bytes: mismatch", n)
				}
			}
		}
	}
}

// FIPS 197 appendix C.1 to C.3.
func TestEncryptBlockFIPS197(t *testing.T) {
	pt := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	for _, tc := range []struct {
		kl   int
		want string
	}{
		{16, "69c4e0d86a7b0430d8cdb78070b4c55a"},
		{24, "dda97ca4864cdfe06eaf70a0ec0d7191"},
		{32, "8ea2b7ca516745bfeafc49904b496089"},
	} {
		c, _ := New(key[:tc.kl])
		out := make([]byte, 16)
		c.Encrypt(out, pt)
		if got := hexString(out); got != tc.want {
			t.Errorf("AES-%d: %s, want %s", tc.kl*8, got, tc.want)
		}
	}
}

func hexString(b []byte) string {
	const h = "0123456789abcdef"
	s := make([]byte, 0, 2*len(b))
	for _, x := range b {
		s = append(s, h[x>>4], h[x&15])
	}
	return string(s)
}

func BenchmarkEncryptBlockAsm(b *testing.B) {
	c, _ := New(make([]byte, 16))
	var blk [BlockSize]byte
	b.SetBytes(BlockSize)
	for i := 0; i < b.N; i++ {
		encryptBlock(c, blk[:], blk[:])
	}
}

func BenchmarkEncryptBlockGeneric(b *testing.B) {
	c, _ := New(make([]byte, 16))
	var blk [BlockSize]byte
	b.SetBytes(BlockSize)
	for i := 0; i < b.N; i++ {
		encryptBlockGeneric(&c.blockExpanded, blk[:], blk[:])
	}
}

func BenchmarkGCMCounterCrypt1200(b *testing.B) {
	c, _ := New(make([]byte, 16))
	var ctr [BlockSize]byte
	buf := make([]byte, 1200)
	b.SetBytes(1200)
	for i := 0; i < b.N; i++ {
		GCMCounterCrypt(c, buf, buf, &ctr)
	}
}
