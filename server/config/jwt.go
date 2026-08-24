package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// secretKeyReader is the entropy source. nil means crypto/rand; tests replace
// it to exercise the failure path.
var secretKeyReader io.Reader

func generateSecretKey(reader io.Reader) (string, error) {
	b := make([]byte, 64)

	var err error
	if reader == nil {
		_, err = rand.Read(b)
	} else {
		_, err = io.ReadFull(reader, b)
	}

	if err != nil {
		// There is no safe fallback. Anything derived from the clock is a key
		// an attacker can search, because they know roughly when the device
		// booted, so this has to fail instead.
		return "", fmt.Errorf("failed to read random bytes for the secret key: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
