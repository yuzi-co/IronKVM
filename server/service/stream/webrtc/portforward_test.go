package webrtc

import (
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pion/webrtc/v4"

	"NanoKVM-Server/config"
)

// The tests never read server.yaml: they start from "nothing configured" and
// set what each case needs.
func init() {
	forwardSettings = func() config.WebRTC { return config.WebRTC{} }
}

func useSettings(t *testing.T, settings config.WebRTC) {
	t.Helper()

	old := forwardSettings
	forwardSettings = func() config.WebRTC { return settings }
	t.Cleanup(func() { forwardSettings = old })
}

// freeUDPPort returns a port nothing listens on at the moment.
func freeUDPPort(t *testing.T) int {
	t.Helper()

	conn, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		t.Skipf("no UDP: %v", err)
	}
	defer func() { _ = conn.Close() }()

	return conn.LocalAddr().(*net.UDPAddr).Port
}

// resetMuxes closes what a test opened, so that its port is free again.
func resetMuxes(t *testing.T) {
	t.Helper()

	t.Cleanup(func() {
		muxMutex.Lock()
		defer muxMutex.Unlock()
		for port, mux := range muxByPort {
			_ = mux.Close()
			delete(muxByPort, port)
		}
		muxActive = 0
	})
}

func TestNothingConfiguredChangesNothing(t *testing.T) {
	if forward := currentPortForward(); forward != nil {
		t.Fatalf("a port forward without a setting: %+v", forward)
	}
	if ActivePort() != 0 {
		t.Fatalf("active port = %d", ActivePort())
	}
}

func TestPublicAddressWithoutPortIsRefused(t *testing.T) {
	if err := ValidatePortForward(config.WebRTC{PublicAddress: "kvm.example.com"}); err == nil {
		t.Fatal("a public address with no port was accepted")
	}
}

func TestValidatePortForward(t *testing.T) {
	good := []config.WebRTC{
		{},
		{UDPPort: 58000},
		{UDPPort: 58000, PublicAddress: "203.0.113.7"},
		{UDPPort: 58000, PublicAddress: "2001:db8::7"},
		{UDPPort: 58000, PublicAddress: "kvm.example.com"},
		{UDPPort: 58000, PublicAddress: "my-kvm.dyndns.org"},
	}
	for _, c := range good {
		if err := ValidatePortForward(c, 80, 443); err != nil {
			t.Errorf("%+v rejected: %v", c, err)
		}
	}

	bad := map[string]config.WebRTC{
		"low port":        {UDPPort: 80},
		"high port":       {UDPPort: 70000},
		"negative":        {UDPPort: -1},
		"web server port": {UDPPort: 443},
		"loopback":        {UDPPort: 58000, PublicAddress: "127.0.0.1"},
		"unspecified":     {UDPPort: 58000, PublicAddress: "0.0.0.0"},
		"multicast":       {UDPPort: 58000, PublicAddress: "224.0.0.1"},
		"link-local":      {UDPPort: 58000, PublicAddress: "169.254.1.1"},
		"half an address": {UDPPort: 58000, PublicAddress: "203.0.113"},
		"bad v6":          {UDPPort: 58000, PublicAddress: "2001:db8::g"},
		"url":             {UDPPort: 58000, PublicAddress: "https://kvm.example.com"},
		"port suffix":     {UDPPort: 58000, PublicAddress: "kvm.example.com:58000"},
		"space":           {UDPPort: 58000, PublicAddress: "kvm example.com"},
		"leading dash":    {UDPPort: 58000, PublicAddress: "-kvm.example.com"},
		"too long":        {UDPPort: 58000, PublicAddress: strings.Repeat("a", 254)},
	}
	for name, c := range bad {
		if err := ValidatePortForward(c, 80, 443); err == nil {
			t.Errorf("%s: %+v accepted", name, c)
		}
	}
}

func TestResolvePublicAddress(t *testing.T) {
	old := lookupIPs
	t.Cleanup(func() {
		lookupIPs = old
		resolveMutex.Lock()
		resolved = map[string]resolvedAddress{}
		resolveMutex.Unlock()
	})

	calls := 0
	answer := []net.IPAddr{
		{IP: net.ParseIP("203.0.113.7")},
		{IP: net.ParseIP("127.0.0.1")},
		{IP: net.ParseIP("::ffff:198.51.100.2")},
	}
	var failure error
	lookupIPs = func(string) ([]net.IPAddr, error) {
		calls++
		if failure != nil {
			return nil, failure
		}
		return answer, nil
	}

	if got := ResolvePublicAddress("203.0.113.9"); len(got) != 1 || got[0] != "203.0.113.9" || calls != 0 {
		t.Fatalf("an IP was looked up: %v (%d lookups)", got, calls)
	}
	if got := ResolvePublicAddress(""); got != nil {
		t.Fatalf("empty = %v", got)
	}

	got := ResolvePublicAddress("kvm.example.com")
	if len(got) != 2 || got[0] != "203.0.113.7" || got[1] != "198.51.100.2" {
		t.Fatalf("resolved = %v", got)
	}
	ResolvePublicAddress("kvm.example.com")
	if calls != 1 {
		t.Fatalf("the answer was not kept: %d lookups", calls)
	}

	// An expired answer is looked up again, and a failed lookup serves the old one.
	resolveMutex.Lock()
	entry := resolved["kvm.example.com"]
	entry.expires = time.Now().Add(-time.Second)
	resolved["kvm.example.com"] = entry
	resolveMutex.Unlock()
	failure = net.UnknownNetworkError("down")
	if got := ResolvePublicAddress("kvm.example.com"); len(got) != 2 || calls != 2 {
		t.Fatalf("after a failed lookup: %v (%d lookups)", got, calls)
	}
	if got := ResolvePublicAddress("other.example.com"); got != nil {
		t.Fatalf("a name that never resolved gave %v", got)
	}
}

