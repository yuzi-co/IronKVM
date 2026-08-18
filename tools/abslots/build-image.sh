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
#   remove <path in the image>                           runs first
#   add    <path in the payload> <path in image> [mode]  runs second
#   touch  <path in the image>                           runs last
#
# remove before add, so a manifest can replace a directory by removing it and
# adding a new one. touch last, because a marker may live inside a directory
# that add created.
#
# The optional mode on add is not decoration. The payload is a copy of a
# Windows checkout on a Linux build host, so its modes and its ownership
# describe the copy and not the image. On 2026-08-16 that shipped a root A with
# /kvmapp/server/NanoKVM-Server at 0644: a slot that boots and then cannot
# start its own server. Declare the mode for anything that must be executable,
# and the builder writes root ownership on every added path itself.
#
# The remove list is the half that matters. cp adds and overwrites and never
# deletes, so a file the base has and the fork does not will survive into every
# image built from it. That is exactly how /etc/kvm/ssh_stop reaches a slot and
# stops sshd while reporting success, and that file is present in the official
# v1.4.3 rootfs.
set -e

HERE=$(cd "$(dirname "$0")" && pwd)

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

while read -r verb a b c; do
    [ "$verb" = remove ] || continue
    rm -rf "$STAGE/tree${a}"
    echo "  remove  $a"
done < "$STAGE/m"

while read -r verb a b c; do
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
    # The image is a root filesystem. Nothing added to it belongs to the person
    # who ran the build, and a uid that only exists on the build host is one
    # more thing that reads as correct until sshd applies StrictModes to it.
    chown -R 0:0 "$STAGE/tree${b}" 2>/dev/null || true
    if [ -n "$c" ]; then
        if [ -d "$STAGE/tree${b}" ]; then
            chmod -R "$c" "$STAGE/tree${b}"
        else
            chmod "$c" "$STAGE/tree${b}"
        fi
        echo "  add     $a -> $b (mode $c)"
    else
        echo "  add     $a -> $b"
    fi
done < "$STAGE/m"

while read -r verb a b c; do
    [ "$verb" = touch ] || continue
    mkdir -p "$(dirname "$STAGE/tree${a}")"
    : > "$STAGE/tree${a}"
    echo "  touch   $a"
done < "$STAGE/m"

echo
echo "############ 2b. correct the base's boot path ownership"
# Sipeed's rootfs ships /etc/init.d, its scripts and /usr/sbin/tailscaled owned
# by uid 1000, which is their build host's user and exists in no passwd file on
# the board. Root ignores the mode when it executes them, so the board boots and
# nothing looks wrong. It is still a boot path owned by a stranger.
#
# Only these two paths are corrected. The rest of the base keeps what Sipeed
# shipped: /etc/bind belongs to named and /var/www to www-data, and a blanket
# chown would break the cases that are already right.
for p in /etc/init.d /usr/sbin/tailscaled; do
    [ -e "$STAGE/tree$p" ] || continue
    n=$(find "$STAGE/tree$p" \( ! -user 0 -o ! -group 0 \) -print 2>/dev/null | wc -l)
    [ "$n" -eq 0 ] && continue
    chown -R 0:0 "$STAGE/tree$p" 2>/dev/null || true
    printf '  %-40s %s path(s) to root\n' "$p" "$n"
done

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
#
# S01fs makes the data partition on the first boot and reads its start sector
# from here. The number is derived from partition.sfdisk rather than written
# down, because the table and this file must never disagree about where the
# partition goes.
DATA_START=$("$HERE/data-start.sh")
cat > "$STAGE/tree/etc/nanokvm-slots.conf" <<CONF
SLOT_A=/dev/mmcblk0p2
SLOT_B=/dev/mmcblk0p3
RECOVERY=/dev/mmcblk0p5
DATA_DEV=/dev/mmcblk0p6
DATA_START=$DATA_START
CONF

echo
echo "############ 5. gates"
fail=0
note() { printf '  %-58s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fail=1; return 0; }
ELFMAGIC=7f454c46

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

# rcS runs every file in /etc/init.d whose name starts with S, so a backup left
# beside a script is not a backup, it is a second script that runs at every
# boot. On 2026-08-16 a board carried /etc/init.d/S02identity.rollback, an older
# identity script that ran after the current one and bound /etc/kvm a second
# time. It was found only because the mount count looked wrong.
stray=0
for f in "$STAGE/tree/etc/init.d"/*; do
    [ -e "$f" ] || continue
    case "${f##*/}" in
        *.rollback|*.bak|*.orig|*.old|*.save|*~|*.dpkg-*|*.rpm*)
            stray=$((stray + 1))
            echo "      stray file in init.d: ${f##*/}" ;;
    esac
