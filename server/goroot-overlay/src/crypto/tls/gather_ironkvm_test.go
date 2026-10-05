package tls

import (
	"bytes"
	"crypto/rand"
	"io"
	mathrand "math/rand"
	"testing"
)

// TestIronKVMGather sends random amounts of application data from a server
// that gathers some of it and not the rest, flushing at several sizes, and
// checks that the client reads exactly what was written, for TLS 1.3 (with
// the suite the platform prefers: ChaCha20-Poly1305 through SealTLS13 on a
// core without AES, AES-GCM through encrypt elsewhere) and for TLS 1.2 with
// an AEAD and a CBC suite. A Close while gathering must deliver what was
// gathered before the alert.
func TestIronKVMGather(t *testing.T) {
	cases := []struct {
		name  string
		vers  uint16
		suite uint16
		noAES bool
	}{
		{"TLS13", VersionTLS13, 0, false},
		// As on the C906: ChaCha20-Poly1305, sealed through SealTLS13.
		{"TLS13-noAES", VersionTLS13, TLS_CHACHA20_POLY1305_SHA256, true},
		{"TLS12-CHACHA20", VersionTLS12, TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256, false},
		{"TLS12-AES128GCM", VersionTLS12, TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, false},
		{"TLS12-AES128CBC", VersionTLS12, TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA, false},
	}
	for _, tc := range cases {
		for _, flushAt := range []int{0, 16384, 65536, 300000} {
			t.Run(tc.name, func(t *testing.T) {
				if tc.noAES {
					saved := hasAESGCMHardwareSupport
					hasAESGCMHardwareSupport = false
					defer func() { hasAESGCMHardwareSupport = saved }()
				}
				testGather(t, tc.vers, tc.suite, flushAt)
			})
		}
	}
}

func testGather(t *testing.T, vers, suite uint16, flushAt int) {
	c, s := localPipe(t)
	cfg := testConfig.Clone()
	cfg.MinVersion, cfg.MaxVersion = vers, vers
	if suite != 0 && vers != VersionTLS13 {
		cfg.CipherSuites = []uint16{suite}
	}
	client := Client(c, cfg)
	server := Server(s, cfg)

	type result struct {
		data []byte
		err  error
	}
	got := make(chan result, 1)
	go func() {
		b, err := io.ReadAll(client)
		got <- result{b, err}
	}()

	if err := server.Handshake(); err != nil {
		t.Fatal(err)
	}
	if suite != 0 && server.ConnectionState().CipherSuite != suite {
		t.Fatalf("negotiated %s", CipherSuiteName(server.ConnectionState().CipherSuite))
	}

	rng := mathrand.New(mathrand.NewSource(int64(flushAt) + int64(vers)))
	src := make([]byte, 1<<20)
	rand.Read(src)
	var sent bytes.Buffer
	gathering := false
	for i := 0; i < 60; i++ {
		switch {
		case !gathering && rng.Intn(4) != 0:
			server.BeginGather(flushAt)
			gathering = true
		case gathering && rng.Intn(3) == 0:
			if err := server.EndGather(); err != nil {
				t.Fatal(err)
			}
			gathering = false
		}
		// Sizes around the record boundaries and whole pictures, at any
		// alignment of the source.
		var n int
		switch rng.Intn(4) {
		case 0:
			n = rng.Intn(64)
		case 1:
			n = 16384*(1+rng.Intn(3)) + rng.Intn(3) - 1
		default:
			n = rng.Intn(300000)
		}
		off := rng.Intn(len(src) - n)
		w, err := server.Write(src[off : off+n])
		if err != nil || w != n {
			t.Fatalf("write %d: %d, %v", n, w, err)
		}
		sent.Write(src[off : off+n])
	}
	// The last picture is still gathered, if gathering: Close sends it,
	// then close_notify.
	if !gathering {
		server.BeginGather(flushAt)
		server.Write(src[:12345])
		sent.Write(src[:12345])
	}
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}

	r := <-got
	if r.err != nil {
		t.Fatalf("client read: %v", r.err)
	}
	if !bytes.Equal(r.data, sent.Bytes()) {
		t.Fatalf("client read %d bytes, %d sent, contents equal %v", len(r.data), sent.Len(),
			bytes.Equal(r.data, sent.Bytes()[:min(len(r.data), sent.Len())]))
	}
}
