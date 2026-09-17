#!/bin/sh
# Check that S01fs never claims the stock data disk before it is formatted.
#
#   test-s01fs-disk0.sh [path-to-S01fs]
#
# This covers the /boot/usb.disk0 branch, which is upstream's own first-boot
# provisioning for a stock partition layout. It is not the A/B path: on an
# IronKVM card the DATA_START in the image's deviceinfo disarms it, and the
# image manifest creates /etc/kvm.disk0 as well, so the branch is dead there
# twice over. It
# is live for a stock-layout board that installs this firmware over the air,
# because the update package carries system/init.d/S01fs and carries neither
# of those two files.
#
# The fault it guards against: the branch used to create /etc/kvm.disk0 first
# and then run mkfs.exfat in the background. The marker therefore existed
# while the filesystem did not. S03usbdev runs later in the same boot and
# hands the mass-storage LUN whatever /boot/usb.disk0 names, so a board that
# lost power during the format came back, read the marker, decided the disk
# was ready and exported a half-made filesystem to the host. Nothing retried
# it, because the marker said the work was done.
#
# So the marker is written last, a separate marker records that a format is in
# flight, and the format runs in the foreground. mkfs.exfat writes metadata
# only, so the cost to boot is small and the alternative is handing a host a
# corrupt disk.
#
# Two more things are asserted here, from upstream pull request #764: the
# partition is made with the parted ntfs type so the MBR id is 0x07 and a
# desktop mounts the card, and the filesystem gets a label. The A/B path in
# this same script already does both.
set -u

S01=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S01fs}
[ -f "$S01" ] || { echo "usage: test-s01fs-disk0.sh <S01fs>"; exit 1; }

