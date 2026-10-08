// NanoKVM (ironkvm-dist#72): the riscv64 GCM against the generic code it
// replaces (sealGeneric and openGeneric, still compiled on every
// architecture), on random keys, nonces, plaintexts and additional data. See
// goroot-overlay/overlay.sh.

//go:build riscv64 && !purego

package gcm

import (
	"bytes"
	"crypto/internal/fips140/aes"
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

func TestSealOpenMatchGeneric(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 66))
	sizes := []int{0, 1, 12, 15, 16, 17, 20, 31, 32, 33, 64, 100, 200, 1180, 1200, 1216, 4096 + 7}
	for _, kl := range []int{16, 24, 32} {
		for _, nonceSize := range []int{12, 8, 16} {
			for _, tagSize := range []int{16, 12} {
				for _, pn := range sizes {
					for _, an := range []int{0, 8, 12, 13, 16, 20, 33} {
						blk, err := aes.New(randBytes(r, kl))
						if err != nil {
							t.Fatal(err)
						}
						g, err := New(blk, nonceSize, tagSize)
						if err != nil {
							t.Fatal(err)
						}
						nonce := randBytes(r, nonceSize)
						if r.IntN(4) == 0 && nonceSize == 12 {
							// The 32-bit counter wraps inside the message.
							nonce[8], nonce[9], nonce[10], nonce[11] = 0xff, 0xff, 0xff, 0xff
						}
						pt := randBytes(r, pn)
						ad := randBytes(r, an)
						got := make([]byte, pn+tagSize)
						want := make([]byte, pn+tagSize)
						seal(got, g, nonce, pt, ad)
						sealGeneric(want, g, nonce, pt, ad)
						if !bytes.Equal(got, want) {
							t.Fatalf("seal key %d nonce %d tag %d pt %d ad %d:\n asm     %x\n generic %x",
								kl, nonceSize, tagSize, pn, an, got, want)
						}
						dec := make([]byte, pn)
						if err := open(dec, g, nonce, got, ad); err != nil || !bytes.Equal(dec, pt) {
							t.Fatalf("open key %d pt %d ad %d: %v", kl, pn, an, err)
						}
						// A changed bit anywhere fails, as in the generic code.
						bad := append([]byte(nil), got...)
						bad[r.IntN(len(bad))] ^= 1 << r.IntN(8)
						if err := open(dec, g, nonce, bad, ad); err == nil {
							t.Fatalf("open accepted a changed ciphertext (pt %d ad %d)", pn, an)
						}
						if err := openGeneric(dec, g, nonce, bad, ad); err == nil {
							t.Fatalf("generic open accepted a changed ciphertext")
						}
						if an > 0 {
							ad2 := append([]byte(nil), ad...)
							ad2[r.IntN(an)] ^= 0x80
							if err := open(dec, g, nonce, got, ad2); err == nil {
								t.Fatalf("open accepted changed additional data")
							}
						}
					}
				}
			}
		}
	}
}

// Through the exported API, in place, as pion srtp calls it: Seal into the
// plaintext's own buffer after the header.
func TestSealInPlace(t *testing.T) {
	r := rand.New(rand.NewPCG(4, 66))
	for i := 0; i < 500; i++ {
		blk, _ := aes.New(randBytes(r, 16))
		g, _ := New(blk, 12, 16)
		nonce := randBytes(r, 12)
		hdr := randBytes(r, 12+r.IntN(20))
		pt := randBytes(r, r.IntN(1300))
		buf := make([]byte, len(hdr)+len(pt), len(hdr)+len(pt)+16)
		copy(buf, hdr)
		copy(buf[len(hdr):], pt)
		want := make([]byte, len(pt)+16)
		sealGeneric(want, g, nonce, pt, hdr)
		out := g.Seal(buf[len(hdr):len(hdr)], nonce, buf[len(hdr):], buf[:len(hdr)])
		if !bytes.Equal(out, want) {
			t.Fatalf("in place seal: mismatch at %d bytes", len(pt))
		}
	}
}

func BenchmarkSeal1200(b *testing.B) {
	blk, _ := aes.New(make([]byte, 16))
	g, _ := New(blk, 12, 16)
	nonce := make([]byte, 12)
	ad := make([]byte, 20)
	pt := make([]byte, 1200)
	out := make([]byte, 1216)
	b.SetBytes(1200)
	for i := 0; i < b.N; i++ {
		seal(out, g, nonce, pt, ad)
	}
}

func BenchmarkSealGeneric1200(b *testing.B) {
	blk, _ := aes.New(make([]byte, 16))
	g, _ := New(blk, 12, 16)
	nonce := make([]byte, 12)
	ad := make([]byte, 20)
	pt := make([]byte, 1200)
	out := make([]byte, 1216)
	b.SetBytes(1200)
	for i := 0; i < b.N; i++ {
		sealGeneric(out, g, nonce, pt, ad)
	}
}

func BenchmarkGHASH1200(b *testing.B) {
	blk, _ := aes.New(make([]byte, 16))
	g, _ := New(blk, 12, 16)
	data := make([]byte, 1200)
	var y gcmFieldElement
	b.SetBytes(1200)
	for i := 0; i < b.N; i++ {
		ghashUpdate8(&g.table, &y, data)
	}
}
