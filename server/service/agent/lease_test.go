package agent

import (
	"testing"
	"time"
)

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

func TestAnAgentThatSaysItSendsTurnDoneKeepsTheFirstTurnsLease(t *testing.T) {
	leases, released := newCountingLeases()
	leases.ExpectTurnDone("s")
	leases.Prompt("s", "r1", 0)

	leases.Observe("s", Event{Type: EventAgentMessage, Kind: KindReply})
	if released["a"] {
		t.Fatal("a reply ended the first turn's lease of an agent that said it sends turn_done")
	}
	leases.Observe("s", Event{Type: EventTurnDone, RequestIDs: []string{"r1"}})
	if !released["a"] {
		t.Fatal("turn_done did not end the lease")
	}

	// Forget drops what the agent said with the session, so a later
	// session of the same id starts from the reply rule again.
	leases.Forget("s")
	leases.Prompt("s", "r2", 0)
	leases.Observe("s", Event{Type: EventAgentMessage, Kind: KindReply})
	if !released["b"] {
		t.Fatal("a reply did not end the lease after the session was forgotten")
	}
}

// A session is in a turn while it holds a lease, whether or not a capture
// lease could be taken, and its timer ends the turn as it ends the lease.
func TestInTurnFollowsTheLeases(t *testing.T) {
	leases := NewTaskLeases(nil)
	if leases.InTurn("s") {
		t.Fatal("in a turn before any prompt")
	}
	leases.Prompt("s", "r1", 0)
	if !leases.InTurn("s") || leases.InTurn("other") {
		t.Fatal("the prompt did not start a turn for its session alone")
	}
	leases.Observe("s", Event{Type: EventTurnDone, RequestIDs: []string{"r1"}})
	if leases.InTurn("s") {
		t.Fatal("turn_done did not end the turn")
	}

	leases.Prompt("s", "r2", 20)
	deadline := time.Now().Add(5 * time.Second)
	for leases.InTurn("s") {
		if time.Now().After(deadline) {
			t.Fatal("the prompt's time limit did not end the turn")
		}
		time.Sleep(5 * time.Millisecond)
	}

	var none *TaskLeases
	if none.InTurn("s") {
		t.Fatal("no tracker means no turn")
	}
}

// A turn keeps its number while prompts join it and gets a new one once it
// has ended, so something tied to one turn can tell the next apart.
func TestTurnNumbersEachTurnOfASession(t *testing.T) {
	leases, _ := newCountingLeases()
	if got := leases.Turn("s"); got != 0 {
		t.Fatalf("turn before any prompt = %d", got)
	}
	leases.Prompt("s", "r1", 0)
	first := leases.Turn("s")
	leases.Prompt("s", "r2", 0)
	if first == 0 || leases.Turn("s") != first {
		t.Fatalf("turn = %d, then %d with a second prompt, want the same nonzero", first, leases.Turn("s"))
	}
	leases.Prompt("other", "r1", 0)
	if leases.Turn("other") == first {
		t.Fatal("another session shares the turn's number")
	}

	leases.Cancel("s", "")
	if got := leases.Turn("s"); got != 0 {
		t.Fatalf("turn after cancel = %d", got)
	}
	leases.Prompt("s", "r3", 0)
	if second := leases.Turn("s"); second == 0 || second == first {
		t.Fatalf("next turn = %d, want a new number after %d", second, first)
	}
}
