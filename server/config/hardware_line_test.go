package config

import "testing"

// The names are the device tree's gpio-line-names on the mainline kernel and
// the function names of ironkvm-dist's pin map files. A wrong one drives the
// wrong line, so each variant's four fields are held to them.
func TestGPIOLineNameNamesEveryLine(t *testing.T) {
	for _, h := range []Hardware{HWAlpha, HWBeta, HWPcie} {
		want := map[string]string{
			h.GPIOPower:    "power",
			h.GPIOReset:    "reset",
			h.GPIOPowerLED: "led-power",
		}
		if h.GPIOHDDLed != "" {
			want[h.GPIOHDDLed] = "led-hdd"
		}
		for p, name := range want {
			if got := h.GPIOLineName(p); got != name {
				t.Errorf("%s: GPIOLineName(%q) = %q, want %q", h.Version, p, got, name)
			}
		}
		if got := h.GPIOLineName(""); got != "" {
			t.Errorf("%s: GPIOLineName(\"\") = %q, want none", h.Version, got)
		}
		if got := h.GPIOLineName("/sys/class/gpio/gpio1/value"); got != "" {
			t.Errorf("%s: an unknown path has the name %q", h.Version, got)
		}
	}
}

// Beta and pcie have no HDD LED. An empty path must not answer to the name of
// a line the board does not have.
func TestGPIOLineNameHasNoHDDLedWithoutOne(t *testing.T) {
	if got := HWBeta.GPIOLineName(HWBeta.GPIOHDDLed); got != "" {
		t.Fatalf("beta's empty HDD LED path has the name %q", got)
	}
}
