package wireguard

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Routing, as wg-quick does it. Each allowed IP is a route through the
// interface in the main table. A peer given 0.0.0.0/0 or ::/0 would take the
// board's default route, and the tunnel's own packets with it, so that route
// goes into a table of its own instead, with two rules: packets the tunnel
// sends are marked and keep the main table; the rest look in the main table
// first, without its default route, and then in the tunnel's table. Routes to
// the local networks stay where they are, but anything that is not on them,
// such as a client on another subnet behind the router, is answered through
// the tunnel.
const (
	interfaceName = "wg0"
	// routeTable is both the table of the default route and the mark on
	// the tunnel's own packets, as in wg-quick.
	routeTable = 51820
	mainTable  = 254
	// rulePriority is the "not marked: tunnel table" rule; the "main table
	// without its default" rule sits just before it.
	rulePriority = 32761
)

// rule is a policy routing rule.
type rule struct {
	family         int // 4 or 6
	priority       int
	table          int
	fwmark         int
	invert         bool
	suppressPrefix bool
}

func (r rule) String() string {
	s := fmt.Sprintf("family %d priority %d table %d", r.family, r.priority, r.table)
	if r.fwmark != 0 {
		if r.invert {
			s += " not"
		}
		s += fmt.Sprintf(" fwmark %d", r.fwmark)
	}
	if r.suppressPrefix {
		s += " suppress_prefixlength 0"
	}
	return s
}

func defaultRouteRules(family int) []rule {
	return []rule{
		{family: family, priority: rulePriority - 1, table: mainTable, suppressPrefix: true},
		{family: family, priority: rulePriority, table: routeTable, fwmark: routeTable, invert: true},
	}
}

// system is what the manager does to the kernel: netlink in netlink_linux.go,
// a fake in the tests.
type system interface {
	linkExists(name string) (bool, error)
	addLink(name string) error
	delLink(name string) error
	configure(name string, c wgtypes.Config) error
	device(name string) (*wgtypes.Device, error)
	setLink(name string, mtu int, up bool) error
	addAddr(name string, p netip.Prefix) error
	// addRoute adds a route through the interface; table 0 is main. It
	// returns EEXIST when the destination is routed already.
	addRoute(name string, p netip.Prefix, table int) error
	addRule(r rule) error
	delRule(r rule) error
	// setSrcValidMark lets the reverse path filter see the mark, as
	// wg-quick does for a default route.
	setSrcValidMark() error
	resolve(endpoint string) (*net.UDPAddr, error)
}

// Manager brings the interface up and down and reports on it.
type Manager struct {
	mu    sync.Mutex
	sys   system
	iface string
	// From the last apply: the routes that were not added, and the error.
	warnings []string
	lastErr  string
}

// Status is what the page shows of the running tunnel. It has no secret.
type Status struct {
	Up         bool         `json:"up"`
	Interface  string       `json:"interface"`
	PublicKey  string       `json:"publicKey,omitempty"`
	ListenPort int          `json:"listenPort,omitempty"`
	Peers      []PeerStatus `json:"peers"`
	Warnings   []string     `json:"warnings,omitempty"`
	Error      string       `json:"error,omitempty"`
}

// PeerStatus is one peer as the kernel sees it.
type PeerStatus struct {
	PublicKey string `json:"publicKey"`
	Endpoint  string `json:"endpoint,omitempty"`
	// LatestHandshake is in Unix seconds, 0 for never.
	LatestHandshake int64    `json:"latestHandshake"`
	RxBytes         int64    `json:"rxBytes"`
	TxBytes         int64    `json:"txBytes"`
	AllowedIPs      []string `json:"allowedIPs"`
}

// Up replaces the interface with one made from c. On a failure nothing is
// left behind. The warnings name the routes that were not added.
func (m *Manager) Up(c Config) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// A configuration that is refused leaves a running tunnel alone.
	if err := c.Validate(); err != nil {
		return nil, err
	}
	warnings, err := m.up(c)
	m.warnings = warnings
	m.lastErr = ""
	if err != nil {
		m.lastErr = err.Error()
		m.down()
	}
	return warnings, err
}

func (m *Manager) up(c Config) ([]string, error) {
	devConf, err := m.deviceConfig(c)
	if err != nil {
		return nil, err
	}
	if err := m.down(); err != nil {
		return nil, err
	}

	if err := m.sys.addLink(m.iface); err != nil {
		return nil, fmt.Errorf("creating %s: %w", m.iface, err)
	}
	if err := m.sys.configure(m.iface, devConf); err != nil {
		return nil, fmt.Errorf("configuring %s: %w", m.iface, err)
	}

	mtu := c.MTU
	if mtu == 0 {
		mtu = defaultMTU
	}
	var own []netip.Prefix
	for _, a := range c.Addresses {
		prefix, _ := parseAddress(a)
		own = append(own, prefix.Masked())
		if err := m.sys.addAddr(m.iface, prefix); err != nil && !errors.Is(err, syscall.EEXIST) {
			return nil, fmt.Errorf("adding address %s: %w", prefix, err)
		}
	}
	if err := m.sys.setLink(m.iface, mtu, true); err != nil {
		return nil, fmt.Errorf("bringing %s up: %w", m.iface, err)
	}

	var warnings []string
	families := map[int]bool{}
	for _, p := range c.Peers {
		for _, a := range p.AllowedIPs {
			prefix, _ := netip.ParsePrefix(a)
			table := 0
			if prefix.Bits() == 0 {
				table = routeTable
				families[family(prefix)] = true
			}
			err := m.sys.addRoute(m.iface, prefix, table)
			switch {
			case err == nil:
			case errors.Is(err, syscall.EEXIST):
				// The kernel routes an address's own subnet
				// through the interface already.
				if !contains(own, prefix) {
					warnings = append(warnings, fmt.Sprintf("%s is routed elsewhere already, so it does not go through the tunnel", prefix))
				}
			default:
				return nil, fmt.Errorf("adding the route to %s: %w", prefix, err)
			}
		}
	}

	if len(families) > 0 {
		if err := m.sys.setSrcValidMark(); err != nil {
			log.Warnf("wireguard: src_valid_mark: %s", err)
		}
		for _, f := range []int{4, 6} {
			if !families[f] {
				continue
			}
			for _, r := range defaultRouteRules(f) {
				if err := m.sys.addRule(r); err != nil && !errors.Is(err, syscall.EEXIST) {
					return nil, fmt.Errorf("adding the IPv%d routing rule: %w", f, err)
				}
			}
		}
	}
	for _, w := range warnings {
		log.Warnf("wireguard: %s", w)
	}
	log.Infof("wireguard: %s up with %d peers", m.iface, len(c.Peers))
	return warnings, nil
}

