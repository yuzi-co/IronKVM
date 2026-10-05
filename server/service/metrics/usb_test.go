package metrics

import (
	"errors"
	"strings"
	"testing"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/hid"
)

// useUSBSources stubs every USB and HID reader with a healthy keyboard, a
// stalled absolute pointer and a configured link.
func useUSBSources(t *testing.T) {
	t.Helper()

	setVar(t, &usbLinkState, func() (string, error) { return "configured", nil })
	setVar(t, &hidStatus, func() []proto.HidDeviceStatus {
		return []proto.HidDeviceStatus{
			{Name: "keyboard", State: "accepting"},
			{Name: "mouse-absolute", State: "stalled"},
		}
	})
	setVar(t, &hidWriteErrors, func() []hid.EndpointWriteErrors {
		return []hid.EndpointWriteErrors{
			{Name: "keyboard"},
			{Name: "mouse-absolute", Stalled: 4, Detached: 1},
		}
	})
	setVar(t, &usbRecoveries, func() (uint64, uint64) { return 2, 1 })
	setVar(t, &usbReenums, func() uint64 { return 5 })
}

func TestUSBWritesEveryFamily(t *testing.T) {
	useUSBSources(t)

	want := `# HELP ironkvm_usb_udc_state The USB device controller's state, 1 for the current one.
# TYPE ironkvm_usb_udc_state gauge
ironkvm_usb_udc_state{state="not attached"} 0
ironkvm_usb_udc_state{state="attached"} 0
ironkvm_usb_udc_state{state="powered"} 0
ironkvm_usb_udc_state{state="reconnecting"} 0
ironkvm_usb_udc_state{state="unauthenticated"} 0
ironkvm_usb_udc_state{state="default"} 0
ironkvm_usb_udc_state{state="addressed"} 0
ironkvm_usb_udc_state{state="configured"} 1
ironkvm_usb_udc_state{state="suspended"} 0
# HELP ironkvm_hid_endpoint_state Each HID endpoint's state, 1 for the current one.
# TYPE ironkvm_hid_endpoint_state gauge
ironkvm_hid_endpoint_state{endpoint="keyboard",state="unknown"} 0
ironkvm_hid_endpoint_state{endpoint="keyboard",state="accepting"} 1
ironkvm_hid_endpoint_state{endpoint="keyboard",state="stalled"} 0
ironkvm_hid_endpoint_state{endpoint="keyboard",state="detached"} 0
ironkvm_hid_endpoint_state{endpoint="keyboard",state="error"} 0
ironkvm_hid_endpoint_state{endpoint="mouse-absolute",state="unknown"} 0
ironkvm_hid_endpoint_state{endpoint="mouse-absolute",state="accepting"} 0
ironkvm_hid_endpoint_state{endpoint="mouse-absolute",state="stalled"} 1
ironkvm_hid_endpoint_state{endpoint="mouse-absolute",state="detached"} 0
ironkvm_hid_endpoint_state{endpoint="mouse-absolute",state="error"} 0
# HELP ironkvm_hid_write_errors_total HID writes that stalled or found the gadget detached, by endpoint.
# TYPE ironkvm_hid_write_errors_total counter
ironkvm_hid_write_errors_total{endpoint="keyboard",kind="stalled"} 0
ironkvm_hid_write_errors_total{endpoint="keyboard",kind="detached"} 0
ironkvm_hid_write_errors_total{endpoint="mouse-absolute",kind="stalled"} 4
ironkvm_hid_write_errors_total{endpoint="mouse-absolute",kind="detached"} 1
# HELP ironkvm_usb_recoveries_total Recoveries the USB watchdog started, by action.
# TYPE ironkvm_usb_recoveries_total counter
ironkvm_usb_recoveries_total{action="rebind"} 2
ironkvm_usb_recoveries_total{action="rebuild"} 1
# HELP ironkvm_usb_reenumerations_total Times the host enumerated the gadget again after it had been configured, not counting the server's own rebinds.
# TYPE ironkvm_usb_reenumerations_total counter
ironkvm_usb_reenumerations_total 5
`
	assertText(t, render(t, collectUSB), want)
}

// No controller to read is a kernel without one. The UDC family goes, and the
// HID families, which do not depend on it, stay.
func TestUSBLeavesOutTheUDCStateWhenNoControllerReads(t *testing.T) {
	useUSBSources(t)
	setVar(t, &usbLinkState, func() (string, error) { return "", errors.New("no controller") })

	got := render(t, collectUSB)
	if strings.Contains(got, "ironkvm_usb_udc_state") {
		t.Fatalf("the UDC family was written without a controller:\n%s", got)
	}
	if !strings.Contains(got, `ironkvm_hid_endpoint_state{endpoint="keyboard",state="accepting"} 1`) {
		t.Fatalf("the HID families went with it:\n%s", got)
	}
}

// A kernel that adds a state must not make the current state vanish.
func TestUSBWritesAnUnknownUDCStateAsCurrent(t *testing.T) {
	useUSBSources(t)
	setVar(t, &usbLinkState, func() (string, error) { return "a new state", nil })

	got := render(t, collectUSB)
	if !strings.Contains(got, `ironkvm_usb_udc_state{state="a new state"} 1`+"\n") {
		t.Fatalf("the unknown state is missing:\n%s", got)
	}
	if strings.Contains(got, `ironkvm_usb_udc_state{state="configured"} 1`) {
		t.Fatalf("a known state was marked current:\n%s", got)
	}
}
