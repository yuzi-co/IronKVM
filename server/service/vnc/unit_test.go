package vnc

import (
	"bytes"
	"crypto/des"
	"encoding/binary"
	"image"
	"image/jpeg"
	"os"
	"testing"
)

func TestParseClientVersion(t *testing.T) {
	cases := []struct {
		in    string
		minor int
		ok    bool
	}{
		{"RFB 003.008\n", 8, true},
		{"RFB 003.007\n", 7, true},
		// macOS Screen Sharing answers 3.889, which is 3.8 with extras.
		{"RFB 003.889\n", 8, true},
		{"RFB 003.003\n", 0, false},
		{"RFB 004.000\n", 0, false},
		{"RFB 003.008 ", 0, false},
		{"HTTP/1.1 200", 0, false},
	}
	for _, c := range cases {
		minor, err := parseClientVersion([]byte(c.in))
		if c.ok != (err == nil) || minor != c.minor {
			t.Errorf("parseClientVersion(%q) = %d, %v", c.in, minor, err)
		}
	}
}

func TestSecurityResultCarriesTheReasonOnlyFrom38(t *testing.T) {
	var buf bytes.Buffer
	if err := writeSecurityResult(&buf, 8, ""); err != nil || !bytes.Equal(buf.Bytes(), []byte{0, 0, 0, 0}) {
		t.Fatalf("success = %v, %v", buf.Bytes(), err)
	}

	buf.Reset()
	_ = writeSecurityResult(&buf, 8, "no")
	if want := []byte{0, 0, 0, 1, 0, 0, 0, 2, 'n', 'o'}; !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("3.8 failure = %v, want %v", buf.Bytes(), want)
	}

	buf.Reset()
	_ = writeSecurityResult(&buf, 7, "no")
	if want := []byte{0, 0, 0, 1}; !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("3.7 failure = %v, want %v", buf.Bytes(), want)
	}
}

func TestServerInitAnnouncesTrueColour(t *testing.T) {
	msg := serverInit(1920, 1080, "kvm")
	if len(msg) != 24+3 {
		t.Fatalf("ServerInit is %d bytes", len(msg))
	}
	if binary.BigEndian.Uint16(msg[0:]) != 1920 || binary.BigEndian.Uint16(msg[2:]) != 1080 {
		t.Fatalf("ServerInit size = %v", msg[:4])
	}
	format := parsePixelFormat(msg[4:20])
	if format != serverPixelFormat || !format.jpegCapable() {
		t.Fatalf("ServerInit pixel format = %+v", format)
	}
	if binary.BigEndian.Uint32(msg[20:]) != 3 || string(msg[24:]) != "kvm" {
		t.Fatalf("ServerInit name = %v", msg[20:])
	}
}

func TestJPEGNeedsTrueColourAt16Or32Bits(t *testing.T) {
	for _, c := range []struct {
		bpp, trueColour uint8
		ok              bool
	}{{32, 1, true}, {16, 1, true}, {8, 1, false}, {32, 0, false}} {
		format := pixelFormat{BitsPerPixel: c.bpp, TrueColour: c.trueColour}
		if format.jpegCapable() != c.ok {
			t.Errorf("%d bpp, true colour %d: jpegCapable = %t", c.bpp, c.trueColour, !c.ok)
		}
	}
}

func TestCompactLength(t *testing.T) {
	cases := []struct {
		n    int
		want []byte
	}{
		{0, []byte{0x00}},
		{127, []byte{0x7f}},
		{128, []byte{0x80, 0x01}},
		{10000, []byte{0x90, 0x4e}},
		{16383, []byte{0xff, 0x7f}},
		{16384, []byte{0x80, 0x80, 0x01}},
		{maxTightLength, []byte{0xff, 0xff, 0xff}},
	}
	for _, c := range cases {
		if got := appendCompactLength(nil, c.n); !bytes.Equal(got, c.want) {
			t.Errorf("compact length %d = %x, want %x", c.n, got, c.want)
		}
		if got := decodeCompactLength(appendCompactLength(nil, c.n)); got != c.n {
			t.Errorf("compact length %d reads back as %d", c.n, got)
		}
	}
}

