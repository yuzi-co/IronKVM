package authn

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOIDCLinkIsMadeOnceAndChecked(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "pwd"))
	if err := store.Create("alice", "password-123", RoleUser); err != nil {
		t.Fatal(err)
	}
	before, _ := store.Get("alice")

	user, previous, err := store.BindOIDC("alice", "https://idp.example", "sub-1")
	if err != nil || previous != "" || user.OIDCSubject != "sub-1" || user.OIDCIssuer != "https://idp.example" {
		t.Fatalf("first link: %+v, %q, %v", user, previous, err)
	}
	// Linking is not a credential change: sessions stay.
	if user.TokenVersion != before.TokenVersion {
		t.Fatal("linking revoked the sessions")
	}
	if _, previous, err = store.BindOIDC("alice", "https://idp.example", "sub-1"); err != nil || previous != "https://idp.example" {
		t.Fatalf("same identity: %q, %v", previous, err)
	}
	if _, _, err = store.BindOIDC("alice", "https://idp.example", "sub-2"); !errors.Is(err, ErrOIDCSubjectMismatch) {
		t.Fatalf("another sub: %v", err)
	}
	if got, _ := store.Get("alice"); got.OIDCSubject != "sub-1" {
		t.Fatalf("a refused link changed the account: %q", got.OIDCSubject)
	}

	// Another issuer replaces the link.
	user, previous, err = store.BindOIDC("alice", "https://new-idp.example", "sub-9")
	if err != nil || previous != "https://idp.example" || user.OIDCIssuer != "https://new-idp.example" || user.OIDCSubject != "sub-9" {
		t.Fatalf("new issuer: %+v, %q, %v", user, previous, err)
	}

	users, _ := store.List()
	for _, u := range users {
		if u.OIDCLinked != (u.Username == "alice") {
			t.Fatalf("list says %s linked=%v", u.Username, u.OIDCLinked)
		}
	}

	if user, err = store.ClearOIDC("alice"); err != nil || user.OIDCIssuer != "" || user.OIDCSubject != "" {
		t.Fatalf("clear: %+v, %v", user, err)
	}
	if _, err = store.ClearOIDC("nobody"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("clear an unknown account: %v", err)
	}
	if _, _, err = store.BindOIDC("alice", "", "sub-1"); err == nil {
		t.Fatal("linked without an issuer")
	}
	if _, _, err = store.BindOIDC("nobody", "https://idp.example", "sub-1"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("link an unknown account: %v", err)
	}
}

func TestOIDCCreatedAccountIsLinkedAndKeptOnDisk(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pwd")
	store := NewStore(path)
	if err := store.CreateOIDC("viewer1", "password-123", RoleViewer, "https://idp.example", "sub-v"); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateOIDC("x", "password-123", RoleUser, "https://idp.example", ""); err == nil {
		t.Fatal("created an OIDC account without a sub")
	}

	// A viewer is rewritten on its way to disk; the link goes with it.
	data, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(data), `"oidcSubject": "sub-v"`) {
		t.Fatalf("file %s, %v", data, err)
	}
	user, err := NewStore(path).Get("viewer1")
	if err != nil || user.Role != RoleViewer || user.OIDCIssuer != "https://idp.example" || user.OIDCSubject != "sub-v" || !user.Enabled {
		t.Fatalf("read back %+v, %v", user, err)
	}
}

// An account file written before the link existed has no such fields. It
// loads, its accounts are unlinked, and a link is added to it in place.
func TestAccountFileWithoutOIDCLinksStillLoads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pwd")
	store := NewStore(path)
	if err := store.Create("alice", "password-123", RoleUser); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || strings.Contains(string(data), "oidc") {
		t.Fatalf("an unlinked account wrote OIDC fields: %s, %v", data, err)
	}
	if _, ok, err := NewStore(path).Authenticate("alice", "password-123"); err != nil || !ok {
		t.Fatalf("login: %v %v", ok, err)
	}
	if _, _, err = NewStore(path).BindOIDC("alice", "https://idp.example", "sub-1"); err != nil {
		t.Fatal(err)
	}
	if user, _ := NewStore(path).Get("alice"); user.OIDCSubject != "sub-1" {
		t.Fatalf("link not kept: %+v", user)
	}
}
