package config

import (
	"errors"
	"strings"
	"testing"
)

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("no entropy")
}

// Falling back to anything derived from the clock would hand an attacker a
// signing key they can search: the device boots, and the key is the nanosecond
// it happened to boot at.
func TestGenerateSecretKeyFailsRatherThanReturningAPredictableValue(t *testing.T) {
	key, err := generateSecretKey(failingReader{})

	if err == nil {
		t.Fatalf("expected an error, got key %q", key)
	}
	if key != "" {
		t.Fatalf("expected no key on failure, got %q", key)
	}
}

func TestGenerateSecretKeyReturnsADifferentKeyEachTime(t *testing.T) {
	first, err := generateSecretKey(nil)
	if err != nil {
		t.Fatalf("failed to generate key: %s", err)
	}

	second, err := generateSecretKey(nil)
	if err != nil {
		t.Fatalf("failed to generate key: %s", err)
	}

	if first == second {
		t.Fatal("two generated keys are identical")
	}
	if strings.TrimSpace(first) == "" {
		t.Fatal("generated key is blank")
	}
}
