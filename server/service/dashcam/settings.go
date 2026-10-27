package dashcam

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"NanoKVM-Server/utils/atomicfile"
)

// File holds the settings. It lives under /etc/kvm, which both slots share:
// keep the format backward compatible, and ignore keys you do not know.
const File = "/etc/kvm/dashcam.json"

// The limits of the settings.
const (
	MinSeconds     = 30
	MaxSeconds     = 120
	DefaultSeconds = 60

	MaxPostSeconds     = 30
	DefaultPostSeconds = 10

	// MinKbps and MaxKbps bound the bitrate cap. 0 means the stream's own
	// bitrate, up to MaxRingBytes.
	MinKbps = 250
	MaxKbps = 20000

	MinStorageMB     = 20
	MaxStorageMB     = 100000
	DefaultStorageMB = 200
)

// Triggers says which events save a clip. A press of the button on the page
// and a call of the save route always do.
type Triggers struct {
	// ATX is a press of the power or reset button, from the page, Redfish,
	// IPMI, a macro or the agent, and a press by the host watchdog.
	ATX bool `json:"atx"`
	// Signal is the HDMI signal going and coming back.
	Signal bool `json:"signal"`
	// Failsafe is a server that starts in failsafe mode.
	Failsafe bool `json:"failsafe"`
}

// Settings are the dashcam's settings as the file keeps them.
type Settings struct {
	// Enabled keeps the last Seconds of the stream in memory, which keeps
	// the encoder running with nobody watching. Off by default.
	Enabled bool `json:"enabled"`
	// Seconds is how much of the stream before a trigger a clip holds.
	Seconds int `json:"seconds"`
	// PostSeconds is how much after the trigger it holds as well.
	PostSeconds int `json:"postSeconds"`
	// MaxKbps caps the memory the ring takes, as the bitrate it can hold
	// Seconds of. The encoder is not touched: a stream above the cap gives a
	// shorter clip. 0 holds the stream as it is, up to MaxRingBytes.
	MaxKbps int `json:"maxKbps"`
	// StorageMB is how much of /data the clips may take. The oldest go
	// first.
	StorageMB int      `json:"storageMB"`
	Triggers  Triggers `json:"triggers"`
}

// Defaults are the settings of a board that has none.
func Defaults() Settings {
	return Settings{
		Seconds:     DefaultSeconds,
		PostSeconds: DefaultPostSeconds,
		StorageMB:   DefaultStorageMB,
		Triggers:    Triggers{ATX: true, Signal: true, Failsafe: true},
	}
}

// Validate checks the limits.
func (s Settings) Validate() error {
	if s.Seconds < MinSeconds || s.Seconds > MaxSeconds {
		return fmt.Errorf("the duration is %d to %d seconds", MinSeconds, MaxSeconds)
	}
	if s.PostSeconds < 0 || s.PostSeconds > MaxPostSeconds {
		return fmt.Errorf("the time after the trigger is 0 to %d seconds", MaxPostSeconds)
	}
	if s.MaxKbps != 0 && (s.MaxKbps < MinKbps || s.MaxKbps > MaxKbps) {
		return fmt.Errorf("the bitrate cap is 0 or %d to %d kbit/s", MinKbps, MaxKbps)
	}
	if s.StorageMB < MinStorageMB || s.StorageMB > MaxStorageMB {
		return fmt.Errorf("the storage cap is %d to %d MB", MinStorageMB, MaxStorageMB)
	}
	return nil
}

// Load reads the settings. A missing file is the defaults; a key the file
// does not have keeps its default.
func Load(path string) (Settings, error) {
	s := Defaults()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return Defaults(), fmt.Errorf("%s is not valid JSON", path)
	}
	if err := s.Validate(); err != nil {
		return Defaults(), fmt.Errorf("%s: %w", path, err)
	}
	return s, nil
}

// Save writes the settings.
func Save(path string, s Settings) error {
	if err := s.Validate(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return atomicfile.WriteFile(path, data, 0o644)
}
