#!/bin/sh
# Build one ext4 root image from a pinned base plus a manifest.
#
#   build-image.sh <base.tar.zst> <manifest> <payload-dir> <size-mib> <out.img>
#
# Needs mke2fs, e2fsck, tar and zstd. Run it in a container or on the Unraid
# host. Nothing here runs on the device.
#
# mke2fs -d populates the image from a directory with no loop mount and no
# privileged mount, which is what makes this buildable through Docker on
# Windows at all.
#
# The manifest has three verbs and the order they run in matters:
#
#   remove <path in the image>                    runs first
#   add    <path in the payload> <path in image>  runs second
#   touch  <path in the image>                    runs last
#
# remove before add, so a manifest can replace a directory by removing it and
# adding a new one. touch last, because a marker may live inside a directory
# that add created.
#
# The remove list is the half that matters. cp adds and overwrites and never
# deletes, so a file the base has and the fork does not will survive into every
# image built from it. That is exactly how /etc/kvm/ssh_stop reaches a slot and
# stops sshd while reporting success, and that file is present in the official
# v1.4.3 rootfs.
set -e

BASE=${1:?usage: build-image.sh <base.tar.zst> <manifest> <payload-dir> <size-mib> <out.img>}
MANIFEST=${2:?usage: build-image.sh <base.tar.zst> <manifest> <payload-dir> <size-mib> <out.img>}
PAYLOAD=${3:?usage: build-image.sh <base.tar.zst> <manifest> <payload-dir> <size-mib> <out.img>}
SIZE_MIB=${4:?usage: build-image.sh <base.tar.zst> <manifest> <payload-dir> <size-mib> <out.img>}
OUT=${5:?usage: build-image.sh <base.tar.zst> <manifest> <payload-dir> <size-mib> <out.img>}

[ -f "$BASE" ]     || { echo "no such base: $BASE"; exit 1; }
[ -f "$MANIFEST" ] || { echo "no such manifest: $MANIFEST"; exit 1; }
[ -d "$PAYLOAD" ]  || { echo "no such payload directory: $PAYLOAD"; exit 1; }

STAGE=$(mktemp -d)
trap 'rm -rf "$STAGE"' EXIT

echo "############ 1. unpack the base"
mkdir -p "$STAGE/tree"
zstd -dc "$BASE" | tar --numeric-owner -xf - -C "$STAGE/tree"
printf '  %-40s %s files\n' "base" "$(find "$STAGE/tree" -type f | wc -l)"

echo
echo "############ 2. apply the manifest"
grep -v '^[[:space:]]*#' "$MANIFEST" | grep -v '^[[:space:]]*$' > "$STAGE/m" || true

while read -r verb a b; do
    [ "$verb" = remove ] || continue
    rm -rf "$STAGE/tree${a}"
    echo "  remove  $a"
done < "$STAGE/m"

