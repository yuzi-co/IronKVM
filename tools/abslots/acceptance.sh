#!/bin/sh
# Build the deliberately broken slot the acceptance test needs, and check the
# board afterwards.
#
#   acceptance.sh build <mute|hang> <base.tar.zst> <payload-dir> <size-mib> <out.img>
#   acceptance.sh check                                  (run this on the board)
#
# The design is not delivered until a slot that cannot come up returns the
# board to the trusted slot by itself, on hardware, with no card removal. This
# script builds that slot and reads the evidence back.
#
# `build` composes the manifest rather than keeping a second copy of it. The
# broken slot must be the real root plus exactly one fault, or the test proves
# something about a different image.
#
# Two variants, and the difference is what rescue door survives the fault:
#
#   mute  removes S50sshd and S95nanokvm. The slot boots all the way, so init
#         starts a getty and the USB console still answers, and it has no ssh
#         and no web. Run this one when nobody can reach the board.
#
#   hang  adds a script that never returns, before S50sshd. rcS blocks, so init
#         never reaches the getty either, and the USB console is dead with the
#         rest. REQUIRES SOMEONE ABLE TO POWER-CYCLE THE BOARD. If the
#         watchdog's reboot -f does not take, nothing else can reach it.
#
# Both reach the identical watchdog decision: carrier and an address, no
# listener, running slot is not the trusted slot, so reboot into the trusted
# slot without setting the recovery marker. mute proves that decision. hang
# additionally proves the watcher survives an rcS that never finishes, which is
# the stronger claim and the more expensive one to get wrong.
set -e

HERE=$(dirname "$0")

cmd_build() {
    VARIANT=${1:?usage: acceptance.sh build <mute|hang> <base> <payload> <size-mib> <out.img>}
    BASE=${2:?usage: acceptance.sh build <mute|hang> <base> <payload> <size-mib> <out.img>}
    PAYLOAD=${3:?usage: acceptance.sh build <mute|hang> <base> <payload> <size-mib> <out.img>}
    SIZE=${4:?usage: acceptance.sh build <mute|hang> <base> <payload> <size-mib> <out.img>}
    OUT=${5:?usage: acceptance.sh build <mute|hang> <base> <payload> <size-mib> <out.img>}

    M=$(mktemp)
    trap 'rm -f "$M"' EXIT

    case "$VARIANT" in
        mute)
            # The add lines have to go, not just be followed by a remove.
            # build-image.sh runs every remove before every add on purpose, so
            # a remove of something the manifest also adds is undone in the
            # same build. Its own gate caught this.
            grep -v '[[:space:]]/etc/init\.d/S95nanokvm' "$HERE/manifest/root.manifest" > "$M"
            printf '\nremove  /etc/init.d/S50sshd\nremove  /etc/init.d/S95nanokvm\n' >> "$M"
            echo "acceptance: root.manifest minus both doors, into $OUT"
            ;;
        hang)
            cat "$HERE/manifest/root.manifest" > "$M"
            printf '\nadd     tools/abslots/device/S45hang              /etc/init.d/S45hang          0755\n' >> "$M"
            echo "acceptance: root.manifest plus one hang, into $OUT"
            echo "acceptance: this variant kills the USB console too; do not run it unattended"
            ;;
        *)
            echo "acceptance: unknown variant '$VARIANT', want mute or hang" >&2
            exit 1
            ;;
    esac

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
