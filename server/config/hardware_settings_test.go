package config

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// loadYAML decodes server.yaml content the way initialize does, through viper.
func loadYAML(t *testing.T, content string) Config {
	t.Helper()

	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewBufferString(content)); err != nil {
		t.Fatalf("read: %s", err)
	}
	var conf Config
	if err := v.Unmarshal(&conf); err != nil {
		t.Fatalf("unmarshal: %s", err)
	}
	return conf
}

// Most boards have no power LED header wired, so the setting is off until the
// owner says otherwise.
func TestThePowerLEDIsNotConnectedByDefault(t *testing.T) {
	if defaultConfig.HardwareSettings.PowerLED {
		t.Fatal("defaultConfig says the power LED is connected")
	}

	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if conf := loadYAML(t, string(data)); conf.HardwareSettings.PowerLED {
		t.Fatal("the default configuration loads with the power LED connected")
	}
}

// A server.yaml written before the setting existed has no hardware block.
func TestAnOldConfigWithoutTheKeyLoadsAsNotConnected(t *testing.T) {
	old := "proto: http\nport:\n  http: 80\n  https: 443\nauthentication: enable\n"

	if conf := loadYAML(t, old); conf.HardwareSettings.PowerLED {
		t.Fatal("an old server.yaml loads with the power LED connected")
	}

	var conf Config
	if err := yaml.Unmarshal([]byte(old), &conf); err != nil {
		t.Fatal(err)
	}
	if conf.HardwareSettings.PowerLED {
		t.Fatal("config.Read of an old server.yaml says the power LED is connected")
	}
}

// The key is hardware.powerLed, both for viper at start and for the
// Read/Write pair that saves settings.
func TestThePowerLEDSettingIsKeptUnderHardware(t *testing.T) {
	conf := loadYAML(t, "hardware:\n  powerLed: true\n")
	if !conf.HardwareSettings.PowerLED {
		t.Fatal("hardware.powerLed: true did not load")
	}

	data, err := yaml.Marshal(&conf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "hardware:\n    powerLed: true\n") {
		t.Fatalf("saved as:\n%s", data)
	}

	var back Config
	if err := yaml.Unmarshal(data, &back); err != nil || !back.HardwareSettings.PowerLED {
		t.Fatalf("round trip lost the setting: %v", err)
	}
}

// The pins are derived from the board version at start and must stay that
// way: a hardware block in server.yaml does not reach them.
func TestAHardwareBlockLeavesThePinsAlone(t *testing.T) {
	conf := loadYAML(t, "hardware:\n  powerLed: true\n  gpioPower: /etc/shadow\n")
	if conf.Hardware.GPIOPower != "" {
		t.Fatalf("GPIOPower loaded as %q from server.yaml", conf.Hardware.GPIOPower)
	}
}
