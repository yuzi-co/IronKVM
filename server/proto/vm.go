package proto

type IP struct {
	Name    string `json:"name"`
	Addr    string `json:"addr"`
	Version string `json:"version"`
	Type    string `json:"type"`
}

type GetInfoRsp struct {
	IPs         []IP   `json:"ips"`
	Mdns        string `json:"mdns"`
	Image       string `json:"image"`
	Application string `json:"application"`
	Base        string `json:"base,omitempty"`
	DeviceKey   string `json:"deviceKey"`
}

type GetHardwareRsp struct {
	Version string `json:"version"`
}

type SetGpioReq struct {
	Type     string `validate:"required"`  // reset / power
	Duration uint   `validate:"omitempty"` // press time (unit: milliseconds)
}

type GetGpioRsp struct {
	PWR bool `json:"pwr"` // power led
	HDD bool `json:"hdd"` // hdd led
}

type SetScreenReq struct {
	Type  string `validate:"required"` // resolution / fps / quality
	Value int    `validate:"number"`   // value
}

type GetScriptsRsp struct {
	Files []string `json:"files"`
}

type UploadScriptRsp struct {
	File string `json:"file"`
}

type RunScriptReq struct {
	Name string `validate:"required"`
	Type string `validate:"required"` // foreground | background
}

type RunScriptRsp struct {
	Log string `json:"log"`
}

type DeleteScriptReq struct {
	Name string `validate:"required"`
}

// autostart
type GetAutostartRsp struct {
	Files []string `json:"files"`
}

type UploadAutostartReq struct {
	Content string `json:"content"`
}

// VirtualDeviceState separates what was asked for from what is running. The two
// differ when the USB controller ran out of endpoints and the boot script gave
// the function up: the marker is still there, and the function is not.
type VirtualDeviceState struct {
	Enabled bool `json:"enabled"`
	Active  bool `json:"active"`
	Cost    int  `json:"cost"`
}

type GetVirtualDeviceRsp struct {
	Console VirtualDeviceState `json:"console"`
	Network VirtualDeviceState `json:"network"`
	Disk    VirtualDeviceState `json:"disk"`
	Audio   VirtualDeviceState `json:"audio"`
	Used    int                `json:"used"`
	Total   int                `json:"total"`
}

type UpdateVirtualDeviceReq struct {
	Device string `validate:"required"`
}

type UpdateVirtualDeviceRsp struct {
	On bool `json:"on"`
}

type SetMemoryLimitReq struct {
	Enabled bool  `validate:"omitempty"`
	Limit   int64 `validate:"omitempty"`
}

type GetMemoryLimitRsp struct {
	Enabled bool  `json:"enabled"`
	Limit   int64 `json:"limit"`
}

type GetIonRsp struct {
	Total       uint64 `json:"total"`
	Used        uint64 `json:"used"`
	Free        uint64 `json:"free"`
	UsageRate   int    `json:"usageRate"`
	Generations int    `json:"generations"`
	Reserve     uint64 `json:"reserve"`
	Measured    bool   `json:"measured"`
	Verdict     string `json:"verdict"`
}

// SetOledReq carries only the settings the caller means to change. The fields
// are pointers because zero is a real value for both of them: a sleep of 0
// keeps the screen on for good, and the request has to say the difference
// between asking for that and not mentioning sleep at all.
type SetOledReq struct {
	Sleep      *int `json:"sleep" validate:"omitempty"`
	Brightness *int `json:"brightness" validate:"omitempty"`
}

type GetOLEDRsp struct {
	Exist      bool `json:"exist"`
	Sleep      int  `json:"sleep"`
	Brightness int  `json:"brightness"`

	// BrightnessSupported reports whether the running kvm_system acts on the
	// brightness setting. A release carries Sipeed's build, which does not, and
	// a control that writes a file nothing reads is worse than no control.
	BrightnessSupported bool `json:"brightnessSupported"`
}

type GetGetHdmiStateRsp struct {
	// Enabled reports whether capture is switched on in software.
	Enabled bool `json:"enabled"`

	// Signal reports whether the port is actually carrying a picture, which
	// is how a caller tells a sleeping machine from an awake one.
	Signal bool `json:"signal"`

	IdleTimeout int `json:"idleTimeout"`
}

type SetHdmiIdleTimeoutReq struct {
	Minutes int `validate:"gte=0,lte=10080"`
}

type GetSSHStateRsp struct {
	Enabled bool `json:"enabled"`
}

type GetSwapRsp struct {
	Size int64 `json:"size"` // unit: MB
}

