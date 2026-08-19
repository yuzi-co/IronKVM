#!/bin/sh
# Check that the boot path notices a gadget the host never enumerated.
#
#   test-usb-enumeration.sh [path-to-init.d-dir]
#
# Binding the gadget to a controller is not the same as the host accepting it.
# On 2026-08-19 a board booted with the gadget bound, the three /dev/hidg*
# present, and every descriptor correct, and no input reached the managed host
# at all. dmesg showed the host resetting the bus nine times between t=3.8s and
# t=5.7s and never issuing SET_ADDRESS. Running stop_start hours later enumerated
# it at high speed on the first attempt and every endpoint came back.
#
# Nothing reported the failure. `cat UDC` names the controller either way, the
# device nodes exist either way, and the server only learns of it when a write
# times out, which is after somebody has already pressed a key that went nowhere.
#
# The controller does report it. /sys/class/udc/<udc>/state reads "configured"
# once the host has accepted the gadget, and something below that when it has
# not. Verified on the device: configured while working, "not attached" while
# unbound, configured again after a rebind.
#
# So the boot path waits for that, and rebinds if it does not arrive. What it
# must not do is delay the boot: rcS waits on S03usbdev, and a board with no
# host attached never reaches "configured" at all. That case is normal, not a
# fault, and it has to cost nothing.
set -u

DIR=$(dirname "$0")
INITD=${1:-$DIR/../../kvmapp/system/init.d}
S03=$INITD/S03usbdev
[ -f "$S03" ] || { echo "usage: test-usb-enumeration.sh <init.d dir>"; exit 1; }