fails=0
note() { printf '  %-66s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

echo "===== the stock data disk is claimed only after it is formatted ====="

sed -n '/^# --- disk0 provisioning ---$/,/^# --- end disk0 provisioning ---$/p' "$S01" > "$WORK/f.sh"
if [ ! -s "$WORK/f.sh" ]; then
    note "S01fs carries a disk0 provisioning block" FAIL
    echo
    echo "$fails case(s) FAILED"
    exit 1
fi
note "S01fs carries a disk0 provisioning block" OK

# The branch now proves the device is not something else before it writes to it,
# and both proofs come from blocks above it: the description says whether the
# device is a slot, and blkid says whether it already holds a filesystem. So the
# blocks are sourced in the order the script has them, exactly as a boot does.
sed -n '/^# --- data device ---/,/^# --- end data device ---/p' "$S01" > "$WORK/dd.sh"
sed -n '/^# --- data geometry ---/,/^# --- end data geometry ---/p' "$S01" > "$WORK/geo.sh"
[ -s "$WORK/dd.sh" ] && [ -s "$WORK/geo.sh" ] \
    && note "the data device and geometry blocks can be extracted" OK \
    || note "the data device and geometry blocks can be extracted" FAIL

# The real reader, not a stub. A stub would not catch a key the guard asks for
# under the wrong name.
READER=${READER:-$(cd "$(dirname "$0")/../.." && pwd)/kvmapp/system/ironkvm-deviceinfo}
[ -f "$READER" ] || { echo "needs kvmapp/system/ironkvm-deviceinfo"; exit 2; }

# Stubs. parted logs its argv and creates the partition node, so the function
# sees the outcome rather than a return code: parted exits 1 on a busy disk
# after it has succeeded, which is measured behaviour on this board.
mkdir -p "$WORK/bin"
cat > "$WORK/bin/parted" <<'EOF'
#!/bin/sh
echo "parted $*" >> "$PLOG"
[ -n "${PARTED_MAKES_DEV:-}" ] && : > "$PARTED_MAKES_DEV"
exit "${PARTED_RC:-1}"
EOF
cat > "$WORK/bin/mkfs.exfat" <<'EOF'
#!/bin/sh
echo "mkfs.exfat $*" >> "$PLOG"
exit "${MKFS_RC:-0}"
EOF
cat > "$WORK/bin/ironkvm-deviceinfo" <<STUB
#!/bin/sh
DEVICEINFO_PATHS="$WORK/deviceinfo" exec sh "$READER" "\$@"
STUB
# One file per device name, so a case says what blkid answers for the device it
# is about. A device no case described reads as empty, which is what blkid
# prints for a partition with no filesystem.
cat > "$WORK/bin/blkid" <<STUB
#!/bin/sh
f="$WORK/fs/\$(basename "\$1")"
[ -f "\$f" ] && cat "\$f"
exit 0
STUB
chmod +x "$WORK/bin/parted" "$WORK/bin/mkfs.exfat" \
    "$WORK/bin/ironkvm-deviceinfo" "$WORK/bin/blkid"
PATH="$WORK/bin:$PATH"
export PATH

# What blkid answers for the device this case is about, and what the in-flight
# marker holds.
fsout()   { printf '%s\n' "$1" > "$WORK/fs/$(basename "$dev")"; }
pends()   { printf '%s\n' "$1" > "$WORK/root/etc/kvm.disk0.formatting"; }

# One run of the function against a fresh fake root.
# $1 = the case; $2 = mkfs exit code
#
# "slot" and "slot-pending" give the device a description that calls it a root
# slot. Their disk is a path under the scratch directory, so nothing in /dev is
# named even if the guard they test is broken.
#
# The "resume" cases are a board that lost power during mkfs.exfat. The device
# holds the half-made filesystem and the marker names it, which is the one state
# that may be formatted again. Two spellings of the same state, because the two
# kinds of board answer blkid differently: util-linux prints a TYPE field and
# busybox never does.
run() {
    rm -rf "$WORK/root"; mkdir -p "$WORK/root/etc"
    : > "$WORK/root/plog"
    rm -rf "$WORK/fs"; mkdir -p "$WORK/fs"
    printf '%s\n' DATA_FS=exfat > "$WORK/deviceinfo"
    dev="$WORK/root/mmcblk0p3"

    case "$1" in
        marked)  : > "$WORK/root/etc/kvm.disk0"; : > "$dev" ;;
        pending) : > "$WORK/root/etc/kvm.disk0.formatting"; : > "$dev" ;;
        nodev)   PARTED_MAKES=""; ;;
        slot)    dev="$WORK/root/fake0p3"
                 printf '%s\n' "DISK=$WORK/root/fake0" BOOT_PART=1 \
                     SLOT_A=2 SLOT_B=3 RECOVERY=5 DATA_FS=exfat > "$WORK/deviceinfo" ;;
        hasfs)   : > "$dev"
                 fsout '/dev/mmcblk0p3: LABEL="data" UUID="EE8B-6CB5" TYPE="exfat"' ;;
        resume)  : > "$dev"
                 fsout '/dev/mmcblk0p3: LABEL="data" UUID="EE8B-6CB5" TYPE="exfat"'
                 pends "$dev" ;;
        resume-bare)
                 # busybox blkid, and the label not written yet.
                 : > "$dev"
                 fsout '/dev/mmcblk0p3: UUID="EE8B-6CB5"'
                 pends "$dev" ;;
        resume-ext4)
                 : > "$dev"
                 fsout '/dev/mmcblk0p3: LABEL="slot-b" UUID="accd0050" TYPE="ext4"'
                 pends "$dev" ;;
        resume-ext4-busybox)
                 : > "$dev"
                 fsout '/dev/mmcblk0p3: LABEL="slot-b" UUID="accd0050"'
                 pends "$dev" ;;
        stale)   : > "$dev"
                 fsout '/dev/mmcblk0p3: LABEL="data" UUID="EE8B-6CB5" TYPE="exfat"'
                 pends /dev/some-other-device ;;
        legacy)  # The empty marker the earlier code wrote with touch.
                 : > "$dev"
                 fsout '/dev/mmcblk0p3: LABEL="data" UUID="EE8B-6CB5" TYPE="exfat"'
                 : > "$WORK/root/etc/kvm.disk0.formatting" ;;
        slot-pending)
                 dev="$WORK/root/fake0p3"
                 printf '%s\n' "DISK=$WORK/root/fake0" BOOT_PART=1 \
                     SLOT_A=2 SLOT_B=3 RECOVERY=5 DATA_FS=exfat > "$WORK/deviceinfo"
                 : > "$dev"
                 fsout '/dev/fake0p3: LABEL="data" UUID="EE8B-6CB5" TYPE="exfat"'
                 pends "$dev" ;;
    esac

    (
        PLOG="$WORK/root/plog"; export PLOG
        MKFS_RC=${2:-0}; export MKFS_RC
        [ "$1" = nodev ] || { PARTED_MAKES_DEV="$dev"; export PARTED_MAKES_DEV; }
        DISK0_DEV="$dev"
        DISK0_MARKER="$WORK/root/etc/kvm.disk0"
        DISK0_PENDING="$WORK/root/etc/kvm.disk0.formatting"
        DISK0_SETTLE=0
        export DISK0_DEV DISK0_MARKER DISK0_PENDING DISK0_SETTLE
        DEVINFO="$WORK/bin/ironkvm-deviceinfo"
        BLKID="$WORK/bin/blkid"
        . "$WORK/dd.sh"
        . "$WORK/geo.sh"
        . "$WORK/f.sh"
        provision_disk0 > "$WORK/root/out" 2>&1
        echo "rc=$?"
    )
}

