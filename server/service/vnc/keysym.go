package vnc

import (
	"NanoKVM-Server/service/hid"
)

// Modifier bits of the keyboard report's first byte.
const (
	modLeftCtrl   byte = 0x01
	modLeftShift  byte = 0x02
	modLeftAlt    byte = 0x04
	modLeftGUI    byte = 0x08
	modRightCtrl  byte = 0x10
	modRightShift byte = 0x20
	modRightAlt   byte = 0x40
	modRightGUI   byte = 0x80
)

// modifierKeysyms maps the X11 modifier keysyms to their bit in the report.
// Meta is the Command key of a Mac client, so it is the GUI key here, as Super
// is. ISO_Level3_Shift is what an X11 client sends for AltGr.
var modifierKeysyms = map[uint32]byte{
	0xffe1: modLeftShift,  // Shift_L
	0xffe2: modRightShift, // Shift_R
	0xffe3: modLeftCtrl,   // Control_L
	0xffe4: modRightCtrl,  // Control_R
	0xffe7: modLeftGUI,    // Meta_L
	0xffe8: modRightGUI,   // Meta_R
	0xffe9: modLeftAlt,    // Alt_L
	0xffea: modRightAlt,   // Alt_R
	0xffeb: modLeftGUI,    // Super_L
	0xffec: modRightGUI,   // Super_R
	0xfe03: modRightAlt,   // ISO_Level3_Shift
}

// functionKeysyms maps the keysyms that name a key rather than a character to
// the key's HID usage. The keypad keysyms with Num Lock off (KP_Home and the
// rest) map to the keypad keys, so the host sees what the client's keypad did.
var functionKeysyms = map[uint32]byte{
	0xff08: 0x2a, // BackSpace
	0xff09: 0x2b, // Tab
	0xfe20: 0x2b, // ISO_Left_Tab, which is Shift+Tab
	0xff0d: 0x28, // Return
	0xff13: 0x48, // Pause
	0xff14: 0x47, // Scroll_Lock
	0xff15: 0x9a, // Sys_Req
	0xff1b: 0x29, // Escape
	0xffff: 0x4c, // Delete

	0xff50: 0x4a, // Home
	0xff51: 0x50, // Left
	0xff52: 0x52, // Up
	0xff53: 0x4f, // Right
	0xff54: 0x51, // Down
	0xff55: 0x4b, // Prior
	0xff56: 0x4e, // Next
	0xff57: 0x4d, // End

	0xff61: 0x46, // Print
	0xff63: 0x49, // Insert
	0xff67: 0x65, // Menu
	0xff6b: 0x48, // Break
	0xff7f: 0x53, // Num_Lock
	0xffe5: 0x39, // Caps_Lock

	0xff8d: 0x58, // KP_Enter
	0xff95: 0x5f, // KP_Home
	0xff96: 0x5c, // KP_Left
	0xff97: 0x60, // KP_Up
	0xff98: 0x5e, // KP_Right
	0xff99: 0x5a, // KP_Down
	0xff9a: 0x61, // KP_Prior
	0xff9b: 0x5b, // KP_Next
	0xff9c: 0x59, // KP_End
	0xff9d: 0x5d, // KP_Begin
	0xff9e: 0x62, // KP_Insert
	0xff9f: 0x63, // KP_Delete
	0xffaa: 0x55, // KP_Multiply
	0xffab: 0x57, // KP_Add
	0xffac: 0x85, // KP_Separator
	0xffad: 0x56, // KP_Subtract
	0xffae: 0x63, // KP_Decimal
	0xffaf: 0x54, // KP_Divide
	0xffb0: 0x62, // KP_0
	0xffb1: 0x59, // KP_1
	0xffb2: 0x5a, // KP_2
	0xffb3: 0x5b, // KP_3
	0xffb4: 0x5c, // KP_4
	0xffb5: 0x5d, // KP_5
	0xffb6: 0x5e, // KP_6
	0xffb7: 0x5f, // KP_7
	0xffb8: 0x60, // KP_8
	0xffb9: 0x61, // KP_9
	0xffbd: 0x67, // KP_Equal
}

func init() {
	// F1 to F12 are 0xffbe to 0xffc9 and usages 0x3a to 0x45. F13 to F24
	// follow at 0xffca and 0x68.
	for i := uint32(0); i < 12; i++ {
		functionKeysyms[0xffbe+i] = byte(0x3a + i)
		functionKeysyms[0xffca+i] = byte(0x68 + i)
	}
}

