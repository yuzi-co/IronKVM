#!/bin/sh
# Tests for release.sh.
#
# Only the parts that can run without a builder image and without a Sipeed base
# image: the argument handling, the manifest it writes, and the self-check that
# stops a mismatched release from being published.
#
# Needs openssl and tar.

set -u

HERE=$(cd "$(dirname "$0")" && pwd)
SCRIPT="$HERE/release.sh"
MANIFEST="$HERE/../abslots/manifest/root.manifest"
pass=0
fail=0

ok()    { pass=$((pass + 1)); echo "  ok    $1"; }
bad()   { fail=$((fail + 1)); echo "  FAIL  $1"; }
check() { if [ "$2" = "$3" ]; then ok "$1"; else bad "$1 (got '$2', want '$3')"; fi; }

skip_missing() {
    echo "$(basename "$0"): needs $1, which is not on PATH." >&2
    echo "the release host image carries it:" >&2
    echo "  docker build -t ironkvm-release-host tools/release" >&2
    echo "  docker run --rm -v \"$PWD:/repo\" -w /repo ironkvm-release-host sh $0" >&2
    exit 2
}

for t in openssl tar; do
    # Exit 2, not 0. A suite that ran no case has not passed, and a green
    # line for a suite that did nothing is worse than a red one: it is
    # counted as coverage that does not exist.
    command -v "$t" > /dev/null 2>&1 || { skip_missing "$t"; }
done

echo "release.sh"

# A version that is not X.Y.Z produces a package name the device refuses, and
# the refusal would happen on the device rather than here.
check "a two-component version is refused" \
    "$(sh "$SCRIPT" --dry-run 1.0 > /dev/null 2>&1; echo $?)" "1"

# A prerelease sorts BELOW the release it came from, so the update page would
# offer the older build as an upgrade for ever.
check "a prerelease version is refused" \
    "$(sh "$SCRIPT" --dry-run 1.0.0-rc1 > /dev/null 2>&1; echo $?)" "1"
# The shape check below would also reject it, for having a letter in it. What
# the dedicated check adds is a message that names the actual reason, so the
# message is what proves it ran.
check "it says why a prerelease is refused" \
    "$(sh "$SCRIPT" --dry-run 1.0.0-rc1 2>&1 | grep -c 'prerelease')" "1"

# Build metadata compares EQUAL under semver, so a release carrying it would be
# invisible to every device already on that version.
check "a version with build metadata is refused" \
    "$(sh "$SCRIPT" --dry-run '1.0.0+iron' > /dev/null 2>&1; echo $?)" "1"

check "no version at all is refused" \
    "$(sh "$SCRIPT" --dry-run > /dev/null 2>&1; echo $?)" "1"

# The manifest and the artifact have to agree. This repository has already lost
# three guards to rot, and a feed pointing at a package it does not describe is
# exactly the failure nobody notices until a device tries to install it.
build_fixture() {
    WORK=$(mktemp -d)
    mkdir -p "$WORK/out" "$WORK/stage/ironkvm_1.0.0"
    printf 'payload\n' > "$WORK/stage/ironkvm_1.0.0/marker"
    tar czf "$WORK/out/ironkvm_1.0.0.tar.gz" -C "$WORK/stage" ironkvm_1.0.0

    sum=$(openssl dgst -sha512 -binary "$WORK/out/ironkvm_1.0.0.tar.gz" | openssl base64 -A)
    size=$(wc -c < "$WORK/out/ironkvm_1.0.0.tar.gz" | tr -d ' ')
    cat > "$WORK/out/latest.json" <<EOF
{
  "manifest_version": 2,
  "version": "1.0.0",
  "name": "ironkvm_1.0.0.tar.gz",
  "url": "https://example.com/ironkvm_1.0.0.tar.gz",
  "sha512": "$sum",
  "size": $size,
  "size_bytes": $size,
  "unpacked_size_bytes": $((size * 2))
}
EOF
}

verify() {
    RELEASE_OUT="$WORK/out" sh "$SCRIPT" --verify-only 1.0.0 > "$WORK/verify.out" 2>&1
    echo $?
}

build_fixture
check "a matching manifest passes the self-check" "$(verify)" "0"

