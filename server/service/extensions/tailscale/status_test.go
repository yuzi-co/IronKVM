package tailscale

import (
	"os"
	"reflect"
	"testing"

	"NanoKVM-Server/proto"
)

func TestParseStatusRunning(t *testing.T) {
	b, err := os.ReadFile("testdata/status-running.json")
	if err != nil {
		t.Fatal(err)
	}
	ts, err := parseStatus(b)
	if err != nil {
		t.Fatal(err)
	}
	got, err := toVpnStatus(ts)
	if err != nil {
		t.Fatal(err)
	}
	want := proto.VpnStatus{
		State:   proto.VpnRunning,
		Version: "1.90.1",
		IP:      "100.101.102.103",
		Name:    "nanokvm",
		Account: "owner@example.com",
		Control: true,
		Peers: []proto.VpnPeer{
			{Name: "laptop", IP: "100.64.0.2", Online: true},
			{Name: "phone", IP: "100.64.0.3", Online: false},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %+v\nwant %+v", got, want)
	}
}

func TestParseStatusNeedsLogin(t *testing.T) {
	ts, err := parseStatus([]byte(`{"BackendState":"NeedsLogin","Self":{"HostName":"nanokvm","TailscaleIPs":null,"Online":false},"CurrentTailnet":null,"Peer":null}`))
	if err != nil {
		t.Fatal(err)
	}
	got, err := toVpnStatus(ts)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != proto.VpnNotLogin || got.Name != "nanokvm" || got.Peers == nil || len(got.Peers) != 0 {
		t.Fatalf("got %+v", got)
	}
}

// The CLI prints a version warning before the JSON when it and the daemon
// differ.
func TestParseStatusSkipsALeadingWarning(t *testing.T) {
	ts, err := parseStatus([]byte("Warning: client version \"1.88.3\" != tailscaled server version \"1.90.1\"\n" +
		`{"BackendState":"Stopped","Self":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	if got, err := toVpnStatus(ts); err != nil || got.State != proto.VpnStopped {
		t.Fatalf("got %+v %v", got, err)
	}
}

func TestParseStatusRejectsAnUnknownState(t *testing.T) {
	ts, err := parseStatus([]byte(`{"BackendState":"Exploded"}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := toVpnStatus(ts); err == nil {
		t.Fatal("an unknown state must be an error")
	}
}

func TestParseStatusRejectsGarbage(t *testing.T) {
	if _, err := parseStatus([]byte("failed to connect to local tailscaled")); err == nil {
		t.Fatal("output with no JSON must be an error")
	}
}
