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

	// A single sign-on does not link an account that has no link yet: an
	// administrator or the account's password does (ironkvm-dist#169).
	if _, err := store.BindOIDC("alice", "https://idp.example", "sub-1"); !errors.Is(err, ErrOIDCNotLinked) {
		t.Fatalf("an unlinked account was bound: %v", err)
	}
	if got, _ := store.Get("alice"); got.OIDCLinked() {
		t.Fatal("a refused bind linked the account")
	}

	user, err := store.ConfirmOIDC("alice", "https://idp.example", "sub-1")
	if err != nil || user.OIDCSubject != "sub-1" || user.OIDCIssuer != "https://idp.example" {
		t.Fatalf("first link: %+v, %v", user, err)
	}
	// Linking is not a credential change: sessions stay.
	if user.TokenVersion != before.TokenVersion {
		t.Fatal("linking revoked the sessions")
	}
	if _, err = store.BindOIDC("alice", "https://idp.example", "sub-1"); err != nil {
		t.Fatalf("same identity: %v", err)
	}
	if _, err = store.BindOIDC("alice", "https://idp.example", "sub-2"); !errors.Is(err, ErrOIDCSubjectMismatch) {
		t.Fatalf("another sub: %v", err)
	}
	if got, _ := store.Get("alice"); got.OIDCSubject != "sub-1" {
		t.Fatalf("a refused link changed the account: %q", got.OIDCSubject)
	}

	// Under another issuer the account is not linked, and stays as it is.
	if _, err = store.BindOIDC("alice", "https://new-idp.example", "sub-9"); !errors.Is(err, ErrOIDCNotLinked) {
		t.Fatalf("new issuer: %v", err)
	}
	if got, _ := store.Get("alice"); got.OIDCIssuer != "https://idp.example" || got.OIDCSubject != "sub-1" {
		t.Fatalf("a refused bind changed the link: %q %q", got.OIDCIssuer, got.OIDCSubject)
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
	if _, err = store.BindOIDC("alice", "", "sub-1"); err == nil {
		t.Fatal("linked without an issuer")
	}
	if _, err = store.BindOIDC("nobody", "https://idp.example", "sub-1"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("link an unknown account: %v", err)
	}
}

// ConfirmOIDC links only an account with no link, or the same one again, or
// one linked under another issuer: a password confirmed while an
// administrator linked the account to another sub must not undo that.
func TestOIDCConfirmLinksOnlyAnUnlinkedAccount(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "pwd"))
	if err := store.Create("alice", "password-123", RoleUser); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ConfirmOIDC("alice", "https://idp.example", "sub-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ConfirmOIDC("alice", "https://idp.example", "sub-1"); err != nil {
		t.Fatalf("the same link again: %v", err)
	}
	if _, err := store.ConfirmOIDC("alice", "https://idp.example", "sub-2"); !errors.Is(err, ErrOIDCSubjectMismatch) {
		t.Fatalf("confirm over another sub: %v", err)
	}
	if got, _ := store.Get("alice"); got.OIDCIssuer != "https://idp.example" || got.OIDCSubject != "sub-1" {
		t.Fatalf("a refused confirm changed the link: %q %q", got.OIDCIssuer, got.OIDCSubject)
	}
	if user, err := store.ConfirmOIDC("alice", "https://other.example", "sub-7"); err != nil || user.OIDCIssuer != "https://other.example" || user.OIDCSubject != "sub-7" {
		t.Fatalf("confirm under another issuer: %+v %v", user, err)
	}
	if _, err := store.ConfirmOIDC("alice", "", "sub-1"); err == nil {
		t.Fatal("confirmed without an issuer")
	}
}

// LinkOIDC is the administrator's pre-link: it sets the link whatever was
// there, and says what it replaced.
func TestOIDCAdministratorLinkReplacesTheLink(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "pwd"))
	if err := store.Create("alice", "password-123", RoleUser); err != nil {
		t.Fatal(err)
	}
	user, previousIssuer, previousSubject, err := store.LinkOIDC("alice", "https://idp.example", "sub-1")
	if err != nil || previousIssuer != "" || previousSubject != "" || user.OIDCSubject != "sub-1" {
		t.Fatalf("pre-link: %+v %q %q %v", user, previousIssuer, previousSubject, err)
	}
	_, previousIssuer, previousSubject, err = store.LinkOIDC("alice", "https://idp.example", "sub-2")
	if err != nil || previousIssuer != "https://idp.example" || previousSubject != "sub-1" {
		t.Fatalf("relink: %q %q %v", previousIssuer, previousSubject, err)
	}
	if _, err = store.BindOIDC("alice", "https://idp.example", "sub-2"); err != nil {
		t.Fatalf("a sign-in as the pre-linked sub: %v", err)
	}
	if _, _, _, err = store.LinkOIDC("alice", "https://idp.example", ""); err == nil {
		t.Fatal("linked without a sub")
	}
	if _, _, _, err = store.LinkOIDC("nobody", "https://idp.example", "sub-1"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("link an unknown account: %v", err)
	}
}

// An account made for a single sign-on has a password nobody knows, so it
// cannot confirm a link with it. Setting a password ends that.
func TestOIDCCreatedAccountHasNoPasswordUntilOneIsSet(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "pwd"))
	if err := store.CreateOIDC("dave", "random-password-1", RoleUser, "https://idp.example", "sub-d"); err != nil {
		t.Fatal(err)
	}
	if err := store.Create("alice", "password-123", RoleUser); err != nil {
		t.Fatal(err)
	}
	if dave, _ := store.Get("dave"); !dave.NoPassword {
		t.Fatal("an account made for OIDC has a usable password")
	}
	if alice, _ := store.Get("alice"); alice.NoPassword {
		t.Fatal("an account made with a password has none")
	}
	if dave, err := store.SetPassword("dave", "a-new-password"); err != nil || dave.NoPassword {
		t.Fatalf("after a password was set: %+v %v", dave, err)
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
	if _, err = NewStore(path).ConfirmOIDC("alice", "https://idp.example", "sub-1"); err != nil {
		t.Fatal(err)
	}
	if user, _ := NewStore(path).Get("alice"); user.OIDCSubject != "sub-1" {
		t.Fatalf("link not kept: %+v", user)
	}
}
