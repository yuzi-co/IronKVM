package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// The watchdog presses buttons on the host, so nothing turns it on but the
// owner.
func TestTheWatchdogIsOffByDefault(t *testing.T) {
	if defaultConfig.Watchdog.Enabled {
		t.Fatal("defaultConfig turns the watchdog on")
	}

	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if conf := loadYAML(t, string(data)); conf.Watchdog.Enabled {
		t.Fatal("the default configuration loads with the watchdog on")
	}
}

// A server.yaml written before the watchdog existed has no watchdog block. It
// loads with the watchdog off and with the defaults.
func TestAnOldConfigLoadsWithTheWatchdogOff(t *testing.T) {
	old := "proto: http\nport:\n  http: 80\n  https: 443\nauthentication: enable\n"

	conf := loadYAML(t, old)
	if conf.Watchdog.Enabled {
		t.Fatal("an old server.yaml loads with the watchdog on")
	}

	got := conf.Watchdog.WithDefaults()
	want := Watchdog{
		TimeoutMinutes:  DefaultWatchdogTimeoutMinutes,
		Action:          DefaultWatchdogAction,
		CooldownMinutes: DefaultWatchdogCooldownMinutes,
		MaxPerHour:      DefaultWatchdogMaxPerHour,
	}
	if got != want {
		t.Fatalf("defaults: got %+v, want %+v", got, want)
	}
}

// The keys live under watchdog, both for viper at start and for the
// Read/Write pair that saves settings.
func TestTheWatchdogSettingsAreKeptUnderWatchdog(t *testing.T) {
	content := "watchdog:\n  enabled: true\n  timeoutMinutes: 7\n  action: power\n" +
		"  cooldownMinutes: 20\n  maxPerHour: 2\n  pingHost: 192.168.1.10\n"
	want := Watchdog{
		Enabled:         true,
		TimeoutMinutes:  7,
		Action:          "power",
		CooldownMinutes: 20,
		MaxPerHour:      2,
		PingHost:        "192.168.1.10",
	}

	conf := loadYAML(t, content)
	if conf.Watchdog != want {
		t.Fatalf("viper loaded %+v", conf.Watchdog)
	}

	data, err := yaml.Marshal(&conf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "watchdog:\n    enabled: true\n    timeoutMinutes: 7\n") {
		t.Fatalf("saved as:\n%s", data)
	}

	var back Config
	if err := yaml.Unmarshal(data, &back); err != nil || back.Watchdog != want {
		t.Fatalf("round trip: %+v, %v", back.Watchdog, err)
	}
}
