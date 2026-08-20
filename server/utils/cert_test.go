package utils

import (
	"crypto/x509"
	"encoding/pem"
	"net"
	"os"
	"path/filepath"
	"testing"
)

// The certificate this device generates for itself decides whether the browser
// will open a websocket to it, and the websocket is how every keystroke and
// every mouse movement travels. A certificate the browser rejects therefore
// does not degrade the product, it removes the input.
//
// It does that quietly. Chrome asks the operator about an untrusted certificate
// when it loads the page and does not ask when it opens a websocket to the same
// origin: it refuses and reports nothing to the page. So the UI appears, the
// REST calls work, the video plays, and the keyboard is dead.
//
// The generator used to name "localhost" and the two loopback addresses and
// nothing else, which cannot match any URL that reaches a network KVM. Enabling
// HTTPS from the settings page was therefore enough to take the input away, on
// every device, every time.

func ips(t *testing.T, in ...string) []net.IP {
	t.Helper()

	out := make([]net.IP, 0, len(in))
	for _, s := range in {
		ip := net.ParseIP(s)
		if ip == nil {
			t.Fatalf("bad test address %q", s)
		}
		out = append(out, ip)
	}

	return out
}

func hasDNS(id certIdentity, want string) bool {
	for _, got := range id.DNS {
		if got == want {
			return true
		}
	}

	return false
}

func hasIP(id certIdentity, want string) bool {
	target := net.ParseIP(want)
	for _, got := range id.IPs {
		if got.Equal(target) {
			return true
		}
	}

	return false
}

func TestCertNamesCoverTheAddressTheDeviceIsReachedOn(t *testing.T) {
	id := certNames("nanokvm", ips(t, "10.0.0.222"))

	if !hasIP(id, "10.0.0.222") {
		t.Fatalf("the device's own address is missing from %v", id.IPs)
	}
	if !hasDNS(id, "nanokvm") {
		t.Errorf("the hostname is missing from %v", id.DNS)
	}
	if !hasDNS(id, "nanokvm.local") {
		t.Errorf("the mDNS name is missing from %v; avahi answers on it", id.DNS)
	}
}

func TestCertNamesKeepLoopbackWorking(t *testing.T) {
	// The server talks to itself over loopback, and so does the PicoClaw
	// runtime. Naming the real address must not cost the old one.
	id := certNames("nanokvm", ips(t, "10.0.0.222"))

	if !hasDNS(id, "localhost") {
		t.Errorf("localhost is missing from %v", id.DNS)
	}
	if !hasIP(id, "127.0.0.1") {
		t.Errorf("127.0.0.1 is missing from %v", id.IPs)
	}
	if !hasIP(id, "::1") {
		t.Errorf("::1 is missing from %v", id.IPs)
	}
}

func TestCertNamesCoverEveryInterface(t *testing.T) {
	// A board can be on ethernet and wifi at once, and the operator may reach
	// it on either.
	id := certNames("nanokvm", ips(t, "10.0.0.222", "192.168.1.50"))

	if !hasIP(id, "10.0.0.222") || !hasIP(id, "192.168.1.50") {
		t.Fatalf("both interfaces must be covered, got %v", id.IPs)
	}
}

func TestCertNamesDropsWhatCannotBeReached(t *testing.T) {
	// An IPv6 link-local address is scoped to one interface and changes with
	// the hardware. It cannot be typed into a browser usefully, and putting it
	// in the certificate only makes the certificate go stale sooner.
	id := certNames("nanokvm", ips(t, "10.0.0.222", "fe80::1"))

	if hasIP(id, "fe80::1") {
		t.Errorf("a link-local address was included: %v", id.IPs)
	}
}

func TestCertNamesDoesNotRepeatItself(t *testing.T) {
	id := certNames("nanokvm", ips(t, "127.0.0.1", "10.0.0.222", "10.0.0.222"))

	seen := map[string]int{}
	for _, ip := range id.IPs {
		seen[ip.String()]++
	}
	for addr, n := range seen {
		if n > 1 {
			t.Errorf("%s appears %d times in %v", addr, n, id.IPs)
		}
	}
}

func TestCertNamesSurvivesAnUnsetHostname(t *testing.T) {
	// os.Hostname can fail. A bare ".local" is not a name and a certificate
	// carrying one is malformed.
	id := certNames("", ips(t, "10.0.0.222"))

	for _, name := range id.DNS {
		if name == "" || name == ".local" {
			t.Errorf("an empty hostname produced the name %q", name)
		}
	}
	if !hasDNS(id, "localhost") {
		t.Errorf("localhost must survive an unset hostname, got %v", id.DNS)
	}
}

// writeCert generates a certificate for the given identity and returns its path.
func writeCert(t *testing.T, id certIdentity) string {
	t.Helper()

	dir := t.TempDir()
	crt := filepath.Join(dir, "server.crt")
	key := filepath.Join(dir, "server.key")

	if err := generateCertTo(crt, key, id); err != nil {
		t.Fatalf("generateCertTo: %v", err)
	}

	return crt
}

func TestCertCoversAcceptsACertificateThatStillFits(t *testing.T) {
	// This is the case that must not regenerate. Regenerating voids whatever
	// trust the operator granted the last certificate, so a restart that
	// changes nothing must leave the file alone.
	id := certNames("nanokvm", ips(t, "10.0.0.222"))

	if !CertCovers(writeCert(t, id), id) {
		t.Fatal("a certificate covering the current identity was reported stale")
	}
}

