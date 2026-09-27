package vpn

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestVersionCacheKeepsAnAnswerForItsTTL(t *testing.T) {
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
	c := &VersionCache{TTL: UpdateTTL, Now: func() time.Time { return now }}
	calls := 0
	fetch := func() (string, error) { calls++; return "0.79.0", nil }

	for i := 0; i < 3; i++ {
		if v, err := c.Latest(fetch); err != nil || v != "0.79.0" {
			t.Fatalf("got %q %v", v, err)
		}
	}
	if calls != 1 {
		t.Fatalf("fetched %d times within the hour", calls)
	}

	now = now.Add(UpdateTTL + time.Second)
	if _, err := c.Latest(fetch); err != nil || calls != 2 {
		t.Fatalf("an hour later it must ask again: calls=%d err=%v", calls, err)
	}

	c.Reset()
	if _, err := c.Latest(fetch); err != nil || calls != 3 {
		t.Fatalf("after Reset it must ask again: calls=%d err=%v", calls, err)
	}
}

// A failed check is kept for five minutes, so a page open on a board without
// network does not run apk or reach the release server on every visit.
func TestVersionCacheKeepsAFailureForFiveMinutes(t *testing.T) {
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
	c := &VersionCache{TTL: UpdateTTL, Now: func() time.Time { return now }}
	calls := 0
	fail := func() (string, error) { calls++; return "", errors.New("offline") }
	for i := 0; i < 2; i++ {
		if _, err := c.Latest(fail); err == nil || err.Error() != "offline" {
			t.Fatalf("the failure must reach the caller: %v", err)
		}
	}
	if calls != 1 {
		t.Fatalf("asked %d times within five minutes of a failure", calls)
	}
	now = now.Add(UpdateFailTTL + time.Second)
	if v, err := c.Latest(func() (string, error) { return "1.90.1", nil }); err != nil || v != "1.90.1" {
		t.Fatalf("after five minutes it must ask again: %q %v", v, err)
	}
}

func TestVersionCacheAsksOnceForConcurrentCallers(t *testing.T) {
	c := &VersionCache{TTL: UpdateTTL}
	var calls atomic.Int32
	release := make(chan struct{})
	fetch := func() (string, error) { calls.Add(1); <-release; return "0.79.0", nil }

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if v, err := c.Latest(fetch); err != nil || v != "0.79.0" {
				t.Errorf("got %q %v", v, err)
			}
		}()
	}
	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()
	if n := calls.Load(); n != 1 {
		t.Fatalf("fetched %d times for callers that came together", n)
	}
}

// The lock guards the cached value only. A slow check must not hold up an
// update, which resets the cache.
func TestVersionCacheDoesNotHoldItsLockWhileFetching(t *testing.T) {
	c := &VersionCache{TTL: UpdateTTL}
	release := make(chan struct{})
	started := make(chan struct{})
	go func() {
		_, _ = c.Latest(func() (string, error) { close(started); <-release; return "1", nil })
	}()
	<-started
	done := make(chan struct{})
	go func() { c.Reset(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Reset waited for the fetch")
	}
	close(release)
}
