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

	// IPMI configures IPMI over LAN, kept under "ipmi" in server.yaml. A
	// file without the block loads with the service off.
	IPMI IPMI `yaml:"ipmi,omitempty" mapstructure:"ipmi"`

	// NetBoot configures network boot of the host, kept under "netboot" in
	// server.yaml. A file without the block loads with both sides off.
	NetBoot NetBoot `yaml:"netboot,omitempty" mapstructure:"netboot"`

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
