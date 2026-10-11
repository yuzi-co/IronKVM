package proto

// SequenceStep is one step of a timed input sequence: a macro, or the presses
// of a key auto-tap. Type picks the fields that apply.
//
//   - key: the keyboard's state after the step. Modifiers is the modifier
//     byte of the boot keyboard report and Keys holds at most six usage IDs
//     from the Keyboard page. An empty step releases every key.
//   - mouse: one pointer report. Absolute reports place the pointer at X, Y
//     (0 to 32767); relative ones move it by X, Y (-127 to 127). Buttons is the
//     button byte, Wheel and HWheel the scroll. HWheel exists on the relative
//     pointer only.
//   - delay: a pause of Ms milliseconds.
//   - atx: a press of the front-panel Button, power or reset, held for Ms
//     milliseconds, through the same call as POST /api/vm/gpio.
type SequenceStep struct {
	Type      string `json:"type"`
	Modifiers int    `json:"modifiers,omitempty"`
	Keys      []int  `json:"keys,omitempty"`
	Absolute  bool   `json:"absolute,omitempty"`
	Buttons   int    `json:"buttons,omitempty"`
	X         int    `json:"x,omitempty"`
	Y         int    `json:"y,omitempty"`
	Wheel     int    `json:"wheel,omitempty"`
	HWheel    int    `json:"hwheel,omitempty"`
	Ms        int    `json:"ms,omitempty"`
	Button    string `json:"button,omitempty"`
}

// AutoTapReq starts a key auto-tap: the server presses and releases one key or
// combination every IntervalMs until a stop condition. TimeoutS bounds it,
// StopOnInput ends it at the first key or click from a browser, and
// StopOnVideo ends it when the host's video mode changes. ATX, power or reset,
// presses that button first, so one request can restart the host and catch
// its firmware menu. Label is the name the status shows, such as "F2".
type AutoTapReq struct {
	Label       string `json:"label" form:"label"`
	Modifiers   int    `json:"modifiers" form:"modifiers"`
	Keys        []int  `json:"keys" form:"keys"`
	IntervalMs  int    `json:"intervalMs" form:"intervalMs"`
	TimeoutS    int    `json:"timeoutS" form:"timeoutS"`
	StopOnInput bool   `json:"stopOnInput" form:"stopOnInput"`
	StopOnVideo bool   `json:"stopOnVideo" form:"stopOnVideo"`
	ATX         string `json:"atx" form:"atx"`
	ATXMs       int    `json:"atxMs" form:"atxMs"`
}

// SequenceStatusRsp reports the sequence running now, or the last one.
//
// Kind is autotap or macro, and empty before the first run. Step counts the
// steps done and Total the steps of a macro; Taps counts the presses of an
// auto-tap. Reason says why the last run ended: finished, stopped, timeout,
// input (a key or click from a browser), video (the video mode changed),
// shutdown, busy (the keyboard could not be taken), hid_error or atx_error.
type SequenceStatusRsp struct {
	Running   bool   `json:"running"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Step      int    `json:"step"`
	Total     int    `json:"total"`
	Taps      int    `json:"taps"`
	ElapsedMs int64  `json:"elapsedMs"`
	TimeoutMs int64  `json:"timeoutMs"`
	Reason    string `json:"reason"`
}

// Macro is a recorded sequence kept on the device.
type Macro struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Steps []SequenceStep `json:"steps"`
}

type GetMacrosRsp struct {
	Macros []Macro `json:"macros"`
}

// SaveMacroReq adds a macro, or replaces the one ID names.
type SaveMacroReq struct {
	ID    string         `json:"id" form:"id"`
	Name  string         `json:"name" form:"name"`
	Steps []SequenceStep `json:"steps" form:"steps"`
}

type DeleteMacroReq struct {
	ID string `json:"id" form:"id" validate:"required"`
}

// PlayMacroReq plays the saved macro ID names, or Steps when ID is empty, so a
// recording can be tried before it is saved.
type PlayMacroReq struct {
	ID    string         `json:"id" form:"id"`
	Name  string         `json:"name" form:"name"`
	Steps []SequenceStep `json:"steps" form:"steps"`
}
