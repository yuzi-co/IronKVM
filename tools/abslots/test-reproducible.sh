#!/bin/sh
# Check that a FIT built twice from one set of inputs comes out byte-identical.
#
#   test-reproducible.sh <any-boot.sd>
#
# Needs dumpimage and mkimage. Run it in the container.
#
# mkimage stamps the current time into the FIT header, so for as long as this
# repository has built boot images, two builds from identical sources produced
# different bytes. A sha256 over a boot.sd therefore said nothing at all about
# whether a board was running this repository's image, and on 2026-09-11 that
# cost a real investigation: the merged tree rebuilt to 6e854fdb while the board
# carried 25b5a8e9 and git reported no difference in any input.
#
# repack-boot.sh now sets SOURCE_DATE_EPOCH, which mkimage writes in place of
# the clock. This file is the proof, and it is deliberately two cases rather
# than one:
#
#   the same epoch twice must produce the same bytes. That is the property.
#
#   two different epochs must produce different bytes. Without this case, a
#   mkimage that ignored the variable and happened to run both builds inside
#   one second would pass the first case and prove nothing. The control is what
#   makes the result mean that SOURCE_DATE_EPOCH is what decides.
#
# The input can be any FIT image. Nothing here repacks an initramfs, because the
# question is about the timestamp in the header and not about the payload.
set -u

ORIG=${1:-}
[ -n "$ORIG" ] && [ -f "$ORIG" ] || {
    echo "$(basename "$0"): needs a boot.sd to measure." >&2
    echo "any FIT image will do: the stock one, or a built one." >&2
    exit 2
}

need() {
    for _cmd in "$@"
    do
        command -v "$_cmd" >/dev/null 2>&1 && continue
        echo "$(basename "$0"): needs $_cmd, which is not on PATH." >&2
        echo "u-boot-tools carries both." >&2
        exit 2
    done
}
need dumpimage mkimage

HERE=$(cd "$(dirname "$0")" && pwd)
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT
FAIL=0

t() {
    if [ "$2" = "0" ]; then echo "ok   - $1"; else echo "FAIL - $1"; FAIL=1; fi
}

dumpimage -T flat_dt -p 0 -o "$WORK/kernel"  "$ORIG" >/dev/null 2>&1
dumpimage -T flat_dt -p 1 -o "$WORK/ramdisk" "$ORIG" >/dev/null 2>&1
dumpimage -T flat_dt -p 2 -o "$WORK/fdt"     "$ORIG" >/dev/null 2>&1
[ -s "$WORK/kernel" ] && [ -s "$WORK/ramdisk" ] && [ -s "$WORK/fdt" ] || {
    echo "$(basename "$0"): $ORIG is not a three part FIT this board would boot." >&2
    exit 2
}

sed -e "s#@KERNEL@#$WORK/kernel#" -e "s#@RAMDISK@#$WORK/ramdisk#" \
    -e "s#@FDT@#$WORK/fdt#" "$HERE/boot.its.in" > "$WORK/boot.its"

build() {
    SOURCE_DATE_EPOCH=$1 mkimage -f "$WORK/boot.its" "$2" >/dev/null 2>&1
}

build 1600000000 "$WORK/a.sd"
build 1600000000 "$WORK/b.sd"
build 1700000000 "$WORK/c.sd"

[ -s "$WORK/a.sd" ] && [ -s "$WORK/b.sd" ] && [ -s "$WORK/c.sd" ]
t "three images were built" $?

cmp -s "$WORK/a.sd" "$WORK/b.sd"
t "one epoch twice gives identical bytes" $?

if cmp -s "$WORK/a.sd" "$WORK/c.sd"; then
    echo "FAIL - a different epoch gives different bytes"
    echo "       mkimage is ignoring SOURCE_DATE_EPOCH, so the case above passed"
    echo "       by running both builds inside one second and proves nothing."
    FAIL=1
else
    echo "ok   - a different epoch gives different bytes"
fi

# The repack script has to be the thing that sets it, or the property belongs to
# this test rather than to the build.
grep -q 'SOURCE_DATE_EPOCH' "$HERE/repack-boot.sh"
t "repack-boot.sh sets SOURCE_DATE_EPOCH" $?

exit $FAIL
