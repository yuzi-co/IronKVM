package install

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/upstream"
	"NanoKVM-Server/utils/atomicfile"
)

// FreeBytes reports the space available to an unprivileged writer in the
// filesystem that holds path. A variable so the tests can fill the disk.
var FreeBytes = statfsFree

// ErrNoSpace is the free space check's refusal.
var ErrNoSpace = errors.New("not enough free space")

// spaceMargin is kept free beyond the download and the files: the licence
// files, the add-on's small files, and room for its settings.
const spaceMargin = 2 << 20

const downloadTimeout = 10 * time.Minute

// Options are what an add-on's package adds to its manifest.
type Options struct {
	// Client downloads the artifact; upstream.Default when nil.
	Client *upstream.Client
	// URL replaces the manifest's URL. The sum is still the manifest's.
	URL string
	// Apk is the package manager for an apk artifact; ApkPath when empty.
	Apk string
	// Workspace is where the artifact is fetched off a distribution image.
	// On one it is .fetch in the add-on's directory, on the same filesystem
	// as the image.
	Workspace string
	// Progress hears the stage and a percentage; nil for none.
	Progress func(stage string, percent int)
	// Fetch replaces the download for a manifest with no pinned sum: it
	// writes the artifact to dst, checks it against whatever vouches for it
	// in place of a pin, and returns its SHA-256, which installed.json
	// records as the artifact's. An apk artifact still has its signature
	// checked by apk before anything is taken out of it.
	Fetch func(ctx context.Context, dst string, progress func(done, total int64)) (string, error)
	// Plain places the checked files the way the add-on did before images:
	// stage holds each member under its saved name and the licence files
	// in stage/licenses. Install calls it off a distribution image, and on
	// one whose kernel or slot cannot make an image (see the package doc).
	Plain func(stage string) error
}

// Result says where the files went.
type Result struct {
	// Image is true when the add-on is now an image, mounted at
	// FilesDir(name); false when Plain placed the files.
	Image bool
	// Sums are the members' sums, keyed by saved name.
	Sums map[string]string
}

// Installed is installed.json: what is installed, from where, under which
// licence, with the sums of the files, so a later check or the web UI need
// not trust the files themselves.
type Installed struct {
	Manifest Manifest          `json:"manifest"`
	Sums     map[string]string `json:"sums"`
	Source   string            `json:"source"`
	URL      string            `json:"url"`
	Licence  string            `json:"licence"`
	Image    string            `json:"image,omitempty"`
	Time     time.Time         `json:"time"`
}

// ReadInstalled reads the add-on's installed.json.
func ReadInstalled(name string) (Installed, bool) {
	var rec Installed
	data, err := os.ReadFile(filepath.Join(addon.Dir(name), recordFile))
	if err != nil || json.Unmarshal(data, &rec) != nil || rec.Manifest.Name != name {
		return Installed{}, false
	}
	return rec, true
}

// Install fetches, checks and places one version of an add-on. Until the
// final switch nothing an earlier install left is touched, so a failed
// download, a mismatch or a full disk leaves the add-on as it was. The
// workspace goes in every case, and so does the add-on's directory when a
// failed first install left it empty.
func Install(ctx context.Context, m Manifest, opt Options) (Result, error) {
	if err := m.validate(); err != nil {
		return Result{}, err
	}
	progress := opt.Progress
	if progress == nil {
		progress = func(string, int) {}
	}
	onData := addon.OnData()
	image := onData && ImageSupported()
	if !image && opt.Plain == nil {
		return Result{}, fmt.Errorf("%s installs only as an image, and this slot cannot make one", m.Title)
	}
	dir := addon.Dir(m.Name)
	ws := opt.Workspace
	if onData {
		ws = filepath.Join(dir, ".fetch")
	}
	if ws == "" {
		return Result{}, fmt.Errorf("%s: no workspace off a distribution image", m.Name)
	}

	if err := checkFreeSpace(ws, m, image); err != nil {
		return Result{}, err
	}
	_ = os.RemoveAll(ws)
	if err := os.MkdirAll(ws, 0o755); err != nil {
		return Result{}, fmt.Errorf("failed to make %s: %w", ws, err)
	}
	defer func() {
		_ = os.RemoveAll(ws)
		if onData {
			// Only removes the add-on's directory when a failed first
			// install left it empty.
			_ = os.Remove(dir)
		}
	}()

	progress("downloading", 5)
	url := m.Artifact.URL
	if opt.URL != "" {
		url = opt.URL
	}
	artifact := filepath.Join(ws, artifactName(url, m))
	downloading := func(done, total int64) {
		if total <= 0 {
			total = m.Artifact.Size
		}
		if total > 0 {
			progress("downloading", 5+int(done*65/total))
		}
	}
	if opt.Fetch != nil {
		fctx, cancel := context.WithTimeout(ctx, downloadTimeout)
		sum, err := opt.Fetch(fctx, artifact, downloading)
		cancel()
		if err != nil {
			return Result{}, fmt.Errorf("failed to download %s %s: %w", m.Title, m.Version, err)
		}
		m.Artifact.SHA256 = sum
	} else if err := download(ctx, opt.Client, url, m, artifact, downloading); err != nil {
		return Result{}, err
	}

	progress("extracting", 75)
	stage := filepath.Join(ws, "stage")
	var sums map[string]string
	var err error
	switch m.Artifact.Format {
	case FormatApk:
		sums, err = ExtractApk(artifact, m.Artifact.Members, stage, opt.Apk)
	default:
		sums, err = ExtractTarGz(artifact, m.Artifact.Members, stage)
	}
	if err != nil {
		return Result{}, fmt.Errorf("%s %s: %w", m.Title, m.Version, err)
	}
	_ = os.Remove(artifact)
	// Each member was checked as it was written; this stage is for the
	// progress bar, which has always shown one.
	progress("verifying", 85)

	progress("installing", 90)
	res := Result{Image: image, Sums: sums}
	rec := Installed{Manifest: m, Sums: sums, Source: m.Source, URL: url, Licence: m.Licence, Time: time.Now().UTC()}
	if image {
		name := imageName(dir, m)
		staged := filepath.Join(ws, name)
		if err := pack(stage, staged, ws, m.Artifact.Members); err != nil {
			return Result{}, err
		}
		if err := os.Rename(staged, filepath.Join(dir, name)); err != nil {
			return Result{}, err
		}
		if err := activate(m.Name, name); err != nil {
			_ = os.Remove(filepath.Join(dir, name))
			_ = os.Remove(filepath.Join(dir, name+sumSuffix))
			return Result{}, err
		}
		rec.Image = name
		dropPlainFiles(dir, m.Artifact.Members)
		if err := copyLicences(stage, dir); err != nil {
			log.Warnf("%s install: the licence files were not kept: %v", m.Name, err)
		}
	} else {
		if onData && !ImageSupported() {
			log.Infof("%s install: this slot has no squashfs or no mksquashfs, so the files are placed as plain files", m.Name)
		}
		if err := opt.Plain(stage); err != nil {
			return Result{}, err
		}
		if onData {
			if err := dropImages(m.Name); err != nil {
				log.Warnf("%s install: the old image was not removed: %v", m.Name, err)
			}
		}
	}
	if onData {
		if err := writeRecord(dir, rec); err != nil {
			return Result{}, fmt.Errorf("failed to record the install: %w", err)
		}
	}
	return res, nil
}

