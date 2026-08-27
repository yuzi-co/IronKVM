#!/bin/sh
# Prove that test-usb-descriptors.sh fails when the gadget scripts are wrong.
#
#   test-usb-descriptors-mutation.sh
#
# Not destructive: every mutation is made on a copy under mktemp, and the
# shipped scripts are only ever read.
#
# A descriptor test is easy to write and easy to write uselessly. Most of its
# assertions read a file the script under test wrote a moment earlier, so a
# case that reads the wrong path, or compares an empty string against an empty
# string, passes on a correct script and on a broken one alike. Running it
# against deliberately broken copies is what tells the two apart.
#
# Each mutation is a defect that has either happened here or is one edit away:
# a wrong descriptor byte, a pointer claiming boot protocol it cannot speak, the
# raw eMMC handed back to the host as a disk.
#
# NOT-APPLIED is a failure, not a skip. A sed that matches nothing leaves the
# copy identical, the test passes because there is nothing wrong with it, and
# the run reads as proof when it proved nothing.
DIR=$(dirname "$0")
INITD=${1:-$DIR/../../kvmapp/system/init.d}
TEST=$DIR/test-usb-descriptors.sh
for f in "$INITD/S03usbdev" "$INITD/S03usbhid" "$TEST"
do
    [ -f "$f" ] || { echo "missing: $f"; exit 1; }
done

pass=0
fail=0

try() {
    desc=$1
    shift

    d=$(mktemp -d)
    cp "$INITD/S03usbdev" "$INITD/S03usbhid" "$d/"
    "$@" "$d"

    if cmp -s "$d/S03usbdev" "$INITD/S03usbdev" && cmp -s "$d/S03usbhid" "$INITD/S03usbhid"
    then
        printf '  %-14s %s\n' NOT-APPLIED "$desc"
        fail=$((fail + 1))
        rm -rf "$d"
        return
    fi

    if sh "$TEST" "$d" > /dev/null 2>&1
    then
        printf '  %-14s %s\n' SURVIVED "$desc"
        fail=$((fail + 1))
    else
        printf '  %-14s %s\n' caught "$desc"
        pass=$((pass + 1))
    fi
    rm -rf "$d"
}

# A byte of the keyboard report descriptor. Both 0xE7 bytes are the top of the
# usage range, so the host stops seeing the highest key codes.
m_desc() { sed -i 's|xE7|xE6|g' "$1/S03usbdev"; }

# The composite class triple. Wrong here and Windows binds one function.
m_class() { sed -i 's|echo 0xEF > bDeviceClass|echo 0xEE > bDeviceClass|' "$1/S03usbdev"; }

# The absolute pointer declaring boot protocol. The boot mouse protocol carries
# no absolute coordinates, so a host that takes the claim at face value gets
# nonsense. This was the state of the shipped script once.
m_boot() {
    sed -i 's|^        mkdir functions/hid.GS2$|        mkdir functions/hid.GS2\n        echo 1 > functions/hid.GS2/subclass|' "$1/S03usbdev"
}

# The raw eMMC partition as the backing file for an empty marker. It has no
# MBR, so a Legacy BIOS finds no 0x55AA signature and hangs.
m_emmc() {
    sed -i 's|^            disk=$(cat /boot/usb.disk0)|            echo /dev/mmcblk0p3 > functions/mass_storage.disk0/lun.0/file\n            disk=$(cat /boot/usb.disk0)|' "$1/S03usbdev"
}

# The two gadget MACs becoming the same address.
m_mac() { sed -i 's|usb_host_mac="48:da:35:6d:|usb_host_mac="48:da:35:6e:|' "$1/S03usbdev"; }

# The report length the host reads for the hid-only keyboard.
m_hidkbd() { sed -i 's|echo 8 > functions/hid.GS0/report_length|echo 9 > functions/hid.GS0/report_length|' "$1/S03usbhid"; }

# The NCM compatible id Windows matches its driver on.
m_ncm() {
    sed -i 's|echo WINNCM > functions/ncm.usb0/os_desc/interface.ncm/compatible_id|echo WINRNDIS > functions/ncm.usb0/os_desc/interface.ncm/compatible_id|' "$1/S03usbdev"
}

