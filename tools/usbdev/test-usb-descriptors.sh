#!/bin/sh
# Assert the USB descriptors that both gadget scripts build.
#
#   test-usb-descriptors.sh [path-to-init.d]
#
# Not destructive: no configfs is touched, no gadget is built, and no marker is
# written. Everything runs against a directory tree under mktemp.
#
# The report descriptors below are the bytes a host reads to learn what the
# keyboard and the two pointers are. Nothing on the device reports them back,
# and a wrong byte does not announce itself: the gadget binds, the host
# enumerates, and one key or one axis is wrong on some machines and not on
# others. That is the class of defect this file exists to catch, so the expected
# values are written out in full rather than derived from the script under test.
#
# They come from sipeed/NanoKVM#814 by Mika Cohen, which asserts the same four
# descriptors against a different implementation of the same gadget. Two
# independent transcriptions that agree are the closest thing to a second
# opinion available here.
#
# That pull request could not supply the harness. Its test drives an
# S03usb-common that this fork does not have, through nine environment
# variables that these scripts do not read. This file uses the harness
# tools/usbdev/test-acm-console.sh established instead: lift every top level
# function out of the shipped script, rewrite the absolute paths as it goes, and
# run the real thing against a fake tree.
DIR=${1:-$(dirname "$0")/../../kvmapp/system/init.d}
S03=$DIR/S03usbdev
HID=$DIR/S03usbhid
for f in "$S03" "$HID"
do
    [ -f "$f" ] || { echo "usage: test-usb-descriptors.sh <init.d directory>"; exit 1; }
done

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# The device runs busybox ash. Its echo takes -ne and writes \xNN as one byte,
# and the gadget scripts build every report descriptor that way. dash does
# neither: it prints "-ne" as text and leaves the escapes alone, so the
# descriptors arrive as a run of literal backslashes and this suite reports a
# dozen failures that say nothing about the script under test. /bin/sh is dash
# on Debian and on Ubuntu, so that is the common case and not a corner.
#
# Pick the shell by what it does with the shipped line, not by its name, and
# say which one won. A harness that chooses in silence is a harness whose
# fidelity nobody can check.
SH=
cat > "$work/probe.sh" <<'PROBE'
echo -ne \\x41 > probe
PROBE
for _candidate in "busybox sh" ash bash sh
do
    command -v "${_candidate%% *}" >/dev/null 2>&1 || continue
    rm -f "$work/probe"
    (cd "$work" && $_candidate probe.sh) 2>/dev/null
    [ "$(od -An -tx1 -v "$work/probe" 2>/dev/null | tr -d ' \n')" = 41 ] || continue
    SH=$_candidate
    break
done
if [ -z "$SH" ]
then
    echo "no shell here writes \xNN as bytes the way busybox ash does." >&2
    echo "install busybox, ash or bash, then run this again." >&2
    exit 2
fi
echo "harness shell: $SH"

# The bound below proves that the retry terminates. It is not a statement about
# how fast this workstation is, and it must not become one: a subshell costs
# about 5ms on a quiet Git Bash and was measured at 124ms on a busy one, so a
# tight bound turns a loaded machine into two failing cases that name a defect
# nobody introduced. A minute still catches a loop that never ends.
# Every case runs the whole script under this bound, and the script itself
# waits for a device controller: usb_bind tries ten times with a sleep between
# them, so the no-controller cases spend about nine seconds there before they
# do anything else.
#
# 60 was not enough. On Git Bash on Windows, where every mkdir into the fake
# configfs is a separate process, both no-controller cases hit the bound on a
# tree with nothing wrong with it and reported the defects they exist to catch.
# A timeout that fires on a healthy tree is worse than a slow true failure, so
# this is set for the slowest host it runs on and overridden downwards, not up.
CASE_TIMEOUT=${CASE_TIMEOUT:-240}

# The descriptors, as a host sees them.
KEYBOARD_DESC=05010906a101050719e029e71500250175019508810295017508810395057501050819012905910295017503910395067508150025e70507190029e78100c0
HID_ONLY_KEYBOARD_DESC=05010906a101050719e029e71500250175019508810295017508810395057501050819012905910295017503910395067508150025650507190029658100c0
RELATIVE_MOUSE_DESC=05010902a1010901a1000509190129031500250195037501810295017505810305010930093109381581257f750895038106050c0a38021581257f750895018106c0c0
ABSOLUTE_MOUSE_DESC=05010902a1010901a10005091901290515002501950575018102950175038101050109300931150026ff7f350046ff7f751095028102050109381581257f35004500750895018106c0c0
HID_ONLY_ABSOLUTE_DESC=05010902a1010901a10005091901290315002501950375018102950175058101050109300931150026ff7f350046ff7f751095028102050109381581257f35004500750895018106c0c0