# Only the recorded checksum is wrong. The package keeps its size, its name and
# its top directory, so every other check still passes and this one is the only
# thing that can catch it.
#
# Two earlier versions of this case did not isolate it. Appending to the tarball
# changed the size, and corrupting it in place broke gzip so the top-directory
# check fired first. Both passed with the checksum comparison deleted.
sed -i 's|"sha512": ".*"|"sha512": "'"$(printf 'wrong' | openssl base64 -A)"'"|' \
    "$WORK/out/latest.json"
check "a mismatched sha512 fails the self-check" "$(verify)" "1"
rm -rf "$WORK"

# The updater derives the directory it expects from the package name, so a
# tarball whose top directory disagrees unpacks into something it refuses.
build_fixture
rm -f "$WORK/out/ironkvm_1.0.0.tar.gz"
mkdir -p "$WORK/stage/wrong_name"
printf 'payload\n' > "$WORK/stage/wrong_name/marker"
tar czf "$WORK/out/ironkvm_1.0.0.tar.gz" -C "$WORK/stage" wrong_name
sum=$(openssl dgst -sha512 -binary "$WORK/out/ironkvm_1.0.0.tar.gz" | openssl base64 -A)
size=$(wc -c < "$WORK/out/ironkvm_1.0.0.tar.gz" | tr -d ' ')
sed -i "s|\"sha512\": \".*\"|\"sha512\": \"$sum\"|; s|\"size\": [0-9]*|\"size\": $size|; s|\"size_bytes\": [0-9]*|\"size_bytes\": $size|" "$WORK/out/latest.json"
check "a wrong tarball top directory fails the self-check" "$(verify)" "1"
rm -rf "$WORK"

# A manifest naming a different package than the one built would send every
# device to a file that is not there.
build_fixture
sed -i 's|"name": "ironkvm_1.0.0.tar.gz"|"name": "ironkvm_9.9.9.tar.gz"|' "$WORK/out/latest.json"
check "a manifest naming another package fails the self-check" "$(verify)" "1"
rm -rf "$WORK"

# A size that disagrees with the artifact makes the device refuse the download
# after spending the bandwidth.
build_fixture
sed -i 's|"size_bytes": [0-9]*|"size_bytes": 12345|' "$WORK/out/latest.json"
check "a mismatched size_bytes fails the self-check" "$(verify)" "1"
rm -rf "$WORK"

# The image and the package must install the same boot scripts. Three of them
# live in tools/ and the image manifest adds them from there, so a package built
# from kvmapp/ alone would ship the scripts that can break a boot and leave out
# the watchdog that undoes them.
check "the payload picks up the boot scripts that live outside kvmapp" \
    "$(grep -c 'tools/abslots/device/S00awatchdog tools/service/S98supervise' "$SCRIPT")" "1"
check "the payload is checked against the image manifest" \
    "$(grep -c 'the image installs /etc/init.d' "$SCRIPT")" "1"

# The identity script is the reason a slot switch does not revert the root
# password to the factory one. An update that ships the fork's boot scripts and
# omits this one is the exact failure it was written to stop.
check "the identity script travels with the package" \
    "$(grep -c 'tools/abslots/device/S02identity' "$SCRIPT")" "1"

# rcS is the single exception, and it has to stay a named one. Every S* script
# is protected by the watchdog; rcS is what STARTS the watchdog, so a valid but
# broken rcS installed by an update leaves a board with no repair path and no
# recovery marker. The exception is only defensible while the reason is written
# down beside it.
check "rcS is held back from the package" \
    "$(grep -c '\[ "\$want" = rcS \] && continue' "$SCRIPT")" "1"
check "the reason rcS is held back is recorded" \
    "$(grep -c 'nothing would run at all, including the watchdog' "$SCRIPT")" "1"

# /kvmapp/system/init.d is the application's own reference copy and carries 20
# scripts. The image installs 10 of them, deliberately: it leaves S50sshd,
# S00kmod, S15kvmhwd and S80dnsmasq at their base versions and never installs
# avahi, ssdpd, tailscaled, picoclaw, wifi or usbhid at all. install.sh had no
# way to know that and installed the directory, so a package update would have
# added six daemons to a 166MB board that its own image never runs.
#
# The list is derived from the manifest so the two cannot name different sets.
check "the package names the boot scripts install.sh may install" \
    "$(grep -c '> "\$PAYLOAD/system/init\.d\.install"' "$SCRIPT")" "1"
