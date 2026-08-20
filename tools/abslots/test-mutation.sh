#!/bin/sh
# Break the dispatch on purpose and fail if test-dispatch.sh does not notice.
#
#   test-mutation.sh
#
# A guard that stops guarding reads exactly like a guard that holds. Two guards
# in this repository had already rotted that way and kept reporting success:
# tools/vidiag/test-restart-space.sh skipped both of its real cases after the
# code it watched moved into a function, and tools/vidiag/test-restart-waits.sh
# extracted a function that no longer existed and exited before checking any
# behaviour at all.
#
# So the dispatch tests are themselves tested. Each mutation below removes one
# property the design depends on, and this file fails if the suite still passes.
HERE=$(cd "$(dirname "$0")" && pwd)
SEL="$HERE/init-slot-selection.inc"
DIS="$HERE/init-mount-dispatch.inc"
SUITE="$HERE/test-dispatch.sh"

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

# caught <label> <selection> <dispatch>
caught() {
    label=$1
    if sh "$SUITE" "$2" "$3" >/dev/null 2>&1; then
        note "$label" FAIL
    else
        note "$label" OK
    fi
}

echo "===== every mutation is caught ====="

# 1. The trial is never disarmed. This is the exact failure the design exists to
#    prevent: a slot that hangs would be retried on every boot for ever.
sed '/"\$BUSYBOX" rm -f/d' "$SEL" > "$WORK/sel-nodisarm.inc"
caught "a trial that is never disarmed" "$WORK/sel-nodisarm.inc" "$DIS"

# 2. The disarm is attempted but not proved, so a FAT write that does not land
#    turns into an endless retry.
sed '/could not be deleted/,+1d' "$SEL" > "$WORK/sel-noproof.inc"
caught "a disarm that is not proved" "$WORK/sel-noproof.inc" "$DIS"

# 3. Recovery is dropped from the ladder, so a trusted slot that will not mount
#    strands the board at msc, which needs a person and a card reader.
sed '/falling back to recovery/,+1d' "$DIS" > "$WORK/dis-norecovery.inc"
caught "a missing recovery rung" "$SEL" "$WORK/dis-norecovery.inc"

# 4. The trial is preferred even when it will not mount, so there is no
#    fallback to the slot that is known to work.
sed 's/^if \[ -z "${bootdev}" \] \&\& \[ -n "${trial}" \]$/if [ -n "${trial}" ]; then :; fi\nif [ -n "${trial}" ]/' "$DIS" > "$WORK/dis-notrusted.inc"
caught "a trial with no fallback to trusted" "$SEL" "$WORK/dis-notrusted.inc"

# 5. An unrecognised trusted marker no longer falls back to slot A, so a
#    corrupted marker strands the board.
sed '/is not a slot name, using slot A/,+1d' "$DIS" > "$WORK/dis-nodefault.inc"
caught "an unrecognised marker with no default" "$SEL" "$WORK/dis-nodefault.inc"

# 6. The recovery marker stops winning, so an operator who asked for recovery
#    gets whatever the trial or the trusted slot is instead.
sed '/^if \[ -n "${recovery_requested}" \]$/,+4d' "$DIS" > "$WORK/dis-norecoveryreq.inc"
caught "an ignored recovery request" "$SEL" "$WORK/dis-norecoveryreq.inc"

echo
echo "===== the unmutated pair still passes ====="
if sh "$SUITE" "$SEL" "$DIS" >/dev/null 2>&1; then
    note "the shipped dispatch passes its own suite" OK
else
    note "the shipped dispatch passes its own suite" FAIL
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
