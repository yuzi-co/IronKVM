package network

import (
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
)

// IPv6 on the wired interface.
//
// Every command here is `ip -6` or a write to the interface's IPv6 sysctls, so
// nothing in this file can touch the IPv4 address, the IPv4 routes or the DHCP
// client that the address trial in ethernet.go manages. A board whose IPv6
// change goes wrong still answers on its IPv4 address.
//
// The setting is saved only after it has been applied, and an apply that fails
// puts the saved setting back. The change is not a timed trial like the IPv4
// one, because it cannot cut the connection the request came in on, except
// when that connection is itself IPv6: see ipv6ChangeAllowed.
//
// The server applies the saved setting at boot (RestoreIPv6). S30eth is not
// touched: it ships in the application tarball too, which the fork cannot
// change on a board that runs Sipeed's firmware, and the kernel default
// (SLAAC) is what a board with no saved setting uses on both slots.

const (
	ipv6ModeSLAAC  = "slaac"
	ipv6ModeDHCPv6 = "dhcpv6"
	ipv6ModeStatic = "static"
	ipv6ModeOff    = "off"

	dhcpv6Client  = "udhcpc6"
	dhcpv6PidFile = "/run/udhcpc6.eth0.pid"

	minIPv6Prefix = 16
	maxIPv6Prefix = 126
)

// These are vars so the tests can point them at a scratch tree.
var (
	ipv6ConfigFile   = "/etc/kvm/ipv6"
	ipv6SysctlDir    = "/proc/sys/net/ipv6/conf"
	ipv6InetFile     = "/proc/net/if_inet6"
	ipv6RouteFile    = "/proc/net/ipv6_route"
	ipv6RestoredMark = "/run/ironkvm-ipv6-restored"
	lookPath         = exec.LookPath
	ipv6ApplyMutex   sync.Mutex
)

// ipv6Config is the saved setting. Address, Prefix and Gateway are kept across
// a switch away from static.
type ipv6Config struct {
	Mode    string `json:"mode"`
	Address string `json:"address,omitempty"`
	Prefix  int    `json:"prefix,omitempty"`
	Gateway string `json:"gateway,omitempty"`
}

func defaultIPv6Config() ipv6Config {
	return ipv6Config{Mode: ipv6ModeSLAAC}
}

// readIPv6Config returns the saved setting. A missing or unreadable file is the
// default, which is what the kernel does without anyone's help.
func readIPv6Config() ipv6Config {
	data, err := os.ReadFile(ipv6ConfigFile)
	if err != nil {
		return defaultIPv6Config()
	}

	var config ipv6Config
	if err := json.Unmarshal(data, &config); err != nil {
		return defaultIPv6Config()
	}

	switch config.Mode {
	case ipv6ModeSLAAC, ipv6ModeDHCPv6, ipv6ModeOff:
	case ipv6ModeStatic:
		// A static setting that no longer validates must not be applied.
		if err := validateIPv6Static(config); err != nil {
			return defaultIPv6Config()
		}
	default:
		return defaultIPv6Config()
	}

	return config
}

func writeIPv6Config(config ipv6Config) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(ipv6ConfigFile), 0o755); err != nil {
		return err
	}

	tmp := ipv6ConfigFile + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}

	return os.Rename(tmp, ipv6ConfigFile)
}

