// Package wol keeps the Wake-on-LAN device list and sends the wake packets.
//
// The list has two files. MacFile is the original format, one "MAC name" line
// per device: it stays as it was so a board that goes back to an older
// application still reads it. What the old format has no room for (the
// broadcast address, the UDP port, the SecureOn password and the host IP)
// lives in OptionsFile, keyed by MAC. An older application ignores that file,
// and a device without an entry in it behaves as it always did.
package wol

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"NanoKVM-Server/utils/atomicfile"
)

const (
	// MacFile is the device list, one "MAC name" line each.
	MacFile = "/etc/kvm/cache/wol"
	// OptionsFile holds the per-device options. It carries the SecureOn
	// passwords, so it is readable by root only.
	OptionsFile = "/etc/kvm/wol-devices.json"

	maxNameRunes = 64
)

var (
	ErrInvalidMAC  = errors.New("invalid MAC address")
	ErrInvalidName = errors.New("invalid name")
	ErrDuplicate   = errors.New("a device with this MAC address or name already exists")
	ErrNotFound    = errors.New("no such device")
)

// Options are the per-device settings. The zero value is the original
// behaviour: an Ethernet frame sent with ether-wake.
type Options struct {
	// Broadcast is an IPv4 address or a CIDR subnet; the packet goes to the
	// address itself or to the broadcast address of the subnet.
	Broadcast string `json:"broadcast,omitempty"`
	// Port is the UDP port, 7 or 9. Zero means an Ethernet frame, unless a
	// Broadcast is set, which implies UDP on port 9.
	Port int `json:"port,omitempty"`
	// SecureOn is the six-byte SecureOn password as a MAC-style string.
	SecureOn string `json:"secureOn,omitempty"`
	// IP is the host's address, for the awake check.
	IP string `json:"ip,omitempty"`
}

// Device is one entry of the list.
type Device struct {
	Mac  string
	Name string
	Options
}

// Store is the device list on disk.
type Store struct {
	macPath string
	optPath string
	mu      sync.Mutex
}

// NewStore returns a store over the two files.
func NewStore(macPath, optPath string) *Store {
	return &Store{macPath: macPath, optPath: optPath}
}

// Default is the store of the device's own files.
var Default = NewStore(MacFile, OptionsFile)

// ParseMAC normalizes a MAC address to upper case with colons.
func ParseMAC(mac string) (string, error) {
	mac = strings.ToUpper(strings.TrimSpace(mac))
	mac = strings.NewReplacer("-", "", ":", "", ".", "").Replace(mac)
	if len(mac) != 12 {
		return "", ErrInvalidMAC
	}
	for _, r := range mac {
		if (r < '0' || r > '9') && (r < 'A' || r > 'F') {
			return "", ErrInvalidMAC
		}
	}

	var b strings.Builder
	for i := 0; i < 12; i += 2 {
		if i > 0 {
			b.WriteByte(':')
		}
		b.WriteString(mac[i : i+2])
	}

	return b.String(), nil
}

// SanitizeName collapses white space, as the list has always done.
func SanitizeName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}

// ValidateOptions normalizes the options and rejects what cannot be sent.
func ValidateOptions(o Options) (Options, error) {
	o.Broadcast = strings.TrimSpace(o.Broadcast)
	o.SecureOn = strings.TrimSpace(o.SecureOn)
	o.IP = strings.TrimSpace(o.IP)

	if o.Broadcast != "" {
		if _, err := BroadcastTarget(o.Broadcast); err != nil {
			return o, err
		}
	}

	if o.Port != 0 && o.Port != 7 && o.Port != 9 {
		return o, inputErr("the port must be 7 or 9")
	}

	if o.SecureOn != "" {
		sec, err := ParseMAC(o.SecureOn)
		if err != nil {
			return o, inputErr("the SecureOn password must be six bytes in hex")
		}
		o.SecureOn = sec
	}

	if o.IP != "" {
		ip := net.ParseIP(o.IP)
		if ip == nil || ip.To4() == nil || ip.IsUnspecified() || ip.IsMulticast() || ip.IsLoopback() {
			return o, inputErr("the host address is not a usable IPv4 address")
		}
		o.IP = ip.To4().String()
	}

	return o, nil
}

