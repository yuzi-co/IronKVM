package agent

import (
	"strings"
	"sync"
	"time"
)

const (
	// DefaultTaskLeaseDuration bounds a task lease whose prompt set no limit.
	DefaultTaskLeaseDuration = 2 * time.Minute
	// MaxTaskLeaseDuration bounds any task lease.
	MaxTaskLeaseDuration = 30 * time.Minute
)

// TaskLeases keeps HDMI capture running while the agent works on a prompt, so
// its screenshots do not wait for capture to start. A lease starts with a
// prompt and lasts for the whole turn. It ends when:
//   - turn_done names its request (or names none: all of the session's),
//   - an error names its request (or names none: all of the session's),
//   - the browser cancels it, the session closes, or its timer runs out,
//   - a reply arrives from an agent that has not sent turn_done in this
//     session, the same rule the sidebar uses for its run state.
type TaskLeases struct {
	acquire func() func()

	mu       sync.Mutex
	leases   map[string]func()
	timers   map[string]*time.Timer
	turnDone map[string]bool
}

// NewTaskLeases returns a lease tracker that takes a capture lease with
// acquire. acquire returns the function that releases the lease.
func NewTaskLeases(acquire func() func()) *TaskLeases {
	return &TaskLeases{
		acquire:  acquire,
		leases:   make(map[string]func()),
		timers:   make(map[string]*time.Timer),
		turnDone: make(map[string]bool),
	}
}

// Prompt starts the lease for a prompt. maxRuntimeMs is the prompt's own
// limit, 0 when unset.
func (l *TaskLeases) Prompt(sessionID, requestID string, maxRuntimeMs int) {
	if l == nil || sessionID == "" {
		return
	}
	duration := DefaultTaskLeaseDuration
	if maxRuntimeMs > 0 {
		maxMilliseconds := int(MaxTaskLeaseDuration / time.Millisecond)
		duration = time.Duration(min(maxRuntimeMs, maxMilliseconds)) * time.Millisecond
	}

	key := taskLeaseKey(sessionID, requestID)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.leases[key] == nil && l.acquire != nil {
		l.leases[key] = l.acquire()
	}
	if timer := l.timers[key]; timer != nil {
		timer.Stop()
	}
	l.timers[key] = time.AfterFunc(duration, func() {
		l.release(key)
	})
}

// Cancel ends the lease of requestID, or every lease of the session when
// requestID is empty.
func (l *TaskLeases) Cancel(sessionID, requestID string) {
	if requestID == "" {
		l.ReleaseSession(sessionID)
		return
	}
	l.releaseRequest(sessionID, requestID)
}

// Observe ends leases as the agent's events say.
func (l *TaskLeases) Observe(sessionID string, event Event) {
	if l == nil || sessionID == "" {
		return
	}
	switch event.Type {
	case EventTurnDone:
		l.mu.Lock()
		l.turnDone[sessionID] = true
		l.mu.Unlock()
		if len(event.RequestIDs) == 0 {
			l.ReleaseSession(sessionID)
			return
		}
		for _, id := range event.RequestIDs {
			l.releaseRequest(sessionID, id)
		}
	case EventError:
		if event.RequestID != "" {
			l.releaseRequest(sessionID, event.RequestID)
			return
		}
		l.ReleaseSession(sessionID)
	case EventAgentMessage:
		if event.Kind != KindReply {
			return
		}
		l.mu.Lock()
		sendsTurnDone := l.turnDone[sessionID]
		l.mu.Unlock()
		if !sendsTurnDone {
			l.ReleaseSession(sessionID)
		}
	}
}

// ReleaseSession ends every lease of the session.
func (l *TaskLeases) ReleaseSession(sessionID string) {
	if l == nil || sessionID == "" {
		return
	}
	prefix := "task:" + sessionID + ":"
	l.mu.Lock()
	keys := make([]string, 0, len(l.leases))
	for key := range l.leases {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	l.mu.Unlock()
	for _, key := range keys {
		l.release(key)
	}
}

// Forget ends every lease of a closed session and drops what the tracker
// knows about it.
func (l *TaskLeases) Forget(sessionID string) {
	if l == nil || sessionID == "" {
		return
	}
	l.ReleaseSession(sessionID)
	l.mu.Lock()
	delete(l.turnDone, sessionID)
	l.mu.Unlock()
}

func (l *TaskLeases) releaseRequest(sessionID, requestID string) {
	if l == nil || sessionID == "" {
		return
	}
	l.release(taskLeaseKey(sessionID, requestID))
}

func (l *TaskLeases) release(key string) {
	l.mu.Lock()
	release := l.leases[key]
	delete(l.leases, key)
	if timer := l.timers[key]; timer != nil {
		timer.Stop()
		delete(l.timers, key)
	}
	l.mu.Unlock()

	if release != nil {
		release()
	}
}

func taskLeaseKey(sessionID, requestID string) string {
	if requestID == "" {
		requestID = "default"
	}
	return "task:" + sessionID + ":" + requestID
}
