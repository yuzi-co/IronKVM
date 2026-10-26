package vnc

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestScancodeMapping(t *testing.T) {
	cases := []struct {
		name string
		code uint32
		want keyAction
	}{
		{"A", 0x1e, keyAction{code: 0x04}},
		{"Q, which is A on AZERTY", 0x10, keyAction{code: 0x14}},
		{"Z, which is Y on QWERTZ", 0x2c, keyAction{code: 0x1d}},
		{"1", 0x02, keyAction{code: 0x1e}},
		{"0", 0x0b, keyAction{code: 0x27}},
		{"grave, a dead key on QWERTZ", 0x29, keyAction{code: 0x35}},
		{"equal, a dead key on QWERTZ", 0x0d, keyAction{code: 0x2e}},
		{"backslash", 0x2b, keyAction{code: 0x31}},
		{"the ISO key left of Z", 0x56, keyAction{code: 0x64}},
		{"Escape", 0x01, keyAction{code: 0x29}},
		{"Return", 0x1c, keyAction{code: 0x28}},
		{"space", 0x39, keyAction{code: 0x2c}},
		{"F1", 0x3b, keyAction{code: 0x3a}},
		{"F10", 0x44, keyAction{code: 0x43}},
		{"F11", 0x57, keyAction{code: 0x44}},
		{"F12", 0x58, keyAction{code: 0x45}},
		{"KP_7", 0x47, keyAction{code: 0x5f}},
		{"KP_Multiply", 0x37, keyAction{code: 0x55}},
		{"Shift_L", 0x2a, keyAction{modifier: modLeftShift}},
		{"Shift_R", 0x36, keyAction{modifier: modRightShift}},
		{"Control_L", 0x1d, keyAction{modifier: modLeftCtrl}},
		{"Alt_L", 0x38, keyAction{modifier: modLeftAlt}},

		// E0-prefixed keys arrive as 0x80 | the second byte.
		{"AltGr, E0 38", 0xb8, keyAction{modifier: modRightAlt}},
		{"Control_R, E0 1D", 0x9d, keyAction{modifier: modRightCtrl}},
		{"Super_L, E0 5B", 0xdb, keyAction{modifier: modLeftGUI}},
		{"Super_R, E0 5C", 0xdc, keyAction{modifier: modRightGUI}},
		{"Menu, E0 5D", 0xdd, keyAction{code: 0x65}},
		{"KP_Enter, E0 1C", 0x9c, keyAction{code: 0x58}},
		{"KP_Divide, E0 35", 0xb5, keyAction{code: 0x54}},
		{"Right, E0 4D", 0xcd, keyAction{code: 0x4f}},
		{"Left, E0 4B", 0xcb, keyAction{code: 0x50}},
		{"Up, E0 48", 0xc8, keyAction{code: 0x52}},
		{"Down, E0 50", 0xd0, keyAction{code: 0x51}},
		{"Home, E0 47", 0xc7, keyAction{code: 0x4a}},
		{"End, E0 4F", 0xcf, keyAction{code: 0x4d}},
		{"Prior, E0 49", 0xc9, keyAction{code: 0x4b}},
		{"Next, E0 51", 0xd1, keyAction{code: 0x4e}},
		{"Insert, E0 52", 0xd2, keyAction{code: 0x49}},
		{"Delete, E0 53", 0xd3, keyAction{code: 0x4c}},

		// The spec's fixed codes for the two multi-byte keys.
		{"Print", 0x54, keyAction{code: 0x46}},
		{"Print, the old E0 37 form", 0xb7, keyAction{code: 0x46}},
		{"Pause", 0xc6, keyAction{code: 0x48}},

		// Keys no keysym reaches on the US layout today.
		{"Ro, Japanese", 0x73, keyAction{code: 0x87}},
		{"Yen, Japanese", 0x7d, keyAction{code: 0x89}},
		{"Henkan, Japanese", 0x79, keyAction{code: 0x8a}},
		{"Muhenkan, Japanese", 0x7b, keyAction{code: 0x8b}},
		{"Katakana/Hiragana", 0x70, keyAction{code: 0x88}},
		{"KP_Equal", 0x59, keyAction{code: 0x67}},
		{"KP_Comma", 0x7e, keyAction{code: 0x85}},
		{"F13", 0x5d, keyAction{code: 0x68}},
		{"F24", 0x6f, keyAction{code: 0x73}},
	}
	for _, c := range cases {
		got, ok := mapScancode(c.code)
		if !ok || got != c.want {
			t.Errorf("%s: mapScancode(%#x) = %+v, %t, want %+v", c.name, c.code, got, ok, c.want)
		}
	}

	for _, code := range []uint32{
		0x00,   // no key
		0x80,   // E0 00, no key
		0xe038, // E0 38 with the prefix byte kept, which the spec does not allow
		0x1e | 0x100,
		0xffffffff,
	} {
		if got, ok := mapScancode(code); ok {
			t.Errorf("mapScancode(%#x) = %+v, want no key", code, got)
		}
	}
}

