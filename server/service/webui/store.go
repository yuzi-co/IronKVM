// Package webui installs web UI packages apart from the server, and decides
// which UI the server serves.
//
// The application package still carries a UI, in <server dir>/web, and that
// one is always there to fall back to. A UI installed on its own lives beside
// it, in <server dir>/ui:
//
//	ui/state.json          {"current": "<id>", "previous": "<id>"}
//	ui/releases/<id>/      one unpacked package: manifest.json and web/
//
// An empty id means the bundled UI. state.json is replaced by a rename, so a
// switch is atomic: a reader, or a power cut, sees the old state or the new
// one. Only the releases the state names are kept: the current one and one
// previous, for a one-step rollback.
//
// An application update replaces the server directory, ui/ with it, and the
// board goes back to the UI the new application carries. That is deliberate:
// a separately installed UI was checked against the server it was installed
// on, not the next one.
package webui

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

const (
	stateFile   = "state.json"
	releasesDir = "releases"
	workPrefix  = ".work-"
	// freeReserve is left free on the filesystem after an install, so the
	// UI never takes the last of the space the rest of the board writes to.
	freeReserve = 16 << 20
)

var releaseIDPattern = regexp.MustCompile(`^[0-9A-Za-z][0-9A-Za-z._-]*$`)

// freeBytes is a variable so a test can run out of space.
var freeBytes = func(p string) (uint64, error) {
	var s unix.Statfs_t
	if err := unix.Statfs(p, &s); err != nil {
		return 0, err
	}
	return uint64(s.Bavail) * uint64(s.Bsize), nil
}

// State is what state.json holds.
type State struct {
	Current  string `json:"current"`
	Previous string `json:"previous"`
}

// Store is the UI directory of one server.
type Store struct {
	// Root is <server dir>/ui and Bundled is <server dir>/web.
	Root    string
	Bundled string
	// API is the server's API version, apicontract.APIVersion.
	API    int
	Policy Policy
	// Probe, when set, is asked to serve the entry document from a web
	// directory before that directory is switched to. The router sets it.
	Probe func(webDir string) error

	mu     sync.Mutex
	active atomic.Pointer[activeUI]
}

type activeUI struct {
	id  string // "" for the bundled UI
	dir string
}

// NewStore returns a store and selects the UI to serve.
func NewStore(root, bundled string, api int, policy Policy) *Store {
	s := &Store{Root: root, Bundled: bundled, API: api, Policy: policy}
	s.Load()
	return s
}

// Dir is the web directory to serve now.
func (s *Store) Dir() string {
	if a := s.active.Load(); a != nil {
		return a.dir
	}
	return s.Bundled
}

func (s *Store) releaseDir(id string) string {
	return filepath.Join(s.Root, releasesDir, id)
}

func (s *Store) readState() State {
	var st State
	data, err := os.ReadFile(filepath.Join(s.Root, stateFile))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Warnf("webui: cannot read %s: %v", stateFile, err)
		}
		return st
	}
	if err := json.Unmarshal(data, &st); err != nil {
		log.Warnf("webui: %s does not parse, serving the bundled UI: %v", stateFile, err)
		return State{}
	}
	if st.Current != "" && !releaseIDPattern.MatchString(st.Current) {
		st.Current = ""
	}
	if st.Previous != "" && !releaseIDPattern.MatchString(st.Previous) {
		st.Previous = ""
	}
	return st
}

// writeState replaces state.json by a rename, after the new content is on
// disk.
func (s *Store) writeState(st State) error {
	data, err := json.Marshal(st)
	if err != nil {
		return err
	}
	tmp := filepath.Join(s.Root, stateFile+".tmp")
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, filepath.Join(s.Root, stateFile)); err != nil {
		return err
	}
	syncDir(s.Root)
	return nil
}

func syncDir(dir string) {
	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		_ = d.Close()
	}
}

// usable reports whether a release passes the basic check. "" is the bundled
// UI, which came with this server and is always usable.
func (s *Store) usable(id string) error {
	if id == "" {
		return nil
	}
	_, err := basicCheck(s.releaseDir(id), s.API)
	return err
}

func (s *Store) webDirFor(id string) string {
	if id == "" {
		return s.Bundled
	}
	return filepath.Join(s.releaseDir(id), webDir)
}

func (s *Store) setActive(id string) {
	s.active.Store(&activeUI{id: id, dir: s.webDirFor(id)})
}

// Load selects the UI to serve: the current release if it passes the basic
// check, else the previous one, else the bundled UI. It changes no file, so a
// UI that fails because the server binary was swapped by hand comes back when
// the binary does.
func (s *Store) Load() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Work directories an interrupted install or upload left behind.
	for _, prefix := range []string{workPrefix, uploadPrefix} {
		stale, _ := filepath.Glob(filepath.Join(s.Root, prefix+"*"))
		for _, dir := range stale {
			_ = os.RemoveAll(dir)
		}
	}

	st := s.readState()
	for _, id := range []string{st.Current, st.Previous} {
		if id == "" {
			break
		}
		if err := s.usable(id); err != nil {
			log.Warnf("webui: not serving UI %s: %v", id, err)
			continue
		}
		s.setActive(id)
		log.Infof("webui: serving installed UI %s", id)
		return
	}
	s.setActive("")
}

