package picoclaw

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type gatewayProbeError struct {
	status      string
	configError string
	lastError   string
	message     string
}

func (s *Service) connectGateway(sessionID string) (*websocket.Conn, *PicoclawError) {
	cfg := s.config.Get()

	gatewayURL, err := buildGatewayURL(cfg, sessionID)
	if err != nil {
		s.runtime.Update(func(status *RuntimeStatus) {
			status.Ready = false
			status.Status = "config_error"
			status.ConfigError = err.Error()
			status.LastError = err.Error()
			status.CheckedAt = time.Now()
			status.CurrentSession = s.lock.Owner()
		})
		return nil, newPicoclawError(CodeRuntimeUnavailable, err.Error())
	}
	header := http.Header{}
	if cfg.Token != "" {
		header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.Token))
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: time.Duration(cfg.ConnectTimeoutMs) * time.Millisecond,
		NetDialContext: (&net.Dialer{
			Timeout: time.Duration(cfg.ConnectTimeoutMs) * time.Millisecond,
		}).DialContext,
	}

	upstream, response, err := dialer.Dial(gatewayURL, header)
	if err != nil {
		s.runtime.Update(func(status *RuntimeStatus) {
			status.Ready = false
			status.Status = "unavailable"
			status.ConfigError = ""
			status.LastError = err.Error()
			status.CheckedAt = time.Now()
			status.CurrentSession = s.lock.Owner()
			if response != nil && response.StatusCode == http.StatusUnauthorized {
				status.Status = "config_error"
				status.ConfigError = "gateway authentication failed"
			}
		})
		return nil, newPicoclawError(CodeRuntimeUnavailable, "gateway is unavailable")
	}

	s.runtime.Update(func(status *RuntimeStatus) {
		status.Ready = true
		status.Status = "ready"
		status.ConfigError = ""
		status.LastError = ""
		status.CheckedAt = time.Now()
		status.CurrentSession = s.lock.Owner()
	})

	return upstream, nil
}

// gatewayReadyPath is PicoClaw's readiness endpoint. Since v0.3 the gateway
// serves it unauthenticated on the same port as the pico channel, and it
// answers 200 only after the channels have started.
const gatewayReadyPath = "/ready"

// gatewayProbeTimeout bounds one readiness request. The gateway listens on
// loopback, so anything slower than this is as good as unavailable.
const gatewayProbeTimeout = 3 * time.Second

// probePicoclawGateway asks the gateway's GET /ready whether it is up. This
// replaces a full WebSocket handshake every few seconds. A gateway without
// the endpoint (404) is probed with the WebSocket handshake as before.
func probePicoclawGateway(cfg Config) *gatewayProbeError {
	readyURL, err := buildGatewayReadyURL(cfg)
	if err != nil {
		return &gatewayProbeError{
			status:      "config_error",
			configError: err.Error(),
			lastError:   err.Error(),
			message:     "gateway config is invalid",
		}
	}

	timeout := time.Duration(cfg.ConnectTimeoutMs) * time.Millisecond
	if timeout <= 0 || timeout > gatewayProbeTimeout {
		timeout = gatewayProbeTimeout
	}
	client := &http.Client{
		Timeout: timeout,
		// Never follow a redirect away from the loopback gateway.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	response, err := client.Get(readyURL)
	if err != nil {
		return &gatewayProbeError{
			status:    "unavailable",
			lastError: err.Error(),
			message:   "gateway is unavailable",
		}
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64*1024))
	_ = response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return probePicoclawGatewayWebSocket(cfg)
	case http.StatusServiceUnavailable:
		return &gatewayProbeError{
			status:    "unavailable",
			lastError: "gateway is not ready",
			message:   "gateway is not ready",
		}
	default:
		return &gatewayProbeError{
			status:    "unavailable",
			lastError: fmt.Sprintf("gateway readiness check failed: HTTP %d", response.StatusCode),
			message:   "gateway readiness check failed",
		}
	}
}

// buildGatewayReadyURL turns the configured ws:// gateway URL into the
// http:// URL of its readiness endpoint on the same host and port.
func buildGatewayReadyURL(cfg Config) (string, error) {
	parsed, err := url.Parse(cfg.GatewayURL)
	if err != nil {
		return "", fmt.Errorf("invalid gateway url: %w", err)
	}
	scheme := ""
	switch parsed.Scheme {
	case "ws":
		scheme = "http"
	case "wss":
		scheme = "https"
	default:
		return "", fmt.Errorf("invalid gateway url scheme: %s", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("invalid gateway url host")
	}
	return (&url.URL{Scheme: scheme, Host: parsed.Host, Path: gatewayReadyPath}).String(), nil
}

func probePicoclawGatewayWebSocket(cfg Config) *gatewayProbeError {
	gatewayURL, err := buildGatewayURL(cfg, "runtime-probe")
	if err != nil {
		return &gatewayProbeError{
			status:      "config_error",
			configError: err.Error(),
			lastError:   err.Error(),
			message:     "gateway config is invalid",
		}
	}

	header := http.Header{}
	if cfg.Token != "" {
		header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.Token))
	}

	timeout := time.Duration(cfg.ConnectTimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	dialer := websocket.Dialer{
		HandshakeTimeout: timeout,
		NetDialContext: (&net.Dialer{
			Timeout: timeout,
		}).DialContext,
	}

	conn, response, err := dialer.Dial(gatewayURL, header)
	if err == nil {
		_ = conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "runtime probe complete"),
			time.Now().Add(time.Second),
		)
		_ = conn.Close()
		return nil
	}

	probeErr := &gatewayProbeError{
		status:    "unavailable",
		lastError: err.Error(),
		message:   "gateway websocket is unavailable",
	}
	if response == nil {
		return probeErr
	}

	switch response.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		probeErr.status = "config_error"
		probeErr.configError = "gateway authentication failed"
		probeErr.message = "gateway authentication failed"
	case http.StatusNotFound:
		probeErr.lastError = "gateway pico channel is unavailable"
		probeErr.message = "gateway pico channel is unavailable"
	default:
		probeErr.lastError = fmt.Sprintf("gateway websocket handshake failed: HTTP %d", response.StatusCode)
		probeErr.message = "gateway websocket handshake failed"
	}

	return probeErr
}

func buildGatewayURL(cfg Config, sessionID string) (string, error) {
	parsed, err := url.Parse(cfg.GatewayURL)
	if err != nil {
		return "", fmt.Errorf("invalid gateway url: %w", err)
	}
	if parsed.Scheme != "ws" && parsed.Scheme != "wss" {
		return "", fmt.Errorf("invalid gateway url scheme: %s", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("invalid gateway url host")
	}

	query := parsed.Query()
	query.Set("session_id", sessionID)
	if cfg.Token != "" && cfg.AllowTokenQuery {
		query.Set("token", cfg.Token)
	}
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}

func gatewayHostPort(parsed *url.URL) (string, error) {
	if parsed == nil || parsed.Host == "" {
		return "", fmt.Errorf("invalid gateway url host")
	}
	if _, _, err := net.SplitHostPort(parsed.Host); err == nil {
		return parsed.Host, nil
	}

	switch parsed.Scheme {
	case "ws":
		return net.JoinHostPort(parsed.Host, "80"), nil
	case "wss":
		return net.JoinHostPort(parsed.Host, "443"), nil
	default:
		return "", fmt.Errorf("invalid gateway url scheme: %s", parsed.Scheme)
	}
}
