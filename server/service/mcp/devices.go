package mcpservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	protocol "github.com/modelcontextprotocol/go-sdk/mcp"

	"NanoKVM-Server/service/hostpower"
)

const (
	powerToolName       = "cube_power"
	powerStatusToolName = "cube_power_status"
	mediaToolName       = "cube_media"
	mediaListToolName   = "cube_media_list"

	// The hold limits are vm.PressButton's: it refuses a press longer than
	// ten seconds.
	minPowerHoldMS = 100
	maxPowerHoldMS = 10000

	pressShort = "short"
	pressLong  = "long"
)

// MediaDrive is one virtual drive and what it holds.
type MediaDrive struct {
	ID       string `json:"id"`
	File     string `json:"file"`
	ReadOnly bool   `json:"readOnly"`
}

// Devices are the board's power button, power LED and virtual media, which
// the router supplies. A nil function leaves its tool out. Reading them is in
// the observe tier; pressing a button and changing a drive are in the power
// and media tiers, which are off unless an administrator turns them on.
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
}

type PowerParams struct {
	Button string `json:"button" jsonschema:"power or reset"`
	Press  string `json:"press,omitempty" jsonschema:"short or long"`
	HoldMS *int   `json:"holdMs,omitempty" jsonschema:"hold time in milliseconds"`
}

type PowerStatusParams struct{}

type PowerStatusResult struct {
	// State is on, off or unknown.
	State        string `json:"state"`
	LEDConnected bool   `json:"ledConnected"`
	Message      string `json:"message,omitempty"`
}

type MediaListParams struct{}

type MediaParams struct {
	Action   string `json:"action" jsonschema:"insert or eject"`
	Drive    string `json:"drive,omitempty" jsonschema:"cdrom or disk"`
	File     string `json:"file,omitempty" jsonschema:"image path as listed"`
	ReadOnly *bool  `json:"readOnly,omitempty" jsonschema:"disk only: false lets the host write to the image"`
}

type MediaResult struct {
	OK     bool         `json:"ok"`
	Drives []MediaDrive `json:"drives,omitempty"`
	Images []string     `json:"images,omitempty"`
}

const powerDescription = "Press the ATX power or reset button of the remote host. " +
	"A short power press asks the operating system to shut down, or switches on a host that is off; " +
	"a long power press holds the button for 5 seconds and forces the host off at once, losing unsaved work. " +
	"A short reset press restarts the host immediately, also losing unsaved work. " +
	"Check cube_power_status first: the same short press switches a host on or off depending on its state. " +
	"Only one press runs at a time; a call made while another press is held is refused."

const powerStatusDescription = "Report whether the remote host is powered on, by its power LED: " +
	"state is on, off or unknown. Unknown means the LED header is not wired to the NanoKVM or could not be read, " +
	"so the power state cannot be told; use a screenshot instead and do not assume the host is off."

const mediaListDescription = "List the virtual drives the remote host sees (cdrom and disk), " +
	"the image each one holds (empty when none) and whether it is read-only, " +
	"and the image files on the NanoKVM that can be inserted."

const mediaDescription = "Insert an image into a virtual drive of the remote host, or eject it. " +
	"Use a file path exactly as cube_media_list returns it. Inserting replaces any image already in the drive. " +
	"The cdrom drive is always read-only. The disk is read-only unless readOnly is false; " +
	"a writable disk lets the host change or format the image, so only ask for it when the user wants that. " +
	"Ejecting an image the host is using can lose its writes."

