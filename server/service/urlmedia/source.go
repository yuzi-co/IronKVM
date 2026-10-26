// Package urlmedia serves an ISO or disk image to a virtual drive straight
// from an HTTP(S) URL, without storing it on the board (ironkvm-dist#106).
//
// The image is read with byte-range requests through a small block cache,
// and served to the kernel as a network block device (NBD) on one end of a
// socket pair; the mass storage gadget then serves /dev/nbdN like any other
// device. The kernel needs CONFIG_BLK_DEV_NBD, which only the mainline
// kernel (slot B) has, so the feature checks for it at run time.
package urlmedia

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"NanoKVM-Server/utils"
)

// MaxSize is the largest image served. The mass storage gadget addresses
// 512-byte sectors with 32 bits (READ CAPACITY(10)), so 2 TiB is the
// largest disk it can show.
const MaxSize = int64(2) << 40

var (
	ErrInvalidURL       = errors.New("give an http or https URL")
	ErrNoRanges         = errors.New("the server does not support byte ranges, so the image cannot be read in place: download it instead")
	ErrUnknownSize      = errors.New("the server did not say how large the file is")
	ErrSize             = fmt.Errorf("the file is empty or larger than %d GiB", MaxSize>>30)
	ErrTooManyRedirects = errors.New("too many redirects")
	ErrDowngrade        = errors.New("a redirect from https to http was refused")
	ErrBlockedAddress   = errors.New("the URL points at the board itself or a link-local address")
	ErrTLS              = errors.New("TLS error")
	ErrChanged          = errors.New("the file changed on the server since it was mounted")
)

// StatusError is an HTTP answer that is neither the range asked for nor a
// redirect.
type StatusError struct{ Code int }

func (e *StatusError) Error() string {
	return fmt.Sprintf("the server answered %d %s", e.Code, http.StatusText(e.Code))
}

// Options tune a source. The zero value is the board's policy.
type Options struct {
	// AllowLoopback lets the tests reach their loopback server.
	AllowLoopback bool
	// Attempts is how many times one read is tried before the host gets an
	// I/O error; AttemptTimeout bounds each; Backoff is the wait before each
	// retry (the last value repeats).
	Attempts       int
	AttemptTimeout time.Duration
	Backoff        []time.Duration
	// MaxRedirects bounds the redirects followed when opening.
	MaxRedirects int
}

// The board's policy: a read gives up after about 25 s (3 attempts of up to
// 8 s, with 0.5 s and 1.5 s between), inside the 30 s a host's SCSI layer
// usually waits for a command, and well inside the kernel's NBD timeout.
const (
	defaultAttempts       = 3
	defaultAttemptTimeout = 8 * time.Second
	defaultMaxRedirects   = 5
	probeTimeout          = 20 * time.Second
)

var defaultBackoff = []time.Duration{500 * time.Millisecond, 1500 * time.Millisecond}

func (o Options) withDefaults() Options {
	if o.Attempts <= 0 {
		o.Attempts = defaultAttempts
	}
	if o.AttemptTimeout <= 0 {
		o.AttemptTimeout = defaultAttemptTimeout
	}
	if len(o.Backoff) == 0 {
		o.Backoff = defaultBackoff
	}
	if o.MaxRedirects <= 0 {
		o.MaxRedirects = defaultMaxRedirects
	}
	return o
}

// Source reads an image over HTTP with range requests.
type Source struct {
	url       string // after redirects
	size      int64
	etag      string // strong ETag, if the server gave one
	client    *http.Client
	opts      Options
	bytesRead atomic.Int64
}

// blockedIP refuses the board itself and link-local addresses, which
// include cloud metadata endpoints. Private ranges are allowed: an image on
// a LAN server is the main use.
func blockedIP(ip net.IP, allowLoopback bool) bool {
	if ip.IsLoopback() {
		return !allowLoopback
	}
	return ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() || ip.IsMulticast()
}

