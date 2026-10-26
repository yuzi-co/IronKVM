//go:build linux

package wireguard

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"time"

	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// netlinkSystem makes the interface with rtnetlink, as `ip` would, and
// configures it with WireGuard's generic netlink family (wgctrl), as `wg`
// would. The kernel loads the wireguard module on the first RTM_NEWLINK of
// that kind where it is a module (slot B).
type netlinkSystem struct{}

func newSystem() system { return netlinkSystem{} }

// fib_rule_hdr flag, from linux/fib_rules.h.
const fibRuleInvert = 0x2

const srcValidMarkFile = "/proc/sys/net/ipv4/conf/all/src_valid_mark"

func execute(typ netlink.HeaderType, flags netlink.HeaderFlags, data []byte) error {
	conn, err := netlink.Dial(unix.NETLINK_ROUTE, nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Execute(netlink.Message{
		Header: netlink.Header{Type: typ, Flags: netlink.Request | netlink.Acknowledge | flags},
		Data:   data,
	})
	return err
}

// ifinfomsg: family, pad, type, index, flags, change.
func ifinfomsg(index int, flags, change uint32) []byte {
	b := make([]byte, unix.SizeofIfInfomsg)
	b[0] = unix.AF_UNSPEC
	binary.NativeEndian.PutUint32(b[4:], uint32(index))
	binary.NativeEndian.PutUint32(b[8:], flags)
	binary.NativeEndian.PutUint32(b[12:], change)
	return b
}

func encode(header []byte, fn func(ae *netlink.AttributeEncoder)) ([]byte, error) {
	ae := netlink.NewAttributeEncoder()
	fn(ae)
	attrs, err := ae.Encode()
	if err != nil {
		return nil, err
	}
	return append(header, attrs...), nil
}

func index(name string) (int, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return 0, err
	}
	return iface.Index, nil
}

func (netlinkSystem) linkExists(name string) (bool, error) {
	_, err := net.InterfaceByName(name)
	if err == nil {
		return true, nil
	}
	// The standard library exports no sentinel for a missing interface.
	var opErr *net.OpError
	if errors.As(err, &opErr) && opErr.Err != nil && opErr.Err.Error() == "no such network interface" {
		return false, nil
	}
	return false, err
}

func (netlinkSystem) addLink(name string) error {
	data, err := encode(ifinfomsg(0, 0, 0), func(ae *netlink.AttributeEncoder) {
		ae.String(unix.IFLA_IFNAME, name)
		ae.Nested(unix.IFLA_LINKINFO, func(nae *netlink.AttributeEncoder) error {
			nae.String(unix.IFLA_INFO_KIND, "wireguard")
			return nil
		})
	})
	if err != nil {
		return err
	}
	return execute(unix.RTM_NEWLINK, netlink.Create|netlink.Excl, data)
}

func (netlinkSystem) delLink(name string) error {
	idx, err := index(name)
	if err != nil {
		return err
	}
	return execute(unix.RTM_DELLINK, 0, ifinfomsg(idx, 0, 0))
}

func (netlinkSystem) setLink(name string, mtu int, up bool) error {
	idx, err := index(name)
	if err != nil {
		return err
	}
	var flags uint32
	if up {
		flags = unix.IFF_UP
	}
	data, err := encode(ifinfomsg(idx, flags, unix.IFF_UP), func(ae *netlink.AttributeEncoder) {
		ae.Uint32(unix.IFLA_MTU, uint32(mtu))
	})
	if err != nil {
		return err
	}
	return execute(unix.RTM_NEWLINK, 0, data)
}

func afOf(addr netip.Addr) byte {
	if addr.Is4() {
		return unix.AF_INET
	}
	return unix.AF_INET6
}

