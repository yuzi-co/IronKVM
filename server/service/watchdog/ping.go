package watchdog

import (
	"context"
	"errors"
	"math/rand/v2"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// pingTimeout bounds one echo request, so a host that does not answer holds
// up a sample by no more than this.
const pingTimeout = 2 * time.Second

// Ping sends one ICMP echo request to host, an IPv4 or IPv6 address, and
// reports whether a reply came back in time.
func Ping(ctx context.Context, host string) bool {
	ok, _ := ping(ctx, host)
	return ok
}

// ping does the work of Ping and says why it failed. It opens a raw socket,
// which the server can because it runs as root.
func ping(ctx context.Context, host string) (bool, error) {
	ip := net.ParseIP(host)
	if ip == nil {
		return false, errors.New("not an IP address")
	}

	network, address, protocol := "ip4:icmp", "0.0.0.0", 1
	var request, reply icmp.Type = ipv4.ICMPTypeEcho, ipv4.ICMPTypeEchoReply
	if ip.To4() == nil {
		network, address, protocol = "ip6:ipv6-icmp", "::", 58
		request, reply = ipv6.ICMPTypeEchoRequest, ipv6.ICMPTypeEchoReply
	}

	conn, err := icmp.ListenPacket(network, address)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	deadline := time.Now().Add(pingTimeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return false, err
	}

	// A raw socket sees every echo reply the board receives, so the ID and
	// the sequence number pick out the answer to this request.
	id := os.Getpid() & 0xffff
	seq := rand.IntN(0x10000)
	message := icmp.Message{
		Type: request,
		Body: &icmp.Echo{ID: id, Seq: seq, Data: []byte("ironkvm watchdog")},
	}
	data, err := message.Marshal(nil)
	if err != nil {
		return false, err
	}
	if _, err := conn.WriteTo(data, &net.IPAddr{IP: ip}); err != nil {
		return false, err
	}

	buf := make([]byte, 1500)
	for {
		n, peer, err := conn.ReadFrom(buf)
		if err != nil {
			return false, err
		}
		parsed, err := icmp.ParseMessage(protocol, buf[:n])
		if err != nil || parsed.Type != reply {
			continue
		}
		echo, ok := parsed.Body.(*icmp.Echo)
		if !ok || echo.ID != id || echo.Seq != seq {
			continue
		}
		if from, ok := peer.(*net.IPAddr); ok && !from.IP.Equal(ip) {
			continue
		}
		return true, nil
	}
}
