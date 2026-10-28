package proto

// VpnState is where a VPN add-on stands, as the settings page shows it.
type VpnState string

const (
	VpnNotInstall VpnState = "notInstall"
	VpnNotRunning VpnState = "notRunning"
	VpnNotLogin   VpnState = "notLogin"
	VpnStopped    VpnState = "stopped"
	VpnRunning    VpnState = "running"
)

// VpnStatus is the status of Tailscale or NetBird, in one shape for both.
type VpnStatus struct {
	State       VpnState  `json:"state"`
	Version     string    `json:"version"`
	IP          string    `json:"ip"`
	Name        string    `json:"name"`    // host name or FQDN
	Account     string    `json:"account"` // tailnet name; NetBird management URL host
	Control     bool      `json:"control"` // connected to the coordination or management server
	Peers       []VpnPeer `json:"peers"`
	UptimeSec   int64     `json:"uptimeSec"` // of the daemon process, 0 when not running
	BootEnabled bool      `json:"bootEnabled"`
	Memory      VpnMemory `json:"memory"`
	BlockedBy   string    `json:"blockedBy"` // "tailscale" or "netbird" when the other one runs or starts at boot
	// LoginServer is Tailscale's control server when it is not Tailscale's
	// own, a Headscale server for example. Empty otherwise, and for NetBird.
	LoginServer string `json:"loginServer,omitempty"`
}

type VpnPeer struct {
	Name   string `json:"name"`
	IP     string `json:"ip"`
	Online bool   `json:"online"`
	// Idle is a NetBird peer with no tunnel yet under lazy connections: the
	// daemon opens one when traffic needs it, and cannot tell whether the
	// peer is up until then.
	Idle bool `json:"idle,omitempty"`
}

// VpnMemory is in bytes. GroupHigh and GroupMax are 0 when the group sets no
// limit or does not exist.
type VpnMemory struct {
	DaemonRSS    int64 `json:"daemonRss"`
	GroupCurrent int64 `json:"groupCurrent"`
	GroupHigh    int64 `json:"groupHigh"`
	GroupMax     int64 `json:"groupMax"`
}

type VpnBootReq struct {
	Enabled bool `json:"enabled"`
}

// VpnUpdateRsp is the installed version and the release an update installs,
// the one the server pins.
type VpnUpdateRsp struct {
	Current string `json:"current"`
	Latest  string `json:"latest"`
	// Pinned is the release the server was tested with, where an add-on can
	// install a newer one than it (NetBird, from Alpine's index).
	Pinned string `json:"pinned,omitempty"`
}

type VpnLoginRsp struct {
	Url string `json:"url"`
}

// NetbirdLoginReq joins with a setup key. An empty key asks for SSO instead.
type NetbirdLoginReq struct {
	SetupKey string `json:"setupKey"`
}
