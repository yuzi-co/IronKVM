package authn

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// A device whose account file is absent builds the default database on every
// read, so deriving it must not cost a bcrypt hash each time. Reusing one hash
// is what makes that observable.
func TestDefaultPasswordHashIsComputedOnce(t *testing.T) {
	first, err := defaultDatabase()
	if err != nil {
		t.Fatal(err)
	}

	second, err := defaultDatabase()
	if err != nil {
		t.Fatal(err)
	}

	if first.Users[0].PasswordHash != second.Users[0].PasswordHash {
		t.Fatal("the default password hash is recomputed on every call")
	}
}

func TestDefaultDatabaseStillAcceptsTheDefaultPassword(t *testing.T) {
	db, err := defaultDatabase()
	if err != nil {
		t.Fatal(err)
	}

	user := db.Users[0]
	if user.Username != defaultUsername {
		t.Fatalf("default username = %q, want %q", user.Username, defaultUsername)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(defaultPassword)); err != nil {
		t.Fatalf("the default password no longer verifies: %s", err)
	}
}
