package proto

// WakeOnLANReq names the device by MAC or, for a listed device, by name.
type WakeOnLANReq struct {
	Mac  string `form:"mac"`
	Name string `form:"name"`
}

// WolDevice is a device of the list. The SecureOn password is never sent back:
// HasSecureOn says only that one is stored.
type WolDevice struct {
	Mac         string `json:"mac"`
	Name        string `json:"name"`
	Broadcast   string `json:"broadcast"`
	Port        int    `json:"port"`
	HasSecureOn bool   `json:"hasSecureOn"`
	IP          string `json:"ip"`
}

type GetWolDevicesRsp struct {
	Devices []WolDevice `json:"devices"`
}

// SaveWolDeviceReq adds a device, or edits the one at OldMac. An edit that
// leaves SecureOn empty keeps the stored password unless ClearSecureOn is set.
type SaveWolDeviceReq struct {
	OldMac        string `json:"oldMac"`
	Mac           string `json:"mac" validate:"required"`
	Name          string `json:"name"`
	Broadcast     string `json:"broadcast"`
	Port          int    `json:"port"`
	SecureOn      string `json:"secureOn"`
	ClearSecureOn bool   `json:"clearSecureOn"`
	IP            string `json:"ip"`
}

type CheckWolDeviceRsp struct {
	Up bool `json:"up"`
}

type WolNeighbor struct {
	Mac  string `json:"mac"`
	IP   string `json:"ip"`
	Name string `json:"name"`
	// Listed says the MAC is already in the device list.
	Listed bool `json:"listed"`
}

type ScanWolRsp struct {
	Neighbors []WolNeighbor `json:"neighbors"`
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
