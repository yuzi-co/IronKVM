package middleware

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"NanoKVM-Server/authn"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const SessionRevokedCloseCode = 4401

// session is one authenticated request that is still running, such as a
// websocket.
type session struct {
	username     string
	tokenVersion uint64
	// store is the account store that authenticated the request. The recheck
	// asks the same one rather than reading authn.DefaultStore again.
	store  *authn.Store
	cancel context.CancelFunc

	// sessionID names the login session the request came through, and
	// sessions is the store that holds it. The recheck asks that too, and
	// RevokeLoginSession ends the requests that carry the id.
	sessionID string
	sessions  *authn.SessionStore

	// A session opened from a share link has no account. shareID names the
	// link and shares is the store that holds it, and the recheck asks that.
	shareID string
	shares  *authn.ShareStore
}

// sessionRegistry holds the running sessions, and while there are any, one
// goroutine that rechecks them all every interval against the account store.
// RevokeUserSessions ends a user's sessions at once and covers every change
// made through the server. The recheck is for the rest: an account file
// removed by the reset button or edited from a shell. It costs one stat of
// the file per distinct token while the file is unchanged.
type sessionRegistry struct {
	mutex    sync.Mutex
	nextID   atomic.Uint64
	byUserID map[string]map[uint64]session
	interval time.Duration
	watching bool
}

func newSessionRegistry(interval time.Duration) *sessionRegistry {
	return &sessionRegistry{byUserID: make(map[string]map[uint64]session), interval: interval}
}

var activeSessions = newSessionRegistry(sessionRecheckDelay)

func (r *sessionRegistry) register(entry session) func() {
	id := r.nextID.Add(1)
	r.mutex.Lock()
	if r.byUserID[entry.username] == nil {
		r.byUserID[entry.username] = make(map[uint64]session)
	}
	r.byUserID[entry.username][id] = entry
	if !r.watching && (entry.store != nil || entry.shares != nil) {
		r.watching = true
		go r.watch()
	}
	r.mutex.Unlock()

	return func() {
		r.mutex.Lock()
		delete(r.byUserID[entry.username], id)
		if len(r.byUserID[entry.username]) == 0 {
			delete(r.byUserID, entry.username)
		}
		r.mutex.Unlock()
	}
}

// watch rechecks the sessions until none are left. register starts it again
// with the next session.
func (r *sessionRegistry) watch() {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for range ticker.C {
		r.recheck()

		r.mutex.Lock()
		if len(r.byUserID) == 0 {
			r.watching = false
			r.mutex.Unlock()
			return
		}
		r.mutex.Unlock()
	}
}

// recheck cancels every session whose token the store no longer accepts. The
// store is asked outside the registry's lock, once per distinct token.
func (r *sessionRegistry) recheck() {
	r.mutex.Lock()
	sessions := make([]session, 0, len(r.byUserID))
	for _, byID := range r.byUserID {
		for _, entry := range byID {
			sessions = append(sessions, entry)
		}
	}
	r.mutex.Unlock()

	type tokenKey struct {
		store        *authn.Store
		username     string
		tokenVersion uint64
	}
	valid := make(map[tokenKey]bool)
	validShare := make(map[string]bool)
	for _, entry := range sessions {
		if entry.shares != nil {
			ok, seen := validShare[entry.shareID]
			if !seen {
				_, ok = entry.shares.Valid(entry.shareID)
				validShare[entry.shareID] = ok
			}
			if !ok {
				entry.cancel()
			}
			continue
		}
		if entry.store == nil {
			continue
		}
		if entry.sessions != nil && !entry.sessions.Valid(entry.sessionID) {
			entry.cancel()
			continue
		}
		key := tokenKey{entry.store, entry.username, entry.tokenVersion}
		ok, seen := valid[key]
		if !seen {
			_, err := entry.store.ValidateToken(entry.username, entry.tokenVersion)
			ok = err == nil
			valid[key] = ok
		}
		if !ok {
			entry.cancel()
		}
	}
}

// RevokeUserSessions ends every session of a user: the login session records
// and the requests running on them. Every account change that should sign the
// user out calls it, including the ones that leave the token version alone.
func RevokeUserSessions(username string) {
	if _, err := authn.DefaultSessions.RevokeUser(username, ""); err != nil {
		log.Warnf("revoke the login sessions of %s: %s", username, err)
	}
	activeSessions.revoke(username, "")
}

// RevokeOtherSessions ends every session of a user but keepID.
func RevokeOtherSessions(username, keepID string) {
	if _, err := authn.DefaultSessions.RevokeUser(username, keepID); err != nil {
		log.Warnf("revoke the other login sessions of %s: %s", username, err)
	}
	activeSessions.revoke(username, keepID)
}

// RevokeLoginSession ends one login session and the requests running on it,
// websockets included. The error is the store's: the session is gone from
// memory either way, but a failed write may let it back after a restart.
func RevokeLoginSession(id string) error {
	_, err := authn.DefaultSessions.Revoke(id)
	activeSessions.revokeSession(id)
	if errors.Is(err, authn.ErrSessionNotFound) {
		return err
	}
	if err != nil {
		log.Warnf("revoke login session: %s", err)
	}
	return err
}

// revoke cancels a user's requests, except those of the session keepID when
// it is not empty.
func (r *sessionRegistry) revoke(username, keepID string) {
	r.mutex.Lock()
	var cancelled []session
	for id, entry := range r.byUserID[username] {
		if keepID != "" && entry.sessionID == keepID {
			continue
		}
		cancelled = append(cancelled, entry)
		delete(r.byUserID[username], id)
	}
	if len(r.byUserID[username]) == 0 {
		delete(r.byUserID, username)
	}
	r.mutex.Unlock()

	for _, entry := range cancelled {
		entry.cancel()
	}
}

// revokeSession cancels the requests of one login session.
func (r *sessionRegistry) revokeSession(sessionID string) {
	if sessionID == "" {
		return
	}
	r.mutex.Lock()
	var cancelled []session
	for username, byID := range r.byUserID {
		for id, entry := range byID {
			if entry.sessionID == sessionID {
				cancelled = append(cancelled, entry)
				delete(byID, id)
			}
		}
		if len(byID) == 0 {
			delete(r.byUserID, username)
		}
	}
	r.mutex.Unlock()

	for _, entry := range cancelled {
		entry.cancel()
	}
}

func WatchWebSocket(ctx context.Context, connection *websocket.Conn) func() {
	stopped := make(chan struct{})
	var stopOnce sync.Once
	go func() {
		select {
		case <-ctx.Done():
			_ = connection.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(SessionRevokedCloseCode, "session expired or revoked"),
				time.Now().Add(2*time.Second),
			)
			_ = connection.Close()
		case <-stopped:
		}
	}()
	return func() { stopOnce.Do(func() { close(stopped) }) }
}
