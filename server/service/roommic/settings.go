package roommic

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// SettingsFile holds the administrator's choice, next to the other device
// toggles in /etc/kvm. A missing file means the defaults: not allowed. A
// variable so tests can point it somewhere writable.
var SettingsFile = "/etc/kvm/room-mic"

const (
	// MinGain and MaxGain bound the codec's capture PGA, 2 dB a step.
	MinGain = 0
	MaxGain = 24

	// DefaultGain is 36 dB. At full gain the room's floor measured -33 dBFS,
	// mostly hum below 500 Hz; at 18 it sits near -47 dBFS with speech and
	// claps well clear of it (ironkvm-dist trials 57 and 58).
	DefaultGain = 18
)

// Settings is what the administrator decides.
type Settings struct {
	// Allowed lets viewers switch the microphone on. Off by default.
	Allowed bool `json:"allowed"`
	// Gain is the capture PGA setting, MinGain to MaxGain.
	Gain int `json:"gain"`
}

// DefaultSettings is what a board that was never configured runs.
func DefaultSettings() Settings {
	return Settings{Allowed: false, Gain: DefaultGain}
}

var errGainRange = fmt.Errorf("gain must be between %d and %d", MinGain, MaxGain)

func (s Settings) validate() error {
	if s.Gain < MinGain || s.Gain > MaxGain {
		return errGainRange
	}

	return nil
}

// loadSettings reads the file, or returns the defaults when there is none. A
// file that cannot be parsed also gives the defaults, which keep the
// microphone off.
func loadSettings() (Settings, error) {
	data, err := os.ReadFile(SettingsFile)
	if errors.Is(err, os.ErrNotExist) {
		return DefaultSettings(), nil
	}
	if err != nil {
		return DefaultSettings(), err
	}

	settings := DefaultSettings()
	if err := json.Unmarshal(data, &settings); err != nil {
		return DefaultSettings(), err
	}
	if err := settings.validate(); err != nil {
		settings.Gain = DefaultGain
	}

	return settings, nil
}

// saveSettings writes the file through a temporary one, so a power cut leaves
// the old choice or the new one and never half of either.
func saveSettings(settings Settings) error {
	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	dir := filepath.Dir(SettingsFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".room-mic-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()

	if _, err := tmp.Write(append(data, '\n')); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), SettingsFile)
}
