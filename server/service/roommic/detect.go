// Package roommic captures the board's own microphone, the room the KVM sits
// in, and shares it with the viewers who switch it on.
//
// The microphone is invisible to whoever is in that room and has no privacy
// switch, so the feature is behind two locks: an administrator allows it for
// the device, and then each viewer turns it on for themselves. The capture
// runs only while at least one viewer has it on, and every viewer is told
// while it runs.
package roommic

import (
	"bufio"
	"bytes"
	"os"
	"strings"
)

// CardName is the ALSA id of the onboard codec under the mainline kernel.
//
// Slot A's vendor kernel names the same codec cv182xa_adc, and recording from
// it there hangs the board until the watchdog resets it. The feature must
// never open that card, so it looks for this id and nothing else.
const CardName = "sg2002onboard"

// Device is the codec's capture device, by card name rather than index: device
// 0 is the DAC and 1 the microphone. The hardware gives 48 kHz S16_LE, one
// channel only.
const Device = "hw:" + CardName + ",1"

// cardsPath is a variable so tests can point it at a fixture.
var cardsPath = "/proc/asound/cards"

// Available reports whether this kernel has the onboard microphone. It reads
// the card list each time; it costs one small read and the answer is the only
// thing between the feature and a kernel where recording hangs the board.
func Available() bool {
	cards, err := os.ReadFile(cardsPath)
	if err != nil {
		return false
	}

	return hasCard(cards, CardName)
}

// hasCard reports whether /proc/asound/cards lists a card with exactly this
// id. Each card's first line reads " 0 [sg2002onboard  ]: simple-card - ...";
// the id is the bracketed word, padded with spaces. Matching the id rather
// than searching the whole text keeps a driver or long name that merely
// contains the word from counting.
func hasCard(cards []byte, id string) bool {
	scanner := bufio.NewScanner(bytes.NewReader(cards))
	for scanner.Scan() {
		line := scanner.Text()

		open := strings.IndexByte(line, '[')
		end := strings.IndexByte(line, ']')
		if open < 0 || end < open {
			continue
		}

		// Only a card's first line starts with its index.
		index := strings.TrimSpace(line[:open])
		if index == "" || strings.Trim(index, "0123456789") != "" {
			continue
		}

		if strings.TrimSpace(line[open+1:end]) == id {
			return true
		}
	}

	return false
}