// decodeCompactLength reads a compact length the way a client does.
func decodeCompactLength(b []byte) int {
	n := int(b[0] & 0x7f)
	if b[0]&0x80 != 0 {
		n |= int(b[1]&0x7f) << 7
		if b[1]&0x80 != 0 {
			n |= int(b[2]) << 14
		}
	}
	return n
}

func TestTightJPEGRectFraming(t *testing.T) {
	frame := testJPEG(t, 64, 48)
	var buf bytes.Buffer
	n, err := writeJPEGUpdate(&buf, 64, 48, frame)
	if err != nil {
		t.Fatal(err)
	}
	if int(n) != buf.Len() {
		t.Fatalf("writeJPEGUpdate reported %d bytes and wrote %d", n, buf.Len())
	}

	msg := buf.Bytes()
	// FramebufferUpdate, padding, one rectangle.
	if !bytes.Equal(msg[:4], []byte{0, 0, 0, 1}) {
		t.Fatalf("update header = %v", msg[:4])
	}
	x, y := binary.BigEndian.Uint16(msg[4:]), binary.BigEndian.Uint16(msg[6:])
	w, h := binary.BigEndian.Uint16(msg[8:]), binary.BigEndian.Uint16(msg[10:])
	encoding := int32(binary.BigEndian.Uint32(msg[12:]))
	if x != 0 || y != 0 || w != 64 || h != 48 || encoding != encodingTight {
		t.Fatalf("rectangle = %d,%d %dx%d encoding %d", x, y, w, h, encoding)
	}
	if msg[16] != 0x90 {
		t.Fatalf("compression control = %#x, want 0x90 (JPEG)", msg[16])
	}
	length := decodeCompactLength(msg[17:])
	lengthBytes := len(appendCompactLength(nil, length))
	data := msg[17+lengthBytes:]
	if length != len(frame) || !bytes.Equal(data, frame) {
		t.Fatalf("JPEG payload is %d bytes, want the %d-byte frame unchanged", len(data), len(frame))
	}
}

func TestAFrameTooLargeForTightIsRefused(t *testing.T) {
	if _, err := jpegUpdateHeader(10, 10, maxTightLength+1); err == nil {
		t.Fatal("a frame over 4MB was framed")
	}
}

func TestJPEGSizeReadsTheFrameHeader(t *testing.T) {
	w, h, err := jpegSize(testJPEG(t, 1280, 720))
	if err != nil || w != 1280 || h != 720 {
		t.Fatalf("jpegSize = %dx%d, %v", w, h, err)
	}

	// An APP segment and fill bytes ahead of the frame header.
	frame := testJPEG(t, 8, 8)
	withApp := append([]byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x04, 'x', 'y', 0xff}, frame[2:]...)
	if w, h, err := jpegSize(withApp); err != nil || w != 8 || h != 8 {
		t.Fatalf("jpegSize with an APP segment = %dx%d, %v", w, h, err)
	}

	for _, bad := range [][]byte{nil, []byte("not a jpeg"), {0xff, 0xd8, 0xff, 0xd9}, frame[:6]} {
		if _, _, err := jpegSize(bad); err == nil {
			t.Errorf("jpegSize(%x) found a size", bad)
		}
	}
}

func testJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for i := range img.Pix {
		img.Pix[i] = byte(i)
	}
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestKeysymMapping(t *testing.T) {
	cases := []struct {
		name string
		sym  uint32
		want keyAction
	}{
		{"a", 'a', keyAction{code: 0x04}},
		{"A needs Shift", 'A', keyAction{code: 0x04, implied: modLeftShift}},
		{"1", '1', keyAction{code: 0x1e}},
		{"! needs Shift", '!', keyAction{code: 0x1e, implied: modLeftShift}},
		{"space", ' ', keyAction{code: 0x2c}},
		{"Return", 0xff0d, keyAction{code: 0x28}},
		{"BackSpace", 0xff08, keyAction{code: 0x2a}},
		{"Escape", 0xff1b, keyAction{code: 0x29}},
		{"Delete", 0xffff, keyAction{code: 0x4c}},
		{"Left", 0xff51, keyAction{code: 0x50}},
		{"F1", 0xffbe, keyAction{code: 0x3a}},
		{"F12", 0xffc9, keyAction{code: 0x45}},
		{"F13", 0xffca, keyAction{code: 0x68}},
		{"KP_5", 0xffb5, keyAction{code: 0x5d}},
		{"KP_Enter", 0xff8d, keyAction{code: 0x58}},
		{"Shift_L", 0xffe1, keyAction{modifier: modLeftShift}},
		{"Control_R", 0xffe4, keyAction{modifier: modRightCtrl}},
		{"Super_L", 0xffeb, keyAction{modifier: modLeftGUI}},
		{"AltGr", 0xfe03, keyAction{modifier: modRightAlt}},
		{"Unicode keysym for a", 0x01000061, keyAction{}},
	}
	for _, c := range cases {
		got, ok := mapKeysym(c.sym)
		if c.want == (keyAction{}) {
			// 0x01000061 is below the Unicode keysym range, which
			// starts at 0x01000100.
			if ok {
				t.Errorf("%s: mapped to %+v", c.name, got)
			}
			continue
		}
		if !ok || got != c.want {
			t.Errorf("%s: mapKeysym(%#x) = %+v, %t, want %+v", c.name, c.sym, got, ok, c.want)
		}
	}

	// A character the US layout types only with a dead key is not typed.
	if _, ok := mapKeysym(0xe9); ok {
		t.Error("é was mapped on the US layout")
	}
	// A keysym that names nothing.
	if _, ok := mapKeysym(0x12345678); ok {
		t.Error("an unknown keysym was mapped")
	}
}

func TestKeyboardReleasesTheKeyItsPressChose(t *testing.T) {
	k := newKeyboard()

	// Shift, then "!", then Shift goes up first, and the client releases
	// "1" because that is what the key types now.
	steps := []struct {
		down bool
		sym  uint32
		want []byte
	}{
		{true, 0xffe1, []byte{modLeftShift, 0, 0, 0, 0, 0, 0, 0}},
		{true, '!', []byte{modLeftShift, 0, 0x1e, 0, 0, 0, 0, 0}},
		// "!" still implies Shift while it is held.
		{false, 0xffe1, nil},
		{false, '1', nil},
		{false, '!', []byte{0, 0, 0, 0, 0, 0, 0, 0}},
	}
	for i, step := range steps {
		got := k.event(step.down, step.sym)
		if !bytes.Equal(got, step.want) {
			t.Fatalf("step %d: report %v, want %v", i, got, step.want)
		}
	}
	if k.held() {
		t.Fatal("a key is still held")
	}
}

func TestKeyboardSendsSixKeysAndIgnoresRepeats(t *testing.T) {
	k := newKeyboard()
	for _, sym := range "abcdefg" {
		k.event(true, uint32(sym))
	}
	report := k.report()
	if !bytes.Equal(report[2:], []byte{0x04, 0x05, 0x06, 0x07, 0x08, 0x09}) {
		t.Fatalf("report = %v, want the first six keys", report)
	}
	// Auto-repeat sends the press again, which changes nothing.
	if got := k.event(true, 'a'); got != nil {
		t.Fatalf("a repeated press sent %v", got)
	}
	// With "a" up, "g" takes the free place.
	if got := k.event(false, 'a'); !bytes.Equal(got[2:], []byte{0x05, 0x06, 0x07, 0x08, 0x09, 0x0a}) {
		t.Fatalf("report after a release = %v", got)
	}
	// A keysym nobody pressed, and one no key types, change nothing.
	if got := k.event(false, 'z'); got != nil {
		t.Fatalf("releasing an unpressed key sent %v", got)
	}
	if got := k.event(true, 0xe9); got != nil {
		t.Fatalf("an untypeable keysym sent %v", got)
	}
}

func TestPointerScaling(t *testing.T) {
	cases := []struct {
		position, size int
		want           uint16
	}{
		{0, 1920, 1},
		{1919, 1920, 0x8000},
		{5000, 1920, 0x8000},
		{959, 1919, 0x7fff/2 + 1},
		{540, 1081, uint16(0x7fff*540/1080) + 1},
		{0, 0, 1},
	}
	for _, c := range cases {
		if got := absoluteCoordinate(c.position, c.size); got != c.want {
			t.Errorf("absoluteCoordinate(%d, %d) = %d, want %d", c.position, c.size, got, c.want)
		}
	}
}

