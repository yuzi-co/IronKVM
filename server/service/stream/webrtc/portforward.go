package webrtc

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pion/ice/v4"
	"github.com/pion/webrtc/v4"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
)

// WebRTC through a router's port forward (ironkvm-dist#113).
//
// By default every viewer's ICE agent listens on its own ephemeral UDP port,
// which no router can forward. With webrtc.udpPort set, all viewers share one
// port through pion's UDP mux (ice.UDPMux), which routes each datagram to the
// right agent by the ICE username fragment. With webrtc.publicAddress set as
// well, the agents offer one more host candidate: that address on the same
// port, which is what a viewer outside the LAN can reach.
//
// With neither set, none of this runs and a peer connection is built exactly as
// it was.
//
// A shared socket cannot use the per-frame sendmmsg batching in udpbatch.go,
// which belongs to one viewer's sockets, so viewers on the fixed port send one
// datagram at a time. How much that costs at 60 fps with several viewers has
// to be measured on the board (#91).

// liveSettings is the setting in force once the API has changed it. Until then
// the one loaded from server.yaml applies.
var liveSettings atomic.Pointer[config.WebRTC]

// forwardSettings is a var so that the tests do not read server.yaml.
var forwardSettings = func() config.WebRTC {
	if live := liveSettings.Load(); live != nil {
		return *live
	}

	return config.GetInstance().WebRTC
}

// PortForwardSettings is the setting new viewers get.
func PortForwardSettings() config.WebRTC {
	return forwardSettings()
}

// publicAddressTTL is how long a resolved host name is kept. A dynamic DNS
// name changes rarely, and a connection should not wait on a lookup each time.
const publicAddressTTL = time.Minute

const resolveTimeout = 2 * time.Second

var (
	muxMutex  sync.Mutex
	muxByPort = map[int]*ice.UDPMuxDefault{}
	muxActive int

	resolveMutex sync.Mutex
	resolved     = map[string]resolvedAddress{}
)

type resolvedAddress struct {
	ips     []string
	expires time.Time
}

// listenFixedPort opens the shared socket. A var for the tests.
var listenFixedPort = func(port int) (net.PacketConn, error) {
	return net.ListenUDP("udp", &net.UDPAddr{Port: port})
}

// lookupIPs resolves a host name. A var for the tests.
var lookupIPs = func(host string) ([]net.IPAddr, error) {
	ctx, cancel := context.WithTimeout(context.Background(), resolveTimeout)
	defer cancel()

	return net.DefaultResolver.LookupIPAddr(ctx, host)
}

// BindPortForward opens the shared port now, so that a port that is taken is
// reported when the setting is saved and not when the first viewer connects.
func BindPortForward(settings config.WebRTC) error {
	if settings.UDPPort == 0 {
		return nil
	}

	_, err := muxForPort(settings.UDPPort)

	return err
}

// SetPortForward makes the setting the one new viewers get. A port change
// applies to viewers that connect afterwards. The socket of the port it
// replaces stays open until the server restarts: closing it would drop the
// viewers still on it, the person changing the setting among them.
func SetPortForward(settings config.WebRTC) {
	liveSettings.Store(&settings)

	muxMutex.Lock()
	defer muxMutex.Unlock()

	if settings.UDPPort == 0 {
		muxActive = 0
	} else if _, ok := muxByPort[settings.UDPPort]; ok {
		muxActive = settings.UDPPort
	}
}

// ActivePort is the shared port new viewers get, zero when there is none.
func ActivePort() int {
	muxMutex.Lock()
	defer muxMutex.Unlock()

	return muxActive
}

func muxForPort(port int) (*ice.UDPMuxDefault, error) {
	muxMutex.Lock()
	defer muxMutex.Unlock()

	if mux, ok := muxByPort[port]; ok && !mux.IsClosed() {
		muxActive = port
		return mux, nil
	}

	conn, err := listenFixedPort(port)
	if err != nil {
		return nil, fmt.Errorf("cannot listen on UDP port %d: %w", port, err)
	}

	mux := ice.NewUDPMuxDefault(ice.UDPMuxParams{UDPConn: conn})
	muxByPort[port] = mux
	muxActive = port
	log.Infof("webrtc: viewers share UDP port %d", port)

	return mux, nil
}

// portForward is what a peer connection takes from the settings.
type portForward struct {
	mux   ice.UDPMux
	rules []webrtc.ICEAddressRewriteRule
}

