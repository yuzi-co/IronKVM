package wireguard

import (
	"encoding/base64"
	"strconv"
	"strings"
	"testing"
)

// Keys for the tests: 32 bytes each, base64.
var (
	testPrivate = base64.StdEncoding.EncodeToString(fill(1))
	testPeerKey = base64.StdEncoding.EncodeToString(fill(2))
	testPSK     = base64.StdEncoding.EncodeToString(fill(3))
	testPeer2   = base64.StdEncoding.EncodeToString(fill(4))
)

func fill(b byte) []byte {
	key := make([]byte, 32)
	for i := range key {
		key[i] = b + byte(i)
	}
	return key
}

const sampleConf = `# exported by the VPN server
[Interface]
PrivateKey = %PRIV%
Address = 10.8.0.2/24, fd00:8::2/64
DNS = 10.8.0.1, lan.example
ListenPort = 51821
MTU = 1380
PostUp = iptables -A FORWARD -i wg0 -j ACCEPT

[Peer]
PublicKey = %PEER%
PresharedKey = %PSK%
Endpoint = vpn.example.com:51820
AllowedIPs = 10.8.0.0/24
AllowedIPs = 192.168.50.0/24 # the office
PersistentKeepalive = 25
`

func sample() string {
	r := strings.NewReplacer("%PRIV%", testPrivate, "%PEER%", testPeerKey, "%PSK%", testPSK)
	return r.Replace(sampleConf)
}

func TestParseConf(t *testing.T) {
	conf, warnings, err := ParseConf(sample())
	if err != nil {
		t.Fatal(err)
	}
	if conf.PrivateKey != testPrivate {
		t.Errorf("private key %q", conf.PrivateKey)
	}
	if got := strings.Join(conf.Addresses, ","); got != "10.8.0.2/24,fd00:8::2/64" {
		t.Errorf("addresses %s", got)
	}
	if got := strings.Join(conf.DNS, ","); got != "10.8.0.1,lan.example" {
		t.Errorf("dns %s", got)
	}
	if conf.ListenPort != 51821 || conf.MTU != 1380 {
		t.Errorf("port %d mtu %d", conf.ListenPort, conf.MTU)
	}
	if len(conf.Peers) != 1 {
		t.Fatalf("%d peers", len(conf.Peers))
	}
	p := conf.Peers[0]
	if p.PublicKey != testPeerKey || p.PresharedKey != testPSK {
		t.Errorf("peer keys %q %q", p.PublicKey, p.PresharedKey)
	}
	if p.Endpoint != "vpn.example.com:51820" || p.PersistentKeepalive != 25 {
		t.Errorf("endpoint %q keepalive %d", p.Endpoint, p.PersistentKeepalive)
	}
	if got := strings.Join(p.AllowedIPs, ","); got != "10.8.0.0/24,192.168.50.0/24" {
		t.Errorf("allowed %s", got)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "PostUp") {
		t.Errorf("warnings %v", warnings)
	}
}

func TestParseConfIsCaseInsensitiveAndAcceptsOff(t *testing.T) {
	text := "[interface]\nprivatekey=" + testPrivate + "\naddress=10.0.0.2\n[PEER]\npublickey=" + testPeerKey +
		"\nallowedips=0.0.0.0/0\npersistentkeepalive=off\n"
	conf, _, err := ParseConf(text)
	if err != nil {
		t.Fatal(err)
	}
	if conf.Addresses[0] != "10.0.0.2/32" {
		t.Errorf("a bare address is a /32: %v", conf.Addresses)
	}
	if conf.Peers[0].PersistentKeepalive != 0 {
		t.Errorf("off is 0")
	}
}

