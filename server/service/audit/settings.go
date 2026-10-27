package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// The settings live in /etc/kvm, which both firmware slots read, as small
// JSON files. A slot that does not know a field ignores it, so a field added
// here never stops the other slot from starting.
//
// The variable lets tests point at a directory.
var SettingsFile = "/etc/kvm/audit.json"

// Settings are the audit log's own options.
type Settings struct {
	// Syslog sends every event to the local syslog socket as well, which
	// forwards it to the collector named under Settings > Network > Remote
	// syslog, if one is.
	Syslog bool `json:"syslog"`
}

var (
	fileMu   sync.Mutex
	settings Settings
	loaded   bool
)

// CurrentSettings returns the settings in force.
func CurrentSettings() Settings {
	fileMu.Lock()
	defer fileMu.Unlock()
	if !loaded {
		_ = readJSON(SettingsFile, &settings)
		loaded = true
	}
	return settings
}

// SaveSettings stores new settings.
func SaveSettings(s Settings) error {
	fileMu.Lock()
	defer fileMu.Unlock()
	if err := writeJSON(SettingsFile, s); err != nil {
		return err
	}
	settings, loaded = s, true
	return nil
}

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".audit-*.tmp")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// Syslog formats an event as one line for the syslog forward.
func Syslog(e Event) string {
	var b strings.Builder
	b.WriteString("audit:")
	add := func(k, v string) {
		if v == "" {
			return
		}
		b.WriteByte(' ')
		b.WriteString(k)
		b.WriteByte('=')
		if strings.ContainsAny(v, " \"=") {
			v = strconv.Quote(v)
		}
		b.WriteString(v)
	}
	add("actor", e.Actor)
	add("via", e.Via)
	add("key", e.Key)
	add("session", e.Session)
	add("role", e.Role)
	add("ip", e.IP)
	add("category", e.Category)
	add("action", e.Action)
	add("target", e.Target)
	add("outcome", e.Outcome)
	add("tier", e.Tier)
	add("detail", e.Detail)
	if e.Repeat > 1 {
		add("repeat", strconv.Itoa(e.Repeat))
	}
	return b.String()
}
