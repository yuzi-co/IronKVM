package proto

// The Tailscale names predate NetBird. They stay, as aliases, for the code
// and the clients that use them.
type TailscaleState = VpnState

const (
	TailscaleNotInstall = VpnNotInstall
	TailscaleNotRunning = VpnNotRunning
	TailscaleNotLogin   = VpnNotLogin
	TailscaleStopped    = VpnStopped
	TailscaleRunning    = VpnRunning
)

// GetTailscaleStatusRsp keeps state, name, ip and account where old clients
// read them, and carries the rest of VpnStatus beside them.
type GetTailscaleStatusRsp = VpnStatus

type LoginTailscaleRsp = VpnLoginRsp

// TailscaleLoginServerReq sets the control server: an https URL, such as a
// Headscale server's, or empty for Tailscale's own.
type TailscaleLoginServerReq struct {
	LoginServer string `json:"loginServer"`
}

// TailscaleLoginServerRsp is the server now in use and, when the change
// started a login, the URL to finish it at.
type TailscaleLoginServerRsp struct {
	LoginServer string `json:"loginServer"`
	Url         string `json:"url"`
}