func registerDeviceTools(server *protocol.Server, devices Devices) {
	if devices.PowerLEDConnected != nil && devices.PowerLED != nil {
		protocol.AddTool(server, &protocol.Tool{
			Name:        powerStatusToolName,
			Description: powerStatusDescription,
			InputSchema: objectSchema(map[string]*jsonschema.Schema{}),
		}, guard(powerStatusToolName, describeNothing[PowerStatusParams], powerStatusHandler(devices)))
	}
	if devices.PressButton != nil {
		protocol.AddTool(server, &protocol.Tool{
			Name:        powerToolName,
			Description: powerDescription,
			InputSchema: powerSchema(),
		}, guard(powerToolName, describePower, powerHandler(devices)))
	}
	if devices.ListMedia != nil {
		protocol.AddTool(server, &protocol.Tool{
			Name:        mediaListToolName,
			Description: mediaListDescription,
			InputSchema: objectSchema(map[string]*jsonschema.Schema{}),
		}, guard(mediaListToolName, describeNothing[MediaListParams], mediaListHandler(devices)))
	}
	if devices.InsertMedia != nil && devices.EjectMedia != nil {
		protocol.AddTool(server, &protocol.Tool{
			Name:        mediaToolName,
			Description: mediaDescription,
			InputSchema: mediaSchema(),
		}, guard(mediaToolName, describeMedia, mediaHandler(devices)))
	}
}

func powerStatusHandler(devices Devices) protocol.ToolHandlerFor[PowerStatusParams, PowerStatusResult] {
	power := hostpower.New(hostpower.Deps{
		PowerLEDConnected: devices.PowerLEDConnected,
		PowerLED:          devices.PowerLED,
	})
	return func(context.Context, *protocol.CallToolRequest, PowerStatusParams) (*protocol.CallToolResult, PowerStatusResult, error) {
		connected, on, err := power.LED()
		switch {
		case !connected:
			return nil, PowerStatusResult{State: hostpower.Unknown.String(), Message: "the power LED is not connected to the NanoKVM"}, nil
		case err != nil:
			return nil, PowerStatusResult{State: hostpower.Unknown.String(), LEDConnected: true, Message: "the power LED could not be read: " + err.Error()}, nil
		case on:
			return nil, PowerStatusResult{State: hostpower.On.String(), LEDConnected: true}, nil
		}
		return nil, PowerStatusResult{State: hostpower.Off.String(), LEDConnected: true}, nil
	}
}

var errPressBusy = errors.New("another button press is still in progress")

// powerHold is how long a call holds its button.
func powerHold(in PowerParams) (time.Duration, error) {
	press := strings.ToLower(strings.TrimSpace(in.Press))
	if press != "" && in.HoldMS != nil {
		return 0, fmt.Errorf("give press or holdMs, not both")
	}
	if in.HoldMS != nil {
		return time.Duration(clampInt(*in.HoldMS, minPowerHoldMS, maxPowerHoldMS)) * time.Millisecond, nil
	}
	switch press {
	case "", pressShort:
		return hostpower.ShortPress, nil
	case pressLong:
		return hostpower.LongPress, nil
	}
	return 0, fmt.Errorf("press must be short or long")
}

func powerHandler(devices Devices) protocol.ToolHandlerFor[PowerParams, ToolOKResult] {
	// vm.PressButton serializes presses from every caller, so a second one
	// waits for the first and then presses. Two presses an agent sent at once
	// are a mistake, since a second short press undoes the first, so MCP
	// refuses a call while one of its own presses is held.
	var pressing atomic.Bool
	return func(_ context.Context, _ *protocol.CallToolRequest, in PowerParams) (*protocol.CallToolResult, ToolOKResult, error) {
		button := strings.ToLower(strings.TrimSpace(in.Button))
		if button != hostpower.ButtonPower && button != hostpower.ButtonReset {
			return nil, ToolOKResult{}, fmt.Errorf("button must be power or reset")
		}
		hold, err := powerHold(in)
		if err != nil {
			return nil, ToolOKResult{}, err
		}
		if !pressing.CompareAndSwap(false, true) {
			return nil, ToolOKResult{}, errPressBusy
		}
		defer pressing.Store(false)
		if err := devices.PressButton(button, hold); err != nil {
			return nil, ToolOKResult{}, err
		}
		return nil, ToolOKResult{OK: true}, nil
	}
}

func mediaListHandler(devices Devices) protocol.ToolHandlerFor[MediaListParams, MediaResult] {
	return func(context.Context, *protocol.CallToolRequest, MediaListParams) (*protocol.CallToolResult, MediaResult, error) {
		drives, images, err := devices.ListMedia()
		if err != nil {
			return nil, MediaResult{}, err
		}
		return nil, MediaResult{OK: true, Drives: drives, Images: images}, nil
	}
}

