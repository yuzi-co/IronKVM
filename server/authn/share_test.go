package authn

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestShares(t *testing.T) (*ShareStore, *time.Time) {
	t.Helper()
	store := NewShareStore(filepath.Join(t.TempDir(), "share_links.json"))
	now := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	store.SetClock(func() time.Time { return now })
	return store, &now
}

func TestShareLinkRedeemsUntilItExpires(t *testing.T) {
	store, now := newTestShares(t)
	link, token, err := store.Create("for Sam", "admin", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := store.Redeem(token); !ok || got.ID != link.ID {
		t.Fatalf("redeem: %+v %v", got, ok)
	}
	if _, ok := store.Redeem(token + "x"); ok {
		t.Fatal("a wrong token redeemed")
	}
	if _, ok := store.Redeem(""); ok {
		t.Fatal("an empty token redeemed")
	}

	*now = now.Add(time.Hour + time.Second)
	if _, ok := store.Redeem(token); ok {
		t.Fatal("an expired link redeemed")
	}
	if _, ok := store.Valid(link.ID); ok {
		t.Fatal("an expired link is still valid")
	}
	if links, _ := store.List(); len(links) != 0 {
		t.Fatalf("an expired link is listed: %+v", links)
	}
}

func TestShareLinkRevokeEndsItAtOnce(t *testing.T) {
	store, _ := newTestShares(t)
	link, token, err := store.Create("", "admin", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Revoke(link.ID); err != nil {
		t.Fatal(err)
	}
	if _, ok := store.Redeem(token); ok {
		t.Fatal("a revoked link redeemed")
	}
	if _, ok := store.Valid(link.ID); ok {
		t.Fatal("a revoked link is still valid")
	}
	if err = store.Revoke(link.ID); !errors.Is(err, ErrShareNotFound) {
		t.Fatalf("second revoke: %v", err)
	}
}

func TestShareFileHoldsOnlyTheTokenHash(t *testing.T) {
	store, _ := newTestShares(t)
	_, token, err := store.Create("x", "admin", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(store.path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), token) {
		t.Fatal("the token is stored in the clear")
	}
	info, _ := os.Stat(store.path)
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", info.Mode().Perm())
	}
}

func TestShareLifetimeAndCountAreBounded(t *testing.T) {
	store, _ := newTestShares(t)
	for _, ttl := range []time.Duration{0, time.Second, MaxShareTTL + time.Second} {
		if _, _, err := store.Create("", "admin", ttl); !errors.Is(err, ErrShareTTL) {
			t.Fatalf("ttl %v: %v", ttl, err)
		}
	}
	for index := 0; index < MaxShareLinks; index++ {
		if _, _, err := store.Create("", "admin", time.Hour); err != nil {
			t.Fatalf("link %d: %v", index, err)
		}
	}
	if _, _, err := store.Create("", "admin", time.Hour); !errors.Is(err, ErrShareLimit) {
		t.Fatalf("over the limit: %v", err)
	}
}

func TestShareStoreSurvivesARestartAndAnUnreadableFileGrantsNothing(t *testing.T) {
	store, now := newTestShares(t)
	_, token, err := store.Create("", "admin", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	again := NewShareStore(store.path)
	again.SetClock(func() time.Time { return *now })
	if _, ok := again.Redeem(token); !ok {
		t.Fatal("the link did not survive a restart")
	}

	if err = os.WriteFile(store.path, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	broken := NewShareStore(store.path)
	broken.SetClock(func() time.Time { return *now })
	if _, ok := broken.Redeem(token); ok {
		t.Fatal("a link redeemed from an unreadable file")
	}
}