BASE_UID="uid 0123456789ABCDEF_0123456789ABCDEF"
SERIAL_FROM_UID=0123456789ABCDEF0123456789ABCDEF
SERIAL_FALLBACK=0123456789ABCDEF

# --- the harness ---------------------------------------------------------

# build_env makes the directories configfs would have made on its own, plus the
# few files outside configfs that the scripts read.
build_env() {
    rm -rf "$work/sys" "$work/boot" "$work/proc" "$work/etc"
    mkdir -p "$work/sys/kernel/config/usb_gadget" \
             "$work/sys/class/udc/4340000.usb" \
             "$work/sys/class/cvi-base" \
             "$work/sys/bus/platform/drivers/dwc2" \
             "$work/boot" "$work/proc/cviusb" "$work/etc"
    : > "$work/etc/profile"
    echo "$BASE_UID" > "$work/sys/class/cvi-base/base_uid"
}

# lift copies every top level function out of a script and points its absolute
# paths at the fake tree. The whole script is taken, not one function: a helper
# left undefined returns 127, which is falsy, so an "off" case would pass for
# the wrong reason and an "on" case would fail for the wrong one.
lift() {
    script=$1
    out=$work/func.sh

    # configfs creates a directory's children as soon as the directory exists.
    # A plain mkdir does not, so every write into lun.0, os_desc and strings
    # would fail against a correct script and a broken one alike.
    cat > "$out" <<'STUB'
mkdir() {
    for _arg in "$@"
    do
        case "$_arg" in -*) continue ;; esac
        /bin/mkdir -p "$_arg" || return 1
        case "$_arg" in
            g0|*/g0)
                /bin/mkdir -p "$_arg/strings" "$_arg/configs" \
                             "$_arg/functions" "$_arg/os_desc" ;;
            functions/mass_storage.*|*/functions/mass_storage.*)
                /bin/mkdir -p "$_arg/lun.0" ;;
            functions/rndis.*|*/functions/rndis.*)
                /bin/mkdir -p "$_arg/os_desc/interface.rndis" ;;
            functions/ncm.*|*/functions/ncm.*)
                /bin/mkdir -p "$_arg/os_desc/interface.ncm" ;;
        esac
    done
}
# configfs does not store a symlink. It resolves the target at the moment of the
# call, against the working directory of the caller, and records an internal
# link. A plain filesystem cannot do that: the stored target is relative to the
# link, so it dangles, and a filesystem without symlinks refuses the call
# outright. Record the link the way a reader of this tree tests for it, and keep
# the one behaviour that matters - configfs refuses a link to a function that
# was never created, so this has to refuse it too.
ln() {
    _target= _dest=
    for _arg in "$@"
    do
        case "$_arg" in -*) continue ;; esac
        if [ -z "$_target" ]
        then _target=$_arg
        else _dest=$_arg
        fi
    done
    [ -n "$_target" ] && [ -n "$_dest" ] || return 1
    [ -e "$_target" ] || return 1
    _dest=${_dest%/}
    [ -d "$_dest" ] && _dest=$_dest/${_target##*/}
    : > "$_dest"
}
sleep() { :; }
STUB

    sed -n '/^[a-z_][a-z_]*() *{$/,/^}$/p' "$script" \
        | sed "s|/sys/|$work/sys/|g; s|/boot/|$work/boot/|g; s|/proc/|$work/proc/|g; s|\. /etc/profile|:|" \
        >> "$out"
}

# run lifts a script, appends one call, and runs the result.
#
# Under a timeout, always. sleep is stubbed above, so a loop in the script that
# lost its bound does not merely run slowly here, it never ends, and the whole
# suite stops with no output. That is a defect worth reporting, not a reason to
# hang; the retry section below turns the timeout into a case of its own.
run() {
    lift "$1"
    echo "$2" >> "$work/func.sh"
    timeout "$CASE_TIMEOUT" $SH "$work/func.sh" > "$work/out" 2>&1
    G=$work/sys/kernel/config/usb_gadget/g0
}

