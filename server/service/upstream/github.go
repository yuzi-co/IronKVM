// Package upstream updates the third-party components IronKVM downloads, such
// as netboot.xyz and Ventoy, to a newer release than the one the server pins.
//
// Each component keeps its pinned release as the built-in default. An update
// is never automatic: the owner checks for one, and asks for it. The new
// release's files are downloaded over HTTPS and checked against the checksum
// file the project publishes in that same release; a component whose releases
// carry no such file is not offered updates. The files are staged next to the
// installed ones, and swapped in only once every one of them is verified, so
// a failed update leaves the installed release as it was.
package upstream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"NanoKVM-Server/utils"
)

const (
	// CacheTTL is how long a release lookup is kept. GitHub allows an
	// unauthenticated client 60 API requests an hour, and releases are rare.
	CacheTTL = 6 * time.Hour
	// FailTTL is how long a failed lookup is kept, so a board without a
	// network does not ask again on every check.
	FailTTL = 5 * time.Minute
	// lookupTimeout bounds one release lookup.
	lookupTimeout = 30 * time.Second
	// apiMax bounds a release's JSON. One with a hundred assets is well
	// under it.
	apiMax = 4 << 20
)

// Release is a project's release as GitHub publishes it.
type Release struct {
	Tag     string
	Version string // the tag without a leading "v"
	// Assets maps each asset's name to its download URL.
	Assets    map[string]string
	Published time.Time
}

// ErrOffline wraps a lookup that could not reach GitHub at all.
var ErrOffline = errors.New("cannot reach GitHub")

// versionPattern is the shape a release's version must have before it is
// used in a URL or a file name.
var versionPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+){1,3}$`)

// Client looks up releases, and keeps each repository's answer for CacheTTL.
type Client struct {
	// APIBase is GitHub's API, without the trailing slash.
	APIBase string
	HTTP    *http.Client
	// Now is the clock, for the tests. Nil means time.Now.
	Now func() time.Time

	mu    sync.Mutex
	cache map[string]*lookup
}

type lookup struct {
	at       time.Time
	release  Release
	err      error
	inflight chan struct{}
}

// Default is the client the components use.
var Default = NewClient("https://api.github.com", &http.Client{
	Transport: &http.Transport{
		Proxy:                 utils.ProxyFromConfig,
		TLSHandshakeTimeout:   30 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		IdleConnTimeout:       90 * time.Second,
	},
})

func NewClient(apiBase string, httpClient *http.Client) *Client {
	return &Client{APIBase: strings.TrimRight(apiBase, "/"), HTTP: httpClient}
}

func (c *Client) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

// Cached returns the last lookup of repo and when it was made, without asking
// GitHub. ok is false when there has been none.
func (c *Client) Cached(repo string) (rel Release, at time.Time, err error, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	l := c.cache[repo]
	if l == nil || l.at.IsZero() {
		return Release{}, time.Time{}, nil, false
	}
	return l.release, l.at, l.err, true
}

// Latest returns repo's latest release, from the cache while the last answer
// is fresh, unless force asks GitHub again. Callers that come while a lookup
// runs wait for it rather than start their own.
func (c *Client) Latest(ctx context.Context, repo string, force bool) (Release, error) {
	c.mu.Lock()
	if c.cache == nil {
		c.cache = map[string]*lookup{}
	}
	l := c.cache[repo]
	if l == nil {
		l = &lookup{}
		c.cache[repo] = l
	}
	for l.inflight != nil {
		wait := l.inflight
		c.mu.Unlock()
		select {
		case <-wait:
		case <-ctx.Done():
			return Release{}, ctx.Err()
		}
		c.mu.Lock()
		// The lookup that just finished is as fresh as a forced one.
		force = false
	}
	if !force && !l.at.IsZero() {
		age := c.now().Sub(l.at)
		if l.err == nil && age < CacheTTL {
			rel := l.release
			c.mu.Unlock()
			return rel, nil
		}
		if l.err != nil && age < FailTTL {
			err := l.err
			c.mu.Unlock()
			return Release{}, err
		}
	}
	done := make(chan struct{})
	l.inflight = done
	c.mu.Unlock()

	rel, err := c.fetchLatest(ctx, repo)

	c.mu.Lock()
	l.inflight = nil
	l.at = c.now()
	l.release, l.err = rel, err
	close(done)
	c.mu.Unlock()
	return rel, err
}

type apiRelease struct {
	TagName     string    `json:"tag_name"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

// fetchLatest asks GitHub for repo's latest release. GitHub's "latest" is the
// newest release that is neither a draft nor a prerelease.
func (c *Client) fetchLatest(ctx context.Context, repo string) (Release, error) {
	ctx, cancel := context.WithTimeout(ctx, lookupTimeout)
	defer cancel()

	endpoint := fmt.Sprintf("%s/repos/%s/releases/latest", c.APIBase, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Release{}, offline(err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return Release{}, fmt.Errorf("%s has no published release", repo)
	case resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests:
		reset := ""
		if n, err := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64); err == nil {
			reset = fmt.Sprintf(" until %s", time.Unix(n, 0).UTC().Format("15:04 UTC"))
		}
		return Release{}, fmt.Errorf("GitHub refused the release lookup (rate limited%s), try again later", reset)
	case resp.StatusCode != http.StatusOK:
		return Release{}, fmt.Errorf("GitHub answered the release lookup with status %d", resp.StatusCode)
	}

	var raw apiRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, apiMax)).Decode(&raw); err != nil {
		return Release{}, fmt.Errorf("read the release of %s: %w", repo, err)
	}
	return parseRelease(raw)
}

func parseRelease(raw apiRelease) (Release, error) {
	if raw.Draft || raw.Prerelease {
		return Release{}, fmt.Errorf("release %s is not a final release", raw.TagName)
	}
	version := strings.TrimPrefix(raw.TagName, "v")
	if !versionPattern.MatchString(version) {
		return Release{}, fmt.Errorf("release tag %q is not a version this server understands", raw.TagName)
	}

	rel := Release{
		Tag:       raw.TagName,
		Version:   version,
		Assets:    map[string]string{},
		Published: raw.PublishedAt,
	}
	for _, a := range raw.Assets {
		u, err := url.Parse(a.URL)
		if err != nil || u.Scheme != "https" || a.Name == "" {
			continue
		}
		rel.Assets[a.Name] = a.URL
	}
	return rel, nil
}

// offline wraps a request that got no answer at all: no route, no name
// server, a proxy that refused, or a timeout.
func offline(err error) error {
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return fmt.Errorf("%w (is the board online?): %v", ErrOffline, err)
	}
	return fmt.Errorf("%w: %v", ErrOffline, err)
}

// CompareVersions compares two dotted versions, each with an optional leading
// "v", number by number: -1 when a is older, 0 when equal, 1 when newer. A
// part that is not a number compares as 0.
func CompareVersions(a, b string) int {
	pa := strings.Split(strings.TrimPrefix(a, "v"), ".")
	pb := strings.Split(strings.TrimPrefix(b, "v"), ".")
	for i := 0; i < max(len(pa), len(pb)); i++ {
		var x, y int
		if i < len(pa) {
			x, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			y, _ = strconv.Atoi(pb[i])
		}
		if x != y {
			if x < y {
				return -1
			}
			return 1
		}
	}
	return 0
}
