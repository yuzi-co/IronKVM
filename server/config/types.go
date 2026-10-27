package config

type Config struct {
	Proto          string   `yaml:"proto"`
	Host           string   `yaml:"host"`
	Port           Port     `yaml:"port"`
	Cert           Cert     `yaml:"cert"`
	Logger         Logger   `yaml:"logger"`
	Authentication string   `yaml:"authentication"`
	JWT            JWT      `yaml:"jwt"`
	Stun           string   `yaml:"stun"`
	Turn           Turn     `yaml:"turn"`
	Security       Security `yaml:"security"`
	Ion            Ion      `yaml:"ion"`
	// Proxy reaches the internet through an intermediary. A complete URL or a
	// bare host:port. Empty means the environment decides.
	Proxy string `yaml:"proxy"`

	// HTTP2 offers HTTP/2 on the HTTPS listener. Off by default, and the
	// default is the recommendation: see utils.NewServer for what it costs on
	// this hardware and for the websocket it takes away. It has no effect on a
	// plain HTTP listener, which never speaks h2 here.
	HTTP2 bool `yaml:"http2"`

	// HardwareSettings is what the owner says about how the board is wired to
	// the host, kept under "hardware" in server.yaml. A file without the
	// block loads with every setting off.
	HardwareSettings HardwareSettings `yaml:"hardware" mapstructure:"hardware"`

	// Redfish configures the Redfish service at /redfish, kept under
	// "redfish" in server.yaml.
	Redfish Redfish `yaml:"redfish,omitempty" mapstructure:"redfish"`

	// Watchdog configures the host watchdog, kept under "watchdog" in
	// server.yaml. A file without the block loads with the watchdog off.
	Watchdog Watchdog `yaml:"watchdog,omitempty" mapstructure:"watchdog"`

	// NUT configures the NUT client, kept under "nut" in server.yaml. A file
	// without the block loads with the client off. The password is kept
	// outside server.yaml.
	NUT NUT `yaml:"nut,omitempty" mapstructure:"nut"`

	// IPMI configures IPMI over LAN, kept under "ipmi" in server.yaml. A
	// file without the block loads with the service off.
	IPMI IPMI `yaml:"ipmi,omitempty" mapstructure:"ipmi"`

	// NetBoot configures network boot of the host, kept under "netboot" in
	// server.yaml. A file without the block loads with both sides off.
	NetBoot NetBoot `yaml:"netboot,omitempty" mapstructure:"netboot"`

	// VNC configures the VNC server, kept under "vnc" in server.yaml. A file
	// without the block loads with the server off.
	VNC VNC `yaml:"vnc,omitempty" mapstructure:"vnc"`

	// WebRTC reaches the video through a port forward, kept under "webrtc"
	// in server.yaml. A file without the block loads with both settings off,
	// which is how WebRTC has always been set up.
	WebRTC WebRTC `yaml:"webrtc,omitempty" mapstructure:"webrtc"`

	// WHEP serves the video to WHEP players, kept under "whep" in
	// server.yaml. A file without the block loads with the endpoint off.
	WHEP WHEP `yaml:"whep,omitempty" mapstructure:"whep"`

	// RTSP serves the video and the host audio to RTSP players, kept under
	// "rtsp" in server.yaml. A file without the block loads with the server
	// off.
	RTSP RTSP `yaml:"rtsp,omitempty" mapstructure:"rtsp"`

	// SRT serves the video and the host audio over SRT, kept under "srt" in
	// server.yaml. A file without the block loads with the listener off.
	SRT SRT `yaml:"srt,omitempty" mapstructure:"srt"`

	// Picoclaw tunes what the server hands the PicoClaw assistant, kept under
	// "picoclaw" in server.yaml. A file without the block loads the defaults.
	Picoclaw Picoclaw `yaml:"picoclaw,omitempty" mapstructure:"picoclaw"`

	// HostLog configures the receiver for the host's logs, kept under
	// "hostlog" in server.yaml. A file without the block loads with both
	// listeners off.
	HostLog HostLog `yaml:"hostlog,omitempty" mapstructure:"hostlog"`

	// Hardware holds the board's pins, derived from its version at start. It
	// is never read from or written to server.yaml.
	Hardware Hardware `yaml:"-" mapstructure:"-"`
}

