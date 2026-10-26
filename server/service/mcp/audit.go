package mcpservice

import (
	"context"
	"fmt"
	"strings"

	"NanoKVM-Server/service/audit"

	protocol "github.com/modelcontextprotocol/go-sdk/mcp"
)

// toolTiers say which capability tier each tool needs. A tool missing from
// the table needs the admin tier, so a tool added without a decision is off
// until an owner opts in.
var toolTiers = map[string]audit.Tier{
	screenshotToolName:  audit.TierObserve,
	typeTextToolName:    audit.TierInput,
	pressKeysToolName:   audit.TierInput,
	moveMouseToolName:   audit.TierInput,
	clickMouseToolName:  audit.TierInput,
	scrollMouseToolName: audit.TierInput,
	powerStatusToolName: audit.TierObserve,
	mediaListToolName:   audit.TierObserve,
	powerToolName:       audit.TierPower,
	mediaToolName:       audit.TierMedia,
	wakeToolName:        audit.TierPower,
}

func toolTier(name string) audit.Tier {
	if tier, ok := toolTiers[name]; ok {
		return tier
	}
	return audit.TierAdmin
}

// clientIP is the address the request came from. The router stores it in a
// header the client cannot set (APIKeyMiddleware overwrites it).
func clientIP(req *protocol.CallToolRequest) string {
	if req == nil || req.Extra == nil {
		return ""
	}
	return req.Extra.Header.Get(audit.ClientHeader)
}

// guard wraps a tool handler with its tier check and its audit record. A call
// whose tier is off never reaches the handler; it is recorded as denied. The
// record names the tool, the tier and a summary of the arguments that leaves
// out what was typed.
func guard[In, Out any](
	name string,
	describe func(In) (target, detail string),
	next protocol.ToolHandlerFor[In, Out],
) protocol.ToolHandlerFor[In, Out] {
	tier := toolTier(name)
	return func(ctx context.Context, req *protocol.CallToolRequest, in In) (*protocol.CallToolResult, Out, error) {
		var none Out
		e := audit.Event{
			Actor:    "mcp",
			Via:      "mcp",
			IP:       clientIP(req),
			Category: audit.CategoryAgent,
			Action:   name,
			Tier:     string(tier),
		}
		if describe != nil {
			e.Target, e.Detail = describe(in)
		}

		if !audit.Allowed(audit.SubjectMCP, tier) {
			e.Outcome = audit.OutcomeDenied
			audit.Record(e)
			return nil, none, fmt.Errorf("%s", audit.DeniedMessage(audit.SubjectMCP, tier))
		}

		result, out, err := next(ctx, req, in)
		switch {
		case err != nil:
			e.Outcome = audit.OutcomeFail
			e.Detail = joinDetail(e.Detail, err.Error())
		case result != nil && result.IsError:
			e.Outcome = audit.OutcomeFail
			e.Detail = joinDetail(e.Detail, resultText(result))
		default:
			e.Outcome = audit.OutcomeOK
		}
		audit.Record(e)
		return result, out, err
	}
}

func joinDetail(parts ...string) string {
	var kept []string
	for _, p := range parts {
		if p != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, ": ")
}

func resultText(result *protocol.CallToolResult) string {
	for _, c := range result.Content {
		if text, ok := c.(*protocol.TextContent); ok {
			return text.Text
		}
	}
	return ""
}

// onlyAllowedTools drops from a tools/list answer the tools whose tier is
// off, so a client is not offered what it cannot use.
func onlyAllowedTools(next protocol.MethodHandler) protocol.MethodHandler {
	return func(ctx context.Context, method string, req protocol.Request) (protocol.Result, error) {
		result, err := next(ctx, method, req)
		if err != nil || method != "tools/list" {
			return result, err
		}
		if list, ok := result.(*protocol.ListToolsResult); ok && list != nil {
			kept := list.Tools[:0:0]
			for _, tool := range list.Tools {
				if audit.Allowed(audit.SubjectMCP, toolTier(tool.Name)) {
					kept = append(kept, tool)
				}
			}
			list.Tools = kept
		}
		return result, err
	}
}

// What each tool's arguments are reduced to in the record. Text typed and
// keys pressed are counted, not kept: they may be a password.

func describeTypeText(in TypeTextParams) (string, string) {
	return "", fmt.Sprintf("chars=%d", len([]rune(in.Text)))
}

func describePressKeys(in PressKeysParams) (string, string) {
	return "", fmt.Sprintf("keys=%d", len(in.Keys))
}

func describeMoveMouse(in MoveMouseParams) (string, string) {
	if normalizedMode(in.Mode, "absolute") == "relative" {
		return "", fmt.Sprintf("relative dx=%d dy=%d", in.DeltaX, in.DeltaY)
	}
	if in.X == nil || in.Y == nil {
		return "", "absolute"
	}
	return "", fmt.Sprintf("absolute x=%.3f y=%.3f", *in.X, *in.Y)
}

func describeClickMouse(in ClickMouseParams) (string, string) {
	button := strings.ToLower(strings.TrimSpace(in.Button))
	if button == "" {
		button = "left"
	}
	clicks := in.Clicks
	if clicks == 0 {
		clicks = 1
	}
	detail := fmt.Sprintf("%s x%d", button, clicks)
	if in.X != nil && in.Y != nil {
		detail += fmt.Sprintf(" at %.3f,%.3f", *in.X, *in.Y)
	}
	return "", detail
}

func describeScroll(in ScrollMouseParams) (string, string) {
	return "", fmt.Sprintf("deltaY=%d", in.DeltaY)
}

func describeScreenshot(ScreenshotParams) (string, string) {
	return "", ""
}
