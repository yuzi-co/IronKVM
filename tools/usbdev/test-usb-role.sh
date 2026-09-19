#!/bin/sh
# Where the USB port's role and its PHY restart come from.
#
#   test-usb-role.sh [path-to-init.d-dir]
#
# Two facts about this board used to be spelled out in S03usbdev: how the port
# changes between the device role and the host role, and which platform device
# the PHY restart unbinds and binds again. Both now come from the device
# description, and the role change comes as an action rather than a value,
# because the write has to be read back and a value cannot do that.
#
# So set_otg_role has two paths and this suite holds both of them:
#
#   a helper at /usr/lib/ironkvm/hw.d/usb-role, which an image built by
#   ironkvm-dist installs, and which carries the read-back loop
#
#   the /proc/cviusb/otg_role code that was always here, which is what a board
#   on Sipeed's firmware runs, because it has this application tarball and
#   nothing else
#
# The second path is the one that must not change. A release goes to boards that
# have no helper at all, and the cases below run it exactly as before.
#
# == Nothing here reaches the real files ==
#
# Every function is lifted out of the script with its paths pointed at a
# sandbox, the way tools/usbdev/test-usb-rebind.sh does it. That matters more
# here than elsewhere: a write to the real /proc/cviusb/otg_role flips the
# hardware ID pin, and a write to the real dwc2 unbind takes the gadget away
# from whoever is logged in through it.
#
# == The description is read with the real reader ==
#
# ironkvm-deviceinfo is copied into the sandbox and given a fixture to read, so
# a key this suite spells wrong is a key the suite fails on, not one a stub
# agrees with.
set -u

DIR=$(dirname "$0")
INITD=${1:-$DIR/../../kvmapp/system/init.d}
READER=${READER:-$INITD/../ironkvm-deviceinfo}

[ -f "$READER" ] || { echo "no ironkvm-deviceinfo at $READER"; exit 2; }

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

mkdir -p "$work/bin" "$work/hook"
cp "$READER" "$work/bin/ironkvm-deviceinfo"
chmod 0755 "$work/bin/ironkvm-deviceinfo"

DEVINFO=$work/bin/ironkvm-deviceinfo
DEVICEINFO_PATHS=$work/deviceinfo
USB_ROLE_HOOK=$work/hook/usb-role

# lift copies every shell function out of the script and points the paths it
# touches at the sandbox. start_usb_host and start_usb_dev are replaced
# afterwards, because restart_phy calls them and this suite is about the two
# writes that come before them.
lift() {
    _s03=$1
    rm -rf "$work/sys" "$work/proc" "$work/boot"
    mkdir -p "$work/sys/bus/platform/drivers/dwc2" \
             "$work/sys/class/udc/4340000.usb" \
             "$work/sys/kernel/config/usb_gadget/g0" \
             "$work/proc/cviusb" "$work/boot"

    printf 'sleep() { :; }\n' > "$work/f.sh"
    sed -n '/^[a-z_][a-z_]*() *{$/,/^}$/p' "$_s03" \
        | sed "s|/sys/|$work/sys/|g; s|/proc/|$work/proc/|g; s|/boot/|$work/boot/|g" \
        >> "$work/f.sh"
    cat >> "$work/f.sh" <<STUB
start_usb_host() { echo "start_usb_host" >> "$work/calls"; return 0; }
start_usb_dev() { echo "start_usb_dev" >> "$work/calls"; return 0; }
STUB
}

# run <function> [arg]: leaves the output and status in $work/out.
run() {
    : > "$work/calls"
    (
        USB_OTG_ROLE_TRIES=1
        export USB_OTG_ROLE_TRIES DEVINFO DEVICEINFO_PATHS USB_ROLE_HOOK
        . "$work/f.sh"
        "$@"
        echo "rc=$?"
    ) > "$work/out" 2>&1
}

# devinfo <line>...: the description the reader is given.
devinfo() { printf '%s\n' "$@" > "$work/deviceinfo"; }

# helper <exit status>: a usb-role helper that records what it was passed.
helper() {
    printf '#!/bin/sh\necho "$@" >> "%s/hook.log"\nexit %s\n' "$work" "$1" > "$USB_ROLE_HOOK"
    chmod 0755 "$USB_ROLE_HOOK"
    : > "$work/hook.log"
}
no_helper() { rm -f "$USB_ROLE_HOOK"; : > "$work/hook.log"; }

# The role file answers a write, or it does not. Making the parent a plain file
# is how a refused /proc write is simulated: the redirect fails, and the
# read-back that follows finds nothing, which is the state the retry loop exists
# for.
role_writable() { rm -rf "$work/proc/cviusb"; mkdir -p "$work/proc/cviusb"; }
role_stuck()    { rm -rf "$work/proc/cviusb"; : > "$work/proc/cviusb"; }
role_says()     { command cat "$work/proc/cviusb/otg_role" 2>/dev/null; }

FULL="USB_ROLE_SWITCH=hook USB_PHY_DRIVER=dwc2 USB_PHY_DEVICE=4340000.usb"