# The twelve character floor on the serial. Mass storage rejects a shorter one.
m_serial() { sed -i 's|-ge 12 \]|-ge 1 ]|' "$1/S03usbdev"; }

# Read-only media becoming writable, so the host can write to a mounted image.
m_ro() { sed -i 's|echo 1 > functions/mass_storage.disk0/lun.0/ro|echo 0 > functions/mass_storage.disk0/lun.0/ro|' "$1/S03usbdev"; }

# The Microsoft OS descriptor signature.
m_osdesc() { sed -i 's|echo MSFT100 > os_desc/qw_sign|echo MSFT200 > os_desc/qw_sign|' "$1/S03usbdev"; }

# Remote wakeup on the relative mouse, which is how a keypress wakes a sleeping
# host.
m_wake() { sed -i 's|echo 1 > functions/hid.GS1/wakeup_on_write|echo 0 > functions/hid.GS1/wakeup_on_write|' "$1/S03usbdev"; }

# The hid-only configuration attributes.
m_hidonly() { sed -i 's|echo 0xA0 > configs/c.1/bmAttributes|echo 0xE0 > configs/c.1/bmAttributes|' "$1/S03usbhid"; }

# The bound on the bind retry. Without it the script waits for a controller for
# ever, and rcS never reaches the network or the server.
m_unbounded() { sed -i 's|if \[ "$n" -ge "$tries" \]|if false|' "$1/S03usbdev"; }

# The retry itself, back to the single write it replaced.
m_noretry() { sed -i 's@^    usb_bind$@    ls /sys/class/udc/ | cat > UDC@' "$1/S03usbdev"; }

# The HID unlink, without which f_hid refuses every attribute write and a mode
# switch rebinds carrying the descriptors of the mode it left.
m_nounlink_dev() { sed -i '/^    rm -f configs\/c.1\/hid.GS0 configs\/c.1\/hid.GS1 configs\/c.1\/hid.GS2$/d' "$1/S03usbdev"; }
m_nounlink_hid() { sed -i '/^    rm -f configs\/c.1\/hid.GS0 configs\/c.1\/hid.GS1 configs\/c.1\/hid.GS2$/d' "$1/S03usbhid"; }

# The prune hid-only does of whatever normal mode linked. Left in place the
# console, the disk and the network are five inbound endpoints of six before
# HID asks for three.
m_noprune_hid() { sed -i '/^    rm -f configs\/c.1\/acm.GS0 configs\/c.1\/mass_storage.disk0/,+1d' "$1/S03usbhid"; }

# The explicit device version, which is what the server reads to report the
# mode. Left to the kernel it keeps whatever the other script wrote.
m_nobcd() { sed -i '/^    echo 0x0510 > bcdDevice$/d' "$1/S03usbdev"; }

# The composite class hid-only clears.
m_noclass() { sed -i '/^    echo 0x00 > bDeviceClass$/d' "$1/S03usbhid"; }

echo "===== every mutation must be caught ====="
try "keyboard report descriptor bytes"      m_desc
try "composite device class"                m_class
try "absolute pointer claims boot HID"      m_boot
try "empty disk marker re-exports the eMMC" m_emmc
try "host MAC stops differing from device"  m_mac
try "hid-only keyboard report length"       m_hidkbd
try "NCM compatible id"                     m_ncm
try "serial length floor removed"           m_serial
try "read-only media becomes writable"      m_ro
try "OS descriptor signature"               m_osdesc
try "relative mouse stops waking the host"  m_wake
try "hid-only config attributes"            m_hidonly
try "bind retry loses its bound"            m_unbounded
try "bind retry removed"                    m_noretry
try "S03usbdev stops unlinking HID"         m_nounlink_dev
try "S03usbhid stops unlinking HID"         m_nounlink_hid
try "hid-only stops pruning the other mode" m_noprune_hid
try "the device version is left to default" m_nobcd
try "hid-only stops clearing device class"  m_noclass

echo
if [ "$fail" -eq 0 ]
then
    echo "===== all $pass mutations caught ====="
else
    echo "===== $fail of $((pass + fail)) mutations were not caught ====="
    exit 1
fi
