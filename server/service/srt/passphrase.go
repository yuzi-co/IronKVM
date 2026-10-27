package srt

import (
	"errors"
	"os"
	"strings"

	"NanoKVM-Server/utils/atomicfile"
)

// PassphraseFile holds the SRT passphrase. SRT derives the stream's AES key
// from the passphrase itself, so the board must keep it, not a hash. The file
// is readable by root only and never part of server.yaml, which is 0644, nor
// of the settings file a backup exports.
const PassphraseFile = "/etc/kvm/srt.passphrase"

// SRT's own bounds on a passphrase (the SRTO_PASSPHRASE socket option).
const (
	MinPassphraseLength = 10
	MaxPassphraseLength = 79
)

var errPassphraseLength = errors.New("the SRT passphrase must be 10 to 79 characters")

// ValidatePassphrase checks a passphrase against SRT's bounds. Printable
// ASCII only, because it is typed into a URL as often as into a form.
func ValidatePassphrase(passphrase string) error {
	if len(passphrase) < MinPassphraseLength || len(passphrase) > MaxPassphraseLength {
		return errPassphraseLength
	}
	for _, r := range passphrase {
		if r < 0x20 || r > 0x7e {
			return errors.New("the SRT passphrase must be printable ASCII")
		}
	}
	return nil
}

// PassphraseStore keeps the passphrase.
type PassphraseStore interface {
	// Get returns the passphrase, and false when none is set.
	Get() (string, bool, error)
	Set(passphrase string) error
}

// FilePassphrase keeps the passphrase in a file.
type FilePassphrase struct {
	Path string
}

func (f FilePassphrase) Get() (string, bool, error) {
	data, err := os.ReadFile(f.Path)
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	passphrase := strings.TrimRight(string(data), "\n")
	if passphrase == "" {
		return "", false, nil
	}
	return passphrase, true, nil
}

// Set writes the passphrase through a temporary file and a rename, so a
// reader never sees half of it and the file is never readable by others.
func (f FilePassphrase) Set(passphrase string) error {
	if err := ValidatePassphrase(passphrase); err != nil {
		return err
	}

	return atomicfile.WriteFile(f.Path, []byte(passphrase+"\n"), 0o600)
}