// A German host with a German client: AltGr+Q types "@". The client sends the
// keysym "@", which the US layout would type as Shift+2. The scancode path must
// send the keys the user pressed and nothing the keysym implies.
func TestKeyboardScancodeIgnoresTheKeysym(t *testing.T) {
	k := newKeyboard()
	steps := []struct {
		down bool
		code uint32
		sym  uint32
		want []byte
	}{
		{true, 0xb8, 0xfe03, []byte{modRightAlt, 0, 0, 0, 0, 0, 0, 0}},
		{true, 0x10, '@', []byte{modRightAlt, 0, 0x14, 0, 0, 0, 0, 0}},
		// The client may name the key differently on release.
		{false, 0x10, 'q', []byte{modRightAlt, 0, 0, 0, 0, 0, 0, 0}},
		{false, 0xb8, 0xfe03, []byte{0, 0, 0, 0, 0, 0, 0, 0}},
	}
	for i, step := range steps {
		got := k.scancodeEvent(step.down, step.code, step.sym)
		if !bytes.Equal(got, step.want) {
			t.Fatalf("step %d: report %v, want %v", i, got, step.want)
		}
	}
	if k.held() {
		t.Fatal("a key is still held")
	}
}

// A dead key has a keysym the US table cannot type, so the keysym path drops
// it. The scancode path presses the key.
func TestKeyboardScancodeTypesADeadKey(t *testing.T) {
	k := newKeyboard()
	if got := k.event(true, 0xfe52); got != nil { // dead_circumflex
		t.Fatalf("the keysym path sent %v for a dead key", got)
	}
	k.event(false, 0xfe52)

	if got := k.scancodeEvent(true, 0x29, 0xfe52); !bytes.Equal(got, []byte{0, 0, 0x35, 0, 0, 0, 0, 0}) {
		t.Fatalf("report %v, want the grave key", got)
	}
	if got := k.scancodeEvent(false, 0x29, 0xfe52); !bytes.Equal(got, make([]byte, 8)) {
		t.Fatalf("release report %v", got)
	}
}

// A keycode the table does not know falls back to the keysym in the same
// message, so the key is not lost.
func TestKeyboardScancodeFallsBackToTheKeysym(t *testing.T) {
	k := newKeyboard()
	if got := k.scancodeEvent(true, 0x7f, 'A'); !bytes.Equal(got, []byte{modLeftShift, 0, 0x04, 0, 0, 0, 0, 0}) {
		t.Fatalf("report %v, want Shift+A from the keysym", got)
	}
	if got := k.scancodeEvent(false, 0x7f, 'A'); !bytes.Equal(got, make([]byte, 8)) {
		t.Fatalf("release report %v", got)
	}
	// Neither a keycode nor a keysym: nothing.
	if got := k.scancodeEvent(true, 0x7f, 0); got != nil {
		t.Fatalf("an unknown key sent %v", got)
	}
}

// Keys pressed through either message are held together, and each release
// finds its own key.
func TestKeyboardMixesScancodeAndKeysymEvents(t *testing.T) {
	k := newKeyboard()
	k.scancodeEvent(true, 0x1d, 0xffe3) // Control_L
	if got := k.event(true, 'c'); !bytes.Equal(got, []byte{modLeftCtrl, 0, 0x06, 0, 0, 0, 0, 0}) {
		t.Fatalf("report %v", got)
	}
	// The keysym of Control_L does not release the key the scancode held.
	if got := k.event(false, 0xffe3); got != nil {
		t.Fatalf("a keysym release let go of a scancode key: %v", got)
	}
	k.event(false, 'c')
	if got := k.scancodeEvent(false, 0x1d, 0xffe3); !bytes.Equal(got, make([]byte, 8)) {
		t.Fatalf("release report %v", got)
	}
	// Auto-repeat of a scancode key changes nothing.
	k.scancodeEvent(true, 0x1e, 'a')
	if got := k.scancodeEvent(true, 0x1e, 'a'); got != nil {
		t.Fatalf("a repeated press sent %v", got)
	}
}

func TestJPEGUpdateCarriesTheExtendedKeyEventAck(t *testing.T) {
	header, err := jpegUpdateHeader(64, 48, 100, true)
	if err != nil {
		t.Fatal(err)
	}
	if header[0] != msgFramebufferUpdate || binary.BigEndian.Uint16(header[2:]) != 2 {
		t.Fatalf("update header %v, want two rectangles", header[:4])
	}
	// The pseudo-rectangle comes first and is empty.
	ack := header[4:16]
	if !bytes.Equal(ack[:8], make([]byte, 8)) {
		t.Fatalf("the pseudo-rectangle has a position or size: %v", ack[:8])
	}
	if enc := int32(binary.BigEndian.Uint32(ack[8:])); enc != encodingQEMUExtendedKeyEvent {
		t.Fatalf("pseudo-rectangle encoding %d", enc)
	}
	if enc := int32(binary.BigEndian.Uint32(header[24:28])); enc != encodingTight {
		t.Fatalf("second rectangle encoding %d, want Tight", enc)
	}

	plain, err := jpegUpdateHeader(64, 48, 100, false)
	if err != nil {
		t.Fatal(err)
	}
	if binary.BigEndian.Uint16(plain[2:]) != 1 || !bytes.Equal(plain[4:], header[16:]) {
		t.Fatal("without the ack the update is not the plain one-rectangle update")
	}
}