// validateIPv6Static rejects a setting that would leave the interface with an
// address nobody can reach or a route that goes nowhere.
func validateIPv6Static(config ipv6Config) error {
	address, err := netip.ParseAddr(config.Address)
	if err != nil || !address.Is6() || address.Is4In6() || address.Zone() != "" {
		return fmt.Errorf("the address is not an IPv6 address")
	}
	if address.IsLoopback() || address.IsUnspecified() || address.IsMulticast() {
		return fmt.Errorf("the address cannot be used on an interface")
	}
	if address.IsLinkLocalUnicast() {
		return fmt.Errorf("a link-local address (fe80::/10) is assigned automatically")
	}

	if config.Prefix < minIPv6Prefix || config.Prefix > maxIPv6Prefix {
		return fmt.Errorf("the prefix length must be between %d and %d", minIPv6Prefix, maxIPv6Prefix)
	}
	prefix, err := address.Prefix(config.Prefix)
	if err != nil {
		return fmt.Errorf("the prefix length does not fit the address")
	}
	if prefix.Addr() == address {
		return fmt.Errorf("the address is the network address of its subnet")
	}

	if config.Gateway == "" {
		return nil
	}

	gateway, err := netip.ParseAddr(config.Gateway)
	if err != nil || !gateway.Is6() || gateway.Is4In6() || gateway.Zone() != "" {
		return fmt.Errorf("the router is not an IPv6 address")
	}
	if gateway.IsLoopback() || gateway.IsUnspecified() || gateway.IsMulticast() {
		return fmt.Errorf("the router cannot be used as a router address")
	}
	if gateway == address {
		return fmt.Errorf("the router cannot be the address of this device")
	}
	// A link-local router is how most networks announce theirs. Any other
	// router has to be on the subnet, or the kernel cannot reach it.
	if !gateway.IsLinkLocalUnicast() && !prefix.Contains(gateway) {
		return fmt.Errorf("the router is outside the subnet of the address")
	}

	return nil
}

// ipv6ChangeAllowed refuses a change that the request's own connection may not
// survive. A client on IPv6 can only start DHCPv6 on top of SLAAC, which adds
// addresses and removes none. Everything else drops or replaces the addresses
// the client is talking to, and the person making it can use the IPv4 address.
func ipv6ChangeAllowed(client netip.Addr, current, next ipv6Config) bool {
	client = client.Unmap()
	if !client.IsValid() || client.Is4() || client.IsLoopback() {
		return true
	}
	if current == next {
		return true
	}

	return current.Mode == ipv6ModeSLAAC && next.Mode == ipv6ModeDHCPv6
}

func ipv6Supported() bool {
	_, err := os.Stat(ipv6InetFile)
	return err == nil
}

func dhcpv6Available() bool {
	_, err := lookPath(dhcpv6Client)
	return err == nil
}

// writeIPv6Sysctl sets one of eth0's IPv6 sysctls. A var for the tests.
var writeIPv6Sysctl = func(key string, value string) error {
	path := filepath.Join(ipv6SysctlDir, ethInterface, key)
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		return fmt.Errorf("failed to set %s: %w", key, err)
	}

	return nil
}

