package agent

import "testing"

func TestReleaseOwnedRejectsStaleSession(t *testing.T) {
	lock := &SessionLock{}
	if err := lock.Ensure("current"); err != nil {
		t.Fatal(err)
	}

	if lock.ReleaseOwned("stale") {
		t.Fatal("stale session released the active lock")
	}
	if got := lock.Owner(); got != "current" {
		t.Fatalf("owner = %q, want current", got)
	}

	if !lock.ReleaseOwned("current") {
		t.Fatal("current owner failed to release the lock")
	}
	if got := lock.Owner(); got != "" {
		t.Fatalf("owner = %q, want empty", got)
	}
	if lock.ReleaseOwned("current") {
		t.Fatal("empty lock was treated as an owned release")
	}
}
