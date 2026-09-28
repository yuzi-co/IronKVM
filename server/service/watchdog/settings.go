package watchdog

import (
	"sync"

	"NanoKVM-Server/config"
)

// settingsMu guards config's Watchdog block, which the owner can change while
// the watchdog reads it on every sample.
var settingsMu sync.RWMutex

// CurrentSettings returns the running watchdog settings, with the defaults
// filled in.
func CurrentSettings() config.Watchdog {
	settingsMu.RLock()
	defer settingsMu.RUnlock()

	return config.GetInstance().Watchdog.WithDefaults()
}

// saveSettings writes the settings to server.yaml. Tests replace it.
var saveSettings = func(s config.Watchdog) error {
	conf, err := config.Read()
	if err != nil {
		return err
	}

	conf.Watchdog = s

	return config.Write(conf)
}

// ApplySettings saves the settings to server.yaml and applies them to the
// running server.
func ApplySettings(s config.Watchdog) error {
	settingsMu.Lock()
	defer settingsMu.Unlock()

	if err := saveSettings(s); err != nil {
		return err
	}
	config.GetInstance().Watchdog = s
	return nil
}
