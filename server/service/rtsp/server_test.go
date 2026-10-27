package rtsp

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bluenviron/gortsplib/v5"
	"github.com/bluenviron/gortsplib/v5/pkg/base"
	"github.com/bluenviron/gortsplib/v5/pkg/description"
	"github.com/bluenviron/gortsplib/v5/pkg/format"
	"github.com/pion/rtp"

	"NanoKVM-Server/config"
	"NanoKVM-Server/service/stream/egress"
	"NanoKVM-Server/service/stream/egress/egresstest"
)

const goodToken = "nkvm_view-only"

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

type harness struct {
	server  *Server
	feed    *egresstest.Feed
	port    int
	tlsPort int
	stop    chan struct{}
	allowed atomic.Bool
	mu      sync.Mutex
	valid   bool
	// cert is what RTSPS serves; a test swaps it as the settings page would.
	cert atomic.Pointer[tls.Certificate]
}

// start runs a plain RTSP server on a free port with a tap the test feeds,
// and pumps 30 frames a second into it, with sound when withAudio.
func start(t *testing.T, withAudio bool) *harness {
	t.Helper()
	return launch(t, withAudio, config.RTSP{Enabled: true})
}

// launch is start with the listeners settings asks for, on free ports, and a
// test certificate for RTSPS.
func launch(t *testing.T, withAudio bool, settings config.RTSP) *harness {
	t.Helper()
	tap, feed := egresstest.New()
	h := &harness{feed: feed, port: freePort(t), tlsPort: freePort(t), stop: make(chan struct{}), valid: true}
	h.allowed.Store(true)
	h.cert.Store(testCertificate(t))

	settings.Port, settings.TLSPort = h.port, h.tlsPort
	settings.UDPPort = config.DefaultRTSPUDPPort
	h.server = New(Deps{
		Settings:      func() config.RTSP { return settings },
		ClientAllowed: func(net.Addr) bool { return h.allowed.Load() },
		Certificate: func() (*tls.Certificate, error) {
			if cert := h.cert.Load(); cert != nil {
				return cert, nil
			}
			return nil, errors.New("no certificate")
		},
		HasCertificate: func() bool { return h.cert.Load() != nil },
		Authenticate: func(token string) (egress.Viewer, bool) {
			if token != goodToken {
				return egress.Viewer{}, false
			}
			return egress.Viewer{Username: "watcher", Role: "viewer", Via: "apikey", Valid: func() bool {
				h.mu.Lock()
				defer h.mu.Unlock()
				return h.valid
			}}, true
		},
		Tap:            tap,
		AudioAvailable: func() bool { return withAudio },
		Recheck:        50 * time.Millisecond,
	})
	if err := h.server.Apply(); err != nil {
		t.Fatal(err)
	}

	feed.Pump(h.stop, 30, 10)
	if withAudio {
		feed.PumpAudio(h.stop, []byte{0xfc, 0xff, 0xfe})
	}

	t.Cleanup(func() {
		close(h.stop)
		h.server.Stop()
	})
	return h
}

func (h *harness) url(token string) string {
	userinfo := ""
	if token != "" {
		userinfo = "obs:" + token + "@"
	}
	return "rtsp://" + userinfo + "127.0.0.1:" + strconv.Itoa(h.port) + Path
}

func (h *harness) tlsURL(token string) string {
	return "rtsps://obs:" + token + "@127.0.0.1:" + strconv.Itoa(h.tlsPort) + Path
}

// testCertificate is a self-signed certificate for 127.0.0.1, as the board's
// generated one is self-signed for its address.
func testCertificate(t *testing.T) *tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: "rtsp test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return &tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: leaf}
}

// trusting is a client TLS configuration that accepts only cert, as a player
// told to trust the board's self-signed certificate does.
func trusting(cert *tls.Certificate) *tls.Config {
	pool := x509.NewCertPool()
	pool.AddCert(cert.Leaf)
	return &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}
}

// dial connects a player over TCP and describes the stream.
func dial(t *testing.T, address string) (*gortsplib.Client, *description.Session, error) {
	t.Helper()
	return dialWith(t, address, gortsplib.ProtocolTCP, nil)
}

// dialWith connects a player over the given transport, with the TLS
// configuration an rtsps:// address needs, and describes the stream.
func dialWith(t *testing.T, address string, protocol gortsplib.Protocol, tlsConfig *tls.Config) (*gortsplib.Client, *description.Session, error) {
	t.Helper()
	u, err := base.ParseURL(address)
	if err != nil {
		t.Fatal(err)
	}
	c := &gortsplib.Client{Scheme: u.Scheme, Host: u.Host, Protocol: &protocol, TLSConfig: tlsConfig}
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	desc, _, err := c.Describe(u)
	if err != nil {
		c.Close()
		return nil, nil, err
	}
	return c, desc, nil
}

