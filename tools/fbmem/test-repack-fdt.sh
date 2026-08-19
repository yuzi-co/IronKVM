#!/bin/sh
# Check that repack-fdt.sh changes the device tree and nothing else.
#
#   test-repack-fdt.sh <stock-boot.sd> [workdir]
#
# Needs u-boot-tools and device-tree-compiler, so run it in the same throwaway
# container that builds the image.
#
# The tool turns off one reserved-memory region. The board boots the kernel and
# the ramdisk that are already in the image, so the only thing that may differ
# between the input and the output is the flattened device tree. Every case
# below exists to hold that line: a tool that quietly re-compresses a kernel or
# rebuilds a ramdisk would still produce a bootable image most of the time, and
# the one time it did not there would be nothing to compare against.
#
# The edit itself is two property values. That is small enough to state exactly,
# so the test states it exactly: two lines change in the decompiled source, both
# from "okay" to "disabled", at /reserved-memory/cvifb and at /cvifb. Anything
# else is a failure, including a change that looks harmless.
ORIG=$1
# Exit 2, not 1: a stock boot.sd is an artefact no checkout carries, and
# "cannot run here" is a different answer from "the tool is broken".
[ -n "$ORIG" ] || { echo "usage: test-repack-fdt.sh <stock-boot.sd> [workdir]" >&2; exit 2; }
WORK=${2:-$(mktemp -d)}
HERE=$(cd "$(dirname "$0")" && pwd)
TOOL="$HERE/repack-fdt.sh"

[ -f "$ORIG" ] || { echo "no such file: $ORIG" >&2; exit 2; }

# The tools the header names. One of them being absent is not a defect in
# repack-fdt.sh, so say which one and exit 2 rather than fail a case.
for _t in dumpimage mkimage dtc fdtdump
do
    command -v "$_t" >/dev/null 2>&1 && continue
    echo "test-repack-fdt.sh: needs $_t, which is not on PATH." >&2
    echo "the release host image carries it:" >&2
    echo "  docker run --rm -v \"$PWD:/repo\" -w /repo ironkvm-release-host sh $0 $ORIG" >&2
    exit 2
done

fails=0
note() { printf '  %-58s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

mkdir -p "$WORK"

echo "===== repack-fdt.sh disables the framebuffer reservation ====="

if [ ! -x "$TOOL" ]; then
    note "repack-fdt.sh exists and is executable" FAIL
    echo
    echo "$fails case(s) FAILED"
    exit "$fails"
fi
note "repack-fdt.sh exists and is executable" OK

if "$TOOL" "$ORIG" "$WORK/build" > "$WORK/run.log" 2>&1; then
    note "the tool succeeds on a stock image" OK
else
    note "the tool succeeds on a stock image" FAIL
    sed 's/^/    /' "$WORK/run.log"
    echo
    echo "$fails case(s) FAILED"
    exit "$fails"
fi

NEW="$WORK/build/boot.sd.new"
[ -f "$NEW" ] && note "it writes boot.sd.new" OK || {
    note "it writes boot.sd.new" FAIL
    echo
    echo "$fails case(s) FAILED"
    exit "$fails"
}

# Pull all three images back out of both files and compare them directly,
# rather than trusting whatever the tool reported about itself.
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

# Two changed lines means two removed and two added. A tool that disabled the
# right node and also dropped an unrelated property would pass a check that
# only looked at the two nodes below, so count the whole diff first.
removed=$(diff "$WORK/orig.dts" "$WORK/new.dts" | grep -c '^<')
added=$(diff "$WORK/orig.dts" "$WORK/new.dts" | grep -c '^>')

[ "$removed" = 2 ] && [ "$added" = 2 ] \
    && note "exactly two lines change in the source" OK \
    || note "exactly two lines change in the source ($removed out, $added in)" FAIL

# dtc indents with tabs, and the two nodes sit at different depths, so the
# marker and every leading blank has to go before the lines can be compared.
changed=$(diff "$WORK/orig.dts" "$WORK/new.dts" | grep -E '^[<>]' | sed 's/^[<>][[:space:]]*//' | sort -u)
expected='status = "disabled";
status = "okay";'

[ "$changed" = "$expected" ] \
    && note "every changed line is a status property" OK \
    || note "every changed line is a status property" FAIL

# Name the two nodes rather than infer them from diff context.
for node in /reserved-memory/cvifb /cvifb; do
    was=$(fdtget "$WORK/orig.2" "$node" status 2>/dev/null)
    now=$(fdtget "$WORK/new.2"  "$node" status 2>/dev/null)
    [ "$was" = okay ] && [ "$now" = disabled ] \
        && note "$node goes from okay to disabled" OK \
        || note "$node goes from okay to disabled (was=$was now=$now)" FAIL
done

# The carveout the capture pipeline allocates from must not move or shrink.
was=$(fdtget "$WORK/orig.2" /reserved-memory/ion size 2>/dev/null)
now=$(fdtget "$WORK/new.2"  /reserved-memory/ion size 2>/dev/null)
[ -n "$was" ] && [ "$was" = "$now" ] \
    && note "the ion reservation is untouched" OK \
    || note "the ion reservation is untouched (was=$was now=$now)" FAIL

was=$(fdtget "$WORK/orig.2" /memory@80000000 reg 2>/dev/null)
now=$(fdtget "$WORK/new.2"  /memory@80000000 reg 2>/dev/null)
[ -n "$was" ] && [ "$was" = "$now" ] \
    && note "the memory node is untouched" OK \
    || note "the memory node is untouched (was=$was now=$now)" FAIL

# mkimage stamps the FIT with the current time and rehashes each image, so
# compare the structure with those fields removed.
fitstruct() { fdtdump "$1" 2>/dev/null | grep -v -E '^\s+data = |timestamp|value = |^// '; }
fitstruct "$ORIG" > "$WORK/struct.orig"
fitstruct "$NEW"  > "$WORK/struct.new"
cmp -s "$WORK/struct.orig" "$WORK/struct.new" \
    && note "the FIT structure is unchanged" OK \
    || note "the FIT structure is unchanged" FAIL

echo
echo "===== an image that is already patched is refused ====="

if "$TOOL" "$NEW" "$WORK/again" > "$WORK/again.log" 2>&1; then
    note "a second pass is refused" FAIL
else
    note "a second pass is refused" OK
fi

grep -qi 'already' "$WORK/again.log" \
    && note "it says why it refused" OK \
    || note "it says why it refused" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
fi
exit "$fails"
