package authn

import (
	"path/filepath"
	"testing"
	"time"
)

// The password check is slow by design. It must not hold the store's lock,
// or every login stalls the web UI's session checks for its whole length.
func TestAuthenticateChecksThePasswordOutsideTheLock(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "pwd"))
	if _, ok, err := store.Authenticate("admin", "admin"); err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}

	inCompare := make(chan struct{})
	release := make(chan struct{})
	original := compareHash
	compareHash = func(hash, password []byte) error {
		close(inCompare)
		<-release
		return original(hash, password)
	}
	t.Cleanup(func() { compareHash = original })

	done := make(chan bool)
	go func() {
		_, ok, _ := store.Authenticate("admin", "admin")
		done <- ok
	}()
	<-inCompare

	got := make(chan error)
	go func() {
		_, err := store.Get("admin")
		got <- err
	}()
	select {
	case err := <-got:
		if err != nil {
			t.Fatalf("Get: %s", err)
		}
	case <-time.After(time.Second):
		close(release)
		t.Fatal("Get waited for the password check of another login")
	}

	close(release)
	if !<-done {
		t.Fatal("the login failed")
	}
}
