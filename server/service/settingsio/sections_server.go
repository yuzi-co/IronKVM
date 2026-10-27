package settingsio

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
	"NanoKVM-Server/service/hostlog"
	"NanoKVM-Server/service/nut"
	"NanoKVM-Server/service/rtsp"
	"NanoKVM-Server/service/stream/webrtc"
)

// The sections kept in server.yaml. Each type names the fields that leave the
// board; the rest of server.yaml (ports, protocol, certificates, the JWT
// secret, TURN credentials, the proxy, the allowed networks) is never read
// here. The values are exported as server.yaml holds them, before defaults,
// so a zero that means "the default" stays a zero on the other board.

// serverSpec builds a section from a getter and a setter on config.Config.
func serverSpec[T any](id string, get func(c *config.Config) T, set func(c *config.Config, v T), check func(s *snapshot, v *T) error) spec[T] {
	return spec[T]{
		id: id,
		get: func(s *snapshot) (T, error) {
			conf, err := s.config()
			if err != nil {
				var zero T
				return zero, err
			}
			return get(conf), nil
		},
		check: check,
		put: func(p *plan, v T) error {
			conf, err := p.config()
			if err != nil {
				return err
			}
			set(conf, v)
			return nil
		},
	}
}

type loggingSettings struct {
	Level string `json:"level"`
}

var loggingSection = serverSpec("logging",
	func(c *config.Config) loggingSettings { return loggingSettings{Level: c.Logger.Level} },
	func(c *config.Config, v loggingSettings) { c.Logger.Level = v.Level },
	func(_ *snapshot, v *loggingSettings) error {
		v.Level = strings.ToLower(strings.TrimSpace(v.Level))
		if _, err := logrus.ParseLevel(v.Level); err != nil {
			return fmt.Errorf("unknown log level %q", v.Level)
		}
		return nil
	})

type loginSecuritySettings struct {
	LoginLockoutDuration int `json:"loginLockoutDuration"`
	LoginMaxFailures     int `json:"loginMaxFailures"`
}

var loginSecuritySection = serverSpec("loginSecurity",
	func(c *config.Config) loginSecuritySettings {
		return loginSecuritySettings{
			LoginLockoutDuration: c.Security.LoginLockoutDuration,
			LoginMaxFailures:     c.Security.LoginMaxFailures,
		}
	},
	func(c *config.Config, v loginSecuritySettings) {
		c.Security.LoginLockoutDuration = v.LoginLockoutDuration
		c.Security.LoginMaxFailures = v.LoginMaxFailures
	},
	func(_ *snapshot, v *loginSecuritySettings) error {
		if v.LoginLockoutDuration < 0 || v.LoginLockoutDuration > 7*24*3600 {
			return errors.New("the lockout is 0 to 604800 seconds")
		}
		if v.LoginMaxFailures < 0 || v.LoginMaxFailures > 1000 {
			return errors.New("the failures before a lockout are 0 to 1000")
		}
		// With a lockout and no failures allowed, the first failed login
		// would lock the account out.
		if v.LoginLockoutDuration > 0 && v.LoginMaxFailures < 1 {
			return errors.New("a lockout needs at least one failure before it")
		}
		return nil
	})

type sessionSettings struct {
	RefreshTokenDuration uint64 `json:"refreshTokenDuration"`
	RevokeTokensOnLogout bool   `json:"revokeTokensOnLogout"`
}

// The limits of a session's length. A session shorter than five minutes
// would sign the administrator out while the import's reboot runs.
const (
	minSessionSeconds = 5 * 60
	maxSessionSeconds = 366 * 24 * 3600
)

var sessionsSection = serverSpec("sessions",
	func(c *config.Config) sessionSettings {
		return sessionSettings{
			RefreshTokenDuration: c.JWT.RefreshTokenDuration,
			RevokeTokensOnLogout: c.JWT.RevokeTokensOnLogout,
		}
	},
	func(c *config.Config, v sessionSettings) {
		c.JWT.RefreshTokenDuration = v.RefreshTokenDuration
		c.JWT.RevokeTokensOnLogout = v.RevokeTokensOnLogout
	},
	func(_ *snapshot, v *sessionSettings) error {
		if v.RefreshTokenDuration < minSessionSeconds || v.RefreshTokenDuration > maxSessionSeconds {
			return fmt.Errorf("the session length is %d to %d seconds", minSessionSeconds, maxSessionSeconds)
		}
		return nil
	})

