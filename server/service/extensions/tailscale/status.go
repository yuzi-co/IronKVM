package tailscale

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"

	"NanoKVM-Server/proto"
)

// TsPeer is one node in `tailscale status --json`: Self, or an entry of Peer.
type TsPeer struct {
	HostName     string   `json:"HostName"`
	DNSName      string   `json:"DNSName"`
	TailscaleIPs []string `json:"TailscaleIPs"`
	Online       bool     `json:"Online"`
}

// TsStatus is the part of `tailscale status --json` the page shows.
type TsStatus struct {
	Version        string `json:"Version"`
	BackendState   string `json:"BackendState"`
	Self           TsPeer `json:"Self"`
	CurrentTailnet struct {
		Name string `json:"Name"`
	} `json:"CurrentTailnet"`
	Peer map[string]TsPeer `json:"Peer"`
}

var StateMap = map[string]proto.VpnState{
	"NoState":          proto.VpnNotRunning,
	"Starting":         proto.VpnNotRunning,
	"NeedsLogin":       proto.VpnNotLogin,
	"NeedsMachineAuth": proto.VpnNotLogin,
	"InUseOtherUser":   proto.VpnNotLogin,
	"Running":          proto.VpnRunning,
	"Stopped":          proto.VpnStopped,
}

// parseStatus reads the CLI's JSON. The CLI can print a warning line before
// it, so everything before the first brace is skipped.
func parseStatus(out []byte) (*TsStatus, error) {
	i := bytes.IndexByte(out, '{')
	if i < 0 {
		return nil, errors.New("unknown output")
	}
	var st TsStatus
	if err := json.Unmarshal(out[i:], &st); err != nil {
		return nil, err
	}
	return &st, nil
}

// toVpnStatus is the page's view of the CLI's status. Uptime, memory, start at
// boot and the blocker are vpn.Fill's.
func toVpnStatus(ts *TsStatus) (proto.VpnStatus, error) {
	state, ok := StateMap[ts.BackendState]
	if !ok {
		return proto.VpnStatus{}, fmt.Errorf("unknown tailscale state: %s", ts.BackendState)
	}
	st := proto.VpnStatus{
		State:   state,
		Version: shortVersion(ts.Version),
		IP:      ipv4(ts.Self.TailscaleIPs),
		Name:    ts.Self.HostName,
		Account: ts.CurrentTailnet.Name,
		Control: ts.Self.Online,
		Peers:   []proto.VpnPeer{},
	}
	for _, p := range ts.Peer {
		st.Peers = append(st.Peers, proto.VpnPeer{Name: peerName(p), IP: ipv4(p.TailscaleIPs), Online: p.Online})
	}
	sort.Slice(st.Peers, func(i, j int) bool { return st.Peers[i].Name < st.Peers[j].Name })
	return st, nil
}

// shortVersion drops the build suffix: 1.90.1-t8b5c4a1e2-g3f6d7c8b9 is 1.90.1.
func shortVersion(v string) string {
	s, _, _ := strings.Cut(v, "-")
	return s
}

// ipv4 is the node's IPv4 address, the one the page has always shown.
func ipv4(ips []string) string {
	for _, s := range ips {
		if ip := net.ParseIP(s); ip != nil && ip.To4() != nil {
			return ip.String()
		}
	}
	return ""
}

// peerName is the first label of the MagicDNS name, which is how the tailnet
// names the machine, or the OS host name when there is none.
func peerName(p TsPeer) string {
	if label, _, _ := strings.Cut(p.DNSName, "."); label != "" {
		return label
	}
	return p.HostName
}
