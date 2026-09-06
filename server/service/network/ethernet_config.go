package network

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"NanoKVM-Server/proto"
)

const (
	ethModeDHCP   = "dhcp"
	ethModeStatic = "static"

	// defaultEthPrefix matches what S30eth assumes for a line that carries no
	// prefix length.
	defaultEthPrefix = 16

	ethInterface = "eth0"
)

// ethConfigFile is what S30eth reads at boot. Its presence is the whole
// difference between a DHCP board and a static one, so the file is written
// only after somebody confirms that the new address answers. A board that
// loses power in the middle of a trial comes back on DHCP.
//
// It is a var so the tests can point it somewhere they may write.
var ethConfigFile = "/boot/eth.nodhcp"

// ethernetConfig is one line of ethConfigFile.
type ethernetConfig struct {
	Address string
	Prefix  int
	Gateway string
}

// readEthernetConfig returns the saved static settings. A missing file is the
// DHCP case and is not an error.
func readEthernetConfig() (ethernetConfig, bool) {
	data, err := os.ReadFile(ethConfigFile)
	if err != nil {
		return ethernetConfig{}, false
	}

	for _, line := range strings.Split(string(data), "\n") {
		config, ok := parseEthernetLine(line)
		if ok {
			return config, true
		}
	}

	return ethernetConfig{}, false
}

// parseEthernetLine reads one `<address>[/<prefix>] [gateway]` line. S30eth
// takes the first line it can use and ignores the rest, so this does too.
func parseEthernetLine(line string) (ethernetConfig, bool) {
	fields := strings.Fields(strings.ReplaceAll(strings.TrimSpace(line), "\r", ""))
	if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
		return ethernetConfig{}, false
	}

	address, prefix := fields[0], defaultEthPrefix
	if index := strings.Index(address, "/"); index >= 0 {
		parsed, err := strconv.Atoi(address[index+1:])
		// S30eth falls back to /16 for a prefix it cannot use, so an
		// unreadable one is not a reason to reject the whole line.
		if err == nil && parsed >= 1 && parsed <= 32 {
			prefix = parsed
		}
		address = address[:index]
	}

	if net.ParseIP(address).To4() == nil {
		return ethernetConfig{}, false
	}

	config := ethernetConfig{Address: address, Prefix: prefix}
	if len(fields) > 1 && net.ParseIP(fields[1]).To4() != nil {
		config.Gateway = fields[1]
	}

	return config, true
}

// render writes the line in the form S30eth reads.
func (c ethernetConfig) render() string {
	line := fmt.Sprintf("%s/%d", c.Address, c.Prefix)
	if c.Gateway != "" {
		line += " " + c.Gateway
	}

	return line + "\n"
}

func writeEthernetConfig(config ethernetConfig) error {
	if err := os.WriteFile(ethConfigFile, []byte(config.render()), 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", ethConfigFile, err)
	}

	return nil
}

func removeEthernetConfig() error {
	if err := os.Remove(ethConfigFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove %s: %w", ethConfigFile, err)
	}

	return nil
}

// validateEthernetConfig rejects a setting that would strand the board. Every
// case here is one that the kernel accepts and a person cannot reach.
func validateEthernetConfig(config ethernetConfig) error {
	address := net.ParseIP(config.Address)
	if address == nil || address.To4() == nil {
		return fmt.Errorf("the address is not an IPv4 address")
	}
	address = address.To4()

	if address.IsLoopback() || address.IsMulticast() || address.IsUnspecified() {
		return fmt.Errorf("the address cannot be used on an interface")
	}

	// A /31 leaves no host bits for a gateway and a /32 leaves no subnet at
	// all. Both need routing this device does not set up.
	if config.Prefix < 1 || config.Prefix > 30 {
		return fmt.Errorf("the prefix length must be between 1 and 30")
	}

	mask := net.CIDRMask(config.Prefix, 32)
	network := address.Mask(mask)

	if address.Equal(network) {
		return fmt.Errorf("the address is the network address of its subnet")
	}
	if address.Equal(broadcastOf(network, mask)) {
		return fmt.Errorf("the address is the broadcast address of its subnet")
	}

	if config.Gateway == "" {
		return nil
	}

	gateway := net.ParseIP(config.Gateway)
	if gateway == nil || gateway.To4() == nil {
		return fmt.Errorf("the router is not an IPv4 address")
	}
	gateway = gateway.To4()

	if gateway.Equal(address) {
		return fmt.Errorf("the router cannot be the address of this device")
	}
	if !gateway.Mask(mask).Equal(network) {
		return fmt.Errorf("the router is outside the subnet of the address")
	}

	return nil
}

func broadcastOf(network net.IP, mask net.IPMask) net.IP {
	broadcast := make(net.IP, len(network))
	for i := range network {
		broadcast[i] = network[i] | ^mask[i]
	}

	return broadcast
}

// liveEthernet reports what the interface carries now. During a trial that is
// not what readEthernetConfig returns.
func liveEthernet() proto.EthernetLive {
	live := proto.EthernetLive{Interface: ethInterface}

	name, gateway := getDefaultIPv4Route()
	if name == ethInterface {
		live.Gateway = gateway
	}

	iface, err := net.InterfaceByName(ethInterface)
	if err != nil {
		return live
	}

	address, prefix := ipv4AddressAndPrefix(*iface)
	live.Address = address
	live.Prefix = prefix
	if prefix > 0 {
		live.Netmask = net.IP(net.CIDRMask(prefix, 32)).String()
	}

	return live
}

func ipv4AddressAndPrefix(iface net.Interface) (string, int) {
	addrs, err := iface.Addrs()
	if err != nil {
		return "", 0
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}

		ip := ipNet.IP.To4()
		if ip == nil {
			continue
		}

		ones, _ := ipNet.Mask.Size()
		return ip.String(), ones
	}

	return "", 0
}
