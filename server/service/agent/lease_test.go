package agent

import "testing"

func newCountingLeases() (*TaskLeases, map[string]bool) {
	released := map[string]bool{}
	next := 0
	leases := NewTaskLeases(func() func() {
		next++
		id := string(rune('a' + next - 1))
		return func() { released[id] = true }
	})
	return leases, released
}

func TestTurnDoneEndsTheLeasesItNames(t *testing.T) {
	leases, released := newCountingLeases()
	leases.Prompt("s", "r1", 0)
	leases.Prompt("s", "r2", 0)
	leases.Prompt("s", "r3", 0)

	leases.Observe("s", Event{Type: EventTurnDone, RequestIDs: []string{"r1", "r2"}})
	if !released["a"] || !released["b"] || released["c"] {
		t.Fatalf("released %v, want a and b", released)
	}
	leases.Observe("s", Event{Type: EventTurnDone})
	if !released["c"] {
		t.Fatal("turn_done without request ids did not end the session's leases")
	}
}

func TestOnlyAReplyFromAnAgentWithoutTurnDoneEndsALease(t *testing.T) {
	leases, released := newCountingLeases()
	leases.Prompt("s", "r1", 0)
	for _, event := range []Event{
		{Type: EventTurnStarted},
		{Type: EventAgentMessage, Kind: KindHidden},
		{Type: EventAgentMessage, Kind: KindPlaceholder},
		{Type: EventToolCall, Title: "click"},
		{Type: EventObservation},
		{Type: EventAgentMessageRemoved},
	} {
		leases.Observe("s", event)
		if released["a"] {
			t.Fatalf("%s ended the lease", event.Type)
		}
	}
	leases.Observe("s", Event{Type: EventAgentMessage, Kind: KindReply})
	if !released["a"] {
		t.Fatal("a reply did not end the lease of an agent without turn_done")
	}

	leases.Observe("s", Event{Type: EventTurnDone})
	leases.Prompt("s", "r2", 0)
	leases.Observe("s", Event{Type: EventAgentMessage, Kind: KindReply})
	if released["b"] {
		t.Fatal("a reply ended the lease of an agent that sends turn_done")
	}
	leases.Forget("s")
	if !released["b"] {
		t.Fatal("closing the session did not end its lease")
	}
}

func TestAnErrorEndsTheLeaseItNames(t *testing.T) {
	leases, released := newCountingLeases()
	leases.Prompt("s", "r1", 0)
	leases.Prompt("s", "r2", 0)
	leases.Observe("s", Event{Type: EventError, RequestID: "r1"})
	if !released["a"] || released["b"] {
		t.Fatalf("released %v, want only a", released)
	}
	leases.Observe("s", Event{Type: EventError})
	if !released["b"] {
		t.Fatal("an error without a request id did not end the session's leases")
	}
}

func TestCancelEndsOneOrAllLeases(t *testing.T) {
	leases, released := newCountingLeases()
	leases.Prompt("s", "r1", 0)
	leases.Prompt("s", "r2", 0)
	leases.Prompt("other", "r1", 0)
	leases.Cancel("s", "r1")
	if !released["a"] || released["b"] {
		t.Fatalf("released %v, want only a", released)
	}
	leases.Cancel("s", "")
	if !released["b"] || released["c"] {
		t.Fatalf("released %v, want b and not c", released)
	}
}

func TestARepeatedPromptKeepsOneLease(t *testing.T) {
	leases, released := newCountingLeases()
	leases.Prompt("s", "r1", 0)
	leases.Prompt("s", "r1", 1000)
	leases.Cancel("s", "r1")
	if !released["a"] || len(released) != 1 {
		t.Fatalf("released %v, want one lease", released)
	}
}
