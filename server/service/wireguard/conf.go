// Package wireguard is the board's built-in WireGuard tunnel: one interface,
// wg0, configured from Go over netlink (no wg or wg-quick on the board), its
// settings in /etc/kvm/wireguard.json, and the settings page's API under
// /api/network/wireguard.
//
// The private key and the preshared keys never leave the board: the API
// answers with the public key and whether a preshared key is set, and the
// settings export (service/settingsio) has no WireGuard section at all, since
// the key and the tunnel address belong to this board.
package wireguard

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	maxConfSize = 64 << 10
	maxPeers    = 32
	// defaultMTU is wg-quick's for an Ethernet link of 1500.
	defaultMTU = 1420
	minMTU     = 1280
	maxMTU     = 1500
	// maxKeepalive is the largest interval WireGuard takes (a 16-bit field).
	maxKeepalive = 65535
)

// Peer is one [Peer] section.
type Peer struct {
	PublicKey           string   `json:"publicKey"`
	PresharedKey        string   `json:"presharedKey,omitempty"`
	Endpoint            string   `json:"endpoint,omitempty"`
	AllowedIPs          []string `json:"allowedIPs"`
	PersistentKeepalive int      `json:"persistentKeepalive,omitempty"`
}

// Config is the tunnel as stored on the board.
type Config struct {
	PrivateKey string   `json:"privateKey"`
	Addresses  []string `json:"addresses"`
	// DNS is kept so the page can show it, but is not applied: see dnsNote.
	DNS        []string `json:"dns,omitempty"`
	ListenPort int      `json:"listenPort,omitempty"`
	MTU        int      `json:"mtu,omitempty"`
	Peers      []Peer   `json:"peers"`
	// StartAtBoot brings the tunnel up when the server starts.
	StartAtBoot bool `json:"startAtBoot,omitempty"`
}

// Configured reports whether there is a tunnel to bring up.
func (c Config) Configured() bool {
	return c.PrivateKey != "" && len(c.Addresses) > 0 && len(c.Peers) > 0
}

// DefaultRoute reports whether a peer is given every address of a family,
// which sends the board's own traffic into the tunnel.
func (c Config) DefaultRoute() bool {
	for _, p := range c.Peers {
		for _, a := range p.AllowedIPs {
			if prefix, err := netip.ParsePrefix(a); err == nil && prefix.Bits() == 0 {
				return true
			}
		}
	}
	return false
}

// parseKey decodes a WireGuard key: base64 of 32 bytes.
func parseKey(what, s string) (wgtypes.Key, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(s))
	if err != nil || len(raw) != wgtypes.KeyLen {
		return wgtypes.Key{}, fmt.Errorf("%s is not a WireGuard key (base64 of 32 bytes)", what)
	}
	var key wgtypes.Key
	copy(key[:], raw)
	return key, nil
}

// PublicKey returns the public key of a private key.
func PublicKey(private string) (string, error) {
	key, err := parseKey("the private key", private)
	if err != nil {
		return "", err
	}
	return key.PublicKey().String(), nil
}

// GeneratePrivateKey returns a new private key, made on the board.
func GeneratePrivateKey() (string, error) {
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return "", err
	}
	return key.String(), nil
}

