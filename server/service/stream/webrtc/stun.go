package webrtc

import (
	"errors"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"sync"

	"NanoKVM-Server/config"
)

// The STUN server (ironkvm-dist#157). The browser and the board each ask it
// for their public address at the start of every WebRTC session, which only
// matters for WebRTC across NAT without a VPN or a port forward. The setting
// is config.Stun: host:port, empty for the default, or config.StunDisabled.

// stunMu guards config.Stun, which a save changes while sessions read it.
var stunMu sync.RWMutex

// CurrentStun returns the STUN setting a new session uses.
func CurrentStun() string {
	stunMu.RLock()
	defer stunMu.RUnlock()

	return config.GetInstance().Stun
}

// SetStun changes the STUN setting for the sessions that start after it.
func SetStun(stun string) {
	stunMu.Lock()
	defer stunMu.Unlock()

	config.GetInstance().Stun = stun
}

var errStunHostPort = errors.New("the STUN server is host:port")

// ValidateStun checks a STUN setting before it is saved or imported: empty
// (the default), config.StunDisabled, or host:port with a port from 1 to 65535
// and a host name or IP address. A user name, a scheme or a path is refused.
func ValidateStun(stun string) error {
	if stun == "" || stun == config.StunDisabled {
		return nil
	}
	if len(stun) > 260 || strings.ContainsAny(stun, " \t@/") {
		return errStunHostPort
	}

	host, portText, err := net.SplitHostPort(stun)
	if err != nil || host == "" || portText == "" {
		return errStunHostPort
	}
	port, err := strconv.Atoi(portText)
	if err != nil || strings.ContainsAny(portText, "+-") {
		return errStunHostPort
	}
	if port < 1 || port > 65535 {
		return errors.New("the STUN port must be from 1 to 65535")
	}

	if ip, err := netip.ParseAddr(host); err == nil {
		if ip.Zone() != "" {
			return errors.New("the STUN server cannot be a link-local address with a zone")
		}
		return nil
	}
	// A colon left in the host means an IPv6 address without brackets.
	if strings.Contains(host, ":") || digitsAndDots.MatchString(host) || !hostNameRe.MatchString(host) {
		return errors.New("the STUN host is not a valid host name or IP address")
	}

	return nil
}