log()      { cat "$WORK/root/plog"; }
ran()      { grep -q "$1" "$WORK/root/plog"; }
said()     { grep -q "$1" "$WORK/root/out"; }
marked()   { [ -e "$WORK/root/etc/kvm.disk0" ]; }
pending()  { [ -e "$WORK/root/etc/kvm.disk0.formatting" ]; }

# --- the boot after provisioning: nothing happens at all -------------------
rc=$(run marked)
[ "$rc" = "rc=0" ] && note "a provisioned board reports success" OK \
                   || note "a provisioned board reports success ($rc)" FAIL
ran parted || ran mkfs.exfat && note "a provisioned board does not touch the disk" FAIL \
                             || note "a provisioned board does not touch the disk" OK

# --- the first boot: partition, format, then claim -------------------------
rc=$(run fresh)
[ "$rc" = "rc=0" ] && note "a fresh board provisions successfully" OK \
                   || note "a fresh board provisions successfully ($rc)" FAIL
ran 'parted'      && note "it makes the partition" OK      || note "it makes the partition" FAIL
ran 'mkfs.exfat'  && note "it makes the filesystem" OK     || note "it makes the filesystem" FAIL
marked            && note "and only then writes the marker" OK || note "and only then writes the marker" FAIL
pending           && note "the in-flight marker is cleared" FAIL || note "the in-flight marker is cleared" OK

order=$(grep -oE 'parted|mkfs.exfat' "$WORK/root/plog" | tr '\n' ' ')
[ "$order" = "parted mkfs.exfat " ] && note "the partition is made before the filesystem" OK \
                                    || note "the partition is made before the filesystem ('$order')" FAIL

log | grep -q 'mkpart primary ntfs' && note "the partition gets the ntfs type, so MBR id 0x07" OK \
                                    || note "the partition gets the ntfs type, so MBR id 0x07" FAIL
log | grep -q 'mkfs.exfat -L data'  && note "the filesystem gets a label" OK \
                                    || note "the filesystem gets a label" FAIL

# --- the format fails: claim nothing, and leave the retry armed ------------
rc=$(run fresh 1)
[ "$rc" = "rc=1" ] && note "a failed format reports failure" OK \
                   || note "a failed format reports failure ($rc)" FAIL
marked  && note "a failed format does not write the marker" FAIL \
        || note "a failed format does not write the marker" OK
pending && note "a failed format leaves the retry armed" OK \
        || note "a failed format leaves the retry armed" FAIL

# --- power lost mid-format: retry the format, do not re-partition ----------
rc=$(run pending)
[ "$rc" = "rc=0" ] && note "an interrupted format completes on the next boot" OK \
                   || note "an interrupted format completes on the next boot ($rc)" FAIL
ran 'parted' && note "and it does not partition a second time" FAIL \
             || note "and it does not partition a second time" OK