func TestBindReportsAPortThatIsTaken(t *testing.T) {
	resetMuxes(t)

	taken, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		t.Skipf("no UDP: %v", err)
	}
	defer func() { _ = taken.Close() }()
	port := taken.LocalAddr().(*net.UDPAddr).Port

	err = BindPortForward(config.WebRTC{UDPPort: port})
	if err == nil || !strings.Contains(err.Error(), "cannot listen on UDP port") {
		t.Fatalf("err = %v", err)
	}

	// A viewer meanwhile gets the default behaviour, not a failure.
	useSettings(t, config.WebRTC{UDPPort: port})
	if forward := currentPortForward(); forward != nil {
		t.Fatalf("a port that could not be opened produced %+v", forward)
	}
}

func gatheredCandidates(t *testing.T, pc *webrtc.PeerConnection) []string {
	t.Helper()

	if _, err := pc.CreateDataChannel("probe", nil); err != nil {
		t.Fatal(err)
	}
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := pc.SetLocalDescription(offer); err != nil {
		t.Fatal(err)
	}

	select {
	case <-webrtc.GatheringCompletePromise(pc):
	case <-time.After(10 * time.Second):
		t.Fatal("candidate gathering did not finish")
	}

	var candidates []string
	for _, line := range strings.Split(pc.LocalDescription().SDP, "\n") {
		if strings.HasPrefix(line, "a=candidate:") {
			candidates = append(candidates, strings.TrimSpace(line))
		}
	}

	return candidates
}

// candidatePort and candidateIP read "a=candidate:<f> <comp> udp <prio> <ip> <port> typ host".
func candidatePort(line string) string { return strings.Fields(line)[5] }
func candidateIP(line string) string   { return strings.Fields(line)[4] }

func TestViewersShareTheFixedPortAndOfferThePublicAddress(t *testing.T) {
	resetMuxes(t)
	port := freeUDPPort(t)
	useSettings(t, config.WebRTC{UDPPort: port, PublicAddress: "203.0.113.7"})

	if err := BindPortForward(forwardSettings()); err != nil {
		t.Fatal(err)
	}
	SetPortForward(forwardSettings())
	if ActivePort() != port {
		t.Fatalf("active port = %d, want %d", ActivePort(), port)
	}

	// Two viewers, one port.
	var all [][]string
	for i := 0; i < 2; i++ {
		pc, udp, err := createPeerConnection(nil)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = pc.Close() }()
		if udp != nil {
			t.Fatal("a viewer on the shared port got its own batching sockets")
		}
		all = append(all, gatheredCandidates(t, pc))
	}

	portText := strconv.Itoa(port)
	for i, candidates := range all {
		if len(candidates) == 0 {
			t.Skip("no usable network interface to gather candidates on")
		}
		sawPublic := false
		for _, c := range candidates {
			if candidatePort(c) != portText {
				t.Errorf("viewer %d offered a candidate off the fixed port: %s", i, c)
			}
			if candidateIP(c) == "203.0.113.7" {
				sawPublic = true
			}
		}
		if !sawPublic {
			t.Errorf("viewer %d was not offered the public address: %v", i, candidates)
		}
	}
}

func TestWithoutASettingTheCandidatesStayEphemeral(t *testing.T) {
	pc, _, err := createPeerConnection(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = pc.Close() }()

	// One socket per address and component: the same address on the RTP and
	// the RTCP component is one socket, so compare addresses.
	portOf := map[string]string{}
	for _, c := range gatheredCandidates(t, pc) {
		portOf[candidateIP(c)] = candidatePort(c)
	}
	if len(portOf) < 2 {
		t.Skip("fewer than two addresses to compare")
	}
	ports := map[string]bool{}
	for _, port := range portOf {
		ports[port] = true
	}
	if len(ports) < 2 {
		t.Errorf("default candidates share a port: %v", portOf)
	}
}