// ReleaseInfo describes one UI for the update page.
type ReleaseInfo struct {
	ID      string   `json:"id"`
	Version string   `json:"version"`
	API     APIRange `json:"api"`
	Bundled bool     `json:"bundled"`
}

// Status is what GET /api/ui answers.
type Status struct {
	ServerAPI      int          `json:"serverApi"`
	Active         ReleaseInfo  `json:"active"`
	Previous       *ReleaseInfo `json:"previous,omitempty"`
	CanRollback    bool         `json:"canRollback"`
	SignatureCheck string       `json:"signatureCheck"`
}

func (s *Store) info(id string) ReleaseInfo {
	if id == "" {
		info := ReleaseInfo{Bundled: true}
		if data, err := os.ReadFile(filepath.Join(s.Bundled, BundledVersionFile)); err == nil {
			var v struct {
				Version string   `json:"version"`
				API     APIRange `json:"api"`
			}
			if json.Unmarshal(data, &v) == nil {
				info.Version, info.API = v.Version, v.API
			}
		}
		return info
	}
	info := ReleaseInfo{ID: id}
	if m, err := readManifest(s.releaseDir(id)); err == nil {
		info.Version, info.API = m.Version, m.API
	}
	return info
}

// BundledVersionFile is written into web/dist by the build, so the bundled UI
// can say which version it is.
const BundledVersionFile = "ui-version.json"

// Status reports what is served and what a rollback would serve.
//
// It takes no lock, so the page can ask while an install runs; what it reads
// is replaced by renames, never written in place.
func (s *Store) Status() Status {
	activeID := ""
	if a := s.active.Load(); a != nil {
		activeID = a.id
	}
	st := s.readState()
	out := Status{
		ServerAPI:      s.API,
		Active:         s.info(activeID),
		SignatureCheck: s.Policy.State(),
	}
	if activeID != "" {
		prev := s.rollbackTarget(st, activeID)
		p := s.info(prev)
		out.Previous = &p
		out.CanRollback = true
	}
	return out
}

// rollbackTarget is what a rollback from activeID serves: the previous
// release when the active one is the current one, else the bundled UI.
func (s *Store) rollbackTarget(st State, activeID string) string {
	if activeID == st.Current && st.Previous != "" && s.usable(st.Previous) == nil {
		return st.Previous
	}
	return ""
}

// InstallResult says what an install did.
type InstallResult struct {
	Installed ReleaseInfo `json:"installed"`
	Verdict   Verdict     `json:"signature"`
}

// ErrBusy means another install or rollback is running.
var ErrBusy = errors.New("another UI install or rollback is running")

// CheckPackage verifies the signature and reads the manifest, without
// unpacking anything. Install runs it first.
func (s *Store) CheckPackage(pkg, sig string) (*Manifest, Verdict, uint64, error) {
	info, err := os.Stat(pkg)
	if err != nil {
		return nil, Verdict{}, 0, err
	}
	if info.Size() > maxPackageBytes {
		return nil, Verdict{}, 0, fmt.Errorf("the package is larger than %d bytes", maxPackageBytes)
	}
	verdict, err := s.Policy.VerifyFile(pkg, sig)
	if err != nil {
		return nil, Verdict{}, 0, err
	}
	ai, err := inspectPackage(pkg)
	if err != nil {
		return nil, verdict, 0, err
	}
	if !ai.manifest.API.Contains(s.API) {
		return ai.manifest, verdict, 0, fmt.Errorf("UI %s supports server API %d..%d, and this server speaks %d",
			ai.manifest.Version, ai.manifest.API.Min, ai.manifest.API.Max, s.API)
	}
	return ai.manifest, verdict, ai.expandedBytes, nil
}

