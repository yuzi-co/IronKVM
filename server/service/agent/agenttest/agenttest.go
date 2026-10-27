// Package agenttest is an agent adapter for tests. It shows what a second
// agent needs to provide: the agent.Agent interface and nothing else.
package agenttest

import (
	"context"
	"sort"
	"sync"

	"NanoKVM-Server/service/agent"
)

// Agent is an in-memory agent. Tests drive its sessions through the
// channels of Session, and its runtime through its exported fields.
type Agent struct {
	mu        sync.Mutex
	installed bool
	running   bool
	// Configured counts ApplyConfig calls.
	Configured int
	// Launches and Halts count Launch and Halt calls.
	Launches, Halts int
	// LaunchErr, when set, makes Launch fail.
	LaunchErr error
	// NeverReady keeps a running agent from answering Check.
	NeverReady bool
	// StaleInstall makes Stale report an update.
	StaleInstall bool
	// InstallErr, when set, makes Install fail.
	InstallErr error
	// Uninstalled counts Uninstall calls.
	Uninstalled int
	// OpenErr, when set, makes OpenSession fail.
	OpenErr error
	// TurnDone makes the sessions OpenSession returns say that the agent
	// ends every turn with turn_done (agent.TurnReporter).
	TurnDone bool
	stored   map[string]agent.SessionDetail
	// Opened receives every session OpenSession returns.
	Opened chan *Session
}

var (
	_ agent.Agent = (*Agent)(nil)
)

// New returns an installed, stopped agent with no stored sessions.
func New() *Agent {
	return &Agent{
		installed: true,
		stored:    make(map[string]agent.SessionDetail),
		Opened:    make(chan *Session, 16),
	}
}

func (a *Agent) Name() string { return "agenttest" }

// InstallPath is where the agent claims to be installed.
const InstallPath = "/usr/bin/agenttest"

// Version is the release the agent claims Install installs.
const Version = "v1.0.0-test"

func (a *Agent) InstallPath() string { return InstallPath }
func (a *Agent) Version() string     { return Version }

func (a *Agent) Installed() (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.installed, nil
}

// SetInstalled installs or removes the agent without Install.
func (a *Agent) SetInstalled(installed bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.installed = installed
}

func (a *Agent) Stale() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.StaleInstall
}

func (a *Agent) Install(progress func(stage string, percent int, lastError string)) error {
	progress("downloading", 50, "")
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.InstallErr != nil {
		return a.InstallErr
	}
	a.installed = true
	a.StaleInstall = false
	return nil
}

func (a *Agent) Uninstall() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.installed = false
	a.Uninstalled++
}

func (a *Agent) Model() (agent.Model, error) {
	return agent.Model{Configured: true, Name: "test-model", Target: "test-model"}, nil
}

func (a *Agent) ApplyConfig(context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Configured++
	return nil
}

func (a *Agent) Running() (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.running, nil
}

// SetRunning starts or stops the process without Launch or Halt.
func (a *Agent) SetRunning(running bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.running = running
}

