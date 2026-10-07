package roommic

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// gainControl is the codec's one mixer control, the capture PGA. It boots at
// 0, where the microphone is at the ADC's floor, so the server sets it before
// every capture rather than trusting whatever state it finds.
const gainControl = "Internal I2S Capture Volume"

// applyGain sets the PGA. A variable so tests can record the calls instead of
// running amixer.
var applyGain = setMixerGain

func setMixerGain(gain int) error {
	cmd := exec.Command("amixer", "-q", "-c", CardName, "cset", "name="+gainControl, strconv.Itoa(gain))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("amixer: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}
