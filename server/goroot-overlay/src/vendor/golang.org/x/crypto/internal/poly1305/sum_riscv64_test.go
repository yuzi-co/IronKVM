// NanoKVM: checks the riscv64 assembly against the generic Go code. The
// toolchain's vendor tree ships without x/crypto's tests. Run it on a device,
// see server/goroot-overlay/overlay.sh.

//go:build gc && !purego

package poly1305

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	mrand "math/rand"
	"testing"
)

// RFC 8439, section 2.5.2.
func TestRFC8439(t *testing.T) {
	key, _ := hex.DecodeString("85d6be7857556d337f4452fe42d506a80103808afb0db2fd4abff6af4149f51b")
	want, _ := hex.DecodeString("a8061dc1305136c6c22b8baf0c0127a9")
	var k [32]byte
	copy(k[:], key)
	var tag [16]byte
	Sum(&tag, []byte("Cryptographic Forum Research Group"), &k)
	if !bytes.Equal(tag[:], want) {
		t.Errorf("tag %x, want %x", tag, want)
	}
}

// TestUpdateMatchesGeneric runs update and updateGeneric from the same random
// state over random, misaligned messages of every length class.
func TestUpdateMatchesGeneric(t *testing.T) {
	r := mrand.New(mrand.NewSource(1))
	for i := 0; i < 20000; i++ {
		var key [32]byte
		rand.Read(key[:])
		var a, b macState
		initialize(&key, &a)
		// A partly reduced accumulator, as update leaves it.
		a.h[0] = r.Uint64()
		a.h[1] = r.Uint64()
		a.h[2] = uint64(r.Intn(8))
		b = a

		off := r.Intn(8)
		n := r.Intn(200)
		if i%5 == 0 {
			n = r.Intn(16)
		}
		buf := make([]byte, off+n)
		rand.Read(buf)
		msg := buf[off:]

		update(&a, msg)
		updateGeneric(&b, msg)
		if a != b {
			t.Fatalf("%d bytes at offset %d: state %x, generic %x", n, off, a.h, b.h)
		}
	}
}

// TestSumMatchesGeneric compares whole tags, written in uneven pieces.
func TestSumMatchesGeneric(t *testing.T) {
	r := mrand.New(mrand.NewSource(2))
	for i := 0; i < 2000; i++ {
		var key [32]byte
		rand.Read(key[:])
		msg := make([]byte, r.Intn(3000))
		rand.Read(msg)

		g := newMACGeneric(&key)
		g.Write(msg)
		var want [16]byte
		g.Sum(&want)

		m := New(&key)
		for j := 0; j < len(msg); {
			k := 1 + r.Intn(len(msg)-j)
			m.Write(msg[j : j+k])
			j += k
		}
		got := m.Sum(nil)
		if !bytes.Equal(got, want[:]) {
			t.Fatalf("%d bytes: tag %x, generic %x", len(msg), got, want)
		}
	}
}

func BenchmarkUpdate16K(b *testing.B) {
	var key [32]byte
	msg := make([]byte, 16384+5)[5:]
	for _, im := range []struct {
		name string
		fn   func(*macState, []byte)
	}{{"generic", updateGeneric}, {"asm", update}} {
		b.Run(im.name, func(b *testing.B) {
			var s macState
			initialize(&key, &s)
			b.SetBytes(int64(len(msg)))
			for i := 0; i < b.N; i++ {
				im.fn(&s, msg)
			}
		})
	}
}
