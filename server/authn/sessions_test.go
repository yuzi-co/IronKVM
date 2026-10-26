package authn

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestSessions(t *testing.T) (*SessionStore, string, *time.Time) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "login_sessions.json")
	store := NewSessionStore(path)
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	store.SetClock(func() time.Time { return now })
	return store, path, &now
}

var testClient = SessionClient{
	IP:        "10.0.30.5",
	UserAgent: "Mozilla/5.0 (X11; Linux x86_64; rv:131.0) Gecko/20100101 Firefox/131.0",
}

func TestLoginSessionTouchesUntilRevoked(t *testing.T) {
	store, _, now := newTestSessions(t)
	secret, created, err := store.Create("alice", 7, time.Hour, testClient)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != SessionID(secret) || created.ID == secret {
		t.Fatalf("id %q is not the digest of the secret", created.ID)
	}
	if created.Client != "Firefox on Linux" || created.IP != "10.0.30.5" {
		t.Fatalf("client recorded as %q from %q", created.Client, created.IP)
	}

	*now = now.Add(time.Minute)
	got, ok := store.Touch(secret, "alice", 7, "10.0.30.6")
	if !ok || got.ID != created.ID {
		t.Fatalf("touch: %+v %v", got, ok)
	}
	if !got.LastUsedAt.Equal(*now) || got.IP != "10.0.30.6" {
		t.Fatalf("touch did not record the use: %+v", got)
	}

	if _, ok = store.Touch(secret, "bob", 7, ""); ok {
		t.Fatal("a session answered for another user")
	}
	if _, ok = store.Touch(secret, "alice", 8, ""); ok {
		t.Fatal("a session answered for another token version")
	}
	if _, ok = store.Touch(created.ID, "alice", 7, ""); ok {
		t.Fatal("the public id worked as the secret")
	}

	if _, err = store.Revoke(created.ID); err != nil {
		t.Fatal(err)
	}
	if _, ok = store.Touch(secret, "alice", 7, ""); ok {
		t.Fatal("a revoked session still answers")
	}
	if store.Valid(created.ID) {
		t.Fatal("a revoked session is still valid")
	}
	if _, err = store.Revoke(created.ID); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("second revoke: %v", err)
	}
}

func TestLoginSessionExpires(t *testing.T) {
	store, _, now := newTestSessions(t)
	secret, created, err := store.Create("alice", 1, time.Hour, testClient)
	if err != nil {
		t.Fatal(err)
	}
	*now = now.Add(time.Hour + time.Second)
	if _, ok := store.Touch(secret, "alice", 1, ""); ok {
		t.Fatal("an expired session answered")
	}
	if store.Valid(created.ID) {
		t.Fatal("an expired session is valid")
	}
	if listed := store.List(""); len(listed) != 0 {
		t.Fatalf("an expired session is listed: %+v", listed)
	}
}

func TestLoginSessionsSurviveARestartWithoutTheSecret(t *testing.T) {
	store, path, now := newTestSessions(t)
	secret, created, err := store.Create("alice", 3, time.Hour, testClient)
	if err != nil {
		t.Fatal(err)
	}
	revokedSecret, revoked, err := store.Create("alice", 3, time.Hour, testClient)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.Revoke(revoked.ID); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), secret) || strings.Contains(string(data), revokedSecret) {
		t.Fatal("the session file holds a secret in the clear")
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("session file mode: %v %v", info, err)
	}

	reopened := NewSessionStore(path)
	reopened.SetClock(func() time.Time { return *now })
	if _, ok := reopened.Touch(secret, "alice", 3, ""); !ok {
		t.Fatal("a session did not survive the restart")
	}
	if _, ok := reopened.Touch(revokedSecret, "alice", 3, ""); ok {
		t.Fatal("a revoked session came back after the restart")
	}
	if got, ok := reopened.Get(created.ID); !ok || got.Client != created.Client {
		t.Fatalf("reopened session: %+v %v", got, ok)
	}
}

func TestLoginSessionsAreBoundedPerUser(t *testing.T) {
	store, _, now := newTestSessions(t)
	var first string
	for i := 0; i < MaxSessionsPerUser+5; i++ {
		*now = now.Add(time.Second)
		secret, _, err := store.Create("alice", 1, time.Hour, testClient)
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first = secret
		}
	}
	if _, _, err := store.Create("bob", 1, time.Hour, testClient); err != nil {
		t.Fatal(err)
	}
	if got := len(store.List("alice")); got != MaxSessionsPerUser {
		t.Fatalf("alice has %d sessions", got)
	}
	if _, ok := store.Touch(first, "alice", 1, ""); ok {
		t.Fatal("the least recently used session was kept past the bound")
	}
	if got := len(store.List("bob")); got != 1 {
		t.Fatalf("another user's sessions were evicted: %d", got)
	}
}

