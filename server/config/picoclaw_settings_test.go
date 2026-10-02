package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// A server.yaml without the block sends the model 960 pixel wide screenshots
// at quality 60, as before the setting existed.
func TestAnOldConfigLoadsThePicoclawScreenshotDefaults(t *testing.T) {
	conf := loadYAML(t, "proto: http\nport:\n  http: 80\n  https: 443\n")
	want := Picoclaw{ScreenshotWidth: 960, ScreenshotQuality: 60}
	if got := conf.Picoclaw.WithDefaults(); got != want {
		t.Fatalf("an old server.yaml loads PicoClaw as %+v, want %+v", got, want)
	}
}

func TestPicoclawSettingsAreKeptUnderPicoclaw(t *testing.T) {
	text := "picoclaw:\n  screenshotWidth: 768\n  screenshotQuality: 50\n"
	want := Picoclaw{ScreenshotWidth: 768, ScreenshotQuality: 50}

	if got := loadYAML(t, text).Picoclaw; got != want {
		t.Fatalf("viper loaded %+v, want %+v", got, want)
	}

	var conf Config
	if err := yaml.Unmarshal([]byte(text), &conf); err != nil {
		t.Fatal(err)
	}
	if conf.Picoclaw != want {
		t.Fatalf("yaml loaded %+v, want %+v", conf.Picoclaw, want)
	}
}

func TestPicoclawScreenshotSettingsAreHeldWithinLimits(t *testing.T) {
	for _, test := range []struct {
		in, want Picoclaw
	}{
		{Picoclaw{ScreenshotWidth: 100, ScreenshotQuality: 1}, Picoclaw{ScreenshotWidth: 320, ScreenshotQuality: 10}},
		{Picoclaw{ScreenshotWidth: 4000, ScreenshotQuality: 500}, Picoclaw{ScreenshotWidth: 1920, ScreenshotQuality: 100}},
		{Picoclaw{ScreenshotWidth: -5, ScreenshotQuality: -5}, Picoclaw{ScreenshotWidth: 960, ScreenshotQuality: 60}},
		{Picoclaw{ScreenshotWidth: 640, ScreenshotQuality: 75}, Picoclaw{ScreenshotWidth: 640, ScreenshotQuality: 75}},
	} {
		if got := test.in.WithDefaults(); got != test.want {
			t.Fatalf("%+v.WithDefaults() = %+v, want %+v", test.in, got, test.want)
		}
	}
}