// HardwareSettings are the wiring choices the owner makes.
type HardwareSettings struct {
	// PowerLED says the host's power LED header is connected to the board.
	// Without it the LED line reads "off" whatever the host is doing, so the
	// power state is unknown, and anything that decides from it, such as
	// Redfish's On and ForceOff, must not press. Off by default: most boards
	// are installed without that header.
	PowerLED bool `yaml:"powerLed" mapstructure:"powerLed"`
}

// WebRTC lets a viewer reach the video from outside the LAN through a router's
// port forward. See service/stream/webrtc/portforward.go.
type WebRTC struct {
	// UDPPort, when not zero, is the one UDP port every viewer's ICE agent
	// shares, so a router can forward it. Zero leaves each viewer on its own
	// ephemeral port, as before.
	UDPPort int `yaml:"udpPort,omitempty" mapstructure:"udpPort"`
	// PublicAddress is the address a viewer outside the LAN reaches the router
	// at: an IP or a host name, resolved when a viewer connects. It is offered
	// as one more host candidate on UDPPort. It needs UDPPort.
	PublicAddress string `yaml:"publicAddress,omitempty" mapstructure:"publicAddress"`
}

// WHEP configures the WHEP endpoint, /api/whep (ironkvm-dist#145).
type WHEP struct {
	// Enabled answers WHEP offers. Off by default: it is one more way in to
	// the picture, for an owner who wants a player such as GStreamer or OBS
	// on it.
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
}

// The RTSP server's defaults (ironkvm-dist#142). A zero value in server.yaml
// means the default. 8554 is the port RTSP servers that cannot bind 554 use,
// and 8000 and 8001 are the RTP and RTCP ports MediaMTX picked, so a player's
// firewall rules for either carry over. 8322 is MediaMTX's RTSPS port, for
// the same reason.
const (
	DefaultRTSPPort    = 8554
	DefaultRTSPUDPPort = 8000
	DefaultRTSPSPort   = 8322
)

// RTSP configures the RTSP server. Read it through WithDefaults.
type RTSP struct {
	// Enabled listens for plain RTSP players on Port. Off by default.
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
	// Port is the TCP port plain RTSP listens on.
	Port int `yaml:"port,omitempty" mapstructure:"port"`
	// UDP offers RTP over UDP beside RTP interleaved in the RTSP connection.
	// Off by default: TCP crosses NAT and firewalls as it is, and a player
	// that asks for UDP falls back to it.
	UDP bool `yaml:"udp,omitempty" mapstructure:"udp"`
	// UDPPort is the RTP port when UDP is on. RTCP takes the port above it.
	UDPPort int `yaml:"udpPort,omitempty" mapstructure:"udpPort"`
	// TLS listens for RTSPS players, RTSP over TLS with the web server's
	// certificate, on TLSPort. Off by default. The media stays inside the TLS
	// connection: UDP above applies to plain RTSP only, so an RTSPS player
	// never sends or receives RTP outside it.
	TLS bool `yaml:"tls,omitempty" mapstructure:"tls"`
	// TLSPort is the TCP port RTSPS listens on.
	TLSPort int `yaml:"tlsPort,omitempty" mapstructure:"tlsPort"`
}

// Listening reports whether either listener is on.
func (r RTSP) Listening() bool {
	return r.Enabled || r.TLS
}

// WithDefaults returns the settings with every zero value replaced by its
// default.
func (r RTSP) WithDefaults() RTSP {
	if r.Port <= 0 {
		r.Port = DefaultRTSPPort
	}
	if r.UDPPort <= 0 {
		r.UDPPort = DefaultRTSPUDPPort
	}
	if r.TLSPort <= 0 {
		r.TLSPort = DefaultRTSPSPort
	}
	return r
}