type SetSwapReq struct {
	Size int64 `validate:"omitempty"` // unit: MB
}

// GetZramRsp separates three questions that a single on/off flag would merge.
// Available and Enabled can each be true while the device does not run.
type GetZramRsp struct {
	Available bool `json:"available"` // the kernel modules are installed
	Enabled   bool `json:"enabled"`   // the setting survives a reboot
	Active    bool `json:"active"`    // compressed swap runs now

	Algorithm  string `json:"algorithm"`
	DiskSize   int64  `json:"diskSize"`   // unit: bytes
	Original   int64  `json:"original"`   // unit: bytes, before compression
	Compressed int64  `json:"compressed"` // unit: bytes, after compression
	MemUsed    int64  `json:"memUsed"`    // unit: bytes
	MemLimit   int64  `json:"memLimit"`   // unit: bytes, 0 when unset

	// SwapIn and SwapOut are system-wide page counters. They cover every swap
	// device, not zram alone, and they do not reset when zram restarts.
	SwapIn  int64 `json:"swapIn"`
	SwapOut int64 `json:"swapOut"`
}

type SetZramReq struct {
	Enabled bool `validate:"omitempty"`
}

// GetCpuFreqRsp reports the CPU clock. Running is what the core runs now, read
// from the clock registers; Target is what the next boot applies. They differ
// after a change until the operator reboots, which is the only safe moment to
// switch the clock.
type GetCpuFreqRsp struct {
	Running        int     `json:"running"`        // MHz now, 0 when the registers cannot be decoded
	Measured       bool    `json:"measured"`       // Running was decoded from the clock registers
	Target         int     `json:"target"`         // MHz the next boot applies
	Temperature    float64 `json:"temperature"`    // CPU temperature, degrees C, 0 when unavailable
	Options        []int   `json:"options"`        // selectable frequencies, MHz
	RebootRequired bool    `json:"rebootRequired"` // Running differs from Target, so a reboot is due
}

type SetCpuFreqReq struct {
	Target int `validate:"required"` // MHz, must be one of GetCpuFreqRsp.Options
}

type GetMouseJigglerRsp struct {
	Enabled bool   `json:"enabled"`
	Mode    string `json:"mode"`
}

type SetMouseJigglerReq struct {
	Enabled bool   `validate:"omitempty"`
	Mode    string `validate:"omitempty"`
}

type GetMdnsStateRsp struct {
	Enabled bool `json:"enabled"`
}

type SetHostnameReq struct {
	Hostname string `validate:"required"`
}

type GetHostnameRsp struct {
	Hostname string `json:"hostname"`
}

type SetWebTitleReq struct {
	Title string `validate:"omitempty"`
}

type GetWebTitleRsp struct {
	Title string `json:"title"`
}

type SetTlsReq struct {
	Enabled bool `validate:"omitempty"`
}

type InputRegion struct {
	Mode               string               `json:"mode"`
	FrameWidth         int                  `json:"frameWidth"`
	FrameHeight        int                  `json:"frameHeight"`
	Left               int                  `json:"left"`
	Top                int                  `json:"top"`
	Width              int                  `json:"width"`
	Height             int                  `json:"height"`
	Resolutions        []OriginalResolution `json:"resolutions,omitempty"`
	SelectedResolution string               `json:"selectedResolution"`
	Regions            []ManualRegion       `json:"regions,omitempty"`
	SelectedRegion     string               `json:"selectedRegion"`
}

type ManualRegion struct {
	FrameWidth  int `json:"frameWidth"`
	FrameHeight int `json:"frameHeight"`
	Left        int `json:"left"`
	Top         int `json:"top"`
	Width       int `json:"width"`
	Height      int `json:"height"`
}

type OriginalResolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type SetInputRegionReq struct {
	Mode               string                `json:"mode"`
	FrameWidth         *int                  `json:"frameWidth,omitempty"`
	FrameHeight        *int                  `json:"frameHeight,omitempty"`
	Left               *int                  `json:"left,omitempty"`
	Top                *int                  `json:"top,omitempty"`
	Width              *int                  `json:"width,omitempty"`
	Height             *int                  `json:"height,omitempty"`
	Resolutions        *[]OriginalResolution `json:"resolutions,omitempty"`
	SelectedResolution *string               `json:"selectedResolution,omitempty"`
	Regions            *[]ManualRegion       `json:"regions,omitempty"`
	SelectedRegion     *string               `json:"selectedRegion,omitempty"`
}

type GetInputRegionRsp struct {
	InputRegion
}

type GetInputResolutionRsp struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}
