package wireguard

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// fakeSystem records the calls an apply makes and fails the ones it is told to.
type fakeSystem struct {
	calls  []string
	links  map[string]bool
	config *wgtypes.Config
	dev    *wgtypes.Device
	// fail maps a call's first word to the error it returns.
	fail map[string]error
	// exists lists routes that are there already.
	exists map[string]bool
}

func newFake() *fakeSystem {
	return &fakeSystem{links: map[string]bool{}, fail: map[string]error{}, exists: map[string]bool{}}
}

func (f *fakeSystem) record(format string, args ...any) error {
	call := fmt.Sprintf(format, args...)
	f.calls = append(f.calls, call)
	return f.fail[strings.Fields(call)[0]]
}

func (f *fakeSystem) linkExists(name string) (bool, error) { return f.links[name], nil }

func (f *fakeSystem) addLink(name string) error {
	if err := f.record("addLink %s", name); err != nil {
		return err
	}
	f.links[name] = true
	return nil
}

func (f *fakeSystem) delLink(name string) error {
	delete(f.links, name)
	return f.record("delLink %s", name)
}

func (f *fakeSystem) configure(name string, c wgtypes.Config) error {
	f.config = &c
	return f.record("configure %s", name)
}

func (f *fakeSystem) device(name string) (*wgtypes.Device, error) {
	if f.dev == nil || !f.links[name] {
		return nil, errors.New("no device")
	}
	return f.dev, nil
}

func (f *fakeSystem) setLink(name string, mtu int, up bool) error {
	return f.record("setLink %s mtu %d up %t", name, mtu, up)
}

func (f *fakeSystem) addAddr(name string, p netip.Prefix) error {
	return f.record("addAddr %s %s", name, p)
}

func (f *fakeSystem) addRoute(name string, p netip.Prefix, table int) error {
	if err := f.record("addRoute %s %s table %d", name, p, table); err != nil {
		return err
	}
	if f.exists[p.String()] {
		return syscall.EEXIST
	}
	return nil
}

func (f *fakeSystem) addRule(r rule) error { return f.record("addRule %s", r) }

func (f *fakeSystem) delRule(r rule) error {
	f.calls = append(f.calls, "delRule "+r.String())
	return syscall.ENOENT
}

func (f *fakeSystem) setSrcValidMark() error { return f.record("srcValidMark") }

func (f *fakeSystem) resolve(endpoint string) (*net.UDPAddr, error) {
	if err := f.record("resolve %s", endpoint); err != nil {
		return nil, err
	}
	return net.ResolveUDPAddr("udp", strings.Replace(endpoint, "vpn.example.com", "198.51.100.7", 1))
}

func (f *fakeSystem) has(call string) bool { return slices.Contains(f.calls, call) }

func TestUpConfiguresTheInterface(t *testing.T) {
	f := newFake()
	m := &Manager{sys: f, iface: "wg0"}

	c := validConfig()
	c.ListenPort = 51821
	c.Peers[0].Endpoint = "vpn.example.com:51820"
	c.Peers[0].PresharedKey = testPSK
	c.Peers[0].PersistentKeepalive = 25
	c.Peers[0].AllowedIPs = []string{"10.8.0.0/24", "192.168.50.0/24"}
	// The address's own subnet is routed by the kernel already.
	f.exists["10.8.0.0/24"] = true

	warnings, err := m.Up(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings %v", warnings)
	}

	for _, want := range []string{
		"addLink wg0",
		"configure wg0",
		"setLink wg0 mtu 1420 up true",
		"addAddr wg0 10.8.0.2/24",
		"addRoute wg0 10.8.0.0/24 table 0",
		"addRoute wg0 192.168.50.0/24 table 0",
		"resolve vpn.example.com:51820",
	} {
		if !f.has(want) {
			t.Errorf("no %q in %v", want, f.calls)
		}
	}
	for _, call := range f.calls {
		if strings.HasPrefix(call, "addRule") || call == "srcValidMark" {
			t.Errorf("%s without a default route", call)
		}
	}

	cfg := f.config
	if cfg.PrivateKey == nil || cfg.PrivateKey.String() != testPrivate {
		t.Error("private key not set")
	}
	if cfg.ListenPort == nil || *cfg.ListenPort != 51821 {
		t.Error("listen port not set")
	}
	if cfg.FirewallMark != nil {
		t.Error("a firewall mark without a default route")
	}
	if !cfg.ReplacePeers || len(cfg.Peers) != 1 {
		t.Fatalf("peers %+v", cfg.Peers)
	}
	p := cfg.Peers[0]
	if p.PublicKey.String() != testPeerKey || p.PresharedKey == nil || p.PresharedKey.String() != testPSK {
		t.Error("peer keys")
	}
	if p.Endpoint == nil || p.Endpoint.String() != "198.51.100.7:51820" {
		t.Errorf("endpoint %v", p.Endpoint)
	}
	if p.PersistentKeepaliveInterval == nil || *p.PersistentKeepaliveInterval != 25*time.Second {
		t.Error("keepalive")
	}
	if !p.ReplaceAllowedIPs || len(p.AllowedIPs) != 2 {
		t.Errorf("allowed IPs %v", p.AllowedIPs)
	}
}

