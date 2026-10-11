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
