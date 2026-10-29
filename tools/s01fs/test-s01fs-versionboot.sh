#!/bin/sh
# Check that S01fs leaves a version boot alone.
#
#   test-s01fs-versionboot.sh [path-to-S01fs]
#
# On a version boot (ironkvm.version= on the kernel command line) the root is a
# read-only squashfs under a tmpfs overlay, and the initramfs has already
# mounted /data and the store. S01fs must not grow a partition, provision or
# format anything, or mount /data a second time. It still mounts /boot, which
# the version tool writes its markers on, and configfs and debugfs, which the
# USB gadget and the hardware tools need.
S01=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S01fs}
[ -f "$S01" ] || { echo "usage: test-s01fs-versionboot.sh <S01fs>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

sed -n '/^# --- version boot ---/,/^# --- end version boot ---/p' "$S01" > "$WORK/vb.sh"
sed -n '/^if \[ "\$1" = "start" \]/,$p' "$S01" > "$WORK/start.sh"
if [ -s "$WORK/vb.sh" ] && [ -s "$WORK/start.sh" ]; then
    note "the version-boot and start blocks can be extracted" OK
else
    note "the version-boot and start blocks can be extracted" FAIL
fi

printf 'console=ttyS0 root=/dev/mmcblk0p2 ro ironkvm.version=1.5.0\n' > "$WORK/cmdline.version"
printf 'console=ttyS0 root=/dev/mmcblk0p2 rootwait\n' > "$WORK/cmdline.slot"

# boot_with <cmdline file>: run the start branch with every action recorded in
# $WORK/calls and none performed.
boot_with() {
    : > "$WORK/calls"
    (
        CMDLINE=$1
        export CMDLINE
        rec() { echo "$*" >> "$WORK/calls"; }
        needs_resize()      { return 0; }
        grow_stock_root()   { rec grow_stock_root; }
        grow_root()         { rec grow_root; }
        may_autopartition() { return 0; }
        provision_disk0()   { rec provision_disk0; }
        provision_data()    { rec provision_data; }
        data_device()       { echo "$WORK/p3"; }
        data_start()        { echo 16000000; }
        devinfo_part()      { echo /dev/mmcblk0p1; }
        mount_data()        { rec "mount_data $*"; }
        mount()             { rec "mount $*"; }
        mkdir()             { :; }
        . "$WORK/vb.sh"
        set -- start
        # /etc/profile is not this test's business.
        . "$WORK/start.sh"
    ) > /dev/null 2>&1
}

# The disk0 branch reads /boot/usb.disk0 from the real filesystem. Point it at a
# file here, so that the branch is taken on every run.
sed -i 's|\[ -e /boot/usb.disk0 \]|[ -e "${USB_DISK0:-/boot/usb.disk0}" ]|' "$WORK/start.sh"
: > "$WORK/usb.disk0"
: > "$WORK/p3"
USB_DISK0=$WORK/usb.disk0
export USB_DISK0

echo
echo "===== a slot boot is unchanged ====="
boot_with "$WORK/cmdline.slot"
for want in grow_stock_root provision_disk0 provision_data "mount_data $WORK/p3 /data" "mount -t vfat /dev/mmcblk0p1 /boot"; do
    grep -qx "$want" "$WORK/calls" && note "slot boot: $want" OK || note "slot boot: $want" FAIL
done

echo
echo "===== a version boot grows, provisions and mounts nothing of /data ====="
boot_with "$WORK/cmdline.version"
for bad in grow_stock_root grow_root provision_disk0 provision_data "mount_data $WORK/p3 /data"; do
    grep -qx "$bad" "$WORK/calls" && note "version boot ran: $bad" FAIL || note "version boot does not run: $bad" OK
done
for want in "mount -t vfat /dev/mmcblk0p1 /boot" "mount -t configfs configfs /sys/kernel/config" "mount -t debugfs debugfs /sys/kernel/debug"; do
    grep -qx "$want" "$WORK/calls" && note "version boot still runs: $want" OK || note "version boot lost: $want" FAIL
done

echo
echo "===== the script parses ====="
sh -n "$S01" 2>/dev/null && note "sh -n accepts the script" OK || note "sh -n accepts the script" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
