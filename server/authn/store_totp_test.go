package authn

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func newTOTPStore(t *testing.T) *Store {
	t.Helper()
	store := NewStore(filepath.Join(t.TempDir(), "pwd"))
	if err := store.Create("alice", "alice-password", RoleUser); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestTOTPSetAndClear(t *testing.T) {
	store := newTOTPStore(t)
	before, err := store.Get("alice")
	if err != nil {
		t.Fatal(err)
	}
	if before.TOTPEnabled() {
		t.Fatal("a new account has TOTP on")
	}

	user, err := store.SetTOTP("alice", "v1:sealed", []string{"h1", "h2"})
	if err != nil {
		t.Fatal(err)
	}
	if !user.TOTPEnabled() || len(user.RecoveryCodes) != 2 {
		t.Fatalf("unexpected user after SetTOTP: %+v", user)
	}
	// Enrolment keeps the session that made it.
	if user.TokenVersion != before.TokenVersion {
		t.Fatal("SetTOTP revoked sessions")
	}

	users, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	for _, info := range users {
		if info.Username == "alice" && !info.TOTPEnabled {
			t.Fatal("List does not report TOTP")
		}
	}

	user, err = store.ClearTOTP("alice")
	if err != nil {
		t.Fatal(err)
	}
	if user.TOTPEnabled() || len(user.RecoveryCodes) != 0 {
		t.Fatalf("unexpected user after ClearTOTP: %+v", user)
	}
	if _, err = store.ClearTOTP("nobody"); err != ErrUserNotFound {
		t.Fatalf("ClearTOTP of a missing user: %v", err)
	}
	if _, err = store.SetTOTP("alice", "", nil); err == nil {
		t.Fatal("an empty secret was stored")
	}
}

// The secret stays sealed on disk, and the file keeps its mode.
func TestTOTPFileContents(t *testing.T) {
	store := newTOTPStore(t)
	if _, err := store.SetTOTP("alice", "v1:sealed-value", []string{"abc"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(store.path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"totpSecret": "v1:sealed-value"`) {
		t.Fatalf("secret not stored as given:\n%s", data)
	}
	info, err := os.Stat(store.path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("account mode = %o, want 600", info.Mode().Perm())
	}
}

func TestConsumeRecoveryCodeOnce(t *testing.T) {
	store := newTOTPStore(t)
	if _, err := store.SetTOTP("alice", "v1:sealed", []string{"h1", "h2"}); err != nil {
		t.Fatal(err)
	}

	ok, err := store.ConsumeRecoveryCode("alice", "h1")
	if err != nil || !ok {
		t.Fatalf("first use: ok=%v err=%v", ok, err)
	}
	ok, err = store.ConsumeRecoveryCode("alice", "h1")
	if err != nil || ok {
		t.Fatalf("second use: ok=%v err=%v", ok, err)
	}
	ok, err = store.ConsumeRecoveryCode("alice", "nope")
	if err != nil || ok {
		t.Fatalf("unknown code: ok=%v err=%v", ok, err)
	}
	user, err := store.Get("alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(user.RecoveryCodes) != 1 || user.RecoveryCodes[0] != "h2" {
		t.Fatalf("unexpected codes left: %v", user.RecoveryCodes)
	}
}

// Two logins that present the same recovery code at once: one wins.
func TestConsumeRecoveryCodeConcurrent(t *testing.T) {
	store := newTOTPStore(t)
	if _, err := store.SetTOTP("alice", "v1:sealed", []string{"h1"}); err != nil {
		t.Fatal(err)
	}
	var wins atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if ok, err := store.ConsumeRecoveryCode("alice", "h1"); err == nil && ok {
				wins.Add(1)
			}
		}()
	}
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatalf("%d requests spent the same code", wins.Load())
	}
}

func TestSetRecoveryCodesNeedsTOTP(t *testing.T) {
	store := newTOTPStore(t)
	user, err := store.SetRecoveryCodes("alice", []string{"h1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(user.RecoveryCodes) != 0 {
		t.Fatal("recovery codes stored for an account without TOTP")
	}
	if _, err = store.SetTOTP("alice", "v1:sealed", []string{"h1"}); err != nil {
		t.Fatal(err)
	}
	user, err = store.SetRecoveryCodes("alice", []string{"h2", "h3"})
	if err != nil {
		t.Fatal(err)
	}
	if len(user.RecoveryCodes) != 2 || user.RecoveryCodes[0] != "h2" {
		t.Fatalf("unexpected codes: %v", user.RecoveryCodes)
	}
}

// A copy handed out by Get does not share its recovery codes with the cache.
func TestRecoveryCodesAreCopied(t *testing.T) {
	store := newTOTPStore(t)
	if _, err := store.SetTOTP("alice", "v1:sealed", []string{"h1", "h2"}); err != nil {
		t.Fatal(err)
	}
	first, err := store.Get("alice")
	if err != nil {
		t.Fatal(err)
	}
	first.RecoveryCodes[0] = "changed"
	second, err := store.Get("alice")
	if err != nil {
		t.Fatal(err)
	}
	if second.RecoveryCodes[0] != "h1" {
		t.Fatal("a caller's change reached the store")
	}
}

func TestKeyringRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "totp.key")
	keyring := NewKeyring(path)
	if _, err := keyring.Open("alice", "v1:AAAA"); err != ErrNoKey {
		t.Fatalf("open before any key: %v", err)
	}
	sealed, err := keyring.Seal("alice", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(sealed, "secret") {
		t.Fatal("sealed value carries the secret")
	}
	opened, err := NewKeyring(path).Open("alice", sealed)
	if err != nil || string(opened) != "secret" {
		t.Fatalf("open: %q %v", opened, err)
	}
	if _, err = keyring.Open("bob", sealed); err == nil {
		t.Fatal("a value sealed for alice opened for bob")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("key mode = %o, want 600", info.Mode().Perm())
	}
}
