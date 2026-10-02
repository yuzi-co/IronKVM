package picoclaw

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func gatewayConfigFor(server *httptest.Server) Config {
	return Config{
		GatewayURL:       "ws://" + strings.TrimPrefix(server.URL, "http://") + picoclawGatewayPath,
		ConnectTimeoutMs: 1000,
	}
}

func TestProbePicoclawGatewayUsesReadyEndpoint(t *testing.T) {
	var readyHits, wsHits atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		readyHits.Add(1)
		if r.Header.Get("Authorization") != "" {
			t.Errorf("readiness probe sent credentials")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	})
	mux.HandleFunc(picoclawGatewayPath, func(w http.ResponseWriter, r *http.Request) {
		wsHits.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	cfg := gatewayConfigFor(server)
	cfg.Token = "placeholder"
	if probeErr := probePicoclawGateway(cfg); probeErr != nil {
		t.Fatalf("probe error = %+v", probeErr)
	}
	if readyHits.Load() != 1 || wsHits.Load() != 0 {
		t.Fatalf("ready hits = %d, websocket hits = %d; want 1 and 0", readyHits.Load(), wsHits.Load())
	}
}

func TestProbePicoclawGatewayReportsNotReady(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	probeErr := probePicoclawGateway(gatewayConfigFor(server))
	if probeErr == nil || probeErr.status != "unavailable" || probeErr.message != "gateway is not ready" {
		t.Fatalf("probe error = %+v, want not ready", probeErr)
	}
}

func TestProbePicoclawGatewayFallsBackToWebSocketWithoutReadyEndpoint(t *testing.T) {
	var wsHits atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc(picoclawGatewayPath, func(w http.ResponseWriter, r *http.Request) {
		wsHits.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	probeErr := probePicoclawGateway(gatewayConfigFor(server))
	if wsHits.Load() != 1 {
		t.Fatalf("websocket hits = %d, want 1", wsHits.Load())
	}
	if probeErr == nil || probeErr.status != "config_error" || probeErr.configError != "gateway authentication failed" {
		t.Fatalf("probe error = %+v, want authentication failure from the websocket fallback", probeErr)
	}
}

func TestProbePicoclawGatewayReportsUnreachableGateway(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	cfg := gatewayConfigFor(server)
	server.Close()

	probeErr := probePicoclawGateway(cfg)
	if probeErr == nil || probeErr.status != "unavailable" || probeErr.message != "gateway is unavailable" {
		t.Fatalf("probe error = %+v, want unavailable", probeErr)
	}
}

func TestBuildGatewayReadyURL(t *testing.T) {
	for input, want := range map[string]string{
		"ws://127.0.0.1:18790/pico/ws?x=1": "http://127.0.0.1:18790/ready",
		"wss://example.invalid/pico/ws":    "https://example.invalid/ready",
	} {
		got, err := buildGatewayReadyURL(Config{GatewayURL: input})
		if err != nil || got != want {
			t.Fatalf("buildGatewayReadyURL(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
	if _, err := buildGatewayReadyURL(Config{GatewayURL: "http://127.0.0.1:18790/pico/ws"}); err == nil {
		t.Fatal("expected an error for a non-websocket scheme")
	}
}

const cachedSettingsConfig = `{
  "agents": {"defaults": {"model_name": "m"}},
  "gateway": {"host": "127.0.0.1", "port": 18790},
  "model_list": [{"model_name": "m", "model": "ollama/m", "api_base": "http://127.0.0.1:11434/v1"}],
  "channel_list": {"pico": {"type": "pico", "enabled": true, "settings": {}}}
}`

func writeConfigKeepingStamp(t *testing.T, path string, data string, stamp time.Time) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, stamp, stamp); err != nil {
		t.Fatal(err)
	}
}

func TestPicoclawGatewaySettingsCachedByModTimeAndSize(t *testing.T) {
	home := t.TempDir()
	t.Setenv("PICOCLAW_HOME", home)
	invalidatePicoclawSettingsCache()
	configPath := filepath.Join(home, "config.json")
	stamp := time.Now().Add(-time.Hour).Truncate(time.Second)
	writeConfigKeepingStamp(t, configPath, cachedSettingsConfig, stamp)

	settings, err := loadPicoclawGatewaySettings()
	if err != nil {
		t.Fatal(err)
	}
	if settings.GatewayURL != "ws://127.0.0.1:18790/pico/ws" || !settings.ModelConfigured {
		t.Fatalf("settings = %+v", settings)
	}

	// Same size and modification time: the cached settings are used and the
	// (now unparsable) file is not read again.
	writeConfigKeepingStamp(t, configPath, strings.Repeat(" ", len(cachedSettingsConfig)), stamp)
	if _, err := loadPicoclawGatewaySettings(); err != nil {
		t.Fatalf("cached load reparsed the file: %v", err)
	}

	// A different modification time makes it parse again.
	writeConfigKeepingStamp(t, configPath, strings.Repeat(" ", len(cachedSettingsConfig)), stamp.Add(time.Second))
	if _, err := loadPicoclawGatewaySettings(); err == nil {
		t.Fatal("changed file was not parsed again")
	}

	// A config edit that keeps size and time is still noticed when
	// .security.yml changes, since both files are part of the key.
	writeConfigKeepingStamp(t, configPath, cachedSettingsConfig, stamp)
	if _, err := loadPicoclawGatewaySettings(); err != nil {
		t.Fatal(err)
	}
	writeConfigKeepingStamp(t, configPath, strings.Replace(cachedSettingsConfig, "18790", "18791", 1), stamp)
	if err := os.WriteFile(filepath.Join(home, ".security.yml"), []byte("model_list: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	settings, err = loadPicoclawGatewaySettings()
	if err != nil {
		t.Fatal(err)
	}
	if settings.GatewayURL != "ws://127.0.0.1:18791/pico/ws" {
		t.Fatalf("gateway url = %q, want the new port", settings.GatewayURL)
	}
}

func TestPicoclawGatewaySettingsCacheInvalidatedBySave(t *testing.T) {
	home := t.TempDir()
	t.Setenv("PICOCLAW_HOME", home)
	invalidatePicoclawSettingsCache()
	configPath := filepath.Join(home, "config.json")
	stamp := time.Now().Add(-time.Hour).Truncate(time.Second)
	writeConfigKeepingStamp(t, configPath, cachedSettingsConfig, stamp)

	if _, err := loadPicoclawGatewaySettings(); err != nil {
		t.Fatal(err)
	}
	doc, err := loadPicoclawConfigDocument()
	if err != nil {
		t.Fatal(err)
	}
	doc.raw["gateway"] = map[string]any{"host": "127.0.0.1", "port": 18792}
	if err := doc.saveConfig(); err != nil {
		t.Fatal(err)
	}
	// Put the old stamp back so the result depends on the save dropping the
	// cache, not only on the new modification time.
	if err := os.Chtimes(configPath, stamp, stamp); err != nil {
		t.Fatal(err)
	}

	settings, err := loadPicoclawGatewaySettings()
	if err != nil {
		t.Fatal(err)
	}
	if settings.GatewayURL != "ws://127.0.0.1:18792/pico/ws" {
		t.Fatalf("gateway url = %q, want the saved port", settings.GatewayURL)
	}
}