ran 'mkfs.exfat' && note "and it does format" OK || note "and it does format" FAIL
marked && note "and it claims the disk afterwards" OK || note "and it claims the disk afterwards" FAIL

# --- the partition never appears: stop, claim nothing ----------------------
rc=$(run nodev)
[ "$rc" = "rc=1" ] && note "a missing partition reports failure" OK \
                   || note "a missing partition reports failure ($rc)" FAIL
ran 'mkfs.exfat' && note "a missing partition is never formatted" FAIL \
                 || note "a missing partition is never formatted" OK
marked && note "a missing partition is never claimed" FAIL \
       || note "a missing partition is never claimed" OK

# --- the device is a slot: refuse, and say so -------------------------------
# On an A/B card p3 is a root filesystem, and mkfs.exfat there formats the
# system the board is running from. Three conditions used to keep this branch
# away from such a card: may_autopartition refuses a description that declares
# DATA_START, /boot/usb.disk0 has to exist, and /etc/kvm.disk0 has to be absent.
# A card laid out before the build wrote DATA_START satisfies the first, so on
# that card one missing marker file was the whole defence. This is the positive
# guard: the description names the device as a slot, so the branch stops.
rc=$(run slot)
[ "$rc" = "rc=1" ] && note "a device the description calls a slot is refused" OK \
                   || note "a device the description calls a slot is refused ($rc)" FAIL
ran 'parted' && note "and the card is not partitioned" FAIL \
             || note "and the card is not partitioned" OK
ran 'mkfs.exfat' && note "AND A ROOT SLOT IS NOT FORMATTED" FAIL \
                 || note "and a root slot is not formatted" OK
said 'slot or recovery' && note "and it says so on the console" OK \
                        || note "and it says so on the console" FAIL
marked  && note "a refused device is never claimed" FAIL \
        || note "a refused device is never claimed" OK
pending && note "and no format is recorded as in flight" FAIL \
        || note "and no format is recorded as in flight" OK

# --- the device already holds a filesystem: refuse --------------------------
# Where the description declares no layout there is nothing to ask, so the
# second guard asks the card. A partition blkid can name is never formatted,
# which is the same test the A/B path uses to decide a card is provisioned.
rc=$(run hasfs)
[ "$rc" = "rc=1" ] && note "a device that already holds a filesystem is refused" OK \
                   || note "a device that already holds a filesystem is refused ($rc)" FAIL
ran 'mkfs.exfat' && note "AND AN EXISTING FILESYSTEM IS NOT FORMATTED" FAIL \
                 || note "and an existing filesystem is not formatted" OK
said 'already holds a filesystem' && note "and it says so on the console" OK \
                                  || note "and it says so on the console" FAIL
marked && note "and the disk is not claimed on the strength of it" FAIL \
       || note "and the disk is not claimed on the strength of it" OK

# --- the interrupted format still completes ---------------------------------
# The refusal above has one exception, and it is the state this whole block was
# restructured for. A board that lost power during mkfs.exfat comes back with
# the half-made filesystem on the device, so a refusal with no exception would
# wedge it for ever: no /etc/kvm.disk0, no disk for the host, and no message
# anybody would go looking for. The marker names the device, so the script knows
# it started that format itself.
rc=$(run resume)
[ "$rc" = "rc=0" ] && note "an interrupted format of the data filesystem retries" OK \
                   || note "an interrupted format of the data filesystem retries ($rc)" FAIL
ran 'mkfs.exfat' && note "and the filesystem is made" OK \
                 || note "and the filesystem is made" FAIL
ran 'parted' && note "and the card is not partitioned a second time" FAIL \
             || note "and the card is not partitioned a second time" OK
marked && note "and the disk is claimed afterwards" OK \
       || note "and the disk is claimed afterwards" FAIL

# busybox blkid prints no TYPE field for any filesystem, so on that board a
# half-made exfat reads as a UUID and nothing else. It must still retry, or the
# exception never fires on the boards this branch actually serves.
rc=$(run resume-bare)
[ "$rc" = "rc=0" ] && note "a half-made filesystem with no TYPE field retries" OK \
                   || note "a half-made filesystem with no TYPE field retries ($rc)" FAIL