func (m *Manager) deviceConfig(c Config) (wgtypes.Config, error) {
	private, _ := parseKey("the private key", c.PrivateKey)
	dc := wgtypes.Config{PrivateKey: &private, ReplacePeers: true}
	if c.ListenPort != 0 {
		port := c.ListenPort
		dc.ListenPort = &port
	}
	if c.DefaultRoute() {
		mark := routeTable
		dc.FirewallMark = &mark
	}
	for _, p := range c.Peers {
		key, _ := parseKey("", p.PublicKey)
		pc := wgtypes.PeerConfig{PublicKey: key, ReplaceAllowedIPs: true}
		if p.PresharedKey != "" {
			psk, _ := parseKey("", p.PresharedKey)
			pc.PresharedKey = &psk
		}
		if p.Endpoint != "" {
			addr, err := m.sys.resolve(p.Endpoint)
			if err != nil {
				return dc, fmt.Errorf("resolving %s: %w", p.Endpoint, err)
			}
			pc.Endpoint = addr
		}
		if p.PersistentKeepalive > 0 {
			interval := time.Duration(p.PersistentKeepalive) * time.Second
			pc.PersistentKeepaliveInterval = &interval
		}
		for _, a := range p.AllowedIPs {
			prefix, _ := netip.ParsePrefix(a)
			pc.AllowedIPs = append(pc.AllowedIPs, net.IPNet{
				IP:   prefix.Addr().AsSlice(),
				Mask: net.CIDRMask(prefix.Bits(), prefix.Addr().BitLen()),
			})
		}
		dc.Peers = append(dc.Peers, pc)
	}
	return dc, nil
}

// Down removes the interface, its routes with it, and the rules.
func (m *Manager) Down() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warnings = nil
	m.lastErr = ""
	return m.down()
}

func (m *Manager) down() error {
	for _, f := range []int{4, 6} {
		for _, r := range defaultRouteRules(f) {
			// Missing rules are the usual case.
			_ = m.sys.delRule(r)
		}
	}
	exists, err := m.sys.linkExists(m.iface)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	if err := m.sys.delLink(m.iface); err != nil {
		return fmt.Errorf("removing %s: %w", m.iface, err)
	}
	return nil
}

// IsUp reports whether the interface exists.
func (m *Manager) IsUp() bool {
	exists, err := m.sys.linkExists(m.iface)
	return err == nil && exists
}

// SetError records an error from outside an apply, such as boot giving up.
func (m *Manager) SetError(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastErr = msg
}

// Status reads the interface. c gives the peers' names when the kernel has
// none: a configured peer the interface lacks is still listed.
func (m *Manager) Status(c Config) Status {
	m.mu.Lock()
	defer m.mu.Unlock()

	s := Status{Interface: m.iface, Peers: []PeerStatus{}, Warnings: m.warnings, Error: m.lastErr}
	exists, err := m.sys.linkExists(m.iface)
	if err != nil || !exists {
		return s
	}
	s.Up = true
	dev, err := m.sys.device(m.iface)
	if err != nil {
		s.Error = fmt.Sprintf("reading %s: %s", m.iface, err)
		return s
	}
	s.PublicKey = dev.PublicKey.String()
	s.ListenPort = dev.ListenPort
	for _, p := range dev.Peers {
		ps := PeerStatus{
			PublicKey:  p.PublicKey.String(),
			RxBytes:    p.ReceiveBytes,
			TxBytes:    p.TransmitBytes,
			AllowedIPs: []string{},
		}
		if p.Endpoint != nil {
			ps.Endpoint = p.Endpoint.String()
		}
		if !p.LastHandshakeTime.IsZero() {
			ps.LatestHandshake = p.LastHandshakeTime.Unix()
		}
		for _, n := range p.AllowedIPs {
			if prefix, ok := ipNetPrefix(n); ok {
				ps.AllowedIPs = append(ps.AllowedIPs, prefix.String())
			}
		}
		s.Peers = append(s.Peers, ps)
	}
	return s
}

func ipNetPrefix(n net.IPNet) (netip.Prefix, bool) {
	addr, ok := netip.AddrFromSlice(n.IP)
	if !ok {
		return netip.Prefix{}, false
	}
	ones, _ := n.Mask.Size()
	return netip.PrefixFrom(addr.Unmap(), ones), true
}

func family(p netip.Prefix) int {
	if p.Addr().Is4() {
		return 4
	}
	return 6
}

func contains(list []netip.Prefix, p netip.Prefix) bool {
	for _, q := range list {
		if q == p {
			return true
		}
	}
	return false
}
