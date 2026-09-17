#!/bin/sh
# Check that S01fs grows a freshly installed slot to fill its partition.
#
#   test-s01fs-grow.sh [path-to-S01fs]
#
# The slot images are built at a few hundred megabytes and written to 2 GiB
# slots, because the difference is what `slot install` has to push to the card.
# That trade is only sound if the filesystem is grown on the first boot. If the
# grow silently does not happen, the slot works and nobody notices: it simply
# has a tenth of the space it should, until the day a package install fails with
# no room on a card that is nine tenths empty.
#
# So every one of these cases is about the grow either happening or being
# reported. The four at the end are about it never being the reason a boot
# stops, because this runs in rcS on a board that cannot be power cycled.
S01=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S01fs}
[ -f "$S01" ] || { echo "usage: test-s01fs-grow.sh <S01fs>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

echo "===== the root filesystem grow ====="

sed -n '/^# --- root filesystem grow ---/,/^# --- end root filesystem grow ---/p' \
    "$S01" > "$WORK/grow.sh"
if [ ! -s "$WORK/grow.sh" ]; then
    note "the grow block can be extracted" FAIL
    echo; echo "$fails case(s) FAILED"; exit 1
fi
note "the grow block can be extracted" OK

# /proc/partitions counts 1024 byte blocks. p2 is a 2 GiB slot, p3 is another,
# and p6 is the data partition.
cat > "$WORK/partitions" <<'PARTS'
major minor  #blocks  name

 179        0   30535680 mmcblk0
 179        1      16384 mmcblk0p1
 179        2    2097152 mmcblk0p2
 179        3    2097152 mmcblk0p3
 179        6   25264128 mmcblk0p6
PARTS

mounts() { printf '%s / ext4 rw,relatime 0 0\n' "$1" > "$WORK/mounts"; }

# A dumpe2fs that reports whatever size the case asks for, in 4096 byte blocks.
cat > "$WORK/dumpe2fs" <<'STUB'
#!/bin/sh
[ -n "$FS_BLOCKS" ] || exit 1
cat <<OUT
Filesystem volume name:   slot-b
Block count:              $FS_BLOCKS
Reserved block count:     0
Block size:               4096
OUT
STUB
chmod +x "$WORK/dumpe2fs"

# A resize2fs that records the device it was asked to grow, and fails when the
# case wants a failure.
cat > "$WORK/resize2fs" <<'STUB'
#!/bin/sh
echo "$1" >> "$RESIZE_LOG"
[ -n "$RESIZE_FAILS" ] && exit 1
exit 0
STUB
chmod +x "$WORK/resize2fs"

# Run grow_root with the stubs in place, and print what resize2fs was called
# with. Prints nothing when it was never called.
run_grow() {
    : > "$WORK/resized"
    RESIZE_LOG="$WORK/resized" \
    MOUNTS="$WORK/mounts" PARTITIONS="$WORK/partitions" \
    DUMPE2FS="$WORK/dumpe2fs" RESIZE2FS="$WORK/resize2fs" \
    FS_BLOCKS="$1" RESIZE_FAILS="${2:-}" \
        sh -c ". $WORK/grow.sh; grow_root" > "$WORK/out" 2>&1
    echo $? > "$WORK/rc"
    cat "$WORK/resized"
}

echo
echo "--- it grows when the filesystem is smaller than its partition ---"

# 256 MiB of filesystem in a 2 GiB partition: the state every slot is in the
# first time it boots after an install.
mounts /dev/mmcblk0p2
got=$(run_grow 65536)
[ "$got" = /dev/mmcblk0p2 ] \
    && note "a 256 MiB filesystem in a 2 GiB slot is grown" OK \
    || note "a 256 MiB filesystem in a 2 GiB slot gave '$got'" FAIL

grep -q '256 MiB to 2048 MiB' "$WORK/out" \
    && note "it says on the console what it grew and by how much" OK \
    || note "it did not report the sizes: $(cat "$WORK/out")" FAIL

echo
echo "--- it takes the device from /proc/mounts, not from a partition number ---"

# The same image is written to either slot, and recovery is a third partition.
# A grow that named p2 would do nothing in slot B and would be wrong in
# recovery.
mounts /dev/mmcblk0p3
got=$(run_grow 65536)
[ "$got" = /dev/mmcblk0p3 ] \
    && note "a root on p3 grows p3" OK \
    || note "a root on p3 gave '$got'" FAIL

if grep -n 'mmcblk0p[0-9]' "$WORK/grow.sh" | grep -qv '^[0-9]*:#'; then
    note "the grow block names no partition of its own" FAIL
else
    note "the grow block names no partition of its own" OK
fi

echo
echo "--- it writes nothing when there is nothing to do ---"

# The second boot, and every boot after it. This is the case that decides
# whether this costs SD wear for ever.
mounts /dev/mmcblk0p2
got=$(run_grow 524288)
[ -z "$got" ] \
    && note "a filesystem that already fills its slot is left alone" OK \
    || note "a full-size filesystem was resized anyway: '$got'" FAIL

# mke2fs rounds down to whole block groups, so a filesystem that fills its
# partition is still a little short of it. Without slack that difference would
# be a resize2fs on every boot for ever.
got=$(run_grow 524160)
[ -z "$got" ] \
    && note "a filesystem 512 KiB short of its slot is left alone" OK \
    || note "a filesystem 512 KiB short was resized: '$got'" FAIL

# Two megabytes short is a real gap, not block group rounding.
got=$(run_grow 523776)
[ "$got" = /dev/mmcblk0p2 ] \
    && note "a filesystem 2 MiB short of its slot is grown" OK \
    || note "a filesystem 2 MiB short was not grown" FAIL

echo
echo "--- it never stops a boot ---"

# rcS runs this. Every path below has to end in a return, because the board it
# runs on cannot be power cycled by anyone who is not standing next to it.

mounts /dev/mmcblk0p2
got=$(run_grow 65536 yes)
[ "$got" = /dev/mmcblk0p2 ] \
    && note "a failing resize2fs is still attempted" OK \
    || note "a failing resize2fs was not attempted" FAIL
grep -q 'the next boot tries again' "$WORK/out" \
    && note "a failing resize2fs says the next boot tries again" OK \
    || note "a failing resize2fs printed: $(cat "$WORK/out")" FAIL

# A root the kernel does not list as a partition. An initramfs root, a root over
# NFS, or the tests' own plain files.
mounts /dev/loop0
got=$(run_grow 65536)
[ -z "$got" ] && [ "$(cat "$WORK/rc")" = 0 ] \
    && note "a root that is not a listed partition is left alone, rc 0" OK \
    || note "a root on /dev/loop0 gave '$got' rc $(cat "$WORK/rc")" FAIL

# Not a device at all. tmpfs, overlay, rootfs: all things that appear on / and
# none of them resizable.
printf 'overlay / overlay rw 0 0\n' > "$WORK/mounts"
got=$(run_grow 65536)
[ -z "$got" ] && [ "$(cat "$WORK/rc")" = 0 ] \
    && note "a root that is not a device node is left alone, rc 0" OK \
    || note "a root on overlay gave '$got' rc $(cat "$WORK/rc")" FAIL

# The base this ships on may not carry resize2fs. Alpine keeps it in
# e2fsprogs-extra and not in e2fsprogs, which is exactly how an image would come
# to be built without it.
mounts /dev/mmcblk0p2
: > "$WORK/resized"
out=$( RESIZE_LOG="$WORK/resized" MOUNTS="$WORK/mounts" PARTITIONS="$WORK/partitions" \
       DUMPE2FS="$WORK/dumpe2fs" RESIZE2FS="$WORK/absent-resize2fs" FS_BLOCKS=65536 \
       sh -c ". $WORK/grow.sh; grow_root" 2>&1 )
rc=$?
[ "$rc" = 0 ] && echo "$out" | grep -q 'no resize2fs' \
    && note "a base with no resize2fs says so and carries on" OK \
    || note "a base with no resize2fs gave rc $rc: $out" FAIL

# The board running the Sipeed root filesystem has resize2fs and no dumpe2fs.
# That base needs no growing, because its filesystem was made at the size of the
# partition, but silence would be the wrong answer for an image that did need it
# and was built without the package.
out=$( RESIZE_LOG="$WORK/resized" MOUNTS="$WORK/mounts" PARTITIONS="$WORK/partitions" \
       DUMPE2FS="$WORK/absent-dumpe2fs" RESIZE2FS="$WORK/resize2fs" FS_BLOCKS=65536 \
       sh -c ". $WORK/grow.sh; grow_root" 2>&1 )
rc=$?
: > "$WORK/resized"
[ "$rc" = 0 ] && echo "$out" | grep -q 'no dumpe2fs' \
    && note "a base with no dumpe2fs says so and carries on" OK \
    || note "a base with no dumpe2fs gave rc $rc: $out" FAIL

# A superblock that cannot be read. A filesystem type this does not understand,
# or a device that is busy.
out=$( RESIZE_LOG="$WORK/resized" MOUNTS="$WORK/mounts" PARTITIONS="$WORK/partitions" \
       DUMPE2FS="$WORK/dumpe2fs" RESIZE2FS="$WORK/resize2fs" \
       sh -c ". $WORK/grow.sh; grow_root" 2>&1 )
rc=$?
[ "$rc" = 0 ] && [ ! -s "$WORK/resized" ] \
    && note "an unreadable superblock is left alone, rc 0" OK \
    || note "an unreadable superblock gave rc $rc" FAIL

echo
echo "--- the caller cannot be stopped by it either ---"

# grow_root returns 1 when resize2fs fails, so the call site has to discard it.
# Without that, `set -e` or a shell that propagates the status ends rcS here and
# /data is never mounted.
if grep -qE '^[[:space:]]*grow_root \|\| true' "$S01"; then
    note "S01fs calls grow_root with its status discarded" OK
else
    note "S01fs does not discard grow_root's status" FAIL
fi

# The stock path backgrounds a resize2fs of its own. Two of them on one device
# at once is not something to discover on a board nobody can reach.
if sed -n '/^if \[ "\$1" = "start" \]/,$p' "$S01" | grep -qE '^[[:space:]]*else[[:space:]]*$'; then
    note "grow_root is in the branch the stock resize does not take" OK
else
    note "grow_root may run beside the stock resize" FAIL
fi

echo
echo "--- and it runs before the watchdog is armed ---"

# The resize is a metadata write during boot, and it is the only new one this
# change introduces. rcS runs /etc/init.d in sorted order, and S01fs sorts before
# S01hwdt, which is what arms the SoC watchdog, so a slow resize cannot be the
# reason a boot misses its first pet.
#
# The two scripts live in different source directories, so the order is a
# property of the names the manifest gives them in the image and not of where
# they are kept. That is what this reads.
#
# The manifest is in the ironkvm-dist repository, which is a sibling checkout
# and not always present. Its absence makes this one case unrunnable, so the
# suite reports it as a skip instead of a failure.
IRONKVM_DIST=${IRONKVM_DIST:-$(dirname "$0")/../../../ironkvm-dist}
MANIFEST=${MANIFEST:-$IRONKVM_DIST/abslots/manifest/root.manifest}
no_manifest=no
if [ -f "$MANIFEST" ]; then
    fs=$(awk '$1 == "add" && $3 ~ /\/etc\/init\.d\/.*fs$/ { n = $3; sub(/.*\//, "", n); print n }' "$MANIFEST" | head -1)
    hwdt=$(awk '$1 == "add" && $3 ~ /\/etc\/init\.d\/.*hwdt$/ { n = $3; sub(/.*\//, "", n); print n }' "$MANIFEST" | head -1)
    if [ -z "$fs" ] || [ -z "$hwdt" ]; then
        note "the manifest installs both an fs and a hwdt script" FAIL
    elif [ "$(printf '%s\n%s\n' "$fs" "$hwdt" | sort | head -1)" = "$fs" ]; then
        note "$fs sorts before $hwdt, so the resize precedes the watchdog" OK
    else
        note "$hwdt sorts before $fs; the resize would run under the watchdog" FAIL
    fi
else
    no_manifest=yes
fi

echo
if [ "$fails" -eq 0 ] && [ "$no_manifest" = yes ]; then
    echo "every other case passed"
    echo "needs the ironkvm-dist checkout for $MANIFEST; set IRONKVM_DIST or MANIFEST"
    exit 2
fi
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
    exit 0
fi
echo "$fails case(s) FAILED"
exit 1
