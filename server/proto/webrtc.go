package proto

// GetWebRTCForwardRsp is the port forward setting. A zero UDPPort leaves
// WebRTC as it has always been.
type GetWebRTCForwardRsp struct {
	// UDPPort and PublicAddress are the saved setting.
	UDPPort       int    `json:"udpPort"`
	PublicAddress string `json:"publicAddress"`
	// ActivePort is the shared port new viewers get now. It differs from
	// UDPPort only when the port could not be opened.
	ActivePort int `json:"activePort"`
	// Resolved are the addresses the public address stands for at this
	// moment, the ones a viewer outside the LAN is offered.
	Resolved []string `json:"resolved"`
}

type SetWebRTCForwardReq struct {
	UDPPort       int    `json:"udpPort"`
	PublicAddress string `json:"publicAddress"`
}

// StunServerRsp is the STUN server setting: host:port, or "disable" when STUN
// is off. Default is the server a board uses until its owner sets another.
type StunServerRsp struct {
	Stun    string `json:"stun"`
	Default string `json:"default"`
}

// SetStunServerReq sets the STUN server: host:port, "disable" to turn STUN
// off, or empty for the default.
type SetStunServerReq struct {
	Stun string `json:"stun"`
}

// WHEPSettingsRsp is the WHEP endpoint's setting and state
// (ironkvm-dist#145). Path is where a player POSTs its offer.
type WHEPSettingsRsp struct {
	Enabled     bool   `json:"enabled"`
	Path        string `json:"path"`
	Sessions    int    `json:"sessions"`
	MaxSessions int    `json:"maxSessions"`
}

// SetWHEPSettingsReq switches the WHEP endpoint on or off.
type SetWHEPSettingsReq struct {
	Enabled *bool `json:"enabled"`
}
