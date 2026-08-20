package utils

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// This file reproduces the outage rather than the helper that caused it.
//
// cert_test.go asserts that certNames builds the right list and that CertCovers
// reads it back. Both would pass against a product that still could not be
// used, because neither of them opens a connection. The fault was never in a
// list: it was that a browser refused the websocket carrying keyboard and mouse
// while every other part of the page kept working, so the board looked healthy
// and accepted no input.
//
// So these tests do what the browser does. They serve the real certificate from
// a real TLS listener, and open a real websocket to it, naming the device by the
// address an operator types and verifying the certificate strictly. The dial is
// redirected to the test listener, which is the only pretence: the name being
// verified, the certificate being served and the handshake being performed are
// all genuine.
//
// The browser's own behaviour is what made this invisible, and it cannot be
// modelled here. Chrome offers the operator a way past an untrusted certificate
// when it loads a page, and offers nothing when it opens a websocket to the same
// origin: it refuses, and the page is told nothing it can catch. So the page
// came up, and only the input was gone.

const wssTestAddr = "192.0.2.10" // TEST-NET-1, RFC 5737: never routable, never resolved

// serveTLS starts an HTTPS listener with the given certificate that upgrades
// every request to a websocket and echoes one message back.
func serveTLS(t *testing.T, certPath, keyPath string) string {
	t.Helper()

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		t.Fatalf("the generated pair does not load as a server certificate: %v", err)
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		t.Fatal(err)
	}

	upgrader := websocket.Upgrader{}
	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			_ = conn.WriteMessage(websocket.TextMessage, msg)
		}),
	}

	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() {
		_ = srv.Close()
	})

	return ln.Addr().String()
}

// rootsFrom builds a pool trusting exactly the generated certificate, which is
// the best case an operator can reach: they installed it. If the handshake still
// fails with this pool, no amount of trusting helps, because the certificate
// does not answer for the name.
func rootsFrom(t *testing.T, certPath string) *x509.CertPool {
	t.Helper()

	raw, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatal(err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(raw) {
		t.Fatal("the generated certificate is not usable as a trust root")
	}

	return pool
}

// openInputSocket dials the input websocket the way a browser would: the URL
// names the device, the certificate is verified against that name, and only the
// TCP destination is redirected to the test listener.
func openInputSocket(t *testing.T, listenAddr, certPath, host string) error {
	t.Helper()

	dialer := websocket.Dialer{
		NetDialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "tcp", listenAddr)
		},
		TLSClientConfig:  &tls.Config{RootCAs: rootsFrom(t, certPath)},
		HandshakeTimeout: 10 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, resp, err := dialer.DialContext(ctx, "wss://"+host+"/api/ws", nil)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return err
	}
	defer conn.Close()

	// The handshake is not the whole claim. Prove a keystroke could travel.
	if err := conn.WriteMessage(websocket.TextMessage, []byte("keydown")); err != nil {
		return err
	}
	_, got, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	if string(got) != "keydown" {
		t.Fatalf("the socket opened but did not carry the message: %q", got)
	}

	return nil
}

func certPairFor(t *testing.T, id certIdentity) (string, string) {
	t.Helper()

	dir := t.TempDir()
	crt := filepath.Join(dir, "server.crt")
	key := filepath.Join(dir, "server.key")

	if err := generateCertTo(crt, key, id); err != nil {
		t.Fatalf("generateCertTo: %v", err)
	}

	return crt, key
}

