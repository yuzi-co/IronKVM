package ws

import (
	"fmt"

	"NanoKVM-Server/service/hid"
)

// touchContactLen is one contact in a touch message: flags, the contact ID,
// a 16-bit X and a 16-bit Y, little-endian like the mouse reports.
const touchContactLen = 6

// touchFlagDown is Tip Switch: the finger is on the screen.
const touchFlagDown = 0x01

// decodeTouchFrame reads the payload of a TouchEvent message:
//
//	count { flags id xLo xHi yLo yHi } x count
//
// The frame lists every contact that is down, and once more with the flag
// clear each one that has just lifted. Unknown flag bits are refused rather
// than ignored, so a later client that means something by them is not
// silently misread.
func decodeTouchFrame(payload []byte) ([]hid.TouchContact, error) {
	if len(payload) == 0 {
		return nil, fmt.Errorf("empty touch frame")
	}

	count := int(payload[0])
	if len(payload) != 1+count*touchContactLen {
		return nil, fmt.Errorf("touch frame of %d bytes for %d contacts", len(payload), count)
	}

	contacts := make([]hid.TouchContact, 0, count)
	for i := range count {
		raw := payload[1+i*touchContactLen : 1+(i+1)*touchContactLen]
		if raw[0]&^touchFlagDown != 0 {
			return nil, fmt.Errorf("touch contact with unknown flags 0x%02x", raw[0])
		}
		contacts = append(contacts, hid.TouchContact{
			ID:   raw[1],
			Down: raw[0]&touchFlagDown != 0,
			X:    uint16(raw[2]) | uint16(raw[3])<<8,
			Y:    uint16(raw[4]) | uint16(raw[5])<<8,
		})
	}

	if err := hid.ValidateTouchFrame(contacts); err != nil {
		return nil, err
	}
	return contacts, nil
}
