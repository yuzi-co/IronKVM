#!/bin/sh
#
# Build base/ from Sipeed's published downloads.
#
#   fetch-base.sh [target-dir]      default: ./base
#
# Needs curl, zstd, and the ability to loop-mount (so: Linux, as root, or a
# privileged container). Downloads about 1.6 GB and leaves about 280 MB.
#
# WHY THIS EXISTS RATHER THAN A HOSTED COPY
#
# The base is Sipeed's build. Its GPL parts could be redistributed, but the
# image also carries vendor binaries for the SG2002 whose terms are not stated,
# so this fork does not republish it. Fetching from the publisher and checking
# the bytes against tools/abslots/BASE.sha256 gets the same reproducibility with
# no licensing question to answer on somebody else's behalf.
#
# Every artifact is verified before it is used. The application tarball has a
# checksum published beside it; the system image does not, so BASE.sha256 records
# its hash as a pin. A pin is weaker than a publisher's signature and is recorded
# as such, but it is what makes two builds comparable.
set -eu

HERE=$(cd "$(dirname "$0")" && pwd)
ROOT=$(cd "$HERE/../.." && pwd)
DEST=${1:-$ROOT/base}
PINS="$ROOT/tools/abslots/BASE.sha256"

IMG_NAME=20260610_NanoKVM_Rev1_4_3.img
APP_NAME=nanokvm_2.5.0.tar.gz
IMG_URL=${IMG_URL:-https://github.com/sipeed/NanoKVM/releases/download/v1.4.3/$IMG_NAME}
APP_URL=${APP_URL:-https://github.com/sipeed/NanoKVM/releases/download/2.5.0/$APP_NAME}

[ -f "$PINS" ] || { echo "no pin file at $PINS" >&2; exit 1; }

for t in curl zstd tar sha256sum; do
    command -v "$t" > /dev/null 2>&1 || { echo "$t is required" >&2; exit 1; }
done

pin_for() {
    grep "  $1\( \|$\)" "$PINS" | grep -o '^[0-9a-f]\{64\}' | head -1
}

# check verifies a file against the pin recorded for its NAME, and refuses to
# continue without one. A download that is merely present proves nothing: the
# reason this script exists is that two builds must be comparable.
check() {
    want=$(pin_for "$2")
    [ -n "$want" ] || { echo "no pin recorded for $2" >&2; return 1; }
    have=$(sha256sum "$1" | cut -d' ' -f1)
    [ "$have" = "$want" ] || {
        echo "$2 is $have, pinned as $want" >&2
        return 1; }
    echo "    $2 matches its pin"
}

mkdir -p "$DEST"
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

echo "==> fetching the official application"
curl -fSL --retry 3 -o "$WORK/$APP_NAME" "$APP_URL"
check "$WORK/$APP_NAME" "$APP_NAME"
cp "$WORK/$APP_NAME" "$DEST/$APP_NAME"

echo "==> fetching the official system image (about 1.6 GB)"
curl -fSL --retry 3 -o "$WORK/$IMG_NAME" "$IMG_URL"
check "$WORK/$IMG_NAME" "$IMG_NAME"

echo "==> reading the image's own partition table"
# The offsets are read from the image rather than assumed. This is Sipeed's
# layout, not the fork's, and it is not the table in partition.sfdisk.
P1_START=$(fdisk -l -o Device,Start "$WORK/$IMG_NAME" 2>/dev/null | awk '/img1/ {print $2}')
P2_START=$(fdisk -l -o Device,Start "$WORK/$IMG_NAME" 2>/dev/null | awk '/img2/ {print $2}')
[ -n "$P1_START" ] && [ -n "$P2_START" ] || {
    echo "could not read the image's partition table" >&2; exit 1; }
echo "    p1 at sector $P1_START, p2 at sector $P2_START"

echo "==> extracting /boot"
mkdir -p "$WORK/p1" "$DEST/boot"
mount -o loop,offset=$((P1_START * 512)),ro "$WORK/$IMG_NAME" "$WORK/p1"
cp -a "$WORK/p1/." "$DEST/boot/"
umount "$WORK/p1"
[ -f "$DEST/boot/fip.bin" ] || { echo "no fip.bin in the boot partition" >&2; exit 1; }
[ -f "$DEST/boot/boot.sd" ] || { echo "no boot.sd in the boot partition" >&2; exit 1; }

echo "==> extracting the root filesystem"
mkdir -p "$WORK/p2"
mount -o loop,offset=$((P2_START * 512)),ro "$WORK/$IMG_NAME" "$WORK/p2"
# --numeric-owner so the archive records uids rather than this host's names.
( cd "$WORK/p2" && tar --numeric-owner -cf - . ) | zstd -q -o "$DEST/rootfs.tar.zst"
umount "$WORK/p2"

echo "==> verifying the extracted root filesystem"
# This one is derived rather than downloaded, so the pin is what says the
# extraction produced the same bytes as the build the fork was developed
# against. tar is not deterministic across every option set, so a mismatch here
# means the extraction differed, not that the image did.
check "$DEST/rootfs.tar.zst" nanokvm-base-official.tar.zst || {
    echo "the extraction did not reproduce the pinned tarball" >&2
    exit 1; }

# The application version, which travels beside the fork's own version because
# semver cannot carry it.
echo "$APP_NAME" | sed 's/^nanokvm_//; s/\.tar\.gz$//' > "$DEST/version"

echo
echo "base/ is ready in $DEST"
ls -la "$DEST"
