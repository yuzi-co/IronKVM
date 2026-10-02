package picoclaw

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"NanoKVM-Server/service/vm"
)

const defaultTaskCaptureLeaseDuration = 2 * time.Minute
const maxTaskCaptureLeaseDuration = 30 * time.Minute

func (s *Service) acquireCaptureLease(ctx context.Context) (func(), func() bool, error) {
	if s != nil && s.acquireHDMIForRead != nil {
		return s.acquireHDMIForRead(ctx)
	}
	return vm.AcquireHdmiCaptureLeaseForRead(ctx)
}

func (s *Service) activateTaskCaptureLease(sessionID string, taskID string, duration time.Duration) {
	if s == nil || sessionID == "" {
		return
	}
	if duration <= 0 {
		duration = defaultTaskCaptureLeaseDuration
	}
	if duration > maxTaskCaptureLeaseDuration {
		duration = maxTaskCaptureLeaseDuration
	}

	key := taskCaptureLeaseKey(sessionID, taskID)
	s.captureLeaseMu.Lock()
	defer s.captureLeaseMu.Unlock()
	if s.captureLeases == nil {
		s.captureLeases = make(map[string]func())
	}
	if s.captureLeaseTimers == nil {
		s.captureLeaseTimers = make(map[string]*time.Timer)
	}

	if s.captureLeases[key] == nil {
		if s.acquireHDMILease != nil {
			s.captureLeases[key] = s.acquireHDMILease()
		} else {
			s.captureLeases[key] = vm.AcquireHdmiCaptureLease()
		}
	}
	if timer := s.captureLeaseTimers[key]; timer != nil {
		timer.Stop()
	}
	s.captureLeaseTimers[key] = time.AfterFunc(duration, func() {
		s.releaseCaptureLease(key)
	})
}

func (s *Service) updateTaskCaptureLease(source string, sessionID string, data []byte) {
	if s == nil || sessionID == "" || len(data) == 0 {
		return
	}

	var message struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Payload struct {
			MaxRuntimeMS int      `json:"max_runtime_ms"`
			RequestID    string   `json:"request_id"`
			RequestIDs   []string `json:"request_ids"`
			Kind         string   `json:"kind"`
			Thought      bool     `json:"thought"`
			Placeholder  bool     `json:"placeholder"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(data, &message); err != nil {
		return
	}

	// A task lease lasts for the whole turn. PicoClaw sends no top-level id on
	// its messages, so releasing on every message.create or typing.stop, as
	// this did before, let the lease go at the agent's first reasoning or
	// tool-call message, long before the turn ended.
	switch {
	case source == "downstream" && message.Type == "message.send":
		duration := defaultTaskCaptureLeaseDuration
		if message.Payload.MaxRuntimeMS > 0 {
			maxMilliseconds := int(maxTaskCaptureLeaseDuration / time.Millisecond)
			duration = time.Duration(min(message.Payload.MaxRuntimeMS, maxMilliseconds)) * time.Millisecond
		}
		s.activateTaskCaptureLease(sessionID, message.ID, duration)
	case source == "downstream" && message.Type == "message.cancel":
		if message.ID == "" {
			s.releaseCaptureLeasesForSession(sessionID)
			return
		}
		s.releaseTaskCaptureLease(sessionID, message.ID)
	case source == "upstream" && message.Type == "turn.done":
		// The IronKVM PicoClaw build ends every turn with turn.done, naming
		// the message.send requests the turn handled.
		s.markSessionSendsTurnDone(sessionID)
		ids := message.Payload.RequestIDs
		if len(ids) == 0 && message.Payload.RequestID != "" {
			ids = []string{message.Payload.RequestID}
		}
		if len(ids) == 0 {
			s.releaseCaptureLeasesForSession(sessionID)
			return
		}
		for _, id := range ids {
			s.releaseTaskCaptureLease(sessionID, id)
		}
	case source == "upstream" && message.Type == "error":
		// An error about one request names it; any other error ends the
		// session's work.
		if message.Payload.RequestID != "" {
			s.releaseTaskCaptureLease(sessionID, message.Payload.RequestID)
			return
		}
		s.releaseCaptureLeasesForSession(sessionID)
	case source == "upstream" && (message.Type == "message.create" || message.Type == "message.update"):
		// A build without turn.done ends a turn with its reply. Reasoning,
		// tool calls and placeholders do not end it.
		p := message.Payload
		if p.Kind == "thought" || p.Kind == "tool_calls" || p.Thought || p.Placeholder {
			return
		}
		if !s.sessionSendsTurnDone(sessionID) {
			s.releaseCaptureLeasesForSession(sessionID)
		}
	}
}

func (s *Service) markSessionSendsTurnDone(sessionID string) {
	s.captureLeaseMu.Lock()
	defer s.captureLeaseMu.Unlock()
	if s.turnDoneSessions == nil {
		s.turnDoneSessions = make(map[string]bool)
	}
	s.turnDoneSessions[sessionID] = true
}

func (s *Service) sessionSendsTurnDone(sessionID string) bool {
	s.captureLeaseMu.Lock()
	defer s.captureLeaseMu.Unlock()
	return s.turnDoneSessions[sessionID]
}

// forgetTaskCaptureSession drops what the lease tracker knows about a closed
// session, after releasing its leases.
func (s *Service) forgetTaskCaptureSession(sessionID string) {
	if s == nil || sessionID == "" {
		return
	}
	s.releaseCaptureLeasesForSession(sessionID)
	s.captureLeaseMu.Lock()
	delete(s.turnDoneSessions, sessionID)
	s.captureLeaseMu.Unlock()
}

func taskCaptureLeaseKey(sessionID string, taskID string) string {
	if taskID == "" {
		taskID = "default"
	}
	return "task:" + sessionID + ":" + taskID
}

func (s *Service) releaseTaskCaptureLease(sessionID string, taskID string) {
	if sessionID == "" {
		return
	}
	s.releaseCaptureLease(taskCaptureLeaseKey(sessionID, taskID))
}

func (s *Service) releaseCaptureLeasesForSession(sessionID string) {
	if s == nil || sessionID == "" {
		return
	}
	prefix := "task:" + sessionID + ":"
	s.captureLeaseMu.Lock()
	keys := make([]string, 0, len(s.captureLeases))
	for key := range s.captureLeases {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	s.captureLeaseMu.Unlock()
	for _, key := range keys {
		s.releaseCaptureLease(key)
	}
}

func (s *Service) releaseCaptureLease(key string) {
	if s == nil || key == "" {
		return
	}
	s.captureLeaseMu.Lock()
	release := s.captureLeases[key]
	delete(s.captureLeases, key)
	if timer := s.captureLeaseTimers[key]; timer != nil {
		timer.Stop()
		delete(s.captureLeaseTimers, key)
	}
	s.captureLeaseMu.Unlock()

	if release != nil {
		release()
	}
}
