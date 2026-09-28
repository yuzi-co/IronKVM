package netbird

import (
	"os"
	"reflect"
	"testing"

	"NanoKVM-Server/proto"
)

func load(t *testing.T, name string) *NbStatus {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	nb, err := parseStatus(b)
	if err != nil {
		t.Fatal(err)
	}
	return nb
}

// The JSON recorded on the board on 2026-09-27.
func TestParseConnectedStatusFromTheBoard(t *testing.T) {
	got, err := toVpnStatus(load(t, "status-connected.json"))
	if err != nil {
		t.Fatal(err)
	}
	want := proto.VpnStatus{
		State:   proto.VpnRunning,
		Version: "0.78.2",
		IP:      "100.73.212.105",
		Name:    "ironkvm.netbird.cloud",
		Account: "api.netbird.io",
		Control: true,
		Peers:   []proto.VpnPeer{},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %+v\nwant %+v", got, want)
	}
}

func TestParsePeers(t *testing.T) {
	got, err := toVpnStatus(load(t, "status-peers.json"))
	if err != nil {
		t.Fatal(err)
	}
	want := []proto.VpnPeer{
		{Name: "laptop", IP: "100.73.10.20", Online: true},
		{Name: "phone", IP: "100.73.10.21", Online: false},
	}
	if !reflect.DeepEqual(got.Peers, want) {
		t.Fatalf("got %+v", got.Peers)
	}
}

func TestParseNeedsLogin(t *testing.T) {
	got, err := toVpnStatus(load(t, "status-needslogin.json"))
	if err != nil || got.State != proto.VpnNotLogin {
		t.Fatalf("got %+v %v", got, err)
	}
}

func TestDaemonStatuses(t *testing.T) {
	for status, want := range map[string]proto.VpnState{
		"Idle":           proto.VpnStopped,
		"Connecting":     proto.VpnRunning,
		"Connected":      proto.VpnRunning,
		"NeedsLogin":     proto.VpnNotLogin,
		"LoginFailed":    proto.VpnNotLogin,
		"SessionExpired": proto.VpnNotLogin,
	} {
		got, err := toVpnStatus(&NbStatus{DaemonStatus: status})
		if err != nil || got.State != want {
			t.Fatalf("%s: got %q %v, want %q", status, got.State, err, want)
		}
	}
	// A status a later NetBird adds must not blank the page: the daemon
	// answered, so it runs.
	for i := 0; i < 2; i++ {
		got, err := toVpnStatus(&NbStatus{DaemonStatus: "Exploded", FQDN: "kvm.netbird.cloud"})
		if err != nil || got.State != proto.VpnRunning || got.Name != "kvm.netbird.cloud" {
			t.Fatalf("an unknown status reads as running: %+v %v", got, err)
		}
	}
}

func TestParseStatusRejectsGarbage(t *testing.T) {
	if _, err := parseStatus([]byte("Error: failed to connect to daemon")); err == nil {
		t.Fatal("output with no JSON must be an error")
	}
}