func TestUpWarnsWhenARouteBelongsElsewhere(t *testing.T) {
	f := newFake()
	m := &Manager{sys: f, iface: "wg0"}
	c := validConfig()
	c.Peers[0].AllowedIPs = []string{"10.8.0.0/24", "192.168.1.0/24"}
	f.exists["10.8.0.0/24"] = true
	f.exists["192.168.1.0/24"] = true

	warnings, err := m.Up(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "192.168.1.0/24") {
		t.Errorf("warnings %v", warnings)
	}
	if got := m.Status(c).Warnings; len(got) != 1 {
		t.Errorf("status warnings %v", got)
	}
}

func TestUpDefaultRouteUsesItsOwnTable(t *testing.T) {
	f := newFake()
	m := &Manager{sys: f, iface: "wg0"}
	c := validConfig()
	c.Peers[0].AllowedIPs = []string{"0.0.0.0/0"}

	if _, err := m.Up(c); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"addRoute wg0 0.0.0.0/0 table 51820",
		"addRule " + rule{family: 4, priority: rulePriority, table: routeTable, fwmark: routeTable, invert: true}.String(),
		"addRule " + rule{family: 4, priority: rulePriority - 1, table: mainTable, suppressPrefix: true}.String(),
		"srcValidMark",
	} {
		if !f.has(want) {
			t.Errorf("no %q in %v", want, f.calls)
		}
	}
	if f.config.FirewallMark == nil || *f.config.FirewallMark != routeTable {
		t.Error("the tunnel's own packets are not marked")
	}
	for _, call := range f.calls {
		if strings.HasPrefix(call, "addRule") && strings.Contains(call, "family 6") {
			t.Errorf("an IPv6 rule without ::/0: %s", call)
		}
	}
}

func TestUpFailureTakesTheInterfaceDown(t *testing.T) {
	f := newFake()
	m := &Manager{sys: f, iface: "wg0"}
	f.fail["addAddr"] = errors.New("boom")

	if _, err := m.Up(validConfig()); err == nil {
		t.Fatal("no error")
	}
	if f.links["wg0"] {
		t.Error("the half-made interface is left")
	}
	if m.Status(validConfig()).Error == "" {
		t.Error("the status does not carry the error")
	}
}

func TestUpReplacesAnExistingInterface(t *testing.T) {
	f := newFake()
	f.links["wg0"] = true
	m := &Manager{sys: f, iface: "wg0"}
	if _, err := m.Up(validConfig()); err != nil {
		t.Fatal(err)
	}
	if !f.has("delLink wg0") || slices.Index(f.calls, "delLink wg0") > slices.Index(f.calls, "addLink wg0") {
		t.Errorf("calls %v", f.calls)
	}
}

func TestUpRefusesAnInvalidConfig(t *testing.T) {
	f := newFake()
	m := &Manager{sys: f, iface: "wg0"}
	c := validConfig()
	c.Peers = nil
	if _, err := m.Up(c); err == nil {
		t.Fatal("no error")
	}
	if len(f.calls) != 0 {
		t.Errorf("calls %v", f.calls)
	}
}

func TestDown(t *testing.T) {
	f := newFake()
	f.links["wg0"] = true
	m := &Manager{sys: f, iface: "wg0"}
	if err := m.Down(); err != nil {
		t.Fatal(err)
	}
	if f.links["wg0"] {
		t.Error("the interface is still there")
	}
	rules := 0
	for _, call := range f.calls {
		if strings.HasPrefix(call, "delRule") {
			rules++
		}
	}
	if rules != 4 {
		t.Errorf("%d rules removed, want both families' two", rules)
	}

	// Down again is not an error.
	if err := m.Down(); err != nil {
		t.Fatal(err)
	}
}

func TestStatus(t *testing.T) {
	f := newFake()
	m := &Manager{sys: f, iface: "wg0"}
	c := validConfig()

	if s := m.Status(c); s.Up {
		t.Error("up without an interface")
	}

	if _, err := m.Up(c); err != nil {
		t.Fatal(err)
	}
	priv, _ := wgtypes.ParseKey(testPrivate)
	peer, _ := wgtypes.ParseKey(testPeerKey)
	handshake := time.Unix(1_800_000_000, 0)
	f.dev = &wgtypes.Device{
		Name:       "wg0",
		PrivateKey: priv,
		PublicKey:  priv.PublicKey(),
		ListenPort: 40000,
		Peers: []wgtypes.Peer{{
			PublicKey:         peer,
			Endpoint:          &net.UDPAddr{IP: net.ParseIP("203.0.113.5"), Port: 51820},
			LastHandshakeTime: handshake,
			ReceiveBytes:      1234,
			TransmitBytes:     5678,
			AllowedIPs:        []net.IPNet{{IP: net.IPv4(10, 8, 0, 0).To4(), Mask: net.CIDRMask(24, 32)}},
		}},
	}

	s := m.Status(c)
	if !s.Up || s.ListenPort != 40000 || s.PublicKey != priv.PublicKey().String() {
		t.Errorf("status %+v", s)
	}
	if len(s.Peers) != 1 {
		t.Fatalf("peers %+v", s.Peers)
	}
	p := s.Peers[0]
	if p.PublicKey != testPeerKey || p.Endpoint != "203.0.113.5:51820" || p.LatestHandshake != handshake.Unix() ||
		p.RxBytes != 1234 || p.TxBytes != 5678 || p.AllowedIPs[0] != "10.8.0.0/24" {
		t.Errorf("peer %+v", p)
	}
}
