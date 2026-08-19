package picoclaw

import (
	"testing"

	serverConfig "NanoKVM-Server/config"
)

// The internal token lives in two places at once. The server generates it into
// /etc/kvm/.picoclaw_internal_token and caches it for the process lifetime, and
// setMCPServer copies it into PicoClaw's own config as a request header. The
// copy is what PicoClaw actually sends back over loopback.
//
// The copy has to follow the token. If it is ever allowed to go stale, every
// MCP call the agent makes returns 401, and it fails the way that is hardest to
// diagnose: no crash, no log on the browser side, the agent simply stops being
// able to touch the machine.
//
// setMCPServer overwrites the header today. It also has a create-only branch
// for the entry itself, and that branch is one careless "do not clobber what
// the operator edited" refactor away from swallowing the header as well. These
// tests are what would notice.

const staleToken = "0000000000000000000000000000000000000000"

// mcpHeader returns the internal token header PicoClaw would send, or "" when
// the config carries no such entry.
func mcpHeader(t *testing.T, raw map[string]any) string {
	t.Helper()

	tools, ok := raw["tools"].(map[string]any)
	if !ok {
		return ""
	}
	mcp, ok := tools["mcp"].(map[string]any)
	if !ok {
		return ""
	}
	servers, ok := mcp["servers"].(map[string]any)
	if !ok {
		return ""
	}
	entry, ok := servers["nanokvm"].(map[string]any)
	if !ok {
		return ""
	}
	headers, ok := entry["headers"].(map[string]any)
	if !ok {
		return ""
	}

	token, _ := headers[serverConfig.PicoclawInternalTokenHeader].(string)
	return token
}

// nanokvmEntry builds the fields defaultPicoclawMCPServer produces, without
// reading the token file, so the test stays off /etc/kvm.
func nanokvmEntry(token string) map[string]any {
	return map[string]any{
		"enabled": true,
		"type":    "http",
		"url":     "http://127.0.0.1:80/api/picoclaw/mcp",
		"headers": map[string]any{
			serverConfig.PicoclawInternalTokenHeader: token,
		},
	}
}

// configWithMCPServer builds the raw document json.Unmarshal produces for a
// config that already declares the nanokvm MCP server.
func configWithMCPServer(token string) map[string]any {
	return map[string]any{
		"tools": map[string]any{
			"mcp": map[string]any{
				"servers": map[string]any{
					"nanokvm": nanokvmEntry(token),
				},
			},
		},
	}
}

func TestSetMCPServerReplacesAStaleInternalToken(t *testing.T) {
	editor := &picoclawConfigEditor{raw: configWithMCPServer(staleToken)}

	editor.setMCPServer("nanokvm", nanokvmEntry("current-token"))

	if got := mcpHeader(t, editor.raw); got != "current-token" {
		t.Fatalf("PicoClaw would send %q, want the current token", got)
	}

	// Without this the new header stays in memory and is never written back,
	// so the config on disk keeps the token that no longer works.
	if !editor.changed {
		t.Fatal("the editor did not record a change, so the config is never saved")
	}
}

func TestSetMCPServerKeepsAnEntryItDidNotWrite(t *testing.T) {
	// Refreshing the token must not cost the operator the rest of the entry.
	raw := configWithMCPServer(staleToken)
	entry := raw["tools"].(map[string]any)["mcp"].(map[string]any)["servers"].(map[string]any)["nanokvm"].(map[string]any)
	entry["timeout"] = float64(30)

	editor := &picoclawConfigEditor{raw: raw}
	editor.setMCPServer("nanokvm", nanokvmEntry("current-token"))

	if got := mcpHeader(t, editor.raw); got != "current-token" {
		t.Fatalf("PicoClaw would send %q, want the current token", got)
	}
	if got := entry["timeout"]; got != float64(30) {
		t.Fatalf("timeout is %v, want the value the operator set", got)
	}
}

func TestSetMCPServerLeavesAMatchingTokenAlone(t *testing.T) {
	// The config lives on the SD card. Rewriting it on every start would wear
	// the card for nothing, so an unchanged token has to read as unchanged.
	editor := &picoclawConfigEditor{raw: configWithMCPServer("current-token")}

	editor.setMCPServer("nanokvm", nanokvmEntry("current-token"))

	if editor.changed {
		t.Fatal("the editor rewrote a config that already held the current token")
	}
}

func TestSetMCPServerCreatesTheEntryWhenTheConfigHasNone(t *testing.T) {
	editor := &picoclawConfigEditor{raw: map[string]any{}}

	editor.setMCPServer("nanokvm", nanokvmEntry("current-token"))

	if got := mcpHeader(t, editor.raw); got != "current-token" {
		t.Fatalf("PicoClaw would send %q, want the current token", got)
	}
	if !editor.changed {
		t.Fatal("the editor did not record a change, so the config is never saved")
	}
}
