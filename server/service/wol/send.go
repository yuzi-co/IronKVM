package wol

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

const (
	defaultUDPPort = 9
	sendTimeout    = 3 * time.Second
)

// Seams for the tests: neither may run for real off the device.
var (
	runEtherWake = func(args ...string) ([]byte, error) {
		return exec.Command("ether-wake", args...).CombinedOutput()
	}
	sendUDP = func(target net.IP, port int, payload []byte) error {
		conn, err := net.DialTimeout("udp4", net.JoinHostPort(target.String(), fmt.Sprint(port)), sendTimeout)
		if err != nil {
			return err
		}
		defer func() { _ = conn.Close() }()
		_ = conn.SetWriteDeadline(time.Now().Add(sendTimeout))
		_, err = conn.Write(payload)

		return err
	}
)

// MagicPacket builds the payload: six 0xFF bytes, the MAC sixteen times, and
// the SecureOn password when there is one.
func MagicPacket(mac, secureOn string) ([]byte, error) {
	hw, err := net.ParseMAC(mac)
	if err != nil || len(hw) != 6 {
		return nil, ErrInvalidMAC
	}

	var b bytes.Buffer
	b.Write(bytes.Repeat([]byte{0xff}, 6))
	for i := 0; i < 16; i++ {
		b.Write(hw)
	}
	if secureOn != "" {
		pw, err := net.ParseMAC(secureOn)
		if err != nil || len(pw) != 6 {
			return nil, fmt.Errorf("invalid SecureOn password")
		}
		b.Write(pw)
	}

	return b.Bytes(), nil
}

// usesUDP says whether the device is woken with a UDP datagram. A device with
// no options keeps the Ethernet frame ether-wake has always sent.
func (d Device) usesUDP() bool {
	return d.Broadcast != "" || d.Port != 0
}

// Send wakes a device. The returned text is the tool's output on failure.
func Send(d Device) (string, error) {
	if !d.usesUDP() {
		args := []string{"-b"}
		if d.SecureOn != "" {
			args = append(args, "-p", d.SecureOn)
		}
		args = append(args, d.Mac)
		out, err := runEtherWake(args...)
		if err != nil {
			// The arguments carry the password, so they stay out of the error.
			return strings.TrimSpace(string(out)), fmt.Errorf("ether-wake failed: %w", err)
		}

		return "", nil
	}

	payload, err := MagicPacket(d.Mac, d.SecureOn)
	if err != nil {
		return "", err
	}

	target := net.IPv4bcast
	if d.Broadcast != "" {
		if target, err = BroadcastTarget(d.Broadcast); err != nil {
			return "", err
		}
	}
	port := d.Port
	if port == 0 {
		port = defaultUDPPort
	}

	if err := sendUDP(target, port, payload); err != nil {
		return "", fmt.Errorf("failed to send the packet: %w", err)
	}

	return "", nil
}
