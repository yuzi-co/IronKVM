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
    # A board that HAS slots. escalate now checks, because install.sh can put
    # this watchdog on a board with stock partitioning where the marker means
    # nothing. command -v finds a function, so this is enough to model one.
    slot() { :; }
    # Nothing to undo. The rollback is tried first now, and it must not stop
    # the fall back to recovery when there is no update to blame.
    restore_initd() { return 1; }
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
echo "===== an outstanding update is undone before recovery is used ====="

# A board that cannot be reached right after an update is very probably broken
# BY that update. Putting the previous boot scripts back costs one reboot and
# keeps video and HID; recovery keeps neither and is sticky on purpose.
#
# This ordering is also what makes the rollback reachable at all. Without it the
# boot counter could never reach its limit on a board with this watchdog: either
# the board answers and the count clears, or it does not and recovery takes it
# on the first boot.
rm -f "$WORK/boot/recovery" "$WORK/rebooted" "$WORK/restored"

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
    slot() { :; }
    clear_attempts() { :; }
    restore_initd() { : > "$WORK/restored"; return 0; }
    . "$WORK/esc.sh"
    escalate
) >/dev/null 2>&1

[ -e "$WORK/restored" ] \
    && note "an outstanding update is undone" OK \
    || note "an outstanding update is undone" FAIL

[ -e "$WORK/boot/recovery" ] \
    && note "undoing an update does NOT also drop to recovery" FAIL \
    || note "undoing an update does NOT also drop to recovery" OK

[ -e "$WORK/rebooted" ] \
    && note "undoing an update reboots" OK \
    || note "undoing an update reboots" FAIL

echo
echo "===== a board with no recovery slot stands down instead of looping ====="

# install.sh copies this watchdog out of an update package, so it can now run on
# a board with stock partitioning. There the marker is an inert file and the
# reboot repeats every DEADLINE seconds for ever, which stops anyone reaching
# the board during boot. Standing down is the same answer the carrier check
# gives to a fault the board cannot fix.
rm -f "$WORK/boot/recovery" "$WORK/rebooted"

(
    BOOT="$WORK/boot"
    LOG="$WORK/wd.log"
    DEADLINE=1
    export BOOT LOG DEADLINE
    PATH=/nonexistent
    carrier_up() { return 1; }
    address_up() { return 1; }
    web_up()     { return 1; }
    ssh_up()     { return 1; }
    ensure_boot_mounted() { return 0; }
    log() { echo "$*" >> "$LOG"; }
    reboot() { : > "$WORK/rebooted"; }
    restore_initd() { return 1; }
    . "$WORK/esc.sh"
    escalate
) >/dev/null 2>&1

[ -e "$WORK/boot/recovery" ] \
    && note "no slot tool means no recovery marker" FAIL \
    || note "no slot tool means no recovery marker" OK

[ -e "$WORK/rebooted" ] \
    && note "no slot tool means no reboot loop" FAIL \
    || note "no slot tool means no reboot loop" OK

echo
echo "===== a failed trial goes back to the trusted slot, not to recovery ====="
#
# Recovery is the fallback for a TRUSTED slot that failed. A trial has a better
# one: the slot that was working minutes ago. The initramfs already deleted
# slot.try before handing over, so a plain reboot lands on the trusted slot.
#
# Sending a failed trial to recovery instead costs the board its whole KVM
# function, since recovery serves no video and no HID, and it costs the
# operator a second reboot to leave a marker that is sticky on purpose.
#
# No new marker is needed to tell the two apart. A boot whose running slot is
# not the trusted slot is a trial, and both halves of that are already on disk.

sed -n '/^# --- trial boot ---/,/^# --- end trial boot ---/p' "$WD" > "$WORK/trial.sh"
if [ ! -s "$WORK/trial.sh" ]; then
    note "the trial-boot block can be extracted" FAIL
else
    note "the trial-boot block can be extracted" OK
fi

# escalate_with <running> <trusted>  -- returns the marker state it left
escalate_with() {
    rm -rf "$WORK/e"; mkdir -p "$WORK/e/boot"
    printf '%s\n' "$2" > "$WORK/e/boot/slot"
    (
        BOOT="$WORK/e/boot"
        LOG="$WORK/e/wd.log"
        RUNNING=$1 TRUSTED=$2
        export BOOT LOG RUNNING TRUSTED
        carrier_up() { return 1; }
        address_up() { return 1; }
        web_up()     { return 1; }
        ssh_up()     { return 1; }
        ensure_boot_mounted() { return 0; }
        log() { echo "$*" >> "$LOG"; }
        reboot() { : > "$WORK/e/rebooted"; }
        # The one external fact the block reads. `slot status` prints a
        # two-column table, and this is that table.
        slot() {
            [ "$1" = status ] || return 1
            printf 'running    %s\ntrusted    %s\n' "$RUNNING" "$TRUSTED"
        }
        . "$WORK/trial.sh"
        . "$WORK/esc.sh"
        escalate
    ) >/dev/null 2>&1
    [ -e "$WORK/e/boot/recovery" ] && echo recovery || echo reboot
}

[ "$(escalate_with b a)" = reboot ] \
    && note "a trial of b with a trusted returns to a" OK \
    || note "a trial of b with a trusted went to recovery" FAIL

[ "$(escalate_with a b)" = reboot ] \
    && note "a trial of a with b trusted returns to b" OK \
    || note "a trial of a with b trusted went to recovery" FAIL

# The trusted slot failing is the case recovery exists for.
[ "$(escalate_with a a)" = recovery ] \
    && note "the trusted slot failing still goes to recovery" OK \
    || note "the trusted slot failing no longer goes to recovery" FAIL

# And anything this cannot read must take the safe branch. A watchdog that
# cannot tell where it is must not assume it has a good slot to fall back to.
[ "$(escalate_with unknown a)" = recovery ] \
    && note "an unreadable running slot takes the safe branch" OK \
    || note "an unreadable running slot skipped recovery" FAIL

[ "$(escalate_with '' '')" = recovery ] \
    && note "an unreadable status takes the safe branch" OK \
    || note "an unreadable status skipped recovery" FAIL

# Half an answer is not an answer. "running is blank, trusted is a" differs as
# a string comparison and would read as a trial, which is the one way this
# could skip recovery on a board it knows nothing about.
[ "$(escalate_with '' a)" = recovery ] \
    && note "a blank running slot takes the safe branch" OK \
    || note "a blank running slot read as a trial" FAIL

[ "$(escalate_with b '')" = recovery ] \
    && note "a blank trusted slot takes the safe branch" OK \
    || note "a blank trusted slot read as a trial" FAIL

# Recovery carries no watchdog, so this cannot happen; it costs one line to
# make sure it stays that way if one is ever added.
[ "$(escalate_with recovery a)" = recovery ] \
    && note "recovery is never treated as a trial" OK \
    || note "recovery was treated as a trial" FAIL

# Either way the board must reboot. A watchdog that decides and then does
# nothing is the dead board it exists to prevent.
escalate_with b a >/dev/null
[ -e "$WORK/e/rebooted" ] \
    && note "a failed trial still reboots" OK \
    || note "a failed trial decided and then sat there" FAIL

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
