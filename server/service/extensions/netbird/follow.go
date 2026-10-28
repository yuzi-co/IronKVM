package netbird

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	installer "NanoKVM-Server/service/addons/install"
	"NanoKVM-Server/service/upstream"

	log "github.com/sirupsen/logrus"
)

// Alpine keeps one build of a package per branch, so when v3.24 community
// takes a newer NetBird the pinned file leaves the mirror (ironkvm-dist #174).
// The install then follows the branch: only when the pinned URL answers 404,
// it reads the branch's index, APKINDEX.tar.gz, and installs the netbird
// build the index names, under these checks, any of which refuses it:
//
//   - apk verifies Alpine's signature on the index against the riscv64 keys
//     in /etc/apk/keys, and the index must be one signature segment and one
//     index segment, with nothing appended that the signature would not cover;
//   - the entry must be netbird for riscv64, at a version no older than the
//     pin, so a stale index cannot downgrade the add-on, and within the size
//     bounds below;
//   - the package must be the size the index gives, and its control segment
//     must hash to the index's C: checksum, which binds the download to the
//     signed index entry; its .PKGINFO must name netbird, riscv64 and the
//     index's version;
//   - apk verifies Alpine's signature on the package itself, as it does the
//     pin's, before anything is taken out of it.
//
// installed.json then records the version, the package's URL and its SHA-256
// as downloaded, and the aports tree of the index's c: commit as the source.
// Every other failure of the pin, such as no network or a sum that does not
// match, stays a failure: only a file that is gone is followed.

const (
	// indexFile is the index's name in the branch's directory.
	indexFile = "APKINDEX.tar.gz"
	// indexMax bounds the index download. v3.24 community riscv64's was
	// 2 MB on 2026-10-28.
	indexMax = 16 << 20
	// indexTextMax bounds APKINDEX inside it, decompressed; 8.8 MB then.
	indexTextMax = 64 << 20
	// installedMax bounds the entry's installed size (I:); 0.81.2-r0's is
	// 43.5 MB.
	installedMax = 96 << 20
	// pkginfoMax bounds the package's .PKGINFO.
	pkginfoMax = 64 << 10
	// arch is the board's Alpine architecture.
	arch = "riscv64"
	// aportsTree is the source of a build, by aports commit.
	aportsTree = "https://gitlab.alpinelinux.org/alpine/aports/-/tree/%s/community/netbird"
)

// indexURL is the branch's index. The package a followed entry names sits
// beside it. A variable only so the tests can serve their own.
var indexURL = "https://dl-cdn.alpinelinux.org/alpine/v3.24/community/riscv64/" + indexFile

// indexEntry is the part of an APKINDEX record the install uses.
type indexEntry struct {
	Name      string // P:
	Release   string // V:, such as 0.81.2-r0
	Arch      string // A:
	Size      int64  // S:, the package file's size
	Installed int64  // I:, the installed size
	Checksum  string // C:, Q1 and the base64 SHA-1 of the control segment
	Commit    string // c:, the aports commit it was built from
}

// Version is the upstream version, as `netbird version` reports it.
func (e indexEntry) Version() string {
	v, _, _ := strings.Cut(e.Release, "-r")
	return v
}

// File is the package's file name in the branch's directory.
func (e indexEntry) File() string { return e.Name + "-" + e.Release + ".apk" }

