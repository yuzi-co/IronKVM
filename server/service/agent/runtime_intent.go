package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"NanoKVM-Server/utils/atomicfile"
)

// IntentStore persists whether the agent should run, so the server can
// restore it at boot. title names the agent in its errors.
type IntentStore struct {
	mu    sync.Mutex
	path  string
	title string
}

// NewIntentStore returns a store that keeps the intent in path. title names
// the agent in errors, such as "PicoClaw".
func NewIntentStore(path, title string) *IntentStore {
	return &IntentStore{path: path, title: title}
}

func (s *IntentStore) Load() (RuntimeIntentStatus, error) {
	if s == nil {
		return RuntimeIntentStatus{}, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

func (s *IntentStore) SetDesiredRunning(desired bool, updatedBy string) error {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	intent, err := s.loadLocked()
	if err != nil {
		intent = RuntimeIntentStatus{
			DesiredRunning: false,
			LastError:      err.Error(),
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	intent.DesiredRunning = desired
	intent.UpdatedAt = now
	intent.UpdatedBy = updatedBy
	intent.LastError = ""
	if desired {
		intent.LastStartedAt = now
	} else {
		intent.LastStoppedAt = now
	}

	return s.saveLocked(intent)
}

func (s *IntentStore) SetLastError(message string) error {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	intent, err := s.loadLocked()
	if err != nil {
		intent = RuntimeIntentStatus{
			DesiredRunning: false,
		}
	}

	intent.LastError = message
	intent.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if intent.UpdatedBy == "" {
		intent.UpdatedBy = "system"
	}

	return s.saveLocked(intent)
}

func (s *IntentStore) loadLocked() (RuntimeIntentStatus, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return RuntimeIntentStatus{DesiredRunning: false}, nil
		}
		return RuntimeIntentStatus{DesiredRunning: false}, fmt.Errorf("read %s runtime intent: %w", s.title, err)
	}

	var intent RuntimeIntentStatus
	if err := json.Unmarshal(data, &intent); err != nil {
		return RuntimeIntentStatus{
			DesiredRunning: false,
			LastError:      err.Error(),
		}, fmt.Errorf("parse %s runtime intent: %w", s.title, err)
	}

	return intent, nil
}

func (s *IntentStore) saveLocked(intent RuntimeIntentStatus) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s runtime intent directory: %w", s.title, err)
	}

	data, err := json.MarshalIndent(intent, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s runtime intent: %w", s.title, err)
	}
	data = append(data, '\n')

	if err := atomicfile.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("replace %s runtime intent: %w", s.title, err)
	}

	directory, err := os.Open(dir)
	if err == nil {
		if syncErr := directory.Sync(); syncErr != nil {
			_ = directory.Close()
			return fmt.Errorf("sync %s runtime intent directory: %w", s.title, syncErr)
		}
		_ = directory.Close()
	}

	return nil
}
