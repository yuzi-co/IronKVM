package redfish

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"sync"
	"time"

	"NanoKVM-Server/authn"
)

const (
	// credentialLifetime is how long a Basic login is remembered. A client
	// that sends credentials on every request then pays one bcrypt a minute.
	credentialLifetime = time.Minute
	// maxCredentials bounds the cache.
	maxCredentials = 64
)

// credentialCache remembers Basic credentials that passed the password
// check. It keeps no password: each entry is an HMAC, under a key drawn at
// start, of the username, the password, and the account's token version and
// password hash. A password change or a revoke changes the last two, so the
// old entry stops matching.
type credentialCache struct {
	mu      sync.Mutex
	key     [32]byte
	now     func() time.Time
	entries map[[sha256.Size]byte]time.Time
}

func newCredentialCache(now func() time.Time) *credentialCache {
	c := &credentialCache{now: now, entries: map[[sha256.Size]byte]time.Time{}}
	if _, err := rand.Read(c.key[:]); err != nil {
		// Without a key the cache would be guessable; run without it.
		c.entries = nil
	}
	return c
}

func (c *credentialCache) digest(user *authn.User, password string) [sha256.Size]byte {
	mac := hmac.New(sha256.New, c.key[:])
	for _, part := range []string{user.Username, password, user.PasswordHash} {
		var length [8]byte
		binary.BigEndian.PutUint64(length[:], uint64(len(part)))
		mac.Write(length[:])
		mac.Write([]byte(part))
	}
	var version [8]byte
	binary.BigEndian.PutUint64(version[:], user.TokenVersion)
	mac.Write(version[:])

	var out [sha256.Size]byte
	copy(out[:], mac.Sum(nil))
	return out
}

// has reports whether the credentials passed the check within the lifetime,
// for the account as it is now.
func (c *credentialCache) has(user *authn.User, password string) bool {
	if c.entries == nil {
		return false
	}
	d := c.digest(user, password)

	c.mu.Lock()
	defer c.mu.Unlock()
	expires, ok := c.entries[d]
	if !ok {
		return false
	}
	if !c.now().Before(expires) {
		delete(c.entries, d)
		return false
	}
	return true
}

// remember records credentials that just passed the check.
func (c *credentialCache) remember(user *authn.User, password string) {
	if c.entries == nil {
		return
	}
	d := c.digest(user, password)

	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	if len(c.entries) >= maxCredentials {
		var oldest [sha256.Size]byte
		var oldestExpiry time.Time
		for key, expires := range c.entries {
			if !now.Before(expires) {
				delete(c.entries, key)
				continue
			}
			if oldestExpiry.IsZero() || expires.Before(oldestExpiry) {
				oldest, oldestExpiry = key, expires
			}
		}
		if len(c.entries) >= maxCredentials {
			delete(c.entries, oldest)
		}
	}
	c.entries[d] = now.Add(credentialLifetime)
}
