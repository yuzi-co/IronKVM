#!/bin/sh
# Check that S01fs reads the card's geometry correctly, and that it makes the
# data partition and its filesystem only when it should.
#
#   test-s01fs-provision.sh [path-to-S01fs]
#
# The 1.0.0 card image declared a data partition and nothing ever made a
# filesystem on it. The board came up with no /data, the factory root password
# and a new ssh host key, and S01fs printed OK. This file is the reason that
# cannot happen quietly again.
#
# Two device behaviours are stubbed exactly as measured, because the obvious
# implementation is wrong for both:
#
#   parted exits 1 on a busy disk AFTER it has succeeded. It cannot make the
#   kernel re-read the whole table while the root filesystem is mounted from the
#   same disk, so it warns and returns 1, having already written the table.
#
#   busybox blkid never prints a TYPE field, for any filesystem, and it exits 0
#   for a partition that holds none. Only an empty stdout means no filesystem.
S01=${1:-$(dirname "$0")/../../../kvmapp/system/init.d/S01fs}
[ -f "$S01" ] || { echo "usage: test-s01fs-provision.sh <S01fs>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

sed -n '/^# --- data geometry ---/,/^# --- end data geometry ---/p' "$S01" > "$WORK/geo.sh"
if [ ! -s "$WORK/geo.sh" ]; then
    note "the data geometry block can be extracted" FAIL
    echo; echo "$fails case(s) FAILED"; exit 1
fi
note "the data geometry block can be extracted" OK

# A parted stub that prints the machine-readable table of the card in use, and
# returns 1 the way the real one does on a busy disk.
cat > "$WORK/parted" <<'STUB'
#!/bin/sh
case "$*" in
    *print*) cat "$PARTED_TABLE" ;;
esac
echo "$*" >> "$PARTED_LOG"
exit "${PARTED_RC:-1}"
STUB
chmod +x "$WORK/parted"

cat > "$WORK/blkid" <<'STUB'
#!/bin/sh
[ -n "$BLKID_OUT" ] && echo "$BLKID_OUT"
exit 0
STUB
chmod +x "$WORK/blkid"

# The table of the board in use: a 32 GB card whose data partition is present.
cat > "$WORK/full.table" <<'T'
BYT;
/dev/mmcblk0:60506112s:sd/mmc:512:512:msdos:SD SA32G:;
1:1s:32768s:32768s:fat16::boot, lba;
2:40960s:4235263s:4194304s:ext4::;
3:4235264s:8429567s:4194304s:::;
4:8429568s:60506111s:52076544s:::;
5:8437760s:10534911s:2097152s:ext4::;
6:10543104s:60506111s:49963008s:::;
T

# A freshly flashed card: no data partition, container stops with the recovery
# slot, and the card is bigger than the image.
cat > "$WORK/fresh.table" <<'T'
BYT;
/dev/mmcblk0:15523840s:sd/mmc:512:512:msdos:SD SA08G:;
1:1s:32768s:32768s:fat16::boot, lba;
2:40960s:4235263s:4194304s:ext4::;
3:4235264s:8429567s:4194304s:::;
4:8429568s:10534911s:2105344s:::;
5:8437760s:10534911s:2097152s:ext4::;
T

geo() {
    PARTED_TABLE=$1; PARTED_LOG=$WORK/plog; shift
    export PARTED_TABLE PARTED_LOG
    sh -c "PARTED=$WORK/parted; . $WORK/geo.sh; $*" 2>/dev/null
}

echo "===== the table is read, not guessed ====="

got=$(geo "$WORK/full.table" 'disk_sectors')
[ "$got" = 60506112 ] && note "disk_sectors reads the disk line" OK \
                      || note "disk_sectors gave '$got', want 60506112" FAIL

got=$(geo "$WORK/full.table" 'part_end_sector 4')
[ "$got" = 60506111 ] && note "part_end_sector reads the container" OK \
                      || note "part_end_sector 4 gave '$got', want 60506111" FAIL

got=$(geo "$WORK/full.table" 'part_size_sectors 6')
[ "$got" = 49963008 ] && note "part_size_sectors reads the data partition" OK \
                      || note "part_size_sectors 6 gave '$got', want 49963008" FAIL

got=$(geo "$WORK/fresh.table" 'part_end_sector 6')
[ -z "$got" ] && note "a partition that is not there reads as nothing" OK \
              || note "part_end_sector 6 gave '$got' on a fresh card" FAIL

echo
echo "===== parted's exit status is never trusted ====="

