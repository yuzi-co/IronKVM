package wol

import "errors"

// Wake sends the packet for a listed device, found by name or by MAC, or for a
// MAC that is not listed (which then joins the list, as it always did). The
// text is the tool's output when it fails.
func (s *Store) Wake(mac, name string) (Device, string, error) {
	var d Device
	var err error

	switch {
	case name != "":
		d, err = s.Find(name)
		if err != nil {
			return Device{}, "", err
		}
	default:
		if d.Mac, err = ParseMAC(mac); err != nil {
			return Device{}, "", err
		}
		if known, gerr := s.Get(d.Mac); gerr == nil {
			d = known
		} else if !errors.Is(gerr, ErrNotFound) {
			// The list could not be read: wake with the defaults rather than
			// refuse a wake the caller can still make by hand.
			d = Device{Mac: d.Mac}
		}
	}

	output, err := Send(d)
	if err != nil {
		return d, output, err
	}
	_ = s.Touch(d.Mac)

	return d, "", nil
}