check "an empty list is refused rather than shipped" \
    "$(grep -c '\[ -s "\$PAYLOAD/system/init\.d\.install" \]' "$SCRIPT")" "1"
check "the list is derived from the image manifest" \
    "$(grep -B3 '"\$PAYLOAD/system/init\.d\.install"' "$SCRIPT" | grep -c 'root\.manifest')" "1"
check "rcS is kept out of the list too" \
    "$(grep -c "grep -v '\^rcS\$'" "$SCRIPT")" "1"

# Every script the manifest installs must exist where the release script expects
# to find it. This is the check that catches a script being moved or renamed.
missing=0
for want in $(sed -n 's|^add \([^ ]*\) /etc/init.d/.*|\1|p' "$HERE/../abslots/manifest/root.manifest"); do
    [ -f "$HERE/../../$want" ] || missing=$((missing + 1))
done
check "every boot script the manifest names exists" "$missing" "0"

# root.manifest names paths that only exist together at the repository root:
# official-kvmapp/, kvmapp/, server/NanoKVM-Server, web/dist, tools/abslots/...
# Handing build-image.sh the staged package instead produces an image missing
# everything the manifest adds from tools/, and the build reports success.
check "the image is built from the repository root, not the package" \
    "$(grep -cE '^ +\. [0-9]+ "\$STAGE/(root|recovery)\.img"' "$SCRIPT")" "2"

# The manifest layers the official 2.5.0 tree under the fork's own, because the
# base rootfs carries 2.4.3. Without it the image ships an older application.
check "the official application is unpacked for the manifest" \
    "$(grep -q 'tar xzf "$OFFICIAL_APP" -C official-kvmapp' "$SCRIPT" && echo yes || echo no)" "yes"

# The package needs the SAME three layers as the image. The updater replaces
# /kvmapp rather than merging into it, so every file the package omits is a file
# the device loses. kvmapp/ alone carries 14 of the 37 libraries in
# server/dl_lib and neither kvm_system nor system/tool, which installs a server
# that cannot load libkvm.so onto a board that still answers ssh.
check "the package is layered the way the image is" \
    "$(grep -c 'cp -a official-kvmapp/\. "\$PAYLOAD/"' "$SCRIPT")" "1"
check "the fork's tree goes on top of the official one" \
    "$(grep -n 'cp -a official-kvmapp/\. "\$PAYLOAD/"\|cp -a kvmapp/\. *"\$PAYLOAD/"' "$SCRIPT" \
       | head -1 | grep -c official-kvmapp)" "1"

# Two builds never collide on a hashed asset name, so a merge would leave the
# official bundle's files beside the fork's and serve both.
check "the web directory is replaced, not merged" \
    "$(grep -c 'rm -rf "\$PAYLOAD/server/web"' "$SCRIPT")" "1"

# The guard that would have caught the missing libraries. A grep proves the
# layering is written; this proves the build refuses to ship without it.
check "the package is checked against the official one" \
    "$(grep -c 'the package drops files the official one carries' "$SCRIPT")" "1"

# repack-boot.sh writes boot.sd.new, deliberately: the name says the image has
# not been accepted yet. Asking for boot.sd fails after every verification in
# that script has already passed.
check "the repacked boot image is taken by the name it is written under" \
    "$(grep -c 'bootbuild/boot\.sd\.new' "$SCRIPT")" "1"

# /boot/ver comes from the Sipeed base and reports v1.4.3 whatever is written
# over it, so the card image was the one artefact a board could not name. The
# application version cannot stand in: an update replaces the application and
# leaves the card alone, which is exactly the half that carries the slots, the
# watchdog and the recovery filesystem.
check "the card image stamps its own version onto /boot" \
    "$(grep -c '> "\$STAGE/boot/ironkvm\.ver"' "$SCRIPT")" "1"
check "the stamp is written before the card is assembled" \
    "$(grep -n 'ironkvm\.ver"\|build-card\.sh "\$STAGE/boot"' "$SCRIPT" \
       | head -1 | grep -c 'ironkvm\.ver')" "1"

# /boot/ironkvm.ver names the card. It does not name the application, and the
# application version is the one the update page compares. The card image is
# built from the repository root and not from the package, so the package's own
# version file never reaches it, and a flashed board reported the official
# application version instead: 2.5.0 against a feed at 1.0.1, which semver.gte
# reads as up to date for ever. Measured on a flashed card on 2026-08-22.
check "the image tree is stamped with the release version" \
    "$(grep -c '^echo "\$VERSION" > kvmapp/version$' "$SCRIPT")" "1"
