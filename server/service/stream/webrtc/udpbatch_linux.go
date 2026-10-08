//go:build linux

package webrtc

import (
	"errors"
	"net"
	"net/netip"
	"sync/atomic"
	"syscall"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

// gsoSocket sends a run of equal datagrams with one sendmsg and UDP_SEGMENT.
// Its methods run under the batchConn's lock.
type gsoSocket struct {
	raw syscall.RawConn
	v4  bool
	oob []byte
	off atomic.Bool
}

var errGSOAddress = errors.New("udp batch: address family does not match the socket")

func newGSOSocket(c *net.UDPConn) *gsoSocket {
	g := &gsoSocket{oob: make([]byte, unix.CmsgSpace(2))}

	raw, err := c.SyscallConn()
	if err != nil {
		g.off.Store(true)
		return g
	}
	g.raw = raw

	if la, ok := c.LocalAddr().(*net.UDPAddr); ok {
		g.v4 = la.IP.To4() != nil
	}

	h := (*unix.Cmsghdr)(unsafePointer(&g.oob[0]))
	h.Level = unix.SOL_UDP
	h.Type = unix.UDP_SEGMENT
	h.SetLen(unix.CmsgLen(2))

	return g
}

func (g *gsoSocket) usable() bool {
	return udpBatchEnabled && !g.off.Load()
}

func (g *gsoSocket) send(p []byte, size int, to netip.AddrPort) error {
	var sa unix.Sockaddr

	addr := to.Addr()
	if g.v4 {
		addr = addr.Unmap()
		if !addr.Is4() {
			return errGSOAddress
		}
		sa = &unix.SockaddrInet4{Port: int(to.Port()), Addr: addr.As4()}
	} else {
		sa = &unix.SockaddrInet6{Port: int(to.Port()), Addr: addr.As16()}
	}

	*(*uint16)(unsafePointer(&g.oob[unix.CmsgLen(0)])) = uint16(size)

	var serr error
	err := g.raw.Write(func(fd uintptr) bool {
		_, serr = unix.SendmsgN(int(fd), p, g.oob, sa, 0)
		return serr != unix.EAGAIN
	})
	if err != nil {
		return err
	}

	return serr
}

// refused turns UDP_SEGMENT off for this socket when the kernel says it
// cannot do it here (EIO: the route's device has no checksum offload, or an
// IPsec route; EINVAL; ENOPROTOOPT: a kernel without it). Other errors (a
// full buffer, an unreachable address) leave it on; the run goes out one
// datagram at a time either way.
func (g *gsoSocket) refused(err error) {
	switch {
	case errors.Is(err, unix.EIO), errors.Is(err, unix.EINVAL), errors.Is(err, unix.ENOPROTOOPT),
		errors.Is(err, unix.EOPNOTSUPP):
		if !g.off.Swap(true) {
			log.Infof("udp batch: UDP_SEGMENT refused (%v); one send per datagram on this socket", err)
		}
	}
}
