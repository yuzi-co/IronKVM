package webrtc

import (
	"bytes"
	"net"
	"net/netip"
	"runtime"
	"testing"
	"time"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

func TestRunEnd(t *testing.T) {
	a := netip.MustParseAddrPort("10.0.0.1:5000")
	b := netip.MustParseAddrPort("10.0.0.2:5000")
	d := func(n int, to netip.AddrPort) heldDatagram { return heldDatagram{n: n, to: to} }

	cases := []struct {
		name string
		held []heldDatagram
		ends []int
	}{
		{"one", []heldDatagram{d(100, a)}, []int{1}},
		{"equal", []heldDatagram{d(1216, a), d(1216, a), d(1216, a)}, []int{3}},
		{"shorter last", []heldDatagram{d(1216, a), d(1216, a), d(400, a), d(1216, a)}, []int{3, 4}},
		{"longer breaks", []heldDatagram{d(400, a), d(1216, a), d(1216, a)}, []int{1, 3}},
		{"two shorter", []heldDatagram{d(1216, a), d(400, a), d(400, a)}, []int{2, 3}},
		{"address", []heldDatagram{d(1216, a), d(1216, b), d(1216, b), d(1216, a)}, []int{1, 3, 4}},
	}
	for _, tc := range cases {
		var ends []int
		for i := 0; i < len(tc.held); {
			i = runEnd(tc.held, i)
			ends = append(ends, i)
		}
		if len(ends) != len(tc.ends) {
			t.Errorf("%s: runs end at %v, want %v", tc.name, ends, tc.ends)
			continue
		}
		for k := range ends {
			if ends[k] != tc.ends[k] {
				t.Errorf("%s: runs end at %v, want %v", tc.name, ends, tc.ends)
				break
			}
		}
	}

	// At most gsoMaxSegments, and within batchMaxBytes.
	var many []heldDatagram
	for i := 0; i < 100; i++ {
		many = append(many, d(500, a))
	}
	if got := runEnd(many, 0); got != gsoMaxSegments {
		t.Errorf("100 small datagrams: first run %d, want %d", got, gsoMaxSegments)
	}
	many = many[:0]
	for i := 0; i < 60; i++ {
		many = append(many, d(1216, a))
	}
	if got := runEnd(many, 0); got*1216 > batchMaxBytes || got != batchMaxBytes/1216 {
		t.Errorf("60 full datagrams: first run %d", got)
	}
}

func listenPair(t *testing.T) (*batchNet, *batchConn, *net.UDPConn) {
	t.Helper()
	n, err := newBatchNet()
	if err != nil {
		t.Fatal(err)
	}
	c, err := n.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	rx, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rx.Close() })
	_ = rx.SetReadBuffer(4 << 20)

	return n, c.(*batchConn), rx
}

func datagram(i, n int) []byte {
	b := make([]byte, n)
	for k := range b {
		b[k] = byte(i*7 + k)
	}
	b[0] = byte(i)

	return b
}

func receive(t *testing.T, rx *net.UDPConn, want [][]byte) {
	t.Helper()
	buf := make([]byte, 65536)
	for i, w := range want {
		_ = rx.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, _, err := rx.ReadFromUDP(buf)
		if err != nil {
			t.Fatalf("datagram %d of %d: %v", i, len(want), err)
		}
		if !bytes.Equal(buf[:n], w) {
			t.Fatalf("datagram %d: %d bytes starting %x, want %d starting %x", i, n, buf[:min(n, 4)], len(w), w[:min(len(w), 4)])
		}
	}
	_ = rx.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	if n, _, err := rx.ReadFromUDP(buf); err == nil {
		t.Fatalf("an extra datagram of %d bytes", n)
	}
}

// A frame's datagrams are held until the flush, then all arrive, whole and
// in order, most of them through UDP_SEGMENT sends.
func TestBatchConnHoldsAndFlushes(t *testing.T) {
	n, c, rx := listenPair(t)
	to := rx.LocalAddr().(*net.UDPAddr).AddrPort()

	var want [][]byte
	n.begin()
	for i := 0; i < 70; i++ {
		size := 1216
		if i == 39 || i == 69 {
			size = 517 // a frame's last packet
		}
		p := datagram(i, size)
		want = append(want, p)
		if _, err := c.WriteToAddrPort(p, to); err != nil {
			t.Fatal(err)
		}
		// The caller may reuse its buffer at once.
		p2 := append([]byte(nil), p...)
		copy(p, make([]byte, len(p)))
		copy(p, p2)
	}
	n.flush()
	receive(t, rx, want)

	if runtime.GOOS == "linux" {
		if c.gsoSends.Load() == 0 {
			t.Fatalf("no UDP_SEGMENT send (gso usable %v)", c.gso.usable())
		}
		if got := c.gsoDatagrams.Load(); got < 60 {
			t.Fatalf("%d of 70 datagrams went through UDP_SEGMENT", got)
		}
		t.Logf("70 datagrams: %d UDP_SEGMENT sends carried %d", c.gsoSends.Load(), c.gsoDatagrams.Load())
	}

	// Outside begin and flush, writes go straight out.
	p := datagram(1, 300)
	if _, err := c.WriteTo(p, rx.LocalAddr()); err != nil {
		t.Fatal(err)
	}
	receive(t, rx, [][]byte{p})
}

