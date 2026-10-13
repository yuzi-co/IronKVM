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
	Kernel      string `json:"kernel"`
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
	// HasHDD says the board has an HDD LED input at all. Only the alpha board
	// does; on the others HDD is always false and means nothing.
	HasHDD bool `json:"hasHdd"`
	// LEDConnected says the power LED header is wired. When it is not, PWR
	// is always false and means nothing: the state is unknown.
	LEDConnected bool `json:"ledConnected"`
}

// PowerLEDSetting is whether the host's power LED header is wired to the
// board, as the owner set it.
type PowerLEDSetting struct {
	Connected bool `json:"connected" validate:"omitempty"`
}

type SetScreenReq struct {
	Type  string `validate:"required"` // resolution / fps / quality
	Value int    `validate:"number"`   // value
}

// GetScreenRsp is what the server currently holds for the capture pipeline.
//
// The browser used to keep its own copy of all of this and push it on load,
// which meant a second viewer saw its own settings rather than the board's.
// This is the read that ends that, so it reports the values a stream would
// actually use and not the raw contents of the settings files.
//
// Quality and BitRate are separate here although one API key carries both: the
// key's value decides which of the two it means, and the browser has no way to
// make that distinction on the way back.
type GetScreenRsp struct {
	Width   uint16 `json:"width"`
	Height  uint16 `json:"height"`
	Quality uint16 `json:"quality"` // JPEG quality, for MJPEG
	BitRate uint16 `json:"bitRate"` // kbit/s, for both H.264 paths
	FPS     int    `json:"fps"`
	GOP     uint8  `json:"gop"`
	Codec   uint8  `json:"codec"` // 1 H.264, 2 H.265
	// Aspect is 0 to keep the source's aspect ratio, 1 to stretch it to the
	// resolution. AspectSupported is false with a capture library that always
	// stretches (Sipeed's), where the menu leaves the choice out.
	Aspect          uint8 `json:"aspect"`
	AspectSupported bool  `json:"aspectSupported"`
	// Type is the mode the encoder runs, "h264" or "mjpeg", or empty when the
	// board has not recorded one. A view-only session cannot change it, so it
	// follows this to choose the stream that has a picture.
	Type string `json:"type,omitempty"`
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
	// Fits lists every largest set of optional functions that fits beside
	// HID, so the UI can say which functions go together before a refusal.
	Fits [][]string `json:"fits"`
}

// GetUSBNetworkRsp is the USB network link to the managed host. Mode is off,
// ncm or ecm, or rndis on a board an older server set up. Fits answers whether
// the network can be turned on now, and Refusal says why when it cannot.
type GetUSBNetworkRsp struct {
	Mode    string `json:"mode"`
	Subnet  string `json:"subnet"`
	Board   string `json:"board"`
	Host    string `json:"host"`
	Active  bool   `json:"active"`
	Fits    bool   `json:"fits"`
	Refusal string `json:"refusal"`
}

// SetUSBNetworkReq picks the mode and, optionally, the subnet. An empty
// subnet keeps the one in use.
type SetUSBNetworkReq struct {
	Mode   string `validate:"required,oneof=off ncm ecm"`
	Subnet string `validate:"omitempty,max=18"`
}

type UpdateVirtualDeviceReq struct {
	Device string `validate:"required"`
}

type UpdateVirtualDeviceRsp struct {
	On bool `json:"on"`
}

// ApplyVirtualDeviceReq is every optional USB function the gadget should carry,
// all at once, so that one rebuild applies them together. Network.Mode may be
// rndis only on a board that already runs it, which keeps the link as it is.
type ApplyVirtualDeviceReq struct {
	Console bool                   `json:"console"`
	Disk    bool                   `json:"disk"`
	Audio   bool                   `json:"audio"`
	Network ApplyUSBNetworkRequest `json:"network"`
}

