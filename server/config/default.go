package config

import "log"

// DefaultStun is the STUN server a board uses until its owner sets another.
// The browser and the board both ask it for their public address at the start
// of every WebRTC session.
const DefaultStun = "stun.l.google.com:19302"

// StunDisabled, saved as the STUN server, turns STUN off.
const StunDisabled = "disable"

var defaultConfig = &Config{
	Proto: "http",
	Host:  "",
	Port: Port{
		Http:  80,
		Https: 443,
	},
	Cert: Cert{
		Crt: "server.crt",
		Key: "server.key",
	},
	Logger: Logger{
		Level: "info",
		File:  "stdout",
	},
	JWT: JWT{
		SecretKey:            "",
		RefreshTokenDuration: 2678400,
		RevokeTokensOnLogout: true,
	},
	Stun: DefaultStun,
	Turn: Turn{
		TurnAddr: "",
		TurnUser: "",
		TurnCred: "",
	},
	Authentication: "enable",
	Security: Security{
		LoginLockoutDuration: 0,
		LoginMaxFailures:     5,
	},
	Ion: Ion{
		ReserveFloor: 12582912,
	},
	HTTP2: false,
	Redfish: Redfish{
		Enabled: boolPtr(true),
	},
}

func boolPtr(v bool) *bool {
	return &v
}

func checkDefaultValue() {
	if instance.JWT.SecretKey == "" {
		key, err := loadOrCreateSecretKey()
		if err != nil {
			// Every session on the device is signed with this. Running with a
			// key we could not generate properly is not an option.
			log.Fatalf("failed to generate a jwt secret key: %s", err)
		}

		instance.JWT.SecretKey = key
		instance.JWT.RevokeTokensOnLogout = true
	}

	if instance.JWT.RefreshTokenDuration == 0 {
		instance.JWT.RefreshTokenDuration = 2678400
	}

	if instance.Stun == "" {
		instance.Stun = DefaultStun
	}

	if instance.Authentication == "" {
		instance.Authentication = "enable"
	}

	// A server.yaml written before this feature existed has no ion block, and a
	// zero floor would report every board as unavailable.
	if instance.Ion.ReserveFloor == 0 {
		instance.Ion.ReserveFloor = 12582912
	}

	instance.Hardware = getHardware()
}