type powerLEDSettings struct {
	PowerLED bool `json:"powerLed"`
}

var powerLEDSection = serverSpec("powerLed",
	func(c *config.Config) powerLEDSettings {
		return powerLEDSettings{PowerLED: c.HardwareSettings.PowerLED}
	},
	func(c *config.Config, v powerLEDSettings) { c.HardwareSettings.PowerLED = v.PowerLED },
	nil)

type enabledSettings struct {
	Enabled bool `json:"enabled"`
}

var redfishSection = serverSpec("redfish",
	func(c *config.Config) enabledSettings { return enabledSettings{Enabled: c.Redfish.IsEnabled()} },
	func(c *config.Config, v enabledSettings) {
		enabled := v.Enabled
		c.Redfish.Enabled = &enabled
	},
	nil)

var ipmiSection = serverSpec("ipmi",
	func(c *config.Config) enabledSettings { return enabledSettings{Enabled: c.IPMI.Enabled} },
	func(c *config.Config, v enabledSettings) { c.IPMI.Enabled = v.Enabled },
	nil)

type watchdogSettings struct {
	Enabled         bool   `json:"enabled"`
	TimeoutMinutes  int    `json:"timeoutMinutes"`
	Action          string `json:"action"`
	CooldownMinutes int    `json:"cooldownMinutes"`
	MaxPerHour      int    `json:"maxPerHour"`
	PingHost        string `json:"pingHost"`
}

func (w watchdogSettings) config() config.Watchdog {
	return config.Watchdog{
		Enabled:         w.Enabled,
		TimeoutMinutes:  w.TimeoutMinutes,
		Action:          w.Action,
		CooldownMinutes: w.CooldownMinutes,
		MaxPerHour:      w.MaxPerHour,
		PingHost:        w.PingHost,
	}
}

var watchdogSection = serverSpec("watchdog",
	func(c *config.Config) watchdogSettings {
		w := c.Watchdog
		return watchdogSettings{
			Enabled:         w.Enabled,
			TimeoutMinutes:  w.TimeoutMinutes,
			Action:          w.Action,
			CooldownMinutes: w.CooldownMinutes,
			MaxPerHour:      w.MaxPerHour,
			PingHost:        w.PingHost,
		}
	},
	func(c *config.Config, v watchdogSettings) { c.Watchdog = v.config() },
	// The limits of the watchdog page (service/watchdog/api.go), checked on
	// the values with their defaults, as the watchdog reads them.
	func(_ *snapshot, v *watchdogSettings) error {
		v.PingHost = strings.TrimSpace(v.PingHost)
		w := v.config().WithDefaults()
		switch {
		case w.TimeoutMinutes < 1 || w.TimeoutMinutes > 1440:
			return errors.New("the timeout is 1 to 1440 minutes")
		case w.Action != "reset" && w.Action != "power":
			return errors.New("the action is reset or power")
		case w.CooldownMinutes < 1 || w.CooldownMinutes > 1440:
			return errors.New("the cooldown is 1 to 1440 minutes")
		case w.MaxPerHour < 1 || w.MaxPerHour > 20:
			return errors.New("the actions per hour are 1 to 20")
		case w.PingHost != "" && net.ParseIP(w.PingHost) == nil:
			return errors.New("the ping address is not an IP address")
		}
		return nil
	})

type nutSettings struct {
	Enabled             bool   `json:"enabled"`
	Host                string `json:"host"`
	Port                int    `json:"port"`
	UPS                 string `json:"ups"`
	Username            string `json:"username"`
	OnBatteryMinutes    int    `json:"onBatteryMinutes"`
	ForceOffMinutes     int    `json:"forceOffMinutes"`
	PowerOnRestore      bool   `json:"powerOnRestore"`
	PowerOnMinCharge    int    `json:"powerOnMinCharge"`
	PowerOnDelaySeconds int    `json:"powerOnDelaySeconds"`
}