// ApplyUSBNetworkRequest is the link part of ApplyVirtualDeviceReq. An empty
// subnet keeps the one in use.
type ApplyUSBNetworkRequest struct {
	Mode   string `json:"mode" validate:"required,oneof=off ncm ecm rndis"`
	Subnet string `json:"subnet" validate:"omitempty,max=18"`
}

// ApplyVirtualDeviceRsp is the state after an apply: what GET
// /api/vm/device/virtual returns, with the USB network link beside it.
type ApplyVirtualDeviceRsp struct {
	GetVirtualDeviceRsp
	USBNetwork GetUSBNetworkRsp `json:"usbNetwork"`
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
	// Enabled is the owner's setting: sshd starts at boot.
	Enabled bool `json:"enabled"`

	// Running is whether sshd is up now.
	Running bool `json:"running"`

	// Port is the first port sshd listens on, 22 when it cannot be read.
	Port int `json:"port"`

	// KeysOnly is the owner's setting that turns password logins off.
	KeysOnly bool `json:"keysOnly"`

	// PasswordAuth is what sshd itself reports: "yes", "no", or "" when
	// sshd could not be asked. It differs from KeysOnly when the image's sshd
	// does not read the drop-in.
	PasswordAuth string `json:"passwordAuth"`

	// KeyCount is the number of usable authorized keys for root.
	KeyCount int `json:"keyCount"`

	HostKeys []SSHKey `json:"hostKeys"`

	// RootPassword is "default" (a factory password), "empty" (no password
	// at all), "set", or "unknown". The hash itself is never returned.
	RootPassword string `json:"rootPassword"`
}

// SSHKey describes a public key. Comment is empty for host keys.
type SSHKey struct {
	Type        string `json:"type"`
	Comment     string `json:"comment"`
	Fingerprint string `json:"fingerprint"`
}

type GetSSHKeysRsp struct {
	Keys []SSHKey `json:"keys"`
}

type AddSSHKeyReq struct {
	Key string `json:"key" validate:"required"`
}

type DeleteSSHKeyReq struct {
	Fingerprint string `json:"fingerprint" validate:"required"`
}

type SetSSHKeysOnlyReq struct {
	Enabled bool `json:"enabled" validate:"omitempty"`
}

// SetSSHPortReq changes the port sshd listens on. The range is checked by the
// service, which answers an out-of-range port with its own code.
type SetSSHPortReq struct {
	Port int `json:"port" validate:"omitempty"`
}

type GetSwapRsp struct {
	Size   int64 `json:"size"`   // unit: MB, the swap file's size on disk
	Active bool  `json:"active"` // the kernel swaps to the file now
	Total  int64 `json:"total"`  // unit: bytes, as /proc/swaps reports it
	Used   int64 `json:"used"`   // unit: bytes
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

// VideoClockRates are the video clocks in MHz: the codec core (vc-src0), the
// VPSS and capture (vip-sys1), and the codec bus (axi-video-codec).
type VideoClockRates struct {
	Codec int `json:"codec"`
	Vpss  int `json:"vpss"`
	Bus   int `json:"bus"`
}

// VideoClockOption is one selectable video clock profile and its rates.
type VideoClockOption struct {
	Name  string          `json:"name"`
	Rates VideoClockRates `json:"rates"`
}

// GetVideoClockRsp reports the video clock profile of the mainline kernel.
// Rates is what the clock framework has now; it lags Profile (Pending) while
// the video that ran before a change still runs.
type GetVideoClockRsp struct {
	Profile     string             `json:"profile"`     // the chosen profile, stored in /etc/kvm/video-clock
	Options     []VideoClockOption `json:"options"`     // selectable profiles
	Rates       VideoClockRates    `json:"rates"`       // MHz now, zeros when unreadable
	Measured    bool               `json:"measured"`    // Rates was read from the kernel
	Pending     bool               `json:"pending"`     // Rates differ from Profile's until the video next starts
	Temperature float64            `json:"temperature"` // SoC temperature, degrees C, 0 when unavailable
}

// GetHealthRsp is what the toolbar's alert icon needs, read in one cheap
// request: no command is run and nothing is fetched from the network.
type GetHealthRsp struct {
	// Temperature is the SoC temperature in degrees C, or null when the board
	// has no sensor to read.
	Temperature *float64 `json:"temperature"`
	// Storage is the filesystem that holds the images, or null when it cannot
	// be read.
	Storage *HealthStorage `json:"storage"`
	// Vpn lists the VPN add-ons set to start at boot, which are the ones the
	// owner expects to be up.
	Vpn []HealthVpn `json:"vpn"`
}

type HealthStorage struct {
	Path      string `json:"path"`
	Total     uint64 `json:"total"`     // bytes
	Available uint64 `json:"available"` // bytes an unprivileged writer may still use
}

type HealthVpn struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Running bool   `json:"running"`
}

