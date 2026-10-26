//go:build linux

package wireguard

import (
	"net"
	"os"
	"strings"
	"testing"
)

// TestNetlinkInterface makes a real interface. It needs CAP_NET_ADMIN and a
// kernel with WireGuard, so it runs only when asked, in a throwaway network
// namespace:
//
//	docker run --rm --cap-add NET_ADMIN -e WG_NETLINK_TEST=1 ... go test -run Netlink ./service/wireguard/
func TestNetlinkInterface(t *testing.T) {
	if os.Getenv("WG_NETLINK_TEST") != "1" {
		t.Skip("set WG_NETLINK_TEST=1, with CAP_NET_ADMIN, to make a real interface")
	}
	m := &Manager{sys: netlinkSystem{}, iface: "wgtest0"}
	t.Cleanup(func() { _ = m.Down() })

	c := validConfig()
	c.ListenPort = 51999
	c.Peers[0].PersistentKeepalive = 25
	c.Peers[0].AllowedIPs = []string{"10.8.0.0/24", "192.0.2.0/24"}
	if _, err := m.Up(c); err != nil {
		t.Fatal(err)
	}

	s := m.Status(c)
	if !s.Up || s.ListenPort != 51999 || s.PublicKey != mustPublic(t, testPrivate) || len(s.Peers) != 1 {
		t.Fatalf("status %+v", s)
	}
	if s.Peers[0].Endpoint != "203.0.113.5:51820" || len(s.Peers[0].AllowedIPs) != 2 {
		t.Errorf("peer %+v", s.Peers[0])
	}

	iface, err := net.InterfaceByName("wgtest0")
	if err != nil {
		t.Fatal(err)
	}
	if iface.MTU != defaultMTU || iface.Flags&net.FlagUp == 0 {
		t.Errorf("mtu %d flags %s", iface.MTU, iface.Flags)
	}
	addrs, _ := iface.Addrs()
	if len(addrs) != 1 || addrs[0].String() != "10.8.0.2/24" {
		t.Errorf("addresses %v", addrs)
	}
	routes, _ := os.ReadFile("/proc/net/route")
	// 192.0.2.0 as /proc/net/route spells it: little-endian hex.
	if !strings.Contains(string(routes), "wgtest0\t000200C0") {
		t.Errorf("no route to 192.0.2.0/24:\n%s", routes)
	}

	// The default route: its own table, the rules, the mark.
	c.Peers[0].AllowedIPs = []string{"0.0.0.0/0", "::/0"}
	if _, err := m.Up(c); err != nil {
		t.Fatal(err)
	}
	dev, err := m.sys.device("wgtest0")
	if err != nil {
		t.Fatal(err)
	}
	if dev.FirewallMark != routeTable {
		t.Errorf("fwmark %d", dev.FirewallMark)
	}
	// The main table keeps its default route.
	routes, _ = os.ReadFile("/proc/net/route")
	if strings.Contains(string(routes), "wgtest0\t00000000") {
		t.Errorf("the default route went into the main table:\n%s", routes)
	}
	for _, r := range append(defaultRouteRules(4), defaultRouteRules(6)...) {
		if err := m.sys.addRule(r); err == nil {
			t.Errorf("rule %s was not there", r)
		}
	}

	if err := m.Down(); err != nil {
		t.Fatal(err)
	}
	if _, err := net.InterfaceByName("wgtest0"); err == nil {
		t.Error("the interface is still there")
	}
	for _, r := range append(defaultRouteRules(4), defaultRouteRules(6)...) {
		if err := m.sys.delRule(r); err == nil {
			t.Errorf("rule %s was left behind", r)
		}
	}
}