// Install verifies a package and its signature, unpacks it beside the
// releases it keeps, checks it, and switches to it. The UI that was served
// becomes the previous one. If the new UI fails the check after it is
// unpacked, nothing is switched; if it fails the probe after the switch, the
// old state is put back.
func (s *Store) Install(pkg, sig string) (InstallResult, error) {
	if !s.mu.TryLock() {
		return InstallResult{}, ErrBusy
	}
	defer s.mu.Unlock()

	m, verdict, expanded, err := s.CheckPackage(pkg, sig)
	if err != nil {
		return InstallResult{}, err
	}
	if verdict.Warning != "" {
		log.Warnf("webui: %s", verdict.Warning)
	}

	if err := os.MkdirAll(filepath.Join(s.Root, releasesDir), 0o755); err != nil {
		return InstallResult{}, err
	}
	s.pruneLocked(s.readState())
	free, err := freeBytes(s.Root)
	if err != nil {
		return InstallResult{}, fmt.Errorf("cannot read free space: %w", err)
	}
	if free < expanded+freeReserve {
		return InstallResult{}, fmt.Errorf("not enough space: the UI needs %d MiB and %d MiB is free",
			(expanded+freeReserve)>>20, free>>20)
	}

	sum, err := fileSHA256(pkg)
	if err != nil {
		return InstallResult{}, err
	}
	id := releaseID(m.Version, sum)
	active := s.active.Load()
	if active != nil && active.id == id {
		return InstallResult{}, fmt.Errorf("UI %s is already the one served", m.Version)
	}

	work, err := os.MkdirTemp(s.Root, workPrefix)
	if err != nil {
		return InstallResult{}, err
	}
	defer func() { _ = os.RemoveAll(work) }()
	if err := extractPackage(pkg, work); err != nil {
		return InstallResult{}, fmt.Errorf("unpack: %w", err)
	}
	if err := verifyTree(filepath.Join(work, webDir), m); err != nil {
		return InstallResult{}, fmt.Errorf("the unpacked UI does not match its manifest: %w", err)
	}
	if _, err := basicCheck(work, s.API); err != nil {
		return InstallResult{}, err
	}

	dest := s.releaseDir(id)
	if err := os.RemoveAll(dest); err != nil {
		return InstallResult{}, err
	}
	if err := os.Rename(work, dest); err != nil {
		return InstallResult{}, err
	}
	syncDir(filepath.Join(s.Root, releasesDir))

	prevActive := ""
	if active != nil {
		prevActive = active.id
	}
	old := s.readState()
	next := State{Current: id, Previous: prevActive}
	if err := s.switchTo(next, old); err != nil {
		// Unless the old state names it too: reinstalling the previous
		// release replaced its directory with an identical one.
		if id != old.Current && id != old.Previous {
			_ = os.RemoveAll(dest)
		}
		return InstallResult{}, err
	}
	s.pruneLocked(next)
	log.Infof("webui: installed UI %s as %s", m.Version, id)
	return InstallResult{Installed: s.info(id), Verdict: verdict}, nil
}

// switchTo makes next the state and serves its current UI, after the probe
// accepts it. When the probe refuses, old is written back and the UI served
// before stays served.
func (s *Store) switchTo(next, old State) error {
	if err := s.usable(next.Current); err != nil {
		return fmt.Errorf("the UI failed its check, keeping the one served: %w", err)
	}
	if err := s.writeState(next); err != nil {
		return err
	}
	if s.Probe != nil {
		if err := s.Probe(s.webDirFor(next.Current)); err != nil {
			if werr := s.writeState(old); werr != nil {
				log.Errorf("webui: cannot restore the previous state: %v", werr)
			}
			return fmt.Errorf("the UI failed to serve, keeping the one served: %w", err)
		}
	}
	s.setActive(next.Current)
	return nil
}

// Rollback goes back one step: from an installed UI to the previous one, or
// to the bundled UI when there is no previous one or it no longer passes the
// check. The UI rolled back from is removed.
func (s *Store) Rollback() (ReleaseInfo, error) {
	if !s.mu.TryLock() {
		return ReleaseInfo{}, ErrBusy
	}
	defer s.mu.Unlock()

	active := s.active.Load()
	if active == nil || active.id == "" {
		return ReleaseInfo{}, errors.New("the bundled UI is served; there is nothing to roll back")
	}
	old := s.readState()
	target := s.rollbackTarget(old, active.id)
	next := State{Current: target}
	if err := s.switchTo(next, old); err != nil {
		if target == "" {
			return ReleaseInfo{}, err
		}
		// The previous release would not serve; the bundled UI always can.
		log.Warnf("webui: previous UI %s failed, rolling back to the bundled UI: %v", target, err)
		next = State{}
		if err := s.switchTo(next, old); err != nil {
			return ReleaseInfo{}, err
		}
	}
	s.pruneLocked(next)
	log.Infof("webui: rolled back from %s", active.id)
	return s.info(next.Current), nil
}

// pruneLocked removes every release the state does not name, and any work
// directory an interrupted install left.
func (s *Store) pruneLocked(st State) {
	keep := map[string]bool{st.Current: true, st.Previous: true}
	if a := s.active.Load(); a != nil {
		keep[a.id] = true
	}
	entries, _ := os.ReadDir(filepath.Join(s.Root, releasesDir))
	for _, e := range entries {
		if !keep[e.Name()] {
			_ = os.RemoveAll(filepath.Join(s.Root, releasesDir, e.Name()))
		}
	}
	works, _ := filepath.Glob(filepath.Join(s.Root, workPrefix+"*"))
	for _, w := range works {
		_ = os.RemoveAll(w)
	}
}

// releaseID names a release by its version and the package it came from, so
// two builds of one version do not share a directory.
func releaseID(version, sha256hex string) string {
	v := strings.NewReplacer("+", "_").Replace(version)
	return v + "-" + sha256hex[:12]
}
