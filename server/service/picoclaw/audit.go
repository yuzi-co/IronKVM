package picoclaw

import (
	"encoding/json"
	"fmt"
	"strings"

	"NanoKVM-Server/service/audit"
)

// agentToolTiers say which capability tier each tool of the on-board agent
// needs. A tool the table does not name is unknown and refused.
var agentToolTiers = map[string]audit.Tier{
	"kvm_screenshot":    audit.TierObserve,
	"kvm_actions":       audit.TierInput,
	powerStatusToolName: audit.TierObserve,
	powerToolName:       audit.TierPower,
	// Waiting for the answer to a press is part of pressing.
	powerConfirmationToolName: audit.TierPower,
	mediaListToolName:         audit.TierObserve,
	mediaToolName:             audit.TierMedia,
}

// agentToolOrder is the order tools/list gives the agent's tools in.
var agentToolOrder = []string{
	"kvm_screenshot",
	"kvm_actions",
	powerStatusToolName,
	powerToolName,
	powerConfirmationToolName,
	mediaListToolName,
	mediaToolName,
}

// describeAgentCall reduces a call's arguments to the audit record's target
// and detail.
func describeAgentCall(name string, arguments json.RawMessage) (target, detail string) {
	if name == "kvm_actions" {
		return "", summarizeAgentCall(name, arguments)
	}
	return describeDeviceCall(name, arguments)
}

// summarizeAgentCall reduces a call's arguments to what a reviewer needs: how
// many actions of which kinds. What was typed and which keys were pressed are
// left out, since they may be a password.
func summarizeAgentCall(name string, arguments json.RawMessage) string {
	if name != "kvm_actions" {
		return ""
	}
	var params struct {
		Actions []struct {
			Action string `json:"action"`
		} `json:"actions"`
		ScreenshotAfter bool `json:"screenshot_after"`
	}
	if json.Unmarshal(arguments, &params) != nil {
		return ""
	}
	kinds := make([]string, 0, len(params.Actions))
	for _, a := range params.Actions {
		kinds = append(kinds, a.Action)
	}
	detail := fmt.Sprintf("actions=%d %s", len(kinds), strings.Join(kinds, ","))
	if params.ScreenshotAfter {
		detail += " screenshot_after"
	}
	return strings.TrimSpace(detail)
}

func joinAuditDetail(parts ...string) string {
	var kept []string
	for _, p := range parts {
		if p != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, ": ")
}