# --- assertions ----------------------------------------------------------

is() {
    got=$(cat "$G/$1" 2>/dev/null)
    [ "$got" = "$2" ] && note "$3" OK || note "$3: got '$got', want '$2'" FAIL
}

absent() {
    if [ -e "$G/$1" ] || [ -L "$G/$1" ]
    then
        note "$2" FAIL
    else
        note "$2" OK
    fi
}

present() {
    if [ -e "$G/$1" ] || [ -L "$G/$1" ]
    then
        note "$2" OK
    else
        note "$2" FAIL
    fi
}

hex_is() {
    got=$(od -An -tx1 -v "$G/$1" 2>/dev/null | tr -d ' \n')
    [ "$got" = "$2" ] && note "$3" OK || note "$3 differs from the expected bytes" FAIL
}

# hid_is checks one HID function. An empty subclass means the function must not
# declare one at all, which is how the absolute pointer says it is not a boot
# device.
hid_is() {
    func=$1; sub=$2; proto=$3; len=$4; desc=$5; label=$6
    present "configs/c.1/$func" "$label is linked into the config"
    if [ -n "$sub" ]
    then
        is "functions/$func/subclass" "$sub" "$label subclass $sub"
    else
        absent "functions/$func/subclass" "$label claims no boot subclass"
    fi
    is "functions/$func/protocol" "$proto" "$label protocol $proto"
    is "functions/$func/report_length" "$len" "$label report length $len"
    hex_is "functions/$func/report_desc" "$desc" "$label report descriptor"
}

# --- the rewrite has to have worked --------------------------------------

echo "===== the lifted copy reaches nothing real ====="
# Every case below runs a rewritten copy of the shipped script. If the rewrite
# misses a path, that case writes into the real /sys or reads the real /boot,
# and it can then pass while proving nothing. Count the paths, do not trust the
# substitution.
build_env
lift "$S03"
for prefix in /sys/ /boot/ /proc/
do
    all=$(grep -o -- "$prefix" "$work/func.sh" | wc -l)
    ours=$(grep -o -- "$work$prefix" "$work/func.sh" | wc -l)
    if [ "$all" -eq "$ours" ]
    then
        note "every $prefix path in the lifted S03usbdev points at the fake tree" OK
    else
        note "$((all - ours)) unrewritten $prefix path(s) in the lifted S03usbdev" FAIL
    fi
done
lifted=$(grep -c '^[a-z_][a-z_]*() *{$' "$work/func.sh")
shipped=$(grep -c '^[a-z_][a-z_]*() *{$' "$S03")
if [ "$lifted" -ge "$shipped" ]
then
    note "all $shipped top level functions were lifted" OK
else
    note "only $lifted of $shipped top level functions were lifted" FAIL
fi

# --- normal mode, nothing switched on ------------------------------------

echo
echo "===== S03usbdev, no markers ====="
build_env
run "$S03" start_usb_dev

is bcdUSB    0x0200 "the USB version is stated, not left to the kernel"
is bcdDevice 0x0510 "the device version carries the HID mode flag"
is bDeviceClass    0xEF "device class 0xEF, a composite device"
is bDeviceSubClass 0x02 "device subclass 0x02"
is bDeviceProtocol 0x01 "device protocol 0x01"
is idVendor  0x3346 "default vendor id"
is idProduct 0x1009 "default product id"
is strings/0x409/manufacturer sipeed  "manufacturer string"
is strings/0x409/product      NanoKVM "product string"
is strings/0x409/serialnumber "$SERIAL_FROM_UID" "the serial follows the chip UID"
is configs/c.1/bmAttributes 0xE0 "config attributes 0xE0, remote wakeup"
is configs/c.1/MaxPower     120  "the config asks the host for 120"
is configs/c.1/strings/0x409/configuration NanoKVM "configuration string"

hid_is hid.GS0 1  1 8 "$KEYBOARD_DESC"        "keyboard"
hid_is hid.GS1 1  2 5 "$RELATIVE_MOUSE_DESC"  "relative mouse"
hid_is hid.GS2 "" 2 6 "$ABSOLUTE_MOUSE_DESC"  "absolute pointer"
absent functions/hid.GS0/wakeup_on_write "no wake on write without /boot/usb.wakeup"

