package picoclaw

import (
	"context"

	"NanoKVM-Server/service/agent"

	log "github.com/sirupsen/logrus"
)

// Adapter is PicoClaw behind the agent interface: the pico channel for chat,
// PicoClaw's session files for history, and for the runtime the pinned
// release, the S96picoclaw script and the gateway's /ready probe.
type Adapter struct {
	service *Service
}

var _ agent.Agent = (*Adapter)(nil)

// Agent returns the service's PicoClaw adapter.
func (s *Service) Agent() *Adapter {
	return &Adapter{service: s}
}

func (a *Adapter) Name() string { return "picoclaw" }

func (a *Adapter) InstallPath() string { return picoclawBinaryPath }

func (a *Adapter) Version() string { return Version }

func (a *Adapter) Installed() (bool, error) { return isPicoclawInstalled() }

func (a *Adapter) Stale() bool { return installedPicoclawIsStale() }

func (a *Adapter) Install(progress func(stage string, percent int, lastError string)) error {
	return installReleasePicoclaw(progress)
}

func (a *Adapter) Uninstall() { removePicoclaw() }

// Model reads the model settings from PicoClaw's configuration. The fields
// are what the configuration gave even when err is set.
func (a *Adapter) Model() (agent.Model, error) {
	settings, err := loadPicoclawGatewaySettings()
	return agent.Model{
		Configured: settings.ModelConfigured,
		Name:       settings.ModelName,
		Target:     settings.TargetModelName,
	}, err
}

func (a *Adapter) ApplyConfig(context.Context) error {
	return applyPicoclawConfig()
}

func (a *Adapter) Running() (bool, error) { return isRuntimeRunning() }

func (a *Adapter) Launch(ctx context.Context) (string, string, error) {
	command, output, err := a.service.launchRuntime(ctx)
	return command, output, agentError(err)
}

// Halt runs the S96picoclaw script's stop. Its error keeps the PicoClaw
// error's text, which leads with the code.
func (a *Adapter) Halt() (string, string, error) {
	command, output, err := a.service.stopRuntime()
	if err != nil {
		return command, output, err
	}
	return command, output, nil
}

func (a *Adapter) Check(ctx context.Context) (*agent.RuntimeStatus, error) {
	status, err := a.service.checkRuntime(ctx)
	return status, agentError(err)
}

func (a *Adapter) Describe(status agent.RuntimeStatus) agent.RuntimeStatus {
	return withAgentProfile(status)
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
	return err.AgentError()
}
