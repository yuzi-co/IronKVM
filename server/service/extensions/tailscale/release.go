package tailscale

import installer "NanoKVM-Server/service/addons/install"

// The Tailscale release the add-on installs, pinned by version and SHA-256.
// It is not in the image: the owner installs it from the web UI, and the
// server downloads the static riscv64 package from pkgs.tailscale.com to
// /data, checks it, takes the two binaries out of it, each checked against
// its own sum, and packs them into the add-on's image (service/addons/install).
// Tailscale is BSD-3-Clause. A new release is a new pin here, its sum
// computed from a download and compared with the .sha256 the release server
// publishes beside the package, and the binaries' sums taken from the
// unpacked download. The web UI's manual install help
// (settings/tailscale/install-help.tsx) links the same package.

// Version is the pinned Tailscale release.
const Version = "1.104.1"

// release is a variable only so the tests can serve their own file; nothing
// else changes it. The package is 37,978,637 bytes, and the two binaries
// 68,402,398 together.
var release = installer.Manifest{
	Name:    "tailscale",
	Title:   "Tailscale",
	Version: Version,
	Licence: "BSD-3-Clause",
	Source:  "https://github.com/tailscale/tailscale/tree/v1.104.1",
	Artifact: installer.Artifact{
		URL:      "https://pkgs.tailscale.com/stable/tailscale_1.104.1_riscv64.tgz",
		SHA256:   "a434d46c212a22f7de897ce3573083739661e7afd7f0ca8b921105c6f966eb13",
		Size:     37978637,
		Unpacked: 68402398,
		Max:      64 << 20,
		Format:   installer.FormatTarGz,
		Members: []installer.Member{
			{
				Path:   "tailscale_1.104.1_riscv64/tailscale",
				Name:   "tailscale",
				SHA256: "668c98e9390b194d72b73325814eb0b84f02682366f555b695c298509bdd0abc",
				Mode:   0o755,
			},
			{
				Path:   "tailscale_1.104.1_riscv64/tailscaled",
				Name:   "tailscaled",
				SHA256: "657f77f8ce21bdc1cd2905f7793ee0754addebd8438393ddf43a6e59763670cd",
				Mode:   0o755,
			},
		},
	},
}