fails=0
note() { printf '  %-60s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

UDC=/sys/class/udc/4340000.usb

# reset <state> <speed> <settles-after-N-sleeps>
#
# The fake controller starts in <state> at <speed>. The sleep stub counts calls
# and settles the controller to "configured" at high speed once it has been
# called N times, so a case can say "the host accepts it on the third poll"
# without waiting three seconds.
reset() {
    rm -rf "$work/sys" "$work/kernel"
    mkdir -p "$work$UDC" "$work/sys/kernel/config/usb_gadget/g0"
    printf '%s\n' "$1" > "$work$UDC/state"
    printf '%s\n' "$2" > "$work$UDC/current_speed"
    : > "$work/sys/kernel/config/usb_gadget/g0/UDC"
    echo "$3" > "$work/flip_at"
    : > "$work/binds"
}

lift() {
    cat > "$work/f.sh" <<STUB
# Every second of waiting is a stub, so the cases are instant. Each sleep is
# also the tick that drives the fake controller forward.
sleep() {
    n=\$(cat "$work/ticks" 2>/dev/null || echo 0)
    n=\$((n + 1))
    echo "\$n" > "$work/ticks"
    if [ "\$n" -ge "\$(cat "$work/flip_at")" ] && [ "\$(cat "$work/flip_at")" -gt 0 ]; then
        echo configured > "$work$UDC/state"
        echo high-speed > "$work$UDC/current_speed"
    fi
}

# usb_bind names a controller by listing this directory, and it does that only
# when it is about to write one into UDC. One call here is one real bind, which
# is what separates a rebind from a message announcing one.
ls() {
    if [ "\${1:-}" = "$work/sys/class/udc/" ]; then
        echo x >> "$work/binds"
        echo 4340000.usb
        return
    fi
    command ls "\$@" 2>/dev/null
}
STUB
    sed -n '/^[a-z_][a-z_]*() *{$/,/^}$/p' "$S03" \
        | sed "s|/sys/|$work/sys/|g; s|\. /etc/profile|:|" >> "$work/f.sh"
}

lift
if ! grep -q '^usb_watch_enumeration() {' "$work/f.sh"; then
    note "S03usbdev defines usb_watch_enumeration" FAIL
    echo; echo "$fails case(s) FAILED"; exit "$fails"
fi
note "S03usbdev defines usb_watch_enumeration" OK

# run: prints the function's own output; leaves the tick count in $work/ticks.
run() {
    : > "$work/ticks"
    (
        USB_ENUM_TRIES=${TRIES:-2}
        USB_ENUM_WAIT=${WAIT:-5}
        export USB_ENUM_TRIES USB_ENUM_WAIT
        . "$work/f.sh"
        usb_watch_enumeration
        echo "rc=$?"
    ) > "$work/out" 2>&1
}

echo "===== a host that accepts the gadget costs nothing ====="

reset configured high-speed 0
run
grep -q 'rc=0' "$work/out" \
    && note "an already configured gadget succeeds" OK \
    || note "an already configured gadget was reported as a failure" FAIL
[ "$(cat "$work/ticks")" = "" ] || [ "$(cat "$work/ticks")" -le 1 ] \
    && note "it does not wait when the gadget is already configured" OK \
    || note "it waited $(cat "$work/ticks") tick(s) for nothing" FAIL
grep -q 'rebinding' "$work/out" \
    && note "it rebound a gadget the host had already accepted" FAIL \
    || note "it does not rebind a gadget the host accepted" OK
[ "$(wc -l < "$work/binds")" -eq 0 ] \
    && note "the controller received no bind it did not need" OK \
    || note "the controller was rebound for nothing" FAIL

echo
echo "===== a host that is slow is waited for, not interrupted ====="

# The managed host may be powering on. Arriving late is not a fault.
reset "not attached" unknown 3
run
grep -q 'rc=0' "$work/out" \
    && note "a host that enumerates on the third poll succeeds" OK \
    || note "a host that arrives late was given up on" FAIL
grep -q 'rebinding' "$work/out" \
    && note "it rebound while the host was still coming" FAIL \
    || note "it waits out the window before rebinding" OK

echo
echo "===== a gadget the host never takes is rebound, and bounded ====="

reset "not attached" unknown 0
TRIES=2 run
grep -q 'rebinding' "$work/out" \
    && note "a gadget that never enumerates is rebound" OK \
    || note "it left an unenumerated gadget alone" FAIL

got=$(grep -c 'rebinding' "$work/out")
[ "$got" -eq 2 ] \
    && note "it rebinds exactly USB_ENUM_TRIES times (2)" OK \
    || note "it rebound $got time(s), want 2" FAIL

# Saying it rebound and rebinding are different things, and only one of them
# gets the gadget back. Count the binds the controller actually received.
bound=$(wc -l < "$work/binds")
[ "$bound" -eq 2 ] \
    && note "the controller actually received 2 binds" OK \
    || note "the controller received $bound bind(s), want 2" FAIL

grep -q 'rc=1' "$work/out" \
    && note "it reports the failure rather than hanging" OK \
    || note "it did not report that the gadget is unenumerated" FAIL

# The console has to name this fault correctly, and asserting that is also the
# only thing left that proves usb_enumerated still reads the controller state.
# Every other case here reaches the same outcome through the speed check, so a
# usb_enumerated that answered yes to everything would go unnoticed.
grep -q 'has not enumerated the gadget' "$work/out" \
    && note "the console says the host never took the gadget" OK \
    || note "the console misreports a gadget the host never took" FAIL

# The board with nothing plugged in is the case this must not punish. It reaches
# the same place, and the only cost that matters is that it ends.
ticks=$(cat "$work/ticks")
[ "$ticks" -le 20 ] \
    && note "an unattached board ends the watch quickly ($ticks tick(s))" OK \
    || note "it spent $ticks tick(s), so the watch is effectively unbounded" FAIL

echo
echo "===== a link the host took at the wrong speed is not a link ====="

# This is the case the first version of the watch walked straight past. A board
# that comes up full-speed does reach "configured": the host addresses it, reads
# the descriptors and selects the configuration. What it then cannot do is carry
# the gadget. At full speed the periodic bandwidth in a frame does not hold three
# HID interrupt endpoints alongside the console, the disk and the speaker, so the
# host schedules what fits and silently stops polling the rest.
#
# On 2026-08-19 that left a working absolute mouse and a dead keyboard, and a
# watch that had already returned success.

reset configured full-speed 0
TRIES=2 run
grep -q 'rebinding' "$work/out" \
    && note "a full-speed link is rebound" OK \
    || note "a full-speed link was accepted, which is the fault this misses" FAIL

got=$(grep -c 'rebinding' "$work/out")
[ "$got" -eq 2 ] \
    && note "a full-speed link is rebound exactly twice, then left alone" OK \
    || note "it rebound $got time(s), want 2" FAIL

# A host that really is full-speed only is a board that works as well as it can.
# It must settle, not rebind for the rest of the uptime.
grep -q 'rc=1' "$work/out" \
    && note "a link that stays full-speed is reported, not retried forever" OK \
    || note "a permanently full-speed link was reported as healthy" FAIL

grep -q 'full-speed' "$work/out" \
    && note "the console line names the speed it settled at" OK \
    || note "the console does not say which of the two faults happened" FAIL

# Rebinding is the recovery. Measured on the device: 60 unbind/rebind cycles all
# came up high-speed. So a link that improves must be taken, and must not burn
# the rest of the bound.
reset configured full-speed 3
TRIES=2 run
grep -q 'rc=0' "$work/out" \
    && note "a link that improves to high-speed succeeds" OK \
    || note "a link that recovered was still reported as a failure" FAIL
[ "$(grep -c 'rebinding' "$work/out")" -le 1 ] \
    && note "it stops rebinding as soon as the speed is right" OK \
    || note "it kept rebinding a link that had already recovered" FAIL

# The reverse guard. A good link must never be rebound for its speed.
reset configured high-speed 0
run
[ "$(wc -l < "$work/binds")" -eq 0 ] \
    && note "a high-speed link is never rebound for its speed" OK \
    || note "a healthy high-speed link was rebound" FAIL

echo
echo "===== the boot path starts the watch without waiting for it ====="

# rcS waits on this script. A watch that runs inline would add its whole budget
# to every boot of a board that has no host attached.
start_body=$(sed -n '/^  start)/,/^  ;;/p' "$S03")
printf '%s\n' "$start_body" | grep -q 'usb_watch_enumeration' \
    && note "start) runs the watch" OK \
    || note "start) never runs the watch" FAIL
printf '%s\n' "$start_body" | grep -qE 'usb_watch_enumeration[[:space:]]*&' \
    && note "start) backgrounds it, so rcS is not delayed" OK \
    || note "start) runs the watch inline and delays the boot" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "===== the boot notices a gadget the host never took ====="
else
    echo "$fails case(s) FAILED"
fi
exit "$fails"
