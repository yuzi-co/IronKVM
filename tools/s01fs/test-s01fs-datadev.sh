#!/bin/sh
# Check that S01fs finds /data from the device description, not from a
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
S01=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S01fs}
[ -f "$S01" ] || { echo "usage: test-s01fs-datadev.sh <S01fs>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

# S01fs reads the device description through ironkvm-deviceinfo now. The real
# reader is used, not a stub, because a stub would not catch a key this script
# asks for under the wrong name.
READER=${READER:-$(cd "$(dirname "$0")/../.." && pwd)/kvmapp/system/ironkvm-deviceinfo}
[ -f "$READER" ] || { echo "needs kvmapp/system/ironkvm-deviceinfo"; exit 2; }
mkdir -p "$WORK/bin"
cat > "$WORK/bin/ironkvm-deviceinfo" <<STUB
#!/bin/sh
DEVICEINFO_PATHS="$WORK/deviceinfo" exec sh "$READER" "\$@"
STUB
chmod 755 "$WORK/bin/ironkvm-deviceinfo"
PATH="$WORK/bin:$PATH"
export PATH

# Write a description for the next case. Each argument is one KEY=value line.
deviceinfo() { printf '%s\n' "$@" > "$WORK/deviceinfo"; }

echo "===== the data device comes from the device description ====="

sed -n '/^# --- data device ---/,/^# --- end data device ---/p' "$S01" > "$WORK/dd.sh"
if [ ! -s "$WORK/dd.sh" ]; then
    note "the data device block can be extracted" FAIL
    echo; echo "$fails case(s) FAILED"; exit 1
fi
note "the data device block can be extracted" OK

run_data_device() { sh -c ". $WORK/dd.sh; data_device" 2>/dev/null; }
run_data_start()  { sh -c ". $WORK/dd.sh; data_start" 2>/dev/null; }

# The board in use. DATA is a partition number, and the reader turns it into a
# device with the separator the disk asks for.
deviceinfo DISK=/dev/mmcblk0 SLOT_A=2 SLOT_B=3 RECOVERY=5 DATA=6 BOOT_PART=1 DATA_START=10543104
got=$(run_data_device)
[ "$got" = /dev/mmcblk0p6 ] \
    && note "the data device comes from DISK and DATA" OK \
    || note "the data device gave '$got', want /dev/mmcblk0p6" FAIL

# Another board, another disk and another partition. This is the whole point of
# reading a description: no edit here describes a second layout.
deviceinfo DISK=/dev/sda SLOT_A=2 SLOT_B=3 DATA=4 BOOT_PART=1 DATA_START=99
got=$(run_data_device)
[ "$got" = /dev/sda4 ] \
    && note "another layout gives another device" OK \
    || note "another layout gave '$got', want /dev/sda4" FAIL

deviceinfo DISK=/dev/mmcblk0 SLOT_A=2 SLOT_B=3 DATA=6 BOOT_PART=1 DATA_START=10543104
got=$(run_data_start)
[ "$got" = 10543104 ] \
    && note "the data start comes from the same file" OK \
    || note "the data start gave '$got', want 10543104" FAIL

# A description with no DATA_START is the copy the application tarball carries.
# The caller makes no partition at all then.
deviceinfo DISK=/dev/mmcblk0 SLOT_A=2 SLOT_B=3 DATA=6 BOOT_PART=1
got=$(run_data_start)
[ -z "$got" ] \
    && note "a description without DATA_START gives nothing" OK \
    || note "a description without DATA_START gave '$got'" FAIL

# No description at all. The old fallback was /dev/mmcblk0p3, which in the A/B
# layout is a root filesystem, so a guess here formats a slot.
rm -f "$WORK/deviceinfo"
got=$(run_data_device)
[ -z "$got" ] \
    && note "no deviceinfo leaves the data device empty rather than guessing" OK \
    || note "no deviceinfo guessed '$got'" FAIL

# And nothing in the block names a device of its own.
if grep -n 'mmcblk0p[0-9]' "$WORK/dd.sh" | grep -qv '^[0-9]*:#'; then
    note "the data device block names no partition of its own" FAIL
else
    note "the data device block names no partition of its own" OK
fi

echo
echo "===== no bare partition number survives in the mount path ====="

# The mount must not name a partition directly any more.
if grep -qE '^[[:space:]]*mount[[:space:]]+/dev/mmcblk0p3[[:space:]]+/data' "$S01"; then
    note "the /data mount no longer hardcodes p3" FAIL
else
    note "the /data mount no longer hardcodes p3" OK
fi

# /boot is where the trial marker and the boot counter live, and its partition
# number is a fact about the layout like any other.
if grep -qE '^[[:space:]]*mount -t vfat[[:space:]]+"\$\(devinfo_part BOOT_PART\)"[[:space:]]+/boot' "$S01"; then
    note "the /boot mount takes its device from BOOT_PART" OK
else
    note "the /boot mount still names a partition" FAIL
fi

echo
echo "===== the auto-partition branch cannot run on a declared layout ====="

# parted mkpart plus mkfs.exfat on p3 formats root B in the A/B layout.
sed -n '/^# --- autopartition guard ---/,/^# --- end autopartition guard ---/p' "$S01" > "$WORK/ap.sh"
[ -s "$WORK/ap.sh" ] \
    && note "the autopartition guard block can be extracted" OK \
    || note "the autopartition guard block can be extracted" FAIL

may_auto() { sh -c ". $WORK/ap.sh; may_autopartition && echo yes || echo no" 2>/dev/null; }

# DATA_START is written into the image by the build, so a description that
# carries it describes a card that was laid out deliberately.
deviceinfo DISK=/dev/mmcblk0 SLOT_A=2 SLOT_B=3 RECOVERY=5 DATA=6 BOOT_PART=1 DATA_START=10543104
got=$(may_auto)
[ "$got" = no ] \
    && note "a declared layout refuses to autopartition" OK \
    || note "a declared layout would autopartition, which formats root B" FAIL

# The copy in the application tarball describes the same board and carries no
# DATA_START. A board on Sipeed's firmware must still provision its stock data
# disk, so this one may.
deviceinfo DISK=/dev/mmcblk0 SLOT_A=2 SLOT_B=3 DATA=6 BOOT_PART=1
got=$(may_auto)
[ "$got" = yes ] \
    && note "a description without DATA_START may still autopartition" OK \
    || note "a tarball description disarmed the stock provisioning" FAIL

rm -f "$WORK/deviceinfo"
got=$(may_auto)
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
    exit 1
fi
