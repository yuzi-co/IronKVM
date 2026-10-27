package utils

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"NanoKVM-Server/utils/atomicfile"
)

// An uploaded certificate lives beside the generated one, in /etc/kvm. On a
// board with slots /etc/kvm is bound from /data/identity, so it is on /data
// and survives a slot switch the same way the account file does. server.yaml's
// cert.crt and cert.key name whichever pair is in force.
const (
	UploadedCertFile = "/etc/kvm/uploaded.crt"
	UploadedKeyFile  = "/etc/kvm/uploaded.key"
)

// Limits on an upload. A chain of a leaf and two intermediates in PEM is
// about 6 KiB, and an RSA-4096 key about 3 KiB.
const (
	MaxCertPEMSize = 64 << 10
	MaxKeyPEMSize  = 16 << 10
)

// certExpiryWarning is how close to its end a certificate has to be for the
// page to warn about it.
const certExpiryWarning = 30 * 24 * time.Hour

// CertError is a validation failure, with a Kind the page translates. Its
// message never quotes the key.
type CertError struct {
	Kind    string
	Message string
}

func (e *CertError) Error() string { return e.Message }

func certErr(kind, format string, args ...any) error {
	return &CertError{Kind: kind, Message: fmt.Sprintf(format, args...)}
}

// The kinds of CertError.
const (
	CertErrTooLarge      = "tooLarge"
	CertErrNoCertificate = "noCertificate"
	CertErrBadCert       = "badCertificate"
	CertErrNoKey         = "noKey"
	CertErrEncryptedKey  = "encryptedKey"
	CertErrBadKey        = "badKey"
	CertErrMismatch      = "keyMismatch"
	CertErrExpired       = "expired"
	CertErrNotYetValid   = "notYetValid"
	CertErrBadChain      = "badChain"
	CertErrWeakKey       = "weakKey"
	CertErrNotServer     = "notServerCert"
)

// The warnings CertInfo can carry. A warning does not stop an install.
const (
	CertWarnExpiresSoon  = "expiresSoon"
	CertWarnNameMismatch = "nameMismatch"
	CertWarnSelfSigned   = "selfSigned"
)

// CertInfo describes a certificate for the settings page.
type CertInfo struct {
	Subject     string    `json:"subject"`
	Issuer      string    `json:"issuer"`
	DNSNames    []string  `json:"dnsNames"`
	IPAddresses []string  `json:"ipAddresses"`
	NotBefore   time.Time `json:"notBefore"`
	NotAfter    time.Time `json:"notAfter"`
	// Fingerprint is the SHA-256 of the leaf, as browsers show it.
	Fingerprint string `json:"fingerprint"`
	KeyType     string `json:"keyType"`
	// ChainLength counts the leaf and the intermediates sent with it.
	ChainLength int `json:"chainLength"`
	// Uncovered lists this device's names and addresses that the
	// certificate does not answer for.
	Uncovered []string `json:"uncovered,omitempty"`
	Warnings  []string `json:"warnings,omitempty"`
}

