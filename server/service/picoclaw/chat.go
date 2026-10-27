package picoclaw

import (
	"time"

	"NanoKVM-Server/service/agent"
	"NanoKVM-Server/service/stream/mjpeg"
)

// SetChat connects the service to the chat bridge, which it uses to publish
// observations and control mode changes and to close chat sessions.
func (s *Service) SetChat(chat agent.ChatHub) {
	if chat == nil {
		chat = agent.NoChat{}
	}
	s.chat = chat
	if s.lifecycle != nil {
		s.lifecycle.SetChat(chat)
	}
}

// ChatHost is the KVM side of the chat bridge for PicoClaw: the agent-neutral
// agent.ChatHost with PicoClaw's own part of a session's start and end.
func (s *Service) ChatHost() *agent.ChatHost {
	s.ensureDependencies()
	return s.host
}

// chatOpened keeps the latest stream frame cached while a chat session is
// open, so the agent's screenshots can reuse it.
func (s *Service) chatOpened(string) {
	mjpeg.EnableLatestFrameCache()
}

// chatClosed drops the frame cache and the media the agent loaded for a
// session that opened.
func (s *Service) chatClosed(_ string, opened bool) {
	if opened {
		mjpeg.DisableLatestFrameCache()
		cleanupPicoclawMediaTempDir()
	}
}

// relayConfig gives the chat socket the limits of PicoClaw's gateway
// connection, read from PicoClaw's configuration when the session opened.
func (s *Service) relayConfig() agent.RelayConfig {
	cfg := s.config.Get()
	return agent.RelayConfig{
		ReadTimeout:     time.Duration(cfg.ReadTimeoutMs) * time.Millisecond,
		WriteTimeout:    time.Duration(cfg.WriteTimeoutMs) * time.Millisecond,
		PingInterval:    time.Duration(cfg.PingIntervalMs) * time.Millisecond,
		MaxMessageBytes: int64(cfg.MaxMessageBytes),
	}
}
