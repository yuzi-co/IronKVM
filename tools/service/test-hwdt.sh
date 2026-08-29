#!/bin/sh
# Check what S01hwdt arms the SoC watchdog with, and what it refuses to arm.
#
#   test-hwdt.sh [path-to-S01hwdt]
#
# Not destructive: no watchdog device is opened and no daemon is started. The
# device and the daemon are both stubs, and every case runs the shipped script
# rather than a copy of its decisions.
#
# The section on forbidden settings is the reason this file exists. Everything
# this script arms can reset the board, and /etc/watchdog.conf ships a sample
# full of plausible lines that would do it for the wrong reason. max-load-1 is
# the dangerous one: this board idles at a load average of about 4 with no CPU
# load, because a reader of /proc/cvitek/vb is counted for ever in
# uninterruptible sleep.
HERE=$(cd "$(dirname "$0")" && pwd)
WD=${1:-$HERE/S01hwdt}
[ -f "$WD" ] || { echo "usage: test-hwdt.sh <S01hwdt>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

# A daemon that records how it was called and starts nothing.
mkdir -p "$WORK/bin"
cat > "$WORK/bin/watchdog" <<'STUB'
#!/bin/sh
echo "$@" > "$ARGV_OUT"
exit 0
STUB
chmod 755 "$WORK/bin/watchdog"

# /dev/null is a character device everywhere this runs, which is all preflight
# asks of the watchdog device.
DEV=/dev/null
[ -c "$DEV" ] || { echo "this host has no character device at $DEV to stand in for /dev/watchdog"; exit 2; }

run() {   # run <action>, with the environment already set by the caller
    HWDOG_DEVICE=${DEVICE_OVERRIDE:-$DEV} \
    HWDOG_DAEMON=${DAEMON_OVERRIDE:-$WORK/bin/watchdog} \
    HWDOG_CONF="$WORK/watchdog.conf" \
    HWDOG_LOG="$WORK/hwdt.log" \
    ARGV_OUT="$WORK/argv" \
    HWDOG_TIMEOUT="$TIMEOUT" \
    HWDOG_INTERVAL="$INTERVAL" \
    HWDOG_RETRY="$RETRY" \
    HWDOG_ALLOCATABLE="$ALLOCATABLE" \
    sh "$WD" "$1" 2>&1
}

reset_env() {
    DEVICE_OVERRIDE=
    DAEMON_OVERRIDE=
    TIMEOUT=60
    INTERVAL=8
    RETRY=120
    ALLOCATABLE=256
    rm -f "$WORK/watchdog.conf" "$WORK/argv" "$WORK/hwdt.log"
}

verdict() {
    run status | sed -n 's/^preflight : //p'
}

echo
echo "===== it refuses to arm rather than guessing ====="

reset_env
[ "$(verdict)" = ok ] \
    && note "the shipped defaults pass preflight" OK \
    || note "the shipped defaults give $(verdict), not ok" FAIL

reset_env
DEVICE_OVERRIDE="$WORK/absent"
[ "$(verdict)" = no-device ] \
    && note "a board with no watchdog device does not arm" OK \
    || note "a board with no watchdog device gives $(verdict)" FAIL

reset_env
DAEMON_OVERRIDE="$WORK/absent"
[ "$(verdict)" = no-daemon ] \
    && note "a board with no daemon does not arm" OK \
    || note "a board with no daemon gives $(verdict)" FAIL

# A daemon that pings less often than the hardware resets is a reset timer with
# extra steps. It would take out a healthy board on a fixed period, for ever.
#
# The margin is six pings against the REQUESTED timeout, not three, because the
# driver arms the largest step below the request: 60 asked for arms 42 on this
# part. Six against the request is three against what the hardware gives.
reset_env
INTERVAL=30
[ "$(verdict)" = interval-too-long ] \
    && note "pinging as often as the timeout is refused" OK \
    || note "an interval of 30s against a 60s timeout gives $(verdict)" FAIL

reset_env
INTERVAL=10
[ "$(verdict)" = ok ] \
    && note "six pings inside the requested window is enough" OK \
    || note "an interval of 10s against a 60s timeout gives $(verdict)" FAIL

reset_env
INTERVAL=11
[ "$(verdict)" = interval-too-long ] \
    && note "fewer than six pings inside the requested window is refused" OK \
    || note "an interval of 11s against a 60s timeout gives $(verdict)" FAIL

# The rule is about the ratio, not about the shipped numbers.
reset_env
TIMEOUT=120
INTERVAL=20
[ "$(verdict)" = ok ] \
    && note "the margin scales with the timeout" OK \
    || note "a 20s interval against a 120s timeout gives $(verdict)" FAIL

# An empty override is not garbage: ${HWDOG_TIMEOUT:-60} reads it as unset,
# which is the convention every script in this directory follows. The values
# below are the ones a person can actually put in front of the script.
reset_env
TIMEOUT=
[ "$(verdict)" = ok ] \
    && note "an empty override falls back to the default" OK \
    || note "an empty override gives $(verdict), not the default" FAIL

for bad in abc -5 6.5; do
    reset_env
    TIMEOUT=$bad
    [ "$(verdict)" = bad-number ] \
        && note "a timeout of [$bad] is refused" OK \
        || note "a timeout of [$bad] gives $(verdict)" FAIL
done

reset_env
TIMEOUT=0
[ "$(verdict)" = bad-number ] \
    && note "a timeout of zero is refused" OK \
    || note "a timeout of zero gives $(verdict)" FAIL

echo
echo "===== a watchdog that cannot arm never stops the boot ====="
# This runs first in the rc sequence. Every branch has to leave the board
# booting, including the branches that decide not to arm anything.

reset_env
DEVICE_OVERRIDE="$WORK/absent"
run start > /dev/null 2>&1
[ $? -eq 0 ] \
    && note "start exits 0 with no watchdog device" OK \
    || note "start exits non-zero with no watchdog device" FAIL

reset_env
DAEMON_OVERRIDE="$WORK/absent"
run start > /dev/null 2>&1
[ $? -eq 0 ] \
    && note "start exits 0 with no daemon" OK \
    || note "start exits non-zero with no daemon" FAIL

reset_env
INTERVAL=999
run start > /dev/null 2>&1
[ $? -eq 0 ] \
    && note "start exits 0 on a configuration it refuses" OK \
    || note "start exits non-zero on a configuration it refuses" FAIL

[ -e "$WORK/argv" ] \
    && note "a refused configuration still started the daemon" FAIL \
    || note "a refused configuration starts no daemon" OK

echo
echo "===== what it arms ====="

reset_env
run start > /dev/null 2>&1

[ -s "$WORK/watchdog.conf" ] \
    && note "a configuration is written" OK \
    || note "no configuration is written" FAIL

[ -s "$WORK/argv" ] \
    && note "the daemon is started" OK \
    || note "the daemon is not started" FAIL

grep -q -- "-c $WORK/watchdog.conf" "$WORK/argv" 2>/dev/null \
    && note "the daemon is pointed at the generated configuration" OK \
    || note "the daemon reads some other configuration [$(cat "$WORK/argv" 2>/dev/null)]" FAIL

# has <key> <value>: the generated file must set it, and set it to this.
has() {
    if grep -qE "^[[:space:]]*$1[[:space:]]*=[[:space:]]*$2[[:space:]]*$" "$WORK/watchdog.conf"; then
        note "$1 is $2" OK
    else
        note "$1 is $(sed -n "s/^[[:space:]]*$1[[:space:]]*=[[:space:]]*//p" "$WORK/watchdog.conf"), want $2" FAIL
    fi
}

has watchdog-device    "$DEV"
has watchdog-timeout   60
has interval           8
has retry-timeout      120
has allocatable-memory 256

# The memory test is the one setting here that has never been proven against
# the fault it is for, and it cannot be: inducing memory pressure on this board
# is what wedges it. Zero has to leave the plain timer armed and drop only that
# line, so an operator can back it out without losing the layer.
reset_env
ALLOCATABLE=0
run start > /dev/null 2>&1
if grep -qE '^[[:space:]]*allocatable-memory' "$WORK/watchdog.conf"; then
    note "zero pages still asks for a memory test" FAIL
else
    note "zero pages drops the memory test" OK
fi
if grep -qE "^[[:space:]]*watchdog-timeout[[:space:]]*=[[:space:]]*60[[:space:]]*$" "$WORK/watchdog.conf"; then
    note "zero pages leaves the plain timer armed" OK
else
    note "zero pages disarmed the timer as well" FAIL
fi

reset_env
run start > /dev/null 2>&1

# The daemon locks itself into memory. A watchdog swapped out under memory
# pressure stops exactly when the fault it exists for arrives.
has realtime           yes

echo
echo "===== what it must never arm ====="
# Each line below is in the sample /etc/watchdog.conf and would reset this board
# for a reason that is not a fault.
#
#   max-load-*        the idle load average here is about 4, and it is not CPU
#   min-memory        reads cache, and MemFree is 14MB on a healthy board
#   max-swap          one wedge on record is not enough to place a threshold
#   ping / interface  an external fault is not a board fault
#   pidfile           S98supervise restarts what dies, which is a better answer
#   repair-binary     a repair has to fork, and the fault is that forking stopped
for key in max-load-1 max-load-5 max-load-15 min-memory max-swap ping \
           ping-count interface pidfile repair-binary test-binary \
           test-directory temperature-sensor max-temperature file change; do
    if grep -qE "^[[:space:]]*$key[[:space:]]*=" "$WORK/watchdog.conf"; then
        note "$key is set, and it can reset a healthy board" FAIL
    else
        note "$key is not set" OK
    fi
done

echo
echo "===== stopping and disarming are opposites ====="
# The kernel keeps the timer running or stops it depending on HOW the device was
# closed. A daemon killed with SIGTERM writes the magic V and the timer stops. A
# daemon killed with SIGKILL cannot, and the timer keeps running.
#
# rcK runs stop on the way to a reboot, and it must leave the timer running: on
# 2026-08-15 this board ran a whole shutdown and never reset, and only removing
# power ended it. disarm is the maintenance action and must stop the timer.
sed -n '/^# --- stopping ---/,/^# --- end stopping ---/p' "$WD" > "$WORK/stopping.sh"
if [ -s "$WORK/stopping.sh" ]; then
    note "the stopping block can be extracted" OK
else
    note "the stopping block can be extracted" FAIL
fi

signal_of() {   # signal_of <stop|disarm>
    (
        LOG="$WORK/hwdt.log"
        TIMEOUT=60
        export LOG TIMEOUT
        running() { return 0; }
        killall() { echo "$@" > "$WORK/signal"; }
        log() { :; }
        . "$WORK/stopping.sh"
        "$1"
    ) > /dev/null 2>&1
    cat "$WORK/signal" 2>/dev/null
}

got=$(signal_of stop)
case "$got" in
    *-9*) note "stop denies the daemon its magic close, so the timer survives" OK ;;
    *)    note "stop calls killall [$got], which lets the timer be switched off" FAIL ;;