func (n nutSettings) config() config.NUT {
	return config.NUT{
		Enabled:             n.Enabled,
		Host:                n.Host,
		Port:                n.Port,
		UPS:                 n.UPS,
		Username:            n.Username,
		OnBatteryMinutes:    n.OnBatteryMinutes,
		ForceOffMinutes:     n.ForceOffMinutes,
		PowerOnRestore:      n.PowerOnRestore,
		PowerOnMinCharge:    n.PowerOnMinCharge,
		PowerOnDelaySeconds: n.PowerOnDelaySeconds,
	}
}

var nutSection = serverSpec("nut",
	func(c *config.Config) nutSettings {
		n := c.NUT
		return nutSettings{
			Enabled:             n.Enabled,
			Host:                n.Host,
			Port:                n.Port,
			UPS:                 n.UPS,
			Username:            n.Username,
			OnBatteryMinutes:    n.OnBatteryMinutes,
			ForceOffMinutes:     n.ForceOffMinutes,
			PowerOnRestore:      n.PowerOnRestore,
			PowerOnMinCharge:    n.PowerOnMinCharge,
			PowerOnDelaySeconds: n.PowerOnDelaySeconds,
		}
	},
	// The password stays in /etc/kvm/nut.passwd. Setting the block leaves
	// that file alone, so the other board keeps its own password, if any.
	func(c *config.Config, v nutSettings) { c.NUT = v.config() },
	// The limits of the NUT page (service/nut/api.go).
	func(_ *snapshot, v *nutSettings) error {
		n := v.config().WithDefaults()
		switch {
		case n.Host != "" && !nut.ValidHost(n.Host):
			return errors.New("invalid server address")
		case n.Port < 1 || n.Port > 65535:
			return errors.New("the port is 1 to 65535")
		case n.UPS != "" && !nut.ValidName(n.UPS):
			return errors.New("invalid UPS name")
		case n.Username != "" && !nut.ValidName(n.Username):
			return errors.New("invalid user name")
		case n.OnBatteryMinutes < 0 || n.OnBatteryMinutes > 1440,
			n.ForceOffMinutes < 0 || n.ForceOffMinutes > 1440:
			return errors.New("a time is 0 to 1440 minutes")
		case n.PowerOnMinCharge < 0 || n.PowerOnMinCharge > 100:
			return errors.New("the charge is 0 to 100 percent")
		case n.PowerOnDelaySeconds < 10 || n.PowerOnDelaySeconds > 3600:
			return errors.New("the power-on delay is 10 to 3600 seconds")
		case n.Enabled && (n.Host == "" || n.UPS == ""):
			return errors.New("the server address and the UPS name are needed")
		}
		return nil
	})

type netbootSettings struct {
	USB bool `json:"usb"`
	LAN bool `json:"lan"`
}

var netbootSection = serverSpec("netboot",
	func(c *config.Config) netbootSettings {
		return netbootSettings{USB: c.NetBoot.USB, LAN: c.NetBoot.LAN}
	},
	func(c *config.Config, v netbootSettings) { c.NetBoot = config.NetBoot{USB: v.USB, LAN: v.LAN} },
	nil)

type vncSettings struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
	MaxFPS  int  `json:"maxFps"`
	VNCAuth bool `json:"vncAuth"`
}

func (v vncSettings) config() config.VNC {
	return config.VNC{Enabled: v.Enabled, Port: v.Port, MaxFPS: v.MaxFPS, VNCAuth: v.VNCAuth}
}

var vncSection = serverSpec("vnc",
	func(c *config.Config) vncSettings {
		return vncSettings{Enabled: c.VNC.Enabled, Port: c.VNC.Port, MaxFPS: c.VNC.MaxFPS, VNCAuth: c.VNC.VNCAuth}
	},
	// The VNC password stays in /etc/kvm/vnc.passwd.
	func(c *config.Config, v vncSettings) { c.VNC = v.config() },
	// The limits of the VNC page (service/vnc/settings.go).
	func(s *snapshot, v *vncSettings) error {
		vnc := v.config().WithDefaults()
		if vnc.Port < 1 || vnc.Port > 65535 {
			return errors.New("the port is 1 to 65535")
		}
		if vnc.MaxFPS < 1 || vnc.MaxFPS > 60 {
			return errors.New("the frame rate is 1 to 60")
		}
		if conf, err := s.config(); err == nil {
			if vnc.Port == conf.Port.Http || vnc.Port == conf.Port.Https {
				return errors.New("the port is used by the web server")
			}
		}
		// The password is a secret and is not in the file, so plain VNC
		// authentication can only be turned on where one is already set.
		if vnc.VNCAuth && !s.exists(vncPasswordFile) {
			return errors.New("plain VNC authentication needs a VNC password on this board; set one first")
		}
		return nil
	})

