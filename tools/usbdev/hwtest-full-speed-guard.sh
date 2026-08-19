#!/bin/sh
# Prove the boot guard against a real full-speed link, on the device.
#
#   hwtest-full-speed-guard.sh
#
# RUN THIS ON THE DEVICE. It has an hwtest- prefix rather than test- because it
# needs the hardware: it creates the fault on the real controller, so the
# workstation suites must not try to run it.
#
# DESTRUCTIVE while it runs. It puts the managed host's keyboard and mice at
# full speed on purpose, which is the fault, and restores through S03usbdev
# stop_start on exit including on interrupt.
#
# ---------------------------------------------------------------------------
# Why this exists
#
# test-usb-enumeration.sh drives usb_watch_enumeration against a directory of
# files under mktemp. That is the right way to test the decision, and it is not
# evidence about the board: every controller state in it is a string somebody
# wrote to a file. The guard was written for a fault nobody could produce, so
# nothing had ever run it against the fault.
#
# It can be produced. dwc2 chirps at whatever DCFG.DevSpd holds when the host
# drives the bus reset, and the reset handler rewrites DCFG on every reset
# except the first:
#
#     u32 connected = hsotg->connected;      /* captured before the disconnect */
#     ...
#     if (usb_status & GOTGCTL_BSESVLD && connected)
#             dwc2_hsotg_core_init_disconnected(hsotg, true);
#
# hsotg->connected is set at ENUMDONE, so after a fresh pull-up it is zero and
# the first reset leaves DCFG alone. Bind the gadget, write DCFG.DevSpd while
# the host is still debouncing the connect, and the chirp that follows is a
# full-speed one. Measured: the link comes up full-speed, configured, with all
# three /dev/hidg* present, which is the shape of the 2026-08-19 fault.
#
# The first assertion below is the one that matters most. It shows the ORIGINAL
# success condition, "state reads configured", answering yes on this board. That
# guard looked straight at the fault it was written for and reported health.
set -u

S03=${1:-/etc/init.d/S03usbdev}
DCFG=0x04340800   # dwc2 device configuration, DevSpd is bits [1:0]
G=/sys/kernel/config/usb_gadget/g0
UDC=$(ls /sys/class/udc/ 2>/dev/null | head -1)
S=/sys/class/udc/$UDC
GAP_MS=${GAP_MS:-30}

[ -f "$S03" ] || { echo "not on the device: $S03 is missing"; exit 1; }
[ -n "$UDC" ] || { echo "no device controller"; exit 1; }
command -v devmem >/dev/null || { echo "devmem is missing"; exit 1; }

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

restore() {
    /etc/init.d/S03usbdev stop_start >/dev/null 2>&1
    sleep 6
}
trap 'echo; echo interrupted; restore; exit 130' INT TERM

# The shipped script, sourced whole. Not a lifted copy: the functions under test
# are the ones the board boots. A positional argument is set first because the
# file ends in a case statement over "$1".
set -- noop
# shellcheck disable=SC1090
. "$S03"

speed() { cat "$S/current_speed"; }
state() { cat "$S/state"; }

force_full_speed() {
    echo "" > $G/UDC 2>/dev/null
    sleep 2
    echo "$UDC" > $G/UDC 2>/dev/null
    usleep $((GAP_MS * 1000))

    cur=$(devmem $DCFG 32)
    devmem $DCFG 32 $(( (cur & ~3) | 1 ))

    sleep 8
}

echo "===== creating the fault ====="
printf '  before: state=%s speed=%s\n' "$(state)" "$(speed)"
force_full_speed
printf '  after:  state=%s speed=%s hid=%s\n' \
    "$(state)" "$(speed)" "$(ls /dev/hidg* 2>/dev/null | tr '\n' ' ')"

if [ "$(speed)" != full-speed ]; then
    echo
    echo "the fault was not created, so nothing below would prove anything"
    restore
    exit 1
fi

echo
echo "===== the original success condition passes the broken board ====="

# usb_enumerated is the whole of what the first version of the guard waited for.
if usb_enumerated; then
    note "usb_enumerated says yes on a board that cannot type" OK
else
    note "usb_enumerated says no, so the fault is not the one this reproduces" FAIL
fi

[ -n "$(ls /dev/hidg* 2>/dev/null)" ] \
    && note "every /dev/hidg* is present, so the device nodes prove nothing" OK \
    || note "the device nodes are missing, which is a different fault" FAIL

echo
echo "===== the current success condition catches it ====="

if usb_link_ok; then
    note "usb_link_ok accepted a full-speed link" FAIL
else
    note "usb_link_ok rejects a full-speed link" OK
fi

case "$(usb_link_fault)" in
    *full-speed*) note "usb_link_fault names the speed: $(usb_link_fault)" OK ;;
    *)            note "usb_link_fault does not name the speed: $(usb_link_fault)" FAIL ;;
esac

echo
echo "===== the guard recovers the link ====="

log=/tmp/hwtest-usb-link.log
: > "$log"

USB_ENUM_TRIES=2 USB_ENUM_WAIT=5 USB_LINK_LOG=$log usb_watch_enumeration
rc=$?
sleep 4

printf '  watch returned %s, link is now state=%s speed=%s\n' "$rc" "$(state)" "$(speed)"

[ "$(speed)" = high-speed ] \
    && note "the link is high-speed again after the guard ran" OK \
    || note "the guard left the link at $(speed)" FAIL

[ "$(state)" = configured ] \
    && note "the host accepted the rebound gadget" OK \
    || note "the gadget is $(state) after the guard ran" FAIL

[ "$rc" -eq 0 ] \
    && note "the guard reports success once the link is good" OK \
    || note "the guard returned $rc after recovering the link" FAIL

n=$(ls /dev/hidg* 2>/dev/null | wc -l)
[ "$n" -eq 3 ] \
    && note "all three HID endpoints are back" OK \
    || note "only $n HID endpoint(s) came back" FAIL

echo
echo "===== the boot record says what happened ====="

grep -q 'speed=' "$log" \
    && note "the boot left a record of the link" OK \
    || note "nothing was recorded, so a later investigation has nothing" FAIL

# The record is written after the watch settles, so it reports the recovered
# link. What must survive is the evidence that the board needed a rebind.
grep -q 'rebinds=[1-9]' "$log" \
    && note "the record says the guard had to intervene" OK \
    || note "the record does not show that a rebind was needed: $(cat "$log")" FAIL

rm -f "$log"
restore

echo
printf 'restored  state=%s speed=%s hid=%s\n' \
    "$(state)" "$(speed)" "$(ls /dev/hidg* 2>/dev/null | tr '\n' ' ')"

echo
if [ "$fails" -eq 0 ]; then
    echo "===== the guard catches and recovers a real full-speed link ====="
else
    echo "$fails case(s) FAILED"
fi
exit "$fails"
