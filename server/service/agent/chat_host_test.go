package agent

import "testing"

func TestStaleSessionDoesNotReleaseHID(t *testing.T) {
	lock := &SessionLock{}
	if err := lock.Ensure("current"); err != nil {
		t.Fatal(err)
	}

	releaseCalls := 0
	released, err := releaseOwnedSession(lock, "stale", func() error {
		releaseCalls++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if released || releaseCalls != 0 {
		t.Fatalf("released=%v HID releases=%d, want false and 0", released, releaseCalls)
	}
	if got := lock.Owner(); got != "current" {
		t.Fatalf("owner = %q, want current", got)
	}

	released, err = releaseOwnedSession(lock, "current", func() error {
		releaseCalls++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !released || releaseCalls != 1 {
		t.Fatalf("released=%v HID releases=%d, want true and 1", released, releaseCalls)
	}
}
