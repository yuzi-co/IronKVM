package proto

type GetHidModeRsp struct {
	Mode string `json:"mode"` // normal or hid-only
	// ExtendedKeys is true when the gadget carries the Consumer and System
	// Control reports, which only normal mode's descriptor declares.
	ExtendedKeys bool `json:"extendedKeys"`
	// HidDisabled is true when /boot/disable_hid leaves the keyboard and both
	// pointers out of the gadget.
	HidDisabled bool `json:"hidDisabled"`
}

// SendHidKeyReq presses and releases one key outside the keyboard: a
// Consumer Control usage (media, volume) or a System Control usage (power,
// sleep, wake).
type SendHidKeyReq struct {
	Page  string `json:"page" form:"page" validate:"required,oneof=consumer system"`
	Usage int    `json:"usage" form:"usage" validate:"required"`
}

type GetKeyboardLedStatusRsp struct {
	NumLock    bool   `json:"numLock"`
	CapsLock   bool   `json:"capsLock"`
	ScrollLock bool   `json:"scrollLock"`
	Known      bool   `json:"known"`
	UpdatedAt  string `json:"updatedAt"`
}

// HidDeviceStatus reports whether the target is fetching reports from one HID
// endpoint.
//
// State carries a code rather than a sentence because the web UI translates it
// into about twenty languages. Detail holds the raw error text for an operator
// reading the response by hand, and is empty for the states that need none.
//
// ObservedMsAgo is the age of the observation. Nothing writes to an endpoint the
// operator has switched away from, so a stalled state goes stale rather than
// clearing, and a consumer must be able to tell the two apart.
//
// WasAccepting says the target has fetched a report from this endpoint since
// the gadget last enumerated. A stall without it is most likely a host with no
// driver for the endpoint, which never polls it and is not at fault, and the
// web UI does not warn about that one.
//
// "stalled" and "detached" are different faults and want different answers.
// Stalled is one endpoint the target has stopped polling on a working link, and
// the operator's remedies are a different mouse mode or a USB reset. Detached
// is no link at all, so every endpoint is dead together, and the server's own
// gadget supervisor repairs that without being asked.
type HidDeviceStatus struct {
	Name          string `json:"name"` // keyboard, mouse-relative, mouse-absolute
	Path          string `json:"path"`
	State         string `json:"state"` // unknown, accepting, stalled, detached, error
	Detail        string `json:"detail,omitempty"`
	StateForMs    int64  `json:"stateForMs"`
	ObservedMsAgo int64  `json:"observedMsAgo"`
	WasAccepting  bool   `json:"wasAccepting"`
}

type GetHidStatusRsp struct {
	Devices []HidDeviceStatus `json:"devices"`
}

type SetHidModeReq struct {
	Mode string `validate:"required"` // normal or hid-only
}

type ShortcutKey struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

type Shortcut struct {
	ID   string        `json:"id"`
	Keys []ShortcutKey `json:"keys"`
}

type GetShortcutsRsp struct {
	Shortcuts []Shortcut `json:"shortcuts"`
}

type AddShortcutReq struct {
	Keys []ShortcutKey `validate:"required"`
}

type DeleteShortcutReq struct {
	ID string `validate:"required"`
}

type SetLeaderKeyReq struct {
	Key string `validate:"omitempty"`
}

type GetLeaderKeyRsp struct {
	Key string `json:"key"`
}

// PasteReq types Content on the target as key presses, in the background.
//
// Layout names the keyboard layout active on the target, because a key press
// types whatever that layout puts on the key. Langue is the name an older web
// client sends it under. Delay is the pause after each key press in
// milliseconds, 0 for the default. A text holding characters the layout
// cannot type is refused with the list of them, unless SkipUntypeable says to
// type the rest.
type PasteReq struct {
	Content        string `json:"content" form:"content" validate:"required"`
	Layout         string `json:"layout" form:"layout"`
	Langue         string `json:"langue" form:"langue"`
	Delay          int    `json:"delay" form:"delay"`
	SkipUntypeable bool   `json:"skipUntypeable" form:"skipUntypeable"`
}

// PasteCheckReq asks what typing Content would take, without typing it.
type PasteCheckReq struct {
	Content string `json:"content" form:"content" validate:"required"`
	Layout  string `json:"layout" form:"layout"`
	Delay   int    `json:"delay" form:"delay"`
}

// PasteUntypeable is one character the layout cannot type. Index counts code
// points from the start of the text; Line and Column count from 1, and a
// CRLF ends one line.
type PasteUntypeable struct {
	Index  int    `json:"index"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Char   string `json:"char"`
}

// PasteCheckRsp describes what typing a text takes. Untypeable lists at most
// the first hundred characters the layout cannot type, and UntypeableCount
// counts all of them.
type PasteCheckRsp struct {
	Characters      int               `json:"characters"`
	Keystrokes      int               `json:"keystrokes"`
	DurationMs      int64             `json:"durationMs"`
	Untypeable      []PasteUntypeable `json:"untypeable"`
	UntypeableCount int               `json:"untypeableCount"`
}

// PasteStatusRsp reports the paste typing in the background, or the last one.
//
// Status is idle, typing, done, canceled or failed. Typed and Total count
// characters. Error is a code the web UI translates, set when Status is
// failed: control_busy when the keyboard could not be taken, hid_error when
// a report could not be written.
type PasteStatusRsp struct {
	Status string `json:"status"`
	Typed  int    `json:"typed"`
	Total  int    `json:"total"`
	Error  string `json:"error,omitempty"`
}