// The SRT listener's defaults (ironkvm-dist#143). 8890 is the port MediaMTX
// listens for SRT on, and 120 ms is SRT's own default latency.
const (
	DefaultSRTPort      = 8890
	DefaultSRTLatencyMs = 120
	MinSRTLatencyMs     = 20
	MaxSRTLatencyMs     = 8000
)

// SRT configures the SRT listener. Read it through WithDefaults. The
// passphrase it requires is kept outside server.yaml.
type SRT struct {
	// Enabled listens for SRT callers. Off by default.
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
	// Port is the UDP port the listener takes.
	Port int `yaml:"port,omitempty" mapstructure:"port"`
	// LatencyMs is how long the receiver holds a packet for a
	// retransmission to arrive before it plays on without it. A caller that
	// asks for more gets more.
	LatencyMs int `yaml:"latencyMs,omitempty" mapstructure:"latencyMs"`
}

// WithDefaults returns the settings with every zero value replaced by its
// default.
func (s SRT) WithDefaults() SRT {
	if s.Port <= 0 {
		s.Port = DefaultSRTPort
	}
	if s.LatencyMs <= 0 {
		s.LatencyMs = DefaultSRTLatencyMs
	}
	return s
}

// Redfish configures the Redfish service.
type Redfish struct {
	// Enabled serves /redfish. It is a pointer so that a server.yaml written
	// before the setting existed, which has no key, keeps the service on as
	// it always was. Read it through IsEnabled.
	Enabled *bool `yaml:"enabled,omitempty" mapstructure:"enabled"`
}

// IsEnabled reports whether the Redfish service answers. A missing key means
// on.
func (r Redfish) IsEnabled() bool {
	return r.Enabled == nil || *r.Enabled
}

// IPMI configures IPMI over LAN.
type IPMI struct {
	// Enabled answers IPMI on UDP port 623. Off by default, because IPMI
	// 2.0 authentication is weak by design: anyone who knows a user name
	// can get a hash of that user's IPMI password to crack offline.
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
}

// The host watchdog's defaults. A zero value in server.yaml means the default.
const (
	DefaultWatchdogTimeoutMinutes  = 5
	DefaultWatchdogAction          = "reset"
	DefaultWatchdogCooldownMinutes = 15
	DefaultWatchdogMaxPerHour      = 3
)

// Watchdog configures the host watchdog, which resets or power-cycles the
// host when its picture stops changing. Read it through WithDefaults.
type Watchdog struct {
	// Enabled runs the watchdog. Off by default: sampling the picture keeps
	// capture running.
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
	// TimeoutMinutes is how long the host may show no sign of life before
	// the watchdog acts.
	TimeoutMinutes int `yaml:"timeoutMinutes,omitempty" mapstructure:"timeoutMinutes"`
	// Action is "reset" or "power".
	Action string `yaml:"action,omitempty" mapstructure:"action"`
	// CooldownMinutes is the least time between two actions.
	CooldownMinutes int `yaml:"cooldownMinutes,omitempty" mapstructure:"cooldownMinutes"`
	// MaxPerHour caps the actions in any hour.
	MaxPerHour int `yaml:"maxPerHour,omitempty" mapstructure:"maxPerHour"`
	// PingHost is an address the watchdog pings; a reply is a sign of life.
	// Empty means no ping.
	PingHost string `yaml:"pingHost,omitempty" mapstructure:"pingHost"`
}

// WithDefaults returns the settings with every zero value replaced by its
// default.
func (w Watchdog) WithDefaults() Watchdog {
	if w.TimeoutMinutes <= 0 {
		w.TimeoutMinutes = DefaultWatchdogTimeoutMinutes
	}
	if w.Action == "" {
		w.Action = DefaultWatchdogAction
	}
	if w.CooldownMinutes <= 0 {
		w.CooldownMinutes = DefaultWatchdogCooldownMinutes
	}
	if w.MaxPerHour <= 0 {
		w.MaxPerHour = DefaultWatchdogMaxPerHour
	}
	return w
}