func readIPv6Sysctl(key string) string {
	data, err := os.ReadFile(filepath.Join(ipv6SysctlDir, ethInterface, key))
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

func stopDHCPv6() {
	data, err := os.ReadFile(dhcpv6PidFile)
	if err != nil {
		return
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err == nil && pid > 1 {
		if process, err := os.FindProcess(pid); err == nil {
			_ = process.Kill()
		}
	}
	_ = os.Remove(dhcpv6PidFile)
}

// applyIPv6 puts the interface in the mode. prev is the mode it is leaving,
// which decides whether the addresses are cleared and the link's IPv6
// restarted. It never touches IPv4.
func applyIPv6(prev string, config ipv6Config) error {
	if !ipv6Supported() {
		return fmt.Errorf("this kernel has no IPv6")
	}

	stopDHCPv6()

	switch config.Mode {
	case ipv6ModeOff:
		return writeIPv6Sysctl("disable_ipv6", "1")

	case ipv6ModeSLAAC, ipv6ModeDHCPv6:
		// SLAAC and DHCPv6 both keep router advertisements on: DHCPv6 does not
		// carry the default route.
		if config.Mode == ipv6ModeDHCPv6 && !dhcpv6Available() {
			return fmt.Errorf("this image has no DHCPv6 client (%s)", dhcpv6Client)
		}

		// From SLAAC to DHCPv6 the addresses the interface has are kept. From
		// anywhere else the old ones are removed and the link's IPv6 is
		// restarted, which sends a router solicitation. Waiting for the next
		// unsolicited advertisement takes minutes, and the interface was seen
		// to come back incomplete without the restart (see applyStatic).
		restart := !(prev == ipv6ModeSLAAC && config.Mode == ipv6ModeDHCPv6)
		if restart {
			if err := writeIPv6Sysctl("disable_ipv6", "1"); err != nil {
				return err
			}
		}
		for _, kv := range [][2]string{{"accept_ra", "1"}, {"autoconf", "1"}} {
			if err := writeIPv6Sysctl(kv[0], kv[1]); err != nil {
				return err
			}
		}
		if restart {
			if err := writeIPv6Sysctl("disable_ipv6", "0"); err != nil {
				return err
			}
		}

		if config.Mode == ipv6ModeDHCPv6 {
			return runCommand(dhcpv6Client, "-i", ethInterface, "-b", "-t", "5", "-T", "3", "-A", "20",
				"-p", dhcpv6PidFile)
		}

		return nil

	case ipv6ModeStatic:
		if err := validateIPv6Static(config); err != nil {
			return err
		}
		// No router advertisement may add an address or replace the router
		// that was set by hand.
		for _, kv := range [][2]string{{"accept_ra", "0"}, {"autoconf", "0"}, {"disable_ipv6", "0"}} {
			if err := writeIPv6Sysctl(kv[0], kv[1]); err != nil {
				return err
			}
		}
		if err := runCommand("ip", "-6", "addr", "flush", "dev", ethInterface, "scope", "global"); err != nil {
			return err
		}
		// Losing the old default route is not an error: there may be none.
		_ = runCommand("ip", "-6", "route", "del", "default", "dev", ethInterface)

		address := fmt.Sprintf("%s/%d", config.Address, config.Prefix)
		if err := runCommand("ip", "-6", "addr", "add", address, "dev", ethInterface); err != nil {
			return err
		}
		if config.Gateway == "" {
			return nil
		}

		return runCommand("ip", "-6", "route", "replace", "default", "via", config.Gateway, "dev", ethInterface)
	}

	return fmt.Errorf("unknown IPv6 mode %q", config.Mode)
}

// changeIPv6 applies the new setting and saves it. If the apply or the save
// fails, the saved setting is applied again, so a failed change leaves the
// interface where it was.
func changeIPv6(config ipv6Config) error {
	ipv6ApplyMutex.Lock()
	defer ipv6ApplyMutex.Unlock()

	previous := readIPv6Config()

	if err := applyIPv6(previous.Mode, config); err != nil {
		restoreIPv6(config.Mode, previous)
		return err
	}
	if err := writeIPv6Config(config); err != nil {
		restoreIPv6(config.Mode, previous)
		return fmt.Errorf("failed to save the setting: %w", err)
	}

	return nil
}

// restoreIPv6 is the way back from a failed change. If that fails too, the
// interface is left to the kernel's own behaviour (SLAAC), the one state that
// needs no help.
func restoreIPv6(from string, previous ipv6Config) {
	err := applyIPv6(from, previous)
	if err == nil {
		return
	}
	log.Errorf("failed to put the previous IPv6 setting back: %s", err)

	if err := applyIPv6(ipv6ModeOff, defaultIPv6Config()); err != nil {
		log.Errorf("failed to fall back to SLAAC: %s", err)
	}
}

// RestoreIPv6 applies the saved setting once per boot. SLAAC is what the
// kernel does with no help, so a board on the default is left alone.
func RestoreIPv6() {
	if _, err := os.Stat(ipv6RestoredMark); err == nil {
		return
	}
	_ = os.WriteFile(ipv6RestoredMark, nil, 0o644)

	config := readIPv6Config()
	if config.Mode == ipv6ModeSLAAC || !ipv6Supported() {
		return
	}

	ipv6ApplyMutex.Lock()
	defer ipv6ApplyMutex.Unlock()

	// A fresh boot is in the kernel's state, which is SLAAC.
	if err := applyIPv6(ipv6ModeSLAAC, config); err != nil {
		log.Errorf("failed to apply the saved IPv6 setting %q: %s", config.Mode, err)
		if err := applyIPv6(ipv6ModeOff, defaultIPv6Config()); err != nil {
			log.Errorf("failed to fall back to SLAAC: %s", err)
		}
		return
	}
	log.Infof("applied the saved IPv6 setting: %s", config.Mode)
}

func (s *Service) GetIPv6(c *gin.Context) {
	var rsp proto.Response

	saved := readIPv6Config()
	data := &proto.GetIPv6Rsp{
		Supported:       ipv6Supported(),
		Mode:            saved.Mode,
		Address:         saved.Address,
		Prefix:          saved.Prefix,
		Gateway:         saved.Gateway,
		DHCPv6Available: dhcpv6Available(),
		Client:          clientAddr(c).String(),
		Live:            liveIPv6(),
	}

	rsp.OkRspWithData(c, data)
}

func (s *Service) SetIPv6(c *gin.Context) {
	var req proto.SetIPv6Req
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	config := ipv6Config{Mode: req.Mode}
	saved := readIPv6Config()
	if req.Mode == ipv6ModeStatic {
		config.Address = strings.TrimSpace(req.Address)
		config.Prefix = req.Prefix
		config.Gateway = strings.TrimSpace(req.Gateway)
		if err := validateIPv6Static(config); err != nil {
			rsp.ErrRsp(c, -2, err.Error())
			return
		}
	} else {
		// Kept for the form, as the saved static settings are for IPv4.
		config.Address, config.Prefix, config.Gateway = saved.Address, saved.Prefix, saved.Gateway
	}

	if req.Mode == ipv6ModeDHCPv6 && !dhcpv6Available() {
		rsp.ErrRsp(c, -4, "this image has no DHCPv6 client")
		return
	}

	if !ipv6ChangeAllowed(clientAddr(c), saved, config) {
		rsp.ErrRsp(c, -5, "this page is open over IPv6: use the IPv4 address to change the IPv6 settings")
		return
	}

	if err := changeIPv6(config); err != nil {
		log.Errorf("failed to change the IPv6 setting: %s", err)
		rsp.ErrRsp(c, -3, err.Error())
		return
	}

	rsp.OkRsp(c)
	log.Infof("ipv6 setting changed: mode=%s", config.Mode)
}

func clientAddr(c *gin.Context) netip.Addr {
	ap, err := netip.ParseAddrPort(c.Request.RemoteAddr)
	if err != nil {
		return netip.Addr{}
	}

	return ap.Addr().Unmap()
}

// liveIPv6 reports what the interface carries now.
func liveIPv6() proto.IPv6Live {
	live := proto.IPv6Live{
		Interface: ethInterface,
		Disabled:  readIPv6Sysctl("disable_ipv6") == "1",
		Addresses: []proto.IPv6Address{},
	}

	if iface, err := net.InterfaceByName(ethInterface); err == nil {
		live.Addresses = ipv6Addresses(*iface)
	}
	if data, err := os.ReadFile(ipv6RouteFile); err == nil {
		live.Gateway = parseIPv6DefaultRoute(string(data), ethInterface)
	}

	return live
}

func ipv6Addresses(iface net.Interface) []proto.IPv6Address {
	out := []proto.IPv6Address{}

	addrs, err := iface.Addrs()
	if err != nil {
		return out
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.To4() != nil {
			continue
		}

		ones, _ := ipNet.Mask.Size()
		out = append(out, proto.IPv6Address{
			Address: ipNet.IP.String(),
			Prefix:  ones,
			Scope:   ipv6Scope(ipNet.IP),
		})
	}

	return out
}

func ipv6Scope(ip net.IP) string {
	switch {
	case ip.IsLinkLocalUnicast():
		return "link"
	case len(ip) == net.IPv6len && ip[0]&0xfe == 0xfc:
		return "ula"
	}

	return "global"
}

// parseIPv6DefaultRoute reads /proc/net/ipv6_route for the next hop of the
// default route on the device. Fields: destination, destination prefix length,
// source, source prefix length, next hop, metric, refcount, use, flags, device.
func parseIPv6DefaultRoute(content string, device string) string {
	zero := strings.Repeat("0", 32)

	for _, line := range strings.Split(content, "\n") {
		f := strings.Fields(line)
		if len(f) < 10 || f[9] != device || f[1] != "00" || f[0] != zero {
			continue
		}
		// RTF_UP (0x1) and RTF_GATEWAY (0x2).
		flags, err := strconv.ParseUint(f[8], 16, 32)
		if err != nil || flags&0x3 != 0x3 {
			continue
		}
		if len(f[4]) != 32 || f[4] == zero {
			continue
		}

		var b [16]byte
		valid := true
		for i := range b {
			v, err := strconv.ParseUint(f[4][i*2:i*2+2], 16, 8)
			if err != nil {
				valid = false
				break
			}
			b[i] = byte(v)
		}
		if valid {
			return net.IP(b[:]).String()
		}
	}

	return ""
}