check "the image tree records the base it came from" \
    "$(grep -c '^echo "\$BASE"    > kvmapp/base-version$' "$SCRIPT")" "1"

# Both writes have to precede the image build, or they land in the tree after
# the manifest has copied it and the card carries the value they replace.
check "the version reaches the tree before the slots are built" \
    "$(grep -n '> kvmapp/version$\|build-image\.sh "\$BASE_TAR"' "$SCRIPT" \
       | head -1 | grep -c 'kvmapp/version')" "1"
check "the base version reaches the tree before the slots are built" \
    "$(grep -n '> kvmapp/base-version$\|build-image\.sh "\$BASE_TAR"' "$SCRIPT" \
       | head -1 | grep -c 'kvmapp/base-version')" "1"

# root.manifest merges kvmapp/ into /kvmapp after official-kvmapp/, and that
# order is the only reason no manifest line is needed. If it ever reverses, the
# official version file covers the one written above and the fault returns.
check "the manifest layers the fork over the official application" \
    "$(grep -n '^add  *official-kvmapp/\|^add  *kvmapp/ ' \
       "$HERE/../abslots/manifest/root.manifest" | head -1 \
       | grep -c official-kvmapp)" "1"

# The fork's whole native-library delta is two files. Everything else in
# server/dl_lib is byte-identical to the official 2.5.0 release, so these two
# are the only ones a build can lose without changing anything visible.
#
# They used to reach both artifacts through kvmapp/server/dl_lib, which
# .gitignore excludes and git therefore does not carry. On the workstation that
# built 1.0.1 the directory happened to hold the right files. On a fresh clone
# it is empty, and the build would ship Sipeed's libkvm.so with the fork's
# server: the H.264 encoder strands its carveout on every stop, and no step
# fails.
check "the package takes the fork libraries from the tracked path" \
    "$(grep -c '^cp server/dl_lib/libkvm\.so server/dl_lib/libkvm_mmf\.so "\$PAYLOAD/server/dl_lib/"$' "$SCRIPT")" "1"
check "and reads back what was assembled rather than trusting the copy" \
    "$(grep -c 'is not the one server/dl_lib holds' "$SCRIPT")" "1"

# The card image is built from the manifest and never sees $PAYLOAD, so it needs
# its own line for each.
check "the image names libkvm.so from server/dl_lib" \
    "$(grep -c '^add  *server/dl_lib/libkvm\.so ' "$MANIFEST")" "1"
check "the image names libkvm_mmf.so from server/dl_lib" \
    "$(grep -c '^add  *server/dl_lib/libkvm_mmf\.so ' "$MANIFEST")" "1"

# Order decides which copy survives. kvmapp/ merges into the same /kvmapp, so a
# line above it would be overwritten by whatever the ignored directory holds,
# which is the state this replaced.
check "the image lines come after the kvmapp merge" \
    "$(grep -n '^add  *kvmapp/ \|^add  *server/dl_lib/libkvm\.so ' "$MANIFEST" \
       | head -1 | grep -c 'kvmapp/ ')" "1"


echo
echo "release notes"

# publish() passes --notes-file $OUT/notes.md to gh, and for four releases
# nothing wrote that file. The order is what makes it expensive: the tag is
# pushed first, so a missing notes file leaves a tag on the remote with no
# release under it and a feed still naming the version before.
#
# So the file is checked in the preflight, before anything is built, and it is a
# file in the repository rather than prose typed at the prompt.
check "the notes are checked before anything is built" \
    "$(grep -n 'NOTES_SRC is missing or empty\|^echo "==> building the slot filesystems"' "$SCRIPT" \
       | head -1 | grep -c 'NOTES_SRC')" "1"
check "a release without notes is refused" \
    "$(grep -c 'write the release notes first' "$SCRIPT")" "1"
check "the notes come from the repository, named by version" \
    "$(grep -c '^NOTES_SRC="tools/release/notes/\${VERSION}\.md"$' "$SCRIPT")" "1"
check "the notes are staged beside the artifacts" \
    "$(grep -c 'cp "\$NOTES_SRC" "\$OUT/notes\.md"' "$SCRIPT")" "1"

