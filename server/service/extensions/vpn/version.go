package vpn

import (
	"sync"
	"time"
)

// UpdateTTL is how long a looked-up latest version is kept. The page asks on
// every visit; the package servers need not hear about each one.
const UpdateTTL = time.Hour

// UpdateFailTTL is how long a failed check is kept, so a board without
// network does not run apk or reach the release server on every visit.
const UpdateFailTTL = 5 * time.Minute

// CheckTimeout bounds one update check, apk update and search or the release
// server's redirect, so the page's request cannot hang on a slow mirror.
const CheckTimeout = 30 * time.Second

// VersionCache keeps the latest version an update check found, or the reason
// it failed.
type VersionCache struct {
	TTL time.Duration
	// Now is the clock, for the tests. Nil means time.Now.
	Now func() time.Time

	mu       sync.Mutex
	at       time.Time
	latest   string
	err      error
	inflight chan struct{}
}

func (c *VersionCache) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

// Latest returns the cached version while it is younger than TTL, the cached
// failure while it is younger than UpdateFailTTL, and asks fetch otherwise.
// Callers that come while a fetch runs wait for it rather than start their
// own, and the lock is not held while fetch runs.
func (c *VersionCache) Latest(fetch func() (string, error)) (string, error) {
	c.mu.Lock()
	for {
		now := c.now()
		if c.latest != "" && now.Sub(c.at) < c.TTL {
			v := c.latest
			c.mu.Unlock()
			return v, nil
		}
		if c.err != nil && now.Sub(c.at) < UpdateFailTTL {
			err := c.err
			c.mu.Unlock()
			return "", err
		}
		if c.inflight == nil {
			break
		}
		wait := c.inflight
		c.mu.Unlock()
		<-wait
		c.mu.Lock()
		if c.latest == "" && c.err == nil {
			// Reset ran while it waited: ask again.
			continue
		}
	}
	done := make(chan struct{})
	c.inflight = done
	c.mu.Unlock()

	v, err := fetch()

	c.mu.Lock()
	c.inflight = nil
	c.at = c.now()
	c.latest, c.err = v, err
	if err != nil {
		c.latest = ""
	}
	close(done)
	c.mu.Unlock()
	return v, err
}

// Reset forgets the cached version or failure, after an update.
func (c *VersionCache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.latest, c.err = "", nil
	c.at = time.Time{}
}
