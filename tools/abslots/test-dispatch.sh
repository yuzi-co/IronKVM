#!/bin/sh
# Check the initramfs slot dispatch, in a sandbox.
#
#   test-dispatch.sh [selection.inc] [dispatch.inc]
#
# The dispatch decides which filesystem the board runs before anything can log,
# so every branch is checked here rather than on hardware.
#
# The case that matters most is the disarm. /boot/slot.try must be deleted
# BEFORE the trial is handed control, because a slot that hangs cannot be relied
# on to disarm itself. That is the whole reason this design exists: the previous
# guard lived in the userspace of the slot under test, was not installed in the
# slot that failed, and could not have run anyway because the boot stopped
# before S00.
SEL=${1:-$(dirname "$0")/init-slot-selection.inc}
DIS=${2:-$(dirname "$0")/init-mount-dispatch.inc}
[ -f "$SEL" ] && [ -f "$DIS" ] || { echo "usage: test-dispatch.sh <selection.inc> <dispatch.inc>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

# The initramfs reaches applets outside its twelve symlinks through busybox by
# name. There is no /busybox here, so a stub stands in. Without it the delete
# fails, slot.try survives, and the disarm-proof check correctly discards every
# trial, which looks exactly like a dispatch bug and is not one.
#
# STUB_RM_FAILS makes the delete a no-op that still reports success, which is
# what a FAT directory entry write that does not land looks like from here.
# Making the directory read-only was tried first and does not work: under
# Windows ACLs a read-only directory still allows the delete, so the case
# silently tested nothing.
cat > "$WORK/busybox" <<'BB'
#!/bin/sh
applet=$1
shift
case "$applet" in
    rm)   [ -n "$STUB_RM_FAILS" ] && exit 0; rm "$@" ;;
    sync) sync ;;
    *)    echo "stub busybox: unsupported applet $applet" >&2; exit 127 ;;
esac
BB
chmod 755 "$WORK/busybox"

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

A=/dev/mmcblk0p2
B=/dev/mmcblk0p3
R=/dev/mmcblk0p5

# The stubs stand in for everything the initramfs would touch. mount succeeds
# only for a device named in MOUNTABLE, which is how a slot is made to fail.
harness() {
    cat <<'STUB'
mount() {
    for d in $MOUNTABLE; do [ "$d" = "$3" ] && return 0; done
    return 1
}
umount() { return 0; }
e2fsck() { return 0; }
msc() { echo "DEVICE=msc"; exit 0; }
STUB
}

# run_case <slot> <try> <recovery-marker> <mountable-devices>
run_case() {
    rm -rf "$WORK/boot"
    mkdir -p "$WORK/boot"
    [ -n "$1" ] && printf '%s\n' "$1" > "$WORK/boot/slot"
    [ -n "$2" ] && printf '%s\n' "$2" > "$WORK/boot/slot.try"
    [ -n "$3" ] && : > "$WORK/boot/recovery"
    : > "$WORK/log"
    (
        BOOT="$WORK/boot"
        LOGFILE="$WORK/log"
        MOUNTABLE="$4"
        BUSYBOX="$WORK/busybox"
        KMSG="$WORK/kmsg"
        export BOOT LOGFILE MOUNTABLE BUSYBOX KMSG
        eval "$(harness)"
        . "$SEL"
        . "$DIS"
        echo "DEVICE=${bootdev}"
    ) 2>/dev/null
}

expect() {
    got=$(run_case "$2" "$3" "$4" "$5")
    if echo "$got" | grep -q "DEVICE=$6"; then
        note "$1" OK
    else
        note "$1 (got: $(echo "$got" | tr '\n' ' '))" FAIL
    fi
}

echo "===== the dispatch chooses the right root ====="

expect "no markers boots slot A"                       ""    ""  "" "$A $B $R" "$A"
expect "slot=b boots slot B"                           "b"   ""  "" "$A $B $R" "$B"
expect "a trial overrides the trusted slot"            "b"   "a" "" "$A $B $R" "$A"
expect "a trial that will not mount falls to trusted"  "b"   "a" "" "$B $R"    "$B"
expect "trusted failing too falls to recovery"         "b"   "a" "" "$R"       "$R"
expect "everything failing reaches msc"                "b"   ""  "" ""         "msc"
expect "an unreadable marker boots slot A"             "wat" ""  "" "$A $B $R" "$A"
expect "the recovery marker wins over everything"      "b"   "a" "1" "$A $B $R" "$R"

echo
echo "===== the trial is disarmed before it is handed control ====="

run_case "b" "a" "" "$A $B $R" >/dev/null
[ -e "$WORK/boot/slot.try" ] \
    && note "slot.try is deleted before handover" FAIL \
    || note "slot.try is deleted before handover" OK

# A trial that failed to mount must still be disarmed. Otherwise a slot that
# mounts intermittently gets tried for ever.
run_case "b" "a" "" "$B $R" >/dev/null
[ -e "$WORK/boot/slot.try" ] \
    && note "a failed trial is disarmed too" FAIL \
    || note "a failed trial is disarmed too" OK

echo
echo "===== a trial that cannot be disarmed is discarded ====="

# Deleting on FAT is a directory entry write. If it does not land and the trial
# hangs on every boot, the board loops for ever, which is worse than the dead
# board this design prevents.
rm -rf "$WORK/boot"; mkdir -p "$WORK/boot"
echo b > "$WORK/boot/slot"
echo a > "$WORK/boot/slot.try"
: > "$WORK/log"
got=$( (
    BOOT="$WORK/boot"; LOGFILE="$WORK/log"; MOUNTABLE="$A $B $R"; BUSYBOX="$WORK/busybox"; KMSG="$WORK/kmsg"
    STUB_RM_FAILS=1
    export BOOT LOGFILE MOUNTABLE BUSYBOX KMSG STUB_RM_FAILS
    eval "$(harness)"
    . "$SEL"; . "$DIS"
    echo "DEVICE=${bootdev}"
) 2>/dev/null )

echo "$got" | grep -q "DEVICE=$B" \
    && note "an undeletable trial marker is ignored, trusted boots" OK \
    || note "an undeletable trial marker is ignored, trusted boots (got: $got)" FAIL

grep -q "could not be deleted" "$WORK/kmsg" \
    && note "and it says why" OK \
    || note "and it says why" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