# A dry run must not demand them, the same way it does not demand gh, or the
# only way to read the notes before publishing is to write them under pressure.
check "a dry run does not demand notes" \
    "$(grep -c '\[ "\$DRY_RUN" = "1" \] || \[ -s "\$NOTES_SRC" \]' "$SCRIPT")" "1"

# set -eu is on, so `[ -s x ] && cp` aborts the whole run when x is absent.
# That is exactly the dry run the check above is meant to allow.
check "the staging copy cannot abort a run that has no notes" \
    "$(grep -c '^\[ -s "\$NOTES_SRC" \] && cp' "$SCRIPT")" "0"

# The notes for the version being released now.
check "1.0.1 has notes to publish" \
    "$([ -s "$HERE/notes/1.0.1.md" ] && echo yes || echo no)" "yes"
check "and they say the 1.0.0 card image is superseded" \
    "$(grep -ci 'superseded' "$HERE/notes/1.0.1.md")" "1"
# The base is Sipeed's, and the system image ships no checksum of its own, so the
# pin in BASE.sha256 is the only statement of which bytes an image came from.
check "the base is verified against the recorded pins" \
    "$(grep -q 'which is not pinned in tools/abslots/BASE.sha256' "$SCRIPT" && echo yes || echo no)" "yes"

# make is not installed on every host that has Docker and the builder image.
check "the build does not depend on make" \
    "$(grep -c '^make ' "$SCRIPT")" "0"

echo
echo "fetch-base.sh"

FETCH="$HERE/fetch-base.sh"
PINS="$HERE/../abslots/BASE.sha256"

# The pin lookup has to resolve against the real file, not a fixture. A grep that
# silently matches nothing would let the script "verify" every download.
for name in 20260610_NanoKVM_Rev1_4_3.img nanokvm_2.5.0.tar.gz nanokvm-base-official.tar.zst; do
    got=$(grep "  $name\( \|\$\)" "$PINS" | grep -o '^[0-9a-f]\{64\}' | head -1)
    check "a pin resolves for $name" "$(printf '%s' "$got" | wc -c | tr -d ' ')" "64"
done

# The derived tarball is the one that could differ without anybody noticing: it
# is produced here rather than downloaded, so its pin is what says the extraction
# reproduced the bytes the fork was developed against.
check "the extracted root filesystem is verified too" \
    "$(grep -q 'check "\$DEST/rootfs.tar.zst" nanokvm-base-official.tar.zst' "$FETCH" && echo yes || echo no)" "yes"

# Sipeed's layout is not the fork's, and it is not partition.sfdisk.
check "the offsets are read from the image, not assumed" \
    "$(grep -c 'fdisk -l -o Device,Start' "$FETCH")" "2"

check "a missing pin refuses rather than passing" \
    "$(grep -q 'no pin recorded for' "$FETCH" && echo yes || echo no)" "yes"

echo
echo "release.sh preflight"

# A build that finds a missing tool at its last step has already spent twenty
# minutes and leaves half a release behind. That is exactly how the signing step
# would have failed: it was the final command, and the signer was not installed.
check "the host is checked before anything is built" \
    "$(grep -q '==> checking the host' "$SCRIPT" && echo yes || echo no)" "yes"
check "the tool check runs before the web build" \
    "$(grep -n 'checking the host\|building the web user interface' "$SCRIPT" \
       | head -1 | grep -c 'checking the host')" "1"
check "the builder image is checked, not assumed" \
    "$(grep -q 'docker image inspect' "$SCRIPT" && echo yes || echo no)" "yes"
check "a dry run does not demand gh" \
    "$(grep -q 'DRY_RUN" = "1" \] || NEED="\$NEED gh"' "$SCRIPT" && echo yes || echo no)" "yes"

# The builder image bakes the ownership of its home directory in at build time.
# A host whose own id differs from it runs the build as a user that cannot write
# $HOME, and go stops at the module cache. The identity must therefore be
# overridable separately from whatever id the host reports.
check "the build identity can be overridden" \
    "$(grep -c '^BUILD_UID=\${BUILD_UID:-\$(id -u)}' "$SCRIPT")" "1"
check "the builder image name follows the build identity" \
    "$(grep -c 'nanokvm-builder-local-\${BUILD_UID}-\${BUILD_GID}' "$SCRIPT")" "1"
