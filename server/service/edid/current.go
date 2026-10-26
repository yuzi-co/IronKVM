package edid

// ReasonBridgeOnly says the EDID the board presents lives only in the HDMI
// bridge. The LT6911 keeps it in its own SPI flash; reading it back means
// driving the bridge over I2C, register writes included (Sipeed's
// lt6911_read_edid in kvm_vision.cpp), while the capture pipeline owns that
// bus. Neither slot's kernel exposes it as a file, and the vendor tool keeps
// no copy of what it wrote.
const ReasonBridgeOnly = "bridge-only"

// CurrentStatus answers what EDID the board presents. Hex and Decoded are
// filled only when Readable is true, which no board is today.
type CurrentStatus struct {
	Readable bool   `json:"readable"`
	Reason   string `json:"reason,omitempty"`
	// BuiltinPreset is the preset a host sees while the bridge's EDID slot
	// is blank, which it is unless someone ran nanokvm_update_edid.
	BuiltinPreset string `json:"builtinPreset,omitempty"`
	Hex           string `json:"hex,omitempty"`
	Decoded       *Info  `json:"decoded,omitempty"`
}

// Current reports the board's EDID. It never touches the hardware.
func Current() CurrentStatus {
	return CurrentStatus{Readable: false, Reason: ReasonBridgeOnly, BuiltinPreset: "lt6911uxc-builtin"}
}

// DecodeInput turns an uploaded EDID into bytes: the binary file as it is, or
// hex text (as edid-decode, xrandr --verbose or a sysfs dump through xxd -p
// print it) with any whitespace between the digits.
func DecodeInput(input []byte) ([]byte, error) {
	if len(input) >= len(header) && string(input[:len(header)]) == string(header) {
		return append([]byte(nil), input...), nil
	}
	digits := make([]byte, 0, len(input))
	for _, c := range input {
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
		case c >= '0' && c <= '9', c >= 'a' && c <= 'f', c >= 'A' && c <= 'F':
			digits = append(digits, c)
		default:
			return nil, ErrNotEDID
		}
	}
	if len(digits) == 0 {
		return nil, ErrTruncated
	}
	if len(digits)%2 != 0 {
		return nil, ErrNotEDID
	}
	out := make([]byte, len(digits)/2)
	for i := range out {
		out[i] = unhex(digits[2*i])<<4 | unhex(digits[2*i+1])
	}
	if len(out) < len(header) || string(out[:len(header)]) != string(header) {
		return nil, ErrNotEDID
	}
	return out, nil
}

func unhex(c byte) byte {
	switch {
	case c >= 'a':
		return c - 'a' + 10
	case c >= 'A':
		return c - 'A' + 10
	}
	return c - '0'
}