// BroadcastTarget returns the address a packet goes to for a setting that is
// an IPv4 address or a CIDR subnet.
func BroadcastTarget(value string) (net.IP, error) {
	if strings.Contains(value, "/") {
		ip, network, err := net.ParseCIDR(value)
		if err != nil || ip.To4() == nil {
			return nil, inputErr("the broadcast subnet is not an IPv4 subnet")
		}
		ones, _ := network.Mask.Size()
		if ones < 1 || ones > 30 {
			return nil, inputErr("the subnet prefix length must be between 1 and 30")
		}
		base := network.IP.To4()
		out := make(net.IP, 4)
		for i := range out {
			out[i] = base[i] | ^network.Mask[i]
		}

		return out, nil
	}

	ip := net.ParseIP(value)
	if ip == nil || ip.To4() == nil || ip.IsUnspecified() || ip.IsMulticast() || ip.IsLoopback() {
		return nil, inputErr("the broadcast address is not a usable IPv4 address")
	}

	return ip.To4(), nil
}

type macLine struct {
	mac  string
	name string
}

func splitLine(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", "", false
	}

	return fields[0], strings.TrimSpace(line[len(fields[0]):]), true
}

func (s *Store) readLines() ([]macLine, error) {
	content, err := os.ReadFile(s.macPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var lines []macLine
	for _, raw := range strings.Split(string(content), "\n") {
		mac, name, ok := splitLine(raw)
		if !ok {
			continue
		}
		if normalized, err := ParseMAC(mac); err == nil {
			mac = normalized
		}
		lines = append(lines, macLine{mac: mac, name: name})
	}

	return lines, nil
}

func (s *Store) writeLines(lines []macLine) error {
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l.mac)
		if l.name != "" {
			b.WriteString(" " + l.name)
		}
		b.WriteByte('\n')
	}

	return writeFile(s.macPath, []byte(b.String()), 0o644)
}

func (s *Store) readOptions() (map[string]Options, error) {
	opts := map[string]Options{}
	content, err := os.ReadFile(s.optPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return opts, nil
		}
		return nil, err
	}
	if strings.TrimSpace(string(content)) == "" {
		return opts, nil
	}
	if err := json.Unmarshal(content, &opts); err != nil {
		return nil, fmt.Errorf("%s: %w", s.optPath, err)
	}

	return opts, nil
}

func (s *Store) writeOptions(opts map[string]Options) error {
	for mac, o := range opts {
		if o == (Options{}) {
			delete(opts, mac)
		}
	}
	data, err := json.MarshalIndent(opts, "", "  ")
	if err != nil {
		return err
	}

	return writeFile(s.optPath, append(data, '\n'), 0o600)
}

// writeFile replaces path, making its directory first: /etc/kvm/cache does
// not exist on a board that has never saved a device.
func writeFile(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return atomicfile.WriteFile(path, data, mode)
}

// List returns the devices in file order.
func (s *Store) List() ([]Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	lines, err := s.readLines()
	if err != nil {
		return nil, err
	}
	opts, err := s.readOptions()
	if err != nil {
		// A damaged options file must not hide the list.
		opts = map[string]Options{}
	}

	out := make([]Device, 0, len(lines))
	for _, l := range lines {
		out = append(out, Device{Mac: l.mac, Name: l.name, Options: opts[l.mac]})
	}

	return out, nil
}

// Get returns one device by MAC.
func (s *Store) Get(mac string) (Device, error) {
	mac, err := ParseMAC(mac)
	if err != nil {
		return Device{}, err
	}
	list, err := s.List()
	if err != nil {
		return Device{}, err
	}
	for _, d := range list {
		if d.Mac == mac {
			return d, nil
		}
	}

	return Device{}, ErrNotFound
}

// Find returns the device with this name, ignoring case.
func (s *Store) Find(name string) (Device, error) {
	name = SanitizeName(name)
	if name == "" {
		return Device{}, ErrNotFound
	}
	list, err := s.List()
	if err != nil {
		return Device{}, err
	}

	var found *Device
	for i := range list {
		if strings.EqualFold(list[i].Name, name) {
			if found != nil {
				return Device{}, fmt.Errorf("more than one device is named %q", name)
			}
			found = &list[i]
		}
	}
	if found == nil {
		return Device{}, ErrNotFound
	}

	return *found, nil
}

