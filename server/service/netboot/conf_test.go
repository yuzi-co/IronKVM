package netboot

import (
	"net/netip"
	"slices"
	"strings"
	"testing"
)

func confLines(conf string) []string {
	var lines []string
	for _, line := range strings.Split(conf, "\n") {
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	return lines
}

func withPrefix(lines []string, prefix string) []string {
	var got []string
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			got = append(got, line)
		}
	}
	return got
}

// The link keeps #10's rules: the host gets an address and a mask, and no
// route and no resolver through a board that forwards nothing.
func TestTheLinkOffersNoRouterAndNoDNS(t *testing.T) {
	lines := confLines(linkConf("/data/ironkvm/addons/netboot/tftp"))

	if !slices.Contains(lines, "port=0") {
		t.Error("the link's dnsmasq runs a DNS server: no port=0")
	}

	// dnsmasq sends its own address as the router and the DNS server unless
	// the option is given with no value. A value would be worse still.
	options := withPrefix(lines, "dhcp-option")
	want := []string{"dhcp-option=3", "dhcp-option=6"}
	if !slices.Equal(options, want) {
		t.Errorf("dhcp options %q, want exactly %q", options, want)
	}
	for _, line := range lines {
		if strings.Contains(line, "router") || strings.Contains(line, "dns-server") {
			t.Errorf("a line names a router or a DNS server: %q", line)
		}
	}
}

// One host, one lease. The range itself comes from S03usbdev, which knows the
// subnet, so the file never holds one that could go stale.
func TestTheLinkLeasesOneAddressAndHoldsNoRange(t *testing.T) {
	lines := confLines(linkConf("/tftp"))

	for _, want := range []string{"dhcp-lease-max=1", "dhcp-authoritative", "bind-interfaces"} {
		if !slices.Contains(lines, want) {
			t.Errorf("no %q", want)
		}
	}
	if got := withPrefix(lines, "dhcp-range"); len(got) != 0 {
		t.Errorf("the file holds a range: %q", got)
	}
	if got := withPrefix(lines, "interface"); len(got) != 0 {
		t.Errorf("the file names an interface, which S03usbdev passes: %q", got)
	}
}

// Option 93 picks the loader; iPXE, by its user class, gets the menu script.
func TestTheLinkPicksTheBootFileByArchitecture(t *testing.T) {
	lines := confLines(linkConf("/data/ironkvm/addons/netboot/tftp"))

	for _, want := range []string{
		"enable-tftp",
		"tftp-root=/data/ironkvm/addons/netboot/tftp",
		"dhcp-userclass=set:ipxe,iPXE",
		"dhcp-match=set:bios,option:client-arch,0",
		"dhcp-match=set:efi-x86_64,option:client-arch,7",
		"dhcp-match=set:efi-x86_64,option:client-arch,9",
		"dhcp-match=set:efi-arm64,option:client-arch,11",
		"dhcp-boot=tag:!ipxe,tag:bios,undionly.kpxe",
		"dhcp-boot=tag:!ipxe,tag:efi-x86_64,ipxe.efi",
		"dhcp-boot=tag:!ipxe,tag:efi-arm64,ipxe-arm64.efi",
		"dhcp-boot=tag:ipxe,boot.ipxe",
	} {
		if !slices.Contains(lines, want) {
			t.Errorf("no %q", want)
		}
	}

	// Every loader offered is one the add-on installs.
	for _, line := range withPrefix(lines, "dhcp-boot=") {
		file := line[strings.LastIndex(line, ",")+1:]
		if !slices.Contains(bootFiles(), file) {
			t.Errorf("%q offers %s, which the add-on does not install", line, file)
		}
	}
}

func TestTheBootScriptChainsToTheMenuOnTheBoard(t *testing.T) {
	script := bootScript()
	if !strings.HasPrefix(script, "#!ipxe\n") {
		t.Fatalf("not an iPXE script:\n%s", script)
	}
	if !strings.Contains(script, "chain http://${next-server}:8069/menu.ipxe") {
		t.Fatalf("does not chain to the menu:\n%s", script)
	}
}

func lan(t *testing.T, iface, prefix string) lanNetwork {
	t.Helper()
	return lanNetwork{Interface: iface, Prefix: netip.MustParsePrefix(prefix)}
}

