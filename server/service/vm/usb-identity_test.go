package vm

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"NanoKVM-Server/proto"
)

func TestParseUSBID(t *testing.T) {
	good := map[string]string{
		"0x046d": "0x046d",
		"046D":   "0x046d",
		"0XC31C": "0xc31c",
		"1":      "0x0001",
		" ffff ": "0xffff",
	}
	for in, want := range good {
		got, err := parseUSBID(in)
		if err != nil || got != want {
			t.Errorf("parseUSBID(%q) = %q, %v; want %q", in, got, err, want)
		}
	}

	for _, in := range []string{"", "0x", "0x0000", "0", "10000", "0x1g00", "-1", "0x 12", "12345"} {
		if got, err := parseUSBID(in); err == nil {
			t.Errorf("parseUSBID(%q) = %q, want an error", in, got)
		}
	}
}

func TestUSBStringsAndSerial(t *testing.T) {
	for _, bad := range []string{"", " lead", "trail ", "new\nline", "tab\there", "café", strings.Repeat("a", usbStringMax+1)} {
		if validUSBString("x", bad) == nil {
			t.Errorf("validUSBString(%q) accepted", bad)
		}
	}
	for _, ok := range []string{"Logitech", "Dell KB216 Wired Keyboard", "a=b", strings.Repeat("a", usbStringMax)} {
		if err := validUSBString("x", ok); err != nil {
			t.Errorf("validUSBString(%q): %v", ok, err)
		}
	}

	for _, bad := range []string{"short", "0123456789A", "0123456789AB-", "0123456789AB CD", strings.Repeat("A", usbStringMax+1)} {
		if validUSBSerial(bad) == nil {
			t.Errorf("validUSBSerial(%q) accepted", bad)
		}
	}
	for _, ok := range []string{"", "0123456789AB", "abcDEF012345xyz"} {
		if err := validUSBSerial(ok); err != nil {
			t.Errorf("validUSBSerial(%q): %v", ok, err)
		}
	}
}

func TestResolveUSBIdentityPresetIgnoresCustomFields(t *testing.T) {
	preset, identity, err := resolveUSBIdentity(proto.SetUSBIdentityReq{
		Preset: "logitech", VendorID: "0x1234", Manufacturer: "ignored", Serial: "0123456789AB",
	})
	if err != nil {
		t.Fatal(err)
	}
	want, _ := usbIdentityPreset("logitech")
	want.Serial = "0123456789AB"
	if preset != "logitech" || identity != want {
		t.Fatalf("got %s %+v, want logitech %+v", preset, identity, want)
	}

	if _, _, err := resolveUSBIdentity(proto.SetUSBIdentityReq{Preset: "custom", VendorID: "0x1234", ProductID: "0x5678", Manufacturer: "Acme"}); err == nil {
		t.Fatal("custom without a product name was accepted")
	}
	if _, _, err := resolveUSBIdentity(proto.SetUSBIdentityReq{Preset: "nope"}); err == nil {
		t.Fatal("an unknown preset was accepted")
	}
}

// fakeIdentityBoard is the file and the rebuilds, in memory.
type fakeIdentityBoard struct {
	file     []byte
	rebuilds [][]string
	failSave bool
}

func (b *fakeIdentityBoard) env() usbIdentityEnv {
	return usbIdentityEnv{
		read: func(path string) ([]byte, error) {
			if path != usbIdentityFile || b.file == nil {
				return nil, os.ErrNotExist
			}
			return b.file, nil
		},
		write: func(data []byte) error {
			if b.failSave {
				return errors.New("read-only")
			}
			b.file = data
			return nil
		},
		remove: func() error {
			b.file = nil
			return nil
		},
		rebuild: func(commands []string) error {
			b.rebuilds = append(b.rebuilds, commands)
			return nil
		},
	}
}

