package config

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

type HWVersion int

const (
	HWVersionAlpha HWVersion = iota
	HWVersionBeta
	HWVersionPcie

	HWVersionFile = "/etc/kvm/hw"
)

var HWAlpha = Hardware{
	Version:      HWVersionAlpha,
	GPIOReset:    "/sys/class/gpio/gpio507/value",
	GPIOPower:    "/sys/class/gpio/gpio503/value",
	GPIOPowerLED: "/sys/class/gpio/gpio504/value",
	GPIOHDDLed:   "/sys/class/gpio/gpio505/value",
}

var HWBeta = Hardware{
	Version:      HWVersionBeta,
	GPIOReset:    "/sys/class/gpio/gpio505/value",
	GPIOPower:    "/sys/class/gpio/gpio503/value",
	GPIOPowerLED: "/sys/class/gpio/gpio504/value",
	GPIOHDDLed:   "",
}

var HWPcie = Hardware{
	Version:      HWVersionPcie,
	GPIOReset:    "/sys/class/gpio/gpio505/value",
	GPIOPower:    "/sys/class/gpio/gpio503/value",
	GPIOPowerLED: "/sys/class/gpio/gpio504/value",
	GPIOHDDLed:   "",
}

// GPIOLineName is the device tree name of the line whose sysfs path is p, or
// "" when p is none of this board's lines.
//
// The sysfs paths above exist on the vendor kernel only. The mainline kernel
// has no sysfs GPIO interface, and its board device tree names each control
// line in gpio-line-names instead, with the function names of the pin map
// files in ironkvm-dist (devices/<device>/pins). The server falls back to the
// line of that name when the sysfs path is absent; see service/vm/gpio.go.
// gates/check-pin-map.sh in ironkvm-dist maps the same four fields to the same
// four names.
func (h Hardware) GPIOLineName(p string) string {
	if p == "" {
		return ""
	}
	switch p {
	case h.GPIOPower:
		return "power"
	case h.GPIOReset:
		return "reset"
	case h.GPIOPowerLED:
		return "led-power"
	case h.GPIOHDDLed:
		return "led-hdd"
	}
	return ""
}

func (h HWVersion) String() string {
	switch h {
	case HWVersionAlpha:
		return "Alpha"
	case HWVersionBeta:
		return "Beta"
	case HWVersionPcie:
		return "PCIE"
	default:
		return "Unknown"
	}
}

func GetHwVersion() HWVersion {
	content, err := os.ReadFile(HWVersionFile)
	if err != nil {
		return HWVersionAlpha
	}

	version := strings.ReplaceAll(string(content), "\n", "")
	switch version {
	case "alpha":
		return HWVersionAlpha
	case "beta":
		return HWVersionBeta
	case "pcie":
		return HWVersionPcie
	default:
		return HWVersionAlpha
	}
}

func getHardware() (h Hardware) {
	version := GetHwVersion()

	switch version {
	case HWVersionAlpha:
		h = HWAlpha

	case HWVersionBeta:
		h = HWBeta

	case HWVersionPcie:
		h = HWPcie

	default:
		h = HWAlpha
		log.Errorf("Unsupported hardware version: %s", version)
	}

	return
}
