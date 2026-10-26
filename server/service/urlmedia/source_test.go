package urlmedia

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// testImage is a pattern every byte of which says where it is, so a read from
// the wrong offset shows.
func testImage(size int) []byte {
	b := make([]byte, size)
	for i := range b {
		b[i] = byte(i*7 + i/251)
	}
	return b
}

// rangeServer serves data with byte ranges, as http.ServeContent does.
func rangeServer(t *testing.T, data []byte) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "image.iso", time.Unix(0, 0), bytes.NewReader(data))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// testOptions lets the tests reach httptest's loopback server, and keeps
// the retries short.
func testOptions() Options {
	return Options{
		AllowLoopback:  true,
		Attempts:       3,
		AttemptTimeout: 2 * time.Second,
		Backoff:        []time.Duration{time.Millisecond, time.Millisecond},
	}
}

func TestOpenReadsSizeFromARangedAnswer(t *testing.T) {
	data := testImage(100_000)
	srv := rangeServer(t, data)

	src, err := Open(context.Background(), srv.URL+"/image.iso", testOptions())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer src.Close()

	if src.Size() != int64(len(data)) {
		t.Fatalf("size = %d, want %d", src.Size(), len(data))
	}
	p := make([]byte, 5000)
	if err := src.ReadAt(context.Background(), p, 12345); err != nil {
		t.Fatalf("ReadAt: %v", err)
	}
	if !bytes.Equal(p, data[12345:12345+5000]) {
		t.Fatal("ReadAt returned the wrong bytes")
	}
	if got := src.BytesRead(); got < 5000 {
		t.Fatalf("BytesRead = %d, want at least 5000", got)
	}
}

func TestOpenRefusesAServerWithoutRanges(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		_, _ = w.Write(make([]byte, 1000))
	}))
	defer srv.Close()

	_, err := Open(context.Background(), srv.URL+"/x.iso", testOptions())
	if !errors.Is(err, ErrNoRanges) {
		t.Fatalf("err = %v, want ErrNoRanges", err)
	}
}

