#!/bin/sh
# Prove that test-usb-enumeration.sh fails when the watch is wrong.
#
#   test-usb-enumeration-mutation.sh
#
# Not destructive: every mutation is made on a copy under mktemp, and the
# shipped script is only ever read.
#
# A watch loop is the easy thing to test uselessly. Cases that only assert "it
# came back" pass on a watch that returns immediately and does nothing, and
# cases that only assert "it rebound" pass on one that rebinds a healthy gadget
# for ever. Both of those are worse than no watch: this one runs unattended on
# every boot, and the second would take the USB gadget away from a board that
# was working.
#
# Each run is wrapped in a timeout. A mutation that removes the bound does not
# fail the cases, it hangs them, and a hung suite is not a caught mutation.
#
# NOT-APPLIED is a failure, not a skip. A sed that matches nothing leaves the
# copy identical, the cases pass because there is nothing wrong with it, and the
# run reads as proof when it proved nothing.
DIR=$(dirname "$0")
INITD=${1:-$DIR/../../kvmapp/system/init.d}
TEST=$DIR/test-usb-enumeration.sh
for f in "$INITD/S03usbdev" "$TEST"; do
    [ -f "$f" ] || { echo "missing: $f"; exit 1; }
done

pass=0
fail=0

try() {
    desc=$1
    shift

    d=$(mktemp -d)
    cp "$INITD/S03usbdev" "$INITD/S03usbhid" "$d/" 2>/dev/null
    "$@" "$d/S03usbdev"

    if cmp -s "$d/S03usbdev" "$INITD/S03usbdev"; then
        printf '  %-14s %s\n' NOT-APPLIED "$desc"
        fail=$((fail + 1))
        rm -rf "$d"
        return
    fi

    if timeout 30 sh "$TEST" "$d" > /dev/null 2>&1; then
        printf '  %-14s %s\n' SURVIVED "$desc"
        fail=$((fail + 1))
    else
        printf '  %-14s %s\n' caught "$desc"
        pass=$((pass + 1))
    fi
    rm -rf "$d"
}

# The bound removed. rcS does not wait on this, so it does not stop the boot,
# but it does leave a shell polling a controller for as long as the board is up.
m_unbounded() {
    sed -i 's@^        if \[ "$rebinds" -ge "$USB_ENUM_TRIES" \]@        if false@' "$1"
}

# The watch becomes a report: it notices, and does nothing about it.
m_norebind() {
    sed -i '\@^        echo "" > /sys/kernel/config/usb_gadget/g0/UDC 2>/dev/null@d' "$1"
}

# Every state counts as enumerated, so the watch always returns happy and the
# board that cannot type still cannot type.
m_anystate() {
    sed -i 's@\[ "$(cat "$f")" = configured \] && return 0@return 0@' "$1"
}

# The window is skipped, so a managed host that is still powering on gets its
# enumeration interrupted by a rebind it did not need.
m_nowait() {
    sed -i 's@^        while \[ "$waited" -lt "$USB_ENUM_WAIT" \]@        while false@' "$1"
}

# The watch runs inline. rcS waits on this script, so every boot of a board with
# no host attached now pays the whole budget before the network starts.
m_inline() {
    sed -i 's@^    usb_watch_enumeration &$@    usb_watch_enumeration@' "$1"
}

# The boot never starts the watch at all.
m_nowatch() {
    sed -i '\@^    usb_watch_enumeration &$@d' "$1"
}

# The speed check goes, so usb_link_ok is back to asking only whether the host
# took the gadget. This is the defect as it shipped: the board that came up
# full-speed reached "configured", and the watch reported it healthy.
m_nospeedcheck() {
    sed -i 's@^    \[ "$(usb_link_speed)" = high-speed \]$@    :@' "$1"
}

# The check is inverted, so a healthy high-speed link is rebound and a useless
# full-speed one is kept.
m_invertedspeed() {
    sed -i 's@^    \[ "$(usb_link_speed)" = high-speed \]$@    [ "$(usb_link_speed)" != high-speed ]@' "$1"
}

# The console reports every fault as "not enumerated", so the one line an
# operator gets names the wrong problem and sends them after the wrong thing.
m_faultlies() {
    sed -i 's@^        echo "the host enumerated the gadget at $(usb_link_speed)"$@        echo "not enumerated"@' "$1"
}

echo "===== every mutation must be caught ====="
try "the rebind bound is removed"              m_unbounded
try "it notices but never rebinds"             m_norebind
try "any controller state counts as enumerated" m_anystate
try "it rebinds without waiting first"         m_nowait
try "the watch delays the boot"                m_inline
try "the boot never starts the watch"          m_nowatch
try "the speed check is dropped"               m_nospeedcheck
try "the speed check is inverted"              m_invertedspeed
try "the console names the wrong fault"        m_faultlies

echo
if [ "$fail" -eq 0 ]; then
    echo "===== all $pass mutations caught ====="
else
    echo "===== $fail of $((pass + fail)) mutations were not caught ====="
    exit 1
fi