func TestParseConfRefuses(t *testing.T) {
	base := func(iface, peer string) string {
		return "[Interface]\nPrivateKey = " + testPrivate + "\nAddress = 10.0.0.2/32\n" + iface +
			"[Peer]\nPublicKey = " + testPeerKey + "\nAllowedIPs = 10.0.0.0/24\n" + peer
	}
	cases := map[string]string{
		"no interface":        "[Peer]\nPublicKey = " + testPeerKey + "\nAllowedIPs = 10.0.0.0/24\n",
		"no peer":             "[Interface]\nPrivateKey = " + testPrivate + "\nAddress = 10.0.0.2/32\n",
		"key before section":  "PrivateKey = " + testPrivate + "\n" + base("", ""),
		"unknown section":     base("", "[Router]\n"),
		"line without equals": base("Address\n", ""),
		"short key":           strings.Replace(base("", ""), testPrivate, base64.StdEncoding.EncodeToString(make([]byte, 31)), 1),
		"bad base64":          strings.Replace(base("", ""), testPeerKey, "not!base64", 1),
		"bad address":         base("Address = 10.0.0.300/24\n", ""),
		"bad allowed ip":      base("", "AllowedIPs = 10.0.0.0/33\n"),
		"endpoint no port":    base("", "Endpoint = vpn.example.com\n"),
		"endpoint port 0":     base("", "Endpoint = 1.2.3.4:0\n"),
		"endpoint big port":   base("", "Endpoint = 1.2.3.4:65536\n"),
		"endpoint bad host":   base("", "Endpoint = bad_host!:51820\n"),
		"keepalive range":     base("", "PersistentKeepalive = 65536\n"),
		"keepalive negative":  base("", "PersistentKeepalive = -1\n"),
		"listen port":         base("ListenPort = 70000\n", ""),
		"mtu":                 base("MTU = 100\n", ""),
		"bad dns":             base("DNS = 1.2.3.4.5.6, -bad-\n", ""),
		"duplicate key":       base("PrivateKey = "+testPrivate+"\n", ""),
	}
	for name, text := range cases {
		if _, _, err := ParseConf(text); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestParseConfSizeAndPeerLimits(t *testing.T) {
	if _, _, err := ParseConf(strings.Repeat("#", maxConfSize+1)); err == nil {
		t.Error("an oversized file was accepted")
	}
	var b strings.Builder
	b.WriteString("[Interface]\nPrivateKey = " + testPrivate + "\nAddress = 10.0.0.2/32\n")
	for i := 0; i <= maxPeers; i++ {
		key := fill(byte(10 + i))
		b.WriteString("[Peer]\nPublicKey = " + base64.StdEncoding.EncodeToString(key) + "\nAllowedIPs = 10.1." +
			strconv.Itoa(i) + ".0/24\n")
	}
	if _, _, err := ParseConf(b.String()); err == nil {
		t.Error("too many peers accepted")
	}
}

func validConfig() Config {
	return Config{
		PrivateKey: testPrivate,
		Addresses:  []string{"10.8.0.2/24"},
		Peers: []Peer{{
			PublicKey:  testPeerKey,
			Endpoint:   "203.0.113.5:51820",
			AllowedIPs: []string{"10.8.0.0/24"},
		}},
	}
}

func TestValidate(t *testing.T) {
	c := validConfig()
	c.Peers[0].AllowedIPs = []string{"10.8.0.7/24", "[::]/0"}
	c.Peers[0].Endpoint = "[2001:db8::1]:51820"
	if err := c.Validate(); err == nil {
		// "[::]/0" is not a prefix.
		t.Error("a bracketed prefix was accepted")
	}

	c = validConfig()
	c.Peers[0].AllowedIPs = []string{"10.8.0.7/24", "::/0"}
	c.Peers[0].Endpoint = "[2001:db8::1]:51820"
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if c.Peers[0].AllowedIPs[0] != "10.8.0.0/24" {
		t.Errorf("allowed IPs are masked: %v", c.Peers[0].AllowedIPs)
	}

	bad := map[string]func(c *Config){
		"no address":        func(c *Config) { c.Addresses = nil },
		"no peers":          func(c *Config) { c.Peers = nil },
		"no allowed ips":    func(c *Config) { c.Peers[0].AllowedIPs = nil },
		"same peer twice":   func(c *Config) { c.Peers = append(c.Peers, c.Peers[0]) },
		"peer is the board": func(c *Config) { c.Peers[0].PublicKey = mustPublic(t, testPrivate) },
		"overlapping routes": func(c *Config) {
			c.Peers = append(c.Peers, Peer{PublicKey: testPeer2, AllowedIPs: []string{"10.8.0.0/24"}})
		},
		"bad psk":          func(c *Config) { c.Peers[0].PresharedKey = "abc" },
		"no private key":   func(c *Config) { c.PrivateKey = "" },
		"zero private key": func(c *Config) { c.PrivateKey = base64.StdEncoding.EncodeToString(make([]byte, 32)) },
	}
	for name, change := range bad {
		c := validConfig()
		change(&c)
		if err := c.Validate(); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func mustPublic(t *testing.T, private string) string {
	t.Helper()
	pub, err := PublicKey(private)
	if err != nil {
		t.Fatal(err)
	}
	return pub
}

func TestDefaultRoute(t *testing.T) {
	c := validConfig()
	if c.DefaultRoute() {
		t.Error("a /24 is not the default route")
	}
	c.Peers[0].AllowedIPs = append(c.Peers[0].AllowedIPs, "0.0.0.0/0")
	if !c.DefaultRoute() {
		t.Error("0.0.0.0/0 is the default route")
	}
}