// TestTheShippedCertificateRemovesTheInput is the outage, reproduced.
//
// It must fail if the generator ever goes back to naming only localhost. That
// is not hypothetical: it is what the product did, and one click on the settings
// page was enough to reach it.
func TestTheShippedCertificateRemovesTheInput(t *testing.T) {
	shipped := certIdentity{
		DNS: []string{"localhost"},
		IPs: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	crt, key := certPairFor(t, shipped)
	addr := serveTLS(t, crt, key)

	err := openInputSocket(t, addr, crt, wssTestAddr)
	if err == nil {
		t.Fatal("the input websocket opened against a certificate that names only localhost")
	}

	// The reason matters. A refusal for some other cause would make this test
	// pass while proving nothing about the fault it exists for.
	if !strings.Contains(err.Error(), "certificate is valid for") &&
		!strings.Contains(err.Error(), "x509") {
		t.Fatalf("the socket failed for an unrelated reason: %v", err)
	}

	t.Logf("reproduced: %v", err)
}

// TestTheGeneratedCertificateCarriesTheInput is the fix, proven at the same
// layer. The certificate is trusted and names the device, so the socket opens
// and a keystroke crosses it.
func TestTheGeneratedCertificateCarriesTheInput(t *testing.T) {
	fixed := certNames("nanokvm", []net.IP{net.ParseIP(wssTestAddr)})

	crt, key := certPairFor(t, fixed)
	addr := serveTLS(t, crt, key)

	if err := openInputSocket(t, addr, crt, wssTestAddr); err != nil {
		t.Fatalf("the input websocket did not open against the device's own address: %v", err)
	}
}

// TestTheGeneratedCertificateAnswersForEveryNameItClaims walks the whole set.
// An operator reaches the board by address, by hostname or by the mDNS name that
// avahi answers, and all three have to carry the input.
func TestTheGeneratedCertificateAnswersForEveryNameItClaims(t *testing.T) {
	fixed := certNames("nanokvm", []net.IP{net.ParseIP(wssTestAddr)})

	crt, key := certPairFor(t, fixed)
	addr := serveTLS(t, crt, key)

	for _, host := range []string{wssTestAddr, "nanokvm", "nanokvm.local", "localhost", "127.0.0.1"} {
		if err := openInputSocket(t, addr, crt, host); err != nil {
			t.Errorf("the input websocket did not open for %q: %v", host, err)
		}
	}
}

// TestTheCertificateGoesStaleWhenTheAddressMoves is the same outage arriving
// without anybody touching a setting. DHCP moves the board, the certificate
// still names where it used to be, and the input stops.
func TestTheCertificateGoesStaleWhenTheAddressMoves(t *testing.T) {
	old := certNames("nanokvm", []net.IP{net.ParseIP("192.0.2.10")})
	crt, key := certPairFor(t, old)
	addr := serveTLS(t, crt, key)

	moved := "192.0.2.99"
	if err := openInputSocket(t, addr, crt, moved); err == nil {
		t.Fatal("a certificate for the previous address carried the input, which cannot be")
	}

	// EnsureCert is what closes it, without an operator being present.
	replaced, err := ensureCert(crt, key, certNames("nanokvm", []net.IP{net.ParseIP(moved)}))
	if err != nil {
		t.Fatalf("ensureCert: %v", err)
	}
	if !replaced {
		t.Fatal("the stale certificate was kept")
	}

	addr = serveTLS(t, crt, key)
	if err := openInputSocket(t, addr, crt, moved); err != nil {
		t.Fatalf("the input websocket did not open after the certificate was refreshed: %v", err)
	}
}

// TestTheRealIdentityCarriesTheInput runs the code the device runs, rather than
// a hand-built identity, against an address this machine actually has.
func TestTheRealIdentityCarriesTheInput(t *testing.T) {
	id := CurrentCertIdentity()

	var host string
	for _, ip := range id.IPs {
		if !ip.IsLoopback() {
			host = ip.String()
			break
		}
	}
	if host == "" {
		// Not a skip dressed as a pass: loopback is a real name the certificate
		// must answer for, and it is the one this machine has.
		host = "127.0.0.1"
		t.Log("this machine has no non-loopback address; verifying the loopback name instead")
	}

	crt, key := certPairFor(t, id)
	addr := serveTLS(t, crt, key)

	if err := openInputSocket(t, addr, crt, host); err != nil {
		t.Fatalf("CurrentCertIdentity does not carry the input for %q: %v", host, err)
	}
	t.Logf("verified against %s, names %v", host, id.DNS)
}

// TestAnExpiredCertificateIsRejected. An expired certificate fails the
// handshake exactly like a mismatched one, and takes the input with it in the
// same silence. EnsureCert has to notice and replace it.
func TestAnExpiredCertificateIsRejected(t *testing.T) {
	id := certNames("nanokvm", []net.IP{net.ParseIP(wssTestAddr)})

	restore := certValidFor
	certValidFor = -time.Hour
	crt, key := certPairFor(t, id)
	certValidFor = restore

	if CertCovers(crt, id) {
		t.Fatal("an expired certificate was reported as covering")
	}

	replaced, err := ensureCert(crt, key, id)
	if err != nil {
		t.Fatalf("ensureCert: %v", err)
	}
	if !replaced {
		t.Fatal("the expired certificate was kept")
	}

	addr := serveTLS(t, crt, key)
	if err := openInputSocket(t, addr, crt, wssTestAddr); err != nil {
		t.Fatalf("the input websocket did not open after the expired certificate was replaced: %v", err)
	}
}
