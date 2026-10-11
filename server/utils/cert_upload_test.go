package utils

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// testPKI is a throwaway CA, an intermediate and a leaf for kvm.example.
type testPKI struct {
	now               time.Time
	caCert, interCert *x509.Certificate
	caKey, interKey   crypto.Signer
	leafCert          *x509.Certificate
	leafKey           crypto.Signer
	leafPEM, interPEM []byte
	caPEM, leafKeyPEM []byte
}

var serial int64

func issue(t *testing.T, template *x509.Certificate, parent *x509.Certificate, key crypto.Signer, signer crypto.Signer) *x509.Certificate {
	t.Helper()
	serial++
	template.SerialNumber = big.NewInt(serial)
	if parent == nil {
		parent = template
	}
	der, err := x509.CreateCertificate(rand.Reader, template, parent, key.Public(), signer)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return cert
}

func ecKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func certPEM(certs ...*x509.Certificate) []byte {
	var out []byte
	for _, cert := range certs {
		out = append(out, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})...)
	}
	return out
}

func keyPEM(t *testing.T, key crypto.Signer) []byte {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

func newTestPKI(t *testing.T, leafNotAfter time.Duration) *testPKI {
	t.Helper()
	now := time.Now()
	p := &testPKI{now: now}
	caKey, interKey, leafKey := ecKey(t), ecKey(t), ecKey(t)
	p.caKey, p.interKey, p.leafKey = caKey, interKey, leafKey
	p.caCert = issue(t, &x509.Certificate{
		Subject:               pkix.Name{CommonName: "Test Root"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(10 * 365 * 24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}, nil, caKey, caKey)
	p.interCert = issue(t, &x509.Certificate{
		Subject:               pkix.Name{CommonName: "Test Intermediate"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(5 * 365 * 24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}, p.caCert, interKey, caKey)
	p.leafCert = issue(t, &x509.Certificate{
		Subject:     pkix.Name{CommonName: "kvm.example"},
		DNSNames:    []string{"kvm.example"},
		IPAddresses: []net.IP{net.ParseIP("192.0.2.10")},
		NotBefore:   now.Add(-time.Hour),
		NotAfter:    now.Add(leafNotAfter),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}, p.interCert, leafKey, interKey)
	p.leafPEM = certPEM(p.leafCert)
	p.interPEM = certPEM(p.interCert)
	p.caPEM = certPEM(p.caCert)
	p.leafKeyPEM = keyPEM(t, leafKey)
	return p
}

var testIdentity = CertIdentity{
	DNS: []string{"localhost", "kvm.example"},
	IPs: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("192.0.2.10")},
}

func wantKind(t *testing.T, err error, kind string) {
	t.Helper()
	var certErr *CertError
	if !errors.As(err, &certErr) {
		t.Fatalf("want a %s CertError, got %v", kind, err)
	}
	if certErr.Kind != kind {
		t.Fatalf("want kind %s, got %s (%s)", kind, certErr.Kind, certErr.Message)
	}
}

func TestValidateAcceptsAChain(t *testing.T) {
	p := newTestPKI(t, 90*24*time.Hour)
	chain := append(append([]byte{}, p.leafPEM...), p.interPEM...)
	info, err := ValidateCertPair(chain, p.leafKeyPEM, p.now, testIdentity)
	if err != nil {
		t.Fatal(err)
	}
	if info.ChainLength != 2 || info.Subject != "CN=kvm.example" || info.Issuer != "CN=Test Intermediate" {
		t.Fatalf("unexpected info %+v", info)
	}
	if len(info.DNSNames) != 1 || info.DNSNames[0] != "kvm.example" || info.IPAddresses[0] != "192.0.2.10" {
		t.Fatalf("unexpected names %+v", info)
	}
	if info.KeyType != "ECDSA P-256" || len(info.Fingerprint) != 64 {
		t.Fatalf("unexpected key or fingerprint %+v", info)
	}
	if len(info.Warnings) != 0 || len(info.Uncovered) != 0 {
		t.Fatalf("unexpected warnings %v %v", info.Warnings, info.Uncovered)
	}
	// The root may come along too.
	full := append(append([]byte{}, chain...), p.caPEM...)
	if _, err := ValidateCertPair(full, p.leafKeyPEM, p.now, testIdentity); err != nil {
		t.Fatalf("chain with its root: %v", err)
	}
}

func TestValidateRefusesAKeyThatDoesNotMatch(t *testing.T) {
	p := newTestPKI(t, 90*24*time.Hour)
	_, err := ValidateCertPair(p.leafPEM, keyPEM(t, ecKey(t)), p.now, testIdentity)
	wantKind(t, err, CertErrMismatch)
	// The intermediate's key does not serve the leaf either.
	_, err = ValidateCertPair(p.leafPEM, keyPEM(t, p.interKey), p.now, testIdentity)
	wantKind(t, err, CertErrMismatch)
}

func TestValidateRefusesAChainOutOfOrder(t *testing.T) {
	p := newTestPKI(t, 90*24*time.Hour)
	// Leaf, then the root: the intermediate is missing in between.
	_, err := ValidateCertPair(append(append([]byte{}, p.leafPEM...), p.caPEM...), p.leafKeyPEM, p.now, testIdentity)
	wantKind(t, err, CertErrBadChain)
	// Intermediate first: the key belongs to the leaf, not to the first certificate.
	_, err = ValidateCertPair(append(append([]byte{}, p.interPEM...), p.leafPEM...), p.leafKeyPEM, p.now, testIdentity)
	wantKind(t, err, CertErrMismatch)
}

func TestValidateRefusesOutsideValidity(t *testing.T) {
	p := newTestPKI(t, 90*24*time.Hour)
	_, err := ValidateCertPair(p.leafPEM, p.leafKeyPEM, p.now.Add(91*24*time.Hour), testIdentity)
	wantKind(t, err, CertErrExpired)
	_, err = ValidateCertPair(p.leafPEM, p.leafKeyPEM, p.now.Add(-2*time.Hour), testIdentity)
	wantKind(t, err, CertErrNotYetValid)
}

func TestValidateWarnsBeforeExpiry(t *testing.T) {
	p := newTestPKI(t, 10*24*time.Hour)
	info, err := ValidateCertPair(p.leafPEM, p.leafKeyPEM, p.now, testIdentity)
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Warnings) != 1 || info.Warnings[0] != CertWarnExpiresSoon {
		t.Fatalf("warnings %v", info.Warnings)
	}
}

func TestValidateReportsNamesItDoesNotCover(t *testing.T) {
	p := newTestPKI(t, 90*24*time.Hour)
	id := CertIdentity{
		DNS: []string{"localhost", "nanokvm", "nanokvm.local"},
		IPs: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("192.0.2.10"), net.ParseIP("192.0.2.11")},
	}
	info, err := ValidateCertPair(p.leafPEM, p.leafKeyPEM, p.now, id)
	if err != nil {
		t.Fatal(err)
	}
	want := "nanokvm,nanokvm.local,192.0.2.11"
	if strings.Join(info.Uncovered, ",") != want {
		t.Fatalf("uncovered %v, want %s", info.Uncovered, want)
	}
	if len(info.Warnings) != 1 || info.Warnings[0] != CertWarnNameMismatch {
		t.Fatalf("warnings %v", info.Warnings)
	}
}

func TestValidateRefusesWhatIsNotAPair(t *testing.T) {
	p := newTestPKI(t, 90*24*time.Hour)
	_, err := ValidateCertPair([]byte("hello"), p.leafKeyPEM, p.now, testIdentity)
	wantKind(t, err, CertErrNoCertificate)
	_, err = ValidateCertPair(p.leafKeyPEM, p.leafKeyPEM, p.now, testIdentity)
	wantKind(t, err, CertErrBadCert)
	_, err = ValidateCertPair(p.leafPEM, []byte("hello"), p.now, testIdentity)
	wantKind(t, err, CertErrNoKey)
	_, err = ValidateCertPair(p.leafPEM, p.leafPEM, p.now, testIdentity)
	wantKind(t, err, CertErrNoKey)
	broken := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("not der")})
	_, err = ValidateCertPair(broken, p.leafKeyPEM, p.now, testIdentity)
	wantKind(t, err, CertErrBadCert)
	_, err = ValidateCertPair(append(append([]byte{}, p.leafPEM...), []byte("trailing junk")...), p.leafKeyPEM, p.now, testIdentity)
	wantKind(t, err, CertErrBadCert)
	_, err = ValidateCertPair(make([]byte, MaxCertPEMSize+1), p.leafKeyPEM, p.now, testIdentity)
	wantKind(t, err, CertErrTooLarge)
}