absent configs/c.1/mass_storage.disk0 "no mass storage without /boot/usb.disk0"
absent configs/c.1/rndis.usb0 "no RNDIS without /boot/usb.rndis0"
absent configs/c.1/ncm.usb0   "no NCM without /boot/usb.ncm"
absent configs/c.1/acm.GS0    "no serial console without /boot/usb.acm"
absent configs/c.1/uac1.usb0  "no sound card without /boot/usb.uac"
absent os_desc/c.1            "no OS descriptor link when no network function is built"

is UDC 4340000.usb "the gadget binds to the controller"
if [ "$(cat "$work/proc/cviusb/otg_role")" = device ]
then
    note "the OTG role is set to device" OK
else
    note "the OTG role is not device" FAIL
fi

# --- identity ------------------------------------------------------------

echo
echo "===== S03usbdev, identity ====="
build_env
echo "uid SHORT" > "$work/sys/class/cvi-base/base_uid"
run "$S03" start_usb_dev
is strings/0x409/serialnumber "$SERIAL_FALLBACK" \
   "a UID under twelve characters falls back, so mass storage still works"

build_env
echo 0x1234 > "$work/boot/usb.vid"
echo 0x5678 > "$work/boot/usb.pid"
run "$S03" start_usb_dev
is idVendor  0x1234 "/boot/usb.vid overrides the vendor id"
is idProduct 0x5678 "/boot/usb.pid overrides the product id"

build_env
: > "$work/boot/usb.wakeup"
run "$S03" start_usb_dev
is functions/hid.GS0/wakeup_on_write 1 "the keyboard wakes the host when asked"
is functions/hid.GS1/wakeup_on_write 1 "the relative mouse wakes the host when asked"
is functions/hid.GS2/wakeup_on_write 1 "the absolute pointer wakes the host when asked"

build_env
: > "$work/boot/disable_hid"
run "$S03" start_usb_dev
absent configs/c.1/hid.GS0 "/boot/disable_hid removes the keyboard"
absent configs/c.1/hid.GS1 "/boot/disable_hid removes the relative mouse"
absent configs/c.1/hid.GS2 "/boot/disable_hid removes the absolute pointer"

# --- mass storage --------------------------------------------------------

echo
echo "===== S03usbdev, mass storage ====="
build_env
: > "$work/boot/usb.disk0"
run "$S03" start_usb_dev
present configs/c.1/mass_storage.disk0 "an empty /boot/usb.disk0 still builds the LUN"
is functions/mass_storage.disk0/lun.0/removable 1 "the LUN is removable, so media can be swapped"
is functions/mass_storage.disk0/lun.0/inquiry_string "NanoKVM USB Mass Storage0520" "inquiry string"
absent functions/mass_storage.disk0/lun.0/file \
   "an empty marker leaves no backing file, so the raw eMMC is never exported"

build_env
echo /data/install.iso > "$work/boot/usb.disk0"
: > "$work/boot/usb.disk0.ro"
run "$S03" start_usb_dev
is functions/mass_storage.disk0/lun.0/file /data/install.iso "the named image is inserted"
is functions/mass_storage.disk0/lun.0/ro    1 "the read-only marker sets ro"
is functions/mass_storage.disk0/lun.0/cdrom 0 "the LUN is not a CD-ROM"

# --- network -------------------------------------------------------------

echo
echo "===== S03usbdev, network ====="
# The MACs must be stable across a re-enumeration, which is what deriving them
# from the chip UID buys. The value itself is not the point; that both come from
# the same UID and differ in one nibble is.
mac_of() {
    uid=$(sha512sum < "$work/sys/class/cvi-base/base_uid" | head -c 4)
    echo "48:da:35:$1:${uid%??}:${uid#??}"
}

build_env
: > "$work/boot/usb.rndis0"
run "$S03" start_usb_dev
present configs/c.1/rndis.usb0 "RNDIS is linked into the config"
is functions/rndis.usb0/class    e0 "RNDIS class"
is functions/rndis.usb0/subclass 01 "RNDIS subclass"
is functions/rndis.usb0/protocol 03 "RNDIS protocol"
is functions/rndis.usb0/os_desc/interface.rndis/compatible_id     RNDIS   "RNDIS compatible id"
is functions/rndis.usb0/os_desc/interface.rndis/sub_compatible_id 5162001 "RNDIS sub-compatible id"
is functions/rndis.usb0/dev_addr  "$(mac_of 6e)" "the device MAC follows the chip UID"
is functions/rndis.usb0/host_addr "$(mac_of 6d)" "the host MAC follows the chip UID"
present os_desc/c.1 "the OS descriptor points at the config"
is os_desc/use 1 "OS descriptors are enabled"
is os_desc/b_vendor_code 0xCD "OS descriptor vendor code"
is os_desc/qw_sign MSFT100 "OS descriptor signature"

