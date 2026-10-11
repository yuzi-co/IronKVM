package webui

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"golang.org/x/crypto/blake2b"
)

// The keys in these tests are made fresh by each test with crypto/rand and
// live only in its temporary directory. No key, test or real, is committed
// for them; testdata/ holds one public key and one signature made by the
// minisign tool itself, whose secret half was thrown away (see
// testdata/README).

type testKey struct {
	id   [keyIDLen]byte
	pub  ed25519.PublicKey
	priv ed25519.PrivateKey
}

func newTestKey(t *testing.T) *testKey {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	k := &testKey{pub: pub, priv: priv}
	if _, err := rand.Read(k.id[:]); err != nil {
		t.Fatal(err)
	}
	return k
}

// pubFile is the key as minisign -G writes it.
func (k *testKey) pubFile() []byte {
	raw := append(append([]byte("Ed"), k.id[:]...), k.pub...)
	return []byte("untrusted comment: minisign public key " + hex.EncodeToString(k.id[:]) + "\n" +
		base64.StdEncoding.EncodeToString(raw) + "\n")
}

func (k *testKey) writePub(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), k.pubFile(), 0o644); err != nil {
		t.Fatal(err)
	}
}

// sign makes a .minisig for content, prehashed ("ED") unless legacy is set.
func (k *testKey) sign(content []byte, trusted string, legacy bool) []byte {
	alg, message := "ED", content
	if legacy {
		alg = "Ed"
	} else {
		h := blake2b.Sum512(content)
		message = h[:]
	}
	sig := ed25519.Sign(k.priv, message)
	raw := append(append([]byte(alg), k.id[:]...), sig...)
	global := ed25519.Sign(k.priv, append(append([]byte(nil), sig...), trusted...))
	return []byte("untrusted comment: signature from minisign secret key\n" +
		base64.StdEncoding.EncodeToString(raw) + "\n" +
		"trusted comment: " + trusted + "\n" +
		base64.StdEncoding.EncodeToString(global) + "\n")
}

func (k *testKey) signFile(t *testing.T, path string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path+".minisig", k.sign(content, "timestamp:1 file:"+filepath.Base(path)+" hashed", false), 0o644); err != nil {
		t.Fatal(err)
	}
}

type tarEntry struct {
	name string
	body []byte
	typ  byte
	link string
}

func writeTarGz(t *testing.T, path string, entries []tarEntry) {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		typ := e.typ
		if typ == 0 {
			typ = tar.TypeReg
		}
		h := &tar.Header{Name: e.name, Typeflag: typ, Mode: 0o644, Linkname: e.link}
		if typ == tar.TypeReg {
			h.Size = int64(len(e.body))
		}
		if typ == tar.TypeDir {
			h.Mode = 0o755
		}
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if typ == tar.TypeReg {
			if _, err := tw.Write(e.body); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

// uiFiles is a small built UI: an index.html that loads one hashed asset.
func uiFiles(tag string) map[string][]byte {
	return map[string][]byte{
		"index.html":                  []byte(`<!doctype html><html><head><script type="module" src="/assets/index-` + tag + `.js"></script></head><body><div id="root"></div></body></html>`),
		"assets/index-" + tag + ".js": []byte("console.log('" + tag + "')"),
		"logo.svg":                    []byte("<svg/>"),
	}
}

func manifestFor(version string, api APIRange, files map[string][]byte) Manifest {
	m := Manifest{Format: ManifestFormat, Name: PackageName, Version: version, API: api, Files: map[string]string{}}
	for name, body := range files {
		sum := sha256.Sum256(body)
		m.Files[name] = hex.EncodeToString(sum[:])
	}
	return m
}

// buildPackage writes ironkvm-ui_<version>.tar.gz into dir from files and m.
func buildPackage(t *testing.T, dir string, m Manifest, files map[string][]byte) string {
	t.Helper()
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	entries := []tarEntry{{name: "manifest.json", body: data}, {name: "web/", typ: tar.TypeDir}}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		entries = append(entries, tarEntry{name: "web/" + name, body: files[name]})
	}
	path := filepath.Join(dir, "ironkvm-ui_"+m.Version+".tar.gz")
	writeTarGz(t, path, entries)
	return path
}

// testEnv is a server directory with a bundled UI and a keyring of one key.
type testEnv struct {
	dir   string
	key   *testKey
	store *Store
	pkgs  string
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	dir := t.TempDir()
	bundled := filepath.Join(dir, "web")
	if err := os.MkdirAll(bundled, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundled, "index.html"), []byte("bundled"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundled, BundledVersionFile), []byte(`{"version":"1.0.0","api":{"min":1,"max":1}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	key := newTestKey(t)
	keys := filepath.Join(dir, "keys")
	key.writePub(t, keys, "test.pub")
	policy := Policy{KeyDirs: []string{keys}, CheckOff: filepath.Join(dir, "no-signature-check")}
	store := NewStore(filepath.Join(dir, "ui"), bundled, 1, policy)
	pkgs := filepath.Join(dir, "pkgs")
	if err := os.MkdirAll(pkgs, 0o755); err != nil {
		t.Fatal(err)
	}
	return &testEnv{dir: dir, key: key, store: store, pkgs: pkgs}
}

// signedPackage builds and signs a good package for version.
func (e *testEnv) signedPackage(t *testing.T, version string) string {
	t.Helper()
	files := uiFiles(version)
	pkg := buildPackage(t, e.pkgs, manifestFor(version, APIRange{Min: 1, Max: 1}, files), files)
	e.key.signFile(t, pkg)
	return pkg
}

func (e *testEnv) install(t *testing.T, pkg string) InstallResult {
	t.Helper()
	res, err := e.store.Install(pkg, pkg+".minisig")
	if err != nil {
		t.Fatalf("install %s: %v", filepath.Base(pkg), err)
	}
	return res
}

func servedIndex(t *testing.T, s *Store) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(s.Dir(), "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
