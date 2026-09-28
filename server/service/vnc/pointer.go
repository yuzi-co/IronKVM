package vnc

import (
	"NanoKVM-Server/service/hid"
)

// VNC button mask bits. Buttons 4 and 5 are the wheel, one step each time they
// go down.
const (
	vncButtonLeft   byte = 1 << 0
	vncButtonMiddle byte = 1 << 1
	vncButtonRight  byte = 1 << 2
	vncWheelUp      byte = 1 << 3
	vncWheelDown    byte = 1 << 4
)

// HID button bits of the absolute pointer report.
const (
	hidButtonLeft   byte = 1 << 0
	hidButtonRight  byte = 1 << 1
	hidButtonMiddle byte = 1 << 2
)

// absoluteCoordinate scales a framebuffer position to the pointer's range the
// same way the web UI does: 0x7fff times the fraction of the screen, plus one.
func absoluteCoordinate(position, size int) uint16 {
	if size <= 1 || position <= 0 {
		return 1
	}
	if position >= size-1 {
		return 0x7fff + 1
	}
	return uint16(0x7fff*position/(size-1)) + 1
}

// pointer turns VNC pointer events into absolute pointer reports.
type pointer struct {
	mask byte
}

// event applies one PointerEvent on a framebuffer of the given size and
// returns the reports to send. A wheel step goes in its own report, sent on
// the press of the wheel button only.
func (p *pointer) event(mask byte, x, y, width, height int) [][]byte {
	pressed := mask &^ p.mask
	p.mask = mask

	var buttons byte
	if mask&vncButtonLeft != 0 {
		buttons |= hidButtonLeft
	}
	if mask&vncButtonMiddle != 0 {
		buttons |= hidButtonMiddle
	}
	if mask&vncButtonRight != 0 {
		buttons |= hidButtonRight
	}

	ax := absoluteCoordinate(x, width)
	ay := absoluteCoordinate(y, height)

	reports := [][]byte{absoluteReport(buttons, ax, ay, 0)}
	// Positive is up on the absolute pointer, as the web UI sends it.
	if pressed&vncWheelUp != 0 {
		reports = append(reports, absoluteReport(buttons, ax, ay, 1))
	}
	if pressed&vncWheelDown != 0 {
		reports = append(reports, absoluteReport(buttons, ax, ay, -1))
	}
	return reports
}

func absoluteReport(buttons byte, x, y uint16, wheel int8) []byte {
	report := make([]byte, hid.AbsoluteMouseReportLen)
	report[0] = buttons
	report[1] = byte(x)
	report[2] = byte(x >> 8)
	report[3] = byte(y)
	report[4] = byte(y >> 8)
	report[5] = byte(wheel)
	return report
}
