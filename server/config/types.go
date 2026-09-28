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
