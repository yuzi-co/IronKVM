package webui

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// A UI package is a gzipped tar with two things at its root:
//
//	manifest.json   what the package is (Manifest)
//	web/...         the built UI, as vite and scripts/precompress.mjs leave it
//
// and a minisign signature beside it, <package>.minisig, over the tarball's
// bytes. The signature makes the manifest trustworthy, and the manifest's
// hashes make every unpacked file trustworthy, so the unpacked tree can be
// checked again later without the tarball.
//
// Only directories and regular files are accepted: no links, devices or
// duplicate paths, and nothing outside the two names above.

const (
	// ManifestFormat is the only manifest format this server reads.
	ManifestFormat = 1
	// PackageName is the manifest's name field, and the file name's prefix.
	PackageName = "ironkvm-ui"

	manifestFile = "manifest.json"
	webDir       = "web"
	indexFile    = "index.html"

	maxPackageBytes  = 64 << 20
	maxExpandedBytes = 256 << 20
	maxManifestBytes = 1 << 20
	maxEntries       = 20000
)

var (
	versionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$`)
	// FileNamePattern is what an uploaded or downloaded package may be called.
	FileNamePattern = regexp.MustCompile(`^ironkvm-ui_[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?\.tar\.gz$`)
	sha256Pattern   = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// APIRange is the server API versions a UI supports, both ends included.
type APIRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// Contains reports whether v is in the range.
func (r APIRange) Contains(v int) bool { return r.Min <= v && v <= r.Max }

// Manifest describes a UI package.
type Manifest struct {
	Format  int      `json:"format"`
	Name    string   `json:"name"`
	Version string   `json:"version"`
	API     APIRange `json:"api"`
	// Files maps each file under web/, slash separated, to its SHA-256 in
	// lower-case hex. It lists every file and nothing else.
	Files map[string]string `json:"files"`
}

// Validate checks the manifest on its own, before anything is unpacked.
func (m *Manifest) Validate() error {
	if m.Format != ManifestFormat {
		return fmt.Errorf("unsupported UI manifest format %d", m.Format)
	}
	if m.Name != PackageName {
		return fmt.Errorf("not a UI package: name is %q", m.Name)
	}
	if !versionPattern.MatchString(m.Version) {
		return fmt.Errorf("invalid UI version %q", m.Version)
	}
	if m.API.Min < 1 || m.API.Max < m.API.Min {
		return fmt.Errorf("invalid API range %d..%d", m.API.Min, m.API.Max)
	}
	if len(m.Files) == 0 {
		return errors.New("the manifest lists no files")
	}
	if _, ok := m.Files[indexFile]; !ok {
		return errors.New("the manifest lists no index.html")
	}
	for name, sum := range m.Files {
		if !cleanRelative(name) {
			return fmt.Errorf("invalid file name in manifest: %q", name)
		}
		if !sha256Pattern.MatchString(sum) {
			return fmt.Errorf("invalid hash for %s", name)
		}
	}
	return nil
}

// cleanRelative reports whether name is a plain relative slash path.
func cleanRelative(name string) bool {
	if name == "" || strings.HasPrefix(name, "/") || strings.Contains(name, "\\") {
		return false
	}
	if path.Clean(name) != name {
		return false
	}
	for _, part := range strings.Split(name, "/") {
		if part == "." || part == ".." || part == "" {
			return false
		}
	}
	return true
}

// archiveInfo is what a scan of the tarball learned.
type archiveInfo struct {
	manifest      *Manifest
	expandedBytes uint64
}

// openTar opens a gzipped tar for reading. close closes both layers.
func openTar(pkg string) (*tar.Reader, func(), error) {
	f, err := os.Open(pkg)
	if err != nil {
		return nil, nil, err
	}
	gz, err := gzip.NewReader(f)
	if err != nil {
		_ = f.Close()
		return nil, nil, fmt.Errorf("not a gzip file: %w", err)
	}
	return tar.NewReader(gz), func() { _ = gz.Close(); _ = f.Close() }, nil
}

// entryPath checks one tar header and returns its cleaned name, which is
// manifest.json or a path under web/.
func entryPath(h *tar.Header) (string, error) {
	name := strings.TrimPrefix(h.Name, "./")
	name = strings.TrimSuffix(name, "/")
	if name == "" || name == "." {
		return "", nil
	}
	if !cleanRelative(name) {
		return "", fmt.Errorf("unsafe path in package: %q", h.Name)
	}
	if name != manifestFile && name != webDir && !strings.HasPrefix(name, webDir+"/") {
		return "", fmt.Errorf("unexpected path in package: %q", h.Name)
	}
	switch h.Typeflag {
	case tar.TypeDir:
		if name == manifestFile {
			return "", errors.New("manifest.json is a directory")
		}
	case tar.TypeReg:
		if name == webDir {
			return "", errors.New("web is a file")
		}
	default:
		return "", fmt.Errorf("package entry %q is not a file or a directory", h.Name)
	}
	return name, nil
}

// inspectPackage reads the whole tarball once, without writing anything, and
// returns the manifest and the unpacked size.
func inspectPackage(pkg string) (*archiveInfo, error) {
	tr, closeTar, err := openTar(pkg)
	if err != nil {
		return nil, err
	}
	defer closeTar()

	info := &archiveInfo{}
	seen := map[string]bool{}
	entries := 0
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read package: %w", err)
		}
		entries++
		if entries > maxEntries {
			return nil, fmt.Errorf("the package has more than %d entries", maxEntries)
		}
		name, err := entryPath(h)
		if err != nil {
			return nil, err
		}
		if name == "" {
			continue
		}
		if seen[name] {
			return nil, fmt.Errorf("duplicate path in package: %q", name)
		}
		seen[name] = true
		if h.Typeflag != tar.TypeReg {
			continue
		}
		if h.Size < 0 {
			return nil, fmt.Errorf("bad size for %q", name)
		}
		info.expandedBytes += uint64(h.Size)
		if info.expandedBytes > maxExpandedBytes {
			return nil, fmt.Errorf("the package unpacks to more than %d bytes", maxExpandedBytes)
		}
		if name == manifestFile {
			if h.Size > maxManifestBytes {
				return nil, errors.New("manifest.json is too large")
			}
			var m Manifest
			dec := json.NewDecoder(io.LimitReader(tr, maxManifestBytes))
			dec.DisallowUnknownFields()
			if err := dec.Decode(&m); err != nil {
				return nil, fmt.Errorf("read manifest.json: %w", err)
			}
			info.manifest = &m
		}
	}
	if info.manifest == nil {
		return nil, errors.New("the package has no manifest.json")
	}
	if err := info.manifest.Validate(); err != nil {
		return nil, err
	}
	return info, nil
}

// extractPackage unpacks a package that inspectPackage accepted into dest,
// which must exist and be empty.
func extractPackage(pkg, dest string) error {
	tr, closeTar, err := openTar(pkg)
	if err != nil {
		return err
	}
	defer closeTar()

	// Every write goes through root, which refuses a name that leaves dest,
	// whether by "..", an absolute path or a symbolic link already on disk.
	// entryPath refuses such names first; root holds if it is ever wrong.
	root, err := os.OpenRoot(dest)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()

	var written uint64
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read package: %w", err)
		}
		name, err := entryPath(h)
		if err != nil {
			return err
		}
		if name == "" {
			continue
		}
		target := filepath.FromSlash(name)
		if h.Typeflag == tar.TypeDir {
			if err := root.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := root.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		out, err := root.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return err
		}
		n, err := io.Copy(out, io.LimitReader(tr, maxExpandedBytes-int64(written)+1))
		closeErr := out.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		written += uint64(n)
		if written > maxExpandedBytes {
			return fmt.Errorf("the package unpacks to more than %d bytes", maxExpandedBytes)
		}
	}
}

// readManifest reads an unpacked release's manifest.
func readManifest(releaseDir string) (*Manifest, error) {
	data, err := os.ReadFile(filepath.Join(releaseDir, manifestFile))
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("read manifest.json: %w", err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

// verifyTree holds the unpacked web/ against the manifest: every listed file
// present with its hash, and no file the manifest does not list.
func verifyTree(webPath string, m *Manifest) error {
	found := 0
	err := filepath.WalkDir(webPath, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("%s is not a regular file", p)
		}
		rel, err := filepath.Rel(webPath, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		want, ok := m.Files[rel]
		if !ok {
			return fmt.Errorf("%s is not in the manifest", rel)
		}
		got, err := fileSHA256(p)
		if err != nil {
			return err
		}
		if got != want {
			return fmt.Errorf("%s does not match its hash in the manifest", rel)
		}
		found++
		return nil
	})
	if err != nil {
		return err
	}
	if found != len(m.Files) {
		return fmt.Errorf("the manifest lists %d files and the package has %d", len(m.Files), found)
	}
	return nil
}

func fileSHA256(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

var assetReference = regexp.MustCompile(`(?:src|href)="/?(assets/[^"?#]+)"`)

// basicCheck is the check a release must pass before it is served, and again
// whenever the server starts: it states a server API this server speaks, its
// index.html is the one the manifest names, and every asset index.html loads
// is present with its hash. It reads index.html and its few direct assets, not
// the whole tree, so it is cheap enough for every start.
func basicCheck(releaseDir string, api int) (*Manifest, error) {
	m, err := readManifest(releaseDir)
	if err != nil {
		return nil, err
	}
	if !m.API.Contains(api) {
		return m, fmt.Errorf("UI %s supports server API %d..%d, and this server speaks %d", m.Version, m.API.Min, m.API.Max, api)
	}
	webPath := filepath.Join(releaseDir, webDir)
	index, err := os.ReadFile(filepath.Join(webPath, indexFile))
	if err != nil {
		return m, fmt.Errorf("no index.html: %w", err)
	}
	sum := sha256.Sum256(index)
	if hex.EncodeToString(sum[:]) != m.Files[indexFile] {
		return m, errors.New("index.html does not match its hash in the manifest")
	}
	for _, match := range assetReference.FindAllStringSubmatch(string(index), -1) {
		rel := match[1]
		want, ok := m.Files[rel]
		if !ok {
			return m, fmt.Errorf("index.html loads %s, which the manifest does not list", rel)
		}
		got, err := fileSHA256(filepath.Join(webPath, filepath.FromSlash(rel)))
		if err != nil {
			return m, fmt.Errorf("index.html loads %s: %w", rel, err)
		}
		if got != want {
			return m, fmt.Errorf("%s does not match its hash in the manifest", rel)
		}
	}
	return m, nil
}