check "the host id does not reach the builder directly" \
    "$(grep -c 'e UID="\$(id -u)"' "$SCRIPT")" "0"

# pnpm asks before it purges a modules directory another platform installed, and
# it refuses to purge without a TTY. A release runs from a script, so the build
# stopped there with the web user interface unbuilt and nothing else attempted.
check "the web build never waits for an answer" \
    "$(grep -c 'CI=true pnpm install --frozen-lockfile' "$SCRIPT")" "1"

# web/node_modules belongs to the developer. Installing into it wrote it for
# whatever host built the release, so a Windows workstation that also runs
# `pnpm dev` had its modules directory replaced with Linux binaries every time,
# and putting it back is a step somebody has to remember.
check "the web user interface is built from a copy" \
    "$(grep -c 'cd "\$STAGE/web" && CI=true pnpm install' "$SCRIPT")" "1"
check "the developer's modules directory is never installed into" \
    "$(grep -c 'cd web && CI=true pnpm install' "$SCRIPT")" "0"
check "node_modules is left out of the copy" \
    "$(grep -c 'exclude=\./node_modules' "$SCRIPT")" "1"

# root.manifest reads web/dist from the repository root, so the build output has
# to come back. It is plain JavaScript with no platform in it.
check "the built output is copied back to web/dist" \
    "$(grep -c 'cp -a "\$STAGE/web/dist" web/dist' "$SCRIPT")" "1"

# Signing was designed in and dropped for 1.0. The checksums must not claim to
# be more than they are, and nothing may reference a signature that is not made.
check "no signature is produced" \
    "$(grep -c 'minisign' "$SCRIPT")" "0"
check "no signature is published" \
    "$(grep -c 'minisig' "$SCRIPT")" "0"
check "the checksums say what they do not prove" \
    "$(grep -q 'nothing about who produced it' "$SCRIPT" && echo yes || echo no)" "yes"

# Dry run is what makes the rest of this script testable at all, so it is the
# guard worth asserting hardest.
check "publishing is guarded by the dry-run flag" \
    "$(sed -n '/^publish()/,/^}/p' "$SCRIPT" | grep -c 'DRY_RUN')" "1"
check "nothing outside publish() calls gh" \
    "$(grep -c '^[^#]*gh release' "$SCRIPT")" "1"

# build-card.sh creates the card at exactly its final size, because the table
# declares no data partition and there is no hole to cut back. The card is still
# assembled on the build host's own filesystem rather than in the output
# directory: the image is 5 GiB and both slots are read back out of it to be
# checked, so the same bytes cross the filesystem three times, and the output
# directory on a Windows workstation is a bind mount.
check "the card is assembled on the build host's own filesystem" \
    "$(grep -c '"\$STAGE/ironkvm-\${VERSION}-sdcard\.img"' "$SCRIPT")" "2"
check "the uncompressed card never lands in the output directory" \
    "$(grep -c '"\$OUT/ironkvm-\${VERSION}-sdcard\.img"' "$SCRIPT")" "0"
check "only the compressed image is moved to the output directory" \
    "$(grep -c '^mv "\$STAGE/ironkvm-\${VERSION}-sdcard\.img\.xz" "\$OUT/\$IMG"' "$SCRIPT")" "1"

echo
echo "the release host image"

# The image has to satisfy the preflight it will be run against. Reading the
# tool list out of release.sh rather than repeating it here is the point: a tool
# added to one and not the other is exactly the drift this catches, and gh was
# already missing when the image was first written.
if command -v docker > /dev/null 2>&1 \
   && docker image inspect ironkvm-release-host > /dev/null 2>&1; then
    want=$(sed -n 's/^NEED="\(.*\)"$/\1/p' "$SCRIPT" | head -1)
    want="$want gh sfdisk mkfs.vfat mcopy mke2fs e2fsck zstd cpio mkimage"
    absent=$(docker run --rm ironkvm-release-host sh -c \
        "for t in $want; do command -v \$t > /dev/null 2>&1 || echo \$t; done" \
        2>/dev/null | tr '\n' ' ' | sed 's/ *$//')
    check "the image carries every tool a release needs" "$absent" ""
else
    echo "  skip  the image carries every tool a release needs (no image built)"
fi

echo
echo "passed $pass, failed $fail"
[ "$fail" -eq 0 ]
