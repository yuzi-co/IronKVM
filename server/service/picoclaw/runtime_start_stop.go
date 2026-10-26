package picoclaw

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// launchRuntime runs the S96picoclaw script's start after checking the
// install and writing the enforced settings. The lifecycle waits for the
// gateway to be ready.
func (s *Service) launchRuntime(ctx context.Context) (string, string, *PicoclawError) {
	if ctx == nil {
		ctx = context.Background()
	}
	if s == nil {
		return "", "", newPicoclawError(CodeRuntimeStartFailed, "picoclaw service is unavailable")
	}
	s.ensureDependencies()

	if installed, statErr := isPicoclawInstalled(); statErr != nil {
		s.runtime.Update(func(status *RuntimeStatus) {
			status.Ready = false
			status.Installed = false
			status.ModelConfigured = false
			status.ModelName = ""
			status.Status = "not_installed"
			status.LastError = statErr.Error()
			status.CheckedAt = time.Now()
		})
		return "", "", newPicoclawError(CodeRuntimeStartFailed, statErr.Error())
	} else if !installed {
		s.runtime.Update(func(status *RuntimeStatus) {
			status.Ready = false
			status.Installed = false
			status.ModelConfigured = false
			status.ModelName = ""
			status.Status = "not_installed"
			status.LastError = "picoclaw is not installed"
			status.CheckedAt = time.Now()
		})
		return "", "", newPicoclawError(CodeRuntimeStartFailed, "picoclaw is not installed")
	}

	scriptPath, err := resolvePicoclawStartScript()
	if err != nil {
		s.runtime.Update(func(status *RuntimeStatus) {
			status.Ready = false
			status.Installed = true
			status.Status = "unavailable"
			status.LastError = err.Error()
			status.CheckedAt = time.Now()
		})
		return "", "", newPicoclawError(CodeRuntimeStartFailed, err.Error())
	}
	if err := applyPicoclawConfig(); err != nil {
		s.runtime.Update(func(status *RuntimeStatus) {
			status.Ready = false
			status.Installed = true
			status.Status = "config_error"
			status.ConfigError = err.Error()
			status.LastError = err.Error()
			status.CheckedAt = time.Now()
		})
		return "", "", newPicoclawError(CodeRuntimeStartFailed, err.Error())
	}
	if err := s.detectGatewayPortConflict(); err != nil {
		s.runtime.Update(func(status *RuntimeStatus) {
			status.Ready = false
			status.Installed = true
			status.Status = "unavailable"
			status.LastError = err.Error()
			status.CheckedAt = time.Now()
		})
		return "", "", newPicoclawError(CodeRuntimeStartFailed, err.Error())
	}

	command := scriptPath + " start"
	startCtx, cancel := context.WithTimeout(ctx, picoclawStartTimeout)
	defer cancel()

	output, execErr := exec.CommandContext(startCtx, "sh", "-c", command).CombinedOutput()
	trimmedOutput := strings.TrimSpace(string(output))
	if execErr != nil {
		if lifecycleErr := s.operationError(ctx); lifecycleErr != nil {
			return command, trimmedOutput, lifecycleErr
		}
		s.runtime.Update(func(status *RuntimeStatus) {
			status.Ready = false
			status.Installed = true
			status.Status = "unavailable"
			status.LastError = trimmedOutput
			if status.LastError == "" {
				status.LastError = execErr.Error()
			}
			status.CheckedAt = time.Now()
		})
		return command, trimmedOutput, newPicoclawError(CodeRuntimeStartFailed, "failed to start picoclaw runtime")
	}

	return command, trimmedOutput, nil
}

