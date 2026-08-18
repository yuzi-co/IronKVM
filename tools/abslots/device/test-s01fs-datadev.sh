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
echo "===== /data is mounted so its identity files are not world readable ====="
#
# /data is exfat, which stores no POSIX mode, so every file on it takes its
# mode from the mount. The board keeps its root password hash and its
# authorized_keys there, and 0755 for those two is wrong even on a board whose
# only login user is root.
#
# fmask and dmask exist only on the FAT family. If /data is ever anything else
# the masked mount fails, and an unmounted /data is the exact fault the block
# above was written for, so the fallback matters more than the mask.

sed -n '/^# --- data mount ---/,/^# --- end data mount ---/p' "$S01" > "$WORK/dm.sh"
if [ ! -s "$WORK/dm.sh" ]; then
    note "the data mount block can be extracted" FAIL
else
    note "the data mount block can be extracted" OK

    cat > "$WORK/mountstub.sh" <<'STUB'
mount() {
    echo "mount $*" >> "$MOUNTLOG"
    case "$*" in
        *fmask*) [ "${STUB_MASK_FAILS:-no}" = yes ] && return 1 ;;
    esac
    [ "${STUB_ALL_FAILS:-no}" = yes ] && return 1
    return 0
}
STUB

    drive() {
        MOUNTLOG="$WORK/mlog"; : > "$MOUNTLOG"
        export MOUNTLOG STUB_MASK_FAILS STUB_ALL_FAILS
        sh -c ". $WORK/mountstub.sh; . $WORK/dm.sh; mount_data /dev/mmcblk0p6 /data" >/dev/null 2>&1
        echo $?
    }

    STUB_MASK_FAILS=no STUB_ALL_FAILS=no
    rc=$(drive)
    [ "$rc" = 0 ] && note "a masked mount that works reports success" OK \
                  || note "a masked mount that works returned $rc" FAIL
    grep -q 'fmask=0077' "$WORK/mlog" \
        && note "the mount masks files to 0077" OK \
        || note "the mount does not mask files, got: $(cat "$WORK/mlog")" FAIL
    grep -q 'dmask=0077' "$WORK/mlog" \
        && note "the mount masks directories to 0077" OK \
        || note "the mount does not mask directories" FAIL
    [ "$(wc -l < "$WORK/mlog")" -eq 1 ] \
        && note "and it does not mount twice" OK \
        || note "it mounted $(wc -l < "$WORK/mlog") times" FAIL

    # A filesystem with no fmask must still end up mounted. This is the case
    # that matters: /data unmounted is worse than /data world readable.
    STUB_MASK_FAILS=yes STUB_ALL_FAILS=no
    rc=$(drive)
    [ "$rc" = 0 ] && note "a filesystem with no fmask still mounts" OK \
                  || note "a filesystem with no fmask left /data unmounted (rc $rc)" FAIL
    [ "$(grep -c 'fmask' "$WORK/mlog")" -eq 1 ] && [ "$(wc -l < "$WORK/mlog")" -eq 2 ] \
        && note "it falls back to a plain mount, once" OK \
        || note "the fallback is wrong, got: $(tr '\n' ';' < "$WORK/mlog")" FAIL

    # And a device that cannot be mounted at all must report failure rather
    # than let the caller print that /data is ready.
    STUB_MASK_FAILS=no STUB_ALL_FAILS=yes
    rc=$(drive)
    [ "$rc" != 0 ] && note "a device that cannot mount reports failure" OK \
                   || note "a device that cannot mount reported success" FAIL
fi

# The call site must go through the function, or none of the above runs on the
# device.
if grep -qE 'mount_data[[:space:]]+"\$DATADEV"[[:space:]]+/data' "$S01"; then
    note "the call site uses mount_data" OK
else
    note "the call site still calls mount directly" FAIL
fi

# A mount that failed must not be reported as a mounted /data. S01fs discarded
# the return value and printed OK, so a board with no /data, the factory root
# password and a new ssh host key reported a clean boot. That is how the 1.0.0
# card image shipped without anybody noticing it made no data partition.
if grep -qE '^[[:space:]]*if[[:space:]]+mount_data[[:space:]]' "$S01"; then
    note "the call site tests whether the mount worked" OK
else
    note "the call site discards the result of the mount" FAIL
fi

if grep -q 'FAILED to mount' "$S01"; then
    note "and it says so when the mount fails" OK
else
    note "a failed mount is still silent" FAIL
fi

# The provisioning has to be wired in, or it never runs on the device.
if grep -qE '^[[:space:]]*provision_data[[:space:]]+"\$DATADEV"' "$S01"; then
    note "the boot path calls provision_data" OK
else
    note "provision_data is defined and never called" FAIL
fi

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