// Proxy DHCP must never hand out an address: the LAN has its own DHCP server,
// and a second one would fight it for every client.
func TestTheLANIsProxyOnlyAndNeverHasAnAddressRange(t *testing.T) {
	for _, tc := range []struct{ prefix, want string }{
		{"192.168.1.37/24", "dhcp-range=192.168.1.0,proxy,255.255.255.0"},
		{"10.20.30.40/20", "dhcp-range=10.20.16.0,proxy,255.255.240.0"},
		{"172.16.5.1/16", "dhcp-range=172.16.0.0,proxy,255.255.0.0"},
	} {
		conf, err := lanConf(lan(t, "eth0", tc.prefix), "/tftp")
		if err != nil {
			t.Fatal(err)
		}
		lines := confLines(conf)

		ranges := withPrefix(lines, "dhcp-range")
		if !slices.Equal(ranges, []string{tc.want}) {
			t.Errorf("%s: ranges %q, want only %q", tc.prefix, ranges, tc.want)
		}
		for _, r := range ranges {
			if !strings.Contains(r, ",proxy,") {
				t.Errorf("%s: %q is an address range", tc.prefix, r)
			}
		}
		for _, banned := range []string{"dhcp-authoritative", "dhcp-option", "dhcp-host", "dhcp-lease", "dhcp-boot"} {
			if got := withPrefix(lines, banned); len(got) != 0 {
				t.Errorf("%s: the LAN side has %q", tc.prefix, got)
			}
		}
		for _, want := range []string{"port=0", "interface=eth0", "bind-interfaces", "enable-tftp", "tftp-root=/tftp"} {
			if !slices.Contains(lines, want) {
				t.Errorf("%s: no %q", tc.prefix, want)
			}
		}
	}
}

// The LAN gets netboot.xyz's own binaries, which load the netboot.xyz menu
// from the internet, and nothing that points back at the board's menu.
func TestTheLANOffersNetbootXYZAndNotTheBoardsMenu(t *testing.T) {
	conf, err := lanConf(lan(t, "eth0", "192.168.1.37/24"), "/tftp")
	if err != nil {
		t.Fatal(err)
	}
	lines := confLines(conf)

	for _, want := range []string{
		`pxe-service=tag:!ipxe,x86PC,"netboot.xyz",netboot.xyz.kpxe`,
		`pxe-service=tag:!ipxe,X86-64_EFI,"netboot.xyz",netboot.xyz.efi`,
		`pxe-service=tag:!ipxe,BC_EFI,"netboot.xyz",netboot.xyz.efi`,
		`pxe-service=tag:!ipxe,ARM64_EFI,"netboot.xyz",netboot.xyz-arm64.efi`,
	} {
		if !slices.Contains(lines, want) {
			t.Errorf("no %q", want)
		}
	}
	for _, line := range lines {
		if strings.Contains(line, "boot.ipxe") || strings.Contains(line, "menu.ipxe") ||
			strings.Contains(line, "undionly") || strings.Contains(line, "8069") {
			t.Errorf("the LAN is pointed at the board's menu: %q", line)
		}
	}
}

func TestTheLANRefusesWhatItCannotWrite(t *testing.T) {
	for _, bad := range []lanNetwork{
		{Interface: "", Prefix: netip.MustParsePrefix("192.168.1.0/24")},
		{Interface: "eth0\ndhcp-range=192.168.1.10,192.168.1.20", Prefix: netip.MustParsePrefix("192.168.1.0/24")},
		{Interface: "eth 0", Prefix: netip.MustParsePrefix("192.168.1.0/24")},
		{Interface: "eth0", Prefix: netip.MustParsePrefix("fd00::/64")},
		{Interface: "eth0"},
	} {
		if conf, err := lanConf(bad, "/tftp"); err == nil {
			t.Errorf("%+v was accepted:\n%s", bad, conf)
		}
	}
}

func TestNetmask(t *testing.T) {
	for bits, want := range map[int]string{
		0: "0.0.0.0", 8: "255.0.0.0", 20: "255.255.240.0", 24: "255.255.255.0",
		30: "255.255.255.252", 32: "255.255.255.255",
	} {
		if got := netmask(bits).String(); got != want {
			t.Errorf("/%d: %s, want %s", bits, got, want)
		}
	}
}