func TestOpenRefusesAnUnknownSize(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Range", "bytes 0-0/*")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write([]byte{0})
	}))
	defer srv.Close()

	_, err := Open(context.Background(), srv.URL+"/x.iso", testOptions())
	if !errors.Is(err, ErrUnknownSize) {
		t.Fatalf("err = %v, want ErrUnknownSize", err)
	}
}

func TestOpenRefusesAnEmptyOrHugeFile(t *testing.T) {
	for _, total := range []string{"0", fmt.Sprint(MaxSize + 1)} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Range", "bytes 0-0/"+total)
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write([]byte{0})
		}))
		_, err := Open(context.Background(), srv.URL+"/x.iso", testOptions())
		srv.Close()
		if !errors.Is(err, ErrSize) {
			t.Fatalf("total %s: err = %v, want ErrSize", total, err)
		}
	}
}

func TestOpenFollowsRedirectsAndReadsFromTheTarget(t *testing.T) {
	data := testImage(4096)
	target := rangeServer(t, data)
	var hops atomic.Int32
	front := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hops.Add(1)
		http.Redirect(w, r, target.URL+"/real.iso", http.StatusFound)
	}))
	defer front.Close()

	src, err := Open(context.Background(), front.URL+"/x.iso", testOptions())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer src.Close()
	before := hops.Load()

	p := make([]byte, 100)
	if err := src.ReadAt(context.Background(), p, 10); err != nil {
		t.Fatalf("ReadAt: %v", err)
	}
	if !bytes.Equal(p, data[10:110]) {
		t.Fatal("wrong bytes after a redirect")
	}
	if hops.Load() != before {
		t.Fatal("a read went through the redirect again instead of to its target")
	}
}

func TestOpenStopsAfterTooManyRedirects(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, srv.URL+r.URL.Path+"x", http.StatusFound)
	}))
	defer srv.Close()

	_, err := Open(context.Background(), srv.URL+"/a", testOptions())
	if !errors.Is(err, ErrTooManyRedirects) {
		t.Fatalf("err = %v, want ErrTooManyRedirects", err)
	}
}

func TestOpenRefusesBadURLs(t *testing.T) {
	for _, raw := range []string{"", "ftp://example.com/x.iso", "http://", "/data/x.iso", "http://exa mple.com/"} {
		if _, err := Open(context.Background(), raw, testOptions()); !errors.Is(err, ErrInvalidURL) {
			t.Errorf("%q: err = %v, want ErrInvalidURL", raw, err)
		}
	}
}

func TestOpenRefusesLoopbackAndLinkLocal(t *testing.T) {
	opts := testOptions()
	opts.AllowLoopback = false
	for _, raw := range []string{
		"http://127.0.0.1/x.iso",
		"http://[::1]/x.iso",
		"http://169.254.169.254/latest/meta-data",
		"http://[fe80::1]/x.iso",
		"http://0.0.0.0/x.iso",
	} {
		if _, err := Open(context.Background(), raw, opts); !errors.Is(err, ErrBlockedAddress) {
			t.Errorf("%s: err = %v, want ErrBlockedAddress", raw, err)
		}
	}
}

func TestBlockedAddressAllowsTheLAN(t *testing.T) {
	for _, ip := range []string{"10.0.30.5", "192.168.1.10", "172.16.0.1", "fd00::1", "8.8.8.8"} {
		if blockedIP(mustIP(t, ip), false) {
			t.Errorf("%s is blocked; the LAN is the main use", ip)
		}
	}
}

func TestOpenNamesTLSErrors(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	_, err := Open(context.Background(), srv.URL+"/x.iso", testOptions())
	if !errors.Is(err, ErrTLS) {
		t.Fatalf("err = %v, want ErrTLS", err)
	}
}

func TestOpenReportsHTTPStatus(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()

	_, err := Open(context.Background(), srv.URL+"/x.iso", testOptions())
	var status *StatusError
	if !errors.As(err, &status) || status.Code != http.StatusNotFound {
		t.Fatalf("err = %v, want a 404 StatusError", err)
	}
}

func TestReadAtRetriesAfterServerErrors(t *testing.T) {
	data := testImage(8192)
	var fail atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Range") != "bytes=0-0" && fail.Add(-1) >= 0 {
			http.Error(w, "busy", http.StatusServiceUnavailable)
			return
		}
		http.ServeContent(w, r, "x.iso", time.Unix(0, 0), bytes.NewReader(data))
	}))
	defer srv.Close()

	src, err := Open(context.Background(), srv.URL+"/x.iso", testOptions())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer src.Close()

	fail.Store(2) // two 503s, then the third attempt works
	p := make([]byte, 512)
	if err := src.ReadAt(context.Background(), p, 1024); err != nil {
		t.Fatalf("ReadAt after two 503s: %v", err)
	}
	if !bytes.Equal(p, data[1024:1536]) {
		t.Fatal("wrong bytes after a retry")
	}

	fail.Store(10) // more failures than attempts: the read fails, it does not hang
	if err := src.ReadAt(context.Background(), p, 0); err == nil {
		t.Fatal("ReadAt succeeded with every attempt failing")
	}

	fail.Store(0) // the server is back: the next read works
	if err := src.ReadAt(context.Background(), p, 2048); err != nil {
		t.Fatalf("ReadAt after recovery: %v", err)
	}
}

func TestReadAtRetriesAShortBody(t *testing.T) {
	data := testImage(8192)
	var short atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Range") != "bytes=0-0" && short.Add(-1) >= 0 {
			// Promise 1000 bytes, send 10, and drop the connection.
			w.Header().Set("Content-Range", fmt.Sprintf("bytes 0-999/%d", len(data)))
			w.Header().Set("Content-Length", "1000")
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write(data[:10])
			if hj, ok := w.(http.Hijacker); ok {
				conn, _, _ := hj.Hijack()
				_ = conn.Close()
			}
			return
		}
		http.ServeContent(w, r, "x.iso", time.Unix(0, 0), bytes.NewReader(data))
	}))
	defer srv.Close()

	src, err := Open(context.Background(), srv.URL+"/x.iso", testOptions())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer src.Close()

	short.Store(1)
	p := make([]byte, 1000)
	if err := src.ReadAt(context.Background(), p, 0); err != nil {
		t.Fatalf("ReadAt after a short body: %v", err)
	}
	if !bytes.Equal(p, data[:1000]) {
		t.Fatal("wrong bytes after a short body")
	}
}

func TestReadAtDoesNotRetryAClientError(t *testing.T) {
	data := testImage(4096)
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Range") != "bytes=0-0" {
			calls.Add(1)
			http.Error(w, "gone", http.StatusGone)
			return
		}
		http.ServeContent(w, r, "x.iso", time.Unix(0, 0), bytes.NewReader(data))
	}))
	defer srv.Close()

	src, err := Open(context.Background(), srv.URL+"/x.iso", testOptions())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer src.Close()

	if err := src.ReadAt(context.Background(), make([]byte, 10), 0); err == nil {
		t.Fatal("ReadAt succeeded on a 410")
	}
	if calls.Load() != 1 {
		t.Fatalf("%d requests for a 410, want 1", calls.Load())
	}
}

func TestReadAtNoticesTheFileChanged(t *testing.T) {
	data := testImage(4096)
	var grown atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := data
		if grown.Load() {
			body = append(append([]byte{}, data...), 1, 2, 3)
		}
		http.ServeContent(w, r, "x.iso", time.Unix(0, 0), bytes.NewReader(body))
	}))
	defer srv.Close()

	src, err := Open(context.Background(), srv.URL+"/x.iso", testOptions())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer src.Close()

	grown.Store(true)
	err = src.ReadAt(context.Background(), make([]byte, 10), 0)
	if !errors.Is(err, ErrChanged) {
		t.Fatalf("err = %v, want ErrChanged", err)
	}
}

func TestReadAtGivesUpWithinItsBudget(t *testing.T) {
	data := testImage(4096)
	var stall atomic.Bool
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if stall.Load() {
			select {
			case <-release:
			case <-r.Context().Done():
			}
			return
		}
		http.ServeContent(w, r, "x.iso", time.Unix(0, 0), bytes.NewReader(data))
	}))
	defer srv.Close()
	defer close(release)

	opts := testOptions()
	opts.AttemptTimeout = 100 * time.Millisecond
	src, err := Open(context.Background(), srv.URL+"/x.iso", opts)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer src.Close()

	stall.Store(true)
	start := time.Now()
	if err := src.ReadAt(context.Background(), make([]byte, 10), 0); err == nil {
		t.Fatal("ReadAt succeeded against a server that never answers")
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("ReadAt took %s against a stalled server", d)
	}
}

func TestRedactDropsCredentialsAndQuery(t *testing.T) {
	got := Redact("https://user:secret@example.com:8443/isos/x.iso?X-Amz-Signature=abc")
	if strings.Contains(got, "secret") || strings.Contains(got, "user") || strings.Contains(got, "abc") {
		t.Fatalf("Redact left a secret in %q", got)
	}
	if !strings.Contains(got, "example.com:8443/isos/x.iso") {
		t.Fatalf("Redact = %q, want the host and path kept", got)
	}
	if h := HostOf("https://user:pw@example.com:8443/x"); h != "example.com:8443" {
		t.Fatalf("HostOf = %q", h)
	}
}