// checkRuntime checks that PicoClaw is installed, has a config (running its
// onboarding when it has none) with a configured model, runs, and answers on
// its gateway. It returns the status to record, nil when there is none to
// record, and why PicoClaw cannot take a chat session.
func (s *Service) checkRuntime(ctx context.Context) (*RuntimeStatus, *PicoclawError) {
	if s == nil {
		return nil, newPicoclawError(CodeRuntimeUnavailable, "picoclaw service is unavailable")
	}
	s.ensureDependencies()

	installed, statErr := isPicoclawInstalled()
	if statErr != nil {
		return &RuntimeStatus{
			Ready:           false,
			Installed:       false,
			Installing:      false,
			InstallProgress: 0,
			Status:          "not_installed",
			LastError:       statErr.Error(),
		}, newPicoclawError(CodeRuntimeUnavailable, "failed to check picoclaw installation")
	}
	if !installed {
		return &RuntimeStatus{
			Ready:           false,
			Installed:       false,
			Installing:      false,
			InstallProgress: 0,
			ModelConfigured: false,
			Status:          "not_installed",
		}, newPicoclawError(CodeRuntimeUnavailable, "picoclaw is not installed")
	}

	configPath, pathErr := resolvePicoclawConfigPath()
	if pathErr != nil {
		return &RuntimeStatus{
			Ready:           false,
			Installed:       true,
			Installing:      false,
			InstallProgress: 0,
			ModelConfigured: false,
			Status:          "model_not_configured",
			LastError:       pathErr.Error(),
		}, newPicoclawError(CodeRuntimeUnavailable, "picoclaw model is not configured")
	}
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			if lifecycleErr := s.operationError(ctx); lifecycleErr != nil {
				return nil, lifecycleErr
			}
			if _, onboardErr := s.runPicoclawOnboardContext(ctx); onboardErr != nil {
				return &RuntimeStatus{
					Ready:           false,
					Installed:       true,
					Installing:      false,
					InstallProgress: 0,
					ModelConfigured: false,
					Status:          "model_not_configured",
					LastError:       onboardErr.Message,
				}, newPicoclawError(CodeRuntimeUnavailable, "picoclaw model is not configured")
			}
			if lifecycleErr := s.operationError(ctx); lifecycleErr != nil {
				return nil, lifecycleErr
			}
			if _, statErr := os.Stat(configPath); statErr == nil {
				goto configReady
			}
			return &RuntimeStatus{
				Ready:           false,
				Installed:       true,
				Installing:      false,
				InstallProgress: 0,
				ModelConfigured: false,
				Status:          "model_not_configured",
			}, newPicoclawError(CodeRuntimeUnavailable, "picoclaw model is not configured")
		}
		return &RuntimeStatus{
			Ready:           false,
			Installed:       true,
			Installing:      false,
			InstallProgress: 0,
			ModelConfigured: false,
			Status:          "config_error",
			ConfigError:     err.Error(),
			LastError:       err.Error(),
		}, newPicoclawError(CodeRuntimeUnavailable, "gateway config is invalid")
	}

configReady:
	if syncErr := s.syncConfigFromPicoclaw(); syncErr != nil {
		return nil, syncErr
	}
	settings, settingsErr := loadPicoclawGatewaySettings()
	if settingsErr != nil {
		return nil, newPicoclawError(CodeRuntimeUnavailable, settingsErr.Error())
	}
	if !settings.ModelConfigured {
		return &RuntimeStatus{
			Ready:           false,
			Installed:       true,
			Installing:      false,
			InstallProgress: 0,
			ModelConfigured: false,
			ModelName:       settings.TargetModelName,
			Status:          "model_not_configured",
		}, newPicoclawError(CodeRuntimeUnavailable, "picoclaw model is not configured")
	}

	running, runningErr := isRuntimeRunning()
	if runningErr != nil {
		return &RuntimeStatus{
			Ready:           false,
			Installed:       true,
			Installing:      false,
			InstallProgress: 0,
			ModelConfigured: true,
			ModelName:       settings.ModelName,
			Status:          "unavailable",
			LastError:       runningErr.Error(),
		}, newPicoclawError(CodeRuntimeUnavailable, "failed to check picoclaw runtime")
	}
	if !running {
		return &RuntimeStatus{
			Ready:           false,
			Installed:       true,
			Installing:      false,
			InstallProgress: 0,
			ModelConfigured: true,
			ModelName:       settings.ModelName,
			Status:          "stopped",
		}, newPicoclawError(CodeRuntimeUnavailable, "picoclaw runtime is stopped")
	}

	cfg := s.config.Get()
	if probeErr := probePicoclawGateway(cfg); probeErr != nil {
		return &RuntimeStatus{
			Ready:           false,
			Installed:       true,
			Installing:      false,
			InstallProgress: 0,
			ModelConfigured: true,
			ModelName:       settings.ModelName,
			Status:          probeErr.status,
			ConfigError:     probeErr.configError,
			LastError:       probeErr.lastError,
		}, newPicoclawError(CodeRuntimeUnavailable, probeErr.message)
	}

	return &RuntimeStatus{
		Ready:           true,
		Installed:       true,
		Installing:      false,
		InstallProgress: 100,
		ModelConfigured: true,
		ModelName:       settings.ModelName,
		Status:          "ready",
	}, nil
}

