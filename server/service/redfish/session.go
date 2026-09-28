package redfish

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

const (
	// sessionIdleTimeout ends a session nobody has used for this long.
	sessionIdleTimeout = 30 * time.Minute
	// maxSessions bounds the store.
	maxSessions = 16
	// maxSessionsPerUser bounds one account. A new session past it drops
	// that account's least recently used one.
	maxSessionsPerUser = 4
)

// errSessionLimit refuses a session when the store is full and the account
// has none of its own to give up: another account's is never ended for it.
var errSessionLimit = errors.New("every session is taken")

// session is one login. The token itself is never kept, only its digest.
type session struct {
	id           string
	tokenHash    [sha256.Size]byte
	username     string
	tokenVersion uint64
	created      time.Time
	lastUsed     time.Time
}

// sessionStore holds the sessions in memory, oldest first. A restart ends
// every one of them.
type sessionStore struct {
	mu       sync.Mutex
	now      func() time.Time
	sessions []*session
}

func newSessionStore(now func() time.Time) *sessionStore {
	return &sessionStore{now: now}
}

func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// create starts a session for username and returns it with its token. This
// is the only place the token exists outside the client.
func (s *sessionStore) create(username string, tokenVersion uint64) (session, string, error) {
	token, err := randomHex(32)
	if err != nil {
		return session{}, "", err
	}
	id, err := randomHex(8)
	if err != nil {
		return session{}, "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	s.pruneLocked(now)

	own := 0
	oldest := -1
	for i, candidate := range s.sessions {
		if candidate.username != username {
			continue
		}
		own++
		if oldest < 0 || candidate.lastUsed.Before(s.sessions[oldest].lastUsed) {
			oldest = i
		}
	}
	switch {
	case own >= maxSessionsPerUser || (len(s.sessions) >= maxSessions && own > 0):
		s.sessions = append(s.sessions[:oldest], s.sessions[oldest+1:]...)
	case len(s.sessions) >= maxSessions:
		return session{}, "", errSessionLimit
	}

	created := &session{
		id:           id,
		tokenHash:    sha256.Sum256([]byte(token)),
		username:     username,
		tokenVersion: tokenVersion,
		created:      now,
		lastUsed:     now,
	}
	s.sessions = append(s.sessions, created)

	return *created, token, nil
}

// byToken returns the live session a token belongs to, and counts the call
// as use of it.
func (s *sessionStore) byToken(token string) (session, bool) {
	hash := sha256.Sum256([]byte(token))

	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	s.pruneLocked(now)

	var found *session
	for _, candidate := range s.sessions {
		if subtle.ConstantTimeCompare(hash[:], candidate.tokenHash[:]) == 1 {
			found = candidate
		}
	}
	if found == nil {
		return session{}, false
	}

	found.lastUsed = now
	return *found, true
}

func (s *sessionStore) get(id string) (session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pruneLocked(s.now())
	for _, candidate := range s.sessions {
		if candidate.id == id {
			return *candidate, true
		}
	}
	return session{}, false
}

// list returns the live sessions, oldest first.
func (s *sessionStore) list() []session {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pruneLocked(s.now())
	out := make([]session, 0, len(s.sessions))
	for _, candidate := range s.sessions {
		out = append(out, *candidate)
	}
	return out
}

// remove ends a session and reports whether it existed.
func (s *sessionStore) remove(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, candidate := range s.sessions {
		if candidate.id == id {
			s.sessions = append(s.sessions[:i], s.sessions[i+1:]...)
			return true
		}
	}
	return false
}

// removeAll ends every session.
func (s *sessionStore) removeAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions = nil
}

func (s *sessionStore) pruneLocked(now time.Time) {
	kept := s.sessions[:0]
	for _, candidate := range s.sessions {
		if now.Sub(candidate.lastUsed) < sessionIdleTimeout {
			kept = append(kept, candidate)
		}
	}
	s.sessions = kept
}
