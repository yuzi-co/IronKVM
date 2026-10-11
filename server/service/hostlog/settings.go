package hostlog

import (
	"sync"

	"NanoKVM-Server/config"
)

// settingsMu guards config's HostLog block, which the owner can change while
// the receiver reads it.
var settingsMu sync.RWMutex

// CurrentSettings returns the running settings, with the defaults filled in.
func CurrentSettings() config.HostLog {
	settingsMu.RLock()
	defer settingsMu.RUnlock()

	return config.GetInstance().HostLog.WithDefaults()
}

// saveSettings writes the settings to server.yaml. Tests replace it.
var saveSettings = func(h config.HostLog) error {
	conf, err := config.Read()
	if err != nil {
		return err
	}

	conf.HostLog = h

	return config.Write(conf)
}

// ApplySettings saves the settings to server.yaml and applies them to the
// running server.
func ApplySettings(h config.HostLog) error {
	settingsMu.Lock()
	defer settingsMu.Unlock()

	if err := saveSettings(h); err != nil {
		return err
	}
	config.GetInstance().HostLog = h
	return nil
}
