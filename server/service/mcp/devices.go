package mcpservice

import (
	"context"

	"github.com/google/jsonschema-go/jsonschema"
	protocol "github.com/modelcontextprotocol/go-sdk/mcp"

	"NanoKVM-Server/service/hostdevices"
)

const (
	powerToolName       = "cube_power"
	powerStatusToolName = "cube_power_status"
	mediaToolName       = "cube_media"
	mediaListToolName   = "cube_media_list"
)

// MediaDrive is one virtual drive and what it holds.
type MediaDrive = hostdevices.MediaDrive

// Devices are the board's power button, power LED and virtual media, which
// the router supplies. A nil function leaves its tool out. Reading them is in
// the observe tier; pressing a button and changing a drive are in the power
// and media tiers, which are off unless an administrator turns them on. The
// checks behind the tools are in hostdevices, which the on-board agent's
// tools share.
type Devices = hostdevices.Devices

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
	URL      string `json:"url,omitempty" jsonschema:"insert only: an http or https URL to serve the image from, instead of a file"`
	ReadOnly *bool  `json:"readOnly,omitempty" jsonschema:"disk only: false lets the host write to the image"`
}

type MediaResult struct {
	OK     bool         `json:"ok"`
	Drives []MediaDrive `json:"drives,omitempty"`
	Images []string     `json:"images,omitempty"`
}

var errPressBusy = hostdevices.ErrPressBusy

func registerDeviceTools(server *protocol.Server, devices Devices) {
	if devices.CanReadPower() {
		protocol.AddTool(server, &protocol.Tool{
			Name:        powerStatusToolName,
			Description: hostdevices.PowerStatusDescription,
			InputSchema: objectSchema(map[string]*jsonschema.Schema{}),
		}, guard(powerStatusToolName, describeNothing[PowerStatusParams], powerStatusHandler(devices)))
	}
	if devices.CanPress() {
		protocol.AddTool(server, &protocol.Tool{
			Name:        powerToolName,
			Description: hostdevices.PowerDescription(powerStatusToolName),
			InputSchema: hostdevices.PowerSchema(true),
		}, guard(powerToolName, describePower, powerHandler(devices)))
	}
	if devices.CanListMedia() {
		protocol.AddTool(server, &protocol.Tool{
			Name:        mediaListToolName,
			Description: hostdevices.MediaListDescription,
			InputSchema: objectSchema(map[string]*jsonschema.Schema{}),
		}, guard(mediaListToolName, describeNothing[MediaListParams], mediaListHandler(devices)))
	}
	if devices.CanChangeMedia() {
		protocol.AddTool(server, &protocol.Tool{
			Name:        mediaToolName,
			Description: hostdevices.MediaDescription(mediaListToolName),
			InputSchema: hostdevices.MediaSchema(mediaListToolName),
		}, guard(mediaToolName, describeMedia, mediaHandler(devices)))
	}
}

func powerStatusHandler(devices Devices) protocol.ToolHandlerFor[PowerStatusParams, PowerStatusResult] {
	return func(context.Context, *protocol.CallToolRequest, PowerStatusParams) (*protocol.CallToolResult, PowerStatusResult, error) {
		return nil, PowerStatusResult(devices.ReadPower()), nil
	}
}

func (in PowerParams) press() hostdevices.Press {
	return hostdevices.Press{Button: in.Button, Press: in.Press, HoldMS: in.HoldMS}
}

// powerHandler refuses a call while one of MCP's own presses is held; see
// hostdevices.Presser.
func powerHandler(devices Devices) protocol.ToolHandlerFor[PowerParams, ToolOKResult] {
	presser := hostdevices.NewPresser(devices)
	return func(_ context.Context, _ *protocol.CallToolRequest, in PowerParams) (*protocol.CallToolResult, ToolOKResult, error) {
		if err := presser.Press(in.press()); err != nil {
			return nil, ToolOKResult{}, err
		}
		return nil, ToolOKResult{OK: true}, nil
	}
}

func mediaListHandler(devices Devices) protocol.ToolHandlerFor[MediaListParams, MediaResult] {
	return func(context.Context, *protocol.CallToolRequest, MediaListParams) (*protocol.CallToolResult, MediaResult, error) {
		state, err := devices.List()
		if err != nil {
			return nil, MediaResult{}, err
		}
		return nil, MediaResult(state), nil
	}
}

func (in MediaParams) change() hostdevices.Change {
	return hostdevices.Change{Action: in.Action, Drive: in.Drive, File: in.File, URL: in.URL, ReadOnly: in.ReadOnly}
}

func mediaHandler(devices Devices) protocol.ToolHandlerFor[MediaParams, MediaResult] {
	return func(_ context.Context, _ *protocol.CallToolRequest, in MediaParams) (*protocol.CallToolResult, MediaResult, error) {
		state, err := devices.ChangeMedia(in.change(), mediaListToolName)
		if err != nil {
			return nil, MediaResult{}, err
		}
		return nil, MediaResult(state), nil
	}
}

func describeNothing[In any](In) (string, string) {
	return "", ""
}

func describePower(in PowerParams) (string, string) {
	return hostdevices.DescribePress(in.press())
}

func describeMedia(in MediaParams) (string, string) {
	return hostdevices.DescribeChange(in.change())
}
