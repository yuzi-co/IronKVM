package mcpservice

import (
	"context"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
	protocol "github.com/modelcontextprotocol/go-sdk/mcp"

	"NanoKVM-Server/service/wol"
)

const wakeToolName = "cube_wake_on_lan"

type WakeOnLANParams struct {
	Name string `json:"name,omitempty" jsonschema:"name of a device in the Wake-on-LAN list"`
	Mac  string `json:"mac,omitempty" jsonschema:"MAC address, when the device is not in the list"`
}

type WakeOnLANResult struct {
	OK  bool   `json:"ok"`
	Mac string `json:"mac"`
}

// wakeOnLAN is a var so the tests do not send a packet.
var wakeOnLAN = func(mac, name string) (string, error) {
	device, _, err := wol.Default.Wake(mac, name)
	return device.Mac, err
}

func wakeSchema() *jsonschema.Schema {
	return objectSchema(map[string]*jsonschema.Schema{
		"name": {Type: "string", Description: "Name of a device in the Wake-on-LAN list, e.g. nas"},
		"mac":  {Type: "string", Description: "MAC address, when the device is not in the list"},
	})
}

func wakeHandler() protocol.ToolHandlerFor[WakeOnLANParams, WakeOnLANResult] {
	return func(_ context.Context, _ *protocol.CallToolRequest, input WakeOnLANParams) (*protocol.CallToolResult, WakeOnLANResult, error) {
		if input.Name == "" && input.Mac == "" {
			return nil, WakeOnLANResult{}, fmt.Errorf("name or mac is required")
		}
		mac, err := wakeOnLAN(input.Mac, input.Name)
		if err != nil {
			return nil, WakeOnLANResult{}, fmt.Errorf("wake failed: %w", err)
		}

		return nil, WakeOnLANResult{OK: true, Mac: mac}, nil
	}
}
