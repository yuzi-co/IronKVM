package chacha20poly1305

import (
	"bytes"
	"crypto/rand"
	"testing"
)

// SealTLS13 must give exactly Seal's output for the joined inner plaintext,
// for every length around the 64-byte ChaCha20 block and 16-byte Poly1305
// block boundaries, at every alignment of source and destination, and the
// result must open.
func TestIronKVMSealTLS13MatchesSeal(t *testing.T) {
	key := make([]byte, KeySize)
	nonce := make([]byte, NonceSize)
	rand.Read(key)
	a, err := New(key)
	if err != nil {
		t.Fatal(err)
	}
	s := a.(*chacha20poly1305)

	src := make([]byte, 16384+64)
	rand.Read(src)
	lengths := []int{0, 1, 2, 15, 16, 17, 31, 32, 33, 62, 63, 64, 65, 127, 128, 129, 1000, 4095, 4096, 16383, 16384}
	for i := 0; i < 200; i++ {
		var b [2]byte
		rand.Read(b[:])
		lengths = append(lengths, int(b[0])<<6|int(b[1])>>2)
	}
	ad := []byte{23, 3, 3, 0, 0}
	cases := 0
	for _, n := range lengths {
		if n > 16384 {
			n = 16384
		}
		for off := 0; off < 8; off++ {
			for doff := 0; doff < 8; doff += 3 {
				rand.Read(nonce)
				ad[3], ad[4] = byte((n+17)>>8), byte(n+17)
				pt := src[off : off+n]
				var cb [1]byte
				rand.Read(cb[:])
				ct := cb[0]
				joined := append(append([]byte{}, pt...), ct)
				want := a.Seal(nil, nonce, joined, ad)

				dst := make([]byte, doff, doff+n+1+16)
				got := s.SealTLS13(dst, nonce, pt, ct, ad)[doff:]
				if !bytes.Equal(got, want) {
					t.Fatalf("len %d off %d doff %d: SealTLS13 differs from Seal", n, off, doff)
				}
				opened, err := a.Open(nil, nonce, got, ad)
				if err != nil || !bytes.Equal(opened, joined) {
					t.Fatalf("len %d off %d doff %d: does not open (%v)", n, off, doff, err)
				}
				cases++
			}
		}
	}
	t.Logf("%d cases", cases)
}

func TestIronKVMSealTLS13Overlap(t *testing.T) {
	a, _ := New(make([]byte, KeySize))
	s := a.(*chacha20poly1305)
	buf := make([]byte, 200)
	defer func() {
		if recover() == nil {
			t.Fatal("inexact overlap accepted")
		}
	}()
	s.SealTLS13(buf[:5], make([]byte, NonceSize), buf[3:100], 23, nil)
}