func TestLoginSessionsAreBoundedOverall(t *testing.T) {
	store, _, now := newTestSessions(t)
	for i := 0; i < MaxSessions+3; i++ {
		*now = now.Add(time.Second)
		username := "user" + string(rune('a'+i%26)) + string(rune('a'+i/26))
		if _, _, err := store.Create(username, 1, time.Hour, testClient); err != nil {
			t.Fatal(err)
		}
	}
	if got := len(store.List("")); got != MaxSessions {
		t.Fatalf("%d sessions kept", got)
	}
}

func TestRevokeUserKeepsTheNamedSession(t *testing.T) {
	store, _, _ := newTestSessions(t)
	_, keep, _ := store.Create("alice", 1, time.Hour, testClient)
	otherSecret, _, _ := store.Create("alice", 1, time.Hour, testClient)
	bobSecret, _, _ := store.Create("bob", 1, time.Hour, testClient)

	revoked, err := store.RevokeUser("alice", keep.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(revoked) != 1 {
		t.Fatalf("revoked %d sessions", len(revoked))
	}
	if !store.Valid(keep.ID) {
		t.Fatal("the kept session was revoked")
	}
	if _, ok := store.Touch(otherSecret, "alice", 1, ""); ok {
		t.Fatal("the other session survived")
	}
	if _, ok := store.Touch(bobSecret, "bob", 1, ""); !ok {
		t.Fatal("another user's session was revoked")
	}

	if _, err = store.RevokeUser("alice", ""); err != nil {
		t.Fatal(err)
	}
	if store.Valid(keep.ID) {
		t.Fatal("revoking every session kept one")
	}
}

func TestPruneDropsSessionsTheCallerRejects(t *testing.T) {
	store, _, _ := newTestSessions(t)
	_, stale, _ := store.Create("alice", 1, time.Hour, testClient)
	_, live, _ := store.Create("alice", 2, time.Hour, testClient)
	if err := store.Prune(func(s LoginSession) bool { return s.TokenVersion == 2 }); err != nil {
		t.Fatal(err)
	}
	if store.Valid(stale.ID) || !store.Valid(live.ID) {
		t.Fatal("prune kept the wrong session")
	}
}

func TestTouchWritesTheFileAtMostEveryInterval(t *testing.T) {
	store, path, now := newTestSessions(t)
	secret, _, err := store.Create("alice", 1, 24*time.Hour, testClient)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)

	*now = now.Add(time.Minute)
	store.Touch(secret, "alice", 1, "")
	if after, _ := os.ReadFile(path); string(after) != string(before) {
		t.Fatal("a touch a minute later wrote the file")
	}

	*now = now.Add(sessionTouchPersist)
	store.Touch(secret, "alice", 1, "")
	if after, _ := os.ReadFile(path); string(after) == string(before) {
		t.Fatal("a touch past the interval did not write the file")
	}
}

func TestAnUnreadableSessionFileGrantsNothing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "login_sessions.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := NewSessionStore(path)
	if listed := store.List(""); len(listed) != 0 {
		t.Fatalf("listed %+v", listed)
	}
	if _, _, err := store.Create("alice", 1, time.Hour, testClient); err != nil {
		t.Fatalf("a broken file blocks new logins: %v", err)
	}
}

func TestSummarizeUserAgent(t *testing.T) {
	cases := map[string]string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36":                         "Chrome on Windows",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36 Edg/130.0":               "Edge on Windows",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_5) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Safari/605.1.15":                      "Safari on macOS",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1": "Safari on iOS",
		"Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Mobile Safari/537.36":                   "Chrome on Android",
		"Mozilla/5.0 (X11; Linux x86_64; rv:131.0) Gecko/20100101 Firefox/131.0":                                                                  "Firefox on Linux",
		"curl/8.5.0": "curl/8.5.0",
		"":           "",
	}
	for ua, want := range cases {
		if got := SummarizeUserAgent(ua); got != want {
			t.Errorf("SummarizeUserAgent(%q) = %q, want %q", ua, got, want)
		}
	}
	long := strings.Repeat("x", 500) + "\x00\n"
	if got := SummarizeUserAgent(long); len(got) > maxSessionClient || strings.ContainsAny(got, "\x00\n") {
		t.Fatalf("an unknown agent is not cut and cleaned: %q", got)
	}
}
