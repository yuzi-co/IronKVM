package vm

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
)

// The USB identity is the vendor and product IDs and the strings the gadget
// shows the host. Some hosts treat an unknown device differently from a
// keyboard they recognise, and some operators would rather the KVM did not
// announce itself, so the settings offer a few common keyboards by name as
// well as custom values.
//
// The choice is saved to /etc/kvm/usb-identity, which is on /data. /boot holds
// the older /boot/usb.vid and /boot/usb.pid, and both partitions are shared by
// the two slots, so either would keep the identity across a slot switch. /data
// is chosen because it is where the other runtime settings of the gadget live
// (hid_only, the USB network subnet), and because the boot partition carries
// the kernel both slots boot: a write cut short by a power loss there costs
// more than a setting. The /boot files keep working, below a saved identity.
//
// S03usbdev and S03usbhid read the file each time they build the gadget and
// check every field again, so a file edited by hand cannot put a malformed ID
// into configfs. The format is one key=value per line.
const usbIdentityFile = "/etc/kvm/usb-identity"

// The boot partition's overrides, read by both scripts when no identity is
// saved.
const (
	usbBootVendorFile  = "/boot/usb.vid"
	usbBootProductFile = "/boot/usb.pid"
)

// usbGadgetDir is the live gadget, read back to show what the host sees now.
const usbGadgetDir = "/sys/kernel/config/usb_gadget/g0"

const (
	usbIdentityDefault = "default"
	usbIdentityCustom  = "custom"

	// usbStringMax bounds a string descriptor. The kernel accepts 126
	// characters; nothing a host shows needs half of that.
	usbStringMax = 64
	// usbSerialMin is what the mass storage class asks of a serial number.
	// S03usbdev falls back to a placeholder for the same reason.
	usbSerialMin = 12
)

// usbIdentityPresets are the identities offered by name. The first is the
// board's own, and its values are repeated in S03usbdev and S03usbhid. The
// others are ordinary wired keyboards that no host binds a vendor driver to.
// A receiver's ID (Logitech's Unifying receiver, for one) would not do: Linux
// loads a driver for it that expects the receiver's own protocol.
var usbIdentityPresets = []proto.USBIdentityPreset{
	{Name: usbIdentityDefault, USBIdentity: proto.USBIdentity{
		VendorID: "0x3346", ProductID: "0x1009", Manufacturer: "sipeed", Product: "NanoKVM",
	}},
	{Name: "logitech", USBIdentity: proto.USBIdentity{
		VendorID: "0x046d", ProductID: "0xc31c", Manufacturer: "Logitech", Product: "USB Keyboard",
	}},
	{Name: "microsoft", USBIdentity: proto.USBIdentity{
		VendorID: "0x045e", ProductID: "0x07f8", Manufacturer: "Microsoft", Product: "Wired Keyboard 600",
	}},
	{Name: "dell", USBIdentity: proto.USBIdentity{
		VendorID: "0x413c", ProductID: "0x2113", Manufacturer: "Dell", Product: "Dell KB216 Wired Keyboard",
	}},
}

func usbIdentityPreset(name string) (proto.USBIdentity, bool) {
	for _, preset := range usbIdentityPresets {
		if preset.Name == name {
			return preset.USBIdentity, true
		}
	}
	return proto.USBIdentity{}, false
}

// parseUSBID accepts one to four hex digits, with or without 0x, and returns
// the four-digit form configfs is written with. Zero is refused: it is no
// vendor and no product.
func parseUSBID(value string) (string, error) {
	digits := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(value)), "0x")
	if digits == "" || len(digits) > 4 {
		return "", fmt.Errorf("%q is not a 16-bit hex ID", value)
	}
	id, err := strconv.ParseUint(digits, 16, 16)
	if err != nil {
		return "", fmt.Errorf("%q is not a 16-bit hex ID", value)
	}
	if id == 0 {
		return "", errors.New("an ID of 0 is not valid")
	}
	return fmt.Sprintf("0x%04x", id), nil
}

