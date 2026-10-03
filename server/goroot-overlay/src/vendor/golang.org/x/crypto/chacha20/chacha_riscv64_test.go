// NanoKVM: checks the riscv64 assembly against the generic Go code. The
// toolchain's vendor tree ships without x/crypto's tests, so this is the only
// test the package has. Run it on the device, see overlay.sh.

//go:build gc && !purego

package chacha20

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	mrand "math/rand"
	"testing"
)

type impl struct {
	name string
	fn   func(c *Cipher, dst, src []byte)
}

func impls(t testing.TB) []impl {
	var list []impl
	if isTHeadC9xx() {
		list = append(list, impl{"xtheadbb", func(c *Cipher, dst, src []byte) {
			xorKeyStreamTH(dst, src, &c.key, &c.nonce, &c.counter)
		}})
	} else {
		t.Skip("not a T-Head C906 or C910: the assembly does not run here")
	}
	return list
}

// RFC 8439, section 2.4.2.
func TestRFC8439(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	nonce := []byte{0, 0, 0, 0, 0, 0, 0, 0x4a, 0, 0, 0, 0}
	plaintext := []byte("Ladies and Gentlemen of the class of '99: If I could offer you only one tip for the future, sunscreen would be it.")
	want, _ := hex.DecodeString("6e2e359a2568f98041ba0728dd0d6981e97e7aec1d4360c20a27afccfd9fae0bf91b65c5524733ab8f593dabcd62b3571639d624e65152ab8f530c359f0861d807ca0dbf500d6a6156a38e088a22b65e52bc514d16ccf806818ce91ab77937365af90bbf74a35be6b40b8eedf2785e42874d")

	for _, im := range impls(t) {
		c, err := NewUnauthenticatedCipher(key, nonce)
		if err != nil {
			t.Fatal(err)
		}
		c.SetCounter(1)
		// Two whole blocks through the assembly, the rest generic.
		in := make([]byte, 128)
		copy(in, plaintext)
		out := make([]byte, 128)
		im.fn(c, out, in)
		if got := out[:len(want)]; !bytes.Equal(got, want) {
			t.Errorf("%s: got %x, want %x", im.name, got, want)
		}
		if c.counter != 3 {
			t.Errorf("%s: counter %d after two blocks from 1, want 3", im.name, c.counter)
		}
	}
}

// TestMatchesGeneric compares each implementation with the generic code over
// random keys, nonces, counters, lengths and misaligned slices, in place and
// not.
func TestMatchesGeneric(t *testing.T) {
	r := mrand.New(mrand.NewSource(1))
	for _, im := range impls(t) {
		for i := 0; i < 2000; i++ {
			var key [32]byte
			var nonce [12]byte
			rand.Read(key[:])
			rand.Read(nonce[:])
			blocks := r.Intn(70)
			off := r.Intn(8)
			buf := make([]byte, off+blocks*64)
			rand.Read(buf)
			src := buf[off:]
			counter := uint32(r.Intn(1 << 20))
			if i%7 == 0 {
				counter = ^uint32(0) - uint32(blocks)
			}

			ref, _ := NewUnauthenticatedCipher(key[:], nonce[:])
			ref.counter = counter
			want := make([]byte, len(src))
			ref.xorKeyStreamBlocksGeneric(want, src)

			c, _ := NewUnauthenticatedCipher(key[:], nonce[:])
			c.counter = counter
			var got []byte
			if i%2 == 0 {
				got = make([]byte, len(src)+off)[off:]
				im.fn(c, got, src)
			} else {
				got = append([]byte(nil), src...)
				im.fn(c, got, got)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("%s: %d blocks at offset %d from counter %d differ from generic", im.name, blocks, off, counter)
			}
			if c.counter != ref.counter {
				t.Fatalf("%s: counter %d, generic %d", im.name, c.counter, ref.counter)
			}
		}
	}
}

// TestStream runs the public API, with its buffering, over uneven chunks.
func TestStream(t *testing.T) {
	r := mrand.New(mrand.NewSource(2))
	for i := 0; i < 300; i++ {
		var key [32]byte
		var nonce [12]byte
		rand.Read(key[:])
		rand.Read(nonce[:])
		in := make([]byte, r.Intn(5000))
		rand.Read(in)

		ref, _ := NewUnauthenticatedCipher(key[:], nonce[:])
		want := make([]byte, len(in))
		for j := 0; j < len(in); j += 64 {
			end := min(j+64, len(in))
			var block [64]byte
			ref.xorKeyStreamBlocksGeneric(block[:], block[:])
			for k := j; k < end; k++ {
				want[k] = in[k] ^ block[k-j]
			}
		}

		c, _ := NewUnauthenticatedCipher(key[:], nonce[:])
		got := make([]byte, len(in))
		for j := 0; j < len(in); {
			n := 1 + r.Intn(len(in)-j)
			c.XORKeyStream(got[j:j+n], in[j:j+n])
			j += n
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("%d bytes: stream differs from generic", len(in))
		}
	}
}

func BenchmarkXOR16K(b *testing.B) {
	for _, im := range append([]impl{{"generic", func(c *Cipher, dst, src []byte) {
		c.xorKeyStreamBlocksGeneric(dst, src)
	}}}, impls(b)...) {
		b.Run(im.name, func(b *testing.B) {
			var key [32]byte
			var nonce [12]byte
			c, _ := NewUnauthenticatedCipher(key[:], nonce[:])
			buf := make([]byte, 16384+5)[5:]
			b.SetBytes(int64(len(buf)))
			for i := 0; i < b.N; i++ {
				c.counter = 0
				im.fn(c, buf, buf)
			}
		})
	}
}
