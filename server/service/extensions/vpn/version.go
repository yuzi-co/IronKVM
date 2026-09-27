package vpn

import (
	"sync"
	"time"
)

// UpdateTTL is how long a looked-up latest version is kept. The page asks on
// every visit; the package servers need not hear about each one.
const UpdateTTL = time.Hour

// VersionCache keeps the latest version an update check found.
type VersionCache struct {
	TTL time.Duration
	// Now is the clock, for the tests. Nil means time.Now.
	Now func() time.Time

	mu     sync.Mutex
	at     time.Time
	latest string
}

// Latest returns the cached version while it is younger than TTL, and asks
// fetch otherwise. A failed fetch is not cached. The lock is held through the
// fetch, so two pages open at once make one request.
func (c *VersionCache) Latest(fetch func() (string, error)) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now
	if c.Now != nil {
		now = c.Now
	}
	if c.latest != "" && now().Sub(c.at) < c.TTL {
		return c.latest, nil
	}
	v, err := fetch()
	if err != nil {
		return "", err
	}
	c.latest, c.at = v, now()
	return v, nil
}

// Reset forgets the cached version, after an update.
func (c *VersionCache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.latest = ""
	c.at = time.Time{}
}
