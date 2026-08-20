#!/bin/sh
# Shrink the ION carveout in a NanoKVM boot image and give the difference to Linux.
#
#   resize-ion.sh <boot.sd> <workdir> <new-size-bytes>
#
# Needs: u-boot-tools and device-tree-compiler. Run it in a throwaway
# container; nothing here has to run on the device.
#
# The board reserves 0x4b00000 bytes (75MB) for the video pipeline at
# /reserved-memory/ion. The region is not CMA, so none of it comes back when
# capture is idle: whatever the reservation says is gone from Linux for the
# whole uptime. On a board with 166MB that Linux can address, the reservation
# is the single largest thing standing between the fork and its memory
# pressure.
#
# The size is a device tree property rather than a build constant, which is
# what makes this a boot image edit instead of an SDK rebuild. The node is
# dynamic - it carries "size" with no fixed "reg" - so the kernel places the
# region itself and only the length changes here.
#
# What the pipeline actually needs, measured on the device 2026-08-20 at
# 1920x1080 60fps, which is the LT6911 sensor's ceiling, with MJPEG, H.264
# direct and WebRTC all exercised from a browser:
#
#     peak                       42,942,464   (55% of the reservation)
#     idle, capture up           19,050,496
#     one orphaned generation     6,516,736   (a crash leaks this; a clean stop
#                                              no longer does)
#
# So the floor below is the peak plus one orphan. Under it the board cannot
# work, and the failure is not graceful: libkvm does not check the result of an
# allocation, so an exhausted carveout is a segfault rather than an error, and
# the server dies before it binds its port. See tools/README.md.
set -e

ORIG=${1:?usage: resize-ion.sh <boot.sd> <workdir> <new-size-bytes>}
OUT=${2:?usage: resize-ion.sh <boot.sd> <workdir> <new-size-bytes>}
WANT=${3:?usage: resize-ion.sh <boot.sd> <workdir> <new-size-bytes>}
HERE=$(cd "$(dirname "$0")" && pwd)

# The measured peak plus one orphaned generation. A size under this is refused
# rather than warned about, because the board that boots it answers nothing.
MIN_SIZE=49459200

# The reservation is placed by the kernel, but a length that is not a whole
# number of megabytes buys nothing and makes the value hard to read against the
# device tree it came from.
ALIGN=1048576

# One FIT layout, described in one place. tools/fbmem and tools/slots build the
# same image with a different payload swapped in, and a second copy of this
# template would drift.
ITS_IN="$HERE/../slots/boot.its.in"

[ -f "$ORIG" ]   || { echo "no such file: $ORIG"; exit 1; }
[ -f "$ITS_IN" ] || { echo "missing FIT template: $ITS_IN"; exit 1; }

case "$WANT" in
    ''|*[!0-9]*) echo "size must be a number of bytes: $WANT"; exit 1 ;;
esac

mkdir -p "$OUT"

# dtc resolves /incbin/ relative to the directory of the .its file, not to the
# working directory, so a relative workdir makes mkimage look for the payloads
# inside itself. Resolve both paths before any of them reaches the template.
OUT=$(cd "$OUT" && pwd)
ORIG=$(cd "$(dirname "$ORIG")" && pwd)/$(basename "$ORIG")

echo "############ 1. take the image apart"
dumpimage -T flat_dt -p 0 -o "$OUT/kernel.orig"  "$ORIG" >/dev/null
dumpimage -T flat_dt -p 1 -o "$OUT/ramdisk.orig" "$ORIG" >/dev/null
dumpimage -T flat_dt -p 2 -o "$OUT/fdt.orig"     "$ORIG" >/dev/null
for f in kernel ramdisk fdt; do
    printf '  %-12s %s bytes\n' "$f" "$(wc -c < "$OUT/$f.orig")"
done

echo
echo "############ 2. refuse a size the board cannot run"

# fdtget prints the two cells of a 64-bit size separated by a space.
cells=$(fdtget "$OUT/fdt.orig" /reserved-memory/ion size 2>/dev/null || echo missing)
if [ "$cells" = missing ]; then
    echo "  /reserved-memory/ion has no size; this is not a NanoKVM boot image"
    exit 1
