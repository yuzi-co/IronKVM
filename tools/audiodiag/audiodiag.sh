#!/bin/sh
# Say whether USB audio capture works, and name the part that fails when it
# does not.
#
#   audiodiag.sh [seconds]
#
# Audio on this board does not come from HDMI. The HDMI audio lines are not
# wired, so the only source is the UAC1 gadget: the managed host plays into the
# KVM as if the KVM were a sound card, the samples arrive on the gadget's
# capture stream, and the server reads them with arecord and sends them to the
# browser.
#
# That path has three owners and only one of them is this board. The host has
# to load a USB audio driver, select a streaming alternate setting, and then
# actually play something. When any of those does not happen the symptom is the
# same on this end: arecord blocks, waits out ALSA's ten second read timeout,
# and exits with "read error: I/O error". The server restarts it and the log
# fills with a failure that says nothing about which end is at fault.
#
# This script separates the ends. It checks the gadget, then it looks at
# whether frames actually move, and it reports digital silence apart from no
# stream at all - a host that streams silence and a host that never opened the
# stream are different problems with different fixes.
#
# Exit codes:
#   0  audio arrives and carries signal
#   1  the KVM side is misconfigured; the host was never reached
#   2  the KVM side is correct, and the host is not streaming
#   3  the host streams, and every sample is silence

SECONDS_TO_CAPTURE=${1:-3}

GADGET=/sys/kernel/config/usb_gadget/g0
UAC1="$GADGET/functions/uac1.usb0"
CARD=UAC1Gadget
STATUS=/proc/asound/$CARD/pcm0c/sub0/status
CAPTURE_DEVICE="hw:$CARD,0"

WORK=$(mktemp -d 2>/dev/null || echo /tmp/audiodiag.$$)
mkdir -p "$WORK"
trap 'rm -rf "$WORK"' EXIT INT TERM

note() { printf '  %-52s %s\n' "$1" "$2"; }

verdict() {
    echo
    echo "VERDICT: $1"
    shift
    for line in "$@"; do
        echo "  $line"
    done
}

echo "NanoKVM USB audio diagnostic"
echo

# ---------------------------------------------------------------- KVM side ---
echo "gadget:"

if [ ! -d "$UAC1" ]; then
    note "uac1.usb0 function" "MISSING"
    verdict "KVM side is misconfigured - no audio function." \
        "The gadget has no uac1.usb0 function, so the host sees no sound card." \
        "S03usbdev creates it. Check that it ran, and check its log."
    exit 1
fi
note "uac1.usb0 function" "present"

# c_chmask is the capture channel mask: what the host may play into us. A zero
# mask is a function that carries no capture stream at all, which looks like a
# working gadget right up to the point where nothing can ever arrive.
c_chmask=$(cat "$UAC1/c_chmask" 2>/dev/null)
c_srate=$(cat "$UAC1/c_srate" 2>/dev/null)
c_ssize=$(cat "$UAC1/c_ssize" 2>/dev/null)
note "capture chmask / rate / sample size" "$c_chmask / $c_srate / $c_ssize"

if [ "$c_chmask" = "0" ] || [ -z "$c_chmask" ]; then
    verdict "KVM side is misconfigured - capture is disabled." \
        "c_chmask is $c_chmask, so the function offers the host no capture stream."
    exit 1
fi

# A function that exists but is not linked into a configuration is never
# enumerated, and the host cannot see it.
linked=no
for cfg in "$GADGET"/configs/*/uac1.usb0; do
    [ -e "$cfg" ] && linked=yes
done
note "linked into a configuration" "$linked"
[ "$linked" = yes ] || {
    verdict "KVM side is misconfigured - function is not in any configuration." \
        "Link it into configs/c.1 and rebind the UDC."
    exit 1
}

udc=$(cat "$GADGET/UDC" 2>/dev/null)
note "UDC bound" "${udc:-NONE}"
[ -n "$udc" ] || {
    verdict "KVM side is misconfigured - the gadget is not bound." \
        "Nothing is enumerated while UDC is empty."
    exit 1
}

echo
echo "alsa:"

command -v arecord >/dev/null 2>&1 || {
    note "arecord" "MISSING"
    verdict "KVM side is misconfigured - arecord is not installed." \
        "The server needs arecord to read the gadget."
    exit 1
}
note "arecord" "present"

[ -d "/proc/asound/$CARD" ] || {
    note "card $CARD" "MISSING"
    verdict "KVM side is misconfigured - the gadget card did not register." \
        "The function is linked, so this is a driver fault rather than a config one."
    exit 1
}
note "card $CARD" "present"

# ------------------------------------------------------------- flow check ---
# Two ways to answer the same question. The server holds the PCM open whenever
# a viewer is listening, and a second reader cannot open it, so ask the kernel
# instead of competing for the device.
echo
echo "flow:"

owner=$(sed -n 's/^owner_pid *: *//p' "$STATUS" 2>/dev/null)
state=$(sed -n 's/^state: *//p' "$STATUS" 2>/dev/null)
note "pcm state" "${state:-closed}"

