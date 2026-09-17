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

# The card under the description. data_device falls back to the partition table
# when no DATA is declared, so every case says what the table holds and what
# blkid answers for each partition on it.
#
# The parted stub prints the table in use and exits 1, exactly as the real one
# does on a disk with a mounted partition. The blkid stub answers from one file
# per device name, and prints nothing for a partition no case described, which
# is what blkid does for a partition with no filesystem.
cat > "$WORK/bin/parted" <<STUB
#!/bin/sh
cat "\$PARTED_TABLE"
exit 1
STUB
chmod 755 "$WORK/bin/parted"

cat > "$WORK/bin/blkid" <<STUB
#!/bin/sh
f="$WORK/fs/\$(basename "\$1")"
[ -f "\$f" ] && cat "\$f"
exit 0
STUB
chmod 755 "$WORK/bin/blkid"

mkdir -p "$WORK/fs"
# card <table-file>; then fs <device name> <blkid output> for each partition.
card() { PARTED_TABLE=$1; export PARTED_TABLE; rm -rf "$WORK/fs"; mkdir -p "$WORK/fs"; }
fs() { printf '%s\n' "$2" > "$WORK/fs/$1"; }

: > "$WORK/no.table"

# The stock layout Sipeed's installer makes: boot, root, data.
cat > "$WORK/stock.table" <<'T'
BYT;
/dev/mmcblk0:15523840s:sd/mmc:512:512:msdos:SD SA08G:;
1:1s:32768s:32768s:fat16::boot, lba;
2:40960s:8429567s:8388608s:ext4::;
3:8429568s:15523839s:7094272s:::;
T

# The A/B layout, on a card laid out before the image carried a description.
# p3 is a root slot here, which is what makes the stock guess dangerous.
cat > "$WORK/ab.table" <<'T'
BYT;
/dev/mmcblk0:60506112s:sd/mmc:512:512:msdos:SD SA32G:;
1:1s:32768s:32768s:fat16::boot, lba;
2:40960s:4235263s:4194304s:ext4::;
3:4235264s:8429567s:4194304s:ext4::;
4:8429568s:60506111s:52076544s:::;
5:8437760s:10534911s:2097152s:ext4::;
6:10543104s:60506111s:49963008s:::;
T

card "$WORK/no.table"

echo "===== the data device comes from the device description ====="

sed -n '/^# --- data device ---/,/^# --- end data device ---/p' "$S01" > "$WORK/dd.sh"
if [ ! -s "$WORK/dd.sh" ]; then
    note "the data device block can be extracted" FAIL
    echo; echo "$fails case(s) FAILED"; exit 1
fi
note "the data device block can be extracted" OK

# data_device falls back to the table, and the one place that reads the table is
# the data geometry block. Both are sourced, in the order the script has them.
sed -n '/^# --- data geometry ---/,/^# --- end data geometry ---/p' "$S01" > "$WORK/geo.sh"
if [ ! -s "$WORK/geo.sh" ]; then
    note "the data geometry block can be extracted" FAIL
    echo; echo "$fails case(s) FAILED"; exit 1
fi
note "the data geometry block can be extracted" OK

run_data_device() {
    sh -c "PARTED=$WORK/bin/parted; BLKID=$WORK/bin/blkid; \
           . $WORK/dd.sh; . $WORK/geo.sh; data_device" 2>/dev/null
}
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
echo "===== a description with no DATA finds the partition on the card ====="
#
# The copy the application tarball carries declares no layout. It describes the
# device, and the card under it is whatever Sipeed's installer or an older build
# made, so the partition number is not something this file can state. It used to
# carry the image's own DATA=6, which named /dev/mmcblk0p6 on a stock card whose
# highest partition is p3: the device did not exist, /data was not mounted, and
# upstream's own behaviour of mounting p3 was lost.
#
# The two cards that reach this path want two different numbers, and on each of
# them the other card's number is wrong in a way that matters. p3 is the data
# partition of a stock card and a root slot of an A/B one.

