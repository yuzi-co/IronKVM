package webui

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func releaseNames(t *testing.T, s *Store) []string {
	t.Helper()
	entries, _ := os.ReadDir(filepath.Join(s.Root, releasesDir))
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestInstallSwitchesAndKeepsOnePrevious(t *testing.T) {
	e := newTestEnv(t)
	if servedIndex(t, e.store) != "bundled" {
		t.Fatal("a fresh store should serve the bundled UI")
	}

	res := e.install(t, e.signedPackage(t, "1.1.0"))
	if res.Installed.Version != "1.1.0" || !res.Verdict.Checked || res.Verdict.Signer != "test.pub" {
		t.Fatalf("install result %+v", res)
	}
	if !strings.Contains(servedIndex(t, e.store), "index-1.1.0.js") {
		t.Fatal("the installed UI is not served")
	}
	st := e.store.Status()
	if st.Active.Version != "1.1.0" || !st.CanRollback || st.Previous == nil || !st.Previous.Bundled {
		t.Fatalf("status after the first install %+v", st)
	}

	e.install(t, e.signedPackage(t, "1.2.0"))
	e.install(t, e.signedPackage(t, "1.3.0"))
	st = e.store.Status()
	if st.Active.Version != "1.3.0" || st.Previous == nil || st.Previous.Version != "1.2.0" {
		t.Fatalf("status after three installs %+v", st)
	}
	if n := len(releaseNames(t, e.store)); n != 2 {
		t.Fatalf("%d releases kept, want the current and one previous: %v", n, releaseNames(t, e.store))
	}
}

func TestRollbackGoesBackOneStepThenToBundled(t *testing.T) {
	e := newTestEnv(t)
	e.install(t, e.signedPackage(t, "1.1.0"))
	e.install(t, e.signedPackage(t, "1.2.0"))

	info, err := e.store.Rollback()
	if err != nil || info.Version != "1.1.0" {
		t.Fatalf("first rollback: %+v %v", info, err)
	}
	if !strings.Contains(servedIndex(t, e.store), "index-1.1.0.js") {
		t.Fatal("rollback did not serve the previous UI")
	}
	if n := len(releaseNames(t, e.store)); n != 1 {
		t.Fatalf("the UI rolled back from should be removed: %v", releaseNames(t, e.store))
	}

	info, err = e.store.Rollback()
	if err != nil || !info.Bundled || info.Version != "1.0.0" {
		t.Fatalf("second rollback: %+v %v", info, err)
	}
	if servedIndex(t, e.store) != "bundled" {
		t.Fatal("the second rollback should serve the bundled UI")
	}
	if _, err := e.store.Rollback(); err == nil {
		t.Fatal("a rollback from the bundled UI should be refused")
	}
}

// A refused package changes nothing: the same UI is served and no release
// directory is left behind.
func assertUnchanged(t *testing.T, e *testEnv, wantIndexPart string) {
	t.Helper()
	if !strings.Contains(servedIndex(t, e.store), wantIndexPart) {
		t.Fatalf("served UI changed: %q", servedIndex(t, e.store))
	}
	works, _ := filepath.Glob(filepath.Join(e.store.Root, workPrefix+"*"))
	if len(works) != 0 {
		t.Fatalf("work directories left: %v", works)
	}
}

func TestTamperedPackageIsRefused(t *testing.T) {
	e := newTestEnv(t)
	e.install(t, e.signedPackage(t, "1.1.0"))

	pkg := e.signedPackage(t, "1.2.0")
	// Re-pack with a changed asset and keep the old signature.
	files := uiFiles("1.2.0")
	files["logo.svg"] = []byte("<svg onload=alert(1)/>")
	sig, _ := os.ReadFile(pkg + ".minisig")
	buildPackage(t, e.pkgs, manifestFor("1.2.0", APIRange{1, 1}, files), files)
	if err := os.WriteFile(pkg+".minisig", sig, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := e.store.Install(pkg, pkg+".minisig"); !errors.Is(err, ErrAltered) {
		t.Fatalf("tampered package: %v", err)
	}
	assertUnchanged(t, e, "index-1.1.0.js")
	if n := len(releaseNames(t, e.store)); n != 1 {
		t.Fatalf("releases after a refusal: %v", releaseNames(t, e.store))
	}
}

func TestUnsignedAndStrangerSignedPackagesAreRefused(t *testing.T) {
	e := newTestEnv(t)
	files := uiFiles("1.1.0")
	pkg := buildPackage(t, e.pkgs, manifestFor("1.1.0", APIRange{1, 1}, files), files)
	if _, err := e.store.Install(pkg, pkg+".minisig"); !errors.Is(err, ErrNoSignature) {
		t.Fatalf("unsigned: %v", err)
	}
	newTestKey(t).signFile(t, pkg)
	if _, err := e.store.Install(pkg, pkg+".minisig"); !errors.Is(err, ErrUnknownKey) {
		t.Fatalf("stranger: %v", err)
	}
	assertUnchanged(t, e, "bundled")
}

func TestIncompatiblePackageIsRefused(t *testing.T) {
	e := newTestEnv(t)
	files := uiFiles("2.0.0")
	pkg := buildPackage(t, e.pkgs, manifestFor("2.0.0", APIRange{2, 3}, files), files)
	e.key.signFile(t, pkg)
	_, err := e.store.Install(pkg, pkg+".minisig")
	if err == nil || !strings.Contains(err.Error(), "server API 2..3") {
		t.Fatalf("incompatible: %v", err)
	}
	assertUnchanged(t, e, "bundled")
}

func TestManifestMustMatchTheFiles(t *testing.T) {
	cases := map[string]func(files map[string][]byte, m *Manifest){
		"wrong hash": func(files map[string][]byte, m *Manifest) {
			files["logo.svg"] = []byte("other")
		},
		"unlisted file": func(files map[string][]byte, m *Manifest) {
			files["extra.js"] = []byte("x")
		},
		"missing file": func(files map[string][]byte, m *Manifest) {
			delete(files, "logo.svg")
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			e := newTestEnv(t)
			files := uiFiles("1.1.0")
			m := manifestFor("1.1.0", APIRange{1, 1}, files)
			mutate(files, &m)
			pkg := buildPackage(t, e.pkgs, m, files)
			e.key.signFile(t, pkg)
			if _, err := e.store.Install(pkg, pkg+".minisig"); err == nil {
				t.Fatal("a package whose files disagree with its manifest was installed")
			}
			assertUnchanged(t, e, "bundled")
		})
	}
}

func TestUnsafeArchivesAreRefused(t *testing.T) {
	m := manifestFor("1.1.0", APIRange{1, 1}, uiFiles("1.1.0"))
	data, _ := json.Marshal(m)
	cases := map[string][]tarEntry{
		"traversal":   {{name: "manifest.json", body: data}, {name: "web/../../evil", body: []byte("x")}},
		"absolute":    {{name: "manifest.json", body: data}, {name: "/etc/evil", body: []byte("x")}},
		"symlink":     {{name: "manifest.json", body: data}, {name: "web/index.html", typ: tar.TypeSymlink, link: "/etc/shadow"}},
		"outside":     {{name: "manifest.json", body: data}, {name: "server/NanoKVM-Server", body: []byte("x")}},
		"duplicate":   {{name: "manifest.json", body: data}, {name: "web/a", body: []byte("x")}, {name: "web/a", body: []byte("y")}},
		"no manifest": {{name: "web/index.html", body: []byte("x")}},
	}
	for name, entries := range cases {
		t.Run(name, func(t *testing.T) {
			e := newTestEnv(t)
			pkg := filepath.Join(e.pkgs, "ironkvm-ui_1.1.0.tar.gz")
			writeTarGz(t, pkg, entries)
			e.key.signFile(t, pkg)
			if _, err := e.store.Install(pkg, pkg+".minisig"); err == nil {
				t.Fatal("an unsafe archive was installed")
			}
			assertUnchanged(t, e, "bundled")
			if _, err := os.Stat(filepath.Join(e.dir, "evil")); err == nil {
				t.Fatal("a file was written outside the store")
			}
		})
	}
}

func TestFailedProbeKeepsTheOldUI(t *testing.T) {
	e := newTestEnv(t)
	e.install(t, e.signedPackage(t, "1.1.0"))
	e.store.Probe = func(dir string) error {
		if strings.Contains(dir, "1.2.0") {
			return errors.New("does not serve")
		}
		return nil
	}
	pkg := e.signedPackage(t, "1.2.0")
	if _, err := e.store.Install(pkg, pkg+".minisig"); err == nil {
		t.Fatal("a UI that failed its probe was switched to")
	}
	assertUnchanged(t, e, "index-1.1.0.js")
	if st := e.store.readState(); !strings.HasPrefix(st.Current, "1.1.0-") {
		t.Fatalf("state was not restored: %+v", st)
	}
	// A restart reads the same.
	e.store.Load()
	assertUnchanged(t, e, "index-1.1.0.js")
}

func TestLoadFallsBackWhenTheCurrentUIIsBroken(t *testing.T) {
	e := newTestEnv(t)
	e.install(t, e.signedPackage(t, "1.1.0"))
	e.install(t, e.signedPackage(t, "1.2.0"))
	st := e.store.readState()

	// The current release's index.html no longer matches its manifest.
	index := filepath.Join(e.store.releaseDir(st.Current), webDir, indexFile)
	if err := os.WriteFile(index, []byte("broken"), 0o644); err != nil {
		t.Fatal(err)
	}
	e.store.Load()
	if !strings.Contains(servedIndex(t, e.store), "index-1.1.0.js") {
		t.Fatal("a broken current UI should fall back to the previous one")
	}

	// And to the bundled UI when the previous one is gone too.
	if err := os.RemoveAll(e.store.releaseDir(st.Previous)); err != nil {
		t.Fatal(err)
	}
	e.store.Load()
	if servedIndex(t, e.store) != "bundled" {
		t.Fatal("with no usable release the bundled UI should be served")
	}
}

func TestLoadRefusesAUIForAnotherServerAPI(t *testing.T) {
	e := newTestEnv(t)
	e.install(t, e.signedPackage(t, "1.1.0"))
	// The same directory, read by a server that speaks API 2 (a server binary
	// replaced by hand without an application update).
	other := NewStore(e.store.Root, e.store.Bundled, 2, e.store.Policy)
	if servedIndex(t, other) != "bundled" {
		t.Fatal("a UI that does not support the server's API was served")
	}
}

func TestCorruptStateServesBundled(t *testing.T) {
	e := newTestEnv(t)
	e.install(t, e.signedPackage(t, "1.1.0"))
	if err := os.WriteFile(filepath.Join(e.store.Root, stateFile), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	e.store.Load()
	if servedIndex(t, e.store) != "bundled" {
		t.Fatal("an unreadable state should serve the bundled UI")
	}
	// An id that would leave the releases directory is ignored.
	if err := os.WriteFile(filepath.Join(e.store.Root, stateFile), []byte(`{"current":"../../web"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	e.store.Load()
	if e.store.Dir() != e.store.Bundled {
		t.Fatalf("a traversing id was followed to %s", e.store.Dir())
	}
}

func TestReinstallingTheServedUIIsRefused(t *testing.T) {
	e := newTestEnv(t)
	pkg := e.signedPackage(t, "1.1.0")
	e.install(t, pkg)
	if _, err := e.store.Install(pkg, pkg+".minisig"); err == nil || !strings.Contains(err.Error(), "already") {
		t.Fatalf("reinstall: %v", err)
	}
}

func TestInstallNeedsSpace(t *testing.T) {
	e := newTestEnv(t)
	old := freeBytes
	freeBytes = func(string) (uint64, error) { return 1 << 20, nil }
	defer func() { freeBytes = old }()
	pkg := e.signedPackage(t, "1.1.0")
	if _, err := e.store.Install(pkg, pkg+".minisig"); err == nil || !strings.Contains(err.Error(), "space") {
		t.Fatalf("no space: %v", err)
	}
	assertUnchanged(t, e, "bundled")
}

func TestOwnerSwitchAllowsAnUnsignedPackage(t *testing.T) {
	e := newTestEnv(t)
	if err := os.WriteFile(e.store.Policy.CheckOff, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	files := uiFiles("1.1.0")
	pkg := buildPackage(t, e.pkgs, manifestFor("1.1.0", APIRange{1, 1}, files), files)
	res, err := e.store.Install(pkg, pkg+".minisig")
	if err != nil || res.Verdict.Checked || res.Verdict.Warning == "" {
		t.Fatalf("owner switch: %+v %v", res, err)
	}
}

func TestLoadRemovesStaleWorkDirectories(t *testing.T) {
	e := newTestEnv(t)
	for _, name := range []string{workPrefix + "1", uploadPrefix + "2"} {
		if err := os.MkdirAll(filepath.Join(e.store.Root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	e.store.Load()
	left, _ := filepath.Glob(filepath.Join(e.store.Root, ".*-*"))
	if len(left) != 0 {
		t.Fatalf("stale directories left: %v", left)
	}
}

func TestReinstallingThePreviousUISwapsThem(t *testing.T) {
	e := newTestEnv(t)
	a := e.signedPackage(t, "1.1.0")
	e.install(t, a)
	e.install(t, e.signedPackage(t, "1.2.0"))
	e.install(t, a)
	st := e.store.Status()
	if st.Active.Version != "1.1.0" || st.Previous == nil || st.Previous.Version != "1.2.0" {
		t.Fatalf("status %+v", st)
	}
}
