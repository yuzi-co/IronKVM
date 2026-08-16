#!/bin/sh
# Assemble a flashable card image from a boot directory and two root filesystems.
#
#   build-card.sh <boot-dir> <root.img> <recovery.img> <out.img>
#
# <boot-dir> holds what belongs in /boot: fip.bin, the repacked boot.sd, and any
# markers. It is a DIRECTORY and not an image, because p1 is a FAT filesystem
# that CONTAINS those files. The boot ROM finds fip.bin by reading that
# filesystem, so writing a boot image over the partition would destroy the FAT
# header the ROM has to read first.
#
# Needs sfdisk, mkfs.vfat, mcopy, e2fsck and truncate. Nothing here runs on the
# device.
#
# build-image.sh produces one ext4 root filesystem. This is what puts those
# filesystems where the boot ROM and the initramfs expect them. Before it
# existed a card had to be partitioned by hand, which is not an install
# instruction that can be published.
#
# Two decisions worth knowing before you read the code.
#
# The image is TRUNCATED at the start of the data partition. Its table still describes
# the data partition, and S01fs formats that device on first boot when its
# marker is present, so the bytes are not needed. Carrying them would make the
# image 28.85 GiB instead of 5.02 GiB, and compressing 23 GiB of zeros costs
# minutes on every release for nothing.
#
# Slot B ships EMPTY. It is what the first update writes. Populating it would
# add 2 GiB to every download to carry a second copy of slot A.
#
# The table is fixed, so the image needs a card of at least the sector count the
# table ends at. A larger card leaves the extra space unused until a release
# grows the data partition on first boot.
#
# Environment, for tests:
#   CARD_TABLE            the sfdisk layout        (default partition.sfdisk)
#   CARD_DROP_FROM        first partition to omit  (default 6)
set -eu

HERE=$(cd "$(dirname "$0")" && pwd)
BOOT=${1:?usage: build-card.sh <boot-dir> <root.img> <recovery.img> <out.img>}
ROOT=${2:?usage: build-card.sh <boot-dir> <root.img> <recovery.img> <out.img>}
RECOVERY=${3:?usage: build-card.sh <boot-dir> <root.img> <recovery.img> <out.img>}
OUT=${4:?usage: build-card.sh <boot-dir> <root.img> <recovery.img> <out.img>}

TABLE=${CARD_TABLE:-$HERE/partition.sfdisk}
DROP_FROM=${CARD_DROP_FROM:-6}

# Every input is checked before anything is written. A build that fails halfway
# leaves a file that looks like a card and is not one.
[ -d "$BOOT" ] || { echo "no such boot directory: $BOOT" >&2; exit 1; }
for f in "$ROOT" "$RECOVERY" "$TABLE"; do
    [ -f "$f" ] || { echo "no such file: $f" >&2; exit 1; }
done

# fip.bin is what the SoC ROM looks for. A card without it does not start, and
# it fails before anything that could report why, so refuse here instead.
[ -f "$BOOT/fip.bin" ] || { echo "$BOOT holds no fip.bin; the card would not start" >&2; exit 1; }

# field reads the layout rather than repeating it. Two copies of a geometry
# drift, and the copy that drifts is the one nobody runs.
field() {
    sed -n "s/^$1 *: *.*$2=\([0-9]*\).*/\1/p" "$TABLE" | head -1
}

P1_START=$(field 1 start); P1_SIZE=$(field 1 size)
P2_START=$(field 2 start); P2_SIZE=$(field 2 size)
P5_START=$(field 5 start); P5_SIZE=$(field 5 size)
P6_START=$(field 6 start); P6_SIZE=$(field 6 size)

for v in "$P1_START" "$P1_SIZE" "$P2_START" "$P2_SIZE" "$P5_START" "$P5_SIZE" \
         "$P6_START" "$P6_SIZE"; do
    [ -n "$v" ] || { echo "$TABLE does not describe the expected layout" >&2; exit 1; }
done

# The file is created at its FULL size first, so sfdisk never writes a table
# that describes sectors past the end of the file it is writing to. It is
# sparse, so those sectors cost nothing until something writes them, and
# nothing does.
FULL=$((P6_START + P6_SIZE))

# Keep everything up to the START of the dropped partition, not to the end of
# the one before it. A logical partition is described by an EBR sector that sits
# ahead of it, in the gap the table leaves for exactly that, so cutting at the
# end of the recovery slot throws away the record of the data partition and
# S01fs then has no device to format. The gap is 4 MiB.
KEEP=$(field "$DROP_FROM" start)
[ -n "$KEEP" ] || { echo "$TABLE has no partition $DROP_FROM to drop" >&2; exit 1; }

# Refuse an oversized filesystem before writing it. dd would otherwise write it
# straight over the next partition, and the first thing to notice would be a
# slot that fails e2fsck on the device.
fits() {
    have=$(( $(wc -c < "$1") / 512 ))
    [ "$have" -le "$2" ] && return 0
    echo "$1 is $have sectors, its partition holds $2" >&2
    return 1
}
fits "$ROOT" "$P2_SIZE"
fits "$RECOVERY" "$P5_SIZE"

WORK="$OUT.partial"
rm -f "$WORK"
# Anything that leaves early from here on must not leave a half-built card
# behind that somebody could flash.
trap 'rm -f "$WORK"' EXIT

truncate -s $((FULL * 512)) "$WORK"
sfdisk --no-reread --no-tell-kernel "$WORK" < "$TABLE" > /dev/null

# p1 is a FAT filesystem holding files, so it is formatted and then filled with
# mcopy. mtools addresses an offset inside an image with image@@offset, which is
# what keeps this working without a loop device and without root.
mkfs.vfat -F 16 -n BOOT --offset "$P1_START" "$WORK" "$((P1_SIZE / 2))" > /dev/null
for f in "$BOOT"/*; do
    [ -e "$f" ] || continue
    MTOOLS_SKIP_CHECK=1 mcopy -s -i "$WORK@@$((P1_START * 512))" "$f" ::
done

# Read one file back out. mcopy reports success for a copy into a filesystem it
# addressed at the wrong offset, and the next reader would be the boot ROM.
MTOOLS_SKIP_CHECK=1 mdir -i "$WORK@@$((P1_START * 512))" ::fip.bin > /dev/null 2>&1 || {
    echo "fip.bin is not readable back from the boot partition" >&2
    exit 1
}

dd if="$ROOT"     of="$WORK" bs=512 seek="$P2_START" conv=notrunc status=none
dd if="$RECOVERY" of="$WORK" bs=512 seek="$P5_START" conv=notrunc status=none

# Read the filesystems back out of the image and check them there. Checking the
# inputs would prove nothing about the copy, and a copy that lands at the wrong
# offset still produces a plausible file.
check_fs() {
    tmp="$WORK.check"
    dd if="$WORK" of="$tmp" bs=512 skip="$1" count="$2" status=none
    if ! e2fsck -fn "$tmp" > /dev/null 2>&1; then
        rm -f "$tmp"
        echo "the filesystem at sector $1 does not pass e2fsck" >&2
        return 1
    fi
    rm -f "$tmp"
}
check_fs "$P2_START" "$P2_SIZE"
check_fs "$P5_START" "$P5_SIZE"

# Drop the tail. The table keeps describing the data partition, and S01fs makes
# a filesystem there on first boot.
truncate -s $((KEEP * 512)) "$WORK"

mv "$WORK" "$OUT"
trap - EXIT
echo "built $OUT: $((KEEP / 2048)) MiB, slot B empty, data partition formatted on first boot"
