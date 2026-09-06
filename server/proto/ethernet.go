package proto

type GetEthernetRsp struct {
	// Mode is what the device does at the next boot: "dhcp" or "static".
	Mode string `json:"mode"`
	// Address, Prefix and Gateway are the saved static settings. They stay
	// filled after a switch back to DHCP, so the form can offer them again.
	Address string `json:"address"`
	Prefix  int    `json:"prefix"`
	Gateway string `json:"gateway"`
	// Live is what the interface carries now, which is not the saved
	// configuration while a trial runs.
	Live EthernetLive `json:"live"`
	// Trial describes an applied change that nobody has confirmed yet.
	Trial *EthernetTrial `json:"trial,omitempty"`
}

type EthernetLive struct {
	Interface string `json:"interface"`
	Address   string `json:"address"`
	Prefix    int    `json:"prefix"`
	Netmask   string `json:"netmask"`
	Gateway   string `json:"gateway"`
}

type EthernetTrial struct {
	// Token identifies this trial. A confirmation carries it, so a late
	// confirmation cannot commit a change that replaced it.
	Token string `json:"token"`
	// Mode, Address, Prefix and Gateway are what the trial applied.
	Mode    string `json:"mode"`
	Address string `json:"address"`
	Prefix  int    `json:"prefix"`
	Gateway string `json:"gateway"`
	// RemainingSeconds counts down to the automatic revert.
	RemainingSeconds int `json:"remainingSeconds"`
}

type SetEthernetReq struct {
	Mode    string `json:"mode" validate:"required,oneof=dhcp static"`
	Address string `json:"address"`
	Prefix  int    `json:"prefix"`
	Gateway string `json:"gateway"`
	// TrialSeconds is how long the device waits for a confirmation before it
	// puts the previous configuration back. Zero takes the default.
	TrialSeconds int `json:"trialSeconds"`
}

type SetEthernetRsp struct {
	Token        string `json:"token"`
	TrialSeconds int    `json:"trialSeconds"`
	// Address and Prefix say where to look for the device after the change.
	Address string `json:"address"`
	Prefix  int    `json:"prefix"`
}

type ConfirmEthernetReq struct {
	Token string `json:"token" validate:"required"`
}