# A declared DATA still wins, whatever the card says. A card the build laid out
# is authoritative about itself.
card "$WORK/stock.table"
fs mmcblk0p3 '/dev/mmcblk0p3: LABEL="data" UUID="EE8B-6CB5" TYPE="exfat"'
deviceinfo DISK=/dev/mmcblk0 SLOT_A=2 SLOT_B=3 RECOVERY=5 DATA=6 BOOT_PART=1
got=$(run_data_device)
[ "$got" = /dev/mmcblk0p6 ] \
    && note "a declared DATA beats what is on the card" OK \
    || note "a declared DATA gave '$got', want /dev/mmcblk0p6" FAIL

# A stock card. p3 holds the data filesystem, and upstream labels it data.
card "$WORK/stock.table"
fs mmcblk0p1 '/dev/mmcblk0p1: LABEL="BOOT" TYPE="vfat"'
fs mmcblk0p2 '/dev/mmcblk0p2: LABEL="rootfs" TYPE="ext4"'
fs mmcblk0p3 '/dev/mmcblk0p3: LABEL="data" UUID="EE8B-6CB5" TYPE="exfat"'
deviceinfo DISK=/dev/mmcblk0 BOOT_PART=1 DATA_FS=exfat
got=$(run_data_device)
[ "$got" = /dev/mmcblk0p3 ] \
    && note "a stock card gives p3" OK \
    || note "a stock card gave '$got', want /dev/mmcblk0p3" FAIL

# An A/B card laid out before the image carried a description. Here p3 is root B
# and the data partition is p6, so the stock number is a running system.
card "$WORK/ab.table"
fs mmcblk0p1 '/dev/mmcblk0p1: LABEL="BOOT" TYPE="vfat"'
fs mmcblk0p2 '/dev/mmcblk0p2: LABEL="slot-a" TYPE="ext4"'
fs mmcblk0p3 '/dev/mmcblk0p3: LABEL="slot-b" TYPE="ext4"'
fs mmcblk0p5 '/dev/mmcblk0p5: LABEL="recovery" TYPE="ext4"'
fs mmcblk0p6 '/dev/mmcblk0p6: LABEL="data" UUID="9DA7-0008" TYPE="exfat"'
deviceinfo DISK=/dev/mmcblk0 BOOT_PART=1 DATA_FS=exfat
got=$(run_data_device)
[ "$got" = /dev/mmcblk0p6 ] \
    && note "an A/B card with no DATA gives p6, not the stock p3" OK \
    || note "an A/B card gave '$got', want /dev/mmcblk0p6" FAIL

# A data partition an older installer made without a label. busybox blkid prints
# no TYPE field, so this case answers only on a board with util-linux, which is
# why the label is the first test and the type is the second.
card "$WORK/stock.table"
fs mmcblk0p2 '/dev/mmcblk0p2: LABEL="rootfs" TYPE="ext4"'
fs mmcblk0p3 '/dev/mmcblk0p3: UUID="EE8B-6CB5" TYPE="exfat"'
deviceinfo DISK=/dev/mmcblk0 BOOT_PART=1 DATA_FS=exfat
got=$(run_data_device)
[ "$got" = /dev/mmcblk0p3 ] \
    && note "an unlabelled partition is found by its filesystem type" OK \
    || note "an unlabelled data partition gave '$got', want /dev/mmcblk0p3" FAIL

# Nothing on the card holds a data filesystem. An empty answer and a message on
# the console are the whole point: the alternative is a typed partition number,
# and on one of the two layouts that number is a root slot.
card "$WORK/stock.table"
fs mmcblk0p2 '/dev/mmcblk0p2: LABEL="rootfs" TYPE="ext4"'
deviceinfo DISK=/dev/mmcblk0 BOOT_PART=1 DATA_FS=exfat
got=$(run_data_device)
[ -z "$got" ] \
    && note "no data filesystem anywhere gives no device" OK \
    || note "no data filesystem anywhere guessed '$got'" FAIL

if grep -q 'no data device: none declared' "$S01"; then
    note "and the boot path says so on the console" OK
else
    note "an empty data device is not reported on the console" FAIL
fi

