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
pass=0
fail=0

ok()    { pass=$((pass + 1)); echo "  ok    $1"; }
bad()   { fail=$((fail + 1)); echo "  FAIL  $1"; }
check() { if [ "$2" = "$3" ]; then ok "$1"; else bad "$1 (got '$2', want '$3')"; fi; }

for t in openssl tar; do
    command -v "$t" > /dev/null 2>&1 || { echo "SKIP: $t is missing"; exit 0; }
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

echo
echo "passed $pass, failed $fail"
[ "$fail" -eq 0 ]