// checkHost resolves the URL's host and refuses it when any address it has
// is blocked. A proxy, when one is set, still gets only allowed targets.
func checkHost(ctx context.Context, host string, allowLoopback bool) error {
	if ip := net.ParseIP(host); ip != nil {
		if blockedIP(ip, allowLoopback) {
			return ErrBlockedAddress
		}
		return nil
	}
	if strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".localhost") {
		if !allowLoopback {
			return ErrBlockedAddress
		}
	}
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		// Behind a proxy the board may not resolve the name itself; the
		// request then fails or succeeds on its own.
		return nil
	}
	for _, a := range addrs {
		if blockedIP(a.IP, allowLoopback) {
			return ErrBlockedAddress
		}
	}
	return nil
}

func parseURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return nil, ErrInvalidURL
	}
	return u, nil
}

func newClient(opts Options) *http.Client {
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	transport := &http.Transport{
		Proxy:                 utils.ProxyFromConfig,
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: opts.AttemptTimeout,
		IdleConnTimeout:       60 * time.Second,
		MaxIdleConnsPerHost:   2,
		// A compressed answer would not be the bytes of the range.
		DisableCompression: true,
	}
	return &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > opts.MaxRedirects {
				return ErrTooManyRedirects
			}
			if via[len(via)-1].URL.Scheme == "https" && req.URL.Scheme != "https" {
				return ErrDowngrade
			}
			return checkHost(req.Context(), req.URL.Hostname(), opts.AllowLoopback)
		},
	}
}

// classify turns a transport error into one of the package's errors.
func classify(err error) error {
	var unknown x509.UnknownAuthorityError
	var hostname x509.HostnameError
	var invalid x509.CertificateInvalidError
	var record tls.RecordHeaderError
	var verify *tls.CertificateVerificationError
	switch {
	case errors.Is(err, ErrTooManyRedirects), errors.Is(err, ErrDowngrade), errors.Is(err, ErrBlockedAddress):
		return unwrapURLError(err)
	case errors.As(err, &unknown), errors.As(err, &hostname), errors.As(err, &invalid),
		errors.As(err, &record), errors.As(err, &verify):
		return fmt.Errorf("%w: %s", ErrTLS, unwrapURLError(err))
	}
	return err
}

// unwrapURLError drops the method and URL that net/http puts in front of an
// error, which would show the URL's credentials.
func unwrapURLError(err error) error {
	var ue *url.Error
	if errors.As(err, &ue) {
		return ue.Err
	}
	return err
}

// Open checks that the server can serve the image in place: it asks for its
// first byte and wants a 206 with the total size.
func Open(ctx context.Context, raw string, opts Options) (*Source, error) {
	opts = opts.withDefaults()
	u, err := parseURL(raw)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	if err := checkHost(ctx, u.Hostname(), opts.AllowLoopback); err != nil {
		return nil, err
	}

	client := newClient(opts)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, ErrInvalidURL
	}
	req.Header.Set("Range", "bytes=0-0")
	resp, err := client.Do(req)
	if err != nil {
		client.CloseIdleConnections()
		return nil, classify(unwrapURLError(err))
	}
	defer drain(resp.Body)

	switch {
	case resp.StatusCode == http.StatusOK:
		client.CloseIdleConnections()
		return nil, ErrNoRanges
	case resp.StatusCode != http.StatusPartialContent:
		client.CloseIdleConnections()
		return nil, &StatusError{Code: resp.StatusCode}
	}
	_, _, total, err := parseContentRange(resp.Header.Get("Content-Range"))
	if err != nil {
		client.CloseIdleConnections()
		return nil, err
	}
	if total <= 0 || total > MaxSize {
		client.CloseIdleConnections()
		return nil, ErrSize
	}

	src := &Source{url: resp.Request.URL.String(), size: total, client: client, opts: opts}
	if etag := resp.Header.Get("ETag"); etag != "" && !strings.HasPrefix(etag, "W/") {
		src.etag = etag
	}
	return src, nil
}

