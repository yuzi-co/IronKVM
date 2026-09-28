package vm

import (
	"bytes"
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/hid"
)

// The USB network is a private link between the board and the managed host.
// S03usbdev builds the function, puts the board on the first address of the
// subnet, and runs a udhcpd that leases the second address to the host with a
// subnet mask and nothing else: no router, no DNS server. The board forwards
// nothing that arrives on the link, so the host never reaches the LAN through
// the KVM.
//
// NCM is the mode an operator gets by default: Linux, macOS and Windows 11
// drive it with a driver they already carry. ECM is the fallback for a host
// with no NCM driver. RNDIS is only ever read back, from a board an older
// server set up; nothing here writes it.
const (
	usbNetworkOff   = "off"
	usbNetworkNCM   = "ncm"
	usbNetworkECM   = "ecm"
	usbNetworkRNDIS = "rndis"
)

// usbNetworkSubnetFile holds the subnet as one line in CIDR form. /etc/kvm is
// on /data, so the choice survives an update of the boot script, and
// S03usbdev reads it at every start. The default below is repeated in
// usb_net_subnet, and a test holds the two copies together.
const (
	usbNetworkSubnetFile    = "/etc/kvm/usb-network.subnet"
	defaultUSBNetworkSubnet = "172.31.255.0/30"
)

// usbNetworkModes lists each mode S03usbdev can build with its marker, in the
// order the script ranks them when more than one marker exists.
var usbNetworkModes = []struct {
	mode   string
	marker string
}{
	{usbNetworkNCM, virtualNetworkNCM},
	{usbNetworkECM, virtualNetworkECM},
	{usbNetworkRNDIS, virtualNetworkRNDIS},
}

// usbNetworkMode reads the mode from the markers the way S03usbdev does.
func usbNetworkMode(present func(string) bool) string {
	for _, candidate := range usbNetworkModes {
		if present(candidate.marker) {
			return candidate.mode
		}
	}

	return usbNetworkOff
}

// usbSubnet is a subnet the link can use, with the two addresses in it.
type usbSubnet struct {
	prefix netip.Prefix
	board  netip.Addr
	host   netip.Addr
}

// parseUSBSubnet accepts what usb_net_parse in S03usbdev accepts, and nothing
// more: a private IPv4 network, written as its network address, with a prefix
// from /24 to /30.
//
// /30 is the smallest subnet with two usable addresses, and /24 keeps the
// script's arithmetic inside one octet. The network has to be private because
// the host takes the lease as a real route, and an address on the internet
// would shadow a real one. The host bits have to be zero because the board
// takes the first address, and "first" means nothing in a network written as
// one of its hosts. netip already refuses an octet with a leading zero, which
// ash would read as octal.
func parseUSBSubnet(value string) (usbSubnet, error) {
	prefix, err := netip.ParsePrefix(strings.TrimSpace(value))
	if err != nil {
		return usbSubnet{}, fmt.Errorf("%q is not a subnet in CIDR form, such as %s", value, defaultUSBNetworkSubnet)
	}

	if !prefix.Addr().Is4() {
		return usbSubnet{}, fmt.Errorf("%s is not an IPv4 subnet", prefix)
	}

	if prefix.Bits() < 24 || prefix.Bits() > 30 {
		return usbSubnet{}, fmt.Errorf("%s is a /%d, and the link takes a /24 to a /30", prefix, prefix.Bits())
	}

	if prefix.Masked() != prefix {
		return usbSubnet{}, fmt.Errorf("%s is not a network address, did you mean %s", prefix, prefix.Masked())
	}

	if !prefix.Addr().IsPrivate() {
		return usbSubnet{}, fmt.Errorf("%s is not a private network (10.0.0.0/8, 172.16.0.0/12 or 192.168.0.0/16)", prefix)
	}

	board := prefix.Addr().Next()

	return usbSubnet{prefix: prefix, board: board, host: board.Next()}, nil
}