// validUSBString accepts printable ASCII. The scripts read each value as one
// line, so a newline would end it early, and a host shows nothing sensible for
// control characters.
func validUSBString(name, value string) error {
	if value == "" {
		return fmt.Errorf("%s is empty", name)
	}
	if len(value) > usbStringMax {
		return fmt.Errorf("%s is longer than %d characters", name, usbStringMax)
	}
	for _, r := range value {
		if r < 0x20 || r > 0x7e {
			return fmt.Errorf("%s may hold printable ASCII only", name)
		}
	}
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("%s starts or ends with a space", name)
	}
	return nil
}

// validUSBSerial accepts an empty serial, which keeps the board's own, or 12 to
// 64 letters and digits.
func validUSBSerial(value string) error {
	if value == "" {
		return nil
	}
	if len(value) < usbSerialMin || len(value) > usbStringMax {
		return fmt.Errorf("the serial number needs %d to %d characters", usbSerialMin, usbStringMax)
	}
	for _, r := range value {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z') {
			return errors.New("the serial number may hold letters and digits only")
		}
	}
	return nil
}

// resolveUSBIdentity turns a request into the preset name and identity to save.
func resolveUSBIdentity(req proto.SetUSBIdentityReq) (string, proto.USBIdentity, error) {
	if err := validUSBSerial(req.Serial); err != nil {
		return "", proto.USBIdentity{}, err
	}

	if preset, ok := usbIdentityPreset(req.Preset); ok {
		preset.Serial = req.Serial
		return req.Preset, preset, nil
	}
	if req.Preset != usbIdentityCustom {
		return "", proto.USBIdentity{}, fmt.Errorf("unknown preset %q", req.Preset)
	}

	vendor, err := parseUSBID(req.VendorID)
	if err != nil {
		return "", proto.USBIdentity{}, fmt.Errorf("vendor ID: %w", err)
	}
	product, err := parseUSBID(req.ProductID)
	if err != nil {
		return "", proto.USBIdentity{}, fmt.Errorf("product ID: %w", err)
	}
	if err := validUSBString("the manufacturer", req.Manufacturer); err != nil {
		return "", proto.USBIdentity{}, err
	}
	if err := validUSBString("the product name", req.Product); err != nil {
		return "", proto.USBIdentity{}, err
	}

	return usbIdentityCustom, proto.USBIdentity{
		VendorID:     vendor,
		ProductID:    product,
		Manufacturer: req.Manufacturer,
		Product:      req.Product,
		Serial:       req.Serial,
	}, nil
}

// formatUSBIdentity is the file's contents. The preset line is for the
// settings page; the scripts read the other five.
func formatUSBIdentity(preset string, identity proto.USBIdentity) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "preset=%s\n", preset)
	fmt.Fprintf(&b, "vid=%s\n", identity.VendorID)
	fmt.Fprintf(&b, "pid=%s\n", identity.ProductID)
	fmt.Fprintf(&b, "manufacturer=%s\n", identity.Manufacturer)
	fmt.Fprintf(&b, "product=%s\n", identity.Product)
	fmt.Fprintf(&b, "serial=%s\n", identity.Serial)
	return b.Bytes()
}

// readUSBIdentity reads the saved identity. A missing file is the default. A
// file with an unusable field reads as the default too, and a warning: the
// scripts drop such a field and build the rest, so this does not claim the
// file's values reach the host.
func readUSBIdentity(read func(string) ([]byte, error)) (string, proto.USBIdentity) {
	fallback, _ := usbIdentityPreset(usbIdentityDefault)

	data, err := read(usbIdentityFile)
	if err != nil {
		return usbIdentityDefault, fallback
	}

	fields := map[string]string{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), "=")
		if !ok {
			continue
		}
		// The first line of a key wins, as the scripts' sed | head does.
		if _, seen := fields[key]; !seen {
			fields[key] = value
		}
	}

	_, identity, err := resolveUSBIdentity(proto.SetUSBIdentityReq{
		Preset:       usbIdentityCustom,
		VendorID:     fields["vid"],
		ProductID:    fields["pid"],
		Manufacturer: fields["manufacturer"],
		Product:      fields["product"],
		Serial:       fields["serial"],
	})
	if err != nil {
		log.Warnf("%s: %s, showing the default identity", usbIdentityFile, err)
		return usbIdentityDefault, fallback
	}

	// A preset name stands only while the values are still the preset's, so
	// a file edited by hand reads back as custom.
	preset := fields["preset"]
	want, ok := usbIdentityPreset(preset)
	want.Serial = identity.Serial
	if !ok || want != identity {
		preset = usbIdentityCustom
	}

	return preset, identity
}

