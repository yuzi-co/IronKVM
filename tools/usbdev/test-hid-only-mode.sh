#!/bin/sh
# S03usbdev hands the boot to S03usbhid when the owner chose HID-only mode.
#
#   test-hid-only-mode.sh [path-to-S03usbdev]
#
# The web UI chose HID-only mode by copying S03usbhid over
# /etc/init.d/S03usbdev. That file is on the root slot, so the next image put
# the normal script back and the board came up in normal mode. The server now
# also writes /etc/kvm/hid_only, which is on /data, and S03usbdev hands every
# action to S03usbhid while that file exists.
#
# No case here may build a gadget. On the reference board a real start rebuilds
# the USB gadget under the running server. So every case passes an action the
# case statement does not know, which the normal script ignores, and the HID-only
# script is a stub that records what it was handed.
S03=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S03usbdev}
[ -f "$S03" ] || { echo "no S03usbdev at $S03"; exit 2; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
t() { if [ "$2" = 0 ]; then echo "ok   - $1"; else echo "FAIL - $1"; fails=1; fi; }

cat > "$WORK/S03usbhid" <<STUB
#!/bin/sh
echo "\$*" >> "$WORK/calls"
exit 0
STUB
chmod +x "$WORK/S03usbhid"

ACTION=test-no-such-action

run() {
    rm -f "$WORK/calls"
    HID_ONLY_FILE="$WORK/hid_only" HID_ONLY_SCRIPT="$WORK/S03usbhid" \
        sh "$S03" "$@" > "$WORK/out" 2>&1
}

run "$ACTION"
[ ! -e "$WORK/calls" ]
t "with no marker the normal script keeps the action" $?

touch "$WORK/hid_only"
run "$ACTION" extra
[ "$(cat "$WORK/calls" 2>/dev/null)" = "$ACTION extra" ]
t "with the marker every argument goes to S03usbhid" $?

rm -f "$WORK/calls"
S03USBDEV_DELEGATED=1 HID_ONLY_FILE="$WORK/hid_only" HID_ONLY_SCRIPT="$WORK/S03usbhid" \
    sh "$S03" "$ACTION" > "$WORK/out" 2>&1
[ ! -e "$WORK/calls" ]
t "a script that was already handed over does not hand over again" $?

rm -f "$WORK/S03usbhid"
run "$ACTION"
grep -q "S03usbhid" "$WORK/out"
t "with the marker and no S03usbhid it says so and stays normal" $?

exit $fails
