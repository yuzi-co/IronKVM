package picoclaw

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	defaultPicoclawGatewayHost = "127.0.0.1"
	defaultPicoclawGatewayPort = 18790
	picoclawGatewayPath        = "/pico/ws"
)

func defaultConfig() Config {
	cfg := Config{
		GatewayURL:       "ws://127.0.0.1:18790/pico/ws",
		ConnectTimeoutMs: 10000,
		ReadTimeoutMs:    60000,
		WriteTimeoutMs:   10000,
		PingIntervalMs:   30000,
		MaxMessageBytes:  1024 * 1024,
		AllowTokenQuery:  false,
	}

	settings, err := loadPicoclawGatewaySettings()
	if err != nil {
		return cfg
	}

	cfg.GatewayURL = settings.GatewayURL
	cfg.Token = settings.Token
	cfg.AllowTokenQuery = settings.AllowTokenQuery
	if settings.PingIntervalMs > 0 {
		cfg.PingIntervalMs = settings.PingIntervalMs
	}
	if settings.ReadTimeoutMs > 0 {
		cfg.ReadTimeoutMs = settings.ReadTimeoutMs
	}

	return cfg
}

func (s *Service) syncConfigFromPicoclaw() *PicoclawError {
	installed, installedKnown := picoclawInstalledState()

	if patchErr := preparePicoclawConfigForRead(); patchErr != nil {
		s.runtime.Update(func(status *RuntimeStatus) {
			status.Ready = false
			status.Status = "config_error"
			status.ConfigError = patchErr.Error()
			status.LastError = patchErr.Error()
			if installedKnown {
				status.Installed = installed
			}
			status.CheckedAt = time.Now()
		})
		return newPicoclawError(CodeRuntimeUnavailable, patchErr.Error())
	}

	settings, err := loadPicoclawGatewaySettings()
	if err != nil {
		s.runtime.Update(func(status *RuntimeStatus) {
			status.Ready = false
			status.Status = "config_error"
			status.ConfigError = err.Error()
			status.LastError = err.Error()
			if installedKnown {
				status.Installed = installed
			}
			status.CheckedAt = time.Now()
		})
		return newPicoclawError(CodeRuntimeUnavailable, err.Error())
	}

	cfg := s.config.Get()
	cfg.GatewayURL = settings.GatewayURL
	cfg.Token = settings.Token
	cfg.AllowTokenQuery = settings.AllowTokenQuery
	if settings.PingIntervalMs > 0 {
		cfg.PingIntervalMs = settings.PingIntervalMs
	}
	if settings.ReadTimeoutMs > 0 {
		cfg.ReadTimeoutMs = settings.ReadTimeoutMs
	}
	s.config.Set(cfg)

	s.runtime.Update(func(status *RuntimeStatus) {
		status.Installed = true
		status.ModelConfigured = settings.ModelConfigured
		if settings.ModelConfigured {
			status.ModelName = settings.ModelName
		} else {
			status.ModelName = settings.TargetModelName
		}
		status.ConfigError = ""
		if status.Status == "config_error" && status.Ready {
			status.Status = "ready"
		}
		status.CheckedAt = time.Now()
	})

	return nil
}

// syncRuntimeConfigMetadataFromPicoclaw refreshes the model metadata exposed
// by RuntimeStatus without changing PicoClaw's config or probing its gateway.
// This is used while PicoClaw does not own the control mode.
func (s *Service) syncRuntimeConfigMetadataFromPicoclaw() *PicoclawError {
	doc, err := loadPicoclawConfigDocument()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.runtime.Update(func(status *RuntimeStatus) {
				status.ModelConfigured = false
				status.ModelName = ""
				status.ConfigError = ""
				if status.Status == "config_error" {
					status.Status = "checking"
					status.LastError = ""
				}
				status.CheckedAt = time.Now()
			})
			return nil
		}

		s.runtime.Update(func(status *RuntimeStatus) {
			status.ModelConfigured = false
			status.ModelName = ""
			status.Status = "config_error"
			status.ConfigError = err.Error()
			status.LastError = err.Error()
			status.CheckedAt = time.Now()
		})
		return newPicoclawError(CodeRuntimeUnavailable, err.Error())
	}

	modelName := resolvePicoclawTargetModelName(doc.config)
	modelConfigured := isPicoclawModelConfigured(doc.config, doc.security, modelName)
	s.runtime.Update(func(status *RuntimeStatus) {
		status.ModelConfigured = modelConfigured
		status.ModelName = modelName
		status.ConfigError = ""
		if status.Status == "config_error" || (modelConfigured && status.Status == "model_not_configured") {
			status.Status = "checking"
			status.LastError = ""
		}
		status.CheckedAt = time.Now()
	})

	return nil
}

type picoclawGatewaySettings struct {
	GatewayURL      string
	Token           string
	AllowTokenQuery bool
	PingIntervalMs  int
	ReadTimeoutMs   int
	ModelConfigured bool
	ModelName       string
	TargetModelName string
}

// picoclawFileStamp identifies one version of a file by modification time and
// size. A missing file has the zero stamp with exists false.
type picoclawFileStamp struct {
	exists  bool
	modTime int64
	size    int64
}

