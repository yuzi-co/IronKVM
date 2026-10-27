package wireguard

import (
	"encoding/json"
	"errors"
	"os"

	"NanoKVM-Server/utils/atomicfile"
)

// configFile holds the private key, so it is 0600 and written by rename. A var
// so the tests can point it at a scratch directory.
var configFile = "/etc/kvm/wireguard.json"

// Load returns the stored configuration; a board without one has none.
func Load() (Config, error) {
	var c Config
	data, err := os.ReadFile(configFile)
	if errors.Is(err, os.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(data, &c)
	return c, err
}

// Save writes c to a temporary file that only root can read, then renames it
// over the stored one, so the key is never readable by others, not even for
// a moment, and a crash leaves either the old file or the new one.
func Save(c Config) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return atomicfile.WriteFile(configFile, append(data, '\n'), 0o600)
}

// Remove deletes the stored configuration.
func Remove() error {
	err := os.Remove(configFile)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
