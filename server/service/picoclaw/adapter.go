package picoclaw

import (
	"context"

	"NanoKVM-Server/service/agent"

	log "github.com/sirupsen/logrus"
)

// Adapter is PicoClaw behind the agent interface: the pico channel for chat,
// PicoClaw's session files for history and the S96picoclaw script for the
// runtime.
type Adapter struct {
	service *Service
}

var _ agent.Agent = (*Adapter)(nil)

// Agent returns the service's PicoClaw adapter.
func (s *Service) Agent() *Adapter {
	return &Adapter{service: s}
}

func (a *Adapter) Name() string { return "picoclaw" }

func (a *Adapter) Start(ctx context.Context) error {
	a.service.ensureDependencies()
	_, _, err := a.service.startRuntimeContext(ctx)
	return agentError(err)
}

func (a *Adapter) Stop(context.Context) error {
	return a.service.stopRuntimeAndVerify(false)
}

func (a *Adapter) Ready(ctx context.Context) error {
	return agentError(a.service.ensureRuntimeReadyWithProbeProtection(ctx, false))
}

func (a *Adapter) ApplyConfig(context.Context) error {
	return applyPicoclawConfig()
}

// OpenSession connects to the pico channel for the session, with the token
// and limits from PicoClaw's current configuration.
func (a *Adapter) OpenSession(_ context.Context, sessionID string) (agent.Session, error) {
	s := a.service
	s.ensureDependencies()
	if err := s.syncConfigFromPicoclaw(); err != nil {
		return nil, agentError(err)
	}
	conn, err := s.connectGateway(sessionID)
	if err != nil {
		return nil, agentError(err)
	}
	return newPicoSession(sessionID, conn, s.config.Get()), nil
}

// applyPicoclawConfig writes the settings the server enforces into
// PicoClaw's config.json and .security.yml and refreshes the workspace
// AGENT.md. A stale AGENT.md only costs the agent newer instructions, so a
// failure there is logged and not returned.
func applyPicoclawConfig() error {
	if err := ensurePicoclawStartupDefaults(); err != nil {
		return err
	}
	if _, err := refreshPicoclawAgentProfile(); err != nil {
		log.Warnf("picoclaw: failed to refresh the workspace AGENT.md: %v", err)
	}
	return nil
}

// agentError returns err as an error, nil when err is nil.
func agentError(err *PicoclawError) error {
	if err == nil {
		return nil
	}
	return &agent.Error{Status: err.StatusCode, Code: err.Code, Message: err.Message}
}