build_env
: > "$work/boot/usb.ncm"
: > "$work/boot/usb.rndis0"
run "$S03" start_usb_dev
present configs/c.1/ncm.usb0 "NCM wins when both markers exist"
absent  configs/c.1/rndis.usb0 "RNDIS is not built beside NCM"
is functions/ncm.usb0/os_desc/interface.ncm/compatible_id WINNCM "NCM compatible id"
is functions/ncm.usb0/dev_addr  "$(mac_of 6e)" "the NCM device MAC follows the chip UID"
is functions/ncm.usb0/host_addr "$(mac_of 6d)" "the NCM host MAC follows the chip UID"

# --- stop and restart ----------------------------------------------------

echo
echo "===== S03usbdev, stop and restart ====="
build_env
run "$S03" start_usb_dev
run "$S03" start_usb_host
is UDC "" "stop releases the controller"
if [ "$(cat "$work/proc/cviusb/otg_role")" = host ]
then
    note "stop puts the port back into host role" OK
else
    note "stop leaves the port in device role" FAIL
fi

run "$S03" restart_usb_dev
is UDC 4340000.usb "restart binds the controller again"

# --- switching modes on a gadget that already exists ---------------------

echo
echo "===== switching modes without a reboot ====="
# The mode switch used to reboot, so each script only ever ran against a gadget
# that did not exist yet. Running one against the other's gadget needs two
# things that neither script did.
#
# f_hid refuses every attribute write while the function is linked into a
# config. subclass, protocol, report_length and report_desc all return EBUSY,
# and hidg_alloc copies the values into the instance at the moment the link is
# made. So the links have to come out before the descriptors are written, or
# the gadget rebinds carrying the descriptors of the mode it just left.
#
# And whatever the previous mode linked stays linked, because stop writes an
# empty UDC and leaves configs/c.1 exactly as it was. hid-only therefore has to
# take out the console, the disk, the network and the speaker itself. Left in,
# they put the gadget over its endpoint budget, it refuses to bind, and every
# /dev/hidg* disappears.

# The disk and the network together are all six inbound endpoints once HID has
# taken three. Adding the console here would put the set over the budget and
# the network would never be linked, so the case would prove nothing about
# unlinking it. The console and the speaker get their own round below.
build_env
: > "$work/boot/usb.disk0"
: > "$work/boot/usb.rndis0"
run "$S03" start_usb_dev
present configs/c.1/mass_storage.disk0 "normal mode starts out with the disk"
present configs/c.1/rndis.usb0 "normal mode starts out with the network"
present os_desc/c.1 "normal mode starts out with the OS descriptor link"

run "$HID" start_usb_dev
is bcdDevice 0x0623 "switching to hid-only moves the mode flag"
hex_is functions/hid.GS0/report_desc "$HID_ONLY_KEYBOARD_DESC" \
   "the keyboard descriptor becomes the hid-only one"
hex_is functions/hid.GS2/report_desc "$HID_ONLY_ABSOLUTE_DESC" \
   "the absolute pointer descriptor becomes the hid-only one"
absent configs/c.1/mass_storage.disk0 "hid-only unlinks the disk"
absent configs/c.1/rndis.usb0 "hid-only unlinks the network"
absent os_desc/c.1 "hid-only removes the OS descriptor link"
is bDeviceClass 0x00 "hid-only clears the composite device class"

run "$S03" start_usb_dev
is bcdDevice 0x0510 "switching back moves the mode flag"
is bcdUSB    0x0200 "switching back restores the USB version"
is bDeviceClass 0xEF "switching back restores the composite device class"
hex_is functions/hid.GS0/report_desc "$KEYBOARD_DESC" \
   "the keyboard descriptor is the normal one again"
hex_is functions/hid.GS2/report_desc "$ABSOLUTE_MOUSE_DESC" \
   "the absolute pointer descriptor is the normal one again"