func mediaHandler(devices Devices) protocol.ToolHandlerFor[MediaParams, MediaResult] {
	return func(_ context.Context, _ *protocol.CallToolRequest, in MediaParams) (*protocol.CallToolResult, MediaResult, error) {
		switch strings.ToLower(strings.TrimSpace(in.Action)) {
		case "insert":
			drive := normalizedDrive(in.Drive)
			if drive == "" || in.File == "" {
				return nil, MediaResult{}, fmt.Errorf("insert needs a drive (cdrom or disk) and a file")
			}
			readOnly := in.ReadOnly == nil || *in.ReadOnly
			if drive == "cdrom" && !readOnly {
				return nil, MediaResult{}, fmt.Errorf("the cdrom drive is read-only; insert into disk for a writable image")
			}
			if err := devices.InsertMedia(drive, in.File, readOnly); err != nil {
				return nil, MediaResult{}, err
			}
			return nil, mediaAfter(devices), nil
		case "eject":
			drive := normalizedDrive(in.Drive)
			if drive == "" {
				return nil, MediaResult{}, fmt.Errorf("eject needs a drive (cdrom or disk)")
			}
			if err := devices.EjectMedia(drive); err != nil {
				return nil, MediaResult{}, err
			}
			return nil, mediaAfter(devices), nil
		default:
			return nil, MediaResult{}, fmt.Errorf("action must be insert or eject; cube_media_list lists the drives and images")
		}
	}
}

// mediaAfter is a change's result: the drives as they now are, when they can
// be listed, so the caller sees what the host has without another call.
func mediaAfter(devices Devices) MediaResult {
	result := MediaResult{OK: true}
	if devices.ListMedia != nil {
		if drives, _, err := devices.ListMedia(); err == nil {
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

func describeNothing[In any](In) (string, string) {
	return "", ""
}

func describePower(in PowerParams) (string, string) {
	button := strings.ToLower(strings.TrimSpace(in.Button))
	hold, err := powerHold(in)
	if err != nil {
		return button, ""
	}
	detail := fmt.Sprintf("hold=%dms", hold.Milliseconds())
	if in.HoldMS == nil {
		press := strings.ToLower(strings.TrimSpace(in.Press))
		if press == "" {
			press = pressShort
		}
		detail = press + " " + detail
	}
	return button, detail
}

func describeMedia(in MediaParams) (string, string) {
	action := strings.ToLower(strings.TrimSpace(in.Action))
	detail := strings.TrimSpace(action + " " + normalizedDrive(in.Drive))
	if action == "insert" {
		if in.ReadOnly != nil && !*in.ReadOnly {
			detail += " rw"
		} else {
			detail += " ro"
		}
	}
	return in.File, detail
}

func powerSchema() *jsonschema.Schema {
	return objectSchema(map[string]*jsonschema.Schema{
		"button": {Type: "string", Enum: []any{hostpower.ButtonPower, hostpower.ButtonReset}, Description: "The front-panel button to press"},
		"press": {Type: "string", Enum: []any{pressShort, pressLong}, Description: fmt.Sprintf(
			"short holds the button %s (the default); long holds it %s, which on the power button forces the host off",
			hostpower.ShortPress, hostpower.LongPress)},
		"holdMs": integerSchema("Exact hold time in milliseconds, instead of press", minPowerHoldMS, maxPowerHoldMS),
	}, "button")
}

func mediaSchema() *jsonschema.Schema {
	return objectSchema(map[string]*jsonschema.Schema{
		"action":   {Type: "string", Enum: []any{"insert", "eject"}},
		"drive":    {Type: "string", Enum: []any{"cdrom", "disk"}, Description: "The drive to insert into or eject"},
		"file":     {Type: "string", Description: "insert: an image path exactly as cube_media_list returns it", MaxLength: jsonschema.Ptr(512)},
		"readOnly": {Type: "boolean", Description: "insert into disk: false lets the host write to and format the image; defaults to true. The cdrom drive is always read-only."},
	}, "action", "drive")
}