// GetMemoryRsp is the board's memory at a glance, for the Performance page.
// Every size is in bytes.
type GetMemoryRsp struct {
	Total     int64 `json:"total"`     // MemTotal
	Available int64 `json:"available"` // MemAvailable: what can be had without swapping
	Free      int64 `json:"free"`      // MemFree

	Swaps []MemorySwap `json:"swaps"` // each device in /proc/swaps
	// ZramMemUsed is the RAM zram holds for what it stores, its own overhead
	// included. 0 without zram.
	ZramMemUsed int64 `json:"zramMemUsed"`

	Processes []MemoryProcess `json:"processes"` // the main consumers that run now
	Addons    *MemoryGroup    `json:"addons"`    // the add-ons' cgroup, null when absent

	// VideoPools and Cma describe the mainline kernel's device memory. The
	// vendor kernel reports its ION carveout through GET /api/vm/ion instead,
	// and leaves both out.
	VideoPools []MemoryPool `json:"videoPools,omitempty"` // the reserved pools the video drivers use
	Cma        []MemoryCma  `json:"cma,omitempty"`        // CMA areas other than those pools
}

// MemoryPool is a reserved pool a device allocates from. When Measured is
// false the kernel publishes only its size, and the other sizes are 0.
type MemoryPool struct {
	Name     string `json:"name"` // the reserved-memory node: video-pool@86800000
	Size     int64  `json:"size"`
	Measured bool   `json:"measured"`
	Used     int64  `json:"used"`
	Free     int64  `json:"free"`
	// LargestFree is the longest free run, absent when unknown.
	LargestFree *int64 `json:"largestFree,omitempty"`
	Peak        int64  `json:"peak"`   // since boot or the last reset; 0 when not kept
	Failed      int64  `json:"failed"` // allocations the pool refused
}

// MemoryCma is one CMA area. Used counts what cma_alloc handed out, unlike
// CmaFree in /proc/meminfo, which also counts the free pages that movable
// allocations borrow.
type MemoryCma struct {
	Name  string `json:"name"`
	Total int64  `json:"total"`
	Used  int64  `json:"used"`
	// LargestFree is the longest free run, absent without CMA debugfs.
	LargestFree *int64 `json:"largestFree,omitempty"`
	Failed      int64  `json:"failed"` // alloc_pages_fail, in pages
}

type MemorySwap struct {
	Name string `json:"name"` // as /proc/swaps names it: /swapfile, /dev/zram0
	Kind string `json:"kind"` // "zram", "file" or "partition"
	Size int64  `json:"size"`
	Used int64  `json:"used"`
}

type MemoryProcess struct {
	Name string `json:"name"`
	RSS  int64  `json:"rss"` // VmRSS
}

// MemoryGroup is a memory cgroup's use and its limits. A limit is 0 when unset.
type MemoryGroup struct {
	Current int64 `json:"current"`
	High    int64 `json:"high"`
	Max     int64 `json:"max"`
}

type SetCpuFreqReq struct {
	Target int `validate:"required"` // MHz, must be one of GetCpuFreqRsp.Options
}

type SetVideoClockReq struct {
	Profile string `validate:"required"` // one of GetVideoClockRsp.Options' names
}

