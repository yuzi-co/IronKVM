package vnc

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// PasswordFile holds the password for plain VNC authentication. The protocol
// needs the password itself, not a hash, to check the client's answer, so the
// file is readable by root only and never part of server.yaml, which is 0644.
const PasswordFile = "/etc/kvm/vnc.passwd"

// Bounds on the VNC password. DES uses the first eight bytes, so a longer
// password would only look stronger than it is.
const (
	minPasswordLength = 6
	maxPasswordLength = vncPasswordLength
)

var errPasswordLength = errors.New("the VNC password must be 6 to 8 characters")

func validatePassword(password string) error {
	if len(password) < minPasswordLength || len(password) > maxPasswordLength {
		return errPasswordLength
	}
	for _, r := range password {
		if r < 0x20 || r > 0x7e {
			return errors.New("the VNC password must be printable ASCII")
		}
	}
	return nil
}

// FilePassword keeps the VNC password in a file.
type FilePassword struct {
	Path string
}

func (f FilePassword) Get() (string, bool, error) {
	data, err := os.ReadFile(f.Path)
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	password := strings.TrimRight(string(data), "\n")
	if password == "" {
		return "", false, nil
	}
	return password, true, nil
}

// Set writes the password through a temporary file and a rename, so a reader
// never sees half of it and the file is never readable by others.
func (f FilePassword) Set(password string) error {
	if err := validatePassword(password); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(f.Path), ".vnc.passwd-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()

	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.WriteString(password + "\n"); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), f.Path)
}
