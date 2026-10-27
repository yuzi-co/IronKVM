package picoclaw

import (
	"context"
	"encoding/json"

	"NanoKVM-Server/service/hostdevices"

	protocol "github.com/modelcontextprotocol/go-sdk/mcp"
)

// The device tools: the host's power state and buttons, and its virtual
// media. They are tools of their own rather than kvm_actions steps because
// each needs a tier of its own (observe, power or media) and is listed only
// when that tier is on, and because a press or an insert is a single
// deliberate act that should not ride in a batch of clicks. The checks behind
// them are hostdevices', which external MCP's cube_* tools share.
const (
	powerStatusToolName = "kvm_power_status"
	powerToolName       = "kvm_power"
	mediaListToolName   = "kvm_media_list"
	mediaToolName       = "kvm_media"
)

// agentPowerNote is what the agent's press tool adds to the shared
// description: the owner, not the agent, decides when the host goes down, and
// the server holds the press until a person confirms it (power_confirm.go).
const agentPowerNote = " This tool does not press at once: it answers " + CodeConfirmationRequired + " with an id, and the press waits up to 60 seconds for a signed-in user to confirm it in the IronKVM web UI. Tell the user which button and press you asked for and ask them to confirm it there (a chat reply, or a reply in a channel such as Telegram, does not confirm it), then call " + powerConfirmationToolName + " with the id to learn whether it was pressed."

// powerConfirmationDescription describes the tool the agent waits with.
const powerConfirmationDescription = "Wait for the user's answer to a press that " + powerToolName + " asked for. Give the id it returned. Waits up to 25 seconds and returns the state: pending (still waiting, call again), pressed (the user confirmed and the button was pressed), denied (the user refused; do not press), expired (nobody confirmed within 60 seconds, or the turn ended; nothing was pressed) or failed (the user confirmed but the press could not be made)."

// SetDevices gives the agent's tools the board's devices. The router calls it
// before the first MCP call; a function left nil leaves its tool out.
func (s *Service) SetDevices(devices hostdevices.Devices) {
	s.devices = devices
	s.presser = hostdevices.NewPresser(devices)
}

type deviceTool struct {
	tool    *protocol.Tool
	handler protocol.ToolHandler
}

// deviceTools are the device tools the board's devices allow. Pressing and
// changing media need PicoClaw to hold device control and a chat session in a
// turn to hold the AI input lock, as kvm_actions does; reading the power LED
// or the drive list does not, since neither shows the user's screen or
// changes the host.
func (s *Service) deviceTools() []deviceTool {
	devices := s.devices
	empty := map[string]any{"type": "object", "properties": map[string]any{}}
	var tools []deviceTool
	if devices.CanReadPower() {
		tools = append(tools, deviceTool{
			tool:    &protocol.Tool{Name: powerStatusToolName, Description: hostdevices.PowerStatusDescription, InputSchema: empty},
			handler: s.mcpPowerStatus,
		})
	}
	if devices.CanPress() {
		tools = append(tools, deviceTool{
			tool: &protocol.Tool{
				Name:        powerToolName,
				Description: hostdevices.PowerDescription(powerStatusToolName) + agentPowerNote,
				InputSchema: hostdevices.PowerSchema(false),
			},
			handler: s.withAgentSession(s.requestPowerPress),
		}, deviceTool{
			tool: &protocol.Tool{
				Name:        powerConfirmationToolName,
				Description: powerConfirmationDescription,
				InputSchema: map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"id": map[string]any{"type": "string", "description": "The id " + powerToolName + " returned."},
					},
					"required": []string{"id"},
				},
			},
			handler: s.mcpPowerConfirmation,
		})
	}
	if devices.CanListMedia() {
		tools = append(tools, deviceTool{
			tool:    &protocol.Tool{Name: mediaListToolName, Description: hostdevices.MediaListDescription, InputSchema: empty},
			handler: s.mcpMediaList,
		})
	}
	if devices.CanChangeMedia() {
		tools = append(tools, deviceTool{
			tool: &protocol.Tool{
				Name:        mediaToolName,
				Description: hostdevices.MediaDescription(mediaListToolName),
				InputSchema: hostdevices.MediaSchema(mediaListToolName),
			},
			handler: s.withAgentSession(s.mcpMedia),
		})
	}
	return tools
}

// The arguments of the press and media tools. The press tool offers short
// and long only: an exact hold time is for scripts, not for a chat.
type powerArguments struct {
	Button string `json:"button"`
	Press  string `json:"press"`
}

type mediaArguments struct {
	Action   string `json:"action"`
	Drive    string `json:"drive"`
	File     string `json:"file"`
	URL      string `json:"url"`
	ReadOnly *bool  `json:"readOnly"`
}

func (a powerArguments) press() hostdevices.Press {
	return hostdevices.Press{Button: a.Button, Press: a.Press}
}

func (a mediaArguments) change() hostdevices.Change {
	return hostdevices.Change{Action: a.Action, Drive: a.Drive, File: a.File, URL: a.URL, ReadOnly: a.ReadOnly}
}

func decodeArguments(req *protocol.CallToolRequest, v any) bool {
	args := toolArguments(req)
	if len(args) == 0 {
		return true
	}
	return json.Unmarshal(args, v) == nil
}

func (s *Service) mcpPowerStatus(context.Context, *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	return mcpJSONResult(s.devices.ReadPower()), nil
}

func (s *Service) mcpMediaList(context.Context, *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	state, err := s.devices.List()
	if err != nil {
		return mcpToolError(err.Error()), nil
	}
	return mcpJSONResult(state), nil
}

func (s *Service) mcpMedia(_ context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var args mediaArguments
	if !decodeArguments(req, &args) {
		return mcpToolError("invalid arguments: give action (insert or eject) and drive (cdrom or disk)"), nil
	}
	state, err := s.devices.ChangeMedia(args.change(), mediaListToolName)
	if err != nil {
		return mcpToolError(err.Error()), nil
	}
	return mcpJSONResult(state), nil
}

func mcpJSONResult(v any) *protocol.CallToolResult {
	data, _ := json.Marshal(v)
	return &protocol.CallToolResult{Content: []protocol.Content{&protocol.TextContent{Text: string(data)}}}
}

// describeDeviceCall reduces a device tool's arguments to what the audit log
// keeps, as external MCP does: a URL leaves only its host.
func describeDeviceCall(name string, arguments json.RawMessage) (target, detail string) {
	switch name {
	case powerToolName:
		var args powerArguments
		if json.Unmarshal(arguments, &args) == nil {
			return hostdevices.DescribePress(args.press())
		}
	case powerConfirmationToolName:
		var args confirmationArguments
		if json.Unmarshal(arguments, &args) == nil {
			return args.ID, ""
		}
	case mediaToolName:
		var args mediaArguments
		if json.Unmarshal(arguments, &args) == nil {
			return hostdevices.DescribeChange(args.change())
		}
	}
	return "", ""
}
