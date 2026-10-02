package picoclaw

import (
	"testing"
	"time"

	"NanoKVM-Server/service/agent"
	"NanoKVM-Server/service/controlmode"
)

func TestControlModeChangedEventIncludesControlMetadata(t *testing.T) {
	changedAt := time.Now().UTC()
	event := controlModeChangedEvent(controlmode.Status{
		Mode:          controlmode.ModeMCP,
		Transitioning: true,
		LastError:     "switch failed",
		ChangedAt:     changedAt,
	}, "mcp_config")

	want := agent.ControlMode{
		Mode:          string(controlmode.ModeMCP),
		Transitioning: true,
		CanControl:    false,
		LastError:     "switch failed",
		ChangedAt:     changedAt,
		Source:        "mcp_config",
	}
	if event.Type != agent.EventControlModeChanged || event.Control == nil || *event.Control != want {
		t.Fatalf("event = %+v, want control %+v", event, want)
	}
}

func TestControlModeChangedEventAllowsPicoclawControlWhenStable(t *testing.T) {
	event := controlModeChangedEvent(controlmode.Status{
		Mode: controlmode.ModePicoclaw,
	}, "")

	if !event.Control.CanControl {
		t.Fatal("can_control = false, want true")
	}
	if event.Control.Source != "" {
		t.Fatalf("source = %q, want empty", event.Control.Source)
	}
}