// Touch records a MAC that was woken, if the list does not hold it yet. This
// is how entries have always appeared.
func (s *Store) Touch(mac string) error {
	mac, err := ParseMAC(mac)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	lines, err := s.readLines()
	if err != nil {
		return err
	}
	for _, l := range lines {
		if l.mac == mac {
			return nil
		}
	}

	return s.writeLines(append(lines, macLine{mac: mac}))
}

// Rename sets the name of a listed device.
func (s *Store) Rename(mac, name string) error {
	mac, err := ParseMAC(mac)
	if err != nil {
		return err
	}
	name = SanitizeName(name)
	if name == "" || utf8.RuneCountInString(name) > maxNameRunes {
		return ErrInvalidName
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	lines, err := s.readLines()
	if err != nil {
		return err
	}
	found := false
	for i := range lines {
		if lines[i].mac == mac {
			lines[i].name = name
			found = true
		}
	}
	if !found {
		return ErrNotFound
	}

	return s.writeLines(lines)
}

// Delete removes a device and its options.
func (s *Store) Delete(mac string) error {
	mac, err := ParseMAC(mac)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	lines, err := s.readLines()
	if err != nil {
		return err
	}
	kept := lines[:0]
	for _, l := range lines {
		if l.mac != mac {
			kept = append(kept, l)
		}
	}
	if err := s.writeLines(kept); err != nil {
		return err
	}

	opts, err := s.readOptions()
	if err != nil {
		return nil
	}
	if _, ok := opts[mac]; ok {
		delete(opts, mac)
		return s.writeOptions(opts)
	}

	return nil
}

// Save adds a device, or edits the one at oldMac when oldMac is not empty.
// MACs and names are unique, ignoring case for names. With keepSecureOn an
// empty SecureOn leaves the stored password alone: the list never sends it
// back, so an edit form cannot know it.
func (s *Store) Save(oldMac string, d Device, keepSecureOn bool) (Device, error) {
	mac, err := ParseMAC(d.Mac)
	if err != nil {
		return Device{}, err
	}
	d.Mac = mac
	d.Name = SanitizeName(d.Name)
	if utf8.RuneCountInString(d.Name) > maxNameRunes {
		return Device{}, fmt.Errorf("%w: at most %d characters", ErrInvalidName, maxNameRunes)
	}
	if d.Options, err = ValidateOptions(d.Options); err != nil {
		return Device{}, err
	}
	if oldMac != "" {
		if oldMac, err = ParseMAC(oldMac); err != nil {
			return Device{}, err
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	lines, err := s.readLines()
	if err != nil {
		return Device{}, err
	}
	opts, err := s.readOptions()
	if err != nil {
		return Device{}, err
	}

	index := -1
	for i, l := range lines {
		if oldMac != "" && l.mac == oldMac {
			index = i
			continue
		}
		if l.mac == d.Mac || (d.Name != "" && strings.EqualFold(l.name, d.Name)) {
			return Device{}, ErrDuplicate
		}
	}
	if oldMac != "" && index < 0 {
		return Device{}, ErrNotFound
	}

	if keepSecureOn && d.SecureOn == "" && oldMac != "" {
		d.SecureOn = opts[oldMac].SecureOn
	}

	line := macLine{mac: d.Mac, name: d.Name}
	if index >= 0 {
		lines[index] = line
		delete(opts, oldMac)
	} else {
		lines = append(lines, line)
	}
	opts[d.Mac] = d.Options

	// The options go first. A failure between the two writes then leaves an
	// option entry for a MAC that is not in the list, which nothing reads,
	// and never a listed device that lost its options.
	if err := s.writeOptions(opts); err != nil {
		return Device{}, err
	}
	if err := s.writeLines(lines); err != nil {
		return Device{}, err
	}

	return d, nil
}

// InputError is a value the caller typed that cannot be used; its text is
// meant for that person.
type InputError struct{ msg string }

func (e *InputError) Error() string { return e.msg }

func inputErr(format string, args ...any) error {
	return &InputError{msg: fmt.Sprintf(format, args...)}
}
