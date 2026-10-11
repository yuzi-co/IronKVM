package nut

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"NanoKVM-Server/config"
)

// PasswordFile holds the NUT password. It is readable by root only and never
// part of server.yaml, which is 0644, and nothing exports it.
const PasswordFile = "/etc/kvm/nut.passwd"

// settingsMu guards config's NUT block, which the owner can change while the
// monitor reads it on every poll.
var settingsMu sync.RWMutex

// CurrentSettings returns the running settings, with the defaults filled in.
func CurrentSettings() config.NUT {
	settingsMu.RLock()
	defer settingsMu.RUnlock()

	return config.GetInstance().NUT.WithDefaults()
}

// saveSettings writes the settings to server.yaml. Tests replace it.
var saveSettings = func(s config.NUT) error {
	conf, err := config.Read()
	if err != nil {
		return err
	}

	conf.NUT = s

	return config.Write(conf)
}

// ApplySettings saves the settings to server.yaml and applies them to the
// running server.
func ApplySettings(s config.NUT) error {
	settingsMu.Lock()
	defer settingsMu.Unlock()

	if err := saveSettings(s); err != nil {
		return err
	}
	config.GetInstance().NUT = s
	return nil
}

// The shapes NUT accepts for the names. A UPS name and a user name are words
// of letters, digits and a few marks; the password is printable ASCII.
var (
	upsNamePattern  = regexp.MustCompile(`^[A-Za-z0-9._@-]{1,64}$`)
	hostNamePattern = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9.-]{0,251}[A-Za-z0-9])?$`)
	ipv6Pattern     = regexp.MustCompile(`^[0-9A-Fa-f:.]+$`)
)

const maxPasswordLength = 128

var errPassword = errors.New("the NUT password must be printable ASCII without spaces, up to 128 characters")

func validPassword(p string) error {
	if p == "" || len(p) > maxPasswordLength {
		return errPassword
	}
	for _, r := range p {
		if r <= 0x20 || r > 0x7e {
			return errPassword
		}
	}
	return nil
}

// ValidHost says whether s can be a NUT server address: a host name or an IP
// address. IPv6 is accepted bare, without brackets.
func ValidHost(s string) bool {
	if s == "" {
		return false
	}
	if strings.Contains(s, ":") {
		return ipv6Pattern.MatchString(s)
	}
	return hostNamePattern.MatchString(s)
}

// FilePassword keeps the NUT password in a file.
type FilePassword struct {
	Path string
}

// Get returns the password, empty when there is none.
func (f FilePassword) Get() (string, error) {
	data, err := os.ReadFile(f.Path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(data), "\n"), nil
}

// Set writes the password through a temporary file and a rename, so a reader
// never sees half of it and the file is never readable by others. An empty
// password removes the file.
func (f FilePassword) Set(password string) error {
	if password == "" {
		err := os.Remove(f.Path)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if err := validPassword(password); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(f.Path), ".nut.passwd-*")
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