esac

got=$(signal_of disarm)
case "$got" in
    *-9*) note "disarm kills the daemon so hard the timer keeps running" FAIL ;;
    *)    note "disarm lets the daemon write its magic close" OK ;;
esac

# disarm has to wait for the daemon to actually go. Without the wait a restart
# is a disarm: start runs while the old daemon is still dying, finds a pid,
# reports "already armed" and starts nothing. Seen on the board on 2026-08-29,
# and the only trace was two log lines in the same second.
rm -f "$WORK/polls"
(
    LOG="$WORK/hwdt.log"
    export LOG
    # Alive for the first three questions, gone after that.
    running() {
        n=$(cat "$WORK/polls" 2>/dev/null)
        case "$n" in ""|*[!0-9]*) n=0 ;; esac
        n=$((n + 1))
        echo "$n" > "$WORK/polls"
        [ "$n" -le 3 ]
    }
    killall() { :; }
    sleep() { :; }
    log() { :; }
    . "$WORK/stopping.sh"
    disarm
) > /dev/null 2>&1

polls=$(cat "$WORK/polls" 2>/dev/null)
case "$polls" in
    ''|0|1) note "disarm returns while the daemon is still dying (polled $polls times)" FAIL ;;
    *)      note "disarm waits for the daemon to leave (polled $polls times)" OK ;;
esac

# And it must say so rather than reporting a disarm that did not happen. A
# daemon that will not leave has left the timer running, and SIGKILL is not the
# escalation here: a hard kill cannot write the magic V, so it would arm the
# board rather than disarm it.
out=$(
    LOG="$WORK/hwdt.log" ; export LOG
    running() { return 0; }
    killall() { :; }
    sleep() { :; }
    log() { :; }
    . "$WORK/stopping.sh"
    disarm
    echo "rc=$?"
)
case "$out" in
    *"STILL ARMED"*rc=1*) note "a daemon that will not leave is reported, not assumed gone" OK ;;
    *)                    note "a daemon that will not leave reports [$out]" FAIL ;;
esac

echo
echo "===== the script parses ====="
if sh -n "$WD"; then
    note "sh -n accepts the script" OK
else
    note "sh -n accepts the script" FAIL
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