// Validate checks c and normalises it in place: addresses and allowed IPs
// in canonical form, the allowed IPs masked as WireGuard keeps them.
func (c *Config) Validate() error {
	private, err := parseKey("the private key", c.PrivateKey)
	if err != nil {
		return err
	}
	if private == (wgtypes.Key{}) {
		return errors.New("the private key is all zeros")
	}
	c.PrivateKey = private.String()

	if len(c.Addresses) == 0 {
		return errors.New("the interface needs an address")
	}
	for i, a := range c.Addresses {
		prefix, err := parseAddress(a)
		if err != nil {
			return err
		}
		c.Addresses[i] = prefix.String()
	}
	for _, d := range c.DNS {
		if !validDNS(d) {
			return fmt.Errorf("DNS %q is neither an IP address nor a domain", d)
		}
	}
	if c.ListenPort < 0 || c.ListenPort > 65535 {
		return errors.New("the listen port is 0 to 65535")
	}
	if c.MTU != 0 && (c.MTU < minMTU || c.MTU > maxMTU) {
		return fmt.Errorf("the MTU is %d to %d", minMTU, maxMTU)
	}

	if len(c.Peers) == 0 {
		return errors.New("the tunnel needs a peer")
	}
	if len(c.Peers) > maxPeers {
		return fmt.Errorf("at most %d peers", maxPeers)
	}
	own := private.PublicKey()
	seen := map[wgtypes.Key]bool{}
	var routes []netip.Prefix
	for i := range c.Peers {
		p := &c.Peers[i]
		key, err := parseKey(fmt.Sprintf("peer %d's public key", i+1), p.PublicKey)
		if err != nil {
			return err
		}
		if key == own {
			return fmt.Errorf("peer %d is this board's own key", i+1)
		}
		if seen[key] {
			return fmt.Errorf("peer %d is listed twice", i+1)
		}
		seen[key] = true
		p.PublicKey = key.String()

		if p.PresharedKey != "" {
			psk, err := parseKey(fmt.Sprintf("peer %d's preshared key", i+1), p.PresharedKey)
			if err != nil {
				return err
			}
			p.PresharedKey = psk.String()
		}
		if p.Endpoint != "" {
			endpoint, err := normaliseEndpoint(p.Endpoint)
			if err != nil {
				return fmt.Errorf("peer %d: %w", i+1, err)
			}
			p.Endpoint = endpoint
		}
		if p.PersistentKeepalive < 0 || p.PersistentKeepalive > maxKeepalive {
			return fmt.Errorf("peer %d: the keepalive is 0 to %d seconds", i+1, maxKeepalive)
		}
		if len(p.AllowedIPs) == 0 {
			return fmt.Errorf("peer %d needs allowed IPs", i+1)
		}
		for j, a := range p.AllowedIPs {
			prefix, err := netip.ParsePrefix(strings.TrimSpace(a))
			if err != nil {
				return fmt.Errorf("peer %d: allowed IP %q is not an address with a prefix length", i+1, a)
			}
			prefix = prefix.Masked()
			for _, r := range routes {
				if r == prefix {
					return fmt.Errorf("peer %d: %s is given to another peer too", i+1, prefix)
				}
			}
			routes = append(routes, prefix)
			p.AllowedIPs[j] = prefix.String()
		}
	}
	return nil
}

// parseAddress reads an interface address: a prefix, or a bare address that
// is a host route, as `ip address add` takes it.
func parseAddress(s string) (netip.Prefix, error) {
	s = strings.TrimSpace(s)
	if prefix, err := netip.ParsePrefix(s); err == nil {
		return prefix, nil
	}
	if addr, err := netip.ParseAddr(s); err == nil && addr.Zone() == "" {
		return netip.PrefixFrom(addr, addr.BitLen()), nil
	}
	return netip.Prefix{}, fmt.Errorf("address %q is not an IP address", s)
}

// normaliseEndpoint checks host:port, where the host is an IP address
// (IPv6 in brackets) or a DNS name.
func normaliseEndpoint(s string) (string, error) {
	host, port, err := net.SplitHostPort(strings.TrimSpace(s))
	if err != nil {
		return "", fmt.Errorf("endpoint %q is not host:port", s)
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return "", fmt.Errorf("endpoint %q: the port is 1 to 65535", s)
	}
	if addr, err := netip.ParseAddr(host); err == nil {
		if addr.Zone() != "" {
			return "", fmt.Errorf("endpoint %q: a zone is not supported", s)
		}
		return netip.AddrPortFrom(addr, uint16(n)).String(), nil
	}
	if !validHostname(host) {
		return "", fmt.Errorf("endpoint %q: the host is not an IP address or a DNS name", s)
	}
	return net.JoinHostPort(strings.ToLower(host), port), nil
}

func validHostname(h string) bool {
	h = strings.TrimSuffix(h, ".")
	if h == "" || len(h) > 253 {
		return false
	}
	for _, label := range strings.Split(h, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, r := range label {
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-') {
				return false
			}
		}
	}
	return true
}

