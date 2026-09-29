package upstream

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func hexSum(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// fakeGitHub serves one repository's latest release and its assets over TLS.
type fakeGitHub struct {
	srv     *httptest.Server
	lookups atomic.Int32
	tag     string
	status  int
	assets  map[string][]byte
}

func newFakeGitHub(t *testing.T, tag string, assets map[string][]byte) (*fakeGitHub, *Client) {
	t.Helper()
	f := &fakeGitHub{tag: tag, status: http.StatusOK, assets: assets}
	f.srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/project/releases/latest" {
			f.lookups.Add(1)
			if f.status != http.StatusOK {
				w.Header().Set("X-RateLimit-Reset", "0")
				w.WriteHeader(f.status)
				return
			}
			type asset struct {
				Name string `json:"name"`
				URL  string `json:"browser_download_url"`
			}
			rel := struct {
				Tag    string  `json:"tag_name"`
				Assets []asset `json:"assets"`
			}{Tag: f.tag}
			for name := range f.assets {
				rel.Assets = append(rel.Assets, asset{name, f.srv.URL + "/download/" + name})
			}
			_ = json.NewEncoder(w).Encode(rel)
			return
		}
		if name, ok := strings.CutPrefix(r.URL.Path, "/download/"); ok {
			if data, ok := f.assets[name]; ok {
				_, _ = w.Write(data)
				return
			}
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(f.srv.Close)
	return f, NewClient(f.srv.URL, f.srv.Client())
}

func TestLatestReadsTheReleaseAndCachesIt(t *testing.T) {
	f, c := newFakeGitHub(t, "v1.2.3", map[string][]byte{"a.bin": []byte("a")})
	now := time.Unix(1_000_000, 0)
	c.Now = func() time.Time { return now }

	if _, _, _, ok := c.Cached("owner/project"); ok {
		t.Fatal("a cache before any lookup")
	}
	rel, err := c.Latest(context.Background(), "owner/project", false)
	if err != nil {
		t.Fatal(err)
	}
	if rel.Tag != "v1.2.3" || rel.Version != "1.2.3" || rel.Assets["a.bin"] != f.srv.URL+"/download/a.bin" {
		t.Fatalf("release %+v", rel)
	}

	if _, err := c.Latest(context.Background(), "owner/project", false); err != nil || f.lookups.Load() != 1 {
		t.Fatalf("a fresh answer was asked for again: %d lookups, %v", f.lookups.Load(), err)
	}
	if _, err := c.Latest(context.Background(), "owner/project", true); err != nil || f.lookups.Load() != 2 {
		t.Fatalf("force did not ask again: %d lookups", f.lookups.Load())
	}
	now = now.Add(CacheTTL + time.Second)
	if _, err := c.Latest(context.Background(), "owner/project", false); err != nil || f.lookups.Load() != 3 {
		t.Fatalf("a stale answer was not renewed: %d lookups", f.lookups.Load())
	}
}

func TestLatestRefusesTagsThatAreNotVersions(t *testing.T) {
	_, c := newFakeGitHub(t, "v1.2/../../x", nil)
	if _, err := c.Latest(context.Background(), "owner/project", false); err == nil {
		t.Fatal("an odd tag was accepted")
	}
}

func TestLatestReportsRateLimitsAndKeepsTheFailureBriefly(t *testing.T) {
	f, c := newFakeGitHub(t, "1.0", nil)
	f.status = http.StatusForbidden
	now := time.Unix(1_000_000, 0)
	c.Now = func() time.Time { return now }

	_, err := c.Latest(context.Background(), "owner/project", false)
	if err == nil || !strings.Contains(err.Error(), "rate limited") {
		t.Fatalf("rate limit: %v", err)
	}
	_, _ = c.Latest(context.Background(), "owner/project", false)
	if f.lookups.Load() != 1 {
		t.Fatal("a fresh failure was asked for again")
	}
	now = now.Add(FailTTL + time.Second)
	_, _ = c.Latest(context.Background(), "owner/project", false)
	if f.lookups.Load() != 2 {
		t.Fatal("an old failure was kept")
	}
}

func TestLatestSaysWhenGitHubCannotBeReached(t *testing.T) {
	f, c := newFakeGitHub(t, "1.0", nil)
	f.srv.Close()
	_, err := c.Latest(context.Background(), "owner/project", false)
	if !errors.Is(err, ErrOffline) {
		t.Fatalf("offline: %v", err)
	}
}

func TestParseChecksums(t *testing.T) {
	a := strings.Repeat("a", 64)
	b := strings.Repeat("B", 64)
	sums, err := ParseChecksums(strings.NewReader(
		"# generated\n\n" + a + " *one.efi\n" + b + "  two.tar.gz\r\nnot a line\n"))
	if err != nil {
		t.Fatal(err)
	}
	if sums["one.efi"] != a || sums["two.tar.gz"] != strings.ToLower(b) || len(sums) != 2 {
		t.Fatalf("sums %v", sums)
	}
	if _, err := ParseChecksums(strings.NewReader("# nothing\n")); err == nil {
		t.Fatal("an empty checksum file was accepted")
	}
	if _, err := ParseChecksums(strings.NewReader(a + " x\n" + strings.Repeat("c", 64) + " x\n")); err == nil {
		t.Fatal("two sums for one file were accepted")
	}
}

func TestDownloadKeepsOnlyAVerifiedFile(t *testing.T) {
	data := []byte("the release's file")
	f, c := newFakeGitHub(t, "1.0", map[string][]byte{"f.bin": data})
	dst := filepath.Join(t.TempDir(), "f.bin")
	url := f.srv.URL + "/download/f.bin"

	err := c.Download(context.Background(), url, strings.Repeat("0", 64), 1<<20, dst, nil)
	if !errors.Is(err, ErrChecksum) {
		t.Fatalf("mismatch: %v", err)
	}
	if _, err := os.Stat(dst); err == nil {
		t.Fatal("a file with the wrong sum was kept")
	}
	if _, err := os.Stat(dst + ".part"); err == nil {
		t.Fatal("the partial file was left")
	}

	if err := c.Download(context.Background(), url, hexSum(data), 4, dst, nil); err == nil {
		t.Fatal("a file over the limit was kept")
	}
	if err := c.Download(context.Background(), strings.Replace(url, "https", "http", 1), hexSum(data), 1<<20, dst, nil); err == nil {
		t.Fatal("a plain http URL was fetched")
	}

	var last int64
	if err := c.Download(context.Background(), url, hexSum(data), 1<<20, dst, func(done, total int64) { last = done }); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(dst); string(got) != string(data) || last != int64(len(data)) {
		t.Fatalf("downloaded %q, progress %d", got, last)
	}
}

func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func readFiles(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		data, _ := os.ReadFile(filepath.Join(dir, e.Name()))
		out[e.Name()] = string(data)
	}
	return out
}

