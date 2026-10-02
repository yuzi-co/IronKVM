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
// channels of Session.
type Agent struct {
	mu      sync.Mutex
	running bool
	// Configured counts ApplyConfig calls.
	Configured int
	// OpenErr, when set, makes OpenSession fail.
	OpenErr error
	stored  map[string]agent.SessionDetail
	// Opened receives every session OpenSession returns.
	Opened chan *Session
}

var _ agent.Agent = (*Agent)(nil)

// New returns a stopped agent with no stored sessions.
func New() *Agent {
	return &Agent{
		stored: make(map[string]agent.SessionDetail),
		Opened: make(chan *Session, 16),
	}
}

func (a *Agent) Name() string { return "agenttest" }

func (a *Agent) Start(ctx context.Context) error {
	if err := a.ApplyConfig(ctx); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.running = true
	return nil
}

func (a *Agent) Stop(context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.running = false
	return nil
}

func (a *Agent) Ready(context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.running {
		return agent.NewError(agent.CodeRuntimeUnavailable, "agent is stopped")
	}
	return nil
}

func (a *Agent) ApplyConfig(context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Configured++
	return nil
}

func (a *Agent) OpenSession(_ context.Context, sessionID string) (agent.Session, error) {
	a.mu.Lock()
	openErr := a.OpenErr
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
}

var _ agent.Session = (*Session)(nil)

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