func TestValidateRefusesAnEncryptedKey(t *testing.T) {
	p := newTestPKI(t, 90*24*time.Hour)
	encrypted := pem.EncodeToMemory(&pem.Block{Type: "ENCRYPTED PRIVATE KEY", Bytes: []byte{1, 2, 3}})
	_, err := ValidateCertPair(p.leafPEM, encrypted, p.now, testIdentity)
	wantKind(t, err, CertErrEncryptedKey)
	legacy := pem.EncodeToMemory(&pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: map[string]string{"Proc-Type": "4,ENCRYPTED", "DEK-Info": "AES-128-CBC,00"},
		Bytes:   []byte{1, 2, 3},
	})
	_, err = ValidateCertPair(p.leafPEM, legacy, p.now, testIdentity)
	wantKind(t, err, CertErrEncryptedKey)
}

func TestValidateRefusesAShortRSAKey(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	cert := issue(t, &x509.Certificate{
		Subject:   pkix.Name{CommonName: "kvm.example"},
		DNSNames:  []string{"kvm.example"},
		NotBefore: now.Add(-time.Hour),
		NotAfter:  now.Add(24 * time.Hour),
	}, nil, key, key)
	_, err = ValidateCertPair(certPEM(cert), keyPEM(t, key), now, testIdentity)
	wantKind(t, err, CertErrWeakKey)
}

