package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Redfish was always on before it had a setting, so it stays on by default.
func TestRedfishIsEnabledByDefault(t *testing.T) {
	if !defaultConfig.Redfish.IsEnabled() {
		t.Fatal("defaultConfig says Redfish is off")
	}

	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "redfish:\n    enabled: true\n") {
		t.Fatalf("a new server.yaml does not show the setting:\n%s", data)
	}
	if conf := loadYAML(t, string(data)); !conf.Redfish.IsEnabled() {
		t.Fatal("the default configuration loads with Redfish off")
	}
}

// A server.yaml written before the setting existed has no redfish block, and
// must keep the service it had.
func TestAnOldConfigWithoutTheKeyLoadsWithRedfishOn(t *testing.T) {
	old := "proto: http\nport:\n  http: 80\n  https: 443\nauthentication: enable\n"

	if conf := loadYAML(t, old); !conf.Redfish.IsEnabled() {
		t.Fatal("an old server.yaml loads with Redfish off")
	}

	var conf Config
	if err := yaml.Unmarshal([]byte(old), &conf); err != nil {
		t.Fatal(err)
	}
	if !conf.Redfish.IsEnabled() {
		t.Fatal("config.Read of an old server.yaml says Redfish is off")
	}

	// Saving such a file for another setting must not write a value the
	// owner never chose.
	data, err := yaml.Marshal(&conf)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "redfish") {
		t.Fatalf("an untouched setting was written:\n%s", data)
	}
}

// Off is kept under redfish.enabled, for viper at start and for the
// Read/Write pair that saves settings.
func TestRedfishOffIsKeptUnderRedfishEnabled(t *testing.T) {
	conf := loadYAML(t, "redfish:\n  enabled: false\n")
	if conf.Redfish.IsEnabled() {
		t.Fatal("redfish.enabled: false did not load")
	}

	data, err := yaml.Marshal(&conf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "redfish:\n    enabled: false\n") {
		t.Fatalf("saved as:\n%s", data)
	}

	var back Config
	if err := yaml.Unmarshal(data, &back); err != nil || back.Redfish.IsEnabled() {
		t.Fatalf("round trip lost the setting: %v", err)
	}

	if on := loadYAML(t, "redfish:\n  enabled: true\n"); !on.Redfish.IsEnabled() {
		t.Fatal("redfish.enabled: true did not load")
	}
}
