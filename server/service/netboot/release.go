package netboot

// The upstream files network boot uses, pinned by version and SHA-256. None
// of them is in the image: the add-on downloads the boot files to /data when
// the owner installs it, and the ISO goes to the image directory when the
// owner asks for it. A new release is a new pin here, checked by hand.

// pinnedFile is a file downloaded as it is published.
type pinnedFile struct {
	Name   string // the name it is saved under
	URL    string
	SHA256 string
}

// archiveMember is a file taken out of a pinned archive.
type archiveMember struct {
	Member string // its path inside the archive
	Name   string // the name it is saved under
	SHA256 string
}

// iPXE v2.0.0 is the first iPXE release that publishes binaries, and
// ipxeboot.tar.gz is its "network boot server files" archive. It holds every
// architecture iPXE builds; network boot takes the three hosts are likely to
// be: a BIOS PC, a UEFI x86-64 PC, and a UEFI arm64 machine.
const (
	ipxeVersion = "2.0.0"
	// ipxeArchiveMax bounds the download. The archive is 12,002,760 bytes.
	ipxeArchiveMax = 32 << 20
)

// The pins below are variables only so the tests can serve their own files;
// nothing else changes them.
var ipxeArchive = pinnedFile{
	Name:   "ipxeboot.tar.gz",
	URL:    "https://github.com/ipxe/ipxe/releases/download/v2.0.0/ipxeboot.tar.gz",
	SHA256: "01a526d4cc791fc30362259c609d6c506cc64a7bdff51b9a5eb788354e17eee1",
}

// The names the link's dnsmasq offers by client architecture (DHCP option
// 93). The files are saved under these names, so conf.go and the list below
// cannot disagree.
const (
	ipxeBIOS     = "undionly.kpxe"
	ipxeEFIx64   = "ipxe.efi"
	ipxeEFIarm64 = "ipxe-arm64.efi"
)

var ipxeMembers = []archiveMember{
	{
		Member: "ipxeboot/x86_64/undionly.kpxe",
		Name:   ipxeBIOS,
		SHA256: "4186562d21ff54e970d905751c9f36d628e73a51a94afe4a6a42f925b0df448c",
	},
	{
		Member: "ipxeboot/x86_64/ipxe.efi",
		Name:   ipxeEFIx64,
		SHA256: "868aa34057ff416ebf2fdfb5781de035e2c540477c04039198a9f8a9c6130034",
	},
	{
		Member: "ipxeboot/arm64/ipxe.efi",
		Name:   ipxeEFIarm64,
		SHA256: "a9cb6df506a68f3afa4bb94cf6cb8e3862a5a2ff3bf1d8027fea83f7e1d49217",
	},
}

// netboot.xyz 3.0.3. Its binaries carry an embedded script that loads the
// netboot.xyz menu from the internet: the LAN's proxy DHCP offers them
// directly, and the link's menu chains them. The sums are the release's own
// netboot.xyz-sha256-checksums.txt.
const netbootXYZVersion = "3.0.3"

// The names of the netboot.xyz binaries, for the LAN's proxy DHCP and the
// link's menu.
const (
	netbootXYZBIOS     = "netboot.xyz.kpxe"
	netbootXYZEFIx64   = "netboot.xyz.efi"
	netbootXYZEFIarm64 = "netboot.xyz-arm64.efi"
)

var netbootXYZFiles = []pinnedFile{
	{
		Name:   netbootXYZBIOS,
		URL:    "https://github.com/netbootxyz/netboot.xyz/releases/download/3.0.3/netboot.xyz.kpxe",
		SHA256: "a1aba898b89a29d165089470ee15a742ebc39529236a06608464d5cb6c892d93",
	},
	{
		Name:   netbootXYZEFIx64,
		URL:    "https://github.com/netbootxyz/netboot.xyz/releases/download/3.0.3/netboot.xyz.efi",
		SHA256: "56bb21e9f6d79eadb947e7031a088c9afcc48aa1eb99b2502d161f0b1c586a82",
	},
	{
		Name:   netbootXYZEFIarm64,
		URL:    "https://github.com/netbootxyz/netboot.xyz/releases/download/3.0.3/netboot.xyz-arm64.efi",
		SHA256: "e8eabbefc83c1db49ae773d9e3318cf091182a51fbb86aff87bc871e867ae61a",
	},
}

// bootFileMax bounds each netboot.xyz binary. The largest is 1,185,792 bytes.
const bootFileMax = 8 << 20

// The netboot.xyz ISO, for the virtual CD. The download service fetches it
// into the image directory and checks this sum, as it checks one an operator
// types. BootMenuISO* are exported for it.
const (
	BootMenuISOURL    = "https://github.com/netbootxyz/netboot.xyz/releases/download/3.0.3/netboot.xyz.iso"
	BootMenuISOSHA256 = "2206b05c1c7a8ec7a7a4356dd5f811e19c8081d7d5e2bc856f1f3f4059c5fed1"
)

// bootFiles lists every file the TFTP root has to hold for the add-on to be
// complete, boot.ipxe included.
func bootFiles() []string {
	names := []string{bootScriptName}
	for _, m := range ipxeMembers {
		names = append(names, m.Name)
	}
	for _, f := range netbootXYZFiles {
		names = append(names, f.Name)
	}
	return names
}
