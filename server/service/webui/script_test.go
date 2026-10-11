package webui

import (
	"os"
	"path/filepath"
	"testing"
)

// TestInstallsAPackageMadeByTheScript installs a package that
// scripts/package-ui.sh built and minisign signed, so the two halves are held
// to one format. It needs the package and the public key, and is skipped
// without them:
//
//	IRONKVM_UI_TEST_PACKAGE=/path/ironkvm-ui_<v>.tar.gz   (its .minisig beside it)
//	IRONKVM_UI_TEST_PUBKEY=/path/test.pub
func TestInstallsAPackageMadeByTheScript(t *testing.T) {
	pkg := os.Getenv("IRONKVM_UI_TEST_PACKAGE")
	pub := os.Getenv("IRONKVM_UI_TEST_PUBKEY")
	if pkg == "" || pub == "" {
		t.Skip("IRONKVM_UI_TEST_PACKAGE and IRONKVM_UI_TEST_PUBKEY are not set")
	}
	dir := t.TempDir()
	keys := filepath.Join(dir, "keys")
	if err := os.MkdirAll(keys, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(pub)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(keys, "test.pub"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	bundled := filepath.Join(dir, "web")
	if err := os.MkdirAll(bundled, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundled, "index.html"), []byte("bundled"), 0o644); err != nil {
		t.Fatal(err)
	}

	s := NewStore(filepath.Join(dir, "ui"), bundled, 1, Policy{KeyDirs: []string{keys}})
	res, err := s.Install(pkg, pkg+".minisig")
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if !res.Verdict.Checked {
		t.Fatalf("the signature was not checked: %+v", res.Verdict)
	}
	if err := verifyTree(s.Dir(), mustManifest(t, filepath.Dir(s.Dir()))); err != nil {
		t.Fatalf("installed tree: %v", err)
	}
	t.Logf("installed %s, signed by %s", res.Installed.Version, res.Verdict.Signer)
}

func mustManifest(t *testing.T, releaseDir string) *Manifest {
	t.Helper()
	m, err := readManifest(releaseDir)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
