package hdmi_state

import "time"

type viewerSource struct {
	count   int
	version uint64
}

// State stores capture demand. Callers must serialize access to State.
type State struct {
	viewerCount     int
	leaseCount      int
	sources         map[string]viewerSource
	readyAt         time.Time
	needsFreshFrame bool
}

func New() State {
	return State{sources: make(map[string]viewerSource)}
}

// UpdateViewer accepts only newer source snapshots so a delayed report cannot
// overwrite a newer client-map snapshot.
func (s *State) UpdateViewer(source string, count int, version uint64) bool {
	if count < 0 {
		count = 0
	}
	if current, ok := s.sources[source]; ok && version <= current.version {
		return false
	}

	s.sources[source] = viewerSource{count: count, version: version}
	s.viewerCount = 0
	for _, current := range s.sources {
		s.viewerCount += current.count
	}
	return true
}

func (s *State) NextVersion(source string) uint64 {
	return s.sources[source].version + 1
}

func (s *State) AcquireLease() {
	s.leaseCount++
}

func (s *State) ReleaseLease() bool {
	if s.leaseCount == 0 {
		return false
	}
	s.leaseCount--
	return true
}

func (s *State) HasDemand() bool {
	return s.viewerCount > 0 || s.leaseCount > 0
}

func (s *State) ViewerCount() int {
	return s.viewerCount
}

func (s *State) LeaseCount() int {
	return s.leaseCount
}

func (s *State) MarkWarming(readyAt time.Time) {
	s.readyAt = readyAt
	s.needsFreshFrame = true
}

func (s *State) ClearReadyAt() {
	s.readyAt = time.Time{}
	s.needsFreshFrame = false
}

func (s *State) ReadyAt() time.Time {
	return s.readyAt
}

func (s *State) NeedsFreshFrame() bool {
	return s.needsFreshFrame
}

// QueueStale reports whether the head of the capture queue predates a read
// about to start. While no stream has a viewer nothing drains the queue
// between one-shot reads, so its head is the frame captured just after the
// previous read, however long ago that was; the frame behind it is current.
func (s *State) QueueStale() bool {
	return s.viewerCount == 0
}

// ClaimFreshFrame returns true for the first successful reader after HDMI
// resumes. Failed reads do not consume the claim.
func (s *State) ClaimFreshFrame() bool {
	if !s.needsFreshFrame {
		return false
	}
	s.needsFreshFrame = false
	return true
}

func (s *State) Viewer(source string) (count int, version uint64, ok bool) {
	value, ok := s.sources[source]
	return value.count, value.version, ok
}
