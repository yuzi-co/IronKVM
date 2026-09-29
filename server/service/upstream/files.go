package upstream

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// checksumsMax bounds a checksum file. netboot.xyz's lists some sixty files
// in 4 KB.
const checksumsMax = 256 << 10

// Checksums maps a file name to its SHA-256, in lower-case hex.
type Checksums map[string]string

// ParseChecksums reads a checksum file in sha256sum's format: a sum, a space,
// a space or a "*" for binary mode, and the name. Blank lines and comments
// are skipped, and so is a line that is not a SHA-256.
func ParseChecksums(r io.Reader) (Checksums, error) {
	sums := Checksums{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		sum, name, ok := strings.Cut(line, " ")
		if !ok {
			continue
		}
		name = strings.TrimPrefix(strings.TrimLeft(name, " "), "*")
		sum = strings.ToLower(sum)
		if b, err := hex.DecodeString(sum); err != nil || len(b) != sha256.Size || name == "" {
			continue
		}
		if prev, dup := sums[name]; dup && prev != sum {
			return nil, fmt.Errorf("the checksum file lists %s twice with different sums", name)
		}
		sums[name] = sum
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(sums) == 0 {
		return nil, errors.New("the checksum file lists no SHA-256 sums")
	}
	return sums, nil
}

// get starts a download over HTTPS. Only https is followed, redirects
// included: the release's files are trusted because the checksum file came
// over the same kind of connection.
func (c *Client) get(ctx context.Context, rawURL string) (*http.Response, error) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" {
		return nil, fmt.Errorf("refusing to download %s: not an https URL", rawURL)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	httpClient := *c.HTTP
	httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if req.URL.Scheme != "https" {
			return fmt.Errorf("refusing a redirect to %s", req.URL)
		}
		if len(via) >= 10 {
			return errors.New("too many redirects")
		}
		return nil
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, offline(err)
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("download %s: status %d", rawURL, resp.StatusCode)
	}
	return resp, nil
}

// FetchChecksums downloads and parses a release's checksum file.
func (c *Client) FetchChecksums(ctx context.Context, rawURL string) (Checksums, error) {
	resp, err := c.get(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	return ParseChecksums(io.LimitReader(resp.Body, checksumsMax))
}

// Download writes rawURL to dst, and keeps it only when its SHA-256 is sum.
// It reads at most limit bytes, and reports the bytes written so far and the
// size the server announced, or -1, to progress, which may be nil.
func (c *Client) Download(ctx context.Context, rawURL, sum string, limit int64, dst string, progress func(done, total int64)) error {
	if b, err := hex.DecodeString(sum); err != nil || len(b) != sha256.Size {
		return fmt.Errorf("no valid SHA-256 for %s", filepath.Base(dst))
	}

	resp, err := c.get(ctx, rawURL)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.ContentLength > limit {
		return fmt.Errorf("download %s: %d bytes, larger than the %d this file may have", rawURL, resp.ContentLength, limit)
	}

	tmp := dst + ".part"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp) }()

	hasher := sha256.New()
	w := io.MultiWriter(out, hasher)
	if progress != nil {
		w = &progressWriter{w: w, total: resp.ContentLength, report: progress}
	}
	n, copyErr := io.Copy(w, io.LimitReader(resp.Body, limit+1))
	closeErr := out.Close()
	if copyErr != nil {
		return fmt.Errorf("download %s: %w", rawURL, copyErr)
	}
	if closeErr != nil {
		return closeErr
	}
	if n > limit {
		return fmt.Errorf("download %s: larger than %d bytes", rawURL, limit)
	}
	if got := hex.EncodeToString(hasher.Sum(nil)); got != strings.ToLower(sum) {
		return fmt.Errorf("%w: %s has sha256 %s, the release's checksum file says %s", ErrChecksum, filepath.Base(dst), got, sum)
	}
	return os.Rename(tmp, dst)
}

// ErrChecksum wraps a download whose sum is not the published one.
var ErrChecksum = errors.New("checksum mismatch")

type progressWriter struct {
	w      io.Writer
	done   int64
	total  int64
	report func(done, total int64)
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n, err := p.w.Write(b)
	p.done += int64(n)
	p.report(p.done, p.total)
	return n, err
}

// Swap moves each named file from staged into dir, keeping the one it
// replaces as name.old until every file is in place and commit, which may be
// nil, has succeeded. On any failure the files already moved are put back,
// so dir holds either the old set or the new one, never a mix.
func Swap(staged, dir string, names []string, commit func() error) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	type moved struct {
		name   string
		hadOld bool
	}
	var done []moved
	rollback := func() {
		for i := len(done) - 1; i >= 0; i-- {
			dst := filepath.Join(dir, done[i].name)
			_ = os.Remove(dst)
			if done[i].hadOld {
				_ = os.Rename(dst+".old", dst)
			}
		}
	}

	for _, name := range names {
		src := filepath.Join(staged, name)
		dst := filepath.Join(dir, name)
		if _, err := os.Stat(src); err != nil {
			rollback()
			return fmt.Errorf("staged %s: %w", name, err)
		}
		_ = os.Remove(dst + ".old")
		hadOld := false
		if err := os.Rename(dst, dst+".old"); err == nil {
			hadOld = true
		} else if !errors.Is(err, os.ErrNotExist) {
			rollback()
			return err
		}
		if err := os.Rename(src, dst); err != nil {
			if hadOld {
				_ = os.Rename(dst+".old", dst)
			}
			rollback()
			return err
		}
		done = append(done, moved{name: name, hadOld: hadOld})
	}

	if commit != nil {
		if err := commit(); err != nil {
			rollback()
			return err
		}
	}
	for _, m := range done {
		_ = os.Remove(filepath.Join(dir, m.name) + ".old")
	}
	return nil
}