// The NUT client's defaults. A zero value in server.yaml means the default.
const (
	DefaultNUTPort                = 3493
	DefaultNUTRestoreDelaySeconds = 60
)

// NUT configures the NUT client, which reads a UPS from a NUT server over the
// network and drives the host's ATX buttons from what it finds. Read it
// through WithDefaults.
type NUT struct {
	// Enabled runs the client. Off by default: it presses the host's power
	// button.
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
	// Host is the NUT server's address, and Port its TCP port.
	Host string `yaml:"host,omitempty" mapstructure:"host"`
	Port int    `yaml:"port,omitempty" mapstructure:"port"`
	// UPS is the name of the UPS on that server.
	UPS string `yaml:"ups,omitempty" mapstructure:"ups"`
	// Username is the NUT user. With a password the client logs in as a
	// secondary, so the primary waits for the host to shut down before it cuts
	// the UPS output. Without them it only reads.
	Username string `yaml:"username,omitempty" mapstructure:"username"`

	// OnBatteryMinutes shuts the host down once the UPS has run on battery
	// this long. Zero means never: low battery and a forced shutdown still do.
	OnBatteryMinutes int `yaml:"onBatteryMinutes,omitempty" mapstructure:"onBatteryMinutes"`
	// ForceOffMinutes presses power for five seconds when the host is still on
	// this long after the shutdown press. Zero means never.
	ForceOffMinutes int `yaml:"forceOffMinutes,omitempty" mapstructure:"forceOffMinutes"`

	// PowerOnRestore presses power when mains is back and the host is off.
	PowerOnRestore bool `yaml:"powerOnRestore,omitempty" mapstructure:"powerOnRestore"`
	// PowerOnMinCharge is the battery charge, in percent, the UPS must reach
	// before the host is switched on. Zero means any.
	PowerOnMinCharge int `yaml:"powerOnMinCharge,omitempty" mapstructure:"powerOnMinCharge"`
	// PowerOnDelaySeconds is how long mains must be back before the press.
	PowerOnDelaySeconds int `yaml:"powerOnDelaySeconds,omitempty" mapstructure:"powerOnDelaySeconds"`
}

// WithDefaults returns the settings with every zero value replaced by its
// default.
func (n NUT) WithDefaults() NUT {
	if n.Port <= 0 {
		n.Port = DefaultNUTPort
	}
	if n.PowerOnDelaySeconds <= 0 {
		n.PowerOnDelaySeconds = DefaultNUTRestoreDelaySeconds
	}
	return n
}

// NetBoot configures network boot of the host. Both sides are off by
// default, and both need the netboot add-on, which carries dnsmasq and the
// boot files.
type NetBoot struct {
	// USB serves PXE, iPXE and a menu of the images on /data to the host over
	// the USB network link. It takes effect only while the link is on.
	USB bool `yaml:"usb" mapstructure:"usb"`
	// LAN answers PXE clients on the LAN with proxy DHCP. It never hands out
	// an address, and it offers a boot loader only.
	LAN bool `yaml:"lan" mapstructure:"lan"`
}

// The VNC server's defaults. A zero value in server.yaml means the default.
const (
	DefaultVNCPort   = 5900
	DefaultVNCMaxFPS = 15
)

// VNC configures the VNC server. Read it through WithDefaults.
type VNC struct {
	// Enabled listens for VNC clients. Off by default.
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
	// Port is the TCP port the server listens on.
	Port int `yaml:"port,omitempty" mapstructure:"port"`
	// MaxFPS caps the frames a second a client is sent, below the stream's
	// own rate.
	MaxFPS int `yaml:"maxFps,omitempty" mapstructure:"maxFps"`
	// VNCAuth offers plain VNC authentication beside VeNCrypt. It sends the
	// picture and the keystrokes without encryption, so it is off by default.
	// The password it checks is kept outside server.yaml.
	VNCAuth bool `yaml:"vncAuth,omitempty" mapstructure:"vncAuth"`
}

