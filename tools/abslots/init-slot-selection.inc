# --- slot selection ---
# Runs while /boot is still mounted, because the dispatch below no longer has
# it. Three markers are read here and one is deleted here.
#
#   /boot/slot        the trusted slot: "a", "b", or absent meaning "a"
#   /boot/slot.try    a slot on trial, written by the operator
#   /boot/recovery    if present, boot recovery and nothing else
#
# The delete is the whole design. A slot on trial is disarmed BEFORE it is
# handed control, so a hang, a panic, an oops, a power cut and a slot that
# starts no listener all reach the same next boot, which is the trusted slot.
# The suspect never votes on its own trial.
#
# This replaced a guard that lived in the userspace of the slot under test. On
# 2026-08-15 that guard was not installed in the failing slot, and a boot that
# stops before S00 could not have reached it in any case.
BOOT=${BOOT:-/boot}

# PATH is "/" in the initramfs and only twelve applets are symlinked, so
# anything outside that set is reached through busybox by name. It is a
# variable purely so test-dispatch.sh can point it at a stub: on the device
# nothing sets it and it is /busybox.
BUSYBOX=${BUSYBOX:-/busybox}

# Defined here rather than in the dispatch block, because this block runs
# first and calls it. It was in the dispatch block once, and the effect was a
# selection that logged nothing at all: the function did not exist yet, so the
# "trial discarded" line that explains an undeletable marker would have gone
# missing on exactly the boot that needed explaining.
#
# The console trace from set -x goes to the serial port, which a deployed board
# does not have. These land in dmesg, so a board that fell back explains itself
# over the network with `dmesg | grep nanokvm-slot`.
# KMSG is a variable so the tests can read what was logged. On the device
# nothing sets it and it is /dev/kmsg.
KMSG=${KMSG:-/dev/kmsg}

slotlog() {
	echo "nanokvm-slot: $*" >> "$KMSG"
}

trusted=a
if [ -e "$BOOT/slot" ]
then
	read trusted < "$BOOT/slot"
fi
if [ -z "${trusted}" ]
then
	trusted=a
fi

trial=
if [ -e "$BOOT/slot.try" ]
then
	read trial < "$BOOT/slot.try"

	"$BUSYBOX" rm -f "$BOOT/slot.try"
	"$BUSYBOX" sync

	# Deleting on FAT is a directory entry write. If it did not land and the
	# trial hangs on every boot, the board loops for ever, which is worse than
	# the dead board this exists to prevent. Proving the disarm before the trial
	# is honoured turns that loop into one wasted boot.
	if [ -e "$BOOT/slot.try" ]
	then
		slotlog "trial ${trial} discarded: slot.try could not be deleted"
		trial=
	fi
fi

recovery_requested=
if [ -e "$BOOT/recovery" ]
then
	recovery_requested=1
fi

slotlog "trusted=${trusted} trial=${trial:-none} recovery=${recovery_requested:-no}"
# --- end slot selection ---
