//go:build !linux

package webrtc

import (
	"errors"
	"net"
	"net/netip"
)

// gsoSocket: UDP_SEGMENT is Linux's. Elsewhere a held run goes out one
// datagram at a time.
type gsoSocket struct{}

func newGSOSocket(*net.UDPConn) *gsoSocket { return &gsoSocket{} }

func (g *gsoSocket) usable() bool { return false }

func (g *gsoSocket) send([]byte, int, netip.AddrPort) error {
	return errors.ErrUnsupported
}

func (g *gsoSocket) refused(error) {}
