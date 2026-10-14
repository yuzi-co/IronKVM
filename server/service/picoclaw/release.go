package picoclaw

// The PicoClaw the add-on installs, pinned by version and SHA-256. It is not
// in the image: the owner installs it from the web UI, and the server
// downloads the release asset of IronKVM's fork to /data, checks it, takes the
// binary and its license files out of it, and removes it, as it does for
// Ventoy, netboot.xyz, Tailscale and NetBird.
//
// The asset is built from the fork's `ironkvm` branch
// (https://github.com/yuzi-co/picoclaw) by ironkvm-dist's
// dist/release-picoclaw.sh: the fork's scripts/build-ironkvm.sh, which is
// reproducible, packed with the fork's LICENSE, Go's LICENSE and the notices of
// the Go modules the binary links. ironkvm-dist's dist/picoclaw.pin holds the
// same numbers, and a rebuild of the commit gives the same two sums. PicoClaw
// is MIT. A new release is a new pin here and there.

// pinnedRelease is the release asset and the binary inside it.
type pinnedRelease struct {
	URL    string
	SHA256 string
	// Size is the asset's size, for the free space check before the
	// download; Max bounds the download itself.
	Size int64
	Max  int64
	// Dir is the one directory the archive holds, and the binary is
	// Dir/picoclaw, with BinarySHA256 and BinarySize.
	Dir          string
	BinarySHA256 string
	BinarySize   int64
}

// Version is the pinned build, the fork's `git describe` of its commit, which
// the binary also reports.
const Version = "v0.3.1-42-gac4740b9"

// release is a variable only so the tests can serve their own archive;
// nothing else changes it.
var release = pinnedRelease{
	URL:          "https://github.com/yuzi-co/picoclaw/releases/download/ironkvm-v0.3.1-42-gac4740b9/picoclaw-v0.3.1-42-gac4740b9-linux-riscv64-ironkvm.tar.gz",
	SHA256:       "e0cace9a384e5f13cd8829c427b00846c2e3a5dd226ec3165d8b491b4456dfd9",
	Size:         5986607,
	Max:          32 << 20,
	Dir:          "picoclaw-v0.3.1-42-gac4740b9-linux-riscv64-ironkvm",
	BinarySHA256: "bd3167405cdb8b05470c3699177d0dd6eee521c8c767b2e7486a719a00c51588",
	BinarySize:   15401144,
}