// parseContentRange reads "bytes first-last/total"; a total of "*" is
// ErrUnknownSize.
func parseContentRange(v string) (first, last, total int64, err error) {
	rest, ok := strings.CutPrefix(strings.TrimSpace(v), "bytes ")
	if !ok {
		return 0, 0, 0, ErrNoRanges
	}
	span, size, ok := strings.Cut(rest, "/")
	if !ok {
		return 0, 0, 0, ErrNoRanges
	}
	if size == "*" {
		return 0, 0, 0, ErrUnknownSize
	}
	a, b, ok := strings.Cut(span, "-")
	if !ok {
		return 0, 0, 0, ErrNoRanges
	}
	first, err1 := strconv.ParseInt(a, 10, 64)
	last, err2 := strconv.ParseInt(b, 10, 64)
	total, err3 := strconv.ParseInt(size, 10, 64)
	if err1 != nil || err2 != nil || err3 != nil || first > last {
		return 0, 0, 0, ErrNoRanges
	}
	return first, last, total, nil
}

func drain(body io.ReadCloser) {
	_, _ = io.CopyN(io.Discard, body, 64<<10)
	_ = body.Close()
}

// Size is the image's size in bytes.
func (s *Source) Size() int64 { return s.size }

// BytesRead is how many bytes of the image have come over the network.
func (s *Source) BytesRead() int64 { return s.bytesRead.Load() }

// Close drops the idle connections.
func (s *Source) Close() { s.client.CloseIdleConnections() }

// permanent marks an error that another attempt would not change.
type permanent struct{ err error }

func (p permanent) Error() string { return p.err.Error() }
func (p permanent) Unwrap() error { return p.err }

// ReadAt fills p from offset off, trying again after network errors, server
// errors and short bodies, and giving up after the attempts.
func (s *Source) ReadAt(ctx context.Context, p []byte, off int64) error {
	if off < 0 || off+int64(len(p)) > s.size {
		return ErrOutOfRange
	}
	if len(p) == 0 {
		return nil
	}
	var err error
	for attempt := 0; attempt < s.opts.Attempts; attempt++ {
		if attempt > 0 {
			wait := s.opts.Backoff[min(attempt-1, len(s.opts.Backoff)-1)]
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		err = s.readOnce(ctx, p, off)
		if err == nil {
			return nil
		}
		var perm permanent
		if errors.As(err, &perm) || ctx.Err() != nil {
			break
		}
	}
	return err
}

func (s *Source) readOnce(ctx context.Context, p []byte, off int64) error {
	ctx, cancel := context.WithTimeout(ctx, s.opts.AttemptTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		return permanent{err}
	}
	last := off + int64(len(p)) - 1
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", off, last))
	if s.etag != "" {
		req.Header.Set("If-Match", s.etag)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		err = classify(unwrapURLError(err))
		if errors.Is(err, ErrTLS) {
			return permanent{err}
		}
		return err
	}
	defer drain(resp.Body)

	switch code := resp.StatusCode; {
	case code == http.StatusPartialContent:
	case code == http.StatusPreconditionFailed:
		return permanent{ErrChanged}
	case code == http.StatusOK:
		return permanent{ErrNoRanges}
	case code >= 500 || code == http.StatusTooManyRequests || code == http.StatusRequestTimeout:
		return &StatusError{Code: code}
	default:
		return permanent{&StatusError{Code: code}}
	}

	first, end, total, err := parseContentRange(resp.Header.Get("Content-Range"))
	if err != nil {
		return permanent{err}
	}
	if total != s.size {
		return permanent{ErrChanged}
	}
	if first != off || end != last {
		return permanent{fmt.Errorf("the server sent bytes %d-%d for %d-%d", first, end, off, last)}
	}
	n, err := io.ReadFull(resp.Body, p)
	s.bytesRead.Add(int64(n))
	if err != nil {
		return fmt.Errorf("short body: %w", err)
	}
	return nil
}

// Redact is the URL without its credentials and query, for status and logs.
func Redact(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	u.User = nil
	if u.RawQuery != "" {
		u.RawQuery = ""
		u.ForceQuery = true
	}
	u.Fragment = ""
	return u.String()
}

// HostOf is the URL's host and port, for the audit log.
func HostOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Host
}
