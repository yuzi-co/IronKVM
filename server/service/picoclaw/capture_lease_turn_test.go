package picoclaw

import "testing"

func TestTurnDoneReleasesTheTurnsCaptureLeases(t *testing.T) {
	released := map[string]bool{}
	s := &Service{}
	next := 0
	s.acquireHDMILease = func() func() {
		next++
		id := []string{"a", "b", "c"}[next-1]
		return func() { released[id] = true }
	}

	s.updateTaskCaptureLease("downstream", "sess", []byte(`{"id":"req-1","type":"message.send","payload":{}}`))
	s.updateTaskCaptureLease("downstream", "sess", []byte(`{"id":"req-2","type":"message.send","payload":{}}`))
	s.updateTaskCaptureLease("downstream", "sess", []byte(`{"id":"req-3","type":"message.send","payload":{}}`))

	s.updateTaskCaptureLease("upstream", "sess", []byte(`{"type":"turn.done","payload":{"request_id":"req-1","request_ids":["req-1","req-2"],"status":"ok"}}`))
	if !released["a"] || !released["b"] {
		t.Fatalf("turn.done did not release the leases of its requests: %v", released)
	}
	if released["c"] {
		t.Fatalf("turn.done released a lease of another request: %v", released)
	}

	s.updateTaskCaptureLease("upstream", "sess", []byte(`{"type":"turn.done","payload":{"status":"error"}}`))
	if !released["c"] {
		t.Fatalf("turn.done without request ids did not release the session's leases: %v", released)
	}
}
