package webrtc

import (
	"net"
	"net/netip"
	"os"
	"sync"
	"sync/atomic"

	"github.com/pion/transport/v4"
	"github.com/pion/transport/v4/stdnet"
	log "github.com/sirupsen/logrus"
)

// Batched UDP sends for the video writer (ironkvm-dist#72, run sheet trial 66).
//
// pion writes every RTP packet with its own sendto, and on the board's single
// C906 core the kernel's UDP send costs about 170 us of CPU a packet whatever
// its size (trial 56): the syscall, routing, netfilter (loaded for the VPN)
// and the socket accounting, each paid per datagram. A 1080p keyframe is 80 to
// 100 packets. With UDP generic segmentation offload (UDP_SEGMENT, Linux 4.18,
// so both of the board's kernels) one sendmsg carries up to 64 datagrams of
// one size to one address; the stack runs once for all of them and splits the
// buffer into datagrams just before the driver.
//
// Each peer connection's ICE agent listens through its own batchNet, so its
// sockets are batchConns. While the video writer writes a frame
// (batchNet.begin to batchNet.flush) a batchConn holds the datagrams written
// to it and sends them at the flush, or as soon as batchMaxPackets are held,
// as runs of equal size to one address: one UDP_SEGMENT send per run, a plain
// send for a run of one. Datagrams written by other goroutines in that window
// (audio, RTCP, STUN) are held with the frame's and keep their order. Outside
// the window a batchConn writes straight through.
//
// A socket whose kernel or route refuses UDP_SEGMENT (no checksum offload on
// the device, an IPsec route) falls back to one send a datagram for good, and
// logs it once. KVM_WEBRTC_UDP_BATCH=off turns batching off.

const (
	// batchMaxPackets bounds how long a held datagram waits: about 30 packets
	// of SRTP work, a few milliseconds on the board.
	batchMaxPackets = 32
	// batchMaxBytes keeps a run within one IP datagram's 64 KB.
	batchMaxBytes = 60000
	// gsoMaxSegments is UDP_MAX_SEGMENTS of Linux 5.10 (later kernels allow
	// more).
	gsoMaxSegments = 64
)

var udpBatchEnabled = os.Getenv("KVM_WEBRTC_UDP_BATCH") != "off"

// batchNet is the transport.Net of one peer connection.
type batchNet struct {
	*stdnet.Net

	mu    sync.Mutex
	conns []*batchConn
}

func newBatchNet() (*batchNet, error) {
	n, err := stdnet.NewNet()
	if err != nil {
		return nil, err
	}

	return &batchNet{Net: n}, nil
}

// ListenUDP is how ICE opens its host and server-reflexive sockets.
func (n *batchNet) ListenUDP(network string, locAddr *net.UDPAddr) (transport.UDPConn, error) {
	conn, err := n.Net.ListenUDP(network, locAddr)
	if err != nil {
		return nil, err
	}

	udp, ok := conn.(*net.UDPConn)
	if !ok {
		return conn, nil
	}

	bc := newBatchConn(udp)

	n.mu.Lock()
	n.conns = append(n.conns, bc)
	n.mu.Unlock()

	return bc, nil
}

func (n *batchNet) each(f func(*batchConn)) {
	if n == nil {
		return
	}

	n.mu.Lock()
	conns := n.conns
	n.mu.Unlock()

	for _, c := range conns {
		f(c)
	}
}

// begin holds what is written to this connection's sockets until flush.
func (n *batchNet) begin() {
	n.each((*batchConn).begin)
}

// flush sends what begin held and lets writes through again.
func (n *batchNet) flush() {
	n.each((*batchConn).flush)
}

type heldDatagram struct {
	off, n int
	to     netip.AddrPort
}

type batchConn struct {
	*net.UDPConn

	gso *gsoSocket

	mu      sync.Mutex
	holding bool
	buf     []byte
	held    []heldDatagram

	// Counters for the tests and the trial's diagnostics: datagrams sent in
	// UDP_SEGMENT sends, and those sends.
	gsoDatagrams, gsoSends atomic.Uint64
}

func newBatchConn(c *net.UDPConn) *batchConn {
	return &batchConn{UDPConn: c, gso: newGSOSocket(c)}
}

func (c *batchConn) begin() {
	if !udpBatchEnabled {
		return
	}

	c.mu.Lock()
	c.holding = true
	c.mu.Unlock()
}

func (c *batchConn) flush() {
	c.mu.Lock()
	c.sendHeldLocked()
	c.holding = false
	c.mu.Unlock()
}

// ReadFromAddrPort and WriteToAddrPort make a batchConn an ICE
// AddrPortReaderWriter, the allocation-free path pion takes with a plain
// *net.UDPConn.
func (c *batchConn) ReadFromAddrPort(b []byte) (int, netip.AddrPort, error) {
	return c.UDPConn.ReadFromUDPAddrPort(b)
}

func (c *batchConn) WriteToAddrPort(b []byte, to netip.AddrPort) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.holding || len(b) > batchMaxBytes {
		c.sendHeldLocked()
		return c.UDPConn.WriteToUDPAddrPort(b, to)
	}

	if len(c.held) >= batchMaxPackets || len(c.buf)+len(b) > batchMaxBytes {
		c.sendHeldLocked()
	}

	off := len(c.buf)
	c.buf = append(c.buf, b...)
	c.held = append(c.held, heldDatagram{off: off, n: len(b), to: to})

	return len(b), nil
}

func (c *batchConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	if ua, ok := addr.(*net.UDPAddr); ok && ua != nil {
		return c.WriteToAddrPort(b, ua.AddrPort())
	}

	c.mu.Lock()
	c.sendHeldLocked()
	c.mu.Unlock()

	return c.UDPConn.WriteTo(b, addr)
}

func (c *batchConn) WriteToUDP(b []byte, addr *net.UDPAddr) (int, error) {
	if addr == nil {
		return c.UDPConn.WriteToUDP(b, addr)
	}

	return c.WriteToAddrPort(b, addr.AddrPort())
}

// runEnd is the end of the run that starts at held[i]: datagrams to the same
// address and of the same size, and at most one shorter one to finish it, as
// UDP_SEGMENT takes them.
func runEnd(held []heldDatagram, i int) int {
	first := held[i]
	size := first.n
	total := size
	j := i + 1

	for j < len(held) && j-i < gsoMaxSegments && total+held[j].n <= batchMaxBytes &&
		held[j].to == first.to && held[j].n <= size {
		total += held[j].n
		j++
		if held[j-1].n < size {
			break
		}
	}

	return j
}

func (c *batchConn) sendHeldLocked() {
	held := c.held

	for i := 0; i < len(held); {
		j := runEnd(held, i)

		if j-i > 1 && c.gso.usable() {
			first, last := held[i], held[j-1]
			err := c.gso.send(c.buf[first.off:last.off+last.n], first.n, first.to)
			if err == nil {
				c.gsoDatagrams.Add(uint64(j - i))
				c.gsoSends.Add(1)
				i = j

				continue
			}

			c.gso.refused(err)
		}

		for ; i < j; i++ {
			d := held[i]
			// pion's own writes log and drop a failed send; so does this.
			if _, err := c.UDPConn.WriteToUDPAddrPort(c.buf[d.off:d.off+d.n], d.to); err != nil {
				log.Debugf("udp batch: send to %s: %v", d.to, err)
			}
		}
	}

	c.buf = c.buf[:0]
	c.held = c.held[:0]
}
