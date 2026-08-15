#!/bin/sh
# Turn off the framebuffer reservation in a NanoKVM boot image.
#
#   repack-fdt.sh <stock-boot.sd> <workdir>
#
# Needs: u-boot-tools and device-tree-compiler. Run it in a throwaway
# container; nothing here has to run on the device.
#
# The board reserves 0x7d0000 bytes (8000 KiB) at 0x8ab30000 for a framebuffer
# it does not have. The region is described only by the device tree, so this
# tool rewrites the device tree inside boot.sd and leaves the kernel and the
# ramdisk exactly as they were.
#
# It sets status = "disabled" on two nodes and changes nothing else:
#
#   /reserved-memory/cvifb    the reservation itself
#   /cvifb                    the "cvitek,fb" device that consumes it
#
# The kernel honours that. __fdt_scan_reserved_mem() in
# linux_5.10/drivers/of/fdt.c returns early for a node that is not available,
# before it reserves anything, so the region is never taken and the memory
# stays with Linux.
#
# Disabling beats deleting. The reservation is dynamic - it carries "size" and
# "alloc-ranges" rather than a fixed "reg" - and /cvifb refers to it by
# phandle. Removing the node would leave that phandle dangling, and a two-value
# edit is one that can be read off a diff and reversed by hand.
#
# Upstream reaches the same end differently: sipeed/LicheeRV-Nano-Build#836
# sets FRAMEBUFFER_SIZE to zero in memmap.py at build time. That is the better
# fix for anyone who builds the image. This fork does not build it.
set -e

ORIG=${1:?usage: repack-fdt.sh <stock-boot.sd> <workdir>}
OUT=${2:?usage: repack-fdt.sh <stock-boot.sd> <workdir>}
HERE=$(cd "$(dirname "$0")" && pwd)

# One FIT layout, described in one place. tools/slots builds the same image
# with a different payload swapped in, and two copies of this template would
# drift.
ITS_IN="$HERE/../slots/boot.its.in"

[ -f "$ORIG" ]   || { echo "no such file: $ORIG"; exit 1; }
[ -f "$ITS_IN" ] || { echo "missing FIT template: $ITS_IN"; exit 1; }

mkdir -p "$OUT"

echo "############ 1. take the image apart"
dumpimage -T flat_dt -p 0 -o "$OUT/kernel.orig"  "$ORIG" >/dev/null
dumpimage -T flat_dt -p 1 -o "$OUT/ramdisk.orig" "$ORIG" >/dev/null
dumpimage -T flat_dt -p 2 -o "$OUT/fdt.orig"     "$ORIG" >/dev/null
for f in kernel ramdisk fdt; do
    printf '  %-12s %s bytes\n' "$f" "$(wc -c < "$OUT/$f.orig")"
done

echo
echo "############ 2. refuse an image that already has the change"
state=$(fdtget "$OUT/fdt.orig" /reserved-memory/cvifb status 2>/dev/null || echo missing)
printf '  %-40s %s\n' "/reserved-memory/cvifb status" "$state"
case "$state" in
    okay)    ;;
    missing) echo "  the node is absent; this is not a NanoKVM boot image"; exit 1 ;;
    *)       echo "  already patched: status is \"$state\", nothing to do"; exit 1 ;;
esac

echo
echo "############ 3. disable the reservation and its consumer"
dtc -I dtb -O dts -o "$OUT/fdt.orig.dts" "$OUT/fdt.orig" 2>/dev/null

# Edit by node name rather than by line number. The two nodes are both called
# "cvifb" and no other node is, so tracking the name of the block a line sits
# in reaches exactly those two status properties. The count is checked after.
awk '
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
        if (depth > 0 && stack[depth] == "cvifb" && $0 ~ /status = "okay";/) {
            sub(/status = "okay";/, "status = \"disabled\";")
            changed++
        }
        print
    }
    END { if (changed != 2) { print "expected 2 edits, made " changed+0 > "/dev/stderr"; exit 1 } }
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

for node in /reserved-memory/cvifb /cvifb; do
    now=$(fdtget "$OUT/fdt.chk" "$node" status 2>/dev/null || echo missing)
    [ "$now" = disabled ] && note "$node is disabled" OK || note "$node is disabled (got $now)" FAIL
done

for probe in "/reserved-memory/ion size" "/memory@80000000 reg" "/reserved-memory/cvifb size"; do
    set -- $probe
    was=$(fdtget "$OUT/fdt.orig" "$1" "$2" 2>/dev/null || echo missing)
    now=$(fdtget "$OUT/fdt.chk"  "$1" "$2" 2>/dev/null || echo missing)
    [ "$was" = "$now" ] && note "$1 $2 unchanged" OK || note "$1 $2 unchanged (was $was, now $now)" FAIL
done

# Only the two status values may differ once both blobs are read back as source.
dtc -I dtb -O dts -o "$OUT/fdt.chk.dts" "$OUT/fdt.chk" 2>/dev/null
removed=$(diff "$OUT/fdt.orig.dts" "$OUT/fdt.chk.dts" | grep -c '^<' || true)
added=$(diff "$OUT/fdt.orig.dts" "$OUT/fdt.chk.dts" | grep -c '^>' || true)
[ "$removed" = 2 ] && [ "$added" = 2 ] \
    && note "exactly two lines differ in the source" OK \
    || note "exactly two lines differ in the source ($removed out, $added in)" FAIL

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
cat <<'EOF'

  Copy it to the device, then:

    tools/slots/device/install-boot.sh /data/boot.sd.new <sha256>
    reboot

  It keeps a stock image at /data/boot.sd.orig the first time it runs.
  To go back, install that file and reboot.

  After the reboot, MemTotal should grow by about 8000 kB and
  /proc/device-tree/reserved-memory/cvifb/status should read "disabled".
EOF