func TestSwapReplacesEveryFile(t *testing.T) {
	staged, dir := t.TempDir(), t.TempDir()
	writeFiles(t, dir, map[string]string{"a": "old a", "b": "old b", "keep": "kept"})
	writeFiles(t, staged, map[string]string{"a": "new a", "b": "new b", "c": "new c"})

	committed := false
	if err := Swap(staged, dir, []string{"a", "b", "c"}, func() error { committed = true; return nil }); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"a": "new a", "b": "new b", "c": "new c", "keep": "kept"}
	if got := readFiles(t, dir); fmt.Sprint(got) != fmt.Sprint(want) || !committed {
		t.Fatalf("after the swap %v", got)
	}
}

func TestSwapRollsBackWhenTheCommitFails(t *testing.T) {
	staged, dir := t.TempDir(), t.TempDir()
	old := map[string]string{"a": "old a", "b": "old b"}
	writeFiles(t, dir, old)
	writeFiles(t, staged, map[string]string{"a": "new a", "b": "new b", "c": "new c"})

	err := Swap(staged, dir, []string{"a", "b", "c"}, func() error { return errors.New("no record") })
	if err == nil {
		t.Fatal("a failed commit was not reported")
	}
	if got := readFiles(t, dir); fmt.Sprint(got) != fmt.Sprint(old) {
		t.Fatalf("after the rollback %v", got)
	}
}

