// Package allowlist restricts which client addresses may reach the board's
// management services: the web UI and its API (and with them WebRTC
// signalling, the websockets and Redfish, which share the HTTP listeners),
// VNC and IPMI.
//
// The list lives in server.yaml as security.allowedNetworks, CIDRs or single
// addresses. An empty list allows every address, which is the default and
// what a board without the setting does.
//
// Two kinds of client always get through, because they are the way back in
// from a list that turned out wrong:
//   - loopback, which the on-device runtime and the board's own scripts use;
//   - whatever Exempt allows, which the router sets to the host on the USB
//     network link, the cable that is already plugged into the machine the
//     board controls.
//
// The serial console is the other way back: delete the allowedNetworks lines
// from /etc/kvm/server.yaml and restart the server.
//
// The check is on the address the connection comes from, never on
// X-Forwarded-For, which the client writes. Behind a reverse proxy that means
// the proxy's address is the one checked.
package allowlist

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
)

// maxEntries bounds the list. Every request walks it, and an owner who needs
// more than this many ranges wants a firewall in front of the board.
const maxEntries = 64

// Exempt reports whether an address may connect whatever the list says. It is
// asked only about addresses the list refuses, so it may cost a file read.
// Nil exempts nothing beyond loopback.
var Exempt func(netip.Addr) bool

// current is the parsed list in force. Nil until first use.
var (
	current  atomic.Pointer[[]netip.Prefix]
	loadOnce sync.Once
)

// prefixes returns the list in force, reading server.yaml the first time.
func prefixes() []netip.Prefix {
	loadOnce.Do(func() {
		if current.Load() != nil {
			return
		}
		entries := config.GetInstance().Security.AllowedNetworks
		list, err := Parse(entries)
		if err != nil {
			// A list edited by hand into something unreadable is ignored
			// rather than read as "allow nothing": the alternative locks
			// everyone out, and the board has no remote way back in.
			log.Errorf("allowlist: ignoring security.allowedNetworks in server.yaml: %s", err)
			list = nil
		}
		current.Store(&list)
	})
	return *current.Load()
}

// set replaces the list in force.
func set(list []netip.Prefix) {
	loadOnce.Do(func() {})
	current.Store(&list)
}

// Parse reads a list of CIDRs and single addresses. A single address stands
// for itself alone, and a CIDR with host bits set is read as its network.
func Parse(entries []string) ([]netip.Prefix, error) {
	if len(entries) > maxEntries {
		return nil, fmt.Errorf("at most %d entries", maxEntries)
	}
	list := make([]netip.Prefix, 0, len(entries))
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		p, err := parseEntry(entry)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func parseEntry(entry string) (netip.Prefix, error) {
	if strings.Contains(entry, "/") {
		p, err := netip.ParsePrefix(entry)
		if err != nil {
			return netip.Prefix{}, fmt.Errorf("%q is not an address or a CIDR", entry)
		}
		// ::ffff:10.0.0.0/104 is 10.0.0.0/8 written the long way.
		if p.Addr().Is4In6() {
			if p.Bits() < 96 {
				return netip.Prefix{}, fmt.Errorf("%q is not an address or a CIDR", entry)
			}
			p = netip.PrefixFrom(p.Addr().Unmap(), p.Bits()-96)
		}
		return p.Masked(), nil
	}
	addr, err := netip.ParseAddr(entry)
	if err != nil || addr.Zone() != "" {
		return netip.Prefix{}, fmt.Errorf("%q is not an address or a CIDR", entry)
	}
	addr = addr.Unmap()
	return netip.PrefixFrom(addr, addr.BitLen()), nil
}

// format writes a list back the way an owner would: a range as a CIDR and a
// single address on its own.
func format(list []netip.Prefix) []string {
	out := make([]string, 0, len(list))
	seen := make(map[netip.Prefix]bool, len(list))
	for _, p := range list {
		if seen[p] {
			continue
		}
		seen[p] = true
		if p.IsSingleIP() {
			out = append(out, p.Addr().String())
		} else {
			out = append(out, p.String())
		}
	}
	return out
}

// permits reports whether list lets addr in.
func permits(list []netip.Prefix, addr netip.Addr) bool {
	if len(list) == 0 {
		return true
	}
	addr = addr.Unmap()
	if addr.IsLoopback() {
		return true
	}
	for _, p := range list {
		if p.Contains(addr) {
			return true
		}
	}
	return Exempt != nil && Exempt(addr)
}

// Allowed reports whether a client at addr may connect.
func Allowed(addr netip.Addr) bool {
	return permits(prefixes(), addr)
}

// AllowedAddr is Allowed for a connection's remote address, as the VNC and
// IPMI listeners have it. An address that is not IP, such as a unix socket,
// is allowed: it cannot come from the network.
func AllowedAddr(a net.Addr) bool {
	addr, ok := addrOf(a)
	return !ok || Allowed(addr)
}

func addrOf(a net.Addr) (netip.Addr, bool) {
	switch a := a.(type) {
	case *net.TCPAddr:
		addr, ok := netip.AddrFromSlice(a.IP)
		return addr, ok
	case *net.UDPAddr:
		addr, ok := netip.AddrFromSlice(a.IP)
		return addr, ok
	case nil:
		return netip.Addr{}, false
	}
	ap, err := netip.ParseAddrPort(a.String())
	if err != nil {
		return netip.Addr{}, false
	}
	return ap.Addr(), true
}

// remoteAddr is the address an HTTP request's connection comes from.
func remoteAddr(r *http.Request) (netip.Addr, bool) {
	ap, err := netip.ParseAddrPort(r.RemoteAddr)
	if err != nil {
		return netip.Addr{}, false
	}
	return ap.Addr().Unmap(), true
}

// errNotAllowed is the answer to a refused client.
var errNotAllowed = errors.New("this address is not allowed to reach the device")

// Middleware refuses requests from addresses the list does not allow, before
// anything else runs: the static UI, the login and every API route.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		addr, ok := remoteAddr(c.Request)
		if !ok || Allowed(addr) {
			c.Next()
			return
		}
		log.Debugf("allowlist: refused %s %s from %s", c.Request.Method, c.Request.URL.Path, addr)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code": -1,
			"msg":  errNotAllowed.Error(),
			"data": nil,
		})
	}
}
