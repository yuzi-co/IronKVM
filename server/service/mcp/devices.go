package mcpservice

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	protocol "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	powerToolName = "cube_power"
	mediaToolName = "cube_media"

	defaultPowerHoldMS = 800
	minPowerHoldMS     = 100
	maxPowerHoldMS     = 10000
)

// MediaDrive is one virtual drive and what it holds.
type MediaDrive struct {
	ID       string `json:"id"`
	File     string `json:"file"`
	ReadOnly bool   `json:"readOnly"`
}

// Devices are the board's power button and virtual media, which the router
// supplies. A nil function leaves its tool out: they are behind the power and
// media tiers, which are off unless an administrator turns them on.
type Devices struct {
	// PressButton holds "power" or "reset" for d.
	PressButton func(kind string, d time.Duration) error
	// ListMedia returns the drives and the image files that can be loaded.
	ListMedia func() (drives []MediaDrive, images []string, err error)
	// InsertMedia loads an image into a drive ("cdrom" or "disk").
	InsertMedia func(drive, file string, readOnly bool) error
	// EjectMedia empties a drive.
	EjectMedia func(drive string) error
}

type PowerParams struct {
	Button string `json:"button" jsonschema:"power or reset"`
	HoldMS *int   `json:"holdMs,omitempty" jsonschema:"hold time in milliseconds"`
}

type MediaParams struct {
	Action string `json:"action" jsonschema:"list, insert or eject"`
	Drive  string `json:"drive,omitempty" jsonschema:"cdrom or disk"`
	File   string `json:"file,omitempty" jsonschema:"image path as listed"`
}

type MediaResult struct {
	OK     bool         `json:"ok"`
	Drives []MediaDrive `json:"drives,omitempty"`
	Images []string     `json:"images,omitempty"`
}

func registerDeviceTools(server *protocol.Server, devices Devices) {
	if devices.PressButton != nil {
		protocol.AddTool(server, &protocol.Tool{
			Name:        powerToolName,
			Description: "Press the ATX power or reset button of the remote host. Holding power for several seconds forces it off.",
			InputSchema: powerSchema(),
		}, guard(powerToolName, describePower, powerHandler(devices)))
	}
	if devices.ListMedia != nil && devices.InsertMedia != nil && devices.EjectMedia != nil {
		protocol.AddTool(server, &protocol.Tool{
			Name:        mediaToolName,
			Description: "List, insert or eject the virtual CD and disk images the remote host sees",
			InputSchema: mediaSchema(),
		}, guard(mediaToolName, describeMedia, mediaHandler(devices)))
	}
}

func powerHandler(devices Devices) protocol.ToolHandlerFor[PowerParams, ToolOKResult] {
	return func(_ context.Context, _ *protocol.CallToolRequest, in PowerParams) (*protocol.CallToolResult, ToolOKResult, error) {
		button := strings.ToLower(strings.TrimSpace(in.Button))
		if button != "power" && button != "reset" {
			return nil, ToolOKResult{}, fmt.Errorf("button must be power or reset")
		}
		hold := defaultPowerHoldMS
		if in.HoldMS != nil {
			hold = clampInt(*in.HoldMS, minPowerHoldMS, maxPowerHoldMS)
		}
		if err := devices.PressButton(button, time.Duration(hold)*time.Millisecond); err != nil {
			return nil, ToolOKResult{}, err
		}
		return nil, ToolOKResult{OK: true}, nil
	}
}

func mediaHandler(devices Devices) protocol.ToolHandlerFor[MediaParams, MediaResult] {
	return func(_ context.Context, _ *protocol.CallToolRequest, in MediaParams) (*protocol.CallToolResult, MediaResult, error) {
		switch strings.ToLower(strings.TrimSpace(in.Action)) {
		case "list":
			drives, images, err := devices.ListMedia()
			if err != nil {
				return nil, MediaResult{}, err
			}
			return nil, MediaResult{OK: true, Drives: drives, Images: images}, nil
		case "insert":
			drive := normalizedDrive(in.Drive)
			if drive == "" || in.File == "" {
				return nil, MediaResult{}, fmt.Errorf("insert needs a drive (cdrom or disk) and a file")
			}
			// Read-only, the way the web UI's legacy mount serves images: the
			// host can neither write nor format an image it was given.
			if err := devices.InsertMedia(drive, in.File, true); err != nil {
				return nil, MediaResult{}, err
			}
			return nil, MediaResult{OK: true}, nil
		case "eject":
			drive := normalizedDrive(in.Drive)
			if drive == "" {
				return nil, MediaResult{}, fmt.Errorf("eject needs a drive (cdrom or disk)")
			}
			if err := devices.EjectMedia(drive); err != nil {
				return nil, MediaResult{}, err
			}
			return nil, MediaResult{OK: true}, nil
		default:
			return nil, MediaResult{}, fmt.Errorf("action must be list, insert or eject")
		}
	}
}

func normalizedDrive(drive string) string {
	switch d := strings.ToLower(strings.TrimSpace(drive)); d {
	case "cdrom", "disk":
		return d
	}
	return ""
}

func describePower(in PowerParams) (string, string) {
	hold := defaultPowerHoldMS
	if in.HoldMS != nil {
		hold = clampInt(*in.HoldMS, minPowerHoldMS, maxPowerHoldMS)
	}
	return strings.ToLower(strings.TrimSpace(in.Button)), fmt.Sprintf("hold=%dms", hold)
}

func describeMedia(in MediaParams) (string, string) {
	return in.File, strings.ToLower(strings.TrimSpace(in.Action)) + " " + normalizedDrive(in.Drive)
}

func powerSchema() *jsonschema.Schema {
	return objectSchema(map[string]*jsonschema.Schema{
		"button": {Type: "string", Enum: []any{"power", "reset"}},
		"holdMs": integerSchema(fmt.Sprintf("Hold time in milliseconds; defaults to %d when omitted", defaultPowerHoldMS), minPowerHoldMS, maxPowerHoldMS),
	}, "button")
}

func mediaSchema() *jsonschema.Schema {
	return objectSchema(map[string]*jsonschema.Schema{
		"action": {Type: "string", Enum: []any{"list", "insert", "eject"}},
		"drive":  {Type: "string", Enum: []any{"cdrom", "disk"}, Description: "insert and eject: the drive"},
		"file":   {Type: "string", Description: "insert: an image path exactly as list returns it", MaxLength: jsonschema.Ptr(512)},
	}, "action")
}
