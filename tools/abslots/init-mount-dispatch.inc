# --- slot dispatch ---
# Device names, fixed here rather than computed anywhere. In this layout p3 is
# a root filesystem, and S01fs used to run mkfs.exfat on p3, so a path that
# derives a device from a partition number is the shape of that bug.
SLOT_A=/dev/mmcblk0p2
SLOT_B=/dev/mmcblk0p3
RECOVERY=/dev/mmcblk0p5

# One mount attempt, with the single e2fsck retry this script has always had.
mount_fs() {
	mount -o rw "$1" /realroot && return 0
	slotlog "mount $1 failed rc=$?, running e2fsck"
	e2fsck -fvp "$1"
	slotlog "e2fsck $1 rc=$?"
	mount -o rw "$1" /realroot
}

slot_device() {
	case "$1" in
		a) echo "${SLOT_A}" ;;
		b) echo "${SLOT_B}" ;;
		recovery) echo "${RECOVERY}" ;;
		/dev/*) echo "$1" ;;
		*) echo "" ;;
	esac
}

# The ladder. Each rung is tried only if every rung above it failed, and the
# bottom two are unconditional.
#
# Note what this ladder does not cover: it only catches a root that will not
# mount. A slot that mounts and then starts no listener is the 2026-08-15
# failure, and it is handled by the disarm in the selection block above, not
# here. Being honest about that boundary is why the rungs are labelled.
bootdev=

# Rung 1: an explicit recovery request beats everything, including a trial.
if [ -n "${recovery_requested}" ]
then
	slotlog "recovery requested"
	mount_fs "${RECOVERY}" && bootdev="${RECOVERY}"
fi

# Rung 2: the trial, if there is one and it was disarmed.
if [ -z "${bootdev}" ] && [ -n "${trial}" ]
then
	dev=$(slot_device "${trial}")
	if [ -n "${dev}" ]
	then
		mount_fs "${dev}" && bootdev="${dev}"
		[ -z "${bootdev}" ] && slotlog "trial ${trial} did not mount"
	else
		slotlog "trial ${trial} is not a slot name"
	fi
fi

# Rung 3: the trusted slot. An unrecognised marker lands on slot A, which is
# what an absent marker means, so a corrupted marker cannot strand the board.
if [ -z "${bootdev}" ]
then
	dev=$(slot_device "${trusted}")
	if [ -z "${dev}" ]
	then
		slotlog "trusted marker ${trusted} is not a slot name, using slot A"
		dev="${SLOT_A}"
	fi
	umount /realroot 2>/dev/null
	mount_fs "${dev}" && bootdev="${dev}"
	[ -z "${bootdev}" ] && slotlog "trusted ${trusted} did not mount"
fi

# Rung 4: recovery, which runs almost nothing and so can fail almost nowhere.
if [ -z "${bootdev}" ]
then
	umount /realroot 2>/dev/null
	slotlog "falling back to recovery"
	mount_fs "${RECOVERY}" && bootdev="${RECOVERY}"
fi

# Rung 5: expose the card over USB and stop. This is the stock initramfs
# behaviour and it is the only rung that needs a person.
if [ -z "${bootdev}" ]
then
	slotlog "no root would mount, entering mass storage mode"
	msc
fi

slotlog "booted ${bootdev}"
# --- end slot dispatch ---
