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

// PicoClaw sends its messages without a top-level id. The lease used to go
// at the first of them, a reasoning or tool-call message early in the turn.
func TestTaskCaptureLeaseLastsUntilTurnDone(t *testing.T) {
	released := false
	s := &Service{}
	s.acquireHDMILease = func() func() { return func() { released = true } }

	s.updateTaskCaptureLease("downstream", "sess", []byte(`{"id":"req-1","type":"message.send","payload":{}}`))
	for _, message := range []string{
		`{"type":"typing.start"}`,
		`{"type":"message.create","payload":{"message_id":"m1","kind":"thought","thought":true,"content":"x"}}`,
		`{"type":"message.create","payload":{"message_id":"m2","kind":"tool_calls","content":""}}`,
		`{"type":"message.create","payload":{"message_id":"m3","placeholder":true,"content":"Thinking"}}`,
		`{"type":"typing.stop"}`,
	} {
		s.updateTaskCaptureLease("upstream", "sess", []byte(message))
		if released {
			t.Fatalf("lease released before turn.done, at %s", message)
		}
	}
	// A turn.done earlier in the session marks a build that sends it.
	s.markSessionSendsTurnDone("sess")
	s.updateTaskCaptureLease("upstream", "sess", []byte(`{"type":"message.update","payload":{"message_id":"m3","content":"Done"}}`))
	if released {
		t.Fatal("a reply released the lease although the agent sends turn.done")
	}
	s.updateTaskCaptureLease("upstream", "sess", []byte(`{"type":"turn.done","payload":{"request_ids":["req-1"],"status":"ok"}}`))
	if !released {
		t.Fatal("turn.done did not release the lease")
	}
}

func TestReplyEndsTheLeaseForAnAgentWithoutTurnDone(t *testing.T) {
	released := false
	s := &Service{}
	s.acquireHDMILease = func() func() { return func() { released = true } }

	s.updateTaskCaptureLease("downstream", "sess", []byte(`{"id":"req-1","type":"message.send","payload":{}}`))
	s.updateTaskCaptureLease("upstream", "sess", []byte(`{"type":"message.create","payload":{"message_id":"m1","content":"Done"}}`))
	if !released {
		t.Fatal("a reply did not release the lease of an agent that never sent turn.done")
	}
}

func TestErrorReleasesTheLeaseItNames(t *testing.T) {
	released := map[string]bool{}
	s := &Service{}
	next := 0
	s.acquireHDMILease = func() func() {
		next++
		id := []string{"a", "b"}[next-1]
		return func() { released[id] = true }
	}

	s.updateTaskCaptureLease("downstream", "sess", []byte(`{"id":"req-1","type":"message.send","payload":{}}`))
	s.updateTaskCaptureLease("downstream", "sess", []byte(`{"id":"req-2","type":"message.send","payload":{}}`))
	s.updateTaskCaptureLease("upstream", "sess", []byte(`{"type":"error","payload":{"code":"command_disabled","message":"no","request_id":"req-1"}}`))
	if !released["a"] || released["b"] {
		t.Fatalf("a request-scoped error released %v, want only a", released)
	}
	s.updateTaskCaptureLease("upstream", "sess", []byte(`{"type":"error","payload":{"code":"x","message":"y"}}`))
	if !released["b"] {
		t.Fatal("an error without a request id did not release the session's leases")
	}
}
