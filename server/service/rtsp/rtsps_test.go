package rtsp

import (
	"bytes"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bluenviron/gortsplib/v5"
	"github.com/bluenviron/gortsplib/v5/pkg/description"
	"github.com/bluenviron/gortsplib/v5/pkg/format"
	"github.com/pion/rtp"

	"NanoKVM-Server/config"
	"NanoKVM-Server/service/stream/egress/egresstest"
)

// playUntilKeyframe sets up every track, plays, and waits for an IDR
// picture.
func playUntilKeyframe(t *testing.T, c *gortsplib.Client, address string, desc *description.Session) {
	t.Helper()
	var h264 *format.H264
	videoMedia := desc.FindFormat(&h264)
	if videoMedia == nil {
		t.Fatalf("%s: the description has no H.264", address)
	}
	if !bytes.Equal(h264.SPS, egresstest.SPS) || !bytes.Equal(h264.PPS, egresstest.PPS) {
		t.Fatalf("%s: the description does not carry the keyframe's parameter sets", address)
	}
	if err := c.SetupAll(desc.BaseURL, desc.Medias); err != nil {
		t.Fatalf("%s: SETUP: %v", address, err)
	}

	decoder, err := h264.CreateDecoder()
	if err != nil {
		t.Fatal(err)
	}
	gotKey := make(chan struct{})
	var once sync.Once
	c.OnPacketRTP(videoMedia, h264, func(pkt *rtp.Packet) {
		au, err := decoder.Decode(pkt)
		if err != nil {
			return
		}
		for _, nalu := range au {
			if len(nalu) > 0 && nalu[0]&0x1f == 5 {
				once.Do(func() { close(gotKey) })
			}
		}
	})
	if _, err := c.Play(nil); err != nil {
		t.Fatalf("%s: PLAY: %v", address, err)
	}

	select {
	case <-gotKey:
	case <-time.After(5 * time.Second):
		t.Fatalf("%s: the player never got an IDR picture", address)
	}
}

// An RTSPS player that trusts the board's certificate gets the injected
// frames over TLS, while a plain RTSP player on the other listener gets the
// same feed.
func TestAnRTSPSPlayerReceivesTheInjectedFrames(t *testing.T) {
	h := launch(t, false, config.RTSP{Enabled: true, TLS: true})
	if h.server.Addr() == nil || h.server.TLSAddr() == nil {
		t.Fatal("both listeners should be up")
	}

	secure, desc, err := dialWith(t, h.tlsURL(goodToken), gortsplib.ProtocolTCP, trusting(h.cert.Load()))
	if err != nil {
		t.Fatalf("DESCRIBE over RTSPS: %v", err)
	}
	defer secure.Close()
	playUntilKeyframe(t, secure, h.tlsURL(goodToken), desc)

	plain, desc, err := dial(t, h.url(goodToken))
	if err != nil {
		t.Fatalf("DESCRIBE over RTSP: %v", err)
	}
	defer plain.Close()
	playUntilKeyframe(t, plain, h.url(goodToken), desc)

	state := h.server.State()
	if !state.TLS || !state.Certificate || !state.Listening {
		t.Fatalf("state %+v does not show RTSPS on with a certificate", state)
	}
	var secureClients, plainClients int
	for _, client := range state.Clients {
		if client.Secure {
			secureClients++
		} else {
			plainClients++
		}
	}
	if secureClients != 1 || plainClients != 1 {
		t.Fatalf("state lists %+v, want one RTSPS and one RTSP player", state.Clients)
	}
}

// With only RTSPS on, the plain port is closed: the token never crosses the
// network in clear text.
func TestRTSPSAloneLeavesThePlainPortClosed(t *testing.T) {
	h := launch(t, false, config.RTSP{TLS: true})
	if h.server.Addr() != nil {
		t.Fatal("the plain listener is up with only RTSPS on")
	}
	if c, _, err := dial(t, h.url(goodToken)); err == nil {
		c.Close()
		t.Fatal("a plain RTSP player was served with only RTSPS on")
	}

	c, desc, err := dialWith(t, h.tlsURL(goodToken), gortsplib.ProtocolTCP, trusting(h.cert.Load()))
	if err != nil {
		t.Fatalf("DESCRIBE over RTSPS: %v", err)
	}
	defer c.Close()
	playUntilKeyframe(t, c, h.tlsURL(goodToken), desc)
}