# The stub always exits 1, exactly as the real one does on a busy disk. Every
# reader above ran against it. If any of them tested the exit status, they
# would have returned nothing.
got=$(geo "$WORK/full.table" 'disk_sectors')
[ "$got" = 60506112 ] \
    && note "a parted that exits 1 is still read" OK \
    || note "a parted that exits 1 was treated as a failure" FAIL

echo
echo "===== a filesystem is recognised by output, not by TYPE or by status ====="

hasfs() {
    BLKID_OUT=$1
    export BLKID_OUT
    sh -c "BLKID=$WORK/blkid; . $WORK/geo.sh; has_filesystem /dev/mmcblk0p6 && echo yes || echo no" 2>/dev/null
}

# This is the exact output busybox blkid gives for the data partition in use.
# There is no TYPE field. An implementation that greps for one formats a live
# /data and destroys the board's identity.
got=$(hasfs '/dev/mmcblk0p6: LABEL="data" UUID="EE8B-6CB5"')
[ "$got" = yes ] \
    && note "busybox output with no TYPE counts as a filesystem" OK \
    || note "busybox output with no TYPE was read as an empty partition" FAIL

got=$(hasfs '/dev/mmcblk0p6: LABEL="data" UUID="6ACE-EE79" TYPE="exfat"')
[ "$got" = yes ] \
    && note "util-linux output also counts as a filesystem" OK \
    || note "util-linux output was read as an empty partition" FAIL

# An empty partition. blkid prints nothing and still exits 0, so the exit status
# says nothing at all.
got=$(hasfs '')
[ "$got" = no ] \
    && note "no output means no filesystem" OK \
    || note "an empty partition was read as holding a filesystem" FAIL

echo
echo "===== the kernel and the table must agree before anything is formatted ====="

# The device node is a real file in the scratch directory, not /dev/mmcblk0p6.
# A workstation has no such node, so the -e test inside data_partition_ready
# would fail on every call and BOTH cases below would pass without the function
# doing anything. The name still ends in p6, which is what the function reads
# the partition number from.
ready() {
    cat > "$WORK/partitions" <<PART
major minor  #blocks  name

 179        0   30253056 mmcblk0
 179        6   $1 mmcblk0p6
PART
    : > "$WORK/mmcblk0p6"
    PARTED_TABLE=$WORK/full.table; PARTED_LOG=$WORK/plog
    export PARTED_TABLE PARTED_LOG
    sh -c "PARTED=$WORK/parted; PARTITIONS=$WORK/partitions; . $WORK/geo.sh; \
           data_partition_ready $WORK/mmcblk0p6 && echo yes || echo no" 2>/dev/null
}

# /proc/partitions counts 1024 byte blocks, so 24981504 blocks is 49963008
# sectors, which is what the table says.
got=$(ready 24981504)
[ "$got" = yes ] && note "a node that matches the table is ready" OK \
                 || note "a node that matches the table was rejected" FAIL

# A stale node. Formatting at this size makes a filesystem that cannot be
# repaired, because resize.exfat does not exist on this device.
got=$(ready 1024)
[ "$got" = no ] && note "a node that disagrees with the table is refused" OK \
                || note "a stale node was accepted, which would format the wrong size" FAIL

echo
echo "===== the start sector comes from the slot configuration ====="

cat > "$WORK/slots.conf" <<'CONF'
SLOT_A=/dev/mmcblk0p2
SLOT_B=/dev/mmcblk0p3
RECOVERY=/dev/mmcblk0p5
DATA_DEV=/dev/mmcblk0p6
DATA_START=10543104
CONF

got=$( SLOT_CONF="$WORK/slots.conf" sh -c ". $WORK/geo.sh; data_start" )
[ "$got" = 10543104 ] \
    && note "a conf with DATA_START is honoured" OK \
    || note "a conf with DATA_START gave '$got'" FAIL

# An image from before this change, or a board that was never migrated. It must
# print nothing, so that the caller makes no partition at all. Partitioning a
# board that did not declare where its data goes is worse than not partitioning.
printf 'DATA_DEV=/dev/mmcblk0p6\n' > "$WORK/old.conf"
got=$( SLOT_CONF="$WORK/old.conf" sh -c ". $WORK/geo.sh; data_start" )
[ -z "$got" ] \
    && note "a conf without DATA_START gives nothing" OK \
    || note "a conf without DATA_START gave '$got'" FAIL

got=$( SLOT_CONF="$WORK/absent.conf" sh -c ". $WORK/geo.sh; data_start" )
[ -z "$got" ] \
    && note "no conf at all gives nothing" OK \
    || note "no conf gave '$got'" FAIL

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