// readUSBSubnet reads the subnet in use, the way usb_net_subnet does: the
// first line of the file, or the default when the file is absent or holds
// something the script would refuse.
func readUSBSubnet(read func(string) ([]byte, error)) usbSubnet {
	fallback, _ := parseUSBSubnet(defaultUSBNetworkSubnet)

	content, err := read(usbNetworkSubnetFile)
	if err != nil {
		return fallback
	}

	line, _, _ := bytes.Cut(content, []byte("\n"))

	subnet, err := parseUSBSubnet(string(line))
	if err != nil {
		return fallback
	}

	return subnet
}

// writeUSBSubnet replaces the subnet file whole. A torn write would leave the
// script a line it refuses, and it would fall back to the default silently
// from the operator's point of view.
func writeUSBSubnet(subnet usbSubnet) error {
	if err := os.MkdirAll(filepath.Dir(usbNetworkSubnetFile), 0o755); err != nil {
		return err
	}

	tmp := usbNetworkSubnetFile + ".tmp"
	if err := os.WriteFile(tmp, []byte(subnet.prefix.String()+"\n"), 0o644); err != nil {
		return err
	}

	return os.Rename(tmp, usbNetworkSubnetFile)
}

// usbSubnetOverlap names the first address on another interface of the board
// that falls inside the subnet, or overlaps it. A link subnet that covers the
// LAN would take the board's own route to the LAN away from it.
//
// The USB interfaces themselves are left out: they hold the subnet in use,
// and a change of subnet would otherwise always collide with the old one.
func usbSubnetOverlap(subnet usbSubnet, local []netip.Prefix) (netip.Prefix, bool) {
	for _, prefix := range local {
		if prefix.Overlaps(subnet.prefix) {
			return prefix, true
		}
	}

	return netip.Prefix{}, false
}

// localPrefixes lists the IPv4 subnets on the board's interfaces other than
// loopback and the USB links.
func localPrefixes() []netip.Prefix {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Errorf("list interfaces: %s", err)
		return nil
	}

	var prefixes []netip.Prefix

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || strings.HasPrefix(iface.Name, "usb") {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip, ok := netip.AddrFromSlice(ipNet.IP)
			if !ok {
				continue
			}

			ip = ip.Unmap()
			if !ip.Is4() {
				continue
			}

			bits, _ := ipNet.Mask.Size()
			prefixes = append(prefixes, netip.PrefixFrom(ip, bits).Masked())
		}
	}

	return prefixes
}

// usbNetworkFits answers whether the link can be in the given mode beside the
// functions that are on now, and why not when it cannot.
//
// Turning it off always fits. A board that already carries the network can
// switch between NCM and ECM freely, because both cost the same two inbound
// endpoints and one outbound. Only turning it on from off is checked, and that
// goes through the same canEnable the other switches use, so the refusal names
// what to turn off first.
func usbNetworkFits(mode string, present func(string) bool) (bool, string) {
	if mode == usbNetworkOff || usbNetworkMode(present) != usbNetworkOff {
		return true, ""
	}

	ok, free, relief := canEnable("network", present)
	if ok {
		return true, ""
	}

	return false, refusalMessage("network", free, relief)
}

// usbNetworkCommands rebuilds the gadget in the given mode. Every network
// marker comes off and at most one goes back on, so a board never carries two
// and S03usbdev never has to choose.
//
// The config symlinks are left to S03usbdev: a start that drops the network
// prunes all three, and a start that keeps it unlinks the two it did not
// choose. The markers go with rm -f, for the reason unmountConsoleCommands
// records: a failed removal would stop the list before `start`.
func usbNetworkCommands(mode string) []string {
	commands := []string{
		"/etc/init.d/S03usbdev stop",
		"rm -f /boot/usb.ncm",
		"rm -f /boot/usb.ecm",
		"rm -f /boot/usb.rndis0",
	}

	switch mode {
	case usbNetworkNCM:
		commands = append(commands, "touch "+virtualNetworkNCM)
	case usbNetworkECM:
		commands = append(commands, "touch "+virtualNetworkECM)
	}

	return append(commands, "/etc/init.d/S03usbdev start")
}

