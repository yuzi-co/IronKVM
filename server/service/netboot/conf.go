package netboot

import (
	"fmt"
	"net/netip"
	"strings"
)

// HTTPPort is where the menu and the images are served on the link. Plain
// HTTP: iPXE's sanboot reads an image by range requests, and the board's own
// certificate is one iPXE would refuse.
const HTTPPort = 8069

// bootScriptName is the file in the TFTP root that iPXE gets once it runs.
const bootScriptName = "boot.ipxe"

// bootScript chains iPXE to the board's HTTP menu. It names the board by
// ${next-server}, which dnsmasq sets to its own address on the link, so the
// file does not change with the link's subnet.
func bootScript() string {
	return "#!ipxe\n" +
		"# Written by the IronKVM server. Loads the network boot menu from the KVM.\n" +
		fmt.Sprintf("chain http://${next-server}:%d/menu.ipxe || shell\n", HTTPPort)
}

// linkConf is dnsmasq's configuration for the USB link. It holds no subnet:
// S85netboot passes the interface and the host's one address, which S03usbdev
// reads from the link, on the command line.
//
// It keeps the rules the link's udhcpd kept. No DNS server runs (port=0). The
// router and DNS options are set with no value, which is how dnsmasq is told
// to send neither: without the two lines it offers its own address as both,
// and the host would route and resolve through a board that forwards nothing.
// One lease, because the host is the only client a USB link can have, and
// authoritative, so a host that kept its lease across a reboot of the board is
// answered at once.
func linkConf(tftpRoot string) string {
	var b strings.Builder

	b.WriteString("# Network boot on the USB link to the host. Written by the IronKVM server.\n")
	b.WriteString("# S85netboot adds --interface and --dhcp-range from S03usbdev.\n")
	commonConf(&b, tftpRoot)
	b.WriteString("dhcp-authoritative\n")
	b.WriteString("dhcp-lease-max=1\n")
	b.WriteString("dhcp-option=3\n")
	b.WriteString("dhcp-option=6\n")

	// A PXE ROM gets iPXE for its architecture, and iPXE, which says so in
	// its user class, gets the script that loads the menu.
	b.WriteString("dhcp-match=set:bios,option:client-arch,0\n")
	b.WriteString("dhcp-match=set:efi-x86_64,option:client-arch,7\n")
	b.WriteString("dhcp-match=set:efi-x86_64,option:client-arch,9\n")
	b.WriteString("dhcp-match=set:efi-arm64,option:client-arch,11\n")
	b.WriteString("dhcp-boot=tag:!ipxe,tag:bios," + ipxeBIOS + "\n")
	b.WriteString("dhcp-boot=tag:!ipxe,tag:efi-x86_64," + ipxeEFIx64 + "\n")
	b.WriteString("dhcp-boot=tag:!ipxe,tag:efi-arm64," + ipxeEFIarm64 + "\n")
	b.WriteString("dhcp-boot=tag:ipxe," + bootScriptName + "\n")

	return b.String()
}

// lanNetwork is the LAN side: the interface that holds the default route and
// its IPv4 network.
type lanNetwork struct {
	Interface string
	Prefix    netip.Prefix
}

// lanConf is dnsmasq's configuration for proxy DHCP on the LAN.
//
// The only dhcp-range is the LAN's network in proxy mode, so dnsmasq never
// offers an address and leaves that to the LAN's own DHCP server. There is no
// DHCP option at all. What it offers is a boot loader by TFTP: netboot.xyz's
// binary for the client's architecture, whose embedded script loads the
// netboot.xyz menu from the internet. With one entry for each architecture a
// PXE ROM boots it at once, with no menu. Clients that already run iPXE get
// nothing, so netboot.xyz's own iPXE goes on with its script.
//
// The board's HTTP menu and the images in /data are never offered here.
func lanConf(lan lanNetwork, tftpRoot string) (string, error) {
	if !plainInterface(lan.Interface) {
		return "", fmt.Errorf("%q is not an interface name", lan.Interface)
	}
	if !lan.Prefix.IsValid() || !lan.Prefix.Addr().Is4() {
		return "", fmt.Errorf("%s is not an IPv4 network", lan.Prefix)
	}

	network := lan.Prefix.Masked()
	mask := netmask(network.Bits())

	var b strings.Builder

	b.WriteString("# Proxy DHCP for network boot on the LAN. Written by the IronKVM server.\n")
	b.WriteString("# It never hands out an address.\n")
	commonConf(&b, tftpRoot)
	b.WriteString("interface=" + lan.Interface + "\n")
	fmt.Fprintf(&b, "dhcp-range=%s,proxy,%s\n", network.Addr(), mask)
	b.WriteString("pxe-service=tag:!ipxe,x86PC,\"netboot.xyz\"," + netbootXYZBIOS + "\n")
	b.WriteString("pxe-service=tag:!ipxe,X86-64_EFI,\"netboot.xyz\"," + netbootXYZEFIx64 + "\n")
	b.WriteString("pxe-service=tag:!ipxe,BC_EFI,\"netboot.xyz\"," + netbootXYZEFIx64 + "\n")
	b.WriteString("pxe-service=tag:!ipxe,ARM64_EFI,\"netboot.xyz\"," + netbootXYZEFIarm64 + "\n")

	return b.String(), nil
}

// commonConf is what both sides share. No DNS server; one interface, bound by
// device, so the two instances and the Wi-Fi access point's udhcpd each keep
// their own; root dropped for nobody once the sockets are bound; TFTP from the
// add-on's boot files; and the iPXE tag by user class.
//
// dnsmasq adds the loopback interface to any --interface on its own, and would
// open its TFTP port there too. Two instances would then both want
// 127.0.0.1:69, and the second would not start, so loopback is left out.
func commonConf(b *strings.Builder, tftpRoot string) {
	b.WriteString("port=0\n")
	b.WriteString("bind-interfaces\n")
	b.WriteString("except-interface=lo\n")
	b.WriteString("user=nobody\n")
	b.WriteString("enable-tftp\n")
	b.WriteString("tftp-root=" + tftpRoot + "\n")
	b.WriteString("dhcp-userclass=set:ipxe,iPXE\n")
}

// netmask writes a prefix length as a dotted netmask.
func netmask(bits int) netip.Addr {
	var m [4]byte
	for i := 0; i < 4; i++ {
		switch {
		case bits >= 8:
			m[i] = 0xff
			bits -= 8
		case bits > 0:
			m[i] = byte(0xff << (8 - bits))
			bits = 0
		}
	}
	return netip.AddrFrom4(m)
}

// plainInterface accepts what S85netboot accepts for an interface name.
func plainInterface(name string) bool {
	if name == "" || len(name) > 15 {
		return false
	}
	for _, r := range name {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}