type webrtcSettings struct {
	UDPPort       int    `json:"udpPort"`
	PublicAddress string `json:"publicAddress"`
	Stun          string `json:"stun"`
}

var webrtcSection = serverSpec("webrtc",
	func(c *config.Config) webrtcSettings {
		return webrtcSettings{UDPPort: c.WebRTC.UDPPort, PublicAddress: c.WebRTC.PublicAddress, Stun: c.Stun}
	},
	// TURN is not part of the section: its address, user and credential go
	// together, and the credential is a secret.
	func(c *config.Config, v webrtcSettings) {
		c.WebRTC = config.WebRTC{UDPPort: v.UDPPort, PublicAddress: v.PublicAddress}
		c.Stun = v.Stun
	},
	func(s *snapshot, v *webrtcSettings) error {
		v.PublicAddress = strings.TrimSpace(v.PublicAddress)
		v.Stun = strings.TrimSpace(v.Stun)
		var reserved []int
		if conf, err := s.config(); err == nil {
			reserved = []int{conf.Port.Http, conf.Port.Https}
		}
		if err := webrtc.ValidatePortForward(config.WebRTC{UDPPort: v.UDPPort, PublicAddress: v.PublicAddress}, reserved...); err != nil {
			return err
		}
		// A STUN server is host:port or "disable". A user name has no place
		// in it, so an @ is refused rather than carried to another board.
		return webrtc.ValidateStun(v.Stun)
	})

type rtspSettings struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
	UDP     bool `json:"udp"`
	UDPPort int  `json:"udpPort"`
	TLS     bool `json:"tls,omitempty"`
	TLSPort int  `json:"tlsPort,omitempty"`
}

func (v rtspSettings) config() config.RTSP {
	return config.RTSP{
		Enabled: v.Enabled, Port: v.Port, UDP: v.UDP, UDPPort: v.UDPPort,
		TLS: v.TLS, TLSPort: v.TLSPort,
	}
}

var rtspSection = serverSpec("rtsp",
	func(c *config.Config) rtspSettings {
		r := c.RTSP
		return rtspSettings{
			Enabled: r.Enabled, Port: r.Port, UDP: r.UDP, UDPPort: r.UDPPort,
			TLS: r.TLS, TLSPort: r.TLSPort,
		}
	},
	// Players authenticate with view-only API keys, which stay on the board
	// they were made on.
	func(c *config.Config, v rtspSettings) { c.RTSP = v.config() },
	// The limits of the RTSP page (service/rtsp/settings.go).
	func(s *snapshot, v *rtspSettings) error {
		r := v.config().WithDefaults()
		if r.Port < 1 || r.Port > 65535 {
			return errors.New("the port is 1 to 65535")
		}
		if r.UDPPort < 1 || r.UDPPort > 65534 {
			return errors.New("the UDP port is 1 to 65534")
		}
		if r.TLSPort < 1 || r.TLSPort > 65535 {
			return errors.New("the RTSPS port is 1 to 65535")
		}
		var reserved []int
		if conf, err := s.config(); err == nil {
			reserved = []int{conf.Port.Http, conf.Port.Https}
		}
		return rtsp.CheckPorts(r, reserved)
	})

type srtSettings struct {
	Enabled   bool `json:"enabled"`
	Port      int  `json:"port"`
	LatencyMs int  `json:"latencyMs"`
}

func (v srtSettings) config() config.SRT {
	return config.SRT{Enabled: v.Enabled, Port: v.Port, LatencyMs: v.LatencyMs}
}

