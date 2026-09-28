package proto

import (
	"encoding/json"
	"strings"
	"testing"
)

// Old clients read state, name, ip and account from the Tailscale status.
// Those four names must not move when the shape grows.
func TestTailscaleStatusKeepsItsFourFields(t *testing.T) {
	b, err := json.Marshal(GetTailscaleStatusRsp{
		State:   TailscaleRunning,
		Name:    "kvm",
		IP:      "100.1.2.3",
		Account: "owner@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"state":"running"`, `"name":"kvm"`, `"ip":"100.1.2.3"`, `"account":"owner@example.com"`,
	} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("%s is missing from %s", want, b)
		}
	}
}

func TestVpnStatusFieldNames(t *testing.T) {
	b, err := json.Marshal(VpnStatus{
		State:       VpnStopped,
		Version:     "0.78.2",
		Control:     true,
		Peers:       []VpnPeer{{Name: "laptop", IP: "100.73.10.20", Online: true}},
		UptimeSec:   61,
		BootEnabled: true,
		Memory:      VpnMemory{DaemonRSS: 1, GroupCurrent: 2, GroupHigh: 3, GroupMax: 4},
		BlockedBy:   "tailscale",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"state":"stopped"`, `"version":"0.78.2"`, `"control":true`,
		`"peers":[{"name":"laptop","ip":"100.73.10.20","online":true}]`,
		`"uptimeSec":61`, `"bootEnabled":true`,
		`"memory":{"daemonRss":1,"groupCurrent":2,"groupHigh":3,"groupMax":4}`,
		`"blockedBy":"tailscale"`,
	} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("%s is missing from %s", want, b)
		}
	}
}