func TestSwapRollsBackWhenAFileIsMissing(t *testing.T) {
	staged, dir := t.TempDir(), t.TempDir()
	old := map[string]string{"a": "old a", "b": "old b"}
	writeFiles(t, dir, old)
	writeFiles(t, staged, map[string]string{"a": "new a"})

	if err := Swap(staged, dir, []string{"a", "b"}, nil); err == nil {
		t.Fatal("a missing staged file was not reported")
	}
	if got := readFiles(t, dir); fmt.Sprint(got) != fmt.Sprint(old) {
		t.Fatalf("after the rollback %v", got)
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"3.0.3", "3.0.3", 0},
		{"3.0.10", "3.0.9", 1},
		{"v1.1.17", "1.1.18", -1},
		{"2.0", "2.0.0", 0},
		{"2.0.1", "2.0", 1},
		{"1.10", "1.9.9", 1},
	}
	for _, c := range cases {
		if got := CompareVersions(c.a, c.b); got != c.want {
			t.Errorf("CompareVersions(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

// fakeComponent is a component with one file, installed at version.
type fakeComponent struct {
	version  string
	inUse    error
	failWith error
	file     []byte
}

func (fc *fakeComponent) component(u **Updater) Component {
	return Component{
		Name:         "fake",
		Repo:         "owner/project",
		Pinned:       "1.0",
		ChecksumFile: func(string) string { return "SHA256SUMS" },
		Assets:       func(v string) []string { return []string{"fake-" + v + ".bin"} },
		Installed:    func() (string, bool) { return fc.version, fc.version != "" },
		InUse:        func() error { return fc.inUse },
		Install: func(ctx context.Context, rel Release, sums Checksums, progress func(int)) error {
			if fc.failWith != nil {
				return fc.failWith
			}
			dst := filepath.Join(os.TempDir(), fmt.Sprintf("upstream-test-%d", time.Now().UnixNano()))
			defer func() { _ = os.Remove(dst) }()
			if err := (*u).Download(ctx, rel, sums, "fake-"+rel.Version+".bin", 1<<20, dst, progress, 0, 100); err != nil {
				return err
			}
			fc.file, _ = os.ReadFile(dst)
			fc.version = rel.Version
			return nil
		},
	}
}

func waitJob(t *testing.T, u *Updater) Job {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if j := u.Status().Job; j.State != JobRunning {
			return j
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("the update did not finish")
	return Job{}
}

func TestUpdaterChecksAndUpdates(t *testing.T) {
	payload := []byte("release 1.1")
	sums := fmt.Sprintf("%s  fake-1.1.bin\n", hexSum(payload))
	_, c := newFakeGitHub(t, "v1.1", map[string][]byte{"fake-1.1.bin": payload, "SHA256SUMS": []byte(sums)})
	fc := &fakeComponent{version: "1.0"}
	var u *Updater
	u = NewUpdater(fc.component(&u), c)

	if st := u.Status(); st.CheckedAt != nil || st.UpdateAvailable || st.Version != "1.0" || st.Pinned != "1.0" {
		t.Fatalf("before a check %+v", st)
	}
	st := u.Check(context.Background(), true)
	if st.Latest != "1.1" || !st.UpdateAvailable || st.CheckedAt == nil {
		t.Fatalf("after a check %+v", st)
	}

	if err := u.Start(); err != nil {
		t.Fatal(err)
	}
	if j := waitJob(t, u); j.State != JobDone || j.Version != "1.1" {
		t.Fatalf("job %+v", j)
	}
	if fc.version != "1.1" || string(fc.file) != string(payload) {
		t.Fatalf("installed %s %q", fc.version, fc.file)
	}
	if st := u.Status(); st.UpdateAvailable || st.Version != "1.1" {
		t.Fatalf("after the update %+v", st)
	}
	if err := u.Start(); !errors.Is(err, ErrUpToDate) {
		t.Fatalf("a second update: %v", err)
	}
}

func TestUpdaterRefusesAReleaseWithoutAChecksumFile(t *testing.T) {
	_, c := newFakeGitHub(t, "1.1", map[string][]byte{"fake-1.1.bin": []byte("x")})
	fc := &fakeComponent{version: "1.0"}
	var u *Updater
	u = NewUpdater(fc.component(&u), c)

	st := u.Check(context.Background(), true)
	if st.UpdateAvailable || !strings.Contains(st.Unverifiable, "no checksum file") {
		t.Fatalf("status %+v", st)
	}
	if err := u.Start(); err == nil || fc.version != "1.0" {
		t.Fatalf("an unverifiable release was installed: %v", err)
	}
}

func TestUpdaterFailsOnAWrongChecksum(t *testing.T) {
	sums := fmt.Sprintf("%s  fake-1.1.bin\n", strings.Repeat("0", 64))
	_, c := newFakeGitHub(t, "1.1", map[string][]byte{"fake-1.1.bin": []byte("tampered"), "SHA256SUMS": []byte(sums)})
	fc := &fakeComponent{version: "1.0"}
	var u *Updater
	u = NewUpdater(fc.component(&u), c)

	if err := u.Start(); err != nil {
		t.Fatal(err)
	}
	j := waitJob(t, u)
	if j.State != JobFailed || !strings.Contains(j.Error, "checksum") {
		t.Fatalf("job %+v", j)
	}
	if fc.version != "1.0" || fc.file != nil {
		t.Fatal("a tampered file was installed")
	}
}

func TestUpdaterRefusesWhileInUseOrNotInstalled(t *testing.T) {
	f, c := newFakeGitHub(t, "1.1", nil)
	fc := &fakeComponent{version: "1.0", inUse: errors.New("eject it first")}
	var u *Updater
	u = NewUpdater(fc.component(&u), c)

	if err := u.Start(); err == nil || err.Error() != "eject it first" {
		t.Fatalf("in use: %v", err)
	}
	if st := u.Status(); st.InUse != "eject it first" {
		t.Fatalf("status %+v", st)
	}
	fc.inUse, fc.version = nil, ""
	if err := u.Start(); !errors.Is(err, ErrNotInstalled) {
		t.Fatalf("not installed: %v", err)
	}
	if f.lookups.Load() != 0 {
		t.Fatal("a refused update asked GitHub")
	}
}