// WithDefaults returns the settings with every zero value replaced by its
// default.
func (v VNC) WithDefaults() VNC {
	if v.Port <= 0 {
		v.Port = DefaultVNCPort
	}
	if v.MaxFPS <= 0 {
		v.MaxFPS = DefaultVNCMaxFPS
	}
	return v
}

type Logger struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

type Port struct {
	Http  int `yaml:"http"`
	Https int `yaml:"https"`
}

type Cert struct {
	Crt string `yaml:"crt"`
	Key string `yaml:"key"`
}

type JWT struct {
	SecretKey            string `yaml:"secretKey"`
	RefreshTokenDuration uint64 `yaml:"refreshTokenDuration"`
	RevokeTokensOnLogout bool   `yaml:"revokeTokensOnLogout"`
}

type Turn struct {
	TurnAddr string `yaml:"turnAddr"`
	TurnUser string `yaml:"turnUser"`
	TurnCred string `yaml:"turnCred"`
}

type Security struct {
	LoginLockoutDuration int `yaml:"loginLockoutDuration"`
	LoginMaxFailures     int `yaml:"loginMaxFailures"`

	// AllowedNetworks are the client addresses, as CIDRs or single
	// addresses, that may reach the web UI, the API, VNC and IPMI. Empty
	// allows every address. See service/allowlist.
	AllowedNetworks []string `yaml:"allowedNetworks,omitempty"`
}

// Ion configures how the carveout is graded.
type Ion struct {
	// ReserveFloor is the cost of starting the stream, in bytes. 12MB: opening a
	// stream on a fresh board took the carveout from 19,050,496 to 30,392,320,
	// so one stream start costs 11,341,824 bytes.
	//
	// It is what a generation that has not opened a stream still has ahead of
	// it. As the generation allocates, that much of the cost stops being
	// pending, and the reserve falls with it until only the orphan cost is
	// left. See service/ion, which holds the reasoning.
	//
	// An earlier value of 24MB was the cost of a whole capture session, which
	// graded a healthy board amber: `ok` needs twice this much free, and twice
	// 24MB is 64% of the 75MB carveout, so a board with video running could
	// never reach it. The same arithmetic reached the verdict a second way
	// through the measured path, and that is fixed in service/ion rather than
	// here.
	ReserveFloor uint64 `yaml:"reserveFloor"`
}

type Hardware struct {
	Version      HWVersion `yaml:"-"`
	GPIOReset    string    `yaml:"-"`
	GPIOPower    string    `yaml:"-"`
	GPIOPowerLED string    `yaml:"-"`
	GPIOHDDLed   string    `yaml:"-"`
}

// The defaults and limits of the screenshots PicoClaw's model is sent. A zero
// value in server.yaml means the default.
const (
	DefaultPicoclawScreenshotWidth   = 960
	DefaultPicoclawScreenshotQuality = 60
	MinPicoclawScreenshotWidth       = 320
	MaxPicoclawScreenshotWidth       = 1920
	MinPicoclawScreenshotQuality     = 10
	MaxPicoclawScreenshotQuality     = 100
)

// Picoclaw configures the server side of the PicoClaw integration. Read it
// through WithDefaults.
type Picoclaw struct {
	// ScreenshotWidth is the largest width, in pixels, of a screenshot sent
	// to the model when it does not ask for a size; the height follows at
	// 16:9 at most. Smaller images cost fewer tokens and less time on small
	// local models, larger ones help them read small text and hit small
	// targets. The model can still ask for any size per screenshot.
	ScreenshotWidth int `yaml:"screenshotWidth,omitempty" mapstructure:"screenshotWidth"`
	// ScreenshotQuality is the JPEG quality, 10 to 100, of those screenshots.
	ScreenshotQuality int `yaml:"screenshotQuality,omitempty" mapstructure:"screenshotQuality"`
}