func TestPointerButtonsAndWheel(t *testing.T) {
	var p pointer

	reports := p.event(vncButtonLeft|vncButtonRight, 0, 0, 100, 100)
	if len(reports) != 1 || reports[0][0] != hidButtonLeft|hidButtonRight {
		t.Fatalf("left and right = %v", reports)
	}
	reports = p.event(vncButtonMiddle, 99, 99, 100, 100)
	if reports[0][0] != hidButtonMiddle || binary.LittleEndian.Uint16(reports[0][1:]) != 0x8000 ||
		binary.LittleEndian.Uint16(reports[0][3:]) != 0x8000 {
		t.Fatalf("middle at the far corner = %v", reports)
	}

	// The wheel steps once on the press, and not while the button stays down.
	reports = p.event(vncWheelUp, 10, 10, 100, 100)
	if len(reports) != 2 || int8(reports[1][5]) != 1 || reports[0][5] != 0 {
		t.Fatalf("wheel up = %v", reports)
	}
	if reports = p.event(vncWheelUp, 10, 10, 100, 100); len(reports) != 1 {
		t.Fatalf("a held wheel button stepped again: %v", reports)
	}
	reports = p.event(vncWheelDown, 10, 10, 100, 100)
	if len(reports) != 2 || int8(reports[1][5]) != -1 {
		t.Fatalf("wheel down = %v", reports)
	}
}

// The DES answer matches one computed the way a client computes it, with the
// key bits reversed by hand.
func TestVNCAuthResponse(t *testing.T) {
	challenge := []byte("0123456789abcdef")
	got, err := vncAuthResponse("secret12", challenge)
	if err != nil {
		t.Fatal(err)
	}
	if want := clientVNCAuthResponse(t, "secret12", challenge); !bytes.Equal(got, want) {
		t.Fatalf("response = %x, want %x", got, want)
	}

	// DES uses eight bytes, so a ninth changes nothing.
	longer, _ := vncAuthResponse("secret123", challenge)
	if !bytes.Equal(got, longer) {
		t.Fatal("a ninth password byte changed the response")
	}
	other, _ := vncAuthResponse("secret13", challenge)
	if bytes.Equal(got, other) {
		t.Fatal("two passwords gave the same response")
	}
}

// clientVNCAuthResponse is the client's side of VNC authentication, written
// apart from the server's.
func clientVNCAuthResponse(t *testing.T, password string, challenge []byte) []byte {
	t.Helper()
	key := make([]byte, 8)
	copy(key, password)
	for i := range key {
		b := key[i]
		b = (b&0xf0)>>4 | (b&0x0f)<<4
		b = (b&0xcc)>>2 | (b&0x33)<<2
		b = (b&0xaa)>>1 | (b&0x55)<<1
		key[i] = b
	}
	block, err := des.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	out := make([]byte, 16)
	block.Encrypt(out[:8], challenge[:8])
	block.Encrypt(out[8:], challenge[8:])
	return out
}

func TestPasswordValidation(t *testing.T) {
	for _, c := range []struct {
		password string
		ok       bool
	}{{"abc123", true}, {"abcd1234", true}, {"abc12", false}, {"abcd12345", false}, {"abc\x00def", false}} {
		if err := validatePassword(c.password); (err == nil) != c.ok {
			t.Errorf("validatePassword(%q) = %v", c.password, err)
		}
	}
}

func TestFilePasswordIsPrivate(t *testing.T) {
	store := FilePassword{Path: t.TempDir() + "/vnc.passwd"}
	if _, ok, err := store.Get(); ok || err != nil {
		t.Fatalf("a missing file gave a password: %t, %v", ok, err)
	}
	if err := store.Set("abc"); err == nil {
		t.Fatal("a short password was saved")
	}
	if err := store.Set("abc12345"); err != nil {
		t.Fatal(err)
	}
	password, ok, err := store.Get()
	if !ok || err != nil || password != "abc12345" {
		t.Fatalf("Get = %q, %t, %v", password, ok, err)
	}
	checkPrivate(t, store.Path)
}

func checkPrivate(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("%s has mode %o, want 600", path, perm)
	}
}
