package webrtc

import (
	"strings"
	"testing"

	"NanoKVM-Server/config"
)

func TestValidateStunAcceptsHostPortAndOff(t *testing.T) {
	for _, value := range []string{
		"",
		"disable",
		"stun.l.google.com:19302",
		"stun.example.com:3478",
		"192.0.2.10:3478",
		"[2001:db8::1]:3478",
		"stun:1",
		"stun.example.com:65535",
	} {
		if err := ValidateStun(value); err != nil {
			t.Errorf("ValidateStun(%q) = %v, want nil", value, err)
		}
	}
}

func TestValidateStunRefusesWhatIsNotHostPort(t *testing.T) {
	for _, value := range []string{
		"stun.example.com",
		"stun.example.com:",
		"stun.example.com:0",
		"stun.example.com:65536",
		"stun.example.com:-1",
		"stun.example.com:port",
		":3478",
		"user:pass@stun.example:3478",
		"stun:stun.example.com:3478",
		"stun.example.com:3478/path",
		"stun example.com:3478",
		"bad_host!:3478",
		"2001:db8::1:3478",
		"[fe80::1%eth0]:3478",
		"disabled",
		strings.Repeat("a", 300) + ":3478",
	} {
		err := ValidateStun(value)
		if err == nil {
			t.Errorf("ValidateStun(%q) = nil, want an error", value)
			continue
		}
		if !strings.Contains(err.Error(), "STUN") {
			t.Errorf("ValidateStun(%q) = %q, want the error to name STUN", value, err)
		}
	}
}

func TestICEServersStun(t *testing.T) {
	for _, tc := range []struct {
		name string
		stun string
		want []string
	}{
		{"default when unset", "", []string{"stun:" + config.DefaultStun}},
		{"default", config.DefaultStun, []string{"stun:stun.l.google.com:19302"}},
		{"own server", "stun.example.com:3478", []string{"stun:stun.example.com:3478"}},
		{"off", config.StunDisabled, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			servers := iceServersFor(tc.stun, config.Turn{})
			var got []string
			for _, server := range servers {
				got = append(got, server.URLs...)
			}
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Fatalf("ICE servers for %q = %v, want %v", tc.stun, got, tc.want)
			}
		})
	}
}

func TestICEServersKeepTurnWithStunOff(t *testing.T) {
	servers := iceServersFor(config.StunDisabled, config.Turn{TurnAddr: "turn.example.com:3478", TurnUser: "u", TurnCred: "c"})
	if len(servers) != 1 || servers[0].URLs[0] != "turn:turn.example.com:3478" {
		t.Fatalf("ICE servers = %+v, want only the TURN server", servers)
	}
}

func TestSetStunAppliesToTheNextSession(t *testing.T) {
	conf := config.GetInstance()
	old := conf.Stun
	t.Cleanup(func() { SetStun(old) })

	SetStun("stun.example.com:3478")
	if got := createICEServers(); len(got) == 0 || got[0].URLs[0] != "stun:stun.example.com:3478" {
		t.Fatalf("after SetStun the next session got %+v", got)
	}
	SetStun(config.StunDisabled)
	for _, server := range createICEServers() {
		if strings.HasPrefix(server.URLs[0], "stun:") {
			t.Fatalf("STUN is off but the next session got %+v", server)
		}
	}
}