# A partition the description calls a slot is never the data device, whatever
# blkid says about it. A card whose old data filesystem was written over by a
# slot image still carries the label until something formats it.
card "$WORK/stock.table"
fs mmcblk0p3 '/dev/mmcblk0p3: LABEL="data" TYPE="exfat"'
deviceinfo DISK=/dev/mmcblk0 BOOT_PART=1 SLOT_A=2 SLOT_B=3 DATA_FS=exfat
got=$(run_data_device)
[ -z "$got" ] \
    && note "a declared slot is never taken as the data device" OK \
    || note "a declared slot was taken as the data device: '$got'" FAIL

# And neither is the boot partition.
card "$WORK/stock.table"
fs mmcblk0p1 '/dev/mmcblk0p1: LABEL="data" TYPE="exfat"'
deviceinfo DISK=/dev/mmcblk0 BOOT_PART=1 DATA_FS=exfat
got=$(run_data_device)
[ -z "$got" ] \
    && note "the boot partition is never taken as the data device" OK \
    || note "the boot partition was taken as the data device: '$got'" FAIL

# The block builds a device path from a partition number, and so does the
# reader's part verb. Two spellings of one rule drift, so they are held against
# each other here.
card "$WORK/no.table"
for disk in /dev/mmcblk0 /dev/sda /dev/nvme0n1; do
    deviceinfo DISK="$disk" DATA=6 BOOT_PART=1
    want=$(ironkvm-deviceinfo part DATA 2>/dev/null)
    got=$(sh -c "PARTED=$WORK/bin/parted; BLKID=$WORK/bin/blkid; \
                 . $WORK/dd.sh; . $WORK/geo.sh; part_device 6" 2>/dev/null)
    [ -n "$want" ] && [ "$got" = "$want" ] \
        && note "part_device agrees with the reader on $disk" OK \
        || note "part_device gave '$got' and the reader '$want' on $disk" FAIL
done

card "$WORK/no.table"

echo
echo "===== the reader is found on a board that has none on PATH ====="
#
# An image built by ironkvm-dist installs the reader at
# /usr/bin/ironkvm-deviceinfo, so the bare name resolves through PATH. A board
# that runs Sipeed's firmware installs the application tarball instead, and its
# install.sh comes from the official base tarball, which the fork cannot change:
# nothing puts the reader on PATH there. The tarball carries the reader at
# /kvmapp/system/ironkvm-deviceinfo, so that is where the script looks next.
#
# Without that second look the reader was never found on such a board, the data
# device came out empty, and /data was not mounted. That is the exact fault this
# whole file exists to prevent.

# A second stub stands at the path the tarball installs, and answers from a
# description of its own, so every case below says which reader replied.
mkdir -p "$WORK/tarball"
cat > "$WORK/tarball/ironkvm-deviceinfo" <<STUB
#!/bin/sh
DEVICEINFO_PATHS="$WORK/tarball-deviceinfo" exec sh "$READER" "\$@"
STUB
chmod 755 "$WORK/tarball/ironkvm-deviceinfo"
TARBALL="$WORK/tarball/ironkvm-deviceinfo"
tarball_deviceinfo() { printf '%s\n' "$@" > "$WORK/tarball-deviceinfo"; }
tarball_deviceinfo DISK=/dev/sda DATA=4 BOOT_PART=1

# A third reader, for the case that names one outright.
mkdir -p "$WORK/other"
cat > "$WORK/other/reader" <<STUB
#!/bin/sh
DEVICEINFO_PATHS="$WORK/other-deviceinfo" exec sh "$READER" "\$@"
STUB
chmod 755 "$WORK/other/reader"
printf '%s\n' DISK=/dev/vda DATA=7 BOOT_PART=1 > "$WORK/other-deviceinfo"

# A PATH with no reader on it, which is what a stock firmware board has.
BARE=/usr/bin:/bin

deviceinfo DISK=/dev/mmcblk0 SLOT_A=2 SLOT_B=3 RECOVERY=5 DATA=6 BOOT_PART=1 DATA_START=10543104

