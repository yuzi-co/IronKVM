package proto

type WakeOnLANReq struct {
	Mac string `form:"mac" validate:"required"`
}

type GetMacRsp struct {
	Macs []string `json:"macs"`
}

type DeleteMacReq struct {
	Mac string `form:"mac" validate:"required"`
}

type SetMacNameReq struct {
	Mac  string `form:"mac" validate:"required"`
	Name string `form:"name" validate:"required"`
}

type GetWifiRsp struct {
	Supported bool   `json:"supported"`
	ApMode    bool   `json:"apMode"`
	Connected bool   `json:"connected"`
	Ssid      string `json:"ssid"`
}

// ConnectWifiReq names the network to join. An empty password joins an open
// network; otherwise it is a WPA passphrase of 8 to 63 characters, which is
// what wpa_passphrase in S30wifi takes.
type ConnectWifiReq struct {
	Ssid     string `validate:"required,max=32"`
	Password string `validate:"omitempty,min=8,max=63"`
}

type GetDNSRsp struct {
	Mode      string   `json:"mode"`
	Servers   []string `json:"servers"`
	Effective []string `json:"effective"`
	DHCP      []string `json:"dhcp"`
	Info      DNSInfo  `json:"info"`
}

type SetDNSReq struct {
	Mode    string   `json:"mode" validate:"required,oneof=manual dhcp"`
	Servers []string `json:"servers"`
}

type DNSInfo struct {
	Interface     string   `json:"interface"`
	Type          string   `json:"type"`
	Address       string   `json:"address"`
	SubnetMask    string   `json:"subnetMask"`
	Gateway       string   `json:"gateway"`
	SearchDomains []string `json:"searchDomains"`
}
