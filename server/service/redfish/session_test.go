package redfish

import (
	"regexp"
	"testing"
	"time"
)

// testClock is a clock a test moves by hand.
type testClock struct{ now time.Time }

func (c *testClock) Now() time.Time { return c.now }

func newTestStore() (*sessionStore, *testClock) {
	clock := &testClock{now: time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)}
	return newSessionStore(clock.Now), clock
}

func TestSessionCreateReturnsARandomTokenThatFindsIt(t *testing.T) {
	store, _ := newTestStore()

	sess, token, err := store.create("admin", 7)
	if err != nil {
		t.Fatalf("create: %s", err)
	}
	if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(token) {
		t.Fatalf("token %q is not 32 random bytes in hex", token)
	}
	if sess.id == "" || sess.username != "admin" || sess.tokenVersion != 7 {
		t.Fatalf("session is %+v", sess)
	}

	found, ok := store.byToken(token)
	if !ok || found.id != sess.id {
		t.Fatalf("byToken = %+v, %v, want the session", found, ok)
	}
	if _, ok := store.byToken(token[:63] + "x"); ok {
		t.Fatal("a different token found the session")
	}

	_, second, _ := store.create("admin", 7)
	if second == token {
		t.Fatal("two sessions got the same token")
	}
}

func TestSessionEndsAfterThirtyIdleMinutes(t *testing.T) {
	store, clock := newTestStore()
	_, token, _ := store.create("admin", 1)

	clock.now = clock.now.Add(29 * time.Minute)
	if _, ok := store.byToken(token); !ok {
		t.Fatal("session ended before 30 idle minutes")
	}

	// Use restarts the idle timer.
	clock.now = clock.now.Add(29 * time.Minute)
	if _, ok := store.byToken(token); !ok {
		t.Fatal("session ended although it was used 29 minutes ago")
	}

	clock.now = clock.now.Add(30 * time.Minute)
	if _, ok := store.byToken(token); ok {
		t.Fatal("session outlived 30 idle minutes")
	}
	if len(store.list()) != 0 {
		t.Fatal("an expired session is still listed")
	}
}

func TestSessionStoreKeepsAtMostSixteenAndDropsTheOldest(t *testing.T) {
	store, clock := newTestStore()

	var tokens []string
	for i := 0; i < maxSessions+1; i++ {
		_, token, err := store.create("admin", 1)
		if err != nil {
			t.Fatal(err)
		}
		tokens = append(tokens, token)
		clock.now = clock.now.Add(time.Second)
	}

	if got := len(store.list()); got != maxSessions {
		t.Fatalf("%d sessions, want %d", got, maxSessions)
	}
	if _, ok := store.byToken(tokens[0]); ok {
		t.Fatal("the oldest session survived the limit")
	}
	for _, token := range tokens[1:] {
		if _, ok := store.byToken(token); !ok {
			t.Fatal("a newer session was dropped")
		}
	}
}

func TestSessionRemoveEndsIt(t *testing.T) {
	store, _ := newTestStore()
	sess, token, _ := store.create("admin", 1)

	if !store.remove(sess.id) {
		t.Fatal("remove reported no session")
	}
	if _, ok := store.byToken(token); ok {
		t.Fatal("a removed session still answers")
	}
	if _, ok := store.get(sess.id); ok {
		t.Fatal("a removed session is still found by id")
	}
	if store.remove(sess.id) {
		t.Fatal("a second remove reported a session")
	}
}
