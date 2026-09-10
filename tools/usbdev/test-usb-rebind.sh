#!/bin/sh
# Check that a rebind which did not happen is reported as a failure.
#
#   test-usb-rebind.sh [path-to-init.d-dir]
#
# Emptying the UDC is how a caller unbinds the gadget, and the write is not
# proof that it worked. configfs answers ENODEV for a gadget that is already
# unbound, and a write that fails for any other reason leaves the controller
# name in place. The shell discards both.
#
# usb_bind then waits only while the UDC is empty. Handed one that is still
# populated it runs no iteration at all and returns 0, so `restart` printed
# "USB Restart OK!" over a gadget it had not touched.
#
# The supervisor in server/service/hid/usb_watchdog.go is what makes that
# expensive. `restart` is the cheap rung it tries twice before escalating to
# stop_start, so two silent no-ops read as two honest repairs that did not
# help, and the escalation was spent on the strength of them.
#
# The controller name is the evidence, so these cases hold the readback stuck
# rather than trying to make a file unwritable: file modes do not survive every
# sandbox this suite runs in, and the readback is the part under test.
set -u

DIR=$(dirname "$0")
INITD=${1:-$DIR/../../kvmapp/system/init.d}

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

UDCFILE=/sys/kernel/config/usb_gadget/g0/UDC

# lift copies every shell function out of the script and points the paths it
# touches at the sandbox. /proc and /boot go with /sys, so a case that runs
# start_usb_dev cannot reach the workstation's own files.
lift() {
    S03=$1
    mkdir -p "$work/sys/class/udc/4340000.usb" \
             "$work/sys/kernel/config/usb_gadget/g0" \
             "$work/proc/cviusb" "$work/boot"
    : > "$work$UDCFILE"
    : > "$work/binds"
    rm -f "$work/stuck"

    cat > "$work/f.sh" <<STUB
sleep() { :; }

# The readback is what usb_unbind judges the write by. A "stuck" file stands
# for a controller that stayed bound however the write was answered.
cat() {
    if [ "\${1:-}" = "$work$UDCFILE" ] && [ -f "$work/stuck" ]
    then
        command cat "$work/stuck"
        return 0
    fi
    command cat "\$@" 2>/dev/null
}

# usb_bind names a controller by listing this directory, and only when it is
# about to write one into UDC. One call here is one real bind, which is what
# separates a rebind from a message announcing one.
ls() {
    if [ "\${1:-}" = "$work/sys/class/udc/" ]
    then
        echo x >> "$work/binds"
        echo 4340000.usb
        return
    fi
    command ls "\$@" 2>/dev/null
}
STUB
    sed -n '/^[a-z_][a-z_]*() *{$/,/^}$/p' "$S03" \
        | sed "s|/sys/|$work/sys/|g; s|/proc/|$work/proc/|g; s|/boot/|$work/boot/|g; s|\. /etc/profile|:|" \
        >> "$work/f.sh"
}

# run <function>: leaves the output in $work/out and the bind count in $work/binds.
run() {
    : > "$work/binds"
    (
        USB_BIND_TRIES=2
        USB_OTG_ROLE_TRIES=1
        export USB_BIND_TRIES USB_OTG_ROLE_TRIES
        . "$work/f.sh"
        "$1"
        echo "rc=$?"
    ) > "$work/out" 2>&1
}

bound() { echo 4340000.usb > "$work$UDCFILE"; rm -f "$work/stuck"; }
stuck() { echo 4340000.usb > "$work$UDCFILE"; echo 4340000.usb > "$work/stuck"; }
free_() { : > "$work$UDCFILE"; rm -f "$work/stuck"; }

for script in S03usbdev S03usbhid
do
    S03=$INITD/$script
    [ -f "$S03" ] || { echo "usage: test-usb-rebind.sh <init.d dir>"; exit 1; }

    echo "===== $script ====="
    lift "$S03"

    if ! grep -q '^usb_unbind() {' "$work/f.sh"
    then
        note "$script defines usb_unbind" FAIL
        continue
    fi
    note "$script defines usb_unbind" OK

    # A gadget that really does come unbound has to stay cheap: this is every
    # rebind that works, and a check that cried wolf here would be worse than
    # the bug it replaces.
    bound; run usb_unbind
    grep -q 'rc=0' "$work/out" \
        && note "an unbind that takes reports success" OK \
        || note "an unbind that takes was reported as a failure" FAIL
    grep -q 'still bound' "$work/out" \
        && note "it complained about a gadget it did unbind" FAIL \
        || note "it says nothing when the unbind takes" OK

    stuck; run usb_unbind
    grep -q 'rc=1' "$work/out" \
        && note "an unbind that does not take reports failure" OK \
        || note "an unbind that did not take was reported as success" FAIL
    grep -q '4340000.usb' "$work/out" \
        && note "it names the controller still holding the gadget" OK \
        || note "it does not say what the gadget is still bound to" FAIL

    # The whole point: the caller must not go on to bind, and must not print
    # the line an operator reads as "your gadget was restarted".
    stuck; run restart_usb_dev
    grep -q 'rc=1' "$work/out" \
        && note "restart reports the failure" OK \
        || note "restart reported success over a gadget it never unbound" FAIL
    grep -q 'USB Restart OK' "$work/out" \
        && note "restart announced a rebind that did not happen" FAIL \
        || note "restart does not announce a rebind that did not happen" OK
    [ "$(wc -l < "$work/binds")" -eq 0 ] \
        && note "restart attempted no bind on a gadget still bound" OK \
        || note "restart wrote the UDC while the gadget was still bound" FAIL

    free_; run restart_usb_dev
    grep -q 'rc=0' "$work/out" \
        && note "restart still succeeds on a gadget it can unbind" OK \
        || note "restart failed on a gadget it could unbind" FAIL
    [ "$(wc -l < "$work/binds")" -eq 1 ] \
        && note "the controller received exactly one bind" OK \
        || note "the controller received $(wc -l < "$work/binds") bind(s), want 1" FAIL

    # stop_start reaches start_usb_dev through start_usb_host. The case arm
    # cannot test that unbind without also making a failed otg role write
    # fatal, so the refusal lives where the invariant does.
    stuck; run start_usb_host
    grep -q 'rc=1' "$work/out" \
        && note "the host switch reports an unbind that did not take" OK \
        || note "the host switch reported success without unbinding" FAIL
    grep -q 'otg role' "$work/out" \
        && note "it went on to write the otg role anyway" FAIL \
        || note "it stops before writing the otg role" OK

    stuck; run start_usb_dev
    grep -q 'rc=1' "$work/out" \
        && note "a rebuild refuses a gadget that is still bound" OK \
        || note "it rebuilt descriptors f_hid would have refused" FAIL
    grep -q 'not rebuilding' "$work/out" \
        && note "the console says why the rebuild stopped" OK \
        || note "the rebuild stopped without saying why" FAIL

    free_; run start_usb_dev
    grep -q 'not rebuilding' "$work/out" \
        && note "it refused an unbound gadget it should have rebuilt" FAIL \
        || note "an unbound gadget is rebuilt as before" OK

    echo
done

if [ "$fails" -eq 0 ]
then
    echo "all cases PASSED"
    exit 0
fi

echo "$fails case(s) FAILED"
exit 1
