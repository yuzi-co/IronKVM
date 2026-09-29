package ventoy

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"NanoKVM-Server/utils"
)

// Variables rather than constants so the tests can point them at a scratch
// directory and a stand-in for xzcat.
var (
	// Dir holds the release's files, the head and the state. It is on
	// /data, and its files have no image suffix, so the image list skips
	// them.
	Dir = "/data/ironkvm/ventoy"
	// XzcatPath decompresses a member. The image has xzcat.
	XzcatPath = "xzcat"
	// httpClient fetches the archive. No overall timeout, as the download
	// service has none: these bound a server that never answers.
	httpClient = &http.Client{
		Transport: &http.Transport{
			Proxy:                 utils.ProxyFromConfig,
			TLSHandshakeTimeout:   30 * time.Second,
			ResponseHeaderTimeout: 60 * time.Second,
			IdleConnTimeout:       90 * time.Second,
		},
	}
)

const fetchTimeout = 10 * time.Minute

func inDir(name string) string { return filepath.Join(Dir, name) }

// installed says whether every file the disk needs from the release is in
// place.
func installed() bool {
	for _, m := range releaseMembers {
		fi, err := os.Stat(inDir(m.Name))
		if err != nil || !fi.Mode().IsRegular() {
			return false
		}
	}
	return true
}

// install downloads the pinned release and keeps the three files the disk
// needs, each checked against its own sum.
func install() error {
	ws := inDir(".fetch")
	_ = os.RemoveAll(ws)
	if err := os.MkdirAll(ws, 0o755); err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(ws) }()

	archive := filepath.Join(ws, "ventoy.tar.gz")
	if err := fetchPinned(releaseArchive, archive); err != nil {
		return err
	}
	return extractMembers(archive, releaseMembers, ws, Dir)
}

// uninstall removes the release's files, the head and the record of an
// update, so a reinstall is the pinned release. The state stays, so the
// owner's selection survives a reinstall.
func uninstall() error {
	names := []string{headName, releaseName}
	for _, m := range releaseMembers {
		names = append(names, m.Name)
	}
	for _, name := range names {
		if err := os.Remove(inDir(name)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

// fetchPinned downloads f to dst, and keeps it only when its SHA-256 is the
// pinned one.
func fetchPinned(f pinnedFile, dst string) error {
	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.URL, nil)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", f.URL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: status %d", f.URL, resp.StatusCode)
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	hasher := sha256.New()
	n, copyErr := io.Copy(io.MultiWriter(out, hasher), io.LimitReader(resp.Body, f.Max+1))
	closeErr := out.Close()
	if copyErr != nil {
		return fmt.Errorf("download %s: %w", f.URL, copyErr)
	}
	if closeErr != nil {
		return closeErr
	}
	if n > f.Max {
		return fmt.Errorf("download %s: larger than %d bytes", f.URL, f.Max)
	}
	if got := hex.EncodeToString(hasher.Sum(nil)); got != f.SHA256 {
		return fmt.Errorf("download %s: sha256 %s, want %s", f.URL, got, f.SHA256)
	}
	return nil
}

// extractMembers takes the named members out of the archive into ws,
// decompresses the xz ones, checks each result and moves it into dir.
// Nothing else in the archive is written, so a member's path never becomes
// a path on the board. A member the archive lacks is an error.
func extractMembers(archive string, members []releaseMember, ws, dir string) error {
	if _, err := extractChecked(archive, members, ws); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, m := range members {
		if err := os.Rename(filepath.Join(ws, m.Name), filepath.Join(dir, m.Name)); err != nil {
			return err
		}
	}
	return nil
}

// extractChecked takes the named members out of the archive into ws under
// their saved names, decompressed and checked, and returns their sums. A
// member with no pinned sum, from a release newer than the pin, is checked
// by the archive's sum alone.
func extractChecked(archive string, members []releaseMember, ws string) (map[string]string, error) {
	f, err := os.Open(archive)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	want := map[string]releaseMember{}
	for _, m := range members {
		want[m.Member] = m
	}

	raw := map[string]string{}
	tr := tar.NewReader(gz)
	for len(raw) < len(members) {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		m, ok := want[hdr.Name]
		if !ok || hdr.Typeflag != tar.TypeReg {
			continue
		}
		if hdr.Size > memberMax {
			return nil, fmt.Errorf("%s: larger than %d bytes", m.Member, memberMax)
		}
		tmp := filepath.Join(ws, m.Name+".raw")
		out, err := os.Create(tmp)
		if err != nil {
			return nil, err
		}
		_, copyErr := io.Copy(out, io.LimitReader(tr, hdr.Size))
		closeErr := out.Close()
		if copyErr != nil {
			return nil, copyErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		raw[m.Member] = tmp
	}

	sums := map[string]string{}
	for _, m := range members {
		src, ok := raw[m.Member]
		if !ok {
			return nil, fmt.Errorf("the release has no %s", m.Member)
		}
		sum, err := writeChecked(src, filepath.Join(ws, m.Name), m)
		if err != nil {
			return nil, err
		}
		_ = os.Remove(src)
		sums[m.Name] = sum
	}
	return sums, nil
}

// writeChecked writes the member at src to dst, decompressed when it is xz,
// and keeps it only when its sum is the pinned one, if it has one. It
// returns the sum.
func writeChecked(src, dst string, m releaseMember) (string, error) {
	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	hasher := sha256.New()
	w := io.MultiWriter(out, hasher)

	var copyErr error
	if m.XZ {
		ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, XzcatPath, src)
		cmd.Stdout = &limitWriter{w: w, left: memberMax}
		var stderr limitBuffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			copyErr = fmt.Errorf("decompress %s: %w: %s", m.Member, err, stderr.String())
		}
	} else {
		in, err := os.Open(src)
		if err != nil {
			_ = out.Close()
			return "", err
		}
		_, copyErr = io.Copy(w, in)
		_ = in.Close()
	}
	closeErr := out.Close()
	if copyErr != nil {
		return "", copyErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	got := hex.EncodeToString(hasher.Sum(nil))
	if m.SHA256 != "" && got != m.SHA256 {
		return "", fmt.Errorf("%s: sha256 %s, want %s", m.Member, got, m.SHA256)
	}
	return got, nil
}

// limitWriter fails a write past its limit, so a member that decompresses
// to more than any pinned one can hold stops early.
type limitWriter struct {
	w    io.Writer
	left int64
}

func (l *limitWriter) Write(p []byte) (int, error) {
	if int64(len(p)) > l.left {
		return 0, fmt.Errorf("larger than %d bytes", memberMax)
	}
	l.left -= int64(len(p))
	return l.w.Write(p)
}

// limitBuffer keeps the first 4 KB of a command's error output.
type limitBuffer struct{ buf []byte }

func (b *limitBuffer) Write(p []byte) (int, error) {
	if room := 4096 - len(b.buf); room > 0 {
		b.buf = append(b.buf, p[:min(room, len(p))]...)
	}
	return len(p), nil
}

func (b *limitBuffer) String() string { return string(b.buf) }
