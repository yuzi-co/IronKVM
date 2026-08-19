#!/bin/sh
# Tests for build-card.sh.
#
# The real geometry is 5 GiB, which is too slow to build in a test loop, so the
# script reads its table from a file and the test hands it a small one. What is
# under test is the assembly and its gates, not the numbers. p1 keeps its real
# size, because FAT16 refuses to make a filesystem much smaller.
#
# Needs sfdisk, mkfs.vfat, mtools and e2fsprogs, so run it in a Linux container.

set -u

HERE=$(cd "$(dirname "$0")" && pwd)
SCRIPT="$HERE/build-card.sh"
pass=0
fail=0

ok()    { pass=$((pass + 1)); echo "  ok    $1"; }
bad()   { fail=$((fail + 1)); echo "  FAIL  $1"; }
check() { if [ "$2" = "$3" ]; then ok "$1"; else bad "$1 (got '$2', want '$3')"; fi; }

# Size of a file in sectors, or 0 when it is not there. Reading it inline killed
# the whole run the first time a build failed: wc printed nothing, the arithmetic
# had no operand, and the shell exited before any later case could report. A
# suite that stops at the first failure hides every fault behind the first one.
sectors() {
    [ -f "$1" ] || { echo 0; return; }
    echo $(( $(wc -c < "$1") / 512 ))
}

skip_missing() {
    echo "$(basename "$0"): needs $1, which is not on PATH." >&2
    echo "the release host image carries it:" >&2
    echo "  docker build -t ironkvm-release-host tools/release" >&2
    echo "  docker run --rm -v \"$PWD:/repo\" -w /repo ironkvm-release-host sh $0" >&2
    exit 2
}

for t in sfdisk mkfs.vfat e2fsck mke2fs mcopy mdir; do
    # Exit 2, not 0. A suite that ran no case has not passed, and a green
    # line for a suite that did nothing is worse than a red one: it is
    # counted as coverage that does not exist.
    command -v "$t" > /dev/null 2>&1 || { skip_missing "$t"; }
done

# The offsets the assertions use, kept beside the table they come from.
P1=1
P2=40960
P3=57344
P5=75776
SLOT=16384
# The end of the image, which is also where the data partition starts. Same rule
# as the real table: the end of the recovery slot plus the 4 MiB EBR gap.
END=100352

setup() {
    WORK=$(mktemp -d)
    # The same shape as the real table: one FAT that must not move, two ext4
    # slots, an extended container that stops with the recovery slot, and no
    # data partition at all. The device makes that one on the first boot.
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
    mkdir -p "$WORK/boot"
    printf 'fip\n'  > "$WORK/boot/fip.bin"
    printf 'boot\n' > "$WORK/boot/boot.sd"
    for img in root recovery; do
        dd if=/dev/zero of="$WORK/$img.img" bs=512 count="$SLOT" 2>/dev/null
        mke2fs -q -t ext4 -F "$WORK/$img.img" > /dev/null 2>&1
    done
}

teardown() { rm -rf "$WORK"; }

run() {
    CARD_TABLE="$WORK/table.sfdisk" \
        sh "$SCRIPT" "$WORK/boot" "$WORK/root.img" "$WORK/recovery.img" \
           "$WORK/card.img" > "$WORK/out" 2>&1
    echo $?
}

echo "build-card.sh"

setup
status=$(run)
check "a card is built" "$status" "0"

# The table has to survive into the image, or the board has no slots to boot.
# Five, not six. A sixth would mean the table still declares a data partition,
# which is what limited the image to a 32 GB card.
check "the image carries five partitions" \
    "$(sfdisk -l "$WORK/card.img" 2>/dev/null | grep -c '^/.*card\.img[0-9]')" "5"

check "no data partition is declared" \
    "$(sfdisk -l "$WORK/card.img" 2>/dev/null | grep -c 'card\.img6')" "0"

# The image ends where the data partition will start. The 4 MiB gap ahead of it
# is carried and zeroed, so a card that previously held a different layout has
# no stale EBR sector left anywhere the new chain could reach.
check "the image ends where the data partition will start" \
    "$(sectors "$WORK/card.img")" "$END"

