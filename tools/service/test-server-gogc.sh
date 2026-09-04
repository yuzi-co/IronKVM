#!/bin/sh
# Check that S95nanokvm gives NanoKVM-Server a collector target before it
# starts.
#
#   test-server-gogc.sh [path-to-S95nanokvm]
#
# GOMEMLIMIT, covered by test-server-memlimit.sh, tells the collector where the
# ceiling is. GOGC tells it how hard to work below that ceiling, and left unset
# it is 100: a collection every time the heap grows by the size of the live
# heap.
#
# On the MJPEG path the capture loop hands over a freshly allocated slice per
# frame, about 150KiB thirty times a second. That allocation cannot be pooled,
# because the slice outlives the loop iteration in each client's FrameSlot and
# in the screenshot cache. So a small live heap is passed several times a
# second and the collector runs about twice a second.
#
# Measured on the reference device on 2026-09-04, with that frame size and rate
# and two client slots holding each frame, as a percentage of the one core:
#
#   live heap   GOGC=100   GOGC=150   GOGC=200   GOGC=400
#   2MiB          4.3%       3.2%       2.6%       1.7%
#   4MiB          5.3%       3.8%       3.1%         -
#   8MiB          5.4%       3.5%       2.8%       2.3%
#
# The peak heap follows the ratio, at about (1 + GOGC/100) times the live heap,
# a model that matched the measurement at every step.
#
# The server peaks at 27.8MB of RSS over 26 hours, so its live heap is at the
# small end of that table. 150 buys back about 1.1 to 1.9 points of the one
# core for half the live heap again. 200 and 400 were measured and neither was
# taken: 200 is another 0.7 points for another 2 to 3MiB, and 400 is 10 to
# 24MiB beyond that, on a board that stops being able to start a process below
# about 30MB free.
#
# The H.264 paths are unaffected. Their frames are a few KiB, and the same
# sweep records no collection at all above GOGC=100.
set -u

S95=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S95nanokvm}
[ -f "$S95" ] || { echo "usage: test-server-gogc.sh <S95nanokvm>"; exit 1; }

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

echo "===== the server starts under a collector target ====="

# Run the shipped function, not a copy of it.
sed -n '/^server_gogc() {/,/^}/p' "$S95" > "$work/f.sh"
if [ ! -s "$work/f.sh" ]; then
    note "S95nanokvm defines server_gogc" FAIL
    echo
    echo "$fails case(s) FAILED"
    exit 1
fi
note "S95nanokvm defines server_gogc" OK

run() {
    # Prints the value as a CHILD process sees it. The server is started as a
    # child, so a value assigned but never exported reaches nothing, and
    # reading it back in the same shell would pass either way.
    (
        MEMLIMIT_DIR=$1
        export MEMLIMIT_DIR
        . "$work/f.sh"
        server_gogc
        sh -c 'echo "${GOGC:-unset}"'
    )
}

mkdir -p "$work/empty"
got=$(run "$work/empty")
[ "$got" = "150" ] && note "with no override the target is 150" OK \
                   || note "with no override the target is 150 (got '$got')" FAIL

mkdir -p "$work/set"
echo 300 > "$work/set/GOGC.server"
got=$(run "$work/set")
[ "$got" = "300" ] && note "an override file sets the target" OK \
                   || note "an override file sets the target (got '$got')" FAIL

# Go refuses to start on a GOGC it cannot parse, so a typo in that file would
# take the server down rather than merely mis-tune it.
for bad in "" "  " "150%" "abc" "-5" "12 34" "1e3"; do
    mkdir -p "$work/bad"
    printf '%s\n' "$bad" > "$work/bad/GOGC.server"
    got=$(run "$work/bad")
    [ "$got" = "150" ] && note "a junk override ('$bad') falls back to the default" OK \
                       || note "a junk override ('$bad') falls back to the default (got '$got')" FAIL
    rm -rf "$work/bad"
done

# "off" parses in Go and is the one value that must not get through. It removes
# the ratio entirely and leaves the memory limit as the only trigger, which is
# the configuration that collects hardest exactly when the board can least
# afford it.
mkdir -p "$work/off"
echo off > "$work/off/GOGC.server"
got=$(run "$work/off")
[ "$got" = "150" ] && note "'off' is refused and falls back to the default" OK \
                   || note "'off' is refused and falls back to the default (got '$got')" FAIL

echo "===== it reaches the server ====="

# Set before the exec line, or it does nothing at all.
gogc_line=$(grep -n 'server_gogc' "$S95" | grep -v '^\s*#' | grep -v 'server_gogc() {' | head -1 | cut -d: -f1)
exec_line=$(grep -n '"\$SERVER_DST/NanoKVM-Server"' "$S95" | head -1 | cut -d: -f1)
if [ -n "$gogc_line" ] && [ -n "$exec_line" ] && [ "$gogc_line" -lt "$exec_line" ]; then
    note "the target is set before the server is started" OK
else
    note "the target is set before the server is started (call=$gogc_line exec=$exec_line)" FAIL
fi

echo "===== the ceiling is still there ====="

# This is the load-bearing pair. GOGC is a ratio with no upper bound of its
# own: a live heap that grows takes the trigger with it. GOMEMLIMIT is what
# stops that, and raising GOGC without it would turn a memory fault this board
# cannot recover from into a likelier one.
grep -q '^server_memlimit() {' "$S95" \
    && note "server_memlimit is still defined" OK \
    || note "server_memlimit is still defined" FAIL

memlimit_line=$(grep -n 'server_memlimit' "$S95" | grep -v '^\s*#' | grep -v 'server_memlimit() {' | head -1 | cut -d: -f1)
if [ -n "$memlimit_line" ] && [ -n "$exec_line" ] && [ "$memlimit_line" -lt "$exec_line" ]; then
    note "the limit is still set before the server is started" OK
else
    note "the limit is still set before the server is started" FAIL
fi

# Both in the same branch, so neither can be moved out from under the other by
# an edit that only looks at one of them.
if [ -n "$gogc_line" ] && [ -n "$memlimit_line" ]; then
    gap=$((gogc_line - memlimit_line))
    [ "$gap" -gt -6 ] && [ "$gap" -lt 6 ] \
        && note "the two are set together" OK \
        || note "the two are $gap lines apart, wanted them together" FAIL
else
    note "the two are set together" FAIL
fi

# The two override files are separate names on purpose, so an operator raising
# one cannot silently disable the other.
grep -q 'GOGC.server' "$S95" \
    && note "the override file is GOGC.server" OK \
    || note "the override file is GOGC.server" FAIL

grep -q 'GOMEMLIMIT.server' "$S95" \
    && note "the limit override file is still GOMEMLIMIT.server" OK \
    || note "the limit override file is still GOMEMLIMIT.server" FAIL

echo "===== what the operator is told ====="

# The startup line names both, because a board behaving oddly is diagnosed from
# that log and a value that is set but never printed is a value nobody checks.
grep -q 'GOMEMLIMIT set to .*GOGC set to' "$S95" \
    && note "the startup line reports both values" OK \
    || note "the startup line reports both values" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
