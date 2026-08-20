#!/bin/sh
# Check that S95nanokvm gives NanoKVM-Server a Go heap limit before it starts.
#
#   test-server-memlimit.sh [path-to-S95nanokvm]
#
# The board has 166MB that Linux can see, and about 96MB of that is available
# once the video carveout, the kernel and the running services are counted.
# NanoKVM-Server peaked at 27.8MB over 26 hours on the reference device, read
# from /proc/<pid>/status VmHWM.
#
# Go grows the heap until the collector decides to run, and the collector's
# only input by default is how much the heap has grown. It does not know what
# the board has left. Below about 30MB free this board stops being able to
# start a process at all: fork fails, the kernel does not choose a victim, and
# nothing recovers it except a power cycle. So the failure this guards against
# is not a killed server, it is a board that answers nothing and cannot be
# reached to fix it.
#
# GOMEMLIMIT gives the collector the missing input. The default here is 64MiB,
# a little over twice the measured peak, which leaves the collector room to
# work before it becomes the thing consuming the board.
#
# tailscaled already gets this treatment at S98tailscaled:46-53. The server,
# the process that actually has to answer, did not.
set -u

S95=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S95nanokvm}
[ -f "$S95" ] || { echo "usage: test-server-memlimit.sh <S95nanokvm>"; exit 1; }

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

echo "===== the server starts under a Go heap limit ====="

# Run the shipped function, not a copy of it.
sed -n '/^server_memlimit() {/,/^}/p' "$S95" > "$work/f.sh"
if [ ! -s "$work/f.sh" ]; then
    note "S95nanokvm defines server_memlimit" FAIL
    echo
    echo "$fails case(s) FAILED"
    exit 1
fi
note "S95nanokvm defines server_memlimit" OK

run() {
    # $1 is the config directory to read from. Prints the value as a CHILD
    # process sees it, because that is the only thing that matters: the server
    # is started as a child, so a value that is assigned but never exported
    # reaches nothing. Reading it back in the same shell would pass either way.
    (
        MEMLIMIT_DIR=$1
        export MEMLIMIT_DIR
        . "$work/f.sh"
        server_memlimit
        sh -c 'echo "${GOMEMLIMIT:-unset}"'
    )
}

# With no override file, the compiled-in default applies.
mkdir -p "$work/empty"
got=$(run "$work/empty")
[ "$got" = "64MiB" ] && note "with no override the limit is 64MiB" OK \
                     || note "with no override the limit is 64MiB (got '$got')" FAIL

# An operator can raise or lower it with a file holding a number of MiB.
mkdir -p "$work/set"
echo 96 > "$work/set/GOMEMLIMIT.server"
got=$(run "$work/set")
[ "$got" = "96MiB" ] && note "an override file sets the limit" OK \
                     || note "an override file sets the limit (got '$got')" FAIL

# A file that does not hold a plain number must not produce a malformed value.
# Go refuses to start when GOMEMLIMIT does not parse, so a typo would take the
# server down rather than merely mis-size it.
for bad in "" "  " "512MiB" "abc" "-5" "12 34"; do
    mkdir -p "$work/bad"
    printf '%s\n' "$bad" > "$work/bad/GOMEMLIMIT.server"
    got=$(run "$work/bad")
    [ "$got" = "64MiB" ] && note "a junk override ('$bad') falls back to the default" OK \
                         || note "a junk override ('$bad') falls back to the default (got '$got')" FAIL
    rm -rf "$work/bad"
done

# The limit has to be in the environment before the server is executed, not
# after. A value exported below the exec line does nothing at all.
limit_line=$(grep -n 'server_memlimit' "$S95" | grep -v '^\s*#' | grep -v 'server_memlimit() {' | head -1 | cut -d: -f1)
exec_line=$(grep -n '"\$SERVER_DST/NanoKVM-Server"' "$S95" | head -1 | cut -d: -f1)
if [ -n "$limit_line" ] && [ -n "$exec_line" ] && [ "$limit_line" -lt "$exec_line" ]; then
    note "the limit is set before the server is started" OK
else
    note "the limit is set before the server is started (call=$limit_line exec=$exec_line)" FAIL
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