// An RTSP player gets the injected H.264 with its parameter sets in the SDP,
// and the Opus beside it.
func TestAPlayerReceivesTheInjectedFrames(t *testing.T) {
	h := start(t, true)

	c, desc, err := dial(t, h.url(goodToken))
	if err != nil {
		t.Fatalf("DESCRIBE: %v", err)
	}
	defer c.Close()

	var h264 *format.H264
	videoMedia := desc.FindFormat(&h264)
	if videoMedia == nil {
		t.Fatalf("the description has no H.264: %+v", desc.Medias)
	}
	if !bytes.Equal(h264.SPS, egresstest.SPS) || !bytes.Equal(h264.PPS, egresstest.PPS) {
		t.Fatal("the description does not carry the keyframe's parameter sets")
	}
	var opus *format.Opus
	audioMedia := desc.FindFormat(&opus)
	if audioMedia == nil {
		t.Fatal("the description has no Opus track")
	}

	if err := c.SetupAll(desc.BaseURL, desc.Medias); err != nil {
		t.Fatalf("SETUP: %v", err)
	}

	decoder, err := h264.CreateDecoder()
	if err != nil {
		t.Fatal(err)
	}
	gotKey := make(chan struct{})
	gotAudio := make(chan struct{})
	var keyOnce, audioOnce sync.Once
	c.OnPacketRTP(videoMedia, h264, func(pkt *rtp.Packet) {
		au, err := decoder.Decode(pkt)
		if err != nil {
			return
		}
		for _, nalu := range au {
			if len(nalu) > 0 && nalu[0]&0x1f == 5 {
				keyOnce.Do(func() { close(gotKey) })
			}
		}
	})
	c.OnPacketRTP(audioMedia, opus, func(pkt *rtp.Packet) {
		if bytes.Equal(pkt.Payload, []byte{0xfc, 0xff, 0xfe}) {
			audioOnce.Do(func() { close(gotAudio) })
		}
	})
	if _, err := c.Play(nil); err != nil {
		t.Fatalf("PLAY: %v", err)
	}

	for name, ch := range map[string]chan struct{}{"an IDR picture": gotKey, "an Opus packet": gotAudio} {
		select {
		case <-ch:
		case <-time.After(5 * time.Second):
			t.Fatalf("the player never got %s", name)
		}
	}

	state := h.server.State()
	if len(state.Clients) != 1 || !state.Clients[0].Playing || state.Clients[0].User != "watcher" || state.Clients[0].Transport != "tcp" {
		t.Fatalf("state lists %+v, want one playing TCP client", state.Clients)
	}
}

func TestAPlayerWithoutAViewTokenIsRefused(t *testing.T) {
	h := start(t, false)

	for _, token := range []string{"", "nkvm_wrong"} {
		if c, _, err := dial(t, h.url(token)); err == nil {
			c.Close()
			t.Fatalf("DESCRIBE with token %q succeeded", token)
		}
	}
}

func TestAPlayerOutsideTheAllowedNetworksIsRefused(t *testing.T) {
	h := start(t, false)
	h.allowed.Store(false)

	if c, _, err := dial(t, h.url(goodToken)); err == nil {
		c.Close()
		t.Fatal("DESCRIBE from outside the allowed networks succeeded")
	}
}

func TestAtMostTwoPlayersAreServed(t *testing.T) {
	h := start(t, false)

	for i := range egress.MaxClients + 1 {
		c, desc, err := dial(t, h.url(goodToken))
		if err != nil {
			t.Fatalf("DESCRIBE %d: %v", i, err)
		}
		defer c.Close()

		err = c.SetupAll(desc.BaseURL, desc.Medias)
		if i < egress.MaxClients && err != nil {
			t.Fatalf("player %d was refused: %v", i, err)
		}
		if i == egress.MaxClients && err == nil {
			t.Fatalf("player %d was served past the limit of %d", i, egress.MaxClients)
		}
	}
}

// A player whose key is revoked is ended at the next check.
func TestARevokedPlayerIsEnded(t *testing.T) {
	h := start(t, false)

	c, desc, err := dial(t, h.url(goodToken))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := c.SetupAll(desc.BaseURL, desc.Medias); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Play(nil); err != nil {
		t.Fatal(err)
	}

	h.mu.Lock()
	h.valid = false
	h.mu.Unlock()

	done := make(chan error, 1)
	go func() { done <- c.Wait() }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the revoked player was not ended")
	}
}

func TestSwitchingOffClosesTheListener(t *testing.T) {
	h := start(t, false)
	if h.server.Addr() == nil {
		t.Fatal("the server is not listening")
	}

	h.server.deps.Settings = func() config.RTSP { return config.RTSP{Port: h.port} }
	if err := h.server.Apply(); err != nil {
		t.Fatal(err)
	}
	if h.server.Addr() != nil {
		t.Fatal("the server still listens after being switched off")
	}
	if conn, err := net.DialTimeout("tcp", "127.0.0.1:"+strconv.Itoa(h.port), time.Second); err == nil {
		conn.Close()
		t.Fatal("the port still answers")
	}
}

// Connections that never reach SETUP are bounded too: one past maxConns is
// closed at once, and a place comes back when one closes.
func TestOpenConnectionsAreBounded(t *testing.T) {
	h := start(t, false)
	address := "127.0.0.1:" + strconv.Itoa(h.port)

	var conns []net.Conn
	defer func() {
		for _, c := range conns {
			c.Close()
		}
	}()
	for range maxConns {
		c, err := net.Dial("tcp", address)
		if err != nil {
			t.Fatal(err)
		}
		conns = append(conns, c)
	}

	refusedAt := func() bool {
		c, err := net.Dial("tcp", address)
		if err != nil {
			return true
		}
		defer c.Close()
		_ = c.SetReadDeadline(time.Now().Add(time.Second))
		_, err = c.Read(make([]byte, 1))
		// A refused connection is closed by the server; an accepted one
		// waits for a request until the deadline.
		var timeout net.Error
		return !(errors.As(err, &timeout) && timeout.Timeout())
	}
	if !refusedAt() {
		t.Fatalf("a connection past the %d open ones was served", maxConns)
	}

	conns[0].Close()
	conns = conns[1:]
	deadline := time.Now().Add(3 * time.Second)
	for refusedAt() {
		if time.Now().After(deadline) {
			t.Fatal("no place came back when a connection closed")
		}
	}
}
