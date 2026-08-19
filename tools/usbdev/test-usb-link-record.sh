#!/bin/sh
# Check that every boot leaves a record of the speed the USB link negotiated.
#
#   test-usb-link-record.sh [path-to-init.d-dir]
#
# The board can boot with the gadget bound, all three /dev/hidg* present, every
# descriptor correct, and the bus running at full speed instead of high speed.
# At full speed the periodic bandwidth in a frame does not carry three HID
# interrupt endpoints plus the console, the disk and the speaker, so the host
# silently stops polling some of them. The keyboard is then dead and nothing
# anywhere reports a fault.
#
# The kernel does report the speed, once, at reset time. It is the only report
# there is, and it does not survive: this board writes a VPSS line to the ring
# buffer every second once the video pipeline runs, and the ring holds about
# forty minutes. Measured on 2026-08-19, forty minutes after boot, the ring had
# no dwc2 line left in it at all.
#
# So the failure is only ever seen live, by a person who notices the keyboard is
# dead and looks within the hour. That is why it has been characterised six
# times and diagnosed none: every investigation begins after its evidence has
# gone. This copies the speed out of the ring while it is still there.
#
# A good boot must cost one line. The log is on the SD card, and a board that
# reboots in a loop must not fill it.
set -u

DIR=$(dirname "$0")
INITD=${1:-$DIR/../../kvmapp/system/init.d}
S03=$INITD/S03usbdev
[ -f "$S03" ] || { echo "usage: test-usb-link-record.sh <init.d dir>"; exit 1; }

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

UDC=/sys/class/udc/4340000.usb

# reset <state> <speed> <configured-after-N-sleeps>
reset() {
    rm -rf "$work/sys" "$work/log"
    mkdir -p "$work$UDC" "$work/sys/kernel/config/usb_gadget/g0" "$work/log"
    printf '%s\n' "$1" > "$work$UDC/state"
    printf '%s\n' "$2" > "$work$UDC/current_speed"
    : > "$work/sys/kernel/config/usb_gadget/g0/UDC"
    echo "$3" > "$work/flip_at"
    : > "$work/binds"
}

lift() {
    cat > "$work/f.sh" <<STUB
sleep() {
    n=\$(cat "$work/ticks" 2>/dev/null || echo 0)
    n=\$((n + 1))
    echo "\$n" > "$work/ticks"
    if [ "\$n" -ge "\$(cat "$work/flip_at")" ] && [ "\$(cat "$work/flip_at")" -gt 0 ]; then
        echo configured > "$work$UDC/state"
    fi
}

ls() {
    if [ "\${1:-}" = "$work/sys/class/udc/" ]; then
        echo x >> "$work/binds"
        echo 4340000.usb
        return
    fi
    command ls "\$@" 2>/dev/null
}

# The ring buffer this reads from is the thing that disappears. Two dwc2 lines
# and one line of the video spam that evicts them.
dmesg() {
    echo "[    3.824601] dwc2 4340000.usb: bound driver configfs-gadget"
    echo "[    3.992529] dwc2 4340000.usb: new device is full-speed"
    echo "[ 2377.803386] base_get_chn_buffer:262(): Mod(VPSS) Grp(0) Chn(0)"
}
STUB
    sed -n '/^[a-z_][a-z_]*() *{$/,/^}$/p' "$S03" \
        | sed "s|/sys/|$work/sys/|g; s|\. /etc/profile|:|" >> "$work/f.sh"
}

lift
for fn in usb_record_link usb_link_speed usb_link_state; do
    grep -q "^$fn() {" "$work/f.sh" \
        && note "S03usbdev defines $fn" OK \
        || note "S03usbdev defines $fn" FAIL
done
[ "$fails" -eq 0 ] || { echo; echo "$fails case(s) FAILED"; exit "$fails"; }

LOG=$work/log/usb-link.log

# record <rebinds> - runs usb_record_link alone against the fake controller.
record() {
    (
        USB_LINK_LOG=$LOG
        USB_LINK_KEEP=${KEEP:-200}
        export USB_LINK_LOG USB_LINK_KEEP
        . "$work/f.sh"
        usb_record_link "${1:-0}"
        echo "rc=$?"
    ) > "$work/out" 2>&1
}

# run - runs the whole watch, which must record whatever it settles on.
run() {
    : > "$work/ticks"
    (
        USB_ENUM_TRIES=${TRIES:-2}
        USB_ENUM_WAIT=${WAIT:-5}
        USB_LINK_LOG=$LOG
        USB_LINK_KEEP=${KEEP:-200}
        export USB_ENUM_TRIES USB_ENUM_WAIT USB_LINK_LOG USB_LINK_KEEP
        . "$work/f.sh"
        usb_watch_enumeration
        echo "rc=$?"
    ) > "$work/out" 2>&1
}

echo "===== a good boot is recorded, and costs one line ====="

reset configured high-speed 0
record 0
[ -f "$LOG" ] \
    && note "a good link is recorded at all" OK \
    || note "a good link left no record, so a later boot has nothing to compare" FAIL
[ "$(wc -l < "$LOG" 2>/dev/null || echo 0)" -eq 1 ] \
    && note "a good link costs exactly one line" OK \
    || note "a good link wrote $(wc -l < "$LOG" 2>/dev/null) lines onto the SD card" FAIL
grep -q 'speed=high-speed' "$LOG" \
    && note "the record names the negotiated speed" OK \
    || note "the record does not name the speed, which is the question it exists for" FAIL