// ValidateCertPair checks an uploaded certificate chain and key and describes
// them. It refuses what would fail every handshake or is unsafe to serve: a
// key that does not match, a chain out of order, a certificate outside its
// validity, an encrypted or weak key. It only warns about what may be
// intended, such as a name the device is not reached by.
func ValidateCertPair(certPEM, keyPEM []byte, now time.Time, id CertIdentity) (*CertInfo, error) {
	if len(certPEM) > MaxCertPEMSize || len(keyPEM) > MaxKeyPEMSize {
		return nil, certErr(CertErrTooLarge, "the certificate or key is too large")
	}

	chain, err := parseCertChain(certPEM)
	if err != nil {
		return nil, err
	}
	if err := checkKeyPEM(keyPEM); err != nil {
		return nil, err
	}
	pair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		if strings.Contains(err.Error(), "does not match") {
			return nil, certErr(CertErrMismatch, "the private key does not belong to the certificate")
		}
		return nil, certErr(CertErrBadKey, "the private key cannot be read")
	}

	leaf := chain[0]
	if err := checkKeyStrength(pair.PrivateKey); err != nil {
		return nil, err
	}
	if !serverAuthAllowed(leaf) {
		return nil, certErr(CertErrNotServer, "the certificate is not for a TLS server")
	}
	for i, cert := range chain {
		which := "the certificate"
		if i > 0 {
			which = fmt.Sprintf("intermediate %d (%s)", i, cert.Subject.CommonName)
		}
		if now.Before(cert.NotBefore) {
			return nil, certErr(CertErrNotYetValid, "%s is not valid until %s", which, cert.NotBefore.UTC().Format(time.RFC3339))
		}
		if now.After(cert.NotAfter) {
			return nil, certErr(CertErrExpired, "%s expired on %s", which, cert.NotAfter.UTC().Format(time.RFC3339))
		}
		if i+1 < len(chain) {
			if err := cert.CheckSignatureFrom(chain[i+1]); err != nil {
				return nil, certErr(CertErrBadChain,
					"certificate %d is not signed by certificate %d; put the leaf first, then each issuer in order", i+1, i+2)
			}
		}
	}

	info := describeCert(chain)
	if leaf.NotAfter.Sub(now) < certExpiryWarning {
		info.Warnings = append(info.Warnings, CertWarnExpiresSoon)
	}
	for _, name := range id.DNS {
		if leaf.VerifyHostname(name) != nil {
			info.Uncovered = append(info.Uncovered, name)
		}
	}
	for _, ip := range id.IPs {
		if ip.IsLoopback() {
			continue
		}
		if leaf.VerifyHostname(ip.String()) != nil {
			info.Uncovered = append(info.Uncovered, ip.String())
		}
	}
	// localhost is in the identity for the server's own loopback calls, which
	// do not check the certificate. A public certificate never names it, so
	// it is not worth a warning.
	info.Uncovered = without(info.Uncovered, "localhost")
	if len(info.Uncovered) > 0 {
		info.Warnings = append(info.Warnings, CertWarnNameMismatch)
	}
	if len(chain) == 1 && bytes.Equal(leaf.RawIssuer, leaf.RawSubject) &&
		leaf.CheckSignature(leaf.SignatureAlgorithm, leaf.RawTBSCertificate, leaf.Signature) == nil {
		info.Warnings = append(info.Warnings, CertWarnSelfSigned)
	}
	return info, nil
}

func without(list []string, drop string) []string {
	out := list[:0]
	for _, item := range list {
		if item != drop {
			out = append(out, item)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// parseCertChain reads every CERTIFICATE block, leaf first. Anything else in
// the certificate field is refused, so a key pasted into the wrong box is not
// stored as if it were a certificate.
func parseCertChain(data []byte) ([]*x509.Certificate, error) {
	var chain []*x509.Certificate
	rest := data
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			if strings.Contains(block.Type, "PRIVATE KEY") {
				return nil, certErr(CertErrBadCert, "the certificate field holds a private key")
			}
			return nil, certErr(CertErrBadCert, "unexpected %q block in the certificate", block.Type)
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, certErr(CertErrBadCert, "certificate %d cannot be parsed", len(chain)+1)
		}
		chain = append(chain, cert)
	}
	if len(chain) == 0 {
		return nil, certErr(CertErrNoCertificate, "no PEM certificate found")
	}
	if len(bytes.TrimSpace(rest)) > 0 {
		return nil, certErr(CertErrBadCert, "the certificate has text after its last PEM block")
	}
	return chain, nil
}

// checkKeyPEM looks at the key's PEM framing before tls.X509KeyPair does, so
// the common mistakes get a message of their own.
func checkKeyPEM(data []byte) error {
	block, rest := pem.Decode(data)
	// openssl ecparam -genkey writes the curve's parameters ahead of the key.
	for block != nil && block.Type == "EC PARAMETERS" {
		block, rest = pem.Decode(rest)
	}
	if block == nil {
		return certErr(CertErrNoKey, "no PEM private key found")
	}
	if block.Type == "ENCRYPTED PRIVATE KEY" || block.Headers["Proc-Type"] != "" {
		return certErr(CertErrEncryptedKey, "the private key is encrypted; upload it without a passphrase")
	}
	if !strings.HasSuffix(block.Type, "PRIVATE KEY") {
		return certErr(CertErrNoKey, "the key field holds a %q block, not a private key", block.Type)
	}
	return nil
}

func checkKeyStrength(key any) error {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		if k.N.BitLen() < 2048 {
			return certErr(CertErrWeakKey, "an RSA key of %d bits is too short; use 2048 or more", k.N.BitLen())
		}
	case *ecdsa.PrivateKey:
		if k.Curve.Params().BitSize < 256 {
			return certErr(CertErrWeakKey, "an EC key of %d bits is too short", k.Curve.Params().BitSize)
		}
	case ed25519.PrivateKey:
	default:
		return certErr(CertErrBadKey, "unsupported key type")
	}
	return nil
}

func serverAuthAllowed(cert *x509.Certificate) bool {
	if len(cert.ExtKeyUsage) == 0 {
		return true
	}
	for _, usage := range cert.ExtKeyUsage {
		if usage == x509.ExtKeyUsageServerAuth || usage == x509.ExtKeyUsageAny {
			return true
		}
	}
	return false
}

