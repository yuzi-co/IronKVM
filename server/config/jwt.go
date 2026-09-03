package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	// jwtSecretFile holds the signing key for boards where server.yaml names
	// none, which is every board by default.
	//
	// /etc/kvm is the right home for it. On an A/B board that directory is a
	// bind mount of /data/identity (tools/abslots/device/S02identity), so a key
	// written here survives a slot switch as well as a restart. It is a
	// dot-file for the same reason .picoclaw_internal_token is: a secret is not
	// a setting, and nothing should list it beside the files an operator edits.
	jwtSecretFile = "/etc/kvm/.jwt_secret"

	// A stored key shorter than this did not come from generateSecretKey, which
	// writes 64 random bytes as 88 base64 characters. 32 bytes is the floor for
	// an HMAC key and base64 turns them into 44 characters, so anything below
	// that is treated as damage and replaced rather than used to sign sessions.
	minSecretKeyLength = 44
)

// secretKeyReader is the entropy source. nil means crypto/rand; tests replace
// it to exercise the failure path.
var secretKeyReader io.Reader

// jwtSecretPath is a variable so a test can point the persistence at a
// temporary directory. Nothing on the device changes it.
var jwtSecretPath = jwtSecretFile

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

// loadOrCreateSecretKey returns the persisted signing key, generating and
// storing one the first time.
//
// The key was generated into memory on every start before this existed, so
// every restart invalidated every session and sent every open browser back to
// the login page. A restart here is not a quick event, which made that logout
// more than an inconvenience.
//
// A key that cannot be written is still better than a server that will not
// start, so a failed write is logged and the generated key is used anyway. The
// cost of that is the behaviour this replaces, not a broken server.
func loadOrCreateSecretKey() (string, error) {
	if key := readSecretKey(); key != "" {
		return key, nil
	}

	key, err := generateSecretKey(secretKeyReader)
	if err != nil {
		return "", err
	}

	if err := writeSecretKey(key); err != nil {
		log.Printf("failed to persist the jwt secret key, sessions will not survive a restart: %s", err)
	}

	return key, nil
}

// readSecretKey returns the stored key, or an empty string when there is no
// usable one. Every failure here has the same answer - generate a new key - so
// none of them is reported as an error to the caller.
func readSecretKey() string {
	data, err := os.ReadFile(jwtSecretPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("failed to read %s, generating a new jwt secret key: %s", jwtSecretPath, err)
		}

		return ""
	}

	key := strings.TrimSpace(string(data))
	if len(key) < minSecretKeyLength {
		log.Printf("%s holds no usable key, generating a new one", jwtSecretPath)

		return ""
	}

	// An earlier release, or a restore that did not preserve the mode, can
	// leave the file readable by more than root. Repair it rather than refuse
	// the key: the key is already exposed and a new one would be no better,
	// while a wrong mode that nothing corrects stays wrong for good.
	if info, err := os.Stat(jwtSecretPath); err == nil && info.Mode().Perm() != 0o600 {
		if err := os.Chmod(jwtSecretPath, 0o600); err != nil {
			log.Printf("failed to tighten the permissions on %s: %s", jwtSecretPath, err)
		}
	}

	return key
}

// writeSecretKey stores the key through a temporary file and a rename, so a
// board that loses power mid-write keeps the old key or gets the new one and
// never a truncated file that would sign every session with a few bytes.
func writeSecretKey(key string) error {
	dir := filepath.Dir(jwtSecretPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	file, err := os.CreateTemp(dir, ".jwt_secret.*")
	if err != nil {
		return err
	}

	name := file.Name()
	defer func() { _ = os.Remove(name) }()

	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()

		return err
	}

	if _, err := file.WriteString(key + "\n"); err != nil {
		_ = file.Close()

		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return os.Rename(name, jwtSecretPath)
}
