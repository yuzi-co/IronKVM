package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// RTSP and SRT each open a port to the picture, so nothing turns them on but
// the owner (ironkvm-dist#142, ironkvm-dist#143).
func TestRTSPAndSRTAreOffByDefault(t *testing.T) {
	if defaultConfig.RTSP.Enabled || defaultConfig.SRT.Enabled {
		t.Fatal("defaultConfig turns RTSP or SRT on")
	}

	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	conf := loadYAML(t, string(data))
	if conf.RTSP.Enabled || conf.SRT.Enabled {
		t.Fatal("the default configuration loads with RTSP or SRT on")
	}
}

// A server.yaml without the blocks loads with both off and the defaults.
func TestAnOldConfigLoadsWithRTSPAndSRTOff(t *testing.T) {
	conf := loadYAML(t, "proto: http\nport:\n  http: 80\n  https: 443\n")

	if got, want := conf.RTSP.WithDefaults(), (RTSP{Port: DefaultRTSPPort, UDPPort: DefaultRTSPUDPPort}); got != want {
		t.Fatalf("an old server.yaml loads RTSP as %+v, want %+v", got, want)
	}
	if got, want := conf.SRT.WithDefaults(), (SRT{Port: DefaultSRTPort, LatencyMs: DefaultSRTLatencyMs}); got != want {
		t.Fatalf("an old server.yaml loads SRT as %+v, want %+v", got, want)
	}
}

// The settings are kept under rtsp and srt, for viper at start and for the
// Read/Write pair that saves settings.
func TestRTSPAndSRTSettingsAreKeptUnderTheirNames(t *testing.T) {
	text := "rtsp:\n  enabled: true\n  port: 8555\n  udp: true\n  udpPort: 9000\n" +
		"srt:\n  enabled: true\n  port: 9001\n  latencyMs: 300\n"
	wantRTSP := RTSP{Enabled: true, Port: 8555, UDP: true, UDPPort: 9000}
	wantSRT := SRT{Enabled: true, Port: 9001, LatencyMs: 300}

	loaded := loadYAML(t, text)
	if loaded.RTSP != wantRTSP || loaded.SRT != wantSRT {
		t.Fatalf("viper loaded %+v and %+v, want %+v and %+v", loaded.RTSP, loaded.SRT, wantRTSP, wantSRT)
	}

	var conf Config
	if err := yaml.Unmarshal([]byte(text), &conf); err != nil {
		t.Fatal(err)
	}
	if conf.RTSP != wantRTSP || conf.SRT != wantSRT {
		t.Fatalf("yaml loaded %+v and %+v, want %+v and %+v", conf.RTSP, conf.SRT, wantRTSP, wantSRT)
	}
}
