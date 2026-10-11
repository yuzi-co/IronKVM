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
	"kvm_screenshot": audit.TierObserve,
	"kvm_actions":    audit.TierInput,
}

// allowedToolDefinitions is the tools/list answer: the tools whose tier is on.
func allowedToolDefinitions() []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(mcpToolDefinitions))
	for _, tool := range mcpToolDefinitions {
		name, _ := tool["name"].(string)
		if tier, ok := agentToolTiers[name]; ok && audit.Allowed(audit.SubjectAgent, tier) {
			out = append(out, tool)
		}
	}
	return out
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

// mcpErrorText is the first text of a tool error result.
func mcpErrorText(result map[string]interface{}) string {
	content, _ := result["content"].([]map[string]interface{})
	for _, item := range content {
		if text, ok := item["text"].(string); ok {
			return text
		}
	}
	return ""
}
