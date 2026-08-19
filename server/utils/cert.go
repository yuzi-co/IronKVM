package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	CertFile = "/etc/kvm/server.crt"
	KeyFile  = "/etc/kvm/server.key"

	certValidFor = time.Hour * 24 * 365 * 10
)

// certIdentity is every name and address a certificate for this device has to
// answer for.
//
// It has to answer for the address the operator typed. The browser compares the
// URL against the certificate, and a certificate that names only "localhost"
// matches no URL that reaches a network KVM. That was the whole of this file
// until now, and enabling HTTPS from the settings page was therefore enough to
// take the keyboard away on every device.
//
// The failure is quiet, which is why it was worth this much code. Keyboard and
// mouse travel over a websocket (/api/ws). A browser asks the operator about an
// untrusted certificate when it loads a page and does NOT ask when it opens a
// websocket to the same origin: it refuses, and reports nothing to the page. So
// the UI appears, the REST calls answer, the video plays, and no key reaches the
// managed host.
type certIdentity struct {
	DNS []string
	IPs []net.IP
}

// certNames builds that set from the device's hostname and its own addresses.
//
// Loopback stays in it. The server talks to itself, and so does the on-device
// PicoClaw runtime, so naming the real address must not cost the old one.
func certNames(hostname string, addrs []net.IP) certIdentity {
	id := certIdentity{
		DNS: []string{"localhost"},
		IPs: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	if hostname != "" && hostname != "localhost" {
		// avahi answers on the mDNS name, and an operator who reached the board
		// that way sees the URL they typed compared against this list.
		id.DNS = append(id.DNS, hostname, hostname+".local")
	}

	for _, ip := range addrs {
		if ip == nil || ip.IsLoopback() {
			continue
		}

		// A link-local address is scoped to one interface and moves with the
		// hardware. It cannot be usefully typed into a browser, and naming it
		// only makes the certificate go stale sooner.
		if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			continue
		}

		id.IPs = append(id.IPs, ip)
	}

	id.dedupe()

	return id
}

func (id *certIdentity) dedupe() {
	seenDNS := make(map[string]bool, len(id.DNS))
	dns := id.DNS[:0]
	for _, name := range id.DNS {
		if name == "" || seenDNS[name] {
			continue
		}
		seenDNS[name] = true
		dns = append(dns, name)
	}
	id.DNS = dns

	seenIP := make(map[string]bool, len(id.IPs))
	ips := id.IPs[:0]
	for _, ip := range id.IPs {
		key := ip.String()
		if seenIP[key] {
			continue
		}
		seenIP[key] = true
		ips = append(ips, ip)
	}
	id.IPs = ips
}

// localIPs collects the unicast addresses of every interface that is up.
func localIPs() []net.IP {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Warnf("failed to list interfaces for the certificate: %v", err)
		return nil
	}

	var out []net.IP

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				out = append(out, v.IP)
			case *net.IPAddr:
				out = append(out, v.IP)
			}
		}
	}

	return out
}

// CurrentCertIdentity is what a certificate generated right now has to cover.
func CurrentCertIdentity() certIdentity {
	hostname, err := os.Hostname()
	if err != nil {
		// A board with no hostname still has addresses, and those are what the
		// operator types anyway.
		log.Warnf("failed to read the hostname for the certificate: %v", err)
		hostname = ""
	}

	return certNames(hostname, localIPs())
}

// CertCovers reports whether the certificate on disk already answers for every
// name and address in id.
//
// A certificate that still fits must be left alone. Generating a new one voids
// whatever trust the operator granted the old one, so a restart that changes
// nothing has to change nothing.
func CertCovers(certPath string, id certIdentity) bool {
	raw, err := os.ReadFile(certPath)
	if err != nil {
		return false
	}

	block, _ := pem.Decode(raw)
	if block == nil {
		return false
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}

	if time.Now().After(cert.NotAfter) {
		return false
	}

	for _, name := range id.DNS {
		if cert.VerifyHostname(name) != nil {
			return false
		}
	}

	for _, ip := range id.IPs {
		if cert.VerifyHostname(ip.String()) != nil {
			return false
		}
	}

	return true
}

// ensureCert writes a new certificate when the one on disk cannot answer for
// this device, and reports whether it wrote one.
func ensureCert(certPath, keyPath string, id certIdentity) (bool, error) {
	if CertCovers(certPath, id) {
		return false, nil
	}

	if err := generateCertTo(certPath, keyPath, id); err != nil {
		return false, err
	}

	return true, nil
}

// EnsureCert keeps the shipped certificate paths answering for this device.
//
// Call it before the HTTPS listener starts. A board that moved to another
// address under DHCP would otherwise keep serving a certificate for the address
// it used to have, which fails in exactly the same silent way as the original
// fault and needs no operator mistake to happen.
func EnsureCert() {
	id := CurrentCertIdentity()

	replaced, err := ensureCert(CertFile, KeyFile, id)
	if err != nil {
		log.Errorf("failed to refresh the TLS certificate: %v", err)
		return
	}

	if replaced {
		log.Warnf("the TLS certificate did not cover %v, so a new one was generated; "+
			"a browser that trusted the old certificate has to be told about this one",
			append(id.DNS, ipStrings(id.IPs)...))
	}
}

func ipStrings(ips []net.IP) []string {
	out := make([]string, 0, len(ips))
	for _, ip := range ips {
		out = append(out, ip.String())
	}

	return out
}

// GenerateCert writes a fresh certificate covering this device, at the shipped
// paths. The settings page calls it when HTTPS is switched on.
func GenerateCert() error {
	return generateCertTo(CertFile, KeyFile, CurrentCertIdentity())
}

func generateCertTo(certFile, keyFile string, id certIdentity) error {
	// P-256 rather than RSA-2048. This now runs at boot when the address has
	// moved, on a single core that is also starting the video pipeline, and an
	// RSA key costs seconds there where this costs none. Every browser has
	// accepted P-256 for over a decade.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Errorf("failed to generate the private key: %v", err)
		return err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Errorf("failed to generate serial number: %v", err)
		return err
	}

	commonName := "NanoKVM"
	if len(id.DNS) > 0 {
		commonName = id.DNS[len(id.DNS)-1]
	}

	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: commonName, Organization: []string{"NanoKVM"}},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(certValidFor),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		// Self-signed and installed directly, so it is its own issuer. A trust
		// store will not accept a leaf that does not say so.
		IsCA:        true,
		DNSNames:    id.DNS,
		IPAddresses: id.IPs,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Errorf("failed to create certificate: %v", err)
		return err
	}

	// The key is written first and with the tight mode from the start. Creating
	// it readable and narrowing it afterwards leaves a window in which any user
	// on the device can take it.
	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		log.Errorf("failed to create %s: %v", keyFile, err)
		return err
	}

	privateBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		_ = keyOut.Close()
		log.Errorf("failed to marshal private key: %v", err)
		return err
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateBytes}); err != nil {
		_ = keyOut.Close()
		log.Errorf("failed to encode %s: %v", keyFile, err)
		return err
	}

	_ = keyOut.Sync()
	_ = keyOut.Close()

	certOut, err := os.OpenFile(certFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		log.Errorf("failed to create %s: %v", certFile, err)
		return err
	}

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		_ = certOut.Close()
		log.Errorf("failed to encode %s: %v", certFile, err)
		return err
	}

	_ = certOut.Sync()
	_ = certOut.Close()

	log.Debugf("%s generated for %v %v", certFile, id.DNS, ipStrings(id.IPs))

	return nil
}
