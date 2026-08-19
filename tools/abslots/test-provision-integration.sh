#!/bin/sh
# Drive a real card image through the real parted and the real mkfs.exfat.
#
#   test-provision-integration.sh
#
# This is the test that would have caught the fault in 1.0.0. That image
# declared a data partition, nothing ever made a filesystem on it, and every
# other suite in this repository passed.
#
# Needs parted, exfatprogs, sfdisk, mtools and e2fsprogs. Run it in the release
# host image, which carries all of them.
#
# What this does NOT cover: whether the kernel exposes the new partition node
# while the root filesystem is mounted from the same disk. parted is driven
# against a FILE here, so there are no partition nodes at all. That step is
# measured separately on the board through a loop device, and it is closed for
# good only by a flashed card. See docs/specs/2026-08-18-small-card-image-design.md.
#
# The versions differ from the device's, and that is worth knowing rather than
# hiding: this image carries parted 3.5 and exfatprogs 1.2.0, the board carries
# 3.6 and 1.2.2. The table this produces is identical on both, which is a useful
# thing to have measured, but it is not evidence about the board.
set -u

HERE=$(cd "$(dirname "$0")" && pwd)
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

for t in parted mkfs.exfat blkid sfdisk mkfs.vfat mke2fs mcopy; do
    # Exit 2, not 0. A suite that ran no case has not passed, and a green
    # line for a suite that did nothing is worse than a red one: it is
    # counted as coverage that does not exist.
    command -v "$t" > /dev/null 2>&1 || { skip_missing "$t"; }
done

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

echo "provisioning, end to end"

# A scaled copy of the shipped table. The real one is 5 GiB, which is too slow
# to build in a test loop. p1 keeps its real size because FAT16 refuses to make
# a filesystem much smaller.
cat > "$WORK/table.sfdisk" <<'EOF'
label: dos
label-id: 0x70781617
unit: sectors

1 : start=1,      size=32768,  type=c, bootable
2 : start=40960,  size=16384,  type=83
3 : start=57344,  size=16384,  type=83
4 : start=73728,  size=18432,  type=5
5 : start=75776,  size=16384,  type=83
EOF

START=$(sh "$HERE/data-start.sh" "$WORK/table.sfdisk")
check "the derived start is the end of p5 plus 4 MiB" "$START" "100352"

mkdir -p "$WORK/boot"
printf 'fip\n' > "$WORK/boot/fip.bin"
for img in root recovery; do
    dd if=/dev/zero of="$WORK/$img.img" bs=512 count=16384 2>/dev/null
    mke2fs -q -t ext4 -F "$WORK/$img.img" > /dev/null 2>&1
done

CARD_TABLE="$WORK/table.sfdisk" sh "$HERE/build-card.sh" \
    "$WORK/boot" "$WORK/root.img" "$WORK/recovery.img" "$WORK/card.img" \
    > "$WORK/build.log" 2>&1
check "a card image is built" "$?" "0"
check "the image ends where the data partition will start" \
    "$(( $(wc -c < "$WORK/card.img") / 512 ))" "$START"

# Write it to a "card". A card is bigger than the image, which is the whole
# point of the change: 393216 sectors is 192 MiB.
CARD=393216
cp "$WORK/card.img" "$WORK/flashed.img"
truncate -s $((CARD * 512)) "$WORK/flashed.img"

# The two calls provision_data makes, in the order it makes them, against the
# real parted. Its exit status is ignored here for the same reason the device
# ignores it.
parted -s "$WORK/flashed.img" resizepart 4 100% > /dev/null 2>&1 || true
parted -s "$WORK/flashed.img" mkpart logical ntfs "${START}s" 100% > /dev/null 2>&1 || true

# sfdisk -d prints "device1 : start= 1, size= 32768, type=c, bootable", padded,
# so "start=" and its value can split into two tokens while "type=c" stays whole.
field() {
    sfdisk -d "$WORK/flashed.img" 2>/dev/null \
        | awk -v p="$1" -v k="$2" '
            $0 ~ (p " *:") {
                n = split($0, t, "[ ,]+")
                for (i = 1; i <= n; i++) if (t[i] == k "=") { print t[i+1]; exit }
                for (i = 1; i <= n; i++) if (t[i] ~ "^" k "=") { sub("^" k "=", "", t[i]); print t[i]; exit }
            }'
}

check "the data partition is made" "$(field "$WORK/flashed.img6" start)" "$START"
check "it is type 7, so the card reads on Windows and macOS" \
    "$(field "$WORK/flashed.img6" type)" "7"
check "it reaches the end of the card" \
    "$(( $(field "$WORK/flashed.img6" start) + $(field "$WORK/flashed.img6" size) ))" "$CARD"

# The entries for /boot, root A and root B must be untouched by the write to
# sector 0. This is the one that would be expensive to get wrong.
check "the boot partition still starts at sector 1" "$(field "$WORK/flashed.img1" start)" "1"
check "the boot partition keeps type c" "$(field "$WORK/flashed.img1" type)" "c"
check "root A is unmoved" "$(field "$WORK/flashed.img2" start)" "40960"
check "root B is unmoved" "$(field "$WORK/flashed.img3" start)" "57344"
check "the recovery slot is unmoved" "$(field "$WORK/flashed.img5" start)" "75776"

# The slots must still pass a check after sector 0 was rewritten.
dd if="$WORK/flashed.img" of="$WORK/roota.out" bs=512 skip=40960 count=16384 status=none
check "root A still passes e2fsck" \
    "$(e2fsck -fn "$WORK/roota.out" > /dev/null 2>&1 && echo clean || echo dirty)" "clean"

# And the filesystem. Carve the data partition out and format it exactly as the
# device does, then ask blkid whether it is there. Before this change nothing
# ever ran this step, on any card.
DSIZE=$(field "$WORK/flashed.img6" size)
dd if="$WORK/flashed.img" of="$WORK/data.out" bs=512 skip="$START" count="$DSIZE" status=none
mkfs.exfat -L data "$WORK/data.out" > /dev/null 2>&1
check "mkfs.exfat makes a filesystem on it" "$?" "0"
check "and blkid can name it" \
    "$([ -n "$(blkid "$WORK/data.out" 2>/dev/null)" ] && echo named || echo silent)" "named"

# The guard that decides whether to format has to say no to this one now.
S01="$HERE/../../kvmapp/system/init.d/S01fs"
sed -n '/^# --- data geometry ---/,/^# --- end data geometry ---/p' "$S01" > "$WORK/geo.sh"
got=$(sh -c ". $WORK/geo.sh; has_filesystem $WORK/data.out && echo yes || echo no")
check "has_filesystem agrees, so a re-flash would not destroy it" "$got" "yes"

# And it has to say yes to a partition that holds nothing, or a card whose
# filesystem did not survive would never be repaired.
dd if=/dev/zero of="$WORK/blank.out" bs=1M count=16 status=none
got=$(sh -c ". $WORK/geo.sh; has_filesystem $WORK/blank.out && echo yes || echo no")
check "and an empty partition still reads as empty" "$got" "no"

echo
echo "passed $pass, failed $fail"
[ "$fail" -eq 0 ]
