package middleware

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const SessionRevokedCloseCode = 4401

type sessionRegistry struct {
	mutex    sync.Mutex
	nextID   atomic.Uint64
	byUserID map[string]map[uint64]context.CancelFunc
}

var activeSessions = &sessionRegistry{byUserID: make(map[string]map[uint64]context.CancelFunc)}

func (r *sessionRegistry) register(username string, cancel context.CancelFunc) func() {
	id := r.nextID.Add(1)
	r.mutex.Lock()
	if r.byUserID[username] == nil {
		r.byUserID[username] = make(map[uint64]context.CancelFunc)
	}
	r.byUserID[username][id] = cancel
	r.mutex.Unlock()

	return func() {
		r.mutex.Lock()
		delete(r.byUserID[username], id)
		if len(r.byUserID[username]) == 0 {
			delete(r.byUserID, username)
		}
		r.mutex.Unlock()
	}
}

func RevokeUserSessions(username string) {
	activeSessions.mutex.Lock()
	sessions := activeSessions.byUserID[username]
	delete(activeSessions.byUserID, username)
	activeSessions.mutex.Unlock()

	for _, cancel := range sessions {
		cancel()
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