func TestValidateRefusesAClientCertificate(t *testing.T) {
	key := ecKey(t)
	now := time.Now()
	cert := issue(t, &x509.Certificate{
		Subject:     pkix.Name{CommonName: "someone"},
		NotBefore:   now.Add(-time.Hour),
		NotAfter:    now.Add(24 * time.Hour),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}, nil, key, key)
	_, err := ValidateCertPair(certPEM(cert), keyPEM(t, key), now, testIdentity)
	wantKind(t, err, CertErrNotServer)
}

func TestValidateAcceptsAnECKeyWithItsParameters(t *testing.T) {
	p := newTestPKI(t, 90*24*time.Hour)
	der, err := x509.MarshalECPrivateKey(p.leafKey.(*ecdsa.PrivateKey))
	if err != nil {
		t.Fatal(err)
	}
	params := pem.EncodeToMemory(&pem.Block{Type: "EC PARAMETERS", Bytes: []byte{0x06, 0x08, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x03, 0x01, 0x07}})
	key := append(params, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der})...)
	if _, err := ValidateCertPair(p.leafPEM, key, p.now, testIdentity); err != nil {
		t.Fatal(err)
	}
}

func TestValidateMarksASelfSignedCertificate(t *testing.T) {
	key := ecKey(t)
	now := time.Now()
	cert := issue(t, &x509.Certificate{
		Subject:     pkix.Name{CommonName: "kvm.example"},
		DNSNames:    []string{"kvm.example"},
		IPAddresses: []net.IP{net.ParseIP("192.0.2.10")},
		NotBefore:   now.Add(-time.Hour),
		NotAfter:    now.Add(365 * 24 * time.Hour),
	}, nil, key, key)
	info, err := ValidateCertPair(certPEM(cert), keyPEM(t, key), now, testIdentity)
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Warnings) != 1 || info.Warnings[0] != CertWarnSelfSigned {
		t.Fatalf("warnings %v", info.Warnings)
	}
}

// The error a refusal returns never quotes the key it was given.
func TestValidateNeverEchoesTheKey(t *testing.T) {
	p := newTestPKI(t, 90*24*time.Hour)
	other := keyPEM(t, ecKey(t))
	_, err := ValidateCertPair(p.leafPEM, other, p.now, testIdentity)
	body := strings.Split(string(other), "\n")[1]
	if err == nil || strings.Contains(err.Error(), body[:16]) {
		t.Fatalf("error %v", err)
	}
}

func TestWriteCertPairModes(t *testing.T) {
	p := newTestPKI(t, 90*24*time.Hour)
	dir := t.TempDir()
	certPath, keyPath := filepath.Join(dir, "up.crt"), filepath.Join(dir, "up.key")
	if err := WriteCertPair(certPath, keyPath, p.leafPEM, p.leafKeyPEM); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("key mode %o", info.Mode().Perm())
	}
	if _, err := tls.LoadX509KeyPair(certPath, keyPath); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 2 {
		t.Fatalf("temporary files left behind: %d entries", len(entries))
	}
}

// A listener that serves from ActiveCert picks up a new certificate on the
// next handshake, without a restart.
func TestServingCertSwapsWithoutRestart(t *testing.T) {
	dir := t.TempDir()
	first := newTestPKI(t, 90*24*time.Hour)
	second := newTestPKI(t, 90*24*time.Hour)
	write := func(p *testPKI, name string) (string, string) {
		certPath, keyPath := filepath.Join(dir, name+".crt"), filepath.Join(dir, name+".key")
		if err := WriteCertPair(certPath, keyPath, append(append([]byte{}, p.leafPEM...), p.interPEM...), p.leafKeyPEM); err != nil {
			t.Fatal(err)
		}
		return certPath, keyPath
	}
	firstCert, firstKey := write(first, "first")
	secondCert, secondKey := write(second, "second")

	serving := &ServingCert{}
	if _, err := serving.GetCertificate(nil); err == nil {
		t.Fatal("a certificate before any load")
	}
	if err := serving.Load(firstCert, firstKey); err != nil {
		t.Fatal(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := NewServer(listener.Addr().String(), http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	server.TLSConfig = &tls.Config{GetCertificate: serving.GetCertificate}
	go func() { _ = server.ServeTLS(listener, "", "") }()
	t.Cleanup(func() { _ = server.Close() })

	leafOf := func() *x509.Certificate {
		conn, err := tls.Dial("tcp", listener.Addr().String(), &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		return conn.ConnectionState().PeerCertificates[0]
	}
	if !leafOf().Equal(first.leafCert) {
		t.Fatal("not serving the first certificate")
	}
	if err := serving.Load(secondCert, secondKey); err != nil {
		t.Fatal(err)
	}
	if !leafOf().Equal(second.leafCert) {
		t.Fatal("not serving the second certificate after the swap")
	}
	if err := serving.Load(filepath.Join(dir, "missing.crt"), secondKey); err == nil {
		t.Fatal("a missing file loaded")
	}
	if !leafOf().Equal(second.leafCert) {
		t.Fatal("a failed load replaced the certificate in force")
	}
}
