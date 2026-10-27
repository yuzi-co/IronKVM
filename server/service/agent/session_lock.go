package agent

import (
	"errors"
	"sync"
	"time"
)

var (
	sessionLockOnce sync.Once
	sessionLock     *SessionLock
)

const sessionLockDuration = 30 * time.Minute

// ErrNoSessionID refuses the AI input lock to a request that names no chat
// session.
var ErrNoSessionID = errors.New("no chat session id")

// LockHeldError refuses the AI input lock to a session while another one
// holds it.
type LockHeldError struct {
	// Owner is the session that holds the lock.
	Owner string
}

func (e *LockHeldError) Error() string {
	return "another chat session holds the AI input lock"
}

// SessionLock is the AI input lock: one chat session at a time may have the
// agent write to the host, and while one holds it the browser's own input
// waits. There is one lock whichever agent runs (GetSessionLock).
type SessionLock struct {
	mu             sync.Mutex
	ownerSessionID string
	acquiredAt     time.Time
	expiresAt      time.Time
}

// GetSessionLock returns the server's AI input lock.
func GetSessionLock() *SessionLock {
	sessionLockOnce.Do(func() {
		sessionLock = &SessionLock{}
	})

	return sessionLock
}

// Ensure takes the lock for sessionID or extends its hold. The error is
// ErrNoSessionID or a *LockHeldError.
func (l *SessionLock) Ensure(sessionID string) error {
	_, err := l.acquire(sessionID)
	return err
}

// AcquireTemporary is Ensure that also reports whether the lock was free, so
// that the caller releases it again when done.
func (l *SessionLock) AcquireTemporary(sessionID string) (bool, error) {
	return l.acquire(sessionID)
}

func (l *SessionLock) Renew(sessionID string) bool {
	if sessionID == "" {
		return false
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.ownerSessionID != sessionID {
		return false
	}

	l.expiresAt = time.Now().Add(sessionLockDuration)
	return true
}

func (l *SessionLock) clearExpiredLocked(now time.Time) {
	if l.ownerSessionID == "" {
		return
	}
	if l.expiresAt.IsZero() || now.Before(l.expiresAt) {
		return
	}

	l.ownerSessionID = ""
	l.acquiredAt = time.Time{}
	l.expiresAt = time.Time{}
}

func (l *SessionLock) acquire(sessionID string) (bool, error) {
	if sessionID == "" {
		return false, ErrNoSessionID
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	l.clearExpiredLocked(now)

	if l.ownerSessionID == "" || l.ownerSessionID == sessionID {
		acquired := l.ownerSessionID == ""
		l.ownerSessionID = sessionID
		if l.acquiredAt.IsZero() {
			l.acquiredAt = now
		}
		l.expiresAt = now.Add(sessionLockDuration)
		return acquired, nil
	}

	return false, &LockHeldError{Owner: l.ownerSessionID}
}

func (l *SessionLock) Release(sessionID string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.ownerSessionID == "" {
		return true
	}
	if sessionID != "" && l.ownerSessionID != sessionID {
		return false
	}

	l.ownerSessionID = ""
	l.acquiredAt = time.Time{}
	l.expiresAt = time.Time{}
	return true
}

// ReleaseOwned releases the lock only when sessionID is the current owner.
// Unlike Release, an already-empty lock is not considered a successful
// release. This distinction prevents stale session cleanup from releasing HID
// state that may now belong to another controller.
func (l *SessionLock) ReleaseOwned(sessionID string) bool {
	if sessionID == "" {
		return false
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.clearExpiredLocked(time.Now())
	if l.ownerSessionID != sessionID {
		return false
	}

	l.ownerSessionID = ""
	l.acquiredAt = time.Time{}
	l.expiresAt = time.Time{}
	return true
}

func (l *SessionLock) ForceTakeover(sessionID string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	l.ownerSessionID = sessionID
	l.acquiredAt = now
	l.expiresAt = now.Add(sessionLockDuration)
}

func (l *SessionLock) Owner() string {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.clearExpiredLocked(time.Now())
	return l.ownerSessionID
}

func (l *SessionLock) BlocksManualInput() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.clearExpiredLocked(time.Now())
	return l.ownerSessionID != ""
}
