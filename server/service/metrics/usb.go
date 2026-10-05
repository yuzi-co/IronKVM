package metrics

import (
	"slices"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/hid"
)

// udcStates are the strings the kernel's usb_state_string gives the UDC state
// attribute, in the kernel's order.
var udcStates = []string{
	"not attached", "attached", "powered", "reconnecting", "unauthenticated",
	"default", "addressed", "configured", "suspended",
}

// The USB and HID readers, variables so tests can stub them. Status and
// WriteErrors take none of the HID locks, so a scrape never delays a key.
var (
	usbLinkState   = hid.USBLinkState
	hidStatus      = func() []proto.HidDeviceStatus { return hid.GetHid().Status() }
	hidWriteErrors = func() []hid.EndpointWriteErrors { return hid.GetHid().WriteErrors() }
	usbRecoveries  = hid.USBRecoveries
	usbReenums     = hid.USBReenumerations
)

const (
	helpUDCState   = "The USB device controller's state, 1 for the current one."
	helpHidState   = "Each HID endpoint's state, 1 for the current one."
	helpHidErrors  = "HID writes that stalled or found the gadget detached, by endpoint."
	helpRecoveries = "Recoveries the USB watchdog started, by action."
	helpReenums    = "Times the host enumerated the gadget again after it had been configured, not counting the server's own rebinds."
)

func collectUSB(w *Writer) {
	if state, err := usbLinkState(); err == nil {
		writeEnum(w, "ironkvm_usb_udc_state", helpUDCState, udcStates, state)
	}

	states := hid.EndpointStates()
	for _, status := range hidStatus() {
		writeEnum(w, "ironkvm_hid_endpoint_state", helpHidState, states, status.State, L("endpoint", status.Name))
	}

	for _, errs := range hidWriteErrors() {
		w.Counter("ironkvm_hid_write_errors_total", helpHidErrors, float64(errs.Stalled),
			L("endpoint", errs.Name), L("kind", "stalled"))
		w.Counter("ironkvm_hid_write_errors_total", helpHidErrors, float64(errs.Detached),
			L("endpoint", errs.Name), L("kind", "detached"))
	}

	rebinds, rebuilds := usbRecoveries()
	w.Counter("ironkvm_usb_recoveries_total", helpRecoveries, float64(rebinds), L("action", "rebind"))
	w.Counter("ironkvm_usb_recoveries_total", helpRecoveries, float64(rebuilds), L("action", "rebuild"))

	w.Counter("ironkvm_usb_reenumerations_total", helpReenums, float64(usbReenums()))
}

// writeEnum writes one sample per known state, 1 for current and 0 for the
// rest, so a state that ends leaves a 0 rather than a stale series. A current
// state the list does not name is added at the end with 1.
func writeEnum(w *Writer, name, help string, states []string, current string, labels ...Label) {
	for _, state := range states {
		value := 0.0
		if state == current {
			value = 1
		}
		w.Gauge(name, help, value, append(labels, L("state", state))...)
	}

	if !slices.Contains(states, current) {
		w.Gauge(name, help, 1, append(labels, L("state", current))...)
	}
}
