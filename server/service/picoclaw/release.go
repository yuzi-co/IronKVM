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

// Version is the pinned build: the fork's release tag on its commit,
// <upstream>-ironkvm.N, which is also the `git describe` the binary
// reports.
const Version = "v0.3.1-ironkvm.1"

// release is a variable only so the tests can serve their own archive;
// nothing else changes it.
var release = pinnedRelease{
	URL:          "https://github.com/yuzi-co/picoclaw/releases/download/v0.3.1-ironkvm.1/picoclaw-v0.3.1-ironkvm.1-linux-riscv64.tar.gz",
	SHA256:       "c3ca1a67549626e3f08260d54c13cb62cd1793c69e5fc0c1872c55d958a61294",
	Size:         5986593,
	Max:          32 << 20,
	Dir:          "picoclaw-v0.3.1-ironkvm.1-linux-riscv64",
	BinarySHA256: "500dffa82a467324da67d617ec7c7475ac15bdb8e34e7bd7b823ca579f5712a7",
	BinarySize:   15401144,
}