grep -q 'state=configured' "$LOG" \
    && note "the record names the controller state" OK \
    || note "the record does not name the controller state" FAIL
grep -q 'dwc2' "$LOG" \
    && note "a good link copied the ring buffer out for nothing" FAIL \
    || note "a good link does not copy the ring buffer" OK

echo
echo "===== a boot that came up full-speed takes the ring with it ====="

# This is the whole point. These lines are gone within the hour.
reset configured full-speed 0
record 0
grep -q 'speed=full-speed' "$LOG" \
    && note "the record names full-speed" OK \
    || note "the record does not name the speed it came up at" FAIL
grep -q 'new device is full-speed' "$LOG" \
    && note "a bad link copies the dwc2 lines before the ring wraps" OK \
    || note "a bad link left the evidence in a ring that wraps in forty minutes" FAIL
grep -q 'bound driver configfs-gadget' "$LOG" \
    && note "it copies every dwc2 line, not just the speed" OK \
    || note "it copied only part of the dwc2 output" FAIL
grep -q 'VPSS' "$LOG" \
    && note "it copied the video spam that evicts the useful lines" FAIL \
    || note "it copies the dwc2 lines only" OK

echo
echo "===== a gadget the host never took is recorded too ====="

reset "not attached" unknown 0
record 2
grep -q 'state=not attached' "$LOG" \
    && note "a gadget the host never took is recorded" OK \
    || note "the case that has no other symptom went unrecorded" FAIL
grep -q 'rebinds=2' "$LOG" \
    && note "the record says how many rebinds it took" OK \
    || note "the record does not say whether the watch had to intervene" FAIL
grep -q 'new device is full-speed' "$LOG" \
    && note "an unenumerated gadget takes the ring with it" OK \
    || note "an unenumerated gadget left no kernel output" FAIL

echo
echo "===== the log is bounded, because it lives on the SD card ====="

reset configured high-speed 0
KEEP=5
i=0
while [ "$i" -lt 12 ]; do i=$((i + 1)); KEEP=5 record 0; done
lines=$(wc -l < "$LOG" 2>/dev/null || echo 0)
[ "$lines" -le 5 ] \
    && note "12 boots at a keep of 5 leave $lines line(s)" OK \
    || note "12 boots left $lines lines, so the file grows without bound" FAIL
[ "$lines" -gt 0 ] \
    && note "trimming keeps the newest records rather than emptying the file" OK \
    || note "the trim emptied the file" FAIL
unset KEEP

echo
echo "===== a board with no /data writes nothing and does not fail ====="

# This script runs early. /data is mounted by S01fs, and a board whose data
# partition is absent or broken must still boot.
reset configured high-speed 0
(
    USB_LINK_LOG=$work/absent/usb-link.log
    export USB_LINK_LOG
    . "$work/f.sh"
    usb_record_link 0
    echo "rc=$?"
) > "$work/out" 2>&1
grep -q 'rc=0' "$work/out" \
    && note "a missing /data is not an error" OK \
    || note "a missing /data made the record fail" FAIL
[ -f "$work/absent/usb-link.log" ] \
    && note "it created a log outside a directory it was given" FAIL \
    || note "it writes nothing when the directory is absent" OK

echo
echo "===== the watch records whatever it settles on ====="

reset configured high-speed 0
run
grep -q 'rc=0' "$work/out" \
    && note "a configured gadget still returns success" OK \
    || note "recording changed the watch's return code" FAIL
grep -q 'speed=high-speed' "$LOG" \
    && note "the success path records" OK \
    || note "a boot that worked left no record" FAIL

reset "not attached" full-speed 0
TRIES=2 run
grep -q 'rc=1' "$work/out" \
    && note "a gadget the host never took still returns failure" OK \
    || note "recording changed the watch's return code on the give-up path" FAIL
grep -q 'rebinds=2' "$LOG" \
    && note "the give-up path records, and says it rebound twice" OK \
    || note "the give-up path left no record" FAIL
[ "$(grep -c 'speed=' "$LOG")" -eq 1 ] \
    && note "the watch records once, not once per rebind" OK \
    || note "the watch wrote $(grep -c 'speed=' "$LOG") records for one boot" FAIL

echo
echo "===== the shipped default points at /data ====="

# The overrides above exist for these cases. They would happily pass against a
# script whose real default wrote to /kvmapp, which is the boot SD card and is
# not where runtime state goes.
grep -q 'USB_LINK_LOG:-/data/usb-link.log' "$S03" \
    && note "the default log lives under /data" OK \
    || note "the default log is not under /data" FAIL

# The default has to live inside the function. test-usb-enumeration.sh lifts
# function bodies only, so a default written as a top-level assignment is absent
# when that suite sources the result, and `set -u` then kills the subshell and
# takes the watch's return code with it.
sed -n '/^usb_record_link() {/,/^}/p' "$S03" | grep -q 'USB_LINK_LOG:-' \
    && note "the default is resolved inside the function that is lifted" OK \
    || note "the default sits outside the function, so a lift loses it" FAIL
# Comments first. A comment that explains why runtime state stays off the boot
# card names the boot card, and a locator that reads prose fails a script for
# saying the right thing.
grep -v '^[[:space:]]*#' "$S03" | grep -q '/kvmapp' \
    && note "the script writes to /kvmapp, which is the boot card" FAIL \
    || note "it never writes to the boot card" OK

echo
if [ "$fails" -eq 0 ]; then
    echo "===== every boot leaves the speed it negotiated ====="
else
    echo "$fails case(s) FAILED"
fi
exit "$fails"
