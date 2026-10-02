package vm

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
	"time"

	"NanoKVM-Server/config"
	"NanoKVM-Server/utils/gpiocdev"
)

func useNoGpioChips(t *testing.T) {
	t.Helper()
	old := gpiocdev.ChipGlob
	gpiocdev.ChipGlob = filepath.Join(t.TempDir(), "gpiochip*")
	t.Cleanup(func() { gpiocdev.ChipGlob = old })
}

// On the vendor kernel the device tree names no lines. A sysfs path that is
// missing there must fail as it always has, with the file's own error, and not
// with something the character-device fallback made up.
func TestMissingSysfsLineWithoutANamedLineFailsAsBefore(t *testing.T) {
	useNoGpioChips(t)
	missing := filepath.Join(t.TempDir(), "gpio503", "value")
	conf := &config.GetInstance().Hardware
	old := *conf
	t.Cleanup(func() { *conf = old })
	conf.GPIOPower = missing
	conf.GPIOPowerLED = missing + ".led"

	if err := writeGpio(missing, time.Millisecond); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("writeGpio on a missing path = %v, want the not-exist error", err)
	}
	if _, err := readGpio(conf.GPIOPowerLED); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("readGpio on a missing path = %v, want the not-exist error", err)
	}
}

// An existing sysfs path never consults the character device.
func TestExistingSysfsPathIsNotReplaced(t *testing.T) {
	_, _, led := useButtons(t, "0\n")
	if line := cdevLine(led, gpiocdev.Input); line != nil {
		t.Fatalf("cdevLine(%q) = %v, want the sysfs path", led, line.Name())
	}
}
