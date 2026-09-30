package netbird

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strings"
	"sync"

	"NanoKVM-Server/proto"

	log "github.com/sirupsen/logrus"
)

// NbPeer is one entry of peers.details in `netbird status --json`. The names
// are PeerStateDetailOutput's in client/status/status.go at v0.78.2.
type NbPeer struct {
	FQDN   string `json:"fqdn"`
	IP     string `json:"netbirdIp"`
	Status string `json:"status"` // Idle, Connecting, Connected
}

// NbStatus is the part of `netbird status --json` the page shows.
type NbStatus struct {
	Peers struct {
		Details []NbPeer `json:"details"`
	} `json:"peers"`
	DaemonVersion string `json:"daemonVersion"`
	DaemonStatus  string `json:"daemonStatus"`
	Management    struct {
		URL       string `json:"url"`
		Connected bool   `json:"connected"`
	} `json:"management"`
	IP   string `json:"netbirdIp"`
	FQDN string `json:"fqdn"`
}

// StateMap maps the daemon statuses of client/status at v0.78.2. A fresh
// daemon with no login reports NeedsLogin; `netbird down` leaves it Idle.
var StateMap = map[string]proto.VpnState{
	"Idle":           proto.VpnStopped,
	"Connecting":     proto.VpnRunning,
	"Connected":      proto.VpnRunning,
	"NeedsLogin":     proto.VpnNotLogin,
	"LoginFailed":    proto.VpnNotLogin,
	"SessionExpired": proto.VpnNotLogin,
}

// unknownStatuses holds the daemon statuses already logged as unknown, so
// each one is logged once and not on every visit to the page.
var unknownStatuses sync.Map

// parseStatus reads the CLI's JSON, skipping anything printed before it.
func parseStatus(out []byte) (*NbStatus, error) {
	i := bytes.IndexByte(out, '{')
	if i < 0 {
		return nil, errors.New("unknown output")
	}
	var st NbStatus
	if err := json.Unmarshal(out[i:], &st); err != nil {
		return nil, err
	}
	return &st, nil
}

// toVpnStatus is the page's view of the CLI's status. Uptime, memory, start at
// boot and the blocker are vpn.Fill's.
func toVpnStatus(nb *NbStatus) (proto.VpnStatus, error) {
	state, ok := StateMap[nb.DaemonStatus]
	if !ok {
		// A later NetBird may add a status. The daemon answered, so it runs,
		// and the page keeps its details rather than an error.
		state = proto.VpnRunning
		if _, seen := unknownStatuses.LoadOrStore(nb.DaemonStatus, true); !seen {
			log.Warnf("unknown netbird daemon status %q, shown as running", nb.DaemonStatus)
		}
	}
	st := proto.VpnStatus{
		State:   state,
		Version: nb.DaemonVersion,
		IP:      stripPrefix(nb.IP),
		Name:    nb.FQDN,
		Account: managementHost(nb.Management.URL),
		Control: nb.Management.Connected,
		Peers:   []proto.VpnPeer{},
	}
	for _, p := range nb.Peers.Details {
		st.Peers = append(st.Peers, proto.VpnPeer{
			Name:   peerName(p.FQDN),
			IP:     stripPrefix(p.IP),
			Online: p.Status == "Connected",
			Idle:   p.Status == "Idle",
		})
	}
	sort.Slice(st.Peers, func(i, j int) bool { return st.Peers[i].Name < st.Peers[j].Name })
	return st, nil
}

// stripPrefix drops the network length: 100.73.212.105/16 is 100.73.212.105.
func stripPrefix(ip string) string {
	addr, _, _ := strings.Cut(ip, "/")
	return addr
}

// managementHost is the account the page shows: the management server's host.
func managementHost(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return raw
	}
	return u.Hostname()
}

// peerName is the first label of the peer's FQDN.
func peerName(fqdn string) string {
	if label, _, _ := strings.Cut(fqdn, "."); label != "" {
		return label
	}
	return fqdn
}