present configs/c.1/mass_storage.disk0 "the disk comes back"
present configs/c.1/rndis.usb0 "the network comes back"
present os_desc/c.1 "the OS descriptor link comes back"

# The console and the speaker are three inbound endpoints between them once HID
# has taken three, so they fit and can be checked on their own.
build_env
: > "$work/boot/usb.acm"
: > "$work/boot/usb.uac"
run "$S03" start_usb_dev
present configs/c.1/acm.GS0   "normal mode starts out with the console"
present configs/c.1/uac1.usb0 "normal mode starts out with the speaker"

# p_chmask 0 removes the sound card's IN endpoint, so the managed host gains a
# speaker and not a microphone. Left at the kernel default, Windows and most
# Linux desktops move the default recording device to the new USB card as
# readily as the playback one, and switching a speaker on would take over the
# microphone on the managed machine.
is functions/uac1.usb0/p_chmask 0 "the sound card offers no microphone"

# req_number is how many isochronous requests u_audio keeps in flight. A
# service interval the host does not fill still completes its request, and
# u_audio copies that request's buffer into the ALSA ring again, so the capture
# repeats the audio from req_number milliseconds earlier. The kernel default of
# 3 is the worst setting available here and the buffer costs 192 bytes each.
is functions/uac1.usb0/req_number 8 "the sound card keeps eight requests in flight"

run "$HID" start_usb_dev
absent configs/c.1/acm.GS0   "hid-only unlinks the console"
absent configs/c.1/uac1.usb0 "hid-only unlinks the speaker"

echo
echo "===== the sound card is configured before it is linked ====="
# configfs answers EBUSY on these attributes once the gadget holds a reference
# to the function, which was confirmed on the board. A tree of plain files
# cannot return EBUSY, so the assertions above would pass whichever order the
# script used. This one reads the shipped text instead.
uac_link=$(grep -n 'ln -s functions/uac1.usb0' "$S03" | head -1 | cut -d: -f1)
uac_unlink=$(grep -n 'rm -f configs/c.1/uac1.usb0' "$S03" | head -1 | cut -d: -f1)
if [ -z "$uac_link" ]
then
    note "S03usbdev never links the sound card" FAIL
else
    for attr in p_chmask req_number
    do
        write=$(grep -n "> functions/uac1.usb0/$attr" "$S03" | head -1 | cut -d: -f1)
        if [ -z "$write" ]
        then
            note "S03usbdev never writes uac1.usb0/$attr" FAIL
        elif [ "$write" -lt "$uac_link" ]
        then
            note "$attr is written before the sound card is linked" OK
        else
            note "$attr is written after the link, where configfs refuses it" FAIL
        fi
    done

    # The function directory and its config entry both survive a stop_start, so
    # every rebuild after the first writes to a function the config still
    # holds, and configfs refuses it. Measured on the board: a stop_start with
    # this block writing 8 left req_number at the kernel default of 3. The
    # entry has to come out before the attributes go in.
    if [ -z "$uac_unlink" ]
    then
        note "the sound card is never unlinked, so a rebuild keeps the old attributes" FAIL
    elif [ "$uac_unlink" -lt "$write" ] && [ "$uac_unlink" -lt "$uac_link" ]
    then
        note "the config entry comes out before the attributes go in" OK
    else
        note "the unlink is at line $uac_unlink, after a write or the link" FAIL
    fi
fi

echo
echo "===== the HID links come out before the descriptors go in ====="
# A tree of plain files cannot return EBUSY, so the section above would pass
# whether or not the links are removed first. This one reads the shipped text
# instead, which is where that ordering actually lives.
order_case() {
    name=$1
    body=$(sed -n '/^start_usb_dev()/,/^}/p' "$2")

    first_of() { printf '%s\n' "$body" | grep -n -- "$1" | head -1 | cut -d: -f1; }

    unlink_at=$(first_of 'rm -f configs/c.1/hid')
    write_at=$(first_of 'functions/hid.GS0/subclass')
    relink_at=$(first_of 'ln -s functions/hid.GS0')

    if [ -z "$unlink_at" ]
    then
        note "$name unlinks the HID functions before it configures them" FAIL
        return
    fi
    if [ -n "$write_at" ] && [ "$unlink_at" -lt "$write_at" ]
    then
        note "$name unlinks the HID functions before it writes their attributes" OK
    else
        note "$name writes HID attributes while the functions are still linked" FAIL
    fi
    if [ -n "$relink_at" ] && [ "$unlink_at" -lt "$relink_at" ]
    then
        note "$name unlinks before it links again" OK
    else
        note "$name links the HID functions before it unlinks them" FAIL
    fi
}
order_case S03usbdev "$S03"
order_case S03usbhid "$HID"