// WithDefaults returns the settings with zero values replaced by their
// defaults and the rest held within their limits.
func (p Picoclaw) WithDefaults() Picoclaw {
	switch {
	case p.ScreenshotWidth <= 0:
		p.ScreenshotWidth = DefaultPicoclawScreenshotWidth
	case p.ScreenshotWidth < MinPicoclawScreenshotWidth:
		p.ScreenshotWidth = MinPicoclawScreenshotWidth
	case p.ScreenshotWidth > MaxPicoclawScreenshotWidth:
		p.ScreenshotWidth = MaxPicoclawScreenshotWidth
	}
	switch {
	case p.ScreenshotQuality <= 0:
		p.ScreenshotQuality = DefaultPicoclawScreenshotQuality
	case p.ScreenshotQuality < MinPicoclawScreenshotQuality:
		p.ScreenshotQuality = MinPicoclawScreenshotQuality
	case p.ScreenshotQuality > MaxPicoclawScreenshotQuality:
		p.ScreenshotQuality = MaxPicoclawScreenshotQuality
	}
	return p
}

// The host log receiver's defaults. A zero value in server.yaml means the
// default.
const (
	DefaultSyslogPort      = 514
	DefaultNetconsolePort  = 6666
	DefaultHostLogMemoryKB = 1024
	DefaultHostLogDiskKB   = 4096
)

// HostLog configures the receiver for the managed host's logs: UDP syslog and
// netconsole. Both listeners are off by default. Read it through WithDefaults.
type HostLog struct {
	Syslog     HostLogListener `yaml:"syslog,omitempty" mapstructure:"syslog"`
	Netconsole HostLogListener `yaml:"netconsole,omitempty" mapstructure:"netconsole"`
	// Bind is where the listeners bind: "" for every interface, or "usb" for
	// the board's address on the USB network link.
	Bind string `yaml:"bind,omitempty" mapstructure:"bind"`
	// Sources are the client addresses, single or CIDR, that may send. Empty
	// means the host on the USB link when it is up, else the watchdog's ping
	// host, else any address.
	Sources []string `yaml:"sources,omitempty" mapstructure:"sources"`
	// MemoryKB caps the in-memory ring.
	MemoryKB int `yaml:"memoryKb,omitempty" mapstructure:"memoryKb"`
	// Disk keeps the log on /data as well, up to DiskKB. Off by default:
	// it writes to the SD card.
	Disk   bool `yaml:"disk,omitempty" mapstructure:"disk"`
	DiskKB int  `yaml:"diskKb,omitempty" mapstructure:"diskKb"`
	// Patterns are regular expressions; a line that matches one raises a
	// notification. Empty means the defaults for a kernel panic.
	Patterns []string `yaml:"patterns,omitempty" mapstructure:"patterns"`
}

// HostLogListener is one UDP listener.
type HostLogListener struct {
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
	// Port is the UDP port; zero means the protocol's default.
	Port int `yaml:"port,omitempty" mapstructure:"port"`
}

// WithDefaults returns the settings with every zero value replaced by its
// default.
func (h HostLog) WithDefaults() HostLog {
	if h.Syslog.Port <= 0 {
		h.Syslog.Port = DefaultSyslogPort
	}
	if h.Netconsole.Port <= 0 {
		h.Netconsole.Port = DefaultNetconsolePort
	}
	if h.MemoryKB <= 0 {
		h.MemoryKB = DefaultHostLogMemoryKB
	}
	if h.DiskKB <= 0 {
		h.DiskKB = DefaultHostLogDiskKB
	}
	if len(h.Patterns) == 0 {
		h.Patterns = append([]string(nil), DefaultHostLogPatterns...)
	}
	return h
}

// DefaultHostLogPatterns are the phrases that mean the host is in trouble.
var DefaultHostLogPatterns = []string{
	`Kernel panic`,
	`\bOops\b`,
	`BUG: `,
	`Out of memory: Kill`,
}
