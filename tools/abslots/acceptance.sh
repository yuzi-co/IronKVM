#!/bin/sh
# Build the deliberately broken slot the acceptance test needs, and check the
# board afterwards.
#
#   acceptance.sh build <base.tar.zst> <payload-dir> <size-mib> <out.img>
#   acceptance.sh check                                  (run this on the board)
#
# The design is not delivered until a slot that cannot come up returns the
# board to the trusted slot by itself, on hardware, with no card removal. This
# script builds that slot and reads the evidence back.
#
# `build` composes the manifest rather than keeping a second copy of it. The
# broken slot must be the real root plus exactly one fault, or the test proves
# something about a different image.
set -e

HERE=$(dirname "$0")

cmd_build() {
    BASE=${1:?usage: acceptance.sh build <base.tar.zst> <payload-dir> <size-mib> <out.img>}
    PAYLOAD=${2:?usage: acceptance.sh build <base.tar.zst> <payload-dir> <size-mib> <out.img>}
    SIZE=${3:?usage: acceptance.sh build <base.tar.zst> <payload-dir> <size-mib> <out.img>}
    OUT=${4:?usage: acceptance.sh build <base.tar.zst> <payload-dir> <size-mib> <out.img>}

    M=$(mktemp)
    trap 'rm -f "$M"' EXIT
    cat "$HERE/manifest/root.manifest" > "$M"
    printf '\nadd     tools/abslots/device/S45hang              /etc/init.d/S45hang          0755\n' >> "$M"

    echo "acceptance: root.manifest plus one hang, into $OUT"
    sh "$HERE/build-image.sh" "$BASE" "$M" "$PAYLOAD" "$SIZE" "$OUT"
}

# Read the evidence on the board after the failed trial has reverted it.
#
# Three independent records must agree, because any one of them alone can be
# read the way you were hoping to read it:
#
#   /proc/mounts   which slot is actually running now
#   dmesg          what the initramfs decided, and why
#   /watchdog.log  what the watchdog decided, and after how long
cmd_check() {
    echo "===== where the board ended up ====="
    slot status

    echo
    echo "===== what the initramfs decided ====="
    dmesg 2>/dev/null | grep nanokvm-slot || echo "  (no nanokvm-slot lines in dmesg)"

    echo
    echo "===== what the watchdog decided ====="
    if [ -r /watchdog.log ]; then
        cat /watchdog.log
    else
        echo "  (no /watchdog.log; this boot never needed it)"
    fi

    echo
    echo "===== the trial marker is gone ====="
    if [ -e /boot/slot.try ]; then
        echo "  FAIL: /boot/slot.try survives, so the trial was never disarmed"
    else
        echo "  OK: /boot/slot.try is absent"
    fi

    echo
    echo "===== the board is not stuck in recovery ====="
    if [ -e /boot/recovery ]; then
        echo "  FAIL: /boot/recovery is set, so a failed trial escalated too far"
    else
        echo "  OK: /boot/recovery is absent"
    fi
}

case "$1" in
    build) shift; cmd_build "$@" ;;
    check) cmd_check ;;
    *) echo "Usage: acceptance.sh {build <base> <payload> <size-mib> <out>|check}"; exit 1 ;;
esac
