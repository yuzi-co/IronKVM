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
