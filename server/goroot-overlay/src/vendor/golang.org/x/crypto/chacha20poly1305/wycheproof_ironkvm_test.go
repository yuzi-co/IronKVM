package chacha20poly1305

import (
	"bytes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

type wpFile struct {
	Algorithm     string `json:"algorithm"`
	NumberOfTests int    `json:"numberOfTests"`
	TestGroups    []struct {
		Tests []struct {
			TcID    int      `json:"tcId"`
			Comment string   `json:"comment"`
			Key     string   `json:"key"`
			Iv      string   `json:"iv"`
			Aad     string   `json:"aad"`
			Msg     string   `json:"msg"`
			Ct      string   `json:"ct"`
			Tag     string   `json:"tag"`
			Result  string   `json:"result"`
			Flags   []string `json:"flags"`
		} `json:"tests"`
	} `json:"testGroups"`
}

func h(t *testing.T, s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func runWycheproof(t *testing.T, env string, mk func([]byte) (cipher.AEAD, error)) {
	path := os.Getenv(env)
	if path == "" {
		t.Skip(env + " not set")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var f wpFile
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	n, pass, skipped, tls13 := 0, 0, 0, 0
	for _, g := range f.TestGroups {
		for _, tc := range g.Tests {
			n++
			key, iv, aad, msg := h(t, tc.Key), h(t, tc.Iv), h(t, tc.Aad), h(t, tc.Msg)
			ct, tag := h(t, tc.Ct), h(t, tc.Tag)
			a, err := mk(key)
			if err != nil || len(iv) != a.NonceSize() {
				// Wrong key or nonce sizes: the API rejects them up front.
				if tc.Result == "valid" {
					t.Errorf("tc %d: valid vector rejected by size (%v)", tc.TcID, err)
					continue
				}
				skipped++
				pass++
				continue
			}
			sealed := append(append([]byte{}, ct...), tag...)
			if tc.Result == "valid" || tc.Result == "acceptable" {
				got := a.Seal(nil, iv, msg, aad)
				if !bytes.Equal(got, sealed) {
					t.Errorf("tc %d (%s): Seal mismatch", tc.TcID, tc.Comment)
					continue
				}
				// SealTLS13 (seal_tls13_ironkvm.go) on the same vector, its
				// last message byte taken as the content type.
				if s, ok := a.(*chacha20poly1305); ok && len(msg) > 0 {
					got := s.SealTLS13(nil, iv, msg[:len(msg)-1], msg[len(msg)-1], aad)
					if !bytes.Equal(got, sealed) {
						t.Errorf("tc %d (%s): SealTLS13 mismatch", tc.TcID, tc.Comment)
						continue
					}
					tls13++
				}
			}
			pt, err := a.Open(nil, iv, sealed, aad)
			ok := err == nil && bytes.Equal(pt, msg)
			switch {
			case tc.Result == "valid" && !ok:
				t.Errorf("tc %d (%s): Open failed on a valid vector", tc.TcID, tc.Comment)
			case tc.Result == "invalid" && err == nil:
				t.Errorf("tc %d (%s): Open accepted an invalid vector", tc.TcID, tc.Comment)
			default:
				pass++
			}
		}
	}
	if n != f.NumberOfTests {
		t.Errorf("%s: ran %d of %d tests", f.Algorithm, n, f.NumberOfTests)
	}
	t.Logf("%s: %d tests, %d passed (%d rejected by size as expected), %d also through SealTLS13", f.Algorithm, n, pass, skipped, tls13)
}

func TestIronKVMWycheproofChaCha20Poly1305(t *testing.T) {
	runWycheproof(t, "WP_CHACHA", New)
}

func TestIronKVMWycheproofXChaCha20Poly1305(t *testing.T) {
	runWycheproof(t, "WP_XCHACHA", NewX)
}
