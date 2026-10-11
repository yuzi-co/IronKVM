package proto

type GetIPv6Rsp struct {
	// Supported says the kernel has IPv6. Without it the rest is empty.
	Supported bool `json:"supported"`
	// Mode is what the device does: "slaac", "dhcpv6", "static" or "off".
	Mode string `json:"mode"`
	// Address, Prefix and Gateway are the saved static settings. They stay
	// filled after a switch to another mode, so the form can offer them again.
	Address string `json:"address"`
	Prefix  int    `json:"prefix"`
	Gateway string `json:"gateway"`
	// DHCPv6Available says the image carries a DHCPv6 client.
	DHCPv6Available bool `json:"dhcpv6Available"`
	// Client is the address this request came from, so the page can tell
	// whether the change it is about to make is made over IPv6.
	Client string   `json:"client"`
	Live   IPv6Live `json:"live"`
}

type IPv6Live struct {
	Interface string        `json:"interface"`
	Disabled  bool          `json:"disabled"`
	Addresses []IPv6Address `json:"addresses"`
	Gateway   string        `json:"gateway"`
}

type IPv6Address struct {
	Address string `json:"address"`
	Prefix  int    `json:"prefix"`
	// Scope is "link", "ula" or "global".
	Scope string `json:"scope"`
}

type SetIPv6Req struct {
	Mode    string `json:"mode" validate:"required,oneof=slaac dhcpv6 static off"`
	Address string `json:"address"`
	Prefix  int    `json:"prefix"`
	Gateway string `json:"gateway"`
}
