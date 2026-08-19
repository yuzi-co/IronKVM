#!/bin/sh
# Check that S95nanokvm waits for the old server to exit before it starts a new
# one.
#
#   test-restart-waits.sh [path-to-S95nanokvm]
#
# killall returns as soon as the signal is sent. The server catches the signal
# and tears the capture pipeline down first, and that teardown reaches libkvm
# through cgo, where it can block. The old process is therefore still running,
# and still owns the VI pipeline, after killall returns.
#
# The next server then builds the media stack while the old one still holds it,
# and its channel enable reports ENOMEM. The old process keeps answering on the
# old build, so the restart looks like it worked.
#
# The staged copies come out before the new ones go in. S98supervise reads a
# staged binary with no process as a crash and starts a server itself, and the
# wait is long enough for it to do that.
#
# This runs the shipped functions against stub processes and checks what they
# do, so it fails if the escalation stops waiting or stops forcing.
#
# An earlier version of this file extracted a single wait_for_stop(). The work
# now lives in wait_for_exit() and stop_process(), and the extraction returned
# nothing: the first case failed and the file exited before it checked any
# behaviour at all. The names are asserted here for that reason - a harness
# that cannot find its subject must say so, not pass.
S95=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S95nanokvm}
[ -f "$S95" ] || { echo "usage: test-restart-waits.sh <S95nanokvm>"; exit 1; }

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

echo "===== the escalation ends, whether the server leaves or not ====="

# Run the shipped text rather than a copy of it.
: > "$work/func.sh"
for fn in wait_for_exit stop_process; do
    sed -n "/^$fn() {/,/^}/p" "$S95" >> "$work/func.sh"
done

defined=$(grep -c '^[a-z_]*() {' "$work/func.sh" || true)
if [ "$defined" != 2 ]; then
    note "the script defines wait_for_exit and stop_process ($defined/2)" FAIL
    echo
    echo "$fails case(s) FAILED"
    exit "$fails"
fi
note "the script defines wait_for_exit and stop_process" OK

# The harness replaces killall, pidof and sleep. Each records what it was asked
# to do, and sleep does not really sleep, so a twenty second timeout costs
# nothing.
cat > "$work/harness.sh" <<'HARNESS'
killall() {
    printf '%s\n' "killall $*" >> "$LOGFILE"
}

# The stub process is alive for EXIT_AFTER polls and gone after that. A value
# larger than every timeout is a process that never leaves. ALIVE=0 is a
# process that was never running.
pidof() {
    [ "$ALIVE" = 0 ] && return 1
    polls=$(cat "$WORK/polls" 2>/dev/null || echo 0)
    [ "$polls" -lt "$EXIT_AFTER" ]
}

sleep() {
    polls=$(cat "$WORK/polls" 2>/dev/null || echo 0)
    echo $((polls + 1)) > "$WORK/polls"
}
HARNESS

# run_case <exit_after> <alive>
run_case() {
    : > "$work/log"
    echo 0 > "$work/polls"
    (
        LOGFILE=$work/log
        WORK=$work
        EXIT_AFTER=$1
        ALIVE=${2:-1}
        export LOGFILE WORK EXIT_AFTER ALIVE
        . "$work/harness.sh"
        . "$work/func.sh"
        stop_process NanoKVM-Server
    ) > "$work/out" 2>&1
}

# --- a server that stops when it is asked ---
run_case 2

grep -q -- '-INT' "$work/log" \
    && note "a running server is asked politely first" OK \
    || note "a running server is asked politely first" FAIL

if grep -q -- '-KILL' "$work/log"; then
    note "a server that stops is not forced" FAIL
    sed 's/^/    /' "$work/log"
else
    note "a server that stops is not forced" OK
fi

polls=$(cat "$work/polls")
[ "$polls" -ge 2 ] \
    && note "it waits for the process to be gone ($polls poll(s))" OK \
    || note "it returns after $polls poll(s), so it does not wait" FAIL

# --- a server whose teardown blocks ---
run_case 999

# The order is the point: a KILL that arrives first gives the server no chance
# to release VI, which is the state this whole sequence protects.
order=$(grep -o -- '-INT\|-TERM\|-KILL' "$work/log" | tr '\n' ' ')
[ "$order" = "-INT -TERM -KILL " ] \
    && note "it escalates INT then TERM then KILL" OK \
    || note "it signals in the order: ${order:-nothing}" FAIL

polls=$(cat "$work/polls")
# 20 + 5 + 2 is the budget the three waits describe. One extra poll per wait is
# the loop noticing the deadline, so 30 is the ceiling, not 27.
[ "$polls" -le 30 ] \
    && note "it gives up after the timeouts ($polls poll(s))" OK \
    || note "it waited $polls poll(s), so the wait is unbounded" FAIL

# --- nothing is running ---
run_case 0 0

[ -s "$work/log" ] \
    && note "a server that is not running is signalled anyway" FAIL \
    || note "a server that is not running is left alone" OK

echo
echo "===== stop_services tears down in the right order ====="

body=$(sed -n '/^stop_services() {/,/^}/p' "$S95")
line_of() { printf '%s\n' "$2" | grep -n -- "$1" | head -1 | cut -d: -f1; }

server_at=$(line_of '^ *stop_process NanoKVM-Server$' "$body")
system_at=$(line_of '^ *stop_process kvm_system$' "$body")
# The removal used to name /tmp directly. It is guarded now, and it names
# SERVER_DST and SYSTEM_DST, so match the removal itself rather than the path it
# used to carry. The leading [^#]* keeps a comment that mentions rm out of it.
rm_at=$(line_of '^[^#]*rm -rf' "$body")

if [ -n "$server_at" ] && [ -n "$system_at" ] && [ "$server_at" -lt "$system_at" ]; then
    note "the server stops before kvm_system, so it can release MMF" OK
else
    note "kvm_system stops first, or one of them not at all" FAIL
fi

if [ -n "$rm_at" ] && [ -n "$server_at" ] && [ "$rm_at" -gt "$server_at" ]; then
    note "the staged copies are removed after the server has gone" OK
else
    note "the staged copies are removed while the server may still run" FAIL
fi

echo
echo "===== restart stops before it starts ====="

# The one that matters: nothing is staged until the old server has gone.
arm=$(sed -n '/^  restart)/,/^   ;;/p' "$S95")
stop_at=$(line_of '^ *stop_services$' "$arm")
start_at=$(line_of '^ *start_services$' "$arm")

if [ -n "$stop_at" ] && [ -n "$start_at" ] && [ "$stop_at" -lt "$start_at" ]; then
    note "restart) stages the new server only after the old one is gone" OK
else
    note "restart) stages the new server while the old one may still run" FAIL
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
fi
exit "$fails"
