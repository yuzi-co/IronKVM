package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// The VNC server opens a port and takes input, so nothing turns it on but the
// owner.
func TestTheVNCServerIsOffByDefault(t *testing.T) {
	if defaultConfig.VNC.Enabled {
		t.Fatal("defaultConfig turns the VNC server on")
	}

	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if conf := loadYAML(t, string(data)); conf.VNC.Enabled {
		t.Fatal("the default configuration loads with the VNC server on")
	}
}

// A server.yaml without the block loads with the server off and the defaults.
func TestAnOldConfigLoadsWithTheVNCServerOff(t *testing.T) {
	conf := loadYAML(t, "proto: http\nport:\n  http: 80\n  https: 443\n")
	want := VNC{Port: DefaultVNCPort, MaxFPS: DefaultVNCMaxFPS}
	if got := conf.VNC.WithDefaults(); got != want {
		t.Fatalf("an old server.yaml loads VNC as %+v, want %+v", got, want)
	}
}

// The settings are kept under vnc, for viper at start and for the Read/Write
// pair that saves settings.
func TestVNCSettingsAreKeptUnderVNC(t *testing.T) {
	text := "vnc:\n  enabled: true\n  port: 5901\n  maxFps: 30\n  vncAuth: true\n"
	want := VNC{Enabled: true, Port: 5901, MaxFPS: 30, VNCAuth: true}

	if got := loadYAML(t, text).VNC; got != want {
		t.Fatalf("viper loaded %+v, want %+v", got, want)
	}

	var conf Config
	if err := yaml.Unmarshal([]byte(text), &conf); err != nil {
		t.Fatal(err)
	}
	if conf.VNC != want {
		t.Fatalf("yaml loaded %+v, want %+v", conf.VNC, want)
	}
}
