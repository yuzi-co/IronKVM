package picoclaw

import (
	"NanoKVM-Server/service/agent"
	"NanoKVM-Server/service/controlmode"
)

func (s *Service) PublishControlModeChanged(status controlmode.Status) {
	s.PublishControlModeChangedFrom(status, "")
}

// PublishControlModeChangedFrom tells every open chat session that the AI
// control mode changed.
func (s *Service) PublishControlModeChangedFrom(status controlmode.Status, source string) {
	if s == nil {
		return
	}
	s.ensureDependencies()
	s.chat.Broadcast(controlModeChangedEvent(status, source))
}

func controlModeChangedEvent(status controlmode.Status, source string) agent.Event {
	return agent.Event{
		Type: agent.EventControlModeChanged,
		Control: &agent.ControlMode{
			Mode:          string(status.Mode),
			Transitioning: status.Transitioning,
			CanControl:    status.Mode == controlmode.ModePicoclaw && !status.Transitioning,
			LastError:     status.LastError,
			ChangedAt:     status.ChangedAt,
			Source:        source,
		},
	}
}