var (
	releasePattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)*-r[0-9]+$`)
	commitPattern  = regexp.MustCompile(`^[0-9a-f]{40}$`)
)

// compareReleases compares two Alpine versions of the form 1.2.3-r4: the
// upstream version first, then the release number.
func compareReleases(a, b string) int {
	av, ar, _ := strings.Cut(a, "-r")
	bv, br, _ := strings.Cut(b, "-r")
	if c := upstream.CompareVersions(av, bv); c != 0 {
		return c
	}
	x, _ := strconv.Atoi(ar)
	y, _ := strconv.Atoi(br)
	switch {
	case x < y:
		return -1
	case x > y:
		return 1
	}
	return 0
}

// check refuses an entry the install must not follow.
func (e indexEntry) check() error {
	if e.Name != release.Name {
		return fmt.Errorf("the index entry is %q, not %s", e.Name, release.Name)
	}
	if e.Arch != arch {
		return fmt.Errorf("the index's %s is for %q, not %s", e.Name, e.Arch, arch)
	}
	if !releasePattern.MatchString(e.Release) {
		return fmt.Errorf("the index's %s has the version %q, which is not of the form 1.2.3-r0", e.Name, e.Release)
	}
	if compareReleases(e.Release, pinRelease) < 0 {
		return fmt.Errorf("the index offers %s %s, older than the pinned %s, and a downgrade is refused", e.Name, e.Release, pinRelease)
	}
	if e.Size <= 0 || e.Size > release.Artifact.Max {
		return fmt.Errorf("the index gives %s %s a size of %d bytes, outside 1 to %d", e.Name, e.Release, e.Size, release.Artifact.Max)
	}
	if e.Installed <= 0 || e.Installed > installedMax {
		return fmt.Errorf("the index gives %s %s an installed size of %d bytes, outside 1 to %d", e.Name, e.Release, e.Installed, installedMax)
	}
	if _, err := q1(e.Checksum); err != nil {
		return fmt.Errorf("the index's %s %s: %w", e.Name, e.Release, err)
	}
	if !commitPattern.MatchString(e.Commit) {
		return fmt.Errorf("the index's %s %s names no aports commit", e.Name, e.Release)
	}
	return nil
}

// q1 decodes a C: checksum. Alpine's v2 packages have Q1 checksums, a SHA-1;
// any other kind is refused rather than left unchecked.
func q1(c string) ([]byte, error) {
	b64, ok := strings.CutPrefix(c, "Q1")
	if !ok {
		return nil, fmt.Errorf("the checksum %q is not a Q1 (SHA-1) checksum", c)
	}
	sum, err := base64.StdEncoding.DecodeString(b64)
	if err != nil || len(sum) != sha1.Size {
		return nil, fmt.Errorf("the checksum %q is not a SHA-1", c)
	}
	return sum, nil
}

// lookupIndex reads the branch's index and returns its checked netbird entry.
// A variable so the tests of the update offer need no network.
var lookupIndex = fetchIndex

// fetchIndex downloads the index to a scratch directory, has apk verify its
// signature, and reads the netbird entry out of it.
func fetchIndex(ctx context.Context) (indexEntry, error) {
	dir, err := os.MkdirTemp("", "netbird-index-")
	if err != nil {
		return indexEntry{}, err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	path := filepath.Join(dir, indexFile)
	if err := downloads.Fetch(ctx, indexURL, indexMax, path, nil); err != nil {
		return indexEntry{}, fmt.Errorf("failed to download Alpine's index: %w", err)
	}
	if err := installer.VerifyApk(path, ApkPath); err != nil {
		return indexEntry{}, fmt.Errorf("Alpine's index failed its signature check: %w", err)
	}
	e, err := readIndex(path)
	if err != nil {
		return indexEntry{}, err
	}
	if err := e.check(); err != nil {
		return indexEntry{}, err
	}
	remember(e)
	return e, nil
}

// readIndex reads the netbird entry out of an APKINDEX.tar.gz: a gzip member
// holding only the signature, then one holding DESCRIPTION and APKINDEX, the
// one the signature covers, and nothing after it.
func readIndex(path string) (indexEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return indexEntry{}, err
	}
	defer func() { _ = f.Close() }()
	segs := newSegments(f)

	if err := segs.next(); err != nil {
		return indexEntry{}, fmt.Errorf("the index has no signature: %w", err)
	}
	if err := signatureOnly(segs.zr); err != nil {
		return indexEntry{}, fmt.Errorf("the index's signature segment: %w", err)
	}
	if err := segs.next(); err != nil {
		return indexEntry{}, fmt.Errorf("the index has no index segment: %w", err)
	}
	var entry indexEntry
	found := false
	tr := tar.NewReader(segs.zr)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return indexEntry{}, fmt.Errorf("the index is not a tar: %w", err)
		}
		if h.Name != "APKINDEX" {
			continue
		}
		if found {
			return indexEntry{}, errors.New("the index holds two APKINDEX files")
		}
		found = true
		if entry, err = findEntry(io.LimitReader(tr, indexTextMax+1), release.Name); err != nil {
			return indexEntry{}, err
		}
	}
	if !found {
		return indexEntry{}, errors.New("the index holds no APKINDEX")
	}
	if _, err := io.Copy(io.Discard, segs.zr); err != nil {
		return indexEntry{}, err
	}
	if !segs.atEnd() {
		return indexEntry{}, errors.New("the index has data after its signed segment")
	}
	return entry, nil
}

// signatureOnly drains a signature segment, refusing anything in it but
// .SIGN. files.
func signatureOnly(r io.Reader) error {
	tr := tar.NewReader(r)
	n := 0
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if !strings.HasPrefix(h.Name, ".SIGN.") {
			return fmt.Errorf("it holds %s, which is not a signature", h.Name)
		}
		n++
	}
	if n == 0 {
		return errors.New("it holds no signature")
	}
	_, err := io.Copy(io.Discard, r)
	return err
}

// findEntry reads APKINDEX records, separated by blank lines, and returns the
// one record of package name. Two records of it are refused, as an index
// that does not say which build it means.
func findEntry(r io.Reader, name string) (indexEntry, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64<<10), 1<<20)
	var cur, found indexEntry
	var n int64
	matches := 0
	end := func() {
		if cur.Name == name {
			found = cur
			matches++
		}
		cur = indexEntry{}
	}
	for sc.Scan() {
		line := sc.Text()
		n += int64(len(line)) + 1
		if n > indexTextMax {
			return indexEntry{}, fmt.Errorf("APKINDEX is larger than %d bytes", indexTextMax)
		}
		if line == "" {
			end()
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		switch key {
		case "P":
			cur.Name = value
		case "V":
			cur.Release = value
		case "A":
			cur.Arch = value
		case "S":
			cur.Size, _ = strconv.ParseInt(value, 10, 64)
		case "I":
			cur.Installed, _ = strconv.ParseInt(value, 10, 64)
		case "C":
			cur.Checksum = value
		case "c":
			cur.Commit = value
		}
	}
	if err := sc.Err(); err != nil {
		return indexEntry{}, fmt.Errorf("failed to read APKINDEX: %w", err)
	}
	end()
	switch {
	case matches == 0:
		return indexEntry{}, fmt.Errorf("Alpine's index has no %s", name)
	case matches > 1:
		return indexEntry{}, fmt.Errorf("Alpine's index has %d entries for %s", matches, name)
	}
	return found, nil
}

// segments walks the gzip members an Alpine v2 file is made of, counting the
// compressed bytes each takes. gzip reads a reader that has ReadByte without
// buffering it, so the count ends exactly where a member does.
type segments struct {
	r     *countingReader
	zr    *gzip.Reader
	start int64
}

func newSegments(r io.Reader) *segments {
	return &segments{r: &countingReader{r: bufio.NewReader(r)}}
}

// next starts the next member; start is its first byte's offset.
func (s *segments) next() error {
	s.start = s.r.n
	var err error
	if s.zr == nil {
		s.zr, err = gzip.NewReader(s.r)
	} else {
		err = s.zr.Reset(s.r)
	}
	if err != nil {
		return err
	}
	s.zr.Multistream(false)
	return nil
}

// end is the offset just past the current member, once it is drained.
func (s *segments) end() int64 { return s.r.n }

// atEnd reports whether nothing follows the current member.
func (s *segments) atEnd() bool {
	_, err := s.r.ReadByte()
	return err == io.EOF
}

type countingReader struct {
	r *bufio.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

func (c *countingReader) ReadByte() (byte, error) {
	b, err := c.r.ReadByte()
	if err == nil {
		c.n++
	}
	return b, err
}

// matchEntry checks a downloaded package against its index entry: its size,
// its control segment's SHA-1 against C:, and what its .PKGINFO says it is.
// It returns the package's SHA-256. The signature is apk's to check, after.
func matchEntry(path string, e indexEntry) (string, error) {
	want, err := q1(e.Checksum)
	if err != nil {
		return "", err
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	fi, err := f.Stat()
	if err != nil {
		return "", err
	}
	if fi.Size() != e.Size {
		return "", fmt.Errorf("%s is %d bytes, and the index says %d", e.File(), fi.Size(), e.Size)
	}

	segs := newSegments(f)
	if err := segs.next(); err != nil {
		return "", fmt.Errorf("%s has no signature segment: %w", e.File(), err)
	}
	if err := signatureOnly(segs.zr); err != nil {
		return "", fmt.Errorf("%s's signature segment: %w", e.File(), err)
	}
	if err := segs.next(); err != nil {
		return "", fmt.Errorf("%s has no control segment: %w", e.File(), err)
	}
	start := segs.start
	info, err := readPkginfo(segs.zr)
	if err != nil {
		return "", fmt.Errorf("%s: %w", e.File(), err)
	}
	if _, err := io.Copy(io.Discard, segs.zr); err != nil {
		return "", err
	}
	control := sha1.New()
	if _, err := io.Copy(control, io.NewSectionReader(f, start, segs.end()-start)); err != nil {
		return "", err
	}
	if got := control.Sum(nil); !bytes.Equal(got, want) {
		return "", fmt.Errorf("%s is not the package the index lists: its checksum is Q1%s, the index's %s",
			e.File(), base64.StdEncoding.EncodeToString(got), e.Checksum)
	}
	if info["pkgname"] != e.Name || info["arch"] != e.Arch || info["pkgver"] != e.Release {
		return "", fmt.Errorf("%s says it is %s %s for %s, not %s %s for %s",
			e.File(), info["pkgname"], info["pkgver"], info["arch"], e.Name, e.Release, e.Arch)
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	sum := sha256.New()
	if _, err := io.Copy(sum, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(sum.Sum(nil)), nil
}

// readPkginfo reads .PKGINFO's "key = value" lines out of a control segment.
func readPkginfo(r io.Reader) (map[string]string, error) {
	tr := tar.NewReader(r)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return nil, errors.New("the control segment has no .PKGINFO")
		}
		if err != nil {
			return nil, fmt.Errorf("the control segment is not a tar: %w", err)
		}
		if h.Name != ".PKGINFO" {
			continue
		}
		if h.Size > pkginfoMax {
			return nil, fmt.Errorf(".PKGINFO is %d bytes", h.Size)
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		info := map[string]string{}
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "#") {
				continue
			}
			if k, v, ok := strings.Cut(line, " = "); ok {
				if _, dup := info[k]; !dup {
					info[k] = v
				}
			}
		}
		return info, nil
	}
}

// followManifest is the pin's manifest moved to the index's build. The sum is
// left for the fetch to fill in.
func followManifest(e indexEntry) installer.Manifest {
	m := release
	m.Version = e.Version()
	m.Source = fmt.Sprintf(aportsTree, e.Commit)
	m.Artifact.URL = strings.TrimSuffix(indexURL, indexFile) + e.File()
	m.Artifact.SHA256 = ""
	m.Artifact.Size = e.Size
	m.Artifact.Unpacked = e.Installed
	m.Artifact.Max = e.Size
	return m
}

// installFollowing installs the build the branch's index names in place of
// the pin.
func installFollowing(ctx context.Context, opt installer.Options) (installer.Manifest, installer.Result, error) {
	e, err := lookupIndex(ctx)
	if err != nil {
		return installer.Manifest{}, installer.Result{}, err
	}
	m := followManifest(e)
	log.Infof("netbird install: the pinned %s is gone from Alpine's mirror, so %s from v3.24's index is installed", packageFile, e.File())
	opt.Fetch = func(ctx context.Context, dst string, progress func(done, total int64)) (string, error) {
		if err := downloads.Fetch(ctx, m.Artifact.URL, e.Size, dst, progress); err != nil {
			return "", err
		}
		return matchEntry(dst, e)
	}
	res, err := installer.Install(ctx, m, opt)
	return m, res, err
}

// The update offer looks at the index now and then, in the background, so the
// page that asks never waits on the download.
const indexCheckEvery = 6 * time.Hour

var indexCheck struct {
	sync.Mutex
	entry   indexEntry
	ok      bool
	at      time.Time
	running bool
}

// remember keeps an entry the index check or an install read.
func remember(e indexEntry) {
	indexCheck.Lock()
	defer indexCheck.Unlock()
	indexCheck.entry, indexCheck.ok, indexCheck.at = e, true, time.Now()
}

// knownIndex is the netbird entry of the last index read, if any.
func knownIndex() (indexEntry, bool) {
	indexCheck.Lock()
	defer indexCheck.Unlock()
	return indexCheck.entry, indexCheck.ok
}

// checkIndexSoon reads the index in the background when the last look is
// older than indexCheckEvery. A failed look waits as long as a good one, so an
// offline board does not download on every visit.
func checkIndexSoon() {
	indexCheck.Lock()
	if indexCheck.running || !indexCheck.at.IsZero() && time.Since(indexCheck.at) < indexCheckEvery {
		indexCheck.Unlock()
		return
	}
	indexCheck.running = true
	indexCheck.Unlock()

	lookup := lookupIndex
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		_, err := lookup(ctx)

		indexCheck.Lock()
		defer indexCheck.Unlock()
		indexCheck.running = false
		indexCheck.at = time.Now()
		if err != nil {
			log.Warnf("netbird: check Alpine's index for a newer build: %s", err)
		}
	}()
}

// latestRelease is the newest NetBird an update can install: the pin, or the
// index's build when the last look found a newer one.
func latestRelease() (version string, fromIndex bool) {
	if e, ok := knownIndex(); ok && upstream.CompareVersions(e.Version(), Version) > 0 {
		return e.Version(), true
	}
	return Version, false
}
