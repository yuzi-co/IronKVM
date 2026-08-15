#!/bin/sh
# Check the slot tool's guards against loop files rather than a card.
#
#   test-slot.sh [path-to-slot]
#
# install writes a whole partition with dd. These guards are the only thing
# between a typo and a root filesystem overwritten while it is running, so they
# are checked here where a mistake costs nothing.
SLOT=${1:-$(dirname "$0")/slot}
[ -f "$SLOT" ] || { echo "usage: test-slot.sh <slot>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

mkdir -p "$WORK/boot"
cat > "$WORK/conf" <<CONF
SLOT_A=$WORK/a.img
SLOT_B=$WORK/b.img
RECOVERY=$WORK/r.img
DATA_DEV=$WORK/d.img
CONF

reset() {
    rm -f "$WORK"/*.img
    truncate -s 8M "$WORK/a.img" "$WORK/b.img" "$WORK/r.img"
    printf 'candidate\n' > "$WORK/cand.img"
    truncate -s 4M "$WORK/cand.img"
    rm -f "$WORK/boot"/*
    echo a > "$WORK/boot/slot"
}

# run [RUNNING=x] <args...>
run() {
    running=a
    case "$1" in
        RUNNING=*) running=${1#RUNNING=}; shift ;;
    esac
    SLOT_CONF="$WORK/conf" BOOT="$WORK/boot" RUNNING_SLOT="$running" sh "$SLOT" "$@" 2>&1
}

echo "===== install refuses what it must ====="
reset

run install a "$WORK/cand.img" >/dev/null 2>&1 \
    && note "installing onto the running slot is refused" FAIL \
    || note "installing onto the running slot is refused" OK

truncate -s 16M "$WORK/toobig.img"
run install b "$WORK/toobig.img" >/dev/null 2>&1 \
    && note "an image larger than the partition is refused" FAIL \
    || note "an image larger than the partition is refused" OK

run install b "$WORK/nosuch.img" >/dev/null 2>&1 \
    && note "a missing image is refused" FAIL \
    || note "a missing image is refused" OK

run install zzz "$WORK/cand.img" >/dev/null 2>&1 \
    && note "an unknown slot name is refused" FAIL \
    || note "an unknown slot name is refused" OK

echo
echo "===== install writes and verifies ====="
reset

if run install b "$WORK/cand.img" >"$WORK/inst.log" 2>&1; then
    note "installing to a free slot succeeds" OK
else
    note "installing to a free slot succeeds" FAIL
    sed 's/^/    /' "$WORK/inst.log" | head -5
fi

# The bytes of the image must actually be at the front of the target.
isize=$(wc -c < "$WORK/cand.img")
head -c "$isize" "$WORK/b.img" | sha256sum | cut -d' ' -f1 > "$WORK/got"
sha256sum "$WORK/cand.img" | cut -d' ' -f1 > "$WORK/want"
cmp -s "$WORK/got" "$WORK/want" \
    && note "the installed bytes match the image" OK \
    || note "the installed bytes match the image" FAIL

echo
echo "===== try and confirm move the right markers ====="
reset

run try b >/dev/null 2>&1
[ "$(cat "$WORK/boot/slot.try" 2>/dev/null)" = b ] \
    && note "try writes slot.try" OK || note "try writes slot.try" FAIL
[ "$(cat "$WORK/boot/slot")" = a ] \
    && note "try leaves the trusted marker alone" OK || note "try leaves the trusted marker alone" FAIL

run RUNNING=b confirm >/dev/null 2>&1
[ "$(cat "$WORK/boot/slot")" = b ] \
    && note "confirm promotes the running slot" OK || note "confirm promotes the running slot" FAIL
[ "$(cat "$WORK/boot/slot.prev")" = a ] \
    && note "confirm records the previous slot" OK || note "confirm records the previous slot" FAIL

# Confirming twice must not lose the real previous slot by recording the
# current one over it.
run RUNNING=b confirm >/dev/null 2>&1
[ "$(cat "$WORK/boot/slot.prev")" = a ] \
    && note "confirming twice keeps the real previous slot" OK \
    || note "confirming twice overwrote prev with $(cat "$WORK/boot/slot.prev")" FAIL

run revert >/dev/null 2>&1
[ "$(cat "$WORK/boot/slot")" = a ] \
    && note "revert restores the previous slot" OK || note "revert restores the previous slot" FAIL

echo
echo "===== recovery does not promote itself ====="
reset

run RUNNING=recovery confirm >/dev/null 2>&1 \
    && note "confirm from recovery is refused" FAIL \
    || note "confirm from recovery is refused" OK

run recovery >/dev/null 2>&1
[ -e "$WORK/boot/recovery" ] \
    && note "recovery writes its marker" OK || note "recovery writes its marker" FAIL

echo
echo "===== status reports the markers ====="
reset
echo b > "$WORK/boot/slot"
echo a > "$WORK/boot/slot.try"

out=$(run RUNNING=b status)
echo "$out" | grep -q "running *b"  && note "status reports the running slot" OK  || note "status reports the running slot" FAIL
echo "$out" | grep -q "trusted *b"  && note "status reports the trusted slot" OK  || note "status reports the trusted slot" FAIL
echo "$out" | grep -q "trial *a"    && note "status reports a pending trial" OK   || note "status reports a pending trial" FAIL

echo
echo "===== the script parses ====="
sh -n "$SLOT" 2>/dev/null && note "sh -n accepts the tool" OK || note "sh -n accepts the tool" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
fi
exit "$fails"
