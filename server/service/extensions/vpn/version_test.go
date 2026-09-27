package vpn

import (
	"errors"
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

func TestVersionCacheDoesNotKeepAFailure(t *testing.T) {
	c := &VersionCache{TTL: UpdateTTL}
	if _, err := c.Latest(func() (string, error) { return "", errors.New("offline") }); err == nil {
		t.Fatal("the failure must reach the caller")
	}
	if v, err := c.Latest(func() (string, error) { return "1.90.1", nil }); err != nil || v != "1.90.1" {
		t.Fatalf("a failure must not be cached: %q %v", v, err)
	}
}
