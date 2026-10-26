// Package hostdevices is what an AI agent may do to the host's hardware beyond
// the keyboard and mouse: read the power LED, press the ATX power and reset
// buttons, list the virtual drives, and insert or eject an image. External MCP
// clients (service/mcp) and the on-board agent (service/picoclaw) both call
// it, so the two check their arguments, refuse a second press and keep URL
// credentials out of the audit log in the same way. Each caller gates the
// calls by its own capability tiers and records them in the audit log.
package hostdevices

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"NanoKVM-Server/service/hostpower"
)

const (
	// The hold limits are vm.PressButton's: it refuses a press longer than
	// ten seconds.
	MinHoldMS = 100
	MaxHoldMS = 10000

	PressShort = "short"
	PressLong  = "long"
)

// MediaDrive is one virtual drive and what it holds.
type MediaDrive struct {
	ID       string `json:"id"`
	File     string `json:"file"`
	ReadOnly bool   `json:"readOnly"`
}

// Devices are the board's power button, power LED and virtual media, which
// the router supplies. A nil function leaves out the tools that need it.
type Devices struct {
	// PressButton holds "power" or "reset" for d. It is vm.PressButton, which
	// the web UI, Redfish and IPMI press through: one press at a time.
	PressButton func(kind string, d time.Duration) error
	// PowerLEDConnected reports whether the owner has said the power LED
	// header is wired. When it is not, the LED is never read.
	PowerLEDConnected func() bool
	// PowerLED reports whether the host's power LED is lit.
	PowerLED func() (bool, error)
	// ListMedia returns the drives and the image files that can be loaded.
	ListMedia func() (drives []MediaDrive, images []string, err error)
	// InsertMedia loads an image into a drive ("cdrom" or "disk").
	InsertMedia func(drive, file string, readOnly bool) error
	// EjectMedia empties a drive.
	EjectMedia func(drive string) error
	// MountURL serves an image straight from an http(s) URL in a drive,
	// read-only, on a board whose kernel has NBD. Nil leaves URLs out.
	MountURL func(drive, url string) error
}

// CanReadPower reports whether the power state can be asked for.
func (d Devices) CanReadPower() bool { return d.PowerLEDConnected != nil && d.PowerLED != nil }

// CanPress reports whether a button can be pressed.
func (d Devices) CanPress() bool { return d.PressButton != nil }

// CanListMedia reports whether the drives can be listed.
func (d Devices) CanListMedia() bool { return d.ListMedia != nil }

// CanChangeMedia reports whether images can be inserted and ejected.
func (d Devices) CanChangeMedia() bool { return d.InsertMedia != nil && d.EjectMedia != nil }

// PowerStatus is the host's power state as its LED shows it.
type PowerStatus struct {
	// State is on, off or unknown.
	State        string `json:"state"`
	LEDConnected bool   `json:"ledConnected"`
	Message      string `json:"message,omitempty"`
}

// ReadPower reads the power LED.
func (d Devices) ReadPower() PowerStatus {
	power := hostpower.New(hostpower.Deps{
		PowerLEDConnected: d.PowerLEDConnected,
		PowerLED:          d.PowerLED,
	})
	connected, on, err := power.LED()
	switch {
	case !connected:
		return PowerStatus{State: hostpower.Unknown.String(), Message: "the power LED is not connected to the NanoKVM"}
	case err != nil:
		return PowerStatus{State: hostpower.Unknown.String(), LEDConnected: true, Message: "the power LED could not be read: " + err.Error()}
	case on:
		return PowerStatus{State: hostpower.On.String(), LEDConnected: true}
	}
	return PowerStatus{State: hostpower.Off.String(), LEDConnected: true}
}

// Press is a call to press a button.
type Press struct {
	Button string
	// Press is short or long; empty is short.
	Press string
	// HoldMS, when set, is the exact hold time instead of Press.
	HoldMS *int
}

// ErrPressBusy refuses a press made while another is held.
var ErrPressBusy = errors.New("another button press is still in progress")

// Hold is how long a press holds its button.
func Hold(in Press) (time.Duration, error) {
	press := strings.ToLower(strings.TrimSpace(in.Press))
	if press != "" && in.HoldMS != nil {
		return 0, fmt.Errorf("give press or holdMs, not both")
	}
	if in.HoldMS != nil {
		return time.Duration(clamp(*in.HoldMS, MinHoldMS, MaxHoldMS)) * time.Millisecond, nil
	}
	switch press {
	case "", PressShort:
		return hostpower.ShortPress, nil
	case PressLong:
		return hostpower.LongPress, nil
	}
	return 0, fmt.Errorf("press must be short or long")
}

// Presser presses the buttons for one caller. vm.PressButton serializes
// presses from every caller, so a second one waits for the first and then
// presses. Two presses an agent sent at once are a mistake, since a second
// short press undoes the first, so a Presser refuses a press while one of its
// own is held.
type Presser struct {
	devices  Devices
	pressing atomic.Bool
}

// NewPresser returns a Presser for the devices.
func NewPresser(devices Devices) *Presser { return &Presser{devices: devices} }

// Press checks the call and presses.
func (p *Presser) Press(in Press) error {
	button := strings.ToLower(strings.TrimSpace(in.Button))
	if button != hostpower.ButtonPower && button != hostpower.ButtonReset {
		return fmt.Errorf("button must be power or reset")
	}
	hold, err := Hold(in)
	if err != nil {
		return err
	}
	if !p.pressing.CompareAndSwap(false, true) {
		return ErrPressBusy
	}
	defer p.pressing.Store(false)
	return p.devices.PressButton(button, hold)
}