func (s *Service) stopRuntime() (string, string, *PicoclawError) {
	if s == nil {
		return "", "", newPicoclawError(CodeRuntimeStartFailed, "picoclaw service is unavailable")
	}
	s.ensureDependencies()

	settings, _ := loadPicoclawGatewaySettings()
	scriptPath, err := resolvePicoclawStartScript()
	if err != nil {
		s.runtime.Set(RuntimeStatus{
			Ready:           false,
			Installed:       true,
			InstallPath:     picoclawBinaryPath,
			ModelConfigured: settings.ModelConfigured,
			ModelName:       settings.ModelName,
			Status:          "unavailable",
			LastError:       err.Error(),
			CheckedAt:       time.Now(),
		})
		return "", "", newPicoclawError(CodeRuntimeStartFailed, err.Error())
	}

	command := scriptPath + " stop"
	ctx, cancel := context.WithTimeout(context.Background(), picoclawStopTimeout)
	defer cancel()

	output, execErr := exec.CommandContext(ctx, "sh", "-c", command).CombinedOutput()
	trimmedOutput := strings.TrimSpace(string(output))
	if execErr != nil {
		status := RuntimeStatus{
			Ready:           false,
			Installed:       true,
			InstallPath:     picoclawBinaryPath,
			ModelConfigured: settings.ModelConfigured,
			ModelName:       settings.ModelName,
			Status:          "error",
			LastError:       trimmedOutput,
			CheckedAt:       time.Now(),
		}
		if trimmedOutput == "" {
			status.LastError = execErr.Error()
		}
		s.runtime.Set(status)
		return command, trimmedOutput, newPicoclawError(CodeRuntimeStartFailed, "failed to stop picoclaw runtime")
	}

	time.Sleep(picoclawStopWaitPeriod)
	s.runtime.Set(RuntimeStatus{
		Ready:           false,
		Installed:       true,
		InstallPath:     picoclawBinaryPath,
		ModelConfigured: settings.ModelConfigured,
		ModelName:       settings.ModelName,
		Status:          "stopped",
		CheckedAt:       time.Now(),
		CurrentSession:  "",
	})

	return command, trimmedOutput, nil
}

func resolvePicoclawStartScript() (string, error) {
	if _, err := os.Stat(kvmappPicoclawScript); err == nil {
		return kvmappPicoclawScript, nil
	}
	if _, err := os.Stat(etcInitPicoclawScript); err == nil {
		return etcInitPicoclawScript, nil
	}
	return "", fmt.Errorf("picoclaw start script not found: %s or %s", etcInitPicoclawScript, kvmappPicoclawScript)
}

func runPicoclawOnboard() (string, *PicoclawError) {
	return runPicoclawOnboardWith(context.Background(), func(context.Context) *PicoclawError { return nil })
}

// runPicoclawOnboardContext runs the onboarding for a check, which the
// lifecycle may cancel.
func (s *Service) runPicoclawOnboardContext(ctx context.Context) (string, *PicoclawError) {
	return runPicoclawOnboardWith(ctx, s.operationError)
}

