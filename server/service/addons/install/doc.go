// Package install puts an add-on the owner installs from the web UI on /data:
// it checks the free space, downloads the pinned artifact through
// upstream.Client (bounded and checked against the pinned SHA-256), has the
// upstream's own check run where there is one (Alpine's signature, through
// apk), takes out only the members the manifest lists, each checked against
// its own sum, and then packs them into one read-only squashfs image.
//
// An artifact with no pin can come in through Options.Fetch, which downloads
// and checks it against whatever vouches for it instead: NetBird's package,
// when Alpine has replaced the pinned build, is checked against the branch's
// signed index before apk checks its own signature.
//
// It replaces five copies of that code, in the Tailscale, NetBird, PicoClaw,
// netboot and Ventoy packages, which now keep only their manifests and their
// configuration. See ironkvm-dist's
// docs/notes/2026-10-27-addons-lxc-ondemand.md, section 5 and "Decisions
// 2026-10-28".
//
// Tailscale, NetBird and PicoClaw go through Install and become images.
// netboot and Ventoy, the data add-ons, use only ExtractTarGz, ExtractApk and
// upstream.Client for now, and keep their files as plain files: their
// updaters write releases newer than the pin into the same directories, and
// netboot writes its generated boot.ipxe into its TFTP root, so packing them
// waits for the updaters to produce images too.
//
// # Layout
//
// An add-on <name> lives in /data/ironkvm/addons/<name>/ (addon.Dir):
//
//	<name>-<version>.sqfs          the add-on's files: a squashfs image, LZ4
//	<name>-<version>.sqfs.sha256   its SHA-256, "<sum>  <file>", checked before every mount
//	current                        the file name of the image in use, one line
//	sqfs/                          where the image named in current is mounted, read-only
//	installed.json                 the manifest, the members' sums, source URL and licence
//	installed.previous.json        the same for the image an update replaced
//	licenses/                      the licence files, copied out of the image
//	links binds initd enabled      the addon package's files, unchanged
//	home/, ...                     writable state, plain directories beside the image
//	.fetch/                        the workspace of an install in progress
//
// The files are packed on the board, from the upstream's own download, so the
// upstream remains the one who conveys them. LZ4 because the pages of a
// program are read back in from the image whenever memory runs short, and LZ4
// is the cheapest of the kernel's decompressors.
//
// The whole image is checked against its sidecar before every mount: here,
// before the mount that follows an install, and by ironkvm-dist's S04addons at
// every boot. /data is exFAT, which anybody holding the card can write, and
// nothing checked a binary on it before it ran until now. Links name a file in
// the image as sqfs/<file>, which addon.Record and S04addons both accept, so
// links, binds and the boot script keep the contract they had.
//
// An update packs the new version beside the current one, switches current,
// and mounts the new image at once, with no reboot; the previous image stays
// until the next update, so Rollback can switch back to it. Removal is
// addon.Remove, which detaches the image and removes the directory.
//
// # Fallback
//
// The image needs squashfs in the kernel and mksquashfs on the slot. The
// vendor kernel of slot A has squashfs; the mainline kernel of slot B does not
// until its configuration gains CONFIG_SQUASHFS, and an image built before
// squashfs-tools joined the package list has no mksquashfs. Without either,
// and off a distribution image, Install hands the checked files to the
// caller's Plain function, which places them exactly as the add-on did before
// images, so nothing that worked regresses.
package install