// Writes from other goroutines while a frame is held keep their order with it.
func TestBatchConnKeepsOrderAcrossWriters(t *testing.T) {
	n, c, rx := listenPair(t)
	to := rx.LocalAddr().(*net.UDPAddr).AddrPort()

	var want [][]byte
	n.begin()
	for i := 0; i < 10; i++ {
		p := datagram(i, 1216)
		want = append(want, p)
		_, _ = c.WriteToAddrPort(p, to)
		if i == 4 {
			done := make(chan struct{})
			q := datagram(100, 90) // an audio packet
			want = append(want, q)
			go func() { _, _ = c.WriteToAddrPort(q, to); close(done) }()
			<-done
		}
	}
	n.flush()
	receive(t, rx, want)
}

// With UDP_SEGMENT refused the same datagrams go out one at a time.
func TestBatchConnFallsBack(t *testing.T) {
	n, c, rx := listenPair(t)
	to := rx.LocalAddr().(*net.UDPAddr).AddrPort()
	c.gso.refusedForTest()

	var want [][]byte
	n.begin()
	for i := 0; i < 40; i++ {
		p := datagram(i, 1216)
		want = append(want, p)
		_, _ = c.WriteToAddrPort(p, to)
	}
	n.flush()
	receive(t, rx, want)
	if c.gsoSends.Load() != 0 {
		t.Fatal("a UDP_SEGMENT send after refusal")
	}
}

// End to end through pion: a peer connection built by createPeerConnection
// sends RTP to a plain pion receiver, a frame at a time between begin and
// flush, over SRTP, and every packet arrives.
func TestPeerConnectionThroughBatchNet(t *testing.T) {
	if !udpBatchEnabled {
		t.Skip("KVM_WEBRTC_UDP_BATCH=off")
	}

	sender, udp, err := createPeerConnection(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sender.Close() }()

	receiver, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = receiver.Close() }()

	track, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "kvm")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sender.AddTrack(track); err != nil {
		t.Fatal(err)
	}
	if _, err := receiver.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo,
		webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly}); err != nil {
		t.Fatal(err)
	}

	got := make(chan uint16, 1024)
	receiver.OnTrack(func(remote *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		for {
			p, _, err := remote.ReadRTP()
			if err != nil {
				return
			}
			got <- p.SequenceNumber
		}
	})

	connected := make(chan struct{})
	sender.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		if s == webrtc.PeerConnectionStateConnected {
			select {
			case <-connected:
			default:
				close(connected)
			}
		}
	})

	// The browser offers; this server answers.
	offer, err := receiver.CreateOffer(nil)
	if err != nil {
		t.Fatal(err)
	}
	gathered := webrtc.GatheringCompletePromise(receiver)
	if err := receiver.SetLocalDescription(offer); err != nil {
		t.Fatal(err)
	}
	<-gathered
	if err := sender.SetRemoteDescription(*receiver.LocalDescription()); err != nil {
		t.Fatal(err)
	}
	answer, err := sender.CreateAnswer(nil)
	if err != nil {
		t.Fatal(err)
	}
	gathered = webrtc.GatheringCompletePromise(sender)
	if err := sender.SetLocalDescription(answer); err != nil {
		t.Fatal(err)
	}
	<-gathered
	if err := receiver.SetRemoteDescription(*sender.LocalDescription()); err != nil {
		t.Fatal(err)
	}

	select {
	case <-connected:
	case <-time.After(10 * time.Second):
		t.Fatal("not connected")
	}

	// Three frames of 40 packets, the last of each shorter.
	const frames, per = 3, 40
	seq := uint16(1000)
	payload := bytes.Repeat([]byte{0x41}, 1180)
	for f := 0; f < frames; f++ {
		udp.begin()
		for i := 0; i < per; i++ {
			pl := payload
			if i == per-1 {
				pl = payload[:300]
			}
			p := &rtp.Packet{Header: rtp.Header{Version: 2, SequenceNumber: seq, Timestamp: uint32(f * 3000),
				Marker: i == per-1}, Payload: pl}
			seq++
			if err := track.WriteRTP(p); err != nil {
				t.Fatal(err)
			}
		}
		udp.flush()
	}

	seen := map[uint16]bool{}
	deadline := time.After(5 * time.Second)
	for len(seen) < frames*per {
		select {
		case s := <-got:
			seen[s] = true
		case <-deadline:
			t.Fatalf("received %d of %d packets", len(seen), frames*per)
		}
	}

	var sends, datagrams uint64
	udp.each(func(c *batchConn) {
		sends += c.gsoSends.Load()
		datagrams += c.gsoDatagrams.Load()
	})
	t.Logf("%d packets over SRTP; %d UDP_SEGMENT sends carried %d datagrams", frames*per, sends, datagrams)
	if runtime.GOOS == "linux" && sends == 0 {
		t.Fatal("no UDP_SEGMENT send")
	}
}
