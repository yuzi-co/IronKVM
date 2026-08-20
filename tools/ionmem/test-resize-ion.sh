#!/bin/sh
# Check that resize-ion.sh changes one number in the device tree and nothing else.
#
#   test-resize-ion.sh <boot.sd> [workdir]
#
# Needs u-boot-tools and device-tree-compiler, so run it in the same throwaway
# container that builds the image.
#
# The tool rewrites the length of one reserved-memory region. The board boots
# the kernel and the ramdisk that are already in the image, so the only thing
# that may differ between input and output is the flattened device tree, and
# within that, one property. Every case below holds that line.
#
# The refusals matter as much as the edit. An undersized carveout does not
# degrade: libkvm never checks an allocation result, so the server takes a
# SIGSEGV on a NULL and dies before it binds its port, and the board answers
# nothing. A tool that accepts a number below the measured floor would turn a
# memory saving into an outage that needs hands on the device.
ORIG=$1
# Exit 2, not 1: a boot.sd is an artefact no checkout carries, and "cannot run
# here" is a different answer from "the tool is broken".
[ -n "$ORIG" ] || { echo "usage: test-resize-ion.sh <boot.sd> [workdir]" >&2; exit 2; }
WORK=${2:-$(mktemp -d)}
HERE=$(cd "$(dirname "$0")" && pwd)
TOOL="$HERE/resize-ion.sh"

[ -f "$ORIG" ] || { echo "no such file: $ORIG" >&2; exit 2; }

for _t in dumpimage mkimage dtc fdtget fdtdump
do
    command -v "$_t" >/dev/null 2>&1 && continue
    echo "test-resize-ion.sh: needs $_t, which is not on PATH." >&2
    echo "the release host image carries it:" >&2
    echo "  docker run --rm -v \"$PWD:/repo\" -w /repo ironkvm-release-host sh $0 $ORIG" >&2
    exit 2
done

fails=0
note() { printf '  %-58s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

mkdir -p "$WORK"

# 56MB. The size this fork installs, and the one the cases below are written
# against.
WANT=58720256

echo "===== resize-ion.sh shrinks the reservation ====="

if [ ! -x "$TOOL" ] && [ ! -f "$TOOL" ]; then
    note "resize-ion.sh exists" FAIL
    echo
    echo "$fails case(s) FAILED"
    exit 1
fi
note "resize-ion.sh exists" OK

if sh "$TOOL" "$ORIG" "$WORK/build" "$WANT" > "$WORK/run.log" 2>&1; then
    note "the tool succeeds on a stock image" OK
else
    note "the tool succeeds on a stock image" FAIL
    sed 's/^/    /' "$WORK/run.log"
    echo
    echo "$fails case(s) FAILED"
    exit 1
fi

NEW="$WORK/build/boot.sd.new"
[ -f "$NEW" ] && note "it writes boot.sd.new" OK || {
    note "it writes boot.sd.new" FAIL
    echo
    echo "$fails case(s) FAILED"
    exit 1
}

# Pull all three images back out of both files and compare them directly,
# rather than trusting what the tool reported about itself.
for part in 0 1 2; do
    dumpimage -T flat_dt -p "$part" -o "$WORK/orig.$part" "$ORIG" >/dev/null 2>&1
    dumpimage -T flat_dt -p "$part" -o "$WORK/new.$part"  "$NEW"  >/dev/null 2>&1
done

cmp -s "$WORK/orig.0" "$WORK/new.0" \
    && note "the kernel is byte-identical" OK \
    || note "the kernel is byte-identical" FAIL

cmp -s "$WORK/orig.1" "$WORK/new.1" \
    && note "the ramdisk is byte-identical" OK \
    || note "the ramdisk is byte-identical" FAIL

cmp -s "$WORK/orig.2" "$WORK/new.2" \
    && note "the device tree differs from the original" FAIL \
    || note "the device tree differs from the original" OK

dtc -I dtb -O dts -o "$WORK/orig.dts" "$WORK/orig.2" 2>/dev/null
dtc -I dtb -O dts -o "$WORK/new.dts"  "$WORK/new.2"  2>/dev/null

removed=$(diff "$WORK/orig.dts" "$WORK/new.dts" | grep -c '^<')
added=$(diff "$WORK/orig.dts" "$WORK/new.dts" | grep -c '^>')
[ "$removed" = 1 ] && [ "$added" = 1 ] \
    && note "exactly one line changes in the source" OK \
    || note "exactly one line changes in the source ($removed out, $added in)" FAIL

# The changed line has to be a size, not something that merely sits near one.
changed=$(diff "$WORK/orig.dts" "$WORK/new.dts" | grep -E '^[<>]' | sed 's/^[<>][[:space:]]*//' | grep -c '^size = <')
[ "$changed" = 2 ] \
    && note "both sides of the change are a size property" OK \
    || note "both sides of the change are a size property ($changed of 2)" FAIL

now=$(fdtget "$WORK/new.2" /reserved-memory/ion size 2>/dev/null)
[ "$now" = "0 $WANT" ] \
    && note "the ion reservation is $WANT" OK \
    || note "the ion reservation is $WANT (got $now)" FAIL

# The framebuffer change this fork already ships lives in the same node list.
# A tool that rewrote the wrong reservation would still produce a bootable
# image, and the board would lose 8MB again without saying so.
for probe in "/reserved-memory/cvifb size" "/reserved-memory/cvifb status" \
             "/reserved-memory/ion compatible" "/memory@80000000 reg"; do
    set -- $probe
    was=$(fdtget "$WORK/orig.2" "$1" "$2" 2>/dev/null || echo missing)
    now=$(fdtget "$WORK/new.2"  "$1" "$2" 2>/dev/null || echo missing)
    [ "$was" = "$now" ] \
        && note "$1 $2 is untouched" OK \
        || note "$1 $2 is untouched (was $was, now $now)" FAIL
done

echo
echo "===== it refuses what the board cannot boot ====="

# Each of these must fail, and must leave no image behind. A tool that exits
# non-zero but still writes boot.sd.new invites an operator to install it.
refuses() {
    _what=$1
    _size=$2
    _dir="$WORK/refuse.$3"
    if sh "$TOOL" "$ORIG" "$_dir" "$_size" > "$WORK/refuse.$3.log" 2>&1; then
        note "$_what" FAIL
        return 0
    fi
    if [ -f "$_dir/boot.sd.new" ]; then
        note "$_what (refused, but left an image behind)" FAIL
        return 0
    fi
    note "$_what" OK
}

# One byte under the floor, so the case tests the boundary rather than an
# obviously silly number.
refuses "a size under the measured floor is refused"     49459199 floor
refuses "a size that is not a whole MB is refused"       58720257 align
refuses "growing the reservation is refused"             83886080 grow
refuses "a size equal to the current one is refused"     78643200 same
refuses "a size that is not a number is refused"         fiftysix  nan

# The floor itself has to be accepted, or the constant in the tool and the
# constant in this file have drifted apart and one of them is wrong.
if sh "$TOOL" "$ORIG" "$WORK/atfloor" 49459200 > "$WORK/atfloor.log" 2>&1; then
    note "a size at the floor is refused for alignment, not for the floor" FAIL
else
    grep -q "whole number of MB" "$WORK/atfloor.log" \
        && note "a size at the floor is refused for alignment, not for the floor" OK \
        || note "a size at the floor is refused for alignment, not for the floor" FAIL
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "===== resize-ion.sh edits one number and refuses the rest ====="
    exit 0
fi

echo "===== $fails case(s) FAILED ====="
exit 1