func download(ctx context.Context, client *upstream.Client, url string, m Manifest, dst string, progress func(done, total int64)) error {
	if client == nil {
		client = upstream.Default
	}
	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()
	if err := client.Download(ctx, url, m.Artifact.SHA256, m.Artifact.Max, dst, progress); err != nil {
		if errors.Is(err, upstream.ErrChecksum) {
			return fmt.Errorf("the download is not the pinned %s %s, so nothing was installed: %w", m.Title, m.Version, err)
		}
		return fmt.Errorf("failed to download %s %s: %w", m.Title, m.Version, err)
	}
	return nil
}

// writeRecord writes installed.json, keeping the one it replaces beside it
// for Rollback.
func writeRecord(dir string, rec Installed) error {
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, recordFile)
	if rec.Image != "" {
		_ = os.Rename(path, filepath.Join(dir, previousFile))
	} else {
		_ = os.Remove(filepath.Join(dir, previousFile))
	}
	return atomicfile.WriteFile(path, append(data, '\n'), 0o644)
}

// copyLicences replaces dir/licenses with the staged licence files, so the
// web UI can show them without the image mounted.
func copyLicences(stage, dir string) error {
	src := filepath.Join(stage, LicencesDir)
	target := filepath.Join(dir, LicencesDir)
	entries, err := os.ReadDir(src)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	staged := target + ".new"
	_ = os.RemoveAll(staged)
	if err := os.MkdirAll(staged, 0o755); err != nil {
		return err
	}
	for _, e := range entries {
		if !e.Type().IsRegular() {
			continue
		}
		if err := os.Rename(filepath.Join(src, e.Name()), filepath.Join(staged, e.Name())); err != nil {
			_ = os.RemoveAll(staged)
			return err
		}
	}
	_ = os.RemoveAll(target)
	return os.Rename(staged, target)
}

// checkFreeSpace refuses the install before the download when the
// filesystem it lands on cannot hold the download and the files at once, and
// the image as well when there is one. The workspace may not exist yet, so
// the nearest directory that does is asked.
func checkFreeSpace(ws string, m Manifest, image bool) error {
	if m.Artifact.Size <= 0 {
		return nil
	}
	need := uint64(m.Artifact.Size + m.Artifact.Unpacked + spaceMargin)
	if image {
		// LZ4 squashfs of a binary is about half of it; the whole size
		// keeps the estimate safe for files that do not compress.
		need += uint64(m.Artifact.Unpacked)
	}
	dir := ws
	for {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil
		}
		dir = parent
	}
	free, err := FreeBytes(dir)
	if err != nil {
		log.Warnf("%s install: cannot read the free space of %s: %v", m.Name, dir, err)
		return nil
	}
	if free < need {
		return fmt.Errorf("%w on %s: %s %s needs %d MiB, %d MiB are free", ErrNoSpace, dir, m.Title, m.Version, mib(need), mib(free))
	}
	return nil
}

func mib(n uint64) uint64 { return (n + 1<<20 - 1) >> 20 }

func statfsFree(path string) (uint64, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, err
	}
	return st.Bavail * uint64(st.Bsize), nil
}

// artifactName is the name the artifact is saved under: the URL's own, which
// is what apk and a person reading a log expect, when it is a plain name.
func artifactName(url string, m Manifest) string {
	if i := strings.LastIndexByte(url, '/'); i >= 0 {
		if base := url[i+1:]; plainName(base) && !strings.ContainsAny(base, "?#") {
			return base
		}
	}
	return m.Name + "." + string(m.Artifact.Format)
}

// dropPlainFiles removes the members a plain install put in the add-on's
// directory, now that the image holds them. The caller records links into the
// image next; a daemon still running from a removed file keeps it until it
// exits.
func dropPlainFiles(dir string, members []Member) {
	for _, m := range members {
		if m.Licences {
			continue
		}
		p := filepath.Join(dir, m.Name)
		if fi, err := os.Lstat(p); err == nil && fi.Mode().IsRegular() {
			_ = os.Remove(p)
		}
	}
}
