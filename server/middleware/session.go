package middleware

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"NanoKVM-Server/authn"

	"github.com/gorilla/websocket"
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
	if !r.watching && entry.store != nil {
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
	for _, entry := range sessions {
		if entry.store == nil {
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

func RevokeUserSessions(username string) {
	activeSessions.revoke(username)
}

func (r *sessionRegistry) revoke(username string) {
	r.mutex.Lock()
	sessions := r.byUserID[username]
	delete(r.byUserID, username)
	r.mutex.Unlock()

	for _, entry := range sessions {
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