// RTSPS keeps the media in the TLS connection even when plain RTSP offers
// UDP: a player that asks for UDP is refused.
func TestRTSPSRefusesUDP(t *testing.T) {
	h := launch(t, false, config.RTSP{Enabled: true, TLS: true, UDP: true, UDPPort: freeUDPPair(t)})

	c, desc, err := dialWith(t, h.tlsURL(goodToken), gortsplib.ProtocolUDP, trusting(h.cert.Load()))
	if err != nil {
		t.Fatalf("DESCRIBE over RTSPS: %v", err)
	}
	defer c.Close()
	if err := c.SetupAll(desc.BaseURL, desc.Medias); err == nil {
		t.Fatal("an RTSPS player was given RTP over UDP")
	}
}

// A player that does not trust the certificate gets nothing, and one that
// sends no TLS at all gets no RTSP answer.
func TestRTSPSNeedsTheCertificateTrusted(t *testing.T) {
	h := launch(t, false, config.RTSP{TLS: true})

	if c, _, err := dialWith(t, h.tlsURL(goodToken), gortsplib.ProtocolTCP, trusting(testCertificate(t))); err == nil {
		c.Close()
		t.Fatal("a player that trusts another certificate was served")
	}

	plainOnTLSPort := strings.Replace(h.tlsURL(goodToken), "rtsps://", "rtsp://", 1)
	if c, _, err := dial(t, plainOnTLSPort); err == nil {
		c.Close()
		t.Fatal("the RTSPS port answered plain RTSP")
	}
}

// A certificate replaced on the settings page is served to the next player
// without a restart, as the web server serves it.
func TestANewCertificateServesTheNextPlayer(t *testing.T) {
	h := launch(t, false, config.RTSP{TLS: true})
	old := h.cert.Load()

	next := testCertificate(t)
	h.cert.Store(next)

	c, _, err := dialWith(t, h.tlsURL(goodToken), gortsplib.ProtocolTCP, trusting(next))
	if err != nil {
		t.Fatalf("a player trusting the new certificate was refused: %v", err)
	}
	c.Close()

	if c, _, err := dialWith(t, h.tlsURL(goodToken), gortsplib.ProtocolTCP, trusting(old)); err == nil {
		c.Close()
		t.Fatal("the old certificate is still served")
	}
}

// RTSPS without a certificate does not start, and says why, rather than
// failing every handshake. Plain RTSP waits for it, so the page never shows
// half of what was asked for.
func TestRTSPSWithoutACertificateDoesNotStart(t *testing.T) {
	h := start(t, false)

	h.cert.Store(nil)
	h.server.deps.Settings = func() config.RTSP {
		return config.RTSP{Enabled: true, Port: h.port, TLS: true, TLSPort: h.tlsPort, UDPPort: config.DefaultRTSPUDPPort}
	}
	err := h.server.Apply()
	if !errors.Is(err, errNoCertificate) {
		t.Fatalf("Apply returned %v, want %v", err, errNoCertificate)
	}
	state := h.server.State()
	if state.Listening || state.Certificate || !strings.Contains(state.Error, "certificate") {
		t.Fatalf("state %+v, want not listening, no certificate, and the reason", state)
	}
	if h.server.Addr() != nil || h.server.TLSAddr() != nil {
		t.Fatal("a listener is up after RTSPS failed to start")
	}
}

func TestCheckPorts(t *testing.T) {
	web := []int{80, 443}
	for _, tc := range []struct {
		name     string
		settings config.RTSP
		ok       bool
	}{
		{"the defaults", config.RTSP{}, true},
		{"RTSP on a web port", config.RTSP{Port: 443}, false},
		{"RTSPS on a web port", config.RTSP{TLSPort: 80}, false},
		{"both on one port", config.RTSP{Port: 9000, TLSPort: 9000}, false},
		{"RTSPS on RTSP's default port", config.RTSP{TLSPort: config.DefaultRTSPPort}, false},
		{"two ports of their own", config.RTSP{Port: 9000, TLSPort: 9001}, true},
	} {
		if err := CheckPorts(tc.settings, web); (err == nil) != tc.ok {
			t.Errorf("%s: CheckPorts returned %v", tc.name, err)
		}
	}
}

// freeUDPPair is an even port with the one above it free, for RTP and RTCP.
func freeUDPPair(t *testing.T) int {
	t.Helper()
	for range 20 {
		c, err := net.ListenPacket("udp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		port := c.LocalAddr().(*net.UDPAddr).Port
		c.Close()
		if port >= 65534 {
			continue
		}
		if next, err := net.ListenPacket("udp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port+1))); err == nil {
			next.Close()
			return port
		}
	}
	t.Fatal("no free pair of UDP ports")
	return 0
}
