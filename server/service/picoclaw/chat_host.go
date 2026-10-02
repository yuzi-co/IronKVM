package picoclaw

import (
	"net/http"
	"time"

	"NanoKVM-Server/service/agent"
	"NanoKVM-Server/service/controlmode"
	"NanoKVM-Server/service/stream/mjpeg"
)

// The Service is the KVM side of the chat bridge: it decides who may chat,
// holds the AI input lock and cleans up when a chat session ends. None of it
// depends on the agent.
var _ agent.Host = (*Service)(nil)

// SetChat connects the service to the chat bridge, which it uses to publish
// observations and control mode changes and to close chat sessions.
func (s *Service) SetChat(chat ChatHub) {
	if chat == nil {
		chat = noChat{}
	}
	s.chat = chat
}

// Admit lets a chat session open while the AI control mode allows chat and
// the session can hold the AI input lock.
func (s *Service) Admit(sessionID string) error {
	s.ensureDependencies()
	modeStatus, err := s.control.Status()
	if err != nil {
		return agent.NewError(CodeRuntimeUnavailable, err.Error())
	}
	if modeStatus.Transitioning || modeStatus.Mode == controlmode.ModeMCP {
		controlErr := s.controlWriteError(controlmode.ModePicoclaw, nil)
		return &agent.Error{Status: http.StatusConflict, Code: controlErr.Code, Message: controlErr.Message}
	}
	if lockErr := s.lock.Ensure(sessionID); lockErr != nil {
		return &agent.Error{Status: http.StatusConflict, Code: CodePicoclawLockHeld, Message: lockErr.Message}
	}
	return nil
}

func (s *Service) Renew(sessionID string) bool {
	s.ensureDependencies()
	return s.lock.Renew(sessionID)
}

// Opened keeps the latest stream frame cached while a chat session is open,
// so the agent's screenshots can reuse it.
func (s *Service) Opened(string) {
	mjpeg.EnableLatestFrameCache()
}

// Closed releases what a chat session held: the frame cache, the media the
// agent loaded, the AI input lock and any input left held on the host. A
// session that ended because the agent was not reachable marks the runtime
// unavailable.
func (s *Service) Closed(sessionID string, closeCode int, reason string, opened bool) {
	s.ensureDependencies()
	if opened {
		mjpeg.DisableLatestFrameCache()
		cleanupPicoclawMediaTempDir()
	}
	s.releaseGatewaySession(sessionID)

	status := s.runtime.Get()
	status.CurrentSession = s.lock.Owner()
	if closeCode == CloseCodeUpstreamClosed || closeCode == CloseCodeRuntimeUnavailable {
		status.Ready = false
		if closeCode == CloseCodeUpstreamClosed {
			status.Status = "unavailable"
		}
		status.LastError = reason
		status.CheckedAt = time.Now()
		s.runtime.Set(status)
	}
}

// RelayConfig gives the chat socket the limits of PicoClaw's gateway
// connection, read from PicoClaw's configuration when the session opened.
func (s *Service) RelayConfig() agent.RelayConfig {
	s.ensureDependencies()
	cfg := s.config.Get()
	return agent.RelayConfig{
		ReadTimeout:     time.Duration(cfg.ReadTimeoutMs) * time.Millisecond,
		WriteTimeout:    time.Duration(cfg.WriteTimeoutMs) * time.Millisecond,
		PingInterval:    time.Duration(cfg.PingIntervalMs) * time.Millisecond,
		MaxMessageBytes: int64(cfg.MaxMessageBytes),
	}
}

// noChat stands in for the bridge until SetChat is called, as in tests.
type noChat struct{}

func (noChat) Publish(string, agent.Event) error     { return agent.ErrSessionNotActive }
func (noChat) Broadcast(agent.Event)                 {}
func (noChat) Prompt(string, agent.Prompt) error     { return agent.ErrSessionNotActive }
func (noChat) IsActive(string) bool                  { return false }
func (noChat) CloseSession(string, int, string) bool { return false }
func (noChat) CloseAll(int, string) int              { return 0 }
func (noChat) SessionCount() int                     { return 0 }
