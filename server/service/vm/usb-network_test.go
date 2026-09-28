package vm

import (
	"errors"
	"net/netip"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestUSBNetworkModeFollowsTheScriptsRanking(t *testing.T) {
	cases := []struct {
		name    string
		markers []string
		want    string
	}{
		{"no marker", nil, usbNetworkOff},
		{"ncm", []string{virtualNetworkNCM}, usbNetworkNCM},
		{"ecm", []string{virtualNetworkECM}, usbNetworkECM},
		{"rndis from an older server", []string{virtualNetworkRNDIS}, usbNetworkRNDIS},
		{"ncm outranks ecm", []string{virtualNetworkECM, virtualNetworkNCM}, usbNetworkNCM},
		{"ecm outranks rndis", []string{virtualNetworkRNDIS, virtualNetworkECM}, usbNetworkECM},
		{"other functions do not count", []string{virtualConsole, virtualDisk, virtualAudio}, usbNetworkOff},
	}

	for _, c := range cases {
		if got := usbNetworkMode(presence(c.markers...)); got != c.want {
			t.Errorf("%s: mode is %q, want %q", c.name, got, c.want)
		}
	}
}

// Every marker the mode reads is a marker the budget counts, or the budget
// would approve a set the gadget cannot carry.
func TestEveryUSBNetworkMarkerIsANetworkMarker(t *testing.T) {
	network, _ := functionForDevice("network")

	var modes []string
	for _, candidate := range usbNetworkModes {
		modes = append(modes, candidate.marker)
	}

	if !reflect.DeepEqual(modes, network.markers) {
		t.Errorf("the modes read %v, the budget counts %v", modes, network.markers)
	}
}

func TestParseUSBSubnetAcceptsPrivateNetworks(t *testing.T) {
	cases := []struct {
		value string
		board string
		host  string
	}{
		{"172.31.255.0/30", "172.31.255.1", "172.31.255.2"},
		{"10.0.0.0/24", "10.0.0.1", "10.0.0.2"},
		{"192.168.7.8/29", "192.168.7.9", "192.168.7.10"},
		{"172.16.0.128/25", "172.16.0.129", "172.16.0.130"},
		{" 10.9.8.0/29 ", "10.9.8.1", "10.9.8.2"},
	}

	for _, c := range cases {
		subnet, err := parseUSBSubnet(c.value)
		if err != nil {
			t.Errorf("%q refused: %s", c.value, err)
			continue
		}

		if subnet.board.String() != c.board || subnet.host.String() != c.host {
			t.Errorf("%q gives board %s and host %s, want %s and %s",
				c.value, subnet.board, subnet.host, c.board, c.host)
		}
	}
}

// The same refusals usb_net_parse makes. A subnet the server accepted and the
// script refused would be saved, reported, and silently replaced by the
// default at the next start.
func TestParseUSBSubnetRefusesWhatTheScriptRefuses(t *testing.T) {
	for _, value := range []string{
		"172.31.255.1/30", // host bits set
		"172.31.255.4/29", // host bits set
		"8.8.8.0/30",      // public
		"100.64.0.0/30",   // shared address space, not private
		"172.15.0.0/30",   // just below 172.16.0.0/12
		"172.32.0.0/30",   // just above it
		"10.0.0.0/31",     // no room for two hosts
		"10.0.0.0/23",     // wider than the script's arithmetic
		"10.0.0.08/30",    // leading zero, octal to ash
		"10.0.0.0",
		"10.0.0/30",
		"10.0.0.256/30",
		"fd00::/64",
		"",
	} {
		if _, err := parseUSBSubnet(value); err == nil {
			t.Errorf("%q was accepted", value)
		}
	}
}

// The refusal for host bits names the network the operator probably meant.
func TestParseUSBSubnetSuggestsTheNetworkAddress(t *testing.T) {
	_, err := parseUSBSubnet("172.31.255.1/30")
	if err == nil || !strings.Contains(err.Error(), "172.31.255.0/30") {
		t.Errorf("the refusal is %v, want it to name 172.31.255.0/30", err)
	}
}

func TestReadUSBSubnet(t *testing.T) {
	cases := []struct {
		name    string
		content string
		err     error
		want    string
	}{
		{"no file", "", os.ErrNotExist, defaultUSBNetworkSubnet},
		{"a valid file", "10.9.8.0/29\n", nil, "10.9.8.0/29"},
		{"CRLF", "10.9.8.0/29\r\n", nil, "10.9.8.0/29"},
		{"only the first line", "10.9.8.0/29\n8.8.8.0/30\n", nil, "10.9.8.0/29"},
		{"a public network", "8.8.8.0/30\n", nil, defaultUSBNetworkSubnet},
		{"an empty file", "", nil, defaultUSBNetworkSubnet},
	}

	for _, c := range cases {
		read := func(string) ([]byte, error) {
			return []byte(c.content), c.err
		}

		if got := readUSBSubnet(read).prefix.String(); got != c.want {
			t.Errorf("%s: subnet is %s, want %s", c.name, got, c.want)
		}
	}
}

func TestReadUSBSubnetReadsTheFileTheScriptReads(t *testing.T) {
	var asked string
	readUSBSubnet(func(path string) ([]byte, error) {
		asked = path
		return nil, errors.New("absent")
	})

	script := readInitScript(t)
	if !strings.Contains(script, "file=${USB_NET_SUBNET_FILE:-"+asked+"}") {
		t.Errorf("the server reads %s, and S03usbdev does not", asked)
	}
}

// Two copies of one default. If they differ, the UI reports one subnet and the
// board uses another on every board that never saved one.
func TestTheDefaultSubnetIsTheScriptsDefault(t *testing.T) {
	match := regexp.MustCompile(`(?m)^\s*fallback=(\S+)$`).FindStringSubmatch(readInitScript(t))
	if match == nil {
		t.Fatal("S03usbdev has no fallback= line in usb_net_subnet")
	}

	if match[1] != defaultUSBNetworkSubnet {
		t.Errorf("S03usbdev defaults to %s, the server to %s", match[1], defaultUSBNetworkSubnet)
	}
}

// The spec asks for a /30 by default: the board, the host, and nothing else
// that a lease could hand out.
func TestTheDefaultSubnetIsASlash30(t *testing.T) {
	subnet, err := parseUSBSubnet(defaultUSBNetworkSubnet)
	if err != nil {
		t.Fatalf("the default is refused: %s", err)
	}

	if subnet.prefix.Bits() != 30 {
		t.Errorf("the default is a /%d", subnet.prefix.Bits())
	}
}

func TestUSBSubnetOverlap(t *testing.T) {
	subnet, _ := parseUSBSubnet("192.168.1.0/30")

	lan := netip.MustParsePrefix("192.168.1.0/24")
	if got, ok := usbSubnetOverlap(subnet, []netip.Prefix{lan}); !ok || got != lan {
		t.Errorf("a subnet inside the LAN was not refused: %v %v", got, ok)
	}

	other := netip.MustParsePrefix("10.0.0.0/24")
	if _, ok := usbSubnetOverlap(subnet, []netip.Prefix{other}); ok {
		t.Error("a subnet beside the LAN was refused")
	}

	if _, ok := usbSubnetOverlap(subnet, nil); ok {
		t.Error("a board with no other address refused the subnet")
	}
}

// Turning the link on from off is the only transition that can overrun the
// budget. With HID and the console on, the network does not fit, and the
// refusal is the sentence the other switches give.
func TestUSBNetworkFitsRefusesTheConsoleAndNetworkTogether(t *testing.T) {
	present := presence(virtualConsole)

	for _, mode := range []string{usbNetworkNCM, usbNetworkECM} {
		ok, refusal := usbNetworkFits(mode, present)
		if ok {
			t.Errorf("%s fits beside the console and HID", mode)
			continue
		}

		if !strings.Contains(refusal, "network needs 2 inbound USB endpoints, 1 free") {
			t.Errorf("%s refusal is %q", mode, refusal)
		}

		if !strings.Contains(refusal, "console") {
			t.Errorf("%s refusal %q does not name the console as the way out", mode, refusal)
		}
	}
}

func TestUSBNetworkFitsWhatTheBudgetAllows(t *testing.T) {
	cases := []struct {
		name    string
		mode    string
		markers []string
		want    bool
	}{
		{"on beside the disk and the speaker", usbNetworkNCM, []string{virtualDisk, virtualAudio}, true},
		{"ecm beside the disk and the speaker", usbNetworkECM, []string{virtualDisk, virtualAudio}, true},
		{"on beside the console", usbNetworkNCM, []string{virtualConsole}, false},
		{"on beside the console without HID", usbNetworkNCM, []string{virtualConsole, disableHid}, true},
		{"off beside everything", usbNetworkOff, []string{virtualConsole, virtualDisk, virtualAudio}, true},
		// A board that already carries the network over budget, from an
		// older server or a hand edit, can still switch mode and still turn
		// the link off: neither asks for anything more.
		{"ncm to ecm over budget", usbNetworkECM, []string{virtualConsole, virtualNetworkNCM}, true},
		{"rndis to ncm", usbNetworkNCM, []string{virtualNetworkRNDIS}, true},
		{"off from over budget", usbNetworkOff, []string{virtualConsole, virtualNetworkECM}, true},
	}

	for _, c := range cases {
		if ok, refusal := usbNetworkFits(c.mode, presence(c.markers...)); ok != c.want {
			t.Errorf("%s: fits is %v (%q), want %v", c.name, ok, refusal, c.want)
		}
	}
}

// Every network marker comes off, then at most one goes on, between stop and
// start.
func TestUSBNetworkCommands(t *testing.T) {
	cases := []struct {
		mode  string
		touch string
	}{
		{usbNetworkNCM, "touch /boot/usb.ncm"},
		{usbNetworkECM, "touch /boot/usb.ecm"},
		{usbNetworkOff, ""},
	}

	for _, c := range cases {
		commands := usbNetworkCommands(c.mode)

		if commands[0] != "/etc/init.d/S03usbdev stop" {
			t.Errorf("%s starts with %q, want the stop", c.mode, commands[0])
		}

		if commands[len(commands)-1] != "/etc/init.d/S03usbdev start" {
			t.Errorf("%s ends with %q, want the start", c.mode, commands[len(commands)-1])
		}

		for _, marker := range []string{virtualNetworkNCM, virtualNetworkECM, virtualNetworkRNDIS} {
			removed := -1
			for i, command := range commands {
				if command == "rm -f "+marker {
					removed = i
				}
			}

			if removed < 0 {
				t.Errorf("%s never removes %s with rm -f", c.mode, marker)
			}
		}

		var touches []string
		for i, command := range commands {
			if !strings.HasPrefix(command, "touch ") {
				continue
			}

			touches = append(touches, command)

			for _, earlier := range commands[i+1:] {
				if strings.HasPrefix(earlier, "rm -f /boot/usb.") {
					t.Errorf("%s removes a marker after it touches one: %v", c.mode, commands)
				}
			}
		}

		switch {
		case c.touch == "" && len(touches) != 0:
			t.Errorf("off touches %v", touches)
		case c.touch != "" && !reflect.DeepEqual(touches, []string{c.touch}):
			t.Errorf("%s touches %v, want [%s]", c.mode, touches, c.touch)
		}

		for _, command := range commands {
			if strings.Contains(command, "rndis0") && strings.HasPrefix(command, "touch") {
				t.Errorf("%s writes the RNDIS marker, which the UI no longer offers", c.mode)
			}

			if strings.Contains(command, "rmdir") || strings.Contains(command, "functions/") {
				t.Errorf("%s runs %q, which can block forever", c.mode, command)
			}
		}
	}
}

func TestNeedsRebuild(t *testing.T) {
	cases := []struct {
		name          string
		current       string
		wanted        string
		subnetChanged bool
		want          bool
	}{
		{"nothing changed", usbNetworkNCM, usbNetworkNCM, false, false},
		{"off stays off", usbNetworkOff, usbNetworkOff, false, false},
		{"a new subnet while off waits for the next start", usbNetworkOff, usbNetworkOff, true, false},
		{"a new subnet while on", usbNetworkECM, usbNetworkECM, true, true},
		{"on", usbNetworkOff, usbNetworkNCM, false, true},
		{"off", usbNetworkNCM, usbNetworkOff, false, true},
		{"ncm to ecm", usbNetworkNCM, usbNetworkECM, false, true},
		{"rndis to ncm", usbNetworkRNDIS, usbNetworkNCM, false, true},
	}

	for _, c := range cases {
		if got := needsRebuild(c.current, c.wanted, c.subnetChanged); got != c.want {
			t.Errorf("%s: rebuild is %v, want %v", c.name, got, c.want)
		}
	}
}

func TestUSBNetworkStateReportsTheSubnetAndTheFit(t *testing.T) {
	read := func(string) ([]byte, error) { return []byte("10.9.8.0/29\n"), nil }

	state := usbNetworkState(presence(virtualConsole), read)

	if state.Mode != usbNetworkOff {
		t.Errorf("mode is %q, want off", state.Mode)
	}

	if state.Subnet != "10.9.8.0/29" || state.Board != "10.9.8.1" || state.Host != "10.9.8.2" {
		t.Errorf("subnet %s, board %s, host %s", state.Subnet, state.Board, state.Host)
	}

	if state.Fits || state.Refusal == "" {
		t.Errorf("beside the console the link fits=%v with refusal %q", state.Fits, state.Refusal)
	}

	if state.Active {
		t.Error("a link that is off is reported active")
	}

	state = usbNetworkState(presence(virtualDisk), read)
	if !state.Fits || state.Refusal != "" {
		t.Errorf("beside the disk the link fits=%v with refusal %q", state.Fits, state.Refusal)
	}
}
