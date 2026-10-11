package webui

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/crypto/blake2b"
)

// This file verifies minisign signatures, the format abslots/sign-image.sh
// writes for slot images and generic/slot checks with the minisign binary. The
// board does not carry that binary on every slot, and the format is small:
//
//	public key   untrusted comment: ...
//	             base64("Ed" | key id (8) | Ed25519 public key (32))
//	signature    untrusted comment: ...
//	             base64(algorithm (2) | key id (8) | Ed25519 signature (64))
//	             trusted comment: <text>
//	             base64(Ed25519 signature over signature (64) | <text>)
//
// The algorithm is "Ed" for a signature over the file itself, or "ED" for one
// over the file's BLAKE2b-512 hash, which minisign has made by default since
// 0.10. Both are accepted. The second signature binds the trusted comment, and
// it is checked too: a comment that has been edited is a signature that fails.

const (
	untrustedPrefix = "untrusted comment: "
	trustedPrefix   = "trusted comment: "
	keyIDLen        = 8
	maxKeyFileSize  = 4 << 10
	maxSigFileSize  = 4 << 10
)

var (
	// ErrNoSignature means there is no signature to check.
	ErrNoSignature = errors.New("the package has no signature beside it")
	// ErrUnknownKey means no key in the keyring made the signature.
	ErrUnknownKey = errors.New("no key in the keyring made the signature")
	// ErrAltered means a key in the keyring made the signature, and the file
	// or the trusted comment changed after it did.
	ErrAltered = errors.New("the package does not match its signature: it was changed after it was signed")
)

// PublicKey is one minisign public key.
type PublicKey struct {
	ID  [keyIDLen]byte
	Key ed25519.PublicKey
	// Name is the file the key came from, for messages.
	Name string
}

// Signature is one parsed .minisig file.
type Signature struct {
	Algorithm      string
	KeyID          [keyIDLen]byte
	Sig            []byte
	TrustedComment string
	GlobalSig      []byte
}

// ParsePublicKey reads a minisign public key file. The base64 line may also be
// given alone, as minisign -P takes it.
func ParsePublicKey(data []byte) (*PublicKey, error) {
	var line string
	for _, l := range splitLines(data) {
		if l == "" || strings.HasPrefix(l, untrustedPrefix) {
			continue
		}
		line = l
		break
	}
	raw, err := base64.StdEncoding.DecodeString(line)
	if err != nil || len(raw) != 2+keyIDLen+ed25519.PublicKeySize || string(raw[:2]) != "Ed" {
		return nil, errors.New("not a minisign public key")
	}
	pk := &PublicKey{Key: ed25519.PublicKey(append([]byte(nil), raw[2+keyIDLen:]...))}
	copy(pk.ID[:], raw[2:2+keyIDLen])
	return pk, nil
}

// ParseSignature reads a .minisig file.
func ParseSignature(data []byte) (*Signature, error) {
	lines := splitLines(data)
	if len(lines) < 4 || !strings.HasPrefix(lines[0], untrustedPrefix) || !strings.HasPrefix(lines[2], trustedPrefix) {
		return nil, errors.New("not a minisign signature")
	}
	raw, err := base64.StdEncoding.DecodeString(lines[1])
	if err != nil || len(raw) != 2+keyIDLen+ed25519.SignatureSize {
		return nil, errors.New("not a minisign signature")
	}
	alg := string(raw[:2])
	if alg != "Ed" && alg != "ED" {
		return nil, fmt.Errorf("unsupported signature algorithm %q", alg)
	}
	global, err := base64.StdEncoding.DecodeString(lines[3])
	if err != nil || len(global) != ed25519.SignatureSize {
		return nil, errors.New("not a minisign signature: bad trusted comment signature")
	}
	sig := &Signature{
		Algorithm:      alg,
		Sig:            append([]byte(nil), raw[2+keyIDLen:]...),
		TrustedComment: strings.TrimPrefix(lines[2], trustedPrefix),
		GlobalSig:      global,
	}
	copy(sig.KeyID[:], raw[2:2+keyIDLen])
	return sig, nil
}

func splitLines(data []byte) []string {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return lines
}

// Verify checks sig against the content of r with key. It returns
// ErrUnknownKey when the key ids differ and ErrAltered when they match and the
// signature does not.
func (pk *PublicKey) Verify(r io.Reader, sig *Signature) error {
	if pk.ID != sig.KeyID {
		return ErrUnknownKey
	}
	message, err := signedMessage(r, sig.Algorithm)
	if err != nil {
		return err
	}
	return pk.verifyMessage(message, sig)
}

// signedMessage is what the signature covers: the BLAKE2b-512 hash of the
// content for "ED", the content itself for the legacy "Ed".
func signedMessage(r io.Reader, algorithm string) ([]byte, error) {
	if algorithm == "ED" {
		h, _ := blake2b.New512(nil)
		if _, err := io.Copy(h, r); err != nil {
			return nil, err
		}
		return h.Sum(nil), nil
	}
	// The size of a package is bounded before it reaches this.
	return io.ReadAll(r)
}