func (netlinkSystem) addAddr(name string, p netip.Prefix) error {
	idx, err := index(name)
	if err != nil {
		return err
	}
	// ifaddrmsg: family, prefixlen, flags, scope, index.
	h := make([]byte, unix.SizeofIfAddrmsg)
	h[0] = afOf(p.Addr())
	h[1] = byte(p.Bits())
	binary.NativeEndian.PutUint32(h[4:], uint32(idx))
	data, err := encode(h, func(ae *netlink.AttributeEncoder) {
		ae.Bytes(unix.IFA_LOCAL, p.Addr().AsSlice())
		ae.Bytes(unix.IFA_ADDRESS, p.Addr().AsSlice())
	})
	if err != nil {
		return err
	}
	return execute(unix.RTM_NEWADDR, netlink.Create|netlink.Excl, data)
}

func (netlinkSystem) addRoute(name string, p netip.Prefix, table int) error {
	idx, err := index(name)
	if err != nil {
		return err
	}
	if table == 0 {
		table = mainTable
	}
	// rtmsg: family, dst_len, src_len, tos, table, protocol, scope, type, flags.
	h := make([]byte, unix.SizeofRtMsg)
	h[0] = afOf(p.Addr())
	h[1] = byte(p.Bits())
	h[4] = unix.RT_TABLE_UNSPEC
	h[5] = unix.RTPROT_BOOT
	h[6] = unix.RT_SCOPE_LINK
	h[7] = unix.RTN_UNICAST
	data, err := encode(h, func(ae *netlink.AttributeEncoder) {
		if p.Bits() > 0 {
			ae.Bytes(unix.RTA_DST, p.Masked().Addr().AsSlice())
		}
		ae.Uint32(unix.RTA_OIF, uint32(idx))
		ae.Uint32(unix.RTA_TABLE, uint32(table))
	})
	if err != nil {
		return err
	}
	return execute(unix.RTM_NEWROUTE, netlink.Create|netlink.Excl, data)
}

func ruleMessage(r rule) ([]byte, error) {
	// fib_rule_hdr: family, dst_len, src_len, tos, table, res1, res2,
	// action, flags.
	h := make([]byte, 12)
	h[0] = unix.AF_INET
	if r.family == 6 {
		h[0] = unix.AF_INET6
	}
	h[4] = unix.RT_TABLE_UNSPEC
	h[7] = unix.FR_ACT_TO_TBL
	if r.invert {
		binary.NativeEndian.PutUint32(h[8:], fibRuleInvert)
	}
	return encode(h, func(ae *netlink.AttributeEncoder) {
		ae.Uint32(unix.FRA_PRIORITY, uint32(r.priority))
		ae.Uint32(unix.FRA_TABLE, uint32(r.table))
		if r.fwmark != 0 {
			ae.Uint32(unix.FRA_FWMARK, uint32(r.fwmark))
		}
		if r.suppressPrefix {
			ae.Uint32(unix.FRA_SUPPRESS_PREFIXLEN, 0)
		}
	})
}

func (netlinkSystem) addRule(r rule) error {
	data, err := ruleMessage(r)
	if err != nil {
		return err
	}
	return execute(unix.RTM_NEWRULE, netlink.Create|netlink.Excl, data)
}

func (netlinkSystem) delRule(r rule) error {
	data, err := ruleMessage(r)
	if err != nil {
		return err
	}
	return execute(unix.RTM_DELRULE, 0, data)
}

func (netlinkSystem) setSrcValidMark() error {
	return os.WriteFile(srcValidMarkFile, []byte("1\n"), 0o644)
}

func (netlinkSystem) configure(name string, c wgtypes.Config) error {
	client, err := wgctrl.New()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ConfigureDevice(name, c)
}

func (netlinkSystem) device(name string) (*wgtypes.Device, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.Device(name)
}

func (netlinkSystem) resolve(endpoint string) (*net.UDPAddr, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	host, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return nil, err
	}
	addrs, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("%s has no address", host)
	}
	// An IPv4 address first, as most boards have no IPv6 route out.
	pick := addrs[0]
	for _, a := range addrs {
		if a.Unmap().Is4() {
			pick = a
			break
		}
	}
	return net.ResolveUDPAddr("udp", net.JoinHostPort(pick.Unmap().String(), port))
}