type GetMouseJigglerRsp struct {
	Enabled bool   `json:"enabled"`
	Mode    string `json:"mode"`
}

type SetMouseJigglerReq struct {
	Enabled bool   `validate:"omitempty"`
	Mode    string `validate:"omitempty"`
}

// Key is the key the key jiggler presses: "f15", "shift" or "ctrl". It is
// reported while the key jiggler is off too.
type GetKeyJigglerRsp struct {
	Enabled bool   `json:"enabled"`
	Key     string `json:"key"`
}

// Key is optional; a request without one keeps the key already chosen.
type SetKeyJigglerReq struct {
	Enabled bool   `validate:"omitempty"`
	Key     string `validate:"omitempty,oneof=f15 shift ctrl"`
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

type GetTlsRsp struct {
	Enabled bool `json:"enabled"`
}

// TlsCertReq is a certificate chain and its private key, both PEM. The chain
// starts with the server's own certificate.
type TlsCertReq struct {
	Cert string `json:"cert" validate:"required"`
	Key  string `json:"key" validate:"required"`
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

// GetStreamInfoRsp is what the web UI's stream status bar shows of the board:
// the HDMI source's size, its refresh rate in millihertz (0 when the capture
// library does not say), and the frames per second the capture loop delivered
// over the last few seconds (-1 before any stream has run).
type GetStreamInfoRsp struct {
	Width      int `json:"width"`
	Height     int `json:"height"`
	Refresh    int `json:"refresh"`
	CaptureFps int `json:"captureFps"`
	// CaptureStamps: the capture loop's frames carry the time they were
	// captured (libkvm-v4l2), and H.264 Direct stamps its frames with it, so
	// a frame's age at drawing counts from its capture, not from the
	// encoder. Off with Sipeed's library.
	CaptureStamps bool `json:"captureStamps"`
	// BoardLatency is the board's share of each delivery path's latency
	// (feature info-latency), absent while no viewer was sent a frame
	// with capture times in the last second.
	BoardLatency *BoardLatency `json:"boardLatency,omitempty"`
}

// BoardLatency is, per delivery path, the median time in milliseconds from a
// frame's first line reaching the capture to the frame being handed to the
// network for a viewer, over the last second. A path no viewer used in that
// second is absent.
type BoardLatency struct {
	WebRTC *float64 `json:"webrtc,omitempty"`
	Direct *float64 `json:"direct,omitempty"`
}

// USBIdentity is what the gadget tells the host it is. IDs are four hex digits
// with a 0x prefix. An empty Serial is the board's own serial number.
type USBIdentity struct {
	VendorID     string `json:"vendorId"`
	ProductID    string `json:"productId"`
	Manufacturer string `json:"manufacturer"`
	Product      string `json:"product"`
	Serial       string `json:"serial"`
}

// USBIdentityPreset is one identity the settings offer by name.
type USBIdentityPreset struct {
	Name string `json:"name"`
	USBIdentity
}

// GetUSBIdentityRsp is the chosen identity and what the live gadget carries.
// Preset is default, a preset's name, or custom. BootOverride is true when
// /boot/usb.vid or /boot/usb.pid exist; they apply only while no identity is
// saved.
type GetUSBIdentityRsp struct {
	Preset       string              `json:"preset"`
	Identity     USBIdentity         `json:"identity"`
	Live         USBIdentity         `json:"live"`
	Presets      []USBIdentityPreset `json:"presets"`
	BootOverride bool                `json:"bootOverride"`
}

// SetUSBIdentityReq picks a preset, or custom with its own IDs and strings.
// For a named preset only Serial is read from the request.
type SetUSBIdentityReq struct {
	Preset       string `json:"preset" validate:"required,oneof=default logitech microsoft dell custom"`
	VendorID     string `json:"vendorId"`
	ProductID    string `json:"productId"`
	Manufacturer string `json:"manufacturer"`
	Product      string `json:"product"`
	Serial       string `json:"serial"`
}