func (a *Agent) Launch(ctx context.Context) (string, string, error) {
	if err := a.ApplyConfig(ctx); err != nil {
		return "", "", err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Launches++
	if a.LaunchErr != nil {
		return "agenttest start", "", a.LaunchErr
	}
	if !a.installed {
		return "", "", agent.NewError(agent.CodeRuntimeStartFailed, "agenttest is not installed")
	}
	a.running = true
	return "agenttest start", "started", nil
}

// Counts reads Launches and Halts under the lock, for a test that checks
// them while the lifecycle may still be calling Launch or Halt.
func (a *Agent) Counts() (launches, halts int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.Launches, a.Halts
}

func (a *Agent) Halt() (string, string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Halts++
	a.running = false
	return "agenttest stop", "", nil
}

func (a *Agent) Check(context.Context) (*agent.RuntimeStatus, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	status := &agent.RuntimeStatus{Installed: a.installed, ModelConfigured: true, ModelName: "test-model"}
	switch {
	case !a.installed:
		status.Installed = false
		status.ModelConfigured = false
		status.ModelName = ""
		status.Status = "not_installed"
		return status, agent.NewError(agent.CodeRuntimeUnavailable, "agenttest is not installed")
	case !a.running:
		status.Status = "stopped"
		return status, agent.NewError(agent.CodeRuntimeUnavailable, "agenttest runtime is stopped")
	case a.NeverReady:
		status.Status = "unavailable"
		return status, agent.NewError(agent.CodeRuntimeUnavailable, "agenttest is not answering")
	}
	status.Ready = true
	status.InstallProgress = 100
	status.Status = "ready"
	return status, nil
}

func (a *Agent) Describe(status agent.RuntimeStatus) agent.RuntimeStatus {
	status.AgentProfile = "test"
	return status
}

func (a *Agent) OpenSession(_ context.Context, sessionID string) (agent.Session, error) {
	a.mu.Lock()
	openErr := a.OpenErr
	turnDone := a.TurnDone
	a.mu.Unlock()
	if openErr != nil {
		return nil, openErr
	}
	session := &Session{
		ID:      sessionID,
		Prompts: make(chan agent.Prompt, 16),
		Cancels: make(chan string, 16),
		events:  make(chan agent.Event, 16),
		ended:   make(chan error, 1),
		closed:  make(chan struct{}),

		turnDone: turnDone,
	}
	a.Opened <- session
	return session, nil
}

// Store adds a stored session to the history.
func (a *Agent) Store(detail agent.SessionDetail) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stored[detail.ID] = detail
}

func (a *Agent) ListSessions(context.Context) ([]agent.SessionSummary, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	items := make([]agent.SessionSummary, 0, len(a.stored))
	for _, detail := range a.stored {
		items = append(items, agent.SessionSummary{
			ID:           detail.ID,
			Title:        detail.Summary,
			MessageCount: len(detail.Messages),
			Created:      detail.Created,
			Updated:      detail.Updated,
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Updated > items[j].Updated })
	return items, nil
}

func (a *Agent) ReadSession(_ context.Context, id string) (agent.SessionDetail, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	detail, ok := a.stored[id]
	if !ok {
		return agent.SessionDetail{}, agent.ErrSessionNotFound
	}
	return detail, nil
}

func (a *Agent) DeleteSession(_ context.Context, id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.stored[id]; !ok {
		return agent.ErrSessionNotFound
	}
	delete(a.stored, id)
	return nil
}

// Session is one open chat session of Agent.
type Session struct {
	ID string
	// Prompts and Cancels receive what the browser sent.
	Prompts chan agent.Prompt
	Cancels chan string

	events    chan agent.Event
	ended     chan error
	closed    chan struct{}
	closeOnce sync.Once
	turnDone  bool
}

var (
	_ agent.Session      = (*Session)(nil)
	_ agent.TurnReporter = (*Session)(nil)
)

// ReportsTurnDone is the agent's TurnDone when the session opened.
func (s *Session) ReportsTurnDone() bool { return s.turnDone }

// Emit makes the agent send an event.
func (s *Session) Emit(event agent.Event) { s.events <- event }

// End makes the agent end the session with err.
func (s *Session) End(err error) { s.ended <- err }

// Closed is closed once the server closed the session.
func (s *Session) Closed() <-chan struct{} { return s.closed }

func (s *Session) Prompt(_ context.Context, prompt agent.Prompt) error {
	s.Prompts <- prompt
	return nil
}

func (s *Session) Cancel(_ context.Context, requestID string) error {
	s.Cancels <- requestID
	return nil
}

func (s *Session) Next() (agent.Event, error) {
	select {
	case event := <-s.events:
		return event, nil
	case err := <-s.ended:
		return agent.Event{}, err
	case <-s.closed:
		return agent.Event{}, &agent.CloseError{Code: agent.CloseUpstreamClosed, Reason: "session closed"}
	}
}

func (s *Session) Close() error {
	s.closeOnce.Do(func() { close(s.closed) })
	return nil
}