func TestApplyUSBIdentityRoundTrip(t *testing.T) {
	board := &fakeIdentityBoard{}

	// The default on a board with nothing saved changes nothing, and costs
	// the host no re-enumeration.
	if f := applyUSBIdentity(proto.SetUSBIdentityReq{Preset: "default"}, board.env()); f != nil || len(board.rebuilds) != 0 {
		t.Fatalf("default on a default board: %+v, %d rebuilds", f, len(board.rebuilds))
	}

	custom := proto.SetUSBIdentityReq{
		Preset: "custom", VendorID: "1234", ProductID: "0xABCD",
		Manufacturer: "Acme", Product: "KVM Keyboard", Serial: "SERIAL000001",
	}
	if f := applyUSBIdentity(custom, board.env()); f != nil {
		t.Fatalf("custom: %+v", f)
	}
	wantFile := "preset=custom\nvid=0x1234\npid=0xabcd\nmanufacturer=Acme\nproduct=KVM Keyboard\nserial=SERIAL000001\n"
	if string(board.file) != wantFile {
		t.Fatalf("file =\n%s\nwant\n%s", board.file, wantFile)
	}
	if want := [][]string{{usbGadgetStop, usbGadgetStart}}; !reflect.DeepEqual(board.rebuilds, want) {
		t.Fatalf("rebuilds = %v, want %v", board.rebuilds, want)
	}

	preset, identity := readUSBIdentity(board.env().read)
	if preset != "custom" || identity.VendorID != "0x1234" || identity.Product != "KVM Keyboard" || identity.Serial != "SERIAL000001" {
		t.Fatalf("read back %s %+v", preset, identity)
	}

	// The same again is no rebuild.
	if f := applyUSBIdentity(custom, board.env()); f != nil || len(board.rebuilds) != 1 {
		t.Fatalf("repeat: %+v, %d rebuilds", f, len(board.rebuilds))
	}

	// A preset reads back by name.
	if f := applyUSBIdentity(proto.SetUSBIdentityReq{Preset: "dell"}, board.env()); f != nil {
		t.Fatal(f)
	}
	if preset, _ := readUSBIdentity(board.env().read); preset != "dell" {
		t.Fatalf("preset read back as %s", preset)
	}

	// The default with a serial is saved; the default without one removes
	// the file.
	if f := applyUSBIdentity(proto.SetUSBIdentityReq{Preset: "default", Serial: "0123456789AB"}, board.env()); f != nil || board.file == nil {
		t.Fatalf("default with a serial: %+v, file %q", f, board.file)
	}
	if f := applyUSBIdentity(proto.SetUSBIdentityReq{Preset: "default"}, board.env()); f != nil || board.file != nil {
		t.Fatalf("restore: %+v, file %q", f, board.file)
	}
	if len(board.rebuilds) != 4 {
		t.Fatalf("%d rebuilds, want 4", len(board.rebuilds))
	}
}

func TestApplyUSBIdentityRefusesBeforeWriting(t *testing.T) {
	board := &fakeIdentityBoard{}
	bad := proto.SetUSBIdentityReq{Preset: "custom", VendorID: "0x0000", ProductID: "1", Manufacturer: "a", Product: "b"}
	if f := applyUSBIdentity(bad, board.env()); f == nil || f.code != -2 {
		t.Fatalf("got %+v, want a -2 refusal", f)
	}
	if board.file != nil || len(board.rebuilds) != 0 {
		t.Fatal("a refused identity was written or rebuilt")
	}

	board.failSave = true
	if f := applyUSBIdentity(proto.SetUSBIdentityReq{Preset: "dell"}, board.env()); f == nil || f.code != -3 || len(board.rebuilds) != 0 {
		t.Fatalf("a failed save: %+v, %d rebuilds", f, len(board.rebuilds))
	}
}

func TestReadUSBIdentityFromHandEditedFiles(t *testing.T) {
	read := func(content string) func(string) ([]byte, error) {
		return func(string) ([]byte, error) { return []byte(content), nil }
	}

	// The values of a preset under the wrong name, or none, are custom.
	preset, _ := readUSBIdentity(read("preset=dell\nvid=0x046d\npid=0xc31c\nmanufacturer=Logitech\nproduct=USB Keyboard\nserial=\n"))
	if preset != "custom" {
		t.Fatalf("mismatched preset read as %s", preset)
	}
	preset, _ = readUSBIdentity(read("vid=0x046d\npid=0xc31c\nmanufacturer=Logitech\nproduct=USB Keyboard\n"))
	if preset != "custom" {
		t.Fatalf("unnamed preset read as %s", preset)
	}

	// An unusable field is the default.
	preset, identity := readUSBIdentity(read("preset=custom\nvid=zz\npid=0x1\nmanufacturer=a\nproduct=b\n"))
	if def, _ := usbIdentityPreset("default"); preset != "default" || identity != def {
		t.Fatalf("broken file read as %s %+v", preset, identity)
	}
}

// TestUSBIdentityDefaultMatchesTheScripts holds the default preset to the
// values S03usbdev and S03usbhid fall back to.
func TestUSBIdentityDefaultMatchesTheScripts(t *testing.T) {
	def, _ := usbIdentityPreset(usbIdentityDefault)
	for _, script := range []string{"../../../kvmapp/system/init.d/S03usbdev", "../../../kvmapp/system/init.d/S03usbhid"} {
		data, err := os.ReadFile(script)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		for _, want := range []string{
			"usb_vid=" + def.VendorID,
			"usb_pid=" + def.ProductID,
			"usb_manufacturer='" + def.Manufacturer + "'",
			"usb_product='" + def.Product + "'",
			usbIdentityFile,
		} {
			if !strings.Contains(text, want) {
				t.Errorf("%s does not contain %q", script, want)
			}
		}
	}
}
