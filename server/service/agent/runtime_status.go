package agent

import (
	"sync"
	"time"
)

// RuntimeStatus is what the web UI reads about the agent: its install, its
// process, its model, the persisted intent to run and the AI control mode.
// The JSON names are the ones the runtime routes have always returned.
type RuntimeStatus struct {
	Ready           bool                `json:"ready"`
	Installed       bool                `json:"installed"`
	Installing      bool                `json:"installing"`
	InstallProgress int                 `json:"install_progress,omitempty"`
	InstallStage    string              `json:"install_stage,omitempty"`
	InstallPath     string              `json:"install_path,omitempty"`
	Version         string              `json:"version,omitempty"`          // the release an install downloads
	UpdateAvailable bool                `json:"update_available,omitempty"` // the installed build is not that one
	AgentProfile    string              `json:"agent_profile,omitempty"`
	ModelConfigured bool                `json:"model_configured"`
	ModelName       string              `json:"model_name,omitempty"`
	Status          string              `json:"status"`
	ConfigError     string              `json:"config_error,omitempty"`
	LastError       string              `json:"last_error,omitempty"`
	CheckedAt       time.Time           `json:"checked_at,omitempty"`
	CurrentSession  string              `json:"current_session,omitempty"`
	Restoring       bool                `json:"restoring,omitempty"`
	RuntimeIntent   RuntimeIntentStatus `json:"runtime_intent"`
	ControlMode     string              `json:"control_mode"`
	Transitioning   bool                `json:"transitioning,omitempty"`
	Control         ControlStatus       `json:"control"`
	Capabilities    RuntimeCapabilities `json:"capabilities"`
}

// RuntimeIntentStatus is the persisted intent to run the agent, which the
// server restores at boot.
type RuntimeIntentStatus struct {
	DesiredRunning bool   `json:"desired_running"`
	UpdatedAt      string `json:"updated_at,omitempty"`
	UpdatedBy      string `json:"updated_by,omitempty"`
	LastStartedAt  string `json:"last_started_at,omitempty"`
	LastStoppedAt  string `json:"last_stopped_at,omitempty"`
	LastError      string `json:"last_error,omitempty"`
}

// ControlStatus is the AI control mode as the runtime status shows it.
type ControlStatus struct {
	Mode          string    `json:"mode"`
	Transitioning bool      `json:"transitioning"`
	CanControl    bool      `json:"can_control"`
	LastError     string    `json:"last_error,omitempty"`
	ChangedAt     time.Time `json:"changed_at,omitempty"`
}

// RuntimeCapabilities says what the agent may do now.
type RuntimeCapabilities struct {
	Chat          bool `json:"chat"`
	ReadOnlyTools bool `json:"read_only_tools"`
	DeviceWrite   bool `json:"device_write"`
}

// RuntimeStartResult answers the start and stop routes.
type RuntimeStartResult struct {
	Started bool          `json:"started"`
	Command string        `json:"command"`
	Output  string        `json:"output,omitempty"`
	Status  RuntimeStatus `json:"status"`
}

// RuntimeInstallResult answers the install and uninstall routes.
type RuntimeInstallResult struct {
	Installed bool          `json:"installed"`
	Binary    string        `json:"binary"`
	Output    string        `json:"output,omitempty"`
	Status    RuntimeStatus `json:"status"`
}

// StatusStore holds the agent's runtime status. The Lifecycle records into
// it, and so does the agent's Runtime for the detail only it knows. Every
// status it holds has the agent's install path.
type StatusStore struct {
	mu          sync.RWMutex
	installPath string
	status      RuntimeStatus
}

// NewStatusStore returns a store holding status, with installPath as its
// install path.
func NewStatusStore(installPath string, status RuntimeStatus) *StatusStore {
	status.InstallPath = installPath
	return &StatusStore{installPath: installPath, status: status}
}

// InstallPath is the install path every status of the store has.
func (s *StatusStore) InstallPath() string {
	if s == nil {
		return ""
	}
	return s.installPath
}

func (s *StatusStore) Get() RuntimeStatus {
	if s == nil {
		return RuntimeStatus{
			Ready:     false,
			Installed: false,
			Status:    "unavailable",
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *StatusStore) Set(status RuntimeStatus) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	status.InstallPath = s.installPath
	s.status = status
}

// SetFromProbe records a status a passive check found. While a start, stop or
// restore is pending, a check that did not find the agent ready only updates
// when it looked, so it cannot overwrite the pending state.
func (s *StatusStore) SetFromProbe(status RuntimeStatus) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if IsLifecyclePending(s.status) && !status.Ready {
		s.status.CheckedAt = status.CheckedAt
		s.status.CurrentSession = status.CurrentSession
		s.status.InstallPath = s.installPath
		return
	}

	status.InstallPath = s.installPath
	s.status = status
}

func (s *StatusStore) Update(update func(*RuntimeStatus)) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	update(&s.status)
	s.status.InstallPath = s.installPath
}

// IsLifecyclePending reports whether a start, stop or restore is under way.
func IsLifecyclePending(status RuntimeStatus) bool {
	if status.Restoring {
		return true
	}
	switch status.Status {
	case "starting", "restoring", "stopping":
		return true
	default:
		return false
	}
}