got=$(DEVINFO_TARBALL="$TARBALL" sh -c ". $WORK/dd.sh; data_device" 2>/dev/null)
[ "$got" = /dev/mmcblk0p6 ] \
    && note "a reader on PATH is the one that answers" OK \
    || note "a reader on PATH gave '$got', want /dev/mmcblk0p6" FAIL

got=$(PATH="$BARE" DEVINFO_TARBALL="$TARBALL" sh -c ". $WORK/dd.sh; data_device" 2>/dev/null)
[ "$got" = /dev/sda4 ] \
    && note "no reader on PATH falls back to the tarball's copy" OK \
    || note "no reader on PATH gave '$got', want /dev/sda4" FAIL

got=$(DEVINFO="$WORK/other/reader" DEVINFO_TARBALL="$TARBALL" sh -c ". $WORK/dd.sh; data_device" 2>/dev/null)
[ "$got" = /dev/vda7 ] \
    && note "a DEVINFO from the environment beats both" OK \
    || note "a DEVINFO from the environment gave '$got', want /dev/vda7" FAIL

# A DEVINFO that names a reader which is not there keeps the behaviour this
# block already has for a missing reader: no device, rather than a guess. The
# suites drive that case this way, so the fallback must not rescue it.
got=$(DEVINFO="$WORK/no-such-reader" DEVINFO_TARBALL="$TARBALL" sh -c ". $WORK/dd.sh; data_device" 2>/dev/null)
[ -z "$got" ] \
    && note "a DEVINFO that names nothing still gives no device" OK \
    || note "a DEVINFO that names nothing gave '$got'" FAIL

# The fallback is in /kvmapp because that is the only path the application
# tarball owns. A file written into the vendor root filesystem is removed by the
# next firmware update and belongs to nobody.
if grep -q '^DEVINFO_TARBALL=${DEVINFO_TARBALL:-/kvmapp/system/ironkvm-deviceinfo}$' "$S01"; then
    note "the fallback is the tarball's own path" OK
else
    note "the fallback is not /kvmapp/system/ironkvm-deviceinfo" FAIL
fi

if grep -v '^[[:space:]]*#' "$S01" | grep -q '/usr/'; then
    note "no line outside a comment names a path in /usr" FAIL
else
    note "no line outside a comment names a path in /usr" OK
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

# The guard reads the reader the data device block resolved, so both blocks are
# sourced here, in the order the script has them.
may_auto() { sh -c ". $WORK/dd.sh; . $WORK/ap.sh; may_autopartition && echo yes || echo no" 2>/dev/null; }

# And it names no reader of its own. Two spellings of the same name drift, and a
# guard that reads a different description from the rest of the script arms the
# branch that formats a root slot.
if grep -v '^[[:space:]]*#' "$WORK/ap.sh" | grep -q 'ironkvm-deviceinfo'; then
    note "the guard names no reader of its own" FAIL
else
    note "the guard names no reader of its own" OK
fi

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

# On a stock firmware board the reader is the tarball's copy, and the guard has
# to read the same one. A description there that declares DATA_START must disarm
# the branch exactly as the image's own description does.
tarball_deviceinfo DISK=/dev/sda DATA=4 BOOT_PART=1 DATA_START=99
got=$(PATH="$BARE" DEVINFO_TARBALL="$TARBALL" \
    sh -c ". $WORK/dd.sh; . $WORK/ap.sh; may_autopartition && echo yes || echo no" 2>/dev/null)
[ "$got" = no ] \
    && note "the guard reads the reader the script resolved" OK \
    || note "the guard gave '$got' for a declared layout with no reader on PATH" FAIL

tarball_deviceinfo DISK=/dev/sda DATA=4 BOOT_PART=1
got=$(PATH="$BARE" DEVINFO_TARBALL="$TARBALL" \
    sh -c ". $WORK/dd.sh; . $WORK/ap.sh; may_autopartition && echo yes || echo no" 2>/dev/null)
[ "$got" = yes ] \
    && note "and a tarball description still permits the stock branch" OK \
    || note "a tarball description disarmed the stock branch, got '$got'" FAIL

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
