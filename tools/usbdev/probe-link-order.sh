#!/bin/sh
# Drive the boot-time bind/role sequence on demand and count how often the bus
# comes up at full speed.
#
#   probe-link-order.sh [iterations]        # default 60
#
# RUN THIS ON THE DEVICE. It takes the USB link down and puts it back about once
# every eight seconds, so the managed host loses its keyboard and mice for the
# duration. It restores a known good gadget through S03usbdev stop_start before
# it exits, including when it is interrupted.
#
# ---------------------------------------------------------------------------
# What it was written to settle
#
# S03usbdev binds the gadget to the controller and then writes the OTG role:
#
#     usb_bind
#     echo device > /proc/cviusb/otg_role
#
# That order looks wrong, and the reasoning behind the suspicion was this. The
# role write is not a note of intent. drivers/usb/dwc2/platform.c gives
# /proc/cviusb/otg_role a write handler that calls dwc2_set_hw_id(), which pokes
# the SoC register at 0x03000048 and flips the hardware ID pin. dwc2 answers an
# ID change with GINTSTS_CONIDSTSCHNG, and the handler cannot do the work in
# interrupt context, so it queues wf_otg. dwc2_conn_id_status_change() then runs
# from a workqueue: it polls for device mode with msleep(20), does a full
# dwc2_core_init(), and ends with dwc2_hsotg_core_connect().
#
# The controller also starts every boot on the wrong side of that switch.
# platform.c initialises cviusb.id_override to 0, and 0 is "host", so from probe
# until S03usbdev runs the board is a USB host with a root hub registered.
#
# And the bind before the role write does not raise the pull-up at all.
# dwc2_hsotg_pullup() returns early with a comment saying it will not modify the
# pullup state while in host mode, so the UDC write only records
# hsotg->enabled = 1. The pull-up the managed host actually sees comes from
# dwc2_hsotg_core_connect() at the end of that workqueue item.
#
# So the whole enumeration hangs off an asynchronous core reset whose timing is
# a workqueue schedule plus a poll loop, and a core reset landing inside the
# host's reset window would lose the high-speed chirp and fall back to full
# speed. That is a plausible account of a failure seen about once in six boots.
#
# It is also wrong. This script drove that exact sequence 60 times, and swept
# the gap between the bind and the role write across 0, 50, 100, 150, 200, 250
# and 300 ms. Every one of the 67 came up high-speed, configured and addressed.
# At a 1-in-6 rate, 60 clean iterations have a probability of about 2 in 100000.
#
# The ordering in S03usbdev is therefore not the cause, and reordering it is not
# a fix. Setting the role first is measurably worse: dwc2_hsotg_core_connect()
# is unconditional, so the pull-up goes up before the gadget is bound and the
# host is shown a device that answers nothing. That costs an extra bus reset,
# which this script reported as resets=3 against resets=2.
#
# What remains is boot-specific state that a running board cannot reproduce: the
# controller having only just probed, and the managed host's own port state when
# the KVM appears. Deciding between those needs records from real boots, which
# is what usb_record_link in S03usbdev now collects.
set -u

G=/sys/kernel/config/usb_gadget/g0
UDC=$(ls /sys/class/udc/ | head -1)
S=/sys/class/udc/$UDC
N=${1:-60}
SETTLE=${SETTLE:-4}

[ -d "$G" ] || { echo "no gadget at $G, run this on the device"; exit 1; }
[ -n "$UDC" ] || { echo "no device controller"; exit 1; }

restore() {
    /etc/init.d/S03usbdev stop_start >/dev/null 2>&1
    sleep 6
    printf 'restored  state=%s speed=%s hid=%s\n' \
        "$(cat "$S/state")" "$(cat "$S/current_speed")" \
        "$(ls /dev/hidg* 2>/dev/null | tr '\n' ' ')"
}
trap 'echo interrupted; restore; exit 130' INT TERM

# cycle <gap-ms> puts the controller back where the driver leaves it at probe,
# then runs the sequence S03usbdev runs, with <gap-ms> between the two steps.
cycle() {
    dmesg -c >/dev/null 2>&1

    echo "" > "$G/UDC" 2>/dev/null
    sleep 1
    echo host > /proc/cviusb/otg_role
    sleep 2

    echo "$UDC" > "$G/UDC" 2>/dev/null
    usleep $(($1 * 1000))
    echo device > /proc/cviusb/otg_role
    sleep "$SETTLE"
}

# verdict prints one line and returns 1 when the link came up wrong.
verdict() {
    speed=$(cat "$S/current_speed")
    state=$(cat "$S/state")
    fs=$(dmesg | grep -c 'new device is full-speed')
    resets=$(dmesg | grep -c 'new device is')
    addr=$(dmesg | grep -c 'new address')

    if [ "$speed" != high-speed ] || [ "$state" != configured ] || [ "$fs" != 0 ]
    then
        printf 'FAIL %-10s state=%s speed=%s full-speed-resets=%s\n' "$1" "$state" "$speed" "$fs"
        dmesg | grep dwc2
        return 1
    fi

    printf '     %-10s state=%s speed=%s resets=%s addressed=%s\n' \
        "$1" "$state" "$speed" "$resets" "$addr"
    return 0
}

echo "udc=$UDC  iterations=$N  settle=${SETTLE}s"
echo
echo "--- gap between the bind and the role write ---"
bad=0
for gap in 0 50 100 150 200 250 300
do
    cycle "$gap"
    verdict "+${gap}ms" || bad=$((bad + 1))
done

echo
echo "--- $N iterations of the shipped sequence ---"
# Only failures are printed. Sixty passing lines say nothing that the count at
# the end does not, and they bury the one line that matters.
out=$(mktemp)
trap 'rm -f "$out"' EXIT

i=0
while [ "$i" -lt "$N" ]
do
    i=$((i + 1))
    cycle 0
    verdict "$i" > "$out" 2>&1 || { cat "$out"; bad=$((bad + 1)); }
    [ $((i % 10)) -eq 0 ] && echo "... $i/$N done, $bad bad"
done

echo
echo "===== $bad bad of $((N + 7)) ====="
restore
[ "$bad" -eq 0 ]