func statPicoclawFile(path string) (picoclawFileStamp, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return picoclawFileStamp{}, nil
		}
		return picoclawFileStamp{}, err
	}
	return picoclawFileStamp{exists: true, modTime: info.ModTime().UnixNano(), size: info.Size()}, nil
}

type picoclawSettingsCacheKey struct {
	configPath string
	config     picoclawFileStamp
	security   picoclawFileStamp
}

// picoclawSettingsCache keeps the gateway settings parsed from config.json and
// .security.yml. The status probe asks for them every few seconds and the
// files rarely change, so they are parsed again only when either file's
// modification time or size differs, or after this server wrote one of them.
var picoclawSettingsCache struct {
	mu         sync.Mutex
	valid      bool
	generation uint64
	key        picoclawSettingsCacheKey
	settings   picoclawGatewaySettings
}

func invalidatePicoclawSettingsCache() {
	picoclawSettingsCache.mu.Lock()
	picoclawSettingsCache.valid = false
	picoclawSettingsCache.generation++
	picoclawSettingsCache.mu.Unlock()
}

func picoclawSettingsKey() (picoclawSettingsCacheKey, error) {
	configPath, err := resolvePicoclawConfigPath()
	if err != nil {
		return picoclawSettingsCacheKey{}, err
	}
	configStamp, err := statPicoclawFile(configPath)
	if err != nil {
		return picoclawSettingsCacheKey{}, err
	}
	securityStamp, err := statPicoclawFile(resolvePicoclawSecurityPath(configPath))
	if err != nil {
		return picoclawSettingsCacheKey{}, err
	}
	return picoclawSettingsCacheKey{configPath: configPath, config: configStamp, security: securityStamp}, nil
}

func loadPicoclawGatewaySettings() (picoclawGatewaySettings, error) {
	key, err := picoclawSettingsKey()
	if err != nil || !key.config.exists {
		// Let the uncached path produce the usual error.
		return loadPicoclawGatewaySettingsUncached()
	}

	picoclawSettingsCache.mu.Lock()
	if picoclawSettingsCache.valid && picoclawSettingsCache.key == key {
		settings := picoclawSettingsCache.settings
		picoclawSettingsCache.mu.Unlock()
		return settings, nil
	}
	generation := picoclawSettingsCache.generation
	picoclawSettingsCache.mu.Unlock()

	// Parse without the lock: loading can enable the pico channel and save
	// config.json, which invalidates the cache.
	settings, err := loadPicoclawGatewaySettingsUncached()
	if err != nil {
		return settings, err
	}

	picoclawSettingsCache.mu.Lock()
	defer picoclawSettingsCache.mu.Unlock()
	// A write during the parse (ours or another goroutine's) bumps the
	// generation; then the result is returned but not kept. A file changed
	// after the stat above only makes the stored key stale, so the next call
	// parses again.
	if picoclawSettingsCache.generation == generation {
		picoclawSettingsCache.valid = true
		picoclawSettingsCache.key = key
		picoclawSettingsCache.settings = settings
	}
	return settings, nil
}

func loadPicoclawGatewaySettingsUncached() (picoclawGatewaySettings, error) {
	doc, err := loadPicoclawConfigDocument()
	if err != nil {
		return picoclawGatewaySettings{}, err
	}
	if err := ensurePicoclawPicoChannelEnabled(doc); err != nil {
		return picoclawGatewaySettings{}, err
	}

	cfg := doc.config

	host := cfg.Gateway.Host
	if host == "" || host == "0.0.0.0" {
		host = defaultPicoclawGatewayHost
	}

	port := cfg.Gateway.Port
	if port <= 0 {
		port = defaultPicoclawGatewayPort
	}

	picoSettings := picoclawPicoSettingsV3{}
	if pico, ok := cfg.Channels["pico"]; ok {
		picoSettings = pico.Settings
	}

	settings := picoclawGatewaySettings{
		GatewayURL:      fmt.Sprintf("ws://%s:%d%s", host, port, picoclawGatewayPath),
		Token:           doc.resolvedGatewayToken(),
		AllowTokenQuery: picoSettings.AllowTokenQuery,
	}
	settings.TargetModelName = resolvePicoclawTargetModelName(cfg)

	if isPicoclawModelConfigured(cfg, doc.security, settings.TargetModelName) {
		settings.ModelConfigured = true
		settings.ModelName = settings.TargetModelName
	}

	if picoSettings.PingInterval > 0 {
		settings.PingIntervalMs = picoSettings.PingInterval * 1000
	}
	if picoSettings.ReadTimeout > 0 {
		settings.ReadTimeoutMs = picoSettings.ReadTimeout * 1000
	}

	return settings, nil
}

func resolvePicoclawTargetModelName(cfg picoclawConfigFile) string {
	return strings.TrimSpace(cfg.Agents.Defaults.ModelName)
}

func preparePicoclawConfigForRead() error {
	configPath, err := resolvePicoclawConfigPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to stat picoclaw config: %w", err)
	}

	return nil
}
