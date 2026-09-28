package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// IPMI authentication is weak by design, so nothing turns the service on
// but the owner.
func TestIPMIIsOffByDefault(t *testing.T) {
	if defaultConfig.IPMI.Enabled {
		t.Fatal("defaultConfig turns IPMI on")
	}

	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if conf := loadYAML(t, string(data)); conf.IPMI.Enabled {
		t.Fatal("the default configuration loads with IPMI on")
	}
}

func TestAnOldConfigLoadsWithIPMIOff(t *testing.T) {
	old := "proto: http\nport:\n  http: 80\n  https: 443\nauthentication: enable\n"
	if conf := loadYAML(t, old); conf.IPMI.Enabled {
		t.Fatal("an old server.yaml loads with IPMI on")
	}
}

func TestIPMIOnIsKeptUnderIPMIEnabled(t *testing.T) {
	conf := loadYAML(t, "ipmi:\n  enabled: true\n")
	if !conf.IPMI.Enabled {
		t.Fatal("ipmi.enabled: true did not load")
	}

	data, err := yaml.Marshal(conf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "ipmi:\n    enabled: true\n") {
		t.Fatalf("ipmi.enabled was not written:\n%s", data)
	}
}