// usLayout is the layout printable keysyms are typed on. The US table in
// service/hid is the one paste uses.
var usLayout = func() *hid.Layout {
	layout, err := hid.GetLayout("us")
	if err != nil {
		panic(err)
	}
	return layout
}()

// keysymRune returns the character a printable keysym stands for. Latin-1
// keysyms are their own code point, and 0x01000000 plus a code point is the
// Unicode form.
func keysymRune(sym uint32) (rune, bool) {
	switch {
	case sym >= 0x20 && sym <= 0x7e, sym >= 0xa0 && sym <= 0xff:
		return rune(sym), true
	case sym >= 0x01000100 && sym <= 0x0110ffff:
		return rune(sym - 0x01000000), true
	}
	return 0, false
}

// keyAction is what one keysym presses: a modifier, or a key with the
// modifiers the US layout needs to type the character on it.
type keyAction struct {
	modifier byte
	code     byte
	implied  byte
}

// mapKeysym turns a keysym into a key press on the US layout. It answers false
// for a keysym no single key types, such as a character that needs a dead key.
func mapKeysym(sym uint32) (keyAction, bool) {
	if bit, ok := modifierKeysyms[sym]; ok {
		return keyAction{modifier: bit}, true
	}
	if code, ok := functionKeysyms[sym]; ok {
		return keyAction{code: code}, true
	}
	r, ok := keysymRune(sym)
	if !ok {
		return keyAction{}, false
	}
	strokes, ok := usLayout.Keystrokes(r)
	if !ok || len(strokes) != 1 {
		return keyAction{}, false
	}
	return keyAction{code: byte(strokes[0].Code), implied: byte(strokes[0].Modifiers)}, true
}

// heldKey is a key down on the host, with the keysym that pressed it.
type heldKey struct {
	sym     uint32
	code    byte
	implied byte
}

// keyboard turns VNC key events into boot keyboard reports. It remembers what
// each keysym pressed, so a release lets go of the same key even if the client
// names it differently by then: a client that pressed "!" and let go of Shift
// first releases "1".
type keyboard struct {
	modifiers map[uint32]byte
	keys      []heldKey
}

func newKeyboard() *keyboard {
	return &keyboard{modifiers: make(map[uint32]byte)}
}

// event applies one KeyEvent and returns the report to send, or nil when the
// report would not change.
func (k *keyboard) event(down bool, sym uint32) []byte {
	before := k.report()

	if down {
		k.press(sym)
	} else {
		k.release(sym)
	}

	after := k.report()
	if string(before) == string(after) {
		return nil
	}
	return after
}

func (k *keyboard) press(sym uint32) {
	if _, ok := k.modifiers[sym]; ok {
		return
	}
	for _, key := range k.keys {
		if key.sym == sym {
			return
		}
	}

	action, ok := mapKeysym(sym)
	if !ok {
		return
	}
	if action.modifier != 0 {
		k.modifiers[sym] = action.modifier
		return
	}
	k.keys = append(k.keys, heldKey{sym: sym, code: action.code, implied: action.implied})
}

func (k *keyboard) release(sym uint32) {
	if _, ok := k.modifiers[sym]; ok {
		delete(k.modifiers, sym)
		return
	}
	for i, key := range k.keys {
		if key.sym == sym {
			k.keys = append(k.keys[:i], k.keys[i+1:]...)
			return
		}
	}
}

// report builds the 8-byte report: the modifiers held, those the held
// characters imply, and up to six keys in the order they went down.
func (k *keyboard) report() []byte {
	report := make([]byte, hid.KeyboardReportLen)
	for _, bit := range k.modifiers {
		report[0] |= bit
	}

	slot := 2
	for _, key := range k.keys {
		report[0] |= key.implied
		if slot < len(report) && !containsCode(report[2:slot], key.code) {
			report[slot] = key.code
			slot++
		}
	}
	return report
}

// held reports whether any key or modifier is down.
func (k *keyboard) held() bool {
	return len(k.modifiers) > 0 || len(k.keys) > 0
}

func containsCode(codes []byte, code byte) bool {
	for _, c := range codes {
		if c == code {
			return true
		}
	}
	return false
}