var srtSection = serverSpec("srt",
	func(c *config.Config) srtSettings {
		return srtSettings{Enabled: c.SRT.Enabled, Port: c.SRT.Port, LatencyMs: c.SRT.LatencyMs}
	},
	// The passphrase stays in /etc/kvm/srt.passphrase.
	func(c *config.Config, v srtSettings) { c.SRT = v.config() },
	// The limits of the SRT page (service/srt/settings.go).
	func(s *snapshot, v *srtSettings) error {
		srt := v.config().WithDefaults()
		if srt.Port < 1 || srt.Port > 65535 {
			return errors.New("the port is 1 to 65535")
		}
		if srt.LatencyMs < config.MinSRTLatencyMs || srt.LatencyMs > config.MaxSRTLatencyMs {
			return fmt.Errorf("the latency is %d to %d ms", config.MinSRTLatencyMs, config.MaxSRTLatencyMs)
		}
		// The passphrase is a secret and is not in the file, so SRT can only
		// be switched on where one is already set.
		if srt.Enabled && !s.exists(srtPassphraseFile) {
			return errors.New("SRT needs a passphrase on this board; set one first")
		}
		return nil
	})

type picoclawSettings struct {
	ScreenshotWidth   int `json:"screenshotWidth"`
	ScreenshotQuality int `json:"screenshotQuality"`
}

var picoclawSection = serverSpec("picoclaw",
	func(c *config.Config) picoclawSettings {
		return picoclawSettings{ScreenshotWidth: c.Picoclaw.ScreenshotWidth, ScreenshotQuality: c.Picoclaw.ScreenshotQuality}
	},
	func(c *config.Config, v picoclawSettings) {
		c.Picoclaw = config.Picoclaw{ScreenshotWidth: v.ScreenshotWidth, ScreenshotQuality: v.ScreenshotQuality}
	},
	func(_ *snapshot, v *picoclawSettings) error {
		if v.ScreenshotWidth != 0 && (v.ScreenshotWidth < config.MinPicoclawScreenshotWidth || v.ScreenshotWidth > config.MaxPicoclawScreenshotWidth) {
			return fmt.Errorf("the screenshot width is %d to %d", config.MinPicoclawScreenshotWidth, config.MaxPicoclawScreenshotWidth)
		}
		if v.ScreenshotQuality != 0 && (v.ScreenshotQuality < config.MinPicoclawScreenshotQuality || v.ScreenshotQuality > config.MaxPicoclawScreenshotQuality) {
			return fmt.Errorf("the screenshot quality is %d to %d", config.MinPicoclawScreenshotQuality, config.MaxPicoclawScreenshotQuality)
		}
		return nil
	})

type hostLogListener struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
}

type hostLogSettings struct {
	Syslog     hostLogListener `json:"syslog"`
	Netconsole hostLogListener `json:"netconsole"`
	Bind       string          `json:"bind"`
	Sources    []string        `json:"sources"`
	MemoryKB   int             `json:"memoryKb"`
	Disk       bool            `json:"disk"`
	DiskKB     int             `json:"diskKb"`
	Patterns   []string        `json:"patterns"`
}

func (h hostLogSettings) config() config.HostLog {
	return config.HostLog{
		Syslog:     config.HostLogListener{Enabled: h.Syslog.Enabled, Port: h.Syslog.Port},
		Netconsole: config.HostLogListener{Enabled: h.Netconsole.Enabled, Port: h.Netconsole.Port},
		Bind:       h.Bind,
		Sources:    h.Sources,
		MemoryKB:   h.MemoryKB,
		Disk:       h.Disk,
		DiskKB:     h.DiskKB,
		Patterns:   h.Patterns,
	}
}

var hostLogSection = serverSpec("hostLog",
	func(c *config.Config) hostLogSettings {
		h := c.HostLog
		return hostLogSettings{
			Syslog:     hostLogListener{Enabled: h.Syslog.Enabled, Port: h.Syslog.Port},
			Netconsole: hostLogListener{Enabled: h.Netconsole.Enabled, Port: h.Netconsole.Port},
			Bind:       h.Bind,
			Sources:    append([]string{}, h.Sources...),
			MemoryKB:   h.MemoryKB,
			Disk:       h.Disk,
			DiskKB:     h.DiskKB,
			Patterns:   append([]string{}, h.Patterns...),
		}
	},
	func(c *config.Config, v hostLogSettings) { c.HostLog = v.config() },
	// The limits of the host log page (service/hostlog/api.go).
	func(_ *snapshot, v *hostLogSettings) error {
		_, err := hostlog.ValidateSettings(v.config().WithDefaults())
		return err
	})