ran 'mkfs.exfat' && note "and that filesystem is made too" OK \
                 || note "and that filesystem is made too" FAIL

# --- but only that state ----------------------------------------------------
# A root slot is ext4. The marker cannot excuse formatting one.
rc=$(run resume-ext4)
[ "$rc" = "rc=1" ] && note "another filesystem is refused despite the marker" OK \
                   || note "another filesystem is refused despite the marker ($rc)" FAIL
ran 'mkfs.exfat' && note "AND A ROOT SLOT IS NOT FORMATTED BY THE RETRY" FAIL \
                 || note "and a root slot is not formatted by the retry" OK
said 'is not exfat' && note "and it names the filesystem it found" OK \
                    || note "and it names the filesystem it found" FAIL

# The same card read by busybox blkid, which prints no TYPE. The label is what
# tells a root slot from the data filesystem there.
rc=$(run resume-ext4-busybox)
[ "$rc" = "rc=1" ] && note "a slot with no TYPE field is refused by its label" OK \
                   || note "a slot with no TYPE field is refused by its label ($rc)" FAIL
ran 'mkfs.exfat' && note "AND THAT ROOT SLOT IS NOT FORMATTED EITHER" FAIL \
                 || note "and that root slot is not formatted either" OK

# A marker left by another device excuses nothing. This is why the marker
# records the device instead of only existing.
rc=$(run stale)
[ "$rc" = "rc=1" ] && note "a marker naming another device excuses no format" OK \
                   || note "a marker naming another device excuses no format ($rc)" FAIL
ran 'mkfs.exfat' && note "and nothing is formatted for it" FAIL \
                 || note "and nothing is formatted for it" OK

# The empty marker the earlier code wrote with touch names no device, so it
# cannot say which device the interrupted format was on.
rc=$(run legacy)
[ "$rc" = "rc=1" ] && note "an empty marker from the earlier code excuses nothing" OK \
                   || note "an empty marker from the earlier code excuses nothing ($rc)" FAIL
ran 'mkfs.exfat' && note "and nothing is formatted for it either" FAIL \
                 || note "and nothing is formatted for it either" OK

# And the description outranks the marker. A device it calls a slot is refused
# whatever state the format was left in.
rc=$(run slot-pending)
[ "$rc" = "rc=1" ] && note "a slot is refused whatever the marker says" OK \
                   || note "a slot is refused whatever the marker says ($rc)" FAIL
said 'slot or recovery' && note "and it is refused for being a slot" OK \
                        || note "and it is refused for being a slot" FAIL
ran 'mkfs.exfat' && note "AND A DECLARED SLOT IS NOT FORMATTED BY THE RETRY" FAIL \
                 || note "and a declared slot is not formatted by the retry" OK

# The marker has to carry the device for any of the above to work. A format that
# failed leaves it behind, which is where its content can be read.
rc=$(run fresh 1)
[ "$(cat "$WORK/root/etc/kvm.disk0.formatting")" = "$WORK/root/mmcblk0p3" ] \
    && note "the in-flight marker records the device" OK \
    || note "the in-flight marker holds '$(cat "$WORK/root/etc/kvm.disk0.formatting")'" FAIL

# --- the shipped text itself ------------------------------------------------
# The background format is the defect. Reading the file is the only way to
# assert it stays gone, because a backgrounded mkfs still logs the same line.
# Comment lines are stripped first. The block above quotes the old code in
# prose to explain what was wrong with it, and an assertion that cannot tell
# code from a comment would fail on the explanation.
code=$WORK/code.sh
grep -v '^[[:space:]]*#' "$S01" > "$code"

if grep -E 'mkfs\.exfat[^&]*\) *&|\( *[^)]*mkfs\.exfat[^)]*\) *&' "$code" > /dev/null 2>&1; then
    note "no mkfs.exfat runs in the background" FAIL
else
    note "no mkfs.exfat runs in the background" OK
fi

if grep -q 'touch /etc/kvm.disk0' "$code"; then
    note "the marker is not written by a hardcoded touch in the branch" FAIL
else
    note "the marker is not written by a hardcoded touch in the branch" OK
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