// MediaState is the drives and, for a listing, the images.
type MediaState struct {
	OK     bool         `json:"ok"`
	Drives []MediaDrive `json:"drives,omitempty"`
	Images []string     `json:"images,omitempty"`
}

// List returns the drives and the images that can be inserted.
func (d Devices) List() (MediaState, error) {
	drives, images, err := d.ListMedia()
	if err != nil {
		return MediaState{}, err
	}
	return MediaState{OK: true, Drives: drives, Images: images}, nil
}

// Change is a call to insert or eject.
type Change struct {
	Action   string
	Drive    string
	File     string
	URL      string
	ReadOnly *bool
}

// ChangeMedia inserts or ejects. listTool names the caller's listing tool,
// which an unknown action's error points to. The result holds the drives as
// they now are, so the caller sees what the host has without another call.
func (d Devices) ChangeMedia(in Change, listTool string) (MediaState, error) {
	switch strings.ToLower(strings.TrimSpace(in.Action)) {
	case "insert":
		drive := normalizedDrive(in.Drive)
		if in.URL != "" {
			return d.insertURL(drive, in)
		}
		if drive == "" || in.File == "" {
			return MediaState{}, fmt.Errorf("insert needs a drive (cdrom or disk) and a file")
		}
		readOnly := in.ReadOnly == nil || *in.ReadOnly
		if drive == "cdrom" && !readOnly {
			return MediaState{}, fmt.Errorf("the cdrom drive is read-only; insert into disk for a writable image")
		}
		if err := d.InsertMedia(drive, in.File, readOnly); err != nil {
			return MediaState{}, err
		}
		return d.after(), nil
	case "eject":
		drive := normalizedDrive(in.Drive)
		if drive == "" {
			return MediaState{}, fmt.Errorf("eject needs a drive (cdrom or disk)")
		}
		if err := d.EjectMedia(drive); err != nil {
			return MediaState{}, err
		}
		return d.after(), nil
	default:
		return MediaState{}, fmt.Errorf("action must be insert or eject; %s lists the drives and images", listTool)
	}
}

func (d Devices) insertURL(drive string, in Change) (MediaState, error) {
	switch {
	case d.MountURL == nil:
		return MediaState{}, fmt.Errorf("this board cannot serve an image from a URL")
	case drive == "":
		return MediaState{}, fmt.Errorf("insert needs a drive (cdrom or disk)")
	case in.File != "":
		return MediaState{}, fmt.Errorf("give a file or a url, not both")
	case in.ReadOnly != nil && !*in.ReadOnly:
		return MediaState{}, fmt.Errorf("an image from a URL is always read-only")
	}
	if err := d.MountURL(drive, in.URL); err != nil {
		return MediaState{}, redactedError{err: err, text: RedactURLIn(err.Error(), in.URL)}
	}
	return d.after(), nil
}

func (d Devices) after() MediaState {
	result := MediaState{OK: true}
	if d.ListMedia != nil {
		if drives, _, err := d.ListMedia(); err == nil {
			result.Drives = drives
		}
	}
	return result
}

func normalizedDrive(drive string) string {
	switch d := strings.ToLower(strings.TrimSpace(drive)); d {
	case "cdrom", "disk":
		return d
	}
	return ""
}

func clamp(value, minValue, maxValue int) int {
	return min(max(value, minValue), maxValue)
}

// redactedError is a mount error whose text has the URL's credentials taken
// out. It still unwraps to the original.
type redactedError struct {
	err  error
	text string
}

func (e redactedError) Error() string { return e.text }
func (e redactedError) Unwrap() error { return e.err }

// RedactURLIn takes the credentials of raw out of text: the URL becomes its
// scheme, host and path, and its password and query, wherever else they
// appear, are replaced.
func RedactURLIn(text, raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return strings.ReplaceAll(text, raw, "(url)")
	}
	text = strings.ReplaceAll(text, raw, (&url.URL{Scheme: u.Scheme, Host: u.Host, Path: u.Path}).String())
	if u.User != nil {
		if password, ok := u.User.Password(); ok && password != "" {
			text = strings.ReplaceAll(text, password, "xxxxx")
		}
		if name := u.User.Username(); name != "" {
			text = strings.ReplaceAll(text, name+"@", "")
		}
	}
	if u.RawQuery != "" {
		text = strings.ReplaceAll(text, u.RawQuery, "xxxxx")
	}
	return text
}

// DescribePress reduces a press to what the audit log keeps.
func DescribePress(in Press) (target, detail string) {
	button := strings.ToLower(strings.TrimSpace(in.Button))
	hold, err := Hold(in)
	if err != nil {
		return button, ""
	}
	detail = fmt.Sprintf("hold=%dms", hold.Milliseconds())
	if in.HoldMS == nil {
		press := strings.ToLower(strings.TrimSpace(in.Press))
		if press == "" {
			press = PressShort
		}
		detail = press + " " + detail
	}
	return button, detail
}

// DescribeChange reduces a media change to what the audit log keeps. A URL
// may carry credentials: the log gets its host only.
func DescribeChange(in Change) (target, detail string) {
	action := strings.ToLower(strings.TrimSpace(in.Action))
	detail = strings.TrimSpace(action + " " + normalizedDrive(in.Drive))
	if action == "insert" {
		if in.ReadOnly != nil && !*in.ReadOnly {
			detail += " rw"
		} else {
			detail += " ro"
		}
	}
	if in.URL != "" {
		host := ""
		if u, err := url.Parse(in.URL); err == nil {
			host = u.Host
		}
		return host, detail + " url"
	}
	return in.File, detail
}
