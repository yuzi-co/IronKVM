package ventoy

// The Ventoy release the disk is built from, pinned by version and SHA-256.
// It is not in the image: the owner installs it from the web UI, and the
// server downloads the archive to /data, takes the three files the disk
// needs out of it, and removes it. Ventoy is GPLv3. A new release is a new
// pin here, with its layout checked by hand against Ventoy2Disk.sh.

// pinnedFile is a file downloaded as it is published.
type pinnedFile struct {
	URL    string
	SHA256 string
	// Max bounds the download. The archive is 20,291,141 bytes.
	Max int64
}

// releaseMember is a file taken out of the archive. An xz member is
// decompressed, and Sum is the sum of what it decompresses to.
type releaseMember struct {
	Member string // its path inside the archive
	Name   string // the name it is saved under
	XZ     bool
	SHA256 string
}

// Version is the pinned Ventoy release.
const Version = "1.1.17"

// The pins are variables only so the tests can serve their own files;
// nothing else changes them.
var releaseArchive = pinnedFile{
	URL:    "https://github.com/ventoy/Ventoy/releases/download/v1.1.17/ventoy-1.1.17-linux.tar.gz",
	SHA256: "7fb4ed08cef6a6b4d39dd19260d8c80291a78dfdf9af7d461571e23cbbc43805",
	Max:    64 << 20,
}

// The saved names avoid .img and .iso: the directory is under the image
// directory, and the image list would show them.
const (
	bootName = "boot.bin"
	coreName = "core.bin"
	efiName  = "vtoyefi.bin"
)

var releaseMembers = []releaseMember{
	{
		// The MBR's boot code. VentoyWorker.sh copies its first 446 bytes.
		Member: "./ventoy-1.1.17/boot/boot.img",
		Name:   bootName,
		SHA256: "f37cbea83596aef9812f4d984d344b5103913505dfee40dc0025742ea54a6113",
	},
	{
		// GRUB's core image, 2047 sectors, for sectors 1 to 2047.
		Member: "./ventoy-1.1.17/boot/core.img.xz",
		Name:   coreName,
		XZ:     true,
		SHA256: "b6581090947e7cacbd3cee23dfe2216aee9ab368c6508c2c5f3490621e969b84",
	},
	{
		// The VTOYEFI partition, 32 MB.
		Member: "./ventoy-1.1.17/ventoy/ventoy.disk.img.xz",
		Name:   efiName,
		XZ:     true,
		SHA256: "871f313d60d865a8ee307bc97c961e6cb619143288b4faf811efe9844ca1a003",
	},
}

// memberMax bounds each member as it is decompressed. The largest is
// VTOYEFI, 33,554,432 bytes.
const memberMax = 64 << 20
