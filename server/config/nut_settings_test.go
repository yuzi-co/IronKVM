package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// The NUT client presses buttons on the host, so nothing turns it on but the
// owner.
func TestTheNUTClientIsOffByDefault(t *testing.T) {
	if defaultConfig.NUT.Enabled {
		t.Fatal("defaultConfig turns the NUT client on")
	}

	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if conf := loadYAML(t, string(data)); conf.NUT.Enabled {
		t.Fatal("the default configuration loads with the NUT client on")
	}
}

// A server.yaml written before the client existed has no nut block. It loads
// with the client off and with the defaults.
func TestAnOldConfigLoadsWithTheNUTClientOff(t *testing.T) {
	old := "proto: http\nport:\n  http: 80\n  https: 443\nauthentication: enable\n"

	conf := loadYAML(t, old)
	if conf.NUT.Enabled {
		t.Fatal("an old server.yaml loads with the NUT client on")
	}
	got := conf.NUT.WithDefaults()
	if got.Port != DefaultNUTPort || got.PowerOnDelaySeconds != DefaultNUTRestoreDelaySeconds {
		t.Fatalf("%+v", got)
	}
}

// The password is kept outside server.yaml.
func TestTheNUTBlockHasNoPassword(t *testing.T) {
	data, err := yaml.Marshal(NUT{Enabled: true, Host: "nas", UPS: "ups", Username: "kvm"})
	if err != nil {
		t.Fatal(err)
	}
	if s := string(data); strings.Contains(s, "password") {
		t.Fatalf("server.yaml would carry a password field:\n%s", s)
	}
}