func keyType(cert *x509.Certificate) string {
	switch k := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		return fmt.Sprintf("RSA %d", k.N.BitLen())
	case *ecdsa.PublicKey:
		return "ECDSA " + k.Curve.Params().Name
	case ed25519.PublicKey:
		return "Ed25519"
	}
	return cert.PublicKeyAlgorithm.String()
}

func describeCert(chain []*x509.Certificate) *CertInfo {
	leaf := chain[0]
	sum := sha256.Sum256(leaf.Raw)
	ips := make([]string, 0, len(leaf.IPAddresses))
	for _, ip := range leaf.IPAddresses {
		ips = append(ips, ip.String())
	}
	return &CertInfo{
		Subject:     leaf.Subject.String(),
		Issuer:      leaf.Issuer.String(),
		DNSNames:    append([]string{}, leaf.DNSNames...),
		IPAddresses: ips,
		NotBefore:   leaf.NotBefore,
		NotAfter:    leaf.NotAfter,
		Fingerprint: strings.ToUpper(hex.EncodeToString(sum[:])),
		KeyType:     keyType(leaf),
		ChainLength: len(chain),
	}
}

// DescribeCertFile describes the certificate in a PEM file, for the page to
// show what is being served.
func DescribeCertFile(path string, id CertIdentity) (*CertInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	chain, err := parseCertChain(data)
	if err != nil {
		return nil, err
	}
	info := describeCert(chain)
	if time.Until(chain[0].NotAfter) < certExpiryWarning {
		info.Warnings = append(info.Warnings, CertWarnExpiresSoon)
	}
	for _, name := range id.DNS {
		if name != "localhost" && chain[0].VerifyHostname(name) != nil {
			info.Uncovered = append(info.Uncovered, name)
		}
	}
	for _, ip := range id.IPs {
		if !ip.IsLoopback() && chain[0].VerifyHostname(ip.String()) != nil {
			info.Uncovered = append(info.Uncovered, ip.String())
		}
	}
	if len(info.Uncovered) > 0 {
		info.Warnings = append(info.Warnings, CertWarnNameMismatch)
	}
	return info, nil
}

// WriteCertPair stores a validated pair at the uploaded paths. The key is
// written 0600 from the start, and each file lands by rename, so a reader
// never sees half of one. A process that dies between the two renames leaves
// a key that does not match its certificate, and LoadServingCert then falls
// back to the generated pair rather than failing to listen.
func WriteCertPair(certPath, keyPath string, certPEM, keyPEM []byte) error {
	for _, dir := range []string{filepath.Dir(keyPath), filepath.Dir(certPath)} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	if err := atomicfile.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		return err
	}
	return atomicfile.WriteFile(certPath, certPEM, 0o644)
}

// ServingCert is the certificate the HTTPS listener and the VNC TLS layer
// hand out. It is swapped whole, so a certificate can be replaced without a
// restart: connections already open keep the one they started with, and the
// next handshake gets the new one.
type ServingCert struct {
	cert atomic.Pointer[tls.Certificate]
}

// ActiveCert is the process's serving certificate.
var ActiveCert = &ServingCert{}

// Load reads a pair from disk and puts it in force.
func (s *ServingCert) Load(certPath, keyPath string) error {
	pair, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return err
	}
	s.cert.Store(&pair)
	return nil
}

// Current is the certificate in force, or nil before the first Load.
func (s *ServingCert) Current() *tls.Certificate {
	return s.cert.Load()
}

// GetCertificate is the tls.Config callback.
func (s *ServingCert) GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	if cert := s.cert.Load(); cert != nil {
		return cert, nil
	}
	return nil, errors.New("no certificate loaded")
}

// LoadServingCert puts the configured pair in force, and falls back to the
// generated one when the configured pair cannot be read. A damaged upload
// must not leave the board without HTTPS, because HTTPS is how the operator
// would reach the page that repairs it.
func LoadServingCert(certPath, keyPath string) error {
	err := ActiveCert.Load(certPath, keyPath)
	if err == nil {
		return nil
	}
	if certPath == CertFile && keyPath == KeyFile {
		return err
	}
	if fallbackErr := EnsureCert(); fallbackErr != nil {
		return fmt.Errorf("configured certificate: %v; generated certificate: %w", err, fallbackErr)
	}
	if fallbackErr := ActiveCert.Load(CertFile, KeyFile); fallbackErr != nil {
		return fmt.Errorf("configured certificate: %v; generated certificate: %w", err, fallbackErr)
	}
	return fmt.Errorf("configured certificate %s unusable, serving the generated one: %w", certPath, err)
}
