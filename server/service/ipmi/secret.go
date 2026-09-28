package ipmi

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// KeyFile holds the key IPMI passwords are sealed under.
const KeyFile = "/etc/kvm/ipmi.key"

// sealedPrefix marks the format of a sealed password, so a later format
// can be told apart.
const sealedPrefix = "v1:"

var errNoKey = errors.New("the IPMI key is missing")

// Keyring seals IPMI passwords for the account file and opens them again
// for the RAKP handshake. The key is 32 random bytes in its own file, made
// on the first seal.
//
// The key sits next to what it protects, so this is no defence against
// someone who can read /etc/kvm. What it does is keep the password out of
// the account file itself, which the legacy account mirror and any copy of
// that file would otherwise carry in the clear.
type Keyring struct {
	path string

	mu  sync.Mutex
	key []byte
}

func NewKeyring(path string) *Keyring {
	return &Keyring{path: path}
}

// load returns the key, reading it the first time. With create it makes the
// key when there is none.
func (k *Keyring) load(create bool) ([]byte, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.key != nil {
		return k.key, nil
	}

	data, err := os.ReadFile(k.path)
	if err == nil {
		if len(data) != 32 {
			return nil, errors.New("the IPMI key file is damaged")
		}
		k.key = data
		return k.key, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if !create {
		return nil, errNoKey
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	if err := writeKey(k.path, key); err != nil {
		return nil, err
	}
	k.key = key
	return k.key, nil
}

// writeKey writes the key readable by root alone, in one rename.
func writeKey(path string, key []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".ipmi-key-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if err = tmp.Chmod(0o600); err == nil {
		_, err = tmp.Write(key)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func (k *Keyring) aead(create bool) (cipher.AEAD, error) {
	key, err := k.load(create)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// Seal encrypts an account's IPMI password. The user name is bound in as
// additional data, so a sealed value copied to another account does not
// open.
func (k *Keyring) Seal(username, password string) (string, error) {
	aead, err := k.aead(true)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := aead.Seal(nonce, nonce, []byte(password), []byte(username))
	return sealedPrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

// Open returns the password Seal sealed for this account.
func (k *Keyring) Open(username, sealed string) ([]byte, error) {
	encoded, ok := strings.CutPrefix(sealed, sealedPrefix)
	if !ok {
		return nil, errors.New("unknown sealed password format")
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	aead, err := k.aead(false)
	if err != nil {
		return nil, err
	}
	if len(raw) < aead.NonceSize() {
		return nil, errors.New("sealed password too short")
	}
	return aead.Open(nil, raw[:aead.NonceSize()], raw[aead.NonceSize():], []byte(username))
}
