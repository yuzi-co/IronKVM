#!/bin/sh
# Check that S01fs finds /data from the slot configuration, not from a
# hardcoded partition number.
#
#   test-s01fs-datadev.sh [path-to-S01fs]
#
# In the A/B layout p3 is root B and /data is p6. S01fs shipped with
# `mount /dev/mmcblk0p3 /data`, so on the first A/B boot /data was not mounted,
# S02identity had nothing to bind, and the board came up with the image's own
# /etc/kvm instead of the device's.
#
# The same hardcode has a worse form a few lines above: the /boot/usb.disk0
# branch runs `parted mkpart` and `mkfs.exfat /dev/mmcblk0p3`. In this layout
# that formats root B. It is gated on /etc/kvm.disk0, and the image manifest
# creates that marker, which is what stopped it on the real first boot. Belt and
# braces: a layout that declares its own devices must never reach that branch at
# all.
S01=${1:-$(dirname "$0")/../../../kvmapp/system/init.d/S01fs}
[ -f "$S01" ] || { echo "usage: test-s01fs-datadev.sh <S01fs>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

echo "===== the data device comes from configuration ====="

sed -n '/^# --- data device ---/,/^# --- end data device ---/p' "$S01" > "$WORK/dd.sh"
if [ ! -s "$WORK/dd.sh" ]; then
    note "the data device block can be extracted" FAIL
    echo; echo "$fails case(s) FAILED"; exit 1
fi
note "the data device block can be extracted" OK

# With a conf that names p6, that is what must be used.
cat > "$WORK/slots.conf" <<'CONF'
SLOT_A=/dev/mmcblk0p2
SLOT_B=/dev/mmcblk0p3
RECOVERY=/dev/mmcblk0p5
DATA_DEV=/dev/mmcblk0p6
CONF

got=$( SLOT_CONF="$WORK/slots.conf" sh -c ". $WORK/dd.sh; data_device" )
[ "$got" = /dev/mmcblk0p6 ] \
    && note "a conf naming p6 is honoured" OK \
    || note "a conf naming p6 gave '$got'" FAIL

# With no conf at all, the old single-root layout must still work, because a
# board that has not been migrated still boots this script.
got=$( SLOT_CONF="$WORK/absent.conf" sh -c ". $WORK/dd.sh; data_device" )
[ "$got" = /dev/mmcblk0p3 ] \
    && note "no conf falls back to p3, the pre-A/B layout" OK \
    || note "no conf gave '$got', want /dev/mmcblk0p3" FAIL

# A conf with no DATA_DEV line is a conf from an older build. Fall back rather
# than mount nothing.
printf 'SLOT_A=/dev/mmcblk0p2\n' > "$WORK/partial.conf"
got=$( SLOT_CONF="$WORK/partial.conf" sh -c ". $WORK/dd.sh; data_device" )
[ "$got" = /dev/mmcblk0p3 ] \
    && note "a conf without DATA_DEV falls back to p3" OK \
    || note "a conf without DATA_DEV gave '$got'" FAIL

echo
echo "===== no bare partition number survives in the mount path ====="

# The mount must not name a partition directly any more.
if grep -qE '^[[:space:]]*mount[[:space:]]+/dev/mmcblk0p3[[:space:]]+/data' "$S01"; then
    note "the /data mount no longer hardcodes p3" FAIL
else
    note "the /data mount no longer hardcodes p3" OK
fi

echo
echo "===== the auto-partition branch cannot run on a declared layout ====="

# parted mkpart plus mkfs.exfat on p3 formats root B in the A/B layout.
sed -n '/^# --- autopartition guard ---/,/^# --- end autopartition guard ---/p' "$S01" > "$WORK/ap.sh"
[ -s "$WORK/ap.sh" ] \
    && note "the autopartition guard block can be extracted" OK \
    || note "the autopartition guard block can be extracted" FAIL

got=$( SLOT_CONF="$WORK/slots.conf" sh -c ". $WORK/ap.sh; may_autopartition && echo yes || echo no" )
[ "$got" = no ] \
    && note "a declared layout refuses to autopartition" OK \
    || note "a declared layout would autopartition, which formats root B" FAIL

got=$( SLOT_CONF="$WORK/absent.conf" sh -c ". $WORK/ap.sh; may_autopartition && echo yes || echo no" )
[ "$got" = yes ] \
    && note "an undeclared layout may still autopartition as before" OK \
    || note "an undeclared layout can no longer autopartition, a regression" FAIL

echo
echo "===== the script still parses ====="
sh -n "$S01" 2>/dev/null && note "sh -n accepts S01fs" OK || note "sh -n accepts S01fs" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
fi
exit "$fails"
