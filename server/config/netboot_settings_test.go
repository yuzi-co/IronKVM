package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Proxy DHCP on the LAN answers every PXE client there, so nothing turns
// network boot on but the owner.
func TestNetworkBootIsOffByDefault(t *testing.T) {
	if defaultConfig.NetBoot.USB || defaultConfig.NetBoot.LAN {
		t.Fatal("defaultConfig turns network boot on")
	}

	old := "proto: http\nport:\n  http: 80\n  https: 443\nauthentication: enable\n"
	if conf := loadYAML(t, old); conf.NetBoot.USB || conf.NetBoot.LAN {
		t.Fatal("a server.yaml without the block loads with network boot on")
	}
}

// The keys live under netboot, both for viper at start and for the Read/Write
// pair that saves settings.
func TestTheNetworkBootSettingsAreKeptUnderNetboot(t *testing.T) {
	want := NetBoot{USB: true, LAN: true}

	conf := loadYAML(t, "netboot:\n  usb: true\n  lan: true\n")
	if conf.NetBoot != want {
		t.Fatalf("viper loaded %+v", conf.NetBoot)
	}

	data, err := yaml.Marshal(&conf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "netboot:\n    usb: true\n    lan: true\n") {
		t.Fatalf("saved as:\n%s", data)
	}

	var back Config
	if err := yaml.Unmarshal(data, &back); err != nil || back.NetBoot != want {
		t.Fatalf("round trip: %+v, %v", back.NetBoot, err)
	}
}