func runPicoclawOnboardWith(ctx context.Context, operationError func(context.Context) *PicoclawError) (string, *PicoclawError) {
	if ctx == nil {
		ctx = context.Background()
	}
	scriptPath, err := resolvePicoclawStartScript()
	if err != nil {
		return "", newPicoclawError(CodeRuntimeUnavailable, err.Error())
	}
	if lifecycleErr := operationError(ctx); lifecycleErr != nil {
		return "", lifecycleErr
	}

	command := scriptPath + " onboard"
	onboardCtx, cancel := context.WithTimeout(ctx, picoclawOnboardTimeout)
	defer cancel()

	output, execErr := exec.CommandContext(onboardCtx, "sh", "-c", command).CombinedOutput()
	trimmedOutput := strings.TrimSpace(string(output))
	if execErr != nil {
		if lifecycleErr := operationError(ctx); lifecycleErr != nil {
			return trimmedOutput, lifecycleErr
		}
		if trimmedOutput == "" {
			trimmedOutput = execErr.Error()
		}
		return trimmedOutput, newPicoclawError(CodeRuntimeUnavailable, "failed to initialize picoclaw config")
	}

	return trimmedOutput, nil
}

func isPicoclawInstalled() (bool, error) {
	info, err := os.Stat(picoclawBinaryPath)
	if err == nil {
		return !info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func isRuntimeRunning() (bool, error) {
	pid, err := runtimeGatewayPID()
	if err != nil {
		return false, err
	}
	return pid > 0, nil
}

func (s *Service) detectGatewayPortConflict() error {
	running, err := isRuntimeRunning()
	if err != nil {
		return fmt.Errorf("check PicoClaw gateway process: %w", err)
	}
	if running {
		return nil
	}

	settings, err := loadPicoclawGatewaySettings()
	if err != nil {
		return nil
	}
	parsed, err := url.Parse(settings.GatewayURL)
	if err != nil {
		return nil
	}
	hostPort, err := gatewayHostPort(parsed)
	if err != nil {
		return nil
	}

	timeout := 2 * time.Second
	if s != nil && s.config != nil {
		if configured := time.Duration(s.config.Get().ConnectTimeoutMs) * time.Millisecond; configured > 0 {
			timeout = configured
		}
	}
	conn, err := net.DialTimeout("tcp", hostPort, timeout)
	if err != nil {
		return nil
	}
	_ = conn.Close()

	return fmt.Errorf("gateway port %s is already in use by another process", hostPort)
}

func resolvePicoclawPIDPath() (string, error) {
	home, err := resolvePicoclawHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, picoclawPIDFileName), nil
}

func runtimeGatewayPID() (int, error) {
	pidPath, err := resolvePicoclawPIDPath()
	if err != nil {
		return 0, err
	}

	pid, err := readRuntimePIDFile(pidPath)
	if err == nil {
		running, processErr := isPicoclawGatewayProcess(pid)
		if processErr != nil {
			return 0, processErr
		}
		if running {
			return pid, nil
		}
		_ = os.Remove(pidPath)
	} else if !os.IsNotExist(err) {
		_ = os.Remove(pidPath)
	}

	pid, err = findPicoclawGatewayProcess()
	if err != nil {
		return 0, err
	}
	if pid > 0 {
		_ = os.MkdirAll(filepath.Dir(pidPath), 0o755)
		_ = os.WriteFile(pidPath, []byte(strconv.Itoa(pid)+"\n"), 0o600)
	}
	return pid, nil
}

func readRuntimePIDFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	pid, parseErr := strconv.Atoi(strings.TrimSpace(string(data)))
	if parseErr != nil || pid <= 0 {
		return 0, fmt.Errorf("invalid PicoClaw pid file: %s", path)
	}
	return pid, nil
}

func isPicoclawGatewayProcess(pid int) (bool, error) {
	if pid <= 0 {
		return false, nil
	}
	cmdlinePath := filepath.Join("/proc", strconv.Itoa(pid), "cmdline")
	data, err := os.ReadFile(cmdlinePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return isPicoclawGatewayCmdline(data), nil
}

func isPicoclawGatewayCmdline(data []byte) bool {
	parts := strings.Split(string(data), "\x00")
	args := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			args = append(args, part)
		}
	}
	if len(args) < 2 {
		return false
	}

	binName := filepath.Base(picoclawBinaryPath)
	if binName == "" || binName == "." || binName == string(filepath.Separator) {
		return false
	}

	return filepath.Base(args[0]) == binName && args[1] == "gateway"
}

func findPicoclawGatewayProcess() (int, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 {
			continue
		}
		running, err := isPicoclawGatewayProcess(pid)
		if err != nil {
			continue
		}
		if running {
			return pid, nil
		}
	}

	return 0, nil
}