// currentPortForward returns nil unless a fixed port is configured and bound.
// A failure leaves the default behaviour in place: a viewer that cannot use
// the forward can still connect on the LAN.
func currentPortForward() *portForward {
	settings := forwardSettings()
	if settings.UDPPort == 0 {
		return nil
	}

	mux, err := muxForPort(settings.UDPPort)
	if err != nil {
		log.Warnf("webrtc: %s; viewers use ephemeral ports", err)
		return nil
	}

	return &portForward{mux: mux, rules: publicAddressRules(settings.PublicAddress)}
}

// apply puts the mux and the extra candidate on a connection's setting engine.
func (p *portForward) apply(engine *webrtc.SettingEngine) {
	engine.SetICEUDPMux(p.mux)

	if len(p.rules) == 0 {
		return
	}
	if err := engine.SetICEAddressRewriteRules(p.rules...); err != nil {
		log.Warnf("webrtc: cannot offer the public address: %s", err)
	}
}

// publicAddressRules turns the public address into a rule that adds a host
// candidate for it, beside the ones for the board's own addresses.
func publicAddressRules(address string) []webrtc.ICEAddressRewriteRule {
	ips := ResolvePublicAddress(address)
	if len(ips) == 0 {
		return nil
	}

	return []webrtc.ICEAddressRewriteRule{{
		External:        ips,
		AsCandidateType: webrtc.ICECandidateTypeHost,
		Mode:            webrtc.ICEAddressRewriteAppend,
	}}
}

// ResolvePublicAddress returns the IP addresses of the public address. An IP
// is returned as it is and a host name is looked up, with the answer kept for
// a minute. A lookup that fails returns the last answer, if there was one.
func ResolvePublicAddress(address string) []string {
	address = strings.TrimSpace(address)
	if address == "" {
		return nil
	}
	if ip, err := netip.ParseAddr(address); err == nil {
		return []string{ip.Unmap().String()}
	}

	resolveMutex.Lock()
	cached, ok := resolved[address]
	resolveMutex.Unlock()
	if ok && time.Now().Before(cached.expires) {
		return cached.ips
	}

	addrs, err := lookupIPs(address)
	var ips []string
	for _, a := range addrs {
		ip, ok := netip.AddrFromSlice(a.IP)
		if !ok {
			continue
		}
		ip = ip.Unmap()
		if ip.IsUnspecified() || ip.IsLoopback() || ip.IsMulticast() || ip.IsLinkLocalUnicast() {
			continue
		}
		ips = append(ips, ip.String())
	}
	if err != nil || len(ips) == 0 {
		log.Warnf("webrtc: cannot resolve the public address %q: %v", address, err)
		return cached.ips
	}

	resolveMutex.Lock()
	resolved[address] = resolvedAddress{ips: ips, expires: time.Now().Add(publicAddressTTL)}
	resolveMutex.Unlock()

	return ips
}

var (
	digitsAndDots = regexp.MustCompile("^[0-9.]+$")
	hostNameRe    = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)
)

// ValidatePortForward checks the settings before they are saved. reserved are
// ports the server already uses.
func ValidatePortForward(settings config.WebRTC, reserved ...int) error {
	if settings.UDPPort != 0 && (settings.UDPPort < 1024 || settings.UDPPort > 65535) {
		return fmt.Errorf("the port must be between 1024 and 65535")
	}
	for _, port := range reserved {
		if settings.UDPPort != 0 && settings.UDPPort == port {
			return fmt.Errorf("the port is one the server already uses")
		}
	}

	address := strings.TrimSpace(settings.PublicAddress)
	if address == "" {
		return nil
	}
	if settings.UDPPort == 0 {
		return fmt.Errorf("a public address needs a UDP port: the address is offered on that port")
	}
	if len(address) > 253 {
		return fmt.Errorf("the public address is too long")
	}

	if ip, err := netip.ParseAddr(address); err == nil {
		if ip.Zone() != "" || ip.IsUnspecified() || ip.IsLoopback() || ip.IsMulticast() || ip.IsLinkLocalUnicast() {
			return fmt.Errorf("the public address cannot be used by a viewer")
		}
		return nil
	}

	// Anything with a colon, or all digits and dots, was meant as an IP.
	if strings.Contains(address, ":") || digitsAndDots.MatchString(address) {
		return fmt.Errorf("the public address is not a valid IP address")
	}
	if !hostNameRe.MatchString(address) {
		return fmt.Errorf("the public address is not a valid host name")
	}

	return nil
}