# p1 is a FAT filesystem holding files, not a raw image. The boot ROM finds
# fip.bin by reading that filesystem, so a build that wrote an image over the
# partition would destroy the header the ROM has to read first.
check "the boot partition is a readable filesystem" \
    "$(MTOOLS_SKIP_CHECK=1 mdir -i "$WORK/card.img@@$((P1 * 512))" :: 2>/dev/null | grep -c 'fip')" "1"
MTOOLS_SKIP_CHECK=1 mcopy -i "$WORK/card.img@@$((P1 * 512))" ::fip.bin "$WORK/read-back" 2>/dev/null
check "the boot files are intact inside it" "$(cat "$WORK/read-back" 2>/dev/null)" "fip"
check "every boot file is copied in" \
    "$(MTOOLS_SKIP_CHECK=1 mdir -i "$WORK/card.img@@$((P1 * 512))" :: 2>/dev/null | grep -cE '^(boot|fip) ')" "2"

# A root written at the wrong offset still produces a plausible file. The board
# is where that would first be noticed, which is too late.
check "root A is at the p2 offset" \
    "$(dd if="$WORK/card.img" bs=512 skip="$P2" count="$SLOT" 2>/dev/null | \
       cmp -s - "$WORK/root.img" && echo placed || echo missing)" "placed"

check "recovery is at the p5 offset" \
    "$(dd if="$WORK/card.img" bs=512 skip="$P5" count="$SLOT" 2>/dev/null | \
       cmp -s - "$WORK/recovery.img" && echo placed || echo missing)" "placed"

# Slot B ships empty on purpose. Writing a second copy of root A would add 2 GiB
# to every download to carry the same bytes twice, and the first update fills it.
check "slot B is left empty" \
    "$(dd if="$WORK/card.img" bs=512 skip="$P3" count="$SLOT" 2>/dev/null | \
       tr -d '\0' | wc -c | tr -d ' ')" "0"

# The gate that matters most. A filesystem that did not survive the copy makes a
# card that boots into nothing, and only e2fsck can say so here.
dd if="$WORK/card.img" bs=512 skip="$P2" count="$SLOT" of="$WORK/check.img" 2>/dev/null
check "the copied root passes e2fsck" \
    "$(e2fsck -fn "$WORK/check.img" > /dev/null 2>&1 && echo clean || echo dirty)" "clean"
teardown

# A root larger than its partition would be written straight over the next one.
# Silent truncation here is a slot that fails e2fsck on the device.
setup
dd if=/dev/zero of="$WORK/root.img" bs=512 count=$((SLOT * 2)) 2>/dev/null
mke2fs -q -t ext4 -F "$WORK/root.img" > /dev/null 2>&1
status=$(run)
check "an oversized root is refused" "$status" "1"
check "no card is left behind when a root is oversized" \
    "$(ls "$WORK"/card.img* 2>/dev/null | wc -l | tr -d " ")" "0"
teardown

# A corrupt filesystem must not be shipped. This is the check that catches a
# copy that landed at the wrong offset.
setup
dd if=/dev/urandom of="$WORK/root.img" bs=512 count="$SLOT" conv=notrunc 2>/dev/null
status=$(run)
check "a corrupt root is refused" "$status" "1"
check "no card is left behind when a root is corrupt" \
    "$(ls "$WORK"/card.img* 2>/dev/null | wc -l | tr -d " ")" "0"
teardown

# A card with no fip.bin does not start, and it fails before anything that could
# report why. Refusing to build one is the only place that can say so.
setup
rm -f "$WORK/boot/fip.bin"
status=$(run)
check "a boot directory with no fip.bin is refused" "$status" "1"
# The readback check later in the build also fails without fip.bin, so the two
# guards hide each other and neither could be tested alone. What the early one
# adds is a message that names the problem before minutes of work are spent, so
# the message is what proves it ran.
check "it says which file is missing, before building anything" \
    "$(grep -c 'holds no fip.bin' "$WORK/out")" "1"
teardown

# A missing input must fail before anything is written, not halfway through.
setup
rm -f "$WORK/recovery.img"
status=$(run)
check "a missing input is refused" "$status" "1"
teardown

echo
echo "passed $pass, failed $fail"
[ "$fail" -eq 0 ]
