#!/bin/sh
# Check that stop_process sends a signal the processes can actually receive.
#
#   test-server-stop-signals.sh [path-to-S95nanokvm]
#
# S95nanokvm starts both processes as background jobs of a shell without job
# control, and POSIX requires such a shell to set SIGINT to SIG_IGN in the
# child. Measured on the device 2026-08-19:
#
#     NanoKVM-Server  SigIgn 0x2   SIGINT ignored, SIGTERM caught
#     kvm_system      SigIgn 0x6   SIGINT ignored, SigCgt 0, SIGTERM default
#
# So SIGINT reaches no code in either process. Sending it first and then waiting
# for the process to act on it spends the whole wait on a signal that cannot
# arrive. Two SIGINTs to the running server produced no log line and no exit
# after 21s; one SIGTERM took 7s.
#
# The wait after SIGTERM matters as much as the signal. The server bounds its
# own teardown with disposeTimeout in server/main.go and then exits regardless,
# so a wait shorter than that races a teardown that was going to finish. This
# suite reads that constant rather than repeating it, because the two drifting
# apart is exactly the failure it exists to stop.
#
# The escalation to SIGKILL stays. A process wedged in uninterruptible sleep
# answers nothing, and the stop has to end.
set -u

DIR=$(dirname "$0")
S95=${1:-$DIR/../../kvmapp/system/init.d/S95nanokvm}
MAIN=$DIR/../../server/main.go

[ -f "$S95" ] || { echo "usage: test-server-stop-signals.sh <S95nanokvm>"; exit 2; }
[ -f "$MAIN" ] || {
    echo "test-server-stop-signals.sh: cannot read $MAIN."
    echo "it holds disposeTimeout, which the wait below is measured against."
    exit 2
}

# disposeTimeout bounds the server's own teardown. Parse it; do not assume it.
DISPOSE=$(sed -n 's/^const disposeTimeout = \([0-9][0-9]*\) \* time.Second$/\1/p' "$MAIN")
case "$DISPOSE" in
    ''|*[!0-9]*)
        echo "test-server-stop-signals.sh: no disposeTimeout constant in $MAIN."
        echo "it changed shape, so the wait below is measured against nothing."
        exit 2
        ;;
esac

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# Lift every top-level function, so a helper the code under test calls is never
# missing and never a local reimplementation of it. stop_process calls
# wait_for_exit, and a suite that supplied its own would be testing its own.
cat > "$work/f.sh" <<'STUB'
# The stubs record rather than act. What this suite judges is the order of the
# signals and the length of the wait between them, and both are visible only as
# a sequence.
pidof() { [ ! -f "$STATE/dead" ]; }
killall() {
    echo "signal $1" >> "$LOG"
    [ "$1" = "${EXIT_ON:-never}" ] && : > "$STATE/dead"
    return 0
}
sleep() { echo "sleep" >> "$LOG"; }
STUB
sed -n '/^[a-z_][a-z_]*() *{$/,/^}$/p' "$S95" >> "$work/f.sh"

for fn in stop_process wait_for_exit; do
    if ! grep -q "^$fn() {" "$work/f.sh"; then
        note "S95nanokvm defines $fn" FAIL
        echo; echo "$fails case(s) FAILED"; exit 1
    fi
done
note "S95nanokvm defines stop_process and wait_for_exit" OK
note "server/main.go sets disposeTimeout to ${DISPOSE}s" OK

# run <exit-on-signal> <already-dead>; leaves the sequence in $work/log.
run() {
    rm -rf "$work/state"; mkdir -p "$work/state"
    : > "$work/log"
    [ "${2:-no}" = yes ] && : > "$work/state/dead"

    (
        STATE=$work/state
        LOG=$work/log
        EXIT_ON=$1
        export STATE LOG EXIT_ON
        . "$work/f.sh"
        stop_process NanoKVM-Server
    ) > /dev/null 2>&1
}

signals()     { grep '^signal' "$work/log" | sed 's/^signal //' | tr '\n' ' '; }
first_signal(){ grep -m1 '^signal' "$work/log" | sed 's/^signal //'; }
# Sleeps between SIGTERM and whatever signal follows it: the wait SIGTERM gets.
# Measured against SIGTERM by name rather than by position, so a script that
# sends something else first cannot report that signal's wait as this one's.
wait_after_term() {
    awk '/^signal -TERM$/{ seen=1; next } /^signal/ && seen { exit } /^sleep/ && seen { c++ } END{ print c+0 }' "$work/log"
}

echo "===== the first signal has to be one the process can receive ====="

run never
if [ "$(first_signal)" = -TERM ]; then
    note "the first signal is SIGTERM" OK
else
    note "the first signal is SIGTERM (got '$(first_signal)')" FAIL
fi

case " $(signals)" in
    *" -INT"*) note "SIGINT is never sent" FAIL ;;
    *)         note "SIGINT is never sent" OK ;;
esac

echo
echo "===== the wait has to outlast the teardown it is waiting for ====="

got=$(wait_after_term)
if [ "$got" -gt "$DISPOSE" ]; then
    note "the wait after SIGTERM (${got}s) outlasts disposeTimeout (${DISPOSE}s)" OK
else
    note "the wait after SIGTERM (${got}s) outlasts disposeTimeout (${DISPOSE}s)" FAIL
fi

# A wait with no end is not a stop. The board has to come back.
if [ "$got" -le 30 ]; then
    note "the wait after SIGTERM is bounded (${got}s)" OK
else
    note "the wait after SIGTERM is bounded (${got}s)" FAIL
fi

echo
echo "===== the stop still ends ====="

run never
case " $(signals)" in
    *" -KILL"*) note "a process that never leaves is sent SIGKILL" OK ;;
    *)          note "a process that never leaves is sent SIGKILL" FAIL ;;
esac

# SIGKILL reaches no code, so it costs the carveout a leaked video buffer pool.
# It is the last resort and it must not become the first.
run -TERM
case " $(signals)" in
    *" -KILL"*) note "a process that leaves on SIGTERM is not sent SIGKILL" FAIL ;;
    *)          note "a process that leaves on SIGTERM is not sent SIGKILL" OK ;;
esac

run never yes
if [ -s "$work/log" ]; then
    note "a process that is not running is not signalled" FAIL
else
    note "a process that is not running is not signalled" OK
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "===== stop_process signals what the processes can receive ====="
else
    echo "$fails case(s) FAILED"
    exit 1
fi