# --- binding when the controller is late ---------------------------------

echo
echo "===== S03usbdev, a controller that is not there yet ====="
# dwc2 can still be probing when this script runs. A single write then puts
# nothing into UDC, and an empty UDC is how a caller deliberately unbinds, so
# the gadget is built and never bound and nothing says so: /dev/hidg* are
# missing and the board looks like it has no keyboard.
no_controller() {
    build_env
    rm -rf "$work/sys/class/udc"
    mkdir -p "$work/sys/class/udc"
}

no_controller
run "$S03" start_usb_dev
is UDC "" "no controller present leaves the gadget unbound"
if grep -q unbound "$work/out"
then
    note "the script reports that it gave up" OK
else
    note "the script gives up silently, so nothing records why HID is missing" FAIL
fi

# The retry has to be bounded. This script runs from the rcS wait entry, so a
# loop that waits for ever stops the boot before the network and the server
# start, and the USB console cannot be reached to find out why.
no_controller
lift "$S03"
echo start_usb_dev >> "$work/func.sh"
if timeout "$CASE_TIMEOUT" $SH "$work/func.sh" > "$work/out" 2>&1
then
    note "the retry gives up rather than holding rcS for ever" OK
elif [ $? -eq 124 ]
then
    note "the retry never returns, so rcS stops here" FAIL
else
    note "the retry gives up rather than holding rcS for ever" OK
fi

# sleep is the retry's only pause, so a stub that brings the controller up on
# the third one shows that a later attempt is what binds.
no_controller
lift "$S03"
{
    printf '%s\n' 'sleep() { _n=$((${_n:-0} + 1)); [ "$_n" -ge 3 ] && mkdir -p "'"$work"'/sys/class/udc/4340000.usb"; return 0; }'
    printf '%s\n' start_usb_dev
} >> "$work/func.sh"
timeout "$CASE_TIMEOUT" $SH "$work/func.sh" > "$work/out" 2>&1
G=$work/sys/kernel/config/usb_gadget/g0
is UDC 4340000.usb "a controller that appears late is still picked up"

# --- the HID-only script -------------------------------------------------

echo
echo "===== S03usbhid ====="
build_env
: > "$work/boot/usb.disk0"
: > "$work/boot/usb.rndis0"
: > "$work/boot/usb.acm"
run "$HID" start_usb_dev

is bcdUSB    0x0101 "hid-only bcdUSB"
is bcdDevice 0x0623 "hid-only bcdDevice"
is idVendor  0x3346 "hid-only default vendor id"
is idProduct 0x1009 "hid-only default product id"
is configs/c.1/bmAttributes 0xA0 "hid-only config attributes"
is configs/c.1/MaxPower     200  "hid-only config asks the host for 200"
is strings/0x409/manufacturer sipeed  "hid-only manufacturer string"
is strings/0x409/product      NanoKVM "hid-only product string"
absent strings/0x409/serialnumber "hid-only sets no serial number"
is bDeviceClass    0x00 "hid-only clears the composite device class"
is bDeviceSubClass 0x00 "hid-only clears the composite device subclass"
is bDeviceProtocol 0x00 "hid-only clears the composite device protocol"

hid_is hid.GS0 1  1 8 "$HID_ONLY_KEYBOARD_DESC" "hid-only keyboard"
hid_is hid.GS1 1  2 5 "$RELATIVE_MOUSE_DESC"    "hid-only relative mouse"
hid_is hid.GS2 "" 2 6 "$HID_ONLY_ABSOLUTE_DESC" "hid-only absolute pointer"

absent configs/c.1/mass_storage.disk0 "hid-only ignores /boot/usb.disk0"
absent configs/c.1/rndis.usb0 "hid-only ignores /boot/usb.rndis0"
absent configs/c.1/acm.GS0    "hid-only ignores /boot/usb.acm"
is UDC 4340000.usb "hid-only binds to the controller"

echo
if [ "$fails" -eq 0 ]
then
    echo "===== all cases passed ====="
else
    echo "===== $fails problem(s) ====="
    exit 1
fi
