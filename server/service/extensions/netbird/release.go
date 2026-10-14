package netbird

// The NetBird package the add-on installs, pinned by version and SHA-256.
// NetBird publishes no riscv64 build of its own, so the pin is Alpine's
// package of it: netbird 0.78.2-r0 from v3.24 community, built from aports
// commit 83cf529a483bcd6efe3180a094a2de8a7cde7446. It is not in the image:
// the owner installs it from the web UI, and the server downloads the
// package to /data, checks it against the pin, has apk check Alpine's
// signature while it takes usr/bin/netbird out of it, and removes it. The
// NetBird client is BSD-3-Clause. A new release is a new pin here, its sum
// computed from a download whose signature apk verify accepted.
//
// Alpine keeps one build of a package in a branch: when v3.24 takes a newer
// one, this file leaves the mirror and the install fails, naming the URL,
// until the pin moves.

// pinnedFile is a file downloaded as it is published.
type pinnedFile struct {
	URL    string
	SHA256 string
	// Max bounds the download. The package is 15,313,755 bytes.
	Max int64
}

// Version is the pinned NetBird release, as `netbird version` reports it.
const Version = "0.78.2"

// releasePackage is a variable only so the tests can serve their own file;
// nothing else changes it.
var releasePackage = pinnedFile{
	URL:    "https://dl-cdn.alpinelinux.org/alpine/v3.24/community/riscv64/netbird-0.78.2-r0.apk",
	SHA256: "b082b6d0f973f4117202268bacc613c6c3fc682266f75e4fa40411252db3b441",
	Max:    32 << 20,
}

// packageFile is the name the package is saved under in the workspace.
const packageFile = "netbird-0.78.2-r0.apk"
