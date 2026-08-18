#!/bin/sh
# Check that S01fs never claims the stock data disk before it is formatted.
#
#   test-s01fs-disk0.sh [path-to-S01fs]
#
# This covers the /boot/usb.disk0 branch, which is upstream's own first-boot
# provisioning for a stock partition layout. It is not the A/B path: on an
# IronKVM card /etc/nanokvm-slots.conf disarms it, and the image manifest
# creates /etc/kvm.disk0 as well, so the branch is dead there twice over. It
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

S01=${1:-$(dirname "$0")/../../../kvmapp/system/init.d/S01fs}
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
    exit "$fails"
fi
note "S01fs carries a disk0 provisioning block" OK

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
chmod +x "$WORK/bin/parted" "$WORK/bin/mkfs.exfat"
PATH="$WORK/bin:$PATH"
export PATH

# One run of the function against a fresh fake root.
# $1 = "fresh" | "marked" | "pending" | "nodev"; $2 = mkfs exit code
run() {
    rm -rf "$WORK/root"; mkdir -p "$WORK/root/etc"
    : > "$WORK/root/plog"
    dev="$WORK/root/mmcblk0p3"

    case "$1" in
        marked)  : > "$WORK/root/etc/kvm.disk0"; : > "$dev" ;;
        pending) : > "$WORK/root/etc/kvm.disk0.formatting"; : > "$dev" ;;
        nodev)   PARTED_MAKES=""; ;;
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
        . "$WORK/f.sh"
        provision_disk0 > /dev/null 2>&1
        echo "rc=$?"
    )
}

log()      { cat "$WORK/root/plog"; }
ran()      { grep -q "$1" "$WORK/root/plog"; }
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
fi
exit "$fails"