// readLiveUSBIdentity reads what the gadget carries now. A field that cannot be
// read is left empty.
func readLiveUSBIdentity(read func(string) ([]byte, error)) proto.USBIdentity {
	field := func(name string) string {
		data, err := read(filepath.Join(usbGadgetDir, name))
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(data))
	}

	return proto.USBIdentity{
		VendorID:     field("idVendor"),
		ProductID:    field("idProduct"),
		Manufacturer: field("strings/0x409/manufacturer"),
		Product:      field("strings/0x409/product"),
		Serial:       field("strings/0x409/serialnumber"),
	}
}

func usbIdentityState(read func(string) ([]byte, error), exists func(string) bool) *proto.GetUSBIdentityRsp {
	preset, identity := readUSBIdentity(read)
	return &proto.GetUSBIdentityRsp{
		Preset:       preset,
		Identity:     identity,
		Live:         readLiveUSBIdentity(read),
		Presets:      usbIdentityPresets,
		BootOverride: exists(usbBootVendorFile) || exists(usbBootProductFile),
	}
}

// usbIdentityEnv is what applying needs from the board, so a test can run it
// with nothing touched.
type usbIdentityEnv struct {
	read    func(string) ([]byte, error)
	write   func([]byte) error
	remove  func() error
	rebuild func([]string) error
}

// applyUSBIdentity saves the identity and rebuilds the gadget with it. The
// default with the board's own serial is no saved identity at all: the file is
// removed, so the board is back where it started, /boot overrides included.
//
// Nothing is rebuilt when the saved identity does not change. A rebuild costs
// the host its keyboard, mouse and disk for a few seconds while it enumerates
// the gadget again.
func applyUSBIdentity(req proto.SetUSBIdentityReq, env usbIdentityEnv) *applyFailure {
	preset, identity, err := resolveUSBIdentity(req)
	if err != nil {
		return &applyFailure{-2, err.Error()}
	}

	_, readErr := env.read(usbIdentityFile)
	saved := readErr == nil
	restore := preset == usbIdentityDefault && identity.Serial == ""

	if restore && !saved {
		return nil
	}
	if !restore && saved {
		oldPreset, oldIdentity := readUSBIdentity(env.read)
		if oldPreset == preset && oldIdentity == identity {
			return nil
		}
	}

	if restore {
		err = env.remove()
	} else {
		err = env.write(formatUSBIdentity(preset, identity))
	}
	if err != nil {
		log.Errorf("save %s: %s", usbIdentityFile, err)
		return &applyFailure{-3, "operation failed"}
	}

	if err := env.rebuild([]string{usbGadgetStop, usbGadgetStart}); err != nil {
		return &applyFailure{-3, "operation failed"}
	}

	return nil
}

func writeUSBIdentityFile(data []byte) error {
	if err := os.MkdirAll(filepath.Dir(usbIdentityFile), 0o755); err != nil {
		return err
	}

	tmp := usbIdentityFile + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}

	return os.Rename(tmp, usbIdentityFile)
}

func removeUSBIdentityFile() error {
	if err := os.Remove(usbIdentityFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func usbIdentityPathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (s *Service) GetUSBIdentity(c *gin.Context) {
	var rsp proto.Response

	rsp.OkRspWithData(c, usbIdentityState(os.ReadFile, usbIdentityPathExists))
}

func (s *Service) SetUSBIdentity(c *gin.Context) {
	var req proto.SetUSBIdentityReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid argument")
		return
	}

	failure := applyUSBIdentity(req, usbIdentityEnv{
		read:    os.ReadFile,
		write:   writeUSBIdentityFile,
		remove:  removeUSBIdentityFile,
		rebuild: rebuildWithHidClosed,
	})
	if failure != nil {
		rsp.ErrRsp(c, failure.code, failure.message)
		return
	}

	rsp.OkRspWithData(c, usbIdentityState(os.ReadFile, usbIdentityPathExists))
	log.Infof("usb identity set to %s", req.Preset)
}