while read -r verb a b; do
    [ "$verb" = add ] || continue
    if [ ! -e "$PAYLOAD/$a" ]; then
        echo "  add     $a -> MISSING in $PAYLOAD"
        echo
        echo "the payload does not carry everything the manifest adds; no image written"
        exit 1
    fi
    mkdir -p "$(dirname "$STAGE/tree${b}")"
    # A trailing slash on both sides means "merge this directory into that one",
    # which is how the official application and the fork's own copy of /kvmapp
    # are layered without one erasing the other.
    case "$a$b" in
        */*/) cp -a "$PAYLOAD/${a%/}/." "$STAGE/tree${b}" ;;
        *)    rm -rf "$STAGE/tree${b}"; cp -a "$PAYLOAD/$a" "$STAGE/tree${b}" ;;
    esac
    echo "  add     $a -> $b"
done < "$STAGE/m"

while read -r verb a b; do
    [ "$verb" = touch ] || continue
    mkdir -p "$(dirname "$STAGE/tree${a}")"
    : > "$STAGE/tree${a}"
    echo "  touch   $a"
done < "$STAGE/m"

echo
echo "############ 3. record provenance"
# This is the question nobody could answer during the 2026-08-15 outage: what
# does that slot contain? `slot status` reads it back.
mkdir -p "$STAGE/tree/etc"
{
    echo "base $(basename "$BASE")"
    echo "basesha $(sha256sum "$BASE" | cut -d' ' -f1)"
    echo "manifest $(basename "$MANIFEST")"
    echo "manifestsha $(sha256sum "$MANIFEST" | cut -d' ' -f1)"
    echo "size ${SIZE_MIB}m"
} > "$STAGE/tree/etc/slot-manifest"
sed 's/^/  /' "$STAGE/tree/etc/slot-manifest"

echo
echo "############ 4. device names, written once"
# No script computes a device from a partition number. In this layout p3 is a
# root filesystem, and S01fs used to run mkfs.exfat on p3.
cat > "$STAGE/tree/etc/nanokvm-slots.conf" <<'CONF'
SLOT_A=/dev/mmcblk0p2
SLOT_B=/dev/mmcblk0p3
RECOVERY=/dev/mmcblk0p5
DATA_DEV=/dev/mmcblk0p6
CONF

echo
echo "############ 5. gates"
fail=0
note() { printf '  %-58s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fail=1; return 0; }

crlf=0
if [ -d "$STAGE/tree/etc/init.d" ]; then
    for f in "$STAGE/tree/etc/init.d"/*; do
        [ -f "$f" ] || continue
        [ "$(tr -cd '\r' < "$f" | wc -c)" -gt 0 ] && crlf=$((crlf + 1))
    done
fi
[ "$crlf" -eq 0 ] && note "every init script is LF only" OK || note "$crlf init script(s) carry CR" FAIL

badmode=0
for f in "$STAGE/tree/etc/init.d"/S*; do
    [ -f "$f" ] || continue
    [ -x "$f" ] || badmode=$((badmode + 1))
done
[ "$badmode" -eq 0 ] && note "every S* script is executable" OK || note "$badmode S* script(s) are not executable" FAIL

badsyntax=0
for f in "$STAGE/tree/etc/init.d"/*; do
    [ -f "$f" ] || continue
    head -1 "$f" | grep -q '^#!' || continue
    sh -n "$f" 2>/dev/null || badsyntax=$((badsyntax + 1))
done
[ "$badsyntax" -eq 0 ] && note "sh -n accepts every init script" OK || note "$badsyntax script(s) fail sh -n" FAIL

[ -e "$STAGE/tree/etc/kvm.disk0" ] \
    && note "/etc/kvm.disk0 is present" OK \
    || note "/etc/kvm.disk0 is missing, the first boot would reformat /data" FAIL

[ -e "$STAGE/tree/etc/kvm/ssh_stop" ] \
    && note "/etc/kvm/ssh_stop is absent" FAIL \
    || note "/etc/kvm/ssh_stop is absent" OK

[ -e "$STAGE/tree/swapfile" ] \
    && note "no /swapfile" FAIL \
    || note "no /swapfile" OK

while read -r verb a b; do
    [ "$verb" = remove ] || continue
    [ -e "$STAGE/tree${a}" ] && note "removed: $a" FAIL || note "removed: $a" OK
done < "$STAGE/m"

if [ "$fail" -ne 0 ]; then
    echo
    echo "gates failed; no image written"
    rm -f "$OUT"
    exit 1
fi

echo
echo "############ 6. build the filesystem"
# The journal stays. This is a root filesystem on a board whose normal way of
# ending is a power cut, and an unclean unmount with no journal is a repair with
# no record of what was in flight.
#
# metadata_csum_seed is dropped so the image is not bound to the UUID it was
# built with, which lets one image be written to either slot.
rm -f "$OUT"
mke2fs -q -F -t ext4 -L "$(basename "$OUT" .img)" -d "$STAGE/tree" -O ^metadata_csum_seed \
       "$OUT" "${SIZE_MIB}m"
e2fsck -fp "$OUT" >/dev/null 2>&1 || true

echo
printf '  %-40s %s\n' "image" "$OUT"
printf '  %-40s %s bytes\n' "size" "$(wc -c < "$OUT")"
printf '  %-40s %s\n' "sha256" "$(sha256sum "$OUT" | cut -d' ' -f1)"
