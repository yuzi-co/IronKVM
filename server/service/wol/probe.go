package wol

import (
	"bufio"
	"context"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	probeTimeout = 3 * time.Second
	arpFile      = "/proc/net/arp"
)

// runProbe runs a short command and reports whether it exited zero. A seam for
// the tests.
var runProbe = func(ctx context.Context, name string, args ...string) bool {
	return exec.CommandContext(ctx, name, args...).Run() == nil
}

// IsUp checks whether a host answers: a ping first, then an ARP request for a
// host that drops ICMP. A host on another subnet only answers the ping.
func IsUp(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil || parsed.To4() == nil {
		return false
	}
	ip = parsed.To4().String()

	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()

	if runProbe(ctx, "ping", "-c", "1", "-W", "1", ip) {
		return true
	}

	return runProbe(ctx, "arping", "-c", "1", "-w", "1", ip)
}

// Neighbor is a host the board has seen on the LAN.
type Neighbor struct {
	Mac  string `json:"mac"`
	IP   string `json:"ip"`
	Name string `json:"name,omitempty"`
}

// ParseARP reads /proc/net/arp. Incomplete entries (flags 0x0) are hosts that
// did not answer, and are left out.
func ParseARP(content string) []Neighbor {
	var out []Neighbor
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Scan() // header
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		// IP address, HW type, Flags, HW address, Mask, Device
		if len(f) < 6 || f[2] == "0x0" {
			continue
		}
		mac, err := ParseMAC(f[3])
		if err != nil || mac == "00:00:00:00:00:00" {
			continue
		}
		out = append(out, Neighbor{Mac: mac, IP: f[0]})
	}

	return out
}

// ParseAvahi reads `avahi-browse -a -t -r -p` output into IP to host name.
// Resolved lines look like
// =;eth0;IPv4;name;type;local;host.local;192.168.1.5;port;txt
func ParseAvahi(content string) map[string]string {
	names := map[string]string{}
	for _, line := range strings.Split(content, "\n") {
		f := strings.Split(line, ";")
		if len(f) < 8 || f[0] != "=" || f[2] != "IPv4" {
			continue
		}
		host := strings.TrimSuffix(f[6], ".local")
		if host == "" || net.ParseIP(f[7]) == nil {
			continue
		}
		if _, ok := names[f[7]]; !ok {
			names[f[7]] = host
		}
	}

	return names
}

// browseMDNS lists mDNS hosts, or nothing when avahi-browse is not installed.
var browseMDNS = func() string {
	path, err := exec.LookPath("avahi-browse")
	if err != nil {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	out, _ := exec.CommandContext(ctx, path, "-a", "-t", "-r", "-p").Output()

	return string(out)
}

var readARP = func() string {
	data, _ := os.ReadFile(arpFile)
	return string(data)
}

// Scan returns the hosts on the LAN from the ARP table, named from mDNS when
// that answers.
func Scan() []Neighbor {
	list := ParseARP(readARP())
	names := ParseAvahi(browseMDNS())
	for i := range list {
		list[i].Name = names[list[i].IP]
	}

	return list
}