// validDNS takes what wg-quick's DNS takes: a server address or a search
// domain.
func validDNS(s string) bool {
	if _, err := netip.ParseAddr(s); err == nil {
		return true
	}
	// A dotted string of digits is a broken address, not a domain.
	if strings.Trim(s, "0123456789.") == "" {
		return false
	}
	return validHostname(s)
}

// ignoredKeys are wg-quick keys the board does not act on. The hooks are
// shell commands, which an imported file must not run as root; Table and
// SaveConfig are about wg-quick's own routing and saving.
var ignoredKeys = map[string]bool{
	"preup": true, "postup": true, "predown": true, "postdown": true,
	"table": true, "saveconfig": true, "fwmark": true,
}

// ParseConf reads a wg-quick configuration file. The warnings name the keys
// it ignores.
func ParseConf(text string) (Config, []string, error) {
	var c Config
	var warnings []string
	if len(text) > maxConfSize {
		return c, nil, fmt.Errorf("the file is larger than %d KiB", maxConfSize>>10)
	}

	section := ""
	sawInterface := false
	seen := map[string]bool{}
	var peer *Peer

	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Buffer(make([]byte, 0, 4096), maxConfSize+1)
	for n := 1; scanner.Scan(); n++ {
		line := scanner.Text()
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			seen = map[string]bool{}
			switch section {
			case "interface":
				if sawInterface {
					return c, nil, fmt.Errorf("line %d: a second [Interface]", n)
				}
				sawInterface = true
			case "peer":
				if len(c.Peers) >= maxPeers {
					return c, nil, fmt.Errorf("at most %d peers", maxPeers)
				}
				c.Peers = append(c.Peers, Peer{})
				peer = &c.Peers[len(c.Peers)-1]
			default:
				return c, nil, fmt.Errorf("line %d: unknown section [%s]", n, line[1:len(line)-1])
			}
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return c, nil, fmt.Errorf("line %d: not key = value", n)
		}
		name := strings.TrimSpace(key)
		key = strings.ToLower(name)
		value = strings.TrimSpace(value)
		if section == "" {
			return c, nil, fmt.Errorf("line %d: %s before any section", n, name)
		}

		list := key == "address" || key == "dns" || key == "allowedips"
		if !list && seen[key] {
			return c, nil, fmt.Errorf("line %d: %s given twice", n, name)
		}
		seen[key] = true

		var err error
		switch section + "." + key {
		case "interface.privatekey":
			c.PrivateKey = value
		case "interface.address":
			c.Addresses = append(c.Addresses, splitList(value)...)
		case "interface.dns":
			c.DNS = append(c.DNS, splitList(value)...)
		case "interface.listenport":
			c.ListenPort, err = parseInt(name, value, 0, 65535)
		case "interface.mtu":
			c.MTU, err = parseInt(name, value, minMTU, maxMTU)
		case "peer.publickey":
			peer.PublicKey = value
		case "peer.presharedkey":
			peer.PresharedKey = value
		case "peer.endpoint":
			peer.Endpoint = value
		case "peer.allowedips":
			peer.AllowedIPs = append(peer.AllowedIPs, splitList(value)...)
		case "peer.persistentkeepalive":
			if strings.EqualFold(value, "off") {
				peer.PersistentKeepalive = 0
			} else {
				peer.PersistentKeepalive, err = parseInt(name, value, 0, maxKeepalive)
			}
		default:
			if section == "interface" && ignoredKeys[key] {
				warnings = append(warnings, fmt.Sprintf("%s is ignored", name))
				continue
			}
			return c, nil, fmt.Errorf("line %d: unknown key %s in [%s]", n, name, section)
		}
		if err != nil {
			return c, nil, fmt.Errorf("line %d: %w", n, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return c, nil, err
	}
	if !sawInterface {
		return c, nil, errors.New("there is no [Interface] section")
	}
	if err := c.Validate(); err != nil {
		return c, nil, err
	}
	return c, warnings, nil
}

func splitList(s string) []string {
	var out []string
	for _, item := range strings.Split(s, ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}

func parseInt(name, value string, min, max int) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil || n < min || n > max {
		return 0, fmt.Errorf("%s is %d to %d", name, min, max)
	}
	return n, nil
}
