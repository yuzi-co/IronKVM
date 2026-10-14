package tailscale

// The Tailscale release the add-on installs, pinned by version and SHA-256.
// It is not in the image: the owner installs it from the web UI, and the
// server downloads the static riscv64 package from pkgs.tailscale.com to
// /data, checks it, takes the two binaries out of it, and removes it.
// Tailscale is BSD-3-Clause. A new release is a new pin here, its sum
// computed from a download and compared with the .sha256 the release server
// publishes beside the package. The web UI's manual install help
// (settings/tailscale/install-help.tsx) links the same package.

// pinnedFile is a file downloaded as it is published.
type pinnedFile struct {
	URL    string
	SHA256 string
	// Max bounds the download. The package is 37,978,637 bytes.
	Max int64
}

// Version is the pinned Tailscale release.
const Version = "1.104.1"

// releasePackage is a variable only so the tests can serve their own file;
// nothing else changes it.
var releasePackage = pinnedFile{
	URL:    "https://pkgs.tailscale.com/stable/tailscale_1.104.1_riscv64.tgz",
	SHA256: "a434d46c212a22f7de897ce3573083739661e7afd7f0ca8b921105c6f966eb13",
	Max:    64 << 20,
}