for script in S03usbdev S03usbhid
do
    S03=$INITD/$script
    [ -f "$S03" ] || { echo "usage: test-usb-role.sh <init.d dir>"; exit 1; }

    echo "===== $script ====="
    lift "$S03"

    if ! grep -q '^usb_devinfo() {' "$work/f.sh"
    then
        note "$script reads the device description" FAIL
        continue
    fi
    note "$script reads the device description" OK

    # --- the helper path ---

    devinfo $FULL
    helper 0
    role_writable
    run set_otg_role device
    grep -q 'rc=0' "$work/out" \
        && note "a helper that takes makes set_otg_role succeed" OK \
        || note "a helper that took was reported as a failure" FAIL
    [ "$(command cat "$work/hook.log")" = device ] \
        && note "the helper is called and passed the role" OK \
        || note "the helper got '$(command cat "$work/hook.log")', want 'device'" FAIL
    [ -z "$(role_says)" ] \
        && note "the proc file is left alone when a helper is there" OK \
        || note "it wrote the proc file although a helper was there" FAIL

    helper 0
    run set_otg_role host
    [ "$(command cat "$work/hook.log")" = host ] \
        && note "the host role reaches the helper as well" OK \
        || note "the helper got '$(command cat "$work/hook.log")', want 'host'" FAIL

    # The status is the whole contract of the helper. A caller that ignored it
    # would report a port that changed role when it did not, which is the
    # failure the read-back was added to catch in the first place.
    helper 1
    role_writable
    run set_otg_role device
    grep -q 'rc=1' "$work/out" \
        && note "a helper that exits 1 makes set_otg_role fail" OK \
        || note "a helper that failed was reported as success" FAIL

    # --- USB_ROLE_SWITCH=none ---

    devinfo "USB_ROLE_SWITCH=none"
    helper 0
    role_writable
    run set_otg_role device
    grep -q 'rc=0' "$work/out" \
        && note "a board with no role switch succeeds" OK \
        || note "a board with no role switch was reported as a failure" FAIL
    [ ! -s "$work/hook.log" ] && [ -z "$(role_says)" ] \
        && note "and it neither calls the helper nor writes the proc file" OK \
        || note "it acted on a board whose port has one role" FAIL

    # --- the old path, which is what a Sipeed board runs ---

    devinfo $FULL
    no_helper
    role_writable
    run set_otg_role device
    grep -q 'rc=0' "$work/out" \
        && note "with no helper the proc write reports success" OK \
        || note "with no helper a good proc write was reported as a failure" FAIL
    [ "$(role_says)" = device ] \
        && note "and the role reaches /proc/cviusb/otg_role" OK \
        || note "the proc file says '$(role_says)', want 'device'" FAIL

    no_helper
    role_stuck
    run set_otg_role device
    grep -q 'rc=1' "$work/out" \
        && note "the read-back still gates the result with no helper" OK \
        || note "a role that never read back was reported as success" FAIL
    grep -q 'usb: otg role stayed at' "$work/out" \
        && note "and the operator's message is unchanged" OK \
        || note "the message an operator searches for is gone" FAIL

    # --- restart_phy ---

    devinfo $FULL
    run restart_phy
    grep -q 'rc=0' "$work/out" \
        && note "restart_phy reports success" OK \
        || note "restart_phy reported a failure" FAIL
    [ "$(command cat "$work/sys/bus/platform/drivers/dwc2/unbind" 2>/dev/null)" = 4340000.usb ] \
        && note "restart_phy writes the device name to the driver's unbind" OK \
        || note "the driver's unbind got '$(command cat "$work/sys/bus/platform/drivers/dwc2/unbind" 2>/dev/null)'" FAIL
    [ "$(command cat "$work/sys/bus/platform/drivers/dwc2/bind" 2>/dev/null)" = 4340000.usb ] \
        && note "restart_phy writes it to the driver's bind as well" OK \
        || note "the driver's bind got '$(command cat "$work/sys/bus/platform/drivers/dwc2/bind" 2>/dev/null)'" FAIL
    grep -q 'this board has no PHY restart' "$work/out" \
        && note "it said the board cannot do what it just did" FAIL \
        || note "it does not claim the board has no PHY restart" OK

    # A description with one key and not the other has said nothing usable, and
    # the path built from the empty half is a write to the wrong file.
    lift "$S03"
    devinfo "USB_ROLE_SWITCH=hook" "USB_PHY_DRIVER=dwc2"
    run restart_phy
    grep -q 'rc=0' "$work/out" \
        && note "restart_phy with no device key exits 0" OK \
        || note "restart_phy with no device key reported a failure" FAIL
    grep -q 'this board has no PHY restart' "$work/out" \
        && note "and it says this board has no PHY restart" OK \
        || note "it stopped without saying why" FAIL
    [ ! -e "$work/sys/bus/platform/drivers/dwc2/unbind" ] \
        && note "and it writes nothing to the driver" OK \
        || note "it wrote '$(command cat "$work/sys/bus/platform/drivers/dwc2/unbind")' with no device key" FAIL
    [ ! -s "$work/calls" ] \
        && note "and it does not rebuild the gadget either" OK \
        || note "it went on to rebuild the gadget: $(tr '\n' ' ' < "$work/calls")" FAIL

    # The empty half is the driver this time, so the path it would build is
    # /sys/bus/platform/drivers//unbind, which the kernel reads as a file called
    # unbind directly under drivers.
    lift "$S03"
    devinfo "USB_ROLE_SWITCH=hook" "USB_PHY_DEVICE=4340000.usb"
    run restart_phy
    grep -q 'rc=0' "$work/out" && grep -q 'this board has no PHY restart' "$work/out" \
        && note "restart_phy with no driver key does the same" OK \
        || note "restart_phy with no driver key did something else" FAIL
    [ ! -e "$work/sys/bus/platform/drivers/unbind" ] \
        && note "and it builds no path from an empty driver" OK \
        || note "it wrote to /sys/bus/platform/drivers//unbind" FAIL
    [ ! -s "$work/calls" ] \
        && note "and it rebuilds no gadget with no driver key either" OK \
        || note "it went on to rebuild the gadget: $(tr '\n' ' ' < "$work/calls")" FAIL

    echo
done

if [ "$fails" -eq 0 ]
then
    echo "all cases PASSED"
    exit 0
fi

echo "$fails case(s) FAILED"
exit 1
