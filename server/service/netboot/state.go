package netboot

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"
)

// procNetRoute is the kernel's IPv4 routing table. A variable for the tests.
var procNetRoute = "/proc/net/route"

// defaultRouteInterface names the interface that holds the IPv4 default
// route, the board's way to the LAN. A USB link never holds it: the board
// routes nothing over the link.
func defaultRouteInterface() (string, error) {
	content, err := os.ReadFile(procNetRoute)
	if err != nil {
		return "", err
	}

	sc := bufio.NewScanner(bytes.NewReader(content))
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		// Iface Destination Gateway Flags ...; the header line has
		// "Destination" in the second field and is skipped by the check.
		if len(fields) < 4 || fields[1] != "00000000" {
			continue
		}
		flags, err := strconv.ParseUint(fields[3], 16, 32)
		if err != nil || flags&0x1 == 0 { // RTF_UP
			continue
		}
		if strings.HasPrefix(fields[0], "usb") {
			continue
		}
		return fields[0], nil
	}

	return "", errors.New("the board has no default route")
}

// currentLAN is the LAN side of network boot: the default route's interface
// and its IPv4 network.
func currentLAN() (lanNetwork, error) {
	name, err := defaultRouteInterface()
	if err != nil {
		return lanNetwork{}, err
	}

	iface, err := net.InterfaceByName(name)
	if err != nil {
		return lanNetwork{}, err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return lanNetwork{}, err
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ip, ok := netip.AddrFromSlice(ipNet.IP)
		if !ok || !ip.Unmap().Is4() {
			continue
		}
		bits, _ := ipNet.Mask.Size()
		return lanNetwork{Interface: name, Prefix: netip.PrefixFrom(ip.Unmap(), bits).Masked()}, nil
	}

	return lanNetwork{}, fmt.Errorf("%s has no IPv4 address", name)
}

// Lease is the host's lease on the link, from dnsmasq's lease file.
type Lease struct {
	MAC      string    `json:"mac"`
	IP       string    `json:"ip"`
	Hostname string    `json:"hostname"`
	Expires  time.Time `json:"expires"`
}

// parseLeases reads dnsmasq's lease file: one lease a line, as
// "<expiry> <mac> <ip> <hostname> <client id>", with * for a missing name.
// An expiry of 0 means the lease never ends.
func parseLeases(content []byte) []Lease {
	var leases []Lease

	sc := bufio.NewScanner(bytes.NewReader(content))
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 4 {
			continue
		}
		expiry, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			continue
		}
		lease := Lease{MAC: fields[1], IP: fields[2]}
		if fields[3] != "*" {
			lease.Hostname = fields[3]
		}
		if expiry > 0 {
			lease.Expires = time.Unix(expiry, 0).UTC()
		}
		leases = append(leases, lease)
	}

	return leases
}

// logTailBytes bounds what is read of a log, from its end.
const logTailBytes = 16 << 10

// tailLines returns the last n lines of a file, oldest first, or none.
func tailLines(name string, n int) []string {
	f, err := os.Open(name)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	fi, err := f.Stat()
	if err != nil {
		return nil
	}
	offset := fi.Size() - logTailBytes
	if offset < 0 {
		offset = 0
	}
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return nil
	}
	content, err := io.ReadAll(f)
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	if offset > 0 && len(lines) > 0 {
		// The first line is cut in the middle.
		lines = lines[1:]
	}
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines
}
