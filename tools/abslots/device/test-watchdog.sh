#!/bin/sh
# Check the watchdog's health rule, extracted from the shipped script.
#
#   test-watchdog.sh [path-to-S00awatchdog]
#
# The rule changed for a reason worth keeping in front of whoever edits it next.
# The old rule was net_up AND (web_up OR ssh_up), and net_up required carrier. An
# unplugged cable therefore rebooted the board into recovery, where the cable was
# still unplugged. That is a reboot loop dressed as a safety feature, and it
# punishes an external fault with the one action that cannot help it.
#
# The block is extracted from the script rather than copied here, so the test
# cannot drift away from what ships. Two guards in this repository had already
# rotted by testing a copy that no longer matched.
WD=${1:-$(dirname "$0")/S00awatchdog}
[ -f "$WD" ] || { echo "usage: test-watchdog.sh <S00awatchdog>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

sed -n '/^# --- health check ---/,/^# --- end health check ---/p' "$WD" > "$WORK/health.sh"
if [ ! -s "$WORK/health.sh" ]; then
    note "the health block can be extracted" FAIL
    echo
    echo "1 case(s) FAILED"
    exit 1
fi
note "the health block can be extracted" OK

# verdict <carrier> <address> <listener>
verdict() {
    (
        CARRIER=$1 ADDRESS=$2 LISTENER=$3
        export CARRIER ADDRESS LISTENER
        carrier_up() { [ "$CARRIER" = yes ]; }
        address_up() { [ "$ADDRESS" = yes ]; }
        web_up()     { [ "$LISTENER" = web ]; }
        ssh_up()     { [ "$LISTENER" = ssh ]; }
        . "$WORK/health.sh"
        if healthy; then echo healthy; else echo escalate; fi
    )
}

echo
echo "===== an external fault is not a slot fault ====="

# No carrier means the cable is out or the switch is down. Rebooting into
# recovery lands on a board whose cable is still out.
[ "$(verdict no no none)" = healthy ] \
    && note "no carrier stands down instead of escalating" OK \
    || note "no carrier escalates, which reboots into a still-unplugged cable" FAIL

[ "$(verdict no yes web)" = healthy ] \
    && note "no carrier stands down even if something answers" OK \
    || note "no carrier stands down even if something answers" FAIL

echo
echo "===== a slot fault does escalate ====="

[ "$(verdict yes no none)" = escalate ] \
    && note "carrier but no address escalates" OK \
    || note "carrier but no address escalates" FAIL

[ "$(verdict yes yes none)" = escalate ] \
    && note "an address but no listener escalates" OK \
    || note "an address but no listener escalates" FAIL

echo
echo "===== either door is enough ====="

[ "$(verdict yes yes web)" = healthy ] && note "web alone is healthy" OK || note "web alone is healthy" FAIL
[ "$(verdict yes yes ssh)" = healthy ] && note "ssh alone is healthy" OK || note "ssh alone is healthy" FAIL

echo
echo "===== the escalation goes to recovery, not to a marker revert ====="

# Reverting the marker lands on another full root that may carry the same
# fault: on 2026-08-15 every candidate slot had the same class of problem.
#
# This case runs the escalation rather than grepping it. A first version of
# this file checked `grep -q recovery` on the block, and an escalation mutated
# to `rm -f $BOOT/slot` still passed, because the word survived in a comment
# and a log line. Behaviour has to be executed to be tested.
sed -n '/^# --- escalate ---/,/^# --- end escalate ---/p' "$WD" > "$WORK/esc.sh"
if [ -s "$WORK/esc.sh" ]; then
    note "the escalation block can be extracted" OK
else
    note "the escalation block can be extracted" FAIL
fi

mkdir -p "$WORK/boot"
echo b > "$WORK/boot/slot"
rm -f "$WORK/boot/recovery" "$WORK/rebooted"

(
    BOOT="$WORK/boot"
    LOG="$WORK/wd.log"
    DEADLINE=1
    export BOOT LOG DEADLINE
    carrier_up() { return 1; }
    address_up() { return 1; }
    web_up()     { return 1; }
    ssh_up()     { return 1; }
    ensure_boot_mounted() { return 0; }
    log() { echo "$*" >> "$LOG"; }
    reboot() { : > "$WORK/rebooted"; }
    . "$WORK/esc.sh"
    escalate
) >/dev/null 2>&1

[ -e "$WORK/boot/recovery" ] \
    && note "escalation creates the recovery marker" OK \
    || note "escalation creates the recovery marker" FAIL

[ "$(cat "$WORK/boot/slot" 2>/dev/null)" = b ] \
    && note "escalation leaves the trusted marker alone" OK \
    || note "escalation changed the trusted marker to $(cat "$WORK/boot/slot" 2>/dev/null)" FAIL

[ -e "$WORK/rebooted" ] \
    && note "escalation reboots" OK \
    || note "escalation reboots" FAIL

echo
echo "===== a healthy trial is confirmed without anyone remembering to ====="

grep -q 'slot confirm' "$WD" \
    && note "the watchdog confirms the running slot when healthy" OK \
    || note "the watchdog confirms the running slot when healthy" FAIL

echo
echo "===== the script parses ====="
sh -n "$WD" 2>/dev/null && note "sh -n accepts the script" OK || note "sh -n accepts the script" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
fi
exit "$fails"