read_hw_ptr() { sed -n 's/^hw_ptr *: *//p' "$STATUS" 2>/dev/null; }

frames_moved=unknown
captured=""

if [ -n "$owner" ] && [ "$owner" != "0" ]; then
    note "pcm owner" "pid $owner (another reader holds it)"

    # Passive path. hw_ptr counts frames the driver has taken from the host. It
    # is the ground truth for "is the host streaming", and it costs nothing.
    first=$(read_hw_ptr)
    sleep "$SECONDS_TO_CAPTURE"
    second=$(read_hw_ptr)
    note "hw_ptr over ${SECONDS_TO_CAPTURE}s" "${first:-none} -> ${second:-none}"

    if [ -n "$first" ] && [ -n "$second" ] && [ "$second" != "$first" ]; then
        frames_moved=yes
    else
        frames_moved=no
    fi
else
    note "pcm owner" "free"

    # Active path. Bound the run: with no host stream arecord sits in pcm_read
    # until ALSA's own ten second timeout, which is far longer than the window
    # this script promises.
    limit=$((SECONDS_TO_CAPTURE + 3))
    raw="$WORK/capture.raw"
    err="$WORK/capture.err"

    if command -v timeout >/dev/null 2>&1; then
        timeout "$limit" arecord -D "$CAPTURE_DEVICE" -f S16_LE -r 48000 -c 2 \
            -t raw -d "$SECONDS_TO_CAPTURE" >"$raw" 2>"$err"
    else
        arecord -D "$CAPTURE_DEVICE" -f S16_LE -r 48000 -c 2 \
            -t raw -d "$SECONDS_TO_CAPTURE" >"$raw" 2>"$err"
    fi

    bytes=$(wc -c <"$raw" 2>/dev/null | tr -d ' ')
    [ -n "$bytes" ] || bytes=0
    expected=$((SECONDS_TO_CAPTURE * 48000 * 4))
    note "captured bytes" "$bytes of $expected expected"

    reason=$(grep -v '^$' "$err" 2>/dev/null | tail -1)
    [ -n "$reason" ] && note "arecord said" "$reason"

    if [ "$bytes" -gt 0 ]; then
        frames_moved=yes
        captured="$raw"
    else
        frames_moved=no
    fi
fi

if [ "$frames_moved" = no ]; then
    verdict "The KVM is correct. The host is not streaming audio." \
        "The gadget is bound and the card is registered, and no frame ever arrives." \
        "The host has to bind a USB audio driver and open the stream. Check on it:" \
        "  for i in /sys/bus/usb/devices/*:1.*; do \\" \
        "    printf '%s class=%s driver=%s\\n' \"\$i\" \\" \
        "      \"\$(cat \$i/bInterfaceClass)\" \"\$(basename \$(readlink \$i/driver))\"; done" \
        "An audio interface is class 01. If its driver is empty, the host has no" \
        "snd-usb-audio, and no change on the KVM can make audio work."
    exit 2
fi

# --------------------------------------------------------------- signal ------
# Frames move, so the host streams. Silence and music are both valid streams,
# and only one of them means the user hears anything.
echo
echo "signal:"

if [ -z "$captured" ]; then
    verdict "Audio is flowing. The host streams into the gadget." \
        "Another reader owns the PCM, so this run measured frames rather than samples." \
        "Stop the listener and run again to test for silence against signal."
    exit 0
fi

total=$(wc -c <"$captured" | tr -d ' ')
# Digital silence is every byte zero, so counting what survives deleting NUL
# separates silence from signal without arithmetic over 500k samples.
nonzero=$(tr -d '\000' <"$captured" | wc -c | tr -d ' ')
note "non-zero bytes" "$nonzero of $total"

if [ "$nonzero" = "0" ]; then
    verdict "The path works, and the host is sending digital silence." \
        "Every sample is zero. The USB side is proven end to end." \
        "Play something on the host and run this again."
    exit 3
fi

# Report a rough level when od can produce signed decimals. Not every busybox
# builds that format in, so treat it as a bonus rather than a step.
peak=""
if od -An -td2 -v -N 8 "$captured" >/dev/null 2>&1; then
    peak=$(od -An -td2 -v -N 200000 "$captured" 2>/dev/null | awk '
        { for (i = 1; i <= NF; i++) { v = $i < 0 ? -$i : $i; if (v > max) max = v } }
        END { printf "%d", max }')
fi
[ -n "$peak" ] && note "peak sample (of 32767)" "$peak"

verdict "AUDIO WORKS. Real samples arrive from the host." \
    "$nonzero of $total bytes carry signal."
exit 0