func TestCertCoversRejectsTheLocalhostOnlyCertificate(t *testing.T) {
	// The exact certificate the product shipped, against a device that is
	// reached on its address. This is the reported fault.
	shipped := certIdentity{
		DNS: []string{"localhost"},
		IPs: ips(t, "127.0.0.1", "::1"),
	}
	current := certNames("nanokvm", ips(t, "10.0.0.222"))

	if CertCovers(writeCert(t, shipped), current) {
		t.Fatal("the localhost-only certificate was accepted for a device reached on 10.0.0.222")
	}
}

func TestCertCoversRejectsACertificateLeftBehindByDHCP(t *testing.T) {
	old := certNames("nanokvm", ips(t, "10.0.0.222"))
	moved := certNames("nanokvm", ips(t, "10.0.0.99"))

	if CertCovers(writeCert(t, old), moved) {
		t.Fatal("a certificate for the previous address was accepted after the address changed")
	}
}

func TestCertCoversRejectsWhatItCannotRead(t *testing.T) {
	id := certNames("nanokvm", ips(t, "10.0.0.222"))

	if CertCovers(filepath.Join(t.TempDir(), "absent.crt"), id) {
		t.Error("a missing certificate was reported as covering")
	}

	junk := filepath.Join(t.TempDir(), "junk.crt")
	if err := os.WriteFile(junk, []byte("not a certificate"), 0o600); err != nil {
		t.Fatal(err)
	}
	if CertCovers(junk, id) {
		t.Error("an unparseable certificate was reported as covering")
	}
}

func TestGeneratedCertificateIsUsableAsAServerCertificate(t *testing.T) {
	id := certNames("nanokvm", ips(t, "10.0.0.222"))
	path := writeCert(t, id)

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		t.Fatal("the generated file is not PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	// The browser checks the name against the URL. If this fails, the operator
	// is told the certificate is for the wrong site, which is the error that
	// started all of this.
	if err := cert.VerifyHostname("10.0.0.222"); err != nil {
		t.Errorf("the certificate does not answer for its own address: %v", err)
	}
	if err := cert.VerifyHostname("nanokvm.local"); err != nil {
		t.Errorf("the certificate does not answer for its mDNS name: %v", err)
	}

	var serverAuth bool
	for _, use := range cert.ExtKeyUsage {
		if use == x509.ExtKeyUsageServerAuth {
			serverAuth = true
		}
	}
	if !serverAuth {
		t.Error("the certificate is not marked for server authentication")
	}
}

func TestEnsureCertLeavesAGoodCertificateAlone(t *testing.T) {
	dir := t.TempDir()
	crt := filepath.Join(dir, "server.crt")
	key := filepath.Join(dir, "server.key")
	id := certNames("nanokvm", ips(t, "10.0.0.222"))

	if err := generateCertTo(crt, key, id); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(crt)
	if err != nil {
		t.Fatal(err)
	}

	replaced, err := ensureCert(crt, key, id)
	if err != nil {
		t.Fatalf("ensureCert: %v", err)
	}
	if replaced {
		t.Error("a certificate that still fits was replaced, which voids the operator's trust")
	}

	after, err := os.ReadFile(crt)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("the certificate file changed when nothing needed to change")
	}
}

func TestEnsureCertReplacesOneThatNoLongerFits(t *testing.T) {
	dir := t.TempDir()
	crt := filepath.Join(dir, "server.crt")
	key := filepath.Join(dir, "server.key")

	shipped := certIdentity{DNS: []string{"localhost"}, IPs: ips(t, "127.0.0.1", "::1")}
	if err := generateCertTo(crt, key, shipped); err != nil {
		t.Fatal(err)
	}

	current := certNames("nanokvm", ips(t, "10.0.0.222"))
	replaced, err := ensureCert(crt, key, current)
	if err != nil {
		t.Fatalf("ensureCert: %v", err)
	}
	if !replaced {
		t.Fatal("a certificate that cannot match the device's address was kept")
	}
	if !CertCovers(crt, current) {
		t.Error("the replacement still does not cover the current identity")
	}
}

func TestEnsureCertBuildsOneFromNothing(t *testing.T) {
	dir := t.TempDir()
	crt := filepath.Join(dir, "server.crt")
	key := filepath.Join(dir, "server.key")
	id := certNames("nanokvm", ips(t, "10.0.0.222"))

	replaced, err := ensureCert(crt, key, id)
	if err != nil {
		t.Fatalf("ensureCert: %v", err)
	}
	if !replaced {
		t.Error("no certificate existed and none was reported as written")
	}
	if !CertCovers(crt, id) {
		t.Error("the generated certificate does not cover the current identity")
	}
}

func TestGeneratedKeyIsNotReadableByOtherUsers(t *testing.T) {
	dir := t.TempDir()
	crt := filepath.Join(dir, "server.crt")
	key := filepath.Join(dir, "server.key")

	if err := generateCertTo(crt, key, certNames("nanokvm", ips(t, "10.0.0.222"))); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(key)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode&0o077 != 0 {
		t.Errorf("the private key is mode %04o, which lets other users read it", mode)
	}
}
