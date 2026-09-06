package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"
)

// selfSignedCert builds a certificate the test client will accept, so the
// handshake gets far enough to report which protocol was negotiated.
func selfSignedCert(t *testing.T) tls.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:         true,
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}

	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: leaf}
}

// TestTLSNegotiatesHTTP11 is the case that matters. Asserting on the field
// alone would pass for a nil map on a Go release that changed the default, and
// the negotiated protocol is what the cost and the websocket both depend on.
func TestTLSNegotiatesHTTP11(t *testing.T) {
	cert := selfSignedCert(t)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	server := NewServer(listener.Addr().String(), http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	server.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}

	go func() { _ = server.ServeTLS(listener, "", "") }()
	defer func() { _ = server.Close() }()

	pool := x509.NewCertPool()
	pool.AddCert(cert.Leaf)

	// The client offers h2 first. A server that left HTTP/2 on picks it.
	conn, err := tls.Dial("tcp", listener.Addr().String(), &tls.Config{
		RootCAs:    pool,
		NextProtos: []string{"h2", "http/1.1"},
		ServerName: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	if got := conn.ConnectionState().NegotiatedProtocol; got != "http/1.1" {
		t.Fatalf("negotiated %q over TLS, want http/1.1: HTTP/2 costs the video "+
			"stream and cannot carry the /api/ws websocket", got)
	}
}

// TestTLSNextProtoIsEmptyNotNil guards the mechanism. A nil map is how Go is
// asked to configure HTTP/2, so the difference between nil and empty is the
// whole change and it is easy to undo by accident.
func TestTLSNextProtoIsEmptyNotNil(t *testing.T) {
	server := NewServer("127.0.0.1:0", http.NotFoundHandler())

	if server.TLSNextProto == nil {
		t.Fatal("TLSNextProto is nil, which is what turns HTTP/2 on")
	}

	if len(server.TLSNextProto) != 0 {
		t.Fatalf("TLSNextProto has %d entries, want none", len(server.TLSNextProto))
	}
}

// TestTimeoutsSurvive keeps the fields the constructor already existed for.
func TestTimeoutsSurvive(t *testing.T) {
	server := NewServer("127.0.0.1:0", http.NotFoundHandler())

	if server.ReadHeaderTimeout != readHeaderTimeout {
		t.Errorf("ReadHeaderTimeout is %s, want %s", server.ReadHeaderTimeout, readHeaderTimeout)
	}

	if server.IdleTimeout != idleTimeout {
		t.Errorf("IdleTimeout is %s, want %s", server.IdleTimeout, idleTimeout)
	}

	if server.ReadTimeout != 0 || server.WriteTimeout != 0 {
		t.Error("a whole-request deadline would cut off the MJPEG stream")
	}
}
