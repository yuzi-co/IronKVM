package vm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteSleepSettingStoresAValueTheFirmwareCanActOn(t *testing.T) {
	// The file is what kvm_system reads, so the clamp has to survive the
	// round trip -- not just live in a helper nobody calls.
	path := filepath.Join(t.TempDir(), "oled_sleep")

	if err := writeSleepSetting(path, 5); err != nil {
		t.Fatalf("expected the setting to be written: %s", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected the setting to be readable: %s", err)
	}

	if string(data) != "10" {
		t.Fatalf("stored %q, want \"10\"", data)
	}
}

func TestClampSleepSecondsRaisesDurationsTheFirmwareIgnores(t *testing.T) {
	// kvm_system treats anything below OLED_SLEEP_DELAY_MIN as "never sleep",
	// so asking for 5 seconds used to turn the screen saver off entirely --
	// the opposite of what was requested.
	for _, seconds := range []int{1, 5, 9} {
		if got := clampSleepSeconds(seconds); got != minSleepSeconds {
			t.Fatalf("clampSleepSeconds(%d) = %d, want %d", seconds, got, minSleepSeconds)
		}
	}
}

func TestClampSleepSecondsKeepsNeverAsNever(t *testing.T) {
	// Zero is the one deliberate way to say "keep the screen on".
	if got := clampSleepSeconds(0); got != 0 {
		t.Fatalf("clampSleepSeconds(0) = %d, want 0", got)
	}
}

func TestClampSleepSecondsKeepsSupportedDurations(t *testing.T) {
	// Every duration the web UI offers, plus the boundary itself.
	for _, seconds := range []int{10, 15, 30, 60, 180, 300, 600, 1800, 3600} {
		if got := clampSleepSeconds(seconds); got != seconds {
			t.Fatalf("clampSleepSeconds(%d) = %d, want it unchanged", seconds, got)
		}
	}
}

func TestClampSleepSecondsLowersDurationsTheFirmwareCannotHold(t *testing.T) {
	// The firmware parses the file into a uint16_t, so 65536 wraps to 0 and
	// silently means "never" again.
	for _, seconds := range []int{65536, 100000} {
		if got := clampSleepSeconds(seconds); got != maxSleepSeconds {
			t.Fatalf("clampSleepSeconds(%d) = %d, want %d", seconds, got, maxSleepSeconds)
		}
	}
}

func TestClampSleepSecondsTreatsNegativeAsNever(t *testing.T) {
	// A negative delay has no sensible reading; keep the screen on rather
	// than inventing a duration.
	if got := clampSleepSeconds(-5); got != 0 {
		t.Fatalf("clampSleepSeconds(-5) = %d, want 0", got)
	}
}

func TestClampBrightnessKeepsThePanelReadable(t *testing.T) {
	// A drive of zero is a legal SSD1306 command and a screen nobody can read.
	// This setting is reachable from a browser, so the floor is what stops a
	// slip of the hand from blanking the panel with no way to see how to put
	// it back.
	for _, level := range []int{-1, 0, 1, 0x0F} {
		if got := clampBrightness(level); got != minBrightness {
			t.Fatalf("clampBrightness(%d) = %d, want %d", level, got, minBrightness)
		}
	}
}

func TestClampBrightnessKeepsLevelsThePanelAccepts(t *testing.T) {
	// The command carries one byte, and both ends of the range are usable.
	for _, level := range []int{0x10, 0x60, 0xCF, 0xFF} {
		if got := clampBrightness(level); got != level {
			t.Fatalf("clampBrightness(%d) = %d, want it unchanged", level, got)
		}
	}
}

func TestClampBrightnessLowersLevelsThatDoNotFitTheCommand(t *testing.T) {
	// kvm_system writes the value as the argument of command 0x81, which is a
	// single byte. 256 would wrap to 0 and blank the panel.
	for _, level := range []int{256, 1000} {
		if got := clampBrightness(level); got != maxBrightness {
			t.Fatalf("clampBrightness(%d) = %d, want %d", level, got, maxBrightness)
		}
	}
}

func TestWriteBrightnessSettingStoresAValueTheFirmwareCanActOn(t *testing.T) {
	path := filepath.Join(t.TempDir(), "oled_contrast")

	if err := writeBrightnessSetting(path, 0); err != nil {
		t.Fatalf("expected the setting to be written: %s", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected the setting to be readable: %s", err)
	}

	if string(data) != "16" {
		t.Fatalf("stored %q, want \"16\"", data)
	}
}

func TestReadIntSettingReportsTheDefaultForAnAbsentFile(t *testing.T) {
	// A board that has never had the setting written is the common case, and
	// it means the firmware default rather than an error.
	got, err := readIntSetting(filepath.Join(t.TempDir(), "absent"), defaultBrightness)
	if err != nil {
		t.Fatalf("expected an absent file to be no error: %s", err)
	}

	if got != defaultBrightness {
		t.Fatalf("read %d, want %d", got, defaultBrightness)
	}
}

func TestReadIntSettingReportsAFileThatCannotBeParsed(t *testing.T) {
	// A file that exists and holds nonsense is a fault. Reporting the default
	// would hide it and leave the UI showing a value the device does not have.
	path := filepath.Join(t.TempDir(), "oled_contrast")
	if err := os.WriteFile(path, []byte("bright"), 0o644); err != nil {
		t.Fatalf("could not write the fixture: %s", err)
	}

	if _, err := readIntSetting(path, defaultBrightness); err == nil {
		t.Fatal("expected an unparseable file to be an error")
	}
}