func (pk *PublicKey) verifyMessage(message []byte, sig *Signature) error {
	if pk.ID != sig.KeyID {
		return ErrUnknownKey
	}
	if !ed25519.Verify(pk.Key, message, sig.Sig) {
		return ErrAltered
	}
	global := append(append([]byte(nil), sig.Sig...), sig.TrustedComment...)
	if !ed25519.Verify(pk.Key, global, sig.GlobalSig) {
		return ErrAltered
	}
	return nil
}

// Policy is where trust comes from: the same keyring and the same switch that
// generic/slot uses for slot images, so one key on the board covers both.
type Policy struct {
	// KeyDirs are searched for *.pub files; their union is the keyring.
	KeyDirs []string
	// CheckOff, when it exists, turns the check off.
	CheckOff string
}

// DefaultPolicy is the slot installer's: the key the image was built with and
// the owner's keys on /data. IRONKVM_KEYS_DIRS (space separated) and
// IRONKVM_SIGNATURE_CHECK_OFF replace the two for a test server; the variables
// are read from the server's own environment, which only root can set.
func DefaultPolicy() Policy {
	p := Policy{
		KeyDirs:  []string{"/usr/share/ironkvm/keys", "/data/ironkvm/keys"},
		CheckOff: "/data/ironkvm/no-signature-check",
	}
	if v, ok := os.LookupEnv("IRONKVM_KEYS_DIRS"); ok {
		p.KeyDirs = strings.Fields(v)
	}
	if v, ok := os.LookupEnv("IRONKVM_SIGNATURE_CHECK_OFF"); ok {
		p.CheckOff = v
	}
	return p
}

// Verdict says what a check established.
type Verdict struct {
	// Checked is false when nothing was checked: the switch is on or the
	// keyring is empty. Warning then says which.
	Checked bool   `json:"checked"`
	Signer  string `json:"signer,omitempty"`
	Warning string `json:"warning,omitempty"`
}

// Keys loads every *.pub in the key directories, in a stable order. A file
// that does not parse is skipped: one bad key must not stop the others.
func (p Policy) Keys() []*PublicKey {
	var keys []*PublicKey
	for _, dir := range p.KeyDirs {
		matches, _ := filepath.Glob(filepath.Join(dir, "*.pub"))
		sort.Strings(matches)
		for _, path := range matches {
			info, err := os.Stat(path)
			if err != nil || !info.Mode().IsRegular() || info.Size() > maxKeyFileSize {
				continue
			}
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			pk, err := ParsePublicKey(data)
			if err != nil {
				continue
			}
			pk.Name = filepath.Base(path)
			keys = append(keys, pk)
		}
	}
	return keys
}

// State describes the check without running it, for the update page.
func (p Policy) State() string {
	if p.CheckOff != "" {
		if _, err := os.Stat(p.CheckOff); err == nil {
			return "off"
		}
	}
	if len(p.Keys()) == 0 {
		return "no-keys"
	}
	return "on"
}

// VerifyFile checks file against sigPath, as generic/slot checks an image:
// the switch turns the check off, an empty keyring checks nothing, and
// otherwise one key in the keyring must verify the signature.
//
// A signature says who made a package; it never decides what the owner may
// run. Both ways out are the owner's, on the owner's board, and each leaves a
// warning in the verdict.
func (p Policy) VerifyFile(file, sigPath string) (Verdict, error) {
	if p.CheckOff != "" {
		if _, err := os.Stat(p.CheckOff); err == nil {
			return Verdict{Warning: fmt.Sprintf("%s exists, so nothing checked who made this package", p.CheckOff)}, nil
		}
	}
	keys := p.Keys()
	if len(keys) == 0 {
		return Verdict{Warning: "the keyring is empty, so nothing checked who made this package"}, nil
	}

	info, err := os.Stat(sigPath)
	if err != nil || !info.Mode().IsRegular() {
		return Verdict{}, ErrNoSignature
	}
	if info.Size() > maxSigFileSize {
		return Verdict{}, errors.New("the signature file is too large")
	}
	sigData, err := os.ReadFile(sigPath)
	if err != nil {
		return Verdict{}, err
	}
	sig, err := ParseSignature(sigData)
	if err != nil {
		return Verdict{}, err
	}

	f, err := os.Open(file)
	if err != nil {
		return Verdict{}, err
	}
	message, err := signedMessage(f, sig.Algorithm)
	_ = f.Close()
	if err != nil {
		return Verdict{}, err
	}
	altered := false
	for _, pk := range keys {
		err := pk.verifyMessage(message, sig)
		if err == nil {
			return Verdict{Checked: true, Signer: pk.Name}, nil
		}
		if errors.Is(err, ErrAltered) {
			altered = true
		}
	}
	if altered {
		return Verdict{}, ErrAltered
	}
	return Verdict{}, ErrUnknownKey
}