// needsRebuild answers whether a save has to rebuild the gadget. A new mode
// always does. A new subnet does only while the link is on, because the
// address and udhcpd are set up by S03usbdev start; with the link off the file
// is simply read at the next start.
func needsRebuild(current string, wanted string, subnetChanged bool) bool {
	if current != wanted {
		return true
	}

	return subnetChanged && wanted != usbNetworkOff
}

func usbNetworkState(present func(string) bool, read func(string) ([]byte, error)) proto.GetUSBNetworkRsp {
	mode := usbNetworkMode(present)
	subnet := readUSBSubnet(read)
	fits, refusal := usbNetworkFits(usbNetworkNCM, present)

	return proto.GetUSBNetworkRsp{
		Mode:    mode,
		Subnet:  subnet.prefix.String(),
		Board:   subnet.board.String(),
		Host:    subnet.host.String(),
		Active:  mode != usbNetworkOff && isFunctionActive("network"),
		Fits:    fits,
		Refusal: refusal,
	}
}

// USBLink is the link as another service needs it, network boot among them:
// the mode, the subnet with the board's and the host's address, and whether
// the function is bound.
type USBLink struct {
	Mode   string
	Prefix netip.Prefix
	Board  netip.Addr
	Host   netip.Addr
	Active bool
}

// CurrentUSBLink reads the link as S03usbdev sets it up.
func CurrentUSBLink() USBLink {
	mode := usbNetworkMode(markerPresent)
	subnet := readUSBSubnet(os.ReadFile)

	return USBLink{
		Mode:   mode,
		Prefix: subnet.prefix,
		Board:  subnet.board,
		Host:   subnet.host,
		Active: mode != usbNetworkOff && isFunctionActive("network"),
	}
}

func markerPresent(marker string) bool {
	exist, _ := isDeviceExist(marker)
	return exist
}

func (s *Service) GetUSBNetwork(c *gin.Context) {
	var rsp proto.Response

	rsp.OkRspWithData(c, usbNetworkState(markerPresent, os.ReadFile))

	log.Debugf("get usb network success")
}

func (s *Service) SetUSBNetwork(c *gin.Context) {
	var req proto.SetUSBNetworkReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid argument")
		return
	}

	current := readUSBSubnet(os.ReadFile)
	subnet := current

	if req.Subnet != "" {
		parsed, err := parseUSBSubnet(req.Subnet)
		if err != nil {
			rsp.ErrRsp(c, -2, err.Error())
			return
		}

		if prefix, overlaps := usbSubnetOverlap(parsed, localPrefixes()); overlaps {
			rsp.ErrRsp(c, -2, fmt.Sprintf("%s overlaps %s, which the board already uses", parsed.prefix, prefix))
			return
		}

		subnet = parsed
	}

	// Refuse rather than drop something, as UpdateVirtualDevice does.
	if ok, refusal := usbNetworkFits(req.Mode, markerPresent); !ok {
		log.Infof("refused usb network %s: %s", req.Mode, refusal)
		rsp.ErrRsp(c, -4, refusal)
		return
	}

	subnetChanged := subnet.prefix != current.prefix
	if subnetChanged {
		if err := writeUSBSubnet(subnet); err != nil {
			log.Errorf("write %s: %s", usbNetworkSubnetFile, err)
			rsp.ErrRsp(c, -3, "operation failed")
			return
		}
	}

	// Re-enumerating the gadget costs the host its keyboard for a few
	// seconds, so a save that changes nothing the gadget carries must not.
	if !needsRebuild(usbNetworkMode(markerPresent), req.Mode, subnetChanged) {
		rsp.OkRspWithData(c, usbNetworkState(markerPresent, os.ReadFile))
		return
	}

	h := hid.GetHid()
	h.Lock()
	h.CloseNoLock()
	defer func() {
		h.OpenNoLock()
		h.Unlock()
	}()

	for _, command := range usbNetworkCommands(req.Mode) {
		if err := exec.Command("sh", "-c", command).Run(); err != nil {
			log.Errorf("usb network: %s: %s", command, err)
			rsp.ErrRsp(c, -3, "operation failed")
			return
		}
	}

	rsp.OkRspWithData(c, usbNetworkState(markerPresent, os.ReadFile))

	log.Debugf("set usb network %s %s success", req.Mode, subnet.prefix)
}
