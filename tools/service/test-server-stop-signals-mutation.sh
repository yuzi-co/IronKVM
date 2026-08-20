#!/bin/sh
# Prove that test-server-stop-signals.sh fails when stop_process is wrong.
#
#   test-server-stop-signals-mutation.sh
#
# Not destructive: every mutation is made on a copy under mktemp, and the
# shipped script is only ever read.
#
# The escalation is six lines, and an order-of-signals assertion is easy to
# write in a way that passes whatever the script does. Running the cases against
# deliberately broken copies is what separates a guard from a comment.
#
# NOT-APPLIED is a failure, not a skip. A sed that matches nothing leaves the
# copy identical, the cases pass because there is nothing wrong with it, and the
# run reads as proof when it proved nothing.
DIR=$(dirname "$0")
S95=${1:-$DIR/../../kvmapp/system/init.d/S95nanokvm}
TEST=$DIR/test-server-stop-signals.sh
for f in "$S95" "$TEST"; do
    [ -f "$f" ] || { echo "missing: $f"; exit 2; }
done

pass=0
fail=0

try() {
    desc=$1
    shift

    d=$(mktemp -d)
    cp "$S95" "$d/S95nanokvm"
    "$@" "$d/S95nanokvm"

    if cmp -s "$d/S95nanokvm" "$S95"; then
        printf '  %-14s %s\n' NOT-APPLIED "$desc"
        fail=$((fail + 1))
        rm -rf "$d"
        return
    fi

    if sh "$TEST" "$d/S95nanokvm" > /dev/null 2>&1; then
        printf '  %-14s %s\n' SURVIVED "$desc"
        fail=$((fail + 1))
    else
        printf '  %-14s %s\n' caught "$desc"
        pass=$((pass + 1))
    fi
    rm -rf "$d"
}

# The defect as it shipped. SIGINT is SIG_IGN in both processes, so this spends
# the whole first wait on a signal that cannot arrive.
m_sigint_first() {
    sed -i 's@^    killall -TERM "\$process"@    killall -INT "$process"@' "$1"
}

# The wait drops back to the server's own teardown bound. A teardown that was
# going to finish then races the script and loses, and the loser is SIGKILLed.
m_wait_equals_dispose() {
    sed -i 's@^    if wait_for_exit "\$process" 10; then@    if wait_for_exit "$process" 5; then@' "$1"
}

# Shorter still: no room for any teardown at all.
m_wait_too_short() {
    sed -i 's@^    if wait_for_exit "\$process" 10; then@    if wait_for_exit "$process" 1; then@' "$1"
}

# A wait with no practical end. The board does not come back from a stop that
# waits ten minutes on a process wedged in uninterruptible sleep.
m_wait_unbounded() {
    sed -i 's@^    if wait_for_exit "\$process" 10; then@    if wait_for_exit "$process" 600; then@' "$1"
}

# Straight to the signal that reaches no code. Every stop then leaks a video
# buffer pool, including the stops the process would have answered.
m_kill_first() {
    sed -i 's@^    killall -TERM "\$process"@    killall -KILL "$process"@' "$1"
}

# The escalation removed. A process that answers nothing is never stopped, and
# the start that follows finds the old one still holding the port.
m_no_kill() {
    sed -i '\@^    killall -KILL "\$process"@d' "$1"
}

# The early return removed, so a stop signals a process that is not there. On
# this board killall matches by name, and the name outlives the process it used
# to mean.
m_no_early_return() {
    sed -i 's@^    if ! pidof "\$process" >/dev/null 2>&1; then@    if false; then@' "$1"
}

echo "===== every mutation must be caught ====="
try "SIGINT comes back as the first signal" m_sigint_first
try "the wait equals disposeTimeout"        m_wait_equals_dispose
try "the wait is shorter than a teardown"   m_wait_too_short
try "the wait is effectively unbounded"     m_wait_unbounded
try "SIGKILL becomes the first signal"      m_kill_first
try "the escalation to SIGKILL is removed"  m_no_kill
try "a stopped process is signalled anyway" m_no_early_return

echo
if [ "$fail" -eq 0 ]; then
    echo "===== all $pass mutations caught ====="
else
    echo "===== $fail of $((pass + fail)) mutations were not caught ====="
    exit 1
fi