done
[ "$stray" -eq 0 ] \
    && note "no backup files in /etc/init.d" OK \
    || note "$stray backup file(s) in /etc/init.d would run at boot" FAIL

badsyntax=0
for f in "$STAGE/tree/etc/init.d"/*; do
    [ -f "$f" ] || continue
    head -1 "$f" | grep -q '^#!' || continue
    sh -n "$f" 2>/dev/null || badsyntax=$((badsyntax + 1))
done
[ "$badsyntax" -eq 0 ] && note "sh -n accepts every init script" OK || note "$badsyntax script(s) fail sh -n" FAIL

# An executable that is not executable. The init.d gate above cannot see this
# one, because the server binary is not an init script, and that is exactly how
# a root A shipped on 2026-08-16 with /kvmapp/server/NanoKVM-Server at 0644.
# Reading four bytes needs no `file` and no assumption about where binaries
# live: anything with the ELF magic is meant to be run.
badelf=0
find "$STAGE/tree" -type f ! -perm -u+x -print > "$STAGE/nonexec" 2>/dev/null || true
while IFS= read -r f; do
    # Three kinds of ELF are read and not run, and all three are correct at
    # 0644: a shared library that ld.so maps, a kernel module that insmod
    # reads, and an object file. Only a program is wrong at 0644.
    case "$f" in *.so|*.so.*|*.ko|*.o|*.a) continue ;; esac
    [ "$(dd if="$f" bs=4 count=1 2>/dev/null | od -An -tx1 | tr -d ' \n')" = "$ELFMAGIC" ] || continue
    badelf=$((badelf + 1))
    echo "      not executable: ${f#$STAGE/tree}"
done < "$STAGE/nonexec"
[ "$badelf" -eq 0 ] \
    && note "every ELF file is executable" OK \
    || note "$badelf ELF file(s) are not executable; declare a mode" FAIL

# Ownership on the paths the manifest added. The base keeps whatever Sipeed
# shipped; everything this repository puts in belongs to root.
badown=0
while read -r verb a b c; do
    [ "$verb" = add ] || continue
    [ -e "$STAGE/tree${b}" ] || continue
    n=$(find "$STAGE/tree${b}" \( ! -user 0 -o ! -group 0 \) -print 2>/dev/null | wc -l)
    [ "$n" -gt 0 ] && { badown=$((badown + n)); echo "      not root-owned: $b ($n)"; }
done < "$STAGE/m"
[ "$badown" -eq 0 ] \
    && note "every added path is owned by root" OK \
    || note "$badown added path(s) are not owned by root" FAIL

# The web UI has to be the fork's, because the fork's server answers a
# different shape. /api/vm/device/virtual returns {enabled, active, cost} per
# device here and a plain boolean upstream, and an object is truthy in
# JavaScript, so the official UI drew the virtual disk and the virtual network
# as permanently ON and read every click to turn them off as a request to turn
# them on. The switches could not be moved.
#
# The first root A shipped exactly that pairing, because root.manifest added no
# web at all and the official one arrived with the application tarball. Neither
# half looks wrong on its own, which is why this is a gate and not a review
# note. The string is an i18n key only the fork's UI has, and it survives
# minification.
if [ -d "$STAGE/tree/kvmapp" ]; then
    if grep -rq 'settings\.device\.endpoints' "$STAGE/tree/kvmapp/server/web" 2>/dev/null; then
        note "the web UI is the fork's build" OK
    else
        note "the web UI is not the fork's; its API shape differs from the server" FAIL
    fi
fi

[ -e "$STAGE/tree/etc/kvm.disk0" ] \
    && note "/etc/kvm.disk0 is present" OK \
    || note "/etc/kvm.disk0 is missing, the first boot would reformat /data" FAIL

[ -e "$STAGE/tree/etc/kvm/ssh_stop" ] \
    && note "/etc/kvm/ssh_stop is absent" FAIL \
    || note "/etc/kvm/ssh_stop is absent" OK

[ -e "$STAGE/tree/swapfile" ] \
    && note "no /swapfile" FAIL \
    || note "no /swapfile" OK

while read -r verb a b c; do
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