fi
hi=${cells% *}
lo=${cells#* }
[ "$hi" = 0 ] || { echo "  reservation is above 4GB ($cells); this tool does not handle that"; exit 1; }
CUR=$lo

printf '  %-40s %s bytes (%s MB)\n' "current /reserved-memory/ion size" "$CUR" "$((CUR / 1048576))"
printf '  %-40s %s bytes (%s MB)\n' "requested size" "$WANT" "$((WANT / 1048576))"

[ "$WANT" -ne "$CUR" ] || { echo "  already that size, nothing to do"; exit 1; }
[ "$WANT" -lt "$CUR" ] || { echo "  this tool only shrinks the reservation"; exit 1; }
[ $((WANT % ALIGN)) -eq 0 ] || { echo "  size must be a whole number of MB"; exit 1; }
if [ "$WANT" -lt "$MIN_SIZE" ]; then
    echo "  refused: under the measured floor of $MIN_SIZE bytes"
    echo "  that is the 42,942,464 peak plus one orphaned generation. A board"
    echo "  that boots a smaller reservation segfaults its server on start."
    exit 1
fi

printf '  %-40s %s bytes (%s MB)\n' "returned to Linux" "$((CUR - WANT))" "$(((CUR - WANT) / 1048576))"

echo
echo "############ 3. rewrite the size"
dtc -I dtb -O dts -o "$OUT/fdt.orig.dts" "$OUT/fdt.orig" 2>/dev/null

# Edit by node name rather than by line number, the same way tools/fbmem does.
# Exactly one node in this tree is called "ion", and the count is checked after,
# so a tree that grows a second one fails here rather than being half-edited.
awk -v want="$WANT" '
    # A node opens as: <name> {   or   <label>: <name> {
    /\{[[:space:]]*$/ {
        name = $0
        sub(/^[[:space:]]*/, "", name)
        sub(/[[:space:]]*\{[[:space:]]*$/, "", name)
        sub(/^[^:]*:[[:space:]]*/, "", name)
        depth++
        stack[depth] = name
    }
    /^[[:space:]]*\};?[[:space:]]*$/ {
        if (depth > 0) { stack[depth] = ""; depth-- }
    }
    {
        if (depth > 0 && stack[depth] == "ion" && $0 ~ /size = <0x[0-9a-f]+ 0x[0-9a-f]+>;/) {
            sub(/size = <0x[0-9a-f]+ 0x[0-9a-f]+>;/, sprintf("size = <0x00 0x%x>;", want))
            changed++
        }
        print
    }
    END { if (changed != 1) { print "expected 1 edit, made " changed+0 > "/dev/stderr"; exit 1 } }
' "$OUT/fdt.orig.dts" > "$OUT/fdt.new.dts"

dtc -I dts -O dtb -o "$OUT/fdt.new" "$OUT/fdt.new.dts" 2>/dev/null
printf '  %-40s %s bytes (was %s)\n' "fdt.new" \
    "$(wc -c < "$OUT/fdt.new")" "$(wc -c < "$OUT/fdt.orig")"

echo
echo "############ 4. build the FIT"
sed -e "s#@KERNEL@#$OUT/kernel.orig#" -e "s#@RAMDISK@#$OUT/ramdisk.orig#" \
    -e "s#@FDT@#$OUT/fdt.new#" "$ITS_IN" > "$OUT/boot.its"
mkimage -f "$OUT/boot.its" "$OUT/boot.sd.new" >/dev/null

echo
echo "############ 5. verify"
fail=0
note() { printf '  %-58s %s\n' "$1" "$2"; [ "$2" = "FAIL" ] && fail=1; return 0; }

dumpimage -T flat_dt -p 0 -o "$OUT/kernel.chk"  "$OUT/boot.sd.new" >/dev/null
dumpimage -T flat_dt -p 1 -o "$OUT/ramdisk.chk" "$OUT/boot.sd.new" >/dev/null
dumpimage -T flat_dt -p 2 -o "$OUT/fdt.chk"     "$OUT/boot.sd.new" >/dev/null

cmp -s "$OUT/kernel.orig"  "$OUT/kernel.chk"  && note "kernel byte-identical to original" OK  || note "kernel byte-identical to original" FAIL
cmp -s "$OUT/ramdisk.orig" "$OUT/ramdisk.chk" && note "ramdisk byte-identical to original" OK || note "ramdisk byte-identical to original" FAIL
cmp -s "$OUT/fdt.new"      "$OUT/fdt.chk"     && note "device tree survives the FIT round-trip" OK || note "device tree survives the FIT round-trip" FAIL

now=$(fdtget "$OUT/fdt.chk" /reserved-memory/ion size 2>/dev/null || echo missing)
[ "$now" = "0 $WANT" ] \
    && note "/reserved-memory/ion size is $WANT" OK \
    || note "/reserved-memory/ion size is $WANT (got $now)" FAIL

# Everything else about the reservation, and the framebuffer change this fork
# already ships, has to survive untouched.
for probe in "/reserved-memory/ion compatible" "/reserved-memory/cvifb status" \
             "/reserved-memory/cvifb size" "/memory@80000000 reg"; do
    set -- $probe
    was=$(fdtget "$OUT/fdt.orig" "$1" "$2" 2>/dev/null || echo missing)
    then_=$(fdtget "$OUT/fdt.chk" "$1" "$2" 2>/dev/null || echo missing)
    [ "$was" = "$then_" ] && note "$1 $2 unchanged" OK || note "$1 $2 unchanged (was $was, now $then_)" FAIL
done

# Only the one size value may differ once both blobs are read back as source.
dtc -I dtb -O dts -o "$OUT/fdt.chk.dts" "$OUT/fdt.chk" 2>/dev/null
removed=$(diff "$OUT/fdt.orig.dts" "$OUT/fdt.chk.dts" | grep -c '^<' || true)
added=$(diff "$OUT/fdt.orig.dts" "$OUT/fdt.chk.dts" | grep -c '^>' || true)
[ "$removed" = 1 ] && [ "$added" = 1 ] \
    && note "exactly one line differs in the source" OK \
    || note "exactly one line differs in the source ($removed out, $added in)" FAIL

# mkimage stamps the FIT with the current time and rehashes each image.
fitstruct() { fdtdump "$1" 2>/dev/null | grep -v -E '^\s+data = |timestamp|value = |^// '; }
fitstruct "$ORIG" > "$OUT/struct.orig"; fitstruct "$OUT/boot.sd.new" > "$OUT/struct.new"
cmp -s "$OUT/struct.orig" "$OUT/struct.new" \
    && note "FIT structure identical (ignoring data, hashes, timestamp)" OK \
    || note "FIT structure differs" FAIL

if [ "$fail" -ne 0 ]; then
    echo
    echo "verification failed; removing the output so it cannot be installed"
    rm -f "$OUT/boot.sd.new"
    exit 1
fi

NEWSZ=$(wc -c < "$OUT/boot.sd.new")
ORIGSZ=$(wc -c < "$ORIG")
echo
echo "############ 6. install"
printf '  %-40s %s\n' "image" "$OUT/boot.sd.new"
printf '  %-40s %s\n' "sha256" "$(sha256sum "$OUT/boot.sd.new" | cut -d' ' -f1)"
printf '  %-40s %s bytes (was %s, growth %s)\n' "size" "$NEWSZ" "$ORIGSZ" "$((NEWSZ - ORIGSZ))"
cat <<EOF

  Copy it to the device, then:

    tools/slots/device/install-boot.sh /data/boot.sd.new <sha256>
    reboot

  It keeps a stock image at /data/boot.sd.orig the first time it runs.
  To go back, install that file and reboot.

  There is one boot image on this board, not one per slot, so nothing reverts
  this automatically. The realistic failure is a kernel that boots and a server
  that cannot allocate, which leaves ssh up and the image restorable from here.

  After the reboot:

    /proc/device-tree/reserved-memory/ion/size    should read $WANT
    MemTotal                                      should grow by $(((CUR - WANT) / 1024)) kB
    alloc_mem under a stream                      should stay near 42,942,464
EOF
