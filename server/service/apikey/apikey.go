package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/authn"
)

// File is a variable rather than a constant so tests can redirect it.
var File = "/etc/kvm/api_keys.json"

const (
	// secretPrefix marks a NanoKVM key wherever it turns up - a log, a shell
	// history, a public repository - so it can be recognised and revoked.
	secretPrefix = "nkvm_"

	// secretBytes is the size of the secret itself. 256 bits of randomness is
	// far past guessing, which is what lets the digest below be a plain hash.
	secretBytes = 32

	// idBytes identifies a key without revealing anything about it.
	idBytes = 8
)

var (
	ErrNotFound    = errors.New("api key not found")
	ErrNameTooLong = errors.New("api key name is too long")
)

// maxNameLength keeps a label from turning the store into a place to park
// arbitrary data.
const maxNameLength = 64

// Key is what the device keeps about an issued key. The secret is not part
// of it: only a digest is stored, so a copy of this file cannot be replayed
// against the device.
type Key struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username,omitempty"`
	Hash      string `json:"hash,omitempty"`
	CreatedAt int64  `json:"createdAt"`
}

// Create issues a key to a user and returns the secret. This is the only time
// the secret exists outside the caller's hands; it is not recoverable
// afterwards.
//
// The key carries the name of the user who asked for it, because the device
// now has more than one account and they do not all have the same authority.
// A key that authenticated as nobody in particular would give whoever holds it
// the run of the API, which is more than the operator who issued it may have.
func Create(name, username string) (string, *Key, error) {
	name = strings.TrimSpace(name)
	if len(name) > maxNameLength {
		return "", nil, ErrNameTooLong
	}

	secret, err := randomToken(secretBytes)
	if err != nil {
		return "", nil, err
	}
	secret = secretPrefix + secret

	id, err := randomToken(idBytes)
	if err != nil {
		return "", nil, err
	}

	key := Key{
		ID:        id,
		Name:      name,
		Username:  username,
		Hash:      hashSecret(secret),
		CreatedAt: time.Now().Unix(),
	}

	keys, err := load()
	if err != nil {
		return "", nil, err
	}

	if err := store(append(keys, key)); err != nil {
		return "", nil, err
	}

	return secret, &key, nil
}

// Verify returns the key a presented secret belongs to, if it matches one that
// has not been revoked. The digest is stripped from what it returns.
func Verify(secret string) (*Key, bool) {
	if !strings.HasPrefix(secret, secretPrefix) {
		return nil, false
	}

	keys, err := load()
	if err != nil {
		return nil, false
	}

	presented := []byte(hashSecret(secret))

	// Every key is checked even once one matches, so the work does not depend
	// on which key was presented or how many are configured.
	var matched *Key
	for _, key := range keys {
		if subtle.ConstantTimeCompare(presented, []byte(key.Hash)) == 1 {
			found := key
			found.Hash = ""
			matched = &found
		}
	}

	return matched, matched != nil
}

// List returns the keys a user holds, with their digests stripped, so the
// result is safe to hand to a client. One operator does not get to see what
// another has issued.
func List(username string) ([]Key, error) {
	keys, err := load()
	if err != nil {
		return nil, err
	}

	listed := make([]Key, 0, len(keys))
	for _, key := range keys {
		if key.Username != username {
			continue
		}

		key.Hash = ""
		listed = append(listed, key)
	}

	return listed, nil
}

// Revoke removes a key the user holds. A key that belongs to another user
// reports ErrNotFound rather than a refusal, so the call says nothing about
// what identifiers exist.
func Revoke(id, username string) error {
	keys, err := load()
	if err != nil {
		return err
	}

	kept := make([]Key, 0, len(keys))
	for _, key := range keys {
		if key.ID == id && key.Username == username {
			continue
		}

		kept = append(kept, key)
	}

	if len(kept) == len(keys) {
		return ErrNotFound
	}

	return store(kept)
}

// hashSecret digests a secret for storage.
//
// A plain hash rather than bcrypt: the secret is 256 random bits, so there is
// nothing to guess and no work factor to hide behind, and this runs on every
// request that presents a key. A password KDF here would cost more per API
// call than serving the request.
func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		log.Errorf("failed to generate random bytes: %s", err)
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func load() ([]Key, error) {
	content, err := os.ReadFile(File)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		log.Errorf("failed to read api keys: %s", err)
		return nil, err
	}

	var keys []Key
	if err := json.Unmarshal(content, &keys); err != nil {
		log.Errorf("failed to parse api keys: %s", err)
		return nil, err
	}

	return adopt(keys), nil
}

// adopt gives a key that names no user to the system account.
//
// Keys issued before the device had more than one account carry no name. The
// account that owned the device then is the one the migration turned into the
// system account, so this restores the identity the key was issued with rather
// than guessing at one. A key that cannot be placed authenticates as nobody,
// which the middleware refuses.
func adopt(keys []Key) []Key {
	owner := ""
	for i, key := range keys {
		if key.Username != "" {
			continue
		}

		if owner == "" {
			owner = systemAccount()
			if owner == "" {
				break
			}
		}

		keys[i].Username = owner
	}

	return keys
}

// systemAccount is a variable so tests can name an owner without a store.
var systemAccount = func() string {
	users, err := authn.DefaultStore.List()
	if err != nil {
		log.Errorf("failed to read accounts while placing an api key: %s", err)
		return ""
	}

	for _, user := range users {
		if user.SystemAccount {
			return user.Username
		}
	}

	return ""
}

func store(keys []Key) error {
	content, err := json.Marshal(keys)
	if err != nil {
		log.Errorf("failed to serialise api keys: %s", err)
		return err
	}

	if err := os.MkdirAll(filepath.Dir(File), 0o755); err != nil {
		log.Errorf("failed to create directory for api keys: %s", err)
		return err
	}

	// The digests are credentials in their own right for anyone who can read
	// them alongside a captured request, so only root gets the file.
	if err := os.WriteFile(File, content, 0o600); err != nil {
		log.Errorf("failed to write api keys: %s", err)
		return err
	}

	return nil
}
